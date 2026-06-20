# Local Auth Smoke Flow

Parent index: [API Server](./!INDEX.md)

## Purpose

This document describes the local development smoke flow for API-server authentication.

It covers the development-only verification path for health, email/password auth, bearer-token current-user lookup, logout, direct Discord OAuth start, and the browser-assisted Discord login-session handoff used by the Godot client.

## Overview

The local auth smoke flow exercises real API-server runtime endpoints. It is not a separate debug auth system and does not bypass Rails auth authority.

The smoke flow has three useful paths:

```text
Email/password API smoke
= health -> register or login -> me -> logout -> me should fail

Direct Discord browser smoke
= discord/start -> Discord redirect -> callback -> auth response

Godot Discord handoff smoke
= login_sessions create -> browser login_url -> callback authenticates session -> exchange returns bearer token
```

Email/password smoke is the fastest check for local bearer-token lifecycle. Discord smoke verifies provider configuration, OAuth state handling, browser redirect behavior, provider profile resolution, login-session handoff, and token exchange.

Both email/password auth and Discord OAuth issue the same Space Rocks opaque bearer token shape. Rails stores only the token digest, returns the raw token once on issue, and verifies later bearer requests through `GET /api/auth/me`.

## Debug-only scope

The smoke flow is a local development procedure. It observes and drives real API-server endpoints with normal HTTP clients such as Bruno, curl, a browser, or the Godot client.

It does not create debug-only auth routes, seed privileged users, skip password validation, bypass OAuth state checks, bypass login-session poll-secret checks, expose token digests, or give the client access to Discord secrets.

The only devtools-specific part is the procedure and tooling used to exercise the surface locally. Auth authority remains in `services/api-server/`.

## Server authority

The API server owns:

* Rails users.
* Password credentials.
* Discord provider identities.
* OAuth state records.
* OAuth login-session records.
* Opaque Space Rocks bearer-token issue, digest storage, verification, expiry, and revocation.
* `GET /api/auth/me` current-user lookup.
* `DELETE /api/auth/logout` token revocation.
* `POST /internal/auth/verify-token` internal token verification for the Go game-server.

Smoke tools only call the API surface. They do not own authentication state.

The direct Discord browser path returns the normal auth response from the callback when the OAuth state is not attached to a login session. The Godot handoff path attaches the OAuth state to an `OauthLoginSession`; the callback authenticates that session and returns a short browser message, then the client exchanges `login_session_id` plus `poll_secret` for the normal auth response.

## Client presentation

Bruno and curl present raw HTTP status codes, headers, and response bodies. They are useful for verifying that the API-server auth surface works before involving the Godot UI.

The browser presents the Discord login and authorization flow. After a successful Godot handoff callback, Rails returns:

```json
{ "message": "You can return to the game." }
```

The Godot client presents signed-in state after it exchanges the login session for a Space Rocks bearer token, stores the token locally, validates it through `GET /api/auth/me`, and updates the main menu.

Current Godot sign-in presentation is Discord-oriented. Email/password endpoints exist at the API level and are useful for smoke testing, but the current client sign-in UI does not expose manual email/password sign-in as the normal player-facing path.

## Commands or controls

Run the API server locally before using the smoke flow:

```bash
cd services/api-server
set -a && source ../../.env && set +a
bundle exec rails db:migrate
bundle exec rails server
```

For test setup, keep the test database prepared:

```bash
cd services/api-server
RAILS_ENV=test bundle exec rails db:test:prepare
bundle exec rails test
```

Discord OAuth also requires the local Discord environment to be loaded before Rails starts:

```text
DISCORD_CLIENT_ID
DISCORD_CLIENT_SECRET
DISCORD_REDIRECT_URI
GAME_SERVER_INTERNAL_TOKEN
```

Do not commit real Discord client secrets, OAuth codes, access tokens, Space Rocks bearer tokens, login-session poll secrets, or internal service tokens.

## Bruno smoke flow

The Bruno collection root is:

```text
bruno-api/
```

The local Bruno environment is:

```text
bruno-api/environments/local.yml
```

Important local variables:

| Variable       | Purpose                                                       |
| -------------- | ------------------------------------------------------------- |
| `base_url`     | Local API base URL, normally `http://localhost:3000`.         |
| `email`        | Email used for register/login smoke requests.                 |
| `password`     | Password used for register/login smoke requests.              |
| `display_name` | Display name used during register smoke.                      |
| `auth_token`   | Raw Space Rocks bearer token captured from register or login. |

Use this order for email/password auth smoke:

1. `health`
2. `register` or `login`
3. `me`
4. `logout`
5. `me` again

Expected behavior:

| Step               | Expected result                                                                      |
| ------------------ | ------------------------------------------------------------------------------------ |
| `health`           | `200` response from `GET /health`.                                                   |
| `register`         | `201` response with `token` and `user`; after-response script captures `auth_token`. |
| `login`            | `200` response with `token` and `user`; after-response script captures `auth_token`. |
| `me` before logout | `200` response with `user`, including `account_id`.                                  |
| `logout`           | `204` response and revocation of the current bearer token.                           |
| `me` after logout  | `401` response with `invalid_token`.                                                 |

`auth_token` must contain only the raw token value. Do not include the `Bearer ` prefix in the Bruno variable.

## Curl smoke flow

Use curl when Bruno is unavailable or when a small terminal-only check is faster.

Health:

```bash
curl -s http://localhost:3000/health
```

Register:

```bash
curl -s -X POST http://localhost:3000/api/auth/register \
  -H 'Content-Type: application/json' \
  -d '{"display_name":"Test Pilot","email":"test@example.com","password":"password123"}'
```

Login:

```bash
TOKEN=$(
  curl -s -X POST http://localhost:3000/api/auth/login \
    -H 'Content-Type: application/json' \
    -d '{"email":"test@example.com","password":"password123"}' \
  | jq -r '.token'
)
```

Verify current user:

```bash
curl -s http://localhost:3000/api/auth/me \
  -H "Authorization: Bearer $TOKEN"
```

Logout:

```bash
curl -i -s -X DELETE http://localhost:3000/api/auth/logout \
  -H "Authorization: Bearer $TOKEN"
```

Verify logout revocation:

```bash
curl -i -s http://localhost:3000/api/auth/me \
  -H "Authorization: Bearer $TOKEN"
```

After logout, the same token should no longer verify.

## Direct Discord OAuth smoke

The direct browser OAuth smoke starts with:

```text
GET /api/auth/discord/start
```

Expected behavior:

* Rails creates an OAuth state record.
* Rails redirects to Discord authorization.
* The redirect URL includes `client_id`, `redirect_uri`, `response_type=code`, `scope=identify email`, and `state`.
* The callback verifies the state, exchanges the Discord code, fetches the Discord profile, resolves or creates a Rails user, and returns the normal auth response.

In Bruno, use the `discord oauth start` request with redirects disabled so the Discord `Location` header can be inspected.

Do not store a real Discord authorization code or provider access token in Bruno files.

## Godot Discord login-session smoke

The Godot-oriented smoke starts with:

```text
POST /api/auth/discord/login_sessions
```

Expected response shape:

```json
{
  "login_session_id": "...",
  "poll_secret": "...",
  "login_url": "...",
  "expires_at": "..."
}
```

Smoke sequence:

```text
1. POST /api/auth/discord/login_sessions
2. Open login_url in a browser.
3. Complete Discord OAuth.
4. Browser reaches /api/auth/discord/callback.
5. Rails authenticates the login session.
6. POST /api/auth/discord/login_sessions/{login_session_id}/exchange with poll_secret.
7. Rails returns the normal auth response with token and user.
8. GET /api/auth/me verifies the returned token.
```

Pending exchange response:

```json
{ "status": "pending" }
```

A pending response uses HTTP `202`. The Godot client polls once per second for up to 120 seconds. A successful exchange returns HTTP `200` with the normal auth response and consumes the login session.

Invalid exchange cases return errors for missing poll secret, wrong poll secret, unknown session, expired session, consumed session, or a session that never authenticated.

## Internal verifier smoke

The public auth smoke is enough to verify local login. Use the internal verifier only when checking the API-server boundary consumed by the Go game-server.

The internal verifier uses two token concepts:

```text
Authorization: Bearer <GAME_SERVER_INTERNAL_TOKEN>
```

authenticates the internal service caller, while:

```json
{ "token": "<space-rocks-user-token>" }
```

is the user bearer token being verified.

Request shape:

```bash
curl -s -X POST http://localhost:3000/internal/auth/verify-token \
  -H "Authorization: Bearer $GAME_SERVER_INTERNAL_TOKEN" \
  -H 'Content-Type: application/json' \
  -d "{\"token\":\"$TOKEN\"}"
```

Expected valid response:

```json
{
  "valid": true,
  "user": {
    "id": 1,
    "account_id": "...",
    "display_name": "Test Pilot"
  }
}
```

Expected invalid user-token response:

```json
{ "valid": false }
```

Invalid internal service auth returns `401`.

## Telemetry and diagnostics

The smoke flow is diagnosed through HTTP responses, response bodies, redirect headers, Rails logs, and persisted database side effects.

Useful observations:

| Signal                                   | Meaning                                                                 |
| ---------------------------------------- | ----------------------------------------------------------------------- |
| `GET /health` returns `200`              | Rails app is booted and reachable.                                      |
| Register/login returns `token`           | Auth token issue path works.                                            |
| `GET /api/auth/me` returns `account_id`  | Bearer-token verification and current-user lookup work.                 |
| Logout returns `204`                     | Current token revocation path works.                                    |
| `me` after logout returns `401`          | Revoked token is rejected.                                              |
| Discord start returns redirect           | Discord config and OAuth state issue are reachable.                     |
| Login-session create returns `login_url` | Godot handoff session and OAuth state issue work.                       |
| Exchange returns `202`                   | Login session exists but browser callback has not authenticated it yet. |
| Exchange returns auth response           | Login session was authenticated and consumed.                           |
| Internal verifier returns `valid: true`  | Game-server-facing token verification boundary works.                   |

Common failures:

| Failure                                                    | Likely cause                                                                |
| ---------------------------------------------------------- | --------------------------------------------------------------------------- |
| Rails migration error                                      | API database is not migrated.                                               |
| Register returns `422 invalid`                             | Duplicate email or failed model validation.                                 |
| Login returns `401 invalid_credentials`                    | Email/password mismatch or missing password credential.                     |
| `me` returns `401 invalid_token`                           | Missing, malformed, unknown, expired, or revoked bearer token.              |
| Discord start raises config error                          | Discord environment variables were not loaded before Rails boot.            |
| Discord callback returns `400 missing_params`              | Callback did not include both `code` and `state`.                           |
| Discord callback returns `422 invalid_state`               | OAuth state is missing, expired, consumed, provider-mismatched, or unknown. |
| Discord callback returns `502`                             | Discord token exchange or current-user profile fetch failed.                |
| Login-session exchange returns `422 invalid_login_session` | Wrong poll secret, expired session, consumed session, or unknown public id. |

## Build/runtime gates

The local auth smoke depends on:

* Rails API server running on the configured `base_url`, normally `http://localhost:3000`.
* PostgreSQL database created and migrated.
* `services/api-server/config/routes.rb` exposing auth routes.
* `shared/contracts/http/openapi.yaml` matching implemented request and response shapes.
* Discord environment variables loaded before Rails starts when OAuth smoke is used.
* `GAME_SERVER_INTERNAL_TOKEN` loaded when the internal verifier smoke is used.
* No committed secrets in `bruno-api/`, `.env`, or documentation.

## Code map

Routing and contract source:

* `services/api-server/config/routes.rb`
* `shared/contracts/http/openapi.yaml`

Bruno collection:

* `bruno-api/opencollection.yml`
* `bruno-api/environments/local.yml`
* `bruno-api/api/health.yml`
* `bruno-api/api/register.yml`
* `bruno-api/api/login.yml`
* `bruno-api/api/me.yml`
* `bruno-api/api/logout.yml`
* `bruno-api/api/discord-oauth-start.yml`

Public auth controllers:

* `services/api-server/app/controllers/api/auth/registrations_controller.rb`
* `services/api-server/app/controllers/api/auth/sessions_controller.rb`
* `services/api-server/app/controllers/api/auth/me_controller.rb`
* `services/api-server/app/controllers/api/auth/discord_controller.rb`
* `services/api-server/app/controllers/api/auth/discord_login_sessions_controller.rb`

Controller concerns:

* `services/api-server/app/controllers/concerns/renders_auth_response.rb`
* `services/api-server/app/controllers/concerns/authenticates_bearer_token.rb`

Auth services:

* `services/api-server/app/services/auth/register_user.rb`
* `services/api-server/app/services/auth/login_user.rb`
* `services/api-server/app/services/auth/issue_access_token.rb`
* `services/api-server/app/services/auth/verify_access_token.rb`
* `services/api-server/app/services/auth/oauth_state_issuer.rb`
* `services/api-server/app/services/auth/oauth_state_verifier.rb`
* `services/api-server/app/services/auth/oauth_login_user.rb`
* `services/api-server/app/services/auth/oauth_resolve_user.rb`
* `services/api-server/app/services/auth/oauth_login_session_issuer.rb`

Discord provider services:

* `services/api-server/app/services/auth/providers/discord_config.rb`
* `services/api-server/app/services/auth/providers/discord_authorization_url.rb`
* `services/api-server/app/services/auth/providers/discord_token_exchange.rb`
* `services/api-server/app/services/auth/providers/discord_current_user.rb`
* `services/api-server/app/services/auth/providers/provider_profile.rb`

Auth models:

* `services/api-server/app/models/user.rb`
* `services/api-server/app/models/password_credential.rb`
* `services/api-server/app/models/user_identity.rb`
* `services/api-server/app/models/access_token.rb`
* `services/api-server/app/models/oauth_state.rb`
* `services/api-server/app/models/oauth_login_session.rb`

Internal verifier:

* `services/api-server/app/controllers/internal/base_controller.rb`
* `services/api-server/app/controllers/internal/auth/verify_tokens_controller.rb`

Client handoff consumers:

* `client/scripts/auth/auth_api_client.gd`
* `client/scripts/auth/auth_session_controller.gd`
* `client/scripts/auth/auth_token_store.gd`
* `client/scripts/auth/auth_session.gd`

## Tests

Controller tests:

* `services/api-server/test/controllers/api/auth/registrations_controller_test.rb`
* `services/api-server/test/controllers/api/auth/sessions_controller_test.rb`
* `services/api-server/test/controllers/api/auth/me_controller_test.rb`
* `services/api-server/test/controllers/api/auth/discord_controller_test.rb`
* `services/api-server/test/controllers/api/auth/discord_login_sessions_controller_test.rb`
* `services/api-server/test/controllers/internal/auth/verify_tokens_controller_test.rb`

Service tests:

* `services/api-server/test/services/auth/verify_access_token_test.rb`
* `services/api-server/test/services/auth/oauth_state_issuer_test.rb`
* `services/api-server/test/services/auth/oauth_state_verifier_test.rb`
* `services/api-server/test/services/auth/oauth_login_user_test.rb`
* `services/api-server/test/services/auth/oauth_login_session_issuer_test.rb`
* `services/api-server/test/services/auth/providers/discord_token_exchange_test.rb`
* `services/api-server/test/services/auth/providers/discord_current_user_test.rb`

Model tests:

* `services/api-server/test/models/user_test.rb`
* `services/api-server/test/models/password_credential_test.rb`
* `services/api-server/test/models/access_token_test.rb`
* `services/api-server/test/models/oauth_state_test.rb`
* `services/api-server/test/models/oauth_login_session_test.rb`

Contract tests:

* `services/api-server/test/contracts/openapi_contract_test.rb`
* `services/api-server/test/support/openapi_contract_assertions.rb`

Client tests for the Godot handoff consumer:

* `client/tests/unit/test_auth_session.gd`
* `client/tests/unit/test_auth_token_store.gd`
* `client/tests/unit/test_auth_session_controller.gd`
* `client/tests/unit/ui/sign_in/test_sign_in_flow.gd`
* `client/tests/unit/ui/sign_in/test_login_window.gd`
* `client/tests/unit/ui/menus/test_main_menu_auth_state.gd`

## Related docs

* [API Server Devtools](./!INDEX.md)
* [API-server auth and OAuth](../../services/api-server/auth-and-oauth.md)
* [API-server runtime and health](../../services/api-server/runtime-and-health.md)
* [API-server internal API surface](../../services/api-server/internal-api-surface.md)
* [Client auth session flow](../../services/client/auth-session-flow.md)
* [Account and identity current state](../../domains/platform/account-and-identity-current-state.md)
* [HTTP contract enforcement](../../protocol/http-contract-enforcement.md)

## Notes

This document is a smoke-flow procedure, not the canonical API auth implementation reference. API-server auth ownership belongs in the API-server service docs.

The legacy API notes correctly identified the current boundary: Rails owns authenticated accounts, OAuth identities, and bearer-token verification; the Go game-server consumes auth through an explicit internal verification API instead of reading Rails tables.

The local smoke flow should never require committing real secrets or captured credentials. Any captured local bearer token, OAuth code, poll secret, provider token, or internal service token should be treated as disposable local credential material.
