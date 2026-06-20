# Toroidal Space And Motion

Parent index: [Game Server Simulation World](./!README.md)

## Purpose

This document describes the game-server toroidal space and motion boundary.

It explains how the server keeps authoritative positions inside wrapped world bounds, how ship, asteroid, and projectile motion are advanced, and how wrapped spatial helpers support related world systems such as spawning, visibility, collisions, radial effects, and respawn safety.

## Overview

The game server stores authoritative world positions as bounded toroidal coordinates.

A position that moves beyond one edge of the world wraps to the opposite edge. The server does not create ghost entities or duplicate bodies at world seams. Instead, it keeps one authoritative position per entity and uses wrapped spatial helpers when systems need distance, direction, or local collision placement across an edge.

The current world size comes from shared constants:

```text
shared/constants/server_constants.toml
```

Generated Go constants provide the runtime values:

```text
constants.WorldWidth  = 17200.0
constants.WorldHeight = 9200.0
```

The simulation loop chooses active bounds once per tick:

```text
Game.Step(delta)
-> bounds := space.DefaultBounds()
```

Those bounds are passed into the motion helpers for active players, asteroids, and projectiles.

The main implementation split is:

```text
space package
= world bounds, wrapping, shortest wrapped deltas, distance, direction, normalization

motion package
= per-entity movement integration plus final position wrapping

game package
= simulation phase order, entity map iteration, movement gates, cleanup, spawning, collisions, scoring, and lifecycle consequences
```

## Code root

```text
services/game-server/internal/game/
```

Primary supporting packages:

```text
services/game-server/internal/game/space/
services/game-server/internal/game/motion/
services/game-server/internal/game/runtime/
services/game-server/internal/game/physics/
services/game-server/internal/constants/
```

## Responsibilities

Toroidal space and motion owns the game-server side of:

* Reading generated world bounds through `space.DefaultBounds`.
* Wrapping positions into current bounds.
* Computing shortest wrapped deltas between bounded positions.
* Computing wrapped distance and direction helpers.
* Advancing player ship motion from stored input and ship stats.
* Gating player movement through the player move policy supplied by the game package.
* Advancing asteroid positions from velocity.
* Advancing projectile positions from velocity.
* Decrementing projectile lifetime during projectile movement.
* Decrementing pending-despawn delay for ships, asteroids, and projectiles.
* Normalizing moved ship, asteroid, and projectile positions after advancement.
* Supporting wrapped spatial checks for visibility, despawn, collision placement, radial effect coverage, asteroid spawn aiming, and respawn safety.
* Keeping reusable motion logic outside the root game aggregate while leaving entity ownership in the game package.

## Does not own

Toroidal space and motion does not own:

* The simulation tick lifecycle or phase order.
* Runtime entity map ownership.
* Player session ownership.
* Input packet routing.
* Player pause or suspension state.
* Weapon fire policy.
* Projectile spawning policy.
* Asteroid spawning policy.
* Asteroid visibility or despawn policy.
* Collision primitive math.
* Collision consequence handling.
* Damage, death, scoring, drops, or pickup effects.
* State packet projection.
* Client continuous visual coordinates.
* Camera, background, interpolation, or render-anchor behavior.
* Shared constant generation.
* Data-sync commands or pipeline configuration.
* Room lifecycle or WebSocket transport.

Those systems may use wrapped positions or motion results, but they own their own boundaries.

## Domain roles

Toroidal space and motion participates in the technical runtime domain by enforcing a consistent authoritative coordinate model for the game server.

The server role is:

```text
store one bounded authoritative position per entity
advance authoritative motion during simulation
wrap moved positions into server world bounds
use shortest wrapped spatial math for cross-edge relationships
publish bounded positions through state packets
```

The client role is separate:

```text
receive bounded server positions
convert them into continuous visual positions
render motion across edges without visible snapping
```

The client presentation model is documented separately. This service doc owns only the server-side world and motion implementation.

## World bounds model

The `space.Bounds` type carries the active width and height:

```go
type Bounds struct {
    Width  float64
    Height float64
}
```

`space.DefaultBounds()` returns bounds from generated constants:

```go
space.Bounds{
    Width:  constants.WorldWidth,
    Height: constants.WorldHeight,
}
```

The current source values live in:

```text
shared/constants/server_constants.toml
```

under:

```text
[constants.shared.world]
world_width = 17200.0
world_height = 9200.0
```

The generated Go output lives in:

```text
services/game-server/internal/constants/constants.go
```

The server uses these constants for authoritative wrapping. Client generated constants must stay aligned so client visual wrapping uses the same world size, but the client owns that presentation behavior.

## Wrapping and shortest-delta helpers

`space.WrapPosition(position, bounds)` wraps each coordinate independently.

A coordinate outside the range is normalized with modulo behavior:

```text
x < 0      -> wraps near right edge
x >= width -> wraps near left edge

y < 0       -> wraps near bottom edge
y >= height -> wraps near top edge
```

Positions more than one world size outside the bounds still wrap correctly.

`space.ShortestDelta(from, to, bounds)` returns the shortest vector from one bounded position to another, accounting for edge wrap. This prevents systems from treating two objects near opposite edges as far apart when the toroidal path between them is short.

The convenience helpers use default bounds:

```text
space.Delta(from, to)
space.Distance(from, to)
space.Direction(from, to)
space.NormalizePosition(position)
```

These helpers are used by server systems that need toroidal spatial relationships without owning the wrap implementation.

## Motion integration

The motion package owns entity-local stepping helpers and advance-with-wrap helpers.

The split is:

```text
Step*
= mutate velocity, timers, and position for one entity

Advance*
= Step* plus position normalization into bounds
```

Current advance helpers are:

```text
motion.AdvanceShip
motion.AdvanceShipWithMovePolicy
motion.AdvanceAsteroid
motion.AdvanceBullet
```

The game package calls the advance helpers while iterating its own entity maps. The motion package does not own entity storage and does not delete, spawn, score, damage, or write packets.

### Ship motion

Player ship motion uses:

```text
motion.AdvanceShipWithMovePolicy(player, delta, bounds, canMove)
```

The game package supplies `canMove` from:

```text
game.playerCanMove(player.ID, player)
```

When the ship is pending despawn, motion only decrements `DespawnDelay` and returns.

When `canMove` is false, `StepShipWithMovePolicy` clears ship input and returns without applying rotation, thrust, damping, velocity movement, or invulnerability countdown. The advance helper still normalizes the ship position afterward.

When movement is allowed, ship stepping:

```text
decrements InvulnerabilityRemaining
reads left/right input as rotation axis
reads back/forward input as thrust axis
applies rotation speed
applies thrust along ship facing
applies damping scaled to 60 Hz baseline
limits velocity to MaxSpeed
integrates X/Y from velocity and delta
wraps final position into bounds
```

Ship movement stats come from `runtime.ShipStats`, which are resolved from generated constants and ship-type modifiers.

### Asteroid motion

Asteroid motion uses:

```text
motion.AdvanceAsteroid(asteroid, delta, bounds)
```

When the asteroid is pending despawn, motion only decrements `DespawnDelay` and returns.

Otherwise, asteroid stepping:

```text
adds Velocity.X * delta to X
adds Velocity.Y * delta to Y
wraps final position into bounds
```

The game package decides whether asteroid motion runs by checking:

```text
worldSimulationOptions.AsteroidsCanMove()
```

Asteroid cleanup and far-from-camera despawn checks remain game-owned.

### Projectile motion

Projectile motion uses:

```text
motion.AdvanceBullet(bullet, delta, bounds)
```

When the projectile is pending despawn, motion only decrements `DespawnDelay` and returns.

Otherwise, projectile stepping:

```text
adds Velocity.X * delta to X
adds Velocity.Y * delta to Y
decrements Life by delta
wraps final position into bounds
```

The game package decides whether projectile motion runs by checking:

```text
worldSimulationOptions.BulletsCanMove()
```

When bullets are frozen, projectile lifetime does not decrement through the motion helper. Projectile cleanup checks still run in the projectile phase, but lifetime-based expiration depends on `Life` having reached zero.

## Simulation phase participation

`Game.Step(delta)` chooses bounds before running the normal active-match phases:

```text
bounds := space.DefaultBounds()
```

Current normal motion-related phase order is:

```text
stepPlayerWeapons(delta)
stepPlayers(delta, bounds)
removeReadyPlayers()
stepAsteroidSpawning(delta)
stepAsteroids(delta, bounds)
stepBullets(delta, bounds)
stepPickups(delta)
stepCollisions()
stepRadialEffects(delta)
```

Player movement runs before player fire checks in `stepPlayers`.

Asteroid spawn planning happens before asteroid movement in the same tick.

Projectile movement runs before collision detection.

Pickup aging does not use the motion package because pickups are currently stationary runtime entities.

The reduced match-over path still advances asteroids, projectiles, pickups, and radial effects, but skips player movement, weapon stepping, new asteroid spawning, and collisions.

## World simulation gates

World freeze behavior is owned by `WorldSimulationOptions`, not by the motion package.

The current gates are:

```text
AsteroidsCanMove()
BulletsCanMove()
CanSpawnAsteroids()
CanRunCollisions()
```

Motion is affected as follows:

```text
FreezeAsteroids
-> skips asteroid motion
-> asteroid removal/despawn checks still run

FreezeBullets
-> skips projectile motion
-> projectile lifetime does not decrement through StepBullet
-> projectile removal/despawn checks still run
-> player fire checks do not spawn projectiles

FreezeSpawning
-> skips timed asteroid spawning

FreezeCollisions
-> skips ship/asteroid, projectile/asteroid, and player/pickup collision families
```

Player pause and dev player-freeze are separate from world freeze. They are player-session suspension concerns. Player movement receives the result as `canMove`; the motion package does not own why movement is blocked.

## Wrapped spatial consumers

Several server systems rely on the same wrapped spatial model.

Asteroid timed spawn planning uses wrapped direction so an asteroid spawned near a world edge can aim toward a camera or player through the shortest toroidal path.

Visibility and despawn checks use wrapped deltas from camera views so entities near opposite world edges can still be considered near the camera.

Collision detection keeps one authoritative body per entity. For cross-edge collision checks, the server computes a wrapped delta and temporarily places the other collision body near the actor being tested before calling physics collision detection.

Current wrapped collision families include:

```text
projectile -> asteroid
player -> asteroid
player -> pickup
```

Radial effect stepping uses wrapped deltas from the effect origin to candidate positions.

Respawn safety uses wrapped distance so threats across an edge are still considered nearby.

## Data ownership

Toroidal space and motion owns no durable persistence.

It reads:

```text
constants.WorldWidth
constants.WorldHeight
runtime.ShipStats
runtime.Ship.Input
runtime entity positions
runtime entity velocities
pending-despawn flags and delays
projectile Life
```

It mutates:

```text
runtime.Ship.X
runtime.Ship.Y
runtime.Ship.Rotation
runtime.Ship.Velocity
runtime.Ship.Input when movement is blocked
runtime.Ship.InvulnerabilityRemaining
runtime.Ship.DespawnDelay

runtime.Asteroid.X
runtime.Asteroid.Y
runtime.Asteroid.DespawnDelay

runtime.Bullet.X
runtime.Bullet.Y
runtime.Bullet.Life
runtime.Bullet.DespawnDelay
```

It does not mutate:

```text
game entity maps
player sessions
camera view ownership
score
lives
health
damage results
spawn counters
pending presentation events
packet state
external storage
```

## Protocols and APIs

Toroidal space and motion exposes no public network API.

Its main service-internal surfaces are Go package functions:

```text
space.DefaultBounds()
space.WrapPosition(position, bounds)
space.ShortestDelta(from, to, bounds)
space.WrappedDistance(from, to, bounds)
space.Delta(from, to)
space.Distance(from, to)
space.Direction(from, to)
space.NormalizePosition(position)

motion.StepShip(ship, delta)
motion.StepShipWithMovePolicy(ship, delta, canMove)
motion.AdvanceShip(ship, delta, bounds)
motion.AdvanceShipWithMovePolicy(ship, delta, bounds, canMove)
motion.StepAsteroid(asteroid, delta)
motion.AdvanceAsteroid(asteroid, delta, bounds)
motion.StepBullet(bullet, delta)
motion.AdvanceBullet(bullet, delta, bounds)
```

These surfaces are for server simulation code. They are consumed by the game package and supporting game-server packages.

Clients observe the results indirectly through gameplay state packets. The packet surface carries bounded authoritative positions; it does not expose motion helper calls, internal bounds objects, or server-side shortest-delta calculations.

## Invariants

Toroidal space and motion should preserve these rules:

* Server gameplay positions are bounded authoritative coordinates.
* Moving ships, asteroids, and projectiles are normalized after movement.
* The game package owns entity map iteration and lifecycle consequences.
* The motion package does not spawn, delete, damage, score, emit events, or write packets.
* Wrapped distance and direction use the shortest toroidal path.
* Collision systems should use wrapped-local placement instead of duplicate ghost entities.
* Client visual continuity does not change server authority.
* World freeze and player suspension are different gates.
* Player movement policy is supplied to motion; motion does not inspect sessions.
* Projectile lifetime decrements only when projectile stepping runs.
* Pending-despawn entities decrement despawn delay instead of applying normal movement.

## Code map

Primary implementation files:

```text
services/game-server/internal/game/space/space.go
services/game-server/internal/game/motion/motion.go
services/game-server/internal/game/simulation.go
services/game-server/internal/game/simulation_players.go
services/game-server/internal/game/simulation_asteroids.go
services/game-server/internal/game/simulation_bullets.go
services/game-server/internal/game/world_simulation_options.go
```

Runtime entity files:

```text
services/game-server/internal/game/runtime/state.go
services/game-server/internal/game/runtime/ship.go
services/game-server/internal/game/runtime/asteroid.go
services/game-server/internal/game/runtime/bullet.go
services/game-server/internal/game/runtime/ship_stats.go
```

Wrapped spatial consumers:

```text
services/game-server/internal/game/spawning.go
services/game-server/internal/game/spawning/spawner.go
services/game-server/internal/game/visibility.go
services/game-server/internal/game/collisions.go
services/game-server/internal/game/session.go
services/game-server/internal/game/effects/radial/step.go
```

Generated and source constants:

```text
shared/constants/server_constants.toml
services/game-server/internal/constants/constants.go
client/scripts/generated/constants/constants.gd
```

Related client presentation files:

```text
client/scripts/world/world_sync.gd
client/scripts/world/world_wrap.gd
client/scripts/world/player_render/player_render_api.gd
```

Important non-ownership boundaries:

```text
services/game-server/internal/game/physics/
services/game-server/internal/game/weapons/
services/game-server/internal/game/damage/
services/game-server/internal/game/scoring/
services/game-server/internal/game/pickups/
services/game-server/internal/game/effects/radial/
services/game-server/internal/rooms/
services/game-server/internal/networking/
client/
tools/data_sync/
```

## Tests and verification

Relevant focused tests include:

```text
services/game-server/tests/space/space_test.go
services/game-server/tests/game/movement_test.go
services/game-server/tests/game/visibility_test.go
services/game-server/tests/game/collision_test.go
services/game-server/tests/game/respawn_test.go
services/game-server/tests/game/pause_test.go
services/game-server/internal/game/world_simulation_options_test.go
```

Current tested behavior includes:

* Flat-world delta, distance, and direction behavior.
* Wrapped direction across world edges.
* Position wrapping at right, left, top, and bottom edges.
* Position wrapping for values more than one world size out of bounds.
* Shortest wrapped delta across horizontal and vertical edges.
* Wrapped distance through the shortest path.
* Default-bound position normalization.
* Player movement wrapping across left and right edges.
* Asteroid movement wrapping across edges.
* Projectile movement wrapping across edges.
* World simulation freeze gates.
* Player pause/suspension blocking movement without owning motion internals.
* Cross-edge collision behavior through wrapped spatial checks.
* Visibility and despawn behavior near wrapped world edges.
* Respawn safety using wrapped distance.

Useful verification command:

```bash
cd services/game-server
go test -buildvcs=false ./...
```

Focused verification for this boundary:

```bash
cd services/game-server
go test -buildvcs=false ./tests/space ./tests/game -run 'Wrap|Delta|Distance|Direction|Movement|Visibility|Collision|Respawn|Pause'
```

## Related docs

* [Game Server Simulation World](./!README.md)
* [Game Server Simulation](../!README.md)
* [Game Server](../../!README.md)
* [Simulation Loop And Phase Order](../runtime/simulation-loop-and-phase-order.md)
* [Runtime Entity Store](../runtime/runtime-entity-store.md)
* [State Packet Projection](../runtime/state-packet-projection.md)
* [Player Pause And Suspension](../players/player-pause-and-suspension.md)
* [Player Camera View State](../players/player-camera-view-state.md)
* [Player Respawn](../players/player-respawn.md)
* [Collision To Damage Flow](../combat/collision-to-damage-flow.md)
* [Radial Effects](../combat/radial-effects.md)
* [Pickup Entity Lifecycle](../pickups/pickup-entity-lifecycle.md)
* [View Anchor And Visual Coordinates](../../../client/world-sync/view-anchor-and-visual-coordinates.md)
* [Toroidal Wrap](../../../../systems-design/world/stubs/toroidal-wrap.md)
* [World Authority](../../../../systems-design/world/stubs/world-authority.md)
* [Constants Pipeline](../../../../data/stubs/constants-pipeline.md)

## Notes

Legacy toroidal-wrap documentation correctly identified the key server/client split: the server stores bounded authoritative coordinates, while the client renders continuous visual coordinates. This document narrows that material to the current game-server service implementation.

The server currently has one authoritative position per entity. Cross-edge collision and distance behavior should continue to use wrapped deltas rather than adding duplicate ghost entities to runtime storage.

`space.WrapPosition` returns the input coordinate unchanged when a bound size is non-positive. Current production bounds come from positive generated constants, so that fallback is defensive behavior rather than normal gameplay configuration.
