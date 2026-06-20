## Realtime WebSocket Protocol

Parent index: [Protocol](./!INDEX.md)

## Purpose

This document describes the current realtime WebSocket protocol between the Godot client and the Go game server.

It covers the transport route, JSON packet framing, connection lifecycle, packet-family routing, session-state requirements, delivery semantics, source-of-truth files, generated outputs, service responsibilities, compatibility expectations, and implementation code paths.

## Overview

The realtime protocol currently uses JSON text messages over a WebSocket connection.

The game server exposes one realtime route:

```text
GET /ws
```

The Godot client selects a WebSocket URL from the requested session mode, opens the connection, optionally sends an auth packet, sends room or gameplay request packets, and receives authoritative server packets.

The route path does not define play mode. Local single-player and multiplayer currently use the same local WebSocket route during development. Single-player versus multiplayer behavior is expressed through packets, session identity, room state, admission policy, and player-data routing.

The WebSocket connection itself is only transport readiness. It does not imply:

```text
authenticated account identity
room membership
ready state
active gameplay player state
durable Local Profile identity
durable account identity
```

The server owns authority behind accepted room, gameplay, auth-result, telemetry, and devtools consequences. The client owns connection initiation, packet emission, inbound packet classification, and presentation routing.

The protocol is best-effort and session-scoped. It currently has no durable acknowledgement, retry, reconnect, replay, sequence, baseline, delta snapshot, or negotiated version semantics.

## Participating systems

```text
client/scripts/networking/
```

Owns the client WebSocket peer, polling, raw send/receive, packet encode/decode handoff, inbound packet dispatch, outbound send wrappers, connection signals, and cached WebSocket auth result state.

```text
client/scripts/boot/
client/scripts/session/
```

Own session mode selection, pending boot request timing, WebSocket URL selection, auth-gated multiplayer boot dispatch, and routing from connection-service signals into room and gameplay controllers.

```text
services/game-server/internal/networking/
```

Owns the server WebSocket upgrade, session object, read loop, write loop, inbound packet-family routing, outbound queue, room-session adapter wiring, auth session state, telemetry pong handling, room request routing, gameplay request routing, and disconnect cleanup.

```text
services/game-server/internal/protocol/packetcodec/
```

Owns the server JSON encode/decode wrapper used by networking.

```text
shared/packets/
```

Owns realtime packet type strings, packet field names, selected generated structs, and client packet builders.

```text
services/game-server/internal/rooms/
```

Owns room membership, lobby rules, room lifecycle, game start, return-to-lobby behavior, match lifecycle state, and room cleanup policy.

```text
services/game-server/internal/game/
```

Owns authoritative gameplay simulation, input handling, respawn handling, pause state, targeting, state packet projection, gameplay events, scoring, lives, death, and match-over facts.

```text
services/game-server/internal/devtools/
```

Owns server-authoritative devtools command behavior and debug presentation inputs.

## Protocol authority

The realtime WebSocket protocol owns communication behavior between the client and game server.

It defines:

```text
transport route
wire framing
packet envelope expectations
packet-family routing order
client-to-server packet categories
server-to-client packet categories
delivery assumptions
session-state requirements
source/generated packet contract boundaries
```

It does not own:

```text
room rules
gameplay simulation rules
auth token issuance
Rails auth storage
Local Profile persistence
player-data store selection
client UI behavior
world rendering
devtools command effects
future delta/compression/binary protocol planning
```

Packet schema owns packet shape. Runtime services own packet meaning.

For example:

```text
shared/packets/gameplay.toml
-> defines input/state packet shape

client outbound flow
-> sends input intent

game-server networking
-> routes input packet

game-server game simulation
-> decides authoritative movement, firing, collision, score, death, and resulting state packet
```

## Wire surface

### Endpoint

The game-server process registers:

```text
GET /ws
```

The route is handled by the networking WebSocket handler and upgraded with Gorilla WebSocket.

### Origin policy

The server allows WebSocket upgrade requests with these origins:

```text
empty Origin header
https://space-rocks-client.local
http://localhost:8080
http://127.0.0.1:8080
http://[::1]:8080
```

The Godot client currently sets the WebSocket handshake origin from generated constants:

```text
Constants.MULTIPLAYER_WS_ORIGIN
```

Origin rejection or upgrade failure prevents session creation.

### Message framing

Each WebSocket message is a text message containing one JSON object.

The packet envelope uses:

```json
{
  "type": "packet_type"
}
```

Many packet types also include additional top-level fields or nested objects.

Client-side packet decode requires:

```text
JSON parses successfully
decoded value is a Dictionary
type exists
type is a String
type is not empty after trimming
payload, when present, is a Dictionary
```

Server-side initial envelope decode unmarshals the `type` field before routing. Invalid JSON or an envelope decode failure logs a warning and skips the message. A valid JSON object with an unknown or empty `type` does not produce an explicit protocol response in the current server path.

### Encoding

Client outbound encoding uses:

```text
JSON.stringify(packet)
```

Server outbound encoding uses:

```text
json.Marshal(packet)
```

The current protocol is JSON-only. There is no binary packet encoding, compression, protobuf encoding, schema negotiation, or version negotiation in the implemented transport.

## Connection lifecycle

Current connection flow:

```text
client selects WebSocket URL from requested session mode
-> client calls WebSocketPeer.connect_to_url(url)
-> server upgrades GET /ws
-> server creates one webSocketSession
-> session starts as Guest identity
-> server starts read loop goroutine
-> server starts gameplay lifecycle ticker goroutine
-> server runs write loop on the connection goroutine
-> client polls socket from ClientConnectionService._process()
-> client sends and receives JSON text packets while socket is open
-> read close, write failure, graceful close, or session exit tears down the connection
-> server leaves the current room if needed
-> server clears session room and active-player routing state
```

A new server session starts with:

```text
sessionID           = "session-" plus an atomic sequence number
identity            = Guest
currentRoomID       = empty
currentGamePlayerID = empty
room                = nil
outbound            = buffered channel, capacity 16
```

The session can later become authenticated by `authenticate_request`, attached to a room by room/lobby packets, and activated into gameplay by start-game or single-player start behavior.

## Client network target selection

The client selects a WebSocket URL from the requested session mode.

Current mapping:

```text
single_player -> SINGLE_PLAYER_WS_URL
multiplayer   -> MULTIPLAYER_WS_URL
unknown       -> ""
```

The current local values for single-player and multiplayer both point at:

```text
ws://localhost:8080/ws
```

The route path is therefore not the mode boundary. The mode boundary is the client request plus server-side packet handling, identity, room, and admission rules.

## Client-to-server packets

Client-originated packets are requests or observations. The server decides whether they are accepted and what state changes result.

Current client-to-server packet families are:

```text
auth
telemetry
lobby and room entry
gameplay
targeting
pause
client config
devtools
```

### Auth

```text
authenticate_request
```

The client sends this after the WebSocket opens when an auth token exists.

The packet does not itself make the client authenticated. The server verifies the token through the configured auth verifier and replies with `authenticate_result`.

Auth packets require only an active WebSocket session.

### Telemetry

```text
telemetry_ping
```

Telemetry ping carries:

```text
sequence
client_sent_msec
```

The server responds to the same WebSocket session with `telemetry_pong`.

Telemetry packets do not require room membership or active gameplay state and do not mutate gameplay.

### Lobby and room entry

```text
create_room_request
join_room_request
leave_room_request
set_ready_request
start_game_request
start_single_player_request
return_to_lobby_request
```

Lobby and room packets route to networking session handlers, which delegate room authority to the room system.

Current multiplayer create and join require Authenticated Account identity. Current single-player start does not require WebSocket authentication.

### Gameplay

```text
input
respawn
client_config
pause_request
set_target_player_request
select_target_at_position_request
clear_target_request
```

Normal gameplay application requires a current room and active game player.

For these packets:

```text
input
respawn
client_config
```

the server consumes the packet even when no room or active game player exists, but applies nothing.

For target and pause packets, the current handler requires a room and active game player before routing them. If that state is missing, the packet falls through unhandled.

### Devtools

Devtools command packets use the same WebSocket transport but route before normal generated `game.ClientPacket` decode.

Current devtools groups include:

```text
simple devtools commands
placement devtools commands
remaining devtools commands
```

Devtools command application requires a current room and active game player. If those are missing, the packet is consumed and no command is applied. This prevents devtools packets from falling through into normal gameplay routing.

Devtools packet behavior remains debug-only and server-authoritative.

## Server-to-client packets

Server-originated packets are authoritative readback, request results, or diagnostics.

Current server-to-client packet families include:

```text
auth result
room state and errors
gameplay state
player pause state
telemetry pong
debug status
debug shape catalog
```

### Queued one-off packets

Queued packets are encoded bytes placed into the session outbound channel and written by the WebSocket write loop.

Current queued producers include:

```text
authenticate_result
room_snapshot
room_error
player_pause_state
telemetry_pong
```

The queue is an in-memory handoff with capacity 16. It is not durable storage and has no retry or acknowledgement behavior.

### Ticker-driven packets

The server write loop runs at:

```text
constants.ServerTickRate
```

On eligible ticks, it writes:

```text
state
```

Gameplay state is eligible only when:

```text
session.currentGamePlayerID is not empty
room has an active game instance
room state is InGame or GameOver
```

The server stamps `server_sent_msec` before encoding the gameplay state packet.

When devtools are enabled and eligible, the same write loop can also send:

```text
debug_shape_catalog
debug_status
```

`debug_shape_catalog` is sent at most once per room ID in the current write-loop context. `debug_status` currently sends every 8 eligible gameplay presentation ticks.

### Room state packets

`room_snapshot` is the current active room-state broadcast packet. It is sent after room lifecycle changes such as create, join, ready change, game start, single-player start, return to lobby, leave, and disconnect broadcasts.

`room_state_changed` is generated and recognized by the client inbound router, but the current server room lifecycle path uses `room_snapshot` as the practical room-state update packet.

## Server inbound routing order

The server inbound routing order is:

```text
simple devtools packets
placement devtools packets
remaining devtools packets
normal game.ClientPacket decode
auth packets
telemetry packets
lobby packets
gameplay packets
```

Devtools packets route before normal packet decode because generated devtools command structs live under the server devtools package, not the normal generated game packet family.

Normal packets decode into:

```text
game.ClientPacket
```

Decode failure logs:

```text
websocket packet decode failed
```

and the packet is not routed further.

If no packet-family handler consumes a decoded packet, the server currently returns without applying it and without sending an unknown-packet response.

## Client inbound routing

The client inbound path begins after raw WebSocket text has decoded into a packet dictionary.

Current flow:

```text
NetworkClient.poll()
-> raw WebSocket text received
-> PacketCodec.decode(text)
-> NetworkClient.packet_received(packet)
-> ClientConnectionService._on_packet_received(packet)
-> ServerPacketDispatcher.dispatch(packet)
-> ServerPacketRouter packet-type checks
-> typed dispatcher signal
-> ClientConnectionService typed signal
-> session, room, gameplay, telemetry, or devtools consumer
```

Current client-recognized inbound packet types include:

```text
authenticate_result
room_snapshot
room_state_changed
room_error
state
debug_shape_catalog
debug_status
player_pause_state
telemetry_pong
```

Unrecognized packets with a valid envelope emit:

```text
unknown_packet_received(packet)
```

Packet parse failures emit:

```text
packet_parse_failed(text)
```

and do not enter typed routing.

## Session state requirements

Packet families have different runtime requirements.

```text
authenticate_request
requires WebSocket session only

telemetry_ping
requires WebSocket session only

create_room_request
requires Authenticated Account identity

join_room_request
requires Authenticated Account identity

start_single_player_request
requires WebSocket session and no current room

set_ready_request
requires current room/session membership

start_game_request
requires current room/session membership and room start rules

return_to_lobby_request
requires current room/session membership and room return rules

input, respawn, client_config
require current room and active game player to apply

target and pause packets
require current room and active game player to route

devtools command packets
require current room and active game player to apply

state output
requires current room, active game player, and eligible room game state
```

The protocol preserves this separation:

```text
WebSocket session ID
!= room member identity
!= active gameplay player ID
!= account ID
!= Local Profile ID
```

## Authentication protocol behavior

Every server WebSocket session starts as Guest.

If the client has an auth token, the client sends:

```json
{
  "type": "authenticate_request",
  "token": "<space-rocks-bearer-token>"
}
```

The server verifies the token through the configured token verifier. When verification succeeds, the session identity becomes Authenticated Account identity and stores:

```text
Rails user_id
cross-system account_id
display name
```

The server replies with:

```json
{
  "type": "authenticate_result",
  "authenticated": true,
  "user_id": 123,
  "display_name": "Ada"
}
```

On failure, the server replies with:

```json
{
  "type": "authenticate_result",
  "authenticated": false,
  "error_code": "invalid_token"
}
```

Current auth failure codes are:

```text
invalid_token
token_verification_unavailable
```

Auth failure does not close the WebSocket. The session remains connected as Guest unless another flow ends the connection.

The game server must not log bearer tokens and must not use bearer tokens as gameplay identity.

## Telemetry protocol behavior

Telemetry ping/pong is diagnostic transport traffic.

Client-to-server:

```json
{
  "type": "telemetry_ping",
  "sequence": 1,
  "client_sent_msec": 123456
}
```

Server-to-client:

```json
{
  "type": "telemetry_pong",
  "sequence": 1,
  "client_sent_msec": 123456,
  "server_received_msec": 123500,
  "server_sent_msec": 123501
}
```

The server replies only to the same WebSocket session. Telemetry does not require room membership, does not require active gameplay state, and does not mutate gameplay.

## Delivery and failure semantics

Current delivery is best-effort over the active WebSocket session.

There is no implemented support for:

```text
packet acknowledgements
server resend
client resend
reconnect recovery
session resume
baseline tracking
packet sequence enforcement
delta snapshot recovery
durable outbound queues
```

Client outbound sends are not queued. If the WebSocket is not open, the packet is not sent.

Server queued outbound messages use a bounded in-memory channel. If a WebSocket write fails, the session write loop exits and normal connection teardown begins.

Decode failure behavior:

```text
client invalid inbound JSON
-> client emits packet_parse_failed

server invalid envelope JSON
-> server logs envelope decode warning and continues

server normal packet decode failure
-> server logs decode warning and continues

server unknown decoded packet
-> no response, no state mutation

client unknown decoded packet
-> unknown_packet_received signal
```

Close behavior:

```text
client graceful close
-> close code 1000, reason "client closed"

server expected read close
-> debug log

server unexpected read failure
-> warn log

server expected write close
-> debug log

server unexpected write failure
-> error log
```

## Source-of-truth files

Realtime packet shapes are sourced from:

```text
shared/packets/gameplay.toml
shared/packets/lobby.toml
shared/packets/debug.toml
shared/packets/outputs.toml
```

Those files define packet structs, packet type strings, JSON field names, output routing, and selected generated client builders.

The transport route and runtime connection lifecycle are not sourced from the packet TOML files. They are implemented by the client and game-server networking services.

## Generated outputs

Current generated outputs used by the realtime WebSocket protocol include:

```text
client/scripts/generated/networking/packets/packets.gd
services/game-server/internal/game/packets.go
services/game-server/internal/game/runtime/packets_generated.go
services/game-server/internal/devtools/packets_generated.go
```

The generated client file provides packet type constants, field constants, and selected outbound packet builder functions.

The generated server files provide packet constants and Go structs for game, lobby, auth, telemetry, runtime state, and devtools packet families.

Generated files are outputs, not edit sources.

## Service responsibilities

### Client

The client owns:

```text
session-mode WebSocket URL selection
WebSocket connection initiation
Origin header setup
socket polling
raw JSON text send/receive
client packet encode/decode wrapper
outbound packet helper calls
inbound packet classification
auth result cache on the connection service
connection and packet signals
routing packets to room, gameplay, telemetry, and devtools consumers
```

The client does not own:

```text
server acceptance of packets
room authority
gameplay authority
auth token verification
durable player-data writes
retry or reconnect semantics
```

### Game server networking

Game-server networking owns:

```text
GET /ws upgrade
origin check
per-connection WebSocket session
server-internal session ID
session identity mutation after auth
read loop
write loop
inbound packet family routing
same-session telemetry pong
outbound queue
room request adapter calls
gameplay request adapter calls
room-session attachment registry
active game-player routing field
disconnect cleanup
```

Game-server networking does not own:

```text
room lifecycle rules
gameplay simulation rules
packet schema source files
auth token issuance
Rails auth tables
player-data store selection
client presentation
future realtime delivery policy
```

### Rooms

Rooms own room authority behind lobby and room-entry packets:

```text
room creation
room join
room leave
ready state
owner selection
start-game acceptance
single-player room creation
return to lobby
room cleanup
match lifecycle state
resolved match summary availability
```

### Game simulation

The game simulation owns gameplay authority behind gameplay packets and state output:

```text
input application
movement
projectile firing
respawn
pause state
target state
collisions
damage
scoring
lives
death
pickup state
event projection
state packet projection
match-over policy integration
```

### Devtools

Devtools own debug command behavior after networking identifies a devtools packet.

Devtools use the normal WebSocket transport but must not bypass real server-owned gameplay seams.

### Packet schema pipeline

The packet schema pipeline owns packet shape and generated outputs.

It does not own runtime authority, WebSocket delivery mechanics, client UI, room rules, or game simulation.

## Compatibility expectations

The current compatibility boundary is the shared packet schema and generated output pipeline.

Stable protocol facts include:

```text
packet type strings come from shared packet source files
JSON field names come from shared packet source files
client and server generated outputs must be updated together
generated files should not be hand-edited
runtime handlers must be updated when new packet types are added
```

There is no runtime version negotiation. A client and server built from mismatched packet outputs can drift.

Packet schema changes should follow the packet schema pipeline:

```text
data-sync -validate -packets
data-sync -diff -packets -go -gds
data-sync -push -packets -go -gds
data-sync -check -packets -go -gds
```

Future lane policy, delta snapshots, sequence numbers, quantization, bit packing, and protobuf migration are planning facts until implemented.

## Code map

### Client transport and codec

```text
client/scripts/networking/network_client.gd
client/scripts/networking/client_connection_service.gd
client/scripts/networking/packets/packet_codec.gd
client/scripts/networking/packets/packet_encode_result.gd
client/scripts/networking/packets/packet_decode_result.gd
```

### Client inbound and outbound routing

```text
client/scripts/networking/inbound/server_packet_dispatcher.gd
client/scripts/networking/inbound/server_packet_router.gd
client/scripts/networking/outbound/client_packet_sender.gd
client/scripts/networking/outbound/gameplay_client_packets.gd
client/scripts/networking/outbound/lobby_client_packets.gd
client/scripts/networking/outbound/devtools_client_packets.gd
client/scripts/networking/outbound/telemetry_client_packets.gd
```

### Client boot/session participants

```text
client/scripts/boot/session_boot_controller.gd
client/scripts/boot/shell_boot_flow.gd
client/scripts/boot/pending_boot_request.gd
client/scripts/boot/session_network_target.gd
client/scripts/session/session_network_controller.gd
client/scripts/session/room_session_controller.gd
client/scripts/session/gameplay_session_controller.gd
```

### Server WebSocket lifecycle

```text
services/game-server/cmd/game-server/main.go
services/game-server/internal/networking/websocket.go
services/game-server/internal/networking/websocket_origin.go
services/game-server/internal/networking/websocket_session.go
services/game-server/internal/networking/websocket_read.go
services/game-server/internal/networking/websocket_write.go
services/game-server/internal/networking/websocket_gameplay_tick.go
services/game-server/internal/networking/websocket_close_logging.go
```

### Server inbound routing

```text
services/game-server/internal/networking/client_packet_router.go
services/game-server/internal/networking/inbound_adapter.go
services/game-server/internal/networking/inbound/client_packet_envelope.go
services/game-server/internal/networking/inbound/router.go
services/game-server/internal/networking/inbound/auth.go
services/game-server/internal/networking/inbound/telemetry.go
services/game-server/internal/networking/inbound/lobby.go
services/game-server/internal/networking/inbound/gameplay.go
services/game-server/internal/networking/inbound/devtools.go
```

### Server outbound routing

```text
services/game-server/internal/networking/outbound/server_message_writer.go
services/game-server/internal/networking/outbound/gameplay_presentation.go
services/game-server/internal/networking/outbound/gameplay_state_metrics.go
services/game-server/internal/networking/outbound/debug_status_presentation.go
services/game-server/internal/networking/outbound/debug_shape_catalog_presentation.go
services/game-server/internal/networking/room_snapshot.go
services/game-server/internal/networking/room_error.go
services/game-server/internal/networking/session_auth.go
services/game-server/internal/networking/player_pause_state.go
```

### Server session, auth, and room adapter files

```text
services/game-server/internal/networking/session_identity.go
services/game-server/internal/networking/session_auth.go
services/game-server/internal/networking/session_admission.go
services/game-server/internal/networking/room_handlers.go
services/game-server/internal/networking/room_sessions.go
services/game-server/internal/networking/player_activation.go
```

### Server packet codec and generated packet files

```text
services/game-server/internal/protocol/packetcodec/codec.go
services/game-server/internal/game/packets.go
services/game-server/internal/game/runtime/packets_generated.go
services/game-server/internal/devtools/packets_generated.go
```

### Packet source files

```text
shared/packets/gameplay.toml
shared/packets/lobby.toml
shared/packets/debug.toml
shared/packets/outputs.toml
```

### Important non-ownership boundaries

```text
services/game-server/internal/rooms/
services/game-server/internal/game/
services/game-server/internal/devtools/
services/player-data/
services/api-server/
docs/data/packet-schemas.md
docs/planning/protocol/realtime-protocol-architecture.md
```

## Validation and testing

Packet schema validation:

```text
data-sync -validate -packets
data-sync -check -packets -go -gds
```

Focused game-server networking validation:

```text
cd services/game-server && go test -buildvcs=false ./internal/networking ./internal/networking/outbound ./internal/rooms ./internal/game/rules ./cmd/game-server
```

Broader game-server validation:

```text
cd services/game-server && go test -buildvcs=false ./...
```

Client packet and networking-adjacent validation:

```text
godot --headless --path client -s res://addons/gut/gut_cmdln.gd -gdir=res://tests/unit -ginclude_subdirs -gexit
```

Relevant server tests include:

```text
services/game-server/internal/networking/websocket_test.go
services/game-server/internal/networking/gameplay_packets_test.go
services/game-server/internal/networking/session_auth_test.go
services/game-server/internal/networking/session_identity_test.go
services/game-server/internal/networking/player_activation_test.go
services/game-server/internal/networking/room_sessions_test.go
services/game-server/internal/networking/room_snapshot_test.go
services/game-server/internal/networking/room_error_test.go
services/game-server/internal/networking/outbound/gameplay_presentation_test.go
services/game-server/internal/networking/outbound/debug_status_presentation_test.go
services/game-server/internal/networking/outbound/debug_shape_catalog_presentation_test.go
```

Relevant client tests include:

```text
client/tests/unit/test_packet_codec.gd
client/tests/unit/test_session_network_controller.gd
client/tests/unit/test_room_session_controller.gd
client/tests/unit/test_gameplay_session_controller.gd
client/tests/unit/boot/test_session_network_target.gd
client/tests/unit/test_shell_boot_flow.gd
client/tests/unit/test_pending_boot_request.gd
client/tests/unit/test_gameplay_input_context.gd
client/tests/unit/test_target_request_flow.gd
client/tests/unit/devtools/telemetry/test_network_telemetry_metrics.gd
client/tests/unit/devtools/telemetry/test_world_telemetry_context.gd
```

`test_packet_codec.gd` verifies client JSON packet encode/decode and envelope validation.

`gameplay_packets_test.go` verifies current gameplay packet routing behavior, including `client_config` routing into the game instance and `start_single_player_request` routing through lobby handling while preserving Local Profile ID on the room member.

`session_auth_test.go` verifies session identity mutation after successful WebSocket auth.

## Active issues

* `start_single_player_request` does not currently reject an already-authenticated WebSocket session at the server boundary. See [Current System Limits](../limits/current-system-limits.md#architecture--networking).

## Related docs

* [Protocol](./!README.md)
* [Realtime Client Server Flow](../domains/technical/realtime-client-server-flow.md)
* [Gameplay Session Flow](../domains/player-experience/gameplay-session-flow.md)
* [Client](../services/client/!README.md)
* [Client Networking Flow](../services/client/networking-flow/!README.md)
* [WebSocket Connection Lifecycle](../services/client/networking-flow/websocket-connection-lifecycle.md)
* [Client Outbound Packet Sending](../services/client/networking-flow/outbound-packet-sending.md)
* [Client Inbound Packet Routing](../services/client/networking-flow/inbound-packet-routing.md)
* [Session Boot And Network Target](../services/client/app-shell-and-session/session-boot-and-network-target.md)
* [Game Server](../services/game-server/!README.md)
* [Game Server Networking](../services/game-server/networking/!README.md)
* [WebSocket Session Lifecycle](../services/game-server/networking/websocket-session-lifecycle.md)
* [Game Server Inbound Packet Routing](../services/game-server/networking/inbound-packet-routing.md)
* [Game Server Outbound Packet Routing](../services/game-server/networking/outbound-message-flow.md)
* [Auth Routing](../services/game-server/networking/auth-routing.md)
* [Telemetry Packet Routing](../services/game-server/networking/telemetry-packet-routing.md)
* [Game Server Rooms](../services/game-server/rooms/!README.md)
* [Game Server Simulation](../services/game-server/simulation/!README.md)
* [Packet Schemas](../data/packet-schemas.md)
* [Source Of Truth Map](../data/source-of-truth-map.md)
* [Realtime Protocol Architecture](../planning/protocol/realtime-protocol-architecture.md)
* [Network Observability And Packet Budget](../planning/domains/technical/network-observability-and-packet-budget.md)
* [Current System Limits](../limits/current-system-limits.md)

## Notes

The current implementation sends full gameplay presentation state on the server tick path. That is current protocol behavior, not the intended final realtime architecture.

The current WebSocket protocol is transport/session scoped. Durable match-result persistence happens through player-data routing after authoritative match facts are produced; it is not a WebSocket delivery guarantee.

The generated packet schema defines the shared packet vocabulary, but service implementation still determines runtime consequences. New packets should update source TOML, generated outputs, runtime handlers, tests, and protocol documentation together.
