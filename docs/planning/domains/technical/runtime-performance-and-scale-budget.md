---
author: brian
created: "2026-07-19"
document_id: 019f7d55-fb2c-7483-b2fd-ff87e5d8c898
document_type: general
policy_exempt: false
summary: This doc plans runtime performance and scale-budget policy for Space Rocks.
---
# Runtime Performance And Scale Budget

Parent index: [Technical Planning](./!INDEX.md)

## Purpose

This doc plans runtime performance and scale-budget policy for Space Rocks.

It defines what runtime pressure must be measured, how those measurements should be gathered, and how runtime evidence gates packaged alpha, dev-hosted multiplayer, staging, and production readiness.

This doc is optimization-adjacent. It defines measurement coverage and decision gates that determine when optimization is needed, but it does not prescribe optimization tactics before evidence identifies the limiting pressure.

## Overview

This doc tracks runtime pressure, measurement coverage, and release-shaped performance gates so growth stays evidence-based.

## Current status

Active planning. Deterministic game-owned RNG, runtime measurement capture, and the scripted runtime scenario harness are implemented. The lifecycle regression scenario exercises real multiplayer admission, WebRTC delivery, server-owned bots, sustained bullet pressure, interest transitions, death, and spectating. A receiver-scaling matrix holds eight total participants constant while varying one, two, and four real clients, and a simulation-heavy isolation scenario now holds receiver count to one while pre-seeding asteroid pressure and increasing projectile streams. Match-churn, soak, multi-room coverage, and release thresholds remain.

## Ownership Boundary

This doc owns planning for server tick cost, client frame cost, entity-count pressure, room/player scale, memory/resource growth, runtime load scenarios, scripted/synthetic runtime scenario harness work, measurement methods, and launch-shaped performance gates.

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

The first goal is not to measure everything. The first goal is to establish stable runtime signals that reveal when gameplay growth, multiplayer growth, or release-shaped builds are becoming unsafe. Runtime measurement stays centered on server tick, client frame, entity counts, room/player scale, memory/resource growth, lifecycle churn, and soak behavior. The implemented scripted runtime harness now exercises those signals repeatably; the next work is broader scenario coverage and evidence-based thresholds.

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

## Current Measured Baseline

Baseline recorded 2026-07-12 from `services/game-server/internal/game/presentation_snapshot_benchmark_test.go` on Linux/amd64 with an 11th Gen Intel Core i9-11900H. These measurements are local benchmark evidence, not release limits.

### Pre-optimization baseline

Before the shared presentation frame, `GameplayPresentationSnapshot` copied the authoritative presentation maps while holding `Game.mu` for every receiver request. This is retained as the pre-optimization baseline that explains the presentation ownership seam. Representative results from three 250 ms runs were:

| Scenario | Representative snapshot cost | Allocations |
| --- | ---: | ---: |
| 1 player, 100 asteroids, 100 bullets | 10.8 us/op | 25,904 B/op, 18 allocs/op |
| 8 players, 100 asteroids, 500 bullets | 46.0 us/op | 129,904 B/op, 33 allocs/op |
| 16 players, 500 asteroids, 2,000 bullets | 242 us/op | 558,944 B/op, 61 allocs/op |

The `Game.Step` contention benchmark freezes world movement, spawning, and collisions so it measures aggregate lock acquisition and basic tick work rather than entity-pair collision work. Representative results from three 100 ms runs were:

| Scenario | 0 readers | 1 reader | 4 readers | 8 readers |
| --- | ---: | ---: | ---: | ---: |
| 1 player, 100 asteroids, 100 bullets | 358 ns/op | 1.18 us/op | 3.85 us/op | 7.21 us/op |
| 8 players, 100 asteroids, 500 bullets | 1.09 us/op | 3.01 us/op | 11.6 us/op | 22.2 us/op |
| 16 players, 500 asteroids, 2,000 bullets | 2.32 us/op | 5.87 us/op | 22.0 us/op | 32.8 us/op |

### Post-implementation baseline

The implemented shared presentation frame changes the measured boundary. Across the same small, medium, and stress scenarios, the receiver-scoped snapshot wrapper is approximately 34 ns/op, 0 B/op, and 0 allocs/op. Shared-frame publication is approximately:

| Scenario | Publication cost | Allocations |
| --- | ---: | ---: |
| 1 player, 100 asteroids, 100 bullets | 11.6 us/op | 25,984 B/op |
| 8 players, 100 asteroids, 500 bullets | 45.7 us/op | about 129,985 B/op |
| 16 players, 500 asteroids, 2,000 bullets | 225.7 us/op | about 559,036 B/op |

Publication occurs once per simulation generation rather than once per receiver request. Per-receiver pending-event copying and realtime lane planning remain separate costs. The duplicated per-receiver snapshot-copy pressure is addressed; future measurement gates remain frame-publication cost, delta/projection cost, event copying, GC behavior, larger rooms, and multi-room process scale.

Contention readers request snapshots continuously and are deliberately harsher than normal per-session requests at 60 Hz. They no longer multiply entity-map copying. Remaining `Step` cost is dominated by once-per-step frame publication and simulation work; reader count adds only brief lock access for frame capture and pending-event copying. The representative 8-player load is healthy relative to the 16.67 ms server tick budget, while the 16-player stress case remains evidence for future scale work.

The benchmark player count is an entity-count dimension, not proof of the same number of connected network clients. Server-owned bots run their input decisions inside `Game.Step()` and do not create WebSocket readers, asynchronous input timing, receiver-scoped realtime state, per-session candidate construction, per-session encoding, or DataChannel buffers. Runtime evidence must report both authoritative player/entity count and actual connected-client count.

The canonical multiplayer regression matrix should hold world pressure approximately constant while varying ingress and receiver count:

```text
A: one connected client plus server-owned bots
B: one actively controlled network client with equivalent entity pressure
C: two or more actively controlled network clients with equivalent entity pressure
```

Each run should capture server tick duration, input mailbox replacements, presentation generation cadence, per-session candidate-build and encode time, skipped send ticks, per-lane buffered amount, client packet-application time, and client frame time. This separates simulation/entity capacity from asynchronous input and per-receiver networking capacity.

Revisit this decision when rooms grow larger, sustained entity counts rise, runtime observation shows GC or memory pressure, snapshot work consumes a material portion of the 16.67 ms tick budget, or real connected-session count materially changes planner, encoding, or write cadence.

### Collision Broad-Phase Measurement Foundation

The implemented collision path now uses a toroidal uniform grid behind the generic `spatial.Index` contract for projectile/asteroid, ship/asteroid, and player/pickup candidate discovery. Exact physics checks remain authoritative after candidate lookup. This removes the direct all-target narrow-phase scan from the current collision handlers, but it does not establish a hard production scale limit or measured multi-room capacity.

`collision_spatial_benchmark_test.go` compares brute-force and spatial candidate paths at 100 asteroids/500 projectiles and 500 asteroids/2,000 projectiles. Each spatial benchmark operation includes one asteroid-index rebuild. No benchmark results are recorded here until the benchmark has been executed in a Go-capable environment; results must be captured from real benchmark output rather than inferred.

The existing `Game.Step` contention benchmark remains separate. It freezes collisions, so it measures aggregate lock contention and basic step work rather than collision broad-phase cost.

The canonical spatial-query ownership and behavior are documented in [Spatial Query Index](../../../services/game-server/simulation/world/spatial-query-index.md).

Exact benchmark commands:

```bash
cd services/game-server
go test ./internal/game -run '^$' -bench '^BenchmarkGameplayPresentationSnapshot$' -benchtime=250ms -count=3
go test ./internal/game -run '^$' -bench '^BenchmarkGameplayPresentationFramePublication$' -benchtime=250ms -count=3
go test ./internal/game -run '^$' -bench '^BenchmarkGameStepWithPresentationSnapshotContention$' -benchtime=100ms -count=3
go test ./internal/game -run '^$' -bench '^BenchmarkProjectileAsteroidCollisionBroadPhase$' -benchtime=250ms -count=3
```

Current presentation-frame ownership and publication behavior are documented canonically in [Game Aggregate](../../../services/game-server/simulation/runtime/game-aggregate.md).

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

## Implemented Runtime Scenario Harness

The initial harness lives under `tools/runtime_scenarios/` and runs the real client/server product path rather than a separate simulation.

The canonical lifecycle regression scenario is `network_interest_lifecycle_v1`. It currently uses:

* one coordinator client,
* one additional real headless client,
* six server-owned bots,
* a fixed simulation seed,
* normal authenticated room create/join/ready/start flow,
* real WebRTC gameplay and tooling channels,
* sustained continuous-bullet-stream pressure,
* scripted movement and firing,
* interest-transition pressure,
* authoritative death and spectate targeting,
* existing client/server runtime measurement exporters.

The runner reserves an isolated loopback port, starts the current game-server source through WSL `go run` on Windows or direct `go run` on non-Windows hosts, enables harness-only authentication and deterministic seed injection through process environment, starts the configured clients, collects status and measurement artifacts, and terminates the complete process trees it started. Run artifacts are written under `.ci-artifacts/runtime-scenarios/`.

The coordinator is visible by default so its client report can measure actual rendering pressure. `--headless-coordinator` is available for unattended orchestration and network/lifecycle verification, but a headless run is not a client render-performance baseline.

Complete server sample history is persisted to the server report on disk. Server measurement report version 6 records per-receiver candidate-build, encoding, and total outbound duration; exclusive candidate-build timing for snapshot capture, pending-event copying, interest filtering, lane-candidate assembly, chunk planning, and scheduling; nested lane-candidate timing for hot-tick state advancement, world/hot/lifecycle construction, player locators, overlay, session, event, and candidate finalization; the complete phase breakdown from the exact peak candidate-build tick; receiver skipped-send ticks; per-lane current and peak DataChannel buffered bytes; and per-lane skipped-send counts. The final `measurement_stopped` tooling response carries a bounded summary plus the export path rather than embedding the entire sample history in one control message. Scenario summaries and `phase-markers.json` preserve configured phase boundaries for later correlation with samples and logs.

Runtime-scenario clients run the source project through the Godot editor executable. Scenario startup skips the bundled local-alpha server, so packaged/exported Space Rocks executables are not launched. On Windows, the harness starts the current server source through WSL `go run ./cmd/game-server`, keeping the server as a Linux process instead of allowing native Windows `go run` to create a temporary `game-server.exe`. Non-Windows hosts use direct `go run`.

The receiver-scaling matrix uses the same seed, eight total participants, matching phase durations, and matching 30-stream pressure:

```text
receiver_scale_1c_7b_v1: one real client and seven bots
receiver_scale_2c_6b_v1: two real clients and six bots
receiver_scale_4c_4b_v1: four real clients and four bots
```

`tools/runtime_scenarios/matrix_summary.py` aggregates the resulting client and server reports into one receiver-scaling summary. Headless runs remain useful for network, packet-application, process, and lifecycle pressure, but their frame values are not render-performance baselines.

### Initial receiver-scaling evidence

The current local headless matrix was recorded on 2026-07-28 with Godot 4.6.3 clients and the game server running through WSL. All clients and the server ran on the same development machine, so client-process and cross-VM host contention remain part of the result.

| Real clients | Bots | Server tick avg | Server tick max | Candidate build avg | Candidate build max | Encode avg | Outbound avg | Outbound max |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 1 | 7 | 209.9 us | 8.22 ms | 1.285 ms | 7.80 ms | 0.187 ms | 1.554 ms | 8.28 ms |
| 2 | 6 | 262.8 us | 4.13 ms | 1.698 ms | 10.43 ms | 0.236 ms | 2.038 ms | 11.60 ms |
| 4 | 4 | 366.2 us | 27.76 ms | 2.856 ms | 43.42 ms | 0.298 ms | 3.281 ms | 44.27 ms |

| Real clients | Receiver bytes | Skipped receiver ticks | Highest lane buffer | Mean headless frame | Worst headless p99 |
| ---: | ---: | ---: | ---: | ---: | ---: |
| 1 | 11.0 MB | 0 | 1,199 B | 7.05 ms | 16 ms |
| 2 | 21.9 MB | 0 | 1,200 B | 7.94 ms | 33 ms |
| 4 | 41.7 MB | 0 | 1,200 B | 12.66 ms | 100 ms |

Receiver bytes scale approximately with real-client count. No client reported a send failure, no receiver tick was skipped for backpressure, and the largest observed lane buffer was 1,200 bytes against the 32 KiB server backpressure boundary. DataChannel congestion is therefore not the current receiver-scale limit in this scenario.

Within the measured receiver pipeline, candidate construction is the dominant cost. Encoding remains comparatively small, while average candidate-build time rises from 1.285 ms with one receiver to 2.856 ms with four. The four-client receiver maximum of 44.27 ms is almost entirely explained by its 43.42 ms candidate-build maximum. The older 173.9 ms server-tick maximum did not reproduce; the controlled rerun recorded 27.76 ms. Because all clients and the WSL server share one host, these results do not yet separate candidate-path algorithmic contention from OS scheduling and client-process contention.

A follow-up report-v4 matrix split candidate construction into exclusive phases. Average candidate time was dominated by lane-candidate assembly, with chunk planning second:

| Real clients | Snapshot | Events | Interest | Lane candidates | Chunk planning | Scheduling |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 1 | 0.004 ms | 0.003 ms | 0.080 ms | 1.359 ms | 0.583 ms | 0.015 ms |
| 2 | 0.007 ms | 0.003 ms | 0.089 ms | 1.780 ms | 0.741 ms | 0.016 ms |
| 4 | 0.005 ms | 0.002 ms | 0.094 ms | 1.994 ms | 0.785 ms | 0.018 ms |

Report version 5 then preserved the phase breakdown from the exact worst candidate-build tick. In the correlated four-client rerun, the peak candidate build was 37.432 ms: 35.369 ms of lane-candidate assembly, 1.328 ms of chunk planning, 0.625 ms of snapshot capture, 0.090 ms of interest filtering, 0.004 ms of event copying, and 0.014 ms of scheduling. Lane-candidate assembly therefore accounted for 94.5% of that exact peak.

Report version 6 split lane-candidate assembly itself. In the correlated four-client run, average lane-candidate construction was 2.066 ms: 1.952 ms for world/hot/lifecycle construction, 0.084 ms for session construction, 0.016 ms for overlay, 0.007 ms for player locators, 0.003 ms for events, and negligible state-advance/finalization work. The exact 42.642 ms peak candidate-build tick contained 42.594 ms of lane-candidate assembly, of which 42.513 ms was world/hot/lifecycle construction. That path therefore accounted for 99.7% of the exact peak candidate build and 95.2% of average lane-candidate construction.

The implemented fix now caches one immutable quantized full-world projection per published `Game` presentation generation. Adjacent generations are retained because receiver write loops can briefly overlap frames. Receiver-specific interest filtering, baselines, sequences, cadence, hot-route state, lifecycle selection, and deltas remain isolated. This removes repeated entity-ID sorting, record construction, and float quantization from each receiver path.

Two post-fix four-client runs confirmed the improvement under differing world pressure. World/hot/lifecycle average fell from 1.952 ms to 1.067 ms and 1.298 ms. Its peak fell from 42.513 ms to 5.849 ms and 15.903 ms. Total candidate-build average fell from 3.038 ms to 1.722 ms and 1.990 ms. The heavier confirmation run moved its largest candidate-build outlier to chunk planning: 25.051 ms of a 29.377 ms peak, while world/hot/lifecycle used 3.943 ms on that exact tick. The repeated per-receiver world projection problem was therefore materially reduced.

Chunk planning was then changed from repeated growing-packet measurement to bounded planning. Hot movement lanes now account for compact-record bytes in one pass. Full-world and lifecycle candidates first measure the complete normalized packet once; packets that already fit no longer rebuild and encode every prefix. Oversized packets use binary range packing, followed by the existing final encoded-size hard-cap validation. In a four-client rerun with up to 127 asteroids and 189 projectiles, chunk-planning average fell from 0.533 ms to 0.069 ms and its peak fell from 25.051 ms to 2.502 ms. Total candidate-build average fell to 0.631 ms and its peak to 4.637 ms. No receiver sends were skipped and no hard-cap or delivery failures occurred.

The `simulation_scale_1c_7b_v1` isolation scenario seeds 192 asteroids before measurement, uses one real headless receiver plus seven bots, and increases continuous projectile streams from 60 to 120. The first run sampled up to 144 surviving asteroids and 518 projectiles. Server tick time remained healthy at 0.280 ms average and 6.554 ms maximum, with 45.2 MiB peak RSS and 0.482 peak CPU cores. Receiver candidate construction averaged 1.345 ms and peaked at 7.811 ms; outbound work averaged 1.637 ms and peaked at 9.436 ms. No sends were skipped and no client send failures occurred. The headless client recorded a 33 ms frame-time p99 and 76.226 ms maximum, so this load does not identify authoritative simulation as the current limit; visible-client rendering and packet-application behavior should be checked separately if high-entity stutter remains.

Canonical commands:

```bash
python tools/runtime_scenarios/main.py \
  tools/runtime_scenarios/scenarios/network_interest_lifecycle_v1.json \
  --validate-only

python tools/runtime_scenarios/main.py \
  tools/runtime_scenarios/scenarios/network_interest_lifecycle_v1.json
```

## Launch-Shape Runtime Expectations

| Build Or Environment   | Runtime Expectation                                                                                                |
| ---------------------- | ------------------------------------------------------------------------------------------------------------------ |
| Local Packaged Alpha    | Single-player is playable, diagnosable, and does not visibly degrade under expected local load.                    |
| Dev-Hosted Multiplayer | Multiplayer flows run under controlled test load with runtime warnings visible.                                    |
| Hosted Staging         | Production-like runtime checks and load scenarios pass before production promotion.                                |
| Hosted Production      | Runtime gates are part of release readiness; degraded performance triggers operational downgrade or release block. |

Current baseline measurement is the first reference point, not the final target.

This plan should support launch-shaped confidence from local packaged alpha through hosted production, even if early implementation starts with smaller measurement slices.

## Required Coverage By Release Shape

### Local Packaged Alpha

Local packaged alpha should have runtime coverage for:

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
* multiple authoritative players or server-owned bots in one room,
* multiple real connected clients in one room as a separate load dimension,
* actively changing network input rather than idle connected sessions only,
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

These are not correctness tests. They are controlled pressure scenarios used to show where runtime cost appears. Scenario evidence should carry the scenario configuration and the seed so the same random-choice stream can be reproduced.

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

The World Telemetry Overlay may grow beyond packet metrics to show useful runtime pressure such as entity counts, frame pressure, and server/runtime timing where available. Seeded scenario evidence can preserve the configuration and seed locally. Seed emission through canonical observability events or diagnostic bundles remains deferred on its own product/diagnostic merits; the observability contract and emitter boundary are settled.

This remains development and diagnostics tooling. It should not become player-facing HUD behavior by default.

## Optimization Boundary

This doc should not choose optimization tactics early.

Optimization work should be selected after measurement identifies the limiting pressure.

Possible future optimization areas may include simulation cost, collision cost, rendering cost, entity lifecycle churn, packet size, room/process scale, or memory growth, but those choices belong in later implementation or system-specific planning after evidence exists. Seeded RNG is a runtime-scenario foundation, not replay support, and it does not by itself guarantee full cross-build simulation determinism.

## Implementation sequence

1. Keep the initial runtime signals lightweight and focused on server tick, client frame, entity-count, and memory pressure.
2. Use the implemented seeded runtime harness for repeatable multiplayer pressure while retaining manual measurement for exploratory checks.
3. Expand the scenario catalog beyond the implemented simulation-heavy and receiver-heavy cases to cover match churn, soak, and eventually multi-room pressure separately.
4. Apply the launch-shape coverage matrix to local packaged alpha, dev-hosted multiplayer, hosted staging, and hosted production.
5. Add evidence-based decision thresholds as the scenario baseline grows.
6. Treat optimization as a follow-on choice after the limiting pressure is measured.

## Related Docs

* [Planning](../../!INDEX.md)
* [Development Roadmap](../../development-roadmap.md)
* [Network Observability And Packet Budget](network-observability-and-packet-budget.md)
* [Operational Readiness And Failure Modes](operational-readiness-and-failure-modes.md)
* [Compatibility Versioning And Migrations](compatibility-versioning-and-migrations.md)
* [Build Release And Environment Matrix](build-release-and-environment-matrix.md)
* [Devtools And Telemetry](../../devtools/devtools-and-telemetry.md)
* [Game Server Simulation](../../services/game-server/simulation/!INDEX.md)
* [Deterministic Gameplay RNG Runtime](../../../services/game-server/simulation/runtime/deterministic-gameplay-rng.md)
* [Game Server Networking](../../services/game-server/networking/!INDEX.md)
* [Gameplay Runtime](../../services/client/gameplay-runtime/!INDEX.md)
* [World Sync](../../services/client/world-sync/!INDEX.md)

## Open decisions

* Which additional scenario should be automated next: match churn, soak, or multi-room pressure?
* Which runtime signals should appear in the World Telemetry Overlay?
* Which slow-tick or frame-pressure thresholds should become release gates?
* Which entity-heavy feature should get the next dedicated load scenario?
* When should multi-room process scale become a gate instead of single-room health?
* Which competitive modes require stricter runtime thresholds?
* What minimum runtime coverage is required before hosted staging can promote to production?

## Notes

Keep this doc focused on measurement, release gates, and decision thresholds rather than early optimization tactics.
