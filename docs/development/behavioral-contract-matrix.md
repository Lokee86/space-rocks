---
author: brian
created: "2026-07-30"
document_id: 2fc9121a-cd79-4a0b-a943-daf6deedbcff
document_type: development
policy_exempt: false
summary: This document maps critical Space Rocks behavioral invariants to focused tests and release gates.
---
# Behavioral Contract Matrix

Parent index: [Development Documentation](./!INDEX.md)

## Purpose

This document maps critical Space Rocks behavioral invariants to focused tests and release gates.

## Overview

The matrix focuses on contracts whose failure can cross runtime, authority, persistence, or compatibility boundaries. Package-level coverage alone is insufficient for these invariants.

## Contracts

| Contract | Primary verification owner |
| --- | --- |
| The Go game server is authoritative for gameplay outcomes; the Godot client owns presentation and input collection | Game-server unit/integration tests, client runtime tests, [source-of-truth flow](../domains/technical/source-of-truth-flow.md) |
| Packet schemas originate under `shared/packets/` and generated client/server packet definitions remain synchronized | `tools/data_sync` tests and CI, [packet schemas](../data/packet-schemas.md) |
| Tunable constants originate under `shared/constants/` and generated outputs are not hand-owned | Data-sync validation and generated-diff gates, [constants](../data/constants.md) |
| Packet-facing wire JSON crosses the server and client codec seams rather than ad hoc serialization | Go packet-codec tests, Godot packet-codec tests, [gameplay packets](../protocol/gameplay-packets.md) |
| Room and member identity remain distinct from player-facing `PlayerID` | Room/networking tests, [room membership and identity](../services/game-server/rooms/room-membership-and-identity.md) |
| Durable player counters live in session state rather than the live ship avatar | Player lifecycle/scoring tests, [player session state](../services/game-server/simulation/players/player-session-state.md) |
| Canonical targets use `target_kind` plus `target_id`; legacy player-only targeting stays quarantined | Targeting tests and packet validation, [canonical target state](../services/game-server/simulation/targeting/canonical-target-state.md) |
| Simulation phases execute in deterministic documented order | Runtime phase-order tests, [simulation loop](../services/game-server/simulation/runtime/simulation-loop-and-phase-order.md) |
| Toroidal distance, direction, movement, visibility, and wrap use the owning `space` and `motion` seams | World/motion tests, [toroidal space and motion](../services/game-server/simulation/world/toroidal-space-and-motion.md) |
| Devtools route through real gameplay owners and do not create parallel debug-only mechanics | Server/client devtools tests and build gates, [devtools authority](../devtools/design/devtools-authority-and-seams.md) |
| Observability events retain canonical names, fields, trace identity, and generated contract parity | Service logging tests, aggregator tests, generated-contract checks, [observability contract](../data/observability-contract.md) |
| Match outcomes and results are mode-defined, server-owned, and consistently projected to clients and persistence | Match/runtime/API tests, [match outcomes and results](../services/game-server/simulation/match/match-outcomes-and-results.md) |
| Friendly fire remains disabled under current team rules | Team/match-rule tests, [team configuration and membership](../services/game-server/rooms/team-configuration-and-membership.md) |
| Repository-wide scans and tooling exclude nested `.worktrees/` | Tool-specific traversal tests and CI configuration, [repo hygiene](../agent/repo-hygiene.md) |

## Release gates

The complete release gate is defined by repository CI and release workflows. At minimum, affected Go, Rails, Godot, data-sync, documentation, protocol-generation, and packaging checks must run for their changed surfaces.

## Maintenance

Update this matrix when an invariant changes, its focused test moves, a new cross-runtime contract is introduced, or a release gate ceases to protect the stated behavior.

## Related docs

- [Documentation coverage](documentation-coverage.md)
- [Agent testing](../agent/testing.md)
- [Architecture rules](../agent/architecture-rules.md)
- [Verification planning](../planning/domains/technical/verification-and-quality-gates.md)

## Notes

This matrix records critical contracts, not every unit test in the repository.
