# Runtime and Report Flow

Parent index: [Diagnostic Aggregator](./!INDEX.md)

## Purpose

The diagnostic aggregator is the bounded service for triggered diagnostic reports. It accepts diagnostic material, applies validation and redaction policy, constructs a report or bundle, stores it for bounded retrieval, and exposes service health and lifecycle state.

## Runtime

The executable is `services/diagnostic-aggregator/cmd/diagnostic-aggregator/`.

Current Stage 2 runtime responsibilities include:

- loading environment-scoped configuration
- writing service-owned rolling logs when file logging is enabled
- serving the diagnostic-report HTTP surface
- exposing readiness and liveness health behavior
- closing the report store during shutdown
- draining the HTTP server with the configured shutdown timeout, then force-closing if needed

The default listen address is `127.0.0.1:8091`. The default diagnostic-report retention is 14 days and can be configured through the service environment.

## Report flow

The intended bounded flow is:

```text
manual bug report or allowlisted severe failure/crash
  -> bounded diagnostic upload
  -> intake authentication and rate limiting [later integration]
  -> request and payload validation
  -> redaction of unsafe or sensitive fields
  -> diagnostic report/bundle construction
  -> bounded durable storage
  -> retrieval by the diagnostic-report surface
```

Validation and redaction are service-owned processing stages. Report construction preserves diagnostic context without making arbitrary application logs the service's data source.

Public client uploads are not a continuous stream. In the later client/API integration, uploads are expected to be bounded, authenticated, and rate-limited. They are triggered by manual bug reports or allowlisted severe failures/crashes, not by every ordinary client event or by continuous telemetry forwarding. The public upload integration is not yet wired into the current Stage 2 service boundary; this document labels that policy explicitly rather than describing it as implemented runtime behavior.

## Logging boundary

Ordinary service logs remain owned by the service that produced them and are written to that service's rolling files. They are not continuously sent to the diagnostic aggregator.

The aggregator may receive selected diagnostic context when a bounded report is created, but it is not a general log-search platform, log warehouse, or continuous log-forwarding target.

## Storage and retrieval

The service owns bounded diagnostic-report storage and retrieval for its accepted reports. The default retention window is 14 days; the configured retention value is applied by the report storage layer.

This storage is diagnostic evidence, not authoritative audit storage. Audit-grade records and their durability guarantees remain owned by the appropriate audit or business-data boundary. A diagnostic report may contain audit-relevant context, but storing a report here does not make it the system of record.

## Health and shutdown

The service reports lifecycle and health state through its runtime HTTP surface. Startup marks the service ready after the listener is bound and required dependencies are available. Server failure marks the service stopping and closes the store.

On context cancellation, the runtime:

1. marks the service stopping
2. calls HTTP server shutdown with the configured timeout
3. force-closes the server if graceful draining fails
4. waits for the serving goroutine to return
5. closes the report store exactly once
6. records service-stopped lifecycle output

This is service-process shutdown. It does not own client disconnect behavior, game-server room cleanup, ordinary producer log rotation, or audit-record finalization.

## Ownership boundaries

This service owns:

- triggered diagnostic-report intake
- validation and rejection of unsafe or invalid payloads
- redaction before storage or export
- diagnostic report and bundle construction
- bounded report storage and retrieval
- service health, readiness, and lifecycle reporting
- HTTP draining and report-store closure during shutdown

It does not own:

- continuous service-log collection
- general log search
- gameplay outcomes or client session state
- public authentication and rate-limit infrastructure before later integration
- authoritative audit storage

## Related docs

- [Diagnostic Aggregator](./!INDEX.md)
- [Services](../!INDEX.md)
- [Observability contract](../../observability/!INDEX.md)
