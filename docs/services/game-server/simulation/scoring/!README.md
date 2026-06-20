# Scoring

This index summarizes the scoring docs.

Parent index: [Game Server Simulation](../!README.md)

## Ownership

This folder owns game-server simulation scoring policy, score event vocabulary, pure award calculation, and game-owned award application boundaries.

## Does Not Belong

- Player counter storage and mutation.
- Combat damage resolution.
- Pickups.
- Room match results.
- Client presentation.
- Player-data persistence.
- Packet schema or data pipeline ownership.

## Direct Files
<!-- doc-ledger:files:start -->

- [scoring-policy-and-awards.md](scoring-policy-and-awards.md) - Scoring Policy And Awards documentation.
<!-- doc-ledger:files:end -->

## Stub Files
<!-- doc-ledger:stubs:start -->
<!-- doc-ledger:stubs:end -->

## Direct Folders
<!-- doc-ledger:folders:start -->
<!-- doc-ledger:folders:end -->

## Related Docs

- [Game Server Simulation](../!README.md)
- [Game Server](../../!README.md)
- [Player Counters](../players/player-counters.md)
- [Collision To Damage Flow](../combat/collision-to-damage-flow.md)
- [Damage Resolution](../combat/damage-resolution.md)
- [Pickup Drop Integration](../pickups/pickup-drop-integration.md)
- [Data](../../../../data/!README.md)

## Notes

Scoring policy is pure, while score storage and mutation belong to player counters.