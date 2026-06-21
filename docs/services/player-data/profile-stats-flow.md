## Profile Stats Flow

Parent index: [Player Data](./!INDEX.md)

## Purpose

This document describes how the player-data service loads, normalizes, routes, and mutates profile stats.

It covers the service-owned stats flow behind profile readout, local profile stat seeding, match-result-driven stat updates, identity-based store routing, and the stats surfaces consumed by the game server, client, and API server.

## Overview

Profile stats are owned by the player-data service runtime, not by client presentation code and not by game-server simulation code.

The current implementation has two main stats directions:

```text
Profile read
client -> hosted player-data HTTP profile handler -> player-data runtime -> selected store -> normalized profile response

Stats write
game-server match result reporter -> player-data packet/runtime sink -> dispatcher -> selected store -> updated stats response
```

The player-data runtime exposes one logical stats contract across all identity kinds:

```text
total_score
high_score
ship_deaths
games_played
wins
```

The backing behavior depends on identity:

```text
guest                 -> guest memory store
local_profile         -> embedded SQLite local store in standard builds
authenticated_account -> Rails adapter to API-server/Postgres when configured
```

The client uses profile stats for readout display. It does not count gameplay facts, mutate aggregate stats, choose backing stores, or call Rails stats endpoints directly for profile readout.

The game server owns authoritative gameplay facts and match-result summaries. It sends resolved match facts into player-data, but it does not update profile stats tables directly.

## Code root

```text
services/player-data/
```

## Responsibilities

The player-data profile stats flow owns:

* loading stats through `Runtime.LoadStats`
* routing stats reads by `identity_kind`
* accepting generated `player_data_load_stats` packets through the dispatcher
* returning generated `player_data_load_stats_result` packets
* accepting generated `player_data_record_match_result` packets
* routing match-result writes by `identity_kind`
* updating aggregate stats from accepted match results
* returning updated stats after accepted match-result writes
* rejecting invalid mode/identity pairs in the packet dispatch path
* loading authenticated-account stats through the Rails adapter
* loading local-profile stats from embedded SQLite in standard builds
* loading guest stats from in-process guest memory
* returning zero stats from the HTTP profile handler when `LoadStats` reports no found stats
* providing guest stats as the optional seed source for local profile creation
* keeping the logical stats contract stable across stores even when physical storage differs

## Does not own

The profile stats flow does not own:

* game-server match simulation
* game-server score calculation
* game-server player death counting
* game-server room match-over decisions
* client profile readout presentation
* client match-results presentation
* client-side stat mutation
* Rails/Postgres physical schema design
* Rails authenticated-account persistence internals
* public Rails `GET /api/player/stats` behavior
* OpenAPI contract ownership
* local profile selector UI behavior
* progression, unlocks, inventory, achievements, rewards, leaderboards, or match-history reads

## Domain roles

The profile stats flow participates in the player-data portion of the player experience and platform domains.

Current roles:

* stats read boundary for pregame profile readout
* stats aggregation boundary for match-result commits
* identity-based routing boundary for guest, local-profile, and authenticated-account stats
* local durable stats owner for Local Profile
* guest transient stats owner for Guest
* Rails adapter for authenticated-account stats
* normalization seam that lets the client display the same logical stats shape regardless of backing store

Identity behavior:

| Identity kind           | Read behavior                             | Write behavior                                   | Backing behavior                             |
| ----------------------- | ----------------------------------------- | ------------------------------------------------ | -------------------------------------------- |
| `guest`                 | loads singleton guest stats               | updates singleton guest stats from match results | in-process transient memory                  |
| `local_profile`         | loads local profile stats                 | updates local profile stats from match results   | embedded SQLite in standard builds           |
| `authenticated_account` | loads account stats through Rails adapter | forwards match results through Rails adapter     | API-server/Postgres when Rails is configured |

## Protocols and APIs

The profile stats flow has three service surfaces: hosted HTTP profile reads, generated player-data packets, and store calls.

### Hosted HTTP profile read

The game server currently hosts the player-data profile handler at:

```text
POST /api/player-data/profile
```

The handler is implemented by:

```text
services/player-data/httpapi/profile_handler.go
```

The HTTP surface exists for client profile readout. The client sends the active play mode and identity context, and the handler resolves the profile identity before loading stats through the player-data runtime. The player-data handler owns request handling and response shape; the game server only hosts the route and supplies runtime/auth dependencies.

Request fields:

```text
play_mode
identity_kind
local_profile_id
```

Authenticated account requests also require a bearer token in the `Authorization` header. The handler verifies that token through the injected auth verifier and receives the authenticated `account_id` and display name from that verifier.

Profile identity resolution:

| Identity kind           | Required input               | Callsign                         | Activity status | Runtime identity                                                 |
| ----------------------- | ---------------------------- | -------------------------------- | --------------- | ---------------------------------------------------------------- |
| `guest`                 | none                         | `Guest`                          | `OFFLINE`       | `identity_kind=guest`                                            |
| `local_profile`         | non-empty `local_profile_id` | `Local Pilot`                    | `LOCAL`         | `identity_kind=local_profile`, `local_profile_id=<id>`           |
| `authenticated_account` | valid bearer token           | verifier display name or `Pilot` | `ACTIVE`        | `identity_kind=authenticated_account`, `account_id=<account_id>` |

Response shape:

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

Error behavior:

| Condition                                             | Status | Error                 |
| ----------------------------------------------------- | ------ | --------------------- |
| non-POST method                                       | `405`  | `method_not_allowed`  |
| nil runtime                                           | `500`  | `profile_unavailable` |
| malformed JSON                                        | `400`  | `invalid_request`     |
| unsupported `play_mode`                               | `400`  | `invalid_request`     |
| unsupported `identity_kind`                           | `400`  | `invalid_request`     |
| missing local profile ID for local profile identity   | `400`  | `invalid_request`     |
| missing, invalid, or unverifiable authenticated token | `401`  | `unauthorized`        |
| runtime stats load error                              | `500`  | `profile_unavailable` |

If `Runtime.LoadStats` succeeds but returns `found=false`, the HTTP profile handler returns zero stats.

### Runtime stats load

The core runtime surface is:

```text
Runtime.LoadStats(identity protocol.PlayerDataIdentity) (protocol.PlayerDataStats, bool, error)
```

`Runtime.LoadStats` delegates to the configured store. Store routing is usually provided by `StoreRouter`, which selects the store from `identity.IdentityKind`.

Direct `Runtime.LoadStats` calls do not apply play-mode policy on their own. Callers that need play-mode validation must validate before calling directly or use the generated packet dispatch path.

### Generated load-stats packet

The generated packet read surface is:

```text
player_data_load_stats
```

Input shape:

```text
type: player_data_load_stats
identity:
  identity_kind: guest | local_profile | authenticated_account
  account_id: <authenticated account id, when authenticated_account>
  local_profile_id: <local profile id, when local_profile>
context:
  play_mode: single_player | multiplayer | multiplayer_simulation
```

Output shape:

```text
type: player_data_load_stats_result
found: true | false
stats: PlayerDataStats
error_code: <empty or error code>
message: <empty or diagnostic message>
```

The dispatcher validates play mode and identity through `ValidateModeIdentity` before store access. Invalid mode/identity pairs return `error_code: invalid_mode_identity` and do not call the store.

Allowed pairs:

| Play mode                | Guest    | Local profile | Authenticated account |
| ------------------------ | -------- | ------------- | --------------------- |
| `single_player`          | allowed  | allowed       | rejected              |
| `multiplayer`            | rejected | rejected      | allowed               |
| `multiplayer_simulation` | rejected | rejected      | allowed               |

### Generated match-result packet

Stats are mutated from match-result commits, not from profile reads.

The generated write surface is:

```text
player_data_record_match_result
```

Input shape:

```text
type: player_data_record_match_result
result_id: <idempotency key>
match_id: <match id>
identity:
  identity_kind: guest | local_profile | authenticated_account
  account_id: <authenticated account id, when authenticated_account>
  local_profile_id: <local profile id, when local_profile>
context:
  play_mode: single_player | multiplayer | multiplayer_simulation
score: <authoritative score>
ship_deaths: <authoritative ship deaths>
won: true | false
```

Output shape:

```text
type: player_data_record_match_result_result
accepted: true | false
duplicate: true | false
stats: PlayerDataStats
error_code: <empty or error code>
message: <empty or diagnostic message>
```

The dispatcher validates play mode and identity before store access. Store errors return `accepted=false` with `error_code: store_error`.

## Data ownership

### Logical stats contract

The logical stats contract is generated into `services/player-data/protocol/packets.go` and currently includes:

```text
total_score
high_score
ship_deaths
games_played
wins
```

The logical schema source is:

```text
shared/player_data/stats.toml
```

The packet source is:

```text
shared/packets/player_data.toml
```

### Aggregation rules

Stores update aggregate stats from accepted match results using the same logical rules:

```text
games_played += 1
total_score += score
high_score = max(high_score, score)
ship_deaths += ship_deaths
wins += 1 when won is true
```

Duplicate handling depends on the store, but the logical contract is the same: duplicate `result_id` submissions return existing stats and do not apply the result a second time.

### Guest stats

Guest stats are transient and process-local.

`GuestMemoryStore` keeps one singleton stats aggregate and an in-memory set of processed `result_id` values. Guest stats are available during the current process lifetime and are not account-shaped durable data.

Guest stats can seed a newly created local profile when local profile creation passes `seed_from_guest_stats=true`. That seed operation calls `Runtime.LocalProfileSeedStats(true)`, which loads the guest stats through the runtime instead of reading guest storage directly.

### Local profile stats

Local profile stats are durable in standard builds where embedded SQLite is compiled in.

The embedded SQLite store owns:

```text
local_profiles
local_profile_default
local_player_stats
local_player_match_results
```

Local profile stats fields stored in SQLite:

```text
local_profile_id
total_score
high_score
ship_deaths
games_played
created_at
updated_at
```

Local profile stats intentionally do not store `wins`. Reads set `wins` to `0` in the returned logical stats payload.

Local profile match-result writes insert a row into `local_player_match_results` and update `local_player_stats` in the same transaction. Duplicate writes are detected by `result_id` in `local_player_match_results`.

When a local profile is deleted, the embedded SQLite store deletes its profile row, stats row, and match-result rows. If that local profile was the default, the default resets to Guest.

In `noembeddedsqlite` builds, the local SQLite path is empty and the configured local store path is unavailable. Local profile management surfaces return unavailable behavior through the local profile store interface.

### Authenticated account stats

Authenticated account stats are loaded and written through `RailsStore` when `PLAYER_DATA_RAILS_BASE_URL` is configured.

Stats read path:

```text
RailsStore.LoadStats
-> POST /api/internal/player-data/stats
```

Match-result write path:

```text
RailsStore.RecordMatchResult
-> POST /internal/player-data/match-results
```

Both paths require `PLAYER_DATA_RAILS_INTERNAL_TOKEN`. `RailsStore` sends the token as an internal bearer token for internal Rails endpoints.

The Rails adapter validates that authenticated account requests include:

```text
identity_kind = authenticated_account
account_id
```

Match-result writes also require:

```text
result_id
match_id
```

The API server owns Rails/Postgres persistence and returns normalized stats to the player-data service.

### In-memory account fallback

When `PLAYER_DATA_RAILS_BASE_URL` is not configured, `NewConfiguredRuntime` uses `MemoryStore` for the account route. This supports local/test operation without Rails-backed account persistence.

`MemoryStore` stores stats by logical identity key and tracks processed result IDs in memory. It is not durable account persistence.

## Code map

Primary player-data runtime files:

```text
services/player-data/playerdata/runtime.go
services/player-data/playerdata/store.go
services/player-data/playerdata/store_router.go
services/player-data/playerdata/dispatcher.go
services/player-data/playerdata/mode_policy.go
services/player-data/playerdata/identity.go
services/player-data/playerdata/runtime_sink.go
services/player-data/playerdata/configured_runtime.go
services/player-data/playerdata/default_runtime.go
```

Store implementations:

```text
services/player-data/playerdata/guest_memory_store.go
services/player-data/playerdata/memory_store.go
services/player-data/playerdata/noop_store.go
services/player-data/playerdata/rails_store.go
services/player-data/playerdata/embeddedsqlite/sqlite_store.go
```

HTTP profile surface:

```text
services/player-data/httpapi/profile_handler.go
services/player-data/httpapi/local_profiles_handler.go
```

Generated/source contract files:

```text
shared/player_data/stats.toml
shared/player_data/match_result.toml
shared/packets/player_data.toml
services/player-data/protocol/packets.go
```

Game-server integration files that host or call the stats flow:

```text
services/game-server/cmd/game-server/main.go
services/game-server/cmd/game-server/player_data_http.go
services/game-server/cmd/game-server/player_data_local_store_dev.go
services/game-server/cmd/game-server/player_data_local_store_noembeddedsqlite.go
services/game-server/internal/matchreporting/runtime_reporter.go
services/game-server/internal/matchreporting/mapper.go
```

Important non-ownership boundaries:

```text
services/game-server/internal/rooms/
services/game-server/internal/playerdata/
services/api-server/app/controllers/api/internal/player_data/stats_controller.rb
services/api-server/app/controllers/internal/player_data/match_results_controller.rb
services/api-server/app/services/player_stats/apply_match_result.rb
client/scripts/profile/profile_stats_provider.gd
client/scripts/profile/player_data_profile_api_client.gd
```

## Tests

Primary player-data tests:

```text
services/player-data/playerdata/runtime_test.go
services/player-data/playerdata/store_router_test.go
services/player-data/playerdata/dispatcher_test.go
services/player-data/playerdata/mode_policy_test.go
services/player-data/playerdata/memory_store_test.go
services/player-data/playerdata/guest_memory_store_test.go
services/player-data/playerdata/noop_store_test.go
services/player-data/playerdata/rails_store_test.go
services/player-data/playerdata/configured_runtime_test.go
services/player-data/playerdata/default_runtime_test.go
services/player-data/playerdata/embeddedsqlite/sqlite_store_test.go
services/player-data/httpapi/local_profiles_handler_test.go
```

Covered behavior includes:

* runtime rejection of nil stores
* direct `Runtime.LoadStats` delegation
* guest stat seeding through `Runtime.LocalProfileSeedStats`
* identity-based store routing for stats reads and match-result writes
* generated load-stats packet handling
* generated record-match-result packet handling
* mode/identity validation in the dispatcher path
* duplicate match-result handling
* guest transient stat aggregation
* local SQLite stat row creation and persistence
* local SQLite duplicate detection
* local SQLite exclusion of `wins`
* local profile deletion removing stats and match-result rows
* Rails internal stats-read request shape
* Rails internal match-result write request shape
* internal bearer-token requirement in the Rails adapter

Related client tests:

```text
client/tests/unit/profile/test_profile_stats_provider.gd
client/tests/unit/profile/test_profile_context_provider.gd
client/tests/unit/profile/test_guest_transient_stats_provider.gd
```

Related API-server tests:

```text
services/api-server/test/controllers/api/internal/player_data/stats_controller_test.rb
services/api-server/test/controllers/internal/player_data/match_results_controller_test.rb
services/api-server/test/services/player_stats/apply_match_result_test.rb
services/api-server/test/models/player_stat_test.rb
services/api-server/test/models/player_match_result_test.rb
```

## Related docs

* [Player Data](./!INDEX.md)
* [Player Data HTTP Hosting](../game-server/integrations/player-data-http-hosting.md)
* [Match Result Reporting](../game-server/integrations/match-result-reporting.md)
* [Player Stats And Match Results](../api-server/player-stats-and-match-results.md)
* [Profile Flow](../client/pregame-menu-flow/profile-flow.md)
* [Player Data HTTP API](../../protocol/stubs/player-data-http-api.md) - Stub: player-data HTTP API protocol documentation.
* [Player Data Schema](../../data/stubs/player-data-schema.md) - Stub: player-data schema documentation.

## Notes

The encoded packet path and the hosted HTTP profile path are related but not identical. The dispatcher validates mode/identity pairs with `ValidateModeIdentity` before store access. The HTTP profile handler validates supported play-mode and identity-kind strings, resolves the identity, and then calls `Runtime.LoadStats` directly.

The hosted HTTP profile handler currently returns `Local Pilot` as the local-profile callsign. Local profile display-name management is owned by the local profile API flow, not by the profile stats read path.

Profile reads do not mutate stats. Aggregate stat mutation happens when authoritative match results are reported into the player-data runtime.
