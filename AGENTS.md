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

Run client GUT tests, if the `godot` CLI is available:

```bash
godot --headless --path client -s res://addons/gut/gut_cmdln.gd -gdir=res://tests/unit -ginclude_subdirs -gexit
```

Run the client constants boundary scan:

```bash
python3 -m pytest tools/tests/test_client_constants_boundary.py
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
client/scripts/constants/constants.gd
services/game-server/internal/constants/constants.go
```

Constants are managed by `tools/data_sync/` using `data-sync` blocks. Do not use `tools/scripts/generate_constants.py` for active constants changes.

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
- Route server packet wire JSON through `services/game-server/internal/protocol/packetcodec` instead of direct `encoding/json` calls. The codec is intentionally JSON-only and generic; do not add format switching, protobuf references, or an interface unless explicitly requested. Non-packet JSON such as collision-shape data-file parsing may still use `encoding/json` directly.
- Route client packet wire JSON through `client/scripts/networking/packet_codec/packet_codec.gd` instead of direct `JSON.stringify` or `JSON.parse_string` calls in websocket packet paths. The client codec is intentionally JSON-only and thin; do not add validation, format switching, typed packet objects, protobuf references, or generator changes unless explicitly requested. `network_client.gd` still owns websocket behavior.
- Keep packet-facing player lifecycle status in `StatePacket.player_lifecycle`, beside `players`. Do not put match lifecycle on `ShipState`; pending-respawn and eliminated players may not have active ship state.
- Client spectate/view-cycle eligibility must use authoritative lifecycle status (`active`) plus visual availability. Do not infer active eligibility solely from remote player positions or ship presence.
- Use `services/game-server/internal/game/motion` for per-entity movement integration and advance-with-wrap behavior.
- Use `services/game-server/internal/game/space` for new gameplay distance, direction, and wrap-aware spatial math for the toroidal world.
- Add focused Go tests for server gameplay rules that can regress.
- Add focused GUT tests for client packet, HUD, `world_sync`, and pure client logic regressions. Keep test-only helpers under `client/tests/`, not `client/scripts/`.
- Be careful with Godot scene diffs. Godot may rewrite `uid`, `unique_id`, offsets, imports, and scene metadata.
- Do not revert user/editor changes unless explicitly requested.
- When the user asks to "answer" or "report", do not edit files.
- Keep changes scoped. The user strongly prefers scalable structure without unnecessary code growth.

## Architecture / Seam Discipline

- Prefer small, explicit ownership seams over broad god files. If a change would add a new responsibility to an already-large file, stop and propose the smallest seam or same-package split first.
- Do not add new behavior to gravity-well files unless the prompt explicitly allows it. Known gravity-well candidates include broad lifecycle, networking, sync, shell, and game-loop files.
- A seam must have a concrete responsibility. Good seams include rooms, scoring, spawning, damage, player lifecycle, codec, domain events, session flow, logging, and presentation controllers. Avoid vague buckets like `utils`, `common`, `manager`, or `helpers` unless the responsibility is specific.
- Keep systems self-contained. A system may emit facts/events, expose narrow methods, or accept policy/config, but it should not reach into unrelated systems to make their decisions.
- Do not let integration seams become god objects. Domain events may define, queue, drain, and translate events, but must not decide scoring, damage, spawning, lives, achievements, API persistence, or other gameplay/business rules.
- When adding a feature, first identify the owning system. If no obvious owner exists, stop and report the missing seam instead of placing code in the nearest working file.
- Defer mechanics, not ownership. If a near-future feature clearly needs a home, add the minimal owning seam early even if the first behavior remains unchanged.
- Prefer behavior-preserving extraction before behavior change. First move or route existing behavior through the correct seam, then add new behavior in a later prompt.
- Same-package Go file splits are preferred for reducing god files when no new package boundary is needed. New Go packages/folders are architecture decisions and require a clear domain boundary.
- In Godot/GDScript, folder moves are less important than scene wiring. Avoid risky scene/node/path/signal changes unless required. Prefer extracting pure/helper/controller logic before changing scene ownership.
- Do not mix unrelated seams in one prompt. Codec, rooms, scoring, spawning, health, domain events, devtools, and client lifecycle should be changed in separate slices unless explicitly instructed otherwise.
- If a prompt cannot be completed with a small reviewable diff, stop and report the smallest next prompt instead.
- If implementation touches more than the named lifecycle/system path, stop and report why before continuing.
- Every architecture/refactor prompt must preserve current behavior unless it explicitly says behavior may change.
- Do not add broad cleanup, formatting-only churn, or opportunistic refactors while implementing a seam.
- Devtools must route through real gameplay seams. Do not create parallel debug-only gameplay logic that bypasses damage, lives, spawning, scoring, movement, room/session, or modifier systems.
- Constants/config should live with the smallest system that owns the decision. Do not globalize local presentation defaults unnecessarily, but do not bury gameplay, protocol, lifecycle, or environment policy in random files.
- Domain gameplay events should be emitted by owning systems and consumed later by achievements, stats, API summaries, logs, or notifications. Do not hardwire those future consumers into combat/scoring/spawning/lives code.
- Networking should transport, decode/route packets, manage websocket session state, and write responses. Room lifecycle policy belongs in rooms. Gameplay simulation belongs in game.
- `rooms` owns room creation, joining, leaving, readiness, lifecycle transitions, cleanup policy, and game instance ownership. `networking` may retain websocket session activation/deactivation when it mutates websocket session fields.
- `game` owns authoritative gameplay simulation, gameplay state mutation, and adapters from game storage into narrower gameplay seams. Match/mode policy evaluation belongs in `services/game-server/internal/game/rules`, which should receive plain snapshots/facts and return decisions/status. `game` should not own websocket transport, API persistence, account/auth concerns, or lobby UI flow.
- Before adding code to a large file, check whether an existing seam already owns the behavior. If it does, add the behavior there. If not, propose the seam first.

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
- `client/scripts/spectate_targets.gd`
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

Current areas include `game`, `networking`, `physics`, `rooms`, `scoring`, and `space`. Do not add new `*_test.go` files beside production packages under `services/game-server/internal/`. For game simulation setup, use the shared harness in `services/game-server/tests/game/helpers_test.go`; keep new helpers intent-level, such as placing entities or sending packets, instead of exposing raw private maps.

For server gameplay changes, run:

```bash
cd services/game-server
env GOCACHE=/tmp/space-rocks-go-build go test -buildvcs=false ./...
```

If the command prints read-only `envman` warnings but tests pass, those warnings have been harmless in this environment.

Godot client tests use GUT and live under `client/tests/`. Unit tests go under `client/tests/unit/`; fixtures go under `client/tests/fixtures/`; reusable test-only helpers go under `client/tests/helpers/`. Keep client tests focused on generated packets, HUD behavior, `world_sync`, and pure client logic. Do not put test helpers in `client/scripts/`.

For client changes, run GUT when the `godot` CLI is available:

```bash
godot --headless --path client -s res://addons/gut/gut_cmdln.gd -gdir=res://tests/unit -ginclude_subdirs -gexit
```

For client constants-boundary changes, run:

```bash
python3 -m pytest tools/tests/test_client_constants_boundary.py
```

Full gameplay/network smoke testing remains manual for now: opening the game scene, websocket connection, asteroid spawning, shooting/effects, pause/debug flow, and the full gameplay loop.

## Known Gaps / TODOs

- Pause/menu UI still needs smoke testing and may still be evolving.
- Window/gameplay balance should move away from raw OS max window pixels toward a logical gameplay viewport cap.
- Collision shape export/import should be verified after the Godot 4.6 upgrade.
- Generated Godot constants/packet files may eventually move into a generated folder, but they currently live under `client/scripts/`.
- API server is planned but not scaffolded.
- Ship variants are planned but not implemented. Toroidal wrapping is implemented and still needs manual gameplay smoke testing after related changes.

## Agent Behavior Notes

- Inspect before editing. This repo changes quickly.
- Prefer `rg`/`rg --files` for searches.
- Use `apply_patch` for manual edits.
- Do not use destructive git commands unless explicitly asked.
- Do not create broad refactors when a small change solves the request.
- If a task starts to balloon, stop and report why before adding large amounts of code.
- Preserve current behavior unless the user explicitly asks to change it.
- Keep implementation slices small enough for quick review. Verification commands may run longer, but the code diff should remain small.
- After each implementation prompt, run the requested validation command and report the exact command, result, and `git status --short`.
- If tests fail, stop and report the failure. Do not continue piling changes onto a failing state unless the prompt explicitly asks for a focused fix.
- Read-only prompts must not edit files, run formatters, or perform cleanup.
- Implementation prompts must not broaden scope beyond the named target. If broader work appears necessary, stop and propose a follow-up prompt.
- When completing a numbered prompt, announce completion at the bottom of the response/report using the exact format `**COMPLETED PROMPT X**`, replacing `X` with the prompt number.
