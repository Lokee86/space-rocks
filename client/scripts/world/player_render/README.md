# Active Player Render API

This folder is the active API over `client/legacy/player_render`.

ViewAnchor is both the camera anchor and the render anchor.

`self_id` is player identity and `anchor_player_id` is render origin.

Normal code must not import `client/legacy/player_render` directly.

## Files

- `player_meaning_api.gd`: player facts from legacy sync
- `view_anchor_sync.gd`: ViewAnchor/render-anchor mapping
- `player_render_api.gd`: coordinator used by `world_sync`
