# Outbound Packet Routing

Parent index: [Game Server Networking](./!INDEX.md)

## Purpose

This document describes per-session outbound delivery for the game server, covering both queued WebSocket responses and active realtime WebRTC lane delivery.

## Overview

Game-server outbound routing is the server-side send path for per-session outbound messages written to a connected client session.

The outbound boundary has three current responsibilities:

1. Queued one-off responses produced by request handlers.
2. Active realtime lane packet writes triggered by the realtime WebRTC send path.
3. Debug shape catalog writes when devtools are enabled.

Queued one-off responses use `session.outbound` and `outbound.WriteServerMessage()` over WebSocket.
Active realtime lane packets are encoded by `services/game-server/internal/protocol/realtime/packetcodec` and sent over lane-specific WebRTC DataChannels through `session.webrtcTransport.SendEncodedLaneJSON()`.
Queued one-off response producers generally encode packet structs through `packetcodec` before enqueueing bytes, while active realtime lane packets are built and encoded by `services/game-server/internal/protocol/realtime/` before the WebRTC transport writes the encoded bytes.
Current physical gameplay channels include `sr.world`, `sr.asteroids`, `sr.bullets`, `sr.overlay`, `sr.session`, and `sr.event`. They are reliable and ordered for now, and unreliable/unordered delivery remains future work.

The networking layer owns connection/session write mechanics and message delivery. The realtime protocol package owns lane packet construction, baseline policy, candidate selection, quantization, event wire shaping, and wire-shape assembly. Outbound routing delivers already projected and quantized gameplay lane packets; it does not decide realtime packet schema policy or quantization policy. Projection and readable record building remain readable all the way through `WireLanePacket`.

Realtime send-plan construction owns current candidate selection and byte-budget planning inputs; networking logs and writes the encoded results but does not decide record/entity-level prioritization.

## Code root

`services/game-server/internal/networking/`

The focused outbound helper package is:

`services/game-server/internal/networking/outbound/`

The realtime packet construction package is:

`services/game-server/internal/protocol/realtime/`

## Responsibilities

The game-server outbound packet routing path owns:

- WebSocket delivery for queued one-off responses.
- The session outbound queue used by one-off responses.
- WebRTC delivery for active realtime gameplay packets over reliable/ordered lane-specific gameplay DataChannels.
- One-time debug shape catalog writes per room connection context when devtools are enabled.
- Encoding queued one-off server packet structs through `packetcodec` where the queued producer owns that packet.
- Writing already-encoded active realtime lane packets produced by the realtime protocol package over WebRTC.
- Writing encoded packets through the active Gorilla WebSocket connection.
- Logging outbound encode failures and write closes.
- Realtime packet write diagnostics for successful gameplay lane sends.
- Invoking the realtime active-result send path from the websocket write loop and WebRTC transport path.
- Active lane metadata persistence, event drain, and baseline persistence only after successful WebRTC writes.

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

## Domain roles

The outbound routing surface is the server-to-client packet path.

The server-to-client WebSocket packet path is for queued one-off responses only. Active gameplay lane packets use the WebRTC transport path.

The client consumes these messages after the Godot networking layer decodes WebSocket text and classifies packets by `type`.

The server owns authority behind the payloads. The client should treat outbound server packets as authoritative readback or authoritative request results, not as local decisions.

The current outbound payloads include queued one-off responses plus lane-native realtime packets and debug packets.

WebRTC signaling is still WebSocket-owned in the current implementation. WebRTC DataChannels are now reusable JSON transport seams for active gameplay lanes, and webrtc_smoke remains a diagnostic packet on the WebRTC transport. The current WebRTC packet types are webrtc_offer, webrtc_answer, webrtc_ice_candidate, webrtc_ready, webrtc_smoke, and webrtc_failed. Active realtime gameplay lane packets are sent over lane-specific WebRTC DataChannels with no WebSocket fallback. Current physical gameplay channels include `sr.world`, `sr.asteroids`, `sr.bullets`, `sr.overlay`, `sr.session`, and `sr.event`; they are reliable and ordered for now. Deployment must keep the advertised WebRTC ICE address and UDP path reachable from clients directly; a proxied HTTP WebSocket route does not carry the UDP data channel.

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

Queued response producers generally encode with `packetcodec` before enqueue, and the queued packets are delivered from `session.outbound` through `outbound.WriteServerMessage()` over WebSocket.

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
5. `WireLanePacket` builds readable long-key maps.
6. Delta serializers in `realtime/wire_packets.go` omit empty delta sections from readable wire maps.
7. Raw-float assertion runs on relevant lane wire maps.
8. `CompactWirePacket` applies final compact key/value aliasing, shared ID compaction, and tuple packing for asteroids, bullets, world ships/player records, session players, session lifecycle, and known event records.
9. `packetcodec` encodes JSON.
10. `session.webrtcTransport.SendEncodedLaneJSON()` writes active realtime lane packets over the selected WebRTC lane channel when the transport is ready.
11. Logs lane wire packet details after successful writes.
12. Drains active event_batch events only after a successful WebRTC write.
13. Persists lane metadata only after successful writes.
14. Stores baseline projections for non-event lane packets only after successful writes.
15. Marks a lane baseline ready after a final full packet.
16. Emits a non-empty per-tick debug summary after packet writes.

The lane packet construction path lives in `services/game-server/internal/protocol/realtime/`. That package owns candidate selection, send-plan records, wire-map construction, sparse delta omission, compact alias preparation, hot asteroid/bullet movement splitting, tuple packing, encoded-byte accounting inputs, and helper metadata or types that support the write path. Realtime owns sparse delta omission, compact alias preparation, tuple packing, and sparse event wire shaping. Networking owns successful WebRTC gameplay lane delivery, successful queued WebSocket delivery, event_batch drain-after-success behavior, post-write lane metadata persistence, and the current successful-write debug logs. `packetcodec` owns JSON encoding only. Active realtime world, overlay, session, and `event_batch` lane packets are compacted at the final outbound encode boundary: `WireLanePacket` builds the readable map, `CompactWirePacket` applies aliases, compact values, shared ID compaction, and tuple packing after raw-float assertion and before `packetcodec` encoding, and the alias contract lives in `docs/services/game-server/networking/realtime-compact-wire-mapping.md`. `event_batch` keeps one ordered batch of pending presentation events. Known event records are sparse and event-type-specific before compacting, and the compact wire path tuple-packs the known event records. Unsupported or future event records may remain map-shaped for compatibility until they are explicitly shaped for the compact path. The client expands compact event tuples back into readable dictionaries before event appliers consume them. It remains one ordered batch of pending presentation events, not one packet per event. Sparse delta omission reduces JSON shape overhead, but it does not implement record-level packet splitting, record-level prioritization, packet baselines, packet deltas, or packet budget enforcement. Active lane packets are not handed to networking for WebSocket delivery.

The networking layer owns successful WebRTC delivery for active realtime gameplay packets, successful WebSocket delivery for queued one-off packets, and the post-write session state changes that follow from those successful writes.
Active lane metadata persistence, event drain, and baseline persistence happen only after a successful WebRTC write.

Chunk metadata exists in the wire shape and scheduler records, but this section does not claim full fragmentation or payload-splitting behavior beyond current final-chunk handling.

### Debug status

`debug_status` is built by `outbound.BuildDebugStatusResponse()` and covered by tests.

The builder requires a non-nil room, a non-nil game instance, `devtools.Enabled()`, and room state `InGame` or `GameOver`.

Current docs must not claim periodic `debug_status` delivery unless the active write loop calls the debug status builder. `debug_status` remains a WebSocket devtools readout packet when delivery is active, and it is not part of active gameplay lane output.

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
- `services/game-server/internal/protocol/realtime/compact_wire_ids.go`
- `services/game-server/internal/protocol/realtime/compact_wire_asteroids.go`
- `services/game-server/internal/game/packets.go`
- `services/game-server/internal/game/runtime/packets_generated.go`
- `services/game-server/internal/devtools/packets_generated.go`

### JSON packet codec

Server outbound encoding uses:

`services/game-server/internal/protocol/packetcodec/codec.go`

`packetcodec.Encode(packet)` currently wraps `json.Marshal(packet)`.

The outbound route does not own the packet schema or wire-format strategy. The current implementation is JSON text over WebSocket for queued one-off packets. For queued one-off packets, networking-side producers often encode packet structs before enqueueing. For active realtime lane packets, the realtime protocol package builds the wire map, applies sparse omission and compact aliases, encodes JSON through `services/game-server/internal/protocol/realtime/packetcodec`, and sends encoded bytes to `session.webrtcTransport.SendEncodedLaneJSON()` for WebRTC delivery.

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

Active gameplay output is written as lane packet families, not as one combined gameplay output payload.

Current lane families are:

- `world_full`
- `world_delta`
- `asteroid_delta`
- `bullet_delta`
- `overlay_full`
- `overlay_delta`
- `session_full`
- `session_delta`
- `event_batch`
- `resync_request`
- `resync_required`

Lane roles at service level are:

- world = authoritative world lane records for ships, pickups, and asteroid/bullet lifecycle creates/deletes
- asteroids = regular asteroid movement updates from `asteroid_delta`
- bullets = regular bullet movement updates from `bullet_delta`
- overlay = receiver-specific overlay lane records
- session = session lane records for player/session/lifecycle and total asteroids
- event = event_batch presentation event delivery
- resync = resync_request/resync_required recovery signaling

Lane packet metadata always carries:

- `sequence`
- `server_sent_msec`

Active runtime world/overlay/session lane packets may also carry inferred-or-conditional metadata when needed:

- `lane` when not inferred from `type`
- `baseline_id` when a numeric baseline dependency cannot represent the current value safely
- `baseline_sequence` for parseable runtime delta dependencies
- `snapshot_id` when not inferred from lane, packet kind, and sequence
- `snapshot_kind` when not inferred from `type`
- `chunk_index` and `chunk_count` when `chunk_count > 1`
- `is_final_chunk` only for legacy/backward-compatible decode support, not preferred active runtime output

`event_batch` now uses compact envelope keys and sparse nested event records. It remains one ordered batch of pending presentation events, not one packet per event. It does not use baselines, deltas, state snapshots, or chunking, and this section does not claim any future scheduler or transport behavior.
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

Debug status is devtools-only presentation. Its builder is eligible only when devtools are enabled and the session has an active game player in an `InGame` or `GameOver` room. Current docs must not claim active periodic delivery unless the write-loop call path exists in code.

### Debug shape catalog

Packet type:

`debug_shape_catalog`

Owned payload builder:

`outbound.BuildDebugShapeCatalogResponse()`

Debug shape catalog is devtools-only shape metadata. It is sent once per room ID in the current write-loop context when eligible.

## Failure behavior

Outbound encode failures are logged and the packet is dropped.

Queued WebSocket write failures end the WebSocket write loop for that session. The connection teardown path closes the socket and leaves the disconnected room when needed.

WebRTC missing, not ready, or send failures prevent metadata advance and event drain for active gameplay packets. There is no WebSocket fallback for active gameplay.

The session outbound queue is not a durable delivery guarantee. It is a bounded in-memory handoff. Senders that write into the queue can block when the buffer is full.

## Observability

Current outbound realtime debug logs are emitted only after successful gameplay lane writes.

Per-packet wire logs are the current source for packet family, lane, sequence, baseline or snapshot metadata, chunk metadata, and encoded byte size. The active per-packet debug message is `lane protocol gameplay wire packet written`, with useful fields such as:

- `packet_family`
- `candidate_lane`
- `candidate_kind`
- `sequence`
- `baseline_id`
- `snapshot_id`
- `snapshot_kind`
- `chunk_index`
- `chunk_count`
- `is_final_chunk`
- `encoded_bytes`

Per-tick summary logs are most useful when more than one gameplay packet is emitted in a tick. The active summary debug message is `lane protocol gameplay written`, with useful fields such as:

- `lane_packet_families`
- `baseline_full_count`
- `event_batch_written`
- `event_batch_drained_count`
- `packet_count`
- `encoded_bytes`

Single-packet ticks may still produce one wire log and one summary log at debug level. No-op ticks are intentionally not logged.

`realtime lane metric` is not active runtime output. The current write path does not emit `packetmetrics.LogSentLaneMetrics(...)`, record or CRUD counters, or scheduler, prioritization, budget, deferred, or superseded fields as active log output.

Broader packet-budget and scheduling work remains planning material elsewhere. This document describes the current service write path only, and does not claim packet budget enforcement, record-level prioritization, or cross-tick replay or supersession guarantees.

## Code map

### Primary implementation files

- `services/game-server/internal/networking/websocket.go` - Creates sessions, starts read/write/lifecycle goroutines, and runs the write loop.
- `services/game-server/internal/networking/websocket_write.go` - Owns the session write loop and ticker-driven outbound writes.
- `services/game-server/internal/networking/websocket_session.go` - Defines `webSocketSession` and the per-session outbound channel.
- `services/game-server/internal/networking/webrtc_transport.go` - Owns the session WebRTC transport seam used by active realtime gameplay delivery.
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
- `services/game-server/internal/game/` owns authoritative simulation state and lane-native realtime projection.
- `services/game-server/internal/devtools/` owns debug status and debug shape payload construction inputs.
- `services/game-server/internal/protocol/packetcodec/` owns JSON encode/decode mechanics.
- `services/game-server/internal/protocol/realtime/` owns realtime lane packet construction, send-plan records, sparse delta omission, compact alias preparation, encoded-byte accounting inputs, and metrics behavior.
- `services/game-server/internal/protocol/realtime/packets_generated.go` owns the generated realtime packet constants output.
- `services/game-server/internal/protocol/realtime/compact_wire_packet.go` owns the hand-authored compact alias runtime mapping at the encode boundary.
- `services/game-server/internal/protocol/realtime/compact_wire_ids.go` owns shared ID compaction for tuple-packed and compacted event records.
- `services/game-server/internal/protocol/realtime/compact_wire_bullets.go` owns bullet tuple packing.
- `services/game-server/internal/protocol/realtime/compact_wire_ships.go` owns world ship/player tuple packing.
- `services/game-server/internal/protocol/realtime/compact_wire_players.go` owns session player and lifecycle tuple packing.
- `services/game-server/internal/protocol/realtime/compact_wire_events.go` owns known event tuple packing.
- `services/game-server/internal/networking/packetmetrics/` - packet observability helper types retained around outbound/realtime seams; not active `realtime lane metric` emission.
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




