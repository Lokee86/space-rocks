---
author: brian
created: "2026-07-19"
document_id: 019f7d55-fb2c-7bbf-a68d-8bc50614ae31
document_type: general
policy_exempt: false
summary: This index summarizes the observability docs.
---
# Observability

This index summarizes the observability docs.

Parent index: [Documentation](../!INDEX.md)

## Ownership

The shared observability contract and canonical event-emission boundary own these docs.

## Does Not Belong

Service-specific runtime behavior belongs in the relevant service documentation. Metrics, tracing infrastructure, diagnostic submission, and durable telemetry storage are outside this boundary.

## Direct Files
<!-- doc-ledger:files:start -->

- [Canonical Event Emission](canonical-event-emission.md) - Cross-language canonical emitter infrastructure, service migration state, workflow rollout, and bridge retirement.
<!-- doc-ledger:files:end -->

## Stub Files
<!-- doc-ledger:stubs:start -->
<!-- doc-ledger:stubs:end -->

## Direct Folders
<!-- doc-ledger:folders:start -->

- [generated](generated/!INDEX.md) - Generated documentation.
<!-- doc-ledger:folders:end -->

## Related Docs

- [Generated Contract Reference](generated/contract-reference.md)
- [Observability Contract](../data/observability-contract.md)
- [API-server Observability And Logging](../services/api-server/observability-and-logging.md)
- [Player-data Observability And Logging](../services/player-data/observability-and-logging.md)
- [Client Logging](../services/client/client-logging.md)
- [Game Server Logging And Diagnostics](../services/game-server/observability/logging-and-diagnostics.md)
- [Diagnostic Aggregator Runtime And Report Flow](../services/diagnostic-aggregator/runtime-and-report-flow.md)

## Notes

Generated references describe the contract data. Handwritten docs describe runtime ownership and integration boundaries.