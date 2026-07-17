# Logging And Diagnostics

Parent index: [Game Server Observability](./!INDEX.md)

## Purpose

This document describes the game-server logging and diagnostics boundary. It explains how game-server owners emit canonical structured runtime events, how compatibility log levels remain configured, which diagnostic events belong in logs, and what this boundary does not own.

## Overview

Game-server logging is a small internal access point for the shared canonical observability emitter. Production call sites submit `observability.Request` values directly through `logging.Emit`; they do not use category logger methods or scatter raw `slog`, `log`, or `fmt.Println` calls through service code.

The current implementation uses the shared `shared/go/servicelog` runtime for generic console/file fanout and active-file lifecycle. Generated observability definitions own canonical event categories and default levels. The game-server adapter owns runtime configuration, compatibility-category filtering, environment variables, and service-specific byte limits.

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
- Compatibility category loggers for server, networking, rooms, and game behavior while repository-wide bridge retirement remains separate.
- Environment-based log-level configuration.
- Shared log field names for common diagnostic dimensions.
- Runtime diagnostics for recoverable errors, lifecycle events, and unusual conditions.
- WebSocket close classification for expected versus unexpected read/write failures.
- Keeping logs useful without enabling per-tick or per-entity output by default.
- Passing file-policy values through to shared servicelog for structured JSONL output.

The game-server adapter owns compatibility category filtering, environment variables, and service-specific byte limits used for file-output policy. Canonical event names, categories, default levels, trace requirements, and retention tiers come from generated observability contract data.

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

The former category loggers remain compatibility surfaces only:

```go
logging.Server
logging.Network
logging.Rooms
logging.Game
```

Their adapter methods are retained only for compatibility with code outside the completed game-server production call-site rollout. New or migrated game-server code must use generated canonical event names and `logging.Emit`. Category logger calls are not the preferred API and no production game-server call sites remain.

Compatibility category methods support:

```go
Debug(message string, args ...any)
Info(message string, args ...any)
Warn(message string, args ...any)
Error(message string, err error, args ...any)
```

Compatibility methods attach their legacy category and enter the bridge-only `log_message` path. They must not be used to introduce new semantic events.

Package-level helpers also exist:

```go
logging.Debug(...)
logging.Info(...)
logging.Warn(...)
logging.Error(...)
```

Package-level compatibility helpers also exist, but they are not a production call-site API. Canonical event filtering uses the generated event definition; `LOG_*` compatibility overrides apply to the retained category adapter methods.

### File output behavior

Game-server startup enables shared structured file output with:

```go
ConfigureFileOutput("logs/game-server", "game-server")
```

When the server runs from `services/game-server`, the active file path is `logs/game-server/game-server.jsonl.open`.

At runtime, the canonical emitter and shared sink fan records to:

- a safe human-readable rendering on stderr
- the canonical JSON envelope in the active JSONL file

Ordinary canonical event emission cannot emit `log_message`; that bridge-only event is reserved for compatibility adapters. Service names, event metadata, trace requirements, limits, rejection codes, and redaction policy come from generated contract data.

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

Category overrides are:

```text
LOG_GAME
LOG_NETWORK
LOG_ROOMS
LOG_SERVER
```

If a category override is empty or unset, that category inherits `LOG_LEVEL`.

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
- `off` maps to a level above `error`, suppressing logs for that scope.
- Any non-empty unrecognized value currently resolves to `info`.

Default behavior is quiet:

```text
LOG_LEVEL unset -> warn
```

That means `debug` and `info` logs are hidden unless enabled globally or by category.

### Example configurations

Default warnings and errors only:

```bash
cd services/game-server
go run ./cmd/game-server
```

Show process-level startup and shutdown logs:

```bash
cd services/game-server
LOG_SERVER=info go run ./cmd/game-server
```

Debug room lifecycle only:

```bash
cd services/game-server
LOG_LEVEL=warn LOG_ROOMS=debug go run ./cmd/game-server
```

Debug WebSocket and packet routing only:

```bash
cd services/game-server
LOG_LEVEL=warn LOG_NETWORK=debug go run ./cmd/game-server
```

Disable all categories except network warnings and errors:

```bash
cd services/game-server
LOG_LEVEL=off LOG_NETWORK=warn go run ./cmd/game-server
```

Startup file-output status is emitted through the canonical observability flow:

- startup event: `service_starting` with `reason_code = "process_start"`
- file-output success adds no duplicate event
- failure event: `observability_unavailable` with a stable `failure_mode` such as `logging_file_open_failed`, `logging_runtime_degraded`, or `logging_file_close_failed`
- file paths and raw errors are not emitted as canonical fields

## Compatibility category mapping

The names below are retained only for compatibility filtering and adapter diagnostics. They are not production call-site APIs. Production records use the generated event definition's canonical category, such as `service_lifecycle`, `game_networking`, `networking`, `room_lifecycle`, `gameplay`, or `devtools_admin`.

### Server

The compatibility `server` category covers process-level and runtime-wiring diagnostics.

Current examples include:

- server starting
- server stopped
- player-data runtime initialization failure
- player-data reporter initialization failure
- auth verifier initialization failure

This category should not become the home for room, packet, or simulation diagnostics.

### Network

The compatibility `network` category covers WebSocket, packet routing, and transport diagnostics.

Current examples include:

- WebSocket upgrade failure
- WebSocket connection and disconnection
- expected WebSocket read/write close
- unexpected WebSocket read failure
- WebSocket write failure
- packet envelope decode failure
- packet decode failure
- room snapshot marshal failure
- pause-state marshal failure
- telemetry pong encode failure
- debug packet encode/load failure
- debug realtime packet write diagnostics when `LOG_NETWORK=debug`

This category should not own gameplay decisions. It should report network-facing symptoms and include room, player, session, and remote address fields where available.

### Rooms

The compatibility `rooms` category covers room manager, room lifecycle, membership, and match-result lifecycle diagnostics.

Current examples include:

- lobby room created
- single-player room created
- room member left
- room snapshot broadcast after leave
- room cleanup scheduled
- room cleanup skipped
- room cleaned up
- room stopped
- room game over detected
- match result report started
- match result report skipped
- match result report failed
- match result report succeeded

This category should not own simulation internals. It should describe room state transitions and room-owned lifecycle effects.

### Game

The compatibility `game` category covers simulation and player lifecycle diagnostics.

Current examples include:

- collision shapes unavailable
- player added
- player removed
- player paused or resumed
- respawn requested
- respawn blocked
- player respawned
- player died
- player game over
- score awarded
- asteroid split
- devtools gameplay effects routed through real game seams

This category should not log every tick, entity update, or normal packet write.

## Diagnostic field rules

Shared field constants exist for common log dimensions:

```go
logging.FieldCategory   // "category"
logging.FieldError      // "error"
logging.FieldPacketType // "packet_type"
logging.FieldPlayerID   // "player_id"
logging.FieldRemoteAddr // "remote_addr"
logging.FieldRoomID     // "room_id"
```

Use these constants instead of spelling common field names by hand.

For fields without constants, use short snake_case names:

```text
session_id
current_room_id
cleanup_version
active_players
remaining_members
packet_size
write_duration_ms
match_id
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

Realtime packet debug output is currently focused on written gameplay packets only, and only when network debug logging is enabled.

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

Under extreme hot-lane stress, `lane protocol gameplay wire packet written` can produce one debug log per emitted hot-lane chunk. That can be hundreds of network debug records per second. This is expected diagnostic volume when `LOG_NETWORK=debug` is enabled; it is not normal quiet-mode output and does not define packet-budget policy.

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
- `go vet -buildvcs=false ./...` and `go vet -tags nodevtools -buildvcs=false ./...` passed in `services/game-server`.
- `gofmt -l` was clean for non-generated rollout source files; one generated packet file reports formatting drift and was left unchanged under the generated-file policy.
- `PYTHONPATH=. python -m pytest tests tools/tests tools/data_sync/tests` passed, including architecture and observability emission guards.
- `python tools/architecture_guard/main.py` passed.
- `python tools/data_sync/main.py -validate` passed.
- Constants, packets, realtime-wire, and drop-table generated-data `-check` gates passed.
- `git diff --check` passed.
- Production scans found zero references to `logging.Server.`, `logging.Network.`, `logging.Rooms.`, or `logging.Game.` under `services/game-server`.

The repository wrapper `bash tools/ci/run_repo_checks.sh` requires `PYTHONPATH=.` for the local Windows Python module-import environment; the equivalent stages were rerun with that environment and passed. No rollout-related verification failure remained.

Representative canonical envelopes confirm stable event names, valid owning-flow traces, required identifiers, stable reason/failure classifications, one emission per tested transition, and no raw error or unsafe payload fields. Compatibility cleanup is intentionally separate from this rollout record.

### Manual compatibility-level checks

These commands exercise the retained compatibility adapter filtering; canonical event emission remains governed by generated event definitions.

```bash
cd services/game-server
LOG_SERVER=info go run ./cmd/game-server
```

```bash
cd services/game-server
LOG_LEVEL=warn LOG_ROOMS=debug go run ./cmd/game-server
```

```bash
cd services/game-server
LOG_LEVEL=warn LOG_NETWORK=debug go run ./cmd/game-server
```

Expected stderr output is the safe compatibility rendering with category and level context. Direct canonical calls write exactly one canonical JSON envelope to the active JSONL file; compatibility calls remain isolated to the bridge-only path.

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

The old legacy server logging notes described the same core design, but the current implementation now routes through the shared servicelog runtime, keeps adapter-owned category/level/environment policy local, and preserves the existing diagnostic fields and call-site behavior.
