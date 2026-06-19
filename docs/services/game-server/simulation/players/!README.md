# Game Server Simulation Players

Player simulation documentation for the game server lives here.

## Ownership

This folder owns durable player sessions, live avatar lifecycle, input/pause/suspension, counters/death/respawn, and camera/player world state documentation for the game server simulation.

## Does Not Belong

- Room membership or identity rules.
- WebSocket transport details.
- External integration internals.
- Process startup or shutdown.
- World motion or asteroid rules.

## Direct Files

- None.

## Stub Files

- [player-session-and-avatar-state.md](stubs/player-session-and-avatar-state.md) - Stub: incomplete player session and avatar state documentation.
- [input-pause-and-suspension.md](stubs/input-pause-and-suspension.md) - Stub: incomplete input pause and suspension documentation.
- [counters-death-and-respawn.md](stubs/counters-death-and-respawn.md) - Stub: incomplete counters death and respawn documentation.
- [camera-view-and-player-world-state.md](stubs/camera-view-and-player-world-state.md) - Stub: incomplete camera view and player world state documentation.

## Direct Folders

- None.

## Related Docs

- [Game Server Simulation](../!README.md)
- [Game Server](../../!README.md)
- [Services index](../../../!README.md)

## Notes

This boundary stays on player-owned simulation state and not room ownership or client UI.
