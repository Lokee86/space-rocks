# Long-Match Store Diagnostics

Parent index: [Client](./!INDEX.md)

## Purpose

This document describes the client diagnostic surface for observing bounded long-match replay and deletion stores.

## Overview

`GameplaySessionController` logs a `long_match_store_summary` network event during reset after gameplay packet acceptance has begun. The record reports the current transient client store sizes for the active match.

The summary includes:

```text
match_id
duration_ms
applied_event_batch_ids
applied_event_ids
logged_applied_batch_ids
world_lane_deleted_bullet_ids
world_lane_pending_bullet_updates
projectile_sync_deleted_projectile_ids
asteroid_sync_deleted_asteroid_ids
total_entries
```

The diagnostic reads transient client state only. It does not mutate gameplay, provide server authority, or create durable analytics.

## Debug-only scope

This surface is for development diagnostics and long-match validation. It reports client-side retention state and does not change packet application, entity synchronization, event presentation, or reset behavior.

## Server authority

The game server remains authoritative for gameplay outcomes, lifecycle truth, entity existence, scoring, lives, respawn, and match state. The diagnostic does not measure or replace server-side state, server memory, or server analytics.

## Client presentation

The stores belong to transient client gameplay application and world-sync presentation:

| Store | Cap |
| --- | ---: |
| Applied event batch IDs | 4096 |
| Applied event IDs | 8192 |
| Logged applied batch IDs | 4096 |
| `WorldLaneState` deleted bullet IDs | 4096 |
| `WorldLaneState` pending unknown bullet updates | 2048 |
| `ProjectileSync` deleted projectile IDs | 4096 |
| `AsteroidSync` deleted asteroid IDs | 2048 |

Each bounded history evicts its oldest retained entry after reaching its cap. Duplicate records do not consume additional insertion capacity. Recreation, clear, and reset paths keep public dictionaries and private insertion-order state synchronized.

## Commands or controls

No in-game command or control is required. The normal client reset path emits the `long_match_store_summary` network event after gameplay packet acceptance has begun.

The standalone probe runs headlessly with:

```text
C:\Godot_v4.6.3-stable_win64.exe --headless --path client --script res://tools/long_match_store_memory_probe.gd
```

## Telemetry

The reset summary contains `match_id`, `duration_ms`, all seven store counts, and `total_entries`. It is a diagnostic network event containing transient client measurements only; it is not durable analytics and does not claim server-side metrics.

`client/tools/long_match_store_memory_probe.gd` is a headless production-store probe. It cycles insertions across all seven real stores, emits threshold memory and store-size records, and demonstrates aggregate bounded-store plateauing at 28672 retained entries after all caps are reached. This demonstrates the bounded stores only; it does not validate total process memory stability.

## Build/runtime gates

The probe requires the Godot client project and its configured scripts. It does not instantiate gameplay scenes or require configured render layers. Run it from the repository root or with `client` as the project path using the command above.

## Code map

* `client/scripts/session/gameplay_session_controller.gd` - Emits the `long_match_store_summary` diagnostic during reset after gameplay packet acceptance begins.
* `client/scripts/devtools/telemetry/long_match_store_metrics.gd` - Builds the bounded-store snapshot and aggregate count.
* `client/scripts/protocol/realtime/event_batch_applier.gd` - Owns applied-batch, applied-event, and logged-batch histories.
* `client/scripts/protocol/realtime/world_lane_state.gd` - Owns deleted-bullet tombstones and pending unknown-bullet updates.
* `client/scripts/world/projectile_sync.gd` - Owns projectile deletion tombstones.
* `client/scripts/world/asteroid_sync.gd` - Owns asteroid deletion tombstones.
* `client/tools/long_match_store_memory_probe.gd` - Exercises the seven production stores headlessly.

## Tests

* `client/tests/unit/devtools/telemetry/test_long_match_store_metrics.gd` - Snapshot counts and aggregate total coverage.
* `client/tests/unit/protocol/realtime/test_event_batch_and_resync.gd` - Event-history caps, eviction, and duplicate-record behavior.
* `client/tests/unit/protocol/realtime/test_world_lane_applier.gd` - Bullet tombstone and pending-update caps, recreation, replacement ordering, and clear behavior.
* `client/tests/unit/world/test_projectile_sync.gd` - Projectile tombstone cap, recreation, and reset behavior.
* `client/tests/unit/world/test_asteroid_sync.gd` - Asteroid tombstone cap, recreation, and reset behavior.

## Related docs

* [Client Devtools](./!INDEX.md)
* [Gameplay State Application](../../services/client/gameplay-runtime/gameplay-state-application.md)
* [Entity Sync Owners](../../services/client/world-sync/entity-sync-owners.md)
* [Gameplay Events and Effects](../../services/client/gameplay-event-presentation/gameplay-events-and-effects.md)

## Notes

The diagnostic is intentionally scoped to bounded client retention stores. A plateau in `total_entries` describes these seven stores after their configured caps are reached; it is not a general memory or performance guarantee for the client.
