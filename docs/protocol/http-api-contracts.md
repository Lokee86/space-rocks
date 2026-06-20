## HTTP API Contracts

Parent index: [Protocol](./!README.md)

## Purpose

This document describes the current Space Rocks HTTP API request and response contracts.

It covers the shared OpenAPI source, the Rails API surface, game-server-hosted player-data HTTP routes, internal service-to-service routes, client HTTP consumption, service responsibilities, validation expectations, and implementation paths that must stay aligned with the contract.

## Overview

HTTP request and response shapes are owned by:

```text
shared/contracts/http/openapi.yaml
```

That file is the source of truth for current JSON HTTP contracts. It defines path and method combinations, request bodies, response bodies, status codes, security declarations, and shared schemas for Rails API routes, player-data profile and local-profile routes, auth verification, authenticated-account stats reads, and match-result submission.

The contract is implemented manually by services. It does not generate Rails controllers, Go handlers, Godot API clients, Go API clients, Rails strong params, database migrations, or runtime middleware.

Current enforcement is test-time enforcement. Rails tests load the OpenAPI definition through `openapi_first`, and controller tests can assert request and response conformance with `assert_openapi_contract!`. Runtime OpenAPI middleware is not active.

HTTP API contracts are separate from realtime WebSocket packet contracts. WebSocket packet shapes are owned by the shared packet schema pipeline under `shared/packets/`, not by OpenAPI.

## Participating systems

```text
shared/contracts/http/openapi.yaml
```

Owns HTTP request and response shape.

```text
services/api-server/
```

Implements Rails-hosted HTTP routes for auth, OAuth handoff, current-user reads, public player stats, internal token verification, authenticated-account stats reads, and authenticated-account match-result persistence.

```text
services/game-server/
```

Hosts the current player-data HTTP facade on the game-server HTTP process, verifies authenticated tokens through Rails for realtime and profile flows, and produces match-result summaries that are reported through player-data.

```text
services/player-data/
```

Implements profile read, local profile management, runtime store routing, authenticated-account Rails adapter calls, guest transient stats, and local profile storage behavior behind the HTTP contract.

```text
client/
```

Consumes HTTP routes through Godot JSON API clients. The client builds endpoint URLs, sends JSON requests, attaches bearer tokens when supplied, parses dictionary JSON responses, and maps request failures into client request results.

## Protocol authority

OpenAPI owns the shape of HTTP messages.

Service implementations own the behavior behind those messages.

The API server owns Rails route behavior, Rails auth, bearer-token issuance and verification, OAuth handoff, authenticated-account stats, authenticated-account match-result persistence, and Rails/Postgres physical schema.

The game server owns its HTTP listener, route mounting, realtime WebSocket entrypoint, auth-verifier dependency wiring, and player-data HTTP handler hosting.

The player-data service owns player-data request handling, profile resolution, local profile operations, guest/local/account store routing, and the Rails adapter used for authenticated-account stats and match results.

The client owns HTTP consumption and presentation-facing error handling. It does not own contract authority.

OpenAPI does not own:

```text
Rails database schema
Rails migrations
embedded SQLite schema
Rails controller generation
Go handler generation
Godot client generation
runtime OpenAPI middleware
WebSocket packet schemas
player-data runtime packets
gameplay simulation authority
```

## Contracted route groups

### Rails public API routes

These routes are hosted by `services/api-server` on the Rails API process.

| Method   | Path                                             | Purpose                                                                       |
| -------- | ------------------------------------------------ | ----------------------------------------------------------------------------- |
| `GET`    | `/health`                                        | Rails API JSON health response.                                               |
| `POST`   | `/api/auth/register`                             | Register a new email/password user.                                           |
| `POST`   | `/api/auth/login`                                | Authenticate with email/password.                                             |
| `GET`    | `/api/auth/me`                                   | Read the current authenticated user from a bearer token.                      |
| `DELETE` | `/api/auth/logout`                               | Log out the current bearer-token session.                                     |
| `GET`    | `/api/auth/discord/start`                        | Start direct browser Discord OAuth.                                           |
| `GET`    | `/api/auth/discord/callback`                     | Complete direct browser Discord OAuth.                                        |
| `POST`   | `/api/auth/discord/login_sessions`               | Create a browser-assisted Discord login session for Godot.                    |
| `POST`   | `/api/auth/discord/login_sessions/{id}/exchange` | Exchange an authenticated Discord login session for the normal auth response. |
| `GET`    | `/api/player/stats`                              | Read public authenticated-player stats through Rails.                         |

Public client auth requests use user bearer tokens where required. OAuth provider secrets stay server-side in Rails and are never sent to Godot.

### Game-server-hosted player-data routes

These routes are reachable through the game-server process on the current `:8080` listener. The game server hosts them, but behavior is implemented by `services/player-data/httpapi`.

| Method   | Path                                                 | Purpose                                                                                |
| -------- | ---------------------------------------------------- | -------------------------------------------------------------------------------------- |
| `POST`   | `/api/player-data/profile`                           | Load a normalized profile for Guest, Local Profile, or Authenticated Account identity. |
| `GET`    | `/api/player-data/local-profiles`                    | List local profiles.                                                                   |
| `POST`   | `/api/player-data/local-profiles`                    | Create a local profile.                                                                |
| `PUT`    | `/api/player-data/local-profiles/{local_profile_id}` | Update a local profile display name.                                                   |
| `DELETE` | `/api/player-data/local-profiles/{local_profile_id}` | Delete a local profile.                                                                |
| `GET`    | `/api/player-data/local-profiles/default`            | Read the default local profile selection.                                              |
| `PUT`    | `/api/player-data/local-profiles/default`            | Persist Guest or a local profile as the default identity.                              |

The profile route accepts `play_mode`, `identity_kind`, and optional `local_profile_id`. Authenticated-account profile reads also require a user bearer token. The hosted handler verifies that token through the game-server auth-verifier adapter, which delegates to the Rails internal token verification route.

The local profile routes do not require bearer-token authentication. They operate on the local profile store selected by the player-data runtime. In standard development builds this is embedded SQLite. In `noembeddedsqlite` builds, local profile management returns unavailable behavior through the player-data handler.

### Internal service-to-service routes

These routes are Rails-hosted and protected by the internal bearer token.

| Method | Path                                  | Caller                  | Purpose                                                                       |
| ------ | ------------------------------------- | ----------------------- | ----------------------------------------------------------------------------- |
| `POST` | `/internal/auth/verify-token`         | Game server auth client | Verify a user bearer token and return minimal account identity.               |
| `POST` | `/api/internal/player-data/stats`     | Player-data RailsStore  | Load authenticated-account stats by `account_id`.                             |
| `POST` | `/internal/player-data/match-results` | Player-data RailsStore  | Persist trusted authenticated-account match results and return updated stats. |

Internal requests send:

```http
Authorization: Bearer <GAME_SERVER_INTERNAL_TOKEN>
```

The game server uses `API_SERVER_BASE_URL` and `GAME_SERVER_INTERNAL_TOKEN` to construct its auth verifier. The player-data Rails adapter uses `PLAYER_DATA_RAILS_BASE_URL` and `PLAYER_DATA_RAILS_INTERNAL_TOKEN` for authenticated-account stats and match-result calls.

## Request and response flow

### Client auth flow

The client auth flow uses `AuthApiClient`, `ApiConfig`, and `ApiHttpClient`.

```text
AuthSessionController
  -> AuthApiClient
  -> ApiHttpClient
  -> Rails API route
```

The client uses Rails auth routes to validate saved tokens, log out, create Discord login sessions, and exchange authenticated login sessions. The API server owns bearer-token validity and returned account identity. The client only stores and supplies the bearer token.

### Discord login-session flow

Godot starts browser-assisted Discord login by calling:

```text
POST /api/auth/discord/login_sessions
```

Rails returns:

```text
login_session_id
poll_secret
login_url
expires_at
```

Godot opens `login_url` in the browser. After the browser completes Discord OAuth, Godot polls:

```text
POST /api/auth/discord/login_sessions/{id}/exchange
```

with `poll_secret`. Pending sessions return `202` with `status: pending`. Authenticated sessions return the normal auth response containing a bearer token and user payload.

### Profile read flow

Profile readout uses:

```text
POST /api/player-data/profile
```

Current request fields:

```text
play_mode
identity_kind
local_profile_id
```

Identity behavior:

| `identity_kind`         | Required request data        | Auth behavior                | Runtime identity                                                          |
| ----------------------- | ---------------------------- | ---------------------------- | ------------------------------------------------------------------------- |
| `guest`                 | none                         | no bearer token required     | `identity_kind=guest`                                                     |
| `local_profile`         | non-empty `local_profile_id` | no bearer token required     | `identity_kind=local_profile`, `local_profile_id=<id>`                    |
| `authenticated_account` | bearer token                 | token verified through Rails | `identity_kind=authenticated_account`, `account_id=<verified account id>` |

The handler resolves identity and presentation fields, loads stats through the player-data runtime, and returns:

```json
{
  "profile": {
    "callsign": "Guest",
    "activity_status": "OFFLINE",
    "identity_kind": "guest",
    "stats": {
      "total_score": 0,
      "high_score": 0,
      "ship_deaths": 0,
      "games_played": 0,
      "wins": 0
    }
  }
}
```

Profile reads do not mutate stats. Stat mutation happens through match-result reporting.

### Local profile flow

Local profile management uses the player-data HTTP facade.

The local profile ID is the durable identity key. `display_name` is mutable presentation data and is not identity.

Create request:

```json
{
  "display_name": "Pilot_1",
  "seed_from_guest_stats": true
}
```

The player-data handler validates the display name, generates `local_profile_id` server-side, optionally asks the runtime for Guest seed stats, creates the profile, and returns the created profile.

Default profile requests use:

```json
{
  "identity_kind": "guest",
  "local_profile_id": ""
}
```

or:

```json
{
  "identity_kind": "local_profile",
  "local_profile_id": "local-profile-..."
}
```

Guest default requires an empty `local_profile_id`. Local profile default requires a non-empty `local_profile_id`.

### Internal token verification flow

The game server verifies user bearer tokens through Rails:

```text
services/game-server/internal/authclient.Client.VerifyToken
  -> POST /internal/auth/verify-token
  -> Internal::Auth::VerifyTokensController
  -> Auth::VerifyAccessToken
```

Request body:

```json
{
  "token": "user-bearer-token"
}
```

Valid user-token response:

```json
{
  "valid": true,
  "user": {
    "id": 1,
    "account_id": "account-uuid",
    "display_name": "Pilot"
  }
}
```

Invalid, missing, revoked, or expired user tokens return `200` with:

```json
{
  "valid": false
}
```

Invalid internal service authentication returns `401`.

### Authenticated-account stats read flow

Authenticated-account stats reads route from player-data to Rails:

```text
RailsStore.LoadStats
  -> POST /api/internal/player-data/stats
  -> Api::Internal::PlayerData::StatsController
```

Request body:

```json
{
  "account_id": "account-uuid"
}
```

Success response:

```json
{
  "stats": {
    "total_score": 0,
    "high_score": 0,
    "ship_deaths": 0,
    "games_played": 0,
    "wins": 0
  }
}
```

The Rails controller looks up `User` by `account_id`. If the user exists but has no `PlayerStat`, Rails creates a zeroed stats row before returning the response.

### Match-result submission flow

Authenticated-account match results route from player-data to Rails:

```text
RailsStore.RecordMatchResult
  -> POST /internal/player-data/match-results
  -> Internal::PlayerData::MatchResultsController
  -> PlayerStats::ApplyMatchResult
```

Request body:

```json
{
  "result_id": "result-1",
  "match_id": "match-1",
  "account_id": "account-uuid",
  "score": 12,
  "ship_deaths": 3,
  "won": true
}
```

Success response:

```json
{
  "accepted": true,
  "duplicate": false,
  "stats": {
    "total_score": 12,
    "high_score": 12,
    "ship_deaths": 3,
    "games_played": 1,
    "wins": 1
  }
}
```

`result_id` is the idempotency key. Duplicate result IDs are accepted with `duplicate: true` and do not apply the same result twice.

## Source-of-truth files

Primary HTTP contract source:

```text
shared/contracts/http/openapi.yaml
```

Rails contract enforcement:

```text
services/api-server/test/contracts/openapi_contract_test.rb
services/api-server/test/support/openapi_contract_assertions.rb
services/api-server/test/test_helper.rb
services/api-server/Gemfile
```

Rails route implementation surface:

```text
services/api-server/config/routes.rb
```

Game-server-hosted player-data route surface:

```text
services/game-server/cmd/game-server/main.go
services/game-server/cmd/game-server/player_data_http.go
```

Client HTTP consumption surface:

```text
client/scripts/api/api_config.gd
client/scripts/api/api_http_client.gd
client/scripts/api/api_request_result.gd
client/scripts/auth/auth_api_client.gd
client/scripts/profile/player_data_profile_api_client.gd
client/scripts/profile/local_pilot_api_client.gd
```

## Service responsibilities

### API server

The API server owns Rails-hosted HTTP behavior.

It owns:

```text
Rails public auth routes
Discord OAuth and login-session routes
current-user route
public player stats route
internal token verification route
internal authenticated-account stats route
internal match-result persistence route
internal bearer-token authentication for service-to-service calls
Rails/Postgres persistence for API-owned records
Rails OpenAPI contract test support
```

It does not own:

```text
gameplay simulation
WebSocket packet transport
game-server route hosting
local profile SQLite persistence
guest transient storage
player-data store routing
Godot request construction
OpenAPI generation
```

### Game server

The game server owns route hosting and integration wiring for its HTTP process.

It owns:

```text
game-server HTTP mux composition
GET /ws realtime WebSocket route
GET /health plaintext process health route
mounting player-data HTTP handlers on :8080
constructing the player-data runtime for hosted handlers
constructing the auth verifier from API_SERVER_BASE_URL and GAME_SERVER_INTERNAL_TOKEN
adapting the game-server auth verifier to the player-data HTTP auth verifier
producing match-result summaries upstream of player-data reporting
```

It does not own:

```text
player-data request semantics
player-data response payload ownership
Rails auth tables
Rails account persistence
OpenAPI contract source
local profile persistence internals
```

### Player data

The player-data service owns behavior behind the hosted player-data HTTP surface and Rails adapter calls.

It owns:

```text
profile HTTP handler behavior
local profile HTTP handler behavior
identity-based store routing
guest transient stats
local profile operations
local profile default selection
guest-to-local stat seeding
authenticated-account Rails adapter requests
match-result stats routing
```

It does not own:

```text
Rails/Postgres physical schema
Rails auth token issuance
game-server HTTP mux ownership
client UI presentation
OpenAPI contract source
WebSocket gameplay packet schema
```

### Client

The client owns HTTP consumption.

It owns:

```text
endpoint URL construction
JSON request helper behavior
Accept and Content-Type JSON headers
optional bearer Authorization header attachment
request body serialization
dictionary JSON response parsing
HTTP and network failure mapping
feature API wrapper methods
```

It does not own:

```text
HTTP contract authority
OpenAPI enforcement
server route definitions
server-side auth verification
local profile persistence
authenticated-account stats persistence
match-result authority
OAuth provider secrets
```

## Compatibility and validation

HTTP request or response shape changes must update the OpenAPI source and affected implementation/tests in the same change.

The minimum contract update rule is:

```text
1. Update shared/contracts/http/openapi.yaml.
2. Update the implementing Rails controller, Go handler, or client/API wrapper.
3. Update affected controller, handler, adapter, or client tests.
4. Run OpenAPI parsing and affected service tests.
```

Rails contract validation uses `openapi_first` in tests.

The basic OpenAPI parse test is:

```text
cd services/api-server && bundle exec rails test test/contracts/openapi_contract_test.rb
```

Rails integration tests can validate the current request and response by calling:

```text
assert_openapi_contract!
```

That assertion validates both:

```text
assert_openapi_request!
assert_openapi_response!
```

Go-hosted player-data HTTP routes are currently listed in the OpenAPI contract and tested through Go handler/runtime tests. They are not currently validated by Rails integration tests because they are not Rails routes.

Recommended verification when HTTP contracts change:

```text
cd services/api-server && bundle exec rails test
cd services/player-data && go test ./...
cd services/game-server && go test -buildvcs=false ./...
```

When local profile availability or build-tag behavior changes, also run:

```text
cd services/player-data && go test -tags noembeddedsqlite ./...
cd services/game-server && go test -tags noembeddedsqlite -buildvcs=false ./cmd/game-server
```

When client HTTP call shapes or response handling change, run the affected Godot tests in addition to service tests.

## Code map

Primary contract:

```text
shared/contracts/http/openapi.yaml
```

Rails contract loading and assertions:

```text
services/api-server/Gemfile
services/api-server/test/test_helper.rb
services/api-server/test/contracts/openapi_contract_test.rb
services/api-server/test/support/openapi_contract_assertions.rb
```

Rails routes and controllers:

```text
services/api-server/config/routes.rb
services/api-server/app/controllers/health_controller.rb
services/api-server/app/controllers/api/auth/registrations_controller.rb
services/api-server/app/controllers/api/auth/sessions_controller.rb
services/api-server/app/controllers/api/auth/me_controller.rb
services/api-server/app/controllers/api/auth/discord_controller.rb
services/api-server/app/controllers/api/auth/discord_login_sessions_controller.rb
services/api-server/app/controllers/api/player/stats_controller.rb
services/api-server/app/controllers/internal/base_controller.rb
services/api-server/app/controllers/api/internal/base_controller.rb
services/api-server/app/controllers/internal/auth/verify_tokens_controller.rb
services/api-server/app/controllers/api/internal/player_data/stats_controller.rb
services/api-server/app/controllers/internal/player_data/match_results_controller.rb
```

Rails service and model behavior behind internal routes:

```text
services/api-server/app/services/auth/verify_access_token.rb
services/api-server/app/services/player_stats/apply_match_result.rb
services/api-server/app/services/player_stats/serialize_stats.rb
services/api-server/app/models/user.rb
services/api-server/app/models/access_token.rb
services/api-server/app/models/player_stat.rb
services/api-server/app/models/player_match_result.rb
```

Rails contract and controller tests:

```text
services/api-server/test/controllers/internal/auth/verify_tokens_controller_test.rb
services/api-server/test/controllers/internal/player_data/match_results_controller_test.rb
services/api-server/test/controllers/api/auth/registrations_controller_test.rb
services/api-server/test/controllers/api/auth/sessions_controller_test.rb
services/api-server/test/controllers/api/auth/discord_login_sessions_controller_test.rb
services/api-server/test/controllers/api/internal/player_data/stats_controller_test.rb
```

Game-server HTTP route hosting and auth client:

```text
services/game-server/cmd/game-server/main.go
services/game-server/cmd/game-server/auth_config.go
services/game-server/cmd/game-server/player_data_http.go
services/game-server/internal/authclient/client.go
services/game-server/internal/authclient/types.go
services/game-server/internal/networking/session_auth.go
```

Player-data HTTP handlers and Rails adapter:

```text
services/player-data/httpapi/profile_handler.go
services/player-data/httpapi/local_profiles_handler.go
services/player-data/playerdata/configured_runtime.go
services/player-data/playerdata/runtime.go
services/player-data/playerdata/store_router.go
services/player-data/playerdata/rails_store.go
services/player-data/playerdata/guest_memory_store.go
services/player-data/playerdata/noop_store.go
services/player-data/playerdata/embeddedsqlite/sqlite_store.go
```

Client HTTP API consumers:

```text
client/scripts/api/api_config.gd
client/scripts/api/api_http_client.gd
client/scripts/api/api_request_result.gd
client/scripts/auth/auth_api_client.gd
client/scripts/auth/auth_session_controller.gd
client/scripts/auth/auth_token_store.gd
client/scripts/profile/player_data_profile_api_client.gd
client/scripts/profile/local_pilot_api_client.gd
client/scripts/profile/profile_stats_provider.gd
client/scripts/profile/profile_context_provider.gd
```

Important non-ownership boundaries:

```text
shared/packets/
= realtime WebSocket packet schema, not HTTP API contracts

services/api-server/db/migrate/
= Rails/Postgres physical schema, not HTTP request/response shape

services/player-data/playerdata/embeddedsqlite/sqlite_store.go
= embedded SQLite physical schema, not HTTP contract source

client/scenes/
= presentation and UI structure, not HTTP contract source
```

## Related docs

* [Protocol](./!README.md)
* [HTTP Contract Enforcement](http-contract-enforcement.md)
* [Source Of Truth Map](../data/source-of-truth-map.md)
* [Data](../data/!README.md)
* [API Server](../services/api-server/!README.md)
* [API-server Auth And OAuth](../services/api-server/auth-and-oauth.md)
* [API-server Internal API Surface](../services/api-server/internal-api-surface.md)
* [API-server Player Stats And Match Results](../services/api-server/player-stats-and-match-results.md)
* [Client HTTP API Flow](../services/client/client-http-api-flow.md)
* [Game Server](../services/game-server/!README.md)
* [Game-server Player Data HTTP Hosting](../services/game-server/integrations/player-data-http-hosting.md)
* [Game-server Route Composition](../services/game-server/process/route-composition.md)
* [Player Data](../services/player-data/!README.md)
* [Player-data Local Profiles HTTP API](../services/player-data/local-profiles-http-api.md)
* [Player-data Profile Stats Flow](../services/player-data/profile-stats-flow.md)
* [Player-data Runtime And Store Routing](../services/player-data/runtime-and-store-routing.md)

## Notes

The OpenAPI `/health` route describes the Rails API JSON health endpoint. The game-server process also exposes `GET /health`, but that route currently returns plaintext `OK` and is documented as a game-server process route rather than the Rails JSON health contract.

The current OpenAPI contract is authoritative for HTTP shape, but only Rails tests currently use OpenAPI assertion helpers directly. Go-hosted player-data routes rely on Go handler/runtime tests and manual alignment with the shared OpenAPI file.

Detailed endpoint behavior belongs in service docs. This protocol doc should remain focused on which systems communicate, what HTTP surfaces exist, who owns the contract, and how request/response shape changes stay aligned.
