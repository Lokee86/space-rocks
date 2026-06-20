# Game Server Networking

Parent index: [Game Server](../!INDEX.md)

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

- [auth-routing.md](auth-routing.md) - Game-server auth packet routing documentation.
- [gameplay-network-adapter.md](gameplay-network-adapter.md) - Game-server gameplay packet adapter documentation.
- [inbound-packet-routing.md](inbound-packet-routing.md) - Game-server inbound packet routing documentation.
- [outbound-message-flow.md](outbound-message-flow.md) - Game-server outbound message flow documentation.
- [room-network-adapter.md](room-network-adapter.md) - Game-server room network adapter documentation.
- [telemetry-packet-routing.md](telemetry-packet-routing.md) - Game-server telemetry packet routing documentation.
- [websocket-session-lifecycle.md](websocket-session-lifecycle.md) - Game-server WebSocket session lifecycle documentation.
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

This boundary focuses on transport and routing responsibilities, not gameplay authority.