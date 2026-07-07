# Network Observability And Packet Budget
Parent index: [Technical Planning](./!INDEX.md)

## Purpose

This doc plans the network-visibility and packet-budget seam for future growth in gameplay and realtime state delivery.

## Overview

This doc keeps packet-size observability, local diagnostic capture, and packet-evidence planning aligned so network pressure can be measured before protocol changes are chosen.

## Current status

Active planning.

## Ownership Boundary

This doc owns planning for gameplay packet budget, outbound byte evidence, local diagnostic capture, validation-support telemetry, and devtools visibility.

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
- outbound byte evidence inputs
- local JSONL diagnostic capture inputs
- validation-support logging inputs
- devtools visibility inputs

## Planned Outputs

- packet-budget planning boundaries
- diagnostic expectations for gameplay packet evidence
- visibility requirements for devtools and logging

## Phase P1 - Network Observability And Packet Budget

P1 answers whether the current architecture can safely support more entities and realtime state growth without flying blind. P1 is now a server-side packet evidence checkpoint. P1 records enough pressure and contributor data to select Phase P2 realtime protocol architecture. P1 is measurement and instrumentation, not optimization.

### Existing Baseline

- Server sequential JSONL diagnostic file output exists.
- Client sequential JSONL diagnostic file output exists.
- `services/game-server/internal/networking/packetmetrics/` remains a helper/support seam for packet observability work, but it is not the current active `realtime lane metric` runtime output.
- `services/game-server/internal/protocol/packetcodec/` owns JSON packet encode/decode.
- `client/scripts/devtools/telemetry/` owns client-side telemetry models.
- `client/scenes/devtools/world_telemetry_overlay.tscn` is the devtools-only overlay.
- `docs/services/game-server/observability/logging-and-diagnostics.md` and `docs/services/client/client-logging.md` already define logging rules.

### Current-State Note

- Active server packet evidence now comes from per-packet wire logs and non-empty per-tick write summaries.
- Active server debug logs expose per-packet `encoded_bytes` plus lane, type, kind, and sequence context.
- Active server debug summaries expose non-empty per-tick `packet_count` and total `encoded_bytes` summaries.
- No-op realtime summaries are intentionally suppressed when no packets were written.
- `realtime lane metric` was removed from active runtime output.
- Scheduler, budget, deferred, superseded, and CRUD-count fields are intentionally not emitted as current packet evidence, even though protocol/realtime has candidate-level send-plan records.
- Active debug output does not prove contributor counts by delta section.
- Active debug output does not itself implement packet-budget policy; current candidate-level include/defer selection and hot-packet encoded-size guards live in protocol/realtime. Record/entity-level prioritization remains future work.
- Focused hot-lane chunking is current for `asteroid_delta` and `bullet_delta`. Packet evidence should expect multiple packets per tick on `sr.asteroids` or `sr.bullets` under stress.
- `packet_count` is a count of encoded packets written, not unique lanes.
- Encoded bandwidth evidence should be interpreted with write cadence: under peak stress, bandwidth may drop because write cadence drops even while entity pressure rises.
- Large-packet warnings and slow-write diagnostics should be treated as partial or seam-specific support only where current code still emits them, not as the complete current evidence story.
- `event_batch` may be selected alongside other active lane candidates in the same tick.
- Compact sparse event records reduce that event-tick spike.
Recent extreme debug bullet-stream stress showed server-side hot-lane delivery sustaining 60Hz writes with complete same-sequence bullet chunks under the encoded hard cap. Client-side projectile rendering anomalies around roughly 450-500 active projectiles are tracked as a stable limitation rather than as active evidence of server packet starvation.
Lane-native deltas, mixed-policy physical WebRTC gameplay DataChannels, dedicated unordered/unreliable asteroid/bullet hot movement lanes, and asteroid, bullet, ship/player, session player/lifecycle, and known event tuple packing are current implementation facts, not future work. Current mixed policy means ordered/reliable sr.world, sr.overlay, sr.session, and sr.event, plus unordered/unreliable sr.asteroids and sr.bullets.

### Future-State Note

- Metrics must eventually prove included, deferred, and superseded counts by record or field group.
- Metrics must compare estimated bytes with encoded bytes.
- Metrics must distinguish target, warning, danger, and hard-cap outcomes.
- Current hot-lane chunking keeps individual asteroid/bullet movement packets under the hard cap. Future metrics still need to prove included/deferred/superseded behavior and contributor counts at record or field-group granularity.

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
- The canonical budget lives here. Remaining telemetry and logging work should support P2 validation when it becomes useful, rather than acting as an endless Phase P1 blocker. Packet evidence can still be used to observe reduced JSON numeric size in float-heavy lanes, but this does not imply fixed savings for every packet mix.
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
2. Keep current packet evidence and local JSONL diagnostic capture aligned with the active runtime output.
3. Add only the cheap build, encode, and write context that is still useful and actually supported by current code.
4. Keep remaining telemetry/logging framed as P2 validation support, not a Phase P1 blocker.

### Phase P1 Completion Criteria

- Packet budget policy is documented.
- Server evidence is enough to select realtime protocol work.
- Remaining telemetry and logging support is scoped to what helps P2 validation, not to proving every future packet policy in Phase P1.
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

- Choose this if packet size is measured but contributor evidence is still too coarse for the next decision.
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
- Event batching, event IDs, batch IDs, duplicate suppression, and drain-after-active-socket-write/enqueue-success behavior if presentation events accumulate or repeat too long. `event_batch` is already an active delivery path, so duplicate suppression and drain semantics should be described as current behavior where implemented, not only future work.
- Debug lane separation, if debug or devtools data leaks into normal gameplay packets.
- Shared room snapshot plus per-client overlay, if most state is duplicated per client but only small portions are player-specific.

The next planning work after Phase P1 should be selected by evidence from the decision gate, not by feature visibility alone. After tuple packing is current, the next likely steps are to verify packet-size gains with stress logs, continue world lane reductions where needed, and keep binary/bitpacking as later work unless measurements justify it sooner.
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
2. Keep active packet evidence and local JSONL diagnostic capture aligned with current runtime logging.
3. Add only the build, encode, and write context that remains cheap, localized, and actually supported.
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
- [Stable Limitations](../../../limits/stable-limitations.md)

## Notes

Preserve the packet-budget policy and Phase P1 structure; this doc owns measurement, diagnostics, and decision gates rather than packet-format redesign.














