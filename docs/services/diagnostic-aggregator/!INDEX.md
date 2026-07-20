---
author: brian
created: "2026-07-19"
document_id: 019f7d55-fb2c-76f4-8e22-65a29146ed9b
document_type: general
policy_exempt: false
summary: Diagnostic-aggregator is a logically independent bounded diagnostic-report service. It is currently co-hosted by the game-server process through the public services/diagnostic-aggregator/hosted package.
---
# Diagnostic Aggregator

Parent index: [Services](../!INDEX.md)

Diagnostic-aggregator is a logically independent bounded diagnostic-report service. It is currently co-hosted by the game-server process through the public `services/diagnostic-aggregator/hosted` package.

## Ownership

The diagnostic-aggregator owns diagnostic-report route registration, bearer authentication, submission validation, safety inspection and rejection, report construction, bounded JSONL report storage, retention enforcement, retrieval, and report-store closure.

The game-server composition root owns the HTTP mux, listener, shared server address, process signals, server shutdown, and the hosted service lifecycle. Co-hosting must not make gameplay, networking, or domain code depend on diagnostic implementation details.

## Permanent Dependency Boundary

Co-hosting permits only composition-root construction, registration, and closure of the hosted diagnostic-aggregator service. The sole external Go import is the game-server composition adapter importing the public `services/diagnostic-aggregator/hosted` package.

Game-server internal/runtime/domain packages and every player-data package must not import any diagnostic-aggregator package. Game-server and player-data code must submit or retrieve reports only through a transport/client implementation of the diagnostic-report HTTP API, never through handler, application-service, or report-store calls. Diagnostic handlers, services, stores, and internal types must not be passed into game-server or player-data constructors.

## Hosted Configuration

Hosted operation is disabled by default. Configuration is loaded from:

- `DIAGNOSTIC_AGGREGATOR_ENABLED`
- `DIAGNOSTIC_AGGREGATOR_TOKEN`
- `DIAGNOSTIC_AGGREGATOR_STORAGE_ROOT`
- `DIAGNOSTIC_AGGREGATOR_RETENTION`
- `DIAGNOSTIC_AGGREGATOR_MAX_REQUEST_BYTES`

The default report retention is 14 days. When enabled, a nonblank whitespace-free bearer token is required. The hosted routes use the game-server's existing base URL, currently `http://127.0.0.1:8080/v1/diagnostic-reports` and `/v1/diagnostic-reports/{diagnostic_report_id}`.

## Does Not Belong

- Ordinary service log ownership or continuous log streaming.
- General-purpose log search or analytics.
- Authoritative audit storage.
- Game-server listener, HTTP server, signal, or process-shutdown ownership.
- A separate standalone executable or standalone health/readiness surface.
- Product-wide planning or domain flow documentation.

## Future Detachment

Future detachment should replace only process composition and addressing. Report producers, the public hosted service contract, diagnostic-report API contract, validation/safety policy, and report storage behavior should remain unchanged. Standalone deployment is not currently implemented.

## Direct Files
<!-- doc-ledger:files:start -->

- [runtime-and-report-flow.md](runtime-and-report-flow.md) - Hosted diagnostic-report intake, processing, storage, retrieval, and ownership boundaries.
<!-- doc-ledger:files:end -->
## Stub Files
<!-- doc-ledger:stubs:start -->
<!-- doc-ledger:stubs:end -->
## Direct Folders
<!-- doc-ledger:folders:start -->
<!-- doc-ledger:folders:end -->
## Related Docs

- [Services index](../!INDEX.md)
- [Game-server diagnostic-aggregator hosting](../game-server/integrations/diagnostic-aggregator-hosting.md)
- [Observability planning](../../planning/domains/technical/observability-logging-and-diagnostics.md)

## Notes

This service boundary describes bounded diagnostic reports. It does not turn the service into a continuously fed log sink or an audit-record authority.