# Player Render Quarantine API

Files in `client/legacy/player_render` are quarantined implementation details.

New code must use active APIs under `client/scripts/world/player_render/`.

Only `client/scripts/world/player_render/*.gd` may preload files from `res://legacy/player_render/`.

Active API responsibilities:

- player meaning and player node facts
- current view target selection
- ViewAnchor/render anchor server position
- ViewAnchor/render anchor visual position
- server-to-visual mapping
- visual-to-server mapping

New feature code must not import this folder directly.
