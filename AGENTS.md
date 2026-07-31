# AGENTS.md

Guidance for Codex and other coding agents working in this repository.

This file is the short, always-read operating manual. Keep it practical and stable. For deeper or more temporary context, read the linked docs only when the task needs them.

## Project Snapshot

Space Rocks is an Asteroids-inspired game in active development.

- Godot client: `client/`
- Go real-time game server: `services/game-server/`
- Planned API/business server placeholder: `services/api-server/`
- Shared generated data sources: `shared/`
- Project docs: `docs/`
- Packet/constants sync tool: `tools/data_sync/`
- Task-specific agent workflows: `skills/`

The current gameplay direction is server-authoritative. The Godot client handles rendering, UI, audio/effects, local input collection, and interpolation. The Go game server owns simulation outcomes: movement, bullets, collisions, scoring, lives, death, respawn, pause safety, rooms, and websocket state.

## Read First

For normal workflow:

- `docs/developer.md`

For documentation work:

- `docs/documentation-policy.md`
- `docs/documentation-procedure.md`
- `docs/development/documentation-coverage.md`
- `docs/development/behavioral-contract-matrix.md`

For project memory and recent volatile context:

- `docs/agent/current-context.md`
- `docs/agent/session-primer.md`
- `docs/notes.md`

For architecture and seam rules:

- `docs/agent/architecture-rules.md`
- `docs/systems-design/!INDEX.md`

For Godot/editor/client-specific notes:

- `docs/agent/godot-editing.md`
- `docs/devtools/server/toggles.md`
- `docs/systems-design/world/toroidal-wrap.md`
- `docs/systems-design/entities/variants.md`

For server/API/logging details:

- `docs/agent/server-editing.md`
- `docs/services/game-server/observability/logging-and-diagnostics.md`
- `docs/services/api-server/!INDEX.md`

## Current Layout Notes

The Go game server lives at:

```text
services/game-server/
```

Its Go module path is:

```text
github.com/Lokee86/space-rocks/services/game-server
```

The module path matches the `services/game-server/` directory. Import paths inside the Go server use `github.com/Lokee86/space-rocks/services/game-server/...`.

`services/api-server/` exists as a Ruby/Rails API-only scaffold. Do not put account, persistence, matchmaking metadata, leaderboard, or other business/backend concerns into the Go game server unless the user explicitly changes that direction.

## Generated Files

Do not hand-edit generated files unless the user explicitly asks for a temporary/manual intervention.

Constants source of truth:

```text
shared/constants/server_constants.toml
shared/constants/server_entities.toml
shared/constants/client/presentation.toml
shared/constants/client/shell.toml
shared/constants/client/lobby.toml
```

Generated constants:

```text
client/scripts/generated/constants/constants.gd
services/game-server/internal/constants/constants.go
```

Packet source of truth:

```text
shared/packets/outputs.toml
shared/packets/gameplay.toml
shared/packets/debug.toml
shared/packets/lobby.toml
```

Generated packets:

```text
client/scripts/generated/networking/packets/packets.gd
services/game-server/internal/game/packets.go
services/game-server/internal/game/runtime/packets_generated.go
services/game-server/internal/devtools/packets_generated.go
```

Collision shape source:

```text
shared/collisions/collision_shapes.json
```

Data sync tool:

```text
tools/data_sync/
```

Use `shared/constants/server_constants.toml`, `shared/constants/server_entities.toml`, `shared/constants/client/presentation.toml`, `shared/constants/client/shell.toml`, and `shared/constants/client/lobby.toml` plus `tools/data_sync/` for active constants. Use `shared/packets/outputs.toml`, `shared/packets/gameplay.toml`, `shared/packets/debug.toml`, and `shared/packets/lobby.toml` plus `tools/data_sync/` for active packets. API-specific output is future/deferred until the Ruby/Rails API-only service grows beyond the scaffold.

Tunable/game-data constants belong in the split constants SoT files under `shared/constants/` and generated scripts under `client/scripts/generated/constants/`. Client constants use nested subcategory sections under `constants.client.presentation.*`, `constants.client.shell.*`, and `constants.client.lobby.*`. Do not create local constants files elsewhere; change generated constants through the data source/regeneration path, not manual edits.

Packet schema changes should be made in the relevant split packet TOML under `shared/packets/` and pushed with `tools/data_sync`. Edit `shared/packets/outputs.toml` only when changing output routing. Packet pull is intentionally unsupported.

## Skills

Task-specific workflows live under `skills/*/SKILL.md`.

Use only the relevant skill for the current task. Do not load every skill for every prompt.

- `skills/micro-prompt/SKILL.md` for normal tiny implementation prompts.
- `skills/seam-first/SKILL.md` for adding/changing behavior without growing gravity-well files.
- `skills/go-gameplay-seam/SKILL.md` for server gameplay ownership changes.
- `skills/packet-schema-change/SKILL.md` for packet/schema/codec changes.
- `skills/godot-ui-scene-edit/SKILL.md` for Godot scene, HUD, menu, and layout changes.

## Important Conventions

- Keep authoritative gameplay logic on the Go game server.
- Keep presentation, UI, audio/effects, and interpolation in the Godot client.
- Keep server websocket/session transport in `services/game-server/internal/networking`.
- Keep server inbound packet handlers in `services/game-server/internal/networking/inbound`.
- Keep server outbound packet/write helpers in `services/game-server/internal/networking/outbound`.
- Keep reusable game simulation in `services/game-server/internal/game`, not `services/game-server/cmd/game-server/main.go`.
- Keep API/business logic out of the Go game server; it belongs in the planned `services/api-server/`.
- Use `shared/constants/server_constants.toml`, `shared/constants/server_entities.toml`, `shared/constants/client/presentation.toml`, `shared/constants/client/shell.toml`, and `shared/constants/client/lobby.toml` plus `tools/data_sync/` for active Go/GDScript constants.
- Use `shared/packets/outputs.toml`, `shared/packets/gameplay.toml`, `shared/packets/debug.toml`, and `shared/packets/lobby.toml` plus `tools/data_sync/` for active packets.
- Route server packet wire JSON through `services/game-server/internal/protocol/packetcodec`.
- Route client packet wire JSON through `client/scripts/networking/packets/packet_codec.gd`.
- Keep `PlayerID` player-facing and readable, for example `Player-1`/`Player-2`; do not convert it to UUID. UUID upgrades are for server-internal identities such as `SessionID` and `MemberID`.
- Keep client websocket transport in `client/scripts/networking/network_client.gd`.
- Keep client inbound server packet dispatch in `client/scripts/networking/inbound`.
- Keep client outbound client packet sends in `client/scripts/networking/outbound`.
- Keep client world sync and entity sync owners under `client/scripts/world/`.
- Keep packet-facing player lifecycle status in `StatePacket.player_lifecycle`, beside `players`.
- Client spectate/view-cycle eligibility must use authoritative lifecycle status (`active`) plus visual availability.
- Durable player counters such as score and lives belong to `playerSession`; `runtime.Ship` is the live avatar state and should not own those counters.
- Canonical targeting is `target_kind` + `target_id`.
- Player targets use `target_kind = "player"` and `target_id = playerID`.
- New gameplay systems must not use `target_player_id`.
- `target_player_id` is legacy and quarantined; it may only remain in devtools/debug player-only command paths until removed.
- Devtools/readouts should display `target_kind` + `target_id`, not `target_player_id`.
- Use `services/game-server/internal/game/motion` for per-entity movement integration and advance-with-wrap behavior.
- Use `services/game-server/internal/game/space` for gameplay distance, direction, and wrap-aware spatial math.
- Be careful with Godot scene diffs. Godot may rewrite `uid`, `unique_id`, offsets, imports, and scene metadata.
- Do not revert user/editor changes unless explicitly requested.
- Keep changes scoped. The user strongly prefers scalable structure without unnecessary code growth.

## Architecture / Seam Discipline

Read `docs/agent/architecture-rules.md` before adding ownership seams, moving packages/folders, changing lifecycle/networking/game-loop responsibilities, or editing broad coordination files. It is the canonical source for the detailed guardrails.

Core rules: identify the owning system first; keep policy in that owner and routing/composition thin; create the smallest concrete seam or stop/report when ownership is unclear; prefer behavior-preserving extraction; keep one seam per scoped change; preserve behavior unless explicitly authorized; and avoid unrelated cleanup, churn, refactors, or moves. Stop when the work crosses seams or expands materially. Generated/schema changes and scene/signal rewiring must be explicitly in scope.

## Where To Look First

Server gameplay:

- `services/game-server/internal/game/game.go`
- `services/game-server/internal/game/motion/`
- `services/game-server/internal/game/combat.go`
- `services/game-server/internal/game/session.go`
- `services/game-server/internal/game/match.go`
- `services/game-server/internal/game/rules/`
- `services/game-server/internal/game/spawning.go`
- `services/game-server/internal/game/spawning/`
- `services/game-server/internal/game/scoring/`
- `services/game-server/internal/game/scoring.go`
- `services/game-server/internal/game/runtime/`

Rooms/networking:

- `services/game-server/internal/networking/websocket.go`
- `services/game-server/internal/networking/rooms.go`

Client runtime:

- `client/scripts/shell/gameplay_hud_flow.gd`
- `client/scripts/shell/gameplay_menu_flow.gd`
- `client/scripts/shell/gameplay_runtime_tick_flow.gd`
- `client/scripts/world/`
- `client/scripts/world/world_sync.gd`
- `client/scripts/devtools/`
- `client/scripts/devtools/context/`
- `client/scenes/devtools/`
- `client/scripts/session/`
- `client/scripts/shell/gameplay_shell_flow.gd`
- `client/scripts/gameplay/runtime/`
- `client/scripts/gameplay/state/`
- `client/scripts/gameplay/input/`
- `client/scripts/ui/hud/`
- `client/scripts/ui/menus/`
- `client/scripts/ui/menu_flow/`
- `client/scripts/main_menu/`
- `client/scripts/gameplay/respawn/`
- `client/scripts/gameplay/spectate/`
- `client/scripts/presentation/background/`
- `client/scripts/gameplay/events/`
- `client/scripts/gameplay/effects/`
- `client/scripts/lobby/`
- `client/scripts/boot/`
- `client/scripts/config/`
- `client/scripts/networking/network_client.gd`
- `client/scripts/networking/packets/packet_codec.gd`
- `client/scripts/entities/player.gd`
- `client/scripts/ui/`

Shared schema/generation:

- `shared/constants/server_constants.toml`
- `shared/constants/server_entities.toml`
- `shared/constants/client/presentation.toml`
- `shared/constants/client/shell.toml`
- `shared/constants/client/lobby.toml`
- `shared/packets/outputs.toml`
- `shared/packets/gameplay.toml`
- `shared/packets/debug.toml`
- `shared/packets/lobby.toml`
- `services/game-server/internal/protocol/packetcodec/`
- `tools/data_sync/!README.md`
- `tools/data_sync/main.py`

## Documentation Discipline

Documentation is part of the implementation.

Update affected canonical service, protocol, data, systems-design, domain, devtools, development, planning, and limits documentation in the same change. Update implementation coverage when ownership changes and the behavioral-contract matrix when a durable invariant or protecting test changes. Do not report documentation as complete or current unless configured checks pass and known semantic gaps are disclosed.

## Agent Behavior Notes

- Open/read only the files needed for the requested edit.
- Do not inspect broadly unless the prompt explicitly asks for a scan or the named file directly points to a needed file.
- Focused, safe terminal checks are allowed when useful for the task.
- Avoid destructive git commands, broad cleanup, dependency upgrades, unrelated formatter runs, or expensive commands unless explicitly requested.
- Use `apply_patch` for manual edits.
- Do not use destructive git commands unless explicitly asked.
- Do not create broad refactors when a small change solves the request.
- If a task starts to balloon, stop and report why before adding large amounts of code.
- Preserve current behavior unless the user explicitly asks to change it.
- Keep implementation slices small enough for quick review.
- Before editing a known gravity-well file, use the line-count guardrails in `docs/agent/architecture-rules.md` as judgment, but do not run `wc -l` unless the prompt allows terminal commands.
- Implementation prompts must not broaden scope beyond the named target.
- If broader work appears necessary, stop and propose a follow-up prompt.
- Do not produce no-work prompts. Verification belongs in commands/checkpoints, not in separate agent prompts.
- When completing a numbered prompt, announce completion status at the bottom of the response/report using the exact format `**NOT COMPLETED PROMPT X**`, replacing `X` with the prompt number and removing `NOT` if the task succeeded.
- Do not report notes or edits unless explicitly asked to do so.

## Default Agent Report

```text
Documentation impact:
- Inspected:
- Updated:
- Not affected:
- Compliance check:
- Known documentation gaps:

Changed files:
- ...

Unexpected files touched:
- none / ...

Notes:
- ...

**<NOT >COMPLETED PROMPT X**
```

