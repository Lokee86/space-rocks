# Player Stats And Match Results

Parent index: [API Server](./!README.md)

## Purpose

This document describes the API-server implementation responsibility for authenticated-account player stats and match-result persistence.

It keeps the Rails/Postgres persistence boundary separate from game-server simulation authority, player-data route selection, local-profile storage, guest transient stats, and client match-results presentation.

## Overview

The API server owns the Rails/Postgres persistence path for authenticated-account stats.

The current implemented flow is:

* game-server match resolution
* services/player-data `RecordMatchResult` command
* `RailsStore` authenticated-account route
* `POST /internal/player-data/match-results`
* Rails transaction creates `PlayerMatchResult` and updates `PlayerStat`
* response returns normalized player stats

The API server does not decide match outcome. It accepts resolved match-result facts from the internal player-data caller after the game-server has already built an authoritative match summary.

The API server persists two related records:

* `PlayerStat` - one aggregate stats row per Rails user.
* `PlayerMatchResult` - one persisted match-result row per submitted `result_id`.

`result_id` is the idempotency key for match-result writes. A duplicate `result_id` is accepted and returns `duplicate: true` without applying the stats again.

Stats reads use two Rails surfaces:

* `GET /api/player/stats` - authenticated public/current-user stats read.
* `POST /api/internal/player-data/stats` - internal player-data stats read by `account_id`.

Both stats-read paths create a zeroed `PlayerStat` row when the user exists but has no stats yet.

## Code root

* `services/api-server/`

## Responsibilities

* Persist authenticated-account aggregate stats in Rails/Postgres.
* Persist authenticated-account match-result rows in Rails/Postgres.
* Accept trusted internal match-result writes from the player-data Rails adapter.
* Load authenticated-account stats for internal player-data callers by `account_id`.
* Load current-user stats for authenticated public API callers.
* Create zeroed stats rows on first stats read or first match-result write.
* Serialize only the normalized V1 stats fields in API responses.
* Enforce internal bearer-token protection on internal player-data endpoints.
* Enforce OpenAPI request/response contract coverage in tests.

## Does not own

* Game simulation.
* Match lifecycle.
* Match-over detection.
* Score calculation during gameplay.
* Winner resolution before persistence.
* Guest transient stats.
* Local Profile stats or embedded SQLite storage.
* Player-data identity route selection.
* Player-data packet dispatch.
* Client profile readout routing.
* Client match-results presentation.
* Leaderboard ranking or public match-history presentation.
* Shared logical player-data schema ownership.

## Domain roles

The API server participates in the account-backed player-data path.

Current roles:

* Authenticated-account persistence owner.
* Rails/Postgres physical schema owner for account stats and match-result rows.
* Internal stats-read target for `services/player-data` authenticated-account routing.
* Internal match-result write target for `services/player-data` authenticated-account routing.
* Current-user stats read target for authenticated public API callers.

The game-server remains the gameplay authority. `services/player-data` remains the routing boundary that decides whether a request goes to guest memory, local profile storage, or Rails/API-backed authenticated-account persistence.

## Protocols and APIs

This surface exposes the authenticated-account stats and match-result endpoints used by the player-data service and by Rails-authenticated callers. `services/player-data` and the game-server consume these endpoints for account-backed stats reads and match-result writes, while Rails owns the authenticated-account persistence authority behind them. The boundary carries account IDs, stats payloads, and match-result fields; it does not own guest/local-profile routing, gameplay authority, or client presentation state.

### `GET /api/player/stats`

Public authenticated endpoint for the current account.

Behavior:

* Requires a user bearer token.
* Uses `current_user`.
* Creates a zeroed `PlayerStat` row if absent.
* Returns `{ stats: ... }`.
* Does not expose Rails internal IDs, credentials, token digests, email, timestamps, or password data.

### `POST /api/internal/player-data/stats`

Internal service-to-service endpoint for authenticated-account stats reads.

Behavior:

* Requires `Authorization: Bearer <GAME_SERVER_INTERNAL_TOKEN>`.
* Requires `account_id`.
* Looks up `User` by `account_id`.
* Creates a zeroed `PlayerStat` row if absent.
* Returns `{ stats: ... }`.

Failure responses:

* `401` for missing, malformed, or incorrect internal bearer token.
* `404` with `error: "unknown_user"` when `account_id` does not match a Rails user.
* `422` with `error: "invalid_input"` when `account_id` is missing.

### `POST /internal/player-data/match-results`

Internal service-to-service endpoint for authenticated-account match-result writes.

Behavior:

* Requires `Authorization: Bearer <GAME_SERVER_INTERNAL_TOKEN>`.
* Requires `result_id`, `match_id`, and `account_id`.
* Looks up `User` by `account_id`.
* Applies the match result through `PlayerStats::ApplyMatchResult`.
* Creates `PlayerStat` if absent.
* Creates `PlayerMatchResult` if `result_id` has not been seen.
* Updates aggregate stats in the same transaction.
* Returns `accepted`, `duplicate`, and serialized `stats`.

Failure responses:

* `401` for missing, malformed, or incorrect internal bearer token.
* `404` with `accepted: false, error: "unknown_user"` when `account_id` does not match a Rails user.
* `422` with `accepted: false, error: "invalid_input"` when required fields are missing or the service rejects input.

### Contract source

`shared/contracts/http/openapi.yaml` owns the HTTP request and response shapes.

Rails controllers implement those shapes. Rails tests enforce them through OpenAPI contract assertions.

## Data ownership

### `users.account_id`

`account_id` is the canonical cross-system authenticated-account UUID.

Rails `users.id` remains an internal database foreign key. Player-data callers address authenticated-account stats through `account_id`, not Rails `user_id`.

### `player_stats`

`player_stats` stores one aggregate stats row per user.

Current fields:

* `user_id`
* `total_score`
* `high_score`
* `ship_deaths`
* `games_played`
* `wins`
* timestamps

Current rules:

* `user_id` is unique.
* numeric stats must be greater than or equal to zero.
* default values are zero.
* `wins` is stored for authenticated-account stats.

### `player_match_results`

`player_match_results` stores accepted match-result rows.

Current fields:

* `result_id`
* `match_id`
* `user_id`
* `score`
* `ship_deaths`
* `won`
* timestamps

Current rules:

* `result_id` is required and unique.
* `match_id` is required.
* `match_id` is indexed.
* `score` and `ship_deaths` must be non-negative integers.
* `won` must be boolean.

### Aggregation rules

`PlayerStats::ApplyMatchResult` currently applies a new result by:

* incrementing `games_played` by `1`
* adding result `score` to `total_score`
* setting `high_score` to the max of existing high score and result score
* adding result `ship_deaths` to aggregate `ship_deaths`
* incrementing `wins` by `1` when `won` is true

Duplicate `result_id` submissions do not apply these aggregation rules again.

## Code map

Primary API-server files:

* `services/api-server/config/routes.rb`
* `services/api-server/app/controllers/internal/base_controller.rb`
* `services/api-server/app/controllers/api/internal/base_controller.rb`
* `services/api-server/app/controllers/api/player/stats_controller.rb`
* `services/api-server/app/controllers/api/internal/player_data/stats_controller.rb`
* `services/api-server/app/controllers/internal/player_data/match_results_controller.rb`
* `services/api-server/app/services/player_stats/apply_match_result.rb`
* `services/api-server/app/services/player_stats/serialize_stats.rb`
* `services/api-server/app/models/user.rb`
* `services/api-server/app/models/player_stat.rb`
* `services/api-server/app/models/player_match_result.rb`

Database schema files:

* `services/api-server/db/migrate/20260608000800_create_player_stats.rb`
* `services/api-server/db/migrate/20260608000900_create_player_match_results.rb`
* `services/api-server/db/migrate/20260608001000_add_account_id_to_users.rb`
* `services/api-server/db/schema.rb`

Contract/source files:

* `shared/contracts/http/openapi.yaml`
* `shared/player_data/stats.toml`
* `shared/player_data/match_result.toml`

Important non-ownership boundaries:

* `services/player-data/playerdata/rails_store.go` consumes the Rails internal stats and match-results endpoints for authenticated-account routing.
* `services/player-data/playerdata/store_router.go` owns identity-based store selection.
* `services/player-data/protocol/packets.go` contains generated player-data packet structs.
* `services/game-server/internal/playerdata/` owns game-server match-summary types.
* `services/game-server/internal/matchreporting/` maps game-server match summaries into player-data record commands.

## Tests

API-server service tests:

* `services/api-server/test/services/player_stats/apply_match_result_test.rb`

API-server controller tests:

* `services/api-server/test/controllers/api/player/stats_controller_test.rb`
* `services/api-server/test/controllers/api/internal/player_data/stats_controller_test.rb`
* `services/api-server/test/controllers/internal/player_data/match_results_controller_test.rb`

API-server model tests:

* `services/api-server/test/models/player_stat_test.rb`
* `services/api-server/test/models/player_match_result_test.rb`

Contract tests:

* `services/api-server/test/contracts/openapi_contract_test.rb`
* `services/api-server/test/support/openapi_contract_assertions.rb`

Related non-API-server tests:

* `services/player-data/playerdata/rails_store_test.go`
* `services/player-data/playerdata/store_router_test.go`
* `services/player-data/playerdata/dispatcher_test.go`
* `services/game-server/internal/matchreporting/mapper_test.go`
* `services/game-server/internal/matchreporting/runtime_reporter_test.go`

## Related docs

* [API Server](./!README.md)
* [Player Data service](../player-data/!README.md)
* [Game Server service](../game-server/!README.md)
* [Client service](../client/!README.md)
* [Account And Identity Current State](../../domains/platform/account-and-identity-current-state.md)
* [HTTP contract enforcement](../../protocol/http-contract-enforcement.md) - Current HTTP request/response contract enforcement documentation.
* [Player Data HTTP API](../../protocol/stubs/player-data-http-api.md) - Stub: player-data HTTP API protocol documentation.
* [Player Data Schema](../../data/stubs/player-data-schema.md) - Stub: player-data schema documentation.

## Notes

This document intentionally stays at the API-server service boundary.

The broader identity-routing flow belongs in domain and player-data service documentation. The HTTP request/response contract belongs in protocol documentation. The shared logical stats and match-result schemas belong in data documentation.

The current API-server implementation does not expose a match-history read endpoint. It records match-result rows to support idempotency, aggregate stat updates, and future account-backed history or leaderboard features.
