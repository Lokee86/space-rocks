# Logging And Diagnostics
Parent index: [Technical Planning](../!INDEX.md)

## Purpose

This doc plans future logging and diagnostics direction across the client and server.

## Ownership Boundary

This doc owns planning for future logging and diagnostics, while treating the existing implemented docs as the current reference surface.

Implemented references:

- [docs/services/game-server/observability/logging-and-diagnostics.md](../../../services/game-server/observability/logging-and-diagnostics.md)
- [docs/services/client/client-logging.md](../../../services/client/client-logging.md)

## Current Inputs

- server logging inputs
- client logging inputs
- structured diagnostic inputs
- packet and network observability inputs

## Planned Outputs

- logging and diagnostics planning boundaries
- a clear link between new planning work and implemented logging guidance
- future cross-system observability questions

## Related Docs

- [Systems Plan Index](systems-plan-index.md)
- [Network Observability And Packet Budget](network-observability-and-packet-budget.md)
- [Devtools And Telemetry](devtools-and-telemetry.md)
- [Testing And Smoke Strategy](testing-and-smoke-strategy.md)

## Open Planning Questions

- Which diagnostics should remain in logs instead of moving into telemetry?
- Which future log fields are needed for packet and replay analysis?
- Which logging rules should stay identical across client and server?
