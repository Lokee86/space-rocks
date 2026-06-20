# Lobby Packets

Parent index: [Protocol](../!README.md)

## Purpose

This stub is incomplete and non-canonical. It points to lobby and room realtime packet behavior between the client and game server.

## Overview

This stub covers the lobby and room packet families at the protocol boundary: create room, join room, leave room, ready toggle, start game, return to lobby, room snapshot, room error, and room state change or update packets.

## Participating systems

- Client lobby networking.
- Client gameplay networking when lobby actions transition into gameplay.
- Game-server room networking.
- Shared lobby packet schema and code generation.
- TODO: any additional lobby packet participants that are verified later.

## Authority

- The protocol boundary owns packet shape and packet semantics for the lobby and room packet families.
- The game server owns authoritative room-state packet contents for room snapshots, room state changes, and room errors.
- The client owns local packet emission for lobby and room requests after user intent is collected.
- TODO: add any confirmed authority split details that are not yet documented elsewhere.

## Message or request flow

- Create room and join room requests flow from client to game server.
- Leave room, ready toggle, start game, and return to lobby requests flow from client to game server.
- Room snapshot, room state changed, and room error packets flow from game server back to the client.
- TODO: add any packet-family sequencing details that remain genuinely unknown.

## Source-of-truth files

- `shared/packets/lobby.toml`
- `shared/packets/outputs.toml`
- `client/scripts/generated/networking/packets/packets.gd`
- `services/game-server/internal/game/packets.go`
- `services/game-server/internal/game/runtime/packets_generated.go`
- TODO: add any other verified packet source files if they become relevant.

## Service responsibilities

- Client lobby networking routes lobby and room requests.
- Game-server networking validates room packet handling and updates room-state output.
- Shared schema files define the packet shapes and generated packet outputs used by both sides.
- This stub does not include room manager implementation, room membership implementation, WebSocket lifecycle implementation, or client UI behavior.

## Validation and testing

- TODO: add packet schema or client/server checks if they become confirmed and relevant.

## Related docs

- [Protocol](../!README.md)
- [Game Server Room Membership And Identity](../../services/game-server/rooms/room-membership-and-identity.md)
- [Room Snapshot Projection](../../services/game-server/rooms/room-snapshot-projection.md)
- [Lobby And Start Rules](../../services/game-server/rooms/lobby-and-start-rules.md)
- [WebSocket Session Lifecycle](../../services/game-server/networking/websocket-session-lifecycle.md)

## Notes

This is a protocol stub, not the canonical packet schema source.
It does not cover room manager implementation, room membership storage, WebSocket lifecycle implementation, or client UI behavior.
