# Networking Flow

Parent index: [Client](../!INDEX.md)

## Purpose

This folder owns client networking-flow documentation.

## Ownership

This folder owns client WebSocket connection, packet classification, packet dispatch, inbound coordinator seam, and outbound packet sending docs.

## Does Not Belong

- Server networking implementation.
- Packet schema authority.
- Gameplay runtime ownership.
- UI or presentation-only flow docs.
- Future planning.

## Direct Files
<!-- doc-ledger:files:start -->

- [inbound-packet-routing.md](inbound-packet-routing.md) - Client inbound packet routing documentation.
- [outbound-packet-sending.md](outbound-packet-sending.md) - Client outbound packet sending documentation.
- [websocket-connection-lifecycle.md](websocket-connection-lifecycle.md) - Client WebSocket connection lifecycle documentation.
<!-- doc-ledger:files:end -->
## Stub Files
<!-- doc-ledger:stubs:start -->
<!-- doc-ledger:stubs:end -->
## Direct Folders
<!-- doc-ledger:folders:start -->
<!-- doc-ledger:folders:end -->
## Related Docs

- [Client](../!INDEX.md)
- [Gameplay Runtime](../gameplay-runtime/!INDEX.md)
- [Presentation Bridge](../gameplay-runtime/presentation-bridge.md)
- [Gameplay state application](../gameplay-runtime/gameplay-state-application.md)
- [Input And Targeting](../input-and-targeting.md)

## Notes

This index stays at the client networking-flow boundary and does not expand into packet schema ownership or server transport behavior. `ServerPacketDispatcher` owns packet classification and typed signal emission. `ClientConnectionService` owns application-facing non-realtime dispatcher bindings and public facade signals. `ClientInboundCoordinator` owns only the five WebRTC control dispatcher bindings: answer, ICE, ready, smoke, and failure. `RealtimePacketPipeline` owns gameplay lane dispatcher bindings, realtime packet application, and protocol state. This index only tracks the network seam that feeds downstream session and gameplay owners.