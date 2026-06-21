# Agent Testing Rules
Parent index: [Agent](./!INDEX.md)

## Purpose

This document owns testing and verification guidance for agents.

## Overview

Use this document for test commands, verification checkpoints, generated data checks, and validation planning.

## Rules

- Focused, safe terminal checks are allowed when useful.
- Commands in this document are usually human-run checkpoints.
- Avoid destructive git commands, broad cleanup, dependency upgrades, unrelated formatter runs, or expensive commands unless explicitly requested.
- If a human-run command fails, stop and diagnose that failure before piling on more changes.
- Keep collision export guidance tied to verification and the data pipeline, not broad implementation documentation.

## Server checks

Run server tests:

```bash
cd services/game-server
go test -buildvcs=false ./...
```

The normal/default server test command exercises the devtools-enabled build.

Run server tests with devtools disabled:

```bash
cd services/game-server
go test -tags nodevtools -buildvcs=false ./...
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

Build the game server with devtools disabled:

```bash
cd services/game-server
go build -tags nodevtools -buildvcs=false -o ./tmp/game-server ./cmd/game-server
```

Run the server devtools boundary static check:

```bash
python3 -m pytest tools/tests/test_server_devtools_boundary.py
```

This check protects the server devtools boundary: `internal/devtools` owns behavior, and `internal/game` exposes only `export_devtools*.go` bridge files.

If the server test command prints read-only `envman` warnings but tests pass, those warnings have been harmless in this environment.

## Client checks

Open the Godot project by opening or importing:

```text
client/
```

The configured main scene is:

```text
res://scenes/game.tscn
```

Run client GUT tests, if the `godot` CLI is available:

```bash
godot --headless --path client -s res://addons/gut/gut_cmdln.gd -gdir=res://tests/unit -ginclude_subdirs -gexit
```

Run the client constants boundary scan:

```bash
python3 -m pytest tools/tests/test_client_constants_boundary.py
```

Full gameplay/network smoke testing remains manual for now: opening the game scene, websocket connection, asteroid spawning, shooting/effects, pause/debug flow, and the full gameplay loop.

## Data-sync checks

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

Check shared packets:

```bash
python3 tools/data_sync/main.py -check -packets -go -gds
```

Packet validate/diff/push/check commands operate on the split packet SoT under `shared/packets/` (`outputs.toml`, `gameplay.toml`, `debug.toml`, and `lobby.toml`). Packet generation/checks include server devtools packet output in `services/game-server/internal/devtools/packets_generated.go`.

Export pickup collision shapes with:

```bash
cd /mnt/d/!bin/space-rocks
godot --headless --path client -s res://tools/export_collision_shapes.gd
```

Pickup collision JSON should use class keys such as `powerup` and `weapon`, not per-type keys such as `1_up` or `torpedo`.

## Test layout

Go server tests live under:

```text
services/game-server/tests/<area>/
```

Current areas include:

- `game`
- `networking`
- `physics`
- `rooms`
- `scoring`
- `space`

Do not add new `*_test.go` files beside production packages under `services/game-server/internal/`.

For game simulation setup, use the shared harness in:

```text
services/game-server/tests/game/helpers_test.go
```

Keep new helpers intent-level, such as placing entities or sending packets, instead of exposing raw private maps.

Godot client tests use GUT and live under:

```text
client/tests/
```

Unit tests go under:

```text
client/tests/unit/
```

Fixtures go under:

```text
client/tests/fixtures/
```

Reusable test-only helpers go under:

```text
client/tests/helpers/
```

Keep client tests focused on:

- generated packets
- HUD behavior
- `world_sync`
- pure client logic

Do not put test helpers in `client/scripts/`.

## Related docs

- [Generated Files](./generated-files.md)
- [Repo Hygiene](./repo-hygiene.md)
- [Documentation Editing](./documentation-editing.md)

## Notes

Human-run checkpoint guidance stays here.

Prompt/report expectations live in [Prompting And Reporting](./prompting-and-reporting.md).
