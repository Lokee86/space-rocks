# Project Notes
Parent index: [Documentation](./!INDEX.md)

This file is a small parking lot for project memory that does not yet belong in a canonical document.

It is not the source of truth for architecture, data-sync, testing, devtools, packet schemas, or current implementation paths. Keep canonical facts in the focused docs below.

## Canonical Docs

Use these instead of expanding this file:

- Architecture: `docs/systems-design/!INDEX.md`
- Developer workflow: `docs/developer.md`
- Testing and verification: `docs/agent/testing.md`
- Architecture rules: `docs/agent/architecture-rules.md`
- Godot-specific notes: `docs/agent/godot-notes.md`
- Server-specific notes: `docs/agent/server-notes.md`
- Current volatile context: `docs/agent/current-context.md`
- Devtools toggles: `docs/devtools/toggles.md`
- Data sync: `tools/data_sync/README.md`
- Toroidal wrap: `docs/systems-design/world/toroidal-wrap.md`
- Ship variants: `docs/systems-design/entities/variants.md`
- Client logging: `docs/services/client/client-logging.md`
- Server logging: `docs/services/game-server/observability/logging-and-diagnostics.md`

## Rules For This File

- Do not duplicate canonical architecture or workflow details here.
- Do not document generated file paths here unless this file is being used only as a temporary parking lot.
- Do not document packet schema locations here; use `docs/developer.md` and `tools/data_sync/README.md`.
- Do not document devtool key bindings here; use `docs/devtools/toggles.md`.
- Delete or move notes once they become stable enough for a focused doc.

## Parking Lot

- Space Rocks is moving quickly; stale notes should be removed aggressively.
- Prefer focused docs over growing this file.

## Asteroid Spawn Regression

- Intended behavior: timed asteroids spawn just outside the active player/camera view, generally drift toward the player/camera area, and despawn using that camera/view area plus despawn margin.
- Timed asteroid spawning is not currently intended to behave like global ambient spawning across the full playfield.
- Observed regression: asteroids appeared to spawn randomly across the whole playfield, while Godot camera controls showed asteroids still existed and were drifting elsewhere; `total_asteroids` telemetry helped confirm spawning had not stopped.
- Cause: after the server networking routing split, `client_config` packets were not forwarded to `Game.HandlePacket`, so the server did not receive or apply visible world dimensions, the camera view fell back to full world dimensions, and spawn/despawn bounds effectively behaved like whole-world bounds.
