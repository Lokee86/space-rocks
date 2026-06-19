# Game Server Rooms

Rooms documentation for the game server lives here.

## Ownership

This folder owns room state, membership, lobby/start rules, match lifecycle, cleanup, and snapshot projection documentation for the game server.

## Does Not Belong

- WebSocket transport details.
- Simulation mechanics.
- External integration internals.
- Process startup or shutdown.
- Logging policy detail beyond room-related diagnostics.

## Direct Files

- None.

## Stub Files

- [room-manager.md](stubs/room-manager.md) - Stub: incomplete room manager documentation.
- [room-membership-and-identity.md](stubs/room-membership-and-identity.md) - Stub: incomplete room membership and identity documentation.
- [lobby-and-start-rules.md](stubs/lobby-and-start-rules.md) - Stub: incomplete lobby and start rules documentation.
- [room-match-lifecycle.md](stubs/room-match-lifecycle.md) - Stub: incomplete room match lifecycle documentation.
- [room-cleanup.md](stubs/room-cleanup.md) - Stub: incomplete room cleanup documentation.
- [room-snapshot-projection.md](stubs/room-snapshot-projection.md) - Stub: incomplete room snapshot projection documentation.

## Direct Folders

- None.

## Related Docs

- [Game Server](../!README.md)
- [Services index](../../!README.md)

## Notes

This boundary stays on room ownership and room-facing lifecycle behavior.
