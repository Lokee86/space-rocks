# Agent Current Context

This file is volatile project memory. Read it only when the task depends on current refactor status, dirty worktree notes, recent Godot/editor changes, or known gaps.

Keep this file shorter than permanent docs. Remove stale notes aggressively.

For stable architecture/runtime maps, use [docs/design/architecture.md](../design/architecture.md).
For stable current-session orientation, use [docs/agent/session-primer.md](session-primer.md).
For current devtool toggle behavior and hotkeys, use [docs/devtools/toggles.md](../devtools/toggles.md).

## Current Context

- The repo may be dirty.
- There may be unrelated Godot/editor asset changes in the worktree.
- Do not clean or revert unrelated user/editor changes casually.
- If gameplay or input looks broken, first confirm the Go server is running and the Godot client is connected.
- Godot was upgraded to 4.6 recently. Scene/import diffs may be noisy.
- The older `space-rocks-(4.3)/` project copy is ignored and should not be used as the active project.
- Generated recordings and build artifacts should not be committed. In particular, avoid committing `*.avi`, `tmp/`, `*/tmp/`, and `client/.godot/`.
- Pickup work for the current `1_up` / drop / devtools / lifespan scope is complete.
- Pickup expiry emits `pickup_expired`.
- Debug hitboxes use `debug_shape_catalog` plus `GameplayState` transforms; there is no live `DebugOverlayStatePacket`.
- `GameplayRuntimeContext` is runtime wiring only; do not treat it as a read-model passthrough bucket.
- Server hitbox overlay data comes through `WorldSync`/devtools seams, not `GameplayRuntimeContext`.
- Targeting now sits above `MouseActionFlow`; `GameplayTargetingContext` owns target selection orchestration and `WorldSync` only exposes `target_source()`.
- Room membership/owner state is behind the room membership owner seam.
- `websocket_write.go` only writes outbound/presentation state now; it does not advance game-over lifecycle.
- Continuous bullet stream runtime state is owned by `services/game-server/internal/devtools/streamruntime`.

## Current Direction Notes

- The user wants docs/plans to reflect a future NestJS API server separated from the Go game server.
- API/business/backend concerns should remain out of the Go real-time game server unless explicitly redirected.
- The user strongly prefers small implementation prompts and quick reviewable diffs.
- The user prefers scalable structure and useful seams over dumping more behavior into existing large files.
- World Telemetry Overlay is implemented behind the devtools seam and toggled by `DevToggle9` / `9`.
- Overlay scene: `client/scenes/devtools/world_telemetry_overlay.tscn`; telemetry scripts live under `client/scripts/devtools/telemetry/`.
- Devtools coordination now lives under `client/scripts/devtools/context/` with `GameplayDevtoolsContext` acting as the facade/composition seam.
- Network telemetry uses `telemetry_ping` / `telemetry_pong`; gameplay state packets include `server_sent_msec`.
- `packet_age_ms` depends on server clock offset estimated from telemetry ping/pong, not raw wall-clock subtraction.

## Known Gaps / TODOs

- Generated Godot constants/packet files may eventually move into a generated folder, but they currently live under `client/scripts/`.
- API server is planned but not scaffolded.
- Ship variants are planned but not implemented.
- Client packet codec callers now consume `PacketEncodeResult` and `PacketDecodeResult`; the codec at `client/scripts/networking/packets/packet_codec.gd` owns JSON parsing plus envelope validation only.
