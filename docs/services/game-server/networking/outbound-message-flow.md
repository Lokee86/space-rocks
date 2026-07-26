---
author: brian
created: "2026-07-19"
document_id: 019f7d55-fb2c-71cf-b081-e875da77284d
document_type: general
policy_exempt: false
summary: This document describes per-session outbound delivery for the game server, covering both queued WebSocket responses and active realtime WebRTC lane delivery.
---
# Outbound Packet Routing

Parent index: [Game Server Networking](./!INDEX.md)

## Purpose

This document describes per-session outbound delivery for the game server, covering both queued WebSocket responses and active realtime WebRTC lane delivery.

## Overview

Game-server outbound routing is the server-side send path for per-session outbound messages written to a connected client session.

The outbound boundary has four current responsibilities:

1. Queued one-off responses produced by request handlers.
2. Active realtime lane packet writes triggered by the realtime WebRTC send path.
3. Debug shape catalog writes when devtools are enabled.
4. Direct WebSocket `resync_required` acknowledgment writes from the write-loop resync branch.

Queued one-off responses use `session.outbound` and `outbound.WriteServerMessage()` over WebSocket.
Active realtime lane packets are built by `services/game-server/internal/protocol/realtime/`, encoded through `services/game-server/internal/protocol/packetcodec`, and sent over lane-specific WebRTC DataChannels through `session.webrtcTransport.SendEncodedLaneJSON()`.
Queued one-off response producers generally encode packet structs through `packetcodec` before enqueueing bytes, while active realtime lane packets are built and encoded by `services/game-server/internal/protocol/realtime/` before the WebRTC transport writes the encoded bytes.
Current physical gameplay channels include `sr.world`, `sr.ships`, `sr.asteroids`, `sr.bullets`, `sr.ships.lifecycle`, `sr.asteroids.lifecycle`, `sr.bullets.lifecycle`, `sr.overlay`, `sr.session`, and `sr.event`. `sr.world`, `sr.overlay`, `sr.session`, `sr.event`, `sr.ships.lifecycle`, `sr.asteroids.lifecycle`, and `sr.bullets.lifecycle` are ordered/reliable lanes, while `sr.ships`, `sr.asteroids`, and `sr.bullets` are unordered/unreliable hot-update lanes.

The networking layer owns connection/session write mechanics and message delivery. The realtime protocol package owns lane packet construction, baseline policy, candidate selection, quantization, event wire shaping, and wire-shape assembly. Outbound routing delivers already projected and quantized gameplay lane packets; it does not decide realtime packet schema policy or quantization policy. Projection and readable record building remain readable all the way through `WireLanePacket`.

The WebSocket write loop also drains the typed resync request channel. Each request carries its captured room and receiver context; stale context is discarded without writing or mutating state. For a current request, the loop writes the exact `resync_required` acknowledgment directly over WebSocket first. Only a successful write invalidates that lane's baseline readiness and projection. The next normal planner pass then emits the lane's full recovery candidate over the existing reliable WebRTC lane at the next sequence and a new sequence-backed baseline ID. Recovery control packets are not active lane candidates, and `resync_write_test.go` covers ordering, failure preservation, and stale-request rejection.

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
- WebRTC delivery for active realtime gameplay packets over ordered/reliable lanes for `sr.world`, `sr.overlay`, `sr.session`, `sr.event`, `sr.ships.lifecycle`, `sr.asteroids.lifecycle`, and `sr.bullets.lifecycle`, and unordered/unreliable hot-update lanes for `sr.ships`, `sr.asteroids`, and `sr.bullets`.
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

The server-to-client WebSocket packet path carries queued one-off responses and direct `resync_required` acknowledgment writes. Active gameplay lane packets use the WebRTC transport path.

The client consumes these messages after the Godot networking layer decodes WebSocket text and classifies packets by `type`.

The server owns authority behind the payloads. The client should treat outbound server packets as authoritative readback or authoritative request results, not as local decisions.

The current outbound payloads include queued one-off responses, direct WebSocket resync control packets, lane-native realtime packets, and debug packets.

WebRTC signaling is still WebSocket-owned in the current implementation. WebRTC DataChannels are now reusable JSON transport seams for active gameplay lanes, and webrtc_smoke remains a diagnostic packet on the WebRTC transport. The current WebRTC packet types are webrtc_offer, webrtc_answer, webrtc_ice_candidate, webrtc_ready, webrtc_smoke, and webrtc_failed. Active realtime gameplay lane packets are sent over lane-specific WebRTC DataChannels with no WebSocket fallback. Current physical gameplay channels include `sr.world`, `sr.ships`, `sr.asteroids`, `sr.bullets`, `sr.ships.lifecycle`, `sr.asteroids.lifecycle`, `sr.bullets.lifecycle`, `sr.overlay`, `sr.session`, and `sr.event`; `sr.ships`, `sr.asteroids`, and `sr.bullets` are unordered/unreliable hot-update lanes, and lower-sequence hot packets are rejected by client sequence guards, while same-sequence `ship_delta`, `asteroid_delta`, and `bullet_delta` chunks remain valid. Deployment must keep the advertised WebRTC ICE address and UDP path reachable from clients directly; a proxied HTTP WebSocket route does not carry the UDP data channel.

## Routing model

### Connection write loop

`handleConnection()` starts the connection runtime by creating a `webSocketSession`, starting `readClientInput()` in a goroutine, starting `tickSessionGameplayLifecycle()` in a goroutine, and running `writeServerMessages()` on the connection goroutine.

`writeServerMessages()` owns outbound delivery for the session. It selects between four inputs: read-loop close/error, queued outbound bytes from `session.outbound`, typed queued resync requests, and server tick events.

If the read loop reports a close or error, the write loop logs the read close and returns.

If a WebSocket write fails, `outbound.WriteServerMessage()` invokes the write-close logger and returns `false`. The write loop then returns and the connection teardown path runs.

### Session outbound queue

Each `webSocketSession` owns:

`outbound chan []byte`

The channel is created with a buffer size of 16 in `newWebSocketSession()`.

Queued responses are already encoded byte payloads. They are written by the `session.outbound` branch in `writeServerMessages()`, which passes the encoded bytes to `outbound.WriteServerMessage(session.conn, message, onWriteClose)`.

The queue is not durable. It is a small in-memory handoff between handlers and the write loop, with no retry or acknowledgement guarantee.

### Resync control write

Typed resync requests use a buffered request channel rather than `session.outbound`. Each envelope captures the room ID and receiver ID at enqueue time. The write loop ignores the request when that context is stale, then writes the exact `resync_required` control packet directly over WebSocket for a current request. Baseline readiness and projection are invalidated only after the write succeeds. A failed write exits the loop and preserves readiness and projection; the next normal server tick creates the recovery full. This is not a general packet-delivery acknowledgment system: `resync_required` acknowledges acceptance of one baseline recovery request only.

The control packets are WebSocket packets, not active realtime lane packets:

```text
client -> server: resync_request
server -> client: resync_required
```

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

- captured `SessionContext.GamePlayerID` is not empty
- captured `SessionContext.Room` is not nil
- captured `SessionContext.Room.GameplayContext()` is current and its match is valid

When eligible, the 60 Hz write loop calls `writeGameplayLaneProtocolMessage(session, remoteAddr)` on every eligible tick. Networking owns this invocation and the successful writes; protocol/realtime advances the independent per-session `HotLaneTick` and may suppress movement-only hot candidates according to their chunk-count cadence.

`writeGameplayLaneProtocolMessage()` currently:

1. Writes debug shape catalog output first when eligible.
2. Resets `session.realtimeState` when the receiver is empty or changes.
3. Calls `realtime.BuildActiveRealtimeResultForGame()`.
4. Advances `HotLaneTick` and applies the shared chunk-pressure cadence policy independently to ships, asteroids, and bullets: one chunk at 60 Hz, two chunks at 30 Hz, three chunks at 20 Hz, and four or more chunks at the 15 Hz floor. All chunks for an eligible sequence are sent in the same tick; additional pressure increases the parallel in-flight chunk burst rather than reducing cadence below 15 Hz.
5. Selects included lane candidates from the send plan, including lifecycle candidates before expanded hot chunks when needed.
6. The typed `RealtimeLanePayload` serializer builds the readable wire map after fail-closed payload, metadata, family, and wire-type validation.
7. Delta serializers in `realtime/wire_packets.go` omit empty delta sections from readable wire maps.
8. `CompactWirePacket` applies generated descriptor-driven aliases, value domains, ID codecs/selectors, record encodings, and event layouts.
9. `packetcodec` encodes each selected candidate into `EncodedLanePackets`.
10. Groups selected packets into independent transport transactions. `world`, `ships.lifecycle`, `asteroids.lifecycle`, and `bullets.lifecycle` form one reliable projection group; each hot, overlay, session, and event lane forms its own group.
11. Preflights each group against the buffered amount of its destination lane, including bytes reserved by earlier chunks in that group. A blocked group is skipped without suppressing unrelated groups.
12. `session.webrtcTransport.SendEncodedLaneJSON()` writes each encoded packet over the selected WebRTC lane channel when the group is eligible.
13. Logs lane wire packet details after successful writes.
14. Drains active event_batch events only after the event group succeeds.
15. Persists metadata and projections only for successful groups. Chunked hot projections commit only after the final chunk of the same-sequence burst succeeds.
16. Seeds independent ship, asteroid, and bullet movement projections after a successful world full; later reliable world commits synchronize entity membership without consuming deferred hot movement.
17. Marks a baseline lane ready after a successful final full packet.
18. Emits a non-empty per-tick debug summary after packet writes; cadence-suppressed or individually backpressure-skipped lanes produce no successful wire log for that lane.

`BuildActiveRealtimeResultForGame` obtains a receiver-scoped view over the current shared immutable presentation frame. Networking and `protocol/realtime` do not copy the complete entity maps per session. Receiver-specific pending events, realtime session state, baseline/delta state, candidate selection, encoding, and post-write persistence remain per session.

The lane packet construction path lives in `services/game-server/internal/protocol/realtime/`. The physical compact-wire contract lives in `shared/packets/realtime_wire.toml`, with generated reference data in `docs/protocol/generated/realtime-wire-reference.md`. Realtime runtime owns projection, sparse omission, generic descriptor application, scheduling, chunking, and encoded-byte accounting. Generated descriptors own physical aliases, value domains, ID codecs/selectors, record encodings, event layouts, quantization assignments, and decode compatibility alternatives. Networking owns successful WebRTC gameplay delivery, queued WebSocket delivery, event-batch drain-after-success behavior, post-write lane metadata persistence, and successful-write diagnostics. `packetcodec` owns JSON encoding only. `event_batch` remains one ordered batch of pending presentation events; known events use registered layouts, while unknown event maps remain compatibility pass-through records. Sparse delta omission does not implement record-level splitting or entity-level prioritization.

Hot ship/asteroid/bullet chunk construction uses conservative compact-JSON byte estimation before scheduling so the write path does not repeatedly JSON-encode trial chunks. The chunker is the hard-size guard for hot movement packets. Active encoding records the final encoded byte size for diagnostics and accounting, but it does not reject already-scheduled hot packets for size.

The networking layer owns successful WebRTC delivery for active realtime gameplay packets, successful WebSocket delivery for queued one-off packets, and the post-write session state changes that follow from those successful writes. `server_sent_msec` is not generated by the writer: it is captured as the server Unix-millisecond wall-clock time when the game publishes its immutable presentation frame, so all receiver snapshots using that frame carry the same timestamp. Networking does not own frame publication or timestamp creation; realtime lane projection carries the frame timestamp into encoded lane metadata.
Active lane metadata persistence, event drain, and baseline persistence happen only after a successful WebRTC write.

Chunk metadata exists in the wire shape and scheduler records. The current active path uses it for focused `ship_delta`, `asteroid_delta`, and `bullet_delta` hot-lane chunking. Oversized hot movement update lists are split into multiple real candidates before scheduling and encoding, then written as separate WebRTC messages on the same hot lane. This does not make networking the owner of general fragmentation or record/entity-level prioritization.

The 500 B scheduler target is not a total-per-tick send ceiling; aggregate encoded bytes may exceed it when multiple hot chunks or required packets are written.

### Independent lane-group preflight and commit

Reliability, ordering, buffered amount, metadata, and movement projections are tracked per physical lane group. A congested ship, asteroid, bullet, overlay, session, or event lane suppresses only that group for the current tick. Other eligible groups continue writing and commit their own state independently.

Reliable world projection changes and the three reliable entity lifecycle lanes remain one group because the world projection records the lifecycle membership boundary. If any lane in that reliable group cannot accept the complete group, none of its metadata or projections advance. Hot movement projections are separate from the world projection, so a successful reliable commit cannot consume movement that was deferred by cadence or hot-lane backpressure.

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

- captured `SessionContext.RoomID` is not empty
- the current room ID has not already received the shape catalog from this write loop
- `outbound.CanSendDebugShapeCatalog(SessionContext.Room)` returns true

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
- `shared/packets/webrtc.toml`
- `shared/packets/player_data.toml`
- `shared/packets/realtime_wire.toml`

Generated and runtime outputs include:

- `services/game-server/internal/protocol/realtime/packets_generated.go`
- `services/game-server/internal/protocol/realtimewire/generated.go`
- `services/game-server/internal/protocol/realtime/compact_wire_packet.go`
- `services/game-server/internal/protocol/realtime/compact_wire_descriptor.go`
- `services/game-server/internal/protocol/realtime/`
- `services/game-server/internal/game/packets.go`
- `services/game-server/internal/game/runtime/packets_generated.go`
- `services/game-server/internal/devtools/packets_generated.go`

### JSON packet codec

Server outbound encoding uses:

`services/game-server/internal/protocol/packetcodec/codec.go`

`packetcodec.Encode(packet)` currently wraps `json.Marshal(packet)`.

The outbound routing path does not own the packet schema or wire-format strategy. The current implementation is JSON text over WebSocket for queued one-off packets. For queued one-off packets, networking-side producers often encode packet structs before enqueueing. For active realtime lane packets, the realtime protocol package builds the wire map, applies sparse omission and generated descriptor-driven compact encoding, encodes JSON through `services/game-server/internal/protocol/packetcodec`, and sends encoded bytes to `session.webrtcTransport.SendEncodedLaneJSON()` for WebRTC delivery. Hot ship, asteroid, and bullet lanes remain supersedable. Ship creates/deletes and non-transform state updates are emitted on ordered/reliable `sr.ships.lifecycle`; asteroid and bullet creates/deletes are emitted on `sr.asteroids.lifecycle` and `sr.bullets.lifecycle`. `sr.world` no longer owns those active entity lifecycle/state records.

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

Active gameplay output is written as lane packet families, not as one combined gameplay output payload. The active encoding flow is `typed candidate payload -> payload wire serializer -> compact descriptor encoder -> JSON transport`.

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

Lane roles at service level are:

- world = ships, pickups, world/match presentation state, full/bootstrap snapshots
- asteroids.lifecycle = asteroid creates/deletes
- bullets.lifecycle = bullet/projectile creates/deletes
- asteroids = asteroid movement updates
- bullets = bullet/projectile movement updates
- overlay = receiver-specific overlay lane records
- session = session lane records for player/session/lifecycle and total asteroids
- event = event_batch presentation event delivery


ship_delta, asteroid_delta, and bullet_delta are high-priority hot-supersedable movement candidates; lifecycle creates/deletes use required/critical reliable lifecycle lanes.

Lane packet metadata always carries:

- `sequence`
- `server_sent_msec`

The lifecycle wire contract is explicit. `ships_lifecycle`, `asteroids_lifecycle`, and `bullets_lifecycle` emit:

- `lane`
- `sequence`
- `baseline_id`
- `snapshot_id`
- `snapshot_kind`
- `server_sent_msec`

For lifecycle candidates, `baseline_id` is inherited from the world baseline used to project the candidate. It is required so the client can gate lifecycle application on the matching world baseline. Each lifecycle lane has its own lane-local sequence. `world_full`, `ships_lifecycle`, `asteroids_lifecycle`, and `bullets_lifecycle` use the general hard-cap chunker; when split, lifecycle metadata conditionally carries `chunk_index`, `chunk_count`, and `is_final_chunk`. Lifecycle chunks are assembled atomically client-side before lifecycle gate validation and application. Reliable/ordered delivery on `sr.ships.lifecycle`, `sr.asteroids.lifecycle`, or `sr.bullets.lifecycle` orders messages only within that DataChannel; it does not establish ordering relative to `sr.world`.

Active runtime world/overlay/session lane packets may also carry inferred-or-conditional metadata when needed:

- `lane` when not inferred from `type`
- `baseline_id` when a numeric baseline dependency cannot represent the current value safely
- `baseline_sequence` for parseable runtime delta dependencies
- `snapshot_id` when not inferred from lane, packet kind, and sequence
- `snapshot_kind` when not inferred from `type`
- `chunk_index` and `chunk_count` when a lifecycle or other hard-capped candidate is split
- `is_final_chunk` is conditionally emitted for split lifecycle packets and may be emitted on chunked hot-lane packets when needed by the runtime or debug path; when absent, the client infers it from `chunk_index`/`chunk_count`.

Lifecycle packets retain explicit gate-relevant metadata, while split lifecycle candidates additionally carry conditional chunk metadata.

`event_batch` now uses compact envelope keys and sparse nested event records. It remains one ordered batch of pending presentation events, not one packet per event: projection preserves pending slice order, event IDs are identity/deduplication keys rather than sort keys, and successful-write drain preserves the relative order of events that remain pending. It does not use baselines, deltas, state snapshots, or chunking, and this section does not claim any future scheduler or transport behavior.
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

The ping/pong timestamps let the client estimate the offset between the server Unix-millisecond clock and its monotonic clock. That estimate is used client-side to interpret snapshot `server_sent_msec`; outbound routing only preserves and delivers the timestamp and does not calculate packet age.

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

World, overlay, and session quantization is fail-closed during candidate construction. The realtime projection package detects and returns quantization failures through planner assembly and active-result construction; a failed lane is not silently omitted while candidates for other lanes continue. Networking owns the existing `lane protocol gameplay build failed` log/boundary that receives the propagated build error. Invalid realtime payload, wire-map, compact, or JSON encoding failures likewise abort the active result build rather than being silently omitted, and `{}` is never sent. This fail-closed behavior is distinct from queued WebSocket behavior: queued messages remain an in-memory queue and their WebSocket write failures close the write loop.

Queued WebSocket write failures end the WebSocket write loop for that session. The connection teardown path closes the socket and leaves the disconnected room when needed.

WebRTC missing, not ready, or send failures prevent metadata advance and event drain for active gameplay packets. There is no WebSocket fallback for active gameplay.

The session outbound queue is not a durable delivery guarantee. It is a bounded in-memory handoff with capacity 16. Every queued WebSocket producer uses the session enqueue seam, which never blocks: when the queue is full, the server logs the slow-client condition once and closes that session's WebSocket. Normal connection teardown then removes the session from its room. This disconnect policy preserves correctness for control and room packets rather than silently dropping them, and prevents one slow client from blocking broadcast delivery to healthy sessions. The outbound channel remains open and is drained only by the writer loop; active gameplay WebRTC delivery is unchanged.

## Observability

Current outbound realtime debug logs are emitted only after successful gameplay lane writes.

Per-packet wire logs are the current source for packet family, lane, sequence, baseline or snapshot metadata, chunk metadata, and encoded byte size. The active per-packet debug message is `lane protocol gameplay wire packet written`, with useful fields such as:

Lifecycle lane writes may also appear here for `sr.ships.lifecycle`, `sr.asteroids.lifecycle`, and `sr.bullets.lifecycle`.

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

For chunked hot lanes and split lifecycle candidates, multiple `lane protocol gameplay wire packet written` entries may appear in the same tick, sharing a logical sequence with distinct `chunk_index`/`chunk_count` values. Lifecycle writes remain reliable/ordered and required; they are not hot-supersedable.

Per-tick summary logs are most useful when more than one gameplay packet is emitted in a tick. The active summary debug message is `lane protocol gameplay written`, with useful fields such as:

- `lane_packet_families`
- `baseline_full_count`
- `event_batch_written`
- `event_batch_drained_count`
- `packet_count`
- `encoded_bytes`

`packet_count` counts encoded packets actually written, not unique logical lanes. Under hot-lane stress it may include multiple `ship_delta`, `asteroid_delta`, or `bullet_delta` packets for the same lane in one tick.

Single-packet ticks may still produce one wire log and one summary log at debug level. No-op ticks are intentionally not logged.

`realtime lane metric` is not active runtime output. The current write path does not emit `packetmetrics.LogSentLaneMetrics(...)`, record or CRUD counters, or scheduler, prioritization, budget, deferred, or superseded fields as active log output.

Deeper packet-budget and scheduling work remains planning material elsewhere. This document describes the current service write path only. It acknowledges the current protocol/realtime candidate-level send plan, hot-supersedable world/overlay/asteroid/bullet deltas, and chunker-owned hot-lane hard-size guarding. Current runtime logs do not emit superseded-count fields, and this document does not claim record/entity-level prioritization or active cross-tick replay.

## Code map

### Primary implementation files

- `services/game-server/internal/networking/websocket.go` - Creates sessions, starts read/write/lifecycle goroutines, and runs the write loop.
- `services/game-server/internal/networking/websocket_write.go` - Owns the session write loop and ticker-driven outbound writes.
- `services/game-server/internal/networking/websocket_session.go` - Defines `webSocketSession` and the per-session outbound channel.
- `services/game-server/internal/networking/websocket_outbound_queue.go` - Owns bounded non-blocking enqueue and disconnect-on-overflow policy.
- `services/game-server/internal/networking/webrtc_transport.go` - Owns the session WebRTC transport seam used by active realtime gameplay delivery.
- `services/game-server/internal/networking/room_snapshot.go` - Builds and enqueues room snapshots.
- `services/game-server/internal/networking/room_error.go` - Builds and enqueues room error packets.
- `services/game-server/internal/networking/session_auth.go` - Builds and enqueues auth result packets.
- `services/game-server/internal/networking/player_pause_state.go` - Builds and enqueues player pause state packets.
- `services/game-server/internal/networking/inbound/telemetry.go` - Builds and queues telemetry pong responses.
- `services/game-server/internal/networking/outbound/server_message_writer.go` - Writes encoded server messages to the WebSocket.
- `services/game-server/internal/networking/outbound/debug_status_presentation.go` - Builds encoded debug status packets via Control/Controller projection.
- `services/game-server/internal/networking/outbound/debug_shape_catalog_presentation.go` - Builds encoded debug shape catalog packets.
- `services/game-server/internal/protocol/realtime/hot_lane_chunker.go` - expands oversized `ship_delta`, `asteroid_delta`, and `bullet_delta` hot movement candidates into bounded real candidate chunks before scheduling and encoding.
- `services/game-server/internal/protocol/realtime/realtime_hardcap_chunker.go` - expands oversized `world_full`, `ships_lifecycle`, `asteroids_lifecycle`, and `bullets_lifecycle` candidates into bounded real candidate chunks before scheduling and encoding.

### Related source and generated files

- `services/game-server/internal/game/presentation_snapshot.go`
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
- `services/game-server/internal/protocol/realtime/` owns realtime lane packet construction, send-plan records, sparse delta omission, generated compact descriptor application, encoded-byte accounting inputs, and metrics behavior.
- `services/game-server/internal/protocol/realtime/lanes.go`
- `services/game-server/internal/protocol/realtime/planner.go` - orchestrates lane candidate builder calls before scheduling and propagates candidate-construction errors.
- `services/game-server/internal/protocol/realtime/lane_candidate_world.go` - builds world lane candidates, returns world quantization failures, integrates hot movement splitting, and chains world projections.
- `services/game-server/internal/protocol/realtime/lane_candidate_lifecycle.go` - builds reliable asteroid and bullet lifecycle candidates.
- `services/game-server/internal/protocol/realtime/lane_candidate_overlay.go` - builds overlay lane candidates and returns overlay quantization failures.
- `services/game-server/internal/protocol/realtime/lane_candidate_session.go` - builds session lane candidates and returns session quantization failures.
- `services/game-server/internal/protocol/realtime/lane_candidate_event.go` - builds event_batch candidates without draining pending events.
- `services/game-server/internal/protocol/realtime/candidate_types.go` - realtime candidate/send-preparation types.
- `services/game-server/internal/protocol/realtime/candidate_policy.go` - delivery-class, priority, schedule-record, and projection helpers; packet-family identity is payload-owned.
- `services/game-server/internal/protocol/realtime/candidate_diagnostics.go` - write diagnostics for selected candidates.
- `services/game-server/internal/protocol/realtime/payload.go` - typed realtime lane payload contract and payload-owned packet-family identity.
- `services/game-server/internal/protocol/realtime/payload_validation.go` - supported concrete-value validation matrix and registry.
- `services/game-server/internal/protocol/realtime/payload_*.go` - lane-specific payload construction and serializer dispatch.
- `services/game-server/internal/protocol/realtime/wire_packets.go` projects readable lane maps.
- `services/game-server/internal/protocol/realtime/wire_reflect.go` owns generic readable-record reflection.
- `shared/packets/realtime_wire.toml` owns the physical compact-wire contract.
- `services/game-server/internal/protocol/realtimewire/generated.go` is generated descriptor data.
- `services/game-server/internal/protocol/realtime/compact_wire_packet.go` is the public compact encode boundary and generic recursive fallback.
- `services/game-server/internal/protocol/realtime/compact_wire_descriptor.go` applies generated bindings and encodings.
- `services/game-server/internal/protocol/realtime/hot_lane_size_estimate.go` - estimates compact hot-lane packet and tuple byte sizes for chunk construction without repeated trial JSON encoding.
- `services/game-server/internal/protocol/realtime/active.go` - active lane packet encoding path, `EncodedLanePackets` list construction, compact/packetcodec boundary, and encoded-byte accounting.
- `services/game-server/internal/protocol/realtime/scheduler.go` - include/defer planning for already-built candidates; real hot-lane chunks are created before scheduling.
- `services/game-server/internal/protocol/realtime/quantize/` - numeric quantization algorithms; generated descriptors own field-path policy assignments.
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
- `services/game-server/internal/networking/outbound_backpressure_test.go`
- `services/game-server/internal/networking/session_auth_test.go`
- `services/game-server/tests/game/pause_test.go`
- `services/game-server/internal/networking/packetmetrics/*_test.go`
- `services/game-server/internal/protocol/realtime/*_test.go`
- `services/game-server/internal/protocol/realtime/quantization_propagation_test.go` - world, overlay, and session quantization error propagation through exported planner and active-result boundaries.
- `services/game-server/internal/protocol/realtime/wire_packets_test.go` - lifecycle wire metadata coverage.
- `services/game-server/internal/game/presentation_snapshot_test.go` - snapshot-time Unix-millisecond timestamp coverage.

## Related docs

- [Game Server Networking](./!INDEX.md)
- [Game Server](../!INDEX.md)
- [Client Outbound Packet Sending](../../client/networking-flow/outbound-packet-sending.md)
- [Client Inbound Packet Routing](../../client/networking-flow/inbound-packet-routing.md)
- [Realtime WebSocket Protocol](../../../protocol/realtime-websocket-protocol.md)
- [Gameplay State Application](../../client/gameplay-runtime/gameplay-state-application.md)
- [Lane Packet Projection](../simulation/runtime/lane-packet-projection.md)
- [Packet Schemas](../../../data/packet-schemas.md)
- [Protocol](../../../protocol/!INDEX.md)
- [Data](../../../data/!INDEX.md)
- [Realtime Protocol Architecture](../../../planning/protocol/realtime-protocol-architecture.md)
- [Network Observability And Packet Budget](../../../planning/domains/technical/network-observability-and-packet-budget.md)
