# Realtime Protocol Architecture
Parent index: [Protocol Planning](./!INDEX.md)

## Purpose

This doc plans Phase P2, which replaces the current combined `state` packet with a lane-scoped realtime protocol for authoritative state delivery.

## Ownership Boundary

This doc owns planning for the broader protocol boundary, with `protocol/realtime` and `protocol/packetcodec` as sibling seams beside `networking/outbound`.

`protocol/realtime` owns replication policy, `protocol/packetcodec` owns representation and encoding, and `networking/outbound` owns delivery mechanics.

It should cover lane-scoped full snapshots, lane-scoped deltas, baselines, resync, priority scheduling, lane/priority metrics, shadow verification, cutover, and old `state` deletion. Deferred next-phase work covers quantization, bit packing, protobuf, and client packet codec relocation.

## Current Inputs

- lane inputs
- full snapshot inputs
- delta snapshot inputs
- baseline inputs
- resync inputs
- priority inputs
- lane/priority metrics inputs
- shadow verification inputs
- cutover inputs
- state deletion inputs

## Planned Outputs

- protocol ownership boundaries
- lane vocabulary for realtime delivery
- the preferred sequence of protocol upgrades

## Related Docs

- [Planning](../!INDEX.md)
- [Network Observability And Packet Budget](../domains/technical/network-observability-and-packet-budget.md)
- [Testing And Smoke Strategy](../domains/technical/verification-and-quality-gates.md)
- [Development Roadmap](../development-roadmap.md)

## Open Planning Questions

- Which lane types need to exist before shadow verification starts?
- Which lane and priority metrics need to exist before cutover?
- Which responsibilities stay in outbound delivery mechanics versus protocol policy?

## Phase P2 - Realtime Protocol Architecture

Phase P2 establishes the lane-scoped semantic realtime protocol seam for authoritative multiplayer state delivery. P2 replaces the current combined `state` path rather than optimizing it. It covers lane-scoped full snapshots, lane-scoped deltas, baselines, resync, priority scheduling, lane/priority metrics, shadow verification, cutover, and old `state` deletion.

### Canonical P2 Decisions

- The current combined `state` packet is replaced, not optimized.
- `state` may remain during migration only as a reference/parity aid.
- `state` is not a fallback, compatibility route, or debug dump after cutover.
- Runtime should have only one active gameplay protocol at a time.
- The new protocol is built in parallel, shadow-verified, cut over, and then the old `state` path is removed immediately.
- P2 targets lane-scoped semantic protocol work, not a small partial delta patch.
- Session is the canonical lane name.
- Every current `StatePacket` field must have a new lane owner before cutover.
- Priority is part of P2.
- Delta and priority are separate seams.
- No full ACK/retry for `event_batch` in P2; `event_batch` uses `batch_id`, each event uses `event_id`, and after socket write/enqueue success delivery is at-most-once with duplicate suppression.
- Stable event identity is required before `event_batch` cutover.
- Shadow may peek/copy pending events but never drains.
- Projection does not drain.
- Scheduling does not drain.
- Encoding does not drain.
- Active `event_batch` drains only after socket write/enqueue success.
- `player_pause_state` remains a separate immediate/control-style output during P2 unless later work proves otherwise.
- Client packet codec movement is deferred.
- WebSocket may remain as the transport during P2; lanes are protocol policy first.

### State Packet Replacement Field Ownership

Every current `StatePacket` field must move to a lane-scoped packet family before cutover. The current `state` packet is a migration reference only and must be removed after the new field ownership is confirmed.

- `players` -> world lane
- `bullets` / projectiles -> world lane
- `asteroids` -> world lane
- `pickups` -> world lane
- `self_id` -> overlay lane
- `lives` -> overlay lane unless later proven shared/session
- `server_sent_msec` -> protocol metadata / overlay timing
- `player_sessions` -> session lane
- `player_lifecycle` -> session lane
- `total_asteroids` -> session lane
- `events` -> event lane
- debug status / shape catalog / packet metrics -> debug/telemetry lane

### Packet Family And Cutover

Replacement packet families:

- `world_full`
- `world_delta`
- `overlay_full`
- `overlay_delta`
- `session_full`
- `session_delta`
- `event_batch`
- `resync_request` / resync control messages
- debug/telemetry packets

`*_full` messages are complete replacement snapshots for that lane only.

`*_delta` messages are lane-specific create/update/delete or dirty-field updates.

Migration and cutover rules:

- Old `state` remains active while the new protocol is built beside it.
- New projections and packets are parity-tested against `state`.
- Runtime switches once the new lane-scoped protocol is complete.
- Old `state` is removed immediately after cutover is confirmed.

Server placement:

- `services/game-server/internal/protocol/realtime/`
- `services/game-server/internal/protocol/realtime/packets_generated.go`
- Keep `services/game-server/internal/networking/outbound/` as delivery mechanics.

Client placement:

- `client/scripts/protocol/realtime/`
- Keep `client/scripts/networking/packets/` unmoved in P2.
- Defer `client/scripts/protocol/packetcodec/` relocation to next-phase compact representation work.

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

- Control lane
- World lane
- Overlay lane
- Event lane
- Session lane
- Debug/telemetry lane

Deferred packet telemetry and logging work may resume during Phase P2 when useful to validate lanes, baselines, resync, priority scheduling, shadow verification, or cutover. Validation work should support protocol decisions and should not become a separate pre-P2 blocker.

Transport architecture is a game-time decision. The protocol lanes should be able to map later to WebSocket, WebRTC DataChannel, UDP, or a hybrid, but this phase does not need to choose transport immediately.

### Snapshot And Delta Model

- Full snapshot on join, start, or resync
- Delta snapshots after baseline
- Shared room/world baselines for common realtime world state
- Per-session overlay baselines for receiver-specific state
- Category-aware hot/cold delta policy
- Monotonically increasing sequence numbers
- Entity create/update/delete records
- Missing-baseline recovery
- Stale update discard
- Explicit resync path
- Missing from full means remove.
- Missing from delta means unchanged.
- Explicit delete record means remove.
- Create and update records apply changes.
- Wrong baseline or missing baseline triggers resync behavior.
- Stale sequence is discarded.
- Full snapshots are lane-scoped, not whole-protocol dumps.
- Universal per-entity versioning is not required for P2.

### State Projection And Priority Policy

Priority is not just labels. `protocol/realtime` must include a priority planner that selects records under a packet budget.

Planner inputs:

- Lane
- Candidate records
- Session context
- Baseline
- Byte budget
- Age since last sent
- Distance / relevance / threat
- Create / delete status

Planner outputs:

- Records sent now
- Records deferred
- Full / delta / resync decision
- Budget status

Priority bands:

- Critical: resync/control, deletes, local player state, player create/despawn, room/control essentials
- High: nearby/high-threat projectiles, nearby players/enemies, nearby asteroid changes, nearby pickups
- Medium: normal visible world updates
- Low: far asteroids, far pickups, session metadata, cosmetic/non-urgent state
- Debug: devtools/telemetry only

When a packet exceeds budget, the planner controls send order, chunking, deferral, supersession, and resync decisions. Required records must be sent, chunked, staged, deferred, or trigger resync; they must not be silently discarded. Critical and high records are sent first, lower-priority records are deferred, and deferred records age upward over time. Lane scope guides what can be considered; the planner decides what fits. `world_delta` does not emit every changed entity regardless of budget.

`protocol/realtime` decides what should be sent and how often, including local player state, active player session state, nearby threats, bullets and projectiles, asteroids, pickups, future enemies, future bullet hell entities, and debug-only data.

### Outbound Collaboration

- `networking/outbound` owns delivery cadence, active sessions and channels, write calls, write failures, and backpressure.
- `protocol/realtime` owns replication policy, including what messages are due, what state they contain, lane assignment, priority, full and delta snapshot rules, sequence and baseline rules, and resync behavior.
- `protocol/packetcodec` owns representation and encoding.

Flow:

1. Authoritative game state
2. `protocol/realtime` projection
3. Priority and lane policy
4. Full snapshot or delta snapshot
5. Lane-scoped protocol message shape
6. `protocol/packetcodec` encoding
7. `networking/outbound` delivery mechanics
8. Transport channel
9. Client `protocol/realtime` application
10. World/gameplay presentation

### Deferred Codec And Compact Representation

Deferred next-phase codec-compact-representation work:

- Quantization
- Bit packing
- Protobuf
- `services/game-server/internal/protocol/packetcodec/`
- `client/scripts/protocol/packetcodec/`
- Client packet codec relocation out of `client/scripts/networking/packets/`

P2 keeps `client/scripts/networking/packets/` unmoved.

### Phase P2 Completion Criteria

- Lane-scoped replacement packets exist.
- Every current `StatePacket` field has a new lane owner.
- The current `state` runtime path is removed after cutover.
- Server and client `protocol/realtime` exist.
- Outbound delivery mechanics and realtime policy are separate.
- The hybrid shared-world plus per-session-overlay baseline model exists.
- Full and delta semantics exist per lane.
- A budgeted priority planner exists.
- Sequence, baseline, resync, and stale-handling behavior exists.
- Lane, priority, and budget metrics exist.
- Shadow verification exists.
- High-frequency state is no longer one full combined packet every tick.
- `event_batch` duplicate suppression and control-path/event-drain ordering are defined.




### Step 8 Deletion Terms

Search and delete references to the old combined `state` path, temporary cutover switches, and old client readiness names:

```text
StatePacket
PacketTypeState
TYPE_STATE
Game.StatePacket
BuildGameplayPresentationStateResponse
CanSendGameplayPresentationState
gameplay_state_packet_reader
GameplayStatePacketReader
GameplayStateFlow
GameplayShellFlow
gameplay_state_received
has_received_gameplay_state
has_received_state
state packet dispatcher route
protocol_mode
use_new_protocol
legacy_state_protocol
lane_protocol
```



