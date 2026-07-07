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

This planning doc keeps the remaining architecture boundary for bit packing, protobuf or future binary representation, deeper record/entity-level prioritization, interest management, deeper packet-budget policy beyond current candidate-level send-plan selection and hot-packet encoded-size guards, resync hardening, transport evolution beyond the current mixed lane policy, and future protocol compatibility/versioning. JSON alias compaction, sparse delta serialization, tuple packing, physical gameplay DataChannels, and subtractive asteroid/bullet movement lanes are already implemented for active realtime gameplay lanes and are documented in [Realtime WebSocket Protocol](../../protocol/realtime-websocket-protocol.md) and [Realtime Compact Wire Mapping](../../services/game-server/networking/realtime-compact-wire-mapping.md).

## Current Inputs

Planning inputs for the remaining protocol work:

- current protocol implementation docs
- packet and state schema constraints
- server projection and outbound flow constraints
- client inbound routing and state application constraints
- compatibility and versioning requirements
- current WebRTC baseline and future channel evolution assumptions
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
- What physical lane/channel evolution is worth planning beyond the current reliable/ordered WebRTC physical gameplay lane baseline?

## Phase P2 - Realtime Protocol Architecture

Lane-native JSON gameplay delivery over ordered/reliable `sr.world`, `sr.overlay`, `sr.session`, and `sr.event` lanes, plus unordered/unreliable `sr.asteroids` and `sr.bullets` hot-update lanes, is implemented, and this doc now tracks the remaining protocol evolution after that cutover.

## Implemented Status

- Lane-scoped runtime packets exist.
- The combined `state` runtime delivery path is removed.
- Server and client `protocol/realtime` packages exist.
- Outbound delivery and realtime policy are separate.
- Lane baselines, deltas, sequence metadata, metrics, and shadow/parity support exist at the current implementation level.
- Delta comparison decides what changed; candidate-level scheduling and estimated byte-budget selection decide which lane candidates fit first.
- Hot asteroid and bullet packets have encoded-size guards before send.
- Record/entity-level prioritization remains future work.
- High-frequency gameplay state is no longer sent as one full combined packet every tick.
- Field-delta update maps are implemented for world ship/pickup updates and dedicated asteroid/bullet movement updates.
- Field-delta update maps are implemented for overlay receiver updates.
- Field-delta update maps are implemented for session player and lifecycle updates.
- Creates remain full records, updates carry identity plus changed fields only, and deletes remain identity lists.
- Realtime numeric wire quantization is implemented for outbound lane projection.
- Compact JSON aliasing is implemented for active realtime gameplay lanes.
- Asteroid tuple packing is implemented for compact world lane asteroid full/create/update/delete records.
- Bullet tuple packing is implemented for compact world lane bullet records.
- World ship/player tuple packing is implemented for compact world lane ship and player records.
- Session player and lifecycle tuple packing is implemented for compact session lane records.
- Known event tuple packing is implemented for compact `event_batch` records.
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

Future planning here remains focused on bit packing, protobuf or custom binary representation, deeper record/entity-level prioritization, interest management, deeper packet-budget behavior beyond current candidate-level send-plan selection and hot-packet encoded-size guards, stronger resync behavior, future physical lane/channel evolution beyond the current reliable/ordered WebRTC lane baseline, and future compatibility/versioning. JSON alias compaction, sparse delta serialization, tuple packing, physical gameplay DataChannels, and subtractive asteroid/bullet movement lanes are already implemented and are documented in [Realtime WebSocket Protocol](../../protocol/realtime-websocket-protocol.md) and [Realtime Compact Wire Mapping](../../services/game-server/networking/realtime-compact-wire-mapping.md).

### Remaining Priority And Packet Budget Work

Delta decides what changed. The current candidate-level send plan decides which lane candidates fit the estimated packet budget first. Record/entity-level priority remains future work.

Current implementation has lane-native packets, baselines, deltas, candidate-level scheduling metadata, estimated byte-budget selection, and hot-packet encoded-size guards. Delta decides what changed; the current send plan decides which lane candidates are included or deferred; future work remains around record/entity-level prioritization and deeper budget policy.

Field-delta update maps are now implemented, sparse delta serialization is already in place for the active realtime gameplay lanes, and JSON alias compaction is already in place. Asteroid, bullet, world ship/player, session player/lifecycle, and known event tuple packing are implemented for compact current lane records. Regular asteroid and bullet movement updates are now subtractively split out of `sr.world` into dedicated hot movement packets. High-density stress cases can still exceed future packet-budget targets even after quantization, compact aliases, sparse deltas, tuple packing, and hot movement lanes; remaining work belongs to packet-size verification with stress logs, prioritization, unreliable/unordered delivery where safe, and binary representation later.

State lanes are quantized during outbound projection before delta comparison.
Presentation-event records are quantized during explicit event wire shaping.
The current quantization contract is described in [Realtime WebSocket Protocol](../../protocol/realtime-websocket-protocol.md), and projection ownership lives in [Lane Packet Projection](../../services/game-server/simulation/runtime/lane-packet-projection.md).

Future planning targets remain:

- deeper packet-budget policy beyond current candidate-level send-plan selection
- record/entity-level prioritization
- interest filtering
- stronger resync behavior
- hot/cold lane separation beyond current asteroid/bullet movement extraction
- Current WebRTC physical gameplay channels split into reliable/ordered lanes (`sr.world`, `sr.overlay`, `sr.session`, `sr.event`) and unordered/unreliable hot-update lanes (`sr.asteroids`, `sr.bullets`)
- client-side monotonic rejection already guards late hot packets

Live priority should stay conservative until required gameplay and presentation truth can be proven safe by metrics.

### Numeric Quantization Note

State lanes are quantized during outbound projection before delta comparison. Presentation-event records are quantized during explicit event wire shaping. The current quantization contract is described in [Realtime WebSocket Protocol](../../protocol/realtime-websocket-protocol.md), and projection ownership lives in [Lane Packet Projection](../../services/game-server/simulation/runtime/lane-packet-projection.md).

Keep this planning doc high-level: it tracks the remaining protocol roadmap, not field policy, code paths, or runtime behavior details.

## Outbound Collaboration

- `networking/outbound` owns delivery mechanics.
- `protocol/realtime` owns replication policy.
- `protocol/packetcodec` owns representation and encoding.

Active server outbound delivery is documented in [Outbound Message Flow](../../services/game-server/networking/outbound-message-flow.md).

Protocol and wire behavior is documented in [Realtime WebSocket Protocol](../../protocol/realtime-websocket-protocol.md).

Client inbound lane routing is documented in [Inbound Packet Routing](../../services/client/networking-flow/inbound-packet-routing.md).

Future packetcodec and transport evolution must preserve these ownership seams. The current baseline includes ordered/reliable lanes for world, overlay, session, and event traffic, plus unordered/unreliable hot lanes for asteroid and bullet movement traffic.
## Notes

The planning sections above intentionally avoid duplicating the runtime manuals in the implementation docs.




