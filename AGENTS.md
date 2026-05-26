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

The current gameplay direction is server-authoritative. The Godot client handles rendering, UI, audio/effects, local input collection, and interpolation. The Go game server owns simulation outcomes: movement, bullets, collisions, scoring, lives, death, respawn, pause safety, rooms, and websocket state.

## Read First

For normal workflow and commands:

- `docs/developer.md`
- `docs/agent/testing.md`

For project memory and recent volatile context:

- `docs/agent/current-context.md`
- `docs/notes.md`

For architecture and seam rules:

- `docs/agent/architecture-rules.md`
- `docs/design/architecture.md`

For Godot/editor/client-specific notes:

- `docs/agent/godot-notes.md`
- `docs/devtools/toggles.md`
- `docs/design/toroidal-wrap.md`
- `docs/design/ship-variants.md`

For server/API/logging details:

- `docs/agent/server-notes.md`
- `docs/server/logging.md`
- `docs/api/nestjs-api-server.md`

## Current Layout Notes

The Go game server lives at:

```text
services/game-server/
```

Its Go module path is still:

```text
github.com/Lokee86/space-rocks/server
```

That mismatch is intentional for now. Import paths inside the Go server still use `github.com/Lokee86/space-rocks/server/...`.

`services/api-server/` exists as an empty placeholder for a future Node.js/TypeScript/NestJS API service. Do not put account, persistence, matchmaking metadata, leaderboard, or other business/backend concerns into the Go game server unless the user explicitly changes that direction.

## Core Commands

Run server tests:

```bash
cd services/game-server
go test -buildvcs=false ./...
```

Preferred server test command when cache/environment issues appear:

```bash
cd services/game-server
env GOCACHE=/tmp/space-rocks-go-build go test -buildvcs=false ./...
```

Run client GUT tests, if the `godot` CLI is available:

```bash
godot --headless --path client -s res://addons/gut/gut_cmdln.gd -gdir=res://tests/unit -ginclude_subdirs -gexit
```

Build the game server:

```bash
cd services/game-server
go build -buildvcs=false -o ./tmp/game-server ./cmd/game-server
```

Validate active shared constants:

```bash
python3 tools/data_sync/main.py -validate -constants
```

Validate shared packets:

```bash
python3 tools/data_sync/main.py -validate -packets
```

For full testing expectations and additional data-sync commands, read `docs/agent/testing.md`.

## Generated Files

Do not hand-edit generated files unless the user explicitly asks for a temporary/manual intervention.

Constants source of truth:

```text
shared/game_data.toml
```

Generated constants:

```text
client/scripts/constants/constants.gd
services/game-server/internal/constants/constants.go
```

Packet source of truth:

```text
shared/packets/packets.toml
```

Generated packets:

```text
client/scripts/networking/packets.gd
services/game-server/internal/game/packets.go
services/game-server/internal/game/entities/packets_generated.go
```

Collision shape source:

```text
shared/collisions/collision_shapes.json
```

Data sync tool:

```text
tools/data_sync/
```

Use `shared/game_data.toml` plus `tools/data_sync/` for active constants. Use `shared/packets/packets.toml` plus `tools/data_sync/` for active packets. TypeScript output is future/deferred until the API service exists.

Tunable/game-data constants belong in `shared/game_data.toml` and generated scripts under `client/scripts/constants/`. Do not create local constants files elsewhere; change generated constants through the data source/regeneration path, not manual edits.

Packet schema changes should be made in `shared/packets/packets.toml` and pushed with `tools/data_sync`. Packet pull is intentionally unsupported.

## Skills

Task-specific workflows live under `skills/*/SKILL.md`.

Use the relevant skill before doing that kind of work:

- `skills/agent-micro-refactor/SKILL.md` for normal tiny implementation prompts.
- `skills/godot-seam-refactor/SKILL.md` for splitting or shrinking Godot scripts.
- `skills/go-gameplay-seam/SKILL.md` for server gameplay ownership changes.
- `skills/packet-schema-change/SKILL.md` for packet/schema/codec changes.
- `skills/godot-ui-scene-edit/SKILL.md` for Godot scene, HUD, menu, and layout changes.

Do not load every skill for every task. Load only the one that matches the current prompt.

## Important Conventions

- Keep authoritative gameplay logic on the Go game server.
- Keep presentation, UI, audio/effects, and interpolation in the Godot client.
- Keep websocket and room transport in `services/game-server/internal/networking`.
- Keep reusable game simulation in `services/game-server/internal/game`, not `cmd/game-server/main.go`.
- Keep API/business logic out of the Go game server; it belongs in the planned `services/api-server/`.
- Use `shared/game_data.toml` plus `tools/data_sync/` for active Go/GDScript constants.
- Use `shared/packets/packets.toml` plus `tools/data_sync/` for active packets.
- Route server packet wire JSON through `services/game-server/internal/protocol/packetcodec`.
- Route client packet wire JSON through `client/scripts/networking/packet_codec/packet_codec.gd`.
- Keep packet-facing player lifecycle status in `StatePacket.player_lifecycle`, beside `players`.
- Client spectate/view-cycle eligibility must use authoritative lifecycle status (`active`) plus visual availability.
- Use `services/game-server/internal/game/motion` for per-entity movement integration and advance-with-wrap behavior.
- Use `services/game-server/internal/game/space` for gameplay distance, direction, and wrap-aware spatial math.
- Add focused Go tests for server gameplay rules that can regress.
- Add focused GUT tests for client packet, HUD, `world_sync`, and pure client logic regressions.
- Keep test-only helpers under `client/tests/`, not `client/scripts/`.
- Be careful with Godot scene diffs. Godot may rewrite `uid`, `unique_id`, offsets, imports, and scene metadata.
- Do not revert user/editor changes unless explicitly requested.
- When the user asks to "answer" or "report", do not edit files.
- Keep changes scoped. The user strongly prefers scalable structure without unnecessary code growth.

## Architecture / Seam Discipline

Read `docs/agent/architecture-rules.md` before adding ownership seams, moving packages/folders, changing lifecycle/networking/game-loop responsibilities, or editing known gravity-well files.

Core rules:

- Prefer small, explicit ownership seams over broad god files.
- If a change would add a new responsibility to an already-large file, stop and propose the smallest seam or same-package split first.
- Do not add new behavior to gravity-well files unless the prompt explicitly allows it.
- Known gravity-well candidates include broad lifecycle, networking, sync, shell, and game-loop files.
- When adding a feature, first identify the owning system.
- If no obvious owner exists, stop and report the missing seam instead of placing code in the nearest working file.
- Defer mechanics, not ownership.
- Prefer behavior-preserving extraction before behavior change.
- Do not mix unrelated seams in one prompt.
- Every architecture/refactor prompt must preserve current behavior unless it explicitly says behavior may change.
- Do not add broad cleanup, formatting-only churn, or opportunistic refactors while implementing a seam.

Line-count guardrails for hand-written production files:

- Prefer files under roughly 200 lines when practical.
- Around 300 lines, check whether the file still has one clear responsibility.
- Around 350 lines, avoid adding new responsibility unless it clearly belongs there.
- Around 500 lines, treat actively changing files as split/refactor candidates.
- Over 500 lines, prefer routing/extraction through an owning seam unless the prompt explicitly says to edit that file.
- Generated files, Godot `.tscn` scene files, `.tres` resources, vendored addons, fixtures, snapshots, and large declarative data files are exempt.

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
- `services/game-server/internal/game/entities/`

Rooms/networking:

- `services/game-server/internal/networking/websocket.go`
- `services/game-server/internal/networking/rooms.go`

Client runtime:

- `client/scripts/ui/game_shell.gd`
- `client/scripts/game.gd`
- `client/scripts/networking/network_client.gd`
- `client/scripts/networking/packet_codec/packet_codec.gd`
- `client/scripts/networking/world_sync.gd`
- `client/scripts/entities/player.gd`
- `client/scripts/ui/hud_controller.gd`

Client tests:

- `client/tests/README.md`
- `client/tests/unit/`
- `client/tests/fixtures/`
- `client/tests/helpers/`

Shared schema/generation:

- `shared/game_data.toml`
- `shared/packets/packets.toml`
- `services/game-server/internal/protocol/packetcodec/`
- `tools/data_sync/README.md`
- `tools/data_sync/main.py`

## Testing Expectations

For detailed testing rules and commands, read `docs/agent/testing.md`.

Minimum expectation after implementation prompts:

- Run the requested validation command.
- Report the exact command, result, and `git status --short`.
- If tests fail, stop and report the failure. Do not continue piling changes onto a failing state unless the prompt explicitly asks for a focused fix.

## Agent Behavior Notes

- Inspect before editing. This repo changes quickly.
- Prefer `rg`/`rg --files` for searches.
- Use `apply_patch` for manual edits.
- Do not use destructive git commands unless explicitly asked.
- Do not create broad refactors when a small change solves the request.
- If a task starts to balloon, stop and report why before adding large amounts of code.
- Preserve current behavior unless the user explicitly asks to change it.
- Keep implementation slices small enough for quick review. Verification commands may run longer, but the code diff should remain small.
- Before editing a known gravity-well file, check its approximate size with `wc -l`.
- If a file is already over 350 lines, avoid adding new responsibility there.
- If a file is over 500 lines, prefer extracting/routing through an owning seam unless the prompt explicitly says to edit that file.
- Read-only prompts must not edit files, run formatters, or perform cleanup.
- Implementation prompts must not broaden scope beyond the named target.
- If broader work appears necessary, stop and propose a follow-up prompt.
- When completing a numbered prompt, announce completion at the bottom of the response/report using the exact format `**COMPLETED PROMPT X**`, replacing `X` with the prompt number.
