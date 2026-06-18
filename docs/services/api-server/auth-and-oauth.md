# Auth And OAuth

Parent index: [API Server](!README.md)

## Purpose

This document describes the current API-server auth and OAuth implementation in `services/api-server/`.

It documents Rails-owned account authentication, Discord OAuth, opaque bearer-token handling, current-user lookup, and the internal token-verification boundary consumed by the Go game-server.

## Overview

The API server is the Rails-owned auth boundary for Space Rocks authenticated accounts.

It currently supports two user-facing login paths:

* email/password registration and login
* Discord OAuth login

Both paths issue the same kind of Space Rocks access token: an opaque bearer token backed by an `access_tokens` database row. The raw token is returned to the caller only when it is issued. Rails stores only a SHA-256 digest of the token, tracks expiry and revocation, and updates `last_used_at` when a token is verified.

The API server also exposes current-user and logout endpoints for clients that already hold a bearer token. `GET /api/auth/me` verifies the bearer token and returns the authenticated user, including the cross-system `account_id`. `DELETE /api/auth/logout` revokes only the bearer token used for that request.

Discord OAuth has two implemented entry paths:

* direct browser OAuth through `/api/auth/discord/start` and `/api/auth/discord/callback`
* browser-assisted Godot handoff through Discord login sessions

The Godot handoff creates an `OauthLoginSession`, returns a public login-session id, a one-time poll secret, a browser login URL, and an expiry. The browser completes Discord OAuth through Rails. The callback either returns the normal auth response or authenticates the login session and returns a short handoff message. Godot then exchanges the login session plus poll secret for the normal auth response.

The Go game-server does not read Rails auth tables directly. When it needs to verify a Space Rocks bearer token, it calls the Rails internal endpoint `POST /internal/auth/verify-token` using the configured `GAME_SERVER_INTERNAL_TOKEN`. Rails returns a minimal authenticated-account identity for valid user tokens and `{ "valid": false }` for invalid, missing, expired, or revoked user tokens.

## Code root

* `services/api-server/`

## Responsibilities

* Own Rails authenticated account records.
* Own email/password credentials.
* Normalize password-credential email addresses before validation.
* Hash and verify passwords through Rails `has_secure_password`.
* Own provider identities for OAuth login.
* Own Discord OAuth state creation, verification, consumption, and expiry.
* Own Discord token exchange and current-user profile fetch.
* Resolve OAuth provider profiles into existing or newly created Rails users.
* Own Discord login-session handoff records used by the Godot browser-assisted OAuth flow.
* Issue opaque Space Rocks bearer access tokens.
* Store bearer-token digests instead of raw bearer tokens.
* Reject revoked, expired, unknown, missing, or malformed bearer tokens.
* Revoke the current token on logout.
* Expose the current authenticated user through `GET /api/auth/me`.
* Verify bearer tokens for internal service consumers through `POST /internal/auth/verify-token`.
* Require the internal service bearer token before serving internal auth verification.
* Keep the API auth physical schema owned by Rails migrations and Rails schema files.
* Keep HTTP request and response shapes aligned with `shared/contracts/http/openapi.yaml`.

## Does not own

* Go game-server real-time simulation.
* Go game-server room lifecycle, websocket session lifecycle, matchmaking admission, or gameplay identity assignment.
* Direct game-server access to Rails auth tables.
* Client-side login UI presentation.
* Client-side token persistence.
* Local single-player auth requirements.
* Local Profile persistence.
* Embedded SQLite Local Profile storage.
* Player-data route selection across guest, local profile, and authenticated account.
* Discord account authority beyond the provider profile returned during OAuth.
* Persistence of Discord access tokens, OAuth client secrets, OAuth refresh tokens, provider secrets, raw OAuth state, or raw poll secrets.
* JWT issuance. Current user tokens are opaque bearer tokens.

## Domain roles

### Authenticated Account

An authenticated account is represented by a Rails `User`.

`users.account_id` is the canonical cross-system authenticated-account UUID. Rails `users.id` remains an internal database id and foreign key. API responses expose both shapes in some contexts:

* normal auth responses currently return `id`, `display_name`, and `email`
* `GET /api/auth/me` returns `id`, `account_id`, `display_name`, and `email`
* internal token verification returns `id`, `account_id`, and `display_name`

### Password credential

`PasswordCredential` owns email/password login material for a `User`.

The model normalizes email addresses by trimming whitespace and lowercasing before validation. The password digest is managed through `has_secure_password`.

### Provider identity

`UserIdentity` links a provider identity to a Rails user.

The current implemented provider is Discord. The identity key is the pair of provider and provider UID. Repeated OAuth login for the same Discord provider UID reuses the existing Rails user.

### Access token

`AccessToken` is the Rails-backed server-side record for a Space Rocks bearer token.

Raw tokens are generated as random hex strings, returned to the caller, and never stored directly. The database stores `token_digest`, `audience`, `expires_at`, `revoked_at`, and `last_used_at`.

### OAuth state

`OauthState` protects the Discord callback.

The raw state is generated per OAuth start, returned only inside the provider redirect URL, and stored by digest. OAuth state is provider-scoped, expires, and is consumed during verification.

### OAuth login session

`OauthLoginSession` supports the Godot browser handoff.

It stores a public id, provider, poll-secret digest, status, expiry, optional authenticated user, and consumed timestamp. The raw poll secret is returned to the caller once and compared by digest during exchange.

### Internal verification

Internal verification is the API-server auth boundary consumed by the Go game-server.

The internal endpoint authenticates the service caller with `GAME_SERVER_INTERNAL_TOKEN`, then verifies the user bearer token passed in the request body. Valid user tokens return a minimal authenticated-account identity. Invalid user tokens return `valid: false` without exposing token internals.

## Protocols and APIs

### Public auth endpoints

| Endpoint                                             | Purpose                                                     | Success behavior                                                          | Failure behavior                                                                                                    |
| ---------------------------------------------------- | ----------------------------------------------------------- | ------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------- |
| `POST /api/auth/register`                            | Create a Rails user and password credential.                | Returns auth response with bearer token and auth user.                    | Returns `422` with `invalid` when validation fails.                                                                 |
| `POST /api/auth/login`                               | Login with email/password.                                  | Returns auth response with bearer token and auth user.                    | Returns `401` with `invalid_credentials`.                                                                           |
| `DELETE /api/auth/logout`                            | Revoke the current bearer token.                            | Returns `204`.                                                            | Returns `401` for missing, malformed, expired, revoked, or unknown bearer token.                                    |
| `GET /api/auth/me`                                   | Return the current authenticated user.                      | Returns account user including `account_id`.                              | Returns `401` with `invalid_token`.                                                                                 |
| `GET /api/auth/discord/start`                        | Create OAuth state and redirect to Discord.                 | Redirects to Discord authorization URL.                                   | Provider config errors raise through normal Rails error handling.                                                   |
| `GET /api/auth/discord/callback`                     | Complete Discord OAuth.                                     | Returns the normal auth response, or returns `{ "message": "You can return to the game." }` after authenticating a login session. | Returns `400`, `422`, or `502` for missing params, invalid state, token exchange failure, or profile fetch failure. |
| `POST /api/auth/discord/login_sessions`              | Create Godot browser-assisted Discord login session.        | Returns `login_session_id`, `poll_secret`, `login_url`, and `expires_at`. | Provider config errors raise through normal Rails error handling.                                                   |
| `POST /api/auth/discord/login_sessions/:id/exchange` | Exchange an authenticated login session for a bearer token. | Returns auth response when authenticated.                                 | Returns `202` while pending, `400` for missing poll secret, or `422` for invalid session.                           |

### Internal auth endpoint

| Endpoint                           | Caller                               | Purpose                            | Success behavior                                  | Failure behavior                                                                                            |
| ---------------------------------- | ------------------------------------ | ---------------------------------- | ------------------------------------------------- | ----------------------------------------------------------------------------------------------------------- |
| `POST /internal/auth/verify-token` | Go game-server internal auth client. | Verify a Space Rocks bearer token. | Returns `valid: true` plus minimal user identity. | Returns `401` for invalid internal service auth; returns `200` with `valid: false` for invalid user tokens. |

The internal endpoint uses two different bearer concepts:

```text
Authorization: Bearer <GAME_SERVER_INTERNAL_TOKEN>
```

authenticates the internal service caller, while:

```json
{ "token": "<user access token>" }
```

is the Space Rocks user token being verified.

### HTTP contract source

`shared/contracts/http/openapi.yaml` defines the request and response shapes for the auth endpoints.

Rails controller tests use OpenAPI contract assertions. This is test-time contract enforcement, not runtime OpenAPI middleware.

### Environment

Discord OAuth and internal service verification depend on these environment variables:

* `DISCORD_CLIENT_ID`
* `DISCORD_CLIENT_SECRET`
* `DISCORD_REDIRECT_URI`
* `GAME_SERVER_INTERNAL_TOKEN`

`GAME_SERVER_INTERNAL_TOKEN` must be present for internal endpoints to authorize service callers.

## Data ownership

Rails migrations and `services/api-server/db/schema.rb` own the API auth physical database schema.

Current auth-owned tables:

* `users`
* `password_credentials`
* `user_identities`
* `access_tokens`
* `oauth_states`
* `oauth_login_sessions`

Auth-adjacent authenticated-account player data tables exist in the same Rails service, but they are not owned by this auth document:

* `player_stats`
* `player_match_results`

The auth implementation persists:

* Rails user display name
* Rails user `account_id`
* password credential email
* password digest
* provider identity provider name
* provider identity UID
* provider identity email when available
* access-token digest
* access-token audience
* access-token expiry, revocation, and last-use timestamps
* OAuth state digest and expiry
* OAuth login-session public id, provider, poll-secret digest, status, expiry, authenticated user reference, and consumed timestamp

The auth implementation does not persist:

* raw bearer tokens
* raw OAuth state
* raw login-session poll secrets
* Discord access tokens
* Discord refresh tokens
* Discord client secrets
* OAuth provider secrets

## Security and failure behavior

Bearer token verification requires a syntactically valid `Authorization: Bearer <token>` header on public bearer-protected endpoints.

Public bearer-protected endpoints return `401` for:

* missing bearer token
* malformed bearer header
* unknown token
* revoked token
* expired token

`Auth::VerifyAccessToken` updates `last_used_at` when a token is valid.

`DELETE /api/auth/logout` revokes only the access token used for that request. Other active tokens for the same user remain active.

Internal auth verification first validates the internal service bearer token. The internal service token comparison checks length before secure comparison and rejects requests when the environment token is unset.

The internal verify endpoint intentionally returns `200` with `valid: false` for invalid user tokens after the internal caller is authenticated. This lets the game-server distinguish service-auth failure from user-token failure.

Discord callback state is single-use. `Auth::OauthStateVerifier` consumes usable state during verification and rejects expired, consumed, missing, provider-mismatched, or unknown state.

OAuth login sessions reject expired, consumed, missing, or wrong-poll-secret exchanges. Pending sessions return `202` with `status: "pending"`.

## Code map

### Routing and contracts

* `services/api-server/config/routes.rb`
* `shared/contracts/http/openapi.yaml`

### Controllers

* `services/api-server/app/controllers/api/auth/registrations_controller.rb`
* `services/api-server/app/controllers/api/auth/sessions_controller.rb`
* `services/api-server/app/controllers/api/auth/me_controller.rb`
* `services/api-server/app/controllers/api/auth/discord_controller.rb`
* `services/api-server/app/controllers/api/auth/discord_login_sessions_controller.rb`
* `services/api-server/app/controllers/internal/base_controller.rb`
* `services/api-server/app/controllers/internal/auth/verify_tokens_controller.rb`

### Controller concerns

* `services/api-server/app/controllers/concerns/authenticates_bearer_token.rb`
* `services/api-server/app/controllers/concerns/renders_auth_response.rb`

### Auth services

* `services/api-server/app/services/auth/result.rb`
* `services/api-server/app/services/auth/register_user.rb`
* `services/api-server/app/services/auth/login_user.rb`
* `services/api-server/app/services/auth/issue_access_token.rb`
* `services/api-server/app/services/auth/verify_access_token.rb`
* `services/api-server/app/services/auth/oauth_state_issuer.rb`
* `services/api-server/app/services/auth/oauth_state_verifier.rb`
* `services/api-server/app/services/auth/oauth_login_user.rb`
* `services/api-server/app/services/auth/oauth_resolve_user.rb`
* `services/api-server/app/services/auth/oauth_login_session_issuer.rb`

### Discord provider services

* `services/api-server/app/services/auth/providers/discord_config.rb`
* `services/api-server/app/services/auth/providers/discord_authorization_url.rb`
* `services/api-server/app/services/auth/providers/discord_token_exchange.rb`
* `services/api-server/app/services/auth/providers/discord_current_user.rb`
* `services/api-server/app/services/auth/providers/provider_profile.rb`

### Models and schema

* `services/api-server/app/models/user.rb`
* `services/api-server/app/models/password_credential.rb`
* `services/api-server/app/models/user_identity.rb`
* `services/api-server/app/models/access_token.rb`
* `services/api-server/app/models/oauth_state.rb`
* `services/api-server/app/models/oauth_login_session.rb`
* `services/api-server/db/migrate/20260608000100_create_users.rb`
* `services/api-server/db/migrate/20260608000200_create_password_credentials.rb`
* `services/api-server/db/migrate/20260608000300_create_user_identities.rb`
* `services/api-server/db/migrate/20260608000400_create_access_tokens.rb`
* `services/api-server/db/migrate/20260608000500_create_oauth_states.rb`
* `services/api-server/db/migrate/20260608000600_create_oauth_login_sessions.rb`
* `services/api-server/db/migrate/20260608000700_add_oauth_login_session_to_oauth_states.rb`
* `services/api-server/db/migrate/20260608001000_add_account_id_to_users.rb`
* `services/api-server/db/schema.rb`

### Non-owning consumers

* `services/game-server/internal/authclient/client.go`
* `services/game-server/internal/authclient/types.go`

The Go auth client consumes the internal verification API. It does not own Rails auth persistence or read Rails tables.

## Tests

### Controller tests

* `services/api-server/test/controllers/api/auth/registrations_controller_test.rb`
* `services/api-server/test/controllers/api/auth/sessions_controller_test.rb`
* `services/api-server/test/controllers/api/auth/me_controller_test.rb`
* `services/api-server/test/controllers/api/auth/discord_controller_test.rb`
* `services/api-server/test/controllers/api/auth/discord_login_sessions_controller_test.rb`
* `services/api-server/test/controllers/internal/auth/verify_tokens_controller_test.rb`

### Service tests

* `services/api-server/test/services/auth/verify_access_token_test.rb`
* `services/api-server/test/services/auth/oauth_state_issuer_test.rb`
* `services/api-server/test/services/auth/oauth_state_verifier_test.rb`
* `services/api-server/test/services/auth/oauth_login_user_test.rb`
* `services/api-server/test/services/auth/oauth_login_session_issuer_test.rb`
* `services/api-server/test/services/auth/providers/discord_token_exchange_test.rb`
* `services/api-server/test/services/auth/providers/discord_current_user_test.rb`

### Model tests

* `services/api-server/test/models/user_test.rb`
* `services/api-server/test/models/password_credential_test.rb`
* `services/api-server/test/models/access_token_test.rb`
* `services/api-server/test/models/oauth_state_test.rb`
* `services/api-server/test/models/oauth_login_session_test.rb`

### Contract tests

* `services/api-server/test/contracts/openapi_contract_test.rb`
* `services/api-server/test/support/openapi_contract_assertions.rb`

## Related docs

* [API Server](!README.md)
* [Game Server](../game-server/!README.md)
* [Player Data](../player-data/!README.md)
* [HTTP contract enforcement](../../protocol/http-contract-enforcement.md) - Current HTTP request/response contract enforcement documentation.
* [API-server internal API surface](internal-api-surface.md)
* [API-server player stats and match results](player-stats-and-match-results.md)
* [Documentation policy](../../documentation-policy.md)
* [Documentation procedure](../../documentation-procedure.md)

## Notes

The legacy docs identified the correct high-level boundary: Rails owns authenticated accounts, OAuth identities, and bearer-token verification, while the Go game-server consumes that boundary through explicit API calls. This document rewrites those facts from current code instead of treating the legacy files as current authority.

Direct Discord browser callback returns the normal JSON auth response. The Godot-oriented browser handoff returns a short JSON message after authenticating the login session, then the client receives the normal Space Rocks bearer token after exchange.

`AuthResponse` does not currently include `account_id`; `GET /api/auth/me` and internal verification do include it.

JWT remains deferred. Current access tokens are opaque bearer tokens.
