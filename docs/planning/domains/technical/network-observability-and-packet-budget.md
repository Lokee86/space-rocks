# Network Observability And Packet Budget
Parent index: [Technical Planning](./!README.md)

## Purpose

This doc plans the network-visibility and packet-budget seam for future growth in gameplay and realtime state delivery.

## Ownership Boundary

This doc owns planning for gameplay packet budget, outbound byte metrics, large-packet diagnostics, contributor counts, and devtools visibility.

It should stay on measurement and observability rather than packet-format redesign.

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

## Related Docs

- [Systems Plan Index](systems-plan-index.md)
- [Realtime Protocol Architecture](realtime-protocol-architecture.md)
- [Devtools And Telemetry](devtools-and-telemetry.md)
- [Logging And Diagnostics](logging-and-diagnostics.md)
- [Platform And Progression Roadmap](platform-and-progression-roadmap.md)

## Open Planning Questions

- Which packet sizes should remain warnings versus blockers?
- Which contributor counts are worth tracking long term?
- Which packet metrics should stay devtools-only versus also land in logs?

## Phase A

Phase A answers whether the current architecture can safely support more entities, enemies, bullet hell patterns, progression events, and online play without flying blind. Phase A is measurement and diagnostics, not optimization. Phase A should make later optimization choices evidence-based.

### Existing Baseline

- `services/game-server/internal/networking/outbound/gameplay_state_metrics.go` already warns on gameplay presentation packets over 4KB.
- The same outbound path warns on gameplay presentation writes slower than 20ms.
- `services/game-server/internal/networking/outbound/` owns outbound gameplay presentation helpers.
- `services/game-server/internal/protocol/packetcodec/` owns JSON packet encode/decode.
- `client/scripts/devtools/telemetry/` owns client-side telemetry models.
- `client/scenes/devtools/world_telemetry_overlay.tscn` is the devtools-only overlay.
- `docs/server/logging.md` and `docs/client/logging.md` already define logging rules.

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
- Surface packet byte pressure in devtools telemetry.
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

### Initial Packet Budget

| Threshold | Policy |
| --- | --- |
| Gameplay packet warning: 4KB | Structured warning with contributor counts. |
| Gameplay packet danger: 8KB | Treat as a blocker before entity-heavy feature growth. |
| Slow gameplay write: 20ms | Structured warning with packet size and route context. |
| Target steady-state gameplay packet: under 4KB | Preferred normal gameplay state. |

These thresholds are provisional. Phase A measures whether they are realistic.

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

### Devtools Display Requirements

- The World Telemetry Overlay should show latest gameplay packet bytes.
- The World Telemetry Overlay should show max gameplay packet bytes.
- The World Telemetry Overlay may show optional average gameplay packet bytes.
- The World Telemetry Overlay should show large packet warning count if cheap to track.
- Existing entity counts and timing values should remain.
- This remains devtools-only and must not affect gameplay.

### Likely Phase A Workstreams

1. Document packet budget policy.
2. Extend server contributor metrics in the outbound gameplay presentation path.
3. Measure client-side inbound raw message byte length by packet type.
4. Surface packet byte metrics in the world telemetry overlay.
5. Update telemetry/logging docs and add a smoke checklist.

### Phase A Completion Criteria

- Packet budget policy is documented.
- Large gameplay packets include contributor-count diagnostics.
- Slow writes include useful context.
- Devtools overlay exposes packet byte pressure.
- Manual smoke can demonstrate packet size changes as bullets and asteroids increase.
- No packet format has changed.
- No gameplay behavior has changed.
- No feature work is mixed in.

### Post-Phase-A Decision Gate

Phase A does not automatically choose the Phase B emphasis. Phase A exists to decide what the next major route inside Phase B should be. Network optimization and related protocol work is the most likely next route if Phase A confirms current packet pressure, because gameplay packets are already known to exceed 4KB at times before enemies or bullet hell mechanics exist.

Route 1 - Network optimization immediately

- Choose this if normal gameplay packets are often over 4KB.
- Choose this if packets spike toward or past 8KB.
- Choose this if packet size grows predictably with bullets, asteroids, pickups, or players.
- Choose this if write times or jitter correlate with packet size.
- Choose this if entity-heavy features would clearly make packet pressure worse.
- This is the likely route if Phase A confirms the current concern.

Route 2 - More observability hardening before optimization

- Choose this if packet size is measured but contributors are unclear.
- Choose this if client overlay and server logs disagree.
- Choose this if slow writes happen without large packets.
- Choose this if packet size is acceptable but tick, build, or write timing is not.
- Choose this if instrumentation is too noisy or incomplete to justify a protocol change.

Route 3 - Move to auth and account identity planning before optimization

- Choose this only if normal gameplay packets stay below 4KB.
- Choose this only if spikes are rare and explainable.
- Choose this only if write timing and jitter show no packet-size pressure.
- Choose this only if packet size is not blocking enemies, bullet hell, or progression soon.

Likely optimization families under Route 1, without choosing one:

- Compact wire shape or generated short field names, if JSON key overhead dominates.
- Delta snapshots, if repeated full entity state dominates.
- Fast/slow packet lane split, if all data is being sent at the same frequency.
- Event queue trimming or acknowledgement, if events accumulate or repeat too long.
- Debug lane separation, if debug or devtools data leaks into normal gameplay packets.
- Shared room snapshot plus per-client overlay, if most state is duplicated per client but only small portions are player-specific.

The next planning work after Phase A should be selected by evidence from the decision gate, not by feature visibility alone.
