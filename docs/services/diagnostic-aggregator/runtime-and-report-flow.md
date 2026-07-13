# Runtime and Report Flow

Parent index: [Diagnostic Aggregator](./!INDEX.md)

## Purpose

The diagnostic aggregator is the bounded service for triggered diagnostic reports. It accepts diagnostic material, applies validation and safety policy, constructs a finalized report, stores it for bounded retrieval, and exposes the report API.

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

## Storage and Retention

The diagnostic-aggregator owns bounded report storage for accepted reports. The default retention window is 14 days and is enforced when the hosted service starts. This storage is diagnostic evidence, not authoritative audit storage.

## Future Detachment

Future detachment should replace process composition and addressing only. Producers and API contracts should continue to use the same hosted service contract and report routes. A future standalone deployment is not implemented by the current architecture.

## Related Docs

- [Diagnostic Aggregator](./!INDEX.md)
- [Services](../!INDEX.md)
- [Observability planning](../../planning/domains/technical/observability-logging-and-diagnostics.md)
