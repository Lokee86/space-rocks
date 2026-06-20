# Game Server Simulation

Parent index: [Game Server](../!INDEX.md)

Simulation documentation for the game server lives here.

## Ownership

This folder owns authoritative game runtime behavior for the game server.

## Does Not Belong

- WebSocket transport details.
- Room membership or lifecycle rules.
- External integration internals.
- Process startup or shutdown.
- Logging policy detail beyond simulation-related diagnostics.

## Direct Files
<!-- doc-ledger:files:start -->
<!-- doc-ledger:files:end -->
## Stub Files
<!-- doc-ledger:stubs:start -->
<!-- doc-ledger:stubs:end -->
## Direct Folders
<!-- doc-ledger:folders:start -->

- [combat](combat/!INDEX.md) - Combat documentation.
- [pickups](pickups/!INDEX.md) - Pickups documentation.
- [players](players/!INDEX.md) - Players documentation.
- [runtime](runtime/!INDEX.md) - Runtime documentation.
- [scoring](scoring/!INDEX.md) - Scoring documentation.
- [targeting](targeting/!INDEX.md) - Targeting documentation.
- [world](world/!INDEX.md) - World documentation.
<!-- doc-ledger:folders:end -->
## Related Docs

- [Game Server](../!INDEX.md)
- [Services index](../../!INDEX.md)

## Notes

This boundary covers the server-owned gameplay runtime and not presentation or client sync concerns.