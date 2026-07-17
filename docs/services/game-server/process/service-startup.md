# Service Startup

Parent index: [Game Server Process](./!INDEX.md)

## Purpose

This document describes the current game-server executable startup composition and the logging/runtime identity gate. The process hosts game-server, player-data, and the diagnostic-aggregator HTTP surface in one process while keeping their ownership boundaries separate.

## Startup identity gate

Before constructing runtimes, startup requires two inputs for each service identity:

```text
BUILD_VERSION
ENVIRONMENT
```

`loadLoggingIdentity` also assigns a UUID service-instance identity. The game-server identity is `game-server`; the player-data identity is `player-data`. Missing `BUILD_VERSION` or `ENVIRONMENT` is a startup configuration failure and prevents serving. These identities are separate even though both runtimes are hosted by the same executable.

`LOG_LEVEL` remains the actual game-server level configuration boundary. The removed category-specific game-server environment controls are not part of current startup. Canonical semantic events use generated event policy and `logging.Emit`.

## Current sequence

```text
run()
  signal.NotifyContext(SIGTERM, interrupt)
  load game-server identity (BUILD_VERSION + ENVIRONMENT)
  load player-data identity (BUILD_VERSION + ENVIRONMENT)
  configure game-server canonical logging runtime
  emit service_starting
  configure game-server rolling output
  emit observability degradation if output is unavailable/degraded
  configure player-data logging runtime
  configure player-data rolling output
  emit player-data observability degradation when needed
  create mux
  construct/register hosted diagnostic aggregator
  construct room manager
  load and set WebRTC transport config
  construct player-data runtime
  construct player-data sink and match-result reporter
  construct optional auth verifier
  mount /health and /ws
  mount player-data HTTP routes
  wrap the complete mux with httpapi.WithRequestContext
  create http.Server
  net.Listen("tcp", ":8080")
  emit service_started
  serveHTTPServer with context-driven shutdown
```

A player-data logging runtime is constructed before the player-data runtime itself. Diagnostic-aggregator registration occurs before route serving and owns its own hosted lifecycle.

## Canonical lifecycle events

Game-server startup and runtime failures use canonical events through `services/game-server/internal/logging`:

- `service_starting` with the startup trace;
- `dependency_initialization_failed` for logging, diagnostic-aggregator, player-data runtime, reporter, or other dependency failures;
- `service_runtime_failed` for listener/serve failures;
- `service_started` after the listener is ready;
- `observability_unavailable` when file output or dependency closure degrades.

The process keeps a separate shutdown trace for `service_stopping`, shutdown failures, and `service_stopped`. Informational success chatter is not a startup contract; the process does not emit claims such as a configured structured log file or loaded WebRTC transport as substitute lifecycle events.

## Dependencies and route preparation

`registerDiagnosticAggregator(mux)` loads hosted configuration, constructs the hosted service, and registers diagnostic report routes before the remainder of the process routes are served. The service is closed during process teardown.

The room manager is created after diagnostic registration. WebRTC configuration is read from its environment seam and installed in networking; transport behavior remains owned by networking after this setup.

`buildPlayerDataRuntime()` constructs the shared in-process player-data runtime before the reporter and handlers. The runtime remains player-data-owned even though the game-server hosts it. The reporter receives a player-data sink, and the auth verifier is passed to WebSocket and authenticated player-data profile handling.

The complete mux is wrapped with:

```go
httpapi.WithRequestContext(mux)
```

This gives player-data HTTP requests a trace/request identity boundary and preserves `X-Trace-ID` continuation/replacement behavior described in the player-data owner document.

## HTTP server lifecycle boundary

Startup creates the configured `http.Server`, then explicitly binds the listener:

```go
server := newHTTPServer(httpapi.WithRequestContext(mux))
listener, err := net.Listen("tcp", ":8080")
serveHTTPServer(ctx, server, listener, 5*time.Second, onStopping)
```

`serveHTTPServer` runs `server.Serve(listener)` and waits for either a serve result or process context cancellation. The context cancellation path performs graceful shutdown and has a forced `server.Close()` fallback when the five-second shutdown deadline is exceeded.

## Failure behavior

Fail-fast startup failures are identity configuration, logging-runtime configuration, diagnostic-aggregator registration, player-data runtime construction, reporter construction, and listener binding. Optional auth-verifier configuration can continue without a verifier; request paths fail closed where authentication is required.

Logging/output degradation emits a bounded canonical status event and leaves the process serving through available console/stderr behavior. A serve failure emits `service_runtime_failed` and returns nonzero.

## Code map

```text
services/game-server/cmd/game-server/main.go
services/game-server/cmd/game-server/logging_identity.go
services/game-server/cmd/game-server/diagnostic_aggregator_host.go
services/game-server/cmd/game-server/http_server.go
services/game-server/cmd/game-server/http_server_lifecycle.go
services/game-server/cmd/game-server/player_data_http.go
services/game-server/cmd/game-server/auth_config.go
services/game-server/cmd/game-server/webrtc_config.go
services/game-server/internal/logging/logger.go
services/player-data/logging/logger.go
services/player-data/playerdata/configured_runtime.go
services/player-data/httpapi/request_context.go
services/game-server/internal/matchreporting/runtime_reporter.go
```

## Tests and verification

Focused process tests cover identity requirements, diagnostic registration, server timeout/lifecycle behavior, and observability event ownership. Service verification is split by module:

```bash
cd services/game-server && go test -buildvcs=false ./...
cd services/player-data && go test ./...
cd services/player-data && go test -tags noembeddedsqlite ./...
```

## Related docs

- [Service shutdown](service-shutdown.md)
- [Route composition](route-composition.md)
- [Game-server logging and diagnostics](../observability/logging-and-diagnostics.md)
- [Player-data observability and logging](../../player-data/observability-and-logging.md)
- [Diagnostic aggregator hosting](../integrations/diagnostic-aggregator-hosting.md)
