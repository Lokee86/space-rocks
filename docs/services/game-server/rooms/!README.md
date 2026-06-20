# Game Server Rooms

Parent index: [Game Server](../!README.md)

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

- [lobby-and-start-rules.md](lobby-and-start-rules.md) - Incomplete lobby and start rules documentation.
- [room-cleanup.md](room-cleanup.md) - Incomplete room cleanup documentation.
- [room-manager.md](room-manager.md) - Incomplete room manager documentation.
- [room-match-lifecycle.md](room-match-lifecycle.md) - Room Match Lifecycle documentation.
- [room-membership-and-identity.md](room-membership-and-identity.md) - Room Membership And Identity documentation.
- [room-snapshot-projection.md](room-snapshot-projection.md) - Room Snapshot Projection documentation.
<!-- doc-ledger:files:end -->## Stub Filesroom-membership-and-identity.md
<!-- doc-ledger:stubs:start -->
<!-- doc-ledger:stubs:end -->
## Direct Folders
<!-- doc-ledger:folders:start -->
<!-- doc-ledger:folders:end -->
## Related Docs

- [Game Server](../!README.md)
- [Services index](../../!README.md)

## Notes

This boundary stays on room ownership and room-facing lifecycle behavior.