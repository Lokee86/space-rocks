# Room Session State

Parent index: [App Shell And Session](./!INDEX.md)

## Purpose

This document describes the current client room-session state implementation.

It covers how the Godot client stores transient room state, applies room snapshots, exposes room-state and match-result providers to gameplay flows, and keeps client session routing separate from server room authority.

## Overview

Client room-session state is owned by `RoomSessionController`.

`RoomSessionController` sits at the app-shell/session boundary. It receives room packet handoff from `SessionNetworkController`, delegates lobby-specific snapshot presentation to lobby flows, caches the latest room state, caches the latest match-result payload, and exposes small provider methods used by gameplay runtime and match-end presentation.

The controller does not own authoritative room state. The game server owns room membership, room lifecycle, room state transitions, match-over decisions, and match-result creation. The client only caches server-observed facts so local UI and gameplay presentation can react consistently.

Current room-session flow:

```text
ClientConnectionService room signal
-> SessionNetworkController room handler
-> RoomSessionController
-> LobbyShellFlow / LobbyFlow / LobbySessionState
-> local room-state and match-result providers
-> GameplaySessionController / GameplayComposition / MatchEndFlow
```

Room snapshots are the broad state update path. They update lobby session data, activate the requested session mode, optionally mount or clear multiplayer lobby presentation, cache the latest room state, cache a valid match-result payload, and trigger client config sending when the room reaches `InGame`.

Room-state-change packets are narrower updates. They update only `latest_room_state` and log the change.

## Code root

```text
client/scripts/
```

## Responsibilities

`RoomSessionController` owns:

* Creating and wiring lobby session collaborators.
* Receiving room snapshot packets from `SessionNetworkController`.
* Receiving room-state-change packets from `SessionNetworkController`.
* Delegating room snapshot application to `LobbyShellFlow`.
* Caching the latest room state observed by the client.
* Caching the latest valid match-result payload from room snapshots.
* Clearing cached match-result data when a snapshot has no valid result.
* Exposing the current room state through `current_room_state()`.
* Exposing the current match result through `current_match_result()`.
* Exposing room capacity through `current_max_players()`.
* Sending client config when a room snapshot transitions the client into `InGame`.
* Forwarding room errors to multiplayer dialog presentation.
* Clearing session context and shell boot flow after lobby leave return.
* Configuring the lobby-leave return destination supplied by app/menu flow.

## Does not own

`RoomSessionController` does not own:

* Server room authority.
* Room membership decisions.
* Room capacity rules.
* Room lifecycle transitions.
* Match-over decisions.
* Match result calculation.
* Match result persistence.
* Raw WebSocket transport.
* Packet classification.
* Packet schema source of truth.
* Gameplay simulation.
* Gameplay packet acceptance policy.
* Match-end presentation orchestration.
* Match-result window presentation.
* Main menu or pregame menu route policy.
* Durable player profile or account data.

## Domain roles

### Client room-session cache

`RoomSessionController` keeps the client’s transient copy of room-session facts needed by app shell, lobby presentation, gameplay lifecycle, and match-end presentation.

The cache currently includes:

```text
latest_room_state
latest_match_result
```

`latest_room_state` is updated from both full room snapshots and room-state-change packets.

`latest_match_result` is updated only from room snapshots. A match result is cached only when the payload is a dictionary with a non-empty `match_id`.

### Lobby session bridge

Room snapshots are delegated to `LobbyShellFlow`.

`LobbyShellFlow` applies the snapshot to `LobbyFlow`, which updates `LobbySessionState` using `LobbyPacketReader`.

Current lobby session state includes:

```text
room_code
room_state
local_player_id
owner_id
max_players
members
```

`LobbySessionState` also owns derived lobby helpers such as local-owner checks, all-members-ready checks, and can-start-game checks.

### Session mode activation bridge

After a room snapshot is applied, `LobbyShellFlow` calls:

```text
session_context.activate_requested_mode()
```

The active session mode then determines whether a multiplayer lobby should be shown.

Current condition:

```text
active_mode == Multiplayer
and room_state == Lobby
```

Single-player room snapshots therefore update session state without mounting multiplayer lobby UI.

### Gameplay lifecycle provider

`AppEntry` wires `RoomSessionController.current_room_state()` into `GameplaySessionController`.

`GameplaySessionController` passes the provider through gameplay composition to match-end flows. This keeps gameplay runtime from owning room packet handling directly.

Room state also controls when gameplay packets become acceptable. `SessionNetworkController` checks `RoomSessionController.current_room_state()` after room snapshots and room-state-change packets. When the state is `InGame`, it calls:

```text
GameplaySessionController.begin_accepting_gameplay_packets()
```

### Match-result provider

`AppEntry` wires `RoomSessionController.current_match_result()` into `GameplaySessionController`.

That provider reaches `MatchEndFlow`, which reads cached result data when authoritative room state becomes `GameOver`.

The client result path is:

```text
room_snapshot.match_result
-> RoomSessionController.latest_match_result
-> current_match_result provider
-> MatchEndFlow
-> MatchResultsFlow
```

`RoomSessionController` does not parse result rows for UI display. It only caches the server-provided payload.

### Client config trigger

When `handle_room_snapshot()` sees the current room state become `InGame`, it calls the configured client config sender.

Current sender:

```text
ClientConfigController.send_client_config
```

This keeps the room-session controller aware of the session transition without making it own client config construction.

### Lobby leave cleanup

When lobby leave returns through `LobbyReturnFlow`, `RoomSessionController._on_lobby_left_room()` clears:

```text
session_context
shell_boot_flow
```

This removes stale requested/active session state and pending boot state after leaving the lobby.

## Protocols and APIs

### Room snapshot input

Room snapshots arrive from client networking signals.

```text
ClientConnectionService.room_snapshot_received
-> SessionNetworkController._on_room_snapshot_received
-> RoomSessionController.handle_room_snapshot
```

`handle_room_snapshot()` performs this sequence:

```text
1. Apply the snapshot through LobbyShellFlow.
2. Read LobbyFlow current state.
3. Cache state.room_state as latest_room_state.
4. Cache or clear match result from the snapshot.
5. If state is InGame, send client config when configured.
```

After `RoomSessionController` handles the snapshot, `SessionNetworkController` opens gameplay packet acceptance when the room state is `InGame` and refreshes match-end state.

### Room-state-change input

Room-state-change packets arrive from client networking signals.

```text
ClientConnectionService.room_state_changed
-> SessionNetworkController._on_room_state_changed
-> RoomSessionController.handle_room_state_changed
```

`handle_room_state_changed()` reads the room state field, updates `latest_room_state` when non-empty, and logs the state change.

Afterward, `SessionNetworkController` opens gameplay packet acceptance when the room state is `InGame` and refreshes match-end state.

### Room error input

Room errors route to the room-session controller for lobby/dialog presentation.

```text
ClientConnectionService.room_error_received
-> SessionNetworkController._on_room_error_received
-> RoomSessionController.handle_room_error
-> MultiplayerDialogStatusPresenter.show_room_error
```

The controller logs the server error code and message, then delegates presentation.

### Match result cache rule

A room snapshot result is cached only when:

```text
packet.match_result is Dictionary
and match_result.match_id is not empty
```

If the result field is missing, empty, non-dictionary, or has an empty `match_id`, the cached result is cleared.

This prevents stale match results from surviving later lobby snapshots.

### Provider APIs

`RoomSessionController` exposes these local provider methods:

```text
current_room_state() -> String
current_match_result() -> Dictionary
current_max_players() -> int
```

`current_room_state()` returns `latest_room_state` first. If no latest state exists and `lobby_flow` is available, it falls back to `lobby_flow.current_state().room_state`.

`current_match_result()` returns the cached dictionary or `{}`.

`current_max_players()` returns the max players from current lobby state, or `0` when lobby flow is unavailable.

## Data ownership

The room-session controller owns only transient client runtime state.

Owned local state:

```text
latest_room_state
latest_match_result
lobby_flow
lobby_network_actions
lobby_return_flow
lobby_shell_flow
multiplayer_lobby_presenter
multiplayer_dialog_status_presenter
```

Held references:

```text
main_menu
canvas_layer
session_context
connection_service
shell_boot_flow
client_config_sender
logger
```

The controller does not persist room state or match results.

Authoritative room and match-result data originate from the game server. The client cache exists only to make local presentation and lifecycle decisions react to the latest observed server packets.

## Code map

### Primary implementation

* `client/scripts/session/room_session_controller.gd` - Client room-session state controller, room snapshot handling, room-state cache, match-result cache, room error presentation handoff, lobby leave cleanup, and provider methods.

### App-shell wiring

* `client/scripts/shell/app_entry.gd` - Creates `RoomSessionController`, configures dependencies, wires client config sender, wires room-state/match-result/max-player providers into `GameplaySessionController`, and connects room packet signals through `SessionNetworkController`.
* `client/scripts/session/session_network_controller.gd` - Receives room packet signals from the connection service, delegates to `RoomSessionController`, opens gameplay packet acceptance when room state reaches `InGame`, and refreshes match-end state.
* `client/scripts/session/client_session_context.gd` - Tracks requested and active session mode and decides whether multiplayer lobby should be shown for a room state.

### Lobby collaborators

* `client/scripts/lobby/lobby_shell_flow.gd` - Applies room snapshots to lobby state, activates requested session mode, shows or clears multiplayer lobby presentation, and routes lobby button callbacks.
* `client/scripts/lobby/lobby_flow.gd` - Applies room snapshot fields to `LobbySessionState`.
* `client/scripts/lobby/lobby_session_state.gd` - Stores room code, room state, local player id, owner id, max players, and member list; owns derived lobby readiness helpers.
* `client/scripts/lobby/lobby_packet_reader.gd` - Reads room snapshot fields from packet dictionaries using generated packet constants.
* `client/scripts/lobby/lobby_network_actions.gd` - Sends ready, start-game, and leave-room requests for lobby UI callbacks.
* `client/scripts/lobby/lobby_return_flow.gd` - Owns local return behavior after lobby leave.
* `client/scripts/lobby/multiplayer_lobby_presenter.gd` - Owns multiplayer lobby presentation mounting and updates.
* `client/scripts/lobby/multiplayer_dialog_status_presenter.gd` - Owns room-error dialog presentation.

### Networking participants

* `client/scripts/networking/client_connection_service.gd` - Emits classified room signals after packet dispatch.
* `client/scripts/networking/inbound/server_packet_dispatcher.gd` - Classifies inbound server packets into room, gameplay, debug, telemetry, auth, and unknown packet signals.
* `client/scripts/networking/inbound/server_packet_router.gd` - Routes packet dictionaries by packet type before dispatcher emission.
* `client/scripts/networking/network_client.gd` - Owns raw WebSocket transport.

### Gameplay and match-end consumers

* `client/scripts/session/gameplay_session_controller.gd` - Receives room-state, match-result, and max-player providers; opens gameplay packet acceptance; refreshes match-end state.
* `client/scripts/gameplay/gameplay_composition.gd` - Passes room-state and match-result providers into gameplay presentation flows.
* `client/scripts/gameplay/match_end/match_end_flow.gd` - Reads room state and match result through providers to present authoritative room match-over and results.
* `client/scripts/ui/match_results/match_results_flow.gd` - Presents match result rows after `MatchEndFlow` adapts cached result payload.

### Generated inputs

* `client/scripts/generated/constants/constants.gd` - Generated constants for room states and session modes.
* `client/scripts/generated/networking/packets/packets.gd` - Generated packet field constants for room state, room code, members, max players, match result, and match id fields.

### Server authority boundary

* `services/game-server/internal/networking/room_snapshot.go` - Builds authoritative room snapshot packets and includes resolved match-result summaries when available.
* `services/game-server/internal/rooms/` - Owns authoritative room membership and room lifecycle.
* `services/game-server/internal/game/` - Owns authoritative gameplay simulation and match summary data.

## Tests

Relevant tests include:

* `client/tests/unit/test_room_session_controller.gd`
* `client/tests/unit/test_session_network_controller.gd`
* `client/tests/unit/test_gameplay_session_controller.gd`
* `services/game-server/internal/networking/room_snapshot_test.go`
* `services/game-server/tests/networking/room_snapshot_test.go`

`test_room_session_controller.gd` verifies:

* Lobby return cleanup clears session context and shell boot flow.
* Lobby leave return destination configuration reaches `LobbyReturnFlow`.
* Valid match-result payloads are cached from room snapshots.
* Empty match-result objects clear cached result data.
* Missing match-result fields clear cached result data.

`test_session_network_controller.gd` currently verifies connection and auth boot routing behavior. It is adjacent to room-session routing because the same controller owns room packet handoff, but the listed tests do not currently cover room snapshot delegation.

Game-server room snapshot tests verify that room snapshots include room state, capacity, member readiness, and resolved match-result data when available.

## Related docs

* [App Shell And Session](./!INDEX.md)
* [Client](../!INDEX.md)
* [Networking Flow](../networking-flow/!INDEX.md)
* [Lobby Flow](../lobby-flow/!INDEX.md)
* [Gameplay Runtime](../gameplay-runtime/!INDEX.md)
* [Gameplay Session Lifecycle](../gameplay-runtime/gameplay-session-lifecycle.md)
* [Match End Flow](../match-end-flow/!INDEX.md)
* [Match End Orchestration](../match-end-flow/match-end-orchestration.md)
* [Match Results Presentation](../match-end-flow/match-results-presentation.md)
* [Menu Flow](../menu-flow.md)
* [Auth Session Flow](../auth-session-flow.md)
* [Realtime WebSocket Protocol](../../../protocol/realtime-websocket-protocol.md) - Stub: realtime websocket protocol documentation.

## Notes

The legacy client docs mixed room-session caching into broader menu-flow and match-end UI topics. The current documentation split keeps this file focused on the app-shell/session room-state cache and its immediate consumers.

`RoomSessionController` currently constructs several lobby collaborators internally. That makes the controller both the room-session cache owner and the lobby shell composition point. Lobby-specific behavior should remain documented under lobby-flow docs; this file should document only the app-shell/session ownership boundary.

`latest_match_result` is intentionally cleared when a later room snapshot does not contain a valid result. This prevents old match results from appearing after the room returns to lobby or a new session begins.

The room-session cache should not be treated as a durable model. It is a local reflection of the latest observed server packets.
