# Game Server Networking

Parent index: [Game Server](../!README.md)

Networking documentation for the game server lives here.

## Ownership

This folder owns WebSocket/session transport and packet routing adapter documentation for the game server.

## Does Not Belong

- Process startup or shutdown.
- Room state or room rules.
- Simulation authority or world rules.
- External integration internals.
- Logging policy detail beyond networking-adjacent concerns.

## Direct Files
<!-- doc-ledger:files:start -->
<!-- doc-ledger:files:end -->
## Stub Files
<!-- doc-ledger:stubs:start -->

- [auth-and-telemetry-packet-routing.md](stubs/auth-and-telemetry-packet-routing.md) - Stub: incomplete auth and telemetry packet routing documentation.
- [gameplay-network-adapter.md](stubs/gameplay-network-adapter.md) - Stub: incomplete gameplay network adapter documentation.
- [inbound-packet-routing.md](stubs/inbound-packet-routing.md) - Stub: incomplete inbound packet routing documentation.
- [outbound-message-flow.md](stubs/outbound-message-flow.md) - Stub: incomplete outbound message flow documentation.
- [room-network-adapter.md](stubs/room-network-adapter.md) - Stub: incomplete room network adapter documentation.
- [websocket-session-lifecycle.md](stubs/websocket-session-lifecycle.md) - Stub: incomplete websocket session lifecycle documentation.
<!-- doc-ledger:stubs:end -->
## Direct Folders
<!-- doc-ledger:folders:start -->
<!-- doc-ledger:folders:end -->
## Related Docs

- [Game Server](../!README.md)
- [Services index](../../!README.md)

## Notes

This boundary focuses on transport and routing responsibilities, not gameplay authority.