# Gameplay Session Lifecycle

Parent index: [Gameplay Runtime](./!README.md)

## Purpose

This document describes the current client gameplay-session lifecycle implementation.

It covers how the Godot client begins accepting gameplay packets, routes gameplay-session packet families, resets gameplay presentation state, and handles gameplay exits such as replay, return to lobby, return to pregame, and quit to main menu.

## Overview

The client gameplay-session lifecycle is owned by `GameplaySessionController` and coordinated with `SessionNetworkController`, `GameplayComposition`, and session boot/context state.

`SessionNetworkController` receives classified packet signals from `ClientConnectionService`. Room packets update the room session first. When the current room state becomes `InGame`, `SessionNetworkController` tells `GameplaySessionController` to begin accepting gameplay packets.

`GameplaySessionController` is the lifecycle bridge between the network/session layer and gameplay runtime composition. It owns the `accepts_gameplay_packets` gate, forwards gameplay state and player pause state into runtime only while that gate is open, forwards debug packets to gameplay composition, and runs gameplay composition processing each frame.

Gameplay exits are routed back through `GameplayComposition` signals. `GameplaySessionController` translates those signals into connection actions, reset behavior, session-context clearing, boot-flow clearing, main-menu visibility updates, and higher-level replay or pregame-return signals.

This lifecycle is client presentation/session orchestration only. The server remains authoritative for room state, match lifecycle, active gameplay state, match-over status, and gameplay outcomes.

## Code root

* `client/scripts/`

## Responsibilities

* Configure gameplay composition from scene, network, session, HUD, world, and UI references.
* Create and connect `GameplayStateFlow`.
* Gate gameplay state and player pause packets behind `accepts_gameplay_packets`.
* Begin accepting gameplay packets after room state enters `InGame`.
* Forward gameplay state packets into `GameplayStateFlow`.
* Forward player pause packets into gameplay composition.
* Forward devtools debug status packets into gameplay composition.
* Forward debug shape catalog packets into gameplay composition.
* Run gameplay composition processing from `_process`.
* Route devtools input before normal gameplay input.
* Apply HUD/gameplay UI mouse gating before gameplay input handling.
* Reset gameplay packet acceptance, gameplay state flow, and gameplay composition.
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
* Gameplay state normalization details.
* World entity rendering or interpolation.
* Menu, HUD, input, match-end, or devtools internals beyond lifecycle routing.

## Domain roles

### Gameplay packet gate

`GameplaySessionController` owns the local client gate that decides whether gameplay and pause packets should be applied.

The gate starts closed. `SessionNetworkController` opens it by calling `begin_accepting_gameplay_packets()` when room state reaches `Constants.ROOM_STATE_IN_GAME`.

This prevents gameplay packet application before the client has entered the gameplay session.

### Session lifecycle bridge

`GameplaySessionController` bridges gameplay presentation signals into session actions.

It does not decide whether the server match is over, whether a room may return to lobby, or whether a replay is valid. It only performs the client-side transition work once the gameplay presentation flow emits the relevant request.

### Gameplay reset owner

`GameplaySessionController.reset()` clears the local gameplay-session gate and delegates deeper cleanup to `GameplayStateFlow` and `GameplayComposition`.

`GameplayComposition.reset()` then clears devtools session state, shell/runtime state, presentation state, match-end state, match-results presentation, and spectate state.

### Main-menu visibility bridge

The controller hides the main menu on `gameplay_started`.

On quit-to-main-menu it resets gameplay/session state and shows the main menu again.

## Protocols and APIs

### Room state to gameplay packet acceptance

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

### Gameplay state packets

Gameplay state packets are forwarded only when the gameplay packet gate is open.

```text
ClientConnectionService.gameplay_state_received
-> SessionNetworkController._on_gameplay_state_received
-> GameplaySessionController.handle_gameplay_state
-> GameplayStateFlow.handle_gameplay_state_packet
```

If `accepts_gameplay_packets` is false, the packet is ignored by `GameplaySessionController`.

### Player pause packets

Player pause packets use the same acceptance gate as gameplay state packets.

```text
ClientConnectionService.player_pause_state_received
-> SessionNetworkController._on_player_pause_state_received
-> GameplaySessionController.handle_player_pause_state
-> GameplayComposition.apply_player_pause_state_packet
```

If `accepts_gameplay_packets` is false, the packet is ignored.

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

* `accepts_gameplay_packets`
* references to connection service, HUD, gameplay UI, main menu, session context, shell boot flow, and logger
* gameplay composition reference
* gameplay state flow reference
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

### Runtime state participants

* `client/scripts/gameplay/state/gameplay_state_flow.gd`
* `client/scripts/gameplay/state/gameplay_state_packet_reader.gd`
* `client/scripts/gameplay/state/gameplay_state_apply_flow.gd`

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
* `client/tests/unit/test_gameplay_state_packet_reader.gd`
* `client/tests/unit/test_gameplay_state_apply_flow.gd`
* `client/tests/unit/gameplay/match_end/test_match_end_flow.gd`
* `client/tests/unit/shell/test_gameplay_menu_flow.gd`

`test_gameplay_session_controller.gd` currently verifies that replay waits for graceful close before emitting `replay_requested`.

`test_gameplay_session_state.gd` verifies helper behavior for gameplay packet processing and game-over classification.

`test_session_network_controller.gd` verifies connection/auth boot routing behavior that precedes gameplay-session packet acceptance.

## Related docs

* [Gameplay Runtime](./!README.md)
* [Runtime composition](runtime-composition.md)
* [Gameplay state application](gameplay-state-application.md)
* [Runtime processing](runtime-processing.md)
* [Menu flow](../menu-flow.md) - Client menu flow documentation.
* [Match End Flow](../match-end-flow/!README.md) - Client match-end orchestration and match-results presentation documentation.
* [Gameplay Menu Flow](../gameplay-menu-flow/!README.md) - Client gameplay menu and match-over overlay menu documentation.
* [Gameplay packets](../../../protocol/stubs/gameplay-packets.md) - Stub: gameplay realtime packet documentation.

## Notes

`GameplaySessionState.can_process_gameplay_packets()` allows blank room state, `InGame`, and `GameOver`, but the current `GameplaySessionController` packet gate is opened explicitly by `begin_accepting_gameplay_packets()` when room state reaches `InGame`.

Return-to-lobby intentionally sends a server request and then resets local gameplay state. It does not locally force room membership or room state.

Replay uses `close_gracefully()` and awaits completion before emitting `replay_requested`; quit-to-main-menu and return-to-pregame use `begin_graceful_close()` and continue local cleanup immediately.
