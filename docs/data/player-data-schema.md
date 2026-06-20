# Player Data Schema

Parent index: [Data](./!README.md)

## Purpose

This document describes the current player-data logical schema sources, the implemented contracts that mirror them, the validation pipeline, and the persistence boundaries that consume them.

It exists to keep logical player-data schema ownership separate from HTTP request/response contracts, runtime packet schemas, Rails/Postgres physical tables, embedded SQLite physical tables, and game-server simulation state.

## Overview

Player-data schema is a logical contract for account-shaped player data. It defines the stable concepts that guest, local-profile, and authenticated-account paths expose through the player-data runtime.

The current logical schema sources are:

```text
shared/player_data/stats.toml
shared/player_data/match_result.toml
```

These files define logical player-data contracts. They do not define raw SQL, Rails migrations, HTTP payloads, or generated runtime packet structs.

Current logical schema scope:

```text
Stats
MatchResultSummary
PlayerMatchSummary
```

Current implemented storage routes consume the logical contract differently:

| Route                 | Runtime identity        | Backing behavior                                                                |
| --------------------- | ----------------------- | ------------------------------------------------------------------------------- |
| Guest                 | `guest`                 | process-local transient memory                                                  |
| Local Profile         | `local_profile`         | embedded SQLite in standard no-tag development builds                           |
| Authenticated Account | `authenticated_account` | Rails/API-backed Postgres through the player-data Rails adapter when configured |

The player-data runtime owns identity-based store routing. The game server owns authoritative gameplay facts and reports resolved match results into player-data. The API server owns authenticated-account physical persistence. Embedded SQLite owns local-profile physical persistence.

## Source files

### `shared/player_data/stats.toml`

`stats.toml` owns the logical aggregate stats contract.

Current schema metadata:

```text
schema_name = "stats"
schema_version = "v1.1"
```

Current fields:

| Field          | Type    | Default | Notes                                             |
| -------------- | ------- | ------- | ------------------------------------------------- |
| `total_score`  | integer | `0`     | Cumulative score across accepted match results.   |
| `high_score`   | integer | `0`     | Highest single-match score seen by the aggregate. |
| `ship_deaths`  | integer | `0`     | Cumulative authoritative ship deaths.             |
| `games_played` | integer | `0`     | Count of accepted match-result commits.           |
| `wins`         | integer | `0`     | Marked `multiplayer_only` in the schema source.   |

`wins` is part of the logical stats payload because authenticated-account multiplayer stats need it. Local-profile SQLite currently does not persist wins and returns `wins = 0` for local-profile stat reads.

### `shared/player_data/match_result.toml`

`match_result.toml` owns the logical match-result summary contract.

Current schema metadata:

```text
schema_name = "match_result"
schema_version = "v1.1"
```

Current match metadata:

```text
winner_rule = "multiplayer_highest_score"
ties_award_no_wins = true
```

Current `MatchResultSummary` fields:

| Field         | Type   | Rule     | Notes                                              |
| ------------- | ------ | -------- | -------------------------------------------------- |
| `match_id`    | string | required | Match identity carried into match-result writes.   |
| `mode`        | string | required | Logical match mode.                                |
| `resolved_at` | string | optional | Optional resolved timestamp in the logical schema. |

Current `PlayerMatchSummary` fields:

| Field              | Type    | Rule            | Notes                                           |
| ------------------ | ------- | --------------- | ----------------------------------------------- |
| `game_player_id`   | string  | required        | Game-server player identity inside the match.   |
| `account_id`       | string  | optional        | Authenticated-account identity.                 |
| `local_profile_id` | string  | optional        | Local-profile identity.                         |
| `score`            | integer | default `0`     | Authoritative score for the match result.       |
| `ship_deaths`      | integer | default `0`     | Authoritative death count for the match result. |
| `won`              | boolean | default `false` | Winner flag after winner resolution.            |

The game server resolves `won` before reporting match results. In single-player mode, the current resolver clears `won`. In multiplayer mode, the highest unique score receives `won = true`; tied highest scores award no wins.

## Logical schema boundaries

The player-data logical schema owns:

* aggregate stat field names and meanings
* match-result summary field names and meanings
* required, optional, and default field rules for schema validation
* cross-store expectations that each backing route must satisfy
* V1.1 contract naming for stats and match-result summaries

The player-data logical schema does not own:

* Rails/Postgres table layouts
* embedded SQLite table layouts
* Rails migrations
* SQLite schema initialization SQL
* HTTP request/response shapes
* runtime packet envelope shape
* generated packet structs
* client presentation labels
* game-server score calculation
* game-server room lifecycle
* live simulation state
* OAuth, bearer-token issuance, or account login data

## Related contract sources

### Runtime packet schema

The player-data runtime packet protocol is sourced from:

```text
shared/packets/player_data.toml
```

Generated output:

```text
services/player-data/protocol/packets.go
```

The generated packet structs include `PlayerDataStats`, `PlayerDataIdentity`, `PlayerDataLoadStats`, and `PlayerDataRecordMatchResult`. These packet structs mirror the current logical stats fields, but their source of truth is the packet schema under `shared/packets/`, not the logical schema under `shared/player_data/`.

Packet schema changes must use the packet workflow. Logical player-data schema changes must update `shared/player_data/*.toml` and then update any mirrored implementation or packet contract deliberately.

### HTTP contracts

HTTP request and response shapes are sourced from:

```text
shared/contracts/http/openapi.yaml
```

The OpenAPI contract owns hosted HTTP surfaces such as:

```text
POST /api/player-data/profile
GET /api/player-data/local-profiles
POST /api/player-data/local-profiles
PUT /api/player-data/local-profiles/{local_profile_id}
DELETE /api/player-data/local-profiles/{local_profile_id}
GET /api/player-data/local-profiles/default
PUT /api/player-data/local-profiles/default
POST /api/internal/player-data/stats
POST /internal/player-data/match-results
```

OpenAPI owns request/response payload shape. It does not own logical player-data schema meaning, store routing, or physical persistence.

## Implemented contract mapping

### Stats aggregation

Stores update aggregate stats from accepted match results using the same logical aggregation rules:

```text
games_played += 1
total_score += score
high_score = max(high_score, score)
ship_deaths += ship_deaths
wins += 1 when won is true and the backing route persists wins
```

Duplicate result handling is idempotent by `result_id`. Duplicate submissions return existing stats and do not apply the match result again.

### Guest route

Guest stats use `GuestMemoryStore`.

Current behavior:

* requires `identity_kind = guest`
* keeps one singleton aggregate stat object per runtime instance
* tracks processed result IDs in memory
* returns guest stats as found
* does not persist account-shaped data
* can seed a new local profile when local profile creation requests guest stat seeding

Guest stats satisfy the logical stats payload shape but are transient.

### Local-profile route

Local-profile stats use embedded SQLite in standard no-tag development builds.

Current local SQLite tables:

```text
local_profiles
local_profile_default
local_player_stats
local_player_match_results
```

Current local stats fields persisted in `local_player_stats`:

```text
local_profile_id
total_score
high_score
ship_deaths
games_played
created_at
updated_at
```

Current local match-result fields persisted in `local_player_match_results`:

```text
result_id
match_id
local_profile_id
score
ship_deaths
created_at
```

Local-profile SQLite does not persist `wins`. Local-profile stat reads return the logical stats payload with `wins = 0`.

The local store also owns local profile list, create, delete, display-name update, default profile read/write, and local match-result duplicate detection.

### Authenticated-account route

Authenticated-account stats use `RailsStore` when `PLAYER_DATA_RAILS_BASE_URL` is configured.

Current Rails read path:

```text
RailsStore.LoadStats
-> POST /api/internal/player-data/stats
```

Current Rails write path:

```text
RailsStore.RecordMatchResult
-> POST /internal/player-data/match-results
```

Both paths require `PLAYER_DATA_RAILS_INTERNAL_TOKEN`.

Rails/Postgres physical persistence includes:

```text
users.account_id
player_stats
player_match_results
```

`users.account_id` is the cross-system authenticated-account identifier. Rails `users.id` remains an internal database foreign key.

`player_stats` stores:

```text
user_id
total_score
high_score
ship_deaths
games_played
wins
timestamps
```

`player_match_results` stores:

```text
result_id
match_id
user_id
score
ship_deaths
won
timestamps
```

Rails owns the authenticated-account physical schema and persistence behavior. The logical player-data schema defines the contract those endpoints and tables must satisfy.

## Configuration

The data-sync config includes the player-data logical schema domain:

```text
tools/data_sync/config.toml

[sot.player_data]
paths = [
  "shared/player_data/stats.toml",
  "shared/player_data/match_result.toml",
]
```

The data-sync CLI currently supports `player_data` as a validation-only domain.

Supported:

```text
data-sync -validate -player_data
```

Not supported for `player_data`:

```text
data-sync -push -player_data ...
data-sync -pull -player_data ...
data-sync -diff -player_data ...
data-sync -check -player_data ...
```

The CLI rejects push, pull, diff, and check for `player_data` with the current implementation.

## Validation rules

The player-data TOML loader accepts these field types:

```text
string
integer
boolean
```

Each field must define at least one of:

```text
required
optional
default
```

Validation rejects:

* unreadable or missing player-data TOML files
* invalid TOML syntax
* missing schema name
* schemas with no field groups
* field groups with no fields
* missing field names
* unsupported field types
* defaults that do not match the declared field type
* invalid `required` or `optional` boolean values

Validation loads top-level `[fields.*]` groups and named groups such as:

```text
[MatchResultSummary.fields.*]
[PlayerMatchSummary.fields.*]
```

Metadata tables may exist beside field groups, but field validation is based on the supported field declarations.

## Pipeline usage

When changing only the logical player-data schema:

1. Edit the relevant source file under `shared/player_data/`.

2. Run:

   ```bash
   data-sync -validate -player_data
   ```

3. Update mirrored implementation contracts where needed.

4. Run the affected service tests.

When changing runtime packet payloads for player-data commands or results:

1. Edit `shared/packets/player_data.toml`.

2. Run:

   ```bash
   data-sync -validate -packets
   data-sync -diff -packets -go -gds
   data-sync -push -packets -go -gds
   data-sync -check -packets -go -gds
   ```

3. Review generated changes in `services/player-data/protocol/packets.go`.

4. Run player-data runtime tests and any affected game-server integration tests.

When changing HTTP payloads:

1. Edit `shared/contracts/http/openapi.yaml`.
2. Update Rails or hosted Go HTTP handlers as needed.
3. Run OpenAPI contract tests and affected service tests.

When changing physical storage:

1. Update the owning physical schema:

   * Rails/Postgres: Rails migrations under `services/api-server/db/migrate/`
   * Local SQLite: schema initialization in `services/player-data/playerdata/embeddedsqlite/sqlite_store.go`
2. Ensure the physical schema still satisfies the logical player-data contract.
3. Run the owning store and API tests.

## Consumers

Current consumers of the logical player-data schema include:

* `services/player-data/playerdata/` runtime and store-routing code
* `services/player-data/protocol/packets.go` packet structs that mirror stats payloads
* `services/player-data/playerdata/embeddedsqlite/sqlite_store.go` local-profile persistence
* `services/player-data/playerdata/rails_store.go` authenticated-account Rails adapter
* `services/player-data/httpapi/` hosted HTTP handlers
* `services/game-server/internal/playerdata/` match-summary types and winner resolution
* `services/game-server/internal/matchreporting/` match-result command mapping
* `services/api-server/app/controllers/api/internal/player_data/` internal stats read endpoint
* `services/api-server/app/controllers/internal/player_data/` internal match-result write endpoint
* `services/api-server/app/services/player_stats/` Rails stat serialization and mutation
* `services/api-server/db/migrate/` Rails/Postgres physical schema
* client profile and match-result presentation through HTTP/runtime outputs

## Failure modes

Common schema and contract failure modes:

* `shared/player_data/*.toml` changes without corresponding runtime, store, or API updates.
* Logical schema fields diverge from `shared/packets/player_data.toml`.
* `services/player-data/protocol/packets.go` becomes stale after packet schema edits.
* Rails serializers omit or rename logical stats fields.
* Rails migrations satisfy Rails locally but no longer satisfy the shared logical contract.
* Embedded SQLite stores a different local stats shape than the runtime returns.
* Local-profile persistence accidentally starts treating `wins` as durable local-profile state without a schema decision.
* HTTP OpenAPI changes are mistaken for logical schema changes.
* Physical database tables are treated as the cross-service source of truth.
* `data-sync -validate` is run without `-player_data`, leaving player-data logical schema validation out of the requested domain set.

## Code or source map

Logical schema sources:

```text
shared/player_data/stats.toml
shared/player_data/match_result.toml
```

Data-sync validation and configuration:

```text
tools/data_sync/config.toml
tools/data_sync/data_sync/cli.py
tools/data_sync/data_sync/config.py
tools/data_sync/data_sync/player_data_toml.py
tools/data_sync/data_sync/validate.py
tools/data_sync/tests/test_player_data_toml.py
tools/data_sync/tests/test_cli.py
tools/data_sync/tests/test_validate.py
```

Runtime packet source and generated output:

```text
shared/packets/player_data.toml
services/player-data/protocol/packets.go
```

Player-data runtime and stores:

```text
services/player-data/playerdata/store.go
services/player-data/playerdata/store_router.go
services/player-data/playerdata/identity.go
services/player-data/playerdata/mode_policy.go
services/player-data/playerdata/dispatcher.go
services/player-data/playerdata/runtime.go
services/player-data/playerdata/configured_runtime.go
services/player-data/playerdata/guest_memory_store.go
services/player-data/playerdata/memory_store.go
services/player-data/playerdata/noop_store.go
services/player-data/playerdata/rails_store.go
services/player-data/playerdata/embeddedsqlite/sqlite_store.go
```

Hosted player-data HTTP handlers:

```text
services/player-data/httpapi/profile_handler.go
services/player-data/httpapi/local_profiles_handler.go
```

Game-server match-result source and reporting:

```text
services/game-server/internal/playerdata/types.go
services/game-server/internal/playerdata/resolve.go
services/game-server/internal/playerdata/summary.go
services/game-server/internal/matchreporting/mapper.go
services/game-server/internal/matchreporting/runtime_reporter.go
```

API-server authenticated-account persistence:

```text
services/api-server/app/controllers/api/internal/player_data/stats_controller.rb
services/api-server/app/controllers/internal/player_data/match_results_controller.rb
services/api-server/app/services/player_stats/apply_match_result.rb
services/api-server/app/services/player_stats/serialize_stats.rb
services/api-server/app/models/player_stat.rb
services/api-server/app/models/player_match_result.rb
services/api-server/db/migrate/20260608000800_create_player_stats.rb
services/api-server/db/migrate/20260608000900_create_player_match_results.rb
services/api-server/db/migrate/20260608001000_add_account_id_to_users.rb
services/api-server/db/schema.rb
```

HTTP contract source:

```text
shared/contracts/http/openapi.yaml
```

Persisted local data output:

```text
services/player-data/data/player-data.sqlite3
```

## Validation commands

Player-data schema validation:

```bash
data-sync -validate -player_data
```

Packet contract verification when player-data packet payloads change:

```bash
data-sync -validate -packets
data-sync -check -packets -go -gds
```

Primary data-sync tests:

```text
tools/data_sync/tests/test_player_data_toml.py
tools/data_sync/tests/test_cli.py
tools/data_sync/tests/test_validate.py
```

Primary player-data runtime and store tests:

```text
services/player-data/playerdata/runtime_test.go
services/player-data/playerdata/store_router_test.go
services/player-data/playerdata/dispatcher_test.go
services/player-data/playerdata/mode_policy_test.go
services/player-data/playerdata/memory_store_test.go
services/player-data/playerdata/guest_memory_store_test.go
services/player-data/playerdata/noop_store_test.go
services/player-data/playerdata/rails_store_test.go
services/player-data/playerdata/embeddedsqlite/sqlite_store_test.go
```

Related API-server tests:

```text
services/api-server/test/controllers/api/internal/player_data/stats_controller_test.rb
services/api-server/test/controllers/internal/player_data/match_results_controller_test.rb
services/api-server/test/services/player_stats/apply_match_result_test.rb
services/api-server/test/models/player_stat_test.rb
services/api-server/test/models/player_match_result_test.rb
services/api-server/test/contracts/openapi_contract_test.rb
```

Related game-server integration tests:

```text
services/game-server/internal/playerdata/types_test.go
services/game-server/internal/playerdata/resolve_test.go
services/game-server/internal/playerdata/summary_test.go
services/game-server/internal/matchreporting/mapper_test.go
services/game-server/internal/matchreporting/runtime_reporter_test.go
```

## Related docs

* [Data](./!README.md)
* [Player Data service](../services/player-data/!README.md)
* [Player Data Runtime And Store Routing](../services/player-data/runtime-and-store-routing.md)
* [Profile Stats Flow](../services/player-data/profile-stats-flow.md)
* [Match Result Sinks](../services/player-data/match-result-sinks.md)
* [Local Profiles HTTP API](../services/player-data/local-profiles-http-api.md)
* [API Server Player Stats And Match Results](../services/api-server/player-stats-and-match-results.md)
* [Game Server Match Result Reporting](../services/game-server/integrations/match-result-reporting.md)
* [Game Server Player Data HTTP Hosting](../services/game-server/integrations/player-data-http-hosting.md)
* [HTTP contract enforcement](../protocol/http-contract-enforcement.md)
* [Protocol](../protocol/!README.md)
* [Data Sync](../../tools/data_sync/!README.md)

## Notes

The player-data logical schema is currently validate-only in data-sync. There is no current generator that produces Go structs, Rails migrations, SQLite schema, or OpenAPI payloads directly from `shared/player_data/*.toml`.

The generated player-data packet file is still important to this contract, but it is generated from `shared/packets/player_data.toml`.

Physical schemas may differ between Rails/Postgres and embedded SQLite as long as each backing route satisfies the shared logical player-data contract exposed by the player-data runtime.
