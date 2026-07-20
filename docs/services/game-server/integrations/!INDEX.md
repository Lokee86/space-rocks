---
author: brian
created: "2026-07-19"
document_id: 019f7d55-fb2c-76a1-b162-19158a904bfd
document_type: general
policy_exempt: false
summary: Integration documentation for the game server lives here.
---
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
- [diagnostic-aggregator-hosting.md](diagnostic-aggregator-hosting.md) - Temporary game-server process co-hosting and dependency boundaries for diagnostic-aggregator.
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