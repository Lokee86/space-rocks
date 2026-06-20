## Source Of Truth Map

Parent index: [Data](./!INDEX.md)

## Purpose

This document maps the current Space Rocks data source-of-truth boundaries.

It identifies which files own editable data contracts, which files are generated or synchronized from those contracts, which services consume the outputs, and which validation paths detect drift.

## Overview

Space Rocks uses multiple source-of-truth styles.

Some data is centralized under `shared/` and synchronized through `tools/data_sync`. Other contracts are authoritative files outside data-sync, such as OpenAPI, Rails migrations, Godot scenes, and embedded SQLite schema setup code.

The main ownership rule is:

```text
Edit the owning source.
Regenerate or update synchronized outputs through the owning workflow.
Do not hand-edit generated output as source material.
```

A generated output may be consumed by runtime code, but it does not own the contract. Runtime code may enforce or interpret a contract, but it should not redefine that contract independently.

## Definitions

```text
Source of truth
= the authoritative editable input for a data shape, schema, contract, or catalog

Generated output
= a file produced from source material and not used as the editable source

Synchronized output
= a consumer-facing file that mirrors a source contract, whether or not it is currently data-sync generated

Implemented contract
= runtime code that accepts, stores, validates, or consumes the source contract

Logical schema
= the meaning and shape of data independent of physical storage

Physical schema
= concrete database tables, indexes, migrations, or file layouts

Drift
= a mismatch between the source contract and generated, synchronized, or implemented consumers
```

## Source files

| Area                            | Source of truth                                                                        | Generated or synchronized output                                                                                                                                                                                                                                                        | Consumers                                                                                                             | Enforcement                                                                                  | Does not own                                                                                       |
| ------------------------------- | -------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------- |
| Gameplay and shared constants   | `shared/constants/**/*.toml` listed by `tools/data_sync/config.toml`                   | `services/game-server/internal/constants/*.go`, `client/scripts/generated/constants/constants.gd`                                                                                                                                                                                       | Game server runtime, Godot client runtime                                                                             | `data-sync -validate -constants`, `data-sync -check -constants -go -gds`                     | Packet schemas, drop tables, database schemas, scene hierarchy                                     |
| Realtime packet schemas         | `shared/packets/*.toml` listed by `tools/data_sync/config.toml`                        | `services/game-server/internal/game/packets.go`, `services/game-server/internal/game/runtime/packets_generated.go`, `services/game-server/internal/devtools/packets_generated.go`, `services/player-data/protocol/packets.go`, `client/scripts/generated/networking/packets/packets.gd` | Game server networking, game runtime packet projection, devtools, player-data runtime protocol, Godot networking      | `data-sync -validate -packets`, `data-sync -check -packets -go -gds`                         | Constants, HTTP contracts, Rails schema, local SQLite schema                                       |
| Drop tables                     | `shared/drop_tables/*.toml` listed by `tools/data_sync/config.toml`                    | `services/game-server/internal/game/drops/drop_tables.go`                                                                                                                                                                                                                               | Game server drop evaluation and pickup spawning handoff                                                               | `data-sync -validate -drop-tables`, `data-sync -check -drop-tables -go`, drop table Go tests | Pickup collection, pickup effects, packet schema, constants                                        |
| Player-data logical schema      | `shared/player_data/stats.toml`, `shared/player_data/match_result.toml`                | Current pipeline validates schema shape only; implemented Go structs and stores must satisfy the logical contract                                                                                                                                                                       | Player-data runtime, game-server match reporting, Rails-backed account persistence, embedded SQLite local persistence | `data-sync -validate -player_data`, player-data Go tests, Rails player-data tests            | HTTP request/response shapes, physical Rails schema, physical SQLite schema, live simulation state |
| Player-data runtime packets     | `shared/packets/player_data.toml` plus `shared/packets/outputs.toml`                   | `services/player-data/protocol/packets.go`                                                                                                                                                                                                                                              | Player-data dispatcher, runtime sink, game-server match reporting                                                     | `data-sync -validate -packets`, `data-sync -check -packets -go`                              | Player-data physical storage, HTTP profile/local-profile contracts                                 |
| HTTP API contracts              | `shared/contracts/http/openapi.yaml`                                                   | Rails controllers and tests implement the contract; no generated controllers or clients currently exist                                                                                                                                                                                 | Rails API server, client HTTP API wrappers, game-server auth and match-result integration, player-data HTTP handlers  | Rails OpenAPI contract tests and controller tests                                            | WebSocket packets, data-sync domains, Rails migrations, runtime OpenAPI middleware                 |
| Asteroid variants               | `shared/asteroids/variants.toml`                                                       | `services/game-server/internal/game/asteroids/variants.go`, `client/scripts/generated/asteroids/asteroid_variants.gd`                                                                                                                                                                   | Game-server asteroid spawning and variant lookup, client asteroid presentation                                        | Go asteroid variant tests, GUT asteroid variant tests                                        | Collision geometry, packet field names, asteroid physics, drop evaluation                          |
| Collision shape export          | Godot scene collision nodes referenced by `client/tools/export_collision_shapes.gd`    | `shared/collisions/collision_shapes.json`                                                                                                                                                                                                                                               | Game-server physics collision-shape loader and collision tests                                                        | Godot export command, game-server collision shape tests                                      | Gameplay rules, packet schemas, database schemas                                                   |
| Rails/Postgres physical schema  | `services/api-server/db/migrate/*.rb`                                                  | `services/api-server/db/schema.rb`, Rails database tables                                                                                                                                                                                                                               | API server auth, token, stats, match-result, and account persistence                                                  | `bundle exec rails db:migrate`, Rails tests                                                  | OpenAPI request/response contract, game simulation, local SQLite schema                            |
| Embedded SQLite physical schema | `services/player-data/playerdata/embeddedsqlite/sqlite_store.go` schema initialization | `services/player-data/data/player-data.sqlite3` runtime database file                                                                                                                                                                                                                   | Local profile store in standard no-tag development build                                                              | Player-data embedded SQLite tests, game-server/player-data no-tag tests                      | Rails/Postgres schema, authenticated-account persistence, OpenAPI schema                           |
| Godot scene/node structure      | `client/scenes/**/*.tscn`                                                              | Runtime scene tree and script consumers under `client/scripts/`                                                                                                                                                                                                                         | Godot client presentation, input, menus, world sync, collision export                                                 | GUT tests and Godot scene/runtime checks                                                     | Server simulation, packet schema, database schema                                                  |

## Generated outputs

The current generated or synchronized outputs are already captured in the source-files ownership map above and the code or source map below.

## Consumers

The current consumer relationships are captured in the source-files ownership map above. That map records which runtime, service, and tool boundaries consume each source area.

## Configuration

`tools/data_sync/config.toml` defines the active data-sync source roots and output targets.

Current source groups:

```text
[sot.constants]
shared/constants/server_constants.toml
shared/constants/server_entities.toml
shared/constants/weapons.toml
shared/constants/client/presentation.toml
shared/constants/client/shell.toml
shared/constants/client/lobby.toml
shared/constants/pickups.toml
shared/constants/weapon_pickups.toml

[sot.packets]
shared/packets/outputs.toml
shared/packets/gameplay.toml
shared/packets/debug.toml
shared/packets/lobby.toml
shared/packets/player_data.toml

[sot.drop_tables]
shared/drop_tables/basicasteroids.toml

[sot.player_data]
shared/player_data/stats.toml
shared/player_data/match_result.toml
```

Active generated output targets:

```text
constants -> Go and GDScript
packets -> Go and GDScript
drop_tables -> Go
player_data -> validation only
```

TypeScript output is disabled in the current default config.

## Constants ownership

Constants are edited in TOML under `shared/constants/`.

The constants pipeline discovers destination blocks in configured Go and GDScript outputs and replaces only the matching managed blocks.

Generated constants files include:

```text
services/game-server/internal/constants/constants.go
services/game-server/internal/constants/weapons.go
services/game-server/internal/constants/powerups.go
services/game-server/internal/constants/weapon_pickups.go
client/scripts/generated/constants/constants.gd
```

Constants sections own numeric and string values used by runtime and presentation code. They do not own packet shape, protocol lifecycle, drop-table behavior, database schema, or scene hierarchy.

When changing constants:

```text
edit shared/constants/**/*.toml
data-sync -validate -constants
data-sync -diff -constants -go -gds
data-sync -push -constants -go -gds
data-sync -check -constants -go -gds
```

## Packet schema ownership

Realtime packet schemas are edited under `shared/packets/`.

The packet schema owns:

```text
packet type constants
packet structs
field names
JSON names
packet builder definitions
language output selections
```

Generated packet outputs include:

```text
services/game-server/internal/game/runtime/packets_generated.go
services/game-server/internal/game/packets.go
services/game-server/internal/devtools/packets_generated.go
services/player-data/protocol/packets.go
client/scripts/generated/networking/packets/packets.gd
```

Packet pull is not supported. Packet schema changes should be made in `shared/packets/`, then pushed through data-sync.

When changing packets:

```text
edit shared/packets/*.toml
data-sync -validate -packets
data-sync -diff -packets -go -gds
data-sync -push -packets -go -gds
data-sync -check -packets -go -gds
```

Packet schemas do not own the server simulation, room rules, HTTP request/response contracts, or database schema.

## Drop-table ownership

Drop-table source data is edited under `shared/drop_tables/`.

The current active table source is:

```text
shared/drop_tables/basicasteroids.toml
```

Generated server output:

```text
services/game-server/internal/game/drops/drop_tables.go
```

The drop-table source owns:

```text
table id
source type
drop mode
max drops per source
max active pickups
entry pickup type
entry chance
entry source-size bounds
```

The game-server `drops` package owns deterministic evaluation of generated tables. Root game simulation code owns spawning pickups after a successful drop result. Pickup collection and pickup effects remain owned by pickup systems, not by the drop-table contract.

When changing drop tables:

```text
edit shared/drop_tables/*.toml
data-sync -validate -drop-tables
data-sync -diff -drop-tables -go
data-sync -push -drop-tables -go
data-sync -check -drop-tables -go
```

## Player-data schema ownership

Logical player-data schema lives under `shared/player_data/`.

Current logical schema files:

```text
shared/player_data/stats.toml
shared/player_data/match_result.toml
```

These files define the logical account-shaped player-data contract. They are not physical database schemas.

Current logical stats fields:

```text
total_score
high_score
ship_deaths
games_played
wins
```

Current match-result schema includes match summary and per-player match summary fields such as:

```text
match_id
mode
resolved_at
game_player_id
account_id
local_profile_id
score
ship_deaths
won
```

`data-sync -validate -player_data` validates the player-data TOML schema shape, field groups, supported types, required/default/optional markers, and parseability. It does not currently generate Go structs, Rails migrations, embedded SQLite migrations, or HTTP contracts from these files.

Physical stores must satisfy the logical player-data contract even when their concrete table layout differs.

Physical schema ownership is separate:

```text
Rails/Postgres physical schema -> services/api-server/db/migrate/*.rb
embedded SQLite physical schema -> services/player-data/playerdata/embeddedsqlite/sqlite_store.go
runtime SQLite file -> services/player-data/data/player-data.sqlite3
```

The runtime SQLite file is data, not source material.

## HTTP contract ownership

HTTP request and response shapes are owned by:

```text
shared/contracts/http/openapi.yaml
```

OpenAPI currently owns:

```text
auth request/response schemas
Discord OAuth HTTP request/response schemas
current-user response schema
player stats response schema
internal player-data stats request/response schema
player-data profile request/response schema
local profile request/response schemas
token verification request/response schema
match-result submission request/response schema
status codes and security declarations
```

OpenAPI is not generated by data-sync. Rails controllers and Go HTTP handlers implement the contract manually.

Current enforcement is test-time enforcement. Runtime OpenAPI middleware is not active.

Minimum verification:

```text
cd services/api-server && bundle exec rails test test/contracts/openapi_contract_test.rb
cd services/api-server && bundle exec rails test
```

When HTTP changes touch player-data or game-server integration, also run:

```text
cd services/player-data && go test ./...
cd services/game-server && go test -buildvcs=false ./...
```

OpenAPI does not own Rails migrations, Rails strong params, generated clients, WebSocket packet contracts, or player-data runtime packets.

## Asteroid variant ownership

Asteroid variant metadata is owned by:

```text
shared/asteroids/variants.toml
```

Current synchronized consumers:

```text
services/game-server/internal/game/asteroids/variants.go
client/scripts/generated/asteroids/asteroid_variants.gd
```

The variant source owns:

```text
stable variant id
zero-based runtime index
client texture path
collision shape key
stats profile key
drop table key
timed spawn weight
fragment spawn weight
debug spawn weight
```

Current variant behavior expects zero-based runtime indexes. Variant IDs such as `asteroid_1` are stable labels, not runtime index values.

The current data-sync config does not include an asteroid-variant domain. Treat the TOML file as the contract owner and keep the Go and GDScript catalogs synchronized with it until a generation path is added or confirmed.

Verification:

```text
cd services/game-server && go test -buildvcs=false ./internal/game/asteroids ./internal/game/spawning ./internal/devtools
cd client && godot --headless --path . -s addons/gut/gut_cmdln.gd -gtest=res://tests/unit/entities/test_asteroid_variants.gd
```

Asteroid variants do not own asteroid collision geometry, packet field names, health rules, physics rules, or drop evaluation.

## Collision shape ownership

Collision geometry starts from Godot scene collision nodes.

The export tool is:

```text
client/tools/export_collision_shapes.gd
```

It reads collision nodes from scenes such as:

```text
client/scenes/bullet.tscn
client/scenes/player.tscn
client/scenes/asteroid.tscn
client/scenes/pickups/powerup_pickup.tscn
client/scenes/pickups/weapon_pickup.tscn
```

It writes the shared collision artifact:

```text
shared/collisions/collision_shapes.json
```

The game server consumes that artifact through:

```text
services/game-server/internal/game/physics/collision_shapes.go
```

The shared JSON file is the import artifact for server collision loading. The editable geometry source is the Godot scene collision data exported into that file.

Export command:

```text
godot --headless --path client -s res://tools/export_collision_shapes.gd
```

Collision shape data does not own gameplay collision policy, damage rules, packet schema, scene presentation behavior, or database schema.

## Physical persistence ownership

### Rails/Postgres

Rails migrations own the Rails/Postgres physical schema:

```text
services/api-server/db/migrate/*.rb
```

Generated Rails schema output:

```text
services/api-server/db/schema.rb
```

Rails/Postgres currently owns physical tables for auth, OAuth, access tokens, account stats, and player match results.

Rails migrations do not own HTTP request/response shape. OpenAPI owns that contract.

### Embedded SQLite

The embedded local-profile schema is currently initialized in code:

```text
services/player-data/playerdata/embeddedsqlite/sqlite_store.go
```

Current tables:

```text
local_profiles
local_profile_default
local_player_stats
local_player_match_results
```

The standard no-tag development build includes the embedded SQLite store. The `noembeddedsqlite` build path omits embedded SQLite and local profile management returns unavailable behavior through the HTTP/runtime path.

The SQLite database file is runtime storage:

```text
services/player-data/data/player-data.sqlite3
```

It should not be treated as source material.

## Pipeline usage

Use this default source-of-truth update sequence:

```text
1. Identify the owning source file.
2. Edit only the owning source and the minimum required implementation consumers.
3. Regenerate or synchronize output through the owning workflow.
4. Run source validation.
5. Run drift checks for generated output.
6. Run service tests for every runtime consumer touched by the change.
7. Do not hand-edit generated files unless the file is only a synchronized mirror and no generator exists.
```

For data-sync managed domains, prefer:

```text
data-sync -validate <domain>
data-sync -diff <domain> <languages>
data-sync -push <domain> <languages>
data-sync -check <domain> <languages>
```

For non-data-sync sources, use the source-specific workflow:

```text
OpenAPI -> Rails contract tests and affected service tests
Collision shapes -> Godot export command and game-server collision tests
Rails schema -> Rails migrations and Rails tests
Embedded SQLite schema -> player-data Go tests and build-tag tests
Asteroid variants -> server asteroid tests and client GUT variant tests
Godot scenes -> GUT tests and scene-specific runtime checks
```

## Validation commands

Core data-sync validation:

```text
data-sync -validate
data-sync -validate -constants
data-sync -validate -packets
data-sync -validate -drop-tables
data-sync -validate -player_data
```

Generated output drift checks:

```text
data-sync -check -constants -go -gds
data-sync -check -packets -go -gds
data-sync -check -drop-tables -go
```

HTTP contract validation:

```text
cd services/api-server && bundle exec rails test test/contracts/openapi_contract_test.rb
cd services/api-server && bundle exec rails test
```

Player-data validation:

```text
cd services/player-data && go test ./...
cd services/player-data && go test -tags noembeddedsqlite ./...
```

Game-server validation:

```text
cd services/game-server && go test -buildvcs=false ./...
cd services/game-server && go test -tags noembeddedsqlite -buildvcs=false ./cmd/game-server
```

Collision export validation:

```text
godot --headless --path client -s res://tools/export_collision_shapes.gd
cd services/game-server && go test -buildvcs=false ./internal/game/physics
```

Client validation for generated consumers:

```text
cd client && godot --headless --path . -s addons/gut/gut_cmdln.gd -gdir=res://tests/unit -ginclude_subdirs -gexit
```

## Failure modes

Duplicate source ownership causes drift. A value should not be independently owned by TOML, Go, and GDScript at the same time.

Stale generated output causes runtime mismatch. If a shared TOML file changes but generated Go or GDScript does not, server/client behavior may diverge.

Packet schema edits in generated files are invalid. Packet pull is not supported; edit `shared/packets/*.toml` instead.

OpenAPI changes do not update Rails controllers automatically. Contract changes must be paired with implementation and tests.

Rails migrations do not update OpenAPI automatically. Database schema changes that affect HTTP payloads must update `shared/contracts/http/openapi.yaml`.

Collision scene edits do not update server collision data until the export command rewrites `shared/collisions/collision_shapes.json`.

Asteroid variant TOML is not currently listed in `tools/data_sync/config.toml`. Keep mirrored Go and GDScript catalogs synchronized with `shared/asteroids/variants.toml` until an automated generator is added or confirmed.

Embedded SQLite runtime data is not source material. Do not treat `services/player-data/data/player-data.sqlite3` as the schema owner.

## Code or source map

Primary data-sync implementation:

```text
tools/data_sync/main.py
tools/data_sync/config.toml
tools/data_sync/data_sync/cli.py
tools/data_sync/data_sync/config.py
tools/data_sync/data_sync/validate.py
tools/data_sync/data_sync/constants_sync.py
tools/data_sync/data_sync/packets_sync.py
tools/data_sync/data_sync/drop_tables_sync.py
tools/data_sync/data_sync/player_data_toml.py
```

Data-sync generators:

```text
tools/data_sync/data_sync/generators/go_constants.py
tools/data_sync/data_sync/generators/gds_constants.py
tools/data_sync/data_sync/generators/ts_constants.py
tools/data_sync/data_sync/generators/go_packets.py
tools/data_sync/data_sync/generators/gds_packets.py
tools/data_sync/data_sync/generators/ts_packets.py
tools/data_sync/data_sync/generators/rich_go_packets.py
tools/data_sync/data_sync/generators/rich_gds_packets.py
tools/data_sync/data_sync/generators/go_drop_tables.py
```

Data-sync parser and validation tests:

```text
tools/data_sync/tests/test_config.py
tools/data_sync/tests/test_validate.py
tools/data_sync/tests/test_constants_sync.py
tools/data_sync/tests/test_packets_sync.py
tools/data_sync/tests/test_drop_tables_sync.py
tools/data_sync/tests/test_player_data_toml.py
```

Source roots:

```text
shared/constants/
shared/packets/
shared/drop_tables/
shared/player_data/
shared/contracts/http/openapi.yaml
shared/asteroids/variants.toml
shared/collisions/collision_shapes.json
client/scenes/
services/api-server/db/migrate/
services/player-data/playerdata/embeddedsqlite/sqlite_store.go
```

Important non-ownership boundaries:

```text
docs/data/ documents ownership and pipeline behavior.
docs/protocol/ documents communication behavior.
docs/services/ documents runtime implementation responsibility.
docs/systems-design/ documents conceptual invariants.
docs/limits/ documents active temporary constraints.
docs/planning/ documents future source-of-truth changes.
```

## Related docs

* [Data](./!INDEX.md)
* [HTTP Contract Enforcement](../protocol/http-contract-enforcement.md)
* [API Server](../services/api-server/!INDEX.md)
* [Client](../services/client/!INDEX.md)
* [Game Server](../services/game-server/!INDEX.md)
* [Player Data](../services/player-data/!INDEX.md)
* [Data Sync](../../tools/data_sync/!INDEX.md)

## Notes

Legacy source-of-truth documentation was used as migration source material only. Current authority comes from the source files, generated outputs, service implementations, and current documentation listed in this document.

Some source areas are not fully automated yet. In particular, asteroid variants are represented by a TOML source plus mirrored runtime catalogs, but they are not currently listed as a data-sync domain in `tools/data_sync/config.toml`.

Player-data logical schema validation exists, but generation from `shared/player_data/*.toml` is not currently implemented. Current runtime structs, Rails persistence, and embedded SQLite storage must be kept aligned through tests and implementation review.
