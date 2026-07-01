# Network Observability And Packet Budget
Parent index: [Technical Planning](./!INDEX.md)

## Purpose

This doc plans the network-visibility and packet-budget seam for future growth in gameplay and realtime state delivery.

## Overview

This doc keeps packet-size observability, contributor counts, and large-message diagnostics aligned so network pressure can be measured before protocol changes are chosen.

## Current status

Active planning.

## Ownership Boundary

This doc owns planning for gameplay packet budget, outbound byte metrics, large-packet diagnostics, contributor counts, and devtools visibility.

It should stay on measurement and observability rather than packet-format redesign.

## Canonical Packet Budget

These are the current project policy numbers for realtime gameplay traffic:

- Client input packets target 32-64 B.
- Client input packets warn above 128 B.
- Normal server snapshots target 250-500 B.
- Busy server snapshots may reach 500-800 B.
- Sustained server snapshots above 800 B should warn.
- Realtime gameplay datagrams must stay below roughly 1,100-1,200 B.
- Non-realtime, control, and debug payloads are separate from gameplay packet budgets and must not redefine the realtime budget.

## Current Inputs

- gameplay packet budget inputs
- outbound byte metric inputs
- large-packet diagnostic inputs
- contributor count inputs
- devtools visibility inputs

## Planned Outputs

- packet-budget planning boundaries
- diagnostic expectations for large gameplay packets
- visibility requirements for devtools and logging

## Phase P1 - Network Observability And Packet Budget

P1 answers whether the current architecture can safely support more entities and realtime state growth without flying blind. P1 is now a server-side packet evidence checkpoint. P1 records enough pressure and contributor data to select Phase P2 realtime protocol architecture. P1 is measurement and instrumentation, not optimization.

### Existing Baseline

- `services/game-server/internal/networking/packetmetrics/` owns gameplay packet severity classification, large-packet diagnostics, contributor counts, and slow-write metric context.
- The same outbound path warns on gameplay presentation writes slower than 20ms.
- `services/game-server/internal/networking/outbound/` owns outbound gameplay presentation helpers.
- `services/game-server/internal/protocol/packetcodec/` owns JSON packet encode/decode.
- `client/scripts/devtools/telemetry/` owns client-side telemetry models.
- `client/scenes/devtools/world_telemetry_overlay.tscn` is the devtools-only overlay.
- `docs/services/game-server/observability/logging-and-diagnostics.md` and `docs/services/client/client-logging.md` already define logging rules.

### Current-State Note

- Active logs report lane packet writes and candidate-level scheduling summaries.
- Encoded byte counts are real after encode.
- Scheduling estimates are not codec-aware.
- Budget status is not yet a reliable hard-enforcement proof.

### Future-State Note

- Metrics must eventually prove included, deferred, and superseded counts by record or field group.
- Metrics must compare estimated bytes with encoded bytes.
- Metrics must distinguish target, warning, danger, and hard-cap outcomes.

### Ownership Rules

- Server networking owns encoded outbound packet size, write duration, packet type/category, room ID, player ID, and presentation-state diagnostics.
- Server gameplay owns authoritative state and entity counts before serialization.
- Server logging owns threshold warnings and structured fields.
- Client devtools owns packet/network metrics display.
- Client HUD does not own packet observability.
- Documentation owns packet-budget policy.

### Goals

- Define an initial gameplay packet budget.
- Measure outbound gameplay packet byte size.
- Identify contributor counts for large gameplay packets.
- Surface packet byte pressure in devtools telemetry when useful during P2 validation.
- Keep observability separate from gameplay behavior.
- Preserve JSON encoding until measurements identify the bottleneck.
- Provide evidence for later packet strategy work.

### Non-Goals

- No packet compression.
- No binary protocol migration.
- No delta-state protocol.
- No gameplay packet lane split.
- No enemies.
- No bullet hell mechanics.
- No progression rewards or live grants.
- No auth expansion.
- No website work.
- No player-facing telemetry.
- No raw full-payload packet dumps by default.
- No gameplay behavior changes.

### Initial Guidance

- Gameplay snapshots have a tight budget on the realtime path.
- Non-realtime, control, and debug payloads are separate from gameplay packet budgets.
- Large gameplay packets are diagnostic signals, not a steady-state allowance.
- The canonical budget lives here; remaining telemetry and logging work is paused until packet-size reduction makes it useful again. Packet metrics and logs can be used to observe reduced JSON numeric size in float-heavy lanes, but this does not imply fixed savings for every packet mix.
- Preferred frequent realtime packets should stay small and predictable.
- Packets that grow noticeably should be justified, lowered in frequency, split, or deferred to later protocol work.

### Required Large-Packet Diagnostics

- Encoded byte size
- Packet type
- Room ID
- Player ID
- Remote address if already available in the outbound path
- Room state
- Players count
- `player_sessions` count
- `player_lifecycle` count
- Asteroid count
- Bullet count
- Pickup count
- Enemy count
- Event count
- Total spawned asteroid count
- Build duration where cheap and localized
- Encode duration where cheap and localized
- Write duration where cheap and localized

Raw packet payloads should not be logged by default.

### Phase P2 Validation Display Requirements
These display requirements are deferred until they are useful during Phase P2 validation; they are not Phase P1 completion blockers.

- The World Telemetry Overlay should show latest gameplay packet bytes.
- The World Telemetry Overlay should show max gameplay packet bytes.
- The World Telemetry Overlay may show optional average gameplay packet bytes.
- The World Telemetry Overlay should show large packet warning count if cheap to track.
- Existing entity counts and timing values should remain.
- This remains devtools-only and must not affect gameplay.

### Likely Phase P1 Workstreams

1. Document packet budget policy.
2. Extend server contributor metrics in the outbound gameplay presentation path.
3. Add cheap build, encode, and write duration context.
4. Keep remaining telemetry/logging paused until P2 validation needs it.

### Phase P1 Completion Criteria

- Packet budget policy is documented.
- Large gameplay packets include contributor-count diagnostics.
- Slow writes include useful context.
- Server evidence is enough to select realtime protocol work.
- No packet format has changed.
- No gameplay behavior has changed.
- No feature work is mixed in.

### Phase P1 Decision Gate

Phase P1 uses server-side packet evidence to decide whether Phase P2 should start immediately. Realtime protocol architecture is the selected Phase P2 route when current packet pressure is confirmed, because gameplay packets are already over budget before enemies or bullet hell mechanics exist.

Outcome 1 - Start Phase P2 realtime protocol work immediately

- Choose this if normal gameplay packets are often large.
- Choose this if packets spike upward under gameplay load.
- Choose this if packet size grows predictably with bullets, asteroids, pickups, or players.
- Choose this if write times or jitter correlate with packet size.
- Choose this if entity-heavy features would clearly make packet pressure worse.
- This is the likely outcome if Phase P1 confirms the current concern.

Outcome 2 - Add only the observability needed before protocol work

- Choose this if packet size is measured but contributors are unclear.
- Choose this if client overlay and server logs disagree.
- Choose this if slow writes happen without large packets.
- Choose this if packet size is acceptable but tick, build, or write timing is not.
- Choose this if instrumentation is too noisy or incomplete to justify a protocol change.

Outcome 3 - Move to account identity planning before protocol work

- Choose this only if normal gameplay packets stay modest and are not trending upward.
- Choose this only if spikes are rare and explainable.
- Choose this only if write timing and jitter show no packet-size pressure.
- Choose this only if packet size is not blocking enemies, bullet hell, or progression soon.

Likely protocol work families under Phase P2, without choosing one:

- Compact wire shape or generated short field names, if JSON key overhead dominates.
- Delta snapshots, if repeated full entity state dominates.
- Session lane split, if all data is being sent at the same frequency.
- Event batching, event IDs, batch IDs, duplicate suppression, and drain-after-active-socket-write/enqueue-success behavior if presentation events accumulate or repeat too long.
- Debug lane separation, if debug or devtools data leaks into normal gameplay packets.
- Shared room snapshot plus per-client overlay, if most state is duplicated per client but only small portions are player-specific.

The next planning work after Phase P1 should be selected by evidence from the decision gate, not by feature visibility alone.
### P2 Validation Support

During P2, deferred network telemetry and logging work can resume when it helps validate protocol changes:

```text
client inbound packet byte tracking when useful
World Telemetry Overlay packet display when useful
client/server packet comparison if needed
packet-pressure smoke checks for protocol changes
logging refinements needed to validate packet-size reduction
```

This support work belongs to P2 when it helps validate lanes, snapshots, deltas, baseline handling, packet-size improvements, or realtime protocol behavior.

## Implementation sequence

1. Document the canonical packet budget and keep measurement and diagnostics current.
2. Extend outbound gameplay presentation metrics with contributor-count diagnostics.
3. Add cheap build, encode, and write duration context.
4. Use server-side packet evidence to select realtime protocol work.
5. Resume remaining telemetry/logging during P2 when it helps validate protocol changes.

## Open decisions

- Which packet sizes should remain warnings versus blockers?
- Which contributor counts are worth tracking long term?
- Which packet metrics should stay devtools-only versus also land in logs?
- Whether Phase P1 evidence pushes the next step toward realtime protocol work, more observability hardening, or other planning.

## Related docs

- [Planning](../../!INDEX.md)
- [Realtime Protocol Architecture](../../protocol/realtime-protocol-architecture.md)
- [Devtools And Telemetry](../../devtools/devtools-and-telemetry.md)
- [Logging And Diagnostics](observability-logging-and-diagnostics.md)
- [Development Roadmap](../../development-roadmap.md)

## Notes

Preserve the packet-budget policy and Phase P1 structure; this doc owns measurement, diagnostics, and decision gates rather than packet-format redesign.






