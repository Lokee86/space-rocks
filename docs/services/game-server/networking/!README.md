# Game Server Networking

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

- None.

## Stub Files

- [websocket-session-lifecycle.md](stubs/websocket-session-lifecycle.md) - Stub: incomplete websocket session lifecycle documentation.
- [inbound-packet-routing.md](stubs/inbound-packet-routing.md) - Stub: incomplete inbound packet routing documentation.
- [outbound-message-flow.md](stubs/outbound-message-flow.md) - Stub: incomplete outbound message flow documentation.
- [room-network-adapter.md](stubs/room-network-adapter.md) - Stub: incomplete room network adapter documentation.
- [gameplay-network-adapter.md](stubs/gameplay-network-adapter.md) - Stub: incomplete gameplay network adapter documentation.
- [auth-and-telemetry-packet-routing.md](stubs/auth-and-telemetry-packet-routing.md) - Stub: incomplete auth and telemetry packet routing documentation.

## Direct Folders

- None.

## Related Docs

- [Game Server](../!README.md)
- [Services index](../../!README.md)

## Notes

This boundary focuses on transport and routing responsibilities, not gameplay authority.
