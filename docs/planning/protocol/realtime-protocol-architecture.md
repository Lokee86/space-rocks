# Realtime Protocol Architecture
Parent index: [Protocol Planning](./!INDEX.md)

## Purpose

This doc plans the future protocol seam for authoritative realtime state delivery.

## Ownership Boundary

This doc owns planning for `protocol/realtime`, `protocol/packetcodec`, and the split with `networking/outbound`.

It should cover lanes, full snapshots, delta snapshots, quantization, bit packing, and protobuf as a later target.

## Current Inputs

- `protocol/realtime`
- `protocol/packetcodec`
- `networking/outbound`
- lane inputs
- full snapshot inputs
- delta snapshot inputs
- quantization inputs
- bit packing inputs
- protobuf later-target inputs

## Planned Outputs

- protocol ownership boundaries
- lane vocabulary for future realtime delivery
- the preferred sequence of protocol upgrades

## Related Docs

- [Planning](../../!INDEX.md)
- [Network Observability And Packet Budget](../domains/technical/network-observability-and-packet-budget.md)
- [Testing And Smoke Strategy](../domains/technical/stubs/testing-and-smoke-strategy.md)
- [Development Roadmap](../development-roadmap.md)

## Open Planning Questions

- Which lane types need to exist before snapshot work starts?
- Which compression or packing ideas belong before protobuf is considered?
- Which responsibilities stay in outbound delivery mechanics versus protocol policy?

## Phase B - Realtime Protocol Architecture

Phase B establishes the end-state realtime protocol seam for authoritative multiplayer state delivery. Phase B replaces the current full-state-per-tick model with a governed realtime protocol boundary.

The central rule is:

- `networking/outbound` owns delivery mechanics.
- `protocol/realtime` owns delivery policy.
- `protocol/packetcodec` owns byte representation.

Server placement:

- `services/game-server/internal/protocol/realtime/`
- `services/game-server/internal/protocol/packetcodec/`
- Keep `services/game-server/internal/networking/outbound/` as delivery mechanics.

Client placement:

- `client/scripts/protocol/realtime/`
- `client/scripts/protocol/packetcodec/`
- Current client packet codec files should move from `client/scripts/networking/packets/` into `client/scripts/protocol/packetcodec/` during Phase B.

Rails/API has no realtime gameplay connection. Rails/API remains auth, account/profile, website/API, and durable player-data persistence. Realtime snapshots, deltas, lanes, replication, and transport stay between the game-server and Godot client.

### Protocol Vocabulary

- Full snapshot
- Delta snapshot
- Baseline ID
- Sequence number
- Create/update/delete records
- Lane
- Priority
- Reliability class
- Resync request
- Forced resync
- Stale update discard

### Lanes And Delivery Policy

Lanes are protocol concepts, not an immediate transport commitment.

- Reliable control lane
- Realtime state lane
- Event lane
- Slow world lane
- Debug/telemetry lane

Transport architecture is a game-time decision. The protocol lanes should be able to map later to WebSocket, WebRTC DataChannel, UDP, or a hybrid, but this phase does not need to choose transport immediately.

### Snapshot And Delta Model

- Full snapshot on join, start, or resync
- Delta snapshots after baseline
- Per-session baseline tracking
- Monotonically increasing sequence numbers
- Entity create/update/delete records
- Missing-baseline recovery
- Stale update discard
- Explicit resync path

### State Projection And Priority Policy

- Critical
- High
- Medium
- Low
- Debug

`protocol/realtime` decides what should be sent and how often, including local player state, active player session state, nearby threats, bullets and projectiles, asteroids, pickups, future enemies, future bullet hell entities, and debug-only data.

### Outbound Collaboration

- `networking/outbound` owns delivery cadence, active sessions and channels, write calls, write failures, and backpressure.
- `protocol/realtime` owns what messages are due, what state they contain, lane assignment, priority, full and delta snapshot rules, sequence and baseline rules, and resync behavior.
- `protocol/packetcodec` owns encode and decode representation only.

Flow:

1. Authoritative game state
2. `protocol/realtime` projection
3. Priority and lane policy
4. Full snapshot or delta snapshot
5. Quantized protocol message shape
6. `protocol/packetcodec` encoding
7. `networking/outbound` delivery mechanics
8. Transport channel
9. Client `protocol/realtime` application
10. World/gameplay presentation

### Quantization And Bit Packing

Quantization and bit packing come before protobuf in Phase B.

Likely candidates:

- Positions as quantized integers
- Velocities as quantized integers
- Rotation as a fixed-range integer
- Enum strings as numeric IDs
- Booleans as flags
- Runtime entity IDs as compact IDs
- Omitted default values

Quantization rules belong to the realtime protocol schema. Encoding mechanics belong to packetcodec. Gameplay continues using normal semantic values.

### Protobuf Target

Protobuf is the final step of Phase B after lanes, snapshots and deltas, priority policy, and quantization rules are defined.

Protobuf should encode the new realtime protocol model, not the old full-state packet.

### Phase B Completion Criteria

- Realtime protocol boundary exists on server and client.
- Client codec files are moved under `client/scripts/protocol/packetcodec/`.
- Client realtime protocol code lives under `client/scripts/protocol/realtime/`.
- Server realtime protocol code lives under `services/game-server/internal/protocol/realtime/`.
- Outbound delivery mechanics and realtime protocol delivery policy are separate.
- Full snapshot and delta snapshot semantics exist.
- Per-session baseline and sequence tracking exists.
- Lane and reliability classes are defined.
- High-frequency state is no longer tied to one full reliable packet every tick.
- Quantization and bit-packing rules are defined before protobuf.
- Protobuf target is staged as the final Phase B encoding step.
- Rails/API remains outside the realtime connection path.
