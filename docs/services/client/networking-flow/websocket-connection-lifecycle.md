# WebSocket Connection Lifecycle

Parent index: [Networking Flow](./!README.md)

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
forward connected and closed signals
forward packet parse failures
dispatch decoded inbound packets
expose connection-aware send methods
send websocket authenticate request after connection when an auth token exists
clear websocket auth identity on close
```

The lifecycle boundary stops when a decoded packet dictionary is emitted or dispatched. Packet-type routing belongs to [Inbound Packet Routing](inbound-packet-routing.md). Client packet helper families belong to [Outbound Packet Sending](outbound-packet-sending.md).

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

The current runtime shape is:

```text
ClientConnectionService
  owns process polling
  owns high-level networking signals
  owns NetworkClient child node
  owns ClientPacketSender child object
  owns ServerPacketDispatcher child node

NetworkClient
  owns WebSocketPeer
  owns connect / poll / close
  owns raw wire send and receive
  owns PacketCodec handoff
```

## Protocols and APIs

### Connection startup

The connection starts through `ClientConnectionService.connect_to_server(url)`.

That method:

```text
sets has_started_connection = true
delegates to NetworkClient.connect_to_server(url)
returns the WebSocket connect error code
```

`NetworkClient.connect_to_server(url)` resets graceful-close state, installs the configured websocket Origin header, and calls `WebSocketPeer.connect_to_url(url)`.

The Origin header is derived from generated constants:

```text
Constants.MULTIPLAYER_WS_ORIGIN
```

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
websocket_auth_authenticated = false
websocket_auth_user_id = null
websocket_auth_display_name = ""
```

Then it emits:

```text
closed
```

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

A decoded packet must be a dictionary and must include a valid packet envelope:

```text
type must exist
type must be a String
type must not be empty after trimming
payload, when present, must be a Dictionary
```

If decoding fails, `NetworkClient` logs a network warning and emits:

```text
packet_parse_failed(text)
```

If decoding succeeds, `NetworkClient` emits:

```text
packet_received(packet)
```

`ClientConnectionService` forwards parse failures through its own `packet_parse_failed` signal.

Decoded packets are dispatched by `ClientConnectionService._on_packet_received(packet)`, but packet classification and packet-type routing are owned by inbound packet routing, not by this lifecycle document.

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

The current codec serializes the dictionary as JSON text. If encoding fails, `NetworkClient` logs a network warning and does not send. If encoding succeeds, the encoded wire message is sent through:

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
NetworkClient.send_authenticate_request(token)
NetworkClient.send_raw_packet(Packets.authenticate_request_packet(token))
```

The lifecycle layer does not validate the token, assign account identity, or enforce authorization. It only sends the authenticate request packet when an auth token is available.

The websocket authenticate result is received through the inbound packet dispatch path and cached by `ClientConnectionService`. That result handling is mentioned here only because close handling clears the cached identity.

### Signals

`NetworkClient` exposes these lifecycle and wire signals:

```text
connected_to_server
connection_closed
packet_received(data: Dictionary)
packet_parse_failed(text: String)
```

`ClientConnectionService` exposes lifecycle-facing signals to the wider client:

```text
connected
closed
packet_parse_failed(text: String)
```

`ClientConnectionService` also exposes packet-specific signals, but those belong to inbound packet routing rather than the WebSocket lifecycle boundary.

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
```

Related generated files:

```text
client/scripts/generated/constants/constants.gd
client/scripts/generated/networking/packets/packets.gd
```

Related tests:

```text
client/tests/unit/test_packet_codec.gd
```

Important adjacent implementation files:

```text
client/scripts/networking/inbound/server_packet_dispatcher.gd
client/scripts/networking/inbound/server_packet_router.gd
client/scripts/networking/outbound/client_packet_sender.gd
client/scripts/session/session_network_controller.gd
```

The adjacent files are listed to show handoff boundaries. They do not belong to this document's lifecycle ownership.

## Tests

Packet codec behavior is covered by `client/tests/unit/test_packet_codec.gd`.

Those tests cover:

```text
JSON encode success
dictionary decode success
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

- [Networking Flow](./!README.md)
- [Inbound Packet Routing](inbound-packet-routing.md) - Client inbound packet routing documentation.
- [Outbound Packet Sending](outbound-packet-sending.md)
- [Session Boot And Network Target](../app-shell-and-session/session-boot-and-network-target.md)
- [Auth Session Flow](../auth-session-flow.md)
- [Realtime WebSocket Protocol](../../../protocol/stubs/realtime-websocket-protocol.md)
- [Packet Schema Pipeline](../../../data/stubs/packet-schema-pipeline.md)

## Notes

`NetworkClient.connected` is a local lifecycle flag used for first-open notification, while `is_connected_to_server()` reads the actual `WebSocketPeer` ready state.

A successful WebSocket connection is only transport readiness. Authentication, room membership, gameplay participation, and player authority are separate states.

`PacketCodec` should stay small. If packet versioning, binary transport, compression, compatibility negotiation, or schema-level validation moves into the client codec, it should receive its own service or protocol document.
