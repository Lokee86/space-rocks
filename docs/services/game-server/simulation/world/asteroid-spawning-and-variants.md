# Asteroid Spawning And Variants

Parent index: [Game Server Simulation World](./!INDEX.md)

## Purpose

This document describes the game-server implementation for asteroid spawning and asteroid variant assignment.

It explains how timed, fragment, and debug asteroid spawns are planned, how variants are selected, how spawned asteroids are stored in runtime state, and how asteroid variant ids participate in state projection and collision-shape lookup.

## Overview

Asteroid spawning is a game-server simulation responsibility.

The current server flow is:

```text
simulation tick
-> step player/session state
-> if match is still active, evaluate asteroid spawn timing
-> for each active camera view, spawn a batch when the interval elapses
-> choose an offscreen spawn position
-> plan asteroid velocity, size, and variant
-> allocate asteroid id
-> create runtime asteroid
-> store asteroid in game.entities.Asteroids
-> project asteroid state through StatePacket.asteroids
```

Asteroid variants are assigned by the server when an asteroid is created. The server stores the assigned variant index on the runtime asteroid and exports that index in asteroid state. The client consumes the exported variant id for presentation, but does not choose authoritative asteroid variants.

The asteroid variant catalog is sourced from shared asteroid data and consumed by server runtime helpers. Spawn paths use weighted catalog helpers instead of raw variant-count randomization.

## Code root

```text
services/game-server/
```

Primary implementation areas:

```text
services/game-server/internal/game/
services/game-server/internal/game/spawning/
services/game-server/internal/game/asteroids/
services/game-server/internal/game/runtime/
services/game-server/internal/game/physics/
services/game-server/internal/devtools/
shared/asteroids/
shared/collisions/
shared/constants/
```

## Responsibilities

Asteroid spawning and variant ownership includes:

* Scheduling timed asteroid spawns during active match simulation.
* Gating timed spawns on camera-view availability, match state, and world simulation options.
* Choosing offscreen spawn positions relative to active camera views.
* Planning asteroid velocity, size, reason, entity type, and variant.
* Selecting timed, fragment, and debug spawn variants through weighted catalog helpers.
* Allocating stable runtime asteroid ids.
* Creating runtime asteroid entities.
* Storing spawned asteroids in `game.entities.Asteroids`.
* Spawning asteroid fragments after projectile-caused asteroid destruction.
* Supporting debug asteroid spawn requests through game-owned apply seams.
* Storing the selected variant on `runtime.Asteroid`.
* Exporting asteroid variant ids through `runtime.AsteroidState`.
* Looking up asteroid collision bodies from runtime asteroid variant and size.
* Reporting total asteroid spawn count through state packet projection.

## Does not own

Asteroid spawning and variants do not own:

* Client asteroid rendering.
* Client texture selection.
* Client scene collision-polygon presentation.
* Packet schema source-of-truth files.
* Data-sync implementation.
* Collision shape export from Godot scenes.
* Collision primitive math.
* Projectile/asteroid collision detection.
* Damage resolution.
* Scoring policy.
* Pickup drop-table evaluation.
* Pickup entity lifecycle, collection, or effects.
* Room membership or match lifecycle ownership.
* WebSocket transport.
* Durable player data or account persistence.
* Devtools command routing.

Those systems may consume asteroid state or trigger asteroid-related effects, but they own their own boundaries.

## Domain roles

The game server is authoritative for asteroid existence and asteroid runtime state.

For asteroid spawning, the server decides:

* when a timed asteroid batch spawns
* which camera view a timed spawn is planned around
* where the asteroid appears
* which direction and speed it moves
* which size it has
* which variant index it receives
* which runtime id identifies it
* when fragments are created
* when spawned asteroids are removed from runtime storage

The client observes asteroid state through normal state packets. It may render asteroid variants differently, but it does not create authoritative asteroids, choose authoritative variants, or decide whether an asteroid exists.

## Timed spawn scheduling

Timed asteroid spawning runs from the simulation step.

The active simulation path calls:

```text
Game.Step
-> stepAsteroidSpawning
```

Timed spawning is skipped when the match is over. After match over, the step path still advances and cleans up existing asteroids, bullets, pickups, and radial effects, but it does not call `stepAsteroidSpawning`.

Timed spawning requires:

```text
worldSimulationOptions.CanSpawnAsteroids() == true
hasCameraViews() == true
```

When there are no camera views, the asteroid spawn timer resets to `0`.

When spawning is allowed, `asteroidSpawnElapsed` accumulates tick delta. Once it reaches `constants.AsteroidSpawnInterval`, the timer resets and the game spawns one asteroid batch for each active camera view.

Current generated constants include:

```text
AsteroidSpawnInterval = 3.0
AsteroidSpawnBatchSize = 3
AsteroidSpawnMargin = 160.0
AsteroidDespawnMargin = 320.0
AsteroidMinSpeed = 115.0
AsteroidMaxSpeed = 210.0
AsteroidAimRandomnessDegrees = 30.0
AsteroidSizeScale = 0.35
```

## Timed spawn position

Timed spawns are planned around active camera views.

For each asteroid in a batch, the game chooses a random offscreen position around the target camera view:

```text
spawnAsteroidBatch
-> spawnAsteroid
-> randomAsteroidSpawnPosition
-> randomOffscreenPosition
```

`randomOffscreenPosition` selects one of four sides around the camera view:

```text
top
right
bottom
left
```

The initial spawn margin is `constants.AsteroidSpawnMargin`.

`randomAsteroidSpawnPosition` rejects positions that are onscreen for any active camera view. If repeated attempts keep landing onscreen, the search margin expands every 16 attempts by another `AsteroidSpawnMargin`.

The selected spawn position is normalized through toroidal world space before the spawn plan is built.

## Timed spawn plan

Timed spawn planning is owned by `spawning.Spawner`.

The root game code supplies:

```text
spawn position
target camera position
```

The spawner builds an `AsteroidSpawnPlan` with:

```text
EntityType = asteroid
Reason     = timed_asteroid
Position   = normalized offscreen position
Velocity   = aimed randomized velocity
Size       = random integer from 1 through 4
Variant    = weighted timed-spawn variant index
```

Velocity is aimed from the spawn position toward the target camera position using wrapped world direction. The direction is rotated by a random offset within `AsteroidAimRandomnessDegrees`, then multiplied by a random speed between `AsteroidMinSpeed` and `AsteroidMaxSpeed`.

Variant selection uses:

```go
asteroids.RandomTimedSpawnVariantIndex()
```

The spawn plan does not mutate game state. Mutation happens only when root game code applies the plan.

## Applying asteroid spawns

Asteroid spawn application is owned by root game code:

```text
applyAsteroidSpawn
```

The apply helper:

```text
allocates the next unique asteroid id
creates runtime.NewAsteroid(...)
stores it in game.entities.Asteroids
returns the runtime asteroid
```

Asteroid ids are allocated by `spawning.Spawner.NextAsteroidID`. The current id format is:

```text
asteroid-1
asteroid-2
asteroid-3
```

The allocator checks the active asteroid map and skips ids that already exist.

`spawning.Spawner.TotalAsteroidsSpawned()` returns the current asteroid id counter. State packet projection exposes this as `StatePacket.total_asteroids`.

## Fragment spawns

Fragment spawning runs after projectile-caused asteroid destruction.

The current destruction path is:

```text
applyProjectileAsteroidDestruction
-> scoring policy evaluation
-> score award application
-> asteroid.MarkPendingDespawn
-> spawnAsteroidFragments
-> maybeDropPickupFromAsteroidLocked
```

Fragment size is:

```text
source asteroid size - 1
```

If the fragment size is `0` or lower, no fragments spawn.

Otherwise, `spawning.Spawner.PlanAsteroidFragmentSpawns` creates two fragment spawn plans. Each fragment plan uses:

```text
EntityType = asteroid
Reason     = asteroid_fragment
Position   = source asteroid position
Velocity   = random unit direction * random asteroid speed
Size       = source size - 1
Variant    = weighted fragment-spawn variant index
```

Fragment variants are selected independently. They do not inherit the source asteroid variant.

Variant selection uses:

```go
asteroids.RandomFragmentSpawnVariantIndex()
```

Each fragment plan is applied through the same `applyAsteroidSpawn` helper used by timed asteroid spawns.

## Debug asteroid spawns

Debug asteroid spawning is routed through devtools, but game-owned mutation still goes through the asteroid spawn apply seam.

The debug spawn path builds an `AsteroidSpawnPlan` in:

```text
services/game-server/internal/devtools/spawn_asteroid.go
```

A debug asteroid spawn plan uses:

```text
EntityType = asteroid
Reason     = debug_asteroid
Position   = normalized requested position
Velocity   = requested direction or random fallback direction, multiplied by random asteroid speed
Size       = random integer from 1 through 4
Variant    = weighted debug-spawn variant index
```

Variant selection uses:

```go
asteroids.RandomDebugSpawnVariantIndex()
```

Devtools applies the plan through:

```text
Game.DevtoolsApplyAsteroidSpawnPlan
-> applyAsteroidSpawn
```

`internal/game` exposes the narrow apply helper for devtools. Devtools command parsing, command routing, and debug UI behavior are not owned by this document.

## Variant catalog

The server asteroid variant catalog lives in:

```text
services/game-server/internal/game/asteroids/variants.go
```

The shared source data lives in:

```text
shared/asteroids/variants.toml
```

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

Variant indexes are zero-based runtime values. Variant ids such as `asteroid_1` are stable data/presentation identifiers, not the runtime value stored on the asteroid entity.

The current variant fields include:

```text
ID
Index
CollisionShape
StatsProfile
DropTable
TimedSpawnWeight
FragmentSpawnWeight
DebugSpawnWeight
```

Current source data gives all eight variants:

```text
collision_shape = "asteroid:0"
stats_profile = "standard"
drop_table = "basicasteroids"
timed_spawn_weight = 1.0
fragment_spawn_weight = 1.0
debug_spawn_weight = 1.0
```

Those equal weights mean all current variants are eligible for timed, fragment, and debug spawns with equal probability.

## Variant selection rules

The asteroid variant package owns weighted selection for each spawn source.

Spawn-source helpers are:

```go
TimedSpawnVariants()
FragmentSpawnVariants()
DebugSpawnVariants()
RandomTimedSpawnVariantIndex()
RandomFragmentSpawnVariantIndex()
RandomDebugSpawnVariantIndex()
```

A variant is eligible for a spawn source when that source weight is greater than `0.0`.

A weight of `0.0` or lower excludes that variant from the spawn source.

Random selection sums all positive weights, chooses a random threshold, and returns the selected variant index.

If total positive weight is `0.0` or lower, the fallback selected index is `0`.

Server spawn code should call the catalog helper for the relevant spawn source. It should not use raw `rand.Intn` variant pools or reintroduce a generated constants variant count.

`rand.Intn(4) + 1` is still used for asteroid size selection. That size randomization is separate from variant selection.

## Runtime asteroid state

Asteroids are stored as `runtime.Asteroid`.

A newly spawned asteroid stores:

```text
ID
X
Y
Velocity
Size
Variant
Health
CollisionDamage
```

`runtime.NewAsteroid` initializes health and collision damage from generated constants:

```text
AsteroidHealth
AsteroidCollisionDamage
```

The runtime `Variant` field is the server-selected variant index.

Asteroid state projection uses:

```text
runtime.Asteroid.State()
```

The exported asteroid state contains:

```text
id
x
y
size
health
scale
variant
```

`scale` is derived from:

```text
float64(size) * constants.AsteroidSizeScale
```

`Game.statePacket` projects all runtime asteroids into:

```text
StatePacket.asteroids
```

The client consumes the projected `variant` value for asteroid presentation.

## Collision-shape lookup

Runtime asteroids build authoritative collision bodies through:

```text
runtime.Asteroid.CollisionBody
```

That method calls:

```go
catalog.AsteroidShape(asteroid.Variant, asteroid.Size)
```

The collision shape catalog is loaded when the game is created:

```text
Game.New
-> physics.LoadCollisionShapeCatalog
-> shared/collisions/collision_shapes.json
```

`CollisionShapeCatalog.AsteroidShape` wraps the runtime variant index against the loaded asteroid collision-shape list, then scales the selected shape by:

```text
float64(size) * constants.AsteroidSizeScale
```

The current collision shape catalog contains one asteroid shape, so all current asteroid variants resolve to the same authoritative server collision polygon after index wrapping.

The server collision body path currently uses the runtime variant index and loaded collision-shape catalog. It does not currently resolve the variant catalog `CollisionShape` string as a separate key during runtime collision-body creation.

## Protocol and APIs

Asteroid spawning has no inbound HTTP API.

Normal clients do not send a spawn request for gameplay asteroids. Timed and fragment asteroid spawning are server-authored simulation effects.

Clients observe asteroid spawning and variants through normal gameplay state output:

```text
StatePacket.asteroids
StatePacket.total_asteroids
```

`StatePacket.asteroids` carries each asteroid runtime state, including the server-selected `variant` index. `StatePacket.total_asteroids` carries the spawner's total asteroid id counter.

Debug asteroid spawning is available only through devtools command handling. The devtools command path is a debug surface, not a gameplay authority surface. Devtools builds a debug asteroid spawn plan, then applies it through game-owned spawn mutation.

The packet schema and WebSocket transport are not owned by asteroid spawning. This document only covers the server runtime behavior behind spawned asteroid state.

## Data ownership

Asteroid spawning reads generated and shared data, but it does not own the data pipeline.

Source data includes:

```text
shared/asteroids/variants.toml
shared/constants/server_constants.toml
shared/constants/server_entities.toml
shared/collisions/collision_shapes.json
```

Runtime server outputs and consumers include:

```text
services/game-server/internal/game/asteroids/variants.go
services/game-server/internal/constants/constants.go
services/game-server/internal/game/physics/collision_shapes.go
```

Asteroid spawning mutates:

```text
game.asteroidSpawnElapsed
game.spawner.nextAsteroidID
game.entities.Asteroids
```

It reads:

```text
game.cameraViews
game.worldSimulationOptions
game.collisionShapes
generated asteroid constants
generated asteroid variant catalog
runtime asteroid state during fragment spawning
```

It does not persist asteroid state outside the active game instance.

## Invariants

Asteroid spawning and variants must preserve these rules:

* The game server is authoritative for asteroid creation.
* The client must observe asteroid existence through server state.
* Timed asteroid spawning must not run after match over.
* Timed asteroid spawning requires at least one active camera view.
* Timed asteroid spawn positions must be offscreen for all active camera views.
* Root game code owns mutation into `game.entities.Asteroids`.
* Spawn planning should remain separate from spawn application.
* Timed, fragment, and debug spawns must select variants through the asteroid catalog helpers.
* Variant randomization must not use raw `rand.Intn` pools.
* Runtime asteroid state stores the selected variant index.
* State packet projection must preserve the runtime asteroid variant.
* Fragment spawns must receive newly selected fragment variants.
* Debug asteroid spawning must apply through game-owned mutation seams.
* Collision-body construction must use server-loaded collision shapes, not client presentation shapes.
* Packet schema ownership remains outside the spawning implementation.

## Code map

Primary game integration files:

```text
services/game-server/internal/game/simulation.go
services/game-server/internal/game/simulation_asteroids.go
services/game-server/internal/game/spawning.go
services/game-server/internal/game/visibility.go
services/game-server/internal/game/asteroid_destruction.go
services/game-server/internal/game/state_packet.go
services/game-server/internal/game/game.go
services/game-server/internal/game/world_simulation_options.go
```

Spawner package:

```text
services/game-server/internal/game/spawning/spawner.go
```

Asteroid variant catalog:

```text
services/game-server/internal/game/asteroids/variants.go
services/game-server/internal/game/asteroids/variants_test.go
```

Runtime asteroid storage and projection:

```text
services/game-server/internal/game/runtime/state.go
services/game-server/internal/game/runtime/asteroid.go
services/game-server/internal/game/runtime/packets_generated.go
```

Collision-shape support:

```text
services/game-server/internal/game/physics/collision_shapes.go
services/game-server/internal/game/physics/collision_shapes_test.go
services/game-server/internal/game/collisions.go
```

Debug spawn support:

```text
services/game-server/internal/devtools/spawn_asteroid.go
services/game-server/internal/game/export_devtools_spawn.go
```

Source and generated data:

```text
shared/asteroids/variants.toml
shared/collisions/collision_shapes.json
shared/constants/server_constants.toml
shared/constants/server_entities.toml
services/game-server/internal/constants/constants.go
```

Relevant tests:

```text
services/game-server/internal/game/asteroids/variants_test.go
services/game-server/internal/game/simulation_match_over_test.go
services/game-server/internal/game/world_simulation_options_test.go
services/game-server/internal/game/physics/collision_shapes_test.go
services/game-server/internal/devtools/clear_entities_test.go
services/game-server/internal/devtools/shape_catalog_test.go
```

Important non-ownership boundaries:

```text
services/game-server/internal/networking/
services/game-server/internal/game/damage/
services/game-server/internal/game/scoring/
services/game-server/internal/game/drops/
services/game-server/internal/game/pickups/
services/game-server/internal/devtools/
client/
shared/packets/
tools/data_sync/
```

## Tests and verification

Current coverage includes:

* asteroid variant count remains eight
* variant indexes remain zero-based
* `ByIndex` wraps across the variant catalog
* timed, fragment, and debug spawn variant lists include all current variants
* current variant entries keep required fields and weights
* weighted selection skips zero-weight variants
* match-over simulation skips timed asteroid spawning
* match-over simulation can clean up safe entities without panic
* world simulation freeze flags gate spawning, asteroid movement, bullet movement, and collisions
* collision-shape catalog loading and conversion behavior
* devtools clear-entity behavior for spawned asteroid state

Useful verification commands:

```bash
cd services/game-server
go test -buildvcs=false ./internal/game/asteroids
go test -buildvcs=false ./internal/game -run 'Asteroid|Spawn|MatchOver|WorldSimulation'
go test -buildvcs=false ./internal/game/physics
go test -buildvcs=false ./internal/devtools
go test -buildvcs=false ./...
```

Useful data verification commands:

```bash
data-sync -check -constants -go
data-sync -check -packets -go
```

## Related docs

* [Game Server Simulation World](./!README.md)
* [Game Server Simulation](../!README.md)
* [Game Server](../../!README.md)
* [Visibility And Despawn](visibility-and-despawn.md)
* [Toroidal Space And Motion](toroidal-space-and-motion.md)
* [Collision Shapes](collision-shapes.md)
* [Physics](physics.md)
* [Collision To Damage Flow](../combat/collision-to-damage-flow.md)
* [Pickup Drop Integration](../pickups/pickup-drop-integration.md)
* [State Packet Projection](../runtime/state-packet-projection.md)
* [Asteroid Variants Data](../../../../data/stubs/asteroid-variants-data.md)
* [Asteroid Variant Contract](../../../../protocol/stubs/asteroid-variant-contract.md)
* [Client Asteroid Variant Presentation](../../../client/world-sync/asteroid-variant-presentation.md)
* [Current System Limits](../../../../limits/current-system-limits.md)

## Notes

The generated server variant catalog includes `CollisionShape`, `StatsProfile`, and `DropTable` fields. Current asteroid spawning stores only the selected variant index on runtime asteroids.

Current pickup drop integration does not select a drop table from the runtime asteroid variant. It uses `basicasteroids` directly.

Current server collision-body lookup does not resolve the variant catalog `CollisionShape` field by string key. It indexes the loaded asteroid collision-shape list by runtime variant index with wrapping.

All current variants use equal spawn weights, the same stats profile, the same drop table, and the same collision shape.
