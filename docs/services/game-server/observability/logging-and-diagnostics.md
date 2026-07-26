---
author: brian
created: "2026-07-19"
document_id: 019f7d55-fb2c-70ee-8da0-4dfdfa6966f6
document_type: general
policy_exempt: false
summary: This document describes the game-server logging and diagnostics boundary. It explains how game-server owners emit canonical structured runtime events, how process-level logging configuration is retained, which diagnostic events belong...
---
# Logging And Diagnostics

Parent index: [Game Server Observability](./!INDEX.md)

## Purpose

This document describes the game-server logging and diagnostics boundary. It explains how game-server owners emit canonical structured runtime events, how process-level logging configuration is retained, which diagnostic events belong in logs, and what this boundary does not own.

## Overview

Game-server logging is a small internal access point for the shared canonical observability emitter. Production call sites submit `observability.Request` values directly through `logging.Emit`; they do not use category logger methods or scatter raw `slog`, `log`, or `fmt.Println` calls through service code.

The current implementation uses the shared `shared/go/servicelog` runtime for generic console/file fanout and active-file lifecycle. Generated observability definitions own canonical event categories and default levels. The game-server adapter owns the process-level configuration entry point, environment variable, and service-specific byte limits.

Current runtime flow:

```text
game-server runtime owner -> logging.Emit(observability.Request{...}) -> shared observability validation/redaction -> shared servicelog sink -> stderr + active JSONL file
```

Logging is observational. It should explain what happened in the process, connection, room, or simulation, but it must not change gameplay state, packet routing, persistence, or auth behavior.

## Code root

```text
services/game-server/
```

## Responsibilities

Logging and diagnostics own the game-server side of:

- Structured canonical event emission through the internal logging package.
- Environment-based log-level configuration.
- Shared log field names for common diagnostic dimensions.
- Runtime diagnostics for recoverable errors, lifecycle events, and unusual conditions.
- WebSocket close classification for expected versus unexpected read/write failures.
- Keeping logs useful without enabling per-tick or per-entity output by default.
- Passing file-policy values through to shared servicelog for structured JSONL output.

The game-server adapter owns the process-level configuration entry point, environment variable, and service-specific byte limits used for file-output policy. Canonical event names, categories, default levels, trace requirements, and retention tiers come from generated observability contract data.

The shared `shared/go/observabilityevent` package owns contract validation, redaction, canonical serialization, and stable rejection outcomes. The shared `shared/go/servicelog` package owns direct canonical-record console/file fanout and active-file lifecycle without reshaping records through `slog`.

## Does not own

Logging and diagnostics does not own:

- Process startup or shutdown behavior.
- HTTP route composition.
- WebSocket transport lifecycle.
- Packet schema or codec ownership.
- Room membership rules.
- Match lifecycle rules.
- Gameplay simulation authority.
- Player-data persistence.
- Auth token verification.
- Devtools command behavior.
- Client-side log presentation.
- Durable telemetry storage.
- Metrics, tracing, or log aggregation infrastructure.

Those systems may emit logs, but they own their own runtime behavior.

## Domain roles

Game-server logging participates in technical diagnostics only. It supports development and runtime investigation across:

- server process initialization
- auth verifier initialization
- player-data runtime initialization
- WebSocket upgrade/read/write behavior
- inbound packet decode failures
- outbound packet encode/write diagnostics
- room creation, cleanup, match-over, and match-result reporting
- player lifecycle, respawn, scoring, pause, and game-over events
- devtools command effects routed through real gameplay seams

It is not a product-facing telemetry system and does not provide durable analytics.

## Protocols and APIs

The logging surface is an internal Go package plus environment variables. It is not a client protocol and is not exposed over HTTP or WebSocket.

Server code consumes the package by importing:

```go
import "github.com/Lokee86/space-rocks/services/game-server/internal/logging"
```

The production call-site API is:

```go
logging.Emit(observability.Request{
    Event: observability.EventNameGameServerWriteFailed,
    Context: observability.Context{TraceID: traceID, SessionID: sessionID},
    Fields: observability.Fields{"failure_mode": "websocket_write_failed"},
})
```

Category logger objects and package-level text logging helpers were removed from the game-server package after the production call-site rollout. `logging.Emit` is the only game-server event emission API. The shared bridge remains available to other services until their compatibility paths are migrated.

### File output behavior

Game-server startup enables shared structured file output with:

```go
ConfigureFileOutput("logs/game-server", "game-server")
```

When the server runs from `services/game-server`, the active file path is `logs/game-server/game-server.jsonl.open`.

At runtime, the canonical emitter and shared sink fan records to:

- a safe human-readable rendering on stderr
- the canonical JSON envelope in the active JSONL file

Game-server canonical emission uses the generated service identity, event metadata, trace requirements, limits, rejection codes, and redaction policy. It does not emit the shared bridge-only `log_message` event.

Current logging package file-output helpers are:

```go
ConfigureFileOutput(baseDir, prefix) (string, error)
CloseFileOutput() error
```

Behavior notes:

- `ConfigureFileOutput` opens the active file path for the requested base directory and returns that active path on success.
- `CloseFileOutput` closes the active file output during shutdown or cleanup.
- File-output setup failure does not stop server startup.
- File-policy values are passed through to shared servicelog.
- The shared runtime enforces size and age rotation, completed-segment compression, interrupted-segment recovery, age and byte retention, degraded-state tracking, and clean closure.

### Environment configuration

Logging is configured during game-server startup:

```go
logging.Configure(os.Getenv(logging.EnvGlobalLevel))
```

The global environment variable is:

```text
LOG_LEVEL
```

No game-server category overrides remain. Canonical event categories and default levels come from generated observability contract data.

Supported level values are:

```text
debug
info
warn
warning
error
off
```

Current parsing behavior:

- Empty level values resolve to `warn`.
- `warning` is treated as `warn`.
- `off` maps to a level above `error` in the retained configuration state.
- Any non-empty unrecognized value currently resolves to `info`.

Default behavior is quiet:

```text
LOG_LEVEL unset -> warn
```

Canonical event emission is not filtered through this setting; generated event definitions supply the emitted level and category.

### Example configuration

Default warnings and errors only:

```bash
cd services/game-server
BUILD_VERSION=dev ENVIRONMENT=development go run ./cmd/game-server
```

Set the retained process-level configuration value:

```bash
cd services/game-server
BUILD_VERSION=dev ENVIRONMENT=development LOG_LEVEL=info go run ./cmd/game-server
```

Startup file-output status is emitted through the canonical observability flow:

- startup event: `service_starting` with `reason_code = "process_start"`
- file-output success adds no duplicate event
- failure event: `observability_unavailable` with a stable `failure_mode` such as `logging_file_open_failed`, `logging_runtime_degraded`, or `logging_file_close_failed`
- file paths and raw errors are not emitted as canonical fields

## Diagnostic field rules

Canonical context fields use generated names such as `trace_id`, `session_id`, `room_id`, `player_id`, `match_id`, `packet_type`, and `duration_ms`. Event-specific scalar fields use short snake_case names such as:

```text
reason_code
failure_mode
error_code
cleanup_version
active_players
remaining_members
packet_size
write_duration_ms
player_count
mode
```

Do not log secrets or credentials.

Never log:

- bearer tokens
- internal service tokens
- Discord access tokens
- OAuth codes
- raw OAuth state
- client secrets
- raw auth headers
- raw packet payloads containing sensitive data

## Canonical error and lifecycle rules

Emit the generated canonical event for the owning subsystem. The event definition supplies the level and category; the call site supplies the trace and approved context:

```go
logging.Emit(observability.Request{
    Event: observability.EventNameGameServerWriteFailed,
    Context: observability.Context{
        TraceID: traceID, SessionID: sessionID, RoomID: roomID, PlayerID: playerID,
    },
    Fields: observability.Fields{"failure_mode": "websocket_write_failed"},
})
```

`reason_code`, `failure_mode`, and `error_code` are stable, bounded classifications such as `game_over`, `outbound_queue_full`, or `websocket_write_failed`. They are contract fields, not places for arbitrary error prose. Raw errors, payloads, tokens, and secrets must not be copied into canonical fields or messages. The emitter validates UUID-shaped trace IDs for events that require them, applies redaction and field limits, and reports stable rejection codes without exposing rejected values.

For normal lifecycle events, use the generated event name and preserve the owning flow's trace. Do not add a second log at an adjacent layer for the same state transition.

### Trace ownership

Trace ownership follows the flow that creates the event:

- The process composition root creates startup and shutdown trace IDs and passes them through dependency and host lifecycle events.
- A WebSocket session owns its connection trace ID and preserves it across transport, packet, and room-boundary diagnostics unless an inbound request already carries the authoritative flow trace.
- A room owns the match trace ID used for game-over and match-result reporting.
- Gameplay owns the current match trace and player context for simulation lifecycle events.
- Devtools command outcomes preserve the command trace when present; the session connection trace remains the fallback boundary trace.

Events marked `trace_required` in generated contract data must receive a valid UUID trace. Events without that requirement still use the owning lifecycle trace when one exists.

## Data ownership

Logging and diagnostics does not own durable data.

It emits transient process output to stderr and, when startup file output is available, to the active JSONL file.

It may include identifiers and runtime facts that help correlate behavior:

- room ID
- player ID
- session ID
- remote address
- packet type
- match ID
- packet size
- write duration
- cleanup version
- player count
- active member counts

It must not persist logs, mutate player data, or become the source of truth for gameplay or account state.

If Space Rocks later adds durable telemetry, tracing, metrics, log shipping, or centralized observability, that should be documented as a separate integration or observability system.

## Diagnostic policy

Good log events are state transitions, failures, and unusual conditions.

Use logs for:

- process initialization failures
- service dependency initialization failures
- WebSocket upgrade/read/write failures
- expected versus unexpected socket closes
- malformed packet input
- packet encode or marshal failures
- room lifecycle transitions
- match result reporting lifecycle
- player spawn, respawn, death, and game-over transitions
- devtools effects applied through real gameplay seams
- non-empty realtime packet write summaries when network debug logging is intentionally enabled

Avoid logs for:

- every simulation tick
- every player position update
- every physics step
- every collision candidate
- every successful input packet
- every asteroid spawn candidate
- broad packet dumps
- duplicate logs for the same event at multiple layers

Logs should make production and development failures easier to diagnose without drowning normal gameplay output.

## Current realtime packet debug logs

Realtime packet debug output is currently focused on written gameplay packets only, and only when the owning realtime runtime enables it.

Current active debug messages:

- `lane protocol gameplay wire packet written` records each written gameplay wire packet.
- `lane protocol gameplay written` records a non-empty per-tick gameplay write summary.

`lane protocol gameplay wire packet written` is debug-only and useful for per-packet inspection. Useful fields include:

```text
wire_type
candidate_lane
candidate_kind
wire_lane
sequence
baseline_id
snapshot_id
snapshot_kind
encoded_bytes
```

`lane protocol gameplay written` is debug-only and emitted only when at least one packet was written for the tick. Useful fields include:

```text
lane_packet_families
baseline_full_count
event_batch_written
event_batch_drained_count
packet_count
encoded_bytes
```

Current behavior notes:

- no-op per-tick summaries are suppressed
- `realtime lane metric` is not current runtime output
- scheduler, budget, deferred, superseded, and CRUD-count fields are intentionally not emitted right now
- current debug output is diagnostic only; it does not itself implement packet-budget policy
- Candidate-level send-plan selection and chunker-owned hot-lane hard-size guarding live in protocol/realtime.
- Current debug output is diagnostic only; scheduler and active encoding should not reject already-chunked hot movement packets for size.
- Record/entity-level prioritization, cross-tick replay, and supersession guarantees remain future or non-current behavior.

Under extreme hot-lane stress, `lane protocol gameplay wire packet written` can produce one debug log per emitted hot-lane chunk. That can be hundreds of network debug records per second. This is expected diagnostic volume when realtime debug output is enabled; it is not normal quiet-mode output and does not define packet-budget policy.

## Code map

Primary implementation files:

```text
services/game-server/internal/logging/logger.go
shared/go/servicelog/
services/game-server/cmd/game-server/main.go
services/game-server/cmd/game-server/auth_config.go
```

Primary logging call-site areas:

```text
services/game-server/internal/networking/
services/game-server/internal/networking/inbound/
services/game-server/internal/networking/outbound/
services/game-server/internal/rooms/
services/game-server/internal/game/
services/game-server/internal/devtools/
```

Representative diagnostic helpers:

```text
services/game-server/internal/networking/websocket_close_logging.go
services/game-server/internal/networking/websocket_read.go
services/game-server/internal/networking/websocket_write.go
services/game-server/internal/rooms/lifecycle_tick.go
```

Important non-ownership boundaries:

```text
services/game-server/internal/protocol/packetcodec/
services/game-server/internal/authclient/
services/game-server/internal/playerdata/
services/player-data/
services/api-server/
client/
```

`packetcodec` owns packet encoding and decoding, not logging policy.

`authclient` owns auth verification requests, not logging output rules.

`playerdata` and `services/player-data/` own player-data contracts and storage behavior, not game-server log configuration.

`services/api-server/` owns Rails logging separately.

`client/` owns client-side logging and diagnostic presentation separately.

## Tests

The game-server observability rollout has dedicated JSONL-backed tests for the canonical emitter and each migrated ownership boundary:

```text
services/game-server/cmd/game-server/observability_test.go
services/game-server/internal/networking/observability_test.go
services/game-server/internal/networking/outbound_observability_test.go
services/game-server/internal/networking/room_observability_test.go
services/game-server/internal/rooms/observability_test.go
services/game-server/internal/rooms/cleanup_observability_test.go
services/game-server/internal/game/observability_test.go
services/game-server/internal/matchreporting/observability_test.go
```

These tests inspect representative JSONL records for startup/degraded logging, packet and WebSocket failures, room and match lifecycle, player lifecycle, devtools outcomes, match-result reporting, outbound encode/write failures, trace ownership, stable reason/failure codes, duplicate suppression, and raw-error/payload safety. The normal logging package tests remain in `services/game-server/internal/logging/logger_test.go`.

### Final integrated verification record

The completed game-server rollout gate was run from the repository root:

- `bash tools/ci/run_go_tests.sh` passed shared servicelog, player-data, diagnostic-aggregator, game-server default, and game-server `nodevtools` stages.
- `services/game-server/internal/logging` tests passed after converting sink/file assertions to canonical event records.
- `go vet -buildvcs=false ./...` and `go vet -tags nodevtools -buildvcs=false ./...` passed in `services/game-server`.
- `gofmt -l` was clean for the cleanup files `internal/logging/logger.go` and `internal/logging/logger_test.go`; the repository-wide scan still reports unrelated existing drift, including generated outputs.
- `PYTHONPATH=. python -m pytest tests tools/tests tools/data_sync/tests` passed, including architecture and observability emission guards.
- `pitlord check --repo . --policy tools/pitlord/policy.json` passed.
- `python tools/data_sync/main.py -validate` passed.
- Constants, packets, realtime-wire, and drop-table generated-data `-check` gates passed.
- `git diff --check` passed.
- Production scans confirm game-server semantic production emission uses the canonical `logging.Emit` boundary.

The repository wrapper `bash tools/ci/run_repo_checks.sh` requires `PYTHONPATH=.` for the local Windows Python module-import environment; the equivalent stages were rerun with that environment and passed. No rollout-related verification failure remained.

Representative canonical envelopes confirm stable event names, valid owning-flow traces, required identifiers, stable reason/failure classifications, one emission per tested transition, and no raw error or unsafe payload fields. The game-server-local compatibility logger API is now absent; the shared bridge remains a separate cross-service cleanup boundary.

### Configuration smoke check

The retained `LOG_LEVEL` configuration entry point can be exercised without changing canonical event selection:

```bash
cd services/game-server
BUILD_VERSION=dev ENVIRONMENT=development LOG_LEVEL=info go run ./cmd/game-server
```

```bash
cd services/game-server
BUILD_VERSION=dev ENVIRONMENT=development LOG_LEVEL=warn go run ./cmd/game-server
```

Expected stderr output is the safe canonical rendering with generated category and level context. Direct canonical calls write exactly one canonical JSON envelope to the active JSONL file.

## Related docs

- [Game Server Observability](./!INDEX.md)
- [Game Server](../!INDEX.md)
- [Game Server Process](../process/!INDEX.md)
- [Game Server Networking](../networking/!INDEX.md)
- [Game Server Rooms](../rooms/!INDEX.md)
- [Game Server Simulation](../simulation/!INDEX.md)
- [Game Server Integrations](../integrations/!INDEX.md)
- [Telemetry Packet Routing](../networking/telemetry-packet-routing.md)
- [Player Data HTTP Hosting](../integrations/player-data-http-hosting.md)
- [Auth Verifier Integration](../integrations/auth-verifier-integration.md)
- [Match Result Reporting](../integrations/match-result-reporting.md)
- [Devtools](../../../devtools/!INDEX.md)

## Notes

The current logging implementation is intentionally small. It should remain a thin service diagnostic layer until the game server needs a durable observability backend.

The old legacy server logging notes described category adapters that no longer exist in the game-server package. The current implementation routes canonical events through the shared servicelog runtime and keeps event policy in generated observability contract data.
