## Stable limitations and hot-lane estimator docs skill

Use this skill to restore documentation consistency after the hot-lane estimator cleanup and accepted projectile stress limitation.

This documentation pass covers:

* Removal of runtime raw-float reflection assertions from active packet encoding.
* Hot-lane chunk sizing now using conservative compact-JSON byte estimation.
* Final encoded-size guards remaining the authoritative send boundary.
* Current straight bullet movement normally emitting x/y only, while retaining optional rotation tuple support for future projectile types.
* Extreme ~450–500 active projectile rendering anomalies being accepted as a stable limitation, not active server packet starvation.

## Scope

Create one new stable limitation document.

Update limits taxonomy docs so stable practical ceilings are distinct from active TODOs and permanent design laws.

Update realtime/server docs so they no longer claim runtime raw-float reflection assertions run in the active encode path.

Update realtime/server docs so hot-lane chunk construction is described as estimator-based, with final encoded-size validation still authoritative.

Do not convert client world-sync docs into local limitation ledgers.

Do not describe the projectile stress ceiling as a permanent design invariant.

## New file

Create:

```text
docs/limits/stable-limitations.md
```

Use this exact content:

```markdown
# Stable Limitations

Parent index: [Current Limits](./!INDEX.md)

## Purpose

This document captures accepted practical limitations in the current system.

Stable limitations are not active bugs, roadmap tasks, or permanent design laws. They describe practical system ceilings that may remain indefinitely because Space Rocks is not designed to support every possible stress case.

Use this document when behavior is understood well enough to accept, unlikely to matter for normal gameplay, and not worth immediate engineering work.

## Overview

Space Rocks is an arcade game with practical client, server, rendering, networking, and simulation limits.

A stable limitation can still be revisited later if normal gameplay begins to approach it, but it is not treated as active drift or required follow-up work.

## Stable limitations

### Extreme projectile rendering pressure

Extreme debug bullet-stream stress around 57 continuous bullet streams, roughly 450–500 active projectiles, can produce client-side projectile rendering anomalies.

In the tested run, the server continued to maintain 60Hz gameplay writes and complete bullet hot-lane chunk delivery. Bullet sequences continued, chunks stayed under the hot-lane cap, and no server-side bullet hot-lane gaps were observed.

Current evidence points away from server packet starvation. The suspected cause is client-side presentation pressure in projectile sync, node pooling or reuse, interpolation, or scene rendering under extreme projectile counts.

This is accepted as an edge-case stress limitation for the current arcade gameplay target. The game is not designed to render unbounded projectile counts.

## Not active work

The following are not required follow-up work unless normal gameplay reaches this limit:

- Further server hot-lane packet chunking changes.
- Async server write queues.
- Projectile rendering rewrite.
- Projectile pooling rewrite.
- Raising projectile stress targets beyond the current arcade gameplay need.

## Revisit triggers

Revisit this limitation if any of these become true:

- Normal gameplay can produce similar active projectile counts.
- Projectile rendering anomalies appear below expected gameplay stress levels.
- Server logs show bullet sequence gaps, incomplete chunks, or write cadence collapse during ordinary gameplay.
- A new weapon, mode, or modifier intentionally targets hundreds of simultaneous projectiles.

## Affected docs/systems

- Client projectile presentation and world sync.
- Game-server realtime hot-lane chunking.
- WebRTC bullet hot-lane delivery.
- Debug bullet-stream stress testing.

## Status

Accepted stable limitation. Not active work.

## Related docs

- [Current System Limits](current-system-limits.md)
- [Realtime WebSocket Protocol](../protocol/realtime-websocket-protocol.md)
- [Realtime WebRTC Gameplay Transport](../protocol/realtime-webrtc-gameplay-transport.md)
- [World Sync Coordinator](../services/client/world-sync/world-sync-coordinator.md)
- [Entity Sync Owners](../services/client/world-sync/entity-sync-owners.md)
- [Outbound Packet Routing](../services/game-server/networking/outbound-message-flow.md)
- [Lane Packet Projection](../services/game-server/simulation/runtime/lane-packet-projection.md)

## Notes

Stable limitations are accepted operating boundaries, not proof that the system should never improve. They should be revisited when normal gameplay design starts to approach the documented ceiling.
```

## Update docs/limits/!INDEX.md

Update the folder description.

Replace:

```markdown
`docs/limits` contains factual current limitations, unavailable features, incomplete integrations, hardcoded fallbacks, and current constraints.
```

With:

```markdown
`docs/limits` contains factual current limitations, unavailable features, incomplete integrations, hardcoded fallbacks, current constraints, and accepted stable limitations.
```

Update Ownership.

Replace:

```markdown
This folder owns the current-limits index for temporary constraints and known gaps.
```

With:

```markdown
This folder owns current and stable limitation docs for constraints, known gaps, unavailable behavior, and accepted practical ceilings.
```

Update Does Not Belong to include:

```markdown
- Permanent design rules.
- Domain flow docs.
- Service implementation docs.
- Planning docs.
- Stub content as canonical limits authority.
- Active work orders that belong in planning or issue tracking.
```

Add this Direct Files entry inside the doc-ledger files block:

```markdown
- [stable-limitations.md](stable-limitations.md) - Accepted practical system ceilings that are not active bugs or roadmap work.
```

Update Notes.

Replace the current note with:

```markdown
Use this folder for what the system can and cannot do right now, plus accepted stable ceilings. Permanent architecture rules belong in systems-design docs, not limits.
```

## Update docs/limits/current-system-limits.md

Do not add the projectile rendering issue to this file.

Add this cross-reference in Overview:

```markdown
Accepted practical ceilings that are not active bugs or roadmap work belong in [Stable Limitations](stable-limitations.md).
```

Add this to Related docs:

```markdown
- [Stable Limitations](stable-limitations.md)
```

## Update docs/documentation-policy.md

Replace the Limits Documentation Policy section with:

````markdown
## Limits Documentation Policy

`docs/limits/` is for documented system limitations.

Limits docs may cover temporary or active problems, and they may also cover stable practical ceilings that are accepted for the current product target.

Limits docs may cover:

```text
temporary implementation gaps
known bugs
dev-blocked issues
blocking issues
incomplete transitional behavior
current constraints that should be fixed later
accepted practical ceilings that are not active work
````

Stable limitations are not active TODOs. They document cases where the system has a practical boundary and the current behavior is acceptable for the intended game target.

Permanent design constraints, intentional architecture boundaries, and design invariants belong in `docs/systems-design/`.

Completed systems should not routinely have local “Known limits” sections. If a limitation belongs to the whole project, place it in the appropriate `docs/limits/` document and link to it only when the local doc needs that context.

If a current doc needs to reference an active problem, it should use an `Active issues` section and link to the relevant limits doc.

If a current doc needs to reference an accepted practical ceiling, it should link to `docs/limits/stable-limitations.md`.

````

## Update docs/!INDEX.md

Replace the Limits entry:

```markdown
- [Limits](limits/!INDEX.md) - Temporary blockers, gaps, and transitional limitations.
````

With:

```markdown
- [Limits](limits/!INDEX.md) - Current constraints, known gaps, transitional limitations, and accepted stable system ceilings.
```

## Update docs/developer.md

Replace the top-level Limits entry:

```markdown
* [Limits](limits/!INDEX.md) - Temporary blockers, known bugs, dev-blocked issues, active gaps, and transitional limitations.
```

With:

```markdown
* [Limits](limits/!INDEX.md) - Current constraints, known gaps, transitional limitations, and accepted stable system ceilings.
```

Replace the policy summary:

```markdown
Limits docs own temporary blockers, known bugs, and transitional issues. They do not own permanent design constraints.
```

With:

```markdown
Limits docs own temporary blockers, known bugs, transitional issues, current constraints, and accepted stable practical ceilings. They do not own permanent design constraints.
```

## Update docs/protocol/realtime-websocket-protocol.md

Remove stale claims that active encoding runs raw-float reflection assertion.

Replace this active.go ownership entry:

```text
services/game-server/internal/protocol/realtime/active.go
-> applies raw-float assertion, compact aliasing, packetcodec JSON encoding, and encoded-byte accounting for active lane packets
```

With:

```text
services/game-server/internal/protocol/realtime/active.go
-> applies compact aliasing, packetcodec JSON encoding, hot-packet encoded-size validation, and encoded-byte accounting for active lane packets
```

Replace the encode-boundary sentence that says raw-float assertion runs before CompactWirePacket with:

```markdown
Compact aliases are applied after `WireLanePacket` builds the readable long-key map. Runtime active encoding no longer performs raw-float reflection scanning at this boundary; actual numeric wire quantization must already have happened during projection or explicit event wire shaping before compaction and `packetcodec` encoding.
```

Replace the hot-lane chunking paragraph with:

```markdown
Chunk metadata exists in the wire shape and scheduler records. For `asteroid_delta` and `bullet_delta`, oversized hot movement update lists are split into multiple real same-sequence lane candidates before scheduling and encoding. Chunk construction uses conservative compact-JSON byte estimation to avoid repeated trial JSON encoding on the hot path. Each final chunk is still encoded normally, and the encoded-size guard remains the authoritative send boundary. Each sent chunk is written as its own WebRTC DataChannel message on `sr.asteroids` or `sr.bullets`. This is focused hot-lane chunking, not general fragmentation for all lane families.
```

Replace the compact bullet movement example:

```json
{"t":"bd","q":3,"bu":[[1,11,21,31]]}
```

With:

```json
{"t":"bd","q":3,"bu":[[1,11,21]]}
```

Add directly after that example:

```markdown
Current straight bullet movement normally emits x/y updates only. The compact bullet movement tuple still supports an optional trailing rotation slot for future projectile types that may turn or home during flight.
```

Replace the scheduling/chunking paragraph beginning with “The active path does not implement general record/entity-level prioritization” with:

```markdown
The active path does not implement general record/entity-level prioritization or arbitrary field-level packet splitting. It does implement focused hot-lane chunking for `asteroid_delta` and `bullet_delta`: oversized hot movement update lists become multiple real candidates before `SelectSendPlan` and before final JSON encoding. Hot-lane chunk sizing uses conservative compact-JSON byte estimation, then the final encoded-size guard classifies the real encoded packet before send. Byte estimates used by scheduling and chunk sizing are advisory and intentionally conservative; they are not the final send authority. Packets over the hard-cap or MTU class are not sent. Deferred and supersession storage exists as protocol plumbing, but active cross-tick replay and supersession are not yet the gameplay delivery guarantee.
```

Add these implementation paths to the server realtime protocol code map if absent:

```text
services/game-server/internal/protocol/realtime/hot_lane_chunker.go
services/game-server/internal/protocol/realtime/hot_lane_size_estimate.go
```

Ensure `hot_lane_size_estimate.go` is described as hand-authored runtime code, not generated output.

## Update docs/services/game-server/networking/outbound-message-flow.md

Remove raw-float assertion from the ticker-driven active lane write steps.

Replace the active write-loop step list around raw-float assertion with:

```markdown
7. `CompactWirePacket` applies final compact key/value aliasing, shared ID compaction, and tuple packing for asteroids, bullets, world ships/player records, session players, session lifecycle, and known event records.
8. `packetcodec` encodes each selected candidate into `EncodedLanePackets`.
9. `session.webrtcTransport.SendEncodedLaneJSON()` writes each encoded packet over the selected WebRTC lane channel when the transport is ready.
10. Logs lane wire packet details after successful writes.
11. Drains active event_batch events only after a successful WebRTC write.
12. Persists lane metadata only after successful writes.
13. Stores baseline projections for non-event lane packets only after successful writes.
14. Marks a lane baseline ready after a final full packet.
15. Emits a non-empty per-tick debug summary after packet writes.
```

Replace:

```markdown
`WireLanePacket` builds the readable map, `CompactWirePacket` applies aliases, compact values, shared ID compaction, and tuple packing after raw-float assertion and before `packetcodec` encoding...
```

With:

```markdown
`WireLanePacket` builds the readable map, `CompactWirePacket` applies aliases, compact values, shared ID compaction, and tuple packing before `packetcodec` encoding...
```

Add after the hot chunking paragraph:

```markdown
Hot asteroid/bullet chunk construction uses conservative compact-JSON byte estimation before scheduling so the write path does not repeatedly JSON-encode trial chunks. The final encoded-size guard still runs after real packet encoding and remains the send authority.
```

Add to the code map:

```markdown
- `services/game-server/internal/protocol/realtime/hot_lane_size_estimate.go` - estimates compact hot-lane packet and tuple byte sizes for chunk construction without repeated trial JSON encoding.
```

Replace active.go code-map text:

```markdown
- `services/game-server/internal/protocol/realtime/active.go` - active lane packet encoding path, `EncodedLanePackets` list construction, raw-float assertion/compact/packetcodec boundary, and encoded-byte accounting.
```

With:

```markdown
- `services/game-server/internal/protocol/realtime/active.go` - active lane packet encoding path, `EncodedLanePackets` list construction, compact/packetcodec boundary, hot-packet encoded-size validation, and encoded-byte accounting.
```

## Update docs/services/game-server/simulation/runtime/lane-packet-projection.md

Remove this active-flow line:

```text
-> raw-float assertion for active world/overlay/session wire maps
```

Use this active flow:

```text
authoritative game state
-> realtime projection / planning
-> raw lane records
-> numeric wire quantization into wire-shaped records
-> lane candidate selection, delta comparison, and hot movement split
-> regular asteroid movement updates move to dedicated hot-lane delta packets on sr.asteroids, and bullet movement updates move to dedicated hot-lane delta packets on sr.bullets
-> oversized asteroid/bullet hot movement update lists expand into real same-sequence candidate chunks using conservative compact-JSON byte estimates
-> sparse readable wire-map serialization
-> compact alias mapping
-> packetcodec JSON encoding
-> encoded-byte accounting
-> networking write integration
-> debug wire/summary logging after successful writes
-> WebRTC gameplay lane write using the current per-lane reliability policy
```

In the sparse serialization order, remove:

```text
-> raw-float assertion checks active world/overlay/session wire maps
```

Add after that flow:

```markdown
Runtime active encoding no longer performs raw-float reflection scanning. Numeric wire quantization remains part of projection and explicit event wire shaping before compacting and JSON encoding.
```

Add to the code map:

```markdown
* `services/game-server/internal/protocol/realtime/hot_lane_size_estimate.go` - conservative compact-JSON byte estimation used by hot-lane chunk construction.
```

Replace active.go code-map text:

```markdown
* `services/game-server/internal/protocol/realtime/active.go` - active lane packet encoding path and raw-float assertion/compact/packetcodec boundary.
```

With:

```markdown
* `services/game-server/internal/protocol/realtime/active.go` - active lane packet encoding path, compact/packetcodec boundary, hot-packet encoded-size validation, and encoded-byte accounting.
```

## Update docs/services/game-server/networking/realtime-compact-wire-mapping.md

Under `## Bullet Tuple Mapping`, after the tuple table, add:

```markdown
`bullet_delta.bullet_updates` documents the maximum supported sparse movement tuple shape. Current straight bullet movement normally emits `[id, x, y]` because rotation does not change after spawn. The optional trailing rotation slot remains supported for future projectile types, such as homing or turning projectiles, that may change rotation during flight.
```

Do not remove rotation support from the documented maximum tuple shape.

## Update docs/planning/protocol/realtime-protocol-architecture.md

Add this sentence near the current implementation summary for hot-lane chunking:

```markdown
Focused asteroid/bullet hot-lane chunking now uses conservative compact-JSON byte estimation for chunk construction, with final encoded-size guards still enforcing the actual send boundary.
```

Do not turn this planning doc into the detailed implementation authority. It should link to realtime protocol and service docs for current behavior.

## Update docs/planning/domains/technical/network-observability-and-packet-budget.md

Add this stress-test interpretation note near the existing hot-lane chunking / packet cap note:

```markdown
Recent extreme debug bullet-stream stress showed server-side hot-lane delivery sustaining 60Hz writes with complete same-sequence bullet chunks under the encoded hard cap. Client-side projectile rendering anomalies around roughly 450–500 active projectiles are tracked as a stable limitation rather than as active evidence of server packet starvation.
```

Add to Related docs if the file has a Related docs section:

```markdown
- [Stable Limitations](../../../limits/stable-limitations.md)
```

## Optional update: docs/services/game-server/observability/logging-and-diagnostics.md

If this doc is touched, add this under `## Current realtime packet debug logs`:

```markdown
Under extreme hot-lane stress, `lane protocol gameplay wire packet written` can produce one debug log per emitted hot-lane chunk. That can be hundreds of network debug records per second. This is expected diagnostic volume when `LOG_NETWORK=debug` is enabled; it is not normal quiet-mode output and does not define packet-budget policy.
```

This update is useful but not required if the documentation pass is already large.

## Do not change

Do not add local Known Limitations sections to:

```text
docs/services/client/world-sync/world-sync-coordinator.md
docs/services/client/world-sync/entity-sync-owners.md
```

These docs already describe the apply/interpolate boundary correctly. The new stable limitation doc links to them.

Do not duplicate the projectile stress limitation in:

```text
docs/limits/current-system-limits.md
```

Only add a cross-link there.

Do not describe the projectile stress limit as a permanent design invariant in:

```text
docs/systems-design/
```

This is not a design law.

Do not delete quantization docs. Quantization still exists and remains part of projection/event wire shaping.

Do not delete files.

## Stale text cleanup

After edits, no docs should claim:

```text
raw-float assertion runs on relevant lane wire maps
raw-float assertion checks active world/overlay/session wire maps
active.go applies raw-float assertion
raw-float assertion/compact/packetcodec boundary
```

Replace those with compact/packetcodec boundary wording and quantization-before-encoding wording.

## Verification

Run:

```bash
cd /mnt/d/\!bin/space-rocks
{
  grep -R "raw-float assertion" docs || true
  echo
  grep -R "hot_lane_size_estimate.go" docs || true
  echo
  grep -R "stable-limitations.md" docs || true
} 2>&1 | tee /dev/tty | clip.exe
```

Expected:

* No stale “raw-float assertion runs” claims remain.
* `hot_lane_size_estimate.go` appears in relevant server/protocol docs.
* `stable-limitations.md` appears in the limits index and relevant related-doc links.

Then run the normal project documentation/index verification. Ensure `docs/limits/!INDEX.md` includes `stable-limitations.md` inside its doc-ledger files block.

## Report

Report:

```text
Changed files
New files
Stale raw-float references removed
Stable limitation doc created
Hot-lane estimator docs updated
Index/related-doc links updated
Verification result
```
