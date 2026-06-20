## Local Profiles HTTP API

Parent index: [Player Data](./!INDEX.md)

## Purpose

This document describes the player-data service implementation responsibility for the local profiles HTTP API.

It covers local profile listing, creation, display-name updates, deletion, default profile selection, request validation, runtime/store delegation, and the boundaries between player-data, game-server hosting, client presentation, and HTTP contract ownership.

## Overview

The local profiles HTTP API is the JSON HTTP surface used by the client local pilot flow to manage single-player local profile identity.

The routes are currently hosted by the game-server HTTP process on `:8080`, but the behavior is implemented by the player-data service:

```text
client local pilot flow
  -> HTTP request to game-server data-handler route
  -> services/player-data/httpapi.LocalProfilesHandler
  -> playerdata.Runtime
  -> playerdata.StoreRouter local store
  -> embedded SQLite store in standard development builds
```

The handler owns HTTP request validation, response shaping, error mapping, local profile ID generation, and calls into the player-data runtime. The runtime owns whether local profile management is available and delegates local profile operations to a store implementing `playerdata.LocalProfileStore`.

In the standard no-tag development build, the game-server composition root injects an embedded SQLite local store. In `noembeddedsqlite` builds, the game-server supplies no SQLite path or local store factory, so local profile management routes return `local_profiles_unavailable`.

The local profile ID is the durable identity key. The display name is callsign/presentation data and can change without changing identity or resetting stats.

## Code root

```text
services/player-data/
```

Primary implementation area:

```text
services/player-data/httpapi/
services/player-data/playerdata/
services/player-data/playerdata/embeddedsqlite/
```

Hosted through:

```text
services/game-server/cmd/game-server/
```

## Responsibilities

The local profiles HTTP API owns:

* serving local profile management requests through `httpapi.NewLocalProfilesHandler`
* listing local profiles by `local_profile_id` and `display_name`
* creating local profiles with server-generated `local_profile_id` values
* validating display names before create or update
* updating only the local profile display name
* deleting local profiles through the runtime/store seam
* reading the current default local profile selection
* setting the default selection to Guest or a local profile
* mapping unavailable local profile storage to `local_profiles_unavailable`
* mapping missing local profiles to `local_profile_not_found`
* returning JSON response bodies for successful reads, creates, updates, and default changes
* returning `204 No Content` for successful deletes
* delegating persistence to the player-data runtime instead of directly opening storage from the handler

## Does not own

The local profiles HTTP API does not own:

* game-server route hosting or process startup
* WebSocket packet routing
* client selector presentation
* client-side local pilot UI state
* Rails account identity
* authenticated account persistence
* API-server auth policy
* OAuth token issuance or verification
* match-result calculation
* general profile readout through `POST /api/player-data/profile`
* HTTP contract source-of-truth
* OpenAPI enforcement
* direct SQLite access from game-server code

## Domain roles

The local profiles HTTP API participates in the local player identity/profile portion of the player experience flow.

Its service role is durable local profile management for single-player identity:

```text
Guest
= fallback local identity, no local_profile_id

Local Profile
= durable local identity, keyed by local_profile_id

display_name
= mutable callsign/presentation value
```

Guest remains the fallback default. A local profile can be selected as the default only by `identity_kind = local_profile` and a non-empty `local_profile_id`.

The service does not treat display name as identity. Renaming a profile leaves its `local_profile_id`, local stats, and match-result history attached to the same profile.

## Protocols and APIs

The local profiles API is a JSON HTTP surface. The client consumes it through `LocalPilotApiClient`, and the game server hosts it on the shared HTTP mux. The handler does not require or consume bearer-token auth for local profile management.

OpenAPI owns the request and response shape contract. The player-data handler owns runtime behavior behind the contract.

### Endpoint summary

| Method   | Path                                                 | Behavior                                      |
| -------- | ---------------------------------------------------- | --------------------------------------------- |
| `GET`    | `/api/player-data/local-profiles`                    | Lists local profiles.                         |
| `POST`   | `/api/player-data/local-profiles`                    | Creates a local profile.                      |
| `PUT`    | `/api/player-data/local-profiles/{local_profile_id}` | Updates local profile display name.           |
| `DELETE` | `/api/player-data/local-profiles/{local_profile_id}` | Deletes a local profile.                      |
| `GET`    | `/api/player-data/local-profiles/default`            | Reads the persisted default identity.         |
| `PUT`    | `/api/player-data/local-profiles/default`            | Persists Guest or a local profile as default. |

### List local profiles

```text
GET /api/player-data/local-profiles
```

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

The handler calls `Runtime.ListLocalProfiles()` and returns profiles in the order supplied by the local store. The embedded SQLite store currently orders profiles by `created_at ASC`.

If the runtime is missing or the active store does not support local profile management, the handler returns:

```json
{
  "error": "local_profiles_unavailable"
}
```

with HTTP `503`.

### Create local profile

```text
POST /api/player-data/local-profiles
```

Request body:

```json
{
  "display_name": "Pilot_1",
  "seed_from_guest_stats": true
}
```

The handler trims `display_name` and accepts only non-empty values matching:

```text
^[A-Za-z0-9_-]+$
```

The handler generates `local_profile_id` server-side using 16 random bytes formatted as:

```text
local-profile-<hex>
```

When `seed_from_guest_stats` is `false`, the new profile starts with zero stats.

When `seed_from_guest_stats` is `true`, the handler asks the runtime for Guest stats and passes those stats into local profile creation. The handler does not read guest storage directly.

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

Invalid JSON, empty display names, or display names outside the accepted pattern return:

```json
{
  "error": "invalid_request"
}
```

with HTTP `400`.

### Update local profile display name

```text
PUT /api/player-data/local-profiles/{local_profile_id}
```

Request body:

```json
{
  "display_name": "Pilot_2"
}
```

The handler requires a non-empty `local_profile_id` path value and validates `display_name` with the same rule used for creation.

Only `display_name` changes. The profile identity and stats remain attached to the same `local_profile_id`.

Successful response:

```json
{
  "profile": {
    "local_profile_id": "local-profile-...",
    "display_name": "Pilot_2"
  }
}
```

Missing local profiles return:

```json
{
  "error": "local_profile_not_found"
}
```

with HTTP `404`.

### Delete local profile

```text
DELETE /api/player-data/local-profiles/{local_profile_id}
```

The handler requires a non-empty `local_profile_id` path value and delegates deletion to `Runtime.DeleteLocalProfile()`.

The embedded SQLite store deletes:

```text
local_profiles row
local_player_stats row
local_player_match_results rows
```

If the deleted profile is the current default local profile, the embedded SQLite store resets the default to Guest.

Successful deletion returns HTTP `204` with no JSON body.

Missing local profiles return `local_profile_not_found` with HTTP `404`.

Unavailable local profile storage returns `local_profiles_unavailable` with HTTP `503`.

### Get default local profile

```text
GET /api/player-data/local-profiles/default
```

Successful response:

```json
{
  "default_profile": {
    "identity_kind": "guest",
    "local_profile_id": "",
    "display_name": "Guest"
  }
}
```

or:

```json
{
  "default_profile": {
    "identity_kind": "local_profile",
    "local_profile_id": "local-profile-...",
    "display_name": "Pilot_1"
  }
}
```

The embedded SQLite store returns Guest when no default row exists, when the stored default is invalid, or when the stored local profile no longer exists.

### Set default local profile

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

Attempting to set a missing local profile as default returns `local_profile_not_found` with HTTP `404`.

### Error responses

The handler returns JSON error bodies for failed requests:

```json
{
  "error": "invalid_request"
}
```

Current local profile handler errors:

| Error                        | HTTP status | Meaning                                                                                               |
| ---------------------------- | ----------: | ----------------------------------------------------------------------------------------------------- |
| `method_not_allowed`         |       `405` | Unsupported HTTP method reached the handler.                                                          |
| `invalid_request`            |       `400` | Invalid JSON, invalid path value, invalid identity kind/default combination, or invalid display name. |
| `local_profile_not_found`    |       `404` | A requested local profile does not exist.                                                             |
| `local_profiles_unavailable` |       `503` | Runtime is missing or local profile storage is unavailable.                                           |

## Data ownership

The local profiles HTTP API owns request-time mutation of local profile management data through the runtime seam.

Data crossing the HTTP boundary:

```text
local_profile_id
display_name
identity_kind
seed_from_guest_stats
profiles[]
default_profile
```

Runtime/store-owned data affected by these routes:

```text
local profile records
local profile default selection
local profile stats seed rows
local match-result rows during delete cleanup
```

The embedded SQLite store currently owns these physical tables:

```text
local_profiles
local_profile_default
local_player_stats
local_player_match_results
```

The HTTP handler does not know table names, database paths, SQL statements, or schema setup. Those belong to `playerdata/embeddedsqlite`.

In the standard no-tag development build, the game-server composition root injects the SQLite store and initializes schema before serving routes. In `noembeddedsqlite` builds, the local store path is empty and local profile management is unavailable.

## Code map

Primary handler implementation:

```text
services/player-data/httpapi/local_profiles_handler.go
```

Runtime and store interfaces:

```text
services/player-data/playerdata/runtime.go
services/player-data/playerdata/store.go
services/player-data/playerdata/store_router.go
services/player-data/playerdata/configured_runtime.go
services/player-data/playerdata/noop_store.go
```

Embedded local persistence:

```text
services/player-data/playerdata/embeddedsqlite/sqlite_store.go
```

Game-server hosting and build-tag composition:

```text
services/game-server/cmd/game-server/main.go
services/game-server/cmd/game-server/player_data_http.go
services/game-server/cmd/game-server/player_data_local_store_dev.go
services/game-server/cmd/game-server/player_data_local_store_noembeddedsqlite.go
```

Client consumers:

```text
client/scripts/profile/local_pilot_api_client.gd
client/scripts/api/api_config.gd
client/scripts/ui/menu_flow/local_pilot_flow.gd
```

HTTP contract source:

```text
shared/contracts/http/openapi.yaml
```

Important non-ownership boundaries:

```text
services/game-server/cmd/game-server/
= hosts routes and injects runtime dependencies

client/scripts/profile/
= consumes API responses and owns client-side local pilot selection state

shared/contracts/http/openapi.yaml
= owns HTTP request and response shapes

services/player-data/playerdata/embeddedsqlite/
= owns physical local SQLite schema and persistence behavior

services/api-server/
= owns Rails auth/account APIs, not local SQLite profile management
```

## Tests

Relevant tests:

```text
services/player-data/httpapi/local_profiles_handler_test.go
services/player-data/playerdata/configured_runtime_test.go
services/player-data/playerdata/store_router_test.go
services/player-data/playerdata/noop_store_test.go
services/player-data/playerdata/embeddedsqlite/sqlite_store_test.go
services/game-server/cmd/game-server/
```

Covered behavior includes:

* local profile routes return `local_profiles_unavailable` when the local profile store is missing
* create with `seed_from_guest_stats = false` creates zero stats
* create with `seed_from_guest_stats = true` copies current Guest stats into the new profile
* configured runtime defaults to a noop local store when no SQLite path is configured
* configured runtime uses an injected local store when SQLite path and local store factory are provided
* embedded SQLite initializes local profile tables
* embedded SQLite does not persist `wins` in local player stats
* embedded SQLite deletes local profile, local stats, and local match-result rows together
* deleting the default local profile resets default selection to Guest
* embedded SQLite persists stats across reopen
* embedded SQLite rejects missing profile IDs and missing display names at the store boundary

Useful verification commands for changes in this area:

```text
cd services/player-data && go test ./...
cd services/player-data && go test -tags noembeddedsqlite ./...
cd services/game-server && go test -buildvcs=false ./cmd/game-server
cd services/game-server && go test -tags noembeddedsqlite -buildvcs=false ./cmd/game-server
```

When HTTP request or response shapes change, also update `shared/contracts/http/openapi.yaml` and run the relevant API contract validation.

## Related docs

* [Player Data](./!INDEX.md)
* [Services](../!INDEX.md)
* [Player Data HTTP Hosting](../game-server/integrations/player-data-http-hosting.md)
* [Client HTTP API Flow](../client/client-http-api-flow.md)
* [Local Pilot Flow](../client/pregame-menu-flow/local-pilot-flow.md)
* [HTTP Contract Enforcement](../../protocol/http-contract-enforcement.md)

## Notes

This document is scoped to local profile management routes. The normalized profile read endpoint, match-result sinks, profile stats flow, and runtime/store routing should remain separate player-data service docs rather than being folded into this API document.
