# Outbound Packet Routing

Parent index: [Game Server Networking](./!INDEX.md)

## Purpose

This document describes the game server outbound WebSocket service boundary for per-session packet delivery.

## Overview

Game-server outbound routing is the server-side send path for WebSocket messages written to a connected client session.

The outbound boundary has three current responsibilities:

1. Queued one-off responses produced by request handlers.
2. Ticker-driven active realtime lane packet writes produced by the websocket write loop.
3. Debug shape catalog writes when devtools are enabled.

Queued responses and lane packets both converge at `outbound.WriteServerMessage()`, which writes a WebSocket text message through the active Gorilla WebSocket connection.

The networking layer owns connection/session write mechanics and message delivery. The realtime protocol package owns lane packet construction, baseline policy, candidate selection, quantization, and wire-shape assembly. Outbound routing delivers already projected and quantized gameplay lane packets; it does not decide realtime packet schema policy or quantization policy.

## Code root

`services/game-server/internal/networking/`

The focused outbound helper package is:

`services/game-server/internal/networking/outbound/`

The realtime packet construction package is:

`services/game-server/internal/protocol/realtime/`

## Responsibilities

The game-server outbound packet routing path owns:

- Per-session outbound message delivery over WebSocket text frames.
- The session outbound queue used by one-off responses.
- Ticker-driven active realtime lane packet writes.
- One-time debug shape catalog writes per room connection context when devtools are enabled.
- Encoding already-built server packet structs through `packetcodec`.
- Writing encoded packets through the active Gorilla WebSocket connection.
- Logging outbound encode failures and write closes.
- Invoking the realtime active-result send path from the websocket write loop.

## Does not own

The outbound packet routing path does not own:

- WebSocket upgrade policy.
- Inbound packet classification.
- Room membership rules.
- Lobby readiness rules.
- Match lifecycle rules.
- Gameplay simulation authority.
- Player-data persistence.
- Packet schema source-of-truth.
- Client-side packet decoding or presentation.
- Retry, acknowledgement, resend, or durable outbound queues.
- Realtime lane candidate selection, baseline policy, or packet prioritization.
- Future compact encoding, transport mapping, or remaining packet-budget work.

## Domain roles

The outbound routing surface is the server-to-client WebSocket packet path.

The client consumes these messages after the Godot networking layer decodes WebSocket text and classifies packets by `type`.

The server owns authority behind the payloads. The client should treat outbound server packets as authoritative readback or authoritative request results, not as local decisions.

The current outbound payloads include queued one-off responses plus lane-native realtime packets and debug packets.

## Routing model

### Connection write loop

`handleConnection()` starts the connection runtime by creating a `webSocketSession`, starting `readClientInput()` in a goroutine, starting `tickSessionGameplayLifecycle()` in a goroutine, and running `writeServerMessages()` on the connection goroutine.

`writeServerMessages()` owns outbound delivery for the session. It selects between read-loop close errors, queued outbound messages from `session.outbound`, and server tick events.

If the read loop reports a close or error, the write loop logs the read close and returns.

If a WebSocket write fails, `outbound.WriteServerMessage()` invokes the write-close logger and returns `false`. The write loop then returns and the connection teardown path runs.

### Session outbound queue

Each `webSocketSession` owns:

`outbound chan []byte`

The channel is created with a buffer size of 16 in `newWebSocketSession()`.

Queued responses are already encoded byte payloads. They are written by the `session.outbound` branch in `writeServerMessages()`, which passes the encoded bytes to `outbound.WriteServerMessage(session.conn, message, onWriteClose)`.

The queue is not durable. It is a small in-memory handoff between handlers and the write loop, with no retry or acknowledgement guarantee.

### Queued response producers

Several server handlers build packets, encode them through `packetcodec.Encode()`, and enqueue the encoded bytes.

Current queued producers include:

- `EnqueueAuthenticateResult()`
- `EnqueueRoomSnapshot()`
- `BroadcastRoomSnapshot()`
- `EnqueueRoomError()`
- `EnqueuePlayerPauseState()`
- telemetry pong handling through `inbound.HandleTelemetryPacket()`

`BroadcastRoomSnapshot()` does not write directly to every socket. It snapshots the room's attached sessions and calls `session.EnqueueRoomSnapshot(room)` for each session. Each session then writes the packet through its own outbound queue and write loop.

Queued response producers generally encode with `packetcodec` before enqueue, and the queued packets converge at `outbound.WriteServerMessage()`.

### Ticker-driven active lane writes

`writeServerMessages()` runs a ticker at `constants.ServerTickRate`.

On each tick, gameplay lane output is eligible only when:

- `session.currentGamePlayerID` is not empty
- `session.room` is not nil
- `session.room.GameInstance()` is not nil

When eligible, `writeServerMessages()` calls `writeGameplayLaneProtocolMessage(session, remoteAddr)`.

`writeGameplayLaneProtocolMessage()` currently:

1. Writes debug shape catalog output first when eligible.
2. Resets `session.realtimeState` when the receiver is empty or changes.
3. Calls `realtime.BuildActiveRealtimeResultForGame()`.
4. Selects included lane candidates from the send plan.
5. Encodes the selected lane candidates through the realtime protocol package.
6. Writes each encoded lane packet individually through `outbound.WriteServerMessage()`.
7. Logs lane wire packet details after successful writes.
8. Drains active event batch events only after a successful event batch write.
9. Persists lane metadata only after successful writes.
10. Stores baseline projections for non-event lane packets after successful writes.
11. Marks a lane baseline ready after a final full packet.
12. Logs sent lane metrics and summary fields.

The lane packet construction path lives in `services/game-server/internal/protocol/realtime/`. That package owns candidate building, full/delta decisions, scheduling, wire-map encoding, and lane metric records.

The networking layer owns successful WebSocket delivery and the post-write session state changes that follow from those successful writes.

Chunk metadata exists in the wire shape and scheduler records, but this section does not claim full fragmentation or payload-splitting behavior beyond current final-chunk handling.

### Debug status

`debug_status` is built by `outbound.BuildDebugStatusResponse()` and covered by tests, but this doc does not describe an active write-loop call path for it.

The builder requires a non-nil room, a non-nil game instance, `devtools.Enabled()`, and room state `InGame` or `GameOver`.

The packet is built with `devtools.StatusFor()` and `devtools.StatusesForAllPlayers()`, then encoded through `packetcodec`.

### Ticker-driven debug shape catalog

`maybeWriteDebugShapeCatalog()` sends a `debug_shape_catalog` packet at most once for the current room ID tracked by that write loop.

It is eligible only when:

- `session.currentRoomID` is not empty
- the current room ID has not already received the shape catalog from this write loop
- `outbound.CanSendDebugShapeCatalog(session.room)` returns true

`CanSendDebugShapeCatalog()` uses the same devtools and room/game-state gates as debug status.

The packet is built from `physics.LoadCollisionShapeCatalog()` and `devtools.BuildShapeCatalog()`, then encoded through `packetcodec`.

## Packet sources

### Shared packet sources and generated outputs

Most outbound packet structs are generated from shared packet definitions and realtime protocol code.

Source-of-truth files include:

- `shared/packets/outputs.toml`
- `shared/packets/gameplay.toml`
- `shared/packets/lobby.toml`
- `shared/packets/debug.toml`

Generated and runtime outputs include:

- `services/game-server/internal/protocol/realtime/packets_generated.go`
- `services/game-server/internal/protocol/realtime/`
- `services/game-server/internal/game/packets.go`
- `services/game-server/internal/game/runtime/packets_generated.go`
- `services/game-server/internal/devtools/packets_generated.go`

### JSON packet codec

Server outbound encoding uses:

`services/game-server/internal/protocol/packetcodec/codec.go`

`packetcodec.Encode(packet)` currently wraps `json.Marshal(packet)`.

The outbound route does not own the packet schema or wire-format strategy. The current implementation is JSON text over WebSocket. The realtime protocol package owns lane packet construction and wire-map assembly; networking only encodes and writes the resulting packets.

## Packet families

### Queued non-lane packets

These packets are queued one-off responses or direct diagnostic packets, not active lane writes:

- `authenticate_result`
- `room_snapshot`
- `room_error`
- `player_pause_state`
- `telemetry_pong`
- `debug_shape_catalog`

`authenticate_result` is queued by `EnqueueAuthenticateResult()`.
`room_snapshot` is queued by `EnqueueRoomSnapshot()` and `BroadcastRoomSnapshot()`.
`room_error` is queued by `EnqueueRoomError()`.
`player_pause_state` is queued by `EnqueuePlayerPauseState()`.
`telemetry_pong` is queued through inbound telemetry handling.
`debug_shape_catalog` is written from the write loop when devtools are enabled and the room gate allows it.

Queued producers generally encode through `packetcodec` before enqueue. The queued packets then converge at `outbound.WriteServerMessage()`.

`BroadcastRoomSnapshot()` still broadcasts by enqueueing per attached room session rather than writing directly to every socket.

### Active realtime lane packets

Active gameplay output is written as lane packet families, not as a single `state` packet.

Current lane families are:

- `world_full`
- `world_delta`
- `overlay_full`
- `overlay_delta`
- `session_full`
- `session_delta`
- `event_batch`
- `resync_request`
- `resync_required`

Lane roles at service level are:

- world = authoritative world entities visible through presentation projection
- overlay = receiver-specific HUD-facing values
- session = player, session, lifecycle, and asteroid-count presentation
- event = one-shot presentation event batches
- control = control/resync lane family placeholder/path

Lane packet metadata carries:

- `lane`
- `sequence`
- `baseline_id`
- `snapshot_id`
- `server_sent_msec`
- `snapshot_kind`
- `chunk_index`
- `chunk_count`
- `is_final_chunk`

The packet-shape details for those lane packets belong in the realtime protocol doc. This service doc only keeps the outbound delivery boundary and the current lane roles.

### Room snapshots

Packet type:

`room_snapshot`

Owned payload builder:

`BuildRoomSnapshot()`

Room snapshots are produced after room lifecycle changes such as create, join, ready, start game, single-player start, return to lobby, and leave/disconnect broadcasts.

`BuildRoomSnapshot()` includes room code, room state, members, local player ID, owner ID, max players, and resolved match result summary when one exists.

### Room errors

Packet type:

`room_error`

Owned payload builder:

`EnqueueRoomError()`

Room errors are per-session one-off responses for rejected room actions or invalid room state.

### Authentication result

Packet type:

`authenticate_result`

Owned payload builder:

`EnqueueAuthenticateResult()`

Authentication results are one-off responses to `authenticate_request`. The auth verifier and identity mutation path decide whether the request succeeds. Outbound routing only encodes and queues the result packet.

### Player pause state

Packet type:

`player_pause_state`

Owned payload builder:

`EnqueuePlayerPauseState()`

The game instance owns the pause state packet through `PlayerPauseStatePacket(playerID)`. Networking encodes and queues it for the current session after inbound pause handling routes through the adapter.

### Telemetry pong

Packet type:

`telemetry_pong`

Owned payload builder:

`inbound.HandleTelemetryPacket()`

Telemetry pong is generated as a direct response to `telemetry_ping`. It is diagnostic transport behavior, not gameplay mutation. It preserves the client sequence and timing fields, stamps server receive/send times, encodes the response, and queues it to the same session's outbound channel.

### Debug status

Packet type:

`debug_status`

Owned payload builder:

`outbound.BuildDebugStatusResponse()`

Debug status is devtools-only presentation. It is sent only when devtools are enabled and the session has an active game player in an `InGame` or `GameOver` room.

### Debug shape catalog

Packet type:

`debug_shape_catalog`

Owned payload builder:

`outbound.BuildDebugShapeCatalogResponse()`

Debug shape catalog is devtools-only shape metadata. It is sent once per room ID in the current write-loop context when eligible.

## Failure behavior

Outbound encode failures are logged and the packet is dropped.

Outbound write failures end the write loop for that WebSocket session. The connection teardown path closes the socket and leaves the disconnected room when needed.

The session outbound queue is not a durable delivery guarantee. It is a bounded in-memory handoff. Senders that write into the queue can block when the buffer is full.

## Observability

Lane writes currently emit current lane metrics through `packetmetrics.LogSentLaneMetrics(result.MetricSummaries, ...)` and structured debug logs such as `lane protocol gameplay wire packet written` and `lane protocol gameplay written`.

Event batch writes also log lane-specific debug context when a batch is written and drained.

Broader packet-budget work is planned separately. This document describes the current implementation, not the future realtime protocol or packet-budget design.

## Code map

### Primary implementation files

- `services/game-server/internal/networking/websocket.go` - Creates sessions, starts read/write/lifecycle goroutines, and runs the write loop.
- `services/game-server/internal/networking/websocket_write.go` - Owns the session write loop and ticker-driven outbound writes.
- `services/game-server/internal/networking/websocket_session.go` - Defines `webSocketSession` and the per-session outbound channel.
- `services/game-server/internal/networking/room_snapshot.go` - Builds and enqueues room snapshots.
- `services/game-server/internal/networking/room_error.go` - Builds and enqueues room error packets.
- `services/game-server/internal/networking/session_auth.go` - Builds and enqueues auth result packets.
- `services/game-server/internal/networking/player_pause_state.go` - Builds and enqueues player pause state packets.
- `services/game-server/internal/networking/inbound/telemetry.go` - Builds and queues telemetry pong responses.
- `services/game-server/internal/networking/outbound/server_message_writer.go` - Writes encoded server messages to the WebSocket.
- `services/game-server/internal/networking/outbound/debug_status_presentation.go` - Builds encoded debug status packets.
- `services/game-server/internal/networking/outbound/debug_shape_catalog_presentation.go` - Builds encoded debug shape catalog packets.

### Related source and generated files

- `services/game-server/internal/protocol/packetcodec/codec.go`
- `services/game-server/internal/game/packets.go`
- `services/game-server/internal/game/runtime/packets_generated.go`
- `services/game-server/internal/devtools/packets_generated.go`
- `shared/packets/outputs.toml`
- `shared/packets/gameplay.toml`
- `shared/packets/lobby.toml`
- `shared/packets/debug.toml`

### Important non-ownership boundaries

- `services/game-server/internal/rooms/` owns room state and room lifecycle rules.
- `services/game-server/internal/game/` owns authoritative simulation state and gameplay packet projection.
- `services/game-server/internal/devtools/` owns debug status and debug shape payload construction inputs.
- `services/game-server/internal/protocol/packetcodec/` owns JSON encode/decode mechanics.
- `services/game-server/internal/protocol/realtime/` owns realtime lane packet construction, scheduling, and metrics behavior.
- `services/game-server/internal/protocol/realtime/packets_generated.go` owns the generated realtime packet constants output.
- `services/game-server/internal/networking/packetmetrics/` owns lane packet metric record/log helpers.
- `docs/planning/protocol/realtime-protocol-architecture.md` owns future realtime protocol delivery policy planning.

## Tests and verification

The documented focused test paths for outbound routing are:

- `services/game-server/internal/networking/websocket_write_test.go`
- `services/game-server/internal/networking/outbound/debug_status_presentation_test.go`
- `services/game-server/internal/networking/outbound/debug_shape_catalog_presentation_test.go`
- `services/game-server/internal/networking/room_snapshot_test.go`
- `services/game-server/tests/networking/room_snapshot_test.go`
- `services/game-server/internal/networking/room_error_test.go`
- `services/game-server/internal/networking/session_auth_test.go`
- `services/game-server/tests/game/pause_test.go`
- `services/game-server/internal/networking/packetmetrics/*_test.go`
- `services/game-server/internal/protocol/realtime/*_test.go`

## Related docs

- [Game Server Networking](./!INDEX.md)
- [Game Server](../!INDEX.md)
- [Client Outbound Packet Sending](../../client/networking-flow/outbound-packet-sending.md)
- [Client Inbound Packet Routing](../../client/networking-flow/inbound-packet-routing.md)
- [Realtime WebSocket Protocol](../../../protocol/realtime-websocket-protocol.md)
- [Gameplay State Application](../../client/gameplay-runtime/gameplay-state-application.md)
- [Lane Packet Projection](../../simulation/runtime/lane-packet-projection.md)
- [Packet Schemas](../../../data/packet-schemas.md)
- [Protocol](../../../protocol/!INDEX.md)
- [Data](../../../data/!INDEX.md)
- [Realtime Protocol Architecture](../../../planning/protocol/realtime-protocol-architecture.md)
- [Network Observability And Packet Budget](../../../planning/domains/technical/network-observability-and-packet-budget.md)

## Notes

`websocket_write.go` owns the outbound write loop and lane packet delivery. Game-over lifecycle advancement still lives in `websocket_gameplay_tick.go`.

The current `debug_shape_catalog` send-once behavior is tracked by room ID inside the write loop, not by a durable client acknowledgement.

This document is scoped to current service implementation. Future transport mapping, compact encoding, quantization, and protobuf migration belong in protocol planning until implemented.



