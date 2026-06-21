## Physics

Parent index: [Game Server Simulation World](./!INDEX.md)

## Purpose

This document describes the game-server physics support boundary.

It covers server-side vector helpers, collision primitive definitions, collision body overlap checks, point containment checks, collision outline projection, and runtime collision-shape catalog access.

## Overview

The game-server physics boundary is a support package, not a full physics engine.

It provides reusable math and collision primitives used by authoritative gameplay systems. It does not own entity movement, world wrapping, collision phase order, damage, scoring, pickup effects, or client presentation.

The current runtime relationship is:

```text
shared collision shape JSON
-> physics.LoadCollisionShapeCatalog
-> Game.collisionShapes
-> runtime entity CollisionBody methods
-> game-owned collision, targeting, radial, respawn, pickup, and devtools paths
-> physics primitive checks
-> game-owned consequences
```

The important ownership split is:

```text
physics package
= primitive math, shape conversion, body overlap, point containment, outline points

motion package
= player, asteroid, and projectile movement integration

space package
= toroidal world wrapping and shortest wrapped deltas

game/collisions.go
= current game-pair collision fact detection

combat and pickup paths
= runtime consequences from collision facts
```

Physics answers narrow questions such as:

```text
Do these two collision bodies overlap?
Does this server-space point fall inside this collision body?
What outline points represent this body for telemetry or radius approximation?
Can this imported collision shape become a runtime primitive?
```

It does not decide what an overlap means.

## Code root

```text
services/game-server/internal/game/physics/
```

Primary consumers live under:

```text
services/game-server/internal/game/
services/game-server/internal/game/runtime/
services/game-server/internal/game/entities/pickups/
services/game-server/internal/devtools/
```

## Responsibilities

The physics boundary owns:

* `Vector2` math helpers used by simulation support code.
* Collision shape type definitions for circles, capsules, rectangles, and polygons.
* Runtime collision body shape, position, rotation, and ID data.
* Primitive overlap checks through `DetectCollision`.
* Point containment checks through `BodyContainsPoint`.
* Collision outline point projection through `CollisionBodyOutlinePoints`.
* Imported collision shape conversion into runtime collision primitives.
* Runtime loading of `shared/collisions/collision_shapes.json`.
* Runtime catalog access for bullet, ship, asteroid, and pickup shapes.
* Asteroid collision-shape scaling by asteroid size.
* Asteroid collision-shape index wrapping when the runtime asteroid variant index exceeds the loaded collision-shape list.
* Current single-ship collision-shape fallback through `ShipShapeByID`.
* Error reporting when imported shapes are missing, malformed, or unsupported.

## Does not own

The physics boundary does not own:

* Simulation tick phase order.
* Entity movement integration.
* World bounds or toroidal coordinate wrapping.
* Stored player, projectile, asteroid, enemy, or pickup maps.
* Collision family routing.
* Combat damage request construction.
* Damage resolution.
* Scoring policy.
* Player lives, death, despawn, or respawn lifecycle.
* Pickup collection or pickup effect rules.
* Radial effect timing, coverage, or hit-intent rules.
* Collision-shape source export from Godot scenes.
* Packet schema or packet projection.
* WebSocket routing.
* Client-side rendering, interpolation, input, or hitbox display.
* Devtools command routing.

Those systems may use physics primitives, but they own their own runtime decisions and consequences.

## Domain roles

Physics participates in the server-authoritative gameplay domain as reusable simulation support.

It helps enforce server authority for:

* projectile/asteroid overlap checks
* player/asteroid overlap checks
* player/pickup overlap checks
* server-side point validation for target selection
* respawn clearance checks through game-owned radius approximation
* radial candidate radius derivation for asteroid bodies
* debug collision-body telemetry

Clients may render local visuals and debug overlays, but authoritative collision and point-selection validation use server-side collision bodies.

## Protocols and APIs

Physics exposes no HTTP or WebSocket API.

Its public surface is internal Go package API consumed by other game-server packages.

Main shape and body surfaces:

```go
type CollisionShapeType string
type CollisionShape struct
type CollisionBody struct
type Collision struct

func NewCircleShape(radius float64) CollisionShape
func NewCapsuleShape(radius float64, height float64) CollisionShape
func NewRectangleShape(width float64, height float64) CollisionShape
func NewPolygonShape(points []Vector2) CollisionShape
```

Main primitive checks:

```go
func DetectCollision(a CollisionBody, b CollisionBody) (Collision, bool)
func BodyContainsPoint(body CollisionBody, point Vector2) bool
func CollisionBodyOutlinePoints(body CollisionBody) []Vector2
```

Main catalog surfaces:

```go
func LoadCollisionShapeCatalog() (CollisionShapeCatalog, error)

func (catalog CollisionShapeCatalog) BulletShape() (CollisionShape, error)
func (catalog CollisionShapeCatalog) ShipShape() (CollisionShape, error)
func (catalog CollisionShapeCatalog) ShipShapeByID(shapeID string) (CollisionShape, error)
func (catalog CollisionShapeCatalog) AsteroidShape(variant int, size int) (CollisionShape, error)
func (catalog CollisionShapeCatalog) PickupShape(pickupType string) (CollisionShape, error)
func (shape ImportedCollisionShape) ToCollisionShape(scale float64) (CollisionShape, error)
```

The surface is used inside the process only. Callers build or retrieve `CollisionBody` values, pass them to physics helpers, and then apply any gameplay result in the owning gameplay boundary.

## Runtime model

A collision body is a value object:

```go
type CollisionBody struct {
    ID       string
    Position Vector2
    Rotation float64
    Shape    CollisionShape
}
```

The body does not point back into the entity store. Physics helpers operate on the body value they receive.

Runtime entities expose collision body helpers:

```text
runtime.Ship.CollisionBody
runtime.Bullet.CollisionBody
runtime.Asteroid.CollisionBody
pickups.Pickup.CollisionBody
```

Those helpers translate authoritative runtime state into physics bodies by combining:

```text
entity id
entity position
entity rotation when relevant
shape from CollisionShapeCatalog
```

If an entity cannot build a body, its helper returns `false`. Collision, targeting, devtools, and pickup paths skip unavailable bodies instead of inventing fallback geometry at the call site.

## Collision primitive behavior

`DetectCollision` supports overlap checks between:

```text
circle    <-> circle
capsule   <-> capsule
circle    <-> capsule
polygon   <-> polygon
circle    <-> polygon
capsule   <-> polygon
rectangle <-> supported polygon combinations through polygon conversion
```

Rectangles are treated as polygons internally.

For polygon bodies, local shape points are rotated by the body rotation and translated by the body position before overlap checks.

For capsules, the capsule segment is derived from body position, body rotation, shape height, and shape radius.

`DetectCollision` returns a `Collision` with both input bodies and a contact point. The current contact point is the midpoint between body positions:

```text
(a.Position + b.Position) * 0.5
```

It is not a true contact manifold, not a swept-impact point, and not an impulse-resolution surface.

`BodyContainsPoint` supports point checks for circle, capsule, rectangle-as-polygon, and polygon bodies. Server-side target selection uses this helper to validate that the client-requested click point is inside the authoritative collision body for the requested target.

`CollisionBodyOutlinePoints` converts collision bodies into outline point lists. Current consumers use those points for devtools collision telemetry, devtools shape catalog output, and asteroid radial radius approximation.

## Shape catalog behavior

`LoadCollisionShapeCatalog` searches upward from the process working directory until it finds:

```text
shared/collisions/collision_shapes.json
```

`game.New()` loads this catalog and stores it on the game aggregate as:

```text
Game.collisionShapes
```

If loading fails, game construction still succeeds and logs a warning. Collision-dependent call sites then fail body construction or skip unavailable collision bodies.

The catalog shape is:

```go
type CollisionShapeCatalog struct {
    Bullet    ImportedCollisionShape            `json:"bullet"`
    Ship      ImportedCollisionShape            `json:"ship"`
    Asteroids []ImportedCollisionShape          `json:"asteroids"`
    Pickups   map[string]ImportedCollisionShape `json:"pickups"`
}
```

Current lookup behavior:

```text
BulletShape
-> one imported bullet shape

ShipShape
-> one imported ship shape

ShipShapeByID
-> default ship shape for empty, default, or unknown ship shape IDs

AsteroidShape
-> wrapped asteroid collision-shape index
-> scaled by size * constants.AsteroidSizeScale

PickupShape
-> imported pickup shape by pickup class key
```

The `PickupShape` parameter is named like a pickup type in code, but current pickup collision lookup passes the pickup class string. This keeps classes such as `powerup` and `weapon` mapped to generic pickup collision bodies rather than per-pickup-type bodies.

## Wrapped-world behavior

Physics does not know about toroidal space.

Wrapped-world collision support is handled by callers. Current projectile/asteroid and player/asteroid collision checks place the asteroid body in wrapped-local space before calling `physics.DetectCollision`:

```text
delta = space.Delta(actor.Position(), asteroid.Position())
asteroidBody.Position = actor.Position().Add(delta)
```

This keeps primitive physics simple while allowing cross-boundary collision checks in the toroidal world.

Position normalization and shortest wrapped delta ownership belongs to `services/game-server/internal/game/space/`.

## Data ownership

Physics owns no durable persistence.

It reads:

```text
shared/collisions/collision_shapes.json
```

It converts imported collision shape data into runtime `CollisionShape` values.

It does not own the Godot export tool, source scene collision nodes, data-sync behavior, or documentation of the collision-shape data pipeline. Those belong to data documentation and client tooling.

Physics mutates no game state. Callers decide how to use the returned collision, containment, outline, or catalog result.

Current imported fields consumed by Go shape conversion include:

```text
name
type
radius
height
size
points
```

Additional JSON fields are ignored unless the Go import shape and conversion path are extended.

## Active issues

* Multi-ship collision shape catalog support is not implemented. `ShipShapeByID` currently falls back to the single imported ship shape for unknown IDs. See [Ship Variants](../../../../limits/player-build-limits.md#ship-variants).

## Invariants

The physics boundary must preserve these rules:

* Physics helpers must not mutate game aggregate state.
* Primitive overlap math must stay separate from gameplay consequences.
* Collision bodies are value objects built from authoritative runtime state.
* Missing collision bodies must not produce false-positive gameplay hits.
* Shape catalog loading failures must not prevent game aggregate construction.
* Collision-shape source data must remain outside the physics runtime package.
* Toroidal wrapping must stay outside primitive collision math.
* Client presentation must not become authoritative for collision outcomes.
* Devtools collision telemetry must use the same body construction path as gameplay support, not a parallel debug-only geometry model.

## Code map

Primary implementation files:

```text
services/game-server/internal/game/physics/vector.go
services/game-server/internal/game/physics/collision.go
services/game-server/internal/game/physics/collision_shapes.go
services/game-server/internal/game/physics/collision_outline.go
```

Game aggregate integration:

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

Collision and gameplay consumers:

```text
services/game-server/internal/game/collisions.go
services/game-server/internal/game/combat.go
services/game-server/internal/game/pickup_collisions.go
services/game-server/internal/game/targeting.go
services/game-server/internal/game/radial_candidates.go
services/game-server/internal/game/session.go
```

Devtools and telemetry consumers:

```text
services/game-server/internal/game/export_devtools_collision_telemetry.go
services/game-server/internal/devtools/shape_catalog.go
```

Source data:

```text
shared/collisions/collision_shapes.json
client/tools/export_collision_shapes.gd
```

Important non-ownership boundaries:

```text
services/game-server/internal/game/motion/
services/game-server/internal/game/space/
services/game-server/internal/game/damage/
services/game-server/internal/game/pickups/
services/game-server/internal/game/effects/radial/
services/game-server/internal/game/scoring/
services/game-server/internal/networking/
client/
docs/data/
```

## Tests and verification

Relevant package tests:

```text
services/game-server/internal/game/physics/collision_test.go
services/game-server/internal/game/physics/collision_outline_test.go
services/game-server/internal/game/physics/collision_shapes_test.go
services/game-server/tests/physics/collision_test.go
services/game-server/tests/physics/collision_shapes_test.go
```

Relevant gameplay integration tests:

```text
services/game-server/tests/game/collision_test.go
services/game-server/tests/game/pickups_test.go
services/game-server/tests/game/respawn_test.go
services/game-server/tests/game/ship_collision_shape_test.go
services/game-server/internal/game/targeting_test.go
services/game-server/internal/game/radial_effects_test.go
```

Current coverage includes:

* point containment for circle, capsule, rectangle, and polygon bodies
* capsule/polygon collision checks
* concave polygon miss behavior
* collision outline point projection for rotated polygons and capsules
* collision-shape catalog loading
* asteroid shape scaling by size
* pickup shape lookup and missing-shape errors
* default and unknown ship collision shape ID fallback
* projectile/asteroid and ship/asteroid collision outcomes
* player/pickup collection through collision body checks
* target selection overlap validation through server collision bodies
* respawn safety behavior using resolved ship collision shapes

Useful verification command:

```bash
cd services/game-server
go test -buildvcs=false ./...
```

Focused verification for this boundary:

```bash
cd services/game-server
go test -buildvcs=false ./internal/game/physics ./tests/physics
```

Focused verification for gameplay consumers:

```bash
cd services/game-server
go test -buildvcs=false ./internal/game ./tests/game -run 'Collision|Pickup|Respawn|Target|ShipCollisionShape'
```

Run collision-shape export or data checks when source scene collision nodes or shared collision shape output changes.

## Related docs

* [Game Server Simulation World](./!INDEX.md)
* [Game Server Simulation](../!INDEX.md)
* [Game Server](../../!INDEX.md)
* [Game Aggregate](../runtime/game-aggregate.md)
* [Simulation Loop And Phase Order](../runtime/simulation-loop-and-phase-order.md)
* [Collision To Damage Flow](../combat/collision-to-damage-flow.md)
* [Pickup Collection](../pickups/pickup-collection.md)
* [Canonical Target State](../targeting/canonical-target-state.md)
* [Player Respawn](../players/player-respawn.md)
* [Collision Shapes](collision-shapes.md)
* [Toroidal Space And Motion](toroidal-space-and-motion.md)
* [Collision Shape Data](../../../../data/collision-shape-data.md)
* [Devtools](../../../../devtools/!INDEX.md)
* [Player Build Limits](../../../../limits/player-build-limits.md)

## Notes

Legacy architecture notes correctly identified the durable split between primitive physics math, game-owned collision facts, and combat consequences. This document narrows that material to the current game-server implementation.

The physics package name should not be read as ownership of a full rigid-body simulation. Current movement is simple per-entity integration in the motion package, collision checks are discrete overlap checks, and gameplay consequences are handled by game-owned adapters after primitive checks return.
