# Game Server Integrations

Parent index: [Game Server](../!INDEX.md)

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
<!-- doc-ledger:files:start -->

- [auth-verifier-integration.md](auth-verifier-integration.md) - Game-server API token verification integration documentation.
- [match-result-reporting.md](match-result-reporting.md) - Game-server match result reporting documentation.
- [player-data-http-hosting.md](player-data-http-hosting.md) - Game-server player-data HTTP hosting documentation.
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

This boundary only covers how the game server connects outward, not the implementation of the external services themselves.