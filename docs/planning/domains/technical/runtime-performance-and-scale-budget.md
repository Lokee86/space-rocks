# Runtime Performance And Scale Budget

Parent index: [Technical Planning](./!INDEX.md)

## Purpose

This doc plans runtime performance and scale-budget policy for Space Rocks.

It defines what runtime pressure must be measured, how those measurements should be gathered, and how runtime evidence gates packaged beta, dev-hosted multiplayer, staging, and production readiness.

This doc is optimization-adjacent. It defines measurement coverage and decision gates that determine when optimization is needed, but it does not prescribe optimization tactics before evidence identifies the limiting pressure.

## Overview

This doc tracks runtime pressure, measurement coverage, and release-shaped performance gates so growth stays evidence-based.

## Current status

Active planning.

## Ownership Boundary

This doc owns planning for server tick cost, client frame cost, entity-count pressure, room/player scale, memory/resource growth, runtime load scenarios, measurement methods, and launch-shaped performance gates.

Packet-size policy and packet-budget measurement stay in [Network Observability And Packet Budget](network-observability-and-packet-budget.md). This doc may treat packet pressure as a related runtime signal, but it does not own packet-format redesign or network optimization strategy.

General testing policy belongs in verification and quality-gates planning. This doc owns runtime pressure measurement, not correctness testing strategy.

## Does Not Belong

* Exact packet format changes.
* Delta compression, protobuf, or lane-splitting strategy.
* Exact Go optimization implementation.
* Exact Godot rendering optimization implementation.
* Exact cloud or hosting capacity planning.
* Gameplay balance numbers.
* General correctness-test policy.
* Current implementation authority.

## Runtime Measurement Model

Runtime-heavy features should not be considered safe to expand until Space Rocks can measure their server tick cost, client frame cost, entity-count pressure, and memory/resource impact.

The first goal is not to measure everything. The first goal is to establish stable runtime signals that reveal when gameplay growth, multiplayer growth, or release-shaped builds are becoming unsafe.

Initial runtime measurement should focus on:

* server tick cost,
* client frame cost,
* entity counts,
* room and player scale,
* memory and resource growth,
* runtime degradation over repeated or long-running sessions.

## Measurement Methods

| Pressure                  | Measurement Method                                                                                    |
| ------------------------- | ----------------------------------------------------------------------------------------------------- |
| Server tick cost          | Server timing counters per room or tick, slow-tick warnings, max and rolling-window timing summaries. |
| Simulation subsystem cost | Timed sections around movement, collision, spawning, effects, scoring, and match state where cheap.   |
| Entity pressure           | Per-room entity counters emitted through runtime summaries, diagnostics, or devtools.                 |
| Room/player scale         | Scripted or synthetic multi-room runs with active, idle, joining, leaving, and ending rooms.          |
| Client frame cost         | Godot frame-time or FPS sampling, world-sync timing, visible entity counts, and node/render pressure. |
| Spawn/despawn churn       | Client and server counters over short rolling windows.                                                |
| Memory/resource growth    | Periodic process memory sampling and longer-running soak scenarios.                                   |
| Release candidate health  | Repeatable runtime scenarios before release-shaped builds.                                            |

Measurements should start lightweight. Logs, counters, devtools overlays, and controlled scenarios are enough before a larger benchmark system exists.

## Server Runtime Signals

Server-side measurement should track pressure in the authoritative simulation.

Useful signals include:

* tick duration,
* simulation/update duration,
* collision duration where cheap,
* active room count,
* active player count,
* entity counts per room,
* match start/end churn,
* slow tick count,
* max tick duration over a short window,
* memory use where cheap.

Entity counts should include current and future pressure sources:

* players,
* bullets,
* asteroids,
* asteroid fragments,
* pickups,
* enemies,
* mines,
* drones,
* radial effects,
* gameplay events,
* spectators.

## Client Runtime Signals

Client-side measurement should track whether the Godot client can render and sync the game smoothly.

Useful signals include:

* frame time or FPS,
* synced entity count,
* visible entity counts,
* spawn/despawn churn,
* world sync update cost where cheap,
* node/render pressure,
* memory use where cheap,
* telemetry overlay impact where relevant.

Client frame health and server tick health are separate gates. A feature can be acceptable on the server and still blocked by client render pressure, or acceptable on the client and still blocked by server simulation cost.

## Measurement Types

| Type                       | Role                                                                                             |
| -------------------------- | ------------------------------------------------------------------------------------------------ |
| Manual Measurement         | Early devtools checks while playing known scenarios.                                             |
| Automated Runtime Scenario | Repeatable scripted scenario for release candidates.                                             |
| Synthetic Load             | Fake players, entities, rooms, or events used to pressure systems without relying on real users. |
| Soak Run                   | Longer run used to catch memory growth, timing drift, lifecycle leaks, or retry buildup.         |

Manual measurement is acceptable early. Release-shaped builds should move toward repeatable scenarios so runtime readiness is not judged only by feel.

## Launch-Shape Runtime Expectations

| Build Or Environment   | Runtime Expectation                                                                                                |
| ---------------------- | ------------------------------------------------------------------------------------------------------------------ |
| Local Packaged Beta    | Single-player is playable, diagnosable, and does not visibly degrade under expected local load.                    |
| Dev-Hosted Multiplayer | Multiplayer flows run under controlled test load with runtime warnings visible.                                    |
| Hosted Staging         | Production-like runtime checks and load scenarios pass before production promotion.                                |
| Hosted Production      | Runtime gates are part of release readiness; degraded performance triggers operational downgrade or release block. |

Current baseline measurement is the first reference point, not the final target.

This plan should support launch-shaped confidence from local packaged beta through hosted production, even if early implementation starts with smaller measurement slices.

## Required Coverage By Release Shape

### Local Packaged Beta

Local packaged beta should have runtime coverage for:

* bundled local server startup,
* single-player room simulation,
* normal asteroid gameplay,
* many bullets,
* many asteroids,
* bullets plus asteroids plus pickups,
* match start/end flow,
* local result/profile flow,
* client world sync and frame pressure,
* local session cleanup.

### Dev-Hosted Multiplayer

Dev-hosted multiplayer should have runtime coverage for:

* multiplayer room creation and admission,
* multiple players in one room,
* ready/start/end flow,
* player join and leave churn,
* match result write pressure,
* packet pressure cross-reference,
* server tick pressure under multiplayer load,
* client sync pressure with multiple players.

### Hosted Staging

Hosted staging should have production-like runtime coverage for:

* hosted API, game-server, and player-data interaction,
* compatibility admission,
* controlled multiplayer load,
* repeated match start/end cycles,
* result write and retry pressure,
* telemetry and logging availability,
* degraded-service visibility where relevant.

### Hosted Production

Hosted production should require runtime gates before promotion.

Production candidates should not pass if runtime measurement shows unresolved blocker-level pressure in core gameplay, hosted multiplayer admission, player-data result flow, or client/server runtime health.

## Load Scenarios

Runtime load scenarios should be fake but realistic.

Useful scenarios include:

* many bullets,
* many asteroids,
* bullets plus asteroids plus pickups,
* multiple players,
* repeated match start/end cycles,
* many spawned/despawned entities,
* long-running local sessions,
* player-data write/retry pressure,
* later enemy and boss pressure,
* later mines, drones, and radial effects,
* later multiple active rooms.

These are not correctness tests. They are controlled pressure scenarios used to show where runtime cost appears.

## Runtime-Heavy Feature Gates

Runtime-heavy features should include measurement before they are treated as safe to expand.

This applies to:

* enemies,
* bosses,
* bullet hell patterns,
* mines,
* drones,
* radial effects,
* larger multiplayer rooms,
* spectators,
* competitive modes,
* progression-heavy match results,
* repeated match start/end churn.

A feature can begin small before full scale testing exists, but it should not expand into a major system without measurement coverage.

## Decision States

Exact numeric thresholds do not need to be final yet.

Initial scale planning should use decision states.

| State           | Meaning                                                 |
| --------------- | ------------------------------------------------------- |
| Healthy         | Works with no meaningful runtime concern.               |
| Warning         | Works, but measurable pressure appears.                 |
| Blocked         | Do not expand this area until performance work happens. |
| Needs Load Test | Cannot judge safely without a pressure scenario.        |

Hard limits should be added only after baseline measurement gives enough evidence.

## Competitive Mode Expectations

Competitive modes need stricter runtime confidence than casual or local modes.

Performance drops in competitive modes can affect fairness, rankings, rewards, disputes, and trust. Competitive mode planning should therefore require stronger tick, frame, recovery, and diagnostic expectations before public use.

Casual and local modes may tolerate looser thresholds during early development and beta testing.

## Devtools And Visibility

Runtime pressure should eventually be visible through devtools or diagnostics.

The World Telemetry Overlay may grow beyond packet metrics to show useful runtime pressure such as entity counts, frame pressure, and server/runtime timing where available.

This remains development and diagnostics tooling. It should not become player-facing HUD behavior by default.

## Optimization Boundary

This doc should not choose optimization tactics early.

Optimization work should be selected after measurement identifies the limiting pressure.

Possible future optimization areas may include simulation cost, collision cost, rendering cost, entity lifecycle churn, packet size, room/process scale, or memory growth, but those choices belong in later implementation or system-specific planning after evidence exists.

## Implementation sequence

1. Keep the initial runtime signals lightweight and focused on server tick, client frame, entity-count, and memory pressure.
2. Use manual measurement and devtools overlays first, then grow toward repeatable runtime scenarios.
3. Apply the launch-shape coverage matrix to local packaged beta, dev-hosted multiplayer, hosted staging, and hosted production.
4. Add decision states and load scenarios as the evidence base grows.
5. Treat optimization as a follow-on choice after the limiting pressure is measured.

## Related Docs

* [Planning](../../!INDEX.md)
* [Development Roadmap](../../development-roadmap.md)
* [Network Observability And Packet Budget](network-observability-and-packet-budget.md)
* [Operational Readiness And Failure Modes](operational-readiness-and-failure-modes.md)
* [Compatibility Versioning And Migrations](compatibility-versioning-and-migrations.md)
* [Build Release And Environment Matrix](build-release-and-environment-matrix.md)
* [Devtools And Telemetry](devtools-and-telemetry.md)
* [Game Server Simulation](../../services/game-server/simulation/!INDEX.md)
* [Game Server Networking](../../services/game-server/networking/!INDEX.md)
* [Gameplay Runtime](../../services/client/gameplay-runtime/!INDEX.md)
* [World Sync](../../services/client/world-sync/!INDEX.md)

## Open decisions

* Which first runtime scenarios should be automated versus manual?
* Which runtime signals should appear in the World Telemetry Overlay?
* Which slow-tick or frame-pressure thresholds should become release gates?
* Which entity-heavy feature should get the first dedicated load scenario?
* When should multi-room process scale become a gate instead of single-room health?
* Which competitive modes require stricter runtime thresholds?
* What minimum runtime coverage is required before hosted staging can promote to production?

## Notes

Keep this doc focused on measurement, release gates, and decision thresholds rather than early optimization tactics.
