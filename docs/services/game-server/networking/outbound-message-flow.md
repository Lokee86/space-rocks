# Outbound Packet Routing

Parent index: [Game Server Networking](./!INDEX.md)

## Purpose

This document describes how the game server routes outbound realtime packets from server-owned state and server-side handlers to a connected WebSocket client.

## Overview

Game-server outbound packet routing is the server-side send path for realtime WebSocket messages.

The outbound route has two main lanes:

1. Queued one-off responses produced by request handlers.
2. Ticker-driven presentation packets produced by the WebSocket write loop.

Both lanes converge at `outbound.WriteServerMessage()`, which writes a WebSocket text message through the active Gorilla WebSocket connection.

The game server remains authoritative for gameplay, room state, auth result interpretation, and debug command effects. Outbound routing does not decide those rules. It encodes already-owned server facts into generated packet shapes and delivers them to the current session.

## Code root

`services/game-server/internal/networking/`

The focused outbound helper package is:

`services/game-server/internal/networking/outbound/`

## Responsibilities

The game-server outbound packet routing path owns:

- Per-session outbound message delivery over WebSocket text frames.
- The session outbound queue used by one-off responses.
- Ticker-driven gameplay state presentation writes.
- Ticker-driven debug status presentation writes when devtools are enabled.
- One-time debug shape catalog writes per room connection context when devtools are enabled.
- Encoding server packet structs through `packetcodec`.
- Logging outbound encode failures, write closes, large gameplay packets, and slow gameplay writes.
- Routing room broadcasts to each current room session by enqueuing per-session packets.

## Non-responsibilities

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
- Future realtime lane, delta snapshot, or binary protocol policy.

## Runtime surface

The outbound routing surface is the server-to-client WebSocket packet path.

The client consumes these messages after the Godot networking layer decodes WebSocket text and classifies packets by `type`.

The server owns authority behind the payloads. The client should treat outbound server packets as authoritative readback or authoritative request results, not as local decisions.

The current outbound payloads include:

- `state`
- `room_snapshot`
- `room_error`
- `authenticate_result`
- `player_pause_state`
- `telemetry_pong`
- `debug_status`
- `debug_shape_catalog`

## Routing model

### Connection write loop

`handleConnection()` starts the connection runtime:

1. Create a `webSocketSession`.
2. Start `readClientInput()` in a goroutine.
3. Start `tickSessionGameplayLifecycle()` in a goroutine.
4. Run `writeServerMessages()` on the connection goroutine.

`writeServerMessages()` owns outbound delivery for the session. It selects between:

- read-loop close errors
- queued outbound messages from `session.outbound`
- server tick events

If the read loop reports a close/error, the write loop logs the read close and returns.

If a WebSocket write fails, `outbound.WriteServerMessage()` invokes the write-close logger and returns `false`. The write loop then returns and the connection teardown path runs.

### Session outbound queue

Each `webSocketSession` owns:

`outbound chan []byte`

The channel is created with a buffer size of 16 in `newWebSocketSession()`.

Queued responses are already encoded byte payloads. They are written by this branch in `writeServerMessages()`:

    case message := <-session.outbound:
        outbound.WriteServerMessage(session.conn, message, onWriteClose)

The queue is not durable. It is a small in-memory handoff between handlers and the write loop.

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

### Ticker-driven gameplay state

`writeServerMessages()` runs a ticker at `constants.ServerTickRate`.

On each tick, gameplay presentation is eligible only when:

- `session.currentGamePlayerID` is not empty
- `outbound.CanSendGameplayPresentationState(session.room)` returns true

`CanSendGameplayPresentationState()` requires:

- a non-nil room
- a non-nil game instance
- room state `InGame` or `GameOver`

When eligible, the write loop builds a per-player `state` packet by calling:

`outbound.BuildGameplayPresentationStateResponse(session.room, session.currentGamePlayerID, session.currentRoomID, remoteAddr)`

That helper:

1. Gets the room game instance.
2. Calls `gameInstance.StatePacket(playerID)`.
3. Stamps `server_sent_msec`.
4. Encodes the packet through `packetcodec.Encode()`.
5. Logs a warning if the encoded gameplay packet is larger than 4KB.
6. Returns the encoded bytes to the write loop.

After writing the packet, the write loop logs a warning if the gameplay presentation write took longer than 20ms.

### Ticker-driven debug status

Debug status is sent from the same write loop, but less often than gameplay state.

The current cadence is every 8 eligible gameplay presentation ticks.

`writeDebugStatusMessage()` is eligible only when:

- `session.currentGamePlayerID` is not empty
- `outbound.CanSendDebugStatus(session.room)` returns true

`CanSendDebugStatus()` requires:

- a non-nil room
- a non-nil game instance
- `devtools.Enabled()`
- room state `InGame` or `GameOver`

The packet is built with `devtools.StatusFor()` and `devtools.StatusesForAllPlayers()`, then encoded through `packetcodec`.

### Ticker-driven debug shape catalog

`writeDebugShapeCatalogMessage()` sends a `debug_shape_catalog` packet at most once for the current room ID tracked by that write loop.

It is eligible only when:

- `session.currentRoomID` is not empty
- the current room ID has not already received the shape catalog from this write loop
- `outbound.CanSendDebugShapeCatalog(session.room)` returns true

`CanSendDebugShapeCatalog()` uses the same devtools and room/game-state gates as debug status.

The packet is built from `physics.LoadCollisionShapeCatalog()` and `devtools.BuildShapeCatalog()`, then encoded through `packetcodec`.

## Packet sources

### Generated packet structs

Most outbound packet structs are generated from shared packet definitions.

Source-of-truth files include:

- `shared/packets/outputs.toml`
- `shared/packets/gameplay.toml`
- `shared/packets/lobby.toml`
- `shared/packets/debug.toml`

Generated Go outputs include:

- `services/game-server/internal/game/packets.go`
- `services/game-server/internal/game/runtime/packets_generated.go`
- `services/game-server/internal/devtools/packets_generated.go`

### JSON packet codec

Server outbound encoding uses:

`services/game-server/internal/protocol/packetcodec/codec.go`

`packetcodec.Encode(packet)` currently wraps `json.Marshal(packet)`.

The outbound route does not own the packet schema or wire-format strategy. The current implementation is JSON text over WebSocket. Future realtime protocol planning may move delivery policy, packet lanes, delta snapshots, quantization, or binary encoding into protocol-specific seams.

## Packet families

### Gameplay state

Packet type:

`state`

Owned payload builder:

`outbound.BuildGameplayPresentationStateResponse()`

Gameplay state is per-player presentation state. The packet includes authoritative state such as active ships, sessions, lifecycle, bullets, asteroids, pickups, events, asteroid totals, and `server_sent_msec`.

The outbound helper does not mutate gameplay state. It projects the current game state for the target player and writes it.

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

Gameplay presentation packets currently emit two outbound diagnostics:

- A warning when encoded gameplay state is larger than 4KB.
- A warning when the gameplay state WebSocket write takes longer than 20ms.

The warning context includes room ID, player ID, and remote address.

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
- `services/game-server/internal/networking/outbound/gameplay_presentation.go` - Builds encoded gameplay state presentation packets.
- `services/game-server/internal/networking/outbound/gameplay_state_metrics.go` - Logs large gameplay packet and slow write warnings.
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
- `docs/planning/protocol/realtime-protocol-architecture.md` owns future realtime protocol delivery policy planning.

## Tests and verification

The documented focused test paths for outbound routing are:

- `services/game-server/internal/networking/outbound/gameplay_presentation_test.go`
- `services/game-server/internal/networking/outbound/debug_status_presentation_test.go`
- `services/game-server/internal/networking/outbound/debug_shape_catalog_presentation_test.go`
- `services/game-server/internal/networking/room_snapshot_test.go`
- `services/game-server/tests/networking/room_snapshot_test.go`
- `services/game-server/internal/networking/room_error_test.go`
- `services/game-server/internal/networking/session_auth_test.go`
- `services/game-server/tests/game/pause_test.go`

## Related docs

- [Game Server Networking](./!INDEX.md)
- [Game Server](../!INDEX.md)
- [Client Outbound Packet Sending](../../client/networking-flow/outbound-packet-sending.md)
- [Client Inbound Packet Routing](../../client/networking-flow/inbound-packet-routing.md)
- [Protocol](../../../protocol/!INDEX.md)
- [Data](../../../data/!INDEX.md)
- [Realtime Protocol Architecture](../../../planning/protocol/realtime-protocol-architecture.md)
- [Network Observability And Packet Budget](../../../planning/technical/network-observability-and-packet-budget.md)

## Notes

`websocket_write.go` only writes outbound presentation state and no longer advances game-over lifecycle. Current game-over lifecycle advancement lives in `websocket_gameplay_tick.go`.

The current `debug_shape_catalog` send-once behavior is tracked by room ID inside the write loop, not by a durable client acknowledgement.

This document is scoped to current service implementation. Future lane policy, packet prioritization, deltas, quantization, and protobuf migration belong in protocol planning until implemented.
