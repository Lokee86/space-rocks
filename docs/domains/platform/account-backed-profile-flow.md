# Account Backed Profile Flow

Parent index: [Platform](./!INDEX.md)

## Purpose

This doc describes the current account backend flow for Space Rocks.

It is a platform domain doc. It explains how the client, API server, game server, player-data runtime, rooms, and gameplay systems participate when an authenticated account signs in, proves identity to the game server, enters multiplayer, reads account-backed profile stats, and records account-backed match results.

This doc does not own service implementation detail, database schema, endpoint contracts, or future account policy.

## Overview

The current account backend flow is built around one durable online identity:

```text
Authenticated Account
```

An Authenticated Account is owned by the Rails API server and identified across systems by `account_id`.

The current backend flow has five main lanes:

```text
Account login
-> bearer token issued by API server

Saved-session validation
-> client validates stored token through API server

WebSocket authentication
-> client proves token to game server
-> game server verifies token through API server
-> game-server session becomes Authenticated Account

Account-backed profile read
-> client asks hosted player-data profile endpoint
-> game-server-hosted handler verifies token
-> player-data routes stats read to Rails/API-backed account store

Account-backed match-result write
-> game-server resolves authoritative match result
-> player-data routes authenticated-account write to Rails/API
-> Rails persists aggregate stats and match-result row
```

The flow keeps these boundaries separate:

```text
Rails/API account identity
!= bearer token
!= WebSocket session
!= room member
!= gameplay player
!= ship entity
!= display name
```

The API server owns account identity and online persistence. The game server owns realtime session identity, multiplayer admission, room participation, gameplay authority, and trusted match facts. The player-data runtime owns identity-based store routing. The client owns presentation, local token storage, and initiating the appropriate requests.

## Associated systems

* [Client](../../services/client/!INDEX.md) - stores the user bearer token, validates saved auth state, starts Discord login-session handoff, sends WebSocket auth, and displays profile state.
* [API Server](../../services/api-server/!INDEX.md) - owns Rails authenticated accounts, OAuth/password auth, bearer tokens, internal token verification, and Rails/Postgres account-backed stats.
* [Game Server](../../services/game-server/!INDEX.md) - owns WebSocket sessions, auth verification handoff, multiplayer admission, room lifecycle, gameplay authority, and match-result reporting.
* [Player Data](../../services/player-data/!INDEX.md) - owns identity-based profile/stat routing across Guest, Local Profile, and Authenticated Account routes.
* [Protocol](../../protocol/!INDEX.md) - owns documented request, response, packet, and contract behavior.
* [Data](../../data/!INDEX.md) - owns shared player-data schema and source-of-truth documentation.

## Authority boundaries

### Client authority

The client may:

* store a Space Rocks bearer token locally
* validate the saved bearer token through the API server
* request Discord browser-assisted login
* call logout and clear local session state
* send `authenticate_request` after WebSocket connection
* wait for WebSocket auth before sending multiplayer create/join requests
* call profile readout routes with the current identity context
* display profile callsign, activity status, and normalized stats

The client does not own:

* account identity
* token verification authority
* OAuth provider secrets
* durable account persistence
* multiplayer admission
* gameplay player identity
* account-backed stat mutation

### API-server authority

The API server owns:

* Rails `User` records
* `users.account_id`
* password credentials
* OAuth provider identities
* Discord OAuth state and login-session handoff
* bearer-token issuance, digest storage, expiry, revocation, and verification
* `/api/auth/me`
* `/api/auth/logout`
* `/internal/auth/verify-token`
* authenticated-account stats and match-result persistence
* Rails/Postgres account-backed player-data storage

The API server does not own:

* realtime simulation
* WebSocket sessions
* room lifecycle
* local single-player identity
* Local Profile SQLite storage
* Guest transient stats
* player-data route selection
* client profile presentation

### Game-server authority

The game server owns:

* WebSocket session state
* auth verification handoff to Rails
* conversion of valid token verification into session Authenticated Account identity
* multiplayer create/join admission checks
* room member account identity attachment
* gameplay authority
* match result summary production
* reporting resolved match results to player-data

The game server does not own:

* bearer-token issuance
* token digest storage
* OAuth login
* Rails auth tables
* Rails/Postgres player-data tables
* SQLite Local Profile tables
* player-data backing-store selection

### Player-data authority

The player-data runtime owns:

* profile/stat load behavior
* match-result stat mutation routing
* `guest`, `local_profile`, and `authenticated_account` route selection
* logical stats normalization
* guest transient stats
* local profile route behavior
* authenticated-account Rails adapter behavior

Player-data does not own:

* account authentication
* bearer-token issuance
* multiplayer admission
* match outcome authority
* live gameplay simulation
* client presentation

### Gameplay and room authority

Rooms own membership and lifecycle. Gameplay owns active match players, ship state, scoring, deaths, wins, and match facts.

Room member identity and gameplay player identity must not be replaced by `account_id`.

## Identifier rules

Current identifier rules:

| Identifier           | Meaning                                                   |
| -------------------- | --------------------------------------------------------- |
| `account_id`         | Canonical cross-system Authenticated Account UUID.        |
| Rails `user_id`      | Rails internal database identity.                         |
| bearer token         | Runtime proof material used to verify an account session. |
| WebSocket session ID | Runtime connection/session identity.                      |
| room member ID       | Room membership identity.                                 |
| gameplay player ID   | Live match/gameplay identity.                             |
| display name         | Presentation identity.                                    |
| callsign             | Presentation identity.                                    |

Rules:

* Use `account_id` for authenticated-account routing.
* Do not use Rails `user_id` as the cross-system account id.
* Do not expose bearer tokens as gameplay identity.
* Do not use display names or callsigns as durable routing identity.
* Do not replace gameplay player IDs with account IDs.
* Do not treat Local Profile identity as an Authenticated Account.

## Flow summary

### Account login and token issue

Current implemented backend auth supports:

* email/password registration and login at the API-server level
* Discord OAuth at the API-server level
* browser-assisted Discord login-session handoff for Godot
* opaque bearer-token issuance

The current client-facing sign-in flow is Discord-only. Manual email/password and Google controls exist as disabled UI controls in the client.

Discord browser handoff flow:

```text
client asks API server to create Discord login session
-> API server returns login_session_id, poll_secret, login_url, expires_at
-> client opens login_url in browser
-> browser completes Discord OAuth through Rails callback
-> Rails marks login session authenticated
-> client exchanges login_session_id + poll_secret
-> API server returns bearer token and auth user payload
-> client stores bearer token locally
```

The API server stores only bearer-token digests, not raw bearer tokens.

### Saved session validation

On startup, the client loads its saved bearer token.

```text
saved token exists
-> client calls GET /api/auth/me
-> API server verifies token
-> valid token returns current user
-> invalid token clears local client auth state
```

`GET /api/auth/me` returns `account_id`. The current client auth session state still stores the auth payload as client session presentation state and does not make the client authoritative for account identity.

### Logout

Logout is client-local first and remote best-effort:

```text
client begins logout
-> client clears saved token
-> client clears in-memory auth session
-> client calls DELETE /api/auth/logout when a token existed
-> API server revokes that bearer token
```

Remote logout revokes only the token used for that request. It does not imply global account deletion or account deactivation.

### WebSocket authentication

Every game-server WebSocket session starts as Guest identity.

Authenticated multiplayer requires an auth upgrade:

```text
client connects to WebSocket
-> client sends authenticate_request with bearer token
-> game-server auth verifier calls API-server /internal/auth/verify-token
-> API server verifies user token
-> API server returns valid account identity or valid=false
-> game server stores Authenticated Account identity on the WebSocket session when valid
-> game server sends authenticate_result
```

The API-server internal verification endpoint is protected by the internal service bearer token. The user bearer token is sent in the JSON request body and is not the same credential as the internal service token.

Valid verification returns minimal account identity:

```text
valid
Rails user id
account_id
display_name
```

The game server stores the verified `account_id` for account-routed runtime behavior. It does not store the raw bearer token as gameplay state.

### Multiplayer admission

Authenticated Account identity is required for current multiplayer create/join.

```text
session authenticated
-> create_room_request or join_room_request may proceed
-> room member receives account_id
```

Failure behavior:

| State                                           | Multiplayer create/join behavior |
| ----------------------------------------------- | -------------------------------- |
| No auth verifier configured                     | reject with `auth_unavailable`   |
| Verifier configured but session unauthenticated | reject with `auth_required`      |
| Valid authenticated-account session             | allow create/join                |
| Guest identity                                  | reject                           |
| Local Profile identity                          | reject                           |

Single-player remains Guest or Local Profile oriented. It does not require account authentication.

### Account-backed profile read

The account-backed profile read is a cross-system read path used by profile readout presentation.

```text
client requests profile readout
-> client calls POST /api/player-data/profile
-> game-server-hosted player-data handler receives request
-> authenticated-account request must include bearer token
-> handler verifies bearer token through game-server auth verifier
-> verifier resolves account_id through API server
-> handler loads stats through player-data runtime
-> player-data store router selects authenticated_account route
-> RailsStore calls API-server internal stats endpoint
-> API server loads or creates zeroed PlayerStat
-> normalized profile stats return to client
```

The client does not call Rails stats endpoints directly for profile readout.

The profile response is presentation-safe and normalized. It includes callsign, activity status, identity kind, and stats. It does not make the client owner of account identity or persistence.

### Account-backed match-result write

The account-backed match-result write starts after game-server match resolution.

```text
gameplay produces authoritative match facts
-> room/game-over lifecycle reaches match result reporting point
-> game server builds match summary
-> game server reports through player-data runtime sink
-> player-data routes by identity_kind
-> authenticated_account route uses RailsStore
-> RailsStore calls API-server internal match-results endpoint
-> API server persists PlayerMatchResult
-> API server updates PlayerStat aggregate
-> API server returns normalized stats
```

The API server does not decide the winner, score, deaths, or match-over state. Those are upstream game-server/gameplay facts.

`result_id` is the idempotency key for Rails match-result persistence. Duplicate submissions are accepted as duplicates and do not apply aggregate stats again.

## Inputs and outputs

Current account backend inputs:

* Discord OAuth callback data
* email/password auth request data at API-server level
* saved client bearer token
* WebSocket `authenticate_request`
* Rails internal service bearer token
* user bearer token submitted for verification
* requested multiplayer create/join action
* selected profile read identity
* authoritative match summary
* `account_id`
* `result_id`
* `match_id`

Current account backend outputs:

* bearer token
* current user payload
* token verification result
* WebSocket Authenticated Account session identity
* room member account reference
* profile readout payload
* normalized account stats
* accepted or duplicate match-result response
* persisted account aggregate stats
* persisted account match-result row

## Data routing

Current player-data routing for account backend behavior:

| Identity kind         | Durable account-shaped data | Runtime route               | Backing behavior                         |
| --------------------- | --------------------------- | --------------------------- | ---------------------------------------- |
| Guest                 | no                          | guest route                 | transient runtime memory                 |
| Local Profile         | yes, local-only             | local profile route         | embedded SQLite in standard local builds |
| Authenticated Account | yes, online                 | authenticated account route | Rails/API and Postgres when configured   |

Authenticated Account stats use the logical player-data stats shape:

```text
total_score
high_score
ship_deaths
games_played
wins
```

Authenticated Account match-result writes currently include:

```text
result_id
match_id
account_id
score
ship_deaths
won
```

The player-data runtime owns route selection. The game server does not choose Rails/Postgres directly, and the client does not choose the backing store.

## Failure behavior

Current account backend failures are intentionally separated by boundary.

Client auth/session failures:

* missing saved token signs the client out
* invalid saved token clears local auth state
* Discord login-session failure clears local auth state
* logout clears local state immediately

WebSocket auth failures:

* empty submitted token returns `invalid_token`
* invalid user token returns `invalid_token`
* verifier unavailable returns `token_verification_unavailable`
* auth failure leaves the WebSocket session connected as Guest

Multiplayer admission failures:

* missing verifier returns `auth_unavailable`
* unauthenticated create/join returns `auth_required`
* Guest and Local Profile identities are not admitted to online multiplayer

Profile read failures:

* missing, invalid, or unverifiable bearer token for authenticated-account profile read returns `unauthorized`
* unavailable profile runtime returns `profile_unavailable`
* invalid request shape returns `invalid_request`

API internal auth failures:

* invalid internal service bearer token returns `401`
* invalid user bearer token returns a successful internal response with `valid: false`

Account persistence failures:

* unknown `account_id` returns `unknown_user`
* missing required match-result fields return `invalid_input`
* duplicate `result_id` returns accepted duplicate behavior without double-counting stats

## Out of scope

This doc does not define:

* future Google OAuth
* future manual login/signup product UI
* email verification
* account recovery
* password reset
* provider linking or unlinking
* account deletion or deactivation
* account merge
* Guest-to-account migration
* Local Profile-to-account migration
* local-to-online progression import
* leaderboard eligibility
* [Trust And Eligibility Policy](trust-and-eligibility-policy.md)
* social graph policy
* matchmaking policy
* exact Rails physical schema
* exact SQLite physical schema
* direct code maps
* OpenAPI request/response ownership
* packet source-of-truth ownership

## Active issues

* Manual email/password auth exists at the API-server level, but the current client sign-in UI keeps manual auth and Google controls disabled. See [Current System Limits](../../limits/current-system-limits.md#client-menu-flow).
* `start_single_player_request` does not currently reject an already-authenticated WebSocket session directly at the server boundary. The intended model remains Guest or Local Profile for local single-player, and player-data mode validation rejects `single_player + authenticated_account`. See [Current System Limits](../../limits/current-system-limits.md#architecture--networking).

## Related docs

* [Platform](./!INDEX.md)
* [Account And Identity Current State](account-and-identity-current-state.md)
* [Client](../../services/client/!INDEX.md)
* [Client Auth Session Flow](../../services/client/auth-session-flow.md)
* [Client Profile Flow](../../services/client/pregame-menu-flow/profile-flow.md)
* [API Server](../../services/api-server/!INDEX.md)
* [API-server Auth And OAuth](../../services/api-server/auth-and-oauth.md)
* [API-server Internal API Surface](../../services/api-server/internal-api-surface.md)
* [Trust And Eligibility Policy](trust-and-eligibility-policy.md)
* [API-server Player Stats And Match Results](../../services/api-server/player-stats-and-match-results.md)
* [Game Server](../../services/game-server/!INDEX.md)
* [Game-server Auth Verifier Integration](../../services/game-server/integrations/auth-verifier-integration.md)
* [Game-server Auth Routing](../../services/game-server/networking/auth-routing.md)
* [Player Data](../../services/player-data/!INDEX.md)
* [Player-data Runtime And Store Routing](../../services/player-data/runtime-and-store-routing.md)
* [Player-data Profile Stats Flow](../../services/player-data/profile-stats-flow.md)
* [HTTP Contract Enforcement](../../protocol/http-contract-enforcement.md)
* [Current System Limits](../../limits/current-system-limits.md)
* [Account And Identity Systems planning](../../planning/domains/platform/account-and-identity-systems.md)

## Notes

This doc is intentionally narrower than the full account and identity domain. The broader current identity model belongs in [Account And Identity Current State](account-and-identity-current-state.md).

Authenticated Account is not a synced Local Profile. Local Profile is not an API-server cache. Guest and Local Profile data do not become online-trusted account data by being displayed next to account UI or by sharing the logical player-data stats shape.

The current token model uses opaque bearer tokens. JWT or another future token model belongs in account and identity planning until implemented.

The account-backed profile endpoint is currently hosted by the game-server process, but profile/stat routing still belongs to the player-data runtime. Hosting does not make the game server the owner of account persistence.

`account_id` is the durable cross-system account identity. Rails `user_id`, display name, bearer token, room member id, and gameplay player id must remain separate.
