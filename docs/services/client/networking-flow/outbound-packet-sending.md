# Outbound Packet Sending

Parent index: [Networking Flow](./!INDEX.md)

## Purpose

This document describes how the Godot client sends outbound realtime packets to the game server.

It covers the client service implementation path from local intent to WebSocket text send. It does not define packet schema authority, server validation, or authoritative gameplay results.

## Overview

Outbound packet sending is the client-side bridge between local client intent and the realtime server.

The common path is:

```text
client caller
-> ClientConnectionService send method
-> ClientPacketSender wrapper
-> packet family builder or generated Packets helper
-> NetworkClient.send_raw_packet(packet)
-> PacketCodec.encode(packet)
-> WebSocketPeer.send_text(wire_message)
-> realtime server
```

`ClientConnectionService` is the service-facing outbound facade used by session boot, lobby, gameplay, devtools, telemetry, and general client flows. `ClientPacketSender` owns focused outbound wrapper methods and delegates raw packet sending to `NetworkClient`. `NetworkClient` owns the final connection-state guard, JSON encoding, and WebSocket text send.

Some flows build generated packet dictionaries before reaching `ClientPacketSender`. Those flows call:

```text
ClientConnectionService.send_packet(packet)
```

instead of a more specific wrapper method.

The outbound send path is best-effort and non-queued. If the packet sender has no `NetworkClient`, or if the raw WebSocket is not open, the packet is not sent. If packet encoding fails, the client logs a warning and drops the packet.

The server remains authoritative for packet acceptance, room state, gameplay simulation, devtools effects, and durable results.

## Code root

```text
client/scripts/networking/
client/scripts/networking/outbound/
```

## Responsibilities

The client outbound packet sending flow owns:

* Exposing outbound send methods through `ClientConnectionService`.
* Keeping caller code away from raw `WebSocketPeer` usage.
* Creating generated packet dictionaries through focused packet-family wrappers.
* Passing already-built packet dictionaries from callers that own context-specific packet construction.
* Sending gameplay input, respawn, pause, target, lobby, room-entry, devtools, telemetry, viewport config, and auth packets through a common raw send path.
* Guarding sends when the packet sender or raw network client is unavailable.
* Letting `NetworkClient` guard sends when the WebSocket is not open.
* Encoding outbound packet dictionaries as JSON text.
* Sending encoded packet text through the active WebSocket.
* Keeping packet schema shape behind generated packet helpers where available.

## Does not own

The client outbound packet sending flow does not own:

* Packet schema source-of-truth files.
* Packet generation.
* Server-side packet validation.
* Server-side room admission.
* Server-side gameplay authority.
* Server-side devtools command effects.
* Inbound packet routing.
* WebSocket URL selection.
* Session mode selection.
* WebSocket auth policy.
* Auth token storage.
* Lobby presentation.
* Gameplay input semantics.
* Target candidate selection.
* Retry, acknowledgement, resend, or durable outbound queues.
* Persistent player, account, match, or room data.

## Domain roles

### Connection-service outbound facade

`ClientConnectionService` is the main client service facade for outbound packets.

It creates and owns:

```text
NetworkClient
ClientPacketSender
ServerPacketDispatcher
```

The outbound methods on `ClientConnectionService` are used by other client service flows instead of calling `NetworkClient` directly.

Current exposed send surfaces include:

```text
send_start_single_player_request(local_profile_id)
send_create_room_request()
send_join_room_request(room_code)
send_set_ready_request(is_ready)
send_start_game_request()
send_input_packet(packet)
send_packet(packet)
send_respawn_request()
send_pause_request()
send_telemetry_ping(sequence, client_sent_msec)
send_debug_kill_player_request(target_scope, target_player_id)
send_debug_kill_target_player_request(target_player_id, target_scope)
send_leave_room_request()
send_return_to_lobby_request()
```

The facade is intentionally broader than one packet family because session, lobby, gameplay, and devtools callers all need one stable service-facing outbound seam.

### Packet sender wrapper

`ClientPacketSender` owns the focused outbound wrapper layer.

It stores:

```text
network_client: NetworkClient
```

and exposes packet-family methods that create packet dictionaries and forward them to:

```text
network_client.send_raw_packet(packet)
```

It also exposes generic forwarding:

```text
send_packet(packet)
send_input_packet(packet)
```

`send_packet()` only checks that `network_client` exists. The final connected-state check belongs to `NetworkClient.send_raw_packet()`.

### Packet-family builders

Outbound packet construction is split by family:

```text
GameplayClientPackets
= gameplay input, respawn, pause, and target request packet helpers

LobbyClientPackets
= room-entry, lobby, ready, start, single-player start, leave, and return-to-lobby packet helpers

DevtoolsClientPackets
= debug command packet helpers and devtools-only target fields

TelemetryClientPackets
= telemetry ping packet helper
```

Gameplay, lobby, and devtools helpers mostly wrap generated packet builders from:

```text
client/scripts/generated/networking/packets/packets.gd
```

Telemetry currently builds its ping packet directly from generated field and type constants.

### Raw transport sender

`NetworkClient.send_raw_packet(packet)` owns the final outbound transport step.

Current behavior:

```text
if WebSocket is not open:
    return

encode packet through PacketCodec.encode(packet)

if encode fails:
    log network warning
    return

socket.send_text(encoded wire message)
```

The raw sender does not queue packets for later delivery.

### Caller-owned packet construction

Some callers construct generated packet dictionaries themselves because the packet belongs to a narrower local flow.

Examples:

```text
TargetRequestFlow
-> Packets.select_target_at_position_request_packet(...)
-> ClientConnectionService.send_packet(packet)

TargetRequestFlow
-> Packets.clear_target_request_packet()
-> ClientConnectionService.send_packet(packet)

ClientViewportConfigFlow
-> Packets.client_config_packet(width, height)
-> ClientConnectionService.send_packet(packet)
```

This is still part of the outbound packet sending path. The difference is only where the dictionary is created.

### Auth packet special case

WebSocket auth send is a special case.

When the socket opens, `ClientConnectionService._on_connected()` calls:

```text
_send_authenticate_request_if_token_exists()
```

That reads the auth session token and calls:

```text
NetworkClient.send_authenticate_request(token)
```

`NetworkClient.send_authenticate_request()` builds the generated auth packet and sends through `send_raw_packet()`.

This bypasses `ClientPacketSender`, but it still uses the same raw send path.

### Timing and gating

Outbound timing is owned by the calling flow before packets reach the raw sender.

Examples:

```text
GameplayInputFlow
= sends input only after gameplay state has been received, a player exists, a connection service exists, and gameplay is not paused

ShellBootFlow / SessionNetworkController
= sends pending boot requests after connection and, for multiplayer, after websocket auth succeeds or token verification is unavailable

ClientViewportConfigFlow
= sends viewport config only when a connection service exists and is connected

NetworkClient
= performs the final open-WebSocket guard
```

The outbound packet layer should not duplicate all caller-specific eligibility rules. It provides the common send path and final connection guard.

### Authority and confirmation

Outbound packets are requests or observations from the client.

The client can request:

```text
movement/fire input
respawn
pause
target selection
target clear
room creation
room join
ready state change
game start
leave room
return to lobby
debug command
telemetry ping
viewport config
authentication
```

The server owns whether those packets are accepted and what state changes result. The client observes confirmation through inbound state, room snapshots, room errors, auth results, debug status, telemetry pong, and gameplay packets.

## Protocols and APIs

The outbound packet surface is the client-side realtime WebSocket send API.

It is used by client service flows and consumed by the realtime game server. Data crossing the boundary is a packet dictionary encoded as JSON text. Each packet has a `type` field and may include additional fields or nested dictionaries depending on the packet family.

The client outbound path does not own the schema or server interpretation of those packets. It only builds known packet dictionaries, encodes them, and sends them when the WebSocket is open.

### Generic raw send API

The generic outbound API is:

```text
ClientConnectionService.send_packet(packet)
-> ClientPacketSender.send_packet(packet)
-> NetworkClient.send_raw_packet(packet)
```

Input has a named forwarding method:

```text
ClientConnectionService.send_input_packet(packet)
-> ClientPacketSender.send_input_packet(packet)
-> ClientPacketSender.send_packet(packet)
```

### Gameplay packets

Gameplay outbound packets include:

```text
input
respawn
pause_request
set_target_player_request
select_target_at_position_request
clear_target_request
client_config
```

`client_config` is generated with gameplay packet helpers but is sent by the viewport config flow through the generic packet sender.

Normal gameplay target identity uses:

```text
target_kind
target_id
```

`target_player_id` is devtools/player-only compatibility data and should not become the normal gameplay targeting model.

### Lobby and room-entry packets

Lobby and room-entry outbound packets include:

```text
create_room_request
join_room_request
leave_room_request
set_ready_request
start_game_request
start_single_player_request
return_to_lobby_request
```

Room-entry and lobby callers do not hand-build these packet dictionaries. They route through connection-service methods and `LobbyClientPackets`.

### Devtools packets

Devtools outbound packets include debug command requests such as:

```text
debug_kill_player
toggle_debug_invincible
toggle_debug_infinite_lives
toggle_debug_freeze_world
toggle_debug_freeze_player
debug_set_score
debug_add_score
debug_set_lives
debug_add_lives
debug_clear_bullets
debug_clear_asteroids
```

Devtools packets may include player-only compatibility fields such as:

```text
target_player_id
target_scope
```

Those fields are devtools-specific and do not replace canonical gameplay target identity.

### Telemetry packets

Telemetry outbound currently includes:

```text
telemetry_ping
```

The current telemetry ping packet carries:

```text
sequence
client_sent_msec
```

The server responds through inbound telemetry pong routing.

### Auth packets

WebSocket auth uses:

```text
authenticate_request
```

Auth send is triggered after the raw WebSocket connection opens when a token exists in the auth session controller.

The auth request uses the same raw send path, but it is sent directly through `NetworkClient.send_authenticate_request()` rather than through `ClientPacketSender`.

## Data ownership

The client outbound packet layer owns transient packet dictionaries only.

The packet schema source of truth is:

```text
shared/packets/*.toml
```

Current relevant packet source files include:

```text
shared/packets/gameplay.toml
shared/packets/lobby.toml
shared/packets/debug.toml
shared/packets/outputs.toml
```

Generated client packet output is:

```text
client/scripts/generated/networking/packets/packets.gd
```

Generated server packet outputs include:

```text
services/game-server/internal/game/packets.go
services/game-server/internal/game/runtime/packets_generated.go
services/game-server/internal/devtools/packets_generated.go
```

Packet pipeline drift is checked through data-sync packet generation and validation.

Current packet validation command pattern:

```text
python3 tools/data_sync/main.py -check -packets -go -gds
```

Client packet senders should prefer generated helpers or focused packet-family wrappers. UI, menu, and gameplay scripts should not grow scattered hand-built packet dictionaries when a generated packet helper exists.

## Outbound send paths

### Gameplay input

Current gameplay input send path:

```text
GameplayInputFlow.process()
-> player.get_input_packet()
-> ClientConnectionService.send_input_packet(packet)
-> ClientPacketSender.send_input_packet(packet)
-> ClientPacketSender.send_packet(packet)
-> NetworkClient.send_raw_packet(packet)
```

`GameplayInputFlow` suppresses input sends when gameplay state has not been received, the local player is missing, the connection service is missing, or gameplay is paused.

### Target selection

Current target selection send path:

```text
MouseActionFlow
-> GameplayTargetingContext
-> TargetRequestFlow.select_target()
-> Packets.select_target_at_position_request_packet(...)
-> ClientConnectionService.send_packet(packet)
-> ClientPacketSender.send_packet(packet)
-> NetworkClient.send_raw_packet(packet)
```

Current target clear path:

```text
TargetRequestFlow.deselect_target()
-> Packets.clear_target_request_packet()
-> ClientConnectionService.send_packet(packet)
-> ClientPacketSender.send_packet(packet)
-> NetworkClient.send_raw_packet(packet)
```

The client sends target intent. Authoritative target state is confirmed later through server-driven gameplay state.

### Session boot requests

Current boot request send path:

```text
ShellBootFlow.send_pending_boot_request()
-> ClientConnectionService.send_start_single_player_request(local_profile_id)
   or ClientConnectionService.send_create_room_request()
   or ClientConnectionService.send_join_room_request(room_code)
-> ClientPacketSender
-> LobbyClientPackets
-> generated Packets helper
-> NetworkClient.send_raw_packet(packet)
```

`SessionNetworkController` owns the connection/auth timing for pending boot requests.

### Lobby UI requests

Current lobby UI request path:

```text
multiplayer_lobby.gd signal
-> MultiplayerLobbyPresenter callback
-> LobbyShellFlow
-> LobbyNetworkActions
-> ClientConnectionService.send_set_ready_request(...)
   or ClientConnectionService.send_start_game_request()
   or ClientConnectionService.send_leave_room_request()
-> ClientPacketSender
-> LobbyClientPackets
-> generated Packets helper
-> NetworkClient.send_raw_packet(packet)
```

The client presents readiness and start eligibility, but server room authority owns acceptance.

### Viewport config

Current viewport config send path:

```text
ClientViewportConfigFlow.send_client_config()
-> Packets.client_config_packet(width, height)
-> ClientConnectionService.send_packet(packet)
-> ClientPacketSender.send_packet(packet)
-> NetworkClient.send_raw_packet(packet)
```

`ClientViewportConfigFlow` checks `connection_service.is_server_connected()` before sending. The raw sender still performs its own final WebSocket-open guard.

### Devtools commands

Current devtools command path is generally:

```text
devtools hotkey or window action
-> devtools command context / connection-service send method
-> ClientPacketSender devtools method
-> DevtoolsClientPackets
-> generated Packets helper
-> NetworkClient.send_raw_packet(packet)
```

Devtools requests mutate gameplay only when the server accepts and applies the debug command through server-owned devtools and gameplay seams.

### Telemetry ping

Current telemetry ping path:

```text
telemetry caller
-> ClientConnectionService.send_telemetry_ping(sequence, client_sent_msec)
-> ClientPacketSender.send_telemetry_ping(sequence, client_sent_msec)
-> TelemetryClientPackets.telemetry_ping_packet(...)
-> NetworkClient.send_raw_packet(packet)
```

The outbound telemetry packet is an observation request. The corresponding server response is routed through inbound telemetry pong handling.

### WebSocket auth request

Current WebSocket auth path:

```text
NetworkClient.connected_to_server
-> ClientConnectionService._on_connected()
-> ClientConnectionService._send_authenticate_request_if_token_exists()
-> NetworkClient.send_authenticate_request(token)
-> Packets.authenticate_request_packet(token)
-> NetworkClient.send_raw_packet(packet)
```

This path bypasses `ClientPacketSender` but still uses generated packets and the common raw send method.

## Connection and send constraints

The current outbound send path has these constraints:

```text
ClientConnectionService method with no ClientPacketSender
-> no-op

ClientPacketSender method with no NetworkClient
-> no-op

NetworkClient.send_raw_packet() while WebSocket is not open
-> no-op

PacketCodec.encode(packet) failure
-> network warning and no send
```

Outbound packets are not queued.

A packet attempted before connection, during connection, after close, or after graceful-close start will not be resent automatically. Calling flows must send again when their lifecycle requires it.

## Code map

### Primary outbound networking implementation

```text
client/scripts/networking/client_connection_service.gd
client/scripts/networking/network_client.gd
client/scripts/networking/outbound/client_packet_sender.gd
client/scripts/networking/outbound/gameplay_client_packets.gd
client/scripts/networking/outbound/lobby_client_packets.gd
client/scripts/networking/outbound/devtools_client_packets.gd
client/scripts/networking/outbound/telemetry_client_packets.gd
client/scripts/networking/packets/packet_codec.gd
```

### Generated packet implementation

```text
client/scripts/generated/networking/packets/packets.gd
```

### Gameplay callers

```text
client/scripts/gameplay/input/gameplay_input_flow.gd
client/scripts/gameplay/input/gameplay_input_context.gd
client/scripts/gameplay/input/mouse_action_flow.gd
client/scripts/gameplay/targeting/gameplay_targeting_context.gd
client/scripts/gameplay/targeting/target_request_flow.gd
```

### Session, boot, lobby, and config callers

```text
client/scripts/boot/shell_boot_flow.gd
client/scripts/session/session_network_controller.gd
client/scripts/session/room_session_controller.gd
client/scripts/config/client_viewport_config_flow.gd
client/scripts/lobby/lobby_network_actions.gd
client/scripts/lobby/lobby_shell_flow.gd
```

### Devtools callers

```text
client/scripts/devtools/
client/scripts/devtools/context/
```

### Packet source files

```text
shared/packets/gameplay.toml
shared/packets/lobby.toml
shared/packets/debug.toml
shared/packets/outputs.toml
```

### Non-ownership boundaries

```text
services/game-server/internal/networking/
services/game-server/internal/rooms/
services/game-server/internal/game/
services/game-server/internal/devtools/
```

These server paths consume and validate outbound packets, but they are not owned by the client outbound packet sending documentation.

## Tests

Relevant current tests include:

```text
client/tests/unit/test_packet_codec.gd
client/tests/unit/test_shell_boot_flow.gd
client/tests/unit/test_session_network_controller.gd
client/tests/unit/test_pending_boot_request.gd
client/tests/unit/boot/test_session_network_target.gd
client/tests/unit/test_gameplay_input_context.gd
client/tests/unit/test_target_request_flow.gd
client/tests/unit/lobby/test_lobby_shell_flow.gd
client/tests/unit/lobby/test_lobby_return_flow.gd
client/tests/unit/test_room_session_controller.gd
client/tests/unit/ui/menu_flow/test_app_entry_menu_flow.gd
client/tests/unit/ui/lobby/test_join_dialog_flow.gd
```

`test_packet_codec.gd` verifies JSON encoding and packet envelope decode validation.

Boot and session-network tests verify pending boot request dispatch and websocket auth gating around outbound room-entry packets.

Gameplay input and target request tests verify caller-side send behavior before packets reach the raw transport layer.

Lobby tests verify lobby UI intent and return flows that call outbound lobby request methods.

No focused `ClientPacketSender` unit test was found during this pass.

## Related docs

* [Networking Flow](./!INDEX.md)
* [Client](../!INDEX.md)
* [Input And Targeting](../input-and-targeting.md)
* [Gameplay Runtime](../gameplay-runtime/!INDEX.md)
* [Client Viewport Config Flow](../app-shell-and-session/client-viewport-config-flow.md)
* [Session Boot And Network Target](../app-shell-and-session/session-boot-and-network-target.md)
* [Lobby Flow](../lobby-flow/!INDEX.md)
* [Room Entry And Join Dialog](../lobby-flow/room-entry-and-join-dialog.md)
* [Lobby Session And Presentation](../lobby-flow/lobby-session-and-presentation.md)
* [Game Server](../../game-server/!INDEX.md)
* [Realtime WebSocket Protocol](../../../protocol/realtime-websocket-protocol.md) - incomplete realtime protocol documentation.
* [Gameplay Packets](../../../protocol/gameplay-packets.md) - incomplete gameplay packet protocol documentation.
* [Lobby Packets](../../../protocol/lobby-packets.md) - incomplete lobby packet protocol documentation.
* [Devtools Packets]() - incomplete devtools packet protocol documentation.
* [Packet Schema Pipeline](../../../data/packet-schemas.md) - incomplete packet schema pipeline documentation.

## Notes

WebSocket packet schemas are sourced from `shared/packets/*.toml`, generated client packet helpers live under `client/scripts/generated/networking/packets/`, and client input and devtools send intent while server systems own authority.

`ClientPacketSender` is not the only path that builds outbound packet dictionaries. Target selection, viewport config, and auth use generated helpers closer to their owning flows, then converge at the same raw send path.
