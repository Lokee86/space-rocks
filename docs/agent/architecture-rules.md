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
