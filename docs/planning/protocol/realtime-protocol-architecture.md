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

This planning doc keeps the remaining architecture boundary for bit packing, protobuf or future binary representation, deeper record/entity-level prioritization, interest management, deeper packet-budget policy beyond current candidate-level send-plan selection, resync hardening, transport evolution beyond the current mixed lane policy, and future protocol compatibility/versioning. JSON alias compaction, sparse delta serialization, tuple packing, physical gameplay DataChannels, subtractive asteroid/bullet movement lanes, focused asteroid/bullet hot-lane chunking, chunker-owned hot-lane hard-size guarding, and dedicated reliable ordered lifecycle lanes are already implemented for active realtime gameplay and are documented in [Realtime WebSocket Protocol](../../protocol/realtime-websocket-protocol.md) and [Realtime Compact Wire Mapping](../../services/game-server/networking/realtime-compact-wire-mapping.md). Dedicated reliable ordered lifecycle lanes are implemented for asteroid and bullet/projectile creates/deletes: `sr.asteroids.lifecycle` and `sr.bullets.lifecycle`.

Focused asteroid/bullet hot-lane chunking now uses conservative compact-JSON byte estimation for chunk construction. The chunker is the hard-size guard for hot movement packets; scheduler and active encoding must not duplicate that guard.

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

## Tooling Transport Foundation And Future Consumers

The `sr.tooling` transport foundation is implemented as the ninth mandatory negotiated DataChannel for every gameplay connection. It uses id 9, is reliable, ordered, bidirectional, and must be ready with the eight gameplay channels for the room/game session. It is transport-only: no tooling packet messages or consumers exist yet, and WebSocket retains auth, signaling, lobby, session/control setup, and existing devtools/admin traffic.

Runtime measurement is the first planned `sr.tooling` consumer. Existing WebSocket devtools traffic migrates later. Consumer policy, attachment modes, permissions, and migration sequencing remain planning work in [Devtools And Telemetry](../devtools/devtools-and-telemetry.md).

## Related Docs

- [Planning](../!INDEX.md)
- [Realtime WebSocket Protocol](../../protocol/realtime-websocket-protocol.md)
- [Gameplay Packets](../../protocol/gameplay-packets.md)
- [Outbound Message Flow](../../services/game-server/networking/outbound-message-flow.md)
- [Inbound Packet Routing](../../services/client/networking-flow/inbound-packet-routing.md)
- [Gameplay State Application](../../services/client/gameplay-runtime/gameplay-state-application.md)
- [Lane Packet Projection](../../services/game-server/simulation/runtime/lane-packet-projection.md)
- [Packet Schemas](../../data/packet-schemas.md)
- [Spatial Query Index](../../services/game-server/simulation/world/spatial-query-index.md)
- [Network Observability And Packet Budget](../domains/technical/network-observability-and-packet-budget.md)
- [Testing And Smoke Strategy](../domains/technical/verification-and-quality-gates.md)
- [Devtools And Telemetry](../devtools/devtools-and-telemetry.md)
- [Development Roadmap](../development-roadmap.md)

## Open Planning Questions

- Which packet-budget policy changes require protocol-version compatibility?
- Which resync hardening behaviors should be treated as mandatory versus optional?
- What additional physical gameplay lane/channel evolution is worth planning beyond the current mixed lane policy?

## Phase P2 - Realtime Protocol Architecture

Lane-native JSON gameplay delivery over ordered/reliable `sr.world`, `sr.overlay`, `sr.session`, `sr.event`, `sr.asteroids.lifecycle`, and `sr.bullets.lifecycle` lanes, plus unordered/unreliable `sr.asteroids` and `sr.bullets` hot-update lanes, is implemented, and this doc now tracks the remaining protocol evolution after that cutover.

## Implemented Status

- Lane-scoped runtime packets exist.
- The combined `state` runtime delivery path is removed.
- Server and client `protocol/realtime` packages exist.
- Outbound delivery and realtime policy are separate.
- Lane baselines, deltas, sequence metadata, metrics, and shadow/parity support exist at the current implementation level.
- Delta comparison decides what changed; candidate-level scheduling and estimated byte-budget selection decide which lane candidates fit first.
- Oversized `asteroid_delta` and `bullet_delta` movement update lists are split into real same-sequence hot-lane candidate chunks before scheduling and encoding.
- Active output can emit multiple encoded packets on `sr.asteroids` or `sr.bullets` in one tick.
- Hot asteroid and bullet movement packets have focused candidate-level chunking before scheduling and encoding; that chunker is the only hard-size guard for those hot movement packets.
- Client hot-lane guards accept distinct valid chunks for each `asteroid_delta` or `bullet_delta` lane sequence when `chunk_count` matches, and reject duplicates, malformed/inconsistent chunk metadata, and lower sequences; gaps remain valid and the two lanes track independently.
- Independent per-session hot movement cadence enforcement is implemented: `HotLaneTick` runs at the active 60 Hz build cadence; asteroid movement emits at 60 Hz when unchunked and 30 Hz when chunking is required, while bullet movement emits at 60 Hz for one chunk, 30 Hz for two chunks, and 20 Hz for three or more chunks. Forced sends bypass cadence suppression.
- Non-hot world and asteroid/bullet lifecycle changes force related hot movement emission and advance the shared world projection chain; this is implemented protocol/realtime policy, not future transport work.
- Record/entity-level prioritization remains future work.
- High-frequency gameplay state is no longer sent as one full combined packet every tick.
- Field-delta update maps are implemented for world ship/pickup updates and dedicated asteroid/bullet movement updates.
- Field-delta update maps are implemented for overlay receiver updates.
- Field-delta update maps are implemented for session player and lifecycle updates.
- Creates remain full records, updates carry identity plus changed fields only, and deletes remain identity lists.
- Realtime numeric wire quantization is implemented for outbound lane projection.
- Compact JSON aliasing is implemented for active realtime gameplay lanes.
- Asteroid tuple packing is implemented for compact asteroid lifecycle/full-bootstrap records, with hot movement tuples limited to movement updates only.
- Bullet tuple packing is implemented for compact bullet lifecycle/full-bootstrap records, with hot movement tuples limited to movement updates only.
- World ship/player tuple packing is implemented for compact world lane ship and player records.
- Session player and lifecycle tuple packing is implemented for compact session lane records.
- Known event tuple packing is implemented for compact `event_batch` records.
- Sparse delta serialization is implemented for active realtime gameplay delta lanes; empty delta sections are omitted from emitted delta wire maps.
- Client lifecycle application validates explicit world-baseline dependencies, queues future or not-yet-active lifecycle packets, drains them after matching `world_full` activation, and enforces strict independent lifecycle sequences.
- The `sr.tooling` transport foundation is implemented at negotiated id 9 with reliable, ordered, bidirectional delivery and mandatory readiness alongside the eight gameplay channels. Lane-aware receive routing separates tooling before normal gameplay dispatch; tooling protocol messages and consumers remain future work.
- Unexpected required-channel close recovery preserves the WebSocket/session/room/game context, replaces only the WebRTC peer, and uses a 10-second deadline. Successful recovery preserves the active match and requests fresh world, overlay, and session baselines; failure disables only single-player replay.

Current implementation details live in:

- [Realtime WebSocket Protocol](../../protocol/realtime-websocket-protocol.md)
- [Gameplay Packets](../../protocol/gameplay-packets.md)
- [Outbound Message Flow](../../services/game-server/networking/outbound-message-flow.md)
- [Inbound Packet Routing](../../services/client/networking-flow/inbound-packet-routing.md)
- [Gameplay State Application](../../services/client/gameplay-runtime/gameplay-state-application.md)
- [Lane Packet Projection](../../services/game-server/simulation/runtime/lane-packet-projection.md)
- [Packet Schemas](../../data/packet-schemas.md)

## Remaining Protocol Evolution

Future planning here remains focused on bit packing, protobuf or custom binary representation, deeper record/entity-level prioritization, interest management, deeper packet-budget behavior beyond current candidate-level send-plan selection, stronger resync behavior, future physical lane/channel evolution beyond the current mixed lane policy, and future compatibility/versioning. JSON alias compaction, sparse delta serialization, tuple packing, physical gameplay DataChannels, subtractive asteroid/bullet movement lanes, focused asteroid/bullet hot-lane chunking, and chunker-owned hot-lane hard-size guarding are already implemented for active realtime gameplay lanes and are documented in [Realtime WebSocket Protocol](../../protocol/realtime-websocket-protocol.md) and [Realtime Compact Wire Mapping](../../services/game-server/networking/realtime-compact-wire-mapping.md).

### Remaining Priority And Packet Budget Work

Delta decides what changed. The current candidate-level send plan uses the estimated packet budget as an advisory candidate-selection target; it does not imply that every included hot chunk collectively fits under one per-tick budget. The hot-lane chunker separately enforces the per-message hard-size threshold before scheduling, and aggregate same-tick output is not currently capped. Focused hot-lane chunking is implemented for asteroid and bullet movement lists. General record/entity-level prioritization, interest filtering, and deeper budget policy remain future work.

Current implementation has lane-native packets, baselines, deltas, candidate-level scheduling metadata, estimated byte-budget selection, focused asteroid/bullet hot-lane chunking, and chunker-owned hot-lane hard-size guarding. Delta decides what changed; the current send plan decides which lane candidates are included or deferred; future work remains around record/entity-level prioritization and deeper budget policy. Scheduler and active encoding do not own a second hard-size rejection step for already-chunked hot movement packets.
Current WebRTC physical gameplay channels are split into reliable/ordered lanes (`sr.world`, `sr.overlay`, `sr.session`, `sr.event`, `sr.asteroids.lifecycle`, `sr.bullets.lifecycle`) and unordered/unreliable hot-update lanes (`sr.asteroids`, `sr.bullets`); the mandatory reliable/ordered/bidirectional `sr.tooling` transport lane is negotiated at id 9. Client-side hot-lane guards accept distinct valid chunks for each `asteroid_delta` or `bullet_delta` lane sequence when `chunk_count` matches, and reject duplicates, malformed/inconsistent chunk metadata, and lower sequences; gaps remain valid and the two lanes track independently.

Reliability and ordering remain per DataChannel; they do not establish cross-channel `sr.world`/lifecycle ordering. The implemented client lifecycle gate handles that arrival race by waiting for the referenced active world baseline.

### Future Interest Management Boundary

The game server now has a generic toroidal `spatial.Index` contract with circular and rectangular queries, backed by the current uniform-grid implementation. This foundation currently supports simulation collision broad phase only; record/entity prioritization and interest filtering remain future work. See [Spatial Query Index](../../services/game-server/simulation/world/spatial-query-index.md).

The active collision index is mutable `Game`-owned simulation state. Realtime projection must not read it directly. When interest management is implemented, presentation-frame publication may build or carry a separate immutable presentation-owned spatial snapshot or index once per presentation generation. `protocol/realtime` would then own receiver-interest policy over that immutable presentation input, without moving interest policy into `game/spatial`.

Future policy may include viewport margin or hysteresis, always-required entities, create/delete transitions, and consistent filtering across full, delta, lifecycle, and hot lanes. None of those receiver-interest sets or packet filters is implemented by this foundation.

Field-delta update maps are now implemented, sparse delta serialization is already in place for the active realtime gameplay lanes, and JSON alias compaction is already in place. Asteroid, bullet, world ship/player, session player/lifecycle, and known event tuple packing are implemented for compact current lane records. Regular asteroid and bullet movement updates are now subtractively split out of `sr.world` into dedicated hot movement packets. High-density stress cases can still exceed future packet-budget targets even after quantization, compact aliases, sparse deltas, tuple packing, and hot movement lanes; remaining work belongs to packet-size verification with stress logs, deeper record/entity-level prioritization, further transport policy beyond the current asteroid/bullet unordered hot lanes where safe, and binary representation later.

State lanes are quantized during outbound projection before delta comparison.
Presentation-event records are quantized during explicit event wire shaping.
The current quantization contract is described in [Realtime WebSocket Protocol](../../protocol/realtime-websocket-protocol.md), and projection ownership lives in [Lane Packet Projection](../../services/game-server/simulation/runtime/lane-packet-projection.md).

Future planning targets remain:

- deeper packet-budget policy beyond current candidate-level send-plan selection
- record/entity-level prioritization
- interest filtering
- stronger resync behavior
- hot/cold lane separation beyond current asteroid/bullet movement extraction

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

Future packetcodec and transport evolution must preserve these ownership seams. The current baseline includes ordered/reliable lanes for world, overlay, session, event, asteroid lifecycle, and bullet lifecycle traffic, plus unordered/unreliable hot lanes for asteroid and bullet movement traffic.
## Notes

The planning sections above intentionally avoid duplicating the runtime manuals in the implementation docs.
