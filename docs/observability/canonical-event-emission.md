---
author: brian
created: "2026-07-19"
document_id: 019f7d55-fb2c-72c5-9204-b0f756654f1f
document_type: general
policy_exempt: false
summary: This document owns the cross-service canonical-emitter boundary. The observability contract is authored under shared/contracts/observability/; data-sync validates it and generates the language metadata consumed by runtime emitters....
---
# Canonical Event Emission

Parent index: [Observability](./!INDEX.md)

## Purpose

This document owns the cross-service canonical-emitter boundary. The observability contract is authored under `shared/contracts/observability/`; `data-sync` validates it and generates the language metadata consumed by runtime emitters. Service documents own file-runtime and workflow details; this document records the shared seam and current migration state.

## Current architecture

The current cross-language infrastructure is:

```text
shared/contracts/observability/*.toml
  -> data-sync
  -> generated Go, GDScript, Ruby, JSON, and Markdown contract metadata

Observability::Emitter       Ruby API emitter
Observability::WorkerRuntime Ruby file runtime
Observability::PumaHooks     Ruby Puma lifecycle/request hook
shared/go/observabilityevent Go emitter
shared/go/servicelog         Go rolling sink/runtime
client/scripts/logging/observability_emitter.gd
                              GDScript emitter
```

Every emitter writes the same canonical JSONL envelope. Generated metadata owns service names, event definitions, allowed services, levels, trace requirements, limits, redaction actions, retention policy, and stable rejection codes. Runtime code owns lifecycle, sink, and workflow context; it must not duplicate contract policy locally.

The canonical emitter boundary owns:

- event/service eligibility and trace-required checks;
- UUID, context, free-form field, type, size, null, and event-count validation;
- contract-defined redaction and unsafe-field rejection;
- deterministic envelope serialization;
- non-raising write-failure results and emitter status counters;
- bounded stderr/warning reporting without writing rejected payloads.

The boundary does not own domain correlation policy. Workflows create or continue the trace and supply domain identifiers.

## Canonical emitter infrastructure

Go services use `shared/go/observabilityevent.Emitter` through service-owned `logging.Emit` functions. `shared/go/servicelog` receives serialized canonical records and owns console/file fanout and rolling file lifecycle.

The API server uses `Observability::Emitter`, `Observability::WorkerRuntime`, and `Observability::PumaHooks`. Puma hooks construct one runtime per process/worker and discard inherited runtime state after fork. API request concern code calls the hook rather than constructing writers or envelopes.

The client uses `observability_emitter.gd` behind `ClientLogger`. `emit_canonical` is the semantic client entry point. Client text/category helpers remain compatibility adapters and route through `emit_legacy`; they are not the model for new semantic events.

## Compatibility bridge and retirement

The compatibility bridge is intentionally separate from semantic canonical emission. Legacy category/text helpers may emit the contract's bridge event through the service adapter, but ordinary canonical calls cannot emit that bridge event. The bridge preserves existing public helpers while workflow owners migrate; it does not authorize new semantic call sites.

Bridge retirement is a separate workstream from emitter infrastructure and workflow rollout:

1. **Canonical emitter infrastructure** — shared contract, generated metadata, validation/redaction, writers, status, and failure behavior. This exists.
2. **Semantic workflow rollout** — domain owners replace meaningful lifecycle, request, auth, player-stat, match-result, connection, and diagnostic events with canonical calls. This is complete in some services and mixed in others.
3. **Compatibility bridge retirement** — remove remaining legacy helpers only after their callers have migrated and compatibility tests no longer need them. This remains future work.

## Current per-service migration state

| Service | Current production state |
| --- | --- |
| `game-server` | Canonical-only production emission through `logging.Emit`. The process identity gate, lifecycle events, dependency failures, runtime failures, shutdown, and service-owned workflow events use the canonical emitter. |
| `diagnostic-aggregator` | Canonical-only production workflow emission. Hosted lifecycle, request/rejection, report intake, storage, and closure use canonical events; no production compatibility-emitter calls remain in that service. |
| `api-server` | Canonical request, auth, player-stat, and match-result emission. The API runtime uses `Observability::Emitter`, `Observability::WorkerRuntime`, and `Observability::PumaHooks`; no removed formatter architecture is part of the current design. |
| `client` | Broad canonical rollout is complete. Legacy text/category helpers remain available and route through `emit_legacy` for compatibility callers. |
| `player-data` | Canonical HTTP and dispatcher events exist, including match-result success, duplicate suppression, and read/write failures. Legacy category methods remain available as a transitional compatibility layer. |

## ClientOperationTrace

`ClientOperationTrace` is the client-owned operation identity object. It creates one UUID-backed trace for an operation and exposes its operation name and `trace_id`. Connection attempts own connection traces; boot and room operations create or continue traces; downstream auth, room, gameplay, and devtools flows receive the owning trace instead of inventing unrelated IDs. The object carries correlation only; it does not own event definitions or file output.

## Verification

The cross-language acceptance, redaction, and rejection fixtures live at:

```text
shared/contracts/observability/fixtures/emitter_cases.json
```

Generated outputs must remain drift-free:

```bash
data-sync -validate -observability
data-sync -diff -observability -go -gds -ruby -json -docs
data-sync -push -observability -go -gds -ruby -json -docs
data-sync -check -observability -go -gds -ruby -json -docs
```

## Related owner documents

- [Observability contract data owner](../data/observability-contract.md)
- [Client logging](../services/client/client-logging.md)
- [API-server observability and logging](../services/api-server/observability-and-logging.md)
- [Player-data observability and logging](../services/player-data/observability-and-logging.md)
- [Game-server logging and diagnostics](../services/game-server/observability/logging-and-diagnostics.md)
- [Diagnostic aggregator runtime and report flow](../services/diagnostic-aggregator/runtime-and-report-flow.md)

## Does not own

Canonical emission does not own diagnostic submission, metrics, tracing infrastructure, durable hosted retention, gameplay behavior, networking behavior, or the retirement schedule for compatibility helpers. Those remain with the owning service or planning document.
