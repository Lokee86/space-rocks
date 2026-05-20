# Server Logging

This document covers only the server logging feature.

## Problem It Solves

The server has several systems that can fail or behave unexpectedly without an obvious client-side symptom: websocket connections, room cleanup, player lifecycle, scoring, spawning, and server startup. The logging feature gives those systems a shared, structured way to report what happened.

The goal is practical development visibility without flooding the terminal during normal play. By default, the logger is quiet and only shows warnings and errors. More detailed logs can be enabled globally or by subsystem when debugging.

## Where It Lives

The logging package lives here:

```text
services/game-server/internal/logging/logger.go
```

The package is internal to the server module. Server code should import it as:

```go
import "github.com/Lokee86/space-rocks/server/internal/logging"
```

Current server call sites use these category loggers:

```go
logging.Server
logging.Network
logging.Rooms
logging.Game
```

The package also still exposes package-level helpers:

```go
logging.Debug(...)
logging.Info(...)
logging.Warn(...)
logging.Error(...)
```

Prefer the category loggers for new server code so logs can be filtered by subsystem.

## Initialization

The logger uses Go's standard library `log/slog` package. It writes text logs to `os.Stderr`.

The default logger and category loggers are initialized when the logging package is loaded. The game server then applies environment configuration in:

```text
services/game-server/cmd/game-server/main.go
```

```go
logging.Configure(os.Getenv(logging.EnvGlobalLevel))
```

`Configure` sets the global default log level from `LOG_LEVEL`, then applies category-specific overrides from the environment.

## Configuration

The current logging configuration is environment-variable based.

| Variable | Scope |
| --- | --- |
| `LOG_LEVEL` | Default level for global logs and every category |
| `LOG_GAME` | Overrides only `logging.Game` |
| `LOG_NETWORK` | Overrides only `logging.Network` |
| `LOG_ROOMS` | Overrides only `logging.Rooms` |
| `LOG_SERVER` | Overrides only `logging.Server` |

If a category variable is empty or unset, that category uses `LOG_LEVEL`.

Examples:

```bash
cd services/game-server
LOG_LEVEL=warn go run ./cmd/game-server
```

```bash
cd services/game-server
LOG_LEVEL=warn LOG_ROOMS=debug go run ./cmd/game-server
```

```bash
cd services/game-server
LOG_LEVEL=off LOG_NETWORK=warn go run ./cmd/game-server
```

## Supported Levels

Supported levels:

| Level | Meaning |
| --- | --- |
| `debug` | Detailed lifecycle and diagnostic logs |
| `info` | Important normal events |
| `warn` or `warning` | Unusual but recoverable situations |
| `error` | Failed operations |
| `off` | Disable logs for that scope |

Default behavior:

```text
LOG_LEVEL defaults to warn
```

That means `debug` and `info` logs are hidden unless enabled.

`off` is supported. Internally it maps to a level above `error`, so no logs pass for that scope.

## Categories

### `logging.Server`

Use for process-level server logs.

Current examples:

```go
logging.Server.Info("server starting", "addr", ":8080")
logging.Server.Error("server stopped", err, "addr", ":8080")
```

### `logging.Network`

Use for websocket and network transport logs.

Current examples include:

- websocket upgrade failure
- websocket connected/disconnected
- packet decode failure
- expected websocket close
- websocket read/write failure

### `logging.Rooms`

Use for room manager lifecycle logs.

Current examples include:

- room created
- room joined/left
- cleanup scheduled/canceled/completed
- cleanup skipped because a room is active or the cleanup version is stale

### `logging.Game`

Use for game simulation and player lifecycle logs.

Current examples include:

- collision shape load warning
- player added/removed
- respawn requested/blocked/successful
- player death/game over
- score awarded
- asteroid split
- state marshal failure

## How To Log Errors

Use the category logger for the subsystem you are in.

For real failed operations, use `Error`. It automatically appends the structured `error` field:

```go
logging.Network.Error("websocket write failed", err,
    logging.FieldRoomID, roomID,
    logging.FieldPlayerID, playerID,
    logging.FieldRemoteAddr, remoteAddr,
)
```

Do not manually add `logging.FieldError` when using `Error`; the helper already does that.

For recoverable warnings where the code continues, use `Warn`. If an error value is useful, include it explicitly:

```go
logging.Network.Warn("websocket packet decode failed",
    logging.FieldError, err,
    logging.FieldRoomID, roomID,
    logging.FieldPlayerID, playerID,
)
```

For normal lifecycle events, use `Debug` unless the event is important enough to show in a normal development run.

## Field Names

Shared field constants currently available:

```go
logging.FieldCategory   // "category"
logging.FieldError      // "error"
logging.FieldPacketType // "packet_type"
logging.FieldPlayerID   // "player_id"
logging.FieldRemoteAddr // "remote_addr"
logging.FieldRoomID     // "room_id"
```

Use these constants instead of spelling common field names by hand. For fields that do not have constants yet, use short snake_case names:

```go
"active_players"
"cleanup_version"
"respawn_cooldown"
```

## Correct Usage Examples

Server startup:

```go
logging.Server.Info("server starting", "addr", ":8080")
```

Fatal server listen failure:

```go
if err := http.ListenAndServe(":8080", mux); err != nil {
    logging.Server.Error("server stopped", err, "addr", ":8080")
    os.Exit(1)
}
```

Room cleanup lifecycle:

```go
logging.Rooms.Debug("room cleanup scheduled",
    logging.FieldRoomID, roomID,
    "cleanup_delay", manager.cleanupDelay.String(),
    "cleanup_version", cleanupVersion,
)
```

Recoverable bad packet:

```go
logging.Network.Warn("websocket packet decode failed",
    logging.FieldError, err,
    logging.FieldRoomID, roomID,
    logging.FieldPlayerID, playerID,
)
```

Game lifecycle:

```go
logging.Game.Info("player died",
    logging.FieldPlayerID, playerID,
    "lives", lives,
    "respawn_delay", respawnDelay,
)
```

## What Not To Log

Avoid logging every tick. The server runs simulation ticks frequently, so per-tick logs will flood output and make real errors harder to see.

Do not log:

- every player position update
- every physics step
- every collision candidate check
- every state packet write when it succeeds
- every asteroid spawn candidate
- raw packet payloads unless a focused debug task requires it
- secrets, tokens, credentials, or personally identifying data

Prefer logging state transitions:

- player joined or removed
- respawn requested or blocked
- room cleanup scheduled or skipped
- websocket write failed
- bad packet received

## Testing Logging

Run the normal server tests:

```bash
cd services/game-server
env GOCACHE=/tmp/space-rocks-go-build go test -buildvcs=false ./...
```

To manually check log levels, run the server with different environment variables.

Default quiet mode:

```bash
cd services/game-server
go run ./cmd/game-server
```

Show server startup info:

```bash
cd services/game-server
LOG_SERVER=info go run ./cmd/game-server
```

Debug room lifecycle only:

```bash
cd services/game-server
LOG_LEVEL=warn LOG_ROOMS=debug go run ./cmd/game-server
```

Debug websocket/network logs only:

```bash
cd services/game-server
LOG_LEVEL=warn LOG_NETWORK=debug go run ./cmd/game-server
```

Disable everything except network warnings/errors:

```bash
cd services/game-server
LOG_LEVEL=off LOG_NETWORK=warn go run ./cmd/game-server
```

Expected output format is `slog` text output on stderr. Category logs include a field like:

```text
category=network
```

## Troubleshooting

### I do not see `server starting`

Default logging is `warn`, and `server starting` is logged at `info`.

Use:

```bash
LOG_SERVER=info go run ./cmd/game-server
```

### I enabled `LOG_ROOMS=debug` but still do not see room logs

Room logs only appear when room events happen. Connect a client to `/ws` or `/ws?room_id=some-room`, then disconnect and wait for cleanup behavior.

### I set `LOG_LEVEL=off` and nothing appears

That is expected. `off` disables logs for that scope. Use a category override if you want one subsystem back:

```bash
LOG_LEVEL=off LOG_NETWORK=warn go run ./cmd/game-server
```

### I see websocket read warnings when clients close

Expected websocket close codes are logged at `debug`. Abrupt or unexpected read failures are logged at `warn`. If a client process is killed or the socket closes without a normal close frame, a warning can be legitimate.

### I see too much room cleanup output

Room lifecycle logs are `debug`. They should only appear if `LOG_LEVEL=debug` or `LOG_ROOMS=debug` is set.

### A category override does not seem to work

Check spelling. Current category env vars are:

```text
LOG_GAME
LOG_NETWORK
LOG_ROOMS
LOG_SERVER
```

Empty category env vars inherit `LOG_LEVEL`.

## Future Evolution

Keep the current logging package small until the server needs more.

Good future additions:

- Tests for `parseLevel` and category override behavior
- Optional JSON output for production-style log ingestion
- Optional file logging for packaged local play
- A request or connection ID for correlating websocket events
- More shared field constants if repeated fields emerge
- Debug-only logs for targeted spawn/collision investigations

Avoid adding:

- per-tick logging
- broad packet dumps by default
- a second logging framework alongside `slog`
- logging behavior that changes game state

Logging should remain observational: it should help explain what happened without affecting gameplay.
