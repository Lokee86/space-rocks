# Agent Micro Refactor Skill

Use this skill for the default Space Rocks implementation workflow: one small, bounded agent task that makes a reviewable edit and proves it with focused verification.

## When to use

Use this skill when the prompt asks for a surgical implementation, small refactor, file move, route-through, cleanup of a known seam, or one prompt from a numbered prompt sequence.

Do not use this skill for broad planning-only work, large architectural redesigns, or read-only reports unless the prompt explicitly asks for a read-only scan.

## Core rules

- Keep the diff small and reviewable.
- Preserve current behavior unless the prompt explicitly allows behavior change.
- Do not broaden scope beyond the named target.
- Do not do opportunistic cleanup, formatting churn, dependency upgrades, or unrelated renames.
- Do not edit generated files unless the prompt explicitly asks for a temporary/manual intervention.
- If the task balloons, stop and report the smallest next prompt instead of continuing.
- If tests fail, stop and report the failure. Do not pile additional changes onto a failing state unless the prompt asks for a focused fix.
- If the prompt says read-only, do not edit files, run formatters, or perform cleanup.

## Line-count guardrails

Line count is a smell, not an automatic failure.

For hand-written production files:

- Prefer files under about 200 lines when practical.
- Treat 300+ lines as a review trigger: confirm the file still has one clear responsibility.
- Treat 350+ lines as a warning: avoid adding new responsibility unless it clearly belongs there.
- Treat 500+ lines as a refactor trigger for actively changing files.
- Files over 500 lines should usually receive behavior-preserving edits, bug fixes, routing, or extraction work, not new responsibilities.

Exemptions: generated files, Godot `.tscn` scenes, `.tres` resources, vendored addons, fixtures, snapshots, and large declarative data files.

## Workflow

1. Inspect only the files needed for the requested edit.
2. Identify the smallest concrete change that satisfies the prompt.
3. Make the edit without unrelated cleanup.
4. Run the exact requested verification command.
5. If no command was requested, run the smallest relevant command:
   - Go server change: `cd services/game-server && env GOCACHE=/tmp/space-rocks-go-build go test -buildvcs=false ./...`
   - Godot client logic change: GUT if `godot` CLI is available.
   - Search/path move: a focused `rg` check for old and new paths.
6. Run `git status --short`.

## Terminal / verification policy

- Do not run terminal commands unless the prompt explicitly says to.
- Do not run `rg`, tests, formatters, generators, or git commands by default.
- The default agent job is editing only.
- Report changed files and unexpected scope changes.
- Verification commands are usually run separately by the user.

## Report format

Report only useful execution evidence:

```text
Changed files:
- ...

Verification:
- command: ...
- result: pass/fail
- relevant output: ...

Git status:
- ...

Notes:
- no unrelated edits / list any unrelated edits
```

When completing a numbered prompt, put this exact heading at the bottom of the report, replacing `X`:

```text
**COMPLETED PROMPT X**
```
