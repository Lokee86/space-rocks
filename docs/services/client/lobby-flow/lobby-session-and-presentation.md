# Lobby Session and Presentation

Parent index: [Lobby Flow](./!INDEX.md)

## Purpose

This document describes the current client lobby session and multiplayer lobby presentation implementation.

It covers how the Godot client applies authoritative room snapshots into local lobby state, decides when to show multiplayer lobby UI, presents member readiness and room status, and routes lobby button intent back to the realtime networking layer.

## Overview

The client lobby flow is a presentation and request layer around server-owned room state.

The game server owns room membership, owner selection, readiness authority, room lifecycle transitions, room capacity, and game start acceptance. The client stores a transient lobby read model from `room_snapshot` packets and uses that read model to update the multiplayer lobby window.

Current lobby session flow:

```text
room_snapshot packet
-> ClientConnectionService.room_snapshot_received
-> SessionNetworkController
-> RoomSessionController
-> LobbyShellFlow
-> LobbyFlow
-> LobbySessionState
-> MultiplayerLobbyPresenter
-> multiplayer_lobby.tscn
```

Current lobby button flow:

```text
multiplayer_lobby.tscn button press
-> multiplayer_lobby.gd signal
-> MultiplayerLobbyPresenter callback
-> LobbyShellFlow
-> LobbyNetworkActions
-> ClientConnectionService
-> outbound lobby packet
```

The lobby flow mounts multiplayer lobby presentation only when the active session mode is multiplayer and the room state is `Lobby`. Single-player snapshots can still update room-session state, but they do not mount the multiplayer lobby UI.

## Code root

```text
client/scripts/lobby/
client/scripts/ui/lobby/
client/scenes/ui/dialogs/
client/scenes/ui/elements/
```

## Responsibilities

The client lobby session and presentation flow owns:

* Reading room snapshot fields into a client-side lobby session read model.
* Storing the current transient lobby state used by presentation.
* Deriving local-owner status from `local_player_id` and `owner_id`.
* Deriving all-ready and can-start presentation state from member readiness.
* Activating the requested session mode after a room snapshot is applied.
* Showing multiplayer lobby UI when the active mode is multiplayer and the room state is `Lobby`.
* Clearing multiplayer lobby UI when the current snapshot should not be presented as a multiplayer lobby.
* Presenting room code, room status, member rows, local-player marker, owner marker, connected state, and readiness state.
* Enabling Start only when the local client is the owner and all members are ready.
* Toggling Ready intent based on the current local-ready state.
* Sending Ready, Start, and Leave requests through the client connection service.
* Clearing lobby presentation and local lobby read model after local leave return.
* Returning to the configured post-leave menu destination when one exists.

## Does not own

The client lobby session and presentation flow does not own:

* Server room authority.
* Room creation or join admission.
* Room code validity.
* Room capacity.
* Owner assignment.
* Member connection authority.
* Readiness acceptance.
* Start-game acceptance.
* Room lifecycle transitions.
* Match state authority.
* Gameplay simulation.
* WebSocket transport.
* Packet schema source of truth.
* Authenticated-account state.
* Join dialog input validation.
* Main menu or pregame route policy outside lobby-specific return handoff.
* Match result presentation.

## Domain roles

### Lobby snapshot reader

`LobbyPacketReader` reads generated packet fields from room snapshot dictionaries.

It currently reads:

```text
room_code
room_state
local_player_id
owner_id
max_players
members
```

The reader uses generated packet constants instead of hardcoded packet field names.

### Lobby session read model

`LobbySessionState` stores the current client-side lobby read model.

Current fields:

```text
room_code
room_state
local_player_id
owner_id
max_players
members
```

It also owns derived presentation helpers:

```text
is_local_owner()
all_members_ready()
can_start_game()
summary()
```

`can_start_game()` is client-side presentation logic only. The server still owns whether a `start_game_request` is accepted.

### Lobby flow

`LobbyFlow` applies a room snapshot into `LobbySessionState` through `LobbyPacketReader`.

It owns the local lobby-state object and exposes:

```text
apply_room_snapshot(packet)
current_state()
clear()
```

It does not send packets or mount UI directly.

### Lobby shell flow

`LobbyShellFlow` coordinates the snapshot-to-presentation flow.

On room snapshot:

```text
1. Apply packet to LobbyFlow.
2. Log the lobby summary.
3. Read current LobbySessionState.
4. Activate the requested session mode.
5. If the active mode should show multiplayer lobby for this room state, show lobby presentation.
6. Otherwise clear existing multiplayer lobby presentation.
```

The show condition is delegated to `ClientSessionContext.should_show_multiplayer_lobby(room_state)`, which currently returns true only for:

```text
active_mode == Multiplayer
room_state == Lobby
```

### Multiplayer lobby presenter

`MultiplayerLobbyPresenter` owns the lifecycle of the `multiplayer_lobby.tscn` instance.

It:

* Instantiates the lobby scene when needed.
* Adds it to the supplied canvas layer.
* Connects lobby UI signals to callbacks supplied by `LobbyShellFlow`.
* Applies the latest lobby state to the scene.
* Updates Start button enabled state.
* Shows the lobby window.
* Frees and clears the lobby window when lobby presentation should be removed.

### Multiplayer lobby UI

`multiplayer_lobby.gd` owns the lobby window presentation surface.

It displays:

```text
room code
room status text
member list
ready button text
start button enabled/disabled state
```

It emits:

```text
ready_requested(ready)
start_game_requested
leave_requested
```

The Ready button sends the opposite of current local-ready state. The Start button is disabled by default and is enabled only when the presenter applies a state where `can_start_game()` is true.

### Lobby status view model

`LobbyStatusViewModel` maps room state and local member context into display text.

Current status behavior:

* `Starting` -> `Starting game...`
* `InGame` -> `Game in progress.`
* `GameOver` -> `Game over.`
* non-`Lobby` unknown state -> raw room state text
* local owner and can start -> `Ready to start.`
* local owner and cannot start -> `Waiting for players to ready.`
* non-owner and ready -> `Waiting for host to start.`
* non-owner and not ready -> `Press READY when ready.`

The display strings come from generated client lobby constants.

### Lobby member view model

`LobbyMemberViewModel` maps packet member dictionaries into row presentation values.

It currently:

* Uses `player_id` as the display name.
* Adds `(You)` to the local member row.
* Reads readiness from `ready`, falling back to legacy alias `is_ready`.
* Reads connected state from `connected`, falling back to legacy alias `is_connected`.
* Identifies the owner row by comparing member `player_id` to `owner_id`.

### Lobby player list view

`LobbyPlayerListView` owns rendering member rows into the player-list container.

On each render, it clears existing child rows, instantiates one row per member, and calls `set_member()` when the row supports it.

### Player row

`player_row.gd` owns one rendered lobby member row.

It displays:

```text
player name
Ready / Not Ready text
owner indicator
ready green indicator
ready red indicator
```

A disconnected member is shown through the not-ready/red state even when the member's readiness value is otherwise true.

### Lobby network actions

`LobbyNetworkActions` adapts lobby UI intent into connection-service calls.

Current actions:

```text
send_ready_requested(ready)
send_start_game_requested()
send_leave_requested()
```

These call:

```text
ClientConnectionService.send_set_ready_request(ready)
ClientConnectionService.send_start_game_request()
ClientConnectionService.send_leave_room_request()
```

### Lobby return flow

`LobbyReturnFlow` owns local cleanup after the client requests lobby leave.

Current return behavior:

```text
1. Clear LobbyFlow state.
2. Clear multiplayer lobby presentation.
3. Call the cleanup callback when configured.
4. Call the configured return destination when configured.
5. Otherwise show the main menu.
```

`RoomSessionController` supplies the cleanup callback that clears session context and shell boot flow. `AppEntry` currently configures the return destination to `MenuFlowController.show_multiplayer_pregame`.

## Protocols and APIs

### Inbound room snapshot

The lobby presentation flow consumes `room_snapshot` packets after networking has decoded and classified them.

Relevant room snapshot fields:

```text
type
room_code
room_state
members
local_player_id
owner_id
max_players
match_result
```

This document covers only the lobby fields. Match-result caching and presentation are covered by room-session and match-end docs.

### Outbound lobby requests

The lobby UI can send these realtime requests:

```text
set_ready_request
start_game_request
leave_room_request
```

The lobby flow does not build packets directly. It routes intent through `LobbyNetworkActions`, `ClientConnectionService`, and the outbound packet sender.

### Packet source of truth

Lobby packets are sourced from:

```text
shared/packets/lobby.toml
```

Generated Godot packet constants and builders live in:

```text
client/scripts/generated/networking/packets/packets.gd
```

### Presentation constants

Lobby status text, dialog status text, and Ready/Unready button labels are sourced from:

```text
shared/constants/client/lobby.toml
```

Generated Godot constants live in:

```text
client/scripts/generated/constants/constants.gd
```

## Data ownership

The lobby flow owns transient client-side presentation state only.

Client-owned transient state:

```text
LobbySessionState.room_code
LobbySessionState.room_state
LobbySessionState.local_player_id
LobbySessionState.owner_id
LobbySessionState.max_players
LobbySessionState.members
MultiplayerLobbyPresenter.multiplayer_lobby
multiplayer_lobby.gd local_ready
```

The client does not persist lobby state.

Server-owned authoritative state:

```text
room membership
room owner
member readiness
member connection state
room capacity
room lifecycle state
start-game acceptance
leave-room acceptance
```

The member array is duplicated into the client read model so presentation can render from a stable snapshot. It should still be treated as server-observed packet data, not durable client authority.

## Code map

### Primary lobby flow implementation

* `client/scripts/lobby/lobby_flow.gd` - Applies room snapshots to the local lobby session read model.
* `client/scripts/lobby/lobby_packet_reader.gd` - Reads generated room snapshot fields from packet dictionaries.
* `client/scripts/lobby/lobby_session_state.gd` - Stores transient lobby state and derived owner/ready/start helpers.
* `client/scripts/lobby/lobby_shell_flow.gd` - Coordinates snapshot application, session-mode activation, lobby presentation mounting, and lobby UI callbacks.
* `client/scripts/lobby/lobby_network_actions.gd` - Sends Ready, Start, and Leave requests through the connection service.
* `client/scripts/lobby/lobby_return_flow.gd` - Clears lobby state and routes after leave.
* `client/scripts/lobby/multiplayer_lobby_presenter.gd` - Instantiates, updates, shows, and clears the multiplayer lobby scene.

### Lobby presentation UI

* `client/scripts/ui/lobby/multiplayer_lobby.gd` - Lobby window script for room status, member list, Ready/Start/Leave signals, and Start enabled state.
* `client/scripts/ui/lobby/lobby_status_view_model.gd` - Maps room/member context into lobby status display text.
* `client/scripts/ui/lobby/lobby_member_view_model.gd` - Maps member dictionaries into display, readiness, connection, and owner values.
* `client/scripts/ui/lobby/lobby_player_list_view.gd` - Renders member rows.
* `client/scripts/ui/lobby/player_row.gd` - Renders a single lobby member row.

### Lobby scenes

* `client/scenes/ui/dialogs/multiplayer_lobby.tscn` - Multiplayer lobby window.
* `client/scenes/ui/elements/player_row.tscn` - Lobby member row scene.

### App-shell/session participants

* `client/scripts/session/room_session_controller.gd` - Constructs lobby collaborators, delegates room snapshots to lobby shell flow, caches room state and match result, handles room errors, and supplies lobby leave cleanup.
* `client/scripts/session/session_network_controller.gd` - Receives classified room packet signals and delegates room snapshots, room-state changes, and room errors.
* `client/scripts/session/client_session_context.gd` - Tracks requested/active session mode and decides whether multiplayer lobby presentation should be shown.
* `client/scripts/shell/app_entry.gd` - Wires room session, menu return destination, connection service, and gameplay providers.

### Networking participants

* `client/scripts/networking/client_connection_service.gd` - Exposes lobby request send methods and emits classified room packet signals.
* `client/scripts/networking/outbound/client_packet_sender.gd` - Sends generated lobby request packets through the raw network client.
* `client/scripts/networking/inbound/server_packet_dispatcher.gd` - Emits room snapshot, room-state-change, and room-error signals.
* `client/scripts/networking/inbound/server_packet_router.gd` - Classifies room packets by generated packet type constants.
* `client/scripts/networking/network_client.gd` - Owns raw WebSocket transport.

### Generated inputs

* `client/scripts/generated/networking/packets/packets.gd` - Generated lobby packet types and field constants.
* `client/scripts/generated/constants/constants.gd` - Generated room states, session modes, lobby status text, dialog status text, and Ready/Unready labels.
* `shared/packets/lobby.toml` - Lobby packet source of truth.
* `shared/constants/client/lobby.toml` - Client lobby presentation constants source of truth.

### Server authority boundary

* `services/game-server/internal/networking/room_snapshot.go` - Builds authoritative room snapshot packets for each client session.
* `services/game-server/internal/rooms/` - Owns authoritative room membership, owner, readiness, lifecycle, and joinability.
* `services/game-server/internal/networking/` - Routes inbound lobby requests to server room authority.

## Tests

Relevant tests include:

* `client/tests/unit/test_lobby_session_state.gd`
* `client/tests/unit/test_lobby_member_view_model.gd`
* `client/tests/unit/test_lobby_status_view_model.gd`
* `client/tests/unit/lobby/test_lobby_shell_flow.gd`
* `client/tests/unit/lobby/test_lobby_return_flow.gd`
* `client/tests/unit/test_room_session_controller.gd`

Current coverage verifies:

* `LobbySessionState.is_local_owner()` uses `local_player_id`.
* Lobby member owner detection uses `player_id`.
* Lobby member display marks the local player with `(You)`.
* Lobby status text uses local player id for owner identity.
* Lobby leave sends a leave request and invokes return-after-leave behavior.
* Lobby return clears lobby state, clears lobby presentation, calls cleanup, and routes to configured destination.
* Lobby return falls back to showing the main menu when no destination is configured.
* Room session controller cleanup clears session context and shell boot flow after lobby leave return.

Adjacent server tests include:

* `services/game-server/internal/networking/room_snapshot_test.go`
* `services/game-server/tests/networking/room_snapshot_test.go`

Those tests verify room snapshot shape and server-owned room snapshot behavior, not client presentation.

## Related docs

* [Lobby Flow](./!INDEX.md)
* [Menu Flow](../menu-flow.md)
* [Room Session State](../app-shell-and-session/room-session-state.md)
* [Session Boot And Network Target](../app-shell-and-session/session-boot-and-network-target.md)
* [Auth Session Flow](../auth-session-flow.md)
* [Networking Flow](../networking-flow/!INDEX.md)
* [Match End Flow](../match-end-flow/!INDEX.md)
* [Lobby Packets](../../protocol/stubs/lobby-packets.md) - Stub: incomplete lobby packet protocol documentation.
* [Current System Limits](../../limits/current-system-limits.md)

## Notes

`client/scripts/lobby/` is the lobby shell, presenter, and network-action flow area, and `UserInterface` is the parent root for app/menu/lobby screens.

This document intentionally does not cover the join dialog flow in detail. Join dialog ownership belongs to `room-entry-and-join-dialog.md`.

`LobbySessionState.can_start_game()` is a client presentation helper. It should not be treated as gameplay or room authority.

The lobby UI currently uses player-facing `PlayerID` values for member display. It does not display account ids, local profile ids, member ids, or websocket session ids.

Room error text presentation is adjacent to lobby flow through `MultiplayerDialogStatusPresenter`, but detailed room-entry and join-dialog behavior belongs in the separate room-entry document.
