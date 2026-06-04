# Client Tests

Client tests use GUT and live under `client/tests/`.

- `unit/`: focused unit-style tests.
- `fixtures/`: small test data and scene fixtures.
- `helpers/`: reusable test-only helpers.

Keep client tests focused on generated packets, HUD behavior, `world_sync` coordination, extracted sync owners, and pure client logic.

## Player/Render Tests

Active behavior tests should target `client/scripts/world/player_render` APIs.

Do not casually preload moved legacy files.

Direct legacy imports are acceptable only for:

- explicit legacy compatibility tests
- low-level legacy math tests that intentionally document the black-box behavior

Stale path warning:

- `res://scripts/world/local_visual_sync.gd` is stale
- `res://scripts/world/visual_sync_positions.gd` is stale
- `res://scripts/world/player_sync.gd` is stale

Expected active test areas:

- `ViewAnchorSync` anchor mapping
- `PlayerRenderApi` `self_id` vs `anchor_player_id` behavior
- `world_sync` coordination through `PlayerRenderApi`

## Manual Smoke Boundary

These remain manual smoke tests for now:

- opening the game scene
- websocket connection
- spawning asteroids
- shooting/effects
- pause/debug flow
- full gameplay loop
