# Command Routing And Build Gates

Parent index: [Server](./!INDEX.md)

## Purpose

This document describes how server devtools command packets are routed, dispatched, and gated in the game server.

It covers the WebSocket inbound route, devtools command classification, command dispatch, server-owned mutation seams, runtime/session gates, compile-time build gates, outbound debug packet gates, telemetry, tests, and implementation boundaries.

## Overview

Server devtools commands are client-requested, server-authoritative debug actions. The client sends generated debug packets over the normal WebSocket connection. The game server classifies those packets before normal gameplay packet decoding, decodes them into the devtools command shape, and dispatches them through `services/game-server/internal/devtools`.

Current high-level command flow:

```text
client debug packet
-> WebSocket read loop
-> minimal packet envelope decode
-> inbound packet router
-> devtools packet group classifier
-> devtools.DebugCommand decode
-> devtools.HandleCommand
-> command-specific devtools handler
-> game-owned export_devtools seam
-> authoritative game state mutation
-> lane-native gameplay readback or debug output back to client
```

Devtools commands intentionally do not route through `Game.HandlePacket`. They are not normal gameplay packets, even when they affect gameplay state. Their command packet structs and packet constants are generated under the devtools package, while normal gameplay/lobby/auth/telemetry packet structs are generated under the game package.

The routing split exists so debug tooling can remain isolated from normal gameplay packet ownership while still using real server-owned gameplay seams for mutations.

## Debug-only scope

Server devtools command routing is for development and diagnostic control over active sessions.

It may route commands that:

* Toggle player invincibility.
* Toggle infinite lives.
* Toggle world, asteroid, bullet, spawn, or collision freeze behavior.
* Toggle player freeze.
* Kill, respawn, or spawn players through game-owned seams.
* Spawn asteroids, bullets, pickups, and continuous bullet streams.
* Set or add score.
* Set or add lives.
* Clear bullets or asteroids.
* Provide debug status and shape-catalog output to the client.

It must not:

* Become the normal gameplay packet path.
* Add parallel debug-only gameplay rules that bypass game-owned systems.
* Let the client apply gameplay mutation locally.
* Move authoritative mutation ownership from `internal/game` into networking.
* Treat `target_player_id` as the general gameplay targeting model.
* Place generated devtools command constants in normal gameplay packet ownership.
* Treat debug status packets as gameplay-state authority.

`target_player_id` remains a devtools/player-command compatibility field. Normal gameplay targeting uses canonical target identity outside this command surface.

## Server authority

The game server owns all gameplay-affecting devtools consequences.

`services/game-server/internal/networking` owns WebSocket read/write loops and packet-family routing. It decides whether a raw message belongs to devtools routing and supplies the current session context.

`services/game-server/internal/networking/inbound` owns the devtools command handoff. It checks the packet envelope, groups recognized devtools packet types, decodes the raw message into `devtools.DebugCommand`, and calls:

```text
devtools.HandleCommand(room.GameInstance(), currentGamePlayerID, command)
```

`services/game-server/internal/devtools` owns command dispatch and command-specific debug behavior. It interprets `DebugCommand` fields, resolves player target scopes, validates command-specific payload needs, logs command outcomes, and calls the narrow game-owned APIs needed for real mutation.

`services/game-server/internal/game/export_devtools*.go` owns the mutation boundary exposed to devtools. These files keep authoritative state changes inside the game package while preventing normal game code from importing devtools.

Examples of game-owned devtools seams include:

```text
DevtoolsSetPlayerInvincible
DevtoolsSetInfiniteLives
DevtoolsSetPlayerFrozen
DevtoolsToggleFreezeWorld
DevtoolsToggleFreezeAsteroids
DevtoolsToggleFreezeBullets
DevtoolsToggleFreezeSpawning
DevtoolsToggleFreezeCollisions
DevtoolsKillPlayer
DevtoolsSpawnBullet
DevtoolsApplyAsteroidSpawnPlan
DevtoolsForceRespawnPlayer
DevtoolsClearBullets
DevtoolsClearAsteroids
DevtoolsRegisterSimulationStepObserver
```

The authority rule is: devtools may request and coordinate debug behavior, but gameplay state still changes through server-owned game seams.

## Client presentation

The client is presentation and request-side only for server devtools commands.

Client devtools hotkeys, placement tools, and window buttons build debug packets and send them over the normal networking path. The client does not apply authoritative score, lives, death, respawn, spawn, freeze, clear, or damage effects locally.

Server confirmation reaches the client through normal server-owned outputs:

```text
lane-native gameplay readback
debug_status packets
debug_shape_catalog packets
entity presence or absence in lane-native sync
player lifecycle/session state
telemetry readouts
server logs during development
```

The server also produces outbound debug presentation packets only when its own runtime gates pass. Client-side input gates and public-build input stripping are useful presentation controls, but they are not the authority boundary for gameplay-affecting commands.

## Commands or controls

### Command packet shape

Generated server devtools command packets live in:

```text
services/game-server/internal/devtools/packets_generated.go
```

The current generated command carrier is:

```text
DebugCommand
```

It includes fields used across multiple command families:

```text
type
target_player_id
target_scope
entity_type
pickup_type
x
y
has_direction
direction_x
direction_y
freeze_target
score
amount
lives
```

Command handlers interpret only the fields relevant to the command type they own.

### Recognized command types

The devtools command classifier currently recognizes:

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

There is no separate current `debug_spawn_player` packet. Player spawning is routed through:

```text
debug_spawn_entity
entity_type = "player"
```

### Routing groups

Inbound devtools routing groups recognized command packet types before normal `game.ClientPacket` decode.

Current groups are:

```text
simple devtools packets
placement devtools packets
remaining devtools packets
```

Simple devtools packets include:

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

Placement devtools packets include:

```text
debug_spawn_entity
debug_spawn_pickup
```

Remaining devtools packets include:

```text
debug_begin_continuous_bullet_stream
debug_respawn_player
```

All groups delegate to the same internal command decode and dispatch path.

### Dispatch table

`devtools.HandleCommand` dispatches by `DebugCommand.Type`.

Current dispatch ownership:

```text
toggle_debug_invincible
-> handleToggleDebugInvincible

toggle_debug_infinite_lives
-> handleToggleDebugInfiniteLives

toggle_debug_freeze_world
-> handleToggleDebugFreezeWorld

toggle_debug_freeze_player
-> handleToggleDebugFreezePlayer

debug_kill_player
-> handleDebugKillPlayer

debug_spawn_entity
-> handleDebugSpawnEntity

debug_spawn_pickup
-> handleDebugSpawnPickup

debug_begin_continuous_bullet_stream
-> handleDebugBeginContinuousBulletStream

debug_respawn_player
-> handleDebugRespawnPlayer

debug_set_score
-> handleDebugSetScore

debug_add_score
-> handleDebugAddScore

debug_set_lives
-> handleDebugSetLives

debug_add_lives
-> handleDebugAddLives

debug_clear_bullets
-> handleDebugClearBullets

debug_clear_asteroids
-> handleDebugClearAsteroids
```

Unknown command types return `false` from `HandleCommand`.

### Target scopes

Player-targeted commands resolve through devtools target helpers.

Current target scopes are:

```text
single_player
all_players
```

`all_players` resolves to `target.DevtoolsTargetPlayerIDs()` and does not use a fake player ID.

When a command does not send `target_player_id`, single-player command resolution falls back to the requesting/current game player ID.

Set-style all-player toggles use shared all-player state logic:

* If any eligible player does not currently have the toggle enabled, the all-player command enables the toggle for all eligible players.
* If every eligible player already has the toggle enabled, the all-player command disables the toggle for all eligible players.

This applies to invincibility, infinite lives, and player freeze.

## Telemetry

Server devtools command routing produces development telemetry through logs and debug presentation packets.

### Command decode logging

If a recognized devtools command packet cannot decode into `devtools.DebugCommand`, inbound routing logs:

```text
websocket devtools command decode failed
```

The log includes error, room ID, player ID, session ID, and remote address fields.

The failed packet is consumed and does not fall through into normal gameplay routing.

### Command outcome logging

Command handlers log meaningful outcomes through the game logger.

Examples include:

```text
debug invincibility set
debug infinite lives set
debug world freeze toggled
debug world freeze target ignored
debug player freeze set
debug player spawned
debug bullet spawned
debug asteroid spawned
debug spawn entity not implemented for entity type
debug begin continuous bullet stream ignored
debug continuous bullet stream started
debug pickup spawned
debug pickup spawn ignored
```

These logs are diagnostics, not protocol responses.

### Debug status output

Debug status output is an outbound diagnostic packet. It is built from:

```text
devtools.StatusFor(gameInstance, playerID)
devtools.StatusesForAllPlayers(gameInstance)
```

and emitted as:

```text
debug_status
```

The WebSocket write loop sends debug status periodically after gameplay presentation state is being written for an active game session. The current write cadence is every 8 server ticks.

Debug status output is gated by:

```text
room exists
game instance exists
devtools.Enabled() is true
room state is InGame or GameOver
current game player ID exists
```

### Debug shape catalog output

Debug shape catalog output is an outbound diagnostic packet. It loads the current collision shape catalog and emits:

```text
debug_shape_catalog
```

The WebSocket write loop sends the shape catalog once per room ID after gameplay presentation state is being written.

Debug shape catalog output is gated by:

```text
room exists
game instance exists
devtools.Enabled() is true
room state is InGame or GameOver
current room ID exists
shape catalog has not already been sent for that room ID
```

## Build/runtime gates

### Compile-time build gate

Server devtools have a Go build-tag gate.

Default builds compile:

```text
services/game-server/internal/devtools/enabled_default.go
```

with:

```go
//go:build !nodevtools

func Enabled() bool { return true }
```

`nodevtools` builds compile:

```text
services/game-server/internal/devtools/enabled_nodevtools.go
```

with:

```go
//go:build nodevtools

func Enabled() bool { return false }
```

The shared gate helper is:

```text
services/game-server/internal/devtools/disabled.go
```

It defines:

```go
func ShouldHandleCommand(packetType string) bool {
	return IsCommandType(packetType) && Enabled()
}
```

Tests verify that `Enabled()` and `ShouldHandleCommand(...)` return true in default builds and false for devtools packets in `nodevtools` builds.

### Current inbound command gate

The current inbound command routing path applies these runtime gates before command dispatch:

```text
packet envelope must decode
packet type must match one of the inbound devtools packet group switches
current room must exist
current game player ID must exist
raw packet must decode into devtools.DebugCommand
```

If the current room or current game player ID is missing, the recognized devtools packet is consumed and no command is applied. This prevents devtools command packets from falling through into normal gameplay routing.

If command decode fails, the packet is logged and consumed.

The current inbound router uses local packet-type switches in `inbound/devtools.go` for command grouping. The build-gated `ShouldHandleCommand` helper exists and is tested, but the inspected inbound routing implementation does not currently use that helper as the classifier. New routing edits should converge command availability through the single build-gated classifier instead of duplicating packet-type lists.

### Handler-level gates

Individual handlers apply command-specific guards.

Examples:

* Continuous bullet streams require `has_direction = true`.
* Continuous bullet streams reject zero-length direction vectors.
* Spawn entity dispatch only handles known entity types.
* Pickup spawning delegates validity to the game pickup spawn path.
* Respawn uses game-owned respawn placement and force-respawn seams.
* Kill-player checks that the target is active before applying debug fatal damage.
* Clear bullets and clear asteroids are safe when the entity maps are empty.
* Player counter commands use game-owned counter mutation seams.
* Freeze commands ignore unknown `freeze_target` values after logging.

### Outbound debug packet gate

Outbound debug status and debug shape catalog packets are gated by `devtools.Enabled()`.

This means `nodevtools` builds suppress current outbound debug presentation packets that use these helpers.

The outbound debug packet gates are separate from inbound command routing. They do not prove that a command packet was rejected; they only decide whether server-generated debug status and shape-catalog packets are written.

## Relationship to gameplay implementation areas

Devtools command handlers should route through real gameplay implementation areas.

Current examples:

* Invincibility mutates player/session damage options and is consumed by normal damage/collision rules.
* Infinite lives mutates session life options and is consumed by normal death/life rules.
* Player freeze mutates session suspension and active ship input state through the same suspension model used by gameplay capability checks.
* World freeze mutates `WorldSimulationOptions`, which simulation phases read before spawning, movement, bullets, and collision work.
* Kill player builds a debug damage request and applies normal fatal player damage consequences.
* Score and lives commands delegate to player counter seams.
* Spawn and respawn commands use game-owned spawn, camera, and entity insertion seams.
* Clear entity commands mutate authoritative server entity maps; clients observe the result through normal state sync.
* Continuous bullet streams are paced by server simulation step observers and server bullet movement gates.

The devtools package coordinates debug intent. It does not own normal gameplay rules.

## Code map

### Command classification and build gates

```text
services/game-server/internal/devtools/command_types.go
services/game-server/internal/devtools/disabled.go
services/game-server/internal/devtools/enabled_default.go
services/game-server/internal/devtools/enabled_nodevtools.go
services/game-server/internal/devtools/packets_generated.go
```

### Inbound routing

```text
services/game-server/internal/networking/websocket_read.go
services/game-server/internal/networking/client_packet_router.go
services/game-server/internal/networking/inbound/client_packet_envelope.go
services/game-server/internal/networking/inbound/router.go
services/game-server/internal/networking/inbound/devtools.go
```

### Command dispatch and handlers

```text
services/game-server/internal/devtools/handler.go
services/game-server/internal/devtools/toggles.go
services/game-server/internal/devtools/spawn_entity.go
services/game-server/internal/devtools/spawn_player.go
services/game-server/internal/devtools/spawn_asteroid.go
services/game-server/internal/devtools/spawn_bullet.go
services/game-server/internal/devtools/spawn_pickup.go
services/game-server/internal/devtools/respawn_player.go
services/game-server/internal/devtools/respawn_handler.go
services/game-server/internal/devtools/player_counters.go
services/game-server/internal/devtools/clear_entities.go
services/game-server/internal/devtools/continuous_bullet_stream.go
services/game-server/internal/devtools/player_ids.go
services/game-server/internal/devtools/target_scopes.go
services/game-server/internal/devtools/target_player_ids.go
services/game-server/internal/devtools/placement_requests.go
```

### Continuous bullet stream runtime

```text
services/game-server/internal/devtools/streamruntime/runtime.go
services/game-server/internal/devtools/streamruntime/simulation.go
services/game-server/internal/devtools/streamruntime/continuous_bullet_streams.go
```

### Debug status and shape output

```text
services/game-server/internal/devtools/status.go
services/game-server/internal/devtools/shape_catalog.go
services/game-server/internal/devtools/shape_ids.go
services/game-server/internal/networking/outbound/debug_status_presentation.go
services/game-server/internal/networking/outbound/debug_shape_catalog_presentation.go
services/game-server/internal/networking/websocket_write.go
```

### Game-owned devtools seams

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

### Source and generated packet boundaries

```text
shared/packets/debug.toml
shared/packets/outputs.toml
services/game-server/internal/devtools/packets_generated.go
client/scripts/generated/networking/packets/packets.gd
```

### Important non-ownership boundaries

```text
services/game-server/internal/networking
```

owns WebSocket sessions, read/write loops, inbound routing, and outbound message writing. It does not own command effects.

```text
services/game-server/internal/game
```

owns authoritative gameplay state and the export seams used by devtools. It should not import `internal/devtools`.

```text
client/scripts/devtools
```

owns client request and presentation behavior. It does not own server command authority.

```text
docs/data
```

owns shared packet schema and generation documentation.

## Tests

Relevant server tests include:

```text
services/game-server/internal/devtools/command_types_test.go
services/game-server/internal/devtools/enabled_default_test.go
services/game-server/internal/devtools/disabled_test.go
services/game-server/internal/devtools/toggles_test.go
services/game-server/internal/devtools/target_player_ids_test.go
services/game-server/internal/devtools/player_counters_test.go
services/game-server/internal/devtools/clear_entities_test.go
services/game-server/internal/devtools/shape_ids_test.go
services/game-server/internal/devtools/shape_catalog_test.go
services/game-server/internal/devtools/streamruntime/runtime_test.go
services/game-server/internal/devtools/streamruntime/continuous_bullet_streams_test.go
services/game-server/internal/game/export_devtools_player_spawn_test.go
services/game-server/internal/game/export_devtools_respawn_test.go
services/game-server/internal/game/export_devtools_streams_test.go
services/game-server/internal/game/export_devtools_clear_entities.go
services/game-server/internal/networking/outbound/debug_status_presentation_test.go
services/game-server/internal/networking/outbound/debug_shape_catalog_presentation_test.go
```

Current focused coverage verifies:

* Devtools command type classification.
* Default build devtools enablement.
* `nodevtools` build devtools disablement.
* `ShouldHandleCommand` behavior under default and `nodevtools` builds.
* Toggle command effects.
* All-player and single-player target resolution.
* Score and lives command behavior.
* Clear bullets and clear asteroids behavior.
* Shape ID and shape catalog output behavior.
* Continuous bullet stream runtime behavior.
* Game-owned player spawn and respawn export seams.
* Outbound debug status and debug shape catalog gates.

The inbound packet-family router itself has thinner direct route-order coverage than the command handlers and outbound gate helpers. Existing routing documentation should therefore distinguish handler/gate coverage from direct inbound router coverage.

## Related docs

* [Server Devtools](./!INDEX.md)
* [Devtools](../!INDEX.md)
* [Client Devtools](../client/!INDEX.md)
* [Packet Routing And Devtools Input](../client/packet-routing-and-devtools-input.md)
* [Game Server](../../services/game-server/!INDEX.md)
* [Inbound Packet Routing](../../services/game-server/networking/inbound-packet-routing.md)
* [Outbound Message Flow](../../services/game-server/networking/outbound-message-flow.md)
* [Game Server Simulation](../../services/game-server/simulation/!INDEX.md)
* [Player Counters](../../services/game-server/simulation/players/player-counters.md)
* [Player Respawn](../../services/game-server/simulation/players/player-respawn.md)
* [Player Pause And Suspension](../../services/game-server/simulation/players/player-pause-and-suspension.md)
* [Packet Schemas](../../data/packet-schemas.md)

## Notes

Legacy docs correctly described the intended boundary: devtools command handling is separate from normal gameplay packet routing, and debug mutations must flow through server-owned gameplay seams.

The current implementation has both a generated devtools command type set and separate inbound devtools packet group switches. Keep those lists synchronized when adding command packets. Prefer a single build-gated command classifier for future routing cleanup.

Do not treat outbound debug status or debug shape catalog packets as proof that a command succeeded. They are diagnostic outputs. Command confirmation should be inferred from authoritative game state, debug status state, entity sync, lifecycle/session read-models, or server logs during development.

