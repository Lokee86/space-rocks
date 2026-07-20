---
author: brian
created: "2026-07-19"
document_id: 019f7d55-fb2c-7909-9da8-9ef434007d10
document_type: general
policy_exempt: false
summary: Protocol documentation covers communication and message-flow behavior.
---
# Protocol

Parent index: [Documentation](../!INDEX.md)

Protocol documentation covers communication and message-flow behavior.

## Ownership

This folder owns protocol documentation indexes for current Space Rocks communication flows.

## Does Not Belong

- Domain flow docs.
- Service implementation docs.
- Planning docs.
- Stub content as canonical protocol authority.
- Direct code maps unless the protocol doc is explicitly covering implementation paths.

## Direct Files
<!-- doc-ledger:files:start -->

- [api-product-surface.md](api-product-surface.md) - API product surface documentation.
- [asteroid-variant-contract.md](asteroid-variant-contract.md) - Asteroid Variant Contract documentation.
- [gameplay-packets.md](gameplay-packets.md) - Gameplay packet ownership and family overview.
- [http-api-contracts.md](http-api-contracts.md) - HTTP API contract documentation.
- [http-contract-enforcement.md](http-contract-enforcement.md) - HTTP request/response contract ownership and enforcement across services.
- [lobby-packets.md](lobby-packets.md) - Lobby packet documentation.
- [player-data-http-api.md](player-data-http-api.md) - Player-data HTTP API documentation.
- [realtime-webrtc-gameplay-transport.md](realtime-webrtc-gameplay-transport.md) - Realtime WebRTC gameplay transport documentation.
- [realtime-websocket-protocol.md](realtime-websocket-protocol.md) - Client websocket and realtime gameplay routing split across dispatcher classification, non-realtime service signals, and direct realtime packet pipeline application.
<!-- doc-ledger:files:end -->
## Stub Files
<!-- doc-ledger:stubs:start -->
<!-- doc-ledger:stubs:end -->
## Direct Folders
<!-- doc-ledger:folders:start -->

- [generated](generated/!INDEX.md) - Generated documentation.
<!-- doc-ledger:folders:end -->
## Related Docs

- [API Product Surface](api-product-surface.md)


## Notes

Protocol docs stay focused on who communicates, what is exchanged, and how the message flow works.