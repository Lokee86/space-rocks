# Realtime Protocol Architecture
Parent index: [Protocol Planning](./!INDEX.md)

## Purpose

This doc tracks the remaining realtime protocol architecture work after the lane-native P2 cutover.

## Ownership Boundary

This doc owns planning for the remaining realtime protocol evolution, not the current lane implementation details.

Current implementation facts belong in the canonical protocol, service, and data docs, including:

- [Realtime WebSocket Protocol](../../protocol/realtime-websocket-protocol.md)
- [Gameplay Packets](../../protocol/gameplay-packets.md)
- [Outbound Message Flow](../../services/game-server/networking/outbound-message-flow.md)
- [Inbound Packet Routing](../../services/client/networking-flow/inbound-packet-routing.md)
- [Gameplay State Application](../../services/client/gameplay-runtime/gameplay-state-application.md)
- [Lane Packet Projection](../../services/game-server/simulation/runtime/lane-packet-projection.md)
- [Packet Schemas](../../data/packet-schemas.md)
- [Realtime Compact Wire Mapping](../../services/game-server/networking/realtime-compact-wire-mapping.md)

This planning doc keeps the remaining architecture boundary for bit packing, protobuf or future binary representation, tuple/array packing where needed, deeper prioritization, interest management, packet budget policy, resync hardening, transport evolution beyond the current WebSocket, and future protocol compatibility/versioning. JSON alias compaction and sparse delta serialization are already implemented for active realtime gameplay lanes and are documented in [Realtime WebSocket Protocol](../../protocol/realtime-websocket-protocol.md) and [Realtime Compact Wire Mapping](../../services/game-server/networking/realtime-compact-wire-mapping.md).

## Current Inputs

Planning inputs for the remaining protocol work:

- current protocol implementation docs
- packet and state schema constraints
- server projection and outbound flow constraints
- client inbound routing and state application constraints
- compatibility and versioning requirements
- transport evolution assumptions
- packet budget and prioritization requirements

## Planned Outputs

Planning outputs for the remaining protocol work:

- a sequenced roadmap for the remaining protocol architecture work
- explicit ownership for future codec, budget, resync, and transport changes
- decision points for representation, compatibility, and versioning changes
- follow-up implementation tasks that move from planning into current docs when shipped

## Related Docs

- [Planning](../!INDEX.md)
- [Realtime WebSocket Protocol](../../protocol/realtime-websocket-protocol.md)
- [Gameplay Packets](../../protocol/gameplay-packets.md)
- [Outbound Message Flow](../../services/game-server/networking/outbound-message-flow.md)
- [Inbound Packet Routing](../../services/client/networking-flow/inbound-packet-routing.md)
- [Gameplay State Application](../../services/client/gameplay-runtime/gameplay-state-application.md)
- [Lane Packet Projection](../../services/game-server/simulation/runtime/lane-packet-projection.md)
- [Packet Schemas](../../data/packet-schemas.md)
- [Network Observability And Packet Budget](../domains/technical/network-observability-and-packet-budget.md)
- [Testing And Smoke Strategy](../domains/technical/verification-and-quality-gates.md)
- [Development Roadmap](../development-roadmap.md)

## Open Planning Questions

- Which packet-budget policy changes require protocol-version compatibility?
- Which resync hardening behaviors should be treated as mandatory versus optional?
- What transport evolution is worth planning beyond the current WebSocket path?

## Phase P2 - Realtime Protocol Architecture

Lane-native JSON WebSocket delivery is implemented, and this doc now tracks the remaining protocol evolution after that cutover.

## Implemented Status

- Lane-scoped runtime packets exist.
- The combined `state` runtime delivery path is removed.
- Server and client `protocol/realtime` packages exist.
- Outbound delivery and realtime policy are separate.
- Lane baselines, deltas, sequence metadata, metrics, and shadow/parity support exist at the current implementation level.
- Delta comparison decides what changed; priority and budget decisions still decide which changed data fits first.
- High-frequency gameplay state is no longer sent as one full combined packet every tick.
- Field-delta update maps are implemented for world entity updates.
- Field-delta update maps are implemented for overlay receiver updates.
- Field-delta update maps are implemented for session player and lifecycle updates.
- Creates remain full records, updates carry identity plus changed fields only, and deletes remain identity lists.
- Realtime numeric wire quantization is implemented for outbound lane projection.
- Compact JSON aliasing is implemented for active realtime gameplay lanes.
- Sparse delta serialization is implemented for active realtime gameplay delta lanes; empty delta sections are omitted from emitted delta wire maps.

Current implementation details live in:

- [Realtime WebSocket Protocol](../../protocol/realtime-websocket-protocol.md)
- [Gameplay Packets](../../protocol/gameplay-packets.md)
- [Outbound Message Flow](../../services/game-server/networking/outbound-message-flow.md)
- [Inbound Packet Routing](../../services/client/networking-flow/inbound-packet-routing.md)
- [Gameplay State Application](../../services/client/gameplay-runtime/gameplay-state-application.md)
- [Lane Packet Projection](../../services/game-server/simulation/runtime/lane-packet-projection.md)
- [Packet Schemas](../../data/packet-schemas.md)

## Remaining Protocol Evolution

Future planning here remains focused on bit packing, protobuf or custom binary representation, tuple/array packing where needed, deeper prioritization, interest management, packet budget behavior, stronger resync behavior, transport evolution beyond WebSocket, and future compatibility/versioning. JSON alias compaction and sparse delta serialization are already implemented and are documented in [Realtime WebSocket Protocol](../../protocol/realtime-websocket-protocol.md) and [Realtime Compact Wire Mapping](../../services/game-server/networking/realtime-compact-wire-mapping.md).

### Remaining Priority And Packet Budget Work

Delta decides what changed. Priority decides which changed data fits the packet budget first.

Current implementation has lane-native packets, baselines, deltas, and candidate-level scheduling metadata. Delta decides what changed; priority decides which changed data fits the packet budget first.

Field-delta update maps are now implemented, sparse delta serialization is already in place for the active realtime gameplay lanes, and JSON alias compaction is already in place. High-density world stress cases can still exceed current packet budget even after quantization, compact aliases, and sparse deltas; remaining work belongs to prioritization, packing, and binary representation.

Server-owned outbound realtime lane state projection quantizes float-like values before delta comparison and JSON encoding, without truncating authoritative simulation state.

Future planning targets remain:

- packet budget selection
- record/entity-level prioritization
- interest filtering
- stronger resync behavior
- hot/cold lane separation
- transport evolution beyond current WebSocket

Live priority should stay conservative until required gameplay and presentation truth can be proven safe by metrics.

### Numeric Quantization Note

Server-owned outbound realtime lane state projection quantizes float-like values before delta comparison and JSON encoding. The current quantization contract is described in [Realtime WebSocket Protocol](../../protocol/realtime-websocket-protocol.md), and projection ownership lives in [Lane Packet Projection](../../services/game-server/simulation/runtime/lane-packet-projection.md).

Keep this planning doc high-level: it tracks the remaining protocol roadmap, not field policy, code paths, or runtime behavior details.

## Outbound Collaboration

- `networking/outbound` owns delivery mechanics.
- `protocol/realtime` owns replication policy.
- `protocol/packetcodec` owns representation and encoding.

Active server outbound delivery is documented in [Outbound Message Flow](../../services/game-server/networking/outbound-message-flow.md).

Protocol and wire behavior is documented in [Realtime WebSocket Protocol](../../protocol/realtime-websocket-protocol.md).

Client inbound lane routing is documented in [Inbound Packet Routing](../../services/client/networking-flow/inbound-packet-routing.md).

Future packetcodec and transport evolution must preserve these ownership seams.
## Notes

The planning sections above intentionally avoid duplicating the runtime manuals in the implementation docs.


