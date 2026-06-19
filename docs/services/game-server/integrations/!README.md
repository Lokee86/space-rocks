# Game Server Integrations

Integration documentation for the game server lives here.

## Ownership

This folder owns external-service integration points used by the game server.

## Does Not Belong

- Process startup or shutdown.
- WebSocket transport details.
- Room rules or simulation mechanics.
- Logging policy detail beyond integration-related diagnostics.
- External service internals.

## Direct Files

- None.

## Stub Files

- [auth-verifier-integration.md](stubs/auth-verifier-integration.md) - Stub: incomplete API token verification integration documentation.
- [player-data-http-hosting.md](stubs/player-data-http-hosting.md) - Stub: incomplete player-data HTTP hosting documentation.
- [match-result-reporting.md](stubs/match-result-reporting.md) - Stub: incomplete match result reporting documentation.

## Direct Folders

- None.

## Related Docs

- [Game Server](../!README.md)
- [Services index](../../!README.md)

## Notes

This boundary only covers how the game server connects outward, not the implementation of the external services themselves.
