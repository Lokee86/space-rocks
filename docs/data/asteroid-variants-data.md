## Asteroid Variants Data

Parent index: [Data](./!INDEX.md)

## Purpose

This document describes the asteroid variant source data, the generated or source-derived outputs that consume it, and the validation expectations that keep the client and game server aligned on asteroid variant meaning.

## Overview

`shared/asteroids/variants.toml` is the current asteroid variant source of truth.

The catalog defines the stable asteroid variant entries used by the game server and the client. The server uses the catalog to choose authoritative runtime variant indexes for timed, fragment, and debug asteroid spawns. The client consumes the server-provided runtime variant index and uses the generated client catalog to select presentation data such as asteroid texture paths.

Asteroid variant indexes are zero-based runtime values. Variant ids such as `asteroid_1` are stable data and presentation identifiers, not the runtime value stored on an asteroid entity or sent through asteroid state.

The current catalog contains eight variants:

```text
asteroid_1 -> index 0
asteroid_2 -> index 1
asteroid_3 -> index 2
asteroid_4 -> index 3
asteroid_5 -> index 4
asteroid_6 -> index 5
asteroid_7 -> index 6
asteroid_8 -> index 7
```

Safe lookup helpers wrap out-of-range indexes. For example, index `8` resolves back to the first catalog entry through helper lookup. Runtime code should not treat wrapped lookup as permission to emit arbitrary variant indexes.

## Source files

Primary source:

```text
shared/asteroids/variants.toml
```

The source file contains one `\[\[variants\]\]` entry per asteroid variant.

Current fields:

```text
id
index
texture
collision_shape
stats_profile
drop_table
timed_spawn_weight
fragment_spawn_weight
debug_spawn_weight
```

Field ownership:

| Field                   | Meaning                                                            |
| ----------------------- | ------------------------------------------------------------------ |
| `id`                    | Stable variant identifier used for data and presentation matching. |
| `index`                 | Zero-based runtime index used by server state and lookup helpers.  |
| `texture`               | Godot resource path used by the client presentation catalog.       |
| `collision_shape`       | Data key for the intended asteroid collision-shape association.    |
| `stats_profile`         | Data key for intended per-variant asteroid stat behavior.          |
| `drop_table`            | Data key for intended per-variant drop-table association.          |
| `timed_spawn_weight`    | Weighted eligibility for normal timed asteroid spawning.           |
| `fragment_spawn_weight` | Weighted eligibility for asteroid fragment spawning.               |
| `debug_spawn_weight`    | Weighted eligibility for debug asteroid spawning.                  |

Current source values give all eight variants:

```text
collision_shape = "asteroid:0"
stats_profile = "standard"
drop_table = "basicasteroids"
timed_spawn_weight = 1.0
fragment_spawn_weight = 1.0
debug_spawn_weight = 1.0
```

Those equal weights make all current variants eligible for timed, fragment, and debug spawning with equal probability.

## Configuration

Asteroid variants are not currently declared as a first-class `tools/data_sync` domain.

`tools/data_sync/config.toml` currently configures source-of-truth paths for constants, packets, drop tables, and player-data schema material. It does not define a `[sot.asteroids]` section, an asteroid-variant output target, or an asteroid-variant CLI flag.

The current asteroid variant pipeline therefore has source data and source-derived outputs, but it is not covered by `data-sync -check`, `data-sync -push`, or `data-sync -diff`.

Related but separate configuration and source data:

```text
shared/packets/gameplay.toml
shared/collisions/collision_shapes.json
shared/drop_tables/basicasteroids.toml
tools/data_sync/config.toml
```

`shared/packets/gameplay.toml` owns the `AsteroidState.variant` packet field. It does not own the variant catalog.

`shared/collisions/collision_shapes.json` owns exported collision geometry. It does not own asteroid variant identity, spawn weights, or texture paths.

`shared/drop_tables/basicasteroids.toml` owns drop-table contents. It does not own asteroid variant identity or spawn weighting.

## Generated outputs

Current generated or source-derived outputs:

```text
client/scripts/generated/asteroids/asteroid_variants.gd
services/game-server/internal/game/asteroids/variants.go
```

The client output exposes:

```text
VARIANTS
count()
texture_path_for_index(index)
collision_shape_for_index(index)
timed_spawn_weight_for_index(index)
fragment_spawn_weight_for_index(index)
debug_spawn_weight_for_index(index)
```

The server output exposes:

```text
Variants
Count()
ByIndex(index)
TimedSpawnVariants()
FragmentSpawnVariants()
DebugSpawnVariants()
RandomTimedSpawnVariantIndex()
RandomFragmentSpawnVariantIndex()
RandomDebugSpawnVariantIndex()
```

The server helpers are the runtime selection surface for spawn code. Server spawn paths should use the helper matching the spawn source instead of raw random selection over a hardcoded variant count.

## Consumers

Game-server consumers:

```text
services/game-server/internal/game/asteroids/variants.go
services/game-server/internal/game/spawning/spawner.go
services/game-server/internal/devtools/spawn_asteroid.go
services/game-server/internal/game/runtime/asteroid.go
```

The game server consumes the catalog to:

* select timed spawn variants
* select fragment spawn variants
* select debug spawn variants
* store the selected variant index on `runtime.Asteroid`
* export the selected variant index through `AsteroidState.variant`

Client consumers:

```text
client/scripts/generated/asteroids/asteroid_variants.gd
client/scripts/world/asteroid_sync.gd
client/scripts/entities/asteroid.gd
client/scenes/asteroid.tscn
client/assets/asteroids/
```

The client consumes the catalog to:

* read server-provided asteroid variant indexes from asteroid state
* select asteroid texture paths
* apply asteroid scene presentation for the selected variant
* wrap lookup indexes safely during presentation lookup

Protocol consumers:

```text
shared/packets/gameplay.toml
services/game-server/internal/game/runtime/packets_generated.go
client/scripts/generated/networking/packets/packets.gd
```

The realtime protocol carries the selected runtime variant as `AsteroidState.variant`. The packet schema owns the field shape; the asteroid variant catalog owns the meaning of the index value.

## Pipeline usage

Current update workflow:

```text
1. Edit shared/asteroids/variants.toml.
2. Update the client catalog output in client/scripts/generated/asteroids/asteroid_variants.gd.
3. Update the server catalog output in services/game-server/internal/game/asteroids/variants.go.
4. Update server and client tests when the expected catalog size, indexes, or fields change.
5. Run focused server and client verification.
6. Check that no hardcoded fallback variant pools were reintroduced.
```

Until asteroid variants are added to `tools/data_sync`, do not assume `data-sync` can regenerate or validate this catalog.

Server runtime code should keep variant selection behind the asteroid catalog helpers:

```text
RandomTimedSpawnVariantIndex()
RandomFragmentSpawnVariantIndex()
RandomDebugSpawnVariantIndex()
```

Client runtime code should keep texture lookup behind the generated asteroid variant catalog:

```text
AsteroidVariants.texture_path_for_index(index)
```

The old constants-owned variant-count model should not return. `constants.AsteroidVariants` is not the source of truth for variant count.

## Validation commands

Focused server verification:

```bash
cd services/game-server
go test -buildvcs=false ./internal/game/asteroids
go test -buildvcs=false ./internal/game -run 'Asteroid|Spawn|MatchOver|WorldSimulation'
go test -buildvcs=false ./internal/devtools
```

Focused client verification:

```bash
cd client
godot --headless --path . -s addons/gut/gut_cmdln.gd -gtest=res://tests/unit/entities/test_asteroid_variants.gd
```

Useful drift checks:

```bash
grep -R "constants.AsteroidVariants" -n services client shared docs
grep -R "Variant: rand.Intn" -n services/game-server/internal/game services/game-server/internal/devtools
grep -R "asteroid_textures" -n client/scripts
```

`data-sync` validation may still be useful for adjacent packet, constants, and drop-table sources, but it does not currently validate asteroid variant catalog drift.

## Failure modes

Common failure modes:

* `shared/asteroids/variants.toml` changes without matching client and server catalog updates.
* Client and server catalogs disagree on variant count.
* Client and server catalogs disagree on index order.
* A variant id is duplicated.
* A variant index is duplicated.
* Indexes stop being zero-based and contiguous.
* A texture path points to a missing Godot asset.
* Server spawning uses raw `rand.Intn` or a hardcoded variant count instead of catalog helpers.
* `constants.AsteroidVariants` is reintroduced as variant-count authority.
* Client asteroid rendering reintroduces a hardcoded texture array.
* Spawn weights are edited in one output but not the other.
* `AsteroidState.variant` packet shape changes without corresponding server and client updates.
* Runtime code assumes `collision_shape`, `stats_profile`, or `drop_table` fields are fully enforced when current runtime paths still use narrower behavior.

## Code or source map

Source data:

```text
shared/asteroids/variants.toml
```

Generated or source-derived outputs:

```text
client/scripts/generated/asteroids/asteroid_variants.gd
services/game-server/internal/game/asteroids/variants.go
```

Server consumers:

```text
services/game-server/internal/game/asteroids/variants.go
services/game-server/internal/game/asteroids/variants_test.go
services/game-server/internal/game/spawning/spawner.go
services/game-server/internal/devtools/spawn_asteroid.go
services/game-server/internal/game/runtime/asteroid.go
```

Client consumers:

```text
client/scripts/world/asteroid_sync.gd
client/scripts/entities/asteroid.gd
client/scenes/asteroid.tscn
client/tests/unit/entities/test_asteroid_variants.gd
client/assets/asteroids/asteroid1.png
client/assets/asteroids/asteroid2.png
client/assets/asteroids/asteroid3.png
client/assets/asteroids/asteroid4.png
client/assets/asteroids/asteroid5.png
client/assets/asteroids/asteroid6.png
client/assets/asteroids/asteroid7.png
client/assets/asteroids/asteroid8.png
```

Related protocol and adjacent data:

```text
shared/packets/gameplay.toml
shared/collisions/collision_shapes.json
shared/drop_tables/basicasteroids.toml
tools/data_sync/config.toml
```

Important non-ownership boundaries:

```text
tools/data_sync/
shared/packets/
shared/collisions/
shared/drop_tables/
services/game-server/internal/game/drops/
services/game-server/internal/game/physics/
client/tools/export_collision_shapes.gd
```

## Related docs

* [Data](./!INDEX.md)
* [Data Sync And SSOT Pipeline](data-sync-and-ssot-pipeline.md)
* [Source Of Truth Map](source-of-truth-map.md)
* [Server Asteroid Spawning And Variants](../services/game-server/simulation/world/asteroid-spawning-and-variants.md)
* [Client Asteroid Variant Presentation](../services/client/world-sync/asteroid-variant-presentation.md)
* [Client Entity Sync Owners](../services/client/world-sync/entity-sync-owners.md)
* [Packet Schemas](packet-schemas.md)
* [Current System Limits](../limits/current-system-limits.md)

## Notes

The current server collision-body path uses the runtime asteroid variant index against the loaded collision-shape catalog. It does not currently resolve the variant catalog `collision_shape` string as a separate runtime key during collision-body construction.

Current pickup drop integration does not select drop tables from the runtime asteroid variant. It uses the basic asteroid drop-table path directly.

All current variants use equal weights, the same collision-shape key, the same stats profile, and the same drop table.
