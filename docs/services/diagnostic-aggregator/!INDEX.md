---
author: brian
created: "2026-07-19"
document_id: 019f7d55-fb2c-76f4-8e22-65a29146ed9b
document_type: general
policy_exempt: false
summary: Diagnostic-aggregator is a logically independent bounded diagnostic-report service with a standalone executable and an optional in-process hosting adapter.
---
# Diagnostic Aggregator

Parent index: [Services](../!INDEX.md)

## Purpose

Diagnostic-aggregator accepts bounded diagnostic submissions, validates and safety-inspects them, constructs reports, stores accepted reports, and serves bounded retrieval.

## Overview

The service has two supported runtime modes:

- Standalone: `services/diagnostic-aggregator/cmd/diagnostic-aggregator` owns its HTTP server, listener, signals, shutdown, configuration, service identity, and operational logging.
- Optional in-process hosting: the public `services/diagnostic-aggregator/hosted` adapter registers the same service on a host-owned mux and lets the host own process composition and shutdown.

Production Compose runs the standalone executable as the `diagnostic-aggregator` service on `127.0.0.1:8083` (container port 8080). Game-server co-hosting is an optional composition boundary, not the only runtime.

## Ownership

Diagnostic-aggregator owns route behavior, bearer authentication, submission validation, safety inspection and rejection, canonical event decoding, report construction, bounded JSONL persistence, recovery, retention enforcement, retrieval, service logging, and report-store closure.

In standalone mode it also owns the HTTP listener and process lifecycle. In hosted mode the game-server composition root owns only the shared mux, listener, process signals, shared server shutdown, and host-level lifecycle wiring.

## Non-ownership and dependency boundary

The service does not own gameplay, networking, rooms, match outcomes, player-data persistence, ordinary service-log ownership, continuous log streaming, general-purpose log search, or authoritative audit storage. Co-hosting does not merge these domains.

Only the game-server composition-root adapter may import `services/diagnostic-aggregator/hosted`. Game-server internal/runtime/domain packages and player-data packages must not import diagnostic-aggregator packages or call handlers, application services, stores, or internal types directly. Producers use the diagnostic-report HTTP API through a transport/client seam.

## Configuration and API surface

Defaults include storage at `data/diagnostic-reports`, operational logs at `logs/diagnostic-aggregator`, a 4 MiB request limit, and 14-day report retention. Configuration is read by the hosted package in both modes: `DIAGNOSTIC_AGGREGATOR_ENABLED`, `DIAGNOSTIC_AGGREGATOR_TOKEN`, `DIAGNOSTIC_AGGREGATOR_STORAGE_ROOT`, `DIAGNOSTIC_AGGREGATOR_LOG_ROOT`, `DIAGNOSTIC_AGGREGATOR_RETENTION`, `DIAGNOSTIC_AGGREGATOR_MAX_REQUEST_BYTES`, `BUILD_VERSION`, and `ENVIRONMENT`; standalone also reads `PORT`.

The HTTP surface is `POST /v1/diagnostic-reports` and `GET /v1/diagnostic-reports/{diagnostic_report_id}`. This service documentation records ownership and flow, not a duplicate protocol or devtools contract.

## Direct Files
<!-- doc-ledger:files:start -->

- [runtime-and-report-flow.md](runtime-and-report-flow.md) - Canonical runtime modes, report processing, storage/recovery/retention, and observability.
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
- [Diagnostic Aggregator smoke tests](../../devtools/diagnostic-aggregator/bruno-smoke-tests.md)
- [Observability planning](../../planning/domains/technical/observability-logging-and-diagnostics.md)

## Notes

Diagnostic reports are bounded diagnostic evidence. They are not a continuous log sink or authoritative audit record.
