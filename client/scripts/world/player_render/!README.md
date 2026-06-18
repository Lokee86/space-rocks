# Active Player Render API

This folder is the active API over `client/legacy/player_render`.

ViewAnchor is both the camera anchor and the render anchor.

`self_id` is player identity and `anchor_player_id` is render origin.

Normal code must not import `client/legacy/player_render` directly.

## Core Invariant

- ViewAnchor is the single render origin.
- Camera follows ViewAnchor.
- Background follows ViewAnchor.
- World rendering uses ViewAnchor-derived server/visual mapping.
- Player identity is separate from render origin.

## Identity vs Render Origin

- `self_id` is the local player identity.
- `anchor_player_id` is the player or target currently defining the render origin.
- During normal play, `self_id` and `anchor_player_id` are usually the same.
- During spectate/view-target behavior, they may differ.

## File Responsibilities

- `player_meaning_api.gd` wraps legacy player sync and exposes player facts.
- `view_anchor_sync.gd` wraps legacy local visual sync and exposes active anchor mapping.
- `player_render_api.gd` coordinates the active player/render seam for `world_sync`.

## Forbidden Dependencies

- active gameplay code must not import `client/legacy/player_render` directly
- `world_sync` should depend on `player_render_api.gd`, not legacy files
- camera/background consumers should use ViewAnchor, not Player/Camera2D

## Where New Work Goes

- new camera/view behavior goes in active API or a new active camera seam
- new render-anchor behavior goes in active API
- new player meaning queries go in `player_meaning_api.gd`
- legacy code remains black-box implementation

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

## Files

- `player_meaning_api.gd`: player facts from legacy sync
- `view_anchor_sync.gd`: ViewAnchor/render-anchor mapping
- `player_render_api.gd`: coordinator used by `world_sync`
