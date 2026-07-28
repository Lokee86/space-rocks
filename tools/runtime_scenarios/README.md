# Runtime Scenarios

This tool runs deterministic, real client/server pressure scenarios against the normal Space Rocks room, WebRTC, gameplay, devtools, and measurement paths.

## Validate a scenario

```bash
python tools/runtime_scenarios/main.py \
  tools/runtime_scenarios/scenarios/network_interest_lifecycle_v1.json \
  --validate-only
```

## Run the lifecycle regression scenario

```bash
python tools/runtime_scenarios/main.py \
  tools/runtime_scenarios/scenarios/network_interest_lifecycle_v1.json
```

## Receiver-scaling matrix

These scenarios hold the room at eight total participants and use the same seed, phase durations, and bullet-stream pressure. Only the number of real network clients changes:

```text
receiver_scale_1c_7b_v1  1 real client + 7 bots
receiver_scale_2c_6b_v1  2 real clients + 6 bots
receiver_scale_4c_4b_v1  4 real clients + 4 bots
```

Run each scenario independently so every run receives its own server process, loopback port, logs, and measurement reports.

After the runs complete, summarize them together:

```bash
python tools/runtime_scenarios/matrix_summary.py \
  .ci-artifacts/runtime-scenarios/<1-client-run> \
  .ci-artifacts/runtime-scenarios/<2-client-run> \
  .ci-artifacts/runtime-scenarios/<4-client-run> \
  --output .ci-artifacts/runtime-scenarios/receiver-scale-matrix.json
```

The summary aggregates total per-receiver server output across all real clients while using one server report as the shared simulation/process reference. Server measurement report version 6 includes receiver candidate-build, encoding, and outbound timing; six top-level candidate-build subphases; nested lane-candidate timing for state advancement, world/hot/lifecycle, locator, overlay, session, event, and finalization work; the complete breakdown from the exact peak candidate-build tick; skipped-send ticks; and per-lane current/peak buffered bytes and skipped-send counts. Headless client frame values measure client processing and host contention, not render performance.

The game server caches one immutable quantized full-world projection per published presentation generation and reuses it across receivers. Receiver-specific interest filtering, baselines, sequences, cadence, hot-route state, lifecycle selection, and deltas remain independent.

Use `--godot <path>` or set `SPACE_ROCKS_GODOT_EXECUTABLE` when Godot is not on `PATH`. The known Windows editor executable is `C:\Godot.exe` (Godot 4.6.3). The path must identify a Godot editor binary; exported Space Rocks executables are rejected.

The coordinator is visible by default so its client report includes real rendering pressure. Use `--headless-coordinator` only for unattended orchestration/network verification; that run is not a render-performance baseline.

Scenario clients always run the source project through the Godot editor executable. Runtime-scenario startup explicitly skips the bundled local-alpha server, so no packaged or exported Space Rocks executable is started. On Windows, the server runs through WSL with `go run ./cmd/game-server`; this launches a Linux process instead of a temporary Windows `game-server.exe`. Non-Windows hosts use direct `go run`.

The runner:

1. Reserves a free loopback port for the run.
2. Starts the current game-server source with WSL `go run ./cmd/game-server` on Windows, or direct `go run` on non-Windows hosts, using harness-only deterministic seed, authentication, output, and listen settings.
3. Starts the source project through one editor-based coordinator and the configured editor-based headless participants.
4. Waits for client status files and combined measurement exports.
5. Stops only the processes it started.
6. Writes logs, configured `phase-markers.json`, and `summary.json` under `.ci-artifacts/runtime-scenarios/`.

Do not use native Windows `go run`, `go build -o`, a packaged executable, or another direct Windows server launch in this harness. Native Windows `go run` still creates a temporary `game-server.exe`, which triggers firewall/UAC approval and breaks unattended remote testing. The Windows harness must keep the server inside WSL.

The harness-only server settings are process environment variables set by this runner. Normal server startup does not enable them, and clients cannot select the simulation seed.
