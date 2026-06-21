## Game Aggregate

Parent index: [Game Server Simulation Runtime](./!INDEX.md)

## Purpose

This document describes the game-server simulation `Game` aggregate.

The aggregate is the in-memory authoritative runtime owner for one active game instance. It coordinates simulation state, player/session maps, entity stores, runtime dependencies, state-packet projection, presentation event lanes, and the lifecycle shell used by rooms.

## Overview

The game-server simulation aggregate is `game.Game` in `services/game-server/internal/game/game.go`.

A `Game` instance represents one running match simulation. It is not the process, not a room, and not a network connection. Rooms own when a game instance is created, started, stopped, cleared, and associated with room lifecycle. Networking owns how decoded client packets reach the current room’s game instance and how projected state packets are encoded and sent to clients.

Inside the simulation boundary, `Game` owns the mutable runtime state needed to advance authoritative gameplay:

```text
Game
-> player sessions
-> active entity store
-> camera views
-> pending presentation events
-> spawn/scoring/drop/radial dependencies
-> simulation options
-> collision shape catalog
-> lifecycle shell
-> public game-facing APIs
```

The aggregate uses a single mutex around public mutation/read surfaces and the simulation step. Package-local helpers split behavior into focused files, but the aggregate remains the owner of the in-memory state they mutate.

The current runtime shape is intentionally direct. `Game` is still the coordination point for many gameplay subsystems, while focused packages own narrower policies such as motion, damage resolution, match rules, scoring policy, spawning construction, drops, radial stepping, collision primitives, and runtime data shapes.

## Code root

```text
services/game-server/internal/game/
```

Supporting runtime package:

```text
services/game-server/internal/game/runtime/
```

## Responsibilities

The game aggregate owns:

* The `Game` struct as the aggregate root for one simulation instance.
* In-memory runtime state for active players, projectiles, asteroids, enemies, and pickups through `runtime.EntityStore`.
* Per-player session records in `playerSessions`.
* Per-player camera views in `cameraViews`.
* Per-player pending presentation event queues in `pendingPresentationEvents`.
* Game-local ID counters for spawned runtime objects.
* Simulation lifecycle shell through `New`, `Start`, `Stop`, `runSimulation`, and `Step`.
* The synchronization boundary around simulation state through `Game.mu`.
* Construction defaults for collision shapes, spawner, scoring policy, drop tables, radial effect store, entity store, and runtime maps.
* The authoritative game-facing API used by rooms, networking, devtools adapters, tests, and outbound state projection.
* Authoritative mutation coordination for player input, respawn, pause, targeting, counters, pickups, combat consequences, radial effects, and world simulation options.
* State packet projection through `StatePacket`.
* Match decision and match fact read models through `MatchDecision`, `IsGameOver`, and `PlayerMatchFacts`.
* Package-local adaptation between pure subsystem results and game-owned state mutation.
* Simulation step observer registration for narrow devtools/runtime hooks.

## Does not own

The game aggregate does not own:

* HTTP process startup, server routing, or process shutdown.
* WebSocket upgrades, connection sessions, read loops, write loops, or packet encoding.
* Room membership, lobby state, ownership, ready state, room cleanup, or room match lifecycle.
* Durable account, profile, progression, or player-data persistence.
* Client rendering, UI, interpolation, audio, HUD, respawn overlays, or match-results presentation.
* Packet source-of-truth TOML files or data-sync generation.
* Pure collision primitive math.
* Pure damage calculation.
* Pure scoring policy calculation.
* Pure match-rule evaluation.
* Weapon profile definitions or projectile spawn-intent policy internals.
* Drop-table source data.
* Radial effect zone/coverage math internals.
* Devtools command routing.

Those systems can call into the aggregate or be called by the aggregate, but they own their own boundaries.

## Domain roles

The game aggregate participates in the server-authoritative gameplay domain.

Its role is to hold and advance the match-local runtime state that clients observe through realtime packets. The core domain flow is:

```text
room starts match
-> room creates or reuses Game
-> room starts Game simulation
-> networking activates connected room members as game players
-> clients send gameplay packets
-> networking routes decoded gameplay packets to Game
-> Game mutates authoritative state
-> Game.Step advances the simulation
-> outbound networking asks Game for StatePacket
-> clients render projected state
-> rules determine match-over from Game state
-> room marks game over and resolves summary
```

The aggregate is match-local. It does not represent platform identity, durable profile state, or account progression.

## Runtime state model

The `Game` struct currently contains:

```go
type Game struct {
    mu                        sync.Mutex
    stopSimulation            chan struct{}
    startSimulationOnce       sync.Once
    stopSimulationOnce        sync.Once
    nextID                    int
    nextPickupID              int
    spawner                   *spawning.Spawner
    scoringPolicy             scoring.Policy
    dropTables                drops.Tables
    radialEffects             radial.Store
    asteroidSpawnElapsed      float64
    worldSimulationOptions    WorldSimulationOptions
    collisionShapes           physics.CollisionShapeCatalog
    entities                  runtime.EntityStore
    simulationStepObservers   []func(float64)
    cameraViews               map[string]*runtime.CameraView
    playerSessions            map[string]*playerSession
    pendingPresentationEvents map[string][]EventState
}
```

These fields are aggregate-owned. Focused files in the `game` package mutate them under the same aggregate boundary.

The main groups are:

```text
lifecycle and synchronization
= mu, stopSimulation, startSimulationOnce, stopSimulationOnce

identity and spawn counters
= nextID, nextPickupID, asteroidSpawnElapsed

composed dependencies
= spawner, scoringPolicy, dropTables, radialEffects, collisionShapes

runtime state stores
= entities, playerSessions, cameraViews, pendingPresentationEvents

dev/runtime controls
= worldSimulationOptions, simulationStepObservers
```

`runtime.EntityStore` groups active entity maps:

```text
Players
Projectiles
Asteroids
Enemies
Pickups
```

The aggregate owns the store. The runtime package owns the data shapes.

## Construction defaults

`game.New()` creates a fresh simulation aggregate.

Current construction behavior:

```text
load collision shape catalog
-> warn if collision shapes are unavailable
-> create stop channel
-> create camera view map
-> create player session map
-> create presentation event map
-> create spawning.Spawner
-> create default scoring policy
-> attach generated drop tables
-> create radial effect store
-> create runtime entity store
```

The constructor does not start the simulation loop. It only prepares the aggregate and its default dependencies.

Collision shape loading is best-effort at construction time. If loading fails, the aggregate still exists and logs a warning through the game logger. Collision-dependent paths must handle missing shape lookup where relevant.

## Lifecycle shell

The aggregate lifecycle surface is intentionally small:

```text
New
Start
Stop
Step
```

`Start` launches the simulation loop once. It uses `startSimulationOnce` so repeated calls do not start duplicate loops.

`Stop` closes the stop channel once. It uses `stopSimulationOnce` so repeated calls do not panic by closing the channel more than once.

`runSimulation` creates a ticker from `constants.ServerTickRate`, derives a fixed delta from the same tick rate, and calls `Step(delta)` for each tick until the stop channel is closed.

`Step` is also callable directly by tests and controlled dev/runtime paths. Direct stepping uses the same aggregate lock and phase coordinator as the ticker-driven loop.

Room lifecycle owns when this lifecycle shell is called. `Room.StartGameForMember` and `Room.StartSinglePlayerGame` call `Game.Start`. `Room.ResetToLobby`, room cleanup, and relevant tests call `Game.Stop`.

## Synchronization boundary

`Game.mu` is the synchronization boundary for aggregate state.

Public methods that read or mutate runtime state lock the aggregate before touching shared fields. Current examples include:

```text
AddPlayer
RemovePlayer
HandlePacket
StatePacket
IsGameOver
MatchDecision
PlayerMatchFacts
SetPlayerScore
AddPlayerScore
SetPlayerLives
AddPlayerLives
SpawnPickup
RemovePickup
targeting APIs
pause packet APIs
devtools export APIs
Step
```

Package-local helper methods generally assume the caller already holds the lock when their names or usage indicate locked aggregate context.

This lock keeps the ticker-driven simulation loop, inbound packet handling, outbound state projection, room match checks, and devtools adapters from concurrently mutating the same maps.

## Public runtime surface

The aggregate exposes the game-facing runtime API used by adjacent service boundaries.

Main room-facing and networking-facing methods:

```text
New
Start
Stop
AddPlayer
RemovePlayer
HandlePacket
StatePacket
IsGameOver
MatchDecision
PlayerMatchFacts
```

Player and gameplay mutation surfaces include:

```text
SetPlayerScore
AddPlayerScore
SetPlayerLives
AddPlayerLives
SetTarget
SetPlayerTarget
SelectTargetAtPosition
ClearTarget
Target
PlayerTarget
ClearPlayerTarget
PlayerPauseStatePacket
SpawnPickup
RemovePickup
```

Devtools-facing surfaces are exposed through `export_devtools_*.go` files. They are intentionally narrow adapters around game-owned state and should not cause `internal/game` to import devtools packages.

## Protocols and APIs

The game aggregate has no HTTP API.

Its runtime surfaces are Go service methods and realtime packet consequences. Networking receives decoded client packets and forwards gameplay requests into the current room’s game instance.

Inbound gameplay packets can reach the aggregate through:

```text
input
respawn
client_config
pause_request
set_target_player_request
select_target_at_position_request
clear_target_request
```

The gameplay network adapter handles routing and request adaptation. The aggregate owns authoritative mutation behind those requests.

Outbound realtime state reaches clients through `Game.StatePacket(playerID)`. Outbound networking calls this method, stamps `server_sent_msec`, encodes the packet through `packetcodec`, and writes it to the websocket session.

`StatePacket` projection includes:

```text
self_id
lives
players
player_sessions
player_lifecycle
bullets
asteroids
pickups
total_asteroids
events
server_sent_msec
```

`Game.StatePacket` also clears that player’s pending presentation events after copying them into the response. This makes the event lane player-specific and packet-facing.

## Data ownership

The game aggregate owns in-memory match-local data only.

It mutates:

```text
runtime entity maps
player sessions
camera views
pending presentation events
spawn counters
pickup counters
asteroid spawn elapsed time
radial effect store
world simulation options
```

It reads or composes data from:

```text
generated constants
generated packet structs
generated drop tables
collision shape catalog
runtime entity types
weapon profiles and weapon state
scoring policy
match rules
damage resolver
motion helpers
spawning helpers
radial effect helpers
pickup rules
```

It does not persist durable profile or account data. Match result persistence is outside the aggregate. The aggregate exposes match facts and decisions; rooms and player-data integration own higher-level match-result routing.

Packet shapes and generated runtime packet structs come from shared packet source files and data-sync output. The aggregate consumes the generated Go types; it does not own the source-of-truth packet definitions.

## Aggregate-owned dependencies

The aggregate composes several focused dependencies:

```text
spawning.Spawner
= asteroid/projectile spawn construction, spawn ID support, total asteroid count

scoring.Policy
= pure score award calculation

drops.Tables
= generated drop table data used when asteroid destruction may create pickups

radial.Store
= active radial effect storage

physics.CollisionShapeCatalog
= loaded collision shapes for collision-dependent gameplay paths

runtime.EntityStore
= active runtime entity maps
```

These dependencies keep policy and data-shape concerns out of the aggregate where possible, while the aggregate remains responsible for applying their outputs to authoritative game state.

## Simulation coordination

`Game.Step(delta)` is the aggregate’s simulation coordinator.

It locks the aggregate, chooses default toroidal world bounds, steps player sessions, then either runs the normal active-match phase order or the reduced match-over phase order.

Normal active-match phase order:

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
-> notify simulation step observers
```

When the match is already over, the aggregate does not run player weapons, player movement, player removal, asteroid spawning, or collision resolution. It still steps asteroids, bullets, pickups, radial effects, and simulation observers before returning.

Detailed phase ownership belongs in [Simulation Loop And Phase Order](simulation-loop-and-phase-order.md). This document only records that the aggregate owns the top-level coordinator and lock boundary.

## Presentation event lane

The aggregate stores generated packet-facing presentation events in:

```text
pendingPresentationEvents map[string][]EventState
```

This queue is per player. Domain events are recorded through game-owned event adapters, translated to packet-facing `EventState`, then appended to every current player session’s pending lane.

`StatePacket(playerID)` copies that player’s pending events into the packet and clears only that player’s lane.

This is not the domain event store. It is a packet presentation queue for client-visible effects such as bullet blasts, ship death, pickup events, radial effect starts, and damage event presentation.

## Match read models

The aggregate exposes match state through read-model methods rather than letting rooms inspect maps directly.

`MatchDecision` locks the aggregate and evaluates match status through the rules package from a plain snapshot. The snapshot is built from:

```text
player sessions
active ship presence
pending-despawn state
remaining lives
```

`IsGameOver` delegates to the same decision path.

`PlayerMatchFacts` projects match summary facts from player sessions:

```text
game player id
score
ship deaths
```

Rooms use these read models to decide when room state can move to game-over and to build match result summaries. The aggregate does not own room state transitions.

## Runtime invariants

The game aggregate must preserve these rules:

* One `Game` instance represents one match-local simulation aggregate.
* Room lifecycle owns game instance creation, start, stop, clear, and room state transitions.
* Networking owns packet transport and only routes decoded gameplay requests into the aggregate.
* `Game.mu` protects aggregate state across simulation, inbound packet mutation, outbound state projection, and devtools adapters.
* `Start` must not launch duplicate simulation loops.
* `Stop` must not close the stop channel more than once.
* `Step` is the only top-level simulation phase coordinator.
* Runtime maps remain aggregate-owned even when package helpers mutate them.
* `runtime` package types are state shapes, not aggregate owners.
* Pending presentation events are packet-facing event lanes, not the domain event source of truth.
* `StatePacket` projection must copy state out of aggregate-owned maps instead of exposing mutable map contents directly.
* Match decisions are evaluated from aggregate state through the rules package.
* Durable profile/account state must not be stored in `Game`.
* Devtools must use narrow exported game-owned adapters and must not become imported aggregate dependencies.

## Code map

Primary implementation files:

```text
services/game-server/internal/game/game.go
services/game-server/internal/game/simulation.go
services/game-server/internal/game/state_packet.go
services/game-server/internal/game/players.go
services/game-server/internal/game/session.go
services/game-server/internal/game/input.go
services/game-server/internal/game/match.go
services/game-server/internal/game/match_facts.go
services/game-server/internal/game/player_counters.go
services/game-server/internal/game/events.go
services/game-server/internal/game/world_simulation_options.go
```

Runtime state and generated packet files:

```text
services/game-server/internal/game/runtime/state.go
services/game-server/internal/game/runtime/packets_generated.go
services/game-server/internal/game/packets.go
```

Simulation helper files under the aggregate package:

```text
services/game-server/internal/game/simulation_players.go
services/game-server/internal/game/simulation_weapons.go
services/game-server/internal/game/simulation_asteroids.go
services/game-server/internal/game/simulation_bullets.go
services/game-server/internal/game/simulation_radial_effects.go
services/game-server/internal/game/pickup_lifecycle.go
services/game-server/internal/game/pickup_collisions.go
services/game-server/internal/game/pickups.go
services/game-server/internal/game/combat.go
services/game-server/internal/game/targeting.go
services/game-server/internal/game/pause.go
```

Composed subsystem packages:

```text
services/game-server/internal/game/runtime/
services/game-server/internal/game/motion/
services/game-server/internal/game/space/
services/game-server/internal/game/physics/
services/game-server/internal/game/spawning/
services/game-server/internal/game/scoring/
services/game-server/internal/game/rules/
services/game-server/internal/game/damage/
services/game-server/internal/game/drops/
services/game-server/internal/game/pickups/
services/game-server/internal/game/effects/radial/
services/game-server/internal/game/events/
services/game-server/internal/game/weapons/
```

Room and networking integration points:

```text
services/game-server/internal/rooms/room_match.go
services/game-server/internal/rooms/room_lifecycle.go
services/game-server/internal/rooms/lifecycle.go
services/game-server/internal/rooms/leave.go
services/game-server/internal/networking/player_activation.go
services/game-server/internal/networking/inbound/gameplay.go
services/game-server/internal/networking/outbound/gameplay_presentation.go
```

Devtools adapter files:

```text
services/game-server/internal/game/export_devtools.go
services/game-server/internal/game/export_devtools_*.go
```

Generated/source files:

```text
shared/packets/gameplay.toml
shared/packets/outputs.toml
shared/constants/server_constants.toml
shared/constants/server_entities.toml
shared/drop_tables/basicasteroids.toml
services/game-server/internal/constants/constants.go
services/game-server/internal/game/packets.go
services/game-server/internal/game/runtime/packets_generated.go
services/game-server/internal/game/drops/drop_tables.go
```

Important non-ownership boundaries:

```text
services/game-server/internal/rooms/
services/game-server/internal/networking/
services/game-server/internal/protocol/packetcodec/
services/player-data/
services/api-server/
client/
```

## Tests and verification

Relevant game integration tests:

```text
services/game-server/tests/game/game_over_test.go
services/game-server/tests/game/match_decision_test.go
services/game-server/tests/game/state_packet_lifecycle_test.go
services/game-server/tests/game/movement_test.go
services/game-server/tests/game/respawn_test.go
services/game-server/tests/game/pause_test.go
services/game-server/tests/game/collision_test.go
services/game-server/tests/game/spawning_test.go
services/game-server/tests/game/visibility_test.go
services/game-server/tests/game/player_counters_test.go
services/game-server/tests/game/pickups_test.go
services/game-server/tests/game/continuous_bullet_stream_test.go
services/game-server/tests/game/devtools_test.go
```

Relevant package tests:

```text
services/game-server/internal/game/...
services/game-server/internal/rooms/...
services/game-server/internal/networking/...
services/game-server/internal/devtools/...
```

Useful verification command:

```bash
cd services/game-server
go test -buildvcs=false ./...
```

Run generated-data checks when packet, constant, or generated runtime shapes change:

```bash
data-sync -check -packets -go -gds
```

Expected behavioral coverage includes:

* game construction through `game.New`
* room-owned game instance start and stop
* player activation into a game instance
* input routing into `Game.HandlePacket`
* state packet projection through `Game.StatePacket`
* match-over decision evaluation
* player match fact projection
* score and lives counter mutation
* simulation stepping
* devtools adapters mutating game-owned state through narrow seams

## Related docs

* [Game Server Simulation Runtime](./!INDEX.md)
* [Game Server Simulation](../!INDEX.md)
* [Game Server](../../!INDEX.md)
* [Game Server Rooms](../../rooms/!INDEX.md)
* [Room Match Lifecycle](../../rooms/room-match-lifecycle.md)
* [Game Server Networking](../../networking/!INDEX.md)
* [Gameplay Network Adapter](../../networking/gameplay-network-adapter.md)
* [Outbound Message Flow](../../networking/outbound-message-flow.md)
* [Game Server Simulation Players](../players/!INDEX.md)
* [Player Session State](../players/player-session-state.md)
* [Active Player Avatar State](../players/active-player-avatar-state.md)
* [Player Counters](../players/player-counters.md)
* [Player Input Routing](../players/player-input-routing.md)
* [Player Pause And Suspension](../players/player-pause-and-suspension.md)
* [Player Respawn](../players/player-respawn.md)
* [Game Server Simulation Combat](../combat/!INDEX.md)
* [Game Server Simulation Pickups](../pickups/!INDEX.md)
* [Game Server Simulation Targeting](../targeting/!INDEX.md)
* [Game Server Simulation World](../world/!INDEX.md)
* [Runtime Entity Store](runtime-entity-store.md)
* [Simulation Loop And Phase Order](simulation-loop-and-phase-order.md)
* [State Packet Projection](state-packet-projection.md)
* [Presentation Event Queue](presentation-event-queue.md)
* [Gameplay Packets](../../../../protocol/gameplay-packets.md)
* [Realtime WebSocket Protocol](../../../../protocol/realtime-websocket-protocol.md)
* [Data Pipeline](../../../../data/!INDEX.md)

## Notes

The legacy architecture material’s useful current facts are that gameplay state is server-authoritative, `Game.Start()` launches the simulation loop at the server tick rate, `Game.Step()` is the same-package simulation coordinator, and `pendingPresentationEvents` is a packet-facing presentation queue rather than the domain event queue.

This document intentionally does not detail the full simulation phase order, state-packet field projection, entity store shape, or presentation event queue mechanics. Those are adjacent runtime docs so the aggregate doc can stay focused on root ownership, lifecycle, synchronization, and service surfaces.
