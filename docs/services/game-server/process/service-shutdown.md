# Service Shutdown

Parent index: [Game Server Process](./!INDEX.md)

## Purpose

This document describes the current graceful game-server process shutdown boundary, including the HTTP lifecycle, service-owned logging closures, hosted diagnostic-aggregator closure, player-data logging runtime closure, and room cleanup.

## Shutdown trigger and ownership

`run()` creates a context with `signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)`. Interrupt and SIGTERM cancel the serving context. The executable does not expose shutdown through an HTTP or packet API; host process signals own the trigger.

`serveHTTPServer` owns the transition from serving to draining. `main.go` owns service traces and dependency teardown. Room manager cleanup remains room-owned and is separate from HTTP/session semantics.

## Current flow

```text
SIGTERM or interrupt
  -> context cancellation
  -> service_stopping on the separate shutdown trace
  -> http.Server.Shutdown(timeout = 5s)
  -> forced http.Server.Close() if Shutdown exceeds/fails
  -> serve goroutine completion
  -> rooms.StopAll()
  -> diagnostic-aggregator closure
  -> player-data logging closure
  -> service_stopped when serving returned normally
  -> game-server logging closure
```

`http_server_lifecycle.go` runs `server.Serve(listener)` in a goroutine. Context cancellation calls `Shutdown` with a five-second timeout. If graceful shutdown returns an error, `Close` is used as a forced fallback and the serve goroutine is joined. `http.ErrServerClosed` is treated as a normal serve result.

The shutdown trace is distinct from the startup trace. `service_stopping` is emitted when draining begins; shutdown/closure degradation events use the shutdown trace. A serving error emits `service_runtime_failed` with the startup trace. `service_stopped` means that serving returned normally; it is emitted before the final game-server logging closure and is not conditional on every deferred closure succeeding.

## Dependency closure order

The process closes dependencies through deferred ownership boundaries. Because these defers execute in LIFO order, the current post-server order is:

1. room manager runs `StopAll`, stopping room cleanup timers, game simulations, and removing rooms;
2. hosted diagnostic aggregator closes its report store/runtime and emits bounded degradation if closure fails;
3. player-data logging runtime closes its active file writer and reports closure degradation through the game-server emitter;
4. `service_stopped` is emitted when serving returned normally;
5. game-server logging runtime closes its active file writer last and reports any close failure through the remaining fallback emitter.

The player-data runtime and stores remain player-data-owned. The game-server process does not call a player-data runtime or store `Close()` method through `playerdata.Runtime`; although the embedded SQLite store exposes a close method, current process teardown does not invoke it through that runtime.

## Canonical shutdown events

The game-server owner emits:

- `service_stopping` when the context begins shutdown;
- `observability_unavailable` for diagnostic-aggregator, player-data logging, or game-server logging closure degradation;
- `service_runtime_failed` if serving or graceful shutdown fails;
- `service_stopped` during deferred teardown, after room cleanup, diagnostic-aggregator closure, and player-data logging closure, but before game-server logging closes.

Room cleanup events remain owned by the room/game workflow. Shutdown does not invent match outcomes, broadcast room snapshots, or turn process shutdown into a gameplay transition.

## Does not own

This process document does not own WebSocket read/write policy, room membership leave rules, match-result resolution, simulation mechanics, player-data persistence, or API auth internals. `http.Server.Shutdown` handles ordinary HTTP connections; hijacked WebSocket connections are not managed by that shutdown call and retain networking-owned lifecycle behavior.

## Code map

```text
services/game-server/cmd/game-server/main.go
services/game-server/cmd/game-server/http_server_lifecycle.go
services/game-server/cmd/game-server/diagnostic_aggregator_host.go
services/game-server/internal/logging/logger.go
services/player-data/logging/logger.go
services/game-server/internal/rooms/manager.go
services/game-server/internal/rooms/room_cleanup.go
services/game-server/internal/game/game.go
services/game-server/internal/game/simulation.go
```

## Tests

Focused process tests cover identity/runtime setup, listener/serve behavior, HTTP timeout configuration, diagnostic registration, and canonical lifecycle ownership. Room and game tests cover `StopAll` and simulation stop primitives. The service-level checks are:

```bash
cd services/game-server && go test -buildvcs=false ./...
cd services/player-data && go test ./...
```

## Related docs

- [Service startup](service-startup.md)
- [Route composition](route-composition.md)
- [Game-server logging and diagnostics](../observability/logging-and-diagnostics.md)
- [Game-server rooms](../rooms/!INDEX.md)
- [Game-server networking](../networking/!INDEX.md)
