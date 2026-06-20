## Packet Schemas

Parent index: [Data](./!INDEX.md)

## Purpose

This document describes the current packet schema source-of-truth pipeline for Space Rocks.

It covers the editable packet schema files under `shared/packets/`, the `data-sync` packet generation workflow, the generated Go and GDScript outputs, validation rules, runtime consumers, and boundaries between packet schemas, HTTP contracts, logical player-data schemas, and runtime protocol behavior.

## Overview

Packet schemas are authored as TOML source files under:

```text
shared/packets/
```

Those TOML files are the canonical source for generated packet constants, packet structs, packet field names, and client packet builders used by the realtime game, lobby/auth flow, devtools, and player-data runtime protocol.

The packet schema pipeline is owned by `tools/data_sync/`. Packet changes are edited in TOML, validated, diffed, pushed into generated language outputs, and then checked for drift.

Current active packet generation targets are:

```text
Go       -> game-server packet structs/constants, devtools packet structs/constants, player-data protocol packets
GDScript -> client packet constants, field constants, and outbound packet builders
TypeScript -> configured but disabled for packets
```

Packet generated files are full generated outputs, not data-sync managed blocks inside hand-written files. They should not be hand-edited.

## Source files

The active packet schema source files are configured in:

```text
tools/data_sync/config.toml
```

Current packet source files:

```text
shared/packets/outputs.toml
shared/packets/gameplay.toml
shared/packets/debug.toml
shared/packets/lobby.toml
shared/packets/player_data.toml
```

### `shared/packets/outputs.toml`

`outputs.toml` owns packet generation routing.

Each `[[outputs]]` entry defines a generated file target, language, package/base class, selected packet type constants, selected structs, selected builders, and imports where needed.

Current output ids:

```text
server_entities_packets
server_game_packets
player_data_packets
client_packets
server_devtools_packets
```

`packet_type_ids` restricts which packet constants are rendered into an output and preserves the listed order. `structs` selects which packet structs are rendered for Go outputs. `builders` selects which GDScript builder functions are rendered for the client output.

### `shared/packets/gameplay.toml`

`gameplay.toml` owns the core realtime gameplay packet and state shapes.

Current schema areas include:

```text
ClientPacket
ClientConfig
InputState
ShipState
PlayerSessionState
AsteroidState
BulletState
PickupState
EventState
PlayerPauseState
StatePacket
```

It also owns packet type values and client builders for gameplay input, respawn, pause, targeting, telemetry, client viewport configuration, state packets, and presentation events.

### `shared/packets/debug.toml`

`debug.toml` owns devtools packet shapes.

Current schema areas include:

```text
DebugCommand
DebugStatus
DebugShapePoint
DebugShapeDefinition
DebugShapeCatalogPacket
DebugStatusPacket
```

It owns devtools packet types for toggles, status, shape catalog output, player mutation commands, entity spawning, pickup spawning, continuous bullet stream commands, respawn commands, score/lives commands, and entity clearing.

Devtools packet schema is shared data, but devtools command behavior is not owned by the packet schema. Runtime behavior is owned by the server devtools implementation and the game-owned devtools export seams.

### `shared/packets/lobby.toml`

`lobby.toml` owns room, lobby, auth, and room match-result packet shapes.

Current schema areas include:

```text
CreateRoomRequest
JoinRoomRequest
LeaveRoomRequest
SetReadyRequest
StartGameRequest
StartSinglePlayerRequest
ReturnToLobbyRequest
AuthenticateRequest
AuthenticateResult
RoomMemberState
RoomPlayerMatchSummary
RoomMatchResultSummary
RoomSnapshot
RoomStateChanged
RoomError
```

It owns packet types and client builders for create/join/leave, ready state, game start, single-player start, return-to-lobby, and authenticate requests.

### `shared/packets/player_data.toml`

`player_data.toml` owns the player-data runtime packet protocol.

Current schema areas include:

```text
PlayerDataIdentity
PlayerDataStats
PlayerDataRequestContext
PlayerDataRecordMatchResult
PlayerDataRecordMatchResultResult
PlayerDataLoadStats
PlayerDataLoadStatsResult
```

This file does not own HTTP request/response shapes and does not own the broader logical player-data schema. HTTP shapes are owned by `shared/contracts/http/openapi.yaml`. Logical player-data schema files are owned separately under `shared/player_data/`.

## Schema model

Packet TOML uses four top-level list types:

```text
[[outputs]]
[[structs]]
[[packet_types]]
[[builders]]
```

### Outputs

An output entry describes one generated file.

Supported output fields include:

```text
id
language
path
package
imports
packet_types
packet_type_ids
structs
base
builders
```

Go outputs require a package name. GDScript outputs may set `base`, currently `RefCounted` for the client packet helper.

Output paths must be relative project paths. Absolute paths and parent-directory traversal are invalid.

### Structs

A struct entry defines a packet or state shape.

Struct ids use exported Go-style names such as:

```text
StatePacket
RoomSnapshot
DebugCommand
PlayerDataLoadStatsResult
```

Fields define the schema-facing field name, JSON key, type, and optional Go-specific overrides.

Supported field patterns include:

```text
bool
int
float
string
map
array
map<string,ShipState>
array<EventState>
custom struct references
```

Field entries may use explicit `key_type`, `value_type`, and `item_type`, or rich type strings such as `map<string,ShipState>` and `array<EventState>` where supported.

Go-specific override fields include:

```text
go_name
go_type
go_item_type
go_value_type
```

These are used when the generated Go field name or type must differ from the schema-level name or type. For example, generated game packet structs can reference runtime package structs such as `runtime.ShipState`.

### Packet types

Packet type entries define stable packet type ids and wire values.

Example shape:

```toml
[[packet_types]]
id = "state"
value = "state"
```

The `id` is used by generators to create language-specific constants. The `value` is the JSON packet type string sent over the wire or through the player-data runtime packet codec.

Packet type ids must be unique. Packet type values must also be unique.

### Builders

Builder entries define generated GDScript outbound packet helper functions.

Example shape:

```toml
[[builders]]
id = "input_packet"
args = ["forward", "back", "right", "left", "primary_fire", "secondary_fire"]

[builders.body]
type = "input"
```

Builder ids must be snake_case and end in `_packet`.

Builder body keys must reference known packet JSON fields. Builder argument references use `$arg_name`. Packet type values used in builder bodies must correspond to known packet type values.

Builders currently generate GDScript dictionary construction helpers only.

## Generated outputs

Current packet outputs are:

| Output id                 | Source routing | Generated file                                                    | Runtime owner                                                                          |
| ------------------------- | -------------- | ----------------------------------------------------------------- | -------------------------------------------------------------------------------------- |
| `server_entities_packets` | `outputs.toml` | `services/game-server/internal/game/runtime/packets_generated.go` | Game-server runtime state structs                                                      |
| `server_game_packets`     | `outputs.toml` | `services/game-server/internal/game/packets.go`                   | Game-server gameplay, lobby, auth, telemetry, room, and state packet structs/constants |
| `server_devtools_packets` | `outputs.toml` | `services/game-server/internal/devtools/packets_generated.go`     | Game-server devtools packet structs/constants                                          |
| `player_data_packets`     | `outputs.toml` | `services/player-data/protocol/packets.go`                        | Player-data runtime packet structs/constants                                           |
| `client_packets`          | `outputs.toml` | `client/scripts/generated/networking/packets/packets.gd`          | Client packet constants, field constants, and outbound packet builders                 |

The generated files begin with a generated-code warning. They are outputs, not edit sources.

## Consumers

### Game server

The game server consumes generated packet code for inbound packet routing, outbound packet writing, state packet projection, telemetry, room snapshots, lobby flow, auth result packets, devtools packets, and player-data match-result reporting.

Primary generated Go files:

```text
services/game-server/internal/game/packets.go
services/game-server/internal/game/runtime/packets_generated.go
services/game-server/internal/devtools/packets_generated.go
```

Primary packet runtime paths include:

```text
services/game-server/internal/networking/
services/game-server/internal/networking/inbound/
services/game-server/internal/networking/outbound/
services/game-server/internal/protocol/packetcodec/
services/game-server/internal/game/state_packet.go
services/game-server/internal/game/input.go
services/game-server/internal/game/events.go
services/game-server/internal/devtools/
services/game-server/internal/matchreporting/
```

The packet schema defines shape and generated constants. It does not own the semantic result of a packet after routing. Gameplay, room, auth, devtools, telemetry, and player-data behavior are owned by their service implementation paths.

### Client

The client consumes generated GDScript packet constants, field constants, and builder functions from:

```text
client/scripts/generated/networking/packets/packets.gd
```

Primary client packet runtime paths include:

```text
client/scripts/networking/
client/scripts/networking/packets/
client/scripts/networking/outbound/
client/scripts/networking/inbound/
client/scripts/gameplay/state/
client/scripts/gameplay/runtime/
client/scripts/world/
client/scripts/lobby/
client/scripts/devtools/
```

The generated GDScript file does not generate full typed packet structs. It provides constants and outbound dictionary builders. Packet reading and interpretation remain in client runtime readers and feature-specific flows.

### Player data

The player-data service consumes generated packet structs and constants from:

```text
services/player-data/protocol/packets.go
```

Primary runtime paths include:

```text
services/player-data/playerdata/dispatcher.go
services/player-data/playerdata/runtime.go
services/player-data/playerdata/store_router.go
services/player-data/playerdata/store.go
services/player-data/playerdata/memory_store.go
services/player-data/playerdata/guest_memory_store.go
services/player-data/playerdata/rails_store.go
services/player-data/playerdata/embeddedsqlite/
services/game-server/internal/matchreporting/
```

The player-data packet schema owns runtime command/result packet shapes such as record-match-result and load-stats packets. It does not own local profile HTTP endpoints, Rails HTTP endpoints, Rails migrations, embedded SQLite physical schema, or the logical player-data schema under `shared/player_data/`.

## Pipeline usage

Packet schema changes follow this workflow:

```text
1. Edit the relevant TOML file under shared/packets/.
2. Run packet validation.
3. Run packet diff for Go and GDScript.
4. Review generated output changes.
5. Push generated Go and GDScript outputs.
6. Check generated outputs for drift.
7. Run affected service tests.
```

Commands:

```bash
data-sync -validate -packets
data-sync -diff -packets -go -gds
data-sync -push -packets -go -gds
data-sync -check -packets -go -gds
```

`-go` currently covers all configured Go packet outputs, including game-server, devtools, runtime entity/state packet structs, and player-data protocol packets.

`-gds` currently covers the generated client packet helper.

`-ts` is configured as disabled for packets. Packet `-push`, `-diff`, or `-check` with `-ts` fails while `[packets.ts]` remains disabled.

Packet pull is intentionally unsupported:

```text
data-sync -pull -packets ...
```

Packet schema changes must be made in `shared/packets/`, not reverse-generated from language outputs.

## Validation commands

Minimum packet validation:

```bash
data-sync -validate -packets
data-sync -check -packets -go -gds
```

Review generated changes before writing them:

```bash
data-sync -diff -packets -go -gds
```

Run data-sync tests when changing the packet pipeline implementation:

```bash
cd tools/data_sync && pytest
```

Run service tests when packet changes affect runtime behavior:

```bash
cd services/game-server && go test -buildvcs=false ./...
cd services/player-data && go test ./...
```

Run client tests when packet constants, builders, or client packet readers change:

```bash
godot --headless --path client -s res://addons/gut/gut_cmdln.gd -gdir=res://tests/unit -ginclude_subdirs -gexit
```

## Failure modes

Common packet schema and pipeline failures include:

```text
duplicate output ids
duplicate output paths
duplicate struct ids
duplicate packet type ids
duplicate packet type values
duplicate builder ids
unknown struct referenced by an output
unknown packet type id referenced by an output
unknown builder referenced by a GDScript output
unsupported output language
disabled TypeScript packet output
absolute or parent-traversing generated output path
invalid struct name
invalid snake_case field name
invalid JSON field name
map fields missing key_type or value_type
array fields missing item_type
unknown custom struct/type reference
builder body references an unknown packet field
builder body references an unknown builder argument
builder body references an unknown packet type value
stale generated output after TOML edits
hand-edited generated packet files
attempted packet pull
```

Generated packet files are fully rewritten by the packet generator. Missing data-sync block markers are not relevant for packet outputs because packets do not use managed block replacement.

## Ownership boundaries

Packet schemas own:

```text
packet type wire strings
packet struct shape
JSON field names
selected generated packet constants
selected generated Go structs
selected generated GDScript constants
selected generated GDScript builders
packet output file routing
```

Packet schemas do not own:

```text
WebSocket upgrade behavior
JSON codec implementation
room lifecycle rules
game simulation authority
input semantics
target selection authority
pause behavior
telemetry meaning
devtools command effects
match result calculation
player-data store routing
HTTP request/response contracts
Rails/Postgres physical schema
embedded SQLite physical schema
logical player-data schema
drop-table data
constant data
collision-shape data
```

HTTP request and response contracts are separate and owned by:

```text
shared/contracts/http/openapi.yaml
```

Logical player-data schemas are separate and owned by:

```text
shared/player_data/
```

Generated packet output should stay aligned with runtime protocol docs and service docs, but protocol behavior is not defined by the data pipeline alone.

## Code or source map

Packet source files:

```text
shared/packets/outputs.toml
shared/packets/gameplay.toml
shared/packets/debug.toml
shared/packets/lobby.toml
shared/packets/player_data.toml
```

Pipeline configuration:

```text
tools/data_sync/config.toml
tools/data_sync/!INDEX.md
tools/data_sync/main.py
tools/data_sync/data_sync/cli.py
```

Packet schema loading and validation:

```text
tools/data_sync/data_sync/packet_toml.py
tools/data_sync/data_sync/model/packets.py
tools/data_sync/data_sync/validate.py
tools/data_sync/data_sync/packet_rendering.py
```

Packet generation:

```text
tools/data_sync/data_sync/packets_sync.py
tools/data_sync/data_sync/generators/rich_go_packets.py
tools/data_sync/data_sync/generators/rich_gds_packets.py
tools/data_sync/data_sync/generators/go_packets.py
tools/data_sync/data_sync/generators/gds_packets.py
tools/data_sync/data_sync/generators/ts_packets.py
```

Generated outputs:

```text
client/scripts/generated/networking/packets/packets.gd
services/game-server/internal/game/packets.go
services/game-server/internal/game/runtime/packets_generated.go
services/game-server/internal/devtools/packets_generated.go
services/player-data/protocol/packets.go
```

Pipeline tests:

```text
tools/data_sync/tests/test_packet_toml_loader.py
tools/data_sync/tests/test_packet_toml_validation.py
tools/data_sync/tests/test_packet_generators.py
tools/data_sync/tests/test_packet_rendering.py
tools/data_sync/tests/test_packets_sync.py
tools/data_sync/tests/test_final_flows.py
tools/data_sync/tests/test_constants_pull.py
```

Important non-ownership boundaries:

```text
services/game-server/internal/protocol/packetcodec/ owns JSON encode/decode helpers.
services/game-server/internal/networking/ owns WebSocket session routing.
client/scripts/networking/ owns client packet send/receive flow.
services/player-data/playerdata/ owns runtime handling of player-data packets.
shared/contracts/http/openapi.yaml owns HTTP contracts.
shared/player_data/ owns logical player-data schema.
```

## Related docs

* [Data](./!INDEX.md)
* [Data Sync and SSoT Pipeline](data-sync-and-ssot-pipeline.md)
* [Source of Truth Map](source-of-truth-map.md)
* [HTTP Contract Enforcement](../protocol/http-contract-enforcement.md)
* [Game Server](../services/game-server/!INDEX.md)
* [Client](../services/client/!INDEX.md)
* [Player Data](../services/player-data/!INDEX.md)

## Notes

`tools/data_sync/!INDEX.md` describes the packet workflow, but `tools/data_sync/config.toml` is the immediate source for the active configured packet source paths and output targets.

The GDScript packet output currently renders field constants from the loaded schema and builders selected by `client_packets`. This makes some field constants available even when the client does not own the corresponding runtime packet family.

Packet data documentation should stay focused on schema source files, generated outputs, and pipeline behavior. Packet lifecycle, routing order, authority, and service-specific runtime consequences belong in protocol and service documentation.
