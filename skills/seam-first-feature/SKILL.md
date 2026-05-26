# Seam-First Feature Skill

Use this skill when adding a new feature, mechanic, UI behavior, gameplay rule, packet behavior, devtool behavior, or visible game behavior.

Do not use this skill for purely mechanical renames, path moves, formatting, generated-file updates, or behavior-preserving extractions where another skill is more specific.

## Goal

Add features through the smallest correct owning seam instead of growing gravity-well files.

Behavior may be small. Ownership should still be explicit.

## Core rules

- Identify the owning system before editing.
- If no clear owner exists, stop and report the missing seam.
- Do not place new behavior in a shell, lifecycle, sync, networking, UI controller, or game-loop file only because it is convenient.
- Defer mechanics, not ownership.
- Prefer a tiny owner/seam now over extracting mature behavior later.
- Preserve existing behavior except for the requested feature slice.
- Do not create vague buckets such as `utils`, `helpers`, `common`, `misc`, or generic `manager`.
- Good seams have concrete responsibilities: damage, scoring, spawning, lives, respawn, spectate, room flow, packet codec, domain events, presentation state, sync ownership, input routing, or menu policy.
- If the feature needs policy, put policy in the owning system, not in transport/rendering glue.
- If the feature only needs routing, keep the routing thin.

## File-size rules

For hand-written production files:

- Prefer files under 200 lines.
- Do not let files grow past 300 lines as a normal accepted state.
- Treat 300+ lines as a hard architecture warning.
- If a feature task would add code to an already-large file, create or use a smaller owning file/seam instead.
- Do not add new behavior to large coordination/gravity-well files for convenience.

Generated files, Godot `.tscn` scenes, `.tres` resources, vendored addons, fixtures, snapshots, and large declarative data files are exempt from strict line-count rules.

## Gravity-well warning files

Be extra cautious before adding behavior to broad files such as:

- Godot gameplay shell files
- Godot lifecycle/menu/session/sync files that already coordinate many systems
- Go game orchestration files
- Go networking/websocket files
- room/session bridge files
- HUD controllers that are already coordinating many UI states

If a file is already large or already coordinates multiple systems, do not add a new responsibility there unless the prompt explicitly says to.

## Feature workflow

1. Name the feature behavior in one sentence.
2. Identify the owning system.
3. Check whether an existing seam already owns it.
4. If an existing seam owns it, add the smallest behavior there.
5. If no seam owns it, create the smallest concrete owner first.
6. Route existing shell/lifecycle/network/UI code through that owner.
7. Keep compatibility wrappers only when needed by existing tests or call sites.
8. Do not combine unrelated seams in the same edit.
9. Leave broad verification for a separate human-run checkpoint unless the prompt explicitly allows terminal commands.

## Stop conditions

Stop and report instead of editing if:

- The correct owner is unclear.
- The prompt requires adding behavior to a gravity-well file without a clear owner.
- The feature appears to involve multiple seams at once.
- The task requires scene/node/signal rewiring not mentioned in the prompt.
- The task requires generated packet/constants changes not mentioned in the prompt.
- The edit would require touching more files than the prompt allows.
- The feature would push a hand-written production file over 300 lines.

## Terminal policy

- Do not run terminal commands unless the prompt explicitly says to.
- Do not run `rg`, tests, formatters, generators, or git commands by default.
- Human-run verification happens separately after the edit.

## Report format

Report only:

```text
Changed files:
- ...

Owning seam used or created:
- ...

Unexpected files touched:
- none / ...

Notes:
- ...

**COMPLETED PROMPT X**
```

When completing a numbered prompt, put the exact completion heading at the bottom of the report, replacing `X` with the prompt number.
