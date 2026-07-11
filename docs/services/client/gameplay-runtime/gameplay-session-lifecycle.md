# Gameplay Session Lifecycle

Parent index: [Gameplay Runtime](./!INDEX.md)

## Purpose

This document describes the current client gameplay-session lifecycle implementation.

It covers how the Godot client begins accepting gameplay packets, bridges gameplay presentation fanout, resets gameplay state, and handles gameplay exits such as replay, return to lobby, return to pregame, and quit to main menu.

## Overview

The client gameplay-session lifecycle is owned by `GameplaySessionController` and coordinated with `SessionNetworkController`, `GameplayComposition`, and session boot/context state.

`SessionNetworkController` receives classified packet signals from `ClientConnectionService`. Room packets update the room session first. When the current room state becomes `InGame`, `SessionNetworkController` tells `GameplaySessionController` to begin accepting gameplay packets.

`GameplaySessionController` is the lifecycle bridge between the network/session layer and gameplay runtime composition. It owns the `accepts_gameplay_packets` gameplay-session/input/pause acceptance state, activates and resets the `PresentationBridge`, forwards player pause state into runtime only while that gate is open, sequences frame processing, forwards control and debug packets into gameplay composition, routes input through HUD and devtools policy, and runs gameplay composition processing each frame.

Gameplay exits are routed back through `GameplayComposition` signals. `GameplaySessionController` translates those signals into connection actions, reset behavior, session-context clearing, boot-flow clearing, main-menu visibility updates, and higher-level replay or pregame-return signals.

This lifecycle is client presentation/session orchestration only. The server remains authoritative for room state, match lifecycle, active gameplay state, match-over status, and gameplay outcomes.

## Code root

* `client/scripts/`

## Responsibilities

* Configure gameplay composition from scene, network, session, HUD, world, and UI references.
* Gate gameplay input and player-pause forwarding behind `accepts_gameplay_packets`; realtime lane packet routing continues through `RealtimePacketPipeline` independently.
* Begin accepting gameplay packets after room state enters `InGame`.
* Activate and flush gameplay presentation bridging only when both the gate and pipeline readiness allow it.
* Forward player pause packets into gameplay composition.
* Forward devtools debug status packets into gameplay composition.
* Forward debug shape catalog packets into gameplay composition.
* Run gameplay composition processing from `_process`.
* Normalize room-state strings used by client session logic.
* Extract room state from packet data with fallback room-state behavior.
* Classify room states that should stop spectating.
* Route devtools input before normal gameplay input.
* Apply HUD/gameplay UI mouse gating before gameplay input handling.
* Reset gameplay packet acceptance, presentation bridge state, and gameplay composition.
* Hide the main menu when gameplay starts.
* Handle quit-to-main-menu by beginning graceful network close, resetting gameplay, clearing session context, clearing boot flow, and showing the main menu.
* Handle return-to-lobby by sending a return-to-lobby request and resetting local gameplay state.
* Handle return-to-pregame by beginning graceful network close, resetting gameplay, clearing session context, clearing boot flow, and emitting `return_to_pregame_requested`.
* Handle replay by waiting for graceful close, resetting gameplay, clearing session context, clearing boot flow, and emitting `replay_requested`.
* Refresh match-end state after room state changes.

## Does not own

* Room membership authority.
* Server room-state authority.
* Server match lifecycle authority.
* Server gameplay simulation.
* Match-over decisions.
* Match-result authority.
* Durable player data.
* Packet schema source of truth.
* Raw WebSocket transport.
* Lane packet normalization details.
* World entity rendering or interpolation.
* Menu, HUD, input, match-end, or devtools internals beyond lifecycle routing.
* Respawn policy.

## Domain roles

### Gameplay session activation

`GameplaySessionController` owns the local client gate that decides whether gameplay input, player pause packets, and gameplay-session activation are allowed.

The gate starts closed. `SessionNetworkController` opens it by calling `begin_accepting_gameplay_packets()` when room state reaches `Constants.ROOM_STATE_IN_GAME`.

This activates the `PresentationBridge` and allows gameplay-session flow to proceed, while `RealtimePacketPipeline` continues to own realtime gameplay packet application.

### Session lifecycle bridge

`GameplaySessionController` bridges gameplay presentation signals into session actions.

It does not decide whether the server match is over, whether a room may return to lobby, or whether a replay is valid. It only performs the client-side transition work once the gameplay presentation flow emits the relevant request.

### Gameplay reset owner

`GameplaySessionController.reset()` clears the local gameplay-session gate, resets the presentation bridge, and resets gameplay composition state. It does not depend on a `GameplayStateFlow` readiness holder.

`GameplayComposition.reset()` then clears devtools session state, shell/runtime state, match-end state, match-results presentation, and spectate state.

Reset also clears any deferred presentation fanout for the current gameplay session. After reset, gameplay packets remain ignored until `begin_accepting_gameplay_packets()` is called again and the realtime pipeline reports gameplay readiness.

### Main-menu visibility bridge

The controller hides the main menu on `gameplay_started`.

On quit-to-main-menu it resets gameplay/session state and shows the main menu again.

### Room-state helpers

`RoomState` normalizes and classifies Lobby, InGame, and GameOver room-state string variants for client session logic.

`GameplayRoomStateFlow` extracts room state from packet dictionaries and delegates game-over classification to `RoomState`.

These helpers support client session flow, but they do not own server room-state authority.

## Protocols and APIs

### Room state to gameplay session activation

Room-state packets are routed through `SessionNetworkController`.

Current flow:

```text
ClientConnectionService.room_snapshot_received
-> SessionNetworkController._on_room_snapshot_received
-> RoomSessionController.handle_room_snapshot
-> GameplaySessionController.begin_accepting_gameplay_packets when room state is InGame
```

```text
ClientConnectionService.room_state_changed
-> SessionNetworkController._on_room_state_changed
-> RoomSessionController.handle_room_state_changed
-> GameplaySessionController.begin_accepting_gameplay_packets when room state is InGame
```

Both paths also refresh match-end state after room state is applied.

### Gameplay presentation bridge

```text
ServerPacketDispatcher
-> ClientInboundCoordinator typed realtime binding
-> RealtimePacketPipeline typed apply entry point
-> RealtimeRouter applies the packet
-> RealtimePresentationState is refreshed
-> RealtimePacketPipeline.gameplay_packet_applied(packet)
-> PresentationBridge.handle_gameplay_packet(packet)
-> inactive bridge ignores routed-packet notifications for presentation scheduling
```

`RealtimePacketPipeline` applies/routes realtime packets regardless of gameplay-session activation and owns the refreshed realtime presentation state.

If `accepts_gameplay_packets` is false, `GameplaySessionController` does not activate or schedule `PresentationBridge`, but realtime lane packets are still received and routed by `RealtimePacketPipeline`. A lifecycle packet may apply, queue, or reject before the historically named `gameplay_packet_applied` notification; that notification means routing and presentation-state refresh completed, not that lifecycle state necessarily mutated.

If gameplay readiness is not yet true, presentation orchestration is skipped even though `ServerPacketDispatcher` has already delivered the packet to `RealtimePacketPipeline` and the pipeline-owned `RealtimeRouter` has routed it. Packet routing and available state application may still occur while gameplay presentation is inactive.

`RealtimePacketPipeline.is_gameplay_ready()` determines whether the client has the required realtime baseline for gameplay presentation. `RealtimePacketPipeline.get_presentation_state()` returns the applied state that gameplay composition later consumes when presentation is allowed.

Packet routing/application and gameplay handoff are separate boundaries. `RealtimePacketPipeline` owns packet routing/application and readiness; `PresentationBridge` owns presentation orchestration; gameplay composition owns the downstream presentation targets.

`RealtimePacketPipeline.gameplay_packet_applied(packet)` hands the routed-packet notification directly to `PresentationBridge.handle_gameplay_packet(packet)` for presentation scheduling when the bridge is active. `GameplaySessionController` does not connect this signal to `GameplayComposition`.

### Active and inactive input gating

* `GameplaySessionController` only forwards player pause packets while `accepts_gameplay_packets` is true.

When the gate is inactive, the controller does not forward gameplay input or player-pause packets and does not activate/schedule the bridge. It does not block realtime lane packets from reaching `RealtimePacketPipeline`; debug packets remain routed independently of this gate.

### Player pause packets

Player pause packets use the `accepts_gameplay_packets` input/pause acceptance gate; realtime lane packets do not.

```text
ClientConnectionService.player_pause_state_received
-> SessionNetworkController._on_player_pause_state_received
-> GameplaySessionController.handle_player_pause_state
-> GameplayComposition.apply_player_pause_state_packet
```

If `accepts_gameplay_packets` is false, the player-pause packet is not forwarded to gameplay composition. Realtime lane packet routing remains independent of this input/pause gate.

### Debug packets

Debug packets route through gameplay composition regardless of `accepts_gameplay_packets`.

```text
ClientConnectionService.debug_status_received
-> SessionNetworkController._on_debug_status_received
-> GameplaySessionController.handle_debug_status_packet
-> GameplayComposition.apply_devtools_debug_status_packet
```

```text
ClientConnectionService.debug_shape_catalog_received
-> SessionNetworkController._on_debug_shape_catalog_received
-> GameplaySessionController.handle_debug_shape_catalog_packet
-> GameplayComposition.apply_debug_shape_catalog_packet
```

Debug command authority remains server/devtools-owned. The gameplay-session lifecycle only forwards client presentation data.

### Replay

Replay is emitted from gameplay composition and handled by `GameplaySessionController._on_gameplay_replay_requested()`.

Current behavior:

```text
1. Log gameplay replay request.
2. Await connection_service.close_gracefully() when available.
3. Reset gameplay lifecycle state.
4. Clear session context.
5. Clear shell boot flow.
6. Emit replay_requested.
```

The replay path waits for graceful close before emitting `replay_requested`.

### Return to lobby

Return to lobby is emitted from gameplay composition and handled by `GameplaySessionController._on_gameplay_return_to_lobby_requested()`.

Current behavior:

```text
1. Log gameplay return-to-lobby request.
2. Send connection_service.send_return_to_lobby_request().
3. Reset local gameplay lifecycle state.
```

This path sends a server request instead of locally deciding lobby return authority.

### Return to pregame

Return to pregame is emitted with a session mode and handled by `GameplaySessionController._on_gameplay_return_to_pregame_requested(session_mode)`.

Current behavior:

```text
1. Log gameplay return-to-pregame request.
2. Begin graceful network close when available.
3. Reset gameplay lifecycle state.
4. Clear session context.
5. Clear shell boot flow.
6. Emit return_to_pregame_requested(session_mode).
```

### Quit to main menu

Quit to main menu is emitted from gameplay composition and handled by `GameplaySessionController._on_gameplay_quit_to_main_menu_requested()`.

Current behavior:

```text
1. Log gameplay quit-to-main-menu request.
2. Begin graceful network close.
3. Reset gameplay lifecycle state.
4. Clear session context.
5. Clear shell boot flow.
6. Show main menu.
```

## Data ownership

The gameplay-session lifecycle owns transient client state only.

Owned local state includes:

* `accepts_gameplay_packets` for gameplay input, player-pause forwarding, and `PresentationBridge` activation/scheduling
* references to connection service, HUD, gameplay UI, main menu, session context, shell boot flow, and logger
* gameplay composition reference
* lifecycle signals for replay and return-to-pregame requests

The lifecycle does not persist data.

The lifecycle does not own durable player identity, account state, local profile data, match result records, room membership, or authoritative game state.

## Code map

### Primary lifecycle files

* `client/scripts/session/gameplay_session_controller.gd`
* `client/scripts/session/session_network_controller.gd`
* `client/scripts/gameplay/gameplay_composition.gd`
* `client/scripts/gameplay/session/gameplay_session_state.gd`

### Packet and connection participants

* `client/scripts/networking/client_connection_service.gd`
* `client/scripts/networking/inbound/server_packet_dispatcher.gd`
* `client/scripts/networking/inbound/server_packet_router.gd`

### Runtime lane state participants

* `client/scripts/protocol/realtime/realtime_router.gd`
* `client/scripts/protocol/realtime/presentation_adapter.gd`
* `client/scripts/protocol/realtime/devtools_lane_state_adapter.gd`
* `client/scripts/session/room_state.gd`
* `client/scripts/gameplay/session/gameplay_room_state_flow.gd`

### Exit and presentation participants

* `client/scripts/gameplay/match_end/match_end_flow.gd`
* `client/scripts/shell/gameplay_menu_flow.gd`
* `client/scripts/shell/gameplay_shell_flow.gd`
* `client/scripts/boot/shell_boot_flow.gd`
* `client/scripts/session/client_session_context.gd`

### Non-ownership boundaries

* `services/game-server/internal/rooms/` owns server room state and room lifecycle.
* `services/game-server/internal/game/` owns authoritative gameplay simulation and match lifecycle.
* `client/scripts/networking/network_client.gd` owns raw WebSocket transport.
* `shared/packets/` owns gameplay packet source definitions.
* `client/scripts/world/` owns world rendering and entity sync.

## Tests

Relevant client tests include:

* `client/tests/unit/test_gameplay_session_controller.gd`
* `client/tests/unit/test_gameplay_session_state.gd`
* `client/tests/unit/test_session_network_controller.gd`
* `client/tests/unit/test_gameplay_room_state_flow.gd`
* `client/tests/unit/protocol/realtime/test_lane_native_presentation_adapters.gd`
* `client/tests/unit/protocol/realtime/test_gameplay_readiness.gd`
* `client/tests/unit/gameplay/match_end/test_match_end_flow.gd`
* `client/tests/unit/shell/test_gameplay_menu_flow.gd`

`test_gameplay_session_controller.gd` currently verifies that replay waits for graceful close before emitting `replay_requested`.

`test_gameplay_session_state.gd` verifies helper behavior for gameplay packet processing and game-over classification.

`test_session_network_controller.gd` verifies connection/auth boot routing behavior that precedes gameplay-session packet acceptance.

## Related docs

* [Gameplay Runtime](./!INDEX.md)
* [Runtime composition](runtime-composition.md)
* [Gameplay state application](gameplay-state-application.md)
* [Runtime processing](runtime-processing.md)
* [Presentation Bridge](presentation-bridge.md) - Applied notification handling, pending/coalescing, readiness-gated flush, and presentation orchestration.
* [Menu flow](../menu-flow.md) - Client menu flow documentation.
* [Match End Flow](../match-end-flow/!INDEX.md) - Client match-end orchestration and match-results presentation documentation.
* [Gameplay Menu Flow](../gameplay-menu-flow/!INDEX.md) - Client gameplay menu and match-over overlay menu documentation.
* [Gameplay packets](../../../protocol/gameplay-packets.md) - Gameplay lane packet documentation.

## Notes

`GameplaySessionState.can_process_gameplay_packets()` allows blank room state, `InGame`, and `GameOver`, but the current `GameplaySessionController` packet gate is opened explicitly by `begin_accepting_gameplay_packets()` when room state reaches `InGame`.

`RealtimePacketPipeline` owns readiness and presentation-state access for active gameplay presentation orchestration. `PresentationBridge` consumes `is_gameplay_ready()` and `get_presentation_state()`; `GameplaySessionController` owns bridge activation, reset, and flush timing.

`accepts_gameplay_packets` is the gameplay-session/input/pause acceptance state that activates and resets `PresentationBridge` and gates gameplay input and player-pause forwarding. It does not gate `RealtimePacketPipeline` lane packet routing.

Dead-HUD recovery, alive-presentation restoration, and respawn-facing presentation do not belong to `GameplaySessionController`; they route through gameplay composition and the focused runtime flows behind it.

Return-to-lobby intentionally sends a server request and then resets local gameplay state. It does not locally force room membership or room state.

Replay uses `close_gracefully()` and awaits completion before emitting `replay_requested`; quit-to-main-menu and return-to-pregame use `begin_graceful_close()` and continue local cleanup immediately.