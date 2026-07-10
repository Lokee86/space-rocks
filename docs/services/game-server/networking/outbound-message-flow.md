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
Active realtime lane packets are built by `services/game-server/internal/protocol/realtime/`, encoded through `services/game-server/internal/protocol/packetcodec`, and sent over lane-specific WebRTC DataChannels through `session.webrtcTransport.SendEncodedLaneJSON()`.
Queued one-off response producers generally encode packet structs through `packetcodec` before enqueueing bytes, while active realtime lane packets are built and encoded by `services/game-server/internal/protocol/realtime/` before the WebRTC transport writes the encoded bytes.
Current physical gameplay channels include `sr.world`, `sr.asteroids`, `sr.bullets`, `sr.asteroids.lifecycle`, `sr.bullets.lifecycle`, `sr.overlay`, `sr.session`, and `sr.event`. `sr.world`, `sr.overlay`, `sr.session`, `sr.event`, `sr.asteroids.lifecycle`, and `sr.bullets.lifecycle` are ordered/reliable lanes, while `sr.asteroids` and `sr.bullets` are unordered/unreliable hot-update lanes.

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
- WebRTC delivery for active realtime gameplay packets over ordered/reliable lanes for `sr.world`, `sr.overlay`, `sr.session`, `sr.event`, `sr.asteroids.lifecycle`, and `sr.bullets.lifecycle`, and unordered/unreliable hot-update lanes for `sr.asteroids` and `sr.bullets`.
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

WebRTC signaling is still WebSocket-owned in the current implementation. WebRTC DataChannels are now reusable JSON transport seams for active gameplay lanes, and webrtc_smoke remains a diagnostic packet on the WebRTC transport. The current WebRTC packet types are webrtc_offer, webrtc_answer, webrtc_ice_candidate, webrtc_ready, webrtc_smoke, and webrtc_failed. Active realtime gameplay lane packets are sent over lane-specific WebRTC DataChannels with no WebSocket fallback. Current physical gameplay channels include `sr.world`, `sr.asteroids`, `sr.bullets`, `sr.asteroids.lifecycle`, `sr.bullets.lifecycle`, `sr.overlay`, `sr.session`, and `sr.event`; `sr.asteroids` and `sr.bullets` are unordered/unreliable hot-update lanes, and lower-sequence hot packets are rejected by client sequence guards, while same-sequence `asteroid_delta` and `bullet_delta` chunks remain valid. Deployment must keep the advertised WebRTC ICE address and UDP path reachable from clients directly; a proxied HTTP WebSocket route does not carry the UDP data channel.

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
4. Selects included lane candidates from the send plan, including lifecycle candidates before expanded asteroid/bullet hot chunks when needed.
5. `WireLanePacket` builds readable long-key maps.
6. Delta serializers in `realtime/wire_packets.go` omit empty delta sections from readable wire maps.
7. `CompactWirePacket` applies final compact key/value aliasing, shared ID compaction, and tuple packing for asteroids, bullets, world ships/player records, session players, session lifecycle, and known event records.
8. `packetcodec` encodes each selected candidate into `EncodedLanePackets`.
9. `session.webrtcTransport.SendEncodedLaneJSON()` writes each encoded packet over the selected WebRTC lane channel when the transport is ready.
10. Logs lane wire packet details after successful writes.
11. Drains active event_batch events only after a successful WebRTC write.
12. Persists lane metadata only after successful writes.
13. Stores baseline projections for non-event lane packets only after successful writes.
14. Marks a lane baseline ready after a final full packet.
15. Emits a non-empty per-tick debug summary after packet writes.

The lane packet construction path lives in `services/game-server/internal/protocol/realtime/`. That package owns candidate selection, send-plan records, wire-map construction, sparse delta omission, compact alias preparation, hot asteroid/bullet movement splitting, tuple packing, encoded-byte accounting inputs, and helper metadata or types that support the write path. Realtime owns sparse delta omission, compact alias preparation, tuple packing, and sparse event wire shaping. Networking owns successful WebRTC gameplay lane delivery, successful queued WebSocket delivery, event_batch drain-after-success behavior, post-write lane metadata persistence, and the current successful-write debug logs. `packetcodec` owns JSON encoding only. Active realtime world, overlay, session, lifecycle, and `event_batch` lane packets are compacted at the final outbound encode boundary: `WireLanePacket` builds the readable map, `CompactWirePacket` applies aliases, compact values, shared ID compaction, and tuple packing before `packetcodec` encoding, and the alias contract lives in `docs/services/game-server/networking/realtime-compact-wire-mapping.md`. `event_batch` keeps one ordered batch of pending presentation events. Known event records are sparse and event-type-specific before compacting, and the compact wire path tuple-packs the known event records. Unsupported or future event records may remain map-shaped for compatibility until they are explicitly shaped for the compact path. The client expands compact event tuples back into readable dictionaries before event appliers consume them. It remains one ordered batch of pending presentation events, not one packet per event. Sparse delta omission reduces JSON shape overhead, but it does not implement record-level packet splitting or record/entity-level prioritization. Packet baselines, packet deltas, candidate-level scheduling, estimated byte-budget selection, and chunker-owned hot-lane hard-size guarding live in protocol/realtime. Active encoding records encoded byte sizes for diagnostics and accounting; it does not reject already-scheduled hot packets for size. Active lane packets are not ...

Hot asteroid/bullet chunk construction uses conservative compact-JSON byte estimation before scheduling so the write path does not repeatedly JSON-encode trial chunks. The chunker is the hard-size guard for hot movement packets. Active encoding records the final encoded byte size for diagnostics and accounting, but it does not reject already-scheduled hot packets for size.

The networking layer owns successful WebRTC delivery for active realtime gameplay packets, successful WebSocket delivery for queued one-off packets, and the post-write session state changes that follow from those successful writes.
Active lane metadata persistence, event drain, and baseline persistence happen only after a successful WebRTC write.

Chunk metadata exists in the wire shape and scheduler records. The current active path uses it for focused `asteroid_delta` and `bullet_delta` hot-lane chunking. Oversized hot movement update lists are split into multiple real candidates before scheduling and encoding, then written as separate WebRTC messages on the same hot lane. This does not make networking the owner of general fragmentation or record/entity-level prioritization.

The 500 B scheduler target is not a total-per-tick send ceiling; aggregate encoded bytes may exceed it when multiple hot chunks or required packets are written.

### Debug status

`debug_status` is built by `outbound.BuildDebugStatusResponse()` and covered by tests.

The builder requires a non-nil room, a non-nil game instance, `devtools.Enabled()`, and room state `InGame` or `GameOver`.

Current docs must not claim periodic `debug_status` delivery unless the active write loop calls the debug status builder. `debug_status` remains a WebSocket devtools readout packet when delivery is active, and it is not part of active gameplay lane output.

The packet is built with `game.NewControl(room.GameInstance())`, `devtools.NewController(...)`, `controller.StatusFor(playerID)`, and `controller.StatusesForAllPlayers()`, then encoded through `packetcodec`.

`controller.StatusesForAllPlayers()` uses `MatchDecision().Players` for all-player status membership.

Command fanout membership is separate from debug-status membership.

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

The outbound routing path does not own the packet schema or wire-format strategy. The current implementation is JSON text over WebSocket for queued one-off packets. For queued one-off packets, networking-side producers often encode packet structs before enqueueing. For active realtime lane packets, the realtime protocol package builds the wire map, applies sparse omission and compact aliases, encodes JSON through `services/game-server/internal/protocol/packetcodec`, and sends encoded bytes to `session.webrtcTransport.SendEncodedLaneJSON()` for WebRTC delivery. Hot asteroid and bullet lanes remain supersedable. Asteroid and bullet lifecycle creates/deletes are emitted on dedicated ordered/reliable lifecycle lanes: `sr.asteroids.lifecycle` and `sr.bullets.lifecycle`; `sr.world` no longer owns those active lifecycle creates/deletes.

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
- `asteroids_lifecycle`
- `bullets_lifecycle`
- `overlay_full`
- `overlay_delta`
- `session_full`
- `session_delta`
- `event_batch`
- `resync_request`
- `resync_required`

Lane roles at service level are:

- world = ships, pickups, world/match presentation state, full/bootstrap snapshots
- asteroids.lifecycle = asteroid creates/deletes
- bullets.lifecycle = bullet/projectile creates/deletes
- asteroids = asteroid movement updates
- bullets = bullet/projectile movement updates
- overlay = receiver-specific overlay lane records
- session = session lane records for player/session/lifecycle and total asteroids
- event = event_batch presentation event delivery
- resync = resync_request/resync_required recovery signaling

asteroid_delta and bullet_delta are high-priority hot-supersedable movement candidates; lifecycle creates/deletes use required/critical reliable lifecycle lanes.

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
- `is_final_chunk` is inferred by the client from `chunk_index`/`chunk_count` when absent and may be emitted on chunked hot-lane packets when needed by the runtime or debug path.

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

Lifecycle lane writes may also appear here for `sr.asteroids.lifecycle` and `sr.bullets.lifecycle`.

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

For chunked hot lanes, multiple `lane protocol gameplay wire packet written` entries may appear for `sr.asteroids` or `sr.bullets` in the same tick, sharing a hot-lane sequence with different `chunk_index`/`chunk_count` values. Lifecycle lanes are not chunked hot lanes.

Per-tick summary logs are most useful when more than one gameplay packet is emitted in a tick. The active summary debug message is `lane protocol gameplay written`, with useful fields such as:

- `lane_packet_families`
- `baseline_full_count`
- `event_batch_written`
- `event_batch_drained_count`
- `packet_count`
- `encoded_bytes`

`packet_count` counts encoded packets actually written, not unique logical lanes. Under hot-lane stress it may include multiple `asteroid_delta` or `bullet_delta` packets for the same lane in one tick.

Single-packet ticks may still produce one wire log and one summary log at debug level. No-op ticks are intentionally not logged.

`realtime lane metric` is not active runtime output. The current write path does not emit `packetmetrics.LogSentLaneMetrics(...)`, record or CRUD counters, or scheduler, prioritization, budget, deferred, or superseded fields as active log output.

Deeper packet-budget and scheduling work remains planning material elsewhere. This document describes the current service write path only. It acknowledges the current protocol/realtime candidate-level send plan, hot-supersedable world/overlay/asteroid/bullet deltas, and chunker-owned hot-lane hard-size guarding. Current runtime logs do not emit superseded-count fields, and this document does not claim record/entity-level prioritization or active cross-tick replay.

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
- `services/game-server/internal/networking/outbound/debug_status_presentation.go` - Builds encoded debug status packets via Control/Controller projection.
- `services/game-server/internal/networking/outbound/debug_shape_catalog_presentation.go` - Builds encoded debug shape catalog packets.
- `services/game-server/internal/protocol/realtime/hot_lane_chunker.go` - expands oversized `asteroid_delta` and `bullet_delta` hot movement candidates into bounded real candidate chunks before scheduling and encoding.

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
- `services/game-server/internal/protocol/realtime/lanes.go`
- `services/game-server/internal/protocol/realtime/planner.go` - orchestrates lane candidate builder calls before scheduling.
- `services/game-server/internal/protocol/realtime/lane_candidate_world.go` - builds world lane candidates, integrates hot movement splitting, and chains world projections.
- `services/game-server/internal/protocol/realtime/lane_candidate_lifecycle.go` - builds reliable asteroid and bullet lifecycle candidates.
- `services/game-server/internal/protocol/realtime/lane_candidate_overlay.go` - builds overlay lane candidates.
- `services/game-server/internal/protocol/realtime/lane_candidate_session.go` - builds session lane candidates.
- `services/game-server/internal/protocol/realtime/lane_candidate_event.go` - builds event_batch candidates without draining pending events.
- `services/game-server/internal/protocol/realtime/candidate_types.go` - realtime candidate/send-preparation types.
- `services/game-server/internal/protocol/realtime/candidate_policy.go` - packet-family, priority, delivery-class, schedule-record, and projection helpers.
- `services/game-server/internal/protocol/realtime/candidate_diagnostics.go` - write diagnostics for selected candidates.
- `services/game-server/internal/protocol/realtime/wire_packets.go`
- `services/game-server/internal/protocol/realtime/packets_generated.go` owns the generated realtime packet constants output.
- `services/game-server/internal/protocol/realtime/compact_wire_packet.go` owns the hand-authored compact alias runtime mapping at the encode boundary.
- `services/game-server/internal/protocol/realtime/compact_wire_ids.go` owns shared ID compaction for tuple-packed and compacted event records.
- `services/game-server/internal/protocol/realtime/compact_wire_asteroids.go`
- `services/game-server/internal/protocol/realtime/compact_wire_bullets.go`
- `services/game-server/internal/protocol/realtime/compact_wire_ships.go`
- `services/game-server/internal/protocol/realtime/compact_wire_players.go`
- `services/game-server/internal/protocol/realtime/compact_wire_events.go`
- `services/game-server/internal/protocol/realtime/hot_lane_size_estimate.go` - estimates compact hot-lane packet and tuple byte sizes for chunk construction without repeated trial JSON encoding.
- `services/game-server/internal/protocol/realtime/active.go` - active lane packet encoding path, `EncodedLanePackets` list construction, compact/packetcodec boundary, and encoded-byte accounting.
- `services/game-server/internal/protocol/realtime/scheduler.go` - include/defer planning for already-built candidates; real hot-lane chunks are created before scheduling.
- `services/game-server/internal/protocol/realtime/quantize/` - numeric wire quantization policies.
- `services/game-server/internal/protocol/realtime/quantize_world.go` - world lane quantization projection.
- `services/game-server/internal/protocol/realtime/quantize_overlay.go` - overlay full-packet wire quantization.
- `services/game-server/internal/protocol/realtime/quantize_session.go` - session full-packet wire quantization.
- `services/game-server/internal/networking/packetmetrics/` - packet observability helper types retained around outbound/realtime seams; not active `realtime lane metric` emission.
- [Realtime WebRTC Gameplay Transport](../../../protocol/realtime-webrtc-gameplay-transport.md)

## Tests and verification

The documented focused test paths for outbound routing are:

- `services/game-server/internal/networking/websocket_write_test.go`
- `services/game-server/internal/networking/outbound/debug_status_presentation_test.go`
- `services/game-server/internal/devtools/controller_status_test.go`
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


