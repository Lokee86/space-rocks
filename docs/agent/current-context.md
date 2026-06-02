# Agent Current Context

This file is volatile project memory. Read it only when the task depends on current refactor status, dirty worktree notes, recent Godot/editor changes, or known gaps.

Keep this file shorter than permanent docs. Remove stale notes aggressively.

For stable architecture/runtime maps, use [docs/design/architecture.md](../design/architecture.md).
For current devtool toggle behavior and hotkeys, use [docs/devtools/toggles.md](../devtools/toggles.md).

## Current Context

- The repo may be dirty.
- There may be unrelated Godot/editor asset changes in the worktree.
- Do not clean or revert unrelated user/editor changes casually.
- If gameplay or input looks broken, first confirm the Go server is running and the Godot client is connected.
- Godot was upgraded to 4.6 recently. Scene/import diffs may be noisy.
- The older `space-rocks-(4.3)/` project copy is ignored and should not be used as the active project.
- Generated recordings and build artifacts should not be committed. In particular, avoid committing `*.avi`, `tmp/`, `*/tmp/`, and `client/.godot/`.

## Current Direction Notes

- The user wants docs/plans to reflect a future NestJS API server separated from the Go game server.
- API/business/backend concerns should remain out of the Go real-time game server unless explicitly redirected.
- The user strongly prefers small implementation prompts and quick reviewable diffs.
- The user prefers scalable structure and useful seams over dumping more behavior into existing large files.
- World Telemetry Overlay is implemented behind the devtools seam and toggled by `DevToggle9` / `9`.
- Overlay scene: `client/scenes/devtools/world_telemetry_overlay.tscn`; telemetry scripts live under `client/scripts/devtools/telemetry/`.
- Network telemetry uses `telemetry_ping` / `telemetry_pong`; gameplay state packets include `server_sent_msec`.
- `packet_age_ms` depends on server clock offset estimated from telemetry ping/pong, not raw wall-clock subtraction.
- `total_asteroids` telemetry only stays visible if both the base `StatePacket` and `WrapStatePacket()` preserve the field.
- Remote player dev labels are implemented behind the client devtools seam. `DevToggle8` / `8` shows basic remote-player labels, and `Shift+DevToggle8` / `Shift+8` shows network telemetry labels. Labels attach to remote player nodes only, exclude the local player, and keep lifecycle, formatting, and mode state under `client/scripts/devtools/` and `client/scenes/devtools/`.
- Server Hitbox Overlay is implemented behind the client devtools seam, toggled by the devtools window checkbox, and draws player/asteroid/bullet outlines from `WorldSync`/runtime draw-entry data using Godot collision resources only as visual templates.
- Eligible devtools player-target lists now default to All Players, using `target_scope=all_players` instead of a fake player ID; Invincible, Infinite Lives, and Freeze Player use set-style all-player activation, and Respawn Player all-player requests still rely on existing per-player respawn guards.

## Known Gaps / TODOs

- Generated Godot constants/packet files may eventually move into a generated folder, but they currently live under `client/scripts/`.
- API server is planned but not scaffolded.
- Ship variants are planned but not implemented.
