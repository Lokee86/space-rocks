# Client Tests

Client tests use GUT and live under `client/tests/`.

- `unit/`: focused unit-style tests.
- `fixtures/`: small test data and scene fixtures.
- `helpers/`: reusable test-only helpers.

Keep client tests focused on generated packets, HUD behavior, `world_sync`, and pure client logic.

## Manual Smoke Boundary

These remain manual smoke tests for now:

- opening the game scene
- websocket connection
- spawning asteroids
- shooting/effects
- pause/debug flow
- full gameplay loop
