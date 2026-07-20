---
author: brian
created: "2026-07-19"
document_id: 019f7d55-fb2c-752a-a0f3-8a5dc38ef098
document_type: general
policy_exempt: false
summary: This document describes the current client-side devtools packet routing and input flow.
---
# Packet Routing And Devtools Input

Parent index: [Client](./!INDEX.md)

## Purpose

This document describes the current client-side devtools packet routing and input flow.

It covers DevToggle hotkeys, devtools window actions, placement input, outbound debug packet construction, inbound debug packet routing, telemetry routing, runtime gates, and the client/server authority boundary for debug tooling.

## Overview

Client devtools are a debug-only input and presentation layer. They collect local development intent, build packets where needed, route those packets through the normal client networking path, and display server-returned diagnostic state and lane-applied read models. They do not own authoritative gameplay mutation.

Current high-level flow:

```text
DevToggle hotkey or devtools window action
-> client devtools context
-> command, placement, overlay, or telemetry route
-> runtime command: GameplayDebugFlow or DevConnectionService
-> ClientConnectionService.send_tooling_packet(packet)
-> sr.tooling
-> server tooling router
-> ToolingPacketRouter for command result/error and developer readouts
-> existing devtools window, overlay, label, or hitbox presentation owner
```

`GameplayDevtoolsContext` is the main client devtools composition facade. It creates and wires focused contexts for cached devtools state, command routing, hotkey routing, window signal routing, placement routing, overlay coordination, and gameplay-state fanout.

Placement tools use an additional session-level flow. `DevToolsSessionFlow` owns active click/drag placement input and runs before normal gameplay input while a placement action is active. It converts visual mouse coordinates into server-space placement results through the world sync coordinate seam, then sends debug spawn or continuous-stream packets through `DevConnectionService`.

Developer readouts are routed through `sr.tooling` and `ToolingPacketRouter`. `debug_status` updates devtools status/readout presentation, `debug_shape_catalog` supports server hitbox overlay presentation, and `telemetry_pong` remains on its current route until continuous telemetry migration. None of those inbound routes grant gameplay authority to the client.

## Debug-only scope

Client devtools are for local development and diagnostics.

They may:

* Open and update the devtools window.
* Toggle remote player dev labels.
* Toggle the world telemetry overlay.
* Toggle the server hitbox overlay presentation.
* Request server-authoritative debug commands.
* Collect click/drag placement coordinates for spawn tools.
* Send telemetry pings while the telemetry overlay is visible.
* Display raw packet/read-model state for local player and target inspection.

They must not:

* Apply authoritative damage, score, lives, respawn, spawn, freeze, or clear-entity effects locally.
* Duplicate server gameplay rules in client-side debug code.
* Treat a client request as confirmed before server state or debug status reflects it.
* Move debug presentation into player-facing HUD ownership.
* Make `target_player_id` the normal gameplay targeting model.
* Bypass the shared `sr.tooling` command send path.

`target_player_id` remains a devtools/player-only compatibility field for player-targeted debug commands. Normal gameplay targeting uses canonical `target_kind` and `target_id`.

## Server authority

See also: [Game Control Devtools Adapter](../server/game-control-devtools-adapter.md)

Gameplay-affecting devtools commands are server-authoritative.

The client sends debug command packets such as:

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

The client sends runtime devtools command payloads through `sr.tooling`. The server tooling router resolves the packet policy, checks the required capability and room attachment, decodes the command, and dispatches it to the existing devtools `Controller.HandleCommand` seam.

Devtools commands do not route through normal client-side authority and should not be documented as gameplay packet ownership. The client only requests a command. The server decides whether the current session, room, player, target scope, and command payload are valid.

Server-side mutation remains behind server-owned handlers and current Control/Controller adapter seams. Command results/errors and developer readouts return through `ToolingPacketRouter`; the existing downstream readout presentation ownership is unchanged.

`GamePlayerID` is passed through to the controller when present, but it is not required merely to dispatch a room-global or explicit-target command.

The server devtools package has build-tag gates for enabled and `nodevtools` builds. Client-side gates are not sufficient security or authority controls; server command availability remains the final boundary.

## Client presentation

`GameplayDevtoolsContext` composes the client devtools presentation and command surface.

Focused ownership is split as follows:

* `DevtoolsStateContext` caches whether gameplay state has been received, the local player ID, canonical game target identity, game-target player compatibility value, and remote-player label mode.
* `DevtoolsCommandContext` owns command request gating and delegates packet sends to `GameplayDebugFlow`, `DevConnectionService`, or the general connection service.
* `DevtoolsWindowActionContext` wires devtools window controller signals to command, placement, and overlay contexts.
* `DevtoolsHotkeyContext` owns DevToggle presentation routing for the devtools window, remote player labels, world telemetry overlay, and delegates command hotkeys to `DevtoolsHotkeyFlow`.
* `DevtoolsPlacementContext` owns placement request routing and placement result forwarding into debug spawn packets.
* `DevtoolsOverlayContext` owns world telemetry, remote player labels, and server hitbox overlay coordination.
* `DevtoolsGameplayStateContext` fans lane-applied gameplay and debug status packets into the devtools window, cached state, target rows, and overlays.

The devtools window is lazily instantiated by `DevtoolsWindowController`. The controller caches current debug status, target rows, telemetry source selections, local player state, target state, and hitbox checkbox state so presentation remains stable whether the window is already open or opened later.

The devtools window presents debug controls and raw readouts. It does not own command authority. Button presses emit signals, the controller resolves target context, and the command or placement contexts decide whether a packet can be sent.

Active devtools placement input is handled while gameplay input routing is active. `GameplaySessionController._input()` gives devtools first refusal before normal HUD/gameplay input only after `accepts_gameplay_packets` has been confirmed true for the active session. If that gate is not open, devtools input does not preempt gameplay/UI handling.

When gameplay tears down, the controller returns before devtools input can inspect or consume the returned-menu click path. This prevents teardown-time menu clicks from being swallowed by the devtools placement flow.

## Commands or controls

### DevToggle map

Current DevToggle routing is:

| Input                 | Behavior                                                       |
| --------------------- | -------------------------------------------------------------- |
| `DevToggle0`          | Toggle the devtools window.                                    |
| `DevToggle1`          | Request invincibility toggle for the local/requesting player.  |
| `DevToggle2`          | Request infinite-lives toggle for the local/requesting player. |
| `DevToggle3`          | Request world-freeze toggle.                                   |
| `DevToggle4`          | Request player-freeze toggle for the local/requesting player.  |
| `DevToggle5`          | Request kill for the local/requesting player.                  |
| `DevToggle6`          | Begin spawn-player placement.                                  |
| `Shift+DevToggle6`    | Begin spawn-asteroid placement.                                |
| `Alt+DevToggle6`      | Begin one-shot spawn-bullet placement.                         |
| `Ctrl+Alt+DevToggle6` | Begin continuous bullet stream placement.                      |
| `DevToggle7`          | Request local player respawn.                                  |
| `DevToggle8`          | Toggle basic remote-player dev labels.                         |
| `Shift+DevToggle8`    | Toggle network remote-player dev labels.                       |
| `DevToggle9`          | Toggle world telemetry overlay.                                |

`DevToggle0`, `DevToggle8`, and `DevToggle9` are presentation or overlay controls. They do not send gameplay mutation packets.

`DevToggle1` through `DevToggle4` route through `GameplayDebugFlow` and are suppressed until gameplay state has been received and a connection service exists.

`DevToggle6` and `DevToggle7` route through `DevtoolsHotkeyFlow` and require gameplay state before they request placement or respawn.

`DevToggle5` routes through `DebugKillInputFlow` and sends a kill request when a connection service exists. Server-side room/player context still determines whether the command has any effect.

### Devtools window controls

The devtools window currently exposes controls for:

```text
invincibility
infinite lives
world freeze
asteroid freeze
bullet freeze
spawn freeze
collision freeze
player freeze
spawn asteroid placement
spawn pickup placement
spawn player placement
spawn bullet placement
respawn player
kill player
set score
add score
set lives
add lives
clear bullets
clear asteroids
set game target
clear game target
show server hitboxes
local telemetry source
target telemetry source
```

Player-targeted controls resolve through `DevtoolsTargetResolver`.

Current target scopes include:

```text
single_player
all_players
```

`All Players` uses `target_scope = "all_players"` and does not send a fake player ID.

The `Game Target` row is valid for player-only controls only when the canonical current target is a player. Non-player canonical targets do not become player IDs for player-only devtools commands.

### Outbound command flow

Most command sends converge on the same raw networking path:

```text
devtools hotkey or window action
-> devtools command/placement context
-> GameplayDebugFlow or DevConnectionService
-> command payload with request_id and trace_id
-> ClientConnectionService.send_tooling_packet(packet)
-> sr.tooling
```

`GameplayDebugFlow` and `DevConnectionService` build runtime command payloads. Runtime command results/errors and developer readouts return through `ToolingPacketRouter`.

### Placement flow

Placement tools collect a server-space position before sending a debug spawn command.

Current placement flow:

```text
DevToggle6 or devtools window placement button
-> DevtoolsPlacementContext.request_placement_action(...)
-> GameplayShellFlow placement route
-> DevToolsSessionFlow.begin_debug_click_placement(...)
-> DebugClickPlacementFlow or DebugContinuousBulletSpawnFlow
-> mouse visual position converted to server position
-> placement result emitted
-> DevConnectionService
-> debug spawn packet
-> ClientConnectionService.send_tooling_packet(packet)
```

Click placement starts on the configured spawn mouse action press and completes on release. If the drag distance is greater than the placement direction threshold, the placement result includes direction data.

One-shot placement can produce:

```text
debug_spawn_entity
debug_spawn_pickup
```

Continuous bullet stream placement produces:

```text
debug_begin_continuous_bullet_stream
```

Continuous bullet stream placement requires a valid drag direction. If no direction is available, the packet builder returns an empty packet and nothing is sent.

### Inbound devtools packet flow

Developer readouts use the tooling route:

```text
sr.tooling
-> RealtimeTransportSession.tooling_packet_received
-> ClientConnectionService._on_tooling_packet_received
-> ToolingPacketRouter.dispatch
```

Current routes preserve the existing public connection-service signals and application owners:

```text
debug_status
-> ToolingPacketRouter.debug_status_received
-> ClientConnectionService.debug_status_received
-> GameplaySessionController
-> GameplayComposition
-> GameplayDevtoolsContext

debug_shape_catalog
-> ToolingPacketRouter.debug_shape_catalog_received
-> ClientConnectionService.debug_shape_catalog_received
-> GameplaySessionController
-> GameplayComposition
-> server hitbox overlay flow

tooling_command_result or tooling_error
-> ToolingPacketRouter
-> command result/error consumers
```

When `sr.tooling` is ready and the room is `in_game` or `game_over`, `ClientConnectionService` sends one `debug_status_subscribe` and one `debug_shape_catalog_request` for that room. Returning to a non-game room state or replacing the realtime transport clears the per-room request marker, allowing a recovered channel to re-establish the readouts.

The normal WebSocket dispatcher/coordinator no longer classify `debug_status` or `debug_shape_catalog`. These packets remain diagnostic presentation data, not gameplay-state authority.

## Telemetry

Devtools telemetry means live debug readouts, not analytics.

The devtools window currently displays:

* Debug status labels for world and freeze-state controls.
* Local player raw telemetry.
* Target raw telemetry.
* Telemetry source selectors for lane-applied state readback and session lane readback. 

`DebugStatusPacketReader` normalizes inbound `debug_status` and `debug_statuses` fields. Non-dictionary values are treated as empty dictionaries before presentation refresh.

The world telemetry overlay is owned by client devtools. It uses gameplay state and telemetry pong responses to show diagnostic world and transport metrics. It is not the gameplay HUD.

World telemetry ping behavior is gated:

```text
overlay must be visible
connection service must exist
server connection must be open
at least 1000 ms must have elapsed since the previous ping
```

When those conditions pass, `WorldTelemetryContext` sends the next telemetry ping packet through `ClientConnectionService.send_packet(packet)`. The server response routes back as `telemetry_pong` and updates network telemetry metrics.

Remote player dev labels can use basic or network mode. Network mode receives metrics from the world telemetry context snapshot. Label lifecycle, formatting, and mode state remain in the client devtools label context, not in normal player or world sync ownership.

Server hitbox overlay presentation is also devtools-owned. The client uses server-derived entity state and debug shape catalog data to draw diagnostic outlines. The overlay does not send gameplay mutation packets.

## Build/runtime gates

Client-side runtime gates include:

```text
DevToggle1-4 command sends require gameplay state and connection service.
DevToggle6 placement requests require gameplay state and a configured placement route.
DevToggle7 local respawn requires gameplay state and a cached local player ID.
Window command requests require gameplay state before packet sends.
Single-player target commands with no effective target player ID are ignored client-side.
Placement result sends require a non-empty result, action name, configured dev connection service, and non-empty built packet.
ClientConnectionService.send_tooling_packet requires an available `sr.tooling` channel.
```

Presentation-only controls have lighter gates:

```text
DevToggle0 can open or close the devtools window.
DevToggle8 can change remote-player label mode.
DevToggle9 can toggle the world telemetry overlay.
The server hitbox checkbox only toggles client overlay presentation.
```

The client includes a devtools build flag script that can erase `DevToggle0` through `DevToggle9` input events when `public_build` is true. This is a client-side input gate only. Server authority and server build gates remain required for gameplay-affecting commands.

Server-side gates include:

```text
devtools command/readout routing requires the packet-specific room attachment and capability; room-global and explicit-target tooling does not require a GamePlayerID
server devtools command handlers own command validity
the server Controller/Target/Control boundary owns command policy and authoritative gameplay mutation
nodevtools server builds disable server devtools availability
```

Outbound devtools packets are not queued. If the client is disconnected, the packet sender is unavailable, the built packet is empty, or encode fails, the command is dropped client-side.

## Code map

### Client devtools composition

```text
client/scripts/devtools/gameplay_devtools_context.gd
client/scripts/devtools/context/devtools_state_context.gd
client/scripts/devtools/context/devtools_command_context.gd
client/scripts/devtools/context/devtools_hotkey_context.gd
client/scripts/devtools/context/devtools_window_action_context.gd
client/scripts/devtools/context/devtools_placement_context.gd
client/scripts/devtools/context/devtools_overlay_context.gd
client/scripts/devtools/context/devtools_gameplay_state_context.gd
client/scripts/devtools/devtools_hotkey_flow.gd
client/scripts/devtools/gameplay_debug_flow.gd
client/scripts/devtools/dev_connection_service.gd
client/scripts/devtools/dev_spawn_packet_builder.gd
client/scripts/devtools/dev_respawn_packet_builder.gd
client/scripts/devtools/devtools_target_resolver.gd
client/scripts/devtools/debug_status_packet_reader.gd
```

### Devtools window presentation

```text
client/scenes/devtools/devtools_window.tscn
client/scripts/devtools/devtools_window.gd
client/scripts/devtools/devtools_window_controller.gd
client/scripts/devtools/devtools_display_refresh_flow.gd
client/scripts/devtools/devtools_player_target_model.gd
```

### Devtools placement and gameplay-session input

```text
client/scripts/devtools/dev_tools_session_flow.gd
client/scripts/gameplay/devtools/debug_kill_input_flow.gd
client/scripts/gameplay/devtools/debug_mouse_world_position.gd
client/scripts/gameplay/devtools/debug_click_placement_flow.gd
client/scripts/gameplay/devtools/debug_continuous_bullet_spawn_flow.gd
client/scripts/session/gameplay_session_controller.gd
client/scripts/gameplay/gameplay_composition.gd
client/scripts/shell/gameplay_shell_flow.gd
client/scripts/gameplay/runtime/gameplay_flow_composer.gd
client/scripts/gameplay/runtime/gameplay_process_flow.gd
client/scripts/gameplay/input/gameplay_input_context.gd
```

### Networking and packet routing

```text
client/scripts/networking/client_connection_service.gd
client/scripts/networking/network_client.gd
client/scripts/networking/packets/packet_codec.gd
client/scripts/networking/outbound/client_packet_sender.gd
client/scripts/networking/inbound/server_packet_router.gd
client/scripts/networking/inbound/server_packet_dispatcher.gd
client/scripts/networking/inbound/tooling_packet_router.gd
client/scripts/devtools/gameplay_debug_flow.gd
client/scripts/devtools/dev_connection_service.gd
client/scripts/generated/networking/packets/packets.gd
```

### Telemetry, labels, and overlays

```text
client/scripts/devtools/telemetry/world_telemetry_context.gd
client/scripts/devtools/telemetry/world_telemetry_overlay_flow.gd
client/scripts/devtools/telemetry/world_telemetry_overlay.gd
client/scripts/devtools/telemetry/network_telemetry_metrics.gd
client/scripts/devtools/telemetry/world_telemetry_metrics.gd
client/scripts/devtools/player_labels/player_dev_labels_context.gd
client/scripts/devtools/player_dev_label.gd
client/scripts/devtools/player_dev_label_formatter.gd
client/scripts/devtools/hitboxes/devtools_server_hitbox_overlay.gd
client/scripts/devtools/hitboxes/debug_shape_catalog_packet_reader.gd
client/scripts/devtools/hitboxes/debug_shape_catalog_store.gd
client/scripts/devtools/hitboxes/debug_shape_id_resolver.gd
client/scripts/gameplay/debug/server_hitbox_overlay_flow.gd
```

### Client build/runtime gate file

```text
client/scripts/devtools/dev_tools_build_flags.gd
```

### Server authority references

```text
services/game-server/internal/networking/client_packet_router.go
services/game-server/internal/networking/inbound/router.go
services/game-server/internal/devtools/controller.go
services/game-server/internal/devtools/target.go
services/game-server/internal/devtools/enabled_default.go
services/game-server/internal/devtools/enabled_nodevtools.go
services/game-server/internal/devtools/packets_generated.go
services/game-server/internal/game/control*.go
```

### Packet source and generated output boundaries

```text
shared/packets/debug.toml
shared/packets/outputs.toml
client/scripts/generated/networking/packets/packets.gd
services/game-server/internal/devtools/packets_generated.go
```

## Tests

Relevant client tests include:

```text
client/tests/unit/devtools/gameplay_debug_flow_test.gd
client/tests/unit/devtools/context/test_devtools_command_context.gd
client/tests/unit/devtools/context/test_devtools_placement_context.gd
client/tests/unit/test_gameplay_devtools_context.gd
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

Current focused coverage verifies:

* Debug flow packet construction for freeze world, invincibility, infinite lives, freeze player, score, and lives.
* All-player target scope packet shape without fake player IDs.
* Single-player target packet shape with target player IDs where required.
* Command context delegation and game-target packet sending.
* Placement context no-op behavior before gameplay state or without a placement route.
* Packet codec encode/decode envelope behavior.
* Telemetry metric and world telemetry context behavior.
* Hitbox overlay flow behavior.

Relevant server-side devtools tests include:

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
```

Those server tests verify command classification, build gates, command effects, targeting/scopes, counter mutation, clear tools, shape catalog behavior, and continuous bullet stream runtime behavior.

## Related docs

* [Client Devtools](./!INDEX.md)
* [Devtools](../!INDEX.md)
* [Server Devtools](../server/!INDEX.md)
* [Client](../../services/client/!INDEX.md)
* [Client Networking Flow](../../services/client/networking-flow/!INDEX.md)
* [Outbound Packet Sending](../../services/client/networking-flow/outbound-packet-sending.md)
* [Inbound Packet Routing](../../services/client/networking-flow/inbound-packet-routing.md)
* [Input And Targeting](../../services/client/input-and-targeting.md)
* [Gameplay Runtime](../../services/client/gameplay-runtime/!INDEX.md)
* [World Sync](../../services/client/world-sync/!INDEX.md)
* [Packet Schemas](../../data/packet-schemas.md)
* [Game Server](../../services/game-server/!INDEX.md)

## Notes

This document intentionally covers the client devtools routing and input side. Server command effects, server build gates, and the server Control adapter belong in server devtools documentation.

Runtime devtools commands and developer readout requests use `ClientConnectionService.send_tooling_packet()` over `sr.tooling`. Command results/errors, `debug_status`, and `debug_shape_catalog` return through `ToolingPacketRouter`; existing downstream presentation ownership is preserved.

Devtools input priority applies to active devtools placement flows. When no devtools placement flow consumes an input event, normal HUD gating and gameplay input continue to own the event path.

Debug status, telemetry, labels, and hitbox overlays are diagnostic presentation. They should remain separate from player-facing HUD and normal gameplay authority.
