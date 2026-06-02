# Developer Handoff

This document is a practical handoff for future development sessions. It focuses on how to work on the project. For broader architecture, see [docs/design/architecture.md](design/architecture.md). For server logging details, see [docs/server/logging.md](server/logging.md).

## Project Overview

Space Rocks is an Asteroids-inspired game with:

- a Godot client in `client/`
- a Go game server in `services/game-server/`
- a planned API server in `services/api-server/`
- shared data sources in `shared/`
- the active constants/packet sync tool in `tools/data_sync/`

The current direction is server-authoritative for gameplay state. The client collects input and renders/interpolates state; the server owns simulation outcomes such as movement, bullets, asteroid collisions, scoring, lives, death, respawn, room state, and pause safety rules.

The project is in active development. Expect rough edges and incomplete UI around newer systems.

## Prerequisites

Install these before running or developing Space Rocks locally:

- **Godot 4.6** for the client project.
  - Open/import the `client/` folder as the Godot project.
  - The configured main scene is `res://scenes/game.tscn`.

- **Go 1.26.3** for the real-time game server.
  - The Go module is in `services/game-server/`.
  - The server entrypoint is `services/game-server/cmd/game-server`.

- **Python 3.10+** for repo tooling and static checks.
  - Install the repo Python dependencies with `python -m pip install -r requirements-dev.txt`.
  - The data-sync tool uses modern Python typing syntax and requires `tomlkit`.
  - The client constants-boundary test uses `pytest`.

- **Git LFS** for binary/source asset files.
  - The repo tracks asset patterns such as PNG, WEBP, WAV, and MP3 through Git LFS.
  - After cloning, run:

```bash
git lfs install
git lfs pull
```

## Repository Structure

- `client/`: Godot project. Scenes, scripts, assets, audio, shaders, client tools, and generated client constants/packet helpers.
- `services/game-server/`: Go module for the real-time game server. Main entrypoint is `services/game-server/cmd/game-server/main.go`.
- `services/api-server/`: empty placeholder for a planned API server service for business/backend systems. Intended stack is Node.js/TypeScript with NestJS.
- `shared/`: source data used by both client and server:
  - `shared/constants/server_constants.toml`, `shared/constants/server_entities.toml`, `shared/constants/client/presentation.toml`, `shared/constants/client/shell.toml`, and `shared/constants/client/lobby.toml` for active constants
  - client constants use nested subcategory sections under `constants.client.presentation.*`, `constants.client.shell.*`, and `constants.client.lobby.*`
  - `shared/packets/outputs.toml`, `shared/packets/gameplay.toml`, `shared/packets/debug.toml`, and `shared/packets/lobby.toml` for active packets
  - debug/devtools packet schema lives in `shared/packets/debug.toml`
  - data-sync output id `server_devtools_packets` generates server devtools packet types into `services/game-server/internal/devtools/packets_generated.go`
  - `shared/collisions/collision_shapes.json`
- `tools/data_sync/`: Python sync/generation tool for constants and packet code.
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

1. Session controllers under `client/scripts/session/` coordinate boot, config, room/lobby, and gameplay session flows.
2. `client/scripts/shell/gameplay_shell_flow.gd` is the narrow gameplay coordinator. It stores references, delegates lane configuration, parses incoming gameplay state packets through `client/scripts/gameplay/state/`, delegates state application to `client/scripts/gameplay/runtime/`, delegates HUD/menu/input/respawn/spectate/events/effects to their owned gameplay seams, and emits outer lifecycle signals.
3. Local gameplay input is routed through `client/scripts/gameplay/input/`. Player movement/shooting packets still originate from `client/scripts/entities/player.gd`, but input polling/routing for pause/menu, respawn, spectate, and devtools is coordinated by the gameplay input seam.
4. `client/scripts/networking/network_client.gd` sends and receives websocket text and routes packet JSON through `client/scripts/networking/packet_codec/packet_codec.gd`.
5. `services/game-server/internal/networking/websocket_read.go` reads client packets and decodes packet JSON through `services/game-server/internal/protocol/packetcodec`.
6. Lobby/lifecycle packets call `services/game-server/internal/rooms` for domain decisions and networking sends snapshots/errors.
7. After a room reaches `InGame`, the room's `*game.Game` handles gameplay packets and advances simulation.
8. The server encodes state packets through `packetcodec` and sends them at the server tick rate.
9. `client/scripts/world/world_sync.gd` coordinates sync ordering and delegates node ownership, packet application, and interpolation to the player, bullet, asteroid, and local-visual sync owners under `client/scripts/world/`.
10. HUD, menu, respawn, spectate, event, death, and effects presentation updates flow through the focused gameplay seams under `client/scripts/gameplay/`.

## Packet Schema And Generated Outputs

Packet schema source of truth is split across:

- `shared/packets/outputs.toml`
- `shared/packets/gameplay.toml`
- `shared/packets/debug.toml`
- `shared/packets/lobby.toml`

Generated packet outputs:

- `services/game-server/internal/game/entities/packets_generated.go`
- `services/game-server/internal/game/packets.go`
- `services/game-server/internal/devtools/packets_generated.go`
- `client/scripts/networking/packets/packets.gd`

Edit the relevant split packet TOML file for schema/content changes. Edit `shared/packets/outputs.toml` only when changing output routing.

Packet schema drift rule:

- Generated packet files are not the source of truth.
- Packet struct fields belong in `shared/packets/*.toml`.
- Output-level Go imports belong in `shared/packets/outputs.toml`.
- The current generator does not support field-level `go_import` for generated Go output.
- Before adding packet fields, run `data-sync -check -packets -go`.
- If check fails, repair schema drift before adding new fields.
- After schema edits, run `data-sync -push -packets -go` and `go test ./...` from `services/game-server`.
- Example drift case: `StatePacket.PlayerWorldStates` existed in generated Go usage but was missing from `shared/packets/gameplay.toml`, so regeneration removed the field and broke handwritten Go code.

Devtools packet boundary rules:

- devtools packet schema lives in `shared/packets/debug.toml`
- server devtools packet output lives in `services/game-server/internal/devtools/packets_generated.go`
- targeted devtools UI packets use `target_player_id` where applicable
- when adding a new generated GDS packet helper, also add its builder mapping in `shared/packets/outputs.toml`
- regenerate Go and GDS packet outputs together when shared packet schema changes
- TS output is currently disabled; do not include TS flags in normal data-sync commands
- packet schema changes normally require editing the relevant `shared/packets/*.toml` source, editing `shared/packets/outputs.toml` when adding generated output routing, then running `data-sync -push -packets -go -gds`
- packet pull remains unsupported; edit shared packet TOML and push generated outputs
- client readers should not depend on generated game packet constants for devtools-only wrapper fields such as `debug_status`

## Devtool Hotkeys

For a focused server devtools reference (commands, boundaries, and checks), see [docs/server/devtools.md](server/devtools.md).
For semantic mouse input behavior, see [docs/client/mouse-input.md](client/mouse-input.md).
For targeting ownership and boundaries, see [docs/server/targeting.md](server/targeting.md).
For devtools telemetry readouts and boundaries, see [docs/devtools/telemetry.md](devtools/telemetry.md).

Canonical gameplay devtool hotkeys:

- `0`: window
- `1`: invincible (self-targeting hotkey)
- `2`: infinite lives (self-targeting hotkey)
- `3`: world freeze
- `4`: player freeze (self-targeting hotkey)
- `5`: kill local player
- `6`: spawn new player
- `7`: force respawn local player
- `8`: reserved
- `9`: reserved

Devtools window targeting notes:

- canonical gameplay target and per-tool devtools target are separate concepts (see [docs/server/targeting.md](server/targeting.md))
- player-only commands use `target_player_id` only after resolver compatibility checks
- invincibility, infinite lives, player freeze, kill, respawn, and score/lives controls can target selected players where wired
- score/lives controls use active-player target dropdowns
- score/lives target dropdown labels are player IDs only (no ALIVE/DEAD or Active/Inactive status text)
- world freeze remains room-wide/global

Devtools command behavior notes:

- Set Score sets the exact authoritative score
- Add Score accepts positive or negative amounts and clamps final score at zero minimum
- Set Lives sets the exact authoritative lives
- Add Lives accepts positive or negative amounts and clamps final lives at zero minimum
- Clear Bullets removes authoritative bullets through normal world sync
- Clear Asteroids removes authoritative asteroids through normal world sync and does not award score or spawn fragments

Client devtools authority note:

- the client devtools UI sends packets only; it does not mutate HUD, score/lives, bullets, asteroids, or `world_sync` locally

Devtools telemetry handoff note:

- raw `LocalPlayerTelemetry` and `TargetTelemetry` readouts live in the devtools window (see [docs/devtools/telemetry.md](devtools/telemetry.md))
- a future world telemetry overlay is separate from HUD and is not implemented yet

Input/targeting handoff note:

- mouse actions in gameplay/devtools flows use semantic InputMap actions (`SelectTarget`, `DeselectTarget`, `SpawnEntity`, `CancelAction`)
- raw left/right mouse buttons should remain only in InputMap bindings (`project.godot`)
- targeting and persistent debug bullet stream cadence stay server-owned

## Run The Server

From the repo root:

```bash
cd services/game-server
go run ./cmd/game-server
```

Normal local server runs include devtools.

To run with server devtools disabled:

```bash
cd services/game-server
go run -tags nodevtools ./cmd/game-server
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

Build the server with devtools disabled:

```bash
cd services/game-server
go build -tags nodevtools -buildvcs=false -o ./tmp/game-server ./cmd/game-server
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

Check generated constants/packets against shared sources:

```bash
data-sync -check -packets -go -gds
data-sync -check -constants -go -gds
```

Show pending generated diffs before pushing:

```bash
data-sync -diff -packets -go -gds
data-sync -diff -constants -go -gds
```

Push regenerated outputs:

```bash
data-sync -push -packets -go -gds
data-sync -push -constants -go -gds
```

Regenerate collision shapes from Godot scenes:

```bash
godot --headless --path client -s res://tools/export_collision_shapes.gd
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
- `services/game-server/tests/protocol/`
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

Keep test-only helpers out of `client/scripts/`. Client tests should focus on generated packets, packet/state reader safety, HUD/menu behavior, `world_sync`, constants-boundary assumptions, and pure client logic. Do not turn these into full gameplay/network integration tests.

Run the GUT suite with:

```bash
godot --headless --path client -s res://addons/gut/gut_cmdln.gd -gdir=res://tests/unit -ginclude_subdirs -gexit
```

A passing run may still report Godot ObjectDB/resource cleanup warnings; treat the suite result as passing when GUT reports all tests passed.

Expected missing-field warnings may appear in tests that intentionally verify safe behavior for missing `lives`, `respawn_delay`, or asteroid `scale`; those warnings are fine when the suite passes.
