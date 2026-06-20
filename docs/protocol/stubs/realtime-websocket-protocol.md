# Realtime WebSocket Protocol

Parent index: [Protocol](../!README.md)

## Purpose

This stub is incomplete and non-canonical. It points to the client/server realtime WebSocket protocol flow.

## Overview

This stub covers connection-level protocol expectations, JSON text packet framing, packet envelope conventions, send and receive sequencing, lobby and gameplay packet routing categories, error packet categories, session-scoped delivery assumptions, and the current absence of durable ack, retry, or reconnect semantics.

## Participating systems

- Client networking.
- Game-server networking.
- Shared packet codec code.
- TODO: any other verified transport participants.

## Authority

- The server owns the accepted connection lifecycle and session-scoped outbound delivery behavior.
- The client owns local connection initiation and inbound packet handling after packets are received.
- TODO: document any confirmed transport authority split details that are not yet captured elsewhere.

## Message or request flow

- The client connects to `/ws`.
- The WebSocket exchange carries JSON text messages.
- The packet envelope type determines which packet handler receives an inbound message.
- Client packets flow inbound to the game server through the realtime dispatch model.
- Server packets flow outbound from the game server to the connected session.
- Delivery is scoped to the active session; durable ack, retry, and reconnect semantics are not currently part of this protocol.

## Source-of-truth files

- `client/scripts/networking/packet_codec/`
- `client/scripts/networking/network_client.gd`
- `services/game-server/internal/protocol/packetcodec/`
- `services/game-server/internal/networking/websocket.go`
- `services/game-server/internal/networking/websocket_read.go`
- `services/game-server/internal/networking/websocket_write.go`
- TODO: add any other verified websocket source files if they become relevant.

## Service responsibilities

- Client networking opens and maintains the session connection and routes inbound packets to local handlers.
- Game-server networking accepts the connection, dispatches inbound packets, and writes outbound packets for the active session.
- Shared codec code defines the wire JSON handling used on both sides.
- This stub does not include game-server WebSocket implementation details, room membership implementation, player input rules, or client UI behavior.

## Validation and testing

- TODO: add websocket protocol or packet codec test references when they are confirmed.
- Stub note: only list verified tests or checks here.

## Related docs

- [Protocol](../!README.md)
- [Game Server Networking WebSocket Session Lifecycle](../../services/game-server/networking/websocket-session-lifecycle.md)
- [Gameplay Network Adapter](../../services/game-server/networking/gameplay-network-adapter.md)
- [Lobby Packets](lobby-packets.md)
- [Gameplay Packets](gameplay-packets.md)

## Notes

This is a protocol stub, not the canonical transport authority.
It does not cover WebSocket handler implementation, room membership implementation, gameplay simulation implementation, or client UI behavior.
