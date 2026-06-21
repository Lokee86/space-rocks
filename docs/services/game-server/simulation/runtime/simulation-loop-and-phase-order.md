# Simulation Loop And Phase Order

Parent index: [Game Server Simulation Runtime](./!INDEX.md)

## Purpose

This document describes the game-server simulation loop and phase-order boundary.

It covers `runSimulation`, `Step(delta)`, server tick cadence, game lock scope, normal active-match phase order, reduced match-over phase order, world simulation freeze gates, collision phase routing, and simulation step observers.

## Overview

The game server is authoritative for live gameplay simulation.

Each `game.Game` owns one in-memory simulation aggregate. Rooms start and stop that aggregate, networking routes decoded player requests into it, and outbound networking later asks it for state packets. The simulation loop itself lives inside the game package.

The runtime loop starts through:

```text
Game.Start
-> runSimulation
-> ticker at constants.ServerTickRate
-> Game.Step(delta)
```

`runSimulation` uses a fixed delta derived from `constants.ServerTickRate`. The current generated server tick rate is `60`, so the normal loop advances simulation at 60 ticks per second.

`Game.Step(delta)` is the authoritative per-tick coordinator. It locks the game aggregate, chooses the wrapped world bounds, advances simulation phases in a fixed order, and invokes registered simulation step observers at the end of the tick.

The phase order is intentionally centralized in `services/game-server/internal/game/simulation.go`. Individual phases delegate into focused helpers, but the root order remains visible in one place.

## Code root

```text
services/game-server/internal/game/
```

Supporting packages include:

```text
services/game-server/internal/game/runtime/
services/game-server/internal/game/motion/
services/game-server/internal/game/space/
services/game-server/internal/game/effects/radial/
services/game-server/internal/game/physics/
services/game-server/internal/constants/
```

## Responsibilities

The simulation loop and phase-order boundary owns:

* Starting the game-owned simulation goroutine through `Game.Start`.
* Stopping the simulation loop through `Game.Stop`.
* Running ticks at `constants.ServerTickRate`.
* Converting tick rate into the fixed simulation delta.
* Locking the `Game` aggregate during each `Step`.
* Preserving the normal active-match phase order.
* Preserving the reduced match-over phase order.
* Keeping match-over simulation from running normal spawn, movement, weapon, and collision phases.
* Applying world simulation option gates for spawning, asteroid movement, bullet movement, and collisions.
* Routing collision phase execution through the current collision-pair handlers.
* Invoking registered simulation step observers after normal or reduced simulation phases.
* Keeping per-phase mutation inside the game aggregate and focused same-package helpers.

## Does not own

This boundary does not own:

* Room lifecycle state.
* Room match start validation.
* WebSocket transport.
* Packet decode or encode mechanics.
* Client-side input collection.
* Client rendering, interpolation, UI, audio, or effects.
* Packet schema source-of-truth files.
* Persistent player data.
* Player-data match result reporting.
* Primitive collision algorithms.
* Pure weapon fire policy.
* Pure damage math.
* Pure score policy.
* Radial effect package internals.
* Devtools command routing.

Those systems participate in or observe simulation, but they own their own boundaries.

## Domain roles

The simulation loop participates in the technical runtime flow for authoritative gameplay.

The room package owns whether a room is in lobby, starting, in game, or game over. The game aggregate owns the live simulation state and match decision used by room lifecycle. Networking owns live WebSocket sessions and outbound state delivery. The client owns presentation only.

The key runtime relationship is:

```text
Room lifecycle
-> starts/stops Game

Networking inbound
-> mutates Game through game-owned APIs

Game.Step
-> advances authoritative state

Networking outbound
-> asks Game for StatePacket

Client
-> renders server results
```

A WebSocket connection does not own the simulation loop. Room membership does not own the simulation loop. The active room owns a `*game.Game` instance, and that game instance owns simulation state.

## Protocols and APIs

This boundary is not a public network protocol by itself.

The main internal runtime surfaces are:

```text
Game.Start()
Game.Stop()
Game.Step(delta)
```

`Game.Start()` is called by room lifecycle when a multiplayer or single-player room starts a match. It uses `sync.Once` so a game instance starts its simulation goroutine only once.

`Game.Stop()` closes the stop channel through `sync.Once`. Room lifecycle calls it when returning a completed room to the lobby or clearing the game instance.

`Game.Step(delta)` is used by `runSimulation` and by tests. It is the simulation coordinator and should remain the place where authoritative phase order is easiest to audit.

Clients do not call these surfaces directly. Clients send gameplay packets through WebSocket networking. Networking resolves room/session/player context and then calls game-owned mutation APIs such as `Game.HandlePacket`. Simulation consumes the resulting stored runtime state during later ticks.

Outbound clients observe simulation indirectly through:

```text
StatePacket players
StatePacket player_sessions
StatePacket player_lifecycle
StatePacket bullets
StatePacket asteroids
StatePacket pickups
StatePacket events
```

State packet projection is a separate runtime responsibility. The simulation loop mutates runtime state; state projection reads that state later and clears per-player pending presentation events after `StatePacket(playerID)` is consumed.

## Tick lifecycle

`Game.Start()` launches `runSimulation()` in a goroutine.

`runSimulation()` creates a ticker using:

```text
time.Second / constants.ServerTickRate
```

It also computes the fixed simulation delta as:

```text
1.0 / constants.ServerTickRate
```

Each ticker event calls:

```text
game.Step(delta)
```

The loop exits when `game.stopSimulation` is closed.

`Game.Stop()` does not directly drain or mutate gameplay state. It stops the simulation loop by closing the stop channel. Room lifecycle owns when a game is stopped and cleared from a room.

## Locking model

`Game.Step(delta)` locks the game aggregate for the full simulation tick:

```text
game.mu.Lock()
defer game.mu.Unlock()
```

The phase helpers called from `Step` run under that lock. They mutate shared runtime maps, player sessions, camera views, radial effects, presentation event queues, and counters without taking separate locks.

The same lock is used by public game APIs that mutate or read live state, including player addition/removal, input routing, pause-state packet generation, match decision reads, counter mutation, targeting, pickups, state packet projection, and devtools adapters.

This means the simulation phase order is serialized against inbound game mutations and outbound state projection.

Simulation step observers are invoked while `Step` still holds the game lock. Current observer usage is narrow devtools integration for continuous bullet streams. Observer callbacks should remain small and route mutations through the intended game-owned devtools adapter functions.

## Normal phase order

For an active match, `Game.Step(delta)` runs this order:

```text
1. stepPlayerSessions(delta)
2. match-over gate
3. stepPlayerWeapons(delta)
4. stepPlayers(delta, bounds)
5. removeReadyPlayers()
6. stepAsteroidSpawning(delta)
7. stepAsteroids(delta, bounds)
8. stepBullets(delta, bounds)
9. stepPickups(delta)
10. stepCollisions()
11. stepRadialEffects(delta)
12. simulationStepObservers
```

### 1. Player sessions

`stepPlayerSessions` advances durable player-session timers.

Current session ticking decrements respawn cooldown toward zero. This happens before the match-over gate.

### 2. Match-over gate

After player sessions tick, `Game.Step` checks:

```text
game.isMatchOverLocked()
```

The match decision is evaluated from player sessions and active ship presence through the rules package. If the match is over, normal active-match phases are skipped and the reduced match-over path runs instead.

### 3. Player weapons

`stepPlayerWeapons` advances cooldown and ammo runtime state for active player ships.

This phase updates per-slot weapon state before player input is consumed for firing in the player phase.

### 4. Players

`stepPlayers` advances active player ships through the motion seam and then consumes stored fire input.

The player phase:

```text
motion.AdvanceShipWithMovePolicy
-> camera view position update
-> skip fire if pending despawn
-> primary fire check
-> secondary fire check
```

Movement is gated by player suspension and pending-despawn state. Shooting is gated by bullet movement options and `playerCanShoot`, then weapon-specific policy applies slot, cooldown, ammo, and equipped-weapon checks.

### 5. Ready player removal

`removeReadyPlayers` removes active player ships whose pending-despawn delay has completed.

This removes the runtime avatar. The durable player session remains available for lifecycle, counters, respawn, and state packet projection.

### 6. Asteroid spawning

`stepAsteroidSpawning` advances timed asteroid spawning.

Asteroid spawning runs only when:

```text
worldSimulationOptions.CanSpawnAsteroids()
game.hasCameraViews()
```

When there are no camera views, the asteroid spawn elapsed timer is reset. When the spawn interval is reached, a batch is spawned for each camera view.

### 7. Asteroids

`stepAsteroids` advances asteroid movement through the motion package when asteroids are not frozen.

It removes asteroids that are ready for removal or far from all camera views.

### 8. Bullets

`stepBullets` advances projectiles through the motion package when bullets are not frozen.

It removes projectiles that are ready for removal, expired, or far from all camera views.

### 9. Pickups

`stepPickups` advances pickup age and expires pickups whose lifespan has elapsed.

Expired pickups record a pickup-expired presentation event before being removed from the runtime pickup map.

### 10. Collisions

`stepCollisions` runs only when:

```text
worldSimulationOptions.CanRunCollisions()
```

When collisions are enabled, the current order is:

```text
handleShipAsteroidCollisions()
handleBulletAsteroidCollisions()
handlePlayerPickupCollisions()
```

This means player/asteroid damage, projectile/asteroid damage, and player/pickup collection share the same collision freeze gate, but each collision family keeps its own consequence logic.

### 11. Radial effects

`stepRadialEffects` advances active radial effects after the normal collision phase.

It builds radial candidates from runtime state, steps each active radial effect, applies returned hit intents through game-owned damage adapters, and removes expired effects.

Radial effects can produce damage and gameplay consequences, but the radial package itself does not mutate the `Game` aggregate.

### 12. Simulation step observers

Registered simulation step observers run last.

Current usage supports devtools continuous bullet streams. Observer callbacks run after regular simulation phases and after radial effect stepping.

## Match-over phase order

If the match is already over after player sessions tick, `Game.Step(delta)` runs a reduced path:

```text
1. stepPlayerSessions(delta)
2. match-over gate
3. stepAsteroids(delta, bounds)
4. stepBullets(delta, bounds)
5. stepPickups(delta)
6. stepRadialEffects(delta)
7. simulationStepObservers
8. return
```

The reduced path intentionally skips:

```text
stepPlayerWeapons
stepPlayers
removeReadyPlayers
stepAsteroidSpawning
stepCollisions
```

This prevents normal active-match behavior from continuing after match completion.

Post-match-over stepping still permits cleanup-safe runtime areas to advance. Pending asteroids and projectiles can finish removal delays, projectiles can expire, pickups can expire, radial effects can finish, and devtools observers can continue to run.

Current tests verify that asteroid spawning does not continue after match over and that cleanup-safe entities do not panic during the reduced match-over step.

## World simulation option gates

`WorldSimulationOptions` owns simulation freeze flags:

```text
FreezeAsteroids
FreezeBullets
FreezeSpawning
FreezeCollisions
```

The current gates are:

```text
AsteroidsCanMove()
BulletsCanMove()
CanSpawnAsteroids()
CanRunCollisions()
```

`SetFreezeWorld(true)` sets all four flags. `SetFreezeWorld(false)` clears all four flags.

These gates do not globally stop the simulation tick. They affect only the phases that explicitly check them.

Current gate effects:

```text
FreezeSpawning
-> disables timed asteroid spawning

FreezeAsteroids
-> disables asteroid movement
-> asteroid cleanup checks still run

FreezeBullets
-> disables projectile movement and lifetime decrement
-> projectile cleanup checks still run
-> player weapon fire checks also require BulletsCanMove()

FreezeCollisions
-> disables ship/asteroid collision
-> disables projectile/asteroid collision
-> disables player/pickup collision
```

Player session timers, pickup aging, radial effect stepping, state packet projection, match decision reads, and simulation step observers are not directly controlled by `WorldSimulationOptions`.

Player pause and dev player-freeze behavior are separate from world simulation options. They route through player suspension state.

## Collision phase routing

The collision phase is deliberately narrow.

`stepCollisions` only decides whether collision families should run. The collision families own their own detection and consequence paths.

Current routing:

```text
stepCollisions
-> if CanRunCollisions
-> handleShipAsteroidCollisions
-> handleBulletAsteroidCollisions
-> handlePlayerPickupCollisions
```

`handleShipAsteroidCollisions` can apply player damage, mark fatal players pending despawn, decrement lives, set respawn cooldown, update camera view, and record ship-death or damage events.

`handleBulletAsteroidCollisions` can apply asteroid damage, mark projectiles pending despawn, spawn projectile impact effects, award score for destroyed asteroids, spawn fragments, and evaluate pickup drops.

`handlePlayerPickupCollisions` can remove collected pickups, resolve pickup collection rules, record pickup collection events, and apply pickup effect intents.

Primitive collision shape math remains in the physics package. Damage math remains in the damage package. Pickup collection rules remain in the pickup rules package. The collision phase only orchestrates the game-owned runtime consequences.

## Simulation observers

`simulationStepObservers` is a game-owned hook list invoked at the end of `Game.Step`.

The current registration surface is:

```text
DevtoolsRegisterSimulationStepObserver(observer func(float64))
```

Current observer usage is the devtools continuous bullet stream path:

```text
devtools continuous stream command
-> ensureContinuousBulletStreamStepObserver
-> DevtoolsRegisterSimulationStepObserver
-> Step observer callback
-> streamruntime.StepContinuousBulletStreams
-> game-owned debug bullet spawn adapter
```

Observers are not a general gameplay scheduling system. They are currently a narrow bridge for devtools behavior that must run inside the authoritative simulation cadence.

## Data ownership

This boundary owns no durable persistence.

It reads and mutates in-memory game runtime data owned by the `Game` aggregate, including:

```text
entities.Players
entities.Projectiles
entities.Asteroids
entities.Pickups
entities.Enemies
playerSessions
cameraViews
pendingPresentationEvents
radialEffects
worldSimulationOptions
asteroidSpawnElapsed
simulationStepObservers
```

It reads generated constants such as:

```text
constants.ServerTickRate
constants.AsteroidSpawnInterval
constants.AsteroidSpawnBatchSize
```

It uses generated or runtime packet state indirectly through state projection, but packet schema ownership belongs to data/protocol documentation, not to the simulation loop.

The simulation loop does not persist profile, account, wallet, or match-result data. Match result reporting happens outside this boundary through room, networking, match-reporting, and player-data integration paths.

## Code map

Primary implementation files:

* `services/game-server/internal/game/game.go` - `Game` aggregate fields, construction defaults, `Start`, and `Stop`.
* `services/game-server/internal/game/simulation.go` - `runSimulation`, `Step(delta)`, normal phase order, match-over phase order, and collision phase routing.
* `services/game-server/internal/game/simulation_players.go` - player session ticking, player movement/fire phase, and ready-player removal.
* `services/game-server/internal/game/simulation_weapons.go` - per-tick weapon state stepping.
* `services/game-server/internal/game/simulation_asteroids.go` - timed asteroid spawning, asteroid movement, and asteroid cleanup.
* `services/game-server/internal/game/simulation_bullets.go` - projectile movement, lifetime stepping, and projectile cleanup.
* `services/game-server/internal/game/pickup_lifecycle.go` - pickup aging, expiration, removal, and expiration events.
* `services/game-server/internal/game/simulation_radial_effects.go` - radial effect stepping, hit application, and expired-effect removal.
* `services/game-server/internal/game/world_simulation_options.go` - world freeze flags and gate helpers.
* `services/game-server/internal/game/match.go` - match-over decision evaluation used by the simulation step gate.
* `services/game-server/internal/game/state_packet.go` - state packet projection that reads post-step runtime state and clears per-player pending events.
* `services/game-server/internal/game/runtime/state.go` - runtime entity store and core runtime entity shapes.
* `services/game-server/internal/game/motion/motion.go` - movement integration and wrapped position advancement for ships, asteroids, and bullets.

Related room and networking files:

* `services/game-server/internal/rooms/room_lifecycle.go` - room lifecycle calls `Game.Start` and `Game.Stop`.
* `services/game-server/internal/rooms/lifecycle_tick.go` - room game-over lifecycle observation.
* `services/game-server/internal/networking/outbound/gameplay_presentation.go` - outbound presentation state reads from `Game.StatePacket`.
* `services/game-server/internal/networking/websocket_gameplay_tick.go` - gameplay presentation tick path.

Related devtools files:

* `services/game-server/internal/game/export_devtools_streams.go` - devtools simulation step observer registration and debug bullet adapter.
* `services/game-server/internal/devtools/continuous_bullet_stream.go` - current observer consumer.
* `services/game-server/internal/devtools/streamruntime/` - continuous bullet stream runtime state outside the game package.

Important non-ownership boundaries:

* `services/game-server/internal/rooms/` owns room lifecycle and active game instance references.
* `services/game-server/internal/networking/` owns WebSocket transport and packet routing.
* `services/game-server/internal/game/weapons/` owns pure weapon fire policy.
* `services/game-server/internal/game/damage/` owns pure damage resolution.
* `services/game-server/internal/game/effects/radial/` owns radial timing, coverage, and hit-intent generation.
* `services/game-server/internal/game/physics/` owns collision primitive math and collision shapes.
* `services/game-server/internal/game/pickups/` owns pure pickup collection/effect rules.
* `client/` owns presentation and input collection, not authoritative phase execution.

## Tests

Relevant focused tests include:

* `services/game-server/internal/game/simulation_match_over_test.go`
* `services/game-server/internal/game/world_simulation_options_test.go`
* `services/game-server/internal/game/player_weapons_test.go`
* `services/game-server/internal/game/radial_effects_test.go`
* `services/game-server/internal/game/radial_projectile_impact_test.go`
* `services/game-server/internal/game/export_devtools_streams_test.go`
* `services/game-server/tests/game/movement_test.go`
* `services/game-server/tests/game/collision_test.go`
* `services/game-server/tests/game/spawning_test.go`
* `services/game-server/tests/game/visibility_test.go`
* `services/game-server/tests/game/pickups_test.go`
* `services/game-server/tests/game/pause_test.go`
* `services/game-server/tests/game/respawn_test.go`
* `services/game-server/tests/game/game_over_test.go`
* `services/game-server/tests/game/state_packet_lifecycle_test.go`
* `services/game-server/tests/game/continuous_bullet_stream_test.go`
* `services/game-server/tests/game/devtools_test.go`

Current verified behavior includes:

* Player, asteroid, and projectile movement wrap through the shared world bounds.
* Timed asteroid spawning depends on camera views and spawn gates.
* Collision consequences stop when collisions are frozen.
* Pickup collection shares the collision freeze gate.
* Paused or suspended players do not move, shoot, or take collision damage.
* Weapon fire from stored input creates projectiles when policy permits.
* Match-over simulation skips normal asteroid spawning.
* Match-over simulation remains cleanup-safe for asteroids, projectiles, and pickups.
* World freeze toggles set and clear spawning, asteroid, bullet, and collision gates together.
* Devtools continuous bullet streams use the simulation observer path.

Useful verification command:

```bash
cd services/game-server
go test -buildvcs=false ./...
```

Focused verification for this boundary:

```bash
cd services/game-server
go test -buildvcs=false ./internal/game ./tests/game
```

## Related docs

* [Game Server Simulation Runtime](./!INDEX.md)
* [Game Server Simulation](../!INDEX.md)
* [Game Server](../../!INDEX.md)
* [Room Match Lifecycle](../../rooms/room-match-lifecycle.md)
* [Player Input Routing](../players/player-input-routing.md)
* [Player Pause And Suspension](../players/player-pause-and-suspension.md)
* [Player Death And Despawn](../players/player-death-and-despawn.md)
* [Player Respawn](../players/player-respawn.md)
* [Weapons And Projectile Fire](../combat/weapons-and-projectile-fire.md)
* [Collision To Damage Flow](../combat/collision-to-damage-flow.md)
* [Radial Effects](../combat/radial-effects.md)
* [Pickup Collection](../pickups/pickup-collection.md)
* [Game Aggregate](game-aggregate.md)
* [Runtime Entity Store](runtime-entity-store.md)
* [State Packet Projection](state-packet-projection.md)
* [Presentation Event Queue](presentation-event-queue.md)
* [Realtime Protocol](../../../../../protocol/!INDEX.md)
* [Data](../../../../../data/!INDEX.md)
* [Devtools](../../../../../devtools/!INDEX.md)

## Notes

Legacy architecture notes correctly identified that `Game.Start()` launches a server-authoritative simulation loop and that `Game.Step()` centralizes phase order while delegating individual phases to focused helpers. This document narrows that legacy material to the current game-server runtime implementation.

The phase order is a service implementation fact. Any change to `Game.Step` order should update this document and any related docs that reference collision order, weapon fire timing, pickup collection timing, radial effects, state packet projection, or match-over behavior.
