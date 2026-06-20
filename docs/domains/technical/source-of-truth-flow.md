## Source Of Truth Flow

Parent index: [Technical](./!README.md)

## Purpose

This document describes the cross-system source-of-truth flow for Space Rocks.

It explains how authoritative editable data, contracts, schemas, generated outputs, and runtime consumers stay aligned across the client, game server, player-data runtime, API server, and data tooling.

## Overview

Space Rocks uses source-of-truth files to prevent the client, game server, player-data runtime, and API server from independently defining the same data shapes.

The core flow is:

```text
owning source file
-> validation, generation, export, migration, or contract-test workflow
-> generated, synchronized, exported, or implemented consumer output
-> runtime service consumption
-> service tests and drift checks
```

Not every source of truth is handled by `tools/data_sync`.

`tools/data_sync` currently owns generated-output workflows for constants, realtime packets, and drop tables. It validates the player-data logical schema, but does not generate player-data runtime structs, Rails migrations, embedded SQLite schema, or HTTP contracts from that schema.

Other source-of-truth flows use adjacent workflows:

```text
OpenAPI HTTP contracts -> Rails/client/Go HTTP implementations and contract tests
Godot scene collision nodes -> exported shared collision JSON -> server collision loading
Rails migrations -> Rails/Postgres physical schema
embedded SQLite setup code -> local profile physical schema
asteroid variant TOML -> mirrored server and client asteroid catalogs
Godot scenes -> client runtime presentation and scene tree behavior
```

The source-of-truth domain flow is not one service and not one pipeline. It is the technical integration pattern that keeps shared contracts from drifting across services.

## Participating systems

### Data source files

The main editable source files live in `shared/`, `client/scenes/`, and service-owned schema locations.

Current source areas include:

```text
shared/constants/**/*.toml
shared/packets/*.toml
shared/drop_tables/*.toml
shared/player_data/*.toml
shared/contracts/http/openapi.yaml
shared/asteroids/variants.toml
client/scenes/**/*.tscn
services/api-server/db/migrate/*.rb
services/player-data/playerdata/embeddedsqlite/sqlite_store.go
```

These files own source material. Generated outputs and runtime consumers must not redefine the same contract independently.

### Data-sync tooling

`tools/data_sync` is the active generation and drift-check pipeline for:

```text
constants
packets
drop_tables
```

It is also the current validation path for:

```text
player_data
```

The data-sync configuration declares the current source roots and output targets. The tool can validate, diff, push, check, and pull where supported. Packet pull is intentionally unsupported, and player-data is validation-only.

### Client

The Godot client consumes generated constants, generated packet helpers, exported scenes, generated or synchronized asteroid variant data, and HTTP contract-shaped responses.

The client owns presentation and request construction. It does not own server simulation facts, packet schema authority, HTTP contract authority, persistent player-data storage, or generated data contracts.

### Game server

The game server consumes generated constants, generated packet structs, generated runtime packet projection types, generated devtools packet types, generated drop-table data, collision shape exports, asteroid variant catalogs, and player-data contracts.

The game server owns authoritative realtime simulation and match-result production. It does not own HTTP API schema, Rails physical storage, client presentation scenes, or editable shared data sources.

### Player-data runtime

The player-data runtime consumes generated player-data runtime packet contracts and implements the logical player-data model for guest, local profile, and authenticated-account paths.

It owns identity-based store routing and player-data persistence behavior. It does not let clients choose backing stores directly, and it does not treat Rails tables or SQLite tables as cross-service logical schema authority.

### API server

The Rails API server consumes and implements the shared HTTP contract.

It owns Rails route/controller behavior, Rails auth/account persistence, internal service endpoints, and Rails/Postgres physical schema through migrations. It does not own WebSocket packet schemas, game simulation state, local SQLite storage, or data-sync generated output.

## Authority boundaries

The main source-of-truth rule is:

```text
Edit the owning source.
Regenerate, export, migrate, or validate through the owning workflow.
Do not hand-edit generated output as authority.
Do not let one source-of-truth area silently own another area.
```

Current authority boundaries:

| Area                            | Authority                                                                                                                                       |
| ------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------- |
| Shared constants                | `shared/constants/**/*.toml` owns editable constants. Generated Go and GDScript files consume them.                                             |
| Realtime packet schemas         | `shared/packets/*.toml` owns WebSocket packet types, packet structs, fields, JSON names, and selected generated outputs.                        |
| Player-data runtime packets     | `shared/packets/player_data.toml` owns player-data runtime transport shapes.                                                                    |
| Drop tables                     | `shared/drop_tables/*.toml` owns drop-table configuration. Generated Go tables are runtime input, not source material.                          |
| Player-data logical schema      | `shared/player_data/*.toml` owns logical stats and match-result shape. Physical stores must satisfy it.                                         |
| HTTP contracts                  | `shared/contracts/http/openapi.yaml` owns HTTP request/response shapes. Rails controllers and Go/client HTTP callers implement it manually.     |
| Asteroid variants               | `shared/asteroids/variants.toml` owns variant metadata. Current server and client catalogs mirror it.                                           |
| Collision geometry              | Godot scene collision nodes are the editable geometry source. `shared/collisions/collision_shapes.json` is the exported server import artifact. |
| Rails physical schema           | Rails migrations own Rails/Postgres physical schema.                                                                                            |
| Embedded SQLite physical schema | Embedded SQLite setup code owns local profile physical schema. The runtime SQLite database file is not source material.                         |
| Godot scene structure           | Godot scenes own client scene/node structure and presentation composition.                                                                      |

## Durable roles

Durable source material lives in committed source files, not in runtime output.

Durable source roles include:

```text
shared TOML files
OpenAPI contract file
Rails migrations
embedded SQLite schema setup code
Godot scene files
Godot-exported shared collision artifact
```

Persistent runtime stores are durable data stores, but they are not automatically source-of-truth files.

For example:

```text
Rails/Postgres stores account-backed data.
Embedded SQLite stores local-profile data.
Guest data is transient memory.
```

The logical schema that those stores must satisfy is separate from their physical storage layout.

## Runtime roles

Runtime services consume source-of-truth outputs after validation, generation, export, or manual implementation.

The game server consumes generated constants, packet structs, drop tables, asteroid variant catalogs, and collision shape data during authoritative simulation.

The client consumes generated constants, packet helper builders, scene structure, and presentation catalogs during input, UI, networking, and world rendering.

The player-data runtime consumes generated player-data packet contracts and implements logical stats/profile behavior across guest, local profile, and authenticated account routes.

The Rails API server implements HTTP request/response shapes defined by OpenAPI and persists authenticated account data through Rails-owned physical schema.

## Presentation roles

Presentation consumers are downstream from source-of-truth ownership.

The Godot client may present values, textures, HUD labels, menu states, local profile status, entity visuals, and network state using generated or synchronized data. It does not become the authority for those contracts merely because the value appears in UI.

Examples:

```text
generated GDScript constants drive client presentation values
generated packet helpers build client outbound requests
asteroid variant texture paths drive asteroid presentation
scene collision nodes can export collision shape data for server use
```

The client can own presentation structure, but not server simulation truth or durable player-data authority.

## Flow summary

### Data-sync managed flow

For constants, packets, and drop tables:

```text
1. Edit the owning source under shared/.
2. Run data-sync validation.
3. Review generated diffs when needed.
4. Push generated output.
5. Run data-sync check to detect drift.
6. Run affected service tests.
7. Commit source and generated output together.
```

Generated outputs are consumed by the client, game server, devtools, and player-data runtime depending on the source area.

### Validation-only player-data schema flow

For logical player-data schema:

```text
1. Edit shared/player_data/*.toml.
2. Validate the logical schema through data-sync.
3. Update player-data runtime structs, stores, Rails persistence, or client/profile consumers as needed.
4. Run player-data, game-server, and Rails tests for affected behavior.
```

The shared player-data schema defines the logical contract. It does not currently generate physical schemas or runtime code.

### HTTP contract flow

For HTTP request/response contracts:

```text
1. Edit shared/contracts/http/openapi.yaml.
2. Update Rails controllers, Go HTTP handlers/adapters, and client API wrappers as needed.
3. Run OpenAPI contract tests and affected service tests.
```

OpenAPI owns HTTP shape. Rails migrations own database shape. Neither one automatically updates the other.

### Collision export flow

For collision geometry:

```text
1. Edit Godot scene collision nodes.
2. Export collision shapes into shared/collisions/collision_shapes.json.
3. Run server collision loading and physics tests.
```

The exported JSON is the server import artifact. The editable geometry source remains the Godot scene collision data.

### Asteroid variant mirror flow

For asteroid variants:

```text
1. Edit shared/asteroids/variants.toml.
2. Keep mirrored server and client catalogs aligned.
3. Run server asteroid/spawn tests and client asteroid variant tests.
```

Asteroid variant TOML is not currently listed as a data-sync domain. Until an automated generation path owns it, mirrored runtime catalogs must remain aligned through implementation review and tests.

### Physical schema flow

For Rails/Postgres:

```text
Rails migrations -> schema.rb -> Rails database -> Rails tests
```

For embedded SQLite local profile storage:

```text
SQLite setup code -> runtime SQLite database -> player-data tests
```

Physical schema updates must still satisfy the relevant logical and HTTP contracts when those contracts are involved.

## Inputs and outputs

| Flow input                 | Output                                                  | Primary consumers                                                                                                      |
| -------------------------- | ------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------- |
| Shared constants TOML      | Generated Go and GDScript constants                     | Game server, client                                                                                                    |
| Shared packet TOML         | Generated Go packet structs and GDScript packet helpers | Game server, client, devtools, player-data runtime                                                                     |
| Shared drop-table TOML     | Generated Go drop table data                            | Game server drop evaluation                                                                                            |
| Shared player-data TOML    | Validated logical schema                                | Player-data runtime, game-server match reporting, Rails-backed account persistence, embedded local-profile persistence |
| OpenAPI YAML               | Manually implemented HTTP endpoints and tests           | Rails API server, client API wrappers, game-server/player-data integrations                                            |
| Godot collision nodes      | Shared collision JSON                                   | Game-server collision loading                                                                                          |
| Asteroid variant TOML      | Mirrored server and client variant catalogs             | Game-server spawning, client asteroid presentation                                                                     |
| Rails migrations           | Rails schema and Postgres tables                        | API server                                                                                                             |
| Embedded SQLite setup code | Local profile SQLite schema                             | Player-data local profile store                                                                                        |
| Godot scenes               | Runtime scene tree and presentation behavior            | Client                                                                                                                 |

## Integration points

Source-of-truth flow intersects the main technical systems at these boundaries:

```text
Client <-> Game Server
```

Realtime packets are defined by shared packet TOML and consumed by both services.

```text
Game Server <-> Player Data
```

Match results and stats requests use player-data runtime packet contracts and logical player-data schemas.

```text
Client <-> HTTP API surfaces
```

Client HTTP wrappers consume request and response shapes defined by OpenAPI.

```text
Game Server <-> API Server
```

Authenticated-account verification and Rails-backed player-data persistence use HTTP contracts and internal service authorization.

```text
Client scene data -> Game Server collision loading
```

Collision shape export moves client-authored geometry into a server-readable shared artifact.

```text
Shared source files -> generated runtime consumers
```

Generated constants, packets, and drop tables are committed runtime inputs, but not editable authority.

## Out of scope

This document does not own the detailed source-of-truth map, generated-output list, command reference, or pipeline implementation map. Those belong in data documentation.

This document does not define realtime packet lifecycle behavior. That belongs in protocol documentation.

This document does not define service implementation responsibility or code paths. Those belong in service documentation.

This document does not define conceptual gameplay rules or design invariants. Those belong in systems-design documentation.

This document does not define future source-of-truth automation plans. Those belong in planning documentation.

## Related docs

* [Technical](./!README.md)
* [Data](../../data/!README.md)
* [Source Of Truth Map](../../data/source-of-truth-map.md)
* [Data Sync And SSoT Pipeline](../../data/data-sync-and-ssot-pipeline.md)
* [Constants Pipeline](../../data/constants.md)
* [Packet Schema Pipeline](../../data/packet-schemas.md)
* [Drop Tables](../../data/drop-tables.md)
* [Player Data Schema](../../data/player-data-schema.md)
* [Collision Shape Data](../../data/collision-shape-data.md)
* [Asteroid Variants Data](../../data/asteroid-variants-data.md)
* [HTTP Contract Enforcement](../../protocol/http-contract-enforcement.md)
* [Client](../../services/client/!README.md)
* [Game Server](../../services/game-server/!README.md)
* [Player Data](../../services/player-data/!README.md)
* [API Server](../../services/api-server/!README.md)

## Notes

Current authority comes from current source files, generated outputs, service implementations, tests, and the current data documentation.

`SSoT` means source of truth. It does not mean every source of truth is generated by `tools/data_sync`.

The current automated generation boundary is intentionally narrower than the project-wide source-of-truth flow. Constants, packets, and drop tables are active data-sync generation domains. Player-data logical schema is validation-only. HTTP contracts, collision export, asteroid variants, Rails migrations, embedded SQLite schema, and Godot scenes have separate workflows.
