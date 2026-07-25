---
author: brian
created: "2026-07-19"
document_id: 019f7d55-fb2c-736b-a46d-7d11abfe6c04
document_type: general
policy_exempt: false
summary: This document describes the current realtime WebSocket protocol between the Godot client and the Go game server.
---
## Realtime WebSocket Protocol

Parent index: [Protocol](./!INDEX.md)

## Purpose

This document describes the current realtime WebSocket protocol between the Godot client and the Go game server.

The protocol is JSON-over-WebSocket for auth, room, lobby, signaling, and other pre-readiness/session-control packets, and JSON-over-lane-specific WebRTC DataChannels for active realtime gameplay and tooling packets: `sr.world`, `sr.overlay`, `sr.session`, `sr.event`, `sr.ships.lifecycle`, `sr.asteroids.lifecycle`, and `sr.bullets.lifecycle` are ordered/reliable, `sr.ships`, `sr.asteroids`, and `sr.bullets` are unordered/unreliable hot-update lanes, and the mandatory `sr.tooling` lane is reliable/ordered/bidirectional with negotiated id 9.

It covers the transport route, JSON packet framing, connection lifecycle, packet-family routing, lane policy, gameplay packet families, session-state requirements, delivery semantics, source-of-truth files, generated outputs, service responsibilities, compatibility expectations, and implementation code paths.

## Overview

The realtime protocol currently uses JSON text messages over a WebSocket connection for signaling and queued one-off packets, and lane-specific WebRTC physical DataChannels for active realtime gameplay packets, with ordered/reliable lanes for `sr.world`, `sr.overlay`, `sr.session`, `sr.event`, `sr.ships.lifecycle`, `sr.asteroids.lifecycle`, and `sr.bullets.lifecycle`, unordered/unreliable hot-update lanes for `sr.ships`, `sr.asteroids`, and `sr.bullets`, and a mandatory reliable/ordered/bidirectional `sr.tooling` transport lane at negotiated id 9.

The game server exposes one realtime route:

```text
GET /ws
```

The Godot client selects a WebSocket URL from the requested session mode, opens the connection, optionally sends an auth packet, sends room or gameplay request packets, and receives authoritative server packets. /ws is the signaling, session, and control route, not the active gameplay state transport.

WebSocket owns auth, room, lobby, signaling, and queued pre-readiness/session-control packets. WebRTC physical DataChannels own active realtime gameplay packets and runtime tooling packets. Current physical gameplay DataChannels are `sr.world`, `sr.overlay`, `sr.session`, `sr.event`, `sr.ships`, `sr.asteroids`, `sr.bullets`, `sr.ships.lifecycle`, `sr.asteroids.lifecycle`, and `sr.bullets.lifecycle`; `sr.tooling` retains negotiated id 9 and is required with all ten gameplay channels and already carries runtime measurement. `control` is a logical recovery category, not a current physical WebRTC gameplay DataChannel. The current generated recovery packet families are `resync_request` and `resync_required`; there is no generated packet family named `control`, and there is no physical `sr.control` channel in the current implementation. The WebSocket URL may point at the normal hosted or proxied service route, but WebRTC DataChannel connectivity is established by ICE candidates rather than by a WebRTC URL. Deployment must allow the advertised WebRTC ICE address to reach the game server directly. Cloudflare-proxied HTTP routes should not be assumed to carry WebRTC DataChannel traffic.

The route path does not define play mode. Local single-player and multiplayer currently use the same local WebSocket route during development. Single-player versus multiplayer behavior is expressed through packets, session identity, room state, admission policy, and player-data routing.

The WebSocket connection itself is only transport readiness for WebSocket-owned packet families. It does not imply:

```text
authenticated account identity
room membership
ready state
active gameplay player state
durable Local Profile identity
durable account identity
```

The server owns authority behind accepted room, gameplay, auth-result, telemetry, and devtools consequences. The client owns connection initiation, packet emission, inbound packet classification, realtime packet application through RealtimePacketPipeline, presentation routing, and WebRTC transport consumption.


Active realtime gameplay packets are not WebSocket-owned anymore. The current server path builds the lane packet, encodes it to JSON, and sends it over the matching session WebRTC gameplay DataChannel when WebRTC is ready. WebSocket still owns auth, room/lobby lifecycle, room snapshots, and WebRTC signaling. There is no WebSocket fallback for active realtime gameplay packets.
The protocol is best-effort and session-scoped, but active gameplay output now uses lane-native packet families and lane policy over WebRTC DataChannels. The current gameplay output lanes are:

```text
world
asteroids
bullets
overlay
session
event
asteroids.lifecycle
bullets.lifecycle
```

`sr.tooling` is not an active gameplay packet lane. Its negotiated channel is mandatory for readiness, and tooling packets are separated before normal gameplay routing. Runtime measurement request/response traffic and runtime devtools command requests are implemented on this channel. Devtools command payloads carry `request_id` and `trace_id`; `services/game-server/internal/networking/tooling` applies packet-policy, room, and capability preflight, decodes `DebugCommand`, dispatches through the existing devtools controller, and returns correlated `tooling_command_result` or `tooling_error` packets. Telemetry routing exists but remains partially wired. Developer readouts now use `sr.tooling`; the legacy continuous telemetry/ping path remains to migrate.

`control` is a logical recovery category, not a current physical WebRTC gameplay DataChannel. The current generated recovery packet families are `resync_request` and `resync_required`. There is no separate generated packet family named `control`, and there is no physical `sr.control` channel in the current implementation.

The active gameplay packet families are:

```text
world_full
world_delta
asteroid_delta
bullet_delta
asteroids_lifecycle
bullets_lifecycle
overlay_full
overlay_delta
session_full
session_delta
event_batch
```

These packets carry current lane snapshots, baseline updates, and event batches instead of one combined lane gameplay output payload. `resync_request` and `resync_required` are WebSocket control packets, not active WebRTC lane outputs or `RealtimeLaneCandidate`s.

Every server-built realtime packet carries the authoritative `match_id` metadata (compact wire key `mid`). A `RealtimeSessionState` is bounded by the identity tuple `(ReceiverID, MatchID)`; changing either component starts clean lane, baseline, projection, and hot-lane cohort state.

WebSocket room snapshots and WebRTC gameplay packets have no cross-transport ordering guarantee. The client pipeline expands compact `mid` to readable `match_id` and validates recognized packet types before applying its match boundary. Before an authoritative active match exists, recognized packets with a non-empty `match_id` are buffered by match and do not mutate lane or presentation state; missing IDs are rejected. When the authoritative `InGame` room snapshot activates `begin_realtime_match(match_id)`, the pipeline clears all pending match buckets and replays only that match's packets through normal lane routing. Pending unrelated matches are discarded, and once a match is active, mismatched IDs are rejected. `GameOver` retains the active match until Lobby/session teardown; `end_match`, reset, and connection teardown clear pending packets and protocol state. This applies to all active packet families, including lifecycle and event packets.

## Canonical baseline recovery loop

`BaselineTracker` owns mismatch detection for the world, overlay, and session baselines. A delta with no usable baseline is rejected with `missing_baseline`; a delta with a different baseline is rejected with `wrong_baseline`. Ordinary stale or duplicate sequence rejection is silent. Only the first transition into pending recovery emits a request; repeated rejected deltas are deduplicated.

The client request contains `type`, `match_id`, `lane`, the last active `baseline_id` (or an empty string), the last accepted `sequence`, and `reason`. The queue envelope captures the room ID, receiver ID, and match ID alongside the request. Its path is `BaselineTracker -> RealtimeRouter -> RealtimePacketPipeline -> ClientConnectionService -> ClientPacketSender -> generated resync_request_packet -> NetworkClient -> WebSocket`. If WebSocket is closed, the request is not queued or automatically retried. The server write loop discards requests with missing, mismatched, or stale room, receiver, or match identities before writing or invalidating realtime state; a valid `resync_required` response carries the validated current `match_id`.

The server accepts only world, overlay, and session requests when the session has an active room, non-nil game instance, and active game player. The read path validates and queues a typed request without mutating `RealtimeSessionState`. Its queue envelope captures room ID, receiver/player ID, and match ID; the write loop discards requests whose context is stale, including requests from a previous match. It writes `resync_required` over WebSocket first, including the validated current match ID, then invalidates only the requested lane's readiness and projection after a successful write. Lane metadata and sequence are preserved and other lanes are untouched.

The next normal planner pass emits a full only for the invalidated lane at previous sequence plus one, with a new sequence-backed baseline ID. That full is a normal required world, overlay, or session candidate sent on the existing ordered/reliable WebRTC lane. Changed-baseline full chunks are accepted only while recovery is pending, with stable baseline ID, sequence, and chunk count, contiguous indices, and correct finality. Readiness remains false until the final chunk; the final accepted full clears tracker and `ResyncState` pending state and restores lane/gameplay readiness when all required lanes are synced.

WebSocket acknowledgment and WebRTC recovery full have no shared delivery order. `resync_required` is acknowledgment-only: it updates reason state only while the lane remains pending and is ignored after recovery, preventing delayed acknowledgment regression. Inbound `resync_request` remains a compatibility client route; the current server sends `resync_required`, not `resync_request`. Baseline recovery covers only world, overlay, and session baselines, not hot asteroid/bullet or lifecycle lanes, and does not add resend, retry, session resume, durable queues, or outbound client queueing.

## WebRTC transport recovery

The WebRTC transport foundation is implemented. An unexpected close of a previously-open required channel preserves the WebSocket connection, session, room membership, and game context. The client replaces only the WebRTC peer, negotiates a fresh offer, and uses a 10-second recovery deadline. Successful replacement readiness preserves the active match ID and requests fresh world, overlay, and session baselines through the existing recovery path. Recovery failure or timeout disables only single-player realtime replay; it does not reset multiplayer/session state. Runtime measurement already uses the recovered `sr.tooling` channel; subscription consumers must explicitly resubscribe after successful recovery.

## Participating systems

```text
client/scripts/networking/
```

Owns the client WebSocket peer, polling, raw send/receive, packet encode/decode handoff, inbound packet dispatch, outbound send wrappers, collaborator composition, connection signals, and cached WebSocket auth result state. `ClientInboundCoordinator` owns dispatcher-consumer bindings for realtime packet families.

```text
client/scripts/boot/
client/scripts/session/
```

Owns session mode selection, pending boot request timing, WebSocket URL selection, auth-gated multiplayer boot dispatch, and routing from connection-service callbacks into room and gameplay controllers.

```text
client/scripts/protocol/realtime/
```

Owns client realtime lane metadata tracking, baseline and readiness tracking, lane packet application, and presentation adapter handoff.

```text
services/game-server/internal/networking/
```

Owns the server WebSocket upgrade, session object, read loop, write loop, inbound packet-family routing, outbound queue, room-session adapter wiring, auth session state, telemetry pong handling, room request routing, lane gameplay packet routing, WebRTC signaling, and disconnect cleanup.

```text
services/game-server/internal/protocol/packetcodec/
```

Owns the server JSON encode/decode wrapper used by networking.

```text
services/game-server/internal/protocol/realtime/
```

Owns lane projection, baseline and delta planning, candidate selection, hot asteroid/bullet movement splitting, send-plan records, numeric wire quantization, sparse delta omission, compact alias preparation, active lane encode-boundary preparation, and active/shadow/parity support.

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

Owns authoritative gameplay simulation, input handling, respawn handling, pause state, targeting, lane state projection, gameplay events, scoring, lives, death, and match-over facts.

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
future packet-encoding or transport-format planning
```

Packet schema owns packet type strings, generated constants, generated struct fields where applicable, and canonical field names. Runtime realtime protocol code owns active lane wire-map emission behavior such as sparse delta omission, numeric wire quantization, and compact alias application. Runtime services own packet consequences and meaning.

Entity lifecycle ownership is split by entity family. The world lane owns pickup, world, and full/bootstrap presentation state. Ship lifecycle packets use `sr.ships.lifecycle`. Asteroid lifecycle packets use `sr.asteroids.lifecycle`. Bullet/projectile lifecycle packets use `sr.bullets.lifecycle`. Hot ship, asteroid, and bullet lanes are unreliable movement/update lanes only and must not create entities implicitly.

For example:

```text
shared/packets/gameplay.toml
-> defines lane packet families and packet shape

services/game-server/internal/protocol/realtime/wire_packets.go
-> decides which active delta sections are emitted after delta planning

services/game-server/internal/protocol/realtime/compact_wire_packet.go
-> aliases remaining emitted keys at the encode boundary

services/game-server/internal/protocol/realtime/active.go
-> applies compact aliasing, packetcodec JSON encoding, and encoded-byte accounting for active lane packets. It does not reject already-scheduled hot packets because of encoded size.

services/game-server/internal/protocol/realtime/quantize/
-> owns numeric wire quantization policies

client outbound flow
-> sends input intent

game-server networking
-> routes input packet

game-server game simulation
-> decides authoritative movement, firing, collision, score, death, and resulting lane outputs
```

## Wire surface

### Endpoint

The game-server process registers:

```text
GET /ws
```

The route is handled by the networking WebSocket handler and upgraded with Gorilla WebSocket.

### Origin policy

The server rejects an absent or empty `Origin` header. The handler builds its policy once. When `SPACE_ROCKS_WEBSOCKET_ALLOWED_ORIGINS` is unset, the exact default allowlist is:

```text
https://space-rocks-client.local
http://localhost:8080
http://127.0.0.1:8080
http://[::1]:8080
```

When set, `SPACE_ROCKS_WEBSOCKET_ALLOWED_ORIGINS` is a comma-separated replacement allowlist; whitespace is trimmed and empty entries are ignored. Origins are matched exactly.

The Godot client currently sets the WebSocket handshake origin from generated constants:

```text
Constants.MULTIPLAYER_WS_ORIGIN
```

Origin rejection or upgrade failure prevents session creation.

### Message framing

Each WebSocket message is a text message containing one JSON object. The server enforces a 256 KiB inbound WebSocket text-message limit and sets a 10-second deadline before every outbound WebSocket write. The process HTTP server uses `:8080` with `ReadHeaderTimeout=5s`, `ReadTimeout=15s`, `WriteTimeout=15s`, and `IdleTimeout=60s`.

The packet envelope uses:

```json
{
  "type": "packet_type"
}
```

Many packet types also include additional top-level fields or nested objects.

Client-side WebSocket packet decode requires:

```text
JSON parses successfully
decoded value is a Dictionary
type exists
type is a String
type is not empty after trimming
payload, when present, is a Dictionary
```

Compact packets may arrive with `t` instead of `type`. Client decode expands compact aliases before applying normal envelope validation, and packets with neither `type` nor compact `t` still fail validation. Readable/logical world records are still maps keyed by `id`. Compact active wire output now tuple-packs asteroid, bullet, ship/player, session player/lifecycle, and known event records on the final wire shape. The client expands compact tuples back into readable dictionaries before lane state appliers run. Compact tuple IDs follow a three-way rule: bare numeric suffix when tuple context determines the prefix, tagged compact ID when the prefix is known but not tuple-determined, and the original string when the prefix is unknown or the suffix is malformed. Compact asteroid tuple IDs are numeric suffixes and the client rehydrates them to `asteroid-<id>` before world lane appliers run. Compact bullet IDs are numeric suffixes and the client rehydrates them to `bullet-<id>` before world lane appliers run. Compact player IDs are numeric suffixes and the client rehydrates them to `player-<id>` before ship, session, lifecycle, and tuple-expanded event appliers run. Compact event batches use tuple event records with compact presentation-event and player IDs where applicable. Compact deletes use numeric suffix IDs for the specific tuple fields that currently own that context. Readable/logical deletes remain full string IDs where the tuple contract does not apply. After client expansion, readable/logical packets still use full string IDs.

Server-side initial envelope decode unmarshals the `type` field before routing. Invalid JSON or an envelope decode failure logs a warning and skips the message. A valid JSON object with an unknown or empty `type` does not produce an explicit protocol response in the current server path.

WebRTC inbound delivery uses WebRTCTransport for DataChannel text receive, PacketCodec.decode for compact alias expansion and envelope decoding, ServerPacketDispatcher for typed dispatch, ClientInboundCoordinator for dispatcher-consumer bindings and branch selection, RealtimePacketPipeline for gameplay packet validation and application ownership, and RealtimeRouter for lane-specific mutation beneath the pipeline.

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

Active realtime gameplay lane packets use compact JSON encoding at the final outbound boundary. The physical contract is owned by `shared/packets/realtime_wire.toml`; generated tables are in [Realtime Wire Contract](./generated/realtime-wire-reference.md), and [Realtime Compact Wire Mapping](../services/game-server/networking/realtime-compact-wire-mapping.md) explains the architecture. The Godot client applies generated descriptors to normalize compact packets to readable dictionaries before typed routing and lane appliers. Current compatibility policy also accepts readable long-key packets.

### Lane metadata

Lane packets always carry these top-level metadata fields:

```text
type
sequence
server_sent_msec
match_id (compact alias: mid)
```

For active world, ship, asteroid, bullet, overlay, and session runtime packets where the fields are present, the client infers additional metadata when the server omits redundant fields. This inference rule does not apply to `ships_lifecycle`, `asteroids_lifecycle`, or `bullets_lifecycle` packets; lifecycle metadata is explicit because the gate must validate the packet's lane and world-baseline dependency before application:

```text
lane
  inferred from type

snapshot_kind
  inferred from type: _full -> full, _delta -> delta

snapshot_id
  inferred from lane + packet kind + sequence

full baseline_id
  inferred as <lane>-baseline-<sequence>

delta baseline_id
  inferred from baseline_sequence when present

chunk_index / chunk_count
  emitted only when chunk_count > 1

is_final_chunk
  inferred from chunk_index and chunk_count
```

Lifecycle packets explicitly carry all gate-relevant metadata:

```text
lane
sequence
baseline_id
snapshot_id
snapshot_kind
server_sent_msec
match_id (compact alias: mid)
```

Lifecycle `baseline_id` is the world baseline identity inherited from the world projection used to build the lifecycle candidate. Lifecycle lanes do not own independent full baselines and do not participate independently in gameplay readiness. Their compact forms use `l`, `q`, `b`, `sid`, `k`, and `ms`; they are not grouped with runtime families whose lane or baseline metadata may be inferred.

Legacy long-key fields and older compact aliases remain accepted during decode for backward-compatible packets, but they are no longer the preferred active runtime output for world, asteroid, bullet, overlay, and session gameplay lanes. `event_batch` runtime metadata is explicit: the readable logical wire map keeps `type`, `sequence`, `server_sent_msec`, `match_id`, `batch_id`, and `events`, while compact runtime output uses `t`, `q`, `ms`, `mid`, `bid`, and `ev`. `event_batch` does not emit `lane`, `baseline_id`, `snapshot_id`, `snapshot_kind`, `chunk_index`, `chunk_count`, or `is_final_chunk` in preferred runtime output, and it remains excluded from baseline/delta/chunk metadata. Control-lane resync packets keep their own current metadata behavior.

The readable server projection comes from `wire_packets.go`, `wire_reflect.go`, and the realtime projection records. The physical compact shape comes from `shared/packets/realtime_wire.toml` and generated descriptors consumed by `compact_wire_packet.go` and `compact_wire_descriptor.go`. Numeric quantization algorithms remain runtime code, while field-path policy assignments come from generated descriptors before compact encoding and `packetcodec` JSON encoding.

Chunk metadata exists in the wire shape and scheduler records. The roughly 1,200 B `HardCapBytes` construction limit covers `world_full`, `ships_lifecycle`, `asteroids_lifecycle`, `bullets_lifecycle`, `ship_delta`, `asteroid_delta`, and `bullet_delta` candidates. Server expansion exact-encodes compact payloads while constructing chunks, preserves logical identity metadata, and splits oversized full/lifecycle or hot candidates before scheduling and encoding. A record that cannot fit individually returns an explicit construction error. This is candidate chunking for the active state families, not arbitrary fragmentation of every lane family.

### Hot movement cadence

Each eligible active build advances an independent per-session 60 Hz `HotLaneTick`. Asteroid movement emits at 60 Hz when unchunked and 30 Hz when chunking is required; bullet movement emits at 60 Hz for one chunk, 30 Hz for two chunks, and 20 Hz for three or more chunks. Forced sends bypass cadence suppression. Sequence numbers advance only for successfully written candidates, while `HotLaneTick` advances on every eligible active build, so skipped sends cannot freeze cadence. This is candidate policy in protocol/realtime; it does not change DataChannel reliability or hot-lane chunk metadata semantics.

### Numeric wire quantization

State-lane records are quantized in the realtime projection and wire-record path before delta comparison and JSON encoding, so deltas compare projected wire-shaped values instead of raw simulation float precision. `event_batch` is not a state lane and does not participate in delta comparison.

Presentation-event records are quantized during explicit event wire shaping before JSON encoding. The domain/game logs may still show raw simulation floats, but those raw logs are not the wire contract.

This currently applies to the server-owned outbound lane state families:

```text
world_full / world_delta
asteroid_delta
bullet_delta
overlay_full / overlay_delta
session_full / session_delta
```

`event_batch` is not a field-delta state lane.

Event wire `x` and `y` values use the same position-scale policy family as realtime entity positions.
`ship_death` `respawn_delay` is quantized as a seconds/duration-style field.
IDs, type names, lives, damage amounts, health, shield, score, and already-integer values are not numeric-quantized.

Example:

```text
simulation x/y = 512.75, 384.25
wire x/y = 512, 384

simulation respawn_delay = 2.75
wire respawn_delay = 3
```

Known float-like fields use lane- and field-specific policies from `services/game-server/internal/protocol/realtime/quantize/`. The active server code paths include:

- `services/game-server/internal/protocol/realtime/quantize/`
- `services/game-server/internal/protocol/realtime/quantized_records.go`
- `services/game-server/internal/protocol/realtime/quantize_world.go`
- `services/game-server/internal/protocol/realtime/quantize_overlay.go`
- `services/game-server/internal/protocol/realtime/quantize_session.go`
- `services/game-server/internal/protocol/realtime/planner.go`
- `services/game-server/internal/protocol/realtime/wire_packets.go`

World, overlay, and session quantization is fail-closed during candidate construction. Their candidate builders return quantization errors, `AssembleRealtimeLaneCandidates` returns `(RealtimeLanePlan, error)`, and active-result construction propagates the same failure to networking's existing `lane protocol gameplay build failed` boundary. A lane that fails quantization is not converted into an empty candidate list or silently omitted while other lanes continue. Realtime projection owns detecting and returning the error; networking owns logging and handling the build-failure boundary.

Client numeric decode is exact-path only. `RealtimeWireGenerated.QUANTIZATION_POLICY_BY_PATH` determines which numeric values are dequantized; unregistered integers and floats are preserved unchanged. Named `float_generic` policies still exist for explicitly registered paths such as rotations and scales, but unmapped floats do not fall back to that policy.

This is still JSON over WebSocket for auth, room, lobby, telemetry, and signaling packets, but active realtime gameplay packets now travel over lane-specific WebRTC DataChannels. WebRTCTransport uses repeated one-packet-per-lane round-robin passes, services reliable lifecycle lanes before the general lane group, rotates group start positions between polls, and preserves the existing `MAX_PACKETS_PER_POLL = 48` total and `MAX_PACKETS_PER_LANE_PER_POLL = 12` per-lane limits. WebRTCTransport does not coalesce bullet_delta packets; packets remain queued and receive pacing is not packet dropping. The current implementation does not have binary packet encoding, compression, protobuf encoding, schema negotiation, or version negotiation.

### Field-delta update semantics

Delta lane update arrays are field-delta aware.

Current delta lane record groups are described here as readable logical shapes before compact aliasing:

```text
creates
= full typed records

updates
= partial field maps containing the identity key plus changed fields only

deletes
= identity string lists
```

Current update identity keys are:

```text
world ship_updates = id
world pickup_updates = id
asteroids_lifecycle asteroid_creates/asteroid_deletes = id
bullets_lifecycle bullet_creates/bullet_deletes = id
asteroid_delta asteroid_updates = id
bullet_delta bullet_updates = id
world bullet_updates and world asteroid_updates = id for legacy/bootstrap/resync-safe compatibility only; regular active movement updates are split into dedicated hot movement lanes

overlay receiver_updates = self_id
session player_session_updates = id
session player_lifecycle_updates = player_id
```

For update maps, omitted fields mean unchanged, not cleared. Clients merge update maps into existing lane records and preserve omitted fields. Clearing or removing a record still requires the explicit delete array for that record group.

`total_asteroids` in `session_delta` remains record-level and is not part of the field-delta update-map conversion.

The server compares projected realtime wire records when building deltas. The client does not own authoritative delta decisions.

The compact asteroid tuple exception is applied only at the final active wire shape. It does not change the readable logical create/update/delete descriptions above.

#### Sparse delta serialization

Delta packet serializers now omit empty create, update, and delete sections.

Sparse omission is implemented in `services/game-server/internal/protocol/realtime/wire_packets.go` by the delta wire serializers, before compact aliasing and `packetcodec` encoding.

Full packets remain complete snapshots and are not sparse.

Missing delta section fields mean empty or no-op, not cleared.

Missing update fields inside a present update record mean unchanged, not cleared.

Meaningful `false` and `0` values inside present records are preserved.

In `session_delta`, `total_asteroids` is omitted when unchanged or nil.

A `total_asteroids: 0` value is meaningful and must be preserved when it is an actual emitted value.

This sparse omission applies to readable long-key maps before compact aliases are applied.
Compact aliases apply after sparse omission at the final outbound encode boundary.

This is not binary packing.
Candidate-level scheduling and estimated byte-budget selection are current protocol/realtime behavior. Record/entity-level prioritization and state-lane record or field sub-packet selection are not implemented here.

### World lane packets

`world_full` carries:

```text
ships
bullets
asteroids
pickups
```

`world_delta` may carry any non-empty subset of:

```text
ship_creates
ship_updates
ship_deletes
bullet_creates
bullet_deletes
asteroid_creates
asteroid_deletes
pickup_creates
pickup_updates
pickup_deletes
```

Absent world delta sections are treated as empty arrays by the client. `world_delta` still has serializer and compatibility support for `bullet_updates` and `asteroid_updates` so bootstrap, compatibility, and resync-safe world deltas can continue to deserialize them, but it does not own active bullet or asteroid lifecycle traffic. Regular active bullet movement is emitted as `bullet_delta` on `sr.bullets`, and regular active asteroid movement is emitted as `asteroid_delta` on `sr.asteroids`. For asteroid records only, the compact active wire form uses tuple arrays with numeric suffix IDs, and the client rehydrates those IDs before world lane application. Readable/logical world records remain id-keyed maps before compact aliasing.

### Asteroid lifecycle lane packets

`asteroids_lifecycle` carries:

```text
asteroid_creates
asteroid_deletes
```

It uses `sr.asteroids.lifecycle`.

It is reliable/ordered.

It is required/critical.

It defines asteroid existence.

It carries create identity such as variant, size, health, scale, and initial position when present.

It is not a hot movement lane.

### Lifecycle synchronization and ordering

`ships_lifecycle`, `asteroids_lifecycle`, and `bullets_lifecycle` each have an independent strict lifecycle sequence. The three lifecycle sequences are separate from the world sequence, the hot ship sequence, the hot asteroid sequence, the hot bullet sequence, and each other. Reliable/ordered delivery orders messages only within one DataChannel: there is no ordering guarantee between `sr.world` and any lifecycle channel, or between a lifecycle channel and its corresponding hot channel.

`RealtimeRouter` submits each lifecycle packet to `LifecycleLaneGate`. A packet applies immediately only when the world is synced and its explicit `baseline_id` matches the active world baseline; otherwise it is queued. Unsupported lanes, non-dictionary packets, missing or empty `baseline_id`, missing sequence, negative sequence, non-integral numeric sequence, strings, and booleans are rejected at the gate. Non-negative integral numeric values such as `1.0` are normalized to integer sequence `1`.

Duplicate or lower lifecycle sequences are rejected when `sequence <= latest applied`. Sequence gaps are valid. Lifecycle lanes are assembled from explicit chunks before gate validation and application; duplicate, mismatched, interrupted, or invalid series fail closed and request world resync. `WorldLaneApplier` owns lifecycle payload validation and `WorldLaneState` mutation after gate acceptance; it does not own lifecycle baseline or sequence policy.

After a matching `world_full` is completely applied and its baseline is recorded, `RealtimeRouter` drains pending packets for that baseline, sorting packets ascending within each lifecycle lane. There is no ordering contract between the two lifecycle lanes. A lifecycle sequence advances only after `WorldLaneApplier` successfully validates and mutates the payload.

The pending gate is bounded to four baseline buckets and eight packets per lifecycle lane per baseline. Per-lane packet-capacity overflow and baseline-bucket capacity overflow do not discard old entries while continuing as synchronized. Instead, capacity loss returns an explicit resync decision with reason `lifecycle_queue_overflow`; it clears all pending lifecycle packets and pending duplicate tracking, preserves latest-applied lifecycle sequence state, marks the world lane unsynchronized, makes gameplay not ready, and emits the existing deduplicated world-lane resync request. While resync is pending, later lifecycle packets may queue for the replacement world baseline. A valid replacement `world_full` restores synchronization and drains only post-overflow lifecycle packets. Parsed `world-baseline-N` buckets older than the accepted active baseline are still discarded during normal cleanup; this is separate from capacity overflow. `RealtimePacketPipeline.reset()` replaces the router and clears lifecycle pending queues, pending duplicate tracking, and latest applied lifecycle sequences.

`BaselineTracker` continues to track only world, overlay, and session state. It supplies world synced state and the active world baseline identity; `record_delta` remains ordinary world/overlay/session baseline tracking and is not the lifecycle gate. Lifecycle lanes do not add independent readiness requirements.

Lifecycle gate failure behavior is local client rejection or explicit world-lane resync on bounded queue capacity loss. Lifecycle lanes do not own independent full baselines or readiness. The existing deduplicated world-lane resync request and replacement `world_full` provide recovery; this does not add ack, resend, reconnect recovery, durable queues, or stale-`world_full` protection.

### Bullet lifecycle lane packets

`bullets_lifecycle` carries:

```text
bullet_creates
bullet_deletes
```

It uses `sr.bullets.lifecycle`.

It is reliable/ordered.

It is required/critical.

It defines projectile existence.

It carries create identity such as owner_id, weapon_id, projectile_type, rotation, and initial position when present.

Torpedo creates must preserve projectile_type.

It is not a hot movement lane.

Lifecycle defines existence. Hot lanes update known entities only.

Cross-lane ordering is not guaranteed between reliable lifecycle lanes and unreliable hot lanes. Clients must tolerate hot updates arriving before lifecycle create packets and after lifecycle delete packets.

Compact asteroid example:

```json
{"t":"wf","q":1,"asteroids":[[1,70,80,2,90,1500,3]],"ships":[],"bullets":[],"pickups":[]}
```

The client expands the tuple-packed asteroid record back into a readable dictionary before world lane application. The client expands ID `1` to `asteroid-1`.

Compact asteroid lifecycle example:

```json
{"t":"al","l":"al","q":1,"b":"world-baseline-7","sid":"asteroids-lifecycle-snapshot-1","k":"d","ms":123455,"mid":"match-1","ac":[[1,10,20,2,90,1500,3]],"ax":[1]}
```

Compact bullet lifecycle example:

```json
{"t":"bl","l":"bl","q":2,"b":"world-baseline-7","sid":"bullets-lifecycle-snapshot-2","k":"d","ms":123456,"mid":"match-1","bc":[[1,"player-1",10,20,30,"pulse","laser"]],"bx":[1]}
```

The client expands the tuple-packed bullet lifecycle records back into readable dictionaries before bullet lifecycle application.

Compact bullet movement example:

```json
{"t":"bd","q":3,"bu":[[1,11,21]]}
```

The client expands the tuple-packed bullet movement records back into readable dictionaries before bullet lane application.

Current straight bullet movement normally emits x/y updates only. The compact bullet movement tuple still supports an optional trailing rotation slot for future projectile types that may turn or home during flight.

Compact asteroid movement example:

```json
{"t":"ad","q":4,"au":[[2,12,22]]}
```

The client expands the tuple-packed asteroid movement records back into readable dictionaries before asteroid lane application.

Compact world ship/player example:

```json
{"t":"wd","q":3,"sc":[[1,"v_wing",10,20,30,100,50,true,"player","player-2"]],"su":[[1,11,21,31,false]],"sx":[1]}
```

The client expands the tuple-packed ship records back into readable dictionaries before world lane application.

Compact world_delta update example:

```json
{"t":"wd","q":5,"su":[[1,11,21,31,false]],"pu":[[2,12,22,32]]}
```

The client expands the tuple-packed `world_delta` update records back into readable dictionaries before world lane application.

Current `world_delta` update maps are partial maps keyed by id:

```text
ship_updates
pickup_updates
bullet_delta.bullet_updates
asteroid_delta.asteroid_updates
```

### Overlay lane packets

`overlay_full` flattens the receiver/HUD fields from `OverlayReceiverRecord` at top level.

`overlay_delta` may carry any non-empty subset of:

```text
receiver_creates
receiver_updates
receiver_deletes
```

Absent overlay delta sections are treated as empty arrays by the client.

Current `overlay_delta` update maps are partial maps keyed by `self_id`:

```text
receiver_updates
```

### Session lane packets

`session_full` carries:

```text
players
player_lifecycle
total_asteroids
```

`session_delta` may carry any non-empty subset of:

```text
players
player_session_updates
player_session_deletes
player_lifecycle
player_lifecycle_updates
player_lifecycle_deletes
total_asteroids
```

Absent session delta sections are treated as empty arrays by the client; absent total_asteroids means unchanged.

Current `session_delta` update maps use lane-specific identity keys:

```text
player_session_updates = id
player_lifecycle_updates = player_id
```

Compact session examples:

```json
{"t":"sd","q":4,"pl":[[1,"v_wing",100,3,250,"pulse","limited","mine","limited",10,20]],"psu":[[1,"sco",100,"lv",2,"rcd",0]],"plc":[[1,"active"]],"plu":[[1,"respawning"]],"plx":[1]}
```

The client expands the tuple-packed session records back into readable dictionaries before session lane application.

`total_asteroids` remains record-level and is not part of the field-delta conversion.

### Event lane packets

`event_batch` carries:

```text
batch_id
events
event_id per event
```

The `events` array preserves the authoritative pending presentation-event slice order copied during server projection. Event IDs are identity and deduplication keys; neither the server nor the client may use them to sort or otherwise reorder events. After a successful `event_batch` write, draining the written event IDs preserves the relative order of events that remain pending.

Readable protocol docs may show expanded logical names, while runtime wire sends compact aliases. Domain logs may still show raw x/y before projection.

Compact event batch example:

```json
{
  "t": "eb",
  "q": 412,
  "ms": 1712345678901,
  "bid": 412,
  "ev": [
    ["shd", 100, 2, 2, 3, 512, 384],
    ["dmg", 101, "pickup", "pickup-4", "impact", 20, 512, 384]
  ]
}
```

The client expands the tuple-packed event records back into readable dictionaries before event lane application.

### Event Batch Example

Readable logical example:

```json
{
  "type": "event_batch",
  "sequence": 412,
  "server_sent_msec": 1712345678901,
  "batch_id": "event-batch-412",
  "events": [
    {
      "type": "ship_death",
      "event_id": "evt-100",
      "player_id": "Player-2",
      "lives": 2,
      "respawn_delay": 3,
      "x": 512,
      "y": 384
    },
    {
      "type": "damage_applied",
      "event_id": "evt-101",
      "source_type": "pickup",
      "source_id": "pickup-4",
      "effect_type": "impact",
      "amount": 20,
      "x": 512,
      "y": 384
    }
  ]
}
```

Compact wire example:

```json
{
  "t": "eb",
  "q": 412,
  "ms": 1712345678901,
  "bid": 412,
  "ev": [
    {
      "t": "shd",
      "ei": "evt-100",
      "pid": "Player-2",
      "lv": 2,
      "rd": 3,
      "x": 512,
      "y": 384
    },
    {
      "t": "dmg",
      "ei": "evt-101",
      "srct": "pickup",
      "src": "pickup-4",
      "fx": "impact",
      "amt": 20,
      "x": 512,
      "y": 384
    }
  ]
}
```

Readable/logical docs may show expanded names, while runtime wire sends compact aliases. Domain logs may still show raw x/y before projection.

### Resync packets

`resync_request` and `resync_required` are the current generated packet families for control-lane recovery signaling.

This does not define a separate generated packet family named `control`.

### Scheduling and delivery classes

The current scheduler assigns delivery classes and priorities at whole-lane-candidate granularity and selects included candidates against an estimated byte budget. Lifecycle logical candidates may expand into one or more hard-capped packet candidates before scheduling and encoding.

The size policies have three separate meanings: the roughly 1,200 B construction cap applies to full, lifecycle, asteroid movement, and bullet movement candidates; the scheduler's 500 B `TargetBytes` is an advisory candidate-selection target; and there is no aggregate per-tick byte cap.

```text
event_batch = critical/event-once
world_delta, overlay_delta, ship_delta, asteroid_delta, and bullet_delta = high priority / hot supersedable
ship_delta, asteroid_delta, and bullet_delta = dedicated hot movement candidates with chunker-owned hard-size guarding and unordered/unreliable WebRTC delivery
asteroids_lifecycle = required / critical
bullets_lifecycle = required / critical
session deltas = medium priority / deferrable
required bootstrap full packets = world, overlay, then session
recovery full candidates = required only for the invalidated world, overlay, or session lane
```

Lifecycle lanes are not hot-supersedable.
Full and lifecycle candidates are split by the general realtime candidate expansion path when needed; hot movement candidates retain their same-sequence chunk semantics.

The active path currently schedules whole lane candidates:

```text
world_delta = one candidate
asteroid_delta = one or more candidates when asteroid hot movement is present; oversized hot movement update lists are split into bounded same-sequence chunks
bullet_delta = one or more candidates when bullet hot movement is present; oversized hot movement update lists are split into bounded same-sequence chunks
asteroids_lifecycle = one logical candidate when asteroid creates/deletes are present; it may expand into one or more hard-capped packet candidates before scheduling/encoding
bullets_lifecycle = one logical candidate when bullet creates/deletes are present; it may expand into one or more hard-capped packet candidates before scheduling/encoding
overlay_delta = one candidate
session_delta = one candidate
event_batch = one candidate
```

Hot ship_delta, asteroid_delta, and bullet_delta packets require finite, non-negative integer-valued numeric sequence values according to the `int` schema. Fractional, negative, missing, non-finite, string, and boolean values are invalid. Sequence gaps are valid because unordered/unreliable hot packets may be dropped.

For `ship_delta`, `asteroid_delta`, and `bullet_delta`, the client accepts a same-sequence packet only when its `chunk_index` has not already been accepted for that lane and sequence and its `chunk_count` matches the count already established for that sequence. Duplicate chunk indices, inconsistent `chunk_count` values, malformed chunk metadata, and lower sequence values are rejected. Distinct chunk indices may arrive in any order, sequence gaps are valid because unordered/unreliable hot packets may be dropped, and ship, asteroid, and bullet chunk tracking are independent.

The active path does not implement general record/entity-level prioritization or arbitrary field-level packet splitting. It does implement focused hot-lane chunking for `ship_delta`, `asteroid_delta`, and `bullet_delta`: oversized hot movement update lists become multiple real candidates before `SelectSendPlan` and before final JSON encoding. Hot-lane chunk sizing uses conservative compact-JSON byte estimation. The chunker is the only hot-lane hard-size guard. Scheduler byte estimates remain advisory for candidate-level include/defer decisions, and active encoding records final encoded bytes for diagnostics rather than rejecting already-scheduled hot packets. Scheduler and active encoding do not reject already-built hot candidates; chunk construction normally prevents multi-update chunks above `HardCapBytes`, while an unsplittable single-update chunk may exceed the threshold and still be encoded and sent, with diagnostics recording it. Deferred and supersession storage exists as protocol plumbing, but active cross-tick replay and supersession are not yet the gameplay delivery guarantee.
### Runtime observability note

Current runtime debug observability is intentionally narrow:

```text
per-packet debug logs
= encoded packet write observations including encoded_bytes

non-empty per-tick debug summaries
= packet_count plus total encoded_bytes for ticks that actually wrote packets
```

Current runtime does not emit `realtime lane metric` logs or scheduler, budget, deferred, superseded, or record-level counter fields as active protocol log output.

Byte estimates in planning and scheduling remain advisory and are not codec-accurate, but current candidate-level send-plan selection uses them for include/defer decisions against the target budget. Record/entity-level prioritization remains future work, and the active path still does not split state-lane deltas into selected record or field sub-packets. Current runtime debug logs do not emit scheduler, budget, deferred, superseded, or record-level counter fields as active protocol log output.

## Server WebSocket inbound routing order

Runtime devtools commands do not participate in the WebSocket envelope route. The remaining WebSocket inbound routing order is:

```text
normal game.ClientPacket decode
auth packets
telemetry packets
lobby packets
gameplay packets
```

Runtime devtools commands use the separate `sr.tooling` route and do not route before normal `game.ClientPacket` decode.

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
NetworkClient.packet_received(packet)
-> ClientConnectionService._on_packet_received(packet)
-> ServerPacketDispatcher.dispatch(packet)
-> ClientInboundCoordinator coordinator signal or typed realtime binding
```

Non-realtime ServerPacketDispatcher outputs:

```text
ServerPacketRouter packet-type checks
typed dispatcher output
ClientInboundCoordinator application-facing signal
ClientConnectionService handler
ClientConnectionService public signal
session, room, telemetry, or devtools consumer
```

ServerPacketDispatcher realtime outputs:

```text
world_full_received
world_delta_received
asteroid_delta_received
bullet_delta_received
asteroids_lifecycle_received
bullets_lifecycle_received
overlay_full_received
overlay_delta_received
session_full_received
session_delta_received
event_batch_received
resync_request_received
resync_required_received
```

ClientInboundCoordinator binds the typed realtime outputs from ServerPacketDispatcher to the matching RealtimePacketPipeline packet-family entry points. ClientConnectionService does not own direct realtime dispatcher bindings; it remains the public facade for coordinator signals.

The pipeline expands and validates recognized packets first. Without an authoritative active match it buffers non-empty-match packets without routing; after activation it routes only matching packets through the active RealtimeRouter, refreshes RealtimePresentationState, emits gameplay_packet_applied(packet), and then PresentationBridge.handle_gameplay_packet(packet) consumes the handoff into presentation layers.

WebSocket and WebRTC gameplay delivery use different transports but converge on the same RealtimePacketPipeline application boundary.

`SessionNetworkController` continues to handle room, session, pause, debug, and other control routing. Its authoritative `InGame` room-snapshot path resets gameplay presentation/composition and calls `begin_realtime_match(match_id)` before gameplay-session acceptance. Protocol match activation is distinct from `PresentationBridge` activation/readiness: matching packets may route after protocol activation while presentation remains inactive, but pre-activation packets are buffered rather than routed.

Client baseline and readiness behavior is currently:

```text
world, overlay, and session baselines must be synced before gameplay readiness is true
deltas require a valid baseline
stale or baseline-mismatched deltas are rejected or ignored by the router/applier path
```

Lane application responsibilities are split by lane:

```text
world lane -> ships, pickups, world/full/bootstrap presentation state
asteroids_lifecycle -> RealtimeRouter -> LifecycleLaneGate -> apply/queue/reject/resync -> WorldLaneApplier.apply_asteroids_lifecycle
bullets_lifecycle -> RealtimeRouter -> LifecycleLaneGate -> apply/queue/reject/resync -> WorldLaneApplier.apply_bullets_lifecycle
asteroid_delta -> regular asteroid movement updates through hot sequence guards
bullet_delta -> regular bullet movement updates through hot sequence guards
overlay lane -> receiver and HUD state
session lane -> player sessions, lifecycle, and total asteroid count
event_batch -> deduped by batch and event identifiers, then drained into event presentation
```

Presentation adapters fan the current lane state into the gameplay runtime surface without turning this protocol doc into a client runtime guide.

Current client-recognized inbound packet types include:

```text
authenticate_result
room_snapshot
room_state_changed
room_error
world_full
world_delta
asteroid_delta
bullet_delta
asteroids_lifecycle
bullets_lifecycle
overlay_full
overlay_delta
session_full
session_delta
event_batch
resync_request
resync_required
player_pause_state
telemetry_pong
```

Developer readouts are recognized separately by `ToolingPacketRouter` after arrival on `sr.tooling`; they are not part of the normal WebSocket dispatcher list.

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

runtime devtools command and developer-readout request packets
use `sr.tooling`; require packet-specific room attachment and tooling capability, but not `GamePlayerID` merely for room-global or explicit-target dispatch

lane gameplay output
requires current room, active game player, and eligible room game state
```

Current gameplay readiness uses the realtime client lane path. World, overlay, and session baselines must be synced before gameplay readiness becomes true. Delta application requires a valid baseline, and stale or baseline-mismatched deltas are rejected or ignored by the router/applier path.

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

The server replies only to the same WebSocket session. Telemetry does not require room membership, does not require active lane gameplay output, and does not mutate gameplay. WebSocket best-effort applies to auth, room, lobby, telemetry, signaling, and queued one-off packets; active realtime gameplay output uses ordered/reliable lanes for `sr.world`, `sr.overlay`, `sr.session`, `sr.event`, `sr.ships.lifecycle`, `sr.asteroids.lifecycle`, and `sr.bullets.lifecycle`, and unordered/unreliable hot-update lanes for `sr.ships`, `sr.asteroids`, and `sr.bullets`. There is no ack, resend, reconnect, session-resume, or durable outbound queue for that delivery path.

## Delivery and failure semantics

Current delivery is best-effort for WebSocket-owned auth, room, lobby, telemetry, signaling, and queued one-off packets. Active realtime gameplay output uses ordered/reliable lanes for `sr.world`, `sr.overlay`, `sr.session`, `sr.event`, `sr.ships.lifecycle`, `sr.asteroids.lifecycle`, and `sr.bullets.lifecycle`, and unordered/unreliable hot-update lanes for `sr.ships`, `sr.asteroids`, and `sr.bullets`.

There is no implemented support for:

```text
general packet-delivery acknowledgements
server resend
client resend
reconnect recovery
session resume
durable outbound queues
```

Current lane-native delivery does include sequence numbers, baseline tracking, and delta snapshots as part of the active gameplay protocol. Those mechanisms support in-session lane ordering and incremental updates. `resync_required` acknowledges acceptance of one baseline recovery request; it does not acknowledge arbitrary packet delivery. In-session world/overlay/session baseline recovery is distinct from resend, reconnect recovery, session resume, or a durable outbound queue.

Client outbound sends are not queued. If the WebSocket is not open, the packet is not sent. Active realtime gameplay output uses ordered/reliable lanes for `sr.world`, `sr.overlay`, `sr.session`, `sr.event`, `sr.ships.lifecycle`, `sr.asteroids.lifecycle`, and `sr.bullets.lifecycle`, and unordered/unreliable hot-update lanes for `sr.ships`, `sr.asteroids`, and `sr.bullets`. There is no ack, resend, reconnect, session-resume, or durable outbound queue for that delivery path.

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
-> unknown packet handling
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
shared/packets/webrtc.toml
shared/packets/outputs.toml
```

Those files define packet structs, packet type strings, JSON field names, output routing, and selected generated client builders. For gameplay output, `shared/packets/gameplay.toml` and `shared/packets/outputs.toml` now include the generated packet type values `ship_delta`, `ships_lifecycle`, `asteroid_delta`, `asteroids_lifecycle`, `bullet_delta`, and `bullets_lifecycle` alongside the existing gameplay families.

The transport route and runtime connection lifecycle are not sourced from the packet TOML files. They are implemented by the client and game-server networking services.

## Generated outputs

Current generated outputs used by the realtime WebSocket protocol include:

```text
shared/packets/outputs.toml
shared/packets/gameplay.toml
shared/packets/lobby.toml
shared/packets/debug.toml
```

Those TOML files are the source inputs for the generated packet outputs below. They define packet type strings, field names, output routing, and selected generated constants and structs.

```text
client/scripts/generated/networking/packets/packets.gd
services/game-server/internal/protocol/realtime/packets_generated.go
services/game-server/internal/game/packets.go
services/game-server/internal/game/runtime/packets_generated.go
services/game-server/internal/devtools/packets_generated.go
```

The generated client file provides packet type constants, field constants, and selected outbound packet builder functions.

Logical packet generation produces packet constants and Go/GDScript structs from the configured packet TOML sources, including `services/game-server/internal/protocol/realtime/packets_generated.go`. Physical realtime-wire generation separately produces `services/game-server/internal/protocol/realtimewire/generated.go`, `client/scripts/generated/networking/realtime_wire_generated.gd`, `shared/packets/generated/realtime_wire.json`, and `docs/protocol/generated/realtime-wire-reference.md`. Runtime projection, descriptor application, scheduling, transport, and packet application files are implementation consumers rather than generated contract owners.

`shared/packets/realtime_wire.toml` owns the physical compact-wire contract. `compact_wire_packet.go` is the public encode boundary and generic fallback; `compact_wire_descriptor.go` applies generated bindings, records, aliases, ID rules, event layouts, and metadata policy.

Runtime wire conversion for lane packets lives in `services/game-server/internal/protocol/realtime/`. Client lane application lives in `client/scripts/protocol/realtime/`.

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

The game simulation owns gameplay authority behind gameplay packets and lane gameplay output:

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
lane-native realtime projection
match-over policy integration
```

### Devtools

Devtools own debug command behavior after `networking/tooling` applies policy, room, and capability preflight and decodes a `DebugCommand`.

Devtools commands and developer-readout requests use `sr.tooling`. Commands dispatch through the existing devtools controller and return correlated `tooling_command_result` or `tooling_error`; readouts preserve their existing application ownership after `ToolingPacketRouter`. Legacy continuous telemetry/ping remains on its existing path. Devtools must not bypass real server-owned gameplay seams.

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

Physical realtime-wire changes use:

```text
data-sync -validate -realtime-wire
data-sync -diff -realtime-wire -go -gds -json -docs
data-sync -push -realtime-wire -go -gds -json -docs
data-sync -check -realtime-wire -go -gds -json -docs
```

Readable packet type strings, structs, fields, and JSON names come from the logical packet schema. Compact aliases, packet metadata, record encodings, tuple/sparse layouts, quantization assignments, ID codecs/selectors, event layouts, and decode alternatives come from `shared/packets/realtime_wire.toml`. Lane scheduling, delta projection, quantization math, transport, baseline behavior, and packet application remain runtime responsibilities. Binary encoding, protobuf migration, schema negotiation, and wire-version negotiation are not implemented.

## Code map

### Client realtime protocol and networking

```text
client/scripts/protocol/realtime/
client/scripts/networking/inbound/server_packet_router.gd
client/scripts/networking/inbound/server_packet_dispatcher.gd
client/scripts/networking/client_connection_service.gd
client/scripts/session/session_network_controller.gd
client/scripts/session/gameplay_session_controller.gd
client/scripts/networking/network_client.gd
client/scripts/networking/webrtc/webrtc_transport.gd
client/scripts/networking/packets/packet_codec.gd
client/scripts/protocol/realtime/compact_lane_packet.gd
client/scripts/protocol/realtime/world_lane_applier.gd
client/scripts/protocol/realtime/realtime_router.gd
client/scripts/protocol/realtime/lifecycle_lane_gate.gd
client/scripts/protocol/realtime/baseline_tracker.gd
client/scripts/networking/packets/packet_encode_result.gd
client/scripts/networking/packets/packet_decode_result.gd
```

PacketCodec.decode owns compact alias expansion before envelope validation; compact_lane_packet.gd owns the alias expansion helper used by PacketCodec and the defensive realtime router normalization path. Client compatibility tests include `client/tests/unit/protocol/realtime/test_compact_lane_packet.gd`, `client/tests/unit/protocol/realtime/test_world_lane_applier.gd`, `client/tests/unit/protocol/realtime/test_overlay_session_lane_applier.gd`, and `client/tests/unit/protocol/realtime/test_event_batch_and_resync.gd`. Related routing and applier paths include `client/scripts/networking/inbound/server_packet_router.gd`, `client/scripts/networking/inbound/server_packet_dispatcher.gd`, `client/scripts/protocol/realtime/world_lane_applier.gd`, and `client/scripts/protocol/realtime/realtime_router.gd`.

### Client boot/session participants

```text
client/scripts/boot/session_boot_controller.gd
client/scripts/boot/shell_boot_flow.gd
client/scripts/boot/pending_boot_request.gd
client/scripts/boot/session_network_target.gd
client/scripts/session/room_session_controller.gd
```
### Server realtime protocol and networking

```text
services/game-server/internal/protocol/realtime/lanes.go
services/game-server/internal/protocol/realtime/metadata.go
services/game-server/internal/protocol/realtime/records.go
services/game-server/internal/protocol/realtime/projection_world.go
services/game-server/internal/protocol/realtime/projection_overlay.go
services/game-server/internal/protocol/realtime/projection_session.go
services/game-server/internal/protocol/realtime/event_projection.go
services/game-server/internal/protocol/realtime/baseline.go
services/game-server/internal/protocol/realtime/delta.go
services/game-server/internal/protocol/realtime/planner.go
services/game-server/internal/protocol/realtime/lane_candidate_world.go
services/game-server/internal/protocol/realtime/lane_candidate_lifecycle.go
services/game-server/internal/protocol/realtime/lane_candidate_overlay.go
services/game-server/internal/protocol/realtime/lane_candidate_session.go
services/game-server/internal/protocol/realtime/lane_candidate_event.go
services/game-server/internal/protocol/realtime/candidate_types.go
services/game-server/internal/protocol/realtime/candidate_policy.go
services/game-server/internal/protocol/realtime/candidate_diagnostics.go
services/game-server/internal/protocol/realtime/quantize_overlay.go
services/game-server/internal/protocol/realtime/quantize_session.go
services/game-server/internal/protocol/realtime/quantization_propagation_test.go
services/game-server/internal/protocol/realtime/scheduler.go
services/game-server/internal/protocol/realtime/priority.go
services/game-server/internal/protocol/realtime/size_estimate.go
services/game-server/internal/protocol/realtime/hot_lane_chunker.go
services/game-server/internal/protocol/realtime/hot_lane_size_estimate.go
services/game-server/internal/protocol/realtime/wire_packets.go
services/game-server/internal/protocol/realtime/lanes.go
services/game-server/internal/protocol/realtime/active.go
services/game-server/internal/protocol/realtime/compact_wire_packet.go
services/game-server/internal/protocol/realtime/compact_wire_descriptor.go
services/game-server/internal/protocol/realtimewire/generated.go
shared/packets/realtime_wire.toml
docs/protocol/generated/realtime-wire-reference.md
services/game-server/internal/protocol/realtime/shadow.go
services/game-server/internal/protocol/realtime/parity.go
services/game-server/internal/protocol/realtime/metrics_bridge.go
services/game-server/internal/protocol/realtime/packets_generated.go
services/game-server/internal/networking/websocket_write.go
services/game-server/internal/networking/websocket.go
services/game-server/internal/networking/websocket_read.go
services/game-server/internal/networking/websocket_session.go
services/game-server/internal/networking/webrtc_transport.go
services/game-server/internal/networking/tooling/router.go
services/game-server/internal/networking/tooling/preflight.go
services/game-server/internal/networking/tooling/commands.go
services/game-server/cmd/game-server/webrtc_config.go
```

### Server packet codec and generated packet files

```text
services/game-server/internal/protocol/packetcodec/codec.go
services/game-server/internal/game/packets.go
services/game-server/internal/game/runtime/packets_generated.go
services/game-server/internal/devtools/packets_generated.go
```

### Server outbound and support boundaries

```text
services/game-server/internal/networking/outbound/server_message_writer.go
services/game-server/internal/networking/outbound/gameplay_presentation.go
services/game-server/internal/networking/packetmetrics/
services/game-server/internal/networking/outbound/debug_status_presentation.go
services/game-server/internal/networking/outbound/debug_shape_catalog_presentation.go
services/game-server/internal/networking/room_snapshot.go
services/game-server/internal/networking/room_error.go
services/game-server/internal/networking/session_auth.go
services/game-server/internal/networking/player_pause_state.go
```

### Shared packet source files

```text
shared/packets/gameplay.toml
shared/packets/lobby.toml
shared/packets/debug.toml
shared/packets/webrtc.toml
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
services/game-server/internal/protocol/realtime/*_test.go
```

Realtime protocol tests cover sparse delta omission in `wire_packets_test.go`, world/overlay/session wire delta serialization, quantized wire values, and planner baseline/delta behavior. `services/game-server/internal/protocol/realtime/quantization_propagation_test.go` verifies that exported planner construction and active-result construction surface world, overlay, and session quantization failures rather than silently omitting a lane. Lifecycle coverage includes `client/tests/unit/protocol/realtime/test_lifecycle_lane_gate.gd`, `client/tests/unit/networking/realtime/test_realtime_packet_pipeline.gd` reset state replacement/clear coverage, and server lifecycle metadata tests in `services/game-server/internal/protocol/realtime/wire_packets_test.go`.

Relevant client tests include:

```text
client/tests/unit/test_packet_codec.gd
client/tests/unit/protocol/realtime/test_compact_lane_packet.gd
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
client/tests/unit/protocol/realtime/test_lane_protocol_routing.gd
client/tests/unit/protocol/realtime/test_lifecycle_lane_gate.gd
client/tests/unit/protocol/realtime/test_world_lane_applier.gd
client/tests/unit/protocol/realtime/test_overlay_session_lane_applier.gd
client/tests/unit/protocol/realtime/test_event_batch_and_resync.gd
client/tests/unit/protocol/realtime/test_gameplay_readiness.gd
client/tests/unit/protocol/realtime/test_lane_native_presentation_adapters.gd
client/tests/unit/protocol/realtime/test_devtools_lane_state_adapter.gd
```

`test_packet_codec.gd` verifies client JSON packet encode/decode and envelope validation. `test_compact_lane_packet.gd` verifies compact alias expansion for realtime lane packets.

`gameplay_packets_test.go` verifies current gameplay packet routing behavior, including `client_config` routing into the game instance and `start_single_player_request` routing through lobby handling while preserving Local Profile ID on the room member.

`session_auth_test.go` verifies session identity mutation after successful WebSocket auth.

## Active issues

* `start_single_player_request` does not currently reject an already-authenticated WebSocket session at the server boundary. See [Current System Limits](../limits/current-system-limits.md#architecture--networking).

## Related docs

* [Protocol](./!INDEX.md)
* [Realtime Client Server Flow](../domains/technical/realtime-client-server-flow.md)
* [Gameplay Session Flow](../domains/player-experience/gameplay-session-flow.md)
* [Client](../services/client/!INDEX.md)
* [Client Networking Flow](../services/client/networking-flow/!INDEX.md)
* [WebSocket Connection Lifecycle](../services/client/networking-flow/websocket-connection-lifecycle.md)
* [Client Outbound Packet Sending](../services/client/networking-flow/outbound-packet-sending.md)
* [Client Inbound Packet Routing](../services/client/networking-flow/inbound-packet-routing.md)
* [Session Boot And Network Target](../services/client/app-shell-and-session/session-boot-and-network-target.md)
* [Game Server](../services/game-server/!INDEX.md)
* [Game Server Networking](../services/game-server/networking/!INDEX.md)
* [WebSocket Session Lifecycle](../services/game-server/networking/websocket-session-lifecycle.md)
* [Game Server Inbound Packet Routing](../services/game-server/networking/inbound-packet-routing.md)
* [Game Server Outbound Packet Routing](../services/game-server/networking/outbound-message-flow.md)
* [Auth Routing](../services/game-server/networking/auth-routing.md)
* [Telemetry Packet Routing](../services/game-server/networking/telemetry-packet-routing.md)
* [Game Server Rooms](../services/game-server/rooms/!INDEX.md)
* [Game Server Simulation](../services/game-server/simulation/!INDEX.md)
* [Packet Schemas](../data/packet-schemas.md)
* [Source Of Truth Map](../data/source-of-truth-map.md)
* [Realtime Protocol Architecture](../planning/protocol/realtime-protocol-architecture.md)
* [Network Observability And Packet Budget](../planning/domains/technical/network-observability-and-packet-budget.md)
* [Current System Limits](../limits/current-system-limits.md)
* [Realtime WebRTC Gameplay Transport](realtime-webrtc-gameplay-transport.md)

## Notes

The current implementation sends lane-native gameplay output on the server tick path over ordered/reliable lanes for `sr.world`, `sr.overlay`, `sr.session`, `sr.event`, `sr.ships.lifecycle`, `sr.asteroids.lifecycle`, and `sr.bullets.lifecycle`, plus unordered/unreliable hot-update lanes for `sr.ships`, `sr.asteroids`, and `sr.bullets`. That is current protocol behavior, not the intended final realtime architecture. The client ICE-server seam exists, but this document does not prescribe a future TURN/STUN topology.

## Client receive hardening

Client pre-match buffering is capped at 4 match buckets, 128 packets and 256 KiB estimated expanded JSON per match, with a 5000 ms lifetime and oldest-bucket eviction. Lost or expired state for the selected authoritative match fails closed and requests world, overlay, and session recovery. `world_full`, asteroid lifecycle, and bullet lifecycle assemblies each allow 128 chunks, 16384 cumulative records, 2 MiB estimated expanded JSON, and 5000 ms. Any limit, expiry, malformed metadata, interrupted, duplicate, mismatched, or non-contiguous failure resets the incomplete assembly, applies no partial state, and requests authoritative recovery.

Baseline IDs are non-empty strings; sequence and chunk values are finite, non-negative, integer-valued numerics; final-chunk metadata is boolean and must agree with index/count. Valid stale deltas remain silently rejected. These are defensive client receive limits, not changes to the server's approximately 1200-byte candidate construction cap.

Deployment knobs currently include SPACE_ROCKS_WEBRTC_ADVERTISED_IPS, SPACE_ROCKS_WEBRTC_UDP_PORT_MIN, SPACE_ROCKS_WEBRTC_UDP_PORT_MAX, and WEBRTC_ICE_SERVERS. The client ICE-server seam exists for future deployment configuration, but this document does not prescribe TURN or other future ICE topology beyond noting that the seam exists.

The current WebSocket protocol is transport/session scoped. Durable match-result persistence happens through player-data routing after authoritative match facts are produced; it is not a WebSocket delivery guarantee.

The generated packet schema defines the shared packet vocabulary, but service implementation still determines runtime consequences. New packets should update source TOML, generated outputs, runtime handlers, tests, and protocol documentation together.
