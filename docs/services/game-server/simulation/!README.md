# Game Server Simulation

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

- None.

## Stub Files

- None.

## Direct Folders

- [Runtime](runtime/!README.md) - Game-server runtime boundary documentation index.
- [Players](players/!README.md) - Game-server player boundary documentation index.
- [World](world/!README.md) - Game-server world boundary documentation index.
- [Targeting](targeting/!README.md) - Game-server targeting boundary documentation index.
- [Combat](combat/!README.md) - Game-server combat boundary documentation index.
- [Pickups](pickups/!README.md) - Game-server pickup boundary documentation index.

## Related Docs

- [Game Server](../!README.md)
- [Services index](../../!README.md)

## Notes

This boundary covers the server-owned gameplay runtime and not presentation or client sync concerns.
