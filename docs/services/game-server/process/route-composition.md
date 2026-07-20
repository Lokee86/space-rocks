---
author: brian
created: "2026-07-19"
document_id: 019f7d55-fb2c-7423-bcca-f5f45ef33451
document_type: general
policy_exempt: false
summary: This document describes the current game-server HTTP route table, hosted diagnostic-report surface, middleware chain, dependency order, and listener/serve boundary.
---
# Route Composition

Parent index: [Game Server Process](./!INDEX.md)

## Purpose

This document describes the current game-server HTTP route table, hosted diagnostic-report surface, middleware chain, dependency order, and listener/serve boundary.

## Composition order

The process composes one mux and one `:8080` listener:

```text
create mux
register diagnostic aggregator and hosted diagnostic-report routes
create room manager
load/set WebRTC transport configuration
construct player-data runtime
construct player-data sink and match-result reporter
construct optional auth verifier
mount /health and /ws
mount player-data HTTP routes
wrap complete mux with httpapi.WithRequestContext
construct http.Server
net.Listen("tcp", ":8080")
serveHTTPServer(context, server, listener, 5s, onStopping)
```

The diagnostic aggregator is constructed and registered before route serving. Player-data runtime construction precedes its HTTP handler construction and match-result reporter construction. The complete mux is wrapped, rather than only one handler, so the player-data HTTP boundary receives the shared request context.

## Routes

| Method | Route | Owner |
| --- | --- | --- |
| `GET` | `/health` | game-server process |
| `GET` | `/ws` | game-server networking |
| `POST` | `/api/player-data/profile` | player-data HTTP API |
| `GET` | `/api/player-data/local-profiles` | player-data HTTP API |
| `POST` | `/api/player-data/local-profiles` | player-data HTTP API |
| `PUT` | `/api/player-data/local-profiles/{local_profile_id}` | player-data HTTP API |
| `DELETE` | `/api/player-data/local-profiles/{local_profile_id}` | player-data HTTP API |
| `GET` | `/api/player-data/local-profiles/default` | player-data HTTP API |
| `PUT` | `/api/player-data/local-profiles/default` | player-data HTTP API |
| `POST` | `/v1/diagnostic-reports` | hosted diagnostic aggregator |
| `GET` | `/v1/diagnostic-reports/{diagnostic_report_id}` | hosted diagnostic aggregator |

The diagnostic report paths are registered by `services/diagnostic-aggregator/hosted.Service` only when `DIAGNOSTIC_AGGREGATOR_ENABLED=true`. The setting defaults to `false`; when disabled, no diagnostic handlers are registered and the routes are unavailable. Exact request validation, bounded report storage, and response behavior remain diagnostic-aggregator-owned.

## Request context and trace identity

`httpapi.WithRequestContext(mux)` wraps the complete mux. At the player-data HTTP boundary, `X-Trace-ID` is continued only when it is a valid UUID; an invalid or absent incoming trace is replaced with a generated UUID. A generated request ID is created for each request and returned in `X-Request-ID`. The owning player-data handler uses both context values for canonical HTTP events. The middleware does not return `X-Trace-ID`.

The middleware preserves one request identity when a handler is reached through an already wrapped path. It does not log request bodies, bearer tokens, raw upstream errors, or private profile data.

## Dependency ownership

```text
mux
  -> diagnostic aggregator hosted service + report routes
  -> room manager + WebSocket handler
  -> player-data runtime + HTTP handlers
  -> auth verifier adapter
  -> match-result reporter
  -> request-context middleware
  -> HTTP server/listener lifecycle
```

The route composer owns reachability and dependency injection. Networking owns WebSocket upgrade/session behavior. Player-data owns HTTP validation, identity routing, store access, and stable response errors. Diagnostic-aggregator owns report contract validation and report storage. The route table owns none of those internal behaviors.

## Failure and serving lifecycle

Dependency construction failures for diagnostic-aggregator, player-data runtime, or match-result reporter fail startup before a listener is bound. Optional auth-verifier construction can degrade to a nil verifier; protected request paths fail closed.

The process calls `net.Listen`, not `ListenAndServe` directly. `serveHTTPServer` serves the listener with the configured `http.Server`, observes context cancellation, performs five-second graceful shutdown, and uses `server.Close` as the forced fallback. `http.ErrServerClosed` is normal during shutdown. Listener and serve failures emit canonical runtime failure events and return nonzero.

## Code map

```text
services/game-server/cmd/game-server/main.go
services/game-server/cmd/game-server/diagnostic_aggregator_host.go
services/game-server/cmd/game-server/http_server_lifecycle.go
services/game-server/cmd/game-server/http_server.go
services/game-server/cmd/game-server/player_data_http.go
services/game-server/cmd/game-server/auth_config.go
services/game-server/cmd/game-server/webrtc_config.go
services/player-data/httpapi/request_context.go
services/player-data/httpapi/profile_handler.go
services/player-data/httpapi/local_profiles_handler.go
services/diagnostic-aggregator/hosted/service.go
```

## Tests

Focused tests cover diagnostic registration, request-context continuation/replacement, player-data canonical HTTP failures, reporter construction, and HTTP server lifecycle. Run:

```bash
cd services/game-server && go test -buildvcs=false ./...
cd services/player-data && go test ./...
cd services/diagnostic-aggregator && go test ./...
```

## Related docs

- [Service startup](service-startup.md)
- [Service shutdown](service-shutdown.md)
- [Player-data observability and logging](../../player-data/observability-and-logging.md)
- [Diagnostic aggregator runtime and report flow](../../diagnostic-aggregator/runtime-and-report-flow.md)
- [Diagnostic aggregator hosting](../integrations/diagnostic-aggregator-hosting.md)
