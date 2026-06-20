# Game Export Devtools Seams

Parent index: [Server](./!INDEX.md)

## Purpose

This document describes the server-side game export seams used by devtools.

It covers the `internal/game/export_devtools*.go` bridge files, the authority boundary between `internal/devtools` and `internal/game`, the gameplay systems those seams touch, and the rules for adding new devtools-facing access without creating parallel debug-only gameplay logic.

## Overview

Server devtools commands are handled outside the normal gameplay packet path, but gameplay-affecting results still belong to the authoritative `game.Game` aggregate.

The current boundary is:

```text
client devtools input
-> debug packet
-> networking inbound devtools route
-> internal/devtools command handler
-> game-owned devtools export seam or existing public game API
-> authoritative game state mutation or read
-> normal state/debug output back to clients
```

`services/game-server/internal/devtools/` owns command classification, command dispatch, command-specific payload interpretation, target-scope handling, logging around debug actions, continuous bullet stream runtime state, and generated debug packet types.

`services/game-server/internal/game/` owns authoritative gameplay state. It exposes narrow methods named `Devtools...` through `export_devtools*.go` files so devtools handlers can request controlled mutations or read debug state without importing devtools into the game package.

The dependency direction is intentional:

```text
internal/devtools imports internal/game
internal/game does not import internal/devtools
```

This keeps devtools as an adapter layer. It can ask the game aggregate to perform debug-oriented operations, but it does not become a second gameplay implementation.

## Debug-only scope

Game export devtools seams are debug-only service APIs. They are not player-facing gameplay features, durable platform APIs, or protocol ownership.

They may expose controlled access for:

* debug status read models
* world and player freeze toggles
* invincibility and infinite-lives toggles
* debug kill behavior
* debug entity spawning
* debug player spawn and respawn
* score and lives counter changes
* clear-bullet and clear-asteroid tools
* continuous bullet stream hooks
* collision-body telemetry reads

They must not:

* move command handling into `internal/game`
* make `internal/game` import `internal/devtools`
* duplicate gameplay damage, respawn, spawn, scoring, collision, or movement rules in devtools-only code
* mutate client state directly
* treat client debug packets as authority
* expose broad map or struct access when a narrow method is enough
* become the normal gameplay API for production feature behavior

When an existing public game-owned method already represents the correct authoritative operation, devtools may use that method directly instead of adding a duplicate `Devtools...` wrapper. The current pickup spawn path uses `Game.SpawnPickup` this way.

## Server authority

The server remains authoritative for all gameplay-affecting devtools behavior.

A debug packet is only a request. The server resolves the current room, current game player, command type, target player or target scope, and command payload before any game state changes.

Current command dispatch starts in:

```text
services/game-server/internal/devtools/handler.go
```

The handler switches on generated debug packet types and delegates to focused handlers for toggles, spawn tools, respawn, counters, clear tools, pickup spawn, and continuous bullet streams.

### Toggle and kill seams

`export_devtools_toggles.go` exposes game-owned operations for:

```text
DevtoolsWorldFrozen
DevtoolsSetWorldFrozen
DevtoolsToggleFreezeWorld
DevtoolsToggleFreezeAsteroids
DevtoolsToggleFreezeBullets
DevtoolsToggleFreezeSpawning
DevtoolsToggleFreezeCollisions
DevtoolsPlayerInvincible
DevtoolsSetPlayerInvincible
DevtoolsInfiniteLives
DevtoolsSetInfiniteLives
DevtoolsPlayerFrozen
DevtoolsSetPlayerFrozen
DevtoolsKillPlayer
```

World freeze commands mutate `WorldSimulationOptions`. Granular freeze targets can affect asteroid movement, bullet movement, asteroid spawning, or collisions independently.

Invincibility is stored on player/session damage options. Infinite lives is stored on the player session life options. Player freeze is stored on session suspension state and clears active input when enabled.

Debug kill does not delete a player directly. It builds a debug damage request, resolves damage through the damage package, updates health and shields, and then applies the normal fatal-player damage path when the result is fatal.

### Spawn seams

`export_devtools_spawn.go` exposes controlled access to spawn-related game behavior:

```text
DevtoolsRandomUnitVector
DevtoolsNextBulletID
DevtoolsAddBullet
DevtoolsSpawnBullet
DevtoolsRandomAsteroidSpeed
DevtoolsApplyAsteroidSpawnPlan
```

Debug bullet spawn routes through the game-owned debug bullet construction path.

Debug asteroid spawn is split. The devtools handler builds a `spawning.AsteroidSpawnPlan` from command input, normalized position, selected direction, random speed, size, and debug variant selection. The game export seam applies the plan through the game-owned asteroid spawn path.

Pickup spawn currently uses:

```text
Game.SpawnPickup
```

That method validates pickup type through pickup definitions, allocates a game-owned pickup ID, and stores the pickup in the authoritative runtime pickup map.

### Player spawn and respawn seams

`export_devtools_player_spawn.go` exposes controlled player spawn support:

```text
DevtoolsEnsurePlayerSession
DevtoolsSpawnPlayerShip
DevtoolsPlayerIDOccupied
DevtoolsReservePlayerID
DevtoolsTargetPlayerIDs
```

Devtools player spawning can create a session, create a ship, reserve or allocate a debug player ID, and ensure a camera view exists for the spawned player.

`export_devtools_respawn.go` exposes controlled respawn support:

```text
DevtoolsSafeRespawnPosition
DevtoolsForceRespawnPlayer
```

Debug respawn uses the same player-session and ship creation model as gameplay respawn. The handler rejects active players before applying a force respawn, so respawn tools do not overwrite already-active ships.

### Counter seams

`export_devtools_player_counters.go` delegates score and lives tools to the shared player counter seam:

```text
DevtoolsSetPlayerScore
DevtoolsAddPlayerScore
DevtoolsSetPlayerLives
DevtoolsAddPlayerLives
```

Those methods call the normal game counter functions. Counter values are clamped by the shared counter implementation and exported through state packets from session-owned counter state.

### Clear entity seams

`export_devtools_clear_entities.go` exposes:

```text
DevtoolsClearBullets
DevtoolsClearAsteroids
```

These methods mutate authoritative server maps. Clients observe the results through normal state updates after the server removes the entities.

Continuous debug bullet stream runtime state is separate from the projectile map. Clear-bullet command handling belongs to devtools command behavior; the game export seam only clears current authoritative projectile entities.

### Continuous bullet stream seams

`export_devtools_streams.go` exposes:

```text
DevtoolsBulletsCanMove
DevtoolsSpawnDebugBullet
DevtoolsRegisterSimulationStepObserver
```

Continuous bullet stream state is owned by:

```text
services/game-server/internal/devtools/streamruntime/
```

The game aggregate does not own stream state. Instead, devtools registers a simulation step observer on the target game instance. On each simulation step, the observer asks the stream runtime to advance streams and uses the game-owned debug bullet adapter to spawn bullets when cadence allows.

This keeps cadence server-paced while avoiding stream state in the core game aggregate.

### Collision telemetry seam

`export_devtools_collision_telemetry.go` exposes a read model for server collision bodies:

```text
DevtoolsCollisionBodies
DevtoolsCollisionBody
DevtoolsCollisionPoint
```

The helper reads player, asteroid, bullet, and pickup collision bodies from the authoritative runtime state and converts outline points into JSON-shaped debug telemetry records.

This is a read-only devtools seam. It does not mutate gameplay and does not move hitbox drawing into the server. Client devtools remains responsible for presentation.

## Client presentation

Client devtools presentation is separate from server authority.

The client can send debug command packets and render debug outputs, but it does not apply authoritative gameplay effects locally. Visible confirmation comes from server state, debug status packets, debug shape catalog packets, entity sync, or the absence/presence of entities after the authoritative server update.

The server-facing outputs tied to these seams include:

```text
normal gameplay state packets
debug_status packets
debug_shape_catalog packets
server logs
```

`debug_status` reflects server-owned devtools state such as invincibility, infinite lives, world freeze, granular freeze flags, and player freeze.

`debug_shape_catalog` provides shape definitions for client-side hitbox presentation. Shape catalog output is built from the physics collision shape catalog, not from client-authored gameplay state.

Collision-body telemetry is exposed as a game-owned read seam. It should remain a diagnostic data surface; drawing, overlay lifecycle, and toggle UI belong to client devtools.

## Commands and controls

Current generated debug command packet types include:

```text
toggle_debug_invincible
toggle_debug_infinite_lives
toggle_debug_freeze_world
toggle_debug_freeze_player
debug_kill_player
debug_spawn_entity
debug_spawn_pickup
debug_begin_continuous_bullet_stream
debug_respawn_player
debug_set_score
debug_add_score
debug_set_lives
debug_add_lives
debug_clear_bullets
debug_clear_asteroids
```

The server command path is:

```text
networking read loop
-> DecodeClientPacketEnvelope
-> inbound.RouteClientPacket
-> inbound devtools packet classifier
-> packetcodec.Decode into devtools.DebugCommand
-> devtools.HandleCommand
-> command-specific handler
-> game export seam or existing game API
```

Devtools command packets are routed before normal gameplay packet decoding. They do not reach `Game.HandlePacket`.

Player-targeted commands use `target_player_id` or `target_scope`. The current all-player scope value is:

```text
all_players
```

All-player toggle behavior for invincibility, infinite lives, and player freeze uses set-style semantics:

```text
if any eligible target is inactive -> enable all eligible targets
if every eligible target is active -> disable all eligible targets
```

Respawn all-player behavior still applies normal respawn eligibility guards per target. Active players are ignored.

World freeze is room/global and does not use a player selector. Granular freeze commands use `freeze_target` values handled by the devtools toggle handler:

```text
all
asteroids
bullets
spawning
spawns
collisions
```

Unknown freeze targets are logged and ignored without changing freeze flags.

## Telemetry

Devtools telemetry here means diagnostic server state, not analytics.

The main status read path is:

```text
devtools.StatusFor
-> game.DevtoolsStatusFor
-> devtools.DebugStatus
```

`DevtoolsStatusFor` reads:

```text
DamageOptions.Invincible
LifeOptions.InfiniteLives
WorldSimulationOptions
Suspension.DevFrozen
```

The outbound debug status path is:

```text
websocket write tick
-> outbound.CanSendDebugStatus
-> outbound.BuildDebugStatusResponse
-> devtools.StatusFor
-> devtools.StatusesForAllPlayers
-> packetcodec.Encode
-> websocket write
```

Debug status is sent periodically while a room has a game instance and is in game or game over. The write loop currently sends it every eight gameplay ticks.

The debug shape catalog path is:

```text
websocket write tick
-> outbound.CanSendDebugShapeCatalog
-> outbound.BuildDebugShapeCatalogResponse
-> physics.LoadCollisionShapeCatalog
-> devtools.BuildShapeCatalog
-> packetcodec.Encode
-> websocket write
```

The shape catalog is sent once per room ID in the current write loop.

Command handlers also log important devtools actions through the game logger, including toggle changes, spawn results, respawn requests/results, ignored commands, and continuous bullet stream creation.

## Build and runtime gates

Server devtools include build-tag gate helpers:

```text
services/game-server/internal/devtools/enabled_default.go
services/game-server/internal/devtools/enabled_nodevtools.go
services/game-server/internal/devtools/disabled.go
```

Default builds return `true` from `devtools.Enabled()`. Builds with the `nodevtools` tag return `false`.

Current outbound debug status and debug shape catalog sending checks `devtools.Enabled()` before sending debug output.

The devtools package also exposes:

```text
ShouldHandleCommand(packetType string) bool
```

That helper combines command-type classification with `devtools.Enabled()` and has default and `nodevtools` tests.

The current inbound router classifies devtools packets in:

```text
services/game-server/internal/networking/inbound/devtools.go
```

It uses route-local packet-type classifiers for simple commands, placement commands, and remaining devtools commands before decoding normal gameplay packets. That route also requires a current room and a non-empty current game player ID before dispatching a command to `devtools.HandleCommand`.

When changing devtools gates, keep these surfaces aligned:

```text
devtools.Enabled
devtools.ShouldHandleCommand
networking inbound devtools classification
networking outbound debug status/catalog sending
```

Runtime gates include:

```text
current room must exist
current game player ID must exist
command type must be recognized by the inbound devtools route
command payload must decode into devtools.DebugCommand
handler-specific target and payload checks must pass
game export seam or public game method must accept the operation
```

## Locking and mutation model

Game export devtools seams operate on `game.Game` aggregate-owned state, so each seam must respect the aggregate’s synchronization model.

Current export methods are mixed:

```text
DevtoolsClearBullets and DevtoolsClearAsteroids lock the game aggregate.
DevtoolsCollisionBodies locks the game aggregate.
Game.SpawnPickup locks the game aggregate.
Player counter exports delegate to public counter methods that lock the game aggregate.
Several toggle, spawn, respawn, and stream adapter methods delegate directly to game-owned fields or helpers without adding their own local lock.
```

When adding or changing a seam, decide explicitly whether the called game method already owns locking or whether the export method must lock before touching aggregate state. Do not assume a method is safe only because it is in an `export_devtools*.go` file.

Simulation step observers run from inside `Game.Step`. Observer callbacks should stay narrow and should route gameplay effects through game-owned adapters.

## Relationship to real gameplay systems

Game export devtools seams exist to expose real gameplay systems to debug tooling. They should not create replacement systems.

Current examples:

* Debug kill uses the damage resolver and fatal-player damage path.
* Invincibility changes damage options consumed by collision/damage behavior.
* Infinite lives changes session life options consumed by death/lives behavior.
* Player freeze changes suspension state consumed by movement, input, shooting, and collision capability checks.
* World freeze changes `WorldSimulationOptions` consumed by simulation phase gates.
* Score and lives commands use shared player counter mutation.
* Debug asteroid spawn applies a normal asteroid spawn plan through the game aggregate.
* Debug bullet spawn uses the game-owned debug bullet spawn helper.
* Debug pickup spawn uses the game-owned pickup spawn API.
* Debug respawn creates a new ship from the player session and updates camera view state.
* Collision telemetry reads server collision bodies from runtime entities and collision shapes.

The rule is:

```text
devtools may choose when to request a debug action
game-owned systems decide what that action actually does
```

## Code map

Primary game export files:

```text
services/game-server/internal/game/export_devtools.go
services/game-server/internal/game/export_devtools_status.go
services/game-server/internal/game/export_devtools_toggles.go
services/game-server/internal/game/export_devtools_spawn.go
services/game-server/internal/game/export_devtools_respawn.go
services/game-server/internal/game/export_devtools_player_spawn.go
services/game-server/internal/game/export_devtools_player_counters.go
services/game-server/internal/game/export_devtools_clear_entities.go
services/game-server/internal/game/export_devtools_streams.go
services/game-server/internal/game/export_devtools_collision_telemetry.go
```

Server devtools command files:

```text
services/game-server/internal/devtools/handler.go
services/game-server/internal/devtools/command_types.go
services/game-server/internal/devtools/packets_generated.go
services/game-server/internal/devtools/toggles.go
services/game-server/internal/devtools/spawn_entity.go
services/game-server/internal/devtools/spawn_asteroid.go
services/game-server/internal/devtools/spawn_bullet.go
services/game-server/internal/devtools/spawn_pickup.go
services/game-server/internal/devtools/spawn_player.go
services/game-server/internal/devtools/respawn_player.go
services/game-server/internal/devtools/respawn_handler.go
services/game-server/internal/devtools/player_counters.go
services/game-server/internal/devtools/clear_entities.go
services/game-server/internal/devtools/continuous_bullet_stream.go
services/game-server/internal/devtools/status.go
services/game-server/internal/devtools/target_player_ids.go
services/game-server/internal/devtools/target_scopes.go
services/game-server/internal/devtools/player_ids.go
services/game-server/internal/devtools/placement_requests.go
```

Build and command gates:

```text
services/game-server/internal/devtools/disabled.go
services/game-server/internal/devtools/enabled_default.go
services/game-server/internal/devtools/enabled_nodevtools.go
services/game-server/internal/networking/inbound/devtools.go
services/game-server/internal/networking/inbound/router.go
services/game-server/internal/networking/client_packet_router.go
```

Outbound debug presentation:

```text
services/game-server/internal/networking/websocket_write.go
services/game-server/internal/networking/outbound/debug_status_presentation.go
services/game-server/internal/networking/outbound/debug_shape_catalog_presentation.go
```

Continuous stream runtime:

```text
services/game-server/internal/devtools/streamruntime/runtime.go
services/game-server/internal/devtools/streamruntime/simulation.go
services/game-server/internal/devtools/streamruntime/continuous_bullet_streams.go
```

Gameplay systems touched by export seams:

```text
services/game-server/internal/game/game.go
services/game-server/internal/game/world_simulation_options.go
services/game-server/internal/game/player_counters.go
services/game-server/internal/game/pickups.go
services/game-server/internal/game/spawning.go
services/game-server/internal/game/simulation.go
services/game-server/internal/game/session.go
services/game-server/internal/game/player_session_state.go
services/game-server/internal/game/runtime/damage_options.go
services/game-server/internal/game/runtime/life_options.go
services/game-server/internal/game/runtime/suspension.go
services/game-server/internal/game/damage/
services/game-server/internal/game/spawning/
services/game-server/internal/game/physics/
services/game-server/internal/game/entities/pickups/
```

Generated/source files:

```text
shared/packets/debug.toml
shared/packets/outputs.toml
services/game-server/internal/devtools/packets_generated.go
client/scripts/generated/networking/packets/packets.gd
```

Important non-ownership boundaries:

```text
client/
services/game-server/internal/networking/
services/game-server/internal/rooms/
services/game-server/internal/protocol/packetcodec/
services/player-data/
services/api-server/
```

## Tests

Focused game export seam tests include:

```text
services/game-server/internal/game/export_devtools_player_spawn_test.go
services/game-server/internal/game/export_devtools_respawn_test.go
services/game-server/internal/game/export_devtools_streams_test.go
services/game-server/internal/game/export_devtools_collision_telemetry_test.go
```

Relevant devtools package tests include:

```text
services/game-server/internal/devtools/command_types_test.go
services/game-server/internal/devtools/enabled_default_test.go
services/game-server/internal/devtools/disabled_test.go
services/game-server/internal/devtools/toggles_test.go
services/game-server/internal/devtools/player_counters_test.go
services/game-server/internal/devtools/clear_entities_test.go
services/game-server/internal/devtools/target_player_ids_test.go
services/game-server/internal/devtools/shape_catalog_test.go
services/game-server/internal/devtools/streamruntime/continuous_bullet_streams_test.go
services/game-server/internal/devtools/streamruntime/runtime_test.go
```

Relevant integration-style game tests include:

```text
services/game-server/tests/game/devtools_test.go
services/game-server/tests/game/continuous_bullet_stream_test.go
services/game-server/tests/game/player_counters_test.go
services/game-server/tests/game/respawn_test.go
services/game-server/tests/game/collision_test.go
services/game-server/tests/game/spawning_test.go
services/game-server/tests/game/pickups_test.go
```

Current coverage verifies:

* debug toggles can enable and disable server-owned flags
* all-player toggle scope applies to eligible player sets
* debug status reflects active debug state
* granular world freeze flags affect only their intended simulation gates
* invincible players do not die from asteroid collision
* infinite-lives players die without losing a life
* debug kill routes through normal death/despawn/event consequences
* debug kill can target another active player
* score and lives tools mutate session-owned counters
* all-player counter scope applies to every target player
* respawn all-player scope ignores active players and respawns eligible players
* debug pickup spawn creates valid pickups and rejects unknown pickup types
* debug player spawn and respawn create camera views with dummy config
* continuous stream seams can query bullet movement and spawn valid debug bullets
* collision telemetry uses server collision bodies and marshals lower-case JSON keys
* default and `nodevtools` gate helpers return expected values in their respective builds

Useful verification command:

```bash
cd services/game-server
go test -buildvcs=false ./...
```

Focused verification for devtools and export seams:

```bash
cd services/game-server
go test -buildvcs=false ./internal/devtools/... ./internal/game ./tests/game
```

Focused `nodevtools` gate verification:

```bash
cd services/game-server
go test -buildvcs=false -tags nodevtools ./internal/devtools/...
```

Run packet checks when debug packet source or generated output changes:

```bash
data-sync -check -packets -go -gds
```

## Related docs

* [Server Devtools](./!INDEX.md)
* [Devtools](../!INDEX.md)
* [Client Packet Routing And Devtools Input](../client/packet-routing-and-devtools-input.md)
* [Target And Placement Debugging](../client/target-and-placement-debugging.md)
* [Debug Status And Target Readmodels](../client/debug-status-and-target-readmodels.md)
* [Hitbox Overlays](../client/hitbox-overlays.md)
* [Game Server](../../services/game-server/!INDEX.md)
* [Game Aggregate](../../services/game-server/simulation/runtime/game-aggregate.md)
* [Simulation Loop And Phase Order](../../services/game-server/simulation/runtime/simulation-loop-and-phase-order.md)
* [Player Counters](../../services/game-server/simulation/players/player-counters.md)
* [Player Respawn](../../services/game-server/simulation/players/player-respawn.md)
* [Player Pause And Suspension](../../services/game-server/simulation/players/player-pause-and-suspension.md)
* [Collision Shapes](../../services/game-server/simulation/world/collision-shapes.md)
* [Packet Schemas](../../data/packet-schemas.md)

## Notes

Legacy devtools docs correctly identified the core boundary: devtools are client-triggered but server-authoritative, `internal/devtools` owns command handling, and `internal/game/export_devtools*.go` exposes only controlled game-owned adapters.

The current implementation has both `devtools.ShouldHandleCommand` gate helpers and route-local inbound devtools packet classifiers. Keep those paths synchronized when changing command availability or build-gate behavior.

This document intentionally focuses on the game export seams. Individual command behavior, client controls, packet schema ownership, and overlay presentation belong in their own devtools, protocol, data, or client docs.
