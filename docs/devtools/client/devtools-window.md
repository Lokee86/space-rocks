# Devtools Window

Parent index: [Client](./!INDEX.md)

## Purpose

This document describes the client devtools window: the Godot debug window used to inspect live gameplay state, send server-authoritative debug commands, select debug targets, request placement-based debug spawns, and toggle client-only diagnostic overlays.

## Overview

The devtools window is a client-side development tool, not player-facing HUD. It is opened through `DevToggle0`, instantiated lazily by `DevtoolsWindowController`, and displayed as an always-on-top Godot `Window` scene.

The window owns presentation and control collection only. It does not apply gameplay mutations locally. Button presses, selectors, telemetry source changes, and checkbox changes emit signals. Those signals are routed through the client devtools context stack into either:

* generated debug packets sent over the normal client networking path
* placement flows that collect a server-space position or direction before sending a debug spawn packet
* client-only overlay toggles such as server hitbox drawing
* normal target-selection packets when setting or clearing the canonical game target

Server-authoritative debug effects remain owned by the game server. The client window may request invincibility, infinite lives, world-freeze changes, player-freeze changes, player kill/respawn, score/lives edits, entity spawns, pickup spawns, bullet streams, and entity clears, but the server decides whether the command is valid and how it mutates gameplay state.

## Debug-only scope

The devtools window is debug tooling for active gameplay sessions. It is separate from production HUD, menu flow, player progression, normal gameplay input, and client presentation state.

The window may display raw lane-applied gameplay dictionaries, debug status, target rows, player lifecycle rows, pickup selector values, and overlay toggles. It should not translate these into player-facing UI, user progression features, or permanent gameplay state.

Client debug input is gated by `client/scripts/devtools/dev_tools_build_flags.gd`. When `public_build` is enabled, the configured `DevToggle0` through `DevToggle9` input events are erased from the `InputMap`. The server-side command path has its own `nodevtools` build gate; disabling server devtools prevents command handling even if a client can still construct debug packets.

Runtime command contexts also require gameplay state before sending most debug actions. If the client has not received gameplay state, or the relevant connection service is not configured, command requests are ignored.

## Server authority

The window follows a request-only model.

Server-owned gameplay mutations include:

* invincibility
* infinite lives
* world freeze
* granular world freeze targets
* player freeze
* kill player
* respawn player
* set score
* add score
* set lives
* add lives
* spawn asteroid
* spawn pickup
* spawn player
* spawn bullet
* begin continuous bullet stream
* clear bullets
* clear asteroids

The window sends generated packets or packet dictionaries through the normal network client path. It does not directly edit `WorldSync`, player nodes, asteroid nodes, bullet nodes, score values, lives values, or lane-applied server-state read models.

The server remains authoritative for pickup spawn validity. The pickup selector is populated from the client pickup presentation catalog so available debug choices match presentation data, but the selector does not make the client authoritative over pickup types.

The server hitbox checkbox is the main exception because it toggles a client-only diagnostic overlay. It does not send a server command and does not mutate gameplay. It only controls whether already-synced server collision telemetry is drawn by the devtools overlay.

## Client presentation

The window scene lives at:

```text
client/scenes/devtools/devtools_window.tscn
```

The root script lives at:

```text
client/scripts/devtools/devtools_window.gd
```

The window is titled `Space Rocks Devtools`, opens centered, stays on top, and hides instead of freeing itself when closed.

The current window surface includes:

* server collision telemetry checkbox
* pickup spawn selector and spawn-pickup button
* invincible toggle and target/status selector
* infinite-lives toggle and target/status selector
* world-freeze toggle and status label
* granular freeze buttons for asteroids, bullets, spawns, and collisions
* player-freeze toggle and target/status selector
* kill-player target selector
* spawn-player selector
* respawn-player selector
* spawn-asteroid and spawn-bullet placement buttons
* clear-asteroids and clear-bullets buttons
* set/add score controls
* set/add lives controls
* game-target selector with set and clear buttons
* local-player telemetry panel
* target telemetry panel

Status labels use `Active` and `Inactive` wording for debug statuses. Player lifecycle selectors may still use lifecycle wording such as `ALIVE` and `DEAD`.

## Commands and controls

`DevToggle0` toggles the devtools window.

Window controls use signals from `devtools_window.gd`. `DevtoolsWindowController` preserves the latest known state so the window can be created after gameplay state or debug status has already arrived.

Player-targeted controls resolve through `DevtoolsTargetResolver` and can use:

* an explicit selected player
* `Game Target`, when the canonical gameplay target is a player
* `All Players`, where the command supports `target_scope=all_players`
* the local player fallback where the command path supports fallback

`All Players` is represented as a scope, not as a fake player ID. It emits `target_scope=all_players` with an empty `target_player_id`.

Player-only commands do not resolve non-player canonical targets. If `Game Target` points at an asteroid, bullet, pickup, or enemy, player-only commands such as kill, invincible, infinite lives, player freeze, score edits, and lives edits do not emit an effective player command.

World-freeze controls use `toggle_debug_freeze_world`. The main world-freeze button sends the default/global freeze request. Granular buttons pass a `freeze_target` value:

```text
asteroids
bullets
spawns
collisions
```

Score and lives inputs must parse as integers before a command is emitted. Empty or invalid integer fields are ignored by the window.

Spawn and respawn buttons do not immediately mutate gameplay. They route into placement or respawn request flows:

* asteroid, bullet, pickup, and player spawn controls request placement
* pickup placement includes the selected `pickup_type`
* player spawn placement may include a selected target player slot
* continuous bullet stream placement is owned by the devtools session placement flow, not the window itself
* respawn requests use the selected target scope and player ID

Set/Clear Game Target uses the normal target packet path to update the canonical player target. This is a target-selection request, not a debug mutation by itself.

## Target rows and read models

`DevtoolsPlayerTargetModel` builds window target rows from applied gameplay lane state and debug status.

Inputs include:

* `self_id`
* `server_players`
* `player_sessions`
* `server_asteroids`
* `server_bullets`
* `server_pickups`
* `server_enemies` or `enemies`
* `player_lifecycle`
* `debug_statuses`
* canonical `target_kind`
* canonical `target_id`

The model creates separate row sets for different control types:

* kill-player rows
* respawn-player rows
* invincibility rows
* infinite-lives rows
* player-frozen rows
* active-player rows for score/lives controls
* game-target options

Feature rows display per-player debug status as `Active` or `Inactive`. Lifecycle rows use lifecycle state where available, and otherwise fall back to whether a player exists in `server_players`.

The canonical target supports display and telemetry for:

```text
player
enemy
pickup
asteroid
bullet
```

Only `player` canonical targets resolve into player-command targets.

## Telemetry behavior

The devtools window owns raw state inspection. It currently exposes two raw telemetry panels:

* local player telemetry
* current target telemetry

Each panel can select a source:

```text
world lane readback
session lane readback
```

The `world lane readback` source reads applied world entity dictionaries. For the local player, this reads from `server_players`. For targets, it can read from player, asteroid, bullet, pickup, or enemy dictionaries according to canonical `target_kind`.

The `session lane readback` source reads durable player/session dictionaries from `player_sessions`. It is only meaningful for player targets. Non-player targets render unavailable output for this source.

Telemetry output is intentionally generic. Dictionary keys are sorted and rendered as key/value lines. Arrays and dictionaries are rendered as JSON strings. Floats are rendered to four decimal places. Target telemetry includes `target_kind` and `target_id` above the raw body when a target is selected.

Debug status packets update:

* global/world freeze status
* granular freeze status
* per-player invincible status
* per-player infinite-lives status
* per-player frozen status

The window telemetry panels are separate from the world telemetry overlay. The world telemetry overlay is a glanceable metrics surface; the devtools window is a raw inspection and command surface.

## Runtime flow

The current client flow is:

```text
DevToggle0
-> DevtoolsHotkeyContext
-> GameplayDevtoolsContext.toggle_devtools_window()
-> DevtoolsWindowController.toggle_window()
-> DevtoolsWindowScene.instantiate()
-> devtools_window.gd
```

Window command flow is:

```text
devtools_window.gd signal
-> DevtoolsWindowController
-> DevtoolsWindowActionContext
-> DevtoolsCommandContext / DevtoolsPlacementContext / DevtoolsOverlayContext
-> GameplayDebugFlow / DevConnectionService / overlay implementation
-> generated packet or client-only overlay update
```

Gameplay state and debug status flow back into the window through:

```text
lane/debug packet
-> GameplayShellFlow
-> GameplayFlowComposer
-> GameplayDevtoolsContext
-> DevtoolsGameplayStateContext
-> DevtoolsDisplayRefreshFlow
-> DevtoolsWindowController
-> devtools_window.gd
```

Placement requests flow through the gameplay shell because placement needs the current world coordinate conversion:

```text
window spawn control
-> DevtoolsWindowController
-> DevtoolsPlacementContext
-> GameplayShellFlow placement route
-> DevToolsSessionFlow
-> DebugClickPlacementFlow or DebugContinuousBulletSpawnFlow
-> placement result
-> GameplayShellFlow.handle_devtools_placement_result()
-> GameplayDevtoolsContext.handle_placement_result()
-> DevConnectionService
-> debug spawn packet
```

## Build and runtime gates

Client-side gate:

```text
client/scripts/devtools/dev_tools_build_flags.gd
```

This file removes `DevToggle0` through `DevToggle9` input events when `public_build` is true.

Server-side gate:

```text
nodevtools
```

Server `nodevtools` builds disable server command handling. The client window must not be treated as authority even when it can still construct a debug packet.

Runtime gates:

* command contexts require gameplay state for most actions
* connection-backed commands require a configured connection service
* respawn and spawn helpers refuse empty packet builds
* player-only commands refuse empty effective player targets
* non-player canonical targets do not resolve into player-only command targets

## Code map

Primary window files:

```text
client/scenes/devtools/devtools_window.tscn
client/scripts/devtools/devtools_window.gd
client/scripts/devtools/devtools_window_controller.gd
```

Composition and action routing:

```text
client/scripts/devtools/gameplay_devtools_context.gd
client/scripts/devtools/context/devtools_window_action_context.gd
client/scripts/devtools/context/devtools_command_context.gd
client/scripts/devtools/context/devtools_placement_context.gd
client/scripts/devtools/context/devtools_overlay_context.gd
client/scripts/devtools/context/devtools_hotkey_context.gd
client/scripts/devtools/context/devtools_gameplay_state_context.gd
client/scripts/devtools/context/devtools_state_context.gd
```

Command and target helpers:

```text
client/scripts/devtools/gameplay_debug_flow.gd
client/scripts/devtools/devtools_player_target_model.gd
client/scripts/devtools/devtools_target_resolver.gd
client/scripts/devtools/devtools_display_refresh_flow.gd
client/scripts/devtools/debug_status_packet_reader.gd
```

Packet and network helpers:

```text
client/scripts/generated/networking/packets/packets.gd
client/scripts/networking/outbound/devtools_client_packets.gd
client/scripts/networking/outbound/client_packet_sender.gd
client/scripts/networking/client_connection_service.gd
client/scripts/devtools/dev_connection_service.gd
client/scripts/devtools/dev_spawn_packet_builder.gd
client/scripts/devtools/dev_respawn_packet_builder.gd
```

Placement support:

```text
client/scripts/devtools/dev_tools_session_flow.gd
client/scripts/gameplay/devtools/debug_click_placement_flow.gd
client/scripts/gameplay/devtools/debug_continuous_bullet_spawn_flow.gd
client/scripts/gameplay/devtools/debug_mouse_world_position.gd
client/scripts/shell/gameplay_shell_flow.gd
client/scripts/gameplay/runtime/gameplay_flow_composer.gd
```

Related overlay files:

```text
client/scripts/devtools/hitboxes/devtools_server_hitbox_overlay.gd
client/scripts/devtools/hitboxes/debug_shape_catalog_packet_reader.gd
client/scripts/devtools/hitboxes/debug_shape_catalog_store.gd
client/scripts/devtools/hitboxes/debug_shape_id_resolver.gd
client/scenes/devtools/server_hitbox_overlay.tscn
client/scripts/devtools/telemetry/world_telemetry_context.gd
client/scripts/devtools/telemetry/world_telemetry_overlay.gd
client/scenes/devtools/world_telemetry_overlay.tscn
client/scripts/devtools/player_labels/player_dev_labels_context.gd
client/scripts/devtools/player_dev_label.gd
client/scripts/devtools/player_dev_label_formatter.gd
client/scenes/devtools/player_dev_label.tscn
```

Important non-ownership boundaries:

* `devtools_window.gd` owns UI controls and signal emission, not gameplay mutation.
* `DevtoolsWindowController` owns window lifecycle, latest cached UI state, and target-scope conversion.
* `DevtoolsCommandContext` owns command routing into debug packet send paths.
* `DevtoolsPlacementContext` owns placement request handoff, not mouse/world coordinate conversion.
* `DevToolsSessionFlow` owns debug placement interaction and continuous bullet placement routing.
* `WorldSync` and gameplay runtime code expose read-only state needed by devtools but do not own devtools rendering.
* server hitbox rendering is client devtools presentation; authoritative shape/state data comes from synced server/debug data.

## Tests and verification

Relevant client tests include:

```text
client/tests/unit/devtools/devtools_window_test.gd
client/tests/unit/test_devtools_window_controller.gd
client/tests/unit/test_devtools_player_target_model.gd
```

Current test coverage verifies:

* granular freeze buttons emit the expected freeze targets
* debug status updates granular freeze labels
* pickup selector values are populated from the pickup catalog and default to `1_up`
* explicit selected players override canonical game targets
* player canonical targets resolve for player-only commands
* non-player canonical targets do not emit player-only commands
* `All Players` emits `target_scope=all_players` with an empty player ID
* target-scope routing is preserved for respawn, toggle, score, and lives actions
* target row and telemetry model behavior is derived from gameplay state and debug status

## Related docs

* [Client Devtools](./!INDEX.md)
* [Server Devtools](../server/!INDEX.md)
* [Devtools Design](../design/!INDEX.md)
* [Client Service](../../services/client/!INDEX.md)
* [Protocol](../../protocol/!INDEX.md)
* [Data](../../data/!INDEX.md)

## Notes

The devtools window should remain a narrow control and inspection surface. New controls should route through the existing devtools contexts, generated packet helpers, placement route, or overlay route instead of adding local gameplay mutation.

Raw telemetry should stay generic. Field-specific gameplay interpretation belongs in the owning gameplay, protocol, data, service, or systems-design documentation rather than in custom devtools-window formatting.
