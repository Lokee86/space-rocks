---
author: brian
created: "2026-07-19"
document_id: 019f7d55-fb2c-77e5-9853-8a7f7473bd74
document_type: general
policy_exempt: false
summary: This document describes the client-side WebSocket connection lifecycle for Space Rocks.
---
# WebSocket Connection Lifecycle

Parent index: [Networking Flow](./!INDEX.md)

## Purpose

This document describes the client-side WebSocket connection lifecycle for Space Rocks.

It covers how the Godot client opens, polls, closes, sends raw packets, receives raw wire messages, and hands decoded packet dictionaries to the client networking flow.

## Overview

The client WebSocket lifecycle is owned by `NetworkClient` and exposed to the rest of the client through `ClientConnectionService`.

`NetworkClient` owns the direct `WebSocketPeer` interaction:

```text
connect to URL
set handshake headers
poll connection state
detect open and closed states
receive raw UTF-8 text packets
decode packet envelopes
emit decoded packet dictionaries
encode outgoing packet dictionaries
send raw wire messages
perform graceful close
```

`ClientConnectionService` owns the client-facing lifecycle facade:

```text
start connection polling
compose NetworkClient / ClientPacketSender / ServerPacketDispatcher / ClientInboundCoordinator / RealtimePacketPipeline / RealtimeTransportSession
own a typed AuthSessionController reference for websocket authentication
configure ClientInboundCoordinator and RealtimeTransportSession callbacks
forward connected and closed signals
forward packet parse failures
dispatch decoded inbound packets
expose connection-aware send methods
reset realtime protocol state on connect and close
send websocket authenticate request after connection when an auth token exists
clear websocket auth identity on close using NO_WEBSOCKET_AUTH_USER_ID (-1)
```

The lifecycle boundary stops when a decoded packet dictionary is emitted or dispatched. Packet-type routing belongs to [Inbound Packet Routing](inbound-packet-routing.md). Client packet helper families belong to [Outbound Packet Sending](outbound-packet-sending.md). Client logging implementation behavior belongs to [Client Logging](../client-logging.md).

## Responsibilities

* Own the client WebSocket connection lifecycle through `NetworkClient`.
* Connect to the configured websocket URL and set the Origin header before handshake.
* Poll the socket, observe open and closed state, and emit lifecycle signals.
* Receive raw UTF-8 wire messages and hand them to `PacketCodec` for decode.
* Encode outbound packet dictionaries and send raw wire messages over the active socket.
* Perform graceful close handoff through `ClientConnectionService` and `NetworkClient`.
* Expose the connection-aware websocket authenticate-request handoff after connect.

## Code root

* `client/scripts/networking/`
* `client/scripts/networking/packets/`

## Domain roles

`NetworkClient` is the transport owner.

It creates and manages the `WebSocketPeer`, but it does not decide what packets mean. Its packet awareness is limited to JSON encoding/decoding and minimal packet envelope validation.

`ClientConnectionService` creates the `NetworkClient`, installs signal connections, and exposes a stable service surface to session, room, gameplay, telemetry, and devtools code.

`ClientConnectionService` configures callbacks on `ClientInboundCoordinator` and `RealtimeTransportSession`. `ClientConnectionService` owns application-facing non-realtime dispatcher bindings and public facade signals. `ClientInboundCoordinator` owns only the five WebRTC control dispatcher bindings and the handlers exposed by `RealtimeTransportSession`. `RealtimePacketPipeline` owns gameplay lane dispatcher bindings. `NetworkClient` does not bind directly to the realtime transport, pipeline, or session collaborators; it only emits transport and decoded-packet signals to the facade.

The current runtime shape is:

```
  ClientConnectionService
    owns logical connection lifecycle and composition
    owns process polling
    owns NetworkClient
    owns ClientPacketSender
    owns ServerPacketDispatcher
    configures ClientInboundCoordinator callbacks
    configures RealtimeTransportSession callbacks

  NetworkClient
    owns WebSocketPeer
    owns connect / poll / close
    owns raw WebSocket wire send and receive
    owns PacketCodec decode handoff
    owns raw packet sending

  ClientInboundCoordinator
    owns the five WebRTC control dispatcher bindings

  ClientConnectionService
    owns application-facing non-realtime dispatcher bindings and public facade signals

  RealtimePacketPipeline
    owns gameplay lane dispatcher bindings, realtime gameplay packet application, and readiness

  RealtimeTransportSession
    owns realtime transport-session lifecycle and handlers
```

## Protocols and APIs

### Connection startup

The connection starts through `ClientConnectionService.connect_to_server(url)`.

That method:

```text
calls reset_realtime_session()
sets has_started_connection = true
delegates to NetworkClient.connect_to_server(url)
returns the WebSocket connect error code
```

`NetworkClient.connect_to_server(url)` resets graceful-close state, installs the configured WebSocket Origin header, and calls `WebSocketPeer.connect_to_url(url)`.

Origin selection follows the transport target:

```text
ws://localhost:<port> target -> Origin: http://localhost
ws://127.0.0.1:<port> target -> Origin: http://127.0.0.1
ws://[::1]:<port> target      -> Origin: http://[::1]
wss:// target                 -> Constants.MULTIPLAYER_WS_ORIGIN
```

Local native-client origins are derived from the target host while intentionally omitting the changing local server port. Hosted secure connections use the official public client origin `https://space-rocks.laughingskull.ca`.

The connection URL itself is not selected by this document's boundary. Session boot and network-target selection are documented separately.

### Poll loop

`ClientConnectionService` polls the network client from `_process()` only after a connection attempt has started.

```text
if has_started_connection && network_client != null:
    network_client.poll()
```

`ClientConnectionService` also sets its process priority from:

```text
Constants.NETWORK_POLL_PROCESS_PRIORITY
```

`NetworkClient.poll()` performs the transport-level work:

```text
socket.poll()
read socket ready state
emit connected_to_server once when state first becomes open
emit connection_closed when state becomes closed unexpectedly
drain all available socket packets
decode each packet as UTF-8 JSON
emit packet_received for valid decoded packets
emit packet_parse_failed for invalid packets
```

The poll loop drains all currently available socket packets during that frame. Invalid packets do not stop later available packets from being decoded.

### Open-state behavior

When `WebSocketPeer.STATE_OPEN` is observed for the first time, `NetworkClient` sets its local `connected` flag and emits:

```text
connected_to_server
```

`ClientConnectionService` handles that signal by:

```text
sending an authenticate request if an auth token exists
emitting connected
```

Authentication is opportunistic at this layer. If no auth session controller exists, no session exists, or the token is empty, no websocket authentication packet is sent.

The connection becoming open does not imply:

```text
authenticated websocket identity
room membership
ready state
active gameplay state
server authority over local player state
```

Those are higher-level states handled by other flow documents.

### Close behavior

`NetworkClient` distinguishes normal graceful close from unexpected close.

Unexpected close:

```text
WebSocketPeer.STATE_CLOSED
closing_gracefully == false
closed_notified == false
```

When those conditions are met, `NetworkClient` emits:

```text
connection_closed
```

`ClientConnectionService` handles that signal by clearing cached websocket auth identity:

```text
calls reset_realtime_session()
websocket_auth_authenticated = false
websocket_auth_user_id = NO_WEBSOCKET_AUTH_USER_ID
websocket_auth_display_name = ""
```

Then it emits:

```text
closed
```

### Realtime reset lifecycle

`ClientConnectionService.reset_realtime_session()` resets the connection-scoped realtime collaborators before a new connection starts or a closed connection clears its remaining state.

The reset sequence is:

```text
RealtimePacketPipeline.reset()
close the current RealtimeTransportSession, when present
clear the RealtimeTransportSession reference
update ClientInboundCoordinator with a null transport-session reference
```

The `RealtimePacketPipeline` object itself is preserved. Only its active realtime protocol state is reset, so consumers holding the pipeline reference continue to use the same object across connection lifecycles.

`RealtimePacketPipeline.reset()` replaces the active `RealtimeRouter`. That replacement clears world/overlay/session protocol state, pending lifecycle baseline buckets, pending lifecycle duplicate tracking, the latest applied asteroid lifecycle sequence, and the latest applied bullet lifecycle sequence. This reset is connection-scoped and is performed on connect and close; `GameplaySessionController.reset()` is a gameplay-session reset and does not perform this protocol reset.

`ClientConnectionService` emits the structured `realtime_protocol_state_reset` network diagnostic after the reset. The diagnostic reports the lifecycle action; it does not make `ClientConnectionService` the owner of realtime pipeline or transport-session state.

On close, `_on_closed()` completes this realtime reset before clearing connection-scoped websocket auth state and emitting `closed`. On connect, `connect_to_server()` performs the same reset before starting the new WebSocket connection.

### Graceful close

`ClientConnectionService.begin_graceful_close()` delegates to `NetworkClient.begin_graceful_close()`.

`NetworkClient.begin_graceful_close()` only proceeds if the socket is currently open or connecting. It then:

```text
sets closing_gracefully = true
sets closed_notified = true
sets connected = false
closes the socket with code 1000
uses close reason "client closed"
polls the socket once
```

The normal close code is:

```text
1000
```

`NetworkClient.close_gracefully()` begins the graceful close and then waits up to:

```text
0.25 seconds
```

During that wait it yields process frames, accumulates elapsed time, and polls the socket until the socket reports closed or the timeout is reached.

Graceful close suppresses the normal unexpected-close signal path.

### Packet decode lifecycle

Incoming socket packets are read as UTF-8 text and passed to `PacketCodec.decode(text)`.

`PacketCodec` owns wire parsing and minimal envelope validation only.

Compact realtime packets may arrive with `t` instead of `type`. `PacketCodec.decode` expands compact aliases to readable long-key dictionaries before envelope validation; see [Realtime Compact Wire Mapping](../../game-server/networking/realtime-compact-wire-mapping.md). Legacy long-key packets remain accepted, and packets with neither `type` nor compact `t` still fail validation.

A decoded packet must be a dictionary and must include a valid packet envelope after compact alias expansion:

```text
type must exist
type must be a String
type must not be empty after trimming
payload, when present, must be a Dictionary
```

If decoding fails, `NetworkClient` records the failure in runtime metrics and emits the canonical failure event through `ClientLogger.emit_canonical`, using the owning connection trace:

```text
event: packet_decode_failed
level: warn
category: network
context:
  trace_id: owning connection trace
fields:
  error
  raw_bytes
  raw_text_length
```

The same decode failure still emits:

```text
packet_parse_failed(text)
```

If decoding succeeds, `NetworkClient` emits:

```text
packet_received(packet)
```

`ClientConnectionService` forwards parse failures through its own `packet_parse_failed` signal.

`ClientConnectionService._on_packet_received(packet)` forwards decoded WebSocket packets to `ServerPacketDispatcher.dispatch(packet)`. Packet-specific application-facing outputs follow this route:

```text
ServerPacketDispatcher typed signal
    -> ClientConnectionService application-facing handler
    -> ClientConnectionService public facade signal

WebRTC control dispatcher signal
    -> ClientInboundCoordinator
    -> RealtimeTransportSession control handling

Gameplay lane dispatcher signal
    -> RealtimePacketPipeline typed apply method
    -> gameplay packet application
```

Application-facing non-realtime packet-specific public facade signals arrive from `ClientConnectionService` bindings. WebRTC control outputs are consumed by `ClientInboundCoordinator`, and realtime gameplay lane outputs are consumed by `RealtimePacketPipeline`. Connection-level `NetworkClient` signals remain separate from these packet routes. Detailed classification belongs to inbound packet routing rather than this lifecycle document.

### Packet encode lifecycle

Outgoing raw packets are sent through:

```text
NetworkClient.send_raw_packet(packet)
```

Before sending, `NetworkClient` checks:

```text
is_connected_to_server()
```

If the socket is not open, the packet is ignored.

If the socket is open, the packet dictionary is passed to:

```text
PacketCodec.encode(packet)
```

The current codec serializes the dictionary as JSON text. If encoding fails, `NetworkClient` records the failure in runtime metrics and emits the canonical failure event through `ClientLogger.emit_canonical`, using the owning connection trace:

```text
event: packet_encode_failed
level: warn
category: network
context:
  trace_id: owning connection trace
fields:
  error
  packet_type
```

If encoding succeeds, the encoded wire message is sent through:

```text
socket.send_text(encode_result.wire_message)
```

Most callers should not call `NetworkClient.send_raw_packet()` directly. The normal path is through `ClientConnectionService` and the outbound packet sender helpers.

### WebSocket authentication handoff

`ClientConnectionService` can receive an auth session controller through:

```text
set_auth_session_controller(auth_session_controller_ref)
```

After the websocket opens, `ClientConnectionService` attempts to send an authenticate request if a token exists.

The send path is:

```text
ClientConnectionService._on_connected()
ClientConnectionService._send_authenticate_request_if_token_exists()
ClientPacketSender.send_authenticate_request(token)
ClientPacketSender.send_packet(packet)
NetworkClient.send_raw_packet(packet)
```

The lifecycle layer does not validate the token, assign account identity, or enforce authorization. It only sends the authenticate request packet when an auth token is available.

The websocket authenticate result is received through the inbound packet dispatch path and cached by `ClientConnectionService`. `websocket_auth_user_id` is an integer cache using `NO_WEBSOCKET_AUTH_USER_ID` (`-1`) for no accepted identity, not `null`. An `authenticate_result` contributes an identity only when `authenticated` is `true` and `user_id` is an actual integer; missing, string, float, or unauthenticated values use the sentinel. That result handling is mentioned here only because close handling clears the cached identity.

### Signals

`NetworkClient` exposes these lifecycle and wire signals:

```text
connected_to_server
connection_closed
packet_received(data: Dictionary)
packet_parse_failed(text: String)
```

`ClientConnectionService` currently exposes this signal surface to the wider client:

```text
connected
closed
packet_parse_failed(text: String)
room_snapshot_received(packet: Dictionary)
websocket_auth_result_received(packet: Dictionary)
room_state_changed(packet: Dictionary)
room_error_received(packet: Dictionary)
debug_shape_catalog_received(packet: Dictionary)
debug_status_received(packet: Dictionary)
player_pause_state_received(packet: Dictionary)
telemetry_pong_received(packet: Dictionary)
realtime_transport_ready
unknown_packet_received(packet: Dictionary)
```

`ClientConnectionService` forwards `packet_parse_failed` from `NetworkClient`, emits `connected` after the transport opens and its realtime/auth handoffs run, emits `closed` after close handling resets cached websocket auth identity and realtime protocol state, and exposes application-facing packet-specific public facade signals through its WebSocket dispatcher and `ToolingPacketRouter` bindings. The semantic `realtime_transport_ready` facade signal is preserved by the coordinator's WebRTC-ready handling. Connection-level `NetworkClient` signals remain separate from packet-specific dispatcher and tooling-router signals.

## Does not own

This document does not own:

```text
session mode selection
single-player versus multiplayer URL policy
server websocket implementation
packet schema source-of-truth
generated packet constants
inbound packet type classification
payload-specific packet readers
gameplay state application
room state ownership
devtools packet semantics
telemetry interpretation
client packet helper families
server authority decisions
auth token validation
account identity authority
```

## Data ownership

`NetworkClient` owns transient transport state such as the live `WebSocketPeer`, connection flags, graceful-close state, and wire send/receive handoff.

`ClientConnectionService` owns the cached websocket auth identity fields that are cleared on close and the lifecycle signals exposed to the rest of the client.

Packet dictionaries and encoded wire text are transport payloads only. This document does not own packet schema, room state, gameplay state, or account identity.

## Code map

Primary implementation files:

```text
client/scripts/networking/network_client.gd
client/scripts/networking/client_connection_service.gd
client/scripts/networking/packets/packet_codec.gd
client/scripts/networking/packets/packet_encode_result.gd
client/scripts/networking/packets/packet_decode_result.gd
client/scripts/protocol/realtime/compact_lane_packet.gd
client/scripts/networking/realtime/realtime_packet_pipeline.gd
client/scripts/protocol/realtime/realtime_router.gd
client/scripts/protocol/realtime/lifecycle_lane_gate.gd
```

Related generated files:

```text
client/scripts/generated/constants/constants.gd
client/scripts/generated/networking/packets/packets.gd
```

Related tests:

```text
client/tests/unit/test_packet_codec.gd
client/tests/unit/networking/realtime/test_realtime_packet_pipeline.gd
client/tests/unit/test_client_connection_service.gd
```

```text
client/scripts/networking/inbound/server_packet_dispatcher.gd
client/scripts/networking/inbound/server_packet_router.gd
client/scripts/networking/outbound/client_packet_sender.gd
client/scripts/session/session_network_controller.gd
```

The adjacent files are listed to show handoff boundaries. They do not belong to this document's lifecycle ownership.

## Tests

Packet codec behavior is covered by `client/tests/unit/test_packet_codec.gd`.

`client/tests/unit/networking/realtime/test_realtime_packet_pipeline.gd` covers connection-scoped pipeline reset and replacement of the router-owned realtime protocol state.

`client/tests/unit/test_client_connection_service.gd` covers accepted integer websocket auth identities, malformed or unauthenticated results using `NO_WEBSOCKET_AUTH_USER_ID`, and close resetting the identity to the sentinel.

Those tests cover:

```text
JSON encode success
dictionary decode success
compact realtime alias decode succeeds
legacy long-key decode still succeeds
missing both `type` and compact `t` fails
invalid JSON rejection
non-dictionary JSON rejection
missing type rejection
empty type rejection
non-string type rejection
non-dictionary payload rejection
packet without payload acceptance
```

The WebSocket transport lifecycle itself depends on Godot `WebSocketPeer` runtime behavior and is primarily verified through integration behavior rather than a dedicated unit test in the current client test surface.

## Related docs

- [Networking Flow](./!INDEX.md)
- [Inbound Packet Routing](inbound-packet-routing.md) - Client inbound packet routing documentation.
- [Outbound Packet Sending](outbound-packet-sending.md)
- [Client Logging](../client-logging.md) - Client logger implementation and output behavior.
- [Session Boot And Network Target](../app-shell-and-session/session-boot-and-network-target.md)
- [Auth Session Flow](../auth-session-flow.md)
- [Realtime Websocket Protocol](../../../protocol/realtime-websocket-protocol.md)
- [Gameplay Packets](../../../protocol/gameplay-packets.md)
- [Lobby Packets](../../../protocol/lobby-packets.md)
- [Devtools Packets](../../../devtools/design/devtools-packet-protocol.md)
- [Packet Schema Pipeline](../../../data/packet-schemas.md) - shared packet schema and generated output documentation.

## Notes

`NetworkClient.connected` is a local lifecycle flag used for first-open notification, while `is_connected_to_server()` reads the actual `WebSocketPeer` ready state.

A successful WebSocket connection is only transport readiness. Authentication, room membership, gameplay participation, and player authority are separate states.

`PacketCodec` should stay small. If packet versioning, binary transport, compression, compatibility negotiation, or schema-level validation moves into the client codec, it should receive its own service or protocol document.

`NetworkClient` records packet decode and encode failures in runtime metrics and reports them through canonical `ClientLogger.emit_canonical(...)` calls with the owning connection trace. Compatibility logging details belong to [Client Logging](../client-logging.md).

`ClientConnectionService.reset_realtime_session()` preserves the `RealtimePacketPipeline` object identity while resetting its state, closing and clearing the active `RealtimeTransportSession`, and updating `ClientInboundCoordinator` with a null transport-session reference. Its websocket auth user-id cache uses the integer `NO_WEBSOCKET_AUTH_USER_ID` (`-1`) rather than `null` when no identity is accepted.

`SessionNetworkController` still uses text-helper logging for some connection and packet parse lifecycle messages through its configured logger callable.

This document describes where networking logs are emitted from current networking paths. Client logging implementation details belong to [Client Logging](../client-logging.md), packet schema details belong to protocol docs, and logs do not own packet routing or gameplay consequences.
