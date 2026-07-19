---
author: brian
created: "2026-07-19"
document_id: 019f7d55-fb2c-7fb8-8287-4676589c833b
document_type: general
policy_exempt: false
summary: This index summarizes scoring, awards, counters, combos, streaks, and contributions.
---
# Scoring And Awards

Parent index: [Game Server Simulation](../!INDEX.md)

Scoring, awards, counters, combos, streaks, and contribution documentation lives here.

## Ownership

This folder owns pure score calculation and the generalized match-local awards/counter runtime.

## Does Not Belong

- Combat damage resolution.
- Team assignment.
- Objective condition evaluation.
- Room match lifecycle.
- Client presentation.
- Player-data persistence.

## Direct Files
<!-- doc-ledger:files:start -->

- [awards-and-counter-runtime.md](awards-and-counter-runtime.md) - Scoped counters, idempotent mutations, contributions, combos, and streaks.
- [scoring-policy-and-awards.md](scoring-policy-and-awards.md) - Pure asteroid-destruction score calculation and application handoff.
<!-- doc-ledger:files:end -->

## Stub Files
<!-- doc-ledger:stubs:start -->
<!-- doc-ledger:stubs:end -->

## Direct Folders
<!-- doc-ledger:folders:start -->
<!-- doc-ledger:folders:end -->

## Related Docs

- [Game Server Simulation](../!INDEX.md)
- [Player Counters](../players/player-counters.md)
- [Damage Resolution](../combat/damage-resolution.md)
- [Objective Runtime](../match/objective-runtime.md)

## Notes

The pure scoring policy calculates asteroid score. The awards runtime owns generalized event idempotence and scoped counter mutation.
