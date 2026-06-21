# Player Data And Persistence
Parent index: [Platform Planning](../!INDEX.md)

## Purpose

This doc plans the shared player-data contract and persistence seam across local and online routes.

## Ownership Boundary

This doc owns planning for shared logical player-data contracts, the local SQLite route, the Rails/Postgres route, player-data service extraction, migrations, schema parity, and contract drift.

It should stay on the data and storage boundary rather than folding in gameplay simulation or UI details.

## Current Inputs

- shared logical player-data contracts
- local SQLite route inputs
- Rails/Postgres route inputs
- player-data service extraction inputs
- migration inputs
- schema parity inputs
- contract drift inputs

## Planned Outputs

- player-data ownership boundaries
- storage-route expectations for local and online persistence
- migration and drift-check planning hooks

## Related Docs

- [Planning](../../../!INDEX.md)
- [Account And Identity Systems](../account-and-identity-systems.md)
- [Progression And Rewards](../../gameplay/progression-and-rewards.md)
- [Player Experience Systems](../../gameplay/player-experience-systems.md)

## Open Planning Questions

- Which logical contracts must remain shared across local and Rails-backed routes?
- Which migration path should keep local and online schema parity easiest to inspect?
- Which contract drift checks should happen before the player-data service is split out?
