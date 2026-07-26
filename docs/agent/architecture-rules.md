---
author: brian
created: "2026-07-19"
document_id: 019f7d55-fb2c-7b69-9108-1c0766b9d878
document_type: general
policy_exempt: false
summary: This document defines mandatory architecture and seam-editing guardrails for agents changing Space Rocks code, structure, or ownership boundaries.
---
# Architecture and Seam Editing Rules
Parent index: [Agent](./!INDEX.md)

## Purpose

This document defines mandatory architecture and seam-editing guardrails for agents changing Space Rocks code, structure, or ownership boundaries.

## Overview

Use these rules to decide where a change belongs before editing. They govern ownership, scope, extraction, and responsibility boundaries; they do not document current implementation facts.

## Rules

- Identify the owning system before editing.
- If no clear owner exists, create the smallest concrete ownership seam or stop and report the missing seam.
- Defer mechanics, not ownership.
- Keep policy in the owning system; keep routing and composition thin.
- Prefer behavior-preserving extraction before behavior change.
- Do not add new responsibilities to broad lifecycle, networking, world-sync, shell/composition, game-loop, HUD/menu coordination, room/session bridge, or game-orchestration files merely because they are convenient.
- Avoid vague buckets or wrappers such as `helpers`, `utils`, `common`, `misc`, or generic managers unless a concrete responsibility truly requires one.
- Keep one ownership seam per scoped change.
- Preserve behavior unless behavior change is explicitly authorized.
- Do not include unrelated cleanup, formatting churn, opportunistic refactors, or package/folder moves.
- Generated/schema changes and scene/signal rewiring must be explicitly in scope.

### Co-hosted service dependency invariant

Process co-hosting permits composition-root construction, registration, and closure only. It does not permit service reach-through or imports from another service's domain or runtime packages. For diagnostic-aggregator, only the game-server composition-root adapter may import the public `services/diagnostic-aggregator/hosted` package. Game-server internal packages and all player-data packages must not import any diagnostic-aggregator package.

Adding a diagnostic dependency to gameplay, networking, rooms, match reporting, player-data runtime/store code, or their constructors is an architecture violation. Stop and report rather than implementing that dependency. Use the transport/API boundary described in [Game-server Diagnostic-Aggregator Hosting](../services/game-server/integrations/diagnostic-aggregator-hosting.md). Pitlord rule `game-runtime-no-diagnostic-aggregator-dependency` enforces this import boundary for game-server internals and player-data source.

### High-risk seam verification

New seams at trust-sensitive boundaries require explicit verification expectations before the seam is considered complete. This includes networking, session and lifecycle ownership, world synchronization, authentication and identity, protocol or schema handling, authoritative simulation, persistence, and similar boundaries where a bypass can invalidate correctness or safety.

For each new high-risk seam, document:

- the owning system;
- the bypass or reach-through behavior that must remain forbidden;
- focused behavioral or contract tests for the seam; and
- whether the invariant is statically detectable and should receive a configuration-driven architecture-guard rule.

Architecture guards should be narrow and evidence-based, not speculative or brittle. Existing guards are intentionally incomplete; expand them when concrete risk justifies it rather than requiring a guard for every ordinary feature. When future cleanup repeatedly finds the same ownership violation, reach-through, forbidden dependency, or duplicated authority, the cleanup should normally add or strengthen a focused guard or test so that class of drift does not return.

Line-count guardrails for handwritten production files:

- Prefer files under roughly 200 lines when practical.
- Around 300 lines, review whether the file still has one clear responsibility.
- Around 350 lines, avoid adding new responsibility unless it clearly belongs there.
- Around 500 lines, treat actively changing files as split candidates.
- For files above 500 lines, prefer extraction or routing over adding more responsibility.
- Generated files, Godot scenes/resources, vendored files, fixtures, snapshots, and large declarative data files are exempt.

Stop and report before proceeding when ownership is unclear, the work crosses multiple seams, or the required scope expands materially beyond the request.

## Related docs

- [Client editing guardrails](./client-editing.md)
- [Server editing guardrails](./server-editing.md)
- [Documentation editing guidance](./documentation-editing.md)
- [Systems design index](../systems-design/!INDEX.md)
- [Seam-first skill](../../skills/seam-first/SKILL.md)

## Notes

This is the canonical agent-facing architecture and seam-editing guardrail document. Keep current implementation facts in their owning canonical documentation instead of duplicating them here.
