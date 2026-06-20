# Match Result Sinks

Parent index: [Player Data](./!README.md)

## Purpose

This document describes how `services/player-data` receives match-result commands, validates identity and play-mode routing, and commits the result into the correct account, local-profile, or guest sink.

It keeps player-data sink behavior separate from game-server match authority, client match-results presentation, Rails/Postgres persistence internals, and the shared player-data schema source of truth.

## Overview

The player-data service is the routing and sink boundary for match-result commits.

The current implemented flow is:

```text
game-server resolved match summary
-> player_data_record_match_result packet
-> RuntimeSink.HandlePlayerDataCommand
-> Runtime.Handle
-> Dispatcher.Handle
-> ValidateModeIdentity
-> StoreRouter.RecordMatchResult
-> account, local, or guest backing store
-> player_data_record_match_result_result packet
```

The game server produces trusted match facts. Player-data does not recompute match outcome, score, deaths, winner state, or match-over eligibility. It validates that the submitted identity kind is allowed for the submitted play mode, chooses the backing route, delegates the write, and returns normalized stats plus duplicate status.

Current sink routes:

| Identity kind           | Allowed mode                            | Route         | Backing behavior                                                                                       |
| ----------------------- | --------------------------------------- | ------------- | ------------------------------------------------------------------------------------------------------ |
| `authenticated_account` | `multiplayer`, `multiplayer_simulation` | Account store | Rails internal match-results endpoint when configured; in-memory account store otherwise               |
| `local_profile`         | `single_player`                         | Local store   | Embedded SQLite in the standard no-tag development build; noop store when no local store is configured |
| `guest`                 | `single_player`                         | Guest store   | Singleton in-memory transient stats                                                                    |

The normal game-server path uses stable `result_id` values derived from match ID and game player ID. Player-data treats `result_id` as the idempotency key where the backing store supports duplicate tracking.

## Code root

`services/player-data/`

## Responsibilities

* Decode generated player-data packets.
* Dispatch `player_data_record_match_result` commands.
* Validate play-mode and identity-kind compatibility before store mutation.
* Route accepted commands by identity kind.
* Keep account, local-profile, and guest stats separate.
* Return `player_data_record_match_result_result` responses.
* Report validation failures as accepted-false packet responses.
* Report store failures as accepted-false packet responses.
* Preserve duplicate semantics returned by the selected store.
* Normalize returned stats through the generated `PlayerDataStats` shape.
* Keep local-profile match-result persistence behind the local store.
* Keep authenticated-account match-result persistence behind the Rails store adapter.
* Keep guest match-result stats transient and process-local.

## Does not own

* Game-server match-over detection.
* Game-server score, death, or winner calculation.
* Game-server result ID generation.
* Room lifecycle or retry policy.
* Client match-results presentation.
* Room snapshot `match_result` projection.
* Rails controller authorization.
* Rails/Postgres physical schema design.
* Embedded SQLite physical schema as a shared source of truth.
* OpenAPI request/response contract ownership.
* Leaderboards.
* Public match-history read APIs.
* Anti-cheat validation beyond current mode/identity gating.
* Account authentication or OAuth state.

## Domain roles

The player-data match-result sink participates in the platform/account and player-experience result-commit path.

Current roles:

* Player-data command receiver.
* Mode/identity validation gate.
* Identity-based store router.
* Authenticated-account adapter caller.
* Local-profile storage boundary.
* Guest transient stats holder.
* Normalized stats response producer.
* Duplicate result response carrier.

The game server remains gameplay authority. The API server remains authenticated-account persistence authority. The client remains presentation-only for match results.

## Protocols and APIs

The player-data match-result sink is consumed through the generated player-data packet surface, not through a public HTTP endpoint.

The current game-server process hosts the player-data runtime in-process. `matchreporting.RuntimeReporter` sends encoded player-data packets into `RuntimeSink.HandlePlayerDataCommand`, and `RuntimeSink` forwards them to `Runtime.Handle`. The boundary carries trusted match-result facts: result ID, match ID, play mode, identity, score, ship deaths, and winner flag. The boundary does not carry client presentation rows, room snapshot state, Rails user IDs, or raw database records.

### Command packet

```text
type: player_data_record_match_result
result_id: <stable idempotency key>
match_id: <authoritative match id>
identity:
  identity_kind: authenticated_account | local_profile | guest
  account_id: <account UUID, authenticated-account route only>
  local_profile_id: <local profile ID, local-profile route only>
context:
  play_mode: single_player | multiplayer | multiplayer_simulation
score: <authoritative score>
ship_deaths: <authoritative ship deaths>
won: <authoritative winner flag>
```

### Response packet

```text
type: player_data_record_match_result_result
accepted: true | false
duplicate: true | false
stats:
  total_score: <normalized total score>
  high_score: <normalized high score>
  ship_deaths: <normalized ship deaths>
  games_played: <normalized games played>
  wins: <normalized wins>
error_code: <machine-readable error, when rejected>
message: <diagnostic message, when rejected>
```

### Validation behavior

`Dispatcher.Handle` validates the `context.play_mode` and `identity.identity_kind` pair before calling the store.

Current mode policy:

* `single_player` allows `guest` and `local_profile`.
* `single_player` rejects `authenticated_account`.
* `multiplayer` allows `authenticated_account`.
* `multiplayer` rejects `guest` and `local_profile`.
* `multiplayer_simulation` allows `authenticated_account`.
* `multiplayer_simulation` rejects `guest` and `local_profile`.
* missing or unknown play modes are rejected.
* missing or unknown identity kinds are rejected.

A mode/identity rejection returns:

```text
accepted: false
duplicate: false
error_code: invalid_mode_identity
```

The backing store is not called when mode/identity validation fails.

### Store error behavior

If the selected store returns an error, the dispatcher returns:

```text
accepted: false
duplicate: false
error_code: store_error
message: <store error>
```

Transport-level decode errors and unknown packet types return Go errors instead of encoded accepted-false result packets.

## Data ownership

### Player-data-owned routing data

Player-data owns runtime routing by:

* `identity.identity_kind`
* `context.play_mode`
* `account_id` for authenticated-account identity
* `local_profile_id` for local-profile identity
* singleton guest identity for guest state

`IdentityKey` produces internal keys for generic in-memory stores:

```text
authenticated_account -> account:<account_id>
local_profile         -> local:<local_profile_id>
guest                 -> guest
```

### Match-result command fields

Player-data accepts these match-result facts from the game server:

* `result_id`
* `match_id`
* `score`
* `ship_deaths`
* `won`
* `identity`
* `context.play_mode`

Player-data does not calculate those values.

### Aggregated stats

The normalized stats shape is:

```text
total_score
high_score
ship_deaths
games_played
wins
```

Generic in-memory stores and guest memory apply new results by:

* incrementing `games_played`
* adding `score` to `total_score`
* setting `high_score` to the maximum seen score
* adding `ship_deaths`
* incrementing `wins` when `won` is true

The embedded SQLite local-profile store applies the same core local stats but does not persist local wins. Local-profile stats returned by the SQLite route set `wins` to `0`.

Authenticated-account wins are handled by the Rails/API-backed store and Rails aggregate logic.

### Account route

When `PLAYER_DATA_RAILS_BASE_URL` is configured, authenticated-account match results route through `RailsStore`.

`RailsStore.RecordMatchResult` requires:

* `identity_kind = authenticated_account`
* non-empty `account_id`
* non-empty `result_id`
* non-empty `match_id`
* non-empty internal token

It sends:

```text
POST /internal/player-data/match-results
Authorization: Bearer <PLAYER_DATA_RAILS_INTERNAL_TOKEN>
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

The Rails response controls accepted and duplicate status for the account route.

When `PLAYER_DATA_RAILS_BASE_URL` is not configured, the account route uses `MemoryStore`. This is a development/default fallback, not Rails/Postgres persistence.

### Local-profile route

In the standard no-tag development build, the game-server composition root injects the embedded SQLite local store into the player-data runtime.

The SQLite local store persists:

* `local_profiles`
* `local_player_stats`
* `local_player_match_results`

For match-result writes, the SQLite store:

* requires `identity_kind = local_profile`
* requires `local_profile_id`
* requires `result_id`
* requires `match_id`
* ensures local profile and local stats rows exist
* checks `local_player_match_results` by `result_id`
* returns duplicate status without reapplying stats if the result already exists
* inserts a local match-result row for new results
* updates aggregate local stats in the same transaction
* returns normalized local stats with `wins = 0`

When built with `noembeddedsqlite`, or when no SQLite path is configured, local-profile routing uses `NoopStore`. `NoopStore` accepts valid store calls without persistence and returns zero stats.

### Guest route

The guest route uses `GuestMemoryStore`.

Guest state is process-local and transient. It tracks:

* one aggregate `PlayerDataStats` value
* processed result IDs for duplicate detection

Guest data is not durable and is not written to SQLite or Rails.

## Idempotency and duplicate handling

`result_id` is the idempotency key.

Duplicate handling is store-specific:

* `MemoryStore` tracks processed result IDs and returns the stats associated with the original stored identity key.
* `GuestMemoryStore` tracks processed result IDs and returns current guest stats on duplicates.
* SQLite stores `result_id` in `local_player_match_results` and returns existing local aggregate stats on duplicates.
* Rails/Postgres owns authenticated-account duplicate handling behind `/internal/player-data/match-results`.
* `NoopStore` does not track duplicates and always returns zero stats with `duplicate = false`.

A duplicate that is accepted by the backing store is surfaced as:

```text
accepted: true
duplicate: true
```

The game-server reporter treats accepted duplicates as successful reports.

## Code map

Primary player-data runtime files:

* `services/player-data/playerdata/runtime.go`
* `services/player-data/playerdata/runtime_sink.go`
* `services/player-data/playerdata/dispatcher.go`
* `services/player-data/playerdata/store.go`
* `services/player-data/playerdata/store_router.go`
* `services/player-data/playerdata/mode_policy.go`
* `services/player-data/playerdata/identity.go`
* `services/player-data/playerdata/configured_runtime.go`
* `services/player-data/playerdata/default_runtime.go`

Backing store files:

* `services/player-data/playerdata/rails_store.go`
* `services/player-data/playerdata/embeddedsqlite/sqlite_store.go`
* `services/player-data/playerdata/guest_memory_store.go`
* `services/player-data/playerdata/memory_store.go`
* `services/player-data/playerdata/noop_store.go`

Generated packet files:

* `services/player-data/protocol/packets.go`

Source-of-truth files:

* `shared/packets/player_data.toml`
* `shared/player_data/match_result.toml`
* `shared/player_data/stats.toml`

Composition and caller files:

* `services/game-server/cmd/game-server/main.go`
* `services/game-server/cmd/game-server/player_data_http.go`
* `services/game-server/cmd/game-server/player_data_local_store_dev.go`
* `services/game-server/cmd/game-server/player_data_local_store_noembeddedsqlite.go`
* `services/game-server/internal/matchreporting/runtime_reporter.go`
* `services/game-server/internal/matchreporting/mapper.go`

Important non-ownership boundaries:

* `services/game-server/internal/rooms/` owns room match lifecycle and resolved summary storage.
* `services/game-server/internal/matchreporting/` owns mapping resolved match summaries into player-data record commands.
* `services/api-server/app/controllers/internal/player_data/match_results_controller.rb` owns authenticated-account HTTP intake.
* `services/api-server/app/services/player_stats/apply_match_result.rb` owns Rails/Postgres aggregate stat mutation.
* `services/api-server/app/models/player_match_result.rb` owns Rails match-result row validation and persistence.
* `services/game-server/internal/networking/room_snapshot.go` owns presentation-safe room snapshot output for client match results.

## Tests

Primary player-data tests:

* `services/player-data/playerdata/dispatcher_test.go`
* `services/player-data/playerdata/store_router_test.go`
* `services/player-data/playerdata/mode_policy_test.go`
* `services/player-data/playerdata/identity_test.go`
* `services/player-data/playerdata/runtime_test.go`
* `services/player-data/playerdata/default_runtime_test.go`
* `services/player-data/playerdata/configured_runtime_test.go`
* `services/player-data/playerdata/configured_runtime_embedded_sqlite_test.go`
* `services/player-data/playerdata/rails_store_test.go`
* `services/player-data/playerdata/embeddedsqlite/sqlite_store_test.go`
* `services/player-data/playerdata/guest_memory_store_test.go`
* `services/player-data/playerdata/memory_store_test.go`
* `services/player-data/playerdata/noop_store_test.go`
* `services/player-data/codec/codec_test.go`

Related game-server tests:

* `services/game-server/internal/matchreporting/mapper_test.go`
* `services/game-server/internal/matchreporting/runtime_reporter_test.go`
* `services/game-server/internal/rooms/room_match_summary_test.go`
* `services/game-server/internal/rooms/lifecycle_tick_test.go`

Related API-server tests:

* `services/api-server/test/controllers/internal/player_data/match_results_controller_test.rb`
* `services/api-server/test/services/player_stats/apply_match_result_test.rb`
* `services/api-server/test/models/player_match_result_test.rb`
* `services/api-server/test/models/player_stat_test.rb`

## Related docs

* [Player Data](./!README.md)
* [Game-server match result reporting](../game-server/integrations/match-result-reporting.md)
* [API-server player stats and match results](../api-server/player-stats-and-match-results.md)
* [API-server internal API surface](../api-server/internal-api-surface.md)
* [Account And Identity Current State](../../domains/platform/account-and-identity-current-state.md)
* [HTTP Contract Enforcement](../../protocol/http-contract-enforcement.md)
* [Player Data HTTP API](../../protocol/stubs/player-data-http-api.md) - Stub: player-data HTTP protocol surface.
* [Player Data Schema](../../data/stubs/player-data-schema.md) - Stub: shared player-data schema and source documentation.

## Notes

This document describes the player-data sink boundary only.

The client match-results window uses presentation-safe room snapshot data. That is separate from player-data match-result sink routing.

The current player-data runtime is hosted in-process by the game-server executable. The service boundary still exists at the packet/runtime-sink layer, and the game server still does not write SQLite or Rails/Postgres player-data tables directly.

`wins` is present in the normalized stats packet shape, but current local SQLite stats intentionally return `wins = 0`. Authenticated-account wins are persisted by Rails.
