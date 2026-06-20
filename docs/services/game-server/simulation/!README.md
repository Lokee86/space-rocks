# Game Server Simulation

Parent index: [Game Server](../!README.md)

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

- [Combat](combat/!README.md) - Game-server combat boundary documentation index.
- [Pickups](pickups/!README.md) - Game-server pickup boundary documentation index.
- [Players](players/!README.md) - Game-server player boundary documentation index.
- [Runtime](runtime/!README.md) - Game-server runtime boundary documentation index.
- [scoring](scoring/!README.md) - Scoring documentation.
- [Targeting](targeting/!README.md) - Game-server targeting boundary documentation index.
- [World](world/!README.md) - Game-server world boundary documentation index.
<!-- doc-ledger:folders:end -->
## Related Docs

- [Game Server](../!README.md)
- [Services index](../../!README.md)

## Notes

This boundary covers the server-owned gameplay runtime and not presentation or client sync concerns.