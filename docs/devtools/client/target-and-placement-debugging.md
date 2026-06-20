# Target And Placement Debugging

Parent index: [Client](./!README.md)

## Purpose

This document describes client-side devtools for inspecting gameplay targets and performing click-based placement actions.

It covers client presentation, input routing, target read models, placement request/result forwarding, and client build/runtime gates. Server-side command execution and authoritative target validation are referenced only as authority boundaries.

## Overview

Target and placement debugging lets a developer inspect the current canonical gameplay target, route player-only debug commands through safe target resolution, and place debug entities into the world with mouse input.

The client owns:

* Opening and refreshing the devtools window.
* Displaying local player and target readouts.
* Resolving devtools UI selections into command target context.
* Starting temporary placement actions from hotkeys or window buttons.
* Translating click or click-drag placement into devtools packets.
* Preventing active placement from falling through into normal target selection.

The server owns:

* Whether a requested target is valid.
* Whether a placed entity may be spawned.
* The resulting gameplay mutation.
* The authoritative state that confirms any target or placement outcome.

The important boundary is that client devtools may request and display debug behavior, but they do not become gameplay authority.

## Debug-only scope

Target and placement debugging is development tooling. It is not player-facing HUD behavior and not normal gameplay UI.

The debug-only surface includes:

* Devtools window target selectors.
* Local player telemetry readouts.
* Target telemetry readouts.
* Game Target rows in player-only command selectors.
* All Players scope rows in supported player command selectors.
* Spawn player, asteroid, bullet, and pickup placement requests.
* Continuous bullet stream placement requests.
* Server hitbox visibility controls when used alongside targeting inspection.

This tooling must not introduce a parallel gameplay system. Debug actions that mutate gameplay still route to server devtools command handling and then through the server-owned gameplay seams.

## Server authority

Target debugging uses canonical server-owned target state.

The generic gameplay target identity is:

```text
target_kind
target_id
```

For player targets, the canonical values are:

```text
target_kind = "player"
target_id = <player id>
```

The client may request a target from local click state, but the request is not authoritative. The server validates the selected target kind, target ID, click position, and current target body before storing target state. The client then reads the confirmed target back through normal gameplay state.

Placement debugging also remains server-authoritative. The client sends a requested entity type, position, direction metadata, and optional player target context. The server decides whether the request applies and mutates the authoritative game state. The client observes the result through normal state packets and world sync.

`target_player_id` is a quarantined compatibility field for debug/player-only command paths. It is not the generic gameplay target model and should not be used for normal gameplay targeting, target readouts, or new target-capable gameplay systems.

## Client presentation

The devtools window presents target and placement debugging through ordinary controls and raw readouts.

Target presentation includes:

* A Game Target selector showing the current canonical target.
* Local player telemetry.
* Target telemetry.
* Player command selectors that may include All Players.
* Player command selectors that may include Game Target only when the canonical target is a player.

Target telemetry displays `target_kind` and `target_id` before the raw target state dictionary. Local and target telemetry can read from either active entity state or player session state, depending on the selected source.

Placement presentation is intentionally minimal. The current implementation does not maintain a separate placement marker or preview overlay. Placement is a temporary input mode: press or drag in the world, emit a placement result, send a devtools packet, then return to normal input handling.

## Target read models

Client devtools target read models are built from normalized gameplay state.

`DevtoolsPlayerTargetModel` reads:

```text
self_id
server_players
player_sessions
server_asteroids
server_bullets
server_pickups
server_enemies / enemies
player_lifecycle
debug_statuses
```

The local player's canonical target is read from the local player's synced active state:

```text
server_players[self_id].target_kind
server_players[self_id].target_id
```

If those generic fields are empty and legacy `target_player_id` is present, the model treats it as a compatibility player target for readback. New code should keep the generic fields as the primary model.

Target rows are built differently depending on command type:

* Raw target rows combine active player state and lifecycle state.
* Player-only command rows may include All Players first.
* Player-only command rows may include Game Target only when the canonical target is a player.
* Non-player canonical targets do not resolve into player-only command rows.
* Respawn rows include lifecycle-aware player rows.
* Feature rows show debug feature state such as `Active` or `Inactive`.

Target state lookup depends on the selected telemetry source:

```text
StatePacket.entities
= active player, asteroid, bullet, pickup, or enemy state

StatePacket.player_world_states
= durable player/session state for player targets only
```

Inactive player session state can be displayed for diagnostics, but displaying it does not make that player active, clickable, collidable, damageable, or targetable.

## Placement flow

Placement debugging starts from either a devtools hotkey or a devtools window button.

The high-level flow is:

```text
hotkey or window button
-> DevtoolsPlacementContext.request_placement_action
-> configured placement route
-> DevToolsSessionFlow.begin_debug_click_placement
-> DebugClickPlacementFlow or DebugContinuousBulletSpawnFlow
-> placement result
-> DevConnectionService
-> devtools packet
-> server devtools command handling
-> normal authoritative state readback
```

`DevtoolsPlacementContext` refuses placement requests until gameplay state has been received. It also refuses requests when no placement route has been configured.

`DebugMouseWorldPosition` reads the current global mouse position from the gameplay scene and converts that visual position to server coordinates through world sync. Placement packets use the server position, not raw screen position.

Normal click placement records the mouse position on press and completes on release. If the release position is far enough from the press position, the client includes a normalized direction vector.

The drag threshold is:

```text
8.0 visual units
```

Placement result fields include:

```text
action_name
server_position
visual_position
has_direction
direction
target_player_id
pickup_type
```

`target_player_id` is included only when the placement action was configured with a player target. `pickup_type` is included only for pickup placement.

Continuous bullet stream placement is separate from one-shot bullet placement. It requires a drag direction and sends a continuous-stream packet directly after the placement flow completes.

## Commands or controls

The current devtools hotkey placement controls are:

```text
DevToggle6
= spawn player placement

Shift + DevToggle6
= spawn asteroid placement

Alt + DevToggle6
= spawn bullet placement

Ctrl + Alt + DevToggle6
= continuous bullet stream placement

DevToggle7
= respawn local player
```

The devtools window exposes placement buttons for:

```text
Spawn Player
Spawn Asteroid
Spawn Bullet
Spawn Pickup
```

Spawn Player may target a specific player slot or create a new player. Spawn Pickup includes a pickup type selector populated from the client pickup presentation catalog, with `1_up` selected when available.

The placement packet boundary is:

```text
debug_spawn_entity
debug_spawn_pickup
debug_begin_continuous_bullet_stream
```

`debug_spawn_entity` carries:

```text
entity_type
x
y
has_direction
direction_x
direction_y
target_player_id
```

`entity_type` may be:

```text
player
asteroid
bullet
```

`debug_spawn_pickup` carries:

```text
pickup_type
x
y
```

`debug_begin_continuous_bullet_stream` carries:

```text
x
y
has_direction
direction_x
direction_y
```

Target commands use separate target request packets:

```text
select_target_at_position_request
clear_target_request
set_target_player_request
```

`select_target_at_position_request` is the generic gameplay target click request. `clear_target_request` clears the current target. `set_target_player_request` is a player-target compatibility request used by the devtools Game Target controls.

## Input priority

Active devtools placement has priority over normal gameplay targeting.

The input order is:

```text
devtools placement input
-> gameplay UI mouse protection
-> normal gameplay mouse actions
```

When placement is active, placement input consumes the relevant mouse event before generic target selection can run. This prevents one click from both placing a debug entity and selecting a gameplay target.

Normal gameplay mouse actions still use the semantic input path:

```text
SelectTarget
DeselectTarget
SpawnEntity
CancelAction
```

Pending placement actions own spawn/cancel input until they complete or cancel.

## Telemetry

Target and placement debugging uses a mix of raw readouts and logs.

Client readouts include:

* Local player state.
* Current canonical target kind and ID.
* Current target state.
* Debug feature status rows.
* Player lifecycle labels.
* Active player versus player session telemetry source selection.

Placement completion itself is not presented as a separate in-world confirmation UI. The client logs placement completion and packet-send details, then relies on normal authoritative state readback to show the spawned entity or resulting state change.

The target telemetry readout should display canonical `target_kind` and `target_id`. It should not display `target_player_id` as the generic target model.

## Build/runtime gates

Client-side devtools hotkeys are gated by `DevToolsBuildFlags`.

In public builds, configured DevToggle input events are removed from the Godot `InputMap`:

```text
DevToggle0
DevToggle1
DevToggle2
DevToggle3
DevToggle4
DevToggle5
DevToggle6
DevToggle7
DevToggle8
DevToggle9
```

Runtime gates also apply:

* Placement requests require received gameplay state.
* Placement requests require a configured placement route.
* Placement packet sends require a configured connection service.
* Placement packet builders return empty packets for unknown or incomplete placement results.
* Continuous bullet stream placement requires a nonzero drag direction.
* Mouse placement helpers are configured only when the gameplay scene has the required world-position conversion route.

Server-side devtools command handling is separately gated by the server devtools build configuration. `nodevtools` server builds disable server devtools command handling.

## Relationship to real gameplay implementation areas

Target and placement debugging must reuse real gameplay and protocol seams.

Normal target selection uses the same target request path as gameplay input:

```text
TargetRequestFlow
-> generated packet builder
-> client connection service
-> server inbound gameplay routing
-> game targeting validation
-> state packet readback
```

Placement debugging uses devtools packets, but the server still applies the effects through server-owned game/devtools adapters. Client placement code does not directly create authoritative entities.

World position conversion also uses the normal world sync conversion path. Placement packets carry server coordinates derived from current visual position and the active world sync transform.

The client should expose only the read models needed by devtools. Rendering, collision outlines, target telemetry, and placement input belong in client devtools and gameplay input seams; authoritative mutation belongs on the server.

## Code map

Primary client implementation paths:

* `client/scripts/devtools/gameplay_devtools_context.gd`
* `client/scripts/devtools/context/devtools_state_context.gd`
* `client/scripts/devtools/context/devtools_gameplay_state_context.gd`
* `client/scripts/devtools/context/devtools_placement_context.gd`
* `client/scripts/devtools/context/devtools_command_context.gd`
* `client/scripts/devtools/context/devtools_window_action_context.gd`
* `client/scripts/devtools/devtools_window.gd`
* `client/scripts/devtools/devtools_window_controller.gd`
* `client/scripts/devtools/devtools_display_refresh_flow.gd`
* `client/scripts/devtools/devtools_player_target_model.gd`
* `client/scripts/devtools/devtools_target_resolver.gd`
* `client/scripts/devtools/devtools_hotkey_flow.gd`
* `client/scripts/devtools/dev_connection_service.gd`
* `client/scripts/devtools/dev_spawn_packet_builder.gd`
* `client/scripts/devtools/dev_respawn_packet_builder.gd`
* `client/scripts/devtools/dev_tools_build_flags.gd`
* `client/scripts/devtools/dev_tools_session_flow.gd`
* `client/scripts/gameplay/devtools/debug_mouse_world_position.gd`
* `client/scripts/gameplay/devtools/debug_click_placement_flow.gd`
* `client/scripts/gameplay/devtools/debug_continuous_bullet_spawn_flow.gd`
* `client/scripts/gameplay/gameplay_composition.gd`
* `client/scripts/session/gameplay_session_controller.gd`

Related gameplay input and targeting paths:

* `client/scripts/gameplay/input/mouse_action_names.gd`
* `client/scripts/gameplay/input/mouse_action_mapper.gd`
* `client/scripts/gameplay/input/mouse_action_flow.gd`
* `client/scripts/gameplay/input/gameplay_input_context.gd`
* `client/scripts/gameplay/input/gameplay_pointer_position_provider.gd`
* `client/scripts/gameplay/targeting/gameplay_targeting_context.gd`
* `client/scripts/gameplay/targeting/gameplay_target_candidate_flow.gd`
* `client/scripts/gameplay/targeting/target_request_flow.gd`
* `client/scripts/gameplay/targeting/target_position_source.gd`
* `client/scripts/gameplay/targeting/target_pick_radius_resolver.gd`
* `client/scripts/world/world_sync.gd`

Related packet and server authority paths:

* `shared/packets/gameplay.toml`
* `shared/packets/debug.toml`
* `client/scripts/generated/networking/packets/packets.gd`
* `services/game-server/internal/networking/inbound/gameplay.go`
* `services/game-server/internal/game/targeting.go`
* `services/game-server/internal/devtools/handler.go`
* `services/game-server/internal/devtools/spawn_entity.go`
* `services/game-server/internal/devtools/spawn_player.go`
* `services/game-server/internal/devtools/spawn_asteroid.go`
* `services/game-server/internal/devtools/spawn_bullet.go`
* `services/game-server/internal/devtools/spawn_pickup.go`
* `services/game-server/internal/devtools/placement_requests.go`
* `services/game-server/internal/devtools/continuous_bullet_stream.go`
* `services/game-server/internal/devtools/streamruntime/`

## Tests

Relevant client tests include:

* `client/tests/unit/devtools/context/test_devtools_placement_context.gd`
* `client/tests/unit/test_devtools_target_resolver.gd`
* `client/tests/unit/test_devtools_player_target_model.gd`
* `client/tests/unit/test_devtools_window_controller.gd`
* `client/tests/unit/test_devtools_display_refresh_flow.gd`
* `client/tests/unit/test_gameplay_devtools_context.gd`
* `client/tests/unit/devtools/devtools_window_test.gd`
* `client/tests/unit/devtools/gameplay_debug_flow_test.gd`
* `client/tests/unit/test_mouse_action_mapper.gd`
* `client/tests/unit/test_mouse_action_flow.gd`
* `client/tests/unit/test_target_request_flow.gd`
* `client/tests/unit/test_target_visual_picker.gd`
* `client/tests/unit/gameplay/test_gameplay_target_candidate_flow.gd`

Relevant server-side verification includes:

* `services/game-server/internal/game/targeting_test.go`
* `services/game-server/internal/devtools/command_types_test.go`
* `services/game-server/internal/devtools/enabled_default_test.go`
* `services/game-server/internal/devtools/disabled_test.go`
* `services/game-server/internal/devtools/target_player_ids_test.go`
* `services/game-server/internal/devtools/toggles_test.go`
* `services/game-server/internal/devtools/player_counters_test.go`
* `services/game-server/internal/devtools/clear_entities_test.go`
* `services/game-server/internal/devtools/streamruntime/*_test.go`

## Related docs

* [Client Devtools](./!README.md)
* [Devtools](../!README.md)
* [Client Input And Targeting](../../services/client/input-and-targeting.md)
* [Game Server](../../services/game-server/!README.md)
* [Realtime Protocol](../../protocol/!README.md)
* [Packet Schemas](../../data/packet-schemas.md)
* [Canonical Target State](../../services/game-server/simulation/targeting/canonical-target-state.md)
* [Target Selection And Status](../../services/game-server/simulation/targeting/target-selection-and-status.md)

## Notes

The current placement flow is intentionally input-and-packet based. There is no separate placement preview layer in the client.

The key invariant is that target readouts may inspect canonical state, and placement tools may request debug mutations, but the server remains the authority for valid targets, spawned entities, and resulting gameplay state.
