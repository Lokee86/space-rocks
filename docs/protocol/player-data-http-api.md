# Player Data HTTP API

Parent index: [Protocol](./!INDEX.md)

## Purpose

This document describes the current player-data HTTP protocol surfaces.

It covers the client-facing player-data HTTP facade hosted by the game-server process, the request and response shapes owned by OpenAPI, the player-data handlers that implement profile and local-profile behavior, and the internal Rails HTTP calls used for authenticated-account persistence.

## Overview

The player-data HTTP API is a JSON HTTP boundary for profile readout and local profile management.

Current client-facing player-data routes are hosted by the game-server process on the data-handler base URL:

```text
http://localhost:8080
```

The game server hosts these routes, but it does not own the player-data behavior behind them. It builds the player-data runtime, adapts the auth verifier, mounts the routes, and delegates request handling to `services/player-data/httpapi`.

Current hosted player-data routes:

```text
POST   /api/player-data/profile
GET    /api/player-data/local-profiles
POST   /api/player-data/local-profiles
PUT    /api/player-data/local-profiles/{local_profile_id}
DELETE /api/player-data/local-profiles/{local_profile_id}
GET    /api/player-data/local-profiles/default
PUT    /api/player-data/local-profiles/default
```

The hosted API is consumed by the Godot client for pregame profile readout and local pilot management. The API carries identity context, local profile identifiers, display names, default-profile selection, and normalized player stats. It does not carry live gameplay state and does not allow the client to choose guest memory, embedded SQLite, Rails, or Postgres directly.

HTTP request and response shapes are sourced from:

```text
shared/contracts/http/openapi.yaml
```

OpenAPI owns HTTP shape. Player-data runtime and stores own behavior behind the shape.

Authenticated-account stats and match-result persistence use separate internal Rails HTTP calls made by the player-data Rails adapter:

```text
POST /api/internal/player-data/stats
POST /internal/player-data/match-results
```

Those routes are owned by the API server. They are part of the player-data HTTP communication flow, but they are not hosted by the game-server player-data facade.

## Participating systems

```text
client/
```

Consumes the hosted player-data HTTP facade through `PlayerDataProfileApiClient`, `LocalPilotApiClient`, and the shared `ApiHttpClient`.

```text
services/game-server/
```

Hosts the HTTP mux on `:8080`, mounts player-data routes, builds the player-data runtime, and adapts the game-server auth verifier for authenticated profile reads.

```text
services/player-data/
```

Implements the hosted profile and local profile handlers, owns player-data runtime behavior, routes identities to stores, and calls Rails internal endpoints for authenticated-account persistence when configured.

```text
services/api-server/
```

Owns Rails auth, token verification, authenticated-account stats persistence, authenticated-account match-result persistence, and Rails/Postgres physical storage.

```text
shared/contracts/http/openapi.yaml
```

Owns the HTTP request and response contract.

## Protocol authority

OpenAPI owns HTTP request and response shapes.

The game server owns route hosting for the current in-process player-data facade.

The player-data service owns hosted profile and local profile HTTP behavior.

The player-data runtime owns identity-based route selection and store delegation.

The API server owns authenticated-account internal HTTP endpoints and Rails/Postgres persistence.

The client owns request consumption and presentation flow, not the authoritative meaning of returned data.

The player-data HTTP API does not own:

```text
WebSocket packet shapes
runtime player-data packet schemas
Rails database schema
embedded SQLite schema
client UI layout
room admission
live gameplay simulation
match outcome calculation
OAuth provider behavior
OpenAPI runtime middleware
generated HTTP clients
```

## Request flow

### Hosted profile read

```text
client
-> POST /api/player-data/profile on game-server data-handler
-> services/player-data/httpapi.ProfileHandler
-> optional bearer-token verification for authenticated_account
-> playerdata.Runtime.LoadStats
-> selected store route
-> normalized profile response
-> client profile readout
```

The profile endpoint exists so the client can request one normalized profile payload regardless of whether the active identity is Guest, Local Profile, or Authenticated Account.

Request body:

```json
{
  "play_mode": "single_player",
  "identity_kind": "guest",
  "local_profile_id": ""
}
```

Supported `play_mode` values:

```text
single_player
multiplayer
multiplayer_simulation
```

Supported `identity_kind` values:

```text
guest
local_profile
authenticated_account
```

Guest profile reads require no bearer token.

Local Profile profile reads require a non-empty `local_profile_id` and require no bearer token.

Authenticated Account profile reads require:

```text
Authorization: Bearer <user token>
```

The hosted profile handler verifies the user token through the injected auth verifier. Token verification returns the authenticated `account_id` and optional display name used for profile presentation and store routing.

Successful response:

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

Current presentation values:

| Identity kind           | Callsign source                            | Activity status |
| ----------------------- | ------------------------------------------ | --------------- |
| `guest`                 | `Guest`                                    | `OFFLINE`       |
| `local_profile`         | `Local Pilot`                              | `LOCAL`         |
| `authenticated_account` | verified display name, or `Pilot` fallback | `ACTIVE`        |

If `Runtime.LoadStats` succeeds but reports no found stats, the profile handler returns zero stats.

The profile HTTP handler validates supported `play_mode` and supported `identity_kind` strings. It does not apply the full player-data mode/identity admission matrix before calling `Runtime.LoadStats`. Packet-dispatched player-data reads and writes validate mode/identity pairs through the player-data dispatcher.

Profile error responses use this shape:

```json
{
  "error": "invalid_request"
}
```

Current profile errors:

| Error                 | HTTP status | Meaning                                                                                                                  |
| --------------------- | ----------: | ------------------------------------------------------------------------------------------------------------------------ |
| `method_not_allowed`  |       `405` | Unsupported method reached the handler.                                                                                  |
| `invalid_request`     |       `400` | Invalid JSON, unsupported play mode, unsupported identity kind, or missing local profile ID for a local-profile request. |
| `unauthorized`        |       `401` | Missing, invalid, unverifiable, or unsupported authenticated-account bearer token.                                       |
| `profile_unavailable` |       `500` | Missing runtime or failed stats load.                                                                                    |

### Local profile list

```text
GET /api/player-data/local-profiles
```

This endpoint lists local profile summaries for the local pilot selector.

Successful response:

```json
{
  "profiles": [
    {
      "local_profile_id": "local-profile-...",
      "display_name": "Pilot_1"
    }
  ]
}
```

`local_profile_id` is durable local identity.

`display_name` is presentation data, not identity.

Unavailable local profile storage returns:

```json
{
  "error": "local_profiles_unavailable"
}
```

with HTTP `503`.

### Local profile create

```text
POST /api/player-data/local-profiles
```

This endpoint creates a new local profile with a server-generated ID.

Request body:

```json
{
  "display_name": "Pilot_1",
  "seed_from_guest_stats": true
}
```

The handler trims `display_name` and accepts only non-empty names matching:

```text
^[A-Za-z0-9_-]+$
```

The handler generates IDs in this format:

```text
local-profile-<16 random bytes as hex>
```

When `seed_from_guest_stats` is `false`, the new profile starts with zero stats.

When `seed_from_guest_stats` is `true`, the handler asks the runtime for current Guest stats and passes those stats into local profile creation. The handler does not read guest storage directly.

Successful response:

```json
{
  "profile": {
    "local_profile_id": "local-profile-...",
    "display_name": "Pilot_1"
  }
}
```

with HTTP `201`.

Invalid JSON or invalid display names return `invalid_request` with HTTP `400`.

Unavailable local profile storage or guest-stat seeding failure returns `local_profiles_unavailable` with HTTP `503`.

### Local profile display-name update

```text
PUT /api/player-data/local-profiles/{local_profile_id}
```

This endpoint updates display name only.

Request body:

```json
{
  "display_name": "Pilot_2"
}
```

The path `local_profile_id` must be non-empty. The new `display_name` uses the same validation rule as create.

Successful response:

```json
{
  "profile": {
    "local_profile_id": "local-profile-...",
    "display_name": "Pilot_2"
  }
}
```

The local profile identity and stats stay attached to the same `local_profile_id`.

Missing profiles return:

```json
{
  "error": "local_profile_not_found"
}
```

with HTTP `404`.

### Local profile delete

```text
DELETE /api/player-data/local-profiles/{local_profile_id}
```

This endpoint deletes a local profile through the player-data runtime and local store.

Successful delete returns:

```text
204 No Content
```

The embedded SQLite store deletes the profile row, local stats row, and local match-result rows. If the deleted profile was the stored default, the default resets to Guest.

The handler requires a non-empty path `local_profile_id`.

Missing profiles return `local_profile_not_found` with HTTP `404`.

Unavailable local profile storage returns `local_profiles_unavailable` with HTTP `503`.

### Default local profile read

```text
GET /api/player-data/local-profiles/default
```

This endpoint reads the persisted default local identity for the local pilot flow.

Guest response:

```json
{
  "default_profile": {
    "identity_kind": "guest",
    "local_profile_id": "",
    "display_name": "Guest"
  }
}
```

Local profile response:

```json
{
  "default_profile": {
    "identity_kind": "local_profile",
    "local_profile_id": "local-profile-...",
    "display_name": "Pilot_1"
  }
}
```

The local store returns Guest when no default row exists, when a stored default is invalid, or when the stored local profile no longer exists.

Unavailable local profile storage returns `local_profiles_unavailable` with HTTP `503`.

### Default local profile update

```text
PUT /api/player-data/local-profiles/default
```

Guest request:

```json
{
  "identity_kind": "guest",
  "local_profile_id": ""
}
```

Local profile request:

```json
{
  "identity_kind": "local_profile",
  "local_profile_id": "local-profile-..."
}
```

The handler accepts only:

```text
identity_kind = guest
identity_kind = local_profile
```

Guest default requires an empty `local_profile_id`.

Local profile default requires a non-empty `local_profile_id`.

Successful response:

```json
{
  "default_profile": {
    "identity_kind": "local_profile",
    "local_profile_id": "local-profile-...",
    "display_name": "Pilot_1"
  }
}
```

Missing local profiles return `local_profile_not_found` with HTTP `404`.

Invalid identity/default combinations return `invalid_request` with HTTP `400`.

Unavailable local profile storage returns `local_profiles_unavailable` with HTTP `503`.

## Hosted local profile error model

Local profile route errors use this shape:

```json
{
  "error": "local_profiles_unavailable"
}
```

Current local profile errors:

| Error                        | HTTP status | Meaning                                                                                                                                           |
| ---------------------------- | ----------: | ------------------------------------------------------------------------------------------------------------------------------------------------- |
| `method_not_allowed`         |       `405` | Unsupported HTTP method reached the handler.                                                                                                      |
| `invalid_request`            |       `400` | Invalid JSON, invalid display name, invalid path value, unsupported default identity kind, or invalid default identity/local-profile combination. |
| `local_profile_not_found`    |       `404` | Requested local profile does not exist.                                                                                                           |
| `local_profiles_unavailable` |       `503` | Runtime is missing or the active store does not support local profile management.                                                                 |

## Internal Rails player-data HTTP calls

Authenticated Account stats are routed through the player-data Rails adapter when `PLAYER_DATA_RAILS_BASE_URL` is configured.

The adapter is an HTTP client inside `services/player-data`. The Rails API server owns the endpoints it calls.

### Authenticated-account stats read

```text
RailsStore.LoadStats
-> POST /api/internal/player-data/stats
```

Request body:

```json
{
  "account_id": "acct-..."
}
```

Required header:

```text
Authorization: Bearer <PLAYER_DATA_RAILS_INTERNAL_TOKEN>
```

Successful response:

```json
{
  "stats": {
    "total_score": 12,
    "high_score": 9,
    "ship_deaths": 3,
    "games_played": 4,
    "wins": 2
  }
}
```

The Rails adapter accepts only `identity_kind = authenticated_account` and requires a non-empty `account_id`.

`POST /api/internal/player-data/stats` is not the same as public `GET /api/player/stats`. The player-data route uses the internal stats endpoint by `account_id`.

### Authenticated-account match-result write

```text
RailsStore.RecordMatchResult
-> POST /internal/player-data/match-results
```

Request body:

```json
{
  "result_id": "result-1",
  "match_id": "match-1",
  "account_id": "acct-...",
  "score": 12,
  "ship_deaths": 2,
  "won": true
}
```

Required header:

```text
Authorization: Bearer <PLAYER_DATA_RAILS_INTERNAL_TOKEN>
```

Successful response:

```json
{
  "accepted": true,
  "duplicate": false,
  "stats": {
    "total_score": 12,
    "high_score": 12,
    "ship_deaths": 2,
    "games_played": 1,
    "wins": 1
  }
}
```

The Rails adapter requires:

```text
identity_kind = authenticated_account
account_id
result_id
match_id
PLAYER_DATA_RAILS_INTERNAL_TOKEN
```

A rejected Rails response with `accepted = false` is converted into a player-data store error.

## Token roles

User bearer tokens are used for authenticated-account profile reads through:

```text
POST /api/player-data/profile
```

The token proves the client’s authenticated account identity to the hosted player-data profile handler. The handler verifies the token through the game-server auth verifier adapter.

Internal bearer tokens are used for player-data service-to-Rails calls:

```text
POST /api/internal/player-data/stats
POST /internal/player-data/match-results
```

Those calls use:

```text
PLAYER_DATA_RAILS_INTERNAL_TOKEN
```

Static user bearer-token configuration such as `PLAYER_DATA_RAILS_BEARER_TOKEN` is not part of the current flow.

Local profile management routes do not require bearer-token auth in the current handler implementation.

## Lifecycle

### Game-server startup

```text
main()
  create HTTP mux
  create room manager
  build player-data runtime
  create player-data sink
  create match reporter
  build auth verifier
  mount health and WebSocket routes
  mount player-data HTTP routes
  listen on :8080
```

The game-server process fails startup if player-data runtime initialization fails.

The game-server process also fails startup if match reporter initialization fails.

### Runtime configuration

The game-server builds the player-data runtime from:

```text
PLAYER_DATA_RAILS_BASE_URL
PLAYER_DATA_RAILS_INTERNAL_TOKEN
playerDataLocalStorePath()
playerDataLocalStoreFactory()
```

Configured runtime routing:

| Identity kind           | Runtime route | Backing behavior                                                                                |
| ----------------------- | ------------- | ----------------------------------------------------------------------------------------------- |
| `guest`                 | Guest store   | process-local guest memory                                                                      |
| `local_profile`         | Local store   | embedded SQLite in standard no-tag development builds                                           |
| `authenticated_account` | Account store | Rails adapter when Rails base URL is configured; in-memory account fallback when not configured |

In `noembeddedsqlite` builds, the embedded SQLite package and dependency are excluded. The local store path is unavailable, and local profile management returns `local_profiles_unavailable`.

## Source-of-truth files

Primary HTTP contract source:

```text
shared/contracts/http/openapi.yaml
```

Hosted player-data HTTP handlers:

```text
services/player-data/httpapi/profile_handler.go
services/player-data/httpapi/local_profiles_handler.go
```

Hosted route composition:

```text
services/game-server/cmd/game-server/main.go
services/game-server/cmd/game-server/player_data_http.go
```

Client HTTP consumers:

```text
client/scripts/api/api_config.gd
client/scripts/api/api_http_client.gd
client/scripts/profile/player_data_profile_api_client.gd
client/scripts/profile/local_pilot_api_client.gd
```

Authenticated-account Rails adapter:

```text
services/player-data/playerdata/rails_store.go
```

Rails internal endpoint ownership:

```text
services/api-server/config/routes.rb
services/api-server/app/controllers/api/internal/player_data/stats_controller.rb
services/api-server/app/controllers/internal/player_data/match_results_controller.rb
```

Related logical schema sources:

```text
shared/player_data/stats.toml
shared/player_data/match_result.toml
```

Related runtime packet source:

```text
shared/packets/player_data.toml
```

## Service responsibilities

### Client

The client owns:

* building hosted player-data request URLs from `ApiConfig`
* sending JSON requests through `ApiHttpClient`
* attaching a user bearer token when a caller supplies one
* constructing profile-read request bodies
* constructing local profile management request bodies
* consuming profile, stats, local profile list, and default-profile responses
* mapping request failures into client flow state

The client does not own:

* HTTP contract source-of-truth
* backing-store selection
* local SQLite persistence
* Rails/Postgres persistence
* token verification
* stat mutation authority
* gameplay fact authority

### Game Server

The game server owns:

* hosting the player-data HTTP routes on the current HTTP mux
* constructing the player-data runtime
* adapting the game-server auth verifier to the player-data HTTP auth verifier interface
* passing the same runtime to profile/local-profile handlers and match-result reporting
* keeping hosted player-data routes separate from WebSocket and simulation behavior

The game server does not own:

* player-data request behavior after delegation to handlers
* local profile persistence internals
* Rails authenticated-account persistence internals
* OpenAPI contract ownership
* client presentation
* live gameplay-to-stats aggregation rules outside match-result reporting

### Player Data

The player-data service owns:

* profile HTTP request validation and response shaping
* authenticated-account identity resolution after token verification
* profile stats loading through the runtime
* local profile list/create/update/delete/default behavior
* local profile ID generation
* local profile display-name validation
* guest-stat seeding for new local profiles
* mapping runtime/store errors into player-data HTTP error responses
* Rails adapter request construction for authenticated-account stats reads and match-result writes

The player-data service does not own:

* Rails auth token issuance
* Rails/Postgres physical schema
* game-server route hosting
* OpenAPI contract enforcement
* client UI behavior
* game simulation or match outcome calculation

### API Server

The API server owns:

* internal token-protected authenticated-account stats reads
* internal token-protected authenticated-account match-result writes
* public authenticated `GET /api/player/stats`
* token verification for game-server authenticated account flow
* Rails/Postgres physical persistence
* Rails OpenAPI contract test support

The API server does not own:

* Guest persistence
* Local Profile persistence
* embedded SQLite storage
* player-data runtime store routing
* game-server-hosted player-data facade routes

## Compatibility and update expectations

HTTP shape changes must update:

```text
shared/contracts/http/openapi.yaml
```

and the affected handlers, client wrappers, Rails controllers, and tests in the same change.

Hosted player-data profile/local-profile HTTP changes should not be made only in client scripts or only in Go handlers. The OpenAPI contract is the HTTP shape source.

Runtime packet changes are separate from HTTP changes and must update:

```text
shared/packets/player_data.toml
```

plus generated outputs.

Logical player-data schema changes are separate from HTTP and packet shape changes and must update:

```text
shared/player_data/stats.toml
shared/player_data/match_result.toml
```

plus any intentionally mirrored implementation contracts.

The current implementation does not generate Godot clients, Go clients, Rails controllers, or database schema from OpenAPI.

## Validation and testing

HTTP contract validation:

```text
cd services/api-server && bundle exec rails test test/contracts/openapi_contract_test.rb
```

Broader API-server validation when Rails internal player-data routes change:

```text
cd services/api-server && bundle exec rails test
```

Player-data validation:

```text
cd services/player-data && go test ./...
```

Restricted build validation:

```text
cd services/player-data && go test -tags noembeddedsqlite ./...
```

Game-server hosted-route and composition validation:

```text
cd services/game-server && go test -buildvcs=false ./cmd/game-server
cd services/game-server && go test -tags noembeddedsqlite -buildvcs=false ./cmd/game-server
```

Relevant current tests:

```text
services/player-data/httpapi/local_profiles_handler_test.go
services/player-data/playerdata/rails_store_test.go
services/player-data/playerdata/configured_runtime_test.go
services/player-data/playerdata/store_router_test.go
services/player-data/playerdata/noop_store_test.go
services/player-data/playerdata/embeddedsqlite/sqlite_store_test.go
services/api-server/test/controllers/api/internal/player_data/stats_controller_test.rb
services/api-server/test/controllers/internal/player_data/match_results_controller_test.rb
services/api-server/test/contracts/openapi_contract_test.rb
client/tests/unit/profile/test_player_data_profile_api_client.gd
```

Covered behavior includes:

* local profile routes returning `local_profiles_unavailable` when no local profile store is configured
* local profile creation with zero seed stats when guest stat seeding is disabled
* local profile creation copying Guest stats when guest stat seeding is enabled
* Rails adapter use of `/api/internal/player-data/stats` instead of public `/api/player/stats`
* Rails adapter use of internal bearer-token headers
* Rails adapter match-result request shape and duplicate response handling
* client profile API request body construction
* client profile bearer-token forwarding

## Code map

Hosted route installation:

```text
services/game-server/cmd/game-server/main.go
services/game-server/cmd/game-server/player_data_http.go
```

Hosted profile and local profile handlers:

```text
services/player-data/httpapi/profile_handler.go
services/player-data/httpapi/local_profiles_handler.go
```

Player-data runtime and local profile store boundary:

```text
services/player-data/playerdata/runtime.go
services/player-data/playerdata/store.go
services/player-data/playerdata/store_router.go
services/player-data/playerdata/configured_runtime.go
services/player-data/playerdata/noop_store.go
services/player-data/playerdata/embeddedsqlite/sqlite_store.go
```

Authenticated-account Rails adapter:

```text
services/player-data/playerdata/rails_store.go
```

Client consumers:

```text
client/scripts/api/api_config.gd
client/scripts/api/api_http_client.gd
client/scripts/profile/player_data_profile_api_client.gd
client/scripts/profile/local_pilot_api_client.gd
client/scripts/ui/menu_flow/local_pilot_flow.gd
```

API-server internal endpoint implementation:

```text
services/api-server/config/routes.rb
services/api-server/app/controllers/api/internal/player_data/stats_controller.rb
services/api-server/app/controllers/internal/player_data/match_results_controller.rb
services/api-server/app/services/player_stats/apply_match_result.rb
services/api-server/app/services/player_stats/serialize_stats.rb
```

Contract and schema sources:

```text
shared/contracts/http/openapi.yaml
shared/player_data/stats.toml
shared/player_data/match_result.toml
shared/packets/player_data.toml
services/player-data/protocol/packets.go
```

Related tests:

```text
services/player-data/httpapi/local_profiles_handler_test.go
services/player-data/playerdata/rails_store_test.go
services/player-data/playerdata/configured_runtime_test.go
services/player-data/playerdata/store_router_test.go
services/player-data/playerdata/embeddedsqlite/sqlite_store_test.go
services/api-server/test/controllers/api/internal/player_data/stats_controller_test.rb
services/api-server/test/controllers/internal/player_data/match_results_controller_test.rb
services/api-server/test/contracts/openapi_contract_test.rb
client/tests/unit/profile/test_player_data_profile_api_client.gd
```

Important non-ownership boundaries:

```text
services/game-server/internal/networking/
= owns the auth verifier consumed by the HTTP adapter

services/game-server/internal/matchreporting/
= owns match-result reporting into player-data

services/api-server/
= owns Rails auth and authenticated-account persistence

shared/contracts/http/openapi.yaml
= owns HTTP request/response shapes

shared/packets/player_data.toml
= owns player-data runtime packet shapes

shared/player_data/*.toml
= owns logical player-data schema
```

## Related docs

* [Protocol](./!README.md)
* [HTTP Contract Enforcement](./http-contract-enforcement.md)
* [Player Data](../services/player-data/!README.md)
* [Local Profiles HTTP API](../services/player-data/local-profiles-http-api.md)
* [Profile Stats Flow](../services/player-data/profile-stats-flow.md)
* [Runtime And Store Routing](../services/player-data/runtime-and-store-routing.md)
* [Game Server Player Data HTTP Hosting](../services/game-server/integrations/player-data-http-hosting.md)
* [Client HTTP API Flow](../services/client/client-http-api-flow.md)
* [API Server Player Stats And Match Results](../services/api-server/player-stats-and-match-results.md)
* [Player Data Schema](../data/player-data-schema.md)
* [Player Data Routing Flow](../domains/platform/player-data-routing-flow.md)

## Notes

The current player-data HTTP facade is hosted in the game-server process, but its behavior remains a player-data service responsibility.

The hosted HTTP profile path and generated player-data packet path are related but not identical. The HTTP profile handler validates supported play-mode and identity-kind strings, resolves identity, and calls `Runtime.LoadStats` directly. The packet dispatcher validates mode/identity combinations before store access.

Local Profile display names are presentation values. `local_profile_id` is the local durable identity.

Authenticated Account routing uses `account_id` resolved from token verification. Rails `user_id` remains an API-server internal database identifier.

Guest stats are process-local and transient. Local Profile stats are durable only when local storage is configured. Authenticated Account stats route through the Rails adapter when Rails configuration is present.
