# Player Render Quarantine API

Files in `client/legacy/player_render` are quarantined implementation details.

New code must use active APIs under `client/scripts/world/player_render/`.

Only `client/scripts/world/player_render/*.gd` may preload files from `res://legacy/player_render/`.

## Do Not Import Directly

New gameplay, HUD, input, targeting, debug, camera, background, bullet, asteroid, and shell code must not import files from `res://legacy/player_render/`.

Only the active API files under `client/scripts/world/player_render/` may preload `res://legacy/player_render/`.

## Active API Files

- `player_meaning_api.gd`: exposes player meaning and legacy player facts
- `view_anchor_sync.gd`: wraps legacy anchor mapping for the active ViewAnchor/render anchor
- `player_render_api.gd`: coordinates player meaning, ViewAnchor state, and `world_sync`-facing methods

## What Belongs Outside Legacy

- camera behavior
- render-anchor policy
- spectate policy
- HUD/debug behavior
- targeting behavior
- gameplay behavior
- new tests for active behavior

## Compatibility-Only Edits

- stale path fixes
- minimal methods needed by the active API
- bug fixes that preserve legacy behavior behind the API

Tests may import legacy directly only when explicitly testing legacy compatibility or low-level legacy math.

## Dependency Guard

This command should only show active API files importing legacy player_render:

```text
grep -R 'res://legacy/player_render' client/scripts --include='*.gd'
```

Expected allowed results:

- `client/scripts/world/player_render/player_meaning_api.gd`
- `client/scripts/world/player_render/view_anchor_sync.gd`

This command should produce no active script results:

```text
grep -R 'res://scripts/world/player_sync.gd\|res://scripts/world/local_visual_sync.gd\|res://scripts/world/visual_sync_positions.gd' client/scripts client/tests --include='*.gd'
```

Tests may intentionally import legacy only when the test name or file makes that purpose explicit.

Active API responsibilities:

- player meaning and player node facts
- current view target selection
- ViewAnchor/render anchor server position
- ViewAnchor/render anchor visual position
- server-to-visual mapping
- visual-to-server mapping

New feature code must not import this folder directly.
