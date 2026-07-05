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
-> lane candidate selection and delta comparison
-> sparse readable wire-map serialization
-> raw-float assertion for active world/overlay/session wire maps
-> compact alias mapping
-> packetcodec JSON encoding
-> encoded-byte accounting
-> networking write integration
-> debug wire/summary logging after successful writes
-> WebRTC sr.reliable write
```

Projection is lane-specific rather than one combined gameplay snapshot.

## Code root

```text
services/game-server/internal/protocol/realtime/
services/game-server/internal/networking/websocket_write.go
services/game-server/internal/networking/webrtc_transport.go
```

The realtime package owns candidate construction, send-plan records, metadata, wire packet assembly, numeric wire quantization, delta comparison, sparse omission, compact alias preparation, and encoded-byte accounting inputs. The session write loop owns tick-driven invocation; active gameplay lane delivery uses WebRTC `sr.reliable` through the networking transport seam. Networking owns successful delivery handling, post-write state changes, and the current successful-write debug wire/summary logging. For `event_batch`, the realtime package shapes sparse event-type-specific wire records rather than broad reflected `EventState` output.

## Responsibilities

The active server projection path owns:

* Projecting authoritative runtime state into lane-native packet families.
* Keeping world, overlay, session, and event ownership separate.
* Producing receiver-specific overlay/session/event output where needed.
* Preserving explicit event-batch drain semantics.
* Leaving JSON encode/decode mechanics to packetcodec and WebRTC active gameplay transport/write success handling to networking.

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

It does not own packet schema source files or generated constants. Runtime wire-map behavior for active realtime lanes lives in protocol/realtime and is specified by the realtime protocol docs; packet schema docs own generated schema inputs and outputs.

## Lane ownership

Current gameplay presentation ownership is split as:

```text
world lane
= active entity presentation state such as ships, asteroids, bullets, pickups

overlay lane
= local-player presentation facts such as lives, score, loadout, cooldown-facing HUD facts

session lane
= durable match-local player session state and lifecycle-oriented read models

event_batch
= transient presentation events sent separately from baseline/delta lanes
```

`player_pause_state` remains a separate same-session packet and is handled independently from lane-native realtime projection.

## Delta projection behavior

The realtime projection path builds lane records from authoritative game state before delta comparison.

Field-delta comparison is current behavior for these update groups:

```text
world lane
= ship, bullet, asteroid, and pickup updates

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
-> raw-float assertion checks active world/overlay/session wire maps
-> packetcodec encodes JSON
-> CompactWirePacket applies aliases, compact values, shared ID compaction, and tuple packing to remaining emitted keys
```

Sparse omission is a realtime wire-map serialization concern. Compact aliasing is a realtime encode-boundary mapping concern. `packetcodec` only encodes the already-shaped map to JSON. Networking only writes encoded bytes after realtime builds them. Full lane-native realtime packets remain complete snapshots. Delta create, update, and delete sections are omitted when empty. Clients treat missing delta sections as empty or no-op, and missing fields inside update records remain unchanged, not cleared. Sparse omission must not drop meaningful `false` or `0` values inside present records.

Implementation ownership for this behavior lives in `services/game-server/internal/protocol/realtime/wire_packets.go`, `services/game-server/internal/protocol/realtime/quantize_world.go`, and `services/game-server/internal/protocol/realtime/quantize/`.

Numeric wire quantization is implemented in the realtime projection and wire-record path before delta comparison. The active server implementation uses `services/game-server/internal/protocol/realtime/quantize/` and `services/game-server/internal/protocol/realtime/quantize_world.go` as the quantization boundary for outbound lane projection. It should not truncate authoritative simulation state for packet-size savings.

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

* `services/game-server/internal/protocol/realtime/` - lane candidates, metadata, send-plan records, baseline/delta planning, wire packets, sparse omission, compact alias preparation, encoded-byte accounting inputs, and shadow/parity helpers.
* `services/game-server/internal/protocol/realtime/wire_packets.go` - readable wire-map construction and sparse delta omission.
* `services/game-server/internal/protocol/realtime/compact_wire_packet.go` - compact alias mapping for emitted active lane keys.
* `services/game-server/internal/protocol/realtime/active.go` - active lane packet encoding path and raw-float assertion/compact/packetcodec boundary.
* `services/game-server/internal/protocol/realtime/quantize/` - numeric wire quantization policies.
* `services/game-server/internal/protocol/realtime/quantize_world.go` - world lane quantization projection.
* `services/game-server/internal/protocol/realtime/quantized_records.go` - quantized wire record types.
* `services/game-server/internal/networking/websocket_write.go` - session write-loop integration and active write triggering.
* `services/game-server/internal/networking/webrtc_transport.go` - active gameplay lane transport delivery over WebRTC `sr.reliable`.
* `services/game-server/internal/networking/packetmetrics/` - packet observability helpers and related support types used by outbound networking seams.
* `services/game-server/internal/networking/` - websocket session, WebRTC transport, and outbound delivery boundaries.
* `shared/packets/gameplay.toml` - shared gameplay schema and realtime packet type values.
* `shared/packets/outputs.toml` - generated output routing for packet constants and builders.

## Tests

Relevant server tests include:

* `services/game-server/internal/protocol/realtime/*_test.go` - lane-native realtime projection coverage, including sparse delta serialization and wire-map omission behavior.
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

## Notes
