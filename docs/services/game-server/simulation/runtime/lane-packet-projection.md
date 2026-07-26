---
author: brian
created: "2026-07-19"
document_id: 019f7d55-fb2c-7ecc-aec2-5ef861a7a273
document_type: general
policy_exempt: false
summary: This document describes the active game-server lane-native realtime projection path for realtime gameplay presentation.
---
# Lane Packet Projection

Parent index: [Game Server Simulation Runtime](./!INDEX.md)

## Purpose

This document describes the active game-server lane-native realtime projection path for realtime gameplay presentation.


## Overview

The game server projects authoritative gameplay state into lane packet candidates and results.

The active flow is:

```text
authoritative game state
-> realtime projection / planning
-> raw lane records
-> numeric wire quantization into wire-shaped records
-> lane candidate selection, delta comparison, and hot movement split
-> regular ship movement updates move to dedicated hot-lane delta packets on sr.ships, asteroid movement updates move to sr.asteroids, and bullet movement updates move to sr.bullets
-> ship/asteroid/bullet creates/deletes split to reliable lifecycle lanes on sr.ships.lifecycle, sr.asteroids.lifecycle, and sr.bullets.lifecycle
-> oversized ship/asteroid/bullet hot movement update lists expand into real same-sequence candidate chunks using conservative compact-JSON byte estimates
-> the chunker is the only hard-size guard for ship/asteroid/bullet hot movement packets; scheduler and active encoding consume already-shaped candidates
-> typed candidate payload
-> payload wire serializer produces the sparse readable wire map
-> compact descriptor encoder
-> packetcodec JSON encoding
-> encoded-byte accounting
-> networking write integration
-> debug wire/summary logging after successful writes
-> WebRTC gameplay lane write using the current per-lane reliability policy
```

Projection is lane-specific rather than one combined gameplay snapshot.

## Code root

```text
services/game-server/internal/protocol/realtime/
services/game-server/internal/networking/websocket_write.go
services/game-server/internal/networking/webrtc_transport.go
```

The realtime package owns candidate construction, send-plan records, metadata, wire packet assembly, numeric wire quantization, delta comparison, subtractive ship/asteroid/bullet movement splitting, sparse omission, generated compact descriptor application, and encoded-byte accounting inputs. Hot-lane hard-size guarding for ship, asteroid, and bullet movement packets belongs to `hot_lane_chunker.go`; scheduler and active encoding must not duplicate that guard. The session write loop owns tick-driven invocation; active gameplay lane delivery uses ordered/reliable WebRTC channels for world, overlay, session, and event traffic, plus unordered/unreliable WebRTC channels for ship, asteroid, and bullet hot movement traffic. Networking owns successful delivery handling, post-write state changes, and the current successful-write debug wire/summary logging. For `event_batch`, the realtime package shapes sparse event-type-specific wire records rather than broad reflected `EventState` output.

## Responsibilities

The active server projection path owns:

* Projecting authoritative runtime state into lane-native packet families.
* Keeping world, overlay, session, and event ownership separate.
* Producing receiver-specific overlay/session/event output where needed.
* Preserving explicit event-batch drain semantics.
* Leaving JSON encode/decode mechanics to packetcodec and WebRTC active gameplay transport/write success handling to networking.

`RealtimeLanePayload` is the owning typed candidate contract in `payload.go`; `payload_validation.go` owns the supported matrix, registry, and invariant validation; and the lane-specific `payload_*.go` files own concrete packet implementations. `RealtimeLaneCandidate` owns only the payload and projection; its lane, kind, packet-family, and metadata methods derive from that payload. Wire encoding validates the supported concrete value payload, metadata/family matrix, non-empty wire map, and matching wire `type` before compact encoding. Invalid payloads fail closed rather than producing an empty map.

Every new realtime packet family must add a supported concrete value payload implementation, compile-time interface assertion, family-matrix and payload-registry entry, wire serializer, and focused invariant coverage.

## Does not own

The lane projection path does not own:

* WebRTC active gameplay transport.
* Packet schema source-of-truth files.
* JSON encode/decode mechanics.
* Room lifecycle.
* Client rendering.
* Match rules or simulation mutation.
* WebRTC delivery scheduling, write integration, write success/failure handling, and post-write state changes in networking.

## Protocols and APIs

Canonical gameplay-family overview: [Gameplay packets](../../../../protocol/gameplay-packets.md)
Canonical detailed lane protocol: [Realtime WebSocket Protocol](../../../../protocol/realtime-websocket-protocol.md)

This doc only covers the projection-side service boundary. It does not define wire lifecycle, transport behavior, baseline rules, sequencing, or resync semantics.

## Data ownership

The lane projection path owns the transient projection results used to build lane-native realtime packets from authoritative game state.

It does not own logical packet schema or physical wire-contract source files. Logical packet types, structs, fields, and JSON names belong to the configured packet TOML. Physical aliases, metadata, record encodings, tuple/sparse layouts, quantization assignments, ID codecs/selectors, event layouts, and decode alternatives belong to `shared/packets/realtime_wire.toml`. Runtime projection and codec algorithms consume generated descriptors.

## Lane ownership

Current gameplay presentation ownership is split as:

```text
world lane
= pickups, reliable world projection state, and full/bootstrap world snapshots

ships.lifecycle lane
= ship creates/deletes and reliable non-transform updates such as health, shields, ship type, and target state

ships lane
= regular ship movement updates

asteroids.lifecycle lane
= asteroid creates/deletes

bullets.lifecycle lane
= bullet/projectile creates/deletes

asteroids lane
= regular asteroid movement updates

bullets lane
= regular bullet movement updates

overlay lane
= local-player presentation facts such as lives, score, loadout, cooldown-facing HUD facts

session lane
= durable match-local player session state and lifecycle-oriented read models
event_batch
= transient presentation events sent separately from baseline/delta lanes
```

Ship, asteroid, and bullet lanes produce hot, high-priority, supersedable movement candidates. Lifecycle defines existence. Hot lanes update known entities only.

Hot movement cadence is enforced independently for ships, asteroids, and bullets during candidate construction using the per-session 60 Hz cadence tick. Every hot lane uses the same chunk-pressure tiers: one chunk emits at 60 Hz, two chunks at 30 Hz, three chunks at 20 Hz, and four or more chunks at the 15 Hz floor. Cadence never drops below 15 Hz; all chunks for an eligible logical sequence are emitted in one same-tick unordered burst so additional pressure increases parallel in-flight chunk count instead of lowering cadence again.

Movement in one entity family never forces another hot lane to bypass its cadence. Reliable world and lifecycle changes also do not force hot emission. Each hot lane compares against and commits its own movement projection, while the reliable world projection commits independently. Reliable lifecycle/world candidates remain eligible on ticks where a related hot lane is cadence-suppressed. Networking preflights and commits each hot lane independently, while the reliable world plus lifecycle candidates form one atomic send group.

Lifecycle candidates are required/critical and must not be treated as hot-supersedable movement candidates.

`player_pause_state` remains a separate same-session packet and is handled independently from lane-native realtime projection.

## Delta projection behavior

The realtime projection path builds lane records from authoritative game state before delta comparison.

Field-delta comparison is current behavior for these update groups:

```text
world lane
= pickup and reliable world updates, plus full/bootstrap/resync-compatible entity sections

ships_lifecycle
= ship creates/deletes and reliable non-transform updates

ships lane
= regular ship movement updates only

asteroids_lifecycle
= asteroid_creates and asteroid_deletes

bullets_lifecycle
= bullet_creates and bullet_deletes

asteroids lane
= regular asteroid movement updates only

bullets lane
= regular bullet movement updates only

overlay lane
= receiver updates

session lane
= player session and player lifecycle updates
```

Creates remain full records. Deletes remain identity lists. Update groups carry partial maps with the identity key plus changed fields only.

Client lane state merges partial update maps into existing records and preserves omitted fields. Omitted fields mean unchanged, not cleared.

Sparse delta serialization is current behavior after projection, quantization, and delta comparison. The order is:

```text
authoritative gameplay state
-> raw lane records
-> numeric wire quantization into wire-shaped records
-> delta comparison on projected wire-shaped values
-> sparse delta serializers emit only non-empty delta sections into readable wire maps
-> CompactWirePacket applies generated packet bindings, aliases, value domains, ID codecs/selectors, record encodings, and event layouts
-> packetcodec encodes JSON
```

Sparse omission is a realtime wire-map serialization concern. The physical compact contract is schema-owned; generic descriptor application is a realtime encode-boundary algorithm concern. `packetcodec` only encodes the already-shaped map to JSON. Networking only writes encoded bytes after realtime builds them. Full lane-native realtime packets remain complete snapshots. Delta create, update, and delete sections are omitted when empty. Clients treat missing delta sections as empty or no-op, and missing fields inside update records remain unchanged, not cleared. Sparse omission must not drop meaningful `false` or `0` values inside present records.

Runtime active encoding no longer performs raw-float reflection scanning. Numeric wire quantization remains part of projection and explicit event wire shaping before compacting and JSON encoding.

Numeric wire quantization is implemented in the realtime projection and wire-record path before delta comparison. The active server implementation uses `services/game-server/internal/protocol/realtime/quantize/`, `services/game-server/internal/protocol/realtime/quantize_world.go`, `services/game-server/internal/protocol/realtime/quantize_overlay.go`, and `services/game-server/internal/protocol/realtime/quantize_session.go` as the quantization boundary for outbound lane projection. World, overlay, and session candidate construction is fail-closed: each lane builder returns a quantization error, planner assembly and active-result construction propagate it, and the failed lane is not silently omitted while other lanes continue. Realtime projection owns detection and error return; networking owns the existing gameplay-build failure log/boundary. Quantization should not truncate authoritative simulation state for packet-size savings. Quantization algorithms remain runtime-owned, while field-path policy assignments come from generated realtime-wire descriptors.

The ownership boundary remains:

```text
simulation
= authoritative gameplay state

realtime projection
= lane packet shaping and delta comparison

packetcodec
= JSON encode/decode mechanics

networking
= WebRTC active gameplay write integration and write success/failure handling
```

## Event semantics

Presentation event projection is non-draining until the active send path explicitly drains after a successful active write.

The important rule is:

```text
projection may inspect or copy pending presentation events
active send/write path is the drain point after a successful active write
```

Projection, shadow, and inspection paths must not treat event access as an implicit flush.

## Code map

Relevant active files include:

* `services/game-server/internal/protocol/realtime/` - lane candidates, metadata, send-plan records, baseline/delta planning, wire packets, sparse omission, generated compact descriptor application, encoded-byte accounting inputs, and shadow/parity helpers.
* `services/game-server/internal/protocol/realtime/hot_lane_allocator.go` - movement/cold-field classification support for dedicated hot and reliable lifecycle records.
* `services/game-server/internal/protocol/realtime/hot_lane_projection.go` - independent ship, asteroid, and bullet movement projections, including reliable-world membership synchronization that preserves deferred hot movement.
* `services/game-server/internal/protocol/realtime/hot_lane_chunker.go` - focused candidate-level chunking for oversized `ship_delta`, `asteroid_delta`, and `bullet_delta` movement update lists; this is the only hard-size guard for hot movement packets.
* `services/game-server/internal/protocol/realtime/hot_lane_size_estimate.go` - conservative compact-JSON byte estimation used by hot-lane chunk construction.
* `services/game-server/internal/protocol/realtime/hot_lane_policy.go` - hot movement lane budget and cadence thresholds.
* `services/game-server/internal/protocol/realtime/hot_lane_cohorts.go` - hot movement lane routing modes and cohort selection support.
* `services/game-server/internal/protocol/realtime/scheduler.go` - lane candidate scheduling and estimated byte-budget include/defer planning for already-built candidates; real hot-lane chunks are created before scheduling.
* `services/game-server/internal/protocol/realtime/lanes.go` - lane definitions and packet-family ownership.
* `services/game-server/internal/protocol/realtime/planner.go` - orchestrates lane candidate builder calls for world, overlay, session, and event lanes and returns candidate-construction errors.
* `services/game-server/internal/protocol/realtime/lane_candidate_world.go` - world lane candidate construction, fail-closed world quantization, world baseline/delta comparison, hot movement split integration, and world projection chaining.
* `services/game-server/internal/protocol/realtime/lane_candidate_lifecycle.go` - asteroid and bullet lifecycle candidate construction for dedicated reliable lifecycle lanes.
* `services/game-server/internal/protocol/realtime/lane_candidate_overlay.go` - overlay lane full/delta candidate construction with fail-closed quantization.
* `services/game-server/internal/protocol/realtime/lane_candidate_session.go` - session lane full/delta candidate construction with fail-closed quantization.
* `services/game-server/internal/protocol/realtime/lane_candidate_event.go` - event_batch candidate construction without draining pending events.
* `services/game-server/internal/protocol/realtime/candidate_types.go` - realtime lane candidate and send-preparation types.
* `services/game-server/internal/protocol/realtime/payload.go` - typed candidate payload contract, compile-time coverage, and candidate constructors.
* `services/game-server/internal/protocol/realtime/payload_validation.go` - supported payload matrix, concrete registry, and payload invariant validation.
* `services/game-server/internal/protocol/realtime/payload_world.go`, `payload_overlay.go`, `payload_session.go`, and `payload_hot_event.go` - lane-specific typed payload implementations.
* `services/game-server/internal/protocol/realtime/candidate_policy.go` - delivery-class, priority, schedule-record, and candidate projection helpers; packet-family identity remains payload-owned.
* `services/game-server/internal/protocol/realtime/candidate_diagnostics.go` - candidate write diagnostics used by active lane metric/debug records.
* `services/game-server/internal/protocol/realtime/quantize_overlay.go` and `services/game-server/internal/protocol/realtime/quantize_session.go` - overlay and session full-packet wire quantization.
* `services/game-server/internal/protocol/realtime/wire_packets.go` - readable wire-map construction and sparse delta omission.
* `shared/packets/realtime_wire.toml` - physical compact-wire contract source of truth.
* `services/game-server/internal/protocol/realtimewire/generated.go` - generated server descriptor data.
* `services/game-server/internal/protocol/realtime/compact_wire_packet.go` - public compact encode boundary and generic recursive fallback.
* `services/game-server/internal/protocol/realtime/compact_wire_descriptor.go` - generic application of generated packet bindings and record encodings.
* `services/game-server/internal/protocol/realtime/active.go` - active lane packet encoding path, compact/packetcodec boundary, and encoded-byte accounting. It does not reject already-scheduled hot packets for size.
* `services/game-server/internal/protocol/realtime/quantize/` - numeric quantization algorithms; generated descriptors own field-path policy assignments.
* `services/game-server/internal/protocol/realtime/quantize_world.go` - world lane quantization projection.
* `services/game-server/internal/protocol/realtime/quantized_records.go` - quantized wire record types.
* `services/game-server/internal/protocol/realtime/baseline.go` - successful-candidate metadata/projection commit, including final-chunk-only hot projection commits and world-full hot projection seeding.
* `services/game-server/internal/networking/websocket_write.go` - session write-loop integration, per-lane-group preflight, independent hot-lane commits, and the atomic reliable world/lifecycle send group.
* `services/game-server/internal/networking/webrtc_transport.go` - active gameplay lane transport delivery over ordered/reliable world, overlay, session, event, ship lifecycle, asteroid lifecycle, and bullet lifecycle channels plus unordered/unreliable ship, asteroid, and bullet hot-update channels.
* `services/game-server/internal/networking/packetmetrics/` - packet observability helpers and related support types used by outbound networking seams.
* `services/game-server/internal/networking/` - websocket session, WebRTC transport, and outbound delivery boundaries.
* `shared/packets/gameplay.toml` - shared gameplay schema and realtime packet type values.
* `shared/packets/outputs.toml` - generated output routing for packet constants and builders.
* `docs/protocol/generated/realtime-wire-reference.md` - generated physical contract reference.

## Tests

Relevant server tests include:

* `services/game-server/internal/protocol/realtime/*_test.go` - lane-native realtime projection coverage, including sparse delta serialization, lifecycle candidate routing/planning, and wire-map omission behavior.
* `services/game-server/internal/protocol/realtime/quantization_propagation_test.go` - exported planner and active-result coverage for surfaced world, overlay, and session quantization failures.
* `services/game-server/internal/networking/websocket_write_test.go`
* `services/game-server/internal/networking/room_snapshot_test.go`
* `services/game-server/internal/networking/room_error_test.go`
* `services/game-server/internal/networking/session_auth_test.go`
* `services/game-server/internal/networking/player_pause_state_test.go`
* `services/game-server/internal/networking/outbound/debug_status_presentation_test.go`
* `services/game-server/internal/networking/outbound/debug_shape_catalog_presentation_test.go`

## Related docs

* [Gameplay packets](../../../../protocol/gameplay-packets.md)
* [Realtime WebSocket Protocol](../../../../protocol/realtime-websocket-protocol.md)
* [Game Server Simulation Runtime](./!INDEX.md)
* [Presentation Event Queue](./presentation-event-queue.md)
* [Packet Schemas](../../../../data/packet-schemas.md)
* [Realtime WebRTC Gameplay Transport](../../../../protocol/realtime-webrtc-gameplay-transport.md)

## Notes
