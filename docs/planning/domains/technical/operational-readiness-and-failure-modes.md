# Operational Readiness And Failure Modes

Parent index: [Technical Planning](./!INDEX.md)

## Purpose

This doc plans operational readiness and technical failure behavior for Space Rocks.

It defines how services report health, how environments enter degraded states, what happens when dependencies fail, and what diagnostic information must exist to understand and recover from failures.

Compatibility planning answers whether versions can safely run together. Operational readiness planning answers what happens when an allowed system fails anyway.

## Overview

This doc keeps health, readiness, degraded-state, and failure-mode policy aligned so player-facing services fail visibly and recoveries stay diagnosable.

## Current status

Active planning.

## Ownership Boundary

This doc owns planning for service health, startup readiness, unavailable-service behavior, degraded modes, restart and recovery assumptions, rollback expectations, operational diagnostics, and future incident-management readiness.

Abuse, moderation, and game-integrity enforcement belong under platform security and admin planning. Exact hosting-provider setup and service-specific implementation details belong in the relevant service docs.

## Does Not Belong

* Abuse or moderation policy.
* Anti-cheat or game-integrity enforcement.
* Exact hosting-provider infrastructure.
* Exact database migration implementation.
* Exact protocol compatibility rules.
* CI command details.
* Player support policy.
* Current implementation authority.

## Readiness Model

Operational readiness should be checked before player-facing services admit users or serve production traffic where practical.

Health checks should not use one vague healthy/unhealthy flag. Different checks should answer different questions.

| Check             | Meaning                                      |
| ----------------- | -------------------------------------------- |
| Liveness          | The process is running.                      |
| Readiness         | The service can accept users or work.        |
| Dependency Health | Required backing services are reachable.     |
| Admission Health  | The environment is allowed to admit players. |

A service can be live but not ready. A hosted environment can be technically running but still block admission because auth, player-data, telemetry, or configuration is not safe.

## Environment States

Hosted environments should support simple operational states.

| State             | Meaning                                                                    |
| ----------------- | -------------------------------------------------------------------------- |
| Healthy           | Normal operation.                                                          |
| Degraded          | Core service works, but something important is impaired.                   |
| Admission Blocked | Existing sessions may continue, but new users or matches should not enter. |
| Offline           | Service should not serve players.                                          |

Maintenance should use an intentional admission-blocked or offline state instead of being treated as an unexpected crash.

## Failure Surfaces

Operational planning applies to:

* client startup and connection failures,
* bundled local server launch failures,
* local profile save or migration failures,
* API-server failures,
* auth-provider failures,
* game-server failures,
* player-data failures,
* web service failures,
* telemetry or logging failures,
* configuration and secret failures,
* hosted dependency failures.

External dependency failures should fail visibly and conservatively. They should not be reported as normal gameplay errors.

## Local Single-Player Failure Policy

Local packaged single-player should not require hosted API, hosted player-data, auth, matchmaking, or web services.

If the bundled local server and local profile storage work, local single-player should work.

If the bundled single-player server cannot start, the client should fail visibly with a clear local-runtime error rather than hanging or pretending the game is loading.

Local profile save, load, and migration failures should also fail visibly. They should avoid silent progress loss and should expose diagnostics where automatic reporting is unavailable.

## Hosted Multiplayer Failure Policy

Hosted multiplayer should fail closed when auth, compatibility admission, or account identity checks are unavailable.

If the game-server is unavailable, hosted multiplayer should block admission or disconnect cleanly with a visible error.

Active match recovery is not required for first implementation. It should remain planned for competitive modes where match loss can create fairness, ranking, reward, or dispute problems.

Casual and local modes may initially use clean failure, visible error handling, and diagnostics rather than full active-match recovery.

## Player-Data Failure Policy

If player-data is unavailable, gameplay may continue where safe, but players must see a clear warning that accumulated stats, items, currency, rankings, or rewards may not be awarded until service recovery.

Player-data failure must not silently discard progress or reward-bearing results.

Local and hosted writes should prefer pending/retry over silent loss where practical.

## Pending And Retry Policy

Pending/retry should be the default direction for result and progression writes.

This applies to:

* match results,
* stats,
* rewards,
* currency,
* item grants,
* achievements,
* ranking changes where applicable.

Retry paths must be idempotent before they become reward-bearing. Retrying a pending write must not duplicate stats, items, currency, achievements, rankings, or other durable grants.

If retry is impossible, the system should fail visibly and preserve enough diagnostic or audit context to investigate the failure.

## Telemetry And Logging Readiness

Before launch or during release-candidate readiness, telemetry or logging failure is a product-readiness block.

After launch, telemetry or logging failure should trigger a high-priority environment downgrade. It does not necessarily require immediate shutdown if core gameplay and data safety remain intact.

Operational logs should identify what failed, where it failed, and which service, build, version, session, or environment was involved. They should not dump secrets, bearer tokens, raw private profile data, or unnecessary full payloads by default.

## Bug Reports And Copy Diagnostics

Major failures that cannot be automatically logged or reported should expose a copy diagnostics or bug-report path.

This applies to failures such as:

* local packaged server launch failure,
* local profile save or migration failure,
* connection failure,
* auth or admission failure,
* player-data write failure,
* unexpected disconnect,
* package or runtime mismatch,
* major client startup failure.

Diagnostics should include enough service, build, session, and environment context to debug the issue while avoiding secrets and unnecessary private data.

## Player-Facing Status

Hosted outages and degraded states should eventually have a player-visible status path.

The first version can be simple, such as an in-client warning or website status note. A dedicated status page can come later if public hosted production needs it.

## Rollback And Recovery

Code rollback and data rollback are separate.

Rolling back a service should not assume migrated data can be downgraded unless a downgrade path was explicitly planned.

Hosted recovery planning should preserve player progress and account integrity where practical. Data recovery boundaries belong in data and service docs, but this doc owns the cross-cutting rule that recovery must be considered before production rollout.

## Configuration And Secrets

Missing, invalid, or unsafe production configuration should fail startup or readiness checks.

A service should not continue into partially broken production operation when required configuration, credentials, secrets, dependency URLs, or environment flags are missing or unsafe.

## Incident Platform Future-Proofing

A full incident-management platform is not required for early local or beta builds.

Hosted production should still be planned with the assumption that service health, readiness checks, degraded states, operational logs, and critical failure events may later feed an incident-management layer.

A future incident layer may include dashboards, alert routing, incident severity levels, runbooks, postmortem records, synthetic monitoring, and deploy or rollback coordination.

The initial goal is not to build all of that immediately. The initial goal is to avoid designing health checks and diagnostics as dead-end local tools that cannot later support production operations.

Where practical, early operational signals should be structured, stable, and service-identifying enough to connect to future incident tooling.

## Implementation sequence

1. Keep local packaged single-player independent from hosted services while preserving visible local failure behavior.
2. Define clear liveness, readiness, dependency, and admission checks for hosted environments.
3. Keep player-data failures visible and retryable without silently discarding reward-bearing results.
4. Make telemetry, logging, and diagnostics part of release-candidate readiness and hosted downgrade behavior.
5. Preserve rollback and recovery assumptions before hosted production rollout.
6. Leave incident-management scaffolding future-proofed without making it a current requirement.

## Related docs

* [Planning](../../!INDEX.md)
* [Development Roadmap](../../development-roadmap.md)
* [Build Release And Environment Matrix](build-release-and-environment-matrix.md)
* [Compatibility Versioning And Migrations](compatibility-versioning-and-migrations.md)
* [Logging And Diagnostics](observability-logging-and-diagnostics.md)
* [Services Planning](../../services/!INDEX.md)
* [Platform Security And Admin](../platform/security-and-admin/!INDEX.md)
* [Network Observability And Packet Budget](network-observability-and-packet-budget.md)

## Open decisions

* Which competitive modes require active match recovery?
* What exact data is stored for pending/retry match results?
* What exact idempotency key protects reward/result retries?
* What exact health checks are required for each service?
* What diagnostics are safe to include in copyable bug reports?
* What downgrade levels are needed after launch?
* What player-visible status path comes first?
* How much incident-platform scaffolding should be built early but left inactive?

## Notes

Keep this doc focused on cross-cutting readiness and failure behavior rather than service-specific implementation detail.
