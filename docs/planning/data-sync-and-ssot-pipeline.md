# Data Sync And SSOT Pipeline

## Purpose

This doc plans the shared source-of-truth pipeline for generated data and schema sync.

## Ownership Boundary

This doc owns planning for constants, packets, drop tables, player-data schema, OpenAPI contracts, generator outputs, and schema drift enforcement.

It should stay focused on the sync pipeline and source-of-truth boundaries, not on feature policy.

## Current Inputs

- constants inputs
- packet inputs
- drop table inputs
- player-data schema inputs
- OpenAPI contract inputs
- generator output inputs
- schema drift inputs

## Planned Outputs

- SSOT pipeline planning boundaries
- generator and sync expectations across data domains
- schema-drift enforcement questions

## Related Docs

- [Systems Plan Index](systems-plan-index.md)
- [Player Data And Persistence](player-data-and-persistence.md)
- [Network Observability And Packet Budget](network-observability-and-packet-budget.md)
- [Deployment And Packaging](deployment-and-packaging.md)

## Open Planning Questions

- Which data domains should share the same sync pipeline?
- Which generator outputs need the strongest drift checks?
- Which schema boundaries should remain separately owned even if the pipeline is shared?
