# Agent Testing Rules

Use this when changing tests, verification commands, generated data, packet/schema code, or anything that needs a validation report.

## Server Commands

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

If the server test command prints read-only `envman` warnings but tests pass, those warnings have been harmless in this environment.

## Client / Godot Commands

Open the Godot project by opening/importing:

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

## Data Sync Commands

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

## Server Test Layout

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

For server gameplay changes, run:

```bash
cd services/game-server
env GOCACHE=/tmp/space-rocks-go-build go test -buildvcs=false ./...
```

## Client Test Layout

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

For client changes, run GUT when the `godot` CLI is available:

```bash
godot --headless --path client -s res://addons/gut/gut_cmdln.gd -gdir=res://tests/unit -ginclude_subdirs -gexit
```

## Reporting Expectations

After each implementation prompt, report:

- changed files
- exact validation command
- pass/fail result
- relevant failure output, if any
- `git status --short`

If tests fail, stop and report the failure. Do not continue piling changes onto a failing state unless the prompt explicitly asks for a focused fix.

Read-only prompts must not edit files, run formatters, or perform cleanup.
