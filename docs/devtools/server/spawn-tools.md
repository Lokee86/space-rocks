# Spawn Tools

Parent index: [Server](./!INDEX.md)

## Purpose

This document describes the server-side devtools spawn tooling.

It covers the debug command surface that spawns players, asteroids, bullets, and pickups into an active game instance. It documents server authority, command behavior, telemetry, build/runtime gates, implementation paths, and verification coverage.

## Overview

Spawn tools are development-only mutation commands routed through the game server devtools surface.

The server receives devtools packets from the client, decodes them as `devtools.DebugCommand`, and applies the requested spawn through server-owned game seams. The client may request a position, direction, entity type, pickup type, or target player id, but the server owns the actual mutation and the resulting authoritative state.

The current spawn packet surface is:

```text
debug_spawn_entity
debug_spawn_pickup
```

`debug_spawn_entity` handles these entity types:

```text
player
asteroid
bullet
```

`debug_spawn_pickup` handles pickup entities separately because pickup spawning needs `pickup_type` instead of the generic `entity_type` field.

Spawn tools are not a production gameplay spawning API. They are debug commands that intentionally converge on real server mutation paths where practical.

## Debug-only scope

Spawn tools are devtools-only behavior.

They are used for:

* creating a debug player ship at a requested position
* creating an asteroid at a requested position
* creating a bullet at a requested position
* creating a pickup at a requested position
* testing world sync, state projection, collision bodies, pickup presentation, target readouts, and placement flows
* reproducing entity-specific gameplay or rendering problems without waiting for normal gameplay conditions

They are not used for:

* normal timed asteroid spawning
* asteroid fragment spawning after destruction
* drop-table pickup spawning
* normal weapon fire
* player join flow
* normal respawn flow
* room membership authority
* match lifecycle authority
* durable player-data mutation

Respawn, clear-entity, and continuous bullet stream tooling are adjacent devtools features with separate command behavior. Spawn tools may share packet fields and placement flows with them, but this document covers only one-shot spawn commands.

## Server authority

The game server remains authoritative for spawned runtime state.

The client may send:

```text
entity_type
pickup_type
x
y
has_direction
direction_x
direction_y
target_player_id
```

The server decides:

* whether the packet type is a spawn command
* whether the command can be decoded
* whether the session has an active room and game player id
* which spawn helper handles the requested entity type
* whether a requested target player id is valid
* whether a new debug player id can be allocated
* how positions are normalized into toroidal world space
* how fallback directions are selected
* how runtime ids are allocated
* how the new entity enters the authoritative entity store
* what world lane records later project to clients

The spawn command path does not give the client direct write access to server entity maps. Client placement requests route through server command handlers and game-owned mutation helpers.

## Client presentation

The client owns spawn UI and placement input, not spawn authority.

Client-side devtools can start placement actions from hotkeys or the devtools window. The client converts mouse placement into server coordinates and sends the appropriate debug packet.

Current client placement actions include:

```text
spawn_player
spawn_asteroid
spawn_bullet
spawn_pickup
```

The server does not send a special spawn-confirmation packet. Confirmation is observed through normal authoritative readback:

```text
world lane ship records
world lane asteroid records
world lane bullet records
world lane pickup records
session lane player records
debug_status packets
event_batch where applicable
```

Spawned entities appear on the client only after the server mutates runtime state and lane-native readback projects that state.

## Command routing

Spawn packets are routed through the inbound devtools packet path before normal gameplay packet handling.

The current high-level route is:

```text
websocket message
-> client packet envelope decode
-> inbound.RouteClientPacket
-> inbound.HandlePlacementDevtoolsPacket
-> packetcodec.Decode into devtools.DebugCommand
-> devtools.HandleCommand
-> spawn command handler
-> game-owned mutation helper
-> lane-native realtime projection
```

`HandlePlacementDevtoolsPacket` recognizes:

```text
debug_spawn_entity
debug_spawn_pickup
```

The handler ignores placement devtools packets when the session has no current room or no current game player id. Decode failures are logged through network logging and stop command execution.

`devtools.HandleCommand` dispatches spawn commands to:

```text
debug_spawn_entity -> handleDebugSpawnEntity
debug_spawn_pickup -> handleDebugSpawnPickup
```

`handleDebugSpawnEntity` then dispatches by `entity_type`:

```text
player   -> applyDebugSpawnPlayer
bullet   -> applyDebugSpawnBullet
asteroid -> applyDebugSpawnAsteroid
```

Unknown generic spawn entity types are logged as unimplemented and otherwise consumed as handled devtools commands.

## Packet fields

The shared debug command struct contains the spawn fields used by both generic entity spawn and pickup spawn:

```text
type
target_player_id
entity_type
pickup_type
x
y
has_direction
direction_x
direction_y
```

`debug_spawn_entity` uses:

```text
type = debug_spawn_entity
entity_type
x
y
has_direction
direction_x
direction_y
target_player_id
```

`debug_spawn_pickup` uses:

```text
type = debug_spawn_pickup
pickup_type
x
y
```

`target_player_id` is meaningful for debug player spawn. It is not the generic gameplay target model.

`has_direction`, `direction_x`, and `direction_y` are meaningful for asteroid and bullet spawn. If direction is not supplied or is a zero vector, the server uses a random unit-vector fallback.

## Position and direction handling

Spawn positions are normalized through toroidal world space before most spawned entities are created.

Generic entity spawn requests use:

```text
SpawnEntityRequest.Position
-> space.NormalizePosition
```

Pickup spawn uses:

```text
DebugCommand x/y
-> space.NormalizePosition
```

Direction handling uses:

```text
SpawnEntityRequest.DirectionOr(fallback)
```

The direction resolver:

* uses the requested direction only when `has_direction` is true
* rejects a zero requested direction
* normalizes accepted requested direction
* otherwise returns a normalized fallback direction

Fallback directions come from the game spawner through the devtools export seam:

```text
Game.DevtoolsRandomUnitVector
```

## Player spawn behavior

Debug player spawn creates or replaces a server-side player session and active ship for a debug player id.

The current player spawn flow is:

```text
handleDebugSpawnEntity
-> applyDebugSpawnPlayer
-> resolveDebugSpawnPlayerID
-> DevtoolsEnsurePlayerSession
-> DevtoolsSpawnPlayerShip
-> ensureDevtoolsPlayerCameraView
```

If `target_player_id` is supplied, the server normalizes it to the `player-N` format and reserves that id through the game-owned id reservation seam. Valid examples include:

```text
player-1
Player-1
```

The normalized stored form is:

```text
player-1
```

Invalid ids are rejected.

If `target_player_id` is not supplied, the server allocates the first available debug gameplay player id from the bounded player id range. The allocation checks both player sessions and active player ships before reserving a candidate.

Player spawn creates a player session through:

```text
Game.DevtoolsEnsurePlayerSession
```

It then creates the active ship through:

```text
Game.DevtoolsSpawnPlayerShip
```

The devtools player ship path resets respawn cooldown, creates a new ship from the session, stores it in `game.entities.Players`, and ensures a camera view exists. Debug-spawned players use the dummy camera config from the devtools player-camera helper.

Debug player spawn does not create a websocket session, authenticate a user, or update durable player data. It creates runtime game state for development inspection.

## Asteroid spawn behavior

Debug asteroid spawn builds an asteroid spawn plan, then applies it through the same game-owned asteroid application seam used by normal asteroid creation.

The current asteroid flow is:

```text
handleDebugSpawnEntity
-> applyDebugSpawnAsteroid
-> buildDebugAsteroidSpawnPlan
-> Game.DevtoolsApplyAsteroidSpawnPlan
-> Game.applyAsteroidSpawn
```

The debug asteroid spawn plan uses:

```text
EntityType = asteroid
Reason     = debug_asteroid
Position   = normalized requested position
Velocity   = requested or fallback direction * random asteroid speed
Size       = random integer from 1 through 4
Variant    = weighted debug-spawn variant index
```

Variant selection uses the asteroid variant catalog helper:

```go
asteroids.RandomDebugSpawnVariantIndex()
```

The devtools path must not reintroduce raw random variant pools. Debug asteroid spawn should remain aligned with the shared asteroid variant contract.

Asteroid id allocation and runtime entity insertion are owned by root game code:

```text
Game.applyAsteroidSpawn
```

The spawned asteroid is stored in:

```text
game.entities.Asteroids
```

## Bullet spawn behavior

Debug bullet spawn creates a basic cannon projectile owned by the requesting game player id.

The current bullet flow is:

```text
handleDebugSpawnEntity
-> applyDebugSpawnBullet
-> Game.DevtoolsSpawnBullet
-> Game.spawnDebugBullet
```

The bullet owner is the `playerID` supplied by the active websocket/game session, not a client-supplied owner field.

The bullet spawn path requires:

```text
target game is not nil
owner player id is not empty
direction normalizes to a nonzero vector
```

The bullet spawn implementation:

* normalizes the requested position
* normalizes the requested or fallback direction
* derives rotation from direction
* uses `constants.BasicCannonProjectileSpeed`
* uses `constants.BasicCannonProjectileLifetime`
* allocates the next bullet id through the game spawner
* creates `runtime.NewBullet`
* stores it in `game.entities.Projectiles`

Debug bullet spawn does not use weapon slot cooldowns, ammo state, loadout state, or normal weapon fire policy. It is a direct debug projectile creation path that still mutates authoritative server runtime state.

## Pickup spawn behavior

Debug pickup spawn creates an authoritative pickup through the game pickup spawn API.

The current pickup flow is:

```text
handleDebugSpawnPickup
-> Game.SpawnPickup
-> Game.spawnPickupLocked
-> pickups.DefinitionFor
-> game.entities.Pickups
```

Pickup spawn uses:

```text
pickup_type
x
y
```

The server converts `pickup_type` into a `pickups.PickupType` and validates it against server pickup definitions. Unknown pickup types are rejected by the pickup spawn API and logged as ignored debug spawns.

A successful pickup spawn:

* allocates the next pickup id
* creates the pickup entity from the definition
* initializes health, age, and lifespan
* stores the pickup in `game.entities.Pickups`
* makes it available for world lane readback and pickup lifecycle handling

Debug pickup spawn does not record a `pickup_dropped` gameplay event. Drop-table integration owns that event for normal asteroid-drop spawns.

## Telemetry

Spawn tools use structured game logs.

Generic entity spawn logs include:

```text
debug player spawn ignored
debug player spawned
debug bullet spawn ignored
debug bullet spawned
debug asteroid spawn ignored
debug asteroid spawned
debug spawn entity not implemented for entity type
```

Pickup spawn logs include:

```text
debug pickup spawn ignored
debug pickup spawned
```

Logged fields may include:

```text
player_id
target_player_id
spawned_player_id
owner_player_id
bullet_id
asteroid_id
pickup_id
pickup_type
entity_type
x
y
has_direction
has_target_player_id
```

Inbound decode failures are logged through network logging as websocket devtools command decode failures.

There is no separate spawn telemetry packet. Spawn outcomes are confirmed by normal logs and by authoritative state readback.

## Build/runtime gates

Server devtools are build-gated by the devtools package.

Default builds use:

```text
services/game-server/internal/devtools/enabled_default.go
```

`nodevtools` builds use:

```text
services/game-server/internal/devtools/enabled_nodevtools.go
```

The package-level command gate is:

```text
devtools.ShouldHandleCommand(packetType)
```

It combines command type recognition with the current devtools build flag.

Spawn command files do not own this gate directly. They are implementation handlers behind devtools packet routing and command dispatch.

Runtime gates also apply:

* inbound routing requires a recognized spawn packet type
* command decode must succeed
* the session must have a current room
* the session must have a current game player id
* generic entity spawn must name a supported entity type
* player spawn must resolve or allocate a valid player id
* bullet spawn must have a nonempty owner id and nonzero direction
* pickup spawn must use a known pickup type
* asteroid spawn requires a live game target and successful game-owned spawn application

## Relationship to real gameplay implementation areas

Spawn tools must not become parallel gameplay implementations.

The intended relationship is:

```text
devtools command
-> debug adapter
-> game-owned mutation seam
-> normal runtime entity storage
-> lane-native realtime projection
```

Current examples:

* Debug asteroid spawn builds a debug asteroid spawn plan, then applies it through `Game.applyAsteroidSpawn`.
* Debug pickup spawn calls `Game.SpawnPickup`, which delegates to the normal pickup creation path.
* Debug player spawn uses game-owned session, ship, and camera-view helpers.
* Debug bullet spawn uses a game-owned debug bullet helper and stores the bullet in the authoritative projectile map.

The devtools package may shape debug requests, validate debug-only fields, and log debug actions. It should not own scoring, normal spawn scheduling, pickup drop policy, room membership, durable profile state, or production gameplay lifecycle rules.

## Code map

Primary server devtools spawn files:

```text
services/game-server/internal/devtools/handler.go
services/game-server/internal/devtools/command_types.go
services/game-server/internal/devtools/packets_generated.go
services/game-server/internal/devtools/placement_requests.go
services/game-server/internal/devtools/spawn_entity.go
services/game-server/internal/devtools/spawn_player.go
services/game-server/internal/devtools/spawn_asteroid.go
services/game-server/internal/devtools/spawn_bullet.go
services/game-server/internal/devtools/spawn_pickup.go
services/game-server/internal/devtools/player_ids.go
services/game-server/internal/devtools/player_camera.go
```

Build gate files:

```text
services/game-server/internal/devtools/disabled.go
services/game-server/internal/devtools/enabled_default.go
services/game-server/internal/devtools/enabled_nodevtools.go
```

Inbound routing files:

```text
services/game-server/internal/networking/client_packet_router.go
services/game-server/internal/networking/inbound/router.go
services/game-server/internal/networking/inbound/devtools.go
services/game-server/internal/networking/inbound/client_packet_envelope.go
```

Game-owned spawn seams:

```text
services/game-server/internal/game/export_devtools_spawn.go
services/game-server/internal/game/export_devtools_player_spawn.go
services/game-server/internal/game/export_devtools_streams.go
services/game-server/internal/game/spawning.go
services/game-server/internal/game/pickups.go
```

Adjacent game implementation paths:

```text
services/game-server/internal/game/spawning/spawner.go
services/game-server/internal/game/asteroids/variants.go
services/game-server/internal/game/runtime/asteroid.go
services/game-server/internal/game/runtime/bullet.go
services/game-server/internal/game/entities/pickups/
services/game-server/internal/game/runtime/state.go
services/game-server/internal/game/state_packet.go
```

Packet source and generated outputs:

```text
shared/packets/debug.toml
shared/packets/outputs.toml
services/game-server/internal/devtools/packets_generated.go
client/scripts/generated/networking/packets/packets.gd
client/scripts/devtools/dev_spawn_packet_builder.gd
```

Related client placement paths:

```text
client/scripts/devtools/dev_spawn_packet_builder.gd
client/scripts/devtools/context/devtools_placement_context.gd
client/scripts/gameplay/devtools/debug_click_placement_flow.gd
client/scripts/gameplay/devtools/debug_mouse_world_position.gd
client/scripts/devtools/dev_connection_service.gd
```

Relevant tests:

```text
services/game-server/internal/devtools/command_types_test.go
services/game-server/internal/devtools/enabled_default_test.go
services/game-server/internal/devtools/disabled_test.go
services/game-server/internal/devtools/target_player_ids_test.go
services/game-server/internal/game/export_devtools_player_spawn_test.go
services/game-server/internal/game/export_devtools_streams_test.go
services/game-server/tests/networking/inbound_devtools_test.go
services/game-server/internal/game/entities/pickups/definitions_test.go
services/game-server/internal/game/pickup_drops_test.go
services/game-server/internal/game/asteroids/variants_test.go
```

Important non-ownership boundaries:

```text
services/game-server/internal/rooms/
services/game-server/internal/game/rules/
services/game-server/internal/game/weapons/
services/game-server/internal/game/drops/
services/game-server/internal/game/scoring/
services/game-server/internal/game/pickups/
services/player-data/
client/
tools/data_sync/
```

## Tests and verification

Current direct and adjacent coverage verifies:

* devtools command type recognition includes spawn command packet types
* devtools build flags are enabled by default
* `nodevtools` builds disable package-level command handling
* inbound placement devtools routing accepts `debug_spawn_entity` and `debug_spawn_pickup`
* devtools player spawn creates sessions, ships, and camera views
* devtools target player id readback includes session-only and active ship targets
* devtools bullet spawn creates a projectile with the expected owner id and origin
* pickup definitions reject unknown pickup types
* asteroid debug spawn variants are selected from the debug variant catalog
* generated debug packet constants include spawn packet types

Useful verification commands from `services/game-server`:

```bash
go test -buildvcs=false ./internal/devtools
go test -buildvcs=false ./internal/game -run 'Devtools|Spawn|Pickup|Asteroid'
go test -buildvcs=false ./tests/networking -run Devtools
go test -buildvcs=false ./...
```

Useful data verification commands from the repository root:

```bash
data-sync -check -packets -go -gds
data-sync -check -constants -go
```

## Related docs

* [Server Devtools](./!INDEX.md)
* [Client Target And Placement Debugging](../client/target-and-placement-debugging.md)
* [Client Packet Routing And Devtools Input](../client/packet-routing-and-devtools-input.md)
* [Game Server](../../services/game-server/!INDEX.md)
* [Inbound Packet Routing](../../services/game-server/networking/inbound-packet-routing.md)
* [Asteroid Spawning And Variants](../../services/game-server/simulation/world/asteroid-spawning-and-variants.md)
* [Pickup Entity Lifecycle](../../services/game-server/simulation/pickups/pickup-entity-lifecycle.md)
* [Weapons And Projectile Fire](../../services/game-server/simulation/combat/weapons-and-projectile-fire.md)
* [Player Session State](../../services/game-server/simulation/players/player-session-state.md)
* [Realtime Websocket Protocol](../../protocol/realtime-websocket-protocol.md)
* [Packet Schemas](../../data/packet-schemas.md)

## Notes

`debug_spawn_entity` and `debug_spawn_pickup` are placement devtools packets on the server inbound path.

Debug bullet spawn intentionally bypasses normal weapon fire policy. That makes it useful for projectile diagnostics, but it should not be treated as proof that normal equipped weapon fire, ammo, cooldown, or loadout behavior is working.

Debug pickup spawn creates a real authoritative pickup entity, but it does not represent normal drop-table behavior.

Debug player spawn creates runtime player state for inspection. It does not represent a normal authenticated join path or durable player profile state.
