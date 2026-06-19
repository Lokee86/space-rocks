# Logging And Diagnostics

Parent index: [Game Server Observability](./!README.md)

## Purpose

This document describes the game-server logging and diagnostics boundary.

It explains how the game server emits structured runtime logs, how log levels are configured, which categories exist, what diagnostic events belong in logs, and what this boundary does not own.

## Overview

Game-server logging is implemented as a small internal wrapper around Go `log/slog`.

The logging package gives game-server code a shared way to report runtime behavior without scattering raw `slog`, `log`, or `fmt.Println` calls through service code. It writes text logs to stderr and supports category-specific loggers so noisy subsystems can be enabled independently during debugging.

The current flow is:

```text
game-server runtime event
-> category logger
-> slog text handler
-> stderr
```

Logging is observational. It should explain what happened in the process, connection, room, or simulation, but it must not change gameplay state, packet routing, persistence, or auth behavior.

## Code root

```text
services/game-server/
```

## Responsibilities

Logging and diagnostics own the game-server side of:

* Structured log emission through the internal logging package.
* Shared category loggers for server, networking, rooms, and game behavior.
* Environment-based log-level configuration.
* Shared log field names for common diagnostic dimensions.
* Runtime diagnostics for recoverable errors, lifecycle events, and unusual conditions.
* Packet-size and slow-write warnings for gameplay presentation packets.
* WebSocket close classification for expected versus unexpected read/write failures.
* Keeping logs useful without enabling per-tick or per-entity output by default.

## Does not own

Logging and diagnostics does not own:

* Process startup or shutdown behavior.
* HTTP route composition.
* WebSocket transport lifecycle.
* Packet schema or codec ownership.
* Room membership rules.
* Match lifecycle rules.
* Gameplay simulation authority.
* Player-data persistence.
* Auth token verification.
* Devtools command behavior.
* Client-side log presentation.
* Durable telemetry storage.
* Metrics, tracing, or log aggregation infrastructure.

Those systems may emit logs, but they own their own runtime behavior.

## Domain roles

Game-server logging participates in technical diagnostics only.

It supports development and runtime investigation across:

* server process initialization
* auth verifier initialization
* player-data runtime initialization
* WebSocket upgrade/read/write behavior
* inbound packet decode failures
* outbound packet encode/write diagnostics
* room creation, cleanup, match-over, and match-result reporting
* player lifecycle, respawn, scoring, pause, and game-over events
* devtools command effects routed through real gameplay seams

It is not a product-facing telemetry system and does not provide durable analytics.

## Protocols and APIs

The logging surface is an internal Go package plus environment variables. It is not a client protocol and is not exposed over HTTP or WebSocket.

Server code consumes the package by importing:

```go
import "github.com/Lokee86/space-rocks/server/internal/logging"
```

Current category loggers are:

```go
logging.Server
logging.Network
logging.Rooms
logging.Game
```

Each category logger supports:

```go
Debug(message string, args ...any)
Info(message string, args ...any)
Warn(message string, args ...any)
Error(message string, err error, args ...any)
```

Category loggers automatically attach a `category` field.

Package-level helpers also exist:

```go
logging.Debug(...)
logging.Info(...)
logging.Warn(...)
logging.Error(...)
```

New game-server code should prefer category loggers because category loggers can be filtered independently.

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

* Empty level values resolve to `warn`.
* `warning` is treated as `warn`.
* `off` maps to a level above `error`, suppressing logs for that scope.
* Any non-empty unrecognized value currently resolves to `info`.

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

## Log categories

### Server

`logging.Server` is for process-level and runtime-wiring diagnostics.

Current examples include:

* server starting
* server stopped
* player-data runtime initialization failure
* player-data reporter initialization failure
* auth verifier initialization failure

This category should not become the home for room, packet, or simulation diagnostics.

### Network

`logging.Network` is for WebSocket, packet routing, and transport diagnostics.

Current examples include:

* WebSocket upgrade failure
* WebSocket connection and disconnection
* expected WebSocket read/write close
* unexpected WebSocket read failure
* WebSocket write failure
* packet envelope decode failure
* packet decode failure
* room snapshot marshal failure
* pause-state marshal failure
* telemetry pong encode failure
* debug packet encode/load failure
* gameplay presentation packet too large
* gameplay presentation packet write too slow

This category should not own gameplay decisions. It should report network-facing symptoms and include room, player, session, and remote address fields where available.

### Rooms

`logging.Rooms` is for room manager, room lifecycle, membership, and match-result lifecycle diagnostics.

Current examples include:

* lobby room created
* single-player room created
* room member left
* room snapshot broadcast after leave
* room cleanup scheduled
* room cleanup skipped
* room cleaned up
* room stopped
* room game over detected
* match result report started
* match result report skipped
* match result report failed
* match result report succeeded

This category should not own simulation internals. It should describe room state transitions and room-owned lifecycle effects.

### Game

`logging.Game` is for simulation and player lifecycle diagnostics.

Current examples include:

* collision shapes unavailable
* player added
* player removed
* player paused or resumed
* respawn requested
* respawn blocked
* player respawned
* player died
* player game over
* score awarded
* asteroid split
* devtools gameplay effects routed through real game seams

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

* bearer tokens
* internal service tokens
* Discord access tokens
* OAuth codes
* raw OAuth state
* client secrets
* raw auth headers
* raw packet payloads containing sensitive data

## Error logging rules

Use the category logger for the subsystem where the event occurs.

For failed operations, use `Error`. The helper automatically appends the structured `error` field:

```go
logging.Network.Error("websocket write failed", err,
    logging.FieldRoomID, roomID,
    logging.FieldPlayerID, playerID,
    logging.FieldRemoteAddr, remoteAddr,
)
```

Do not manually add `logging.FieldError` when using category `Error`.

For recoverable warnings where the code continues, use `Warn`. If an error value matters, include it explicitly:

```go
logging.Network.Warn("websocket packet envelope decode failed",
    logging.FieldError, err,
    logging.FieldRoomID, roomID,
    logging.FieldPlayerID, playerID,
    "session_id", sessionID,
    logging.FieldRemoteAddr, remoteAddr,
)
```

For normal lifecycle events, prefer `Debug` unless the event is important during normal development.

## Data ownership

Logging and diagnostics does not own durable data.

It emits transient process output to stderr.

It may include identifiers and runtime facts that help correlate behavior:

* room ID
* player ID
* session ID
* remote address
* packet type
* match ID
* packet size
* write duration
* cleanup version
* player count
* active member counts

It must not persist logs, mutate player data, or become the source of truth for gameplay or account state.

If Space Rocks later adds durable telemetry, tracing, metrics, or log aggregation, that should be documented as a separate integration or observability system.

## Diagnostic policy

Good log events are state transitions, failures, and unusual conditions.

Use logs for:

* process initialization failures
* service dependency initialization failures
* WebSocket upgrade/read/write failures
* expected versus unexpected socket closes
* malformed packet input
* packet encode or marshal failures
* packet size or slow-write warnings
* room lifecycle transitions
* match result reporting lifecycle
* player spawn, respawn, death, and game-over transitions
* devtools effects applied through real gameplay seams

Avoid logs for:

* every simulation tick
* every player position update
* every physics step
* every collision candidate
* every successful state packet write
* every successful input packet
* every asteroid spawn candidate
* broad packet dumps
* duplicate logs for the same event at multiple layers

Logs should make production and development failures easier to diagnose without drowning normal gameplay output.

## Code map

Primary implementation files:

```text
services/game-server/internal/logging/logger.go
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
services/game-server/internal/networking/outbound/gameplay_state_metrics.go
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

There are no dedicated tests for `services/game-server/internal/logging/logger.go` in the current tree.

Current verification is mostly indirect through normal game-server tests and manual environment checks.

Relevant test areas include:

```text
services/game-server/internal/networking/
services/game-server/internal/networking/outbound/
services/game-server/internal/rooms/
services/game-server/internal/game/
services/game-server/internal/devtools/
services/game-server/tests/
```

Useful verification commands:

```bash
cd services/game-server
go test -buildvcs=false ./...
```

Manual log-level checks should run the game server with targeted environment variables:

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

Expected output is `slog` text output on stderr. Category logs include a `category` field.

## Related docs

* [Game Server Observability](./!README.md)
* [Game Server](../!README.md)
* [Game Server Process](../process/!README.md)
* [Game Server Networking](../networking/!README.md)
* [Game Server Rooms](../rooms/!README.md)
* [Game Server Simulation](../simulation/!README.md)
* [Game Server Integrations](../integrations/!README.md)
* [Telemetry Packet Routing](../networking/telemetry-packet-routing.md)
* [Player Data HTTP Hosting](../integrations/player-data-http-hosting.md)
* [Auth Verifier Integration](../integrations/auth-verifier-integration.md)
* [Match Result Reporting](../integrations/match-result-reporting.md)
* [Devtools](../../../devtools/!README.md)

## Notes

The current logging implementation is intentionally small. It should remain a thin service diagnostic layer until the game server needs a durable observability backend.

The old legacy server logging notes described the same core design, but current implementation has additional call sites for player-data initialization, auth verifier setup, match-result reporting lifecycle, telemetry pong encoding failures, and gameplay presentation packet size/write-duration diagnostics.
