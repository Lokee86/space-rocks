# Session Boot And Network Target

Parent index: [App Shell And Session](./!INDEX.md)

## Purpose

This document describes the current client session boot flow, websocket network-target selection, and semantic `realtime_transport_ready` readiness gate.

It covers how the Godot client turns menu requests into a requested session mode, pending boot request, websocket target URL, connection attempt, inbound signaling ownership, readiness gate, and eventual boot packet dispatch.

## Overview

Session boot is the client-side bridge between menu intent and server room or gameplay entry.

The client does not enter single-player or multiplayer by directly changing scenes into gameplay. It records the requested mode, stores a pending boot request, selects the websocket target for that mode, starts or reuses the websocket connection, and sends the appropriate boot packet only after the readiness gate allows it. Session boot controls boot dispatch timing; gameplay presentation activation remains owned by gameplay runtime flows. `ClientConnectionService` owns the connection-facing handshake plumbing, while `ClientInboundCoordinator` owns inbound WebRTC signaling packet handling and emits the semantic `realtime_transport_ready` signal when the transport is ready.

Current boot request types are:

```text
single_player
create_room
join_room
```

Current session modes are:

```text
none
single_player
multiplayer
```

Network target selection is deliberately small. `SessionNetworkTarget.websocket_url_for_mode()` maps `single_player` to `SINGLE_PLAYER_WS_URL`, maps `multiplayer` to `MULTIPLAYER_WS_URL`, and returns an empty string for unknown modes.

Both current websocket URLs point at the same local Go server route:

```text
ws://localhost:8080/ws
```

The route path does not define play mode. Play mode is expressed through the client request and server-side session, room, and admission policy. Single-player and multiplayer can later point at different infrastructure without changing the server route model.

WebSocket target selection is separate from WebRTC connectivity. The WebSocket URL may point at a normal hosted or proxied service route, but WebRTC DataChannel connectivity is established by ICE candidates rather than by a WebRTC URL. Deployment therefore must allow the advertised WebRTC ICE address and UDP path to reach the game server directly. Cloudflare-proxied HTTP routes should not be assumed to carry UDP WebRTC traffic.

## Code root

* `client/scripts/boot/`
* `client/scripts/session/`
* `client/scripts/shell/`

## Responsibilities

* Accept single-player, create-room, and join-room requests from menu/session callers.
* Set the requested client session mode before starting boot.
* Store one pending boot request until it is sent or cleared.
* Preserve the selected Local Profile id for single-player boot requests.
* Preserve the room code for join-room boot requests.
* Select the websocket URL from generated client constants based on requested session mode.
* Start a WebSocket connection through `ClientConnectionService`. `ClientInboundCoordinator` handles inbound WebRTC signaling packets and forwards them to `RealtimeTransportSession`.
* Send a pending single-player boot request after `realtime_transport_ready` when active gameplay will use WebRTC.
* Hold pending multiplayer boot requests until websocket auth succeeds or token verification is unavailable.
* Send client viewport config after a boot request is sent.
* Clear pending boot request state during session reset paths.

## Does not own

* Raw websocket transport.
* Websocket packet encoding or decoding.
* Server websocket route ownership.
* Realtime packet schema ownership.
* Auth token storage.
* Rails auth/session identity.
* Multiplayer admission authority.
* Room membership authority.
* Gameplay session lifecycle.
* Gameplay session activation.
* Scene/menu presentation.
* Local Profile persistence.
* Account persistence.
* Server-side mode enforcement.

## Domain roles

### Session boot controller

`SessionBootController` is the app-shell facade for boot requests.

It creates:

```text
ClientSessionContext
ClientConnectionService
ShellBootFlow
```

It exposes request methods for:

```text
request_single_player(local_profile_id)
request_create_room()
request_join_room(room_code)
```

Each request method records the requested session mode, prepares the matching pending boot request, selects the websocket URL for that mode, and starts a connection attempt.

### Session context

`ClientSessionContext` tracks requested and active client mode.

It owns:

```text
requested_mode
active_mode
```

Current boot requests set `requested_mode`. Other session flows activate or clear this state as the client moves through room and gameplay lifecycle.

### Pending boot request

`PendingBootRequest` is the small state holder for the boot request waiting to be sent.

It stores:

```text
request_type
join_room_code
local_profile_id
```

It does not connect to the server and does not send packets. It only records, exposes, consumes, and clears the pending request.

### Shell boot flow

`ShellBootFlow` owns pending boot request dispatch.

`RealtimeTransportSession` owns WebRTCTransport start, poll, close, and replacement. `ClientConnectionService` remains the boot-facing coordinator, and `ClientInboundCoordinator` owns inbound WebRTC signaling routing.

It stores a `PendingBootRequest`, starts the websocket connection through the connection service, and sends the pending request once the connection/auth/readiness gate allows it.

It sends one of:

```text
send_start_single_player_request(local_profile_id)
send_create_room_request()
send_join_room_request(room_code)
```

When a boot request is sent, it emits:

```text
boot_request_sent
```

`AppEntry` connects that signal to client viewport config sending.

### Network target selection

`SessionNetworkTarget` is the narrow mapping seam for websocket target selection.

Current mapping:

```text
single_player -> SINGLE_PLAYER_WS_URL
multiplayer   -> MULTIPLAYER_WS_URL
unknown       -> ""
```

The generated constants come from shared client shell constants. The server-advertised ICE address and UDP-path reachability belong to server WebRTC config, not `SessionNetworkTarget`.

### Future environment configuration contract

The client will eventually need one centralized environment-configuration contract for these independently deployable targets:

```text
Rails API base URL
player-data API base URL
single-player WebSocket URL
multiplayer WebSocket URL
```

`SessionNetworkTarget` remains the mode-to-WebSocket mapping seam. It must consume the resolved single-player and multiplayer WebSocket targets; it must not grow scattered environment lookup or service-discovery behavior.

The centralized resolver must apply one documented precedence order:

```text
packaged/build configuration
-> command-line or environment override
-> local development defaults
```

The exact Godot configuration mechanism remains an implementation decision, but callers must receive a resolved contract rather than independently reading process environment state. A missing or malformed hosted target must be visible validation failure at configuration/boot time, with the target name and reason exposed to the user or diagnostic log. It must not silently fall back to a localhost URL or proceed with an empty target. Local defaults are valid only for explicitly local development and local packaged single-player shapes.

### Connection, auth, and readiness gate

`SessionNetworkController` participates in boot only after the websocket connection emits connection/auth/readiness signals.

Current behavior:

```text
websocket connected + `realtime_transport_ready` + pending single-player request
-> send pending request immediately

connected + pending multiplayer request + websocket auth already authenticated
-> send pending request

connected + pending multiplayer request + websocket auth not authenticated
-> wait for authenticate_result

authenticate_result authenticated=true
-> send pending multiplayer request

authenticate_result authenticated=false, error_code=token_verification_unavailable
-> send pending multiplayer request so server-side admission can fail explicitly

authenticate_result authenticated=false, other error
-> keep pending multiplayer request unsent
```

This means multiplayer boot is gated by websocket auth, while single-player boot waits for semantic `realtime_transport_ready` when active gameplay will use WebRTC. Boot request dispatch must not race ahead of that signal.

## Protocols and APIs

### Boot request flow

Single-player flow:

```text
Pregame/Menu caller
-> MainMenuSessionController.request_single_player(local_profile_id)
-> SessionBootController.request_single_player(local_profile_id)
-> ClientSessionContext.request_single_player()
-> ShellBootFlow.request_single_player(local_profile_id)
-> ShellBootFlow.set_websocket_url(SINGLE_PLAYER_WS_URL)
-> ShellBootFlow.connect_to_game_server("single player")
-> ClientConnectionService.connect_to_server(url)
-> NetworkClient.connect_to_server(url)
-> ClientConnectionService ensures RealtimeTransportSession
-> RealtimeTransportSession.start()
-> WebRTCTransport offer, ICE, and ready flow
```

When the connection opens:

```text
  server webrtc_ready packet
  -> ServerPacketDispatcher.webrtc_ready_received
  -> ClientInboundCoordinator.handle_webrtc_ready
  -> ClientInboundCoordinator.realtime_transport_ready
  -> ClientConnectionService._on_realtime_transport_ready
  -> ClientConnectionService.realtime_transport_ready
  -> SessionNetworkController._on_realtime_transport_ready
  -> pending boot request gate
  -> _webrtc_gameplay_ready = true
  -> _try_send_pending_boot_request()
  -> ShellBootFlow.send_pending_boot_request()
```

Local `WebRTCTransport.ready` starts transport-local smoke diagnostics and does not directly release the boot gate. The server-confirmed `webrtc_ready` packet is what becomes semantic `realtime_transport_ready`.

Create-room flow:

```text
Pregame/Menu caller
-> MainMenuSessionController.request_create_room()
-> SessionBootController.request_create_room()
-> ClientSessionContext.request_multiplayer()
-> ShellBootFlow.request_create_room()
-> ShellBootFlow.set_websocket_url(MULTIPLAYER_WS_URL)
-> ShellBootFlow.connect_to_game_server("multiplayer create")
```

Join-room flow:

```text
Pregame/Menu caller
-> MainMenuSessionController.request_join_room(room_code)
-> trim and reject empty room code
-> SessionBootController.request_join_room(room_code)
-> ClientSessionContext.request_multiplayer()
-> ShellBootFlow.request_join_room(room_code)
-> ShellBootFlow.set_websocket_url(MULTIPLAYER_WS_URL)
-> ShellBootFlow.connect_to_game_server("multiplayer join: <room_code>")
```

Multiplayer requests are sent only after websocket auth succeeds or token verification is unavailable. If a multiplayer boot also requires active gameplay transport readiness, it still waits for `realtime_transport_ready` before dispatch.

### Connect result values

`ShellBootFlow.connect_to_game_server()` returns generated client shell result strings:

```text
started_connecting
already_connected
failed
```

If already connected, it sends the pending request immediately and returns `already_connected`.

If `ClientConnectionService.connect_to_server(url)` returns `OK`, it returns `started_connecting`.

Any other connection result returns `failed`.

### Client config handoff

`boot_request_sent` is connected by `AppEntry` to `ClientConfigController.send_client_config()`.

That sends viewport size after the boot request is sent, not when the websocket first opens.

The viewport config flow lives under client config ownership, not session boot ownership.

## Data ownership

The boot flow owns transient client-side request state only.

Owned local state includes:

```text
requested session mode
active session mode
pending boot request type
pending join room code
pending local profile id
selected websocket URL
```

The websocket target constants are generated from:

```text
shared/constants/client/shell.toml
shared/constants/client/presentation.toml
shared/constants/client/lobby.toml
```

Generated client output:

```text
client/scripts/generated/constants/constants.gd
```

Current generated constants used by this flow include:

```text
SINGLE_PLAYER_WS_URL
MULTIPLAYER_WS_URL
BOOT_REQUEST_NONE
BOOT_REQUEST_SINGLE_PLAYER
BOOT_REQUEST_CREATE_ROOM
BOOT_REQUEST_JOIN_ROOM
SESSION_MODE_NONE
SESSION_MODE_SINGLE_PLAYER
SESSION_MODE_MULTIPLAYER
CONNECT_RESULT_STARTED_CONNECTING
CONNECT_RESULT_ALREADY_CONNECTED
CONNECT_RESULT_FAILED
WEBRTC_ICE_SERVERS
```

The client shell constants also own the current client ICE server configuration seam. That seam is currently empty by default and does not require local development configuration. `SINGLE_PLAYER_WS_URL` and `MULTIPLAYER_WS_URL` are WebSocket signaling/control targets, not WebRTC UDP or data-channel addresses.

The client boot flow does not persist boot state.

## Ownership assessment

`session_network_target.gd` is not overloaded. It is appropriately small and currently only owns mode-to-URL resolution.

The broader session boot path has three ownership pressures:

```text
SessionBootController
= facade plus mode setting plus target selection plus connection start

ShellBootFlow
= pending request dispatch plus connection attempt result handling

SessionNetworkController
= boot auth gate plus room routing plus gameplay routing
```

The largest pressure is `SessionNetworkController`, because boot/auth gating is mixed with room and gameplay signal routing. That is current implementation, not a documentation problem.

This document should not hide that pressure, but it should not prescribe a refactor as current fact. A future split could move boot connection/auth-gate behavior into a narrower boot gate flow while leaving room/gameplay packet routing under networking/session routing.

## Code map

### Primary boot files

* `client/scripts/boot/session_boot_controller.gd`
* `client/scripts/boot/shell_boot_flow.gd`
* `client/scripts/boot/pending_boot_request.gd`
* `client/scripts/boot/session_network_target.gd`

### Session participants

* `client/scripts/session/client_session_context.gd`
* `client/scripts/session/session_network_controller.gd`
* `client/scripts/session/client_config_controller.gd`
* `client/scripts/config/client_viewport_config_flow.gd`

### App-shell participants

* `client/scripts/shell/app_entry.gd`
* `client/scripts/main_menu/main_menu_session_controller.gd`
* `client/scripts/ui/menu_flow/menu_flow_controller.gd`
* `client/scripts/ui/menu_flow/multiplayer_entry_flow.gd`

### Networking participants

* `client/scripts/networking/client_connection_service.gd`
* `client/scripts/networking/network_client.gd`
* `client/scripts/networking/inbound/client_inbound_coordinator.gd`
* `client/scripts/networking/inbound/server_packet_dispatcher.gd`
* `client/scripts/networking/webrtc/realtime_transport_session.gd`
* `client/scripts/networking/outbound/client_packet_sender.gd`
* `client/scripts/generated/networking/packets/packets.gd`

### Data and generated files

* `shared/constants/client/shell.toml`
* `shared/constants/client/presentation.toml`
* `shared/constants/client/lobby.toml`
* `client/scripts/generated/constants/constants.gd`

### Non-ownership boundaries

* `services/game-server/internal/networking/` owns server websocket session handling.
* `services/game-server/internal/rooms/` owns room membership and room lifecycle authority.
* `services/game-server/internal/game/` owns authoritative gameplay simulation.
* `shared/packets/` owns realtime packet source definitions.
* `docs/services/client/auth-session-flow.md` owns client auth session documentation.
* `docs/services/client/networking-flow/!INDEX.md` owns client websocket transport and packet-routing documentation.

## Tests

Relevant tests include:

* `client/tests/unit/boot/test_session_network_target.gd`
* `client/tests/unit/test_pending_boot_request.gd`
* `client/tests/unit/test_shell_boot_flow.gd`
* `client/tests/unit/test_session_network_controller.gd`
* `client/tests/unit/networking/inbound/test_client_inbound_coordinator.gd`
* `client/tests/unit/test_client_connection_service_webrtc.gd`
* `client/tests/unit/networking/test_webrtc_transport.gd`
* `client/tests/unit/test_client_connection_service_webrtc.gd`
* `client/tests/unit/ui/menu_flow/test_app_entry_menu_flow.gd`

`test_session_network_target.gd` verifies mode-to-URL mapping and unknown-mode fallback.

`test_pending_boot_request.gd` verifies request state recording, consumption, and clearing.

`test_shell_boot_flow.gd` verifies pending request classification and dispatch behavior.

`test_session_network_controller.gd` verifies that SessionNetworkController waits for semantic `realtime_transport_ready`, single-player sends without websocket auth, multiplayer waits for auth, auth success sends multiplayer boot, token verification unavailable still sends for server-side admission, and invalid-token auth failure leaves the multiplayer request unsent.

`test_app_entry_menu_flow.gd` verifies menu-to-boot routing for single-player, multiplayer create, and multiplayer join flows.

## Active issues

* `start_single_player_request` does not currently reject an already-authenticated websocket session at the server boundary. See [Current System Limits](../../../limits/current-system-limits.md#architecture--networking).

## Related docs

* [App Shell And Session](./!INDEX.md)
* [Client](../!INDEX.md)
* [Auth Session Flow](../auth-session-flow.md)
* [Menu Flow](../menu-flow.md)
* [Networking Flow](../networking-flow/!INDEX.md)
* [Gameplay Runtime](../gameplay-runtime/!INDEX.md)
* [Lobby Flow](../lobby-flow/!INDEX.md)
* [Current System Limits](../../../limits/current-system-limits.md)

## Notes

Legacy documentation correctly identified that the Go server route remains `/ws`, while the client selects a websocket target by session mode. That fact has been rewritten here against current client code and generated constants.

The current local development URLs for single-player and multiplayer are identical. The distinction still matters because the mode request and server-side policy are separate from the physical websocket route.

The filename should remain `session-boot-and-network-target.md`. The implementation file is `session_network_target.gd`, but this document covers the surrounding boot flow because the target selector only makes sense inside that flow.
