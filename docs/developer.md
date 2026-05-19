# Developer Handoff

This document is a practical handoff for future development sessions. It focuses on how to work on the project. For broader architecture, see [docs/design/architecture.md](design/architecture.md). For server logging details, see [docs/server/logging.md](server/logging.md).

## Project Overview

Space Rocks is an Asteroids-inspired game with:

- a Godot client in `client/`
- a Go game server in `server/`
- shared JSON sources in `shared/`
- Python generation scripts in `tools/scripts/`

The current direction is server-authoritative for gameplay state. The client collects input and renders/interpolates state; the server owns simulation outcomes such as movement, bullets, asteroid collisions, scoring, lives, death, respawn, room state, and pause safety rules.

The project is in active development. Expect rough edges and incomplete UI around newer systems.

## Repository Structure

- `client/`: Godot project. Scenes, scripts, assets, audio, shaders, client tools, and generated client constants/packet helpers.
- `server/`: Go module for the game server. Main entrypoint is `server/cmd/game-server/main.go`.
- `shared/`: JSON source data used by both client and server:
  - `shared/constants/constants.json`
  - `shared/packets/packets.json`
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

Multiplayer rooms use an optional query parameter:

```text
ws://localhost:8080/ws?room_id=ROOM_ID
```

Blank room IDs map to the default room. Non-blank room IDs create or join separate server rooms.

Runtime flow:

1. Godot collects input in `client/scripts/player.gd`.
2. `client/scripts/game.gd` sends input/config/respawn/pause packets through `client/scripts/network_client.gd`.
3. `server/internal/networking/websocket.go` reads client packets.
4. The room's `*game.Game` handles packets and advances simulation.
5. The server sends state packets at the server tick rate.
6. `client/scripts/world_sync.gd` applies/interpolates state to Godot nodes.
7. HUD and effects update from state/events.

## Run The Server

From the repo root:

```bash
cd server
go run ./cmd/game-server
```

With Air hot reload, if installed:

```bash
cd server
air
```

The Air config is `server/.air.toml`. It builds:

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
cd server
go run ./cmd/game-server
```

Run all server tests:

```bash
cd server
go test -buildvcs=false ./...
```

Use an explicit cache path when the shell environment has cache or permission issues:

```bash
cd server
env GOCACHE=/tmp/space-rocks-go-build go test -buildvcs=false ./...
```

Build the server binary:

```bash
cd server
go build -buildvcs=false -o ./tmp/game-server ./cmd/game-server
```

Regenerate shared constants:

```bash
python3 tools/scripts/generate_constants.py
```

Regenerate shared packets:

```bash
python3 tools/scripts/generate_packets.py
```

## Generated Files

Do not hand-edit generated files unless there is a specific reason and the source/generator will be updated later.

Constants source:

```text
shared/constants/constants.json
```

Generated outputs:

```text
client/scripts/constants.gd
server/internal/constants/constants.go
```

Packet source:

```text
shared/packets/packets.json
```

Generated outputs:

```text
client/scripts/packets.gd
server/internal/game/packets.go
server/internal/game/entities/packets_generated.go
```

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
- Go commands can run from WSL inside `server/`.
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

Known current cleanup note: generated recordings/build artifacts such as `client/game-clip.avi`, `server/tmp/game-server`, and old `server/cmd/.../tmp` outputs should not be committed.

## Logging And Config

The server has a custom structured logging package:

```text
server/internal/logging/logger.go
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

## Important Conventions

- Keep authoritative gameplay rules on the server.
- Keep rendering, local audio/effects, UI, and interpolation in the Godot client.
- Keep websocket/room transport in `server/internal/networking`.
- Keep game rules in `server/internal/game`.
- Keep reusable simulation code out of `cmd/game-server/main.go`.
- Use `shared/` JSON plus generators when Go and Godot must agree on constants or packet structures.
- Add focused Go tests for server gameplay rules that can regress: collisions, scoring, respawn, spawning, rooms, packet handling, and pause/safety states.
- Avoid per-tick logs. Prefer logs around state transitions, warnings, and real failures.
- Be careful with Godot scene diffs. Godot may add `uid` or `unique_id` fields; do not remove unrelated scene changes casually.
- Do not revert user/editor changes unless explicitly asked.

## Where To Look First

For a server gameplay bug:

- `server/internal/game/game.go`
- `server/internal/game/combat.go`
- `server/internal/game/session.go`
- `server/internal/game/spawning.go`
- `server/internal/game/scoring.go`
- `server/internal/game/entities/`
- relevant tests under `server/internal/game/*_test.go`

For rooms/networking:

- `server/internal/networking/websocket.go`
- `server/internal/networking/rooms.go`
- `server/internal/networking/rooms_test.go`

For client runtime flow:

- `client/scripts/ui/game_shell.gd`
- `client/scripts/game.gd`
- `client/scripts/network_client.gd`
- `client/scripts/world_sync.gd`
- `client/scripts/player.gd`
- `client/scripts/ui/hud_controller.gd`

For shared schema/code generation:

- `shared/constants/constants.json`
- `shared/packets/packets.json`
- `tools/scripts/generate_constants.py`
- `tools/scripts/generate_packets.py`

For docs:

- [docs/design/architecture.md](design/architecture.md)
- [docs/server/logging.md](server/logging.md)
- [docs/devtools/toggles.md](devtools/toggles.md)
- [docs/notes.md](notes.md)

## Known TODOs / Gaps

TODO: add an actual pause/menu scene. The current pause support has packet/server/client state plumbing, but no dedicated menu overlay scene.

TODO: manually smoke test the pause flow in Godot with the server actually running. Recent work re-enabled client pause after a false alarm caused by testing without the server started.

TODO: revisit max window sizing. OS window pixels are not a reliable balance boundary across monitors/DPI; gameplay balance should use a logical/capped viewport size rather than raw window max size.

TODO: clean up generated/recorded artifacts if they are present in Git status, especially `*.avi` and server `tmp` binaries.

TODO: decide whether `client/scripts/constants.gd` should move into a generated/constants folder. This was considered low effort but has not been done.

TODO: verify collision shape export covers the ship shape before relying on exported collision data after Godot upgrades.

Unknown: exact Godot command-line availability in every developer environment. The documented command is best-effort.
