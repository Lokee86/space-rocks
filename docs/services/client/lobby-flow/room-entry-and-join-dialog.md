# Room Entry and Join Dialog

Parent index: [Lobby Flow](./!INDEX.md)

## Purpose

This document describes the current client room-entry flow for multiplayer room creation and joining.

It covers how multiplayer pregame intent reaches create-room and join-room boot requests, how the Join Dialog validates and forwards room codes, how the client gates multiplayer boot requests behind websocket auth, and where room snapshots and room errors hand off to lobby presentation.

## Overview

Client room entry starts after the player enters multiplayer pregame. Authenticated multiplayer entry is documented separately; this document starts at the client-side room-entry surface.

The current flow has two room-entry paths:

```text
Multiplayer Pregame
  Create Game
    -> clear menu UI
    -> queue create-room boot request
    -> connect to multiplayer websocket target
    -> wait for websocket auth when required
    -> send create_room_request

Multiplayer Pregame
  Join Game
    -> show Join Dialog
    -> collect room code
    -> reject blank/whitespace-only room code locally
    -> clear menu UI
    -> queue join-room boot request
    -> connect to multiplayer websocket target
    -> wait for websocket auth when required
    -> send join_room_request
```

The client owns presentation, local input collection, route clearing, pending boot request storage, websocket connection initiation, and outbound packet dispatch. The game server owns room-code validation, room creation, room membership, room joinability, room capacity, lobby authority, and room snapshots.

A successful create or join returns a `room_snapshot`. The client applies that snapshot through the lobby session flow and mounts the multiplayer lobby when the current client session mode should show it.

A failed join returns a `room_error`. The client maps known room error codes to friendly status text and shows that status through the multiplayer dialog status presenter.

## Code root

- `client/scripts/ui/lobby/`
- `client/scripts/ui/menu_flow/`
- `client/scripts/lobby/`
- `client/scripts/boot/`
- `client/scripts/session/`
- `client/scripts/main_menu/`
- `client/scripts/networking/`
- `client/scenes/ui/dialogs/`

## Responsibilities

- Expose multiplayer create and join intent from the pregame menu.
- Gate create and join button handling to multiplayer pregame mode.
- Mount the Join Dialog as a client menu route.
- Collect and trim room-code input.
- Reject blank or whitespace-only room codes before sending a join request.
- Keep the Join Dialog open when local validation fails.
- Clear pregame, sign-in, and join-dialog UI before a valid room transition.
- Queue create-room and join-room boot requests before websocket connection completion.
- Select the multiplayer websocket target for create and join requests.
- Hold multiplayer boot requests until websocket auth succeeds.
- Allow pending multiplayer boot requests to continue when websocket token verification is unavailable so the server can return the authoritative admission error.
- Send generated lobby packets through the client outbound packet sender.
- Apply successful room snapshots through the lobby session flow.
- Surface room-entry failures through friendly status text.
- Keep room-entry presentation separate from server room authority.

## Does not own

- Signed-in account state or Discord login.
- Bearer-token validation.
- Websocket auth authority.
- Room-code generation.
- Room-code format authority.
- Room membership authority.
- Room joinability policy.
- Room capacity policy.
- Room owner selection.
- Ready-state policy.
- Start-game authority.
- Server packet schema source-of-truth ownership.
- Raw websocket transport behavior.
- Lobby member-list presentation after a successful room snapshot.
- Gameplay startup once the room enters `InGame`.
- Persistent account, profile, or match-result data.

## Domain roles

### Pregame entry surface

`pregame_menu.gd` emits multiplayer create and join intent from pregame buttons.

In multiplayer mode:

- `EndlessCreateButton` emits `create_game_requested`.
- `CampaignJoinButton` emits `join_game_requested`.

`pregame_menu_flow.gd` checks the current pregame mode before acting on those signals. Single-player mode cannot use these multiplayer room-entry callbacks.

### Menu route owner

`menu_flow_controller.gd` owns the menu route for the Join Dialog.

It uses the `JOIN_DIALOG` route value from `menu_route.gd`, instantiates `join_dialog.tscn`, wires `JoinDialogFlow`, and clears conflicting menu surfaces before showing the dialog.

When the player cancels the Join Dialog, `close_join_dialog()` removes the dialog and returns to the existing pregame menu when that menu is still alive. If no pregame menu is active, it falls back to the main menu.

### Join Dialog presentation

`join_dialog.gd` owns the direct Join Dialog presentation script.

It resolves:

```text
%RoomCodeInput
%StatusLabel
%JoinButton
%CancelButton
```

It emits:

```text
join_requested(room_code)
cancel_requested
```

The room code emitted from the dialog is stripped of leading and trailing whitespace before it leaves the dialog.

The dialog also exposes:

```text
set_status(message)
clear_status()
```

These methods update only the local Join Dialog status label.

### Join Dialog flow

`join_dialog_flow.gd` owns Join Dialog interaction policy.

On cancel, it calls the injected close-dialog callback.

On join:

```text
strip room code
if blank:
    set Join Dialog status to "Must enter an ID to join."
    stop
clear menu UI for room transition
call injected join-room callback with stripped room code
```

The Join Dialog flow does not send packets directly. It only validates local UI input and delegates the room transition to the injected callback.

### App and boot bridge

The valid join callback flows through app/session boot code:

```text
JoinDialogFlow
MenuFlowController.join_room_callable
AppEntry._request_join_room_from_pregame(room_code)
MainMenuSessionController.request_join_room(room_code)
SessionBootController.request_join_room(room_code)
ShellBootFlow.request_join_room(room_code)
```

Create-room follows the same boot bridge without a room code:

```text
PregameMenuFlow._on_create_game_requested()
MenuFlowController.create_room_callable
AppEntry._request_create_room_from_pregame()
MainMenuSessionController.request_create_room()
SessionBootController.request_create_room()
ShellBootFlow.request_create_room()
```

`MainMenuSessionController.request_join_room()` performs a second non-empty room-code guard. This duplicates the Join Dialog guard intentionally at the app/session boundary.

### Pending boot request

`pending_boot_request.gd` stores the pending boot request until it can be sent.

For room entry, it stores:

```text
BOOT_REQUEST_CREATE_ROOM
BOOT_REQUEST_JOIN_ROOM + room_code
```

Both create-room and join-room are classified as multiplayer boot requests.

`ShellBootFlow.send_pending_boot_request()` consumes the pending request and calls the matching connection-service send method:

```text
send_create_room_request()
send_join_room_request(room_code)
```

### Websocket auth gate

`session_network_controller.gd` decides when pending boot requests are sent after websocket connection.

Single-player requests send immediately after connection.

Multiplayer create and join requests wait for websocket authentication unless the connection is already authenticated.

Current multiplayer behavior:

```text
on websocket connected:
    if pending request is multiplayer:
        if websocket auth already authenticated:
            send pending boot request
        else:
            wait for authenticate_result

on authenticate_result authenticated=true:
    send pending boot request

on authenticate_result authenticated=false, error_code=token_verification_unavailable:
    send pending boot request for server-side admission handling

on authenticate_result authenticated=false, other error:
    do not send pending boot request
```

The client does not decide whether the authenticated account may create or join the room. It only decides whether the pending multiplayer request should be sent to the server.

### Outbound packet sending

Room-entry packets flow through:

```text
ClientConnectionService
ClientPacketSender
LobbyClientPackets
generated Packets helper
NetworkClient
```

The Join Dialog and menu flow do not construct packet dictionaries directly.

### Room snapshot handoff

A successful create or join produces a `room_snapshot`.

The client handles the snapshot through:

```text
ClientConnectionService.room_snapshot_received
SessionNetworkController._on_room_snapshot_received()
RoomSessionController.handle_room_snapshot()
LobbyShellFlow.apply_room_snapshot()
LobbyFlow.apply_room_snapshot()
LobbySessionState.apply_snapshot()
MultiplayerLobbyPresenter.show_lobby()
```

The room-entry flow ends once the room snapshot has handed off to lobby session and lobby presentation.

### Room error handoff

A failed create or join produces a `room_error`.

The client handles the error through:

```text
ClientConnectionService.room_error_received
SessionNetworkController._on_room_error_received()
RoomSessionController.handle_room_error()
MultiplayerDialogStatusPresenter.show_room_error()
```

Known room-entry error codes are mapped to friendly status text:

```text
invalid_room_code -> Invalid room ID.
room_not_found -> Room not found.
room_full -> Room is full.
room_in_game -> Room is already in game.
already_in_room -> Already in a room.
invalid_room_state -> Room is not joinable.
```

If the server supplies an unmapped error with a message, the client shows that message. Otherwise it falls back to `Could not join room.`

## Protocols and APIs

### Local UI signal flow

The Join Dialog uses local Godot signals, not a network request.

```text
JoinButton.pressed
JoinDialog._on_join_pressed()
join_requested(stripped_room_code)
JoinDialogFlow._on_join_requested(room_code)
```

Cancel uses:

```text
CancelButton.pressed
JoinDialog._on_cancel_pressed()
cancel_requested
JoinDialogFlow._on_cancel_requested()
MenuFlowController.close_join_dialog()
```

### Outbound create-room packet

Create-room sends a generated lobby packet:

```json
{
  "type": "create_room_request"
}
```

The client calls this through:

```text
LobbyClientPackets.create_room_request_packet()
```

### Outbound join-room packet

Join-room sends a generated lobby packet:

```json
{
  "type": "join_room_request",
  "room_code": "ABC123"
}
```

The client calls this through:

```text
LobbyClientPackets.join_room_request_packet(room_code)
```

The client only strips whitespace and rejects empty room codes. The server owns room-code normalization and full validation.

### Inbound room snapshot

A successful room-entry response is a `room_snapshot`.

The client reads the snapshot fields needed for lobby state:

```text
room_code
room_state
local_player_id
owner_id
max_players
members
match_result
```

`match_result` may be present after match-end flows, but it is not owned by room entry.

### Inbound room error

A failed room-entry response is a `room_error`.

The client reads:

```text
error_code
message
```

The server owns the authoritative error code. The client only maps it to presentation text.

### Packet source of truth

The realtime lobby packet source of truth is:

```text
shared/packets/lobby.toml
```

Generated client packet helpers live at:

```text
client/scripts/generated/networking/packets/packets.gd
```

Room-entry code should use generated helpers instead of hand-building packet dictionaries in UI or menu scripts.

## Data ownership

The client owns only transient room-entry state.

Client-owned transient state includes:

```text
MenuFlowController.current_route
MenuFlowController.join_dialog
PendingBootRequest.request_type
PendingBootRequest.join_room_code
ClientSessionContext requested multiplayer mode
ClientConnectionService websocket auth result cache
```

The client does not persist room-entry data.

The client does not own durable room data. After a successful server snapshot, `LobbySessionState` stores a client read model of server room state for lobby presentation, but server rooms remain authoritative.

The client does not own account identity or multiplayer admission data. Authenticated account state is provided by auth/session flows and verified by the server.

## Code map

### Pregame and menu entry

- `client/scenes/ui/pregame_menu.tscn`
- `client/scripts/ui/menus/pregame_menu.gd`
- `client/scripts/ui/menu_flow/pregame_menu_flow.gd`
- `client/scripts/ui/menu_flow/menu_flow_controller.gd`
- `client/scripts/ui/menu_flow/menu_route.gd`
- `client/scripts/shell/app_entry.gd`
- `client/scripts/main_menu/main_menu_session_controller.gd`

### Join Dialog

- `client/scenes/ui/dialogs/join_dialog.tscn`
- `client/scripts/ui/lobby/join_dialog.gd`
- `client/scripts/ui/lobby/join_dialog_flow.gd`

### Session boot and websocket auth gate

- `client/scripts/boot/session_boot_controller.gd`
- `client/scripts/boot/shell_boot_flow.gd`
- `client/scripts/boot/pending_boot_request.gd`
- `client/scripts/boot/session_network_target.gd`
- `client/scripts/session/client_session_context.gd`
- `client/scripts/session/session_network_controller.gd`
- `client/scripts/networking/client_connection_service.gd`

### Outbound packet sending

- `client/scripts/networking/outbound/client_packet_sender.gd`
- `client/scripts/networking/outbound/lobby_client_packets.gd`
- `client/scripts/networking/network_client.gd`
- `client/scripts/networking/packets/packet_codec.gd`
- `client/scripts/generated/networking/packets/packets.gd`

### Lobby handoff

- `client/scripts/session/room_session_controller.gd`
- `client/scripts/lobby/lobby_shell_flow.gd`
- `client/scripts/lobby/lobby_flow.gd`
- `client/scripts/lobby/lobby_packet_reader.gd`
- `client/scripts/lobby/lobby_session_state.gd`
- `client/scripts/lobby/multiplayer_dialog_status_presenter.gd`
- `client/scripts/lobby/multiplayer_lobby_presenter.gd`

### Generated constants and packet sources

- `client/scripts/generated/constants/constants.gd`
- `shared/constants/client/lobby.toml`
- `shared/constants/client/shell.toml`
- `shared/packets/lobby.toml`

### Non-owning server authority paths

These paths are listed to show the client/server boundary. The client documentation here does not own server behavior.

- `services/game-server/internal/networking/room_handlers.go`
- `services/game-server/internal/networking/inbound/lobby.go`
- `services/game-server/internal/networking/inbound_adapter.go`
- `services/game-server/internal/rooms/manager.go`
- `services/game-server/internal/rooms/room_join.go`
- `services/game-server/internal/rooms/roomrules/join.go`
- `services/game-server/internal/rooms/constants.go`

## Tests

### Join Dialog UI

- `client/tests/unit/ui/lobby/test_join_dialog.gd`

Verifies:

- join button emits a trimmed room code
- blank input emits an empty room code
- cancel emits `cancel_requested`
- status text can be set and cleared

### Join Dialog flow

- `client/tests/unit/ui/lobby/test_join_dialog_flow.gd`

Verifies:

- cancel calls the close callback
- empty and whitespace-only room codes set local status and do not join
- valid room codes clear UI and call the join callback with the stripped code

### Menu route and room-entry UI clearing

- `client/tests/unit/ui/menu_flow/test_menu_flow_controller.gd`

Verifies:

- Join Dialog route creation
- sign-in and Join Dialog route clearing
- cancel return to multiplayer pregame
- valid join code calls the join callback and clears menu UI
- empty join code shows status and keeps the Join Dialog open
- room-transition clearing removes pregame, sign-in, and Join Dialog surfaces

### Pending boot request and session boot

- `client/tests/unit/test_pending_boot_request.gd`
- `client/tests/unit/test_shell_boot_flow.gd`
- `client/tests/unit/test_session_network_controller.gd`
- `client/tests/unit/boot/test_session_network_target.gd`

Verifies:

- create-room and join-room are multiplayer pending boot requests
- join-room stores the requested room code
- pending room-entry requests are consumed when sent
- single-player requests do not wait for websocket auth
- multiplayer requests wait for websocket auth success
- token-verification-unavailable still sends pending multiplayer requests
- invalid-token auth failure does not send pending multiplayer requests
- multiplayer uses the configured multiplayer websocket target

### Lobby handoff

- `client/tests/unit/lobby/test_lobby_shell_flow.gd`
- `client/tests/unit/lobby/test_lobby_return_flow.gd`
- `client/tests/unit/test_lobby_session_state.gd`
- `client/tests/unit/test_lobby_member_view_model.gd`
- `client/tests/unit/test_lobby_status_view_model.gd`

Verifies room snapshot application, lobby presentation state, lobby return behavior, and member/status view models after room entry succeeds.

### Non-owning server boundary tests

These tests verify server-side authority that the client consumes but does not own:

- `services/game-server/internal/rooms/manager_test.go`
- `services/game-server/internal/rooms/room_join_test.go`

## Related docs

- [Lobby Flow](./!INDEX.md)
- [Client](../!INDEX.md)
- [Menu Flow](../menu-flow.md)
- [Networking Flow](../networking-flow/!INDEX.md)
- [Auth Session Flow](../auth-session-flow.md)
- [App Shell And Session](../app-shell-and-session/!INDEX.md)
- [Session Boot And Network Target](../app-shell-and-session/session-boot-and-network-target.md)
- [Room Session State](../app-shell-and-session/room-session-state.md)
- [Lobby Session And Presentation](lobby-session-and-presentation.md)
- [Game Server](../../game-server/!INDEX.md)

## Notes

Join Game opens `join_dialog.tscn`, multiplayer create/join route through multiplayer flow, and lobby behavior begins after create or join succeeds.

The Join Dialog is a menu route, not the multiplayer lobby. It is cleared before the network join request is sent. Server-side room errors therefore route through the multiplayer dialog status presenter rather than being shown inside the already-cleared Join Dialog.

The client currently performs only blank-room-code validation before sending a join request. Full room-code validation remains server-owned.

`Join Dialog` and `room entry` are documented together because the Join Dialog is the only current room-entry UI that owns local input validation. Create-room entry has no separate dialog and shares the same boot, websocket-auth, outbound-packet, and lobby handoff path.
