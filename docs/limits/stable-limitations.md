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

Extreme debug bullet-stream stress around 57 continuous bullet streams, roughly 450-500 active projectiles, can produce client-side projectile rendering anomalies.

In the tested run, the server continued to maintain 60Hz gameplay writes and complete bullet hot-lane chunk delivery. Bullet sequences continued, chunks stayed under the hot-lane cap, and no server-side bullet hot-lane gaps were observed.

Current evidence points away from server packet starvation. The suspected cause is client-side presentation pressure in projectile sync, node pooling or reuse, interpolation, or scene rendering under extreme projectile counts.

Client inbound apply cost or same-sequence hot-chunk merge behavior may also contribute, but the tested server logs did not show missing bullet sequences, incomplete chunks, or write cadence collapse.

This is accepted as an edge-case stress limitation for the current arcade gameplay target. The game is not designed to render unbounded projectile counts.

### Match-scoped replay and deletion store growth

The client keeps match-scoped dictionaries for applied event batch IDs, applied event IDs, logged applied batch IDs, world-lane deleted bullet IDs, projectile-sync deleted projectile IDs, and asteroid-sync deleted asteroid IDs. They prevent duplicate event presentation and stale/deleted entity resurrection. They are cleared by match/session reset, so this is within-match growth, not cross-match accumulation.

The July 12, 2026 smoke measurement accumulated 3,706 entries over 211.2 seconds during normal gameplay, approximately 1,053 entries/minute. Sustained stress accumulated 44,929 entries over 163.5 seconds, approximately 16,488 entries/minute. In the stress run, the two bullet-deletion stores contained 18,768 entries each, about 83.6% of all measured entries combined.

The synthetic Godot memory probe measured approximately 195-206 bytes per representative string-ID dictionary entry over 1,000 to 250,000 entries. Linear projections from those short runs are:

| Match duration | Normal | Stress |
| --- | ---: | ---: |
| 1 hour | 11.7-12.4 MB | 184-194 MB |
| 4 hours | 47.0-49.6 MB | 736-777 MB |
| 8 hours | 94.0-99.3 MB | 1.47-1.55 GB |
| 24 hours | 282-298 MB | 4.42-4.66 GB |

These are linear projections from short runs, not long-duration validation. The accepted boundary is that ordinary match lengths are safe; endurance matches or modes sustaining extreme projectile churn are not supported without bounded replay windows or protocol-safe tombstone pruning.

Revisit this limitation for endurance or persistent matches, normal gameplay approaching materially higher churn, sustained sessions causing noticeable memory growth, or protocol work that provides a safe pruning/replay window. No pruning work is active solely because of this limitation.

## Not active work

The following are not required follow-up work unless normal gameplay reaches this limit:

- Further server realtime candidate chunking changes beyond the current full/lifecycle and asteroid/bullet movement hard-cap path.
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
