# Verification And Quality Gates

Parent index: [Technical Planning](./!INDEX.md)

## Purpose

This doc plans the verification and quality-gate policy for Space Rocks.

It defines which gates exist, what each gate protects, what evidence is acceptable, and when checks must pass before work becomes release-shaped or production-capable.

Verification is broader than automated tests. It includes tests, contract checks, documentation checks, smoke checks, packaged-build checks, runtime scenarios, migration checks, health checks, and manual confirmation where automation would be brittle or premature.

## Overview

This doc keeps verification and release gates aligned so code, docs, contracts, runtime health, and migrations are proven at the right seam without brittle all-in-one tests.

## Current status

Active planning.

## Ownership Boundary

This doc owns planning for verification strategy, quality-gate boundaries, release-shaped gates, smoke checks, contract checks, documentation checks, and anti-brittle-test policy.

It should describe what must be proven, when it must be proven, and how proof should be gathered at a planning level.

Exact test implementation belongs in the relevant service, protocol, data, client, devtools, or tooling docs.

## Does Not Belong

* Exact test code.
* Exact CI provider setup.
* Exact performance thresholds.
* Exact protocol design.
* Exact migration implementation.
* Exact hosting-provider infrastructure.
* Giant end-to-end test plans that try to prove every system at once.
* Current implementation authority.

## Verification Model

Verification should answer three questions:

* What risk is being checked?
* What evidence proves the risk is controlled?
* When does that evidence become required?

Quality gates are the pass/fail boundaries that use that evidence.

A check can begin as manual verification and become automated later. Automation should be added where it protects important seams without creating brittle broad tests.

## Evidence Types

| Evidence Type        | Role                                                                        |
| -------------------- | --------------------------------------------------------------------------- |
| Automated Tests      | Prove isolated behavior or stable seams.                                    |
| Contract Checks      | Prove generated files and source-of-truth contracts are current.            |
| Documentation Checks | Prove indexes, links, and documentation structure remain valid.             |
| Manual Smoke         | Prove important user flows still work when automation would be too brittle. |
| Packaged Build Smoke | Prove exported or packaged builds work outside the editor/dev runner.       |
| Runtime Scenario     | Prove runtime pressure is measurable and not blocker-level.                 |
| Health Check         | Prove services are live, ready, and able to admit users.                    |
| Migration Check      | Prove old data can migrate, back up, fail safely, or be blocked visibly.    |

## Gate Types

| Gate                             | Purpose                                                                 |
| -------------------------------- | ----------------------------------------------------------------------- |
| Local Development Sanity         | Basic confidence before trusting local changes.                         |
| Documentation And Contract Gate  | Prevent stale docs, stale generated files, and source-of-truth drift.   |
| Local Packaged Beta Gate         | Prove the first packaged single-player testing build works outside dev. |
| Dev-Hosted Multiplayer Gate      | Prove multiplayer works against hosted or semi-hosted services.         |
| Hosted Staging Gate              | Prove a production-like environment is ready for validation.            |
| Production Candidate Gate        | Prove a build/environment is safe enough for public production.         |
| Runtime-Heavy Feature Gate       | Prove heavy gameplay systems are measured before expansion.             |
| Migration And Compatibility Gate | Prove version, profile, and data changes fail or migrate safely.        |
| Operational Readiness Gate       | Prove services fail visibly and health/readiness checks behave.         |

## Local Development Sanity Gate

Local development sanity should remain lightweight.

It may use:

* Go tests,
* GUT tests,
* Rails/API tests where relevant,
* targeted manual smoke,
* data-sync checks when contract surfaces change,
* doc-ledger checks when documentation changes.

Local development does not need to enforce every release-shaped gate. It should stay fast enough to support iteration.

## Documentation And Contract Gate

Documentation and contract gates protect source-of-truth surfaces.

They should verify:

* documentation indexes and required sections are valid,
* links are not broken where tooling can detect them,
* OpenAPI contracts match expected generated output,
* data-sync outputs are current,
* generated Go and Godot contract files are not stale,
* shared data definitions match generated consumers.

Generated contract drift should block release-shaped builds. Local development can remain looser unless the changed surface directly depends on regenerated output.

## Local Packaged Beta Gate

Local packaged beta is the first release-shaped testing target.

This gate should prove that the game works outside the normal editor/dev-runner path.

Acceptable evidence includes packaged-build smoke plus targeted checks for:

* client export success,
* bundled local server startup,
* client connection to the bundled local server,
* local-only server behavior,
* local profile create/load/save,
* single-player match start,
* single-player match end,
* results screen display,
* local stats persistence,
* server process cleanup on quit,
* devtools policy matching beta expectations,
* no hosted auth or hosted multiplayer requirement.

These builds are beta/testing builds, not production builds. Devtools may remain open or enabled for diagnostics and bug reporting.

## Dev-Hosted Multiplayer Gate

Dev-hosted multiplayer is a multiplayer testing target, not production.

This gate should prove that hosted or semi-hosted multiplayer flows work under controlled test conditions.

Acceptable evidence includes:

* client can reach configured services,
* auth/admission works for the test environment,
* room creation works,
* join/ready/start/end flow works,
* disconnect and failed-join paths are visible,
* player-data unavailable behavior warns clearly,
* match-result write or pending/retry behavior is visible,
* server runtime and packet pressure are observable,
* no SQLite is used in the multiplayer build.

Forced upgrades are acceptable in this environment.

## Hosted Staging Gate

Hosted staging should resemble production closely enough to validate release candidates.

This gate should prove:

* API server readiness,
* game-server readiness,
* player-data readiness,
* web service readiness where player-facing,
* auth integration behavior,
* compatibility admission,
* generated contract currency,
* basic operational diagnostics,
* telemetry/logging availability,
* controlled runtime/load scenarios,
* player-data write and retry behavior,
* visible degraded-state behavior.

Hosted staging should follow production expectations where practical, even if production scale is not yet required.

## Production Candidate Gate

Production candidate gates should be stricter than staging gates.

A production candidate should not pass if:

* incompatible clients can reach normal login or gameplay paths,
* production client devtools capability is present,
* generated contracts are stale,
* required migrations are missing,
* local or hosted data migration risks silent data loss,
* player-data failure silently discards reward-bearing results,
* required health/readiness checks fail,
* telemetry/logging required for readiness is unavailable,
* runtime scenarios show blocker-level pressure,
* required operational diagnostics are missing.

Hosted production should initially support the current version only.

## Runtime-Heavy Feature Gate

Runtime-heavy features need measurement before expansion.

This applies to:

* enemies,
* bosses,
* bullet hell patterns,
* drones,
* mines,
* radial effects,
* larger multiplayer rooms,
* spectators,
* competitive modes,
* repeated match start/end churn,
* progression-heavy result flows.

Acceptable evidence may include manual measurement, automated runtime scenarios, synthetic load, or soak runs.

The runtime performance plan owns the measurement details. This doc owns the rule that runtime-heavy work must pass an appropriate gate before becoming release-shaped or production-capable.

## Migration And Compatibility Gate

Migration and compatibility gates protect users from broken upgrades and data loss.

They should verify:

* incompatible production clients are blocked early,
* local packaged builds use matching client/server pieces,
* local profile migration creates or offers a backup,
* failed local profile migration preserves old data,
* hosted migrations include backup and rollback planning where practical,
* service, protocol, and contract versions are treated separately,
* unknown or incompatible realtime/API behavior fails safely.

Exact migration implementation belongs in data and service docs.

## Operational Readiness Gate

Operational readiness gates protect player-facing environments from silent failure.

They should verify:

* liveness checks exist where useful,
* readiness checks block unsafe admission,
* dependency health is visible,
* admission health can block new users or matches,
* maintenance/admission-blocked states work,
* major failures fail visibly,
* copy diagnostics or bug-report paths exist where automatic reporting is unavailable,
* player-data downtime warns users about stats, items, currency, rankings, or rewards,
* telemetry/logging failure blocks pre-launch readiness or downgrades launched environments.

The operational readiness plan owns failure-mode policy. This doc owns the verification gate around that policy.

## Manual Versus Automated Checks

Not every check should be automated immediately.

| Check Status                    | Use                                                                         |
| ------------------------------- | --------------------------------------------------------------------------- |
| Automated Now                   | Stable seams that are cheap and valuable to verify repeatedly.              |
| Manual For Now                  | Flows that matter but are still changing too often for reliable automation. |
| Must Automate Before Production | Release-critical seams that cannot depend on memory or manual discipline.   |
| Do Not Automate Yet             | Broad or unstable flows where automation would be brittle and noisy.        |

Manual checks should not be treated as permanent when they protect production-critical behavior. They are acceptable while the system is still changing.

## Anti-Brittle-Test Policy

Space Rocks should prefer small, targeted verification around important seams over broad fragile tests that try to prove everything at once.

Good verification targets include:

* packet codec behavior,
* local profile migration,
* match-result idempotency,
* generated contract drift,
* bundled server startup,
* auth/admission failure,
* player-data unavailable warning,
* health/readiness state,
* focused runtime pressure scenarios.

Risky verification targets include:

* giant end-to-end tests that depend on every service for one assertion,
* brittle UI pixel/state tests,
* long gameplay scripts that fail for unrelated presentation changes,
* tests that require hosted infrastructure to verify a local seam,
* broad checks that are hard to diagnose when they fail.

Broad smoke tests can exist, but they should not replace seam-focused verification.

## Launch-Shaped Verification

Verification should grow toward launch-shaped confidence, not only first-slice confidence.

Local packaged beta, dev-hosted multiplayer, hosted staging, and production candidate builds each need their own gate because they protect different risks.

A feature or build can be correct in local development and still fail release-shaped verification if it breaks packaging, migration, runtime health, operational readiness, generated contracts, or production admission rules.

## Implementation sequence

1. Keep local development sanity checks fast and seam-focused.
2. Enforce documentation and contract checks wherever source-of-truth drift can block release-shaped work.
3. Use packaged-build smoke for the first local packaged beta gate.
4. Require hosted staging and production candidate checks to cover runtime, migration, and operational readiness risks.
5. Keep runtime-heavy feature, migration, and operational gates separate so each risk is measured at the right level.

## Related docs

* [Planning](../../!INDEX.md)
* [Development Roadmap](../../development-roadmap.md)
* [Build Release And Environment Matrix](build-release-and-environment-matrix.md)
* [Compatibility Versioning And Migrations](compatibility-versioning-and-migrations.md)
* [Operational Readiness And Failure Modes](operational-readiness-and-failure-modes.md)
* [Runtime Performance And Scale Budget](runtime-performance-and-scale-budget.md)
* [Network Observability And Packet Budget](network-observability-and-packet-budget.md)
* [Logging And Diagnostics](observability-logging-and-diagnostics.md)
* [Data Sync And SSoT Pipeline](../../../data/data-sync-and-ssot-pipeline.md)

## Open decisions

* Which checks should stay manual through local packaged beta?
* Which checks must become automated before hosted production?
* Which contract checks protect the highest-risk seams first?
* Which runtime scenarios are required for each release-shaped build?
* Which operational failure paths need automated verification versus manual smoke?
* Which broad smoke checks are useful enough despite brittleness risk?
* What is the minimum production candidate gate before public hosted multiplayer?

## Notes

Keep verification policy scoped to proof and gating decisions, not implementation detail or test-suite design.
