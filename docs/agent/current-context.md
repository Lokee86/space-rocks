# Agent Current Context

This file is volatile project memory. Read it only when the task depends on current refactor status, dirty worktree notes, recent Godot/editor changes, or known gaps.

Keep this file shorter than the permanent architecture docs. Remove stale notes aggressively.

## Current Context

- The repo may be dirty.
- There may be unrelated Godot/editor asset changes in the worktree.
- Do not clean or revert unrelated user/editor changes casually.
- If gameplay or input looks broken, first confirm the Go server is running and the Godot client is connected.
- Godot was upgraded to 4.6 recently. Scene/import diffs may be noisy.
- The older `space-rocks-(4.3)/` project copy is ignored and should not be used as the active project.
- Generated recordings and build artifacts should not be committed. In particular, avoid committing `*.avi`, `tmp/`, `*/tmp/`, and `client/.godot/`.

## Recent / Important Project Direction

- The user wants docs/plans to reflect a future NestJS API server separated from the Go game server.
- API/business/backend concerns should remain out of the Go real-time game server unless explicitly redirected.
- The user strongly prefers small implementation prompts and quick reviewable diffs.
- The user prefers scalable structure and useful seams over dumping more behavior into existing large files.
- The user prefers files under roughly 200 lines when practical and treats roughly 500 lines as a refactor trigger for actively changing production files.
- Agent prompts should be short work orders, not mini-policy documents.
- Agents should not run terminal commands by default. Verification commands are usually human-run checkpoints.
- Agent reports should focus on changed files, unexpected files touched, and concise notes. Include command/test/git output only when the prompt explicitly allowed terminal commands.
- Do not produce no-work prompts. Verification belongs in commands/checkpoints, not in a separate agent prompt.

## Current Client Runtime Areas

Common starting points after the client runtime refactor:

- `client/scripts/session/`: session-level coordinators, including gameplay, room, config, and client session context.
- `client/scripts/shell/gameplay_shell_flow.gd`: narrow gameplay shell coordinator. It should stay mostly as orchestration and delegation.
- `client/scripts/gameplay/runtime/`: gameplay runtime/state-application context.
- `client/scripts/gameplay/state/`: gameplay packet/state readers and normalized state helpers.
- `client/scripts/gameplay/input/`: local gameplay input polling/routing, including movement, pause/menu, respawn, spectate input routes, and devtools input ownership.
- `client/scripts/shell/gameplay_hud_flow.gd`: gameplay HUD flow.
- `client/scripts/shell/gameplay_menu_flow.gd`: gameplay menu flow and semantic menu lifecycle signal routing.
- `client/scripts/shell/gameplay_respawn_flow.gd`: respawn request and confirmation state.
- `client/scripts/shell/gameplay_runtime_tick_flow.gd`: runtime HUD ticking.
- `client/scripts/gameplay/spectate/`: spectate state, menu requests, and camera target cycling.
- `client/scripts/gameplay/events/`: server event lane and death/game-over consequences.
- `client/scripts/gameplay/effects/`: gameplay effects helper used by event/effects flows.
- `client/scripts/lobby/`: lobby shell/presenter/network action flows.
- `client/scripts/boot/`: boot flow and pending boot request.
- `client/scripts/config/`: client config flows.
- `client/scripts/networking/`: websocket client and packet transport.
- `client/scripts/networking/packets/`: generated packets and packet codec.
- `client/scripts/world/`: world sync and entity sync owners.
- `client/scripts/entities/player.gd`: local player node and packet-facing input state.
- `client/scripts/ui/`: UI nodes/controllers.

## Implemented Developer Toggles

Current hardcoded Godot hotkeys:

- `F1`: toggle debug invincibility for the player
- `F2`: toggle debug infinite lives for the player
- `F3`: toggle room-wide debug world freeze
- `F4`: toggle the player's paused state

These are server-authoritative toggles sent through generated packets where applicable. See `docs/devtools/toggles.md`.

## Pause / Menu Context

Pause plumbing exists:

- packets: `pause_player`, `resume_player`
- server player fields include paused/invulnerability state
- paused players should ignore input, not shoot/score, not take asteroid damage, and be hidden by client world sync
- resume starts a short invulnerability window
- pause/menu UI exists but is still evolving

If pause behavior seems wrong, inspect current Godot scenes/scripts before changing code. The HUD/menu scenes have been changed multiple times.

## Future Plans Already Documented

Toroidal/wrapped world:

- Use `services/game-server/internal/game/space` as the abstraction point.
- Future/current server coordinates should be bounded/wrapped.
- Future/current client rendering should use unwrapped visual positions relative to the local player so border crossing is invisible.
- See `docs/design/toroidal-wrap.md`.

Ship variants:

- Future ships may use different client scenes and server collision maps.
- See `docs/design/ship-variants.md`.

API server:

- Planned as Node.js/TypeScript/NestJS in `services/api-server/`.
- It should own business/backend concerns, not real-time simulation.
- See `docs/api/nestjs-api-server.md`.

## Known Gaps / TODOs

- Window/gameplay balance should move away from raw OS max window pixels toward a logical gameplay viewport cap.
- Generated Godot constants/packet files may eventually move into a generated folder, but they currently live under `client/scripts/`.
- API server is planned but not scaffolded.
- Ship variants are planned but not implemented.
