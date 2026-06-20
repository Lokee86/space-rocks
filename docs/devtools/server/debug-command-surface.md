# Debug Command Surface

Parent index: [Server](./!INDEX.md)

## Purpose

This document describes the server-side debug command surface for Space Rocks devtools.

It covers the command packets accepted by the game server, how those packets route into server devtools handlers, which gameplay seams perform authoritative mutation, what diagnostic output the server emits, and which files define the command surface.

## Overview

Server devtools are a debug-only control surface for local development and diagnostics. The client may request debug actions, but the Go game server owns whether those actions are accepted and how they affect authoritative gameplay state.

Current command flow:

```text
client devtools hotkey or window action
-> generated or devtools-built debug packet
-> websocket read path
-> inbound packet envelope decode
-> devtools packet-family routing
-> devtools.DebugCommand decode
-> devtools.HandleCommand
-> game-owned export seam
-> authoritative game state mutation
-> normal gameplay state packet or debug output packet
```

The command surface is not a normal gameplay packet path. Devtools command packets are detected before normal `game.ClientPacket` decoding and do not route through `Game.HandlePacket`.

The server command surface currently covers:

```text
player debug toggles
world/simulation freeze controls
player kill and respawn tools
entity and pickup spawn tools
continuous bullet streams
score and lives mutation tools
clear bullet and asteroid tools
debug status output
debug shape catalog output
```

`debug_status` and `debug_shape_catalog` are outbound diagnostic packets, not inbound command handlers. Inbound mutation commands use `DebugCommand`.

## Debug-only scope

Server devtools may mutate authoritative state for development purposes only.

Allowed debug behavior:

* Toggle player invincibility.
* Toggle infinite lives.
* Toggle world or simulation sub-lane freeze state.
* Toggle player freeze through the normal player suspension state.
* Kill active players through the damage/death path.
* Spawn players, asteroids, bullets, and pickups.
* Start continuous bullet streams.
* Force respawn inactive players.
* Set or add score.
* Set or add lives.
* Clear all bullets.
* Clear all asteroids.
* Emit debug status and shape catalog output for client devtools presentation.

Server devtools must not:

* Become player-facing gameplay features.
* Duplicate gameplay rules in a parallel debug-only implementation.
* Let the client authoritatively mutate gameplay state.
* Route gameplay-affecting changes around game-owned seams.
* Put devtools-only command constants into generated normal game packet ownership.
* Make `internal/game` import `internal/devtools`.

When a command affects gameplay, `services/game-server/internal/devtools` handles command interpretation and then calls a narrow game-owned `Devtools...` export method. The owning gameplay system still performs the actual mutation.

## Server authority

The server command surface is authoritative at three levels.

First, networking decides whether the packet is a devtools command before normal packet handling. The inbound router checks devtools packet families before auth, telemetry, lobby, or gameplay packet handlers.

Second, the devtools package decodes the packet into `DebugCommand` and dispatches by command type through `HandleCommand`.

Third, the game package applies the mutation through game-owned export seams such as:

```text
DevtoolsSetPlayerInvincible
DevtoolsSetInfiniteLives
DevtoolsToggleFreezeWorld
DevtoolsSetPlayerFrozen
DevtoolsKillPlayer
DevtoolsSpawnBullet
DevtoolsApplyAsteroidSpawnPlan
DevtoolsForceRespawnPlayer
DevtoolsSetPlayerScore
DevtoolsAddPlayerScore
DevtoolsClearBullets
DevtoolsClearAsteroids
```

The client does not receive a command-specific acknowledgement packet. Confirmation is observed through normal state projection, debug status output, entity sync, or visible absence/presence of entities after the server applies the change.

If a websocket session has no current room or no current game player ID, the inbound devtools command path consumes the packet and applies no command. This prevents debug command packets from falling through into normal gameplay handling when there is no active game context.

## Client presentation

The client presents this command surface through DevToggle hotkeys, the devtools window, placement tools, overlays, labels, and debug readouts.

The server does not own client presentation. It only receives command packets and emits authoritative state or diagnostic output.

Client-side controls may include local gates such as “gameplay state must exist” or “placement route must be configured,” but those gates are convenience checks only. Server command handlers remain the authority boundary for gameplay-affecting actions.

Player-targeted client controls may send:

```text
target_player_id
target_scope
```

`target_player_id` is a devtools/player-only compatibility field. It is not the normal gameplay targeting model.

`target_scope = "all_players"` requests an all-player operation. The server resolves that scope through current game/session player IDs instead of accepting a fake player ID.

## Command routing

### Inbound router order

The websocket read path first decodes the packet envelope:

```text
services/game-server/internal/networking/websocket_read.go
-> inbound.DecodeClientPacketEnvelope
```

Then `handleClientPacket` builds an inbound session adapter and calls:

```text
inbound.RouteClientPacket
```

The router attempts devtools families before normal packet decode:

```text
HandleSimpleDevtools
HandlePlacementDevtools
HandleRemainingDevtools
DecodePacket
HandleAuth
HandleTelemetry
HandleLobby
HandleGameplay
```

This order keeps debug command packets out of normal `game.ClientPacket` handling.

### Devtools packet families

`services/game-server/internal/networking/inbound/devtools.go` splits command detection into three inbound families.

Simple devtools commands:

```text
toggle_debug_invincible
toggle_debug_infinite_lives
toggle_debug_freeze_world
toggle_debug_freeze_player
debug_kill_player
debug_set_score
debug_add_score
debug_set_lives
debug_add_lives
debug_clear_bullets
debug_clear_asteroids
```

Placement devtools commands:

```text
debug_spawn_entity
debug_spawn_pickup
```

Remaining devtools commands:

```text
debug_begin_continuous_bullet_stream
debug_respawn_player
```

The grouping is a routing detail. It does not define authority or gameplay semantics.

### Handler dispatch

All routed mutation commands decode into:

```go
type DebugCommand struct {
    Type           string  `json:"type"`
    TargetPlayerID string  `json:"target_player_id"`
    TargetScope    string  `json:"target_scope"`
    EntityType     string  `json:"entity_type"`
    PickupType     string  `json:"pickup_type"`
    X              float64 `json:"x"`
    Y              float64 `json:"y"`
    HasDirection   bool    `json:"has_direction"`
    DirectionX     float64 `json:"direction_x"`
    DirectionY     float64 `json:"direction_y"`
    FreezeTarget   string  `json:"freeze_target"`
    Score          int     `json:"score"`
    Amount         int     `json:"amount"`
    Lives          int     `json:"lives"`
}
```

`devtools.HandleCommand` switches on `command.Type` and delegates to focused handlers.

Unknown command types return `false` from `HandleCommand`. The inbound networking path does not serialize that boolean to the client.

## Command surface

### Player toggle commands

| Packet type                   | Target fields                                        | Server behavior                                                                         |
| ----------------------------- | ---------------------------------------------------- | --------------------------------------------------------------------------------------- |
| `toggle_debug_invincible`     | optional `target_player_id`, optional `target_scope` | Toggles or sets `DamageOptions.Invincible` through `DevtoolsSetPlayerInvincible`.       |
| `toggle_debug_infinite_lives` | optional `target_player_id`, optional `target_scope` | Toggles or sets session `LifeOptions.InfiniteLives` through `DevtoolsSetInfiniteLives`. |
| `toggle_debug_freeze_player`  | optional `target_player_id`, optional `target_scope` | Toggles or sets session `Suspension.DevFrozen` through `DevtoolsSetPlayerFrozen`.       |

For single-player targeting, an empty `target_player_id` falls back to the requesting player ID.

For `target_scope = "all_players"`, the server resolves all known target player IDs from player sessions and active players.

All-player toggle behavior for invincibility, infinite lives, and player freeze is set-style:

```text
if any eligible player is inactive -> enable all
if every eligible player is active -> disable all
```

Player freeze uses the same suspension model as gameplay pause. It sets the dev-freeze cause and clears active ship input when enabling freeze.

### World and simulation freeze command

`toggle_debug_freeze_world` toggles world simulation freeze state.

Supported `freeze_target` values:

```text
all
asteroids
bullets
spawning
spawns
collisions
```

If `freeze_target` is empty, the server treats it as `all`.

Effects:

| `freeze_target`        | Server behavior                                                         |
| ---------------------- | ----------------------------------------------------------------------- |
| `all`                  | Toggles whole-world freeze through `DevtoolsToggleFreezeWorld`.         |
| `asteroids`            | Toggles asteroid movement through `DevtoolsToggleFreezeAsteroids`.      |
| `bullets`              | Toggles bullet movement/lifetime through `DevtoolsToggleFreezeBullets`. |
| `spawning` or `spawns` | Toggles asteroid spawning through `DevtoolsToggleFreezeSpawning`.       |
| `collisions`           | Toggles collision passes through `DevtoolsToggleFreezeCollisions`.      |
| unknown value          | Logs and consumes the command without mutation.                         |

The actual simulation gates live in game-owned world simulation options. Devtools only requests toggle changes through exported game methods.

### Kill command

`debug_kill_player` kills active player targets.

Target fields:

```text
target_player_id
target_scope
```

For an all-player request, each resolved player is evaluated individually.

The handler only applies kill to players whose match decision status is `active`. It calls `DevtoolsKillPlayer`, which builds a debug damage request and resolves it through the damage path before applying fatal player damage.

The source player ID is the requesting/current game player ID. The target player ID is the player being killed.

### Spawn commands

`debug_spawn_entity` handles server-authoritative debug entity creation.

Payload fields:

```text
entity_type
x
y
has_direction
direction_x
direction_y
target_player_id
```

Supported `entity_type` values:

| Entity type | Server behavior                                                                                                                                 |
| ----------- | ----------------------------------------------------------------------------------------------------------------------------------------------- |
| `player`    | Creates or replaces a debug player session and active ship at the normalized position.                                                          |
| `asteroid`  | Builds an asteroid spawn plan with debug reason, random size, random debug variant, requested or fallback direction, and random asteroid speed. |
| `bullet`    | Spawns a debug bullet owned by the requesting player at the normalized position with requested or fallback direction.                           |

Spawn positions are normalized into the wrapped game world.

If `has_direction` is false or the requested direction is zero-length, asteroid and bullet commands use a server-generated random unit vector.

Debug player spawn can use `target_player_id` to request a specific debug player ID. The accepted debug player ID format is `player-N` or `Player-N`; the game stores the normalized lowercase form. If no target player ID is supplied, the server allocates the first available debug gameplay player ID.

`debug_spawn_pickup` handles pickup creation separately.

Payload fields:

```text
pickup_type
x
y
```

The server normalizes the requested position, converts `pickup_type` to the gameplay pickup type, and calls `Game.SpawnPickup`. Invalid pickup requests are ignored after debug logging.

### Continuous bullet stream command

`debug_begin_continuous_bullet_stream` starts a persistent debug bullet stream.

Required payload fields:

```text
x
y
has_direction = true
direction_x
direction_y
```

If `has_direction` is false or the direction vector is zero-length, the command is ignored.

When accepted, the handler:

```text
normalizes the stream direction
registers the stream in devtools streamruntime
registers a game simulation step observer if needed
spawns stream bullets during simulation steps while bullets can move
```

Continuous bullet stream runtime state lives in:

```text
services/game-server/internal/devtools/streamruntime/
```

The game package does not own stream runtime state. It only exposes the simulation-step observer hook and debug bullet spawn method used by the stream runtime.

### Respawn command

`debug_respawn_player` forces inactive players through the server respawn path.

Payload fields:

```text
target_player_id
target_scope
x
y
```

For a single-player request, `target_player_id` is required. For `target_scope = "all_players"`, the server resolves all target player IDs and applies the same guard per player.

Current server behavior logs the provided `x` and `y`, but applies the game-owned safe respawn position rather than trusting the payload position. The respawn path calls:

```text
DevtoolsSafeRespawnPosition
DevtoolsForceRespawnPlayer
```

Active players are ignored. Missing sessions or invalid target IDs are ignored.

### Player counter commands

Player counter commands mutate durable player/session counters through game-owned counter seams.

| Packet type       | Payload field | Server behavior                                 |
| ----------------- | ------------- | ----------------------------------------------- |
| `debug_set_score` | `score`       | Sets score through `DevtoolsSetPlayerScore`.    |
| `debug_add_score` | `amount`      | Adds to score through `DevtoolsAddPlayerScore`. |
| `debug_set_lives` | `lives`       | Sets lives through `DevtoolsSetPlayerLives`.    |
| `debug_add_lives` | `amount`      | Adds to lives through `DevtoolsAddPlayerLives`. |

All four commands support `target_player_id` and `target_scope = "all_players"`.

Score and lives clamping is owned by the shared player counter mutation seam, not by the devtools handler. Current tests verify that negative set/add outcomes clamp below zero instead of producing negative packet/session values.

### Clear entity commands

Clear commands mutate authoritative server entity storage directly through game-owned export methods.

| Packet type             | Server behavior                                                   |
| ----------------------- | ----------------------------------------------------------------- |
| `debug_clear_bullets`   | Removes all current projectiles from `game.entities.Projectiles`. |
| `debug_clear_asteroids` | Removes all current asteroids from `game.entities.Asteroids`.     |

Clients observe the result through normal state/world sync. There is no separate clear acknowledgement packet.

## Telemetry and output

Server devtools output is diagnostic presentation, not analytics.

Current output packets:

```text
debug_status
debug_shape_catalog
```

`debug_status` includes:

```text
debug_status
debug_statuses
```

`debug_status` is the receiving/current player status. `debug_statuses` is a per-player status map keyed by player ID.

Current status fields:

```text
invincible
infinite_lives
world_frozen
asteroids_frozen
bullets_frozen
spawning_frozen
collisions_frozen
player_frozen
```

`StatusFor` reads from game-owned debug status seams. `StatusesForAllPlayers` builds per-player statuses from `MatchDecision().Players`.

The websocket write loop sends debug status periodically while:

```text
the session has a current game player ID
the room has a game instance
server devtools are enabled
the room is InGame or GameOver
```

`debug_shape_catalog` sends server collision-shape catalog data used by client devtools hitbox presentation. The websocket write loop sends the shape catalog once per room ID while the same room/game/devtools availability conditions pass.

Both output packet types encode through `packetcodec`.

## Build/runtime gates

Server devtools availability is controlled by Go build tags:

```text
default build: devtools.Enabled() == true
nodevtools build: devtools.Enabled() == false
```

Files:

```text
services/game-server/internal/devtools/enabled_default.go
services/game-server/internal/devtools/enabled_nodevtools.go
```

`ShouldHandleCommand(packetType)` combines command classification with `Enabled()`:

```text
IsCommandType(packetType) && Enabled()
```

`IsCommandType` includes mutation commands only. It does not classify `debug_status` or `debug_shape_catalog` as inbound commands.

Outbound debug status and debug shape catalog presentation check `devtools.Enabled()` before sending.

Inbound command routing also has runtime context gates:

```text
current room must exist
current game player ID must exist
command JSON must decode into DebugCommand
```

If the room or current game player ID is missing, the packet is consumed without mutation. If command decode fails, networking logs a warning and consumes the packet.

When adding or changing a command type, keep these surfaces aligned:

```text
shared/packets/debug.toml
shared/packets/outputs.toml
services/game-server/internal/devtools/packets_generated.go
services/game-server/internal/devtools/command_types.go
services/game-server/internal/networking/inbound/devtools.go
services/game-server/internal/devtools/handler.go
client/scripts/generated/networking/packets/packets.gd
client devtools packet builders or generated packet builders
tests for routing, command classification, and command behavior
```

## Code map

### Packet source and generated output

```text
shared/packets/debug.toml
shared/packets/outputs.toml
services/game-server/internal/devtools/packets_generated.go
client/scripts/generated/networking/packets/packets.gd
```

### Networking route

```text
services/game-server/internal/networking/websocket_read.go
services/game-server/internal/networking/client_packet_router.go
services/game-server/internal/networking/inbound/client_packet_envelope.go
services/game-server/internal/networking/inbound/router.go
services/game-server/internal/networking/inbound/devtools.go
services/game-server/internal/networking/websocket_write.go
services/game-server/internal/networking/outbound/debug_status_presentation.go
services/game-server/internal/networking/outbound/debug_shape_catalog_presentation.go
```

### Server devtools command handlers

```text
services/game-server/internal/devtools/command_types.go
services/game-server/internal/devtools/handler.go
services/game-server/internal/devtools/disabled.go
services/game-server/internal/devtools/enabled_default.go
services/game-server/internal/devtools/enabled_nodevtools.go
services/game-server/internal/devtools/toggles.go
services/game-server/internal/devtools/player_counters.go
services/game-server/internal/devtools/clear_entities.go
services/game-server/internal/devtools/spawn_entity.go
services/game-server/internal/devtools/spawn_player.go
services/game-server/internal/devtools/spawn_asteroid.go
services/game-server/internal/devtools/spawn_bullet.go
services/game-server/internal/devtools/spawn_pickup.go
services/game-server/internal/devtools/continuous_bullet_stream.go
services/game-server/internal/devtools/respawn_handler.go
services/game-server/internal/devtools/respawn_player.go
services/game-server/internal/devtools/target_scopes.go
services/game-server/internal/devtools/target_player_ids.go
services/game-server/internal/devtools/player_ids.go
services/game-server/internal/devtools/status.go
services/game-server/internal/devtools/shape_catalog.go
```

### Continuous bullet stream runtime

```text
services/game-server/internal/devtools/streamruntime/runtime.go
services/game-server/internal/devtools/streamruntime/simulation.go
services/game-server/internal/devtools/streamruntime/continuous_bullet_streams.go
```

### Game-owned export seams

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

### Related gameplay seams

```text
services/game-server/internal/game/world_simulation_options.go
services/game-server/internal/game/player_counters.go
services/game-server/internal/game/runtime/damage_options.go
services/game-server/internal/game/runtime/life_options.go
services/game-server/internal/game/runtime/suspension.go
services/game-server/internal/game/spawning/
services/game-server/internal/game/entities/pickups/
services/game-server/internal/game/pickups/
services/game-server/internal/game/damage/
services/game-server/internal/game/physics/
```

### Client command sources

```text
client/scripts/devtools/dev_spawn_packet_builder.gd
client/scripts/devtools/dev_respawn_packet_builder.gd
client/scripts/devtools/gameplay_debug_flow.gd
client/scripts/devtools/dev_connection_service.gd
client/scripts/devtools/context/
client/scripts/gameplay/devtools/
```

## Tests

Relevant server tests include:

```text
services/game-server/internal/devtools/command_types_test.go
services/game-server/internal/devtools/enabled_default_test.go
services/game-server/internal/devtools/disabled_test.go
services/game-server/internal/devtools/toggles_test.go
services/game-server/internal/devtools/player_counters_test.go
services/game-server/internal/devtools/target_player_ids_test.go
services/game-server/internal/devtools/clear_entities_test.go
services/game-server/internal/devtools/shape_catalog_test.go
services/game-server/internal/devtools/shape_ids_test.go
services/game-server/internal/devtools/streamruntime/runtime_test.go
services/game-server/internal/devtools/streamruntime/continuous_bullet_streams_test.go
```

Relevant game export seam tests include:

```text
services/game-server/internal/game/export_devtools_player_spawn_test.go
services/game-server/internal/game/export_devtools_respawn_test.go
services/game-server/internal/game/export_devtools_streams_test.go
services/game-server/internal/game/export_devtools_collision_telemetry_test.go
services/game-server/internal/game/devtools_dummy_camera_test.go
```

Relevant networking output tests include:

```text
services/game-server/internal/networking/outbound/debug_status_presentation_test.go
services/game-server/internal/networking/outbound/debug_shape_catalog_presentation_test.go
```

Current focused coverage verifies:

* Command type classification.
* Default and `nodevtools` availability helpers.
* All-player target resolution.
* Single-player target fallback behavior.
* Invincibility, infinite lives, and player-freeze all-player toggle behavior.
* Debug kill all-player behavior.
* Score and lives set/add behavior.
* Score and lives lower-bound clamping through the player counter seam.
* Clear bullets and clear asteroids behavior.
* Shape catalog output construction.
* Continuous bullet stream runtime behavior.

Useful verification command:

```bash
cd services/game-server
go test -buildvcs=false ./internal/devtools ./internal/devtools/streamruntime ./internal/game ./internal/networking/outbound
```

## Related docs

* [Server Devtools](./!INDEX.md)
* [Devtools](../!INDEX.md)
* [Client Devtools](../client/!INDEX.md)
* [Packet Routing And Devtools Input](../client/packet-routing-and-devtools-input.md)
* [Game Server](../../services/game-server/!INDEX.md)
* [Inbound Packet Routing](../../services/game-server/networking/inbound-packet-routing.md)
* [Outbound Message Flow](../../services/game-server/networking/outbound-message-flow.md)
* [Game Aggregate](../../services/game-server/simulation/runtime/game-aggregate.md)
* [Player Counters](../../services/game-server/simulation/players/player-counters.md)
* [Player Pause And Suspension](../../services/game-server/simulation/players/player-pause-and-suspension.md)
* [Player Respawn](../../services/game-server/simulation/players/player-respawn.md)
* [Asteroid Spawning And Variants](../../services/game-server/simulation/world/asteroid-spawning-and-variants.md)
* [Pickup Entity Lifecycle](../../services/game-server/simulation/pickups/pickup-entity-lifecycle.md)
* [Packet Schemas](../../data/packet-schemas.md)
* [Source Of Truth Map](../../data/source-of-truth-map.md)

## Notes

The command surface deliberately keeps interpretation in `internal/devtools` and authoritative mutation in `internal/game`. New commands should follow the same split.

`target_player_id` is valid only for devtools/player-only compatibility commands. New gameplay targeting should use canonical target identity instead.

`debug_status` is status projection, not command acknowledgement. Command results should be inferred from authoritative state or debug status changes.

`debug_respawn_player` currently receives position fields but applies a server-selected safe respawn position. Do not document the payload position as authoritative respawn placement unless the implementation changes.

When adding a new command, update packet source data, generated outputs, inbound routing, command classification, handler dispatch, game export seams, client send paths, and focused tests together.
