---
author: brian
created: "2026-07-19"
document_id: 019f7d55-fb2c-7e16-9711-cfc452b6e940
document_type: general
policy_exempt: false
summary: This document defines the optional game-server composition boundary for hosting the independent diagnostic-aggregator service.
---
# Diagnostic-Aggregator Hosting

Parent index: [Game Server Integrations](./!INDEX.md)

## Purpose

This document records the optional game-server composition boundary for in-process diagnostic-aggregator hosting. It does not describe the standalone runtime or duplicate the diagnostic HTTP contract.

## Overview

`diagnostic-aggregator` is an independent service with a standalone executable and production Compose service. The game-server may also host the service in-process when `services/game-server/cmd/game-server/` imports the public `services/diagnostic-aggregator/hosted` adapter.

The only permitted import is:

```text
services/game-server/cmd/game-server/ -> services/diagnostic-aggregator/hosted
```

No package under `services/game-server/internal/`, gameplay, networking, rooms, match reporting, or player-data may import diagnostic-aggregator packages or call diagnostic handlers, report services, stores, or internals directly.

## Hosted flow and lifecycle

```text
host composition root
  -> load hosted configuration
  -> construct hosted service
  -> register diagnostic routes on shared mux
  -> host owns listener, signals, and server shutdown
  -> close hosted service during teardown
```

The hosted service owns diagnostic route behavior, processing, storage, retention, operational logging, and report-store closure. A process-local registration remains an HTTP/API boundary; it is not permission for service reach-through.

## Responsibilities

The game-server composition root owns:

- construction through `hosted.Service`;
- registration on the shared HTTP mux;
- host-level configuration and lifecycle wiring;
- shared listener, signals, and HTTP server shutdown.

Diagnostic-aggregator owns authentication, validation, safety policy, report construction, persistence, retrieval, retention, service observability, and closure. Producers use the HTTP transport/API.

## Non-ownership

This integration does not own diagnostic report policy, handlers, application services, report stores, standalone process behavior, Docker/Compose deployment, gameplay, networking, rooms, match reporting, player-data behavior, or direct in-process diagnostic calls.

## Failure and degradation

Hosted configuration, construction, registration, and closure errors are handled by the game-server composition lifecycle. Diagnostic request and storage failures remain diagnostic-aggregator behavior and must not grant the game-server runtime diagnostic implementation ownership. The standalone process has its own listener and shutdown failure behavior; it is outside this integration boundary.

## Code map

- `services/game-server/cmd/game-server/diagnostic_aggregator_host.go` - sole game-server hosting adapter.
- `services/diagnostic-aggregator/hosted/` - public hosted construction, registration, and close surface.
- `services/diagnostic-aggregator/cmd/diagnostic-aggregator/` - independent standalone executable; not a game-server dependency.

## Tests

Game-server hosting tests verify construction, registration, request routing, and close behavior at the composition root. Diagnostic API, report-processing, storage/recovery/retention, logging, and standalone producer coverage belongs to the diagnostic-aggregator service suite. The Pitlord rule `game-runtime-no-diagnostic-aggregator-dependency` enforces the import boundary.

## Related docs

- [Game Server Integrations](./!INDEX.md)
- [Diagnostic Aggregator](../../diagnostic-aggregator/!INDEX.md)
- [Diagnostic Aggregator runtime and report flow](../../diagnostic-aggregator/runtime-and-report-flow.md)
- [Game Server](../!INDEX.md)
- [Services index](../../../services/!INDEX.md)

## Notes

Co-hosting is a composition option, not the canonical deployment assumption and not an ownership merger.
