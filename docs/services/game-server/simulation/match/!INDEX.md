# Match

Parent index: [Game Server Simulation](../!INDEX.md)

Mode resolution, objectives, and match-result documentation lives here.

## Ownership

This folder owns current simulation documentation for resolved match rules, mode evaluation, objective runtime, locked match decisions, and result summary construction.

## Does Not Belong

- Room membership and assignment execution.
- Client results presentation.
- Player-data persistence.
- Combat damage mechanics.
- Future campaign content planning.

## Direct Files
<!-- doc-ledger:files:start -->

- [match-outcomes-and-results.md](match-outcomes-and-results.md) - Locked decisions, participant and team outcomes, and idempotent summaries.
- [modes-and-match-rules.md](modes-and-match-rules.md) - Room mode configuration and immutable resolved match rules.
- [objective-runtime.md](objective-runtime.md) - Objective definitions, instances, facts, timers, visibility, and snapshots.
<!-- doc-ledger:files:end -->

## Stub Files
<!-- doc-ledger:stubs:start -->
<!-- doc-ledger:stubs:end -->

## Direct Folders
<!-- doc-ledger:folders:start -->
<!-- doc-ledger:folders:end -->

## Related Docs

- [Game Server Simulation](../!INDEX.md)
- [Game Server Rooms](../../rooms/!INDEX.md)
- [Scoring](../scoring/!INDEX.md)

## Notes

Modes select policies; focused owner systems execute them and return normalized facts for match evaluation.
