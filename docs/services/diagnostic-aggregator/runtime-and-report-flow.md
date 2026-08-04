---
author: brian
created: "2026-07-19"
document_id: 019f7d55-fb2c-70db-aab8-778efa779304
document_type: general
policy_exempt: false
summary: The diagnostic aggregator has a standalone executable and an optional hosted adapter. Both modes share bounded report intake, processing, storage, recovery, retention, and observability ownership.
---
# Runtime and Report Flow

Parent index: [Diagnostic Aggregator](./!INDEX.md)

## Purpose

This document is the canonical owner for diagnostic-aggregator runtime modes, report processing, storage/recovery/retention, and service observability.

## Overview

The executable and hosted adapter construct the same bounded diagnostic-report service. The runtime boundary changes who owns process composition; it does not change report policy or storage behavior.

## Runtime modes

Standalone startup is:

```text
cmd/diagnostic-aggregator
  -> hosted.LoadConfig and hosted.New
  -> standalone http.Server and :PORT listener
  -> route registration
  -> signal-driven server shutdown
  -> service/store/log closure
```

The standalone Docker image builds `./cmd/diagnostic-aggregator`, exposes port 8080, and persists `/data` for reports and operational logs. Production Compose runs it as its own `diagnostic-aggregator` service and the game-server depends on that service being started.

Optional hosted startup is:

```text
game-server composition root
  -> hosted.LoadConfig and hosted.New
  -> shared game-server mux registration
  -> host-owned listener, signals, and server shutdown
  -> hosted service closure
```

Hosted operation is disabled when `DIAGNOSTIC_AGGREGATOR_ENABLED` is false. The standalone executable is still a valid independent runtime; this setting controls the hosted construction path.

## Ownership boundary

The aggregator owns route registration, authentication, envelope and payload validation, safety inspection, canonical event decoding, report construction, storage/retrieval, retention, its operational logger, and closure. In standalone mode it additionally owns the listener and process lifecycle. A game-server host owns only shared process composition and must not reach through the HTTP/API boundary.

Reports are not a continuous stream. Ordinary service logs remain owned by their producing services.

## Report flow and state

```text
bounded producer submission
  -> bearer authentication and HTTP limits
  -> strict JSON/envelope validation
  -> safety inspection and rejection
  -> event decoding and report validation
  -> report ID and summary construction
  -> durable JSONL save and canonical stored event
  -> authenticated retrieval and stored-report validation
```

Accepted reports are the service's durable state. Rejected material is not stored. Diagnostic failure is degraded independently and must not affect gameplay or player-data behavior.

## Storage, recovery, and retention

The JSONL store creates root, active, archive, and quarantine directories. It appends reports with durable flushes, rotates active segments by size or age, and optionally gzip-compresses archived segments. Startup recovery examines an interrupted active segment before reopening it. Retrieval scans active and archived segments and validates decoded report identity and shape.

Retention is enforced during service initialization using the configured report window (14 days by default). Expired report segments are deleted under the store lock. Close finalizes the active segment and makes the store unavailable for further operations. Storage errors map to bounded service-unavailable behavior; not-found remains a normal retrieval result.

## Failure and recovery

Configuration, logger, store, retention, authorizer, report-service, or handler initialization failure prevents that runtime from becoming ready. Standalone listener or shutdown failure produces a non-zero process result. Hosted construction failure is returned to the game-server composition root.

Malformed, oversized, unauthorized, unsafe, invalid, corrupt, or unavailable requests receive bounded HTTP errors. Stored-report decode/validation failures are treated as unavailable and are emitted as storage failures. Operational-event write failure does not change report or lifecycle behavior.

## Observability

The diagnostic-aggregator logger owns service lifecycle events, request rejection, accepted intake, durable storage, and storage-failure events. Startup and shutdown use separate traces. Each HTTP operation gets an aggregator-owned request ID and provisional trace; a valid submitted correlation trace may replace the provisional trace for the remaining intake operation.

Operational fields are bounded classifications. Tokens, authorization headers, bodies, rejected values, unrestricted errors, descriptions, and embedded event collections are excluded. There are no production `log_message` bridge sites in the current aggregator implementation.

## Code map

- `services/diagnostic-aggregator/cmd/diagnostic-aggregator/` - standalone executable and process lifecycle.
- `services/diagnostic-aggregator/hosted/` - shared construction, configuration, route registration, logger, and close adapter.
- `services/diagnostic-aggregator/internal/diagnosticapi/` - HTTP transport behavior.
- `services/diagnostic-aggregator/internal/diagnosticreports/` - report validation, construction, summary, and storage orchestration.
- `services/diagnostic-aggregator/internal/diagnostics/` and `internal/redaction/` - diagnostic data validation, decoding, and safety policy.
- `services/diagnostic-aggregator/internal/storage/jsonlstore/` - JSONL layout, append, rotation, recovery, archive, retrieval, retention, and close.
- `services/diagnostic-aggregator/internal/logging/` - service logging and observability emission.
- `services/diagnostic-aggregator/cmd/diagnostic-submit/` - manual HTTP producer.

## Tests

Focused coverage exists for standalone producer end-to-end behavior, hosted configuration/construction/registration/closure, API authentication and contracts, report processing, diagnostics and redaction, JSONL layout/codec/rotation/recovery/retention, and logging. Run `cd services/diagnostic-aggregator && go test ./...` for the service suite.

## Related docs

- [Diagnostic Aggregator](./!INDEX.md)
- [Game-server diagnostic-aggregator hosting](../game-server/integrations/diagnostic-aggregator-hosting.md)
- [Diagnostic Aggregator smoke tests](../../devtools/diagnostic-aggregator/bruno-smoke-tests.md)
- [Observability canonical event emission](../../observability/canonical-event-emission.md)
- [Services](../!INDEX.md)

## Notes

The hosted adapter is an optional in-process boundary. It is not evidence that the standalone executable, Docker image, or production service is absent.
