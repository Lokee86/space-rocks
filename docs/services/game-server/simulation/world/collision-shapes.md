# Collision Shapes

Parent index: [Game Server Simulation World](./!README.md)

## Purpose

This document describes the game-server collision-shape support boundary.

It explains how the server loads shared collision-shape data, converts imported shapes into runtime collision bodies, and exposes those bodies to collision detection, targeting, respawn safety, radial effects, pickup collection, and devtools collision telemetry.

## Overview

Collision shapes are a game-server simulation support system implemented primarily under:

```text
services/game-server/internal/game/physics/
```

The server loads the shared collision-shape catalog from:

```text
shared/collisions/collision_shapes.json
```

That JSON is generated from Godot scene collision nodes by the client-side export tool. The game server consumes the generated output; it does not author or own the source collision-shape data.

At runtime, the server stores the loaded catalog on each `Game` instance:

```go
collisionShapes physics.CollisionShapeCatalog
```

Simulation systems then ask runtime entities to build `physics.CollisionBody` values from that catalog. Those bodies carry:

```text
entity id
world position
rotation when relevant
collision shape
```

The collision-shape boundary is deliberately narrower than gameplay collision ownership:

```text
shared collision data
-> physics.CollisionShapeCatalog
-> runtime entity CollisionBody(...)
-> collision, targeting, respawn, radial, pickup, or devtools consumer
```

The `physics` package owns imported shape conversion, primitive shape representation, outline generation, point containment, and primitive overlap math. It does not decide which gameplay pairs should be checked or what happens after a collision is detected.

## Code root

```text
services/game-server/internal/game/
```

Primary support package:

```text
services/game-server/internal/game/physics/
```

## Responsibilities

Collision-shape support owns the game-server side of:

* Loading the shared collision-shape catalog.
* Finding `shared/collisions/collision_shapes.json` by walking upward from the current working directory.
* Representing imported collision shapes from JSON.
* Converting imported shape records into runtime `CollisionShape` values.
* Supporting circle, capsule, rectangle, and polygon shape types.
* Building bullet, ship, asteroid, and pickup shapes from the catalog.
* Scaling asteroid shapes by runtime asteroid size.
* Wrapping asteroid variant indexes when selecting asteroid shape entries.
* Falling back to the default ship shape for ship shape IDs that do not currently have keyed catalog support.
* Returning errors for missing pickup shape catalogs or unknown pickup shape keys.
* Providing primitive body point-containment checks for server-side target click validation.
* Providing body outline points for devtools collision telemetry and radial candidate radius derivation.
* Letting gameplay consumers skip collision behavior when a collision body cannot be built.

## Does not own

Collision-shape support does not own:

* Client-side collision shape export.
* The shared collision-shape data source or generation workflow.
* Data-sync workflow or generated-data validation.
* Which collision pairs are checked during simulation.
* Collision phase ordering.
* Damage request construction.
* Damage resolution.
* Player lives, death, respawn cooldown, or score consequences.
* Pickup collection rules or pickup effects.
* Target identity ownership.
* Radial effect timing, zones, or hit-intent generation.
* Devtools command routing.
* WebSocket packets or packet encoding.
* Client rendering, interpolation, or presentation hitboxes.

Those systems consume collision bodies but own their own runtime behavior.

## Domain roles

Collision-shape support participates in server-authoritative simulation by supplying the geometry used for:

```text
projectile -> asteroid collision checks
player -> asteroid collision checks
player -> pickup collection checks
server-side target click validation
safe player initial spawn and respawn placement
asteroid radial candidate radius calculation
devtools collision overlay telemetry
```

The client may render scene collision shapes, devtools overlays, and visual hit feedback, but server collision bodies remain the authority for gameplay outcomes.

## Shape catalog model

The imported catalog shape is:

```go
type CollisionShapeCatalog struct {
    Bullet    ImportedCollisionShape
    Ship      ImportedCollisionShape
    Asteroids []ImportedCollisionShape
    Pickups   map[string]ImportedCollisionShape
}
```

Imported shapes support these fields:

```go
type ImportedCollisionShape struct {
    Name   string
    Type   string
    Radius float64
    Height float64
    Size   []float64
    Points [][]float64
}
```

Supported imported `type` values are:

```text
circle
capsule
rectangle
polygon
```

Conversion rules:

* `circle` requires a positive `radius`.
* `capsule` requires positive `radius` and `height`.
* `rectangle` requires exactly two `size` values.
* `polygon` converts each two-number point into a `physics.Vector2`.
* Unsupported shape types return an error.
* Invalid shape data returns an error.

The current shared catalog contains:

```text
bullet   -> capsule
ship     -> polygon
asteroid -> polygon list
pickups  -> circle shapes keyed by pickup class
```

## Runtime body construction

Runtime entities expose collision bodies through focused methods.

Current body builders are:

```text
runtime.Ship.CollisionBody(...)
runtime.Bullet.CollisionBody(...)
runtime.Asteroid.CollisionBody(...)
pickups.Pickup.CollisionBody(...)
```

Ship bodies use:

```text
ship id
ship position
ship rotation
catalog.ShipShapeByID(ship.Stats.CollisionShapeID)
```

Bullet bodies use:

```text
bullet id
bullet position
bullet rotation
catalog.BulletShape()
```

Asteroid bodies use:

```text
asteroid id
asteroid position
catalog.AsteroidShape(asteroid.Variant, asteroid.Size)
```

Pickup bodies use:

```text
pickup id
pickup position
catalog.PickupShape(string(pickup.Class()))
```

Pickup shape lookup uses pickup class, not pickup type. For example, separate pickup types can share a class-level shape such as `powerup` or `weapon`.

If a shape lookup fails, the body builder returns `false`. Consumers treat that as no usable collision body and skip that entity for the current check.

## Runtime consumers

### Collision detection

`services/game-server/internal/game/collisions.go` builds collision bodies for current gameplay collision pairs:

```text
projectile -> asteroid
player -> asteroid
player -> pickup
```

For wrapped-world collision checks, the asteroid or pickup body is temporarily placed in actor-local wrapped space before primitive detection runs. This keeps authoritative entity storage wrapped while still allowing cross-boundary hits.

### Targeting

Server-side target selection uses collision bodies to validate click positions.

`SelectTargetAtPosition` checks that:

```text
requested target exists
target candidate has a collision body
click point is inside that body
```

The final containment check uses:

```go
physics.BodyContainsPoint(...)
```

This makes click-target selection server-validated instead of trusting only client-side presentation.

### Spawn and respawn safety

Initial spawn and respawn planning use ship, asteroid, and player collision shapes to avoid unsafe placements.

The respawn clearance path derives a bounding radius from each shape type:

```text
circle    -> radius
capsule   -> height * 0.5
rectangle -> half-size vector length
polygon   -> farthest point distance from origin
```

The clearance check then compares wrapped-space distance against both radii plus the configured respawn buffer.

If the current ship shape cannot be built, the respawn safety check treats the position as safe rather than blocking respawn indefinitely.

### Pickup collection

Player/pickup collection uses the same collision body model as damage-producing collision checks, but collection is not a damage flow.

A pickup with an unknown type, missing definition, missing class, missing pickup shape catalog, or unknown pickup class shape does not produce a usable collision body and cannot be collected through that check.

### Radial effects

Radial effect candidate generation uses collision-shape outlines to derive asteroid candidate radius when a direct radius is not available.

This keeps radial candidate coverage tied to imported asteroid geometry without making the radial package depend on the full game runtime or collision-shape catalog.

### Devtools collision telemetry

Devtools collision telemetry builds collision bodies for active players, asteroids, projectiles, and pickups.

The telemetry adapter converts each body into:

```text
kind
id
shape type
outline points
```

Outline points are generated by `physics.CollisionBodyOutlinePoints`.

Devtools uses these points for presentation only. Devtools does not own authoritative collision behavior.

## Data ownership

Collision-shape support reads generated shared data from:

```text
shared/collisions/collision_shapes.json
```

The game server stores the loaded catalog in memory per `Game` instance.

Collision-shape support does not persist data.

Collision-shape support mutates no durable account, profile, room, or player-data state.

The data/source-of-truth and export workflow belongs under data documentation, not this service implementation document.

## Protocol and API surfaces

Collision-shape support has no direct HTTP API, WebSocket packet, or protocol surface.

External clients observe collision-shape effects indirectly through normal game-server output after simulation systems consume collision bodies.

Examples:

```text
StatePacket players
StatePacket asteroids
StatePacket bullets
StatePacket pickups
StatePacket events
devtools collision telemetry output
```

Client requests can influence later collision outcomes through movement, shooting, targeting, respawn, pickup collection opportunities, or devtools actions, but no client packet directly calls collision-shape loading or primitive shape conversion.

## Failure behavior

`Game.New()` attempts to load the collision-shape catalog. If loading fails, the server logs a warning and still constructs the game with an empty catalog.

With an empty or incomplete catalog:

* Bullet, ship, asteroid, or pickup body construction can fail.
* Collision consumers skip entities whose bodies cannot be built.
* Missing collision bodies do not produce damage, pickup collection, target selection, or devtools body output.
* Respawn safety treats an unbuildable ship shape as safe rather than blocking all respawns.

Catalog conversion errors are local to the shape lookup that encounters invalid data.

## Invariants

Collision-shape support must preserve these rules:

* The server consumes generated shared collision data; it does not hand-author authoritative shape data.
* Primitive collision math stays separate from gameplay consequences.
* Runtime entities build collision bodies from the shared catalog rather than embedding duplicate shape constants.
* Collision bodies carry runtime identity and position, while the catalog carries reusable shape definitions.
* Missing or invalid collision bodies must not panic normal gameplay paths.
* Pickup shape lookup uses pickup class, not pickup type.
* Asteroid shape lookup scales imported shape geometry by runtime asteroid size.
* Collision shape lookup must not decide damage, scoring, death, pickup effects, or packet output.
* Client presentation must not be treated as authoritative collision geometry.
* Devtools outlines are presentation/telemetry output only.

## Active issues

* Ship shape lookup currently falls back to the default `v_wing` shape for unknown or empty ship shape IDs. A keyed multi-ship collision catalog is not implemented yet. See [Player Build Limits](../../../../limits/player-build-limits.md#ship-variants).

## Code map

Primary collision-shape and primitive support:

```text
services/game-server/internal/game/physics/collision_shapes.go
services/game-server/internal/game/physics/collision.go
services/game-server/internal/game/physics/collision_outline.go
services/game-server/internal/game/physics/vector.go
```

Game aggregate catalog storage:

```text
services/game-server/internal/game/game.go
```

Runtime entity body builders:

```text
services/game-server/internal/game/runtime/ship.go
services/game-server/internal/game/runtime/bullet.go
services/game-server/internal/game/runtime/asteroid.go
services/game-server/internal/game/entities/pickups/pickup.go
```

Gameplay consumers:

```text
services/game-server/internal/game/collisions.go
services/game-server/internal/game/combat.go
services/game-server/internal/game/pickup_collisions.go
services/game-server/internal/game/session.go
services/game-server/internal/game/targeting.go
services/game-server/internal/game/radial_candidates.go
```

Devtools consumer:

```text
services/game-server/internal/game/export_devtools_collision_telemetry.go
```

Shared/generated data:

```text
shared/collisions/collision_shapes.json
client/tools/export_collision_shapes.gd
```

Related tests:

```text
services/game-server/internal/game/physics/collision_shapes_test.go
services/game-server/internal/game/physics/collision_test.go
services/game-server/internal/game/physics/collision_outline_test.go
services/game-server/tests/physics/collision_shapes_test.go
services/game-server/tests/physics/collision_test.go
services/game-server/tests/game/collision_test.go
services/game-server/tests/game/respawn_test.go
services/game-server/tests/game/pickups_test.go
services/game-server/tests/game/devtools_test.go
services/game-server/tests/game/ship_collision_shape_test.go
services/game-server/internal/game/entities/pickups/pickup_test.go
services/game-server/internal/game/export_devtools_collision_telemetry_test.go
```

Important non-ownership boundaries:

```text
shared/collisions/
client/tools/export_collision_shapes.gd
services/game-server/internal/game/damage/
services/game-server/internal/game/pickups/
services/game-server/internal/game/effects/radial/
services/game-server/internal/devtools/
services/game-server/internal/networking/
client/
```

## Tests and verification

Focused physics and collision-shape tests cover:

* collision shape catalog loading
* imported asteroid shape scaling
* default ship shape fallback behavior
* pickup shape lookup
* missing pickup shape errors
* point containment for circle, capsule, rectangle, and polygon bodies
* capsule/polygon primitive collision
* concave polygon miss behavior
* outline point generation for rotated polygons and capsules

Game integration tests cover collision-shape consumption through:

* projectile/asteroid collision behavior
* player/asteroid collision behavior
* wrapped-world collision behavior
* safe spawn and respawn placement
* pickup collection checks
* devtools collision telemetry
* ship collision shape fallback behavior

Useful verification command:

```bash
cd services/game-server
go test -buildvcs=false ./...
```

Focused physics verification:

```bash
cd services/game-server
go test -buildvcs=false ./internal/game/physics ./tests/physics
```

Focused gameplay consumer verification:

```bash
cd services/game-server
go test -buildvcs=false ./tests/game -run 'Collision|Respawn|Pickup|Devtools|ShipCollisionShape'
```

## Related docs

* [Game Server Simulation World](./!README.md)
* [Game Server Simulation](../!README.md)
* [Game Server](../../!README.md)
* [Collision To Damage Flow](../combat/collision-to-damage-flow.md)
* [Pickup Collection](../pickups/pickup-collection.md)
* [Pickup Entity Lifecycle](../pickups/pickup-entity-lifecycle.md)
* [Target Selection And Status](../targeting/target-selection-and-status.md)
* [Active Player Avatar State](../players/active-player-avatar-state.md)
* [Player Respawn](../players/player-respawn.md)
* [Radial Effects](../combat/radial-effects.md)
* [Collision Shape Data](../../../../data/stubs/collision-shape-data.md)
* [Player Build Limits](../../../../limits/player-build-limits.md)

## Notes

The legacy architecture documentation correctly identified the intended split: `physics` owns primitive collision and shape support, `collisions.go` owns game-pair collision facts, and combat/pickup/targeting/respawn/devtools systems consume those facts or bodies for their own purposes.

`ImportedCollisionShape.Name` is retained from exported data, but runtime lookup is currently keyed by catalog location or pickup class rather than by the exported node name.

The shared collision-shape JSON currently includes pickup shape entries with an `offset` field, but the server import struct does not consume offset. Current server collision bodies use runtime entity position plus imported shape geometry.
