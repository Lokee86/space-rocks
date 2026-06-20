# Game Server Simulation Players

Parent index: [Game Server Simulation](../!INDEX.md)

Player simulation documentation for the game server lives here.

## Ownership

This folder owns simulation-local player session state, active player avatar state, player input routing, pause/suspension gates, player counters, death/despawn, respawn, and per-player camera-view state for the game-server simulation.

## Does Not Belong

- Room membership or identity rules.
- WebSocket transport details.
- External integration internals.
- Process startup or shutdown.
- World motion or asteroid rules.

## Direct Files
<!-- doc-ledger:files:start -->

- [active-player-avatar-state.md](active-player-avatar-state.md) - Active player avatar state documentation.
- [player-camera-view-state.md](player-camera-view-state.md) - Player camera view state documentation.
- [player-counters.md](player-counters.md) - Player counters documentation.
- [player-death-and-despawn.md](player-death-and-despawn.md) - Player death and despawn documentation.
- [player-input-routing.md](player-input-routing.md) - Player input routing documentation.
- [player-lifecycle-overview.md](player-lifecycle-overview.md) - Player lifecycle overview documentation.
- [player-pause-and-suspension.md](player-pause-and-suspension.md) - Player pause and suspension documentation.
- [player-respawn.md](player-respawn.md) - Player respawn documentation.
- [player-session-state.md](player-session-state.md) - Player session state documentation.
<!-- doc-ledger:files:end -->
## Stub Files
<!-- doc-ledger:stubs:start -->
<!-- doc-ledger:stubs:end -->
## Direct Folders
<!-- doc-ledger:folders:start -->
<!-- doc-ledger:folders:end -->
## Related Docs

- [Game Server Simulation](../!INDEX.md)
- [Game Server](../../!INDEX.md)
- [Services index](../../../!INDEX.md)

## Notes

This boundary stays on player-owned simulation state and not room ownership or client UI.