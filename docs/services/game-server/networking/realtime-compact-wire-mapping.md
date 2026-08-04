---
author: brian
created: "2026-07-19"
document_id: 019f7d55-fb2c-7484-bb8d-bef7e73827f7
document_type: general
policy_exempt: false
summary: This document explains realtime compact-wire architecture and ownership only. The authoritative physical contract is shared/packets/realtimewire.toml. Generated tables are in docs/protocol/generated/realtime-wire-reference.md.
---
# Realtime Compact Wire Mapping

Parent index: [Game Server Networking](./!INDEX.md)

## Purpose

This document explains realtime compact-wire architecture and ownership only. The authoritative physical contract is `shared/packets/realtime_wire.toml`. Generated tables are in `docs/protocol/generated/realtime-wire-reference.md`.

Do not reproduce aliases, tuple orders, event layouts, ID tags, quantization tables, or packet metadata here. Those details belong to the contract and generated reference.

## Overview

This document describes the current Realtime Compact Wire Mapping behavior, ownership boundaries, state flow, failure behavior, implementation owners, and verification surfaces.

## Ownership

- Logical packet types, struct fields, and JSON names belong to the configured `shared/packets/*.toml` packet schema.
- The physical compact-wire contract belongs to `shared/packets/realtime_wire.toml`.
- Generated consumers are:
  - `services/game-server/internal/protocol/realtimewire/generated.go`
  - `client/scripts/generated/networking/realtime_wire_generated.gd`
  - `shared/packets/generated/realtime_wire.json`
  - `docs/protocol/generated/realtime-wire-reference.md`
- Runtime algorithms apply generated descriptors; they do not own contract data.

## Server Encode Flow

```text
RealtimeLaneCandidate -> typed RealtimeLanePayload validation -> payload WirePacket serializer -> readable map -> CompactWirePacket -> compactWirePacketFromDescriptors -> packetcodec JSON transport
```

- `payload.go`, `payload_validation.go`, and the lane-specific `payload_*.go` files own the typed payload contract, supported concrete-value registry/matrix, candidate construction, and per-payload serializer dispatch.
- `wire_packets.go` and `wire_reflect.go` provide readable-map and generic record-shaping support; they do not own candidate-family dispatch.
- `compact_wire_packet.go` is the public encode boundary and provides the generic recursive fallback.
- `compact_wire_descriptor.go` applies generated packet bindings, records, aliases, IDs, selectors, event layouts, and value domains.
- Quantization algorithms remain in the `quantize` package; path assignments come from generated descriptors.

Reflection uses JSON tags when present and `snake_case` only as a compatibility fallback.

## Client Decode Flow

```text
JSON -> PacketCodec.decode -> CompactLanePacket -> CompactWireDescriptorDecoder -> readable dictionary -> RealtimeQuantize -> RealtimePacketPipeline -> RealtimeRouter/appliers
```

- `compact_lane_packet.gd` is a small public facade.
- The descriptor index, IDs, records, and decoder apply generated data.
- Generated quantization paths drive `realtime_quantize.gd`; the formulas remain handwritten.
- The pipeline accepts schema-registered packets. Router match branches are explicit application behavior.

## Compatibility Policy

The current schema-declared behavior is:

- Readable long keys are accepted.
- Explicit metadata is accepted.
- Unknown keys and unknown event records pass through.
- Runtime metadata inference is packet-descriptor driven.
- Lifecycle packets retain explicit metadata.
- Lifecycle creates and deletes encode in the current primary map/readable-ID shapes, while tuple/numeric forms are decode-only compatibility alternatives declared by `decode_record_ids`.
- Compatibility policy is contract data and must not be added as hidden entity-specific branches.

## Contract Versus Algorithms

Schema owns:

- aliases, packet/lane/snapshot metadata, record encodings, tuple/sparse order, and bindings;
- quantization assignments, ID codecs/selectors, events, decode alternatives, and compatibility flags.

Runtime owns:

- projection from game state and reflection mechanics;
- tuple execution algorithms and ID parsing/formatting algorithms;
- quantization math and metadata reconstruction algorithms;
- transport, scheduling, baseline, and application behavior.

## Adding A Packet Family

1. Add logical packets and structs in the packet TOML.
2. Add physical declarations in `realtime_wire.toml`.
3. Add a concrete `RealtimeLanePayload` implementation, compile-time interface assertion, validation matrix/registry entry, serializer, and focused invariant tests.
4. Run validation and generation.
5. Add server projection/planning and client application behavior.
6. Add fixtures and tests.

Generic codec and generator source must not receive an entity-specific branch. The isolated extensibility proof is `tools/data_sync/tests/test_realtime_wire_enemy_extensibility.py`.

## Generation And Validation

```text
data-sync -validate -realtime-wire
data-sync -diff -realtime-wire -go -gds -json -docs
data-sync -push -realtime-wire -go -gds -json -docs
data-sync -check -realtime-wire -go -gds -json -docs
```

Top-level `data-sync -validate` also validates the realtime-wire domain.

## Code Paths

- `shared/packets/realtime_wire.toml`
- `tools/data_sync/data_sync/model/realtime_wire.py`
- `tools/data_sync/data_sync/realtime_wire_toml.py`
- `tools/data_sync/data_sync/realtime_wire_validate.py`
- `tools/data_sync/data_sync/realtime_wire_sync.py`
- `tools/data_sync/data_sync/generators/realtime_wire_*.py`
- `services/game-server/internal/protocol/realtimewire/generated.go`
- `services/game-server/internal/protocol/realtime/payload.go`
- `services/game-server/internal/protocol/realtime/payload_validation.go`
- `services/game-server/internal/protocol/realtime/payload_*.go`
- `services/game-server/internal/protocol/realtime/wire_packets.go`
- `services/game-server/internal/protocol/realtime/wire_reflect.go`
- `services/game-server/internal/protocol/realtime/compact_wire_packet.go`
- `services/game-server/internal/protocol/realtime/compact_wire_descriptor.go`
- `services/game-server/internal/protocol/realtime/quantize/`
- `client/scripts/generated/networking/realtime_wire_generated.gd`
- `client/scripts/protocol/realtime/compact_lane_packet.gd`
- `client/scripts/protocol/realtime/compact_wire_descriptor_index.gd`
- `client/scripts/protocol/realtime/compact_wire_descriptor_ids.gd`
- `client/scripts/protocol/realtime/compact_wire_descriptor_records.gd`
- `client/scripts/protocol/realtime/compact_wire_descriptor_decoder.gd`
- `client/scripts/protocol/realtime/realtime_quantize.gd`

## Tests

- Shared fixtures directory.
- Server `realtime_wire_fixture_test.go` and `compact_wire_descriptor_test.go`.
- Client `test_realtime_wire_fixtures.gd`, descriptor ID/record/decoder tests, compact lane packet tests, and generated quantization tests.
- Data-sync realtime-wire tests and `test_realtime_wire_enemy_extensibility.py`.

## Related docs

- [Game Server Networking](./!INDEX.md)

## Notes

Changes to this boundary should update its canonical owner, code map or source map, verification evidence, and related documentation in the same change.
