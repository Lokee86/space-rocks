# Developer Handoff

This document is a practical handoff for future development sessions. It focuses on how to work on the project. For broader architecture, see [docs/design/architecture.md](design/architecture.md). For server logging details, see [docs/server/logging.md](server/logging.md).

## Project Overview

Space Rocks is an Asteroids-inspired game with:

- a Godot client in `client/`
- a Go game server in `services/game-server/`
- a planned API server in `services/api-server/`
- shared data sources in `shared/`
- the active constants sync tool in `tools/data_sync/`
- packet generation scripts in `tools/scripts/`

The current direction is server-authoritative for gameplay state. The client collects input and renders/interpolates state; the server owns simulation outcomes such as movement, bullets, asteroid collisions, scoring, lives, death, respawn, room state, and pause safety rules.

The project is in active development. Expect rough edges and incomplete UI around newer systems.

## Repository Structure

- `client/`: Godot project. Scenes, scripts, assets, audio, shaders, client tools, and generated client constants/packet helpers.
- `services/game-server/`: Go module for the real-time game server. Main entrypoint is `services/game-server/cmd/game-server/main.go`.
- `services/api-server/`: empty placeholder for a planned API server service for business/backend systems. Intended stack is Node.js/TypeScript with NestJS.
- `shared/`: source data used by both client and server:
  - `shared/game_data.toml` for active constants
  - `shared/packets/packets.toml` for active packets
  - `shared/collisions/collision_shapes.json`
- `tools/scripts/`: Python generators and conversion helpers.
- `docs/`: Documentation.
- `SourceAssets/`: Source art files. This path is ignored by Git.
- `space-rocks-(4.3)/`: ignored older Godot project copy.

## Client And Server Fit

The Godot client connects to the Go server over websocket:

```text
ws://localhost:8080/ws
```

Websocket connection is session-only. Multiplayer rooms are created and joined with generated packets after connecting to `/ws`; the old `room_id` query path is not used by the real UI.

Legacy direct-room compatibility is quarantined in `services/game-server/internal/rooms`: `GetOrCreate()` and `Join()` create or join already-started direct game rooms and should not be used for lobby-created multiplayer flow. `DefaultRoom()` has been removed; keep any future room lifecycle work on the explicit create/join/start/return APIs.

Current multiplayer lifecycle:

```text
Main Menu -> Multiplayer Dialog -> /ws session -> CreateRoomRequest/JoinRoomRequest -> Lobby -> SetReadyRequest -> StartGameRequest -> InGame -> GameOver -> ReturnToLobbyRequest or LeaveRoomRequest
```

Runtime flow:

1. Godot collects input in `client/scripts/entities/player.gd`.
2. `client/scripts/game.gd` sends input/config/respawn/pause packets through `client/scripts/networking/network_client.gd`.
3. `services/game-server/internal/networking/websocket.go` reads client packets.
4. Lobby/lifecycle packets call `services/game-server/internal/rooms` for domain decisions and networking sends snapshots/errors.
5. After a room reaches `InGame`, the room's `*game.Game` handles gameplay packets and advances simulation.
6. The server sends state packets at the server tick rate.
7. `client/scripts/networking/world_sync.gd` applies/interpolates state to Godot nodes.
8. HUD and effects update from state/events.

## Run The Server

From the repo root:

```bash
cd services/game-server
go run ./cmd/game-server
```

With Air hot reload, if installed:

```bash
cd services/game-server
air
```

The Air config is `services/game-server/.air.toml`. It builds:

```bash
go build -buildvcs=false -o ./tmp/game-server ./cmd/game-server
```

The server listens on `:8080` and exposes:

- `GET /health`
- `GET /ws`

## Open/Run The Godot Client

Open Godot and import/open:

```text
client/
```

The configured main scene is:

```text
res://scenes/game.tscn
```

If the `godot` command is available locally, this may work:

```bash
godot --path client
```

The client expects the Go server to already be running for gameplay.

## Common Development Commands

Run the server:

```bash
cd services/game-server
go run ./cmd/game-server
```

Run all server tests:

```bash
cd services/game-server
go test -buildvcs=false ./...
```

Use an explicit cache path when the shell environment has cache or permission issues:

```bash
cd services/game-server
env GOCACHE=/tmp/space-rocks-go-build go test -buildvcs=false ./...
```

Run client GUT tests, if the `godot` CLI is available:

```bash
godot --headless --path client -s res://addons/gut/gut_cmdln.gd -gdir=res://tests/unit -ginclude_subdirs -gexit
```

Run the client constants-boundary scan:

```bash
python3 -m pytest tools/tests/test_client_constants_boundary.py
```

## Server Test Layout

Go server tests are kept out of production package folders. Put server tests under:

```text
services/game-server/tests/<area>/
```

Current areas:

- `services/game-server/tests/game/`
- `services/game-server/tests/networking/`
- `services/game-server/tests/physics/`
- `services/game-server/tests/rooms/`
- `services/game-server/tests/scoring/`
- `services/game-server/tests/space/`

Do not add new `*_test.go` files beside production code under `services/game-server/internal/`.

Game simulation tests should use the shared harness:

```text
services/game-server/tests/game/helpers_test.go
```

Keep harness helpers gameplay-oriented and deliberate: create a scenario, add players, send packets, step simulation, decode state, place entities, set collision presets, or adjust session state needed for precise behavior tests. Avoid exposing raw private maps directly to individual tests.

Use same-package tests under `services/game-server/internal/` only for tiny unexported seams that should not become production API just to make tests compile. Keep those exceptions focused on pure conversion or helper behavior. The current collision detection seam is covered by existing game behavior tests; do not export its helpers only to test them directly.

## Client Test Layout

Godot client tests use GUT and live under:

```text
client/tests/
```

Current layout:

- `client/tests/unit/`: focused unit-style GUT tests.
- `client/tests/fixtures/`: small test data and scene fixtures.
- `client/tests/helpers/`: reusable test-only helpers.

Keep test-only helpers out of `client/scripts/`. Client tests should focus on generated packets, HUD behavior, `world_sync`, missing server-field safety, constants-boundary assumptions, and pure client logic. Do not turn these into full gameplay/network integration tests.

Run the GUT suite with:

```bash
godot --headless --path client -s res://addons/gut/gut_cmdln.gd -gdir=res://tests/unit -ginclude_subdirs -gexit
```

Expected missing-field warnings may appear in tests that intentionally verify safe behavior for missing `lives`, `respawn_delay`, or asteroid `scale`; those warnings are fine when the suite passes.

The static client constants-boundary test lives at:

```text
tools/tests/test_client_constants_boundary.py
```

Run it with:

```bash
python3 -m pytest tools/tests/test_client_constants_boundary.py
```

Manual smoke testing remains the boundary for opening the game scene, websocket connection, asteroid spawning, shooting/effects, pause/debug flow, and the full gameplay loop.

Build the server binary:

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

Preview shared packet generation:

```bash
python3 tools/data_sync/main.py -diff -packets -go -gds
```

Apply shared packets:

```bash
python3 tools/data_sync/main.py -push -packets -go -gds
```

## Generated Files

Do not hand-edit generated files unless there is a specific reason and the source/generator will be updated later.

Constants source:

```text
shared/game_data.toml
```

Generated outputs:

```text
client/scripts/constants/constants.gd
services/game-server/internal/constants/constants.go
```

Constants are managed by `tools/data_sync/` using marked `data-sync` blocks. Do not use `tools/scripts/generate_constants.py` for active constants changes.

Server-owned constants live under `constants.server.*` in `shared/game_data.toml`. World size is an intentional exception to the usual server-only filtering: `constants.server.world` is generated to both Go and GDScript because client visual wrapping must use the same bounds as server simulation. `player_starting_lives` and `player_respawn_delay` live under `constants.server.player_lifecycle`; `asteroid_size_scale` lives under `constants.server.asteroids`. The client must not import those gameplay-rule constants directly: it receives lives through player/state packets, respawn delay through death events, and asteroid visual scale through asteroid state.

Client constants output is filtered by `tools/data_sync/config.toml`; a constant may remain in the source of truth while being intentionally omitted from `client/scripts/constants/constants.gd`. If world size changes, run `python3 tools/data_sync/main.py -push -constants -go -gds` so both server and client wrap bounds update together.

Packet source:

```text
shared/packets/packets.toml
```

Generated outputs:

```text
client/scripts/networking/packets.gd
services/game-server/internal/game/packets.go
services/game-server/internal/game/entities/packets_generated.go
```

TOML data sync paths:

```text
shared/game_data.toml
shared/packets/packets.toml
tools/data_sync/
```

`shared/game_data.toml` is active for constants. `shared/packets/packets.toml` is active for packets.

The packet TOML schema preserves the old rich JSON behavior: outputs, structs, packet_types, builders, imports, Go package mappings, GDScript builders, arrays/maps, custom struct references, Go type overrides, and rich type strings such as `map<string,ShipState>` and `array<EventState>`.

Packet-facing player lifecycle status lives on `StatePacket.player_lifecycle`, beside `players`. Do not put lifecycle status on `ShipState`: pending-respawn and eliminated players may not have active ship state.

`shared/game_data.toml` contains constants only. Obsolete packet reference sections were removed when the packet TOML pipeline was adopted.

The `tools/data_sync/` tool supports:

```text
-push
-pull
-diff
-check
-validate
```

The active constants path uses `-constants -go -gds`. The active packet path uses `-packets -go -gds`. TypeScript output is future/deferred. Full packet schema pull is not supported; edit packet schema in `shared/packets/packets.toml` and push from TOML. See `tools/data_sync/README.md` for config format, markers, ownership rules, and migration details.

The one-time JSON to TOML migrations seeded `shared/game_data.toml` and `shared/packets/packets.toml`. The old constants and packet JSON sources have been retired.

Collision shapes source:

```text
shared/collisions/collision_shapes.json
```

Godot-side collision export helper:

```text
client/tools/export_collision_shapes.gd
```

TODO: verify whether the collision export tool currently exports every shape needed by the server. Earlier inspection showed bullet/asteroid export paths clearly; ship collision shape freshness should be checked before relying on regenerated collision data.

## Windows / WSL Notes

The current workspace path indicates WSL accessing a Windows drive:

```text
/mnt/d/!bin/space-rocks
```

Useful notes:

- Godot runs on Windows in the current workflow.
- Go commands can run from WSL inside `services/game-server/`.
- The Godot editor may rewrite `.tscn`, `.godot`, `.uid`, or import metadata when opened/upgraded.
- Use `go test -buildvcs=false ./...` because the repo uses this consistently and it avoids build VCS stamping problems.
- If shell startup prints read-only `envman` warnings, those have not prevented Go tests from passing.

## Git LFS And Ignored Files

`.gitattributes` configures Git LFS for:

```text
*.png
*.webp
*.wav
*.mp3
```

Ignored or should-not-commit items include:

- `client/.godot/`
- `client/.export/`
- `client/.import/`
- `tmp/`
- `*/tmp/`
- `temp/`
- `*/temp/`
- `.cache/`
- `*.avi`
- `*.tmp`
- `*.temp`
- `SourceAssets/`
- `space-rocks-(4.3)/`
- secrets: `.env`, `secrets/`, `keys/`

Known current cleanup note: generated recordings/build artifacts such as `client/game-clip.avi`, `services/game-server/tmp/game-server`, and old command `tmp` outputs should not be committed.

## Logging And Config

The server has a custom structured logging package:

```text
services/game-server/internal/logging/logger.go
```

It wraps Go `log/slog` and exposes category loggers:

- `logging.Server`
- `logging.Network`
- `logging.Rooms`
- `logging.Game`

Use the category loggers for new server code. Do not introduce a second logging framework.

Environment variables:

- `LOG_LEVEL`
- `LOG_SERVER`
- `LOG_NETWORK`
- `LOG_ROOMS`
- `LOG_GAME`

Supported levels include `debug`, `info`, `warn`, `warning`, `error`, and `off`. Default logging is quiet: unset `LOG_LEVEL` resolves to `warn`.

See [docs/server/logging.md](server/logging.md) for usage, examples, field names, and troubleshooting.

The client has a lightweight GDScript logger:

```text
client/scripts/logging/logger.gd
```

Prefer `ClientLogger` over raw `print()` for new client lifecycle, UI, networking, packet, HUD, input, and world-sync diagnostics. See [docs/client/logging.md](client/logging.md).

## Important Conventions

- Keep authoritative gameplay rules on the server.
- Keep rendering, local audio/effects, UI, and interpolation in the Godot client.
- Keep room/domain lifecycle ownership in `services/game-server/internal/rooms`: create/join/leave, readiness, start-game, single-player startup, return-to-lobby, game-over transition, game ownership, and cleanup policy.
- Keep websocket/session/packet transport in `services/game-server/internal/networking`: websocket upgrade/read/write loops, packet dispatch, per-connection session fields, player activation/deactivation, snapshots, and errors.
- Keep reusable simulation and gameplay state mutation in `services/game-server/internal/game`.
- Keep match/mode policy decisions in `services/game-server/internal/game/rules`. Rules should receive plain snapshots/facts and return decisions/status; they should not import `game` or `rooms`.
- Keep pure score policy in `services/game-server/internal/game/scoring`. Game code should create scoring events and apply returned awards through game-owned state/session seams.
- Keep packet-facing player lifecycle status sourced from `Game.MatchDecision()` or the same game-owned projection seam. The client should consume `StatePacket.player_lifecycle`, not infer lifecycle from `StatePacket.players` or rendered ship presence.
- Keep client spectate/view-cycle eligibility based on authoritative lifecycle status plus visual availability. Eligible targets should be `active`; `pending_respawn`, `eliminated`, and missing lifecycle status should not be treated as active targets.
- Keep per-entity movement integration and advance-with-wrap behavior in `services/game-server/internal/game/motion`. `Game.Step()` should call the motion seam for individual entities while retaining map iteration, gates, deletion, spawning, collision, scoring, and lifecycle order.
- Use `services/game-server/internal/game/space` for new gameplay distance, direction, and position-normalization logic. It is the wrap-aware server spatial layer for toroidal world behavior.
- Keep reusable simulation code out of `cmd/game-server/main.go`.
- Keep business/backend API logic out of the Go game server. The planned `services/api-server/` service should own accounts, persistence, matchmaking metadata, leaderboards, and other non-real-time concerns.
- Use `shared/game_data.toml` plus `tools/data_sync/` for active constants. Use `shared/packets/packets.toml` plus `tools/data_sync/` for active packets.
- Add focused Go tests for server gameplay rules that can regress: collisions, scoring, respawn, spawning, rooms, packet handling, and pause/safety states.
- Avoid per-tick logs. Prefer logs around state transitions, warnings, and real failures.
- Be careful with Godot scene diffs. Godot may add `uid` or `unique_id` fields; do not remove unrelated scene changes casually.
- Do not revert user/editor changes unless explicitly asked.

## Where To Look First

For a server gameplay bug:

- `services/game-server/internal/game/game.go`
- `services/game-server/internal/game/combat.go`
- `services/game-server/internal/game/collisions.go`
- `services/game-server/internal/game/damage.go`
- `services/game-server/internal/game/match.go`
- `services/game-server/internal/game/motion/`
- `services/game-server/internal/game/rules/`
- `services/game-server/internal/game/session.go`
- `services/game-server/internal/game/spawn_types.go`
- `services/game-server/internal/game/spawning.go`
- `services/game-server/internal/game/spawning/`
- `services/game-server/internal/game/scoring/`
- `services/game-server/internal/game/scoring.go`
- `services/game-server/internal/game/entities/`
- relevant tests under `services/game-server/tests/game/`

For rooms/networking:

- `services/game-server/internal/rooms/`
- `services/game-server/internal/networking/websocket.go`
- `services/game-server/internal/networking/rooms.go`
- relevant tests under `services/game-server/tests/networking/`
- room-domain tests under `services/game-server/tests/rooms/`

For client runtime flow:

- `client/scripts/ui/game_shell.gd`
- `client/scripts/game.gd`
- `client/scripts/spectate_targets.gd`
- `client/scripts/networking/network_client.gd`
- `client/scripts/networking/world_sync.gd`
- `client/scripts/entities/player.gd`
- `client/scripts/ui/hud_controller.gd`
- relevant GUT tests under `client/tests/unit/`

For shared schema/code generation:

- `shared/game_data.toml`
- `shared/packets/packets.toml`
- `tools/data_sync/main.py`

For docs:

- [docs/design/architecture.md](design/architecture.md)
- [docs/api/nestjs-api-server.md](api/nestjs-api-server.md)
- [docs/server/logging.md](server/logging.md)
- [docs/devtools/toggles.md](devtools/toggles.md)
- [docs/notes.md](notes.md)

## Known TODOs / Gaps

TODO: add an actual pause/menu scene. The current pause support has packet/server/client state plumbing, but no dedicated menu overlay scene.

TODO: manually smoke test the pause flow in Godot with the server actually running. Recent work re-enabled client pause after a false alarm caused by testing without the server started.

TODO: revisit max window sizing. OS window pixels are not a reliable balance boundary across monitors/DPI; gameplay balance should use a logical/capped viewport size rather than raw window max size.

TODO: clean up generated/recorded artifacts if they are present in Git status, especially `*.avi` and server `tmp` binaries.

TODO: verify collision shape export covers the ship shape before relying on exported collision data after Godot upgrades.

Unknown: exact Godot command-line availability in every developer environment. The documented command is best-effort.
