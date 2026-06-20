# Devtools Authority and Seams

Parent index: [Design](./!INDEX.md)

## Purpose

This document defines the authority rules and seam boundaries for Space Rocks devtools.

It explains how debug controls can inspect or request changes to gameplay state without becoming a parallel gameplay system.

## Overview

Devtools are a debug-only control and diagnostic layer around real gameplay systems.

The client owns devtools input, presentation, overlays, readmodels, and packet construction. The server owns gameplay mutation, command validation, debug status projection, and the controlled adapters into the authoritative game aggregate.

Current authority flow:

```text
client devtools input or window action
-> client devtools context
-> generated or hand-built debug packet
-> normal client networking send path
-> game-server inbound devtools routing
-> services/game-server/internal/devtools command handler
-> services/game-server/internal/game/export_devtools*.go seam
-> owning gameplay state, simulation option, spawn, damage, counter, respawn, or clear-entity path
-> outbound gameplay/debug presentation packets
-> client devtools readmodels and overlays
```

The devtools seam is allowed to request and display debug-only behavior. It is not allowed to duplicate gameplay rules, mutate client gameplay state as authority, or reach into core gameplay internals without a named game-owned export seam.

## Debug-only scope

Devtools may provide:

```text
debug command hotkeys
devtools window controls
target selectors and status rows
raw local/target telemetry readouts
world telemetry overlay
remote player dev labels
server hitbox overlay
debug spawn and placement tools
debug respawn tools
clear-entity tools
score and lives mutation tools
continuous bullet stream tools
```

These surfaces are for development and diagnostics only.

Devtools must not provide:

```text
player-facing gameplay HUD behavior
production gameplay rules
durable player progression
normal match rules
client-authoritative damage, lives, score, spawn, respawn, target, or freeze behavior
parallel debug-only implementations of gameplay systems
```

Debug-only scope does not mean debug-only logic can bypass ownership. It means the tool surface is not player-facing while the effect still routes through the real owning system.

## Authority model

### Server authority

Gameplay-affecting devtools commands are server-authoritative.

The server decides:

```text
whether a session has a current room
whether a session has a current game player id
whether a command packet is a devtools command
how target scopes resolve
whether a target player exists
whether a spawn, respawn, counter, freeze, clear, or kill operation applies
how the authoritative game state changes
what debug status or gameplay state is emitted afterward
```

The server command package owns devtools command dispatch:

```text
services/game-server/internal/devtools/
```

The game aggregate owns gameplay state and exposes only narrow devtools adapters:

```text
services/game-server/internal/game/export_devtools*.go
```

This keeps `internal/devtools` from owning normal game state and keeps ordinary `internal/game` files from importing or depending on devtools command handling.

### Client authority

The client owns devtools presentation and command intent.

The client may:

```text
open and refresh the devtools window
toggle local devtools overlays
build debug command packets
collect click or drag placement input
select a target row
render raw packet/state readouts
render debug labels and outlines
drop a command request when required local state is missing
```

The client may not:

```text
apply authoritative gameplay effects locally
treat a sent command as confirmed
invent server debug status
mutate synced gameplay entities as a substitute for server state
turn a non-player canonical target into a player-only command target
```

Client confirmation comes from later server output: gameplay state, debug status, shape catalog data, entity presence/absence, or other server-owned presentation packets.

## Packet and command boundary

Devtools command packets are separate from normal gameplay packet handling.

The game server decodes the packet envelope first. Devtools packet families route before normal `game.ClientPacket` decode because debug command packet structs and constants are generated under the devtools package, not the normal game packet family.

Current server command route:

```text
raw websocket message
-> inbound.DecodeClientPacketEnvelope
-> inbound.RouteClientPacket
-> inbound.HandleSimpleDevtoolsPacket / HandlePlacementDevtoolsPacket / HandleRemainingDevtoolsPacket
-> packetcodec.Decode(raw message, devtools.DebugCommand)
-> devtools.HandleCommand(room.GameInstance(), currentGamePlayerID, command)
```

Devtools commands do not route through `Game.HandlePacket`.

Normal gameplay packets still route through the gameplay packet boundary:

```text
input
respawn
client_config
pause_request
target requests
```

Those routes belong to game-server networking and game simulation docs, not to devtools command ownership.

## Relationship to gameplay seams

Devtools must reuse real gameplay seams for gameplay effects.

Current server-owned seams include:

```text
DamageOptions
LifeOptions
Suspension
WorldSimulationOptions
player counter mutation
spawn plans and entity insertion
safe respawn position selection
player ship creation
camera view update
clear-entity mutation
continuous bullet stream runtime registration
collision body telemetry projection
debug status projection
```

Examples:

```text
debug kill player
-> devtools command handler
-> Game.DevtoolsKillPlayer
-> damage.ResolveSingle
-> applyFatalPlayerDamage
```

```text
debug set score
-> devtools command handler
-> Game.DevtoolsSetPlayerScore
-> normal player counter mutation seam
```

```text
debug freeze player
-> devtools command handler
-> Game.DevtoolsSetPlayerFrozen
-> player session suspension state
-> normal ship capability checks
```

```text
debug freeze world
-> devtools command handler
-> WorldSimulationOptions
-> normal simulation phase gates
```

```text
debug respawn player
-> devtools command handler
-> Game.DevtoolsSafeRespawnPosition / DevtoolsForceRespawnPlayer
-> normal player session and camera state
```

The debug command layer can select the operation. The operation itself belongs to the gameplay system that owns the state.

## Client presentation seams

`GameplayDevtoolsContext` is a client-side composition facade. It wires focused contexts and keeps the public devtools API stable while ownership remains in smaller files.

Current client context split:

```text
DevtoolsStateContext
-> cached gameplay-state availability, local player id, canonical target, label mode

DevtoolsCommandContext
-> command request gating and packet send delegation

DevtoolsWindowActionContext
-> devtools window signal wiring

DevtoolsHotkeyContext
-> DevToggle routing

DevtoolsPlacementContext
-> placement request routing and spawn-placement result forwarding

DevtoolsOverlayContext
-> telemetry overlay, remote labels, and hitbox overlay coordination

DevtoolsGameplayStateContext
-> gameplay/debug packet fanout into window, cache, and overlays
```

Presentation seams must stay presentation-only. The devtools window and overlays may display state, request commands, or send diagnostic pings. They must not own simulation state or policy.

## Targeting seams

Devtools uses two related target concepts:

```text
canonical gameplay target
-> target_kind + target_id

player-only devtools command target
-> target_scope + target_player_id
```

The canonical gameplay target is normal gameplay state. It may point at a player, asteroid, bullet, pickup, enemy, or nothing.

Player-only devtools commands only accept player targets. A `Game Target` selector row is valid only when the canonical target kind is `player`.

`All Players` is represented as:

```text
target_scope = "all_players"
target_player_id = ""
```

It is not a fake player ID.

If `target_player_id` is omitted for a single-player command, the server command target resolver falls back to the requesting player where that command supports self-targeting.

## Placement seams

Placement tools are split between client input collection and server mutation.

The client owns:

```text
active click or drag placement state
mouse visual position capture
conversion into server-space placement coordinates
direction detection for drag-based tools
debug spawn packet construction
```

The server owns:

```text
whether the command is accepted
entity construction
spawn plan application
pickup creation
bullet creation
continuous stream registration
game state mutation
```

Current placement flow:

```text
DevToggle6 or devtools window placement button
-> DevtoolsPlacementContext.request_placement_action
-> gameplay shell placement route
-> DevToolsSessionFlow
-> DebugClickPlacementFlow or DebugContinuousBulletSpawnFlow
-> placement result
-> DevConnectionService
-> debug spawn or continuous stream packet
-> server devtools command handler
```

Continuous bullet streams have their own server devtools runtime state under:

```text
services/game-server/internal/devtools/streamruntime/
```

That runtime is devtools-owned. Normal game code exposes only the observer and bullet-spawn seams needed for the stream runtime to integrate with simulation ticks.

## Telemetry seams

Devtools telemetry is diagnostic presentation, not analytics and not player-facing HUD.

Current telemetry surfaces include:

```text
debug_status packet
debug_statuses per-player map
debug_shape_catalog packet
world telemetry overlay
telemetry_ping / telemetry_pong RTT metrics
raw local player telemetry
raw target telemetry
remote player dev labels
server hitbox overlay
collision body telemetry
```

Server-owned telemetry includes debug status, debug shape catalog output, and collision body telemetry projection.

Client-owned telemetry includes presentation readmodels, overlay visibility, label lifecycle, telemetry source selectors, and network metric display.

The devtools window may combine debug status packets with normalized gameplay state to build readmodels, but those readmodels remain transient presentation state.

## Build and runtime gates

Client-side gates are convenience and presentation gates.

Current client-side gates include:

```text
public-build input-map removal for DevToggle0 through DevToggle9
gameplay-state-required checks before command sends
connection-service-required checks before packet sends
placement-route-required checks before placement tools
open-websocket checks before raw packet sending
```

Server-side gates are the authority gates.

Current server-side gates include:

```text
devtools packet classification in inbound networking
current room required before command application
current game player id required before command application
command handlers deciding whether targets and payloads apply
devtools.Enabled() build-tag switch for debug presentation output
```

`nodevtools` switches `devtools.Enabled()` through Go build tags. Debug status and debug shape catalog output check that flag before sending.

Client public-build behavior must not be treated as security or authority. Server routing, server command handlers, and game-owned seams remain the meaningful boundary for gameplay-affecting devtools behavior.

## Non-ownership boundaries

Devtools does not own normal gameplay systems.

Important boundaries:

```text
services/game-server/internal/game/
-> owns authoritative simulation state and gameplay mechanics

services/game-server/internal/devtools/
-> owns debug command handling, debug status projection, and devtools runtime helpers

services/game-server/internal/networking/
-> owns WebSocket read/write loops and packet-family routing

client/scripts/devtools/
-> owns client devtools presentation, input coordination, overlays, readmodels, and packet requests

client/scripts/networking/
-> owns client packet send/receive plumbing

client/scripts/world/
-> owns normal entity presentation and world sync

client/scripts/ui/
-> owns player-facing UI and HUD behavior

shared/packets/
-> owns packet source definitions
```

Devtools may observe or request through these boundaries. It should not absorb their responsibilities.

## Code map

### Shared packet source and generated outputs

```text
shared/packets/debug.toml
shared/packets/outputs.toml
client/scripts/generated/networking/packets/packets.gd
services/game-server/internal/devtools/packets_generated.go
```

### Server command routing

```text
services/game-server/internal/networking/websocket_read.go
services/game-server/internal/networking/client_packet_router.go
services/game-server/internal/networking/inbound/router.go
services/game-server/internal/networking/inbound/devtools.go
services/game-server/internal/devtools/command_types.go
services/game-server/internal/devtools/handler.go
```

### Server authority and game export seams

```text
services/game-server/internal/devtools/
services/game-server/internal/devtools/streamruntime/
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

### Server debug output and gates

```text
services/game-server/internal/devtools/enabled_default.go
services/game-server/internal/devtools/enabled_nodevtools.go
services/game-server/internal/devtools/disabled.go
services/game-server/internal/devtools/status.go
services/game-server/internal/devtools/shape_catalog.go
services/game-server/internal/networking/outbound/debug_status_presentation.go
services/game-server/internal/networking/outbound/debug_shape_catalog_presentation.go
```

### Client composition and command request seams

```text
client/scripts/devtools/gameplay_devtools_context.gd
client/scripts/devtools/context/devtools_state_context.gd
client/scripts/devtools/context/devtools_command_context.gd
client/scripts/devtools/context/devtools_hotkey_context.gd
client/scripts/devtools/context/devtools_window_action_context.gd
client/scripts/devtools/context/devtools_placement_context.gd
client/scripts/devtools/context/devtools_overlay_context.gd
client/scripts/devtools/context/devtools_gameplay_state_context.gd
client/scripts/devtools/gameplay_debug_flow.gd
client/scripts/devtools/dev_connection_service.gd
client/scripts/devtools/devtools_hotkey_flow.gd
client/scripts/devtools/dev_spawn_packet_builder.gd
client/scripts/devtools/dev_respawn_packet_builder.gd
client/scripts/devtools/devtools_target_resolver.gd
```

### Client window, readmodels, overlays, and telemetry

```text
client/scenes/devtools/devtools_window.tscn
client/scripts/devtools/devtools_window.gd
client/scripts/devtools/devtools_window_controller.gd
client/scripts/devtools/devtools_display_refresh_flow.gd
client/scripts/devtools/devtools_player_target_model.gd
client/scripts/devtools/debug_status_packet_reader.gd
client/scripts/devtools/telemetry/
client/scripts/devtools/player_labels/
client/scripts/devtools/hitboxes/
```

### Client placement and input routing

```text
client/scripts/devtools/dev_tools_session_flow.gd
client/scripts/gameplay/devtools/debug_click_placement_flow.gd
client/scripts/gameplay/devtools/debug_continuous_bullet_spawn_flow.gd
client/scripts/gameplay/devtools/debug_mouse_world_position.gd
client/scripts/gameplay/devtools/debug_kill_input_flow.gd
client/scripts/session/gameplay_session_controller.gd
client/scripts/gameplay/gameplay_composition.gd
client/scripts/shell/gameplay_shell_flow.gd
client/scripts/gameplay/input/gameplay_input_context.gd
```

### Client networking

```text
client/scripts/networking/client_connection_service.gd
client/scripts/networking/network_client.gd
client/scripts/networking/packets/packet_codec.gd
client/scripts/networking/outbound/client_packet_sender.gd
client/scripts/networking/outbound/devtools_client_packets.gd
client/scripts/networking/inbound/server_packet_router.gd
client/scripts/networking/inbound/server_packet_dispatcher.gd
```

## Tests and verification

Relevant server tests include:

```text
services/game-server/internal/devtools/command_types_test.go
services/game-server/internal/devtools/disabled_test.go
services/game-server/internal/devtools/enabled_default_test.go
services/game-server/internal/devtools/toggles_test.go
services/game-server/internal/devtools/player_counters_test.go
services/game-server/internal/devtools/target_player_ids_test.go
services/game-server/internal/devtools/clear_entities_test.go
services/game-server/internal/devtools/shape_catalog_test.go
services/game-server/internal/devtools/streamruntime/continuous_bullet_streams_test.go
services/game-server/internal/devtools/streamruntime/runtime_test.go
services/game-server/internal/game/export_devtools_player_spawn_test.go
services/game-server/internal/game/export_devtools_respawn_test.go
services/game-server/internal/game/export_devtools_streams_test.go
services/game-server/internal/game/export_devtools_collision_telemetry_test.go
```

Relevant client tests include:

```text
client/tests/unit/test_gameplay_devtools_context.gd
client/tests/unit/devtools/context/test_devtools_state_context.gd
client/tests/unit/devtools/context/test_devtools_command_context.gd
client/tests/unit/devtools/context/test_devtools_placement_context.gd
client/tests/unit/test_devtools_window_controller.gd
client/tests/unit/devtools/devtools_window_test.gd
client/tests/unit/test_devtools_target_resolver.gd
client/tests/unit/test_devtools_player_target_model.gd
client/tests/unit/devtools/debug_status_packet_reader_test.gd
client/tests/unit/devtools/telemetry/test_world_telemetry_context.gd
client/tests/unit/devtools/telemetry/test_network_telemetry_metrics.gd
client/tests/unit/gameplay/debug/test_server_hitbox_overlay_flow.gd
client/tests/unit/test_gameplay_input_context.gd
client/tests/unit/test_packet_codec.gd
```

Run server devtools tests after changing command classification, target scopes, command handlers, game export seams, debug status, debug shape output, or continuous stream runtime behavior.

Run client devtools tests after changing context wiring, window controls, target resolution, placement routing, overlays, telemetry, or packet builders.

Run packet generation checks after changing `shared/packets/debug.toml` or output routing.

## Related docs

* [Devtools](../!INDEX.md)
* [Devtools Client](../client/!INDEX.md)
* [Devtools Server](../server/!INDEX.md)
* [Client Packet Routing And Devtools Input](../client/packet-routing-and-devtools-input.md)
* [Client Debug Status And Target Readmodels](../client/debug-status-and-target-readmodels.md)
* [Client Devtools Window](../client/devtools-window.md)
* [Client Target And Placement Debugging](../client/target-and-placement-debugging.md)
* [Client Telemetry Overlays](../client/telemetry-overlays.md)
* [Client Hitbox Overlays](../client/hitbox-overlays.md)
* [Game Server Inbound Packet Routing](../../services/game-server/networking/inbound-packet-routing.md)
* [Game Server Outbound Message Flow](../../services/game-server/networking/outbound-message-flow.md)
* [Game Server Simulation](../../services/game-server/simulation/!INDEX.md)
* [Client Gameplay Runtime](../../services/client/gameplay-runtime/!INDEX.md)
* [Client Networking Flow](../../services/client/networking-flow/!INDEX.md)
* [Packet Schemas](../../data/packet-schemas.md)
* [Source Of Truth Map](../../data/source-of-truth-map.md)

## Notes

Devtools authority is intentionally asymmetric: the client can request and display; the server decides and mutates.

The legacy devtools notes correctly identified the core rule that still applies: debug gameplay effects stay server-side, and gameplay-affecting state lives in owning gameplay seams rather than in a parallel debug system.

`target_player_id` remains a devtools/player-only command field. Normal gameplay targeting is `target_kind` plus `target_id`.

Debug status, telemetry overlays, player dev labels, and hitbox overlays are diagnostic presentation. They should remain separate from player-facing HUD and normal gameplay authority.
