# Collision To Damage Flow

Parent index: [Game Server Simulation Combat](./!README.md)

## Purpose

This document describes the game-server collision-to-damage flow.

It explains how authoritative collision facts are detected, converted into damage requests, resolved through the damage seam, and then applied back into gameplay consequences such as health changes, despawn, scoring, fragment spawning, pickup drops, radial impact effects, and player death.

## Overview

Collision-to-damage flow is a game-server simulation responsibility owned by `services/game-server/internal/game`.

The current runtime flow is:

```text
simulation tick
-> collision phase gate
-> collision fact detection
-> damage request construction
-> damage.ResolveSingle
-> runtime entity damage application
-> collision consequences
-> domain event recording
-> later state-packet projection
```

The collision phase is authoritative. The client may render ships, bullets, asteroids, hit effects, and UI feedback, but it does not decide whether a projectile hit, whether a player died, whether score was awarded, or whether an asteroid split.

Current damage-producing collision pairs are:

```text
projectile -> asteroid
asteroid -> player
```

Player/pickup collision also runs in the collision phase, but it is a collection/effect flow rather than a damage flow.

Collision detection and damage resolution are separate seams:

```text
physics package
= primitive overlap math and collision shape support

collisions.go
= game-pair collision fact detection

combat.go
= collision-to-damage orchestration and consequences

damage package
= pure damage result calculation
```

The damage package does not inspect world maps, collision shapes, runtime stores, scoring, sessions, packets, or client state. Game-owned combat code supplies a `DamageResolutionRequest`, receives a `DamageResult`, and applies the result.

## Code root

```text
services/game-server/internal/game/
```

Primary supporting packages:

```text
services/game-server/internal/game/physics/
services/game-server/internal/game/damage/
services/game-server/internal/game/runtime/
services/game-server/internal/game/scoring/
services/game-server/internal/game/events/
services/game-server/internal/game/space/
```

## Responsibilities

Collision-to-damage flow owns the game-server side of:

* Running destructive collision checks during the authoritative simulation step.
* Respecting the world collision freeze gate.
* Respecting match-over simulation phase order.
* Detecting projectile/asteroid collision facts.
* Detecting player/asteroid collision facts.
* Using wrapped-space local placement for cross-edge collision checks.
* Building damage requests from projectile, asteroid, and player runtime state.
* Resolving damage through `damage.ResolveSingle`.
* Applying returned damage results to runtime asteroid, player, or enemy health/shield fields.
* Recording `damage_applied` events when damage actually changes health or shield.
* Marking hit bullets for delayed despawn.
* Applying asteroid destruction consequences after fatal projectile hits.
* Applying fatal player damage consequences after lethal asteroid/player collisions.
* Spawning projectile impact effects from projectile collision metadata.
* Keeping collision detection, damage math, and gameplay consequences separate.

## Does not own

Collision-to-damage flow does not own:

* Primitive collision algorithms.
* Collision shape source data or export/import pipelines.
* Weapon firing policy.
* Weapon profile lookup.
* Projectile spawn intent construction.
* Pure damage math.
* Damage modifier math.
* Area/radial effect timing and coverage.
* Scoring policy evaluation internals.
* Pickup collection rules.
* Player input, pause, or suspension ownership.
* Room membership or match lifecycle ownership.
* WebSocket transport.
* Packet codec behavior.
* Client rendering, interpolation, audio, or effects.

Those systems may participate in collision outcomes, but they own their own boundaries.

## Domain roles

Collision-to-damage flow participates in the player-facing combat domain by enforcing server authority over:

* projectile hits
* asteroid damage
* asteroid destruction
* asteroid splitting
* pickup drops from destroyed asteroids
* score awards from destroyed asteroids
* player collision damage
* player death
* player lives decrement
* respawn cooldown setup
* combat presentation events

It also participates in technical simulation flow by preserving deterministic phase boundaries inside the game-server tick.

## Simulation phase position

`Game.Step` runs collision checks after player, weapon, movement, spawning, projectile, asteroid, and pickup stepping for active matches.

The active-match path is:

```text
step player sessions
-> step player weapons
-> step players
-> remove ready players
-> step asteroid spawning
-> step asteroids
-> step bullets
-> step pickups
-> step collisions
-> step radial effects
-> simulation step observers
```

If the match is already over, the simulation does not run the collision phase. It still steps cleanup-oriented runtime areas such as asteroids, bullets, pickups, and radial effects before returning.

The collision phase itself is gated by:

```go
game.worldSimulationOptions.CanRunCollisions()
```

When collisions are frozen, projectile/asteroid damage, ship/asteroid damage, scoring from projectile hits, asteroid splitting, and player/pickup collection do not run.

## Collision detection model

`collisions.go` defines narrow collision fact types for current game-pair checks:

```go
type ProjectileAsteroidCollision struct {
    ProjectileID   string
    AsteroidID     string
    ImpactPosition physics.Vector2
}

type PlayerAsteroidCollision struct {
    PlayerID       string
    AsteroidID     string
    ImpactPosition physics.Vector2
}

type PlayerPickupCollision struct {
    PlayerID       string
    PickupID       string
    ImpactPosition physics.Vector2
}
```

The damage-producing detection helpers are:

```go
detectProjectileAsteroidCollision(...)
detectPlayerAsteroidCollision(...)
```

Each helper asks the runtime entities for collision bodies using the loaded `CollisionShapeCatalog`.

For wrapped-world checks, the asteroid body is temporarily placed in wrapped-local space near the actor being tested:

```text
delta = space.Delta(actor.Position(), asteroid.Position())
asteroidBody.Position = actor.Position().Add(delta)
```

This lets collisions work across world boundaries without duplicating stored entities as ghost bodies.

If a required collision body cannot be built, detection returns no collision. Missing shapes do not produce damage.

## Projectile to asteroid flow

Projectile/asteroid collision runs in `handleBulletAsteroidCollisions`.

The flow is:

```text
for each projectile
-> skip projectile if already hit this pass
-> skip projectile if pending despawn
-> for each asteroid
-> skip asteroid if already destroyed by another projectile this pass
-> skip asteroid if pending despawn
-> detect projectile/asteroid collision
-> build projectile asteroid damage request
-> resolve damage
-> apply damage to asteroid health
-> record damage_applied event when useful
-> spawn projectile impact effect from bullet metadata
-> mark projectile as hit
-> if asteroid survived, stop processing that projectile
-> if asteroid was destroyed, record hit consequences
-> after scan, mark hit projectiles pending despawn
-> apply destroyed-asteroid consequences
```

A projectile can hit at most one asteroid in a collision pass.

An asteroid can be selected for destruction consequences at most once in a collision pass.

Nonfatal projectile hits are valid. A projectile can reduce asteroid health, emit `damage_applied`, trigger projectile impact metadata, and despawn without awarding score, spawning fragments, or dropping pickups.

### Projectile damage request

Projectile/asteroid damage requests are built in `projectileAsteroidDamageRequest`.

The source is the projectile:

```text
source entity id   = projectile collision ProjectileID
source entity type = projectile
source cause       = projectile
```

The target is the asteroid:

```text
target entity id   = collision AsteroidID
target entity type = asteroid
target health      = asteroid.Health
target modifiers   = asteroid.DamageModifiers
```

The damage spec normally comes from:

```go
bullet.DamageSpec
```

If the runtime projectile has no explicit damage amount, the adapter falls back to the older bullet damage field as kinetic projectile damage:

```text
amount = bullet.Damage
type   = kinetic
cause  = projectile
```

That fallback keeps older/default bullet construction compatible with the newer damage seam.

### Projectile impact effects

After a projectile/asteroid collision resolves damage, combat asks the game to spawn radial impact effects from the projectile metadata:

```go
game.spawnRadialEffectFromBullet(bullet, bullet.OwnerID, collision.ImpactPosition)
```

This is collision-driven impact behavior. The weapon package does not execute impact effects when firing, and the damage package does not execute radial timing or coverage.

### Asteroid destruction consequences

Destroyed asteroids are passed to `applyProjectileAsteroidDestruction`.

That consequence path:

```text
evaluate score policy
-> apply score awards through game-owned score counter seam
-> mark asteroid pending despawn
-> spawn asteroid fragments
-> maybe drop pickup from asteroid
```

Asteroid destruction is delayed visually through `constants.CollisionDespawnDelay`.

Score is awarded only through the game-owned scoring application seam. The scoring policy computes awards, but game-owned code decides whether the player can receive score.

A paused, suspended, invulnerable, missing, or otherwise ineligible player does not receive score even if their projectile destroys an asteroid.

## Asteroid to player flow

Player/asteroid collision runs in `handleShipAsteroidCollisions`.

The flow is:

```text
for each active player ship
-> skip player if pending despawn
-> skip player if player cannot take collision damage
-> for each asteroid
-> skip asteroid if pending despawn
-> detect player/asteroid collision
-> build player asteroid damage request
-> resolve damage
-> apply damage to player health/shields
-> record damage_applied event when useful
-> if result is fatal player damage, remember player for fatal handling
-> after scan, apply fatal player consequences
```

Ship/asteroid collision damages the player. It does not damage or destroy the asteroid.

A nonfatal asteroid collision can reduce player health or shields, emit `damage_applied`, and leave the player active with lives unchanged.

### Player collision eligibility

A player can take asteroid collision damage only when all of these are true:

```text
player is not pending despawn
player session exists
player session is not suspended
player is not temporarily invulnerable
player damage options allow damage
```

The suspension check includes pause and dev freeze.

The damage options check includes debug invincibility.

This means paused players, dev-frozen players, temporarily invulnerable players, and debug-invincible players do not receive asteroid collision damage.

### Player damage request

Player/asteroid damage requests are built in `playerAsteroidDamageRequest`.

The source is the asteroid:

```text
source entity id   = asteroid id
source entity type = asteroid
source cause       = collision
```

The target is the player:

```text
target entity id   = player id
target entity type = player
target health      = player.Health
target shield      = player.Shields
target modifiers   = player.DamageModifiers
```

The damage spec comes from the asteroid runtime state:

```text
amount = asteroid.CollisionDamage
type   = kinetic
cause  = collision
```

### Fatal player consequences

When a player collision damage result is fatal, `applyFatalPlayerDamage` handles the gameplay consequences.

That path:

```text
store or update camera view at death position
-> mark player pending despawn
-> clear movement through pending-despawn behavior
-> increment session ship deaths
-> decrement lives if life options allow it
-> set respawn cooldown when lives remain
-> log player died or player game over
-> record ship_death event
```

The player entity remains during the collision despawn delay and is removed later when ready for removal.

The fatal flow mutates durable session counters through game-owned player counter seams rather than through the damage package.

## Damage application

Damage results are applied by game-owned helpers:

```go
applyDamageResultToAsteroid(...)
applyDamageResultToPlayer(...)
applyDamageResultToEnemy(...)
```

Each helper ignores results marked `Ignored`.

Asteroid application writes:

```text
asteroid.Health = result.RemainingHealth
```

Player and enemy application write:

```text
ship.Health  = result.RemainingHealth
ship.Shields = result.RemainingShield
```

The damage package calculates the result, but game-owned code mutates runtime state.

## Event recording

Collision-to-damage flow records domain events through game-owned event recording.

Current collision-related events include:

```text
damage_applied
bullet_blast
ship_death
```

`damage_applied` is emitted only when the result is not ignored and damage affected health or shield.

Projectile/asteroid destruction records `bullet_blast` at the projectile impact position.

Fatal player damage records `ship_death` at the player's death position, with the remaining lives and respawn delay.

Events are not packets at the point of damage resolution. They are domain/presentation facts that are later projected into state-packet output.

## Data ownership

Collision-to-damage flow reads and mutates authoritative in-memory game runtime state.

It reads:

```text
game.entities.Projectiles
game.entities.Asteroids
game.entities.Players
game.collisionShapes
game.worldSimulationOptions
player sessions
runtime bullet damage specs
runtime asteroid collision damage
runtime damage modifiers
runtime player health/shields/modifiers/options
```

It mutates:

```text
runtime asteroid health
runtime player health
runtime player shields
projectile pending-despawn state
asteroid pending-despawn state
player pending-despawn state
player session lives
player session ship death count
player session respawn cooldown
camera views after fatal player damage
pending presentation/domain events
score counters through the player counter seam
```

It may trigger:

```text
asteroid fragment spawning
pickup drop evaluation and pickup spawning
radial effect spawning from projectile impact metadata
```

It does not persist durable account/profile data.

## Protocol and API surfaces

Collision-to-damage flow has no direct HTTP or WebSocket API.

External clients observe its results indirectly through normal game-server output:

```text
StatePacket players
StatePacket asteroids
StatePacket bullets
StatePacket player_sessions
StatePacket events
```

Relevant event packet projections include damage, bullet blast, and ship death presentation facts.

Inbound client packets can cause movement, shooting, pause, respawn, and devtools actions that affect later collision outcomes, but clients do not call the collision-to-damage flow directly.

## Invariants

Collision-to-damage flow must preserve these rules:

* The server is authoritative for combat collision outcomes.
* Collision detection does not perform damage math.
* The damage package does not mutate runtime entities.
* Game-owned combat code adapts collision facts into damage requests.
* Game-owned combat code applies damage results to runtime state.
* Projectile impact effects happen after collision, not during weapon firing.
* Asteroid destruction consequences require a destroyed damage result.
* Nonfatal damage must not award score, spawn fragments, or drop pickups.
* Ship/asteroid collision does not destroy the asteroid.
* Fatal player damage must route through player session/lives/respawn ownership.
* Paused, suspended, invulnerable, or debug-invincible players do not take asteroid collision damage.
* Collision checks use wrapped-local placement so cross-boundary hits work.
* Client presentation must observe server outcomes rather than re-decide them.

## Code map

Primary implementation files:

```text
services/game-server/internal/game/simulation.go
services/game-server/internal/game/collisions.go
services/game-server/internal/game/combat.go
services/game-server/internal/game/combat_damage_requests.go
services/game-server/internal/game/combat_damage_application.go
services/game-server/internal/game/damage_events.go
services/game-server/internal/game/asteroid_destruction.go
services/game-server/internal/game/scoring.go
services/game-server/internal/game/pause.go
services/game-server/internal/game/world_simulation_options.go
```

Damage resolver files:

```text
services/game-server/internal/game/damage/types.go
services/game-server/internal/game/damage/request.go
services/game-server/internal/game/damage/result.go
services/game-server/internal/game/damage/resolve.go
services/game-server/internal/game/damage/modifiers.go
services/game-server/internal/game/damage/area.go
services/game-server/internal/game/damage/dot.go
```

Runtime entity files:

```text
services/game-server/internal/game/runtime/state.go
services/game-server/internal/game/runtime/bullet.go
services/game-server/internal/game/runtime/asteroid.go
services/game-server/internal/game/runtime/ship.go
services/game-server/internal/game/runtime/damage_options.go
services/game-server/internal/game/runtime/suspension.go
```

Collision and spatial support files:

```text
services/game-server/internal/game/physics/collision.go
services/game-server/internal/game/physics/collision_shapes.go
services/game-server/internal/game/space/space.go
```

Related consequence files:

```text
services/game-server/internal/game/radial_spawning.go
services/game-server/internal/game/radial_damage_requests.go
services/game-server/internal/game/pickup_drops.go
services/game-server/internal/game/spawning.go
services/game-server/internal/game/player_counters.go
services/game-server/internal/game/events.go
services/game-server/internal/game/events/events.go
```

Important non-ownership boundaries:

```text
services/game-server/internal/game/weapons/
services/game-server/internal/game/effects/radial/
services/game-server/internal/game/scoring/
services/game-server/internal/game/pickups/
services/game-server/internal/networking/
client/
shared/collisions/
```

`weapons` owns fire policy and projectile spawn intent, not collision damage application.

`effects/radial` owns timed radial coverage and hit-intent generation, not direct projectile collision or asteroid destruction consequences.

`scoring` owns pure score policy, not score counter mutation.

`pickups` owns pickup collection/effect rules, not damage-producing collision flow.

`networking` owns packet transport, not combat decisions.

`client` owns presentation only.

`shared/collisions` owns generated collision shape data consumed by server physics.

## Tests and verification

Relevant game integration tests:

```text
services/game-server/tests/game/collision_test.go
services/game-server/tests/game/respawn_test.go
services/game-server/tests/game/devtools_test.go
services/game-server/tests/game/pickups_test.go
```

Relevant package tests:

```text
services/game-server/internal/game/damage/
services/game-server/internal/game/physics/
services/game-server/tests/physics/
```

Current test coverage includes:

* projectile/asteroid delayed despawn
* asteroid splitting after fatal projectile hits
* nonfatal projectile damage without score or fragments
* asteroid damage modifiers
* projectile damage events
* score by asteroid size
* cross-boundary projectile/asteroid collisions
* paused or invulnerable players not receiving score
* ship/asteroid delayed player removal and death events
* cross-boundary ship/asteroid collisions
* nonfatal ship collision damage
* player damage modifiers
* ship collision damage events
* paused players skipping asteroid collision damage
* invulnerable players skipping asteroid collision damage
* collision after invulnerability expires
* debug invincibility blocking asteroid death
* frozen world or frozen collisions blocking collision consequences
* pickup collision respecting frozen collisions
* physics primitive collision behavior
* collision shape catalog loading

Useful verification command:

```bash
cd services/game-server
go test -buildvcs=false ./...
```

Focused verification for damage math:

```bash
cd services/game-server
go test -buildvcs=false ./internal/game/damage
```

Focused verification for game collision behavior:

```bash
cd services/game-server
go test -buildvcs=false ./tests/game -run 'Collision|Respawn|Devtools'
```

## Related docs

* [Game Server Simulation Combat](./!README.md)
* [Game Server Simulation](../!README.md)
* [Game Server](../../!README.md)
* [Game Server Simulation World](../world/!README.md)
* [Game Server Simulation Runtime](../runtime/!README.md)
* [Game Server Simulation Players](../players/!README.md)
* [Game Server Simulation Pickups](../pickups/!README.md)
* [Game Server Simulation Scoring](../scoring/!README.md)
* [Damage Resolution](damage-resolution.md)
* [Weapons And Projectile Fire](weapons-and-projectile-fire.md)
* [Radial Effects](radial-effects.md)
* [Collision Shapes And Physics](../world/stubs/collision-shapes-and-physics.md)
* [Toroidal Space And Motion](../world/stubs/toroidal-space-and-motion.md)
* [Player Pause And Suspension](../players/stubs/player-pause-and-suspension.md)
* [Player Counters](../players/stubs/player-counters.md)
* [Player Death And Despawn](../players/stubs/player-death-and-despawn.md)
* [State Packet Projection](../runtime/stubs/state-packet-projection.md)
* [Data](../../../../data/!README.md)
* [Devtools](../../../../devtools/!README.md)

## Notes

Collision facts currently store impact position as the actor position used for the check: projectile position for projectile/asteroid hits and player position for player/asteroid hits.

Projectile impact effects are collision-triggered, not destruction-triggered. A future weapon or effect may need a stricter distinction between contact, damage applied, and target destroyed if impact behavior becomes more varied.

Player/pickup collision shares the collision phase gate but is not part of the damage path documented here.
