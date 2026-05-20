# AGENTS.md

Guidance for Codex and other coding agents working in this repository.

This file is intentionally practical. Read it before making changes, then use the linked docs for deeper context.

## Project Snapshot

Space Rocks is an Asteroids-inspired game in active development.

- Godot client: `client/`
- Go real-time game server: `services/game-server/`
- Planned API/business server placeholder: `services/api-server/`
- Shared generated data sources: `shared/`
- Project docs: `docs/`
- Packet generation scripts: `tools/scripts/`
- Active constants sync tool: `tools/data_sync/`

The current gameplay direction is server-authoritative. The Godot client handles rendering, UI, audio/effects, local input collection, and interpolation. The Go game server owns simulation outcomes: movement, bullets, collisions, scoring, lives, death, respawn, pause safety, rooms, and websocket state.

## Read First

For practical development workflow:

- `docs/developer.md`

For project memory and recent context:

- `docs/notes.md`

For architecture:

- `docs/design/architecture.md`

For logging:

- `docs/server/logging.md`

For developer toggles:

- `docs/devtools/toggles.md`

For planned API service:

- `docs/api/nestjs-api-server.md`

For planned future systems:

- `docs/design/toroidal-wrap.md`
- `docs/design/ship-variants.md`

## Current Layout Notes

The Go game server was moved from `server/` to:

```text
services/game-server/
```

Its Go module path is still:

```text
github.com/Lokee86/space-rocks/server
```

That mismatch is currently intentional. Import paths inside the Go server still use `github.com/Lokee86/space-rocks/server/...`.

`services/api-server/` exists as an empty placeholder directory for a future Node.js/TypeScript NestJS API service. Do not put account, persistence, matchmaking metadata, leaderboard, or other business/backend concerns into the Go game server unless the user explicitly changes that direction.

## Commands

Run the Go game server:

```bash
cd services/game-server
go run ./cmd/game-server
```

Run with Air, if installed:

```bash
cd services/game-server
air
```

Run server tests:

```bash
cd services/game-server
go test -buildvcs=false ./...
```

Preferred test command when cache/environment issues appear:

```bash
cd services/game-server
env GOCACHE=/tmp/space-rocks-go-build go test -buildvcs=false ./...
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

Preview active shared constants:

```bash
python3 tools/data_sync/main.py -diff -constants -go -gds
```

Apply active shared constants:

```bash
python3 tools/data_sync/main.py -push -constants -go -gds
```

Validate shared packets:

```bash
python3 tools/data_sync/main.py -validate -packets
```

Preview shared packets:

```bash
python3 tools/data_sync/main.py -diff -packets -go -gds
```

Apply shared packets:

```bash
python3 tools/data_sync/main.py -push -packets -go -gds
```

Open the Godot project by opening/importing:

```text
client/
```

The configured main scene is:

```text
res://scenes/game.tscn
```

## Generated Files

Do not hand-edit generated files unless the user explicitly asks for a temporary/manual intervention.

Constants source of truth:

```text
shared/game_data.toml
```

Generated constants:

```text
client/scripts/constants.gd
services/game-server/internal/constants/constants.go
```

Constants are managed by `tools/data_sync/` using `data-sync` blocks. Do not use `tools/scripts/generate_constants.py` for active constants changes.

Packet source of truth:

```text
shared/packets/packets.toml
```

Generated packets:

```text
client/scripts/packets.gd
services/game-server/internal/game/packets.go
services/game-server/internal/game/entities/packets_generated.go
```

Collision shape source:

```text
shared/collisions/collision_shapes.json
```

Godot export helper:

```text
client/tools/export_collision_shapes.gd
```

Data sync tool:

```text
shared/game_data.toml
shared/packets/packets.toml
tools/data_sync/
```

`shared/game_data.toml` is active for constants. `shared/packets/packets.toml` is active for packets. TypeScript output is future/deferred until the API service exists.

The packet TOML schema preserves outputs, structs, packet_types, builders, imports, Go package mappings, GDScript builders, arrays/maps/custom struct refs, and rich type strings such as `map<string,ShipState>` and `array<EventState>`. `shared/game_data.toml` should contain constants only; obsolete packet reference data was removed when the packet TOML pipeline was adopted.

Packet pull is intentionally unsupported. Packet schema changes should be made in `shared/packets/packets.toml` and pushed with `tools/data_sync`.

## Logging

The server has a custom structured logging wrapper:

```text
services/game-server/internal/logging/logger.go
```

Use it for server logs. Do not add raw `log.Println` or a second logging package.

Category loggers:

- `logging.Server`
- `logging.Network`
- `logging.Rooms`
- `logging.Game`

Environment variables:

- `LOG_LEVEL`
- `LOG_SERVER`
- `LOG_NETWORK`
- `LOG_ROOMS`
- `LOG_GAME`

Supported levels include `debug`, `info`, `warn`, `warning`, `error`, and `off`. Default is quiet: unset `LOG_LEVEL` resolves to `warn`.

Normal lifecycle logs should usually be `Debug`. Warnings are for unusual recoverable situations. Errors are for real failed operations.

## Important Conventions

- Keep authoritative gameplay logic on the Go game server.
- Keep presentation, UI, audio/effects, and interpolation in the Godot client.
- Keep websocket and room transport in `services/game-server/internal/networking`.
- Keep reusable game simulation in `services/game-server/internal/game`, not `cmd/game-server/main.go`.
- Keep API/business logic out of the Go game server; it belongs in the planned `services/api-server/`.
- Use `shared/game_data.toml` plus `tools/data_sync/` for active Go/GDScript constants. Use `shared/packets/packets.toml` plus `tools/data_sync/` for active packets. TypeScript output is future/deferred.
- Use `services/game-server/internal/game/space` for new gameplay distance, direction, and position-normalization logic. It is flat/infinite today, but exists to contain future wrapped-world support.
- Add focused Go tests for server gameplay rules that can regress.
- Be careful with Godot scene diffs. Godot may rewrite `uid`, `unique_id`, offsets, imports, and scene metadata.
- Do not revert user/editor changes unless explicitly requested.
- When the user asks to "answer" or "report", do not edit files.
- Keep changes scoped. The user strongly prefers scalable structure without unnecessary code growth.

## Where To Look First

Server gameplay:

- `services/game-server/internal/game/game.go`
- `services/game-server/internal/game/combat.go`
- `services/game-server/internal/game/session.go`
- `services/game-server/internal/game/spawning.go`
- `services/game-server/internal/game/scoring.go`
- `services/game-server/internal/game/entities/`

Rooms/networking:

- `services/game-server/internal/networking/websocket.go`
- `services/game-server/internal/networking/rooms.go`

Client runtime:

- `client/scripts/ui/game_shell.gd`
- `client/scripts/game.gd`
- `client/scripts/network_client.gd`
- `client/scripts/world_sync.gd`
- `client/scripts/player.gd`
- `client/scripts/ui/hud_controller.gd`

Shared schema/generation:

- `shared/game_data.toml`
- `shared/packets/packets.toml`
- `tools/data_sync/README.md`
- `tools/data_sync/main.py`

## Critical Current Context

- The repo is likely dirty. The `server/` to `services/game-server/` move may show as many deletes plus a new `services/` tree until committed.
- There are unrelated Godot/editor asset changes in the worktree. Do not clean or revert them casually.
- The user recently moved backend structure and wants docs/plans to reflect a future NestJS API server separated from the Go game server.
- If gameplay or input looks broken, first confirm the Go server is running and the Godot client is connected. This caused a false pause-feature debugging path before.
- Godot was upgraded to 4.6 recently. Scene/import diffs may be noisy.
- The older `space-rocks-(4.3)/` project copy is ignored and should not be used as the active project.
- Generated recordings and build artifacts should not be committed. In particular, avoid committing `*.avi`, `tmp/`, `*/tmp/`, and `client/.godot/`.

## Implemented Developer Toggles

Current hardcoded Godot hotkeys:

- `F1`: toggle debug invincibility for the player
- `F2`: toggle debug infinite lives for the player
- `F3`: toggle room-wide debug world freeze

These are server-authoritative toggles sent through generated packets. See `docs/devtools/toggles.md`.

## Pause State Context

Pause plumbing exists:

- packets: `pause_player`, `resume_player`
- server player fields include paused/invulnerability state
- paused players should ignore input, not shoot/score, not take asteroid damage, and be hidden by client world sync
- resume starts a short invulnerability window
- menu UI has been in flux

If pause behavior seems wrong, inspect current Godot scenes/scripts before changing code. The HUD/menu scenes have been changed multiple times.

## Future Plans Already Documented

Toroidal/wrapped world:

- Use `services/game-server/internal/game/space` as the abstraction point.
- Future server coordinates should become bounded/wrapped.
- Future client rendering should use unwrapped visual positions relative to the local player so border crossing is invisible.
- See `docs/design/toroidal-wrap.md`.

Ship variants:

- Future ships may use different client scenes and server collision maps.
- See `docs/design/ship-variants.md`.

API server:

- Planned as Node.js/TypeScript/NestJS in `services/api-server/`.
- It should own business/backend concerns, not real-time simulation.
- See `docs/api/nestjs-api-server.md`.

## Testing Expectations

Go server tests live under:

```text
services/game-server/tests/<area>/
```

Current areas include `game`, `networking`, `physics`, and `space`. Do not add new `*_test.go` files beside production packages under `services/game-server/internal/`. For game simulation setup, use the shared harness in `services/game-server/tests/game/helpers_test.go`; keep new helpers intent-level, such as placing entities or sending packets, instead of exposing raw private maps.

For server gameplay changes, run:

```bash
cd services/game-server
env GOCACHE=/tmp/space-rocks-go-build go test -buildvcs=false ./...
```

If the command prints read-only `envman` warnings but tests pass, those warnings have been harmless in this environment.

For Godot/client changes, there is no established automated test command. Prefer careful file inspection, then tell the user what needs manual smoke testing in Godot.

## Known Gaps / TODOs

- Pause/menu UI still needs smoke testing and may still be evolving.
- Window/gameplay balance should move away from raw OS max window pixels toward a logical gameplay viewport cap.
- Collision shape export/import should be verified after the Godot 4.6 upgrade.
- Generated Godot constants/packet files may eventually move into a generated folder, but they currently live under `client/scripts/`.
- API server is planned but not scaffolded.
- Toroidal wrapping and ship variants are planned but not implemented.

## Agent Behavior Notes

- Inspect before editing. This repo changes quickly.
- Prefer `rg`/`rg --files` for searches.
- Use `apply_patch` for manual edits.
- Do not use destructive git commands unless explicitly asked.
- Do not create broad refactors when a small change solves the request.
- If a task starts to balloon, stop and report why before adding large amounts of code.
- Preserve current behavior unless the user explicitly asks to change it.
