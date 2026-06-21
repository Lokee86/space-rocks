## Asteroid Variant Contract

Parent index: [Protocol](./!INDEX.md)

## Purpose

This document describes the asteroid variant protocol contract between the game server and the client.

It explains how the server assigns asteroid variant indexes, how those indexes cross the realtime gameplay state packet boundary, how the client consumes them for presentation, and which source data defines their meaning.

## Overview

Asteroid variants cross the protocol boundary as an integer field on asteroid state:

```text
AsteroidState.variant
```

The game server owns authoritative asteroid creation and variant assignment. When the server creates a timed, fragment, or debug asteroid, it selects a variant index from the server asteroid catalog and stores that index on the runtime asteroid. State packet projection includes the selected index in each asteroid's exported state.

The client consumes the `variant` field from server state. It does not choose authoritative variants. It uses the received index to select asteroid presentation data from the client asteroid catalog, including the texture path and scene-level collision-polygon presentation.

The packet schema owns the existence, field name, and wire type of `AsteroidState.variant`. The asteroid variant catalog owns what each integer value means.

## Participating systems

```text
shared/asteroids/variants.toml
```

Owns the asteroid variant catalog values: stable variant ids, zero-based indexes, texture paths, collision-shape keys, stats-profile keys, drop-table keys, and spawn weights.

```text
shared/packets/gameplay.toml
```

Owns the realtime gameplay packet shape, including the `AsteroidState.variant` integer field.

```text
services/game-server/
```

Owns authoritative asteroid spawning, runtime asteroid state, variant selection, and outbound state packet projection.

```text
client/
```

Owns client-side consumption of server asteroid state and presentation lookup for the received variant index.

## Protocol authority

The game server is authoritative for:

```text
asteroid existence
asteroid runtime id
asteroid position
asteroid size
asteroid health
asteroid scale
asteroid variant index
```

The realtime packet schema is authoritative for:

```text
AsteroidState.variant field name
AsteroidState.variant JSON key
AsteroidState.variant wire type
StatePacket.asteroids map shape
```

The asteroid variant catalog is authoritative for:

```text
which indexes exist
which stable id each index maps to
which texture each index maps to
which spawn weights affect server selection
which intended collision-shape, stats-profile, and drop-table keys belong to each variant
```

The client is authoritative only for presentation consumption of the received index. It can wrap lookup helpers for safe presentation, but it must not treat wrapping as permission to invent protocol values.

## Message flow

The normal runtime flow is:

```text
server simulation decides to spawn an asteroid
-> server spawner selects a weighted variant index
-> server creates runtime.Asteroid with Variant set
-> server projects runtime.Asteroid.State()
-> server includes asteroid state in StatePacket.asteroids
-> packetcodec encodes the state packet as JSON
-> client receives the state packet
-> AsteroidSync reads AsteroidState.variant
-> asteroid scene receives the variant index
-> client asteroid catalog resolves presentation data for that index
```

There is no normal client-to-server asteroid variant request. Clients do not request asteroid variants during gameplay.

Debug asteroid spawning is different only at the command entry point. Devtools may request a debug asteroid spawn, but the server still builds a game-owned spawn plan, selects the debug-spawn variant through the asteroid catalog, applies the spawn through game-owned mutation, and exports the resulting asteroid through normal state packets.

## Packet surface

The asteroid variant protocol surface is part of the gameplay state packet.

Schema owner:

```text
shared/packets/gameplay.toml
```

Relevant struct:

```text
AsteroidState
```

Relevant field:

```text
name = "variant"
json = "variant"
type = "int"
```

Runtime JSON shape inside `StatePacket.asteroids` is:

```json
{
  "id": "asteroid-1",
  "x": 100.0,
  "y": 200.0,
  "size": 4,
  "health": 1,
  "scale": 1.4,
  "variant": 0
}
```

The `variant` value is a zero-based runtime index. It is not the stable catalog id string.

Current index mapping:

```text
0 -> asteroid_1
1 -> asteroid_2
2 -> asteroid_3
3 -> asteroid_4
4 -> asteroid_5
5 -> asteroid_6
6 -> asteroid_7
7 -> asteroid_8
```

## Source-of-truth files

Asteroid variant semantic source:

```text
shared/asteroids/variants.toml
```

Realtime packet field source:

```text
shared/packets/gameplay.toml
```

Current synchronized asteroid catalog outputs:

```text
services/game-server/internal/game/asteroids/variants.go
client/scripts/generated/asteroids/asteroid_variants.gd
```

Current packet generated outputs that carry or name the field:

```text
services/game-server/internal/game/runtime/packets_generated.go
client/scripts/generated/networking/packets/packets.gd
```

`shared/asteroids/variants.toml` is not currently listed as a first-class `tools/data_sync` domain. The Go and GDScript asteroid catalogs must be kept synchronized with the TOML source until an automated asteroid-variant generation path exists.

## Contract fields

The catalog fields relevant to this protocol are:

| Field                   | Contract role                                                                                     |
| ----------------------- | ------------------------------------------------------------------------------------------------- |
| `id`                    | Stable data/presentation identifier such as `asteroid_1`. Not sent as the runtime protocol value. |
| `index`                 | Zero-based runtime integer sent through `AsteroidState.variant`.                                  |
| `texture`               | Client presentation path resolved from the received index.                                        |
| `collision_shape`       | Intended collision-shape key associated with the variant data.                                    |
| `stats_profile`         | Intended stats-profile key associated with the variant data.                                      |
| `drop_table`            | Intended drop-table key associated with the variant data.                                         |
| `timed_spawn_weight`    | Server selection weight for timed asteroid spawns.                                                |
| `fragment_spawn_weight` | Server selection weight for fragment asteroid spawns.                                             |
| `debug_spawn_weight`    | Server selection weight for debug asteroid spawns.                                                |

Current source values make all eight variants eligible for timed, fragment, and debug spawns with equal weight:

```text
timed_spawn_weight = 1.0
fragment_spawn_weight = 1.0
debug_spawn_weight = 1.0
```

All current variants use:

```text
collision_shape = "asteroid:0"
stats_profile = "standard"
drop_table = "basicasteroids"
```

## Server responsibilities

The game server owns runtime variant assignment.

Timed asteroid spawns select variants through:

```go
asteroids.RandomTimedSpawnVariantIndex()
```

Asteroid fragment spawns select variants through:

```go
asteroids.RandomFragmentSpawnVariantIndex()
```

Debug asteroid spawns select variants through:

```go
asteroids.RandomDebugSpawnVariantIndex()
```

The server must not use raw hardcoded variant pools for asteroid variant selection. `rand.Intn(4) + 1` is still used for asteroid size selection, but asteroid size is not the variant contract.

The server stores the selected index on:

```text
runtime.Asteroid.Variant
```

It exports that value through:

```text
runtime.Asteroid.State()
runtime.AsteroidState.Variant
StatePacket.asteroids
```

The server packet projection must include the selected variant index for asteroid state sent to clients.

## Client responsibilities

The client consumes the server-provided variant index from asteroid state.

The current client flow is:

```text
AsteroidSync.apply()
-> read Packets.FIELD_VARIANT from asteroid state
-> store the integer by asteroid id
-> initialize the asteroid scene
-> call asteroid_node.set_asteroid_variant(...)
-> resolve texture through AsteroidVariants.texture_path_for_index(index)
```

The client asteroid scene applies the selected texture and scene-level collision polygon presentation. This is presentation behavior. Authoritative collision behavior remains server-owned.

Client lookup helpers wrap indexes for safe presentation lookup. For example, index `8` resolves back to the first catalog entry through helper lookup. This protects presentation code from out-of-range access, but valid server-emitted protocol values should still come from the current catalog index set.

## Compatibility expectations

The compatibility contract is:

```text
AsteroidState.variant remains an integer field.
Runtime variant indexes remain zero-based.
Client and server catalogs keep the same index order.
Catalog ids remain stable labels.
Packet schema changes and catalog semantic changes are updated together when required.
```

Adding, removing, or reordering variants is a protocol-affecting change because existing server/client builds must agree on index meaning.

Safe lookup wrapping is not a version-negotiation mechanism. If a server emits an index that the client only handles through wrapping because the catalogs are stale or mismatched, that is catalog drift to investigate.

`constants.AsteroidVariants` must not be reintroduced as variant-count authority. The asteroid variant catalog owns the list and count.

## Validation and testing

Server tests cover:

```text
variant count
zero-based indexes
ByIndex wrapping
timed spawn eligibility
fragment spawn eligibility
debug spawn eligibility
required current fields
current equal weights
weighted selection skipping zero-weight variants
```

Client tests cover:

```text
catalog count
texture lookup for index 0
texture lookup for index 7
wrapped lookup for index 8
current collision-shape lookup
current spawn-weight lookup
```

Focused server verification:

```bash
cd services/game-server
go test -buildvcs=false ./internal/game/asteroids
go test -buildvcs=false ./internal/game/spawning
go test -buildvcs=false ./internal/devtools
```

Focused client verification:

```bash
cd client
godot --headless --path . -s addons/gut/gut_cmdln.gd -gtest=res://tests/unit/entities/test_asteroid_variants.gd
```

Packet-shape verification when the `variant` field or state packet shape changes:

```bash
data-sync -validate -packets
data-sync -check -packets -go -gds
```

Useful drift checks:

```bash
grep -R "constants.AsteroidVariants" -n services client shared docs
grep -R "Variant: rand.Intn" -n services/game-server/internal/game services/game-server/internal/devtools
grep -R "asteroid_textures" -n client/scripts
```

## Code map

Packet schema and generated packet outputs:

```text
shared/packets/gameplay.toml
services/game-server/internal/game/runtime/packets_generated.go
client/scripts/generated/networking/packets/packets.gd
```

Asteroid variant source and synchronized catalogs:

```text
shared/asteroids/variants.toml
services/game-server/internal/game/asteroids/variants.go
client/scripts/generated/asteroids/asteroid_variants.gd
```

Server runtime and projection:

```text
services/game-server/internal/game/spawning/spawner.go
services/game-server/internal/devtools/spawn_asteroid.go
services/game-server/internal/game/runtime/state.go
services/game-server/internal/game/runtime/asteroid.go
services/game-server/internal/game/state_packet.go
services/game-server/internal/networking/outbound/gameplay_presentation.go
```

Client consumption and presentation:

```text
client/scripts/world/asteroid_sync.gd
client/scripts/entities/asteroid.gd
client/scenes/asteroid.tscn
client/tests/unit/entities/test_asteroid_variants.gd
```

Tests:

```text
services/game-server/internal/game/asteroids/variants_test.go
client/tests/unit/entities/test_asteroid_variants.gd
```

Important non-ownership boundaries:

```text
shared/collisions/collision_shapes.json
shared/drop_tables/basicasteroids.toml
services/game-server/internal/game/physics/
services/game-server/internal/game/drops/
client/tools/export_collision_shapes.gd
tools/data_sync/
```

## Related docs

* [Protocol](./!INDEX.md)
* [Asteroid Variants Data](../data/asteroid-variants-data.md)
* [Packet Schemas](../data/packet-schemas.md)
* [Source Of Truth Map](../data/source-of-truth-map.md)
* [Server Asteroid Spawning And Variants](../services/game-server/simulation/world/asteroid-spawning-and-variants.md)
* [Client Asteroid Variant Presentation](../services/client/world-sync/asteroid-variant-presentation.md)
* [Client Entity Sync Owners](../services/client/world-sync/entity-sync-owners.md)
* [Game Server](../services/game-server/!INDEX.md)
* [Client](../services/client/!INDEX.md)

## Notes

The current protocol sends only the runtime variant index. It does not send the stable variant id, texture path, collision-shape key, stats-profile key, drop-table key, or spawn weight.

Current server collision-body construction uses the runtime asteroid variant index against the loaded collision-shape catalog. It does not currently resolve the asteroid variant catalog `collision_shape` string as a separate key during collision-body construction.

Current pickup drop integration does not select a drop table from the runtime asteroid variant. It uses the basic asteroid drop-table path directly.

All current variants share the same collision-shape key, stats profile, drop table, and spawn weights.
