# Game Server Simulation Runtime

Parent index: [Game Server Simulation](../!README.md)

Runtime documentation for the game server simulation lives here.

## Ownership

This folder owns Game aggregate, simulation loop/phase order, entity store, state packet projection, and presentation event queue documentation for the game server simulation.

## Does Not Belong

- Room membership or lifecycle rules.
- WebSocket transport details.
- External integration internals.
- Process startup or shutdown.
- Player session or avatar ownership.

## Direct Files
<!-- doc-ledger:files:start -->

- [game-aggregate.md](game-aggregate.md) - Game aggregate documentation.
- [presentation-event-queue.md](presentation-event-queue.md) - Presentation event queue documentation.
- [runtime-entity-store.md](runtime-entity-store.md) - Runtime Entity Store documentation.
- [simulation-loop-and-phase-order.md](simulation-loop-and-phase-order.md) - Simulation loop and phase order documentation.
- [state-packet-projection.md](state-packet-projection.md) - State packet projection documentation.
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
- [Services index](../../../!README.md)

## Notes

This boundary focuses on the simulation runtime shell rather than individual gameplay subsystems.