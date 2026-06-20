# Game Server Rooms

Parent index: [Game Server](../!INDEX.md)

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
<!-- doc-ledger:files:start -->

- [lobby-and-start-rules.md](lobby-and-start-rules.md) - Lobby admission, ready state, and start rules documentation.
- [room-cleanup.md](room-cleanup.md) - Room cleanup, empty-room cleanup, and cleanup timer/version behavior documentation.
- [room-manager.md](room-manager.md) - Room manager registry and lifecycle entry points documentation.
- [room-match-lifecycle.md](room-match-lifecycle.md) - Room Match Lifecycle documentation.
- [room-membership-and-identity.md](room-membership-and-identity.md) - Room Membership And Identity documentation.
- [room-snapshot-projection.md](room-snapshot-projection.md) - Room Snapshot Projection documentation.
<!-- doc-ledger:files:end -->

## Stub Files
<!-- doc-ledger:stubs:start -->
<!-- doc-ledger:stubs:end -->
## Direct Folders
<!-- doc-ledger:folders:start -->
<!-- doc-ledger:folders:end -->
## Related Docs

- [Game Server](../!INDEX.md)
- [Services index](../../!INDEX.md)

## Notes

This boundary stays on room ownership and room-facing lifecycle behavior.