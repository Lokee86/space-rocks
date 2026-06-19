# Game Server Integrations

Parent index: [Game Server](../!README.md)

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
<!-- doc-ledger:files:end -->
## Stub Files
<!-- doc-ledger:stubs:start -->

- [auth-verifier-integration.md](stubs/auth-verifier-integration.md) - Stub: incomplete API token verification integration documentation.
- [match-result-reporting.md](stubs/match-result-reporting.md) - Stub: incomplete match result reporting documentation.
- [player-data-http-hosting.md](stubs/player-data-http-hosting.md) - Stub: incomplete player-data HTTP hosting documentation.
<!-- doc-ledger:stubs:end -->
## Direct Folders
<!-- doc-ledger:folders:start -->
<!-- doc-ledger:folders:end -->
## Related Docs

- [Game Server](../!README.md)
- [Services index](../../!README.md)

## Notes

This boundary only covers how the game server connects outward, not the implementation of the external services themselves.