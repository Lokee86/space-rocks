# Devtools And Telemetry
Parent index: [Devtools Planning](../!INDEX.md)

## Purpose

This doc plans the devtools telemetry seam for future diagnostics and internal visibility.

## Ownership Boundary

This doc owns planning for the world telemetry overlay, gameplay telemetry, network telemetry, debug-only presentation, and separation from the player-facing HUD.

It should stay focused on internal visibility rather than player-facing UI design.

## Current Inputs

- world telemetry overlay inputs
- gameplay telemetry inputs
- network telemetry inputs
- debug-only presentation inputs
- player-facing HUD separation inputs

## Planned Outputs

- telemetry ownership boundaries
- debug-only presentation expectations
- future internal-visibility questions for gameplay and networking

## Related Docs

- [Systems Plan Index](systems-plan-index.md)
- [Network Observability And Packet Budget](network-observability-and-packet-budget.md)
- [Logging And Diagnostics](logging-and-diagnostics.md)
- [Testing And Smoke Strategy](testing-and-smoke-strategy.md)

## Open Planning Questions

- Which telemetry belongs in the world overlay versus logs?
- Which network metrics are useful enough for long-term devtools display?
- Which presentation rules should keep devtools fully separate from the HUD?
