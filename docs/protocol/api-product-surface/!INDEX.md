# Api Product Surface

This folder indexes the current API product-surface protocol docs for Space Rocks.
It covers HTTP/API communication surfaces, request/response shape ownership, service responsibility boundaries, and API-facing lifecycle and compatibility expectations.

Parent index: [Protocol](../!INDEX.md)

## Ownership

This folder owns current HTTP/API product-surface protocol docs.
It documents which services own each current API-facing surface, how requests and responses are shaped, and which lifecycle or compatibility rules apply at the protocol boundary.

## Does Not Belong

This folder does not own:

- planning API surface docs
- service implementation detail that belongs under `docs/services/`
- data pipeline or source-generation detail that belongs under `docs/data/`
- broad domain or product strategy that belongs under `docs/domains/` or `docs/planning/`
- realtime WebSocket packet details that belong in the realtime protocol docs

## Direct Files
<!-- doc-ledger:files:start -->

- [http-api-contracts.md](http-api-contracts.md) - HTTP API contract documentation.
- [player-data-http-api.md](player-data-http-api.md) - Player-data HTTP API documentation.
<!-- doc-ledger:files:end -->

## Stub Files
<!-- doc-ledger:stubs:start -->
<!-- doc-ledger:stubs:end -->

## Direct Folders
<!-- doc-ledger:folders:start -->
<!-- doc-ledger:folders:end -->

## Related Docs

- [Protocol](../!INDEX.md)
- [HTTP Contract Enforcement](../http-contract-enforcement.md)
- [API Server](../../services/api-server/!INDEX.md)
- [Game Server](../../services/game-server/!INDEX.md)
- [Player Data](../../services/player-data/!INDEX.md)
- [Client HTTP API Flow](../../services/client/client-http-api-flow.md)
- [Data](../../data/!INDEX.md)

## Notes

This folder is for current protocol and API product-surface docs. Future or unresolved API surface work belongs under `docs/planning/`.