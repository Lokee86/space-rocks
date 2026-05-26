# Godot UI Scene Edit Skill

Use this skill when editing Godot UI scenes, HUD/menu layout, labels, visibility, anchors, containers, or scene-connected UI behavior.

## When to use

Use this skill for work involving:

- `.tscn` UI scenes
- `.tres` UI resources
- HUD/menu layout
- game over UI
- lobby UI
- labels, buttons, containers, anchors, offsets, theme overrides
- scene paths and signal connections
- `client/scripts/ui/**`
- UI-facing controller scripts

## Core rules

- Inspect the scene before editing.
- Prefer script/controller changes over noisy scene rewrites when possible.
- Avoid normalizing or reformatting entire `.tscn` files.
- Do not rewrite unrelated offsets, anchors, unique IDs, imports, or metadata.
- Be careful with Godot-generated `uid`, `unique_id`, offsets, imports, and scene metadata.
- Do not move nodes or change scene ownership unless required.
- Preserve behavior unless the prompt explicitly says layout/behavior may change.
- Manual visual smoke testing may be required; automated tests cannot fully prove layout.

## UI ownership rules

- Presentation, UI, audio/effects, local input collection, and interpolation belong in the Godot client.
- Authoritative gameplay outcomes belong on the Go game server.
- UI should present state from real gameplay/client seams, not invent parallel gameplay state.
- Devtools UI must route through real gameplay seams and generated packets when applicable.
- Client spectate/view-cycle UI should use authoritative lifecycle status plus visual availability.

## Scene edit checklist

Before editing:

1. Locate the scene and attached scripts.
2. Identify the node path that actually owns the target UI element.
3. Check whether a controller script already owns the behavior.
4. Prefer changing the controller if scene data does not need to change.

After editing, check relevant code/diff for:

- node names
- `%UniqueName` references
- signal connections
- attached scripts
- `preload(...)` paths
- visibility defaults
- container/layout side effects
- `uid://` and imported resource changes
- whether Godot rewrote unrelated scene metadata

## Line-count guardrails

Line-count limits do not apply directly to `.tscn` or `.tres` files.

For attached hand-written UI scripts:

- Prefer under about 200 lines.
- 300+ lines requires responsibility scrutiny.
- 500+ lines is a gravity-well warning.
- Do not add new responsibilities to large UI controllers; prefer extracting a small presentation controller or helper with a concrete owner.

## Testing

For UI behavior changes, add or update focused GUT tests where practical when the prompt asks for test edits.

For pure layout-only changes, report that manual smoke testing is still needed.

## Human-run verification

Suggested human-run GUT command:

```bash
godot --headless --path client -s res://addons/gut/gut_cmdln.gd -gdir=res://tests/unit -ginclude_subdirs -gexit
```

Do not run this command by default as the agent.

## Stop conditions

Stop and report instead of continuing if:

- The scene diff becomes broad/noisy.
- A small UI change requires moving many nodes.
- The change requires altering gameplay authority or server-owned state.
- The target node path is unclear.
- Godot generated unrelated import/metadata churn that the prompt did not ask for.
