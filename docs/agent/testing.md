# Agent Testing Rules
Parent index: [Agent](./!INDEX.md)

## Purpose

This document owns testing and verification guidance for agents.

## Overview

Use this document for test commands, verification checkpoints, generated data checks, and validation planning.

GitHub Actions runs three independent jobs: repository checks, Go tests, and Godot/GUT client tests. The jobs use read-only repository permissions, bounded job timeouts, and cancellation of superseded runs. The repository and client jobs use Python tooling where needed; the Go job reads its Go version from `services/game-server/go.mod`.

## Rules

- Focused, safe terminal checks are allowed when useful.
- Commands in this document are usually human-run checkpoints.
- Avoid destructive git commands, broad cleanup, dependency upgrades, unrelated formatter runs, or expensive commands unless explicitly requested.
- If a human-run command fails, stop and diagnose that failure before piling on more changes.
- Keep collision export guidance tied to verification and the data pipeline, not broad implementation documentation.

## Server checks

The shared Go CI runner is:

```bash
bash tools/ci/run_go_tests.sh
```

It executes all three stages, even if an earlier stage fails:

- `player-data`: `go test -buildvcs=false ./...`
- `game-server default`: `go test -buildvcs=false ./...`
- `game-server nodevtools`: `go test -tags nodevtools -buildvcs=false ./...`

The runner reports each stage and returns nonzero if any stage fails.

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

This check protects the server devtools boundary: `internal/devtools` owns controller, command, targeting, stream-runtime, and debug DTO behavior; `internal/game` exposes authoritative capabilities through `Control`; package dependencies remain one-way; and legacy game-side debug adapters remain absent.

A separate focused test guards canonical devtools documentation against removed architecture names and paths.

If the server test command prints read-only `envman` warnings but tests pass, those warnings have been harmless in this environment.

For focused server logger verification, check `services/game-server/internal/logging` and run the server with info-level server logs enabled:

```bash
cd services/game-server
LOG_SERVER=info go run ./cmd/game-server
```

Then inspect the sequential JSONL output under `services/game-server/logs/game-server/game-server-*.jsonl` to confirm file output is being written.


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


For focused client logger verification, use `client/tests/unit/test_client_logger.gd`. It uses `user://logger_test_output` with the `client-test` prefix and cleans up its own JSONL files, so it is a good place to confirm file-output and formatting behavior without broadening into a full client testing catalog.

Run the configuration-driven architecture guard:

```bash
python tools/architecture_guard/main.py
```

Selected scene-backed client integration tests run headlessly and without a server. They cover full game-scene boot/reset propagation and the weapon cooldown visual transition lifecycle.

The guard reads `tools/architecture_guard/rules.toml`; repository-specific invariants remain narrow TOML rules rather than Python logic. Current high-confidence boundaries prevent client scripts from directly reading selected server-owned player, respawn, and asteroid constants, and prevent non-owner client scripts from reaching through `runtime_context.world_sync`. Keep the full rule definitions and exclusions in the TOML file rather than duplicating them here.

Shared local/CI runners are available from the repository root:

```bash
bash tools/ci/run_repo_checks.sh
bash tools/ci/run_go_tests.sh
bash tools/ci/run_client_tests.sh
```

The client runner uses Godot 4.6.3 in CI, checks out LFS assets, performs a 1200-frame first-run editor bootstrap under Xvfb with the compatibility renderer and OpenGL 3 software driver, imports headlessly, and runs the existing GUT selection under `tests/unit`. Each stage has a timeout, streams stdout to the console and an artifact directory, and records a Godot engine log; CI uploads that directory unconditionally. Local runs use Xvfb when available and retain a headless bootstrap fallback. Bootstrap, import, and GUT timeouts are configurable with `GODOT_BOOTSTRAP_TIMEOUT`, `GODOT_IMPORT_TIMEOUT`, and `GODOT_GUT_TIMEOUT`; `GODOT_ARTIFACT_DIR` stores streamed stage output and Godot engine logs.

The remaining gameplay/network smoke boundary is manual: websocket connection, asteroid spawning, actual shooting/effects, pause/debug flow, and the full gameplay loop.

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

## Test design guidance

Keep tests aligned with the seam whose behavior they claim to verify:

- Mutation-semantics tests should call direct mutation methods and assert their returned change (`Found` and resulting value), rather than expecting a presentation snapshot to update immediately.
- Presentation tests should establish state through a publishing `game.Control` seam, or run an explicit `Game.Step` (including a zero-delta step when appropriate) before reading a presentation snapshot.
- When respawn placement policy is the subject, call `Control.SafeRespawnPosition` directly and assert its success boolean before checking the position. Use packet/request plus snapshot coverage when testing publication or lifecycle behavior.
- Do not parameterize a known-supported implementation only to skip it at runtime. Avoidable collected skips are not acceptable; the repository gate is expected to complete with zero skips.

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
