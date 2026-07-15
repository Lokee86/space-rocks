# Runtime and Report Flow

Parent index: [Diagnostic Aggregator](./!INDEX.md)

## Purpose

The diagnostic aggregator is the bounded service for triggered diagnostic reports. It accepts diagnostic material, applies validation and safety policy, constructs a finalized report, stores it for bounded retrieval, and exposes the report API.

## Overview

Diagnostic-aggregator is logically independent and currently co-hosted by the game-server process. The game-server composition root hosts the routes and process lifecycle; diagnostic-aggregator owns report processing and storage. There is no standalone diagnostic-aggregator process. Producer integration beyond the manual producer remains deferred.

## Co-hosted Runtime

There is no standalone diagnostic-aggregator executable, listener, process configuration package, process logger, service identity package, or standalone health/readiness surface. The manual producer remains at `services/diagnostic-aggregator/cmd/diagnostic-submit/`.

The current runtime is co-hosted by the game-server process:

```text
game-server composition root
  -> creates the shared HTTP mux, listener, and server
  -> loads DIAGNOSTIC_AGGREGATOR_* hosted configuration
  -> constructs hosted diagnostic-aggregator when enabled
  -> registers /v1/diagnostic-reports and /v1/diagnostic-reports/
  -> owns process signals and server shutdown
```

Hosted operation is disabled by default. The current environment variables are:

- `DIAGNOSTIC_AGGREGATOR_ENABLED`
- `DIAGNOSTIC_AGGREGATOR_TOKEN`
- `DIAGNOSTIC_AGGREGATOR_STORAGE_ROOT`
- `DIAGNOSTIC_AGGREGATOR_RETENTION`
- `DIAGNOSTIC_AGGREGATOR_MAX_REQUEST_BYTES`

The default storage root is `data/diagnostic-reports`, the default request limit is 4 MiB, and the default report retention is 14 days. The shared game-server base URL is currently `http://127.0.0.1:8080`; report routes are under `/v1/diagnostic-reports`.

## Permitted Co-hosting Boundary

Co-hosting permits only composition-root construction, registration, and closure. The sole external Go import is the game-server composition adapter importing the public `services/diagnostic-aggregator/hosted` package. Game-server internal/runtime/domain packages and every player-data package must not import any diagnostic-aggregator package.

Game-server and player-data must submit or retrieve reports only through a transport/client implementation of the diagnostic-report HTTP API. They must never call diagnostic handlers, application services, or report stores directly. Diagnostic handlers, services, stores, and internal types must not be passed into game-server or player-data constructors.

The game-server composition root owns:

- the HTTP mux and listener
- the shared HTTP server and address
- process signals
- server shutdown and process-level error reporting

The diagnostic-aggregator owns:

- route registration
- bearer authentication
- submission envelope validation
- complete-payload safety inspection and rejection
- canonical event decoding
- finalized report construction and validation
- bounded report storage and retrieval
- startup retention enforcement
- report-store closure

The diagnostic package does not listen, create an HTTP server, handle signals, forward continuous logs, or shut down the game-server process.

## Report Flow

The bounded flow is:

```text
manual bug report or allowlisted severe failure/crash
  -> bounded diagnostic submission to the shared game-server base URL
  -> bearer authentication
  -> request and payload validation
  -> safety inspection and rejection of unsafe material
  -> canonical event decoding
  -> finalized diagnostic report construction
  -> bounded durable JSONL storage
  -> retrieval by the diagnostic-report surface
```

Reports are not a continuous stream. Ordinary service logs remain owned by their producing services and are not continuously forwarded to diagnostic-aggregator. Diagnostic failure must not affect gameplay.

## Code root

`services/diagnostic-aggregator/`

## Responsibilities

Diagnostic-aggregator owns diagnostic-report route registration, bearer authentication, submission validation, safety inspection, report construction, bounded storage and retrieval, retention enforcement, and report-store closure. The game-server composition root owns shared HTTP process composition, route hosting, process signals, and server shutdown.

## Does not own

Diagnostic-aggregator does not own the game-server HTTP listener or process lifecycle, gameplay, networking, rooms, match reporting, player-data behavior, continuous log streaming, general-purpose log search, authoritative audit storage, or a standalone executable and health/readiness surface. Producer integration beyond the current manual producer remains deferred.

## Domain roles

Diagnostic-aggregator is the bounded diagnostic-report processing and storage authority. The game-server composition root is a host and lifecycle coordinator only; it does not gain report-processing authority. Producers and clients provide bounded diagnostic submissions through the HTTP API.

Co-hosting permits only composition-root construction, registration, and closure. The sole external Go import is the game-server composition adapter importing the public `services/diagnostic-aggregator/hosted` package. Game-server internal/runtime/domain packages and every player-data package must not import any diagnostic-aggregator package. Direct handler, application-service, report-store, or internal-type calls and constructor injection remain forbidden.

## Protocols and APIs

The diagnostic-report HTTP API exists to submit bounded diagnostic material and retrieve accepted reports. Manual producers and future transport/client implementations are callers. Diagnostic-aggregator handlers and application services are authoritative for validation, safety policy, report construction, and processing; the report store is authoritative for bounded persistence and retrieval.

Data crossing the boundary includes the authenticated submission envelope, bounded diagnostic payload, correlation fields, and report retrieval identifiers/results. The API does not own gameplay state, player-data persistence, match outcomes, continuous logs, or caller-side retry and degradation policy.

Game-server and player-data must submit or retrieve reports only through a bounded transport/client implementation of this HTTP API, never through direct imports or calls to diagnostic handlers, application services, stores, or internal types. Those diagnostic objects must not be passed into their constructors.

## Canonical Operational Events

The hosted service owns diagnostic-aggregator lifecycle emission. Startup creates one trace for `service_starting`, initialization failures, and `service_started`. Shutdown creates a different trace for `service_stopping`, report-store close failure, and `service_stopped`. The game-server host does not duplicate the aggregator's successful-start event; it emits only the disabled configuration decision and any additional host-level close dependency failure.

Each aggregator HTTP request receives a new aggregator-owned request ID and provisional trace ID. Once a request body has been strictly decoded, a valid submitted `correlation.trace_id` replaces the provisional trace for the remaining intake operation. Submitted request, session, room, match, player, and account correlation stays inside the diagnostic report and does not become the aggregator HTTP request identity.

Emission ownership is intentionally singular:

- the HTTP handler emits `diagnostic_report_rejected` for transport, decoding, envelope, safety, and useful identifier rejection paths;
- the report service emits one `aggregator_event_accepted` after successful validation, `diagnostic_report_stored` after durable save, and `aggregator_storage_failed` for save, non-not-found load, or corrupt stored-report failures;
- the hosted service emits lifecycle events and report-store shutdown failure;
- the storage package returns errors and does not duplicate owning-boundary events.

Operational fields are bounded classifications. Bearer tokens, authorization headers, request bodies, rejected values, unrestricted errors, user descriptions, and embedded event collections are never copied into operational events. Event rejection or write failure does not change startup, shutdown, HTTP, or report-storage behavior.

Current production diagnostic-aggregator code has zero `log_message` bridge sites. Repository-wide bridge retirement remains deferred until the other services complete their workflow migrations.

## Data ownership

Diagnostic-aggregator owns validated report contents, safety decisions, finalized report shape, report identifiers, bounded JSONL storage, and retention metadata. Producers own the source diagnostic material and correlation context they submit. The game-server owns process and gameplay data; player-data owns player-data aggregates and persistence. Neither service gains diagnostic-storage ownership by co-hosting or producing a request.

## Storage and Retention

The diagnostic-aggregator owns bounded report storage for accepted reports. The default retention window is 14 days and is enforced when the hosted service starts. This storage is diagnostic evidence, not authoritative audit storage.

## Future Detachment

Future detachment should replace process composition and addressing only. Producers and API contracts should continue to use the same hosted service contract and report routes. A future standalone deployment is not implemented by the current architecture.

## Code map

- `services/diagnostic-aggregator/` - diagnostic-report service implementation.
- `services/diagnostic-aggregator/hosted` - public hosted-service adapter.
- `services/diagnostic-aggregator/cmd/diagnostic-submit/` - current manual producer.
- `services/game-server/cmd/game-server/` - composition-root route hosting and lifecycle wiring.

## Tests

Tests should cover bounded HTTP submission and retrieval, validation and safety rejection, report processing, storage/retention behavior, and hosted construction/registration/closure. Boundary tests must preserve transport-only producer access and must not introduce direct diagnostic dependencies into game-server or player-data runtime/domain code.

## Related docs

- [Diagnostic Aggregator](./!INDEX.md)
- [Game-server diagnostic-aggregator hosting](../game-server/integrations/diagnostic-aggregator-hosting.md)
- [Services](../!INDEX.md)
- [Observability planning](../../planning/domains/technical/observability-logging-and-diagnostics.md)

## Notes

This document describes the current co-hosted implementation without treating co-hosting as an ownership merger. Configuration defaults and deferred producer integration remain unchanged.
