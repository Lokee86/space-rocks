## State Packet Projection

Parent index: [Game Server Simulation Runtime](./!README.md)

## Purpose

This document describes game-server gameplay state packet projection.

It covers how the authoritative `Game` aggregate projects runtime simulation state into per-player outbound `StatePacket` payloads for realtime client presentation and synchronization.

## Overview

State packet projection is the game-server simulation boundary that turns current authoritative game state into the packet-facing `state` payload consumed by the client.

The current flow is:

```text
networking write tick
-> outbound.BuildGameplayPresentationStateResponse
-> room.GameInstance().StatePacket(playerID)
-> Game.statePacket(playerID)
-> packetcodec.Encode
-> WebSocket write
```

`Game.StatePacket(playerID)` is per-player because the packet includes:

```text
self_id
lives
events
server_sent_msec
```

`self_id`, `lives`, and `events` are specific to the requesting player. Most world maps in the packet are shared authoritative read models for the current game instance.

The state packet currently includes:

```text
type
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

The game projection builds the packet while holding the game mutex. It snapshots active entity maps, session read models, lifecycle classification, asteroid total count, and the requesting player's pending presentation events. After the packet is built, `Game.StatePacket` flushes only that player's pending presentation event queue.

`server_sent_msec` is not stamped by `Game.statePacket`. It is stamped by the outbound networking helper immediately before encoding.

## Code root

```text
services/game-server/internal/game/
services/game-server/internal/game/runtime/
services/game-server/internal/networking/outbound/
```

## Responsibilities

State packet projection owns:

* Locking the game aggregate while building a packet-facing snapshot.
* Returning a `StatePacket` for one requesting player ID.
* Setting the packet type to `state`.
* Projecting the requesting player ID into `self_id`.
* Projecting the requesting player's current lives into the top-level `lives` field.
* Projecting active ships from `game.entities.Players` into `players`.
* Projecting durable per-player session state from `game.playerSessions` into `player_sessions`.
* Projecting match lifecycle classification into `player_lifecycle`.
* Projecting active projectiles from `game.entities.Projectiles` into `bullets`.
* Projecting active asteroids from `game.entities.Asteroids` into `asteroids`.
* Projecting active pickups from `game.entities.Pickups` into `pickups`.
* Projecting the spawner's total spawned asteroid count into `total_asteroids`.
* Copying the requesting player's pending presentation events into `events`.
* Flushing the requesting player's pending presentation events after packet construction.

## Does not own

State packet projection does not own:

* Simulation stepping or phase order.
* Entity creation, movement, collision, despawn, or removal.
* Player input handling.
* Player lifecycle decisions.
* Score, lives, death, respawn, or pause mutation.
* Event creation.
* Event queue semantics outside the packet drain point.
* Room membership or room lifecycle.
* WebSocket transport.
* Packet encoding.
* Packet schema source-of-truth files.
* Client packet decoding.
* Client rendering, interpolation, HUD, effects, or audio.
* Durable account/profile/player-data persistence.
* Future realtime protocol delivery policy, deltas, binary encoding, or packet prioritization.

## Domain roles

State packet projection is a simulation read-model boundary.

It answers:

```text
What authoritative game state should this connected player receive right now?
```

It does not decide the gameplay state. The simulation, player lifecycle, combat, pickups, weapons, spawning, and rules packages mutate or classify state before projection.

The projection boundary keeps these concepts separate:

```text
game.entities.Players
= active ship/avatar state

game.playerSessions
= match-local durable player session state

rules.MatchDecision
= lifecycle classification from plain facts

pendingPresentationEvents[playerID]
= per-player event lane waiting to be sent once

StatePacket
= packet-facing read model for one player
```

## Projection flow

The public packet builder is:

```go
func (game *Game) StatePacket(playerID string) StatePacket
```

It performs the lock and flush behavior:

```text
lock game
-> build response with statePacket(playerID)
-> clear pendingPresentationEvents[playerID]
-> unlock game
-> return response
```

The unexported builder is:

```go
func (game *Game) statePacket(playerID string) StatePacket
```

It builds new maps for packet output instead of exposing the live runtime maps directly.

Current packet projection order:

```text
1. Copy active player ships into players.
2. Evaluate match decision.
3. Convert match player decisions into player_lifecycle.
4. Copy player session read models into player_sessions.
5. Copy active asteroids into asteroids.
6. Copy active pickups into pickups.
7. Copy active projectiles into bullets.
8. Copy pending events for the requesting player.
9. Return StatePacket.
```

## Projected packet fields

### `type`

`type` is set to:

```text
state
```

The constant is generated as `PacketTypeState`.

### `self_id`

`self_id` is the `playerID` argument passed into `Game.StatePacket`.

It identifies the receiving player's game-player identity inside this game instance.

### `lives`

The top-level `lives` field is the requesting player's current session lives.

It is resolved through `game.playerLives(playerID)`. If the player session is missing, the projected lives value is `0`.

Durable lives are session-owned. The top-level field is a convenience projection for the local receiver.

### `players`

`players` is projected from:

```text
game.entities.Players
```

Each active ship is converted through:

```go
runtime.Ship.State()
```

Current `ShipState` fields are:

```text
id
ship_type
x
y
rotation
health
shields
thrusting
target_kind
target_id
primary_weapon_id
primary_ammo_policy
primary_cooldown_remaining
primary_ammo_remaining
secondary_weapon_id
secondary_ammo_policy
secondary_cooldown_remaining
secondary_ammo_remaining
```

`StatePacket.players` is active avatar/render state only.

It does not own durable score, durable lives, respawn cooldown, spawn position, pause state, or match lifecycle status. Those belong to `player_sessions`, `player_lifecycle`, or adjacent player docs.

Pending-respawn and eliminated players may be absent from `players` while still present in `player_sessions` and `player_lifecycle`.

### `player_sessions`

`player_sessions` is projected from:

```text
game.playerSessions
```

Each non-nil session becomes `PlayerSessionState`.

Current projected session fields are:

```text
id
ship_type
score
lives
respawn_cooldown
primary_weapon_id
primary_ammo_policy
secondary_weapon_id
secondary_ammo_policy
spawn_x
spawn_y
```

This is match-local runtime state, not durable platform profile data.

### `player_lifecycle`

`player_lifecycle` is projected from:

```text
game.matchDecisionLocked()
```

`matchDecisionLocked` evaluates a plain match snapshot through the rules package.

Lifecycle projection uses one string status per player ID:

```text
active
pending_respawn
eliminated
```

The match snapshot is built from session and active-avatar facts:

```text
session exists
active ship exists
active ship is not pending despawn
session has remaining lives
```

Do not infer lifecycle from `players` alone. `players` is active ship state; `player_lifecycle` is the lifecycle read model.

### `bullets`

`bullets` is projected from:

```text
game.entities.Projectiles
```

Each projectile is converted through:

```go
runtime.Bullet.State()
```

Current `BulletState` fields are:

```text
id
owner_id
x
y
rotation
weapon_id
projectile_type
```

The packet projection does not own bullet creation, lifetime stepping, collision, impact effects, or despawn behavior.

### `asteroids`

`asteroids` is projected from:

```text
game.entities.Asteroids
```

Each asteroid is converted through:

```go
runtime.Asteroid.State()
```

Current `AsteroidState` fields are:

```text
id
x
y
size
health
scale
variant
```

`scale` is derived from asteroid size and server constants during projection from runtime state.

The packet projection does not own spawning, movement, splitting, damage, variant selection, or despawn behavior.

### `pickups`

`pickups` is projected by:

```go
game.pickupStatesLocked()
```

Current `PickupState` fields are:

```text
id
type
pickup_class
x
y
health
age_seconds
lifespan_seconds
```

Pickup packet state is active pickup state only. Removed pickups disappear from later `pickups` maps. Semantic pickup feedback, such as collection or expiry, is carried by events when applicable.

### `total_asteroids`

`total_asteroids` is projected from:

```go
game.spawner.TotalAsteroidsSpawned()
```

It is a cumulative spawned-asteroid count from the current game spawner, not the current active asteroid map length.

### `events`

`events` is copied from:

```text
game.pendingPresentationEvents[playerID]
```

The projection uses a new slice so the returned packet keeps the events even after the game queue is flushed.

Current event payloads use the generated `EventState` shape. Event types currently include:

```text
bullet_blast
ship_death
pickup_dropped
pickup_collected
pickup_effect_applied
pickup_expired
radial_effect_started
damage_applied
damage_over_time_started
damage_over_time_tick
```

The event queue is per player. `recordDomainEvent` broadcasts packet-facing events to every current player session by appending to each player's pending event queue. `Game.StatePacket(playerID)` drains only the queue for the player whose packet was requested.

### `server_sent_msec`

`server_sent_msec` exists on `StatePacket`, but the game projection leaves it at the zero value.

The networking outbound helper stamps it in:

```text
services/game-server/internal/networking/outbound/gameplay_presentation.go
```

The stamp happens after `Game.StatePacket(playerID)` returns and immediately before packet encoding.

## Protocols and APIs

Gameplay state packets are sent by the WebSocket write loop at the server tick rate when gameplay presentation state is eligible.

Eligibility requires:

```text
session.currentGamePlayerID is not empty
room exists
room has a game instance
room state is InGame or GameOver
```

The write loop calls:

```go
outbound.BuildGameplayPresentationStateResponse(room, playerID, roomID, remoteAddr)
```

That helper:

```text
1. Gets the room's game instance.
2. Calls Game.StatePacket(playerID).
3. Stamps server_sent_msec.
4. Encodes the packet through packetcodec.
5. Logs packet-size diagnostics.
6. Returns encoded bytes to the write loop.
```

The outbound helper does not mutate simulation state except for the event-queue drain that occurs inside `Game.StatePacket`.

## Data ownership

State packet projection reads in-memory game-server runtime state.

It reads:

```text
game.entities.Players
game.entities.Projectiles
game.entities.Asteroids
game.entities.Pickups
game.playerSessions
game.pendingPresentationEvents
game.spawner
rules.MatchDecision
```

It mutates:

```text
game.pendingPresentationEvents[playerID]
```

The mutation is limited to flushing the requesting player's event queue after packet construction.

Packet shape source data lives in:

```text
shared/packets/gameplay.toml
```

Generated server packet outputs include:

```text
services/game-server/internal/game/packets.go
services/game-server/internal/game/runtime/packets_generated.go
```

Generated client packet output includes:

```text
client/scripts/generated/networking/packets/packets.gd
```

State packet projection is not persistence. It does not write account, profile, progression, match-result, or player-data records.

## Invariants

State packet projection must preserve these rules:

* Build packets while holding the game mutex.
* Do not return references to live mutable game maps as packet state.
* Keep active ship state in `players`.
* Keep durable match-local player state in `player_sessions`.
* Keep lifecycle classification in `player_lifecycle`.
* Do not infer lifecycle from `players`.
* Drain only the requesting player's pending presentation events after building that player's packet.
* Stamp `server_sent_msec` in outbound networking, not in game simulation.
* Keep packet schema ownership in shared packet source files and generated outputs.
* Keep client presentation and interpolation out of server projection code.
* Keep room/lobby snapshots separate from gameplay state packets.

## Code map

Primary implementation files:

```text
services/game-server/internal/game/state_packet.go
```

Owns `Game.StatePacket`, `Game.statePacket`, per-player state projection, and pending event flush after projection.

```text
services/game-server/internal/game/player_session_state.go
```

Projects `playerSession` records into `PlayerSessionState`.

```text
services/game-server/internal/game/pickups.go
```

Projects active pickup entities into packet-facing `runtime.PickupState`.

```text
services/game-server/internal/game/events.go
```

Converts gameplay domain events into packet-facing `EventState` values and appends them to per-player pending presentation queues.

```text
services/game-server/internal/game/match.go
```

Builds match snapshots and evaluates match decisions used by `player_lifecycle`.

```text
services/game-server/internal/game/game.go
```

Defines the `Game` aggregate fields read by state packet projection, including `entities`, `playerSessions`, `pendingPresentationEvents`, and `spawner`.

Runtime state projection files:

```text
services/game-server/internal/game/runtime/ship.go
services/game-server/internal/game/runtime/bullet.go
services/game-server/internal/game/runtime/asteroid.go
services/game-server/internal/game/runtime/state.go
```

Generated packet files:

```text
services/game-server/internal/game/packets.go
services/game-server/internal/game/runtime/packets_generated.go
client/scripts/generated/networking/packets/packets.gd
```

Packet source files:

```text
shared/packets/gameplay.toml
shared/packets/outputs.toml
```

Outbound networking files:

```text
services/game-server/internal/networking/websocket_write.go
services/game-server/internal/networking/outbound/gameplay_presentation.go
services/game-server/internal/networking/outbound/gameplay_state_metrics.go
```

Client consumer files:

```text
client/scripts/gameplay/state/gameplay_state_flow.gd
client/scripts/gameplay/state/gameplay_state_packet_reader.gd
client/scripts/gameplay/state/gameplay_state_apply_flow.gd
client/scripts/gameplay/runtime/gameplay_world_state_apply_flow.gd
client/scripts/world/world_sync.gd
```

Important non-ownership boundaries:

```text
services/game-server/internal/rooms/
services/game-server/internal/networking/
services/game-server/internal/protocol/packetcodec/
services/game-server/internal/game/rules/
services/game-server/internal/game/events/
client/
shared/packets/
```

`rooms` owns room state and room lifecycle.

`networking` owns WebSocket write timing, `server_sent_msec`, packet encoding handoff, and transport.

`packetcodec` owns JSON encode/decode mechanics.

`rules` owns pure lifecycle classification.

`events` owns domain event vocabulary.

`client` owns packet normalization, state application, world rendering, HUD, and presentation events.

`shared/packets` owns the packet schema source of truth.

## Tests and verification

Relevant game-server tests include:

```text
services/game-server/tests/game/state_packet_lifecycle_test.go
services/game-server/internal/game/events_test.go
services/game-server/internal/game/pickup_drops_test.go
services/game-server/internal/game/targeting_test.go
services/game-server/tests/game/packets_generated_test.go
```

Relevant outbound networking tests include:

```text
services/game-server/internal/networking/outbound/gameplay_presentation_test.go
```

Relevant client consumer tests include:

```text
client/tests/unit/test_gameplay_state_packet_reader.gd
client/tests/unit/test_gameplay_state_apply_flow.gd
```

Run game-server verification after changing projection behavior:

```bash
go test -buildvcs=false ./services/game-server/internal/game/...
go test -buildvcs=false ./services/game-server/tests/game/...
go test -buildvcs=false ./services/game-server/internal/networking/outbound/...
```

Run generated packet checks when packet shape or generated packet output changes:

```bash
data-sync -check -packets -go -gds
```

Run client state-reader and state-application tests after changing packet fields consumed by the client.

## Related docs

* [Game Server Simulation Runtime](./!README.md)
* [Game Server Simulation](../!README.md)
* [Game Server](../../!README.md)
* [Outbound Packet Routing](../../networking/outbound-message-flow.md)
* [Gameplay Network Adapter](../../networking/gameplay-network-adapter.md)
* [Player Session State](../players/player-session-state.md)
* [Active Player Avatar State](../players/active-player-avatar-state.md)
* [Player Lifecycle Overview](../players/player-lifecycle-overview.md)
* [Player Respawn](../players/player-respawn.md)
* [Pickup Entity Lifecycle](../pickups/pickup-entity-lifecycle.md)
* [Pickup Collection](../pickups/pickup-collection.md)
* [Pickup Effects](../pickups/pickup-effects.md)
* [Weapons And Projectile Fire](../combat/weapons-and-projectile-fire.md)
* [Damage Resolution](../combat/damage-resolution.md)
* [Radial Effects](../combat/radial-effects.md)
* [Gameplay State Application](../../../client/gameplay-runtime/gameplay-state-application.md)
* [World Sync Coordinator](../../../client/world-sync/world-sync-coordinator.md)
* [Gameplay Events And Effects](../../../client/gameplay-event-presentation/gameplay-events-and-effects.md)
* [Gameplay packets](../../../../protocol/stubs/gameplay-packets.md) - Gameplay realtime packet documentation.
* [Packet schema pipeline](../../../../data/stubs/packet-schema-pipeline.md) - Packet schema data documentation.
* [Realtime Protocol Architecture](../../../../planning/protocol/realtime-protocol-architecture.md)

## Notes

The legacy architecture docs correctly warn that consumers and wrappers must deliberately preserve new `StatePacket` fields. Current gameplay state output is built through the outbound gameplay presentation helper, while debug status is sent as a separate packet lane.

`runtime.EntityStore` contains an `Enemies` map, but the current `StatePacket` does not project enemies.

`StatePacket.events` is a presentation/event lane embedded in the gameplay state packet. It is not the source domain event queue. Domain event vocabulary and event creation live outside state packet projection.
