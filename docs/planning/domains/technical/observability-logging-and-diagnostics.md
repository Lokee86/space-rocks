---
author: brian
created: "2026-07-19"
document_id: 019f7d55-fb2c-75de-8a9c-dfb4e8f30cbf
document_type: general
policy_exempt: false
summary: This document owns remaining future work for observability, logging, and diagnostics. Current contract, emitter, service runtime, and diagnostic-aggregator behavior belongs in current data/service documents linked below.
---
# Observability, Logging, And Diagnostics

Parent index: [Technical Planning](./!INDEX.md)

## Purpose

This document owns remaining future work for observability, logging, and diagnostics. Current contract, emitter, service runtime, and diagnostic-aggregator behavior belongs in current data/service documents linked below.

## Overview

The canonical event contract, generated bindings, service-owned rolling log runtimes, and bounded diagnostic-report service are implemented. This plan is limited to future product-triggered report creation, client-facing diagnostic workflows, resilient producer transport, bundle correlation, remaining compatibility retirement, stronger release verification, optional collection, and audit-domain durability.

## Current baseline (implemented)

The current baseline is not a design proposal:

- The observability SSoT exists under `shared/contracts/observability/`.
- `tools/data_sync` owns validation and generation for Go, GDScript, Ruby, JSON, and docs outputs.
- Generated consumers exist for Go, GDScript, Ruby, JSON, and the generated Markdown reference.
- Canonical emitters exist across the game-server, player-data, diagnostic-aggregator, API-server, and client components.
- Service-owned rolling runtimes exist across those five components, with active/archive layout, rotation, compression of completed segments, recovery, retention, and bounded degraded behavior.
- The diagnostic aggregator accepts, validates, stores, and retrieves bounded diagnostic reports.
- Completed targeted workflow migrations include game-server canonical-only emission and diagnostic-aggregator canonical-only workflow emission. API request/auth/player-stat/match-result workflows and broad client canonical rollout are current; player-data has canonical HTTP/dispatcher coverage while its category compatibility layer remains transitional.

Current ownership is documented in:

- [Canonical event emission](../../../observability/canonical-event-emission.md)
- [Observability contract](../../../data/observability-contract.md)
- [API-server observability](../../../services/api-server/observability-and-logging.md)
- [Player-data observability](../../../services/player-data/observability-and-logging.md)
- [Client logging](../../../services/client/client-logging.md)
- [Game-server logging and diagnostics](../../../services/game-server/observability/logging-and-diagnostics.md)
- [Diagnostic aggregator runtime and report flow](../../../services/diagnostic-aggregator/runtime-and-report-flow.md)

## Planning boundary

This document does not reopen settled baseline decisions. The contract files, generator ownership, machine-consumed generated outputs, canonical envelope, service writer model, and diagnostic-aggregator intake/storage boundary are implementation facts. Planning work starts after those seams.

## Remaining future work

### Product-triggered diagnostic producers and uploads

Add product-triggered report creation and upload beyond the current manual/operational producer path. Producers must be bounded, explicit about user intent, non-blocking to gameplay, and safe when the aggregator is unavailable.

### Client-facing diagnostic surfaces

Add client-facing bug-report, copy-diagnostics, and consent/presentation surfaces. These should select bounded local evidence and identify the owning operation trace without exposing raw private data.

### Non-blocking producer transports

Add resilient asynchronous producer transports with bounded queues, retry/backoff policy, cancellation, and visible drop/rejection accounting. Transport failure must degrade diagnostics rather than gameplay or player-data behavior.

### Diagnostic bundles and correlation

Complete bundle construction and correlation across selected client, game-server, player-data, API, and diagnostic-aggregator evidence. Define the product-facing bundle shape, size budget, redaction boundary, and trace/report linkage without changing the canonical envelope.

### Compatibility bridge retirement

Retire remaining legacy adapters after service workflow migrations are complete and compatibility tests can be removed. The client and player-data adapters are current transitional examples. This is separate from canonical emitter infrastructure and from already-completed workflow rollouts.

### Hosted release verification

Add hosted staging/release checks for identity configuration, generated-contract drift, per-process active writer ownership, retention cleanup, compression/recovery, degraded behavior, trace propagation, and bounded aggregator intake.

### Optional centralized collection

Evaluate a future centralized collection path only after product, privacy, operational, and retention requirements justify it. Local service-owned logs remain independently useful and are not replaced by a central collector.

### Audit-domain durability

Plan stronger durability and retention for audit-domain events where product/security requirements demand it. This is distinct from operational diagnostics and must keep its own ownership and access policy.

## Explicit non-goals

- Do not redesign the SSoT or generator ownership.
- Do not add a second contract source.
- Do not make the diagnostic aggregator the only diagnostic source.
- Do not turn ordinary observability into an unbounded telemetry stream.
- Do not make logging failure block gameplay, HTTP request handling, or player-data persistence.

## Related docs

- [Canonical event emission](../../../observability/canonical-event-emission.md)
- [Observability contract](../../../data/observability-contract.md)
- [Diagnostic aggregator runtime](../../../services/diagnostic-aggregator/runtime-and-report-flow.md)
- [Operations](../../../operations/!INDEX.md)

## Notes

Current operational and service behavior belongs in the linked current owners. Future work should extend those seams without turning this planning document into a second contract or runtime authority.

## Related planning

- [Development roadmap](../../development-roadmap.md)
- [Runtime performance and scale budget](runtime-performance-and-scale-budget.md)
- [Operational readiness and failure modes](operational-readiness-and-failure-modes.md)
- [Build, release, environment, and compatibility matrix](build-release-and-environment-matrix.md)
- [Verification and quality gates](verification-and-quality-gates.md)
