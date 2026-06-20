# Gameplay Packets

Parent index: [Protocol](../!README.md)

## Purpose

This stub is incomplete and non-canonical. It points to gameplay realtime packet behavior between the client and game server.

## Overview

This stub covers the gameplay packet families at the protocol boundary: input packets, client config packets, pause request packets, respawn request packets, gameplay `state` packets, and gameplay event or presentation packets.

## Participating systems

- Client gameplay networking.
- Game-server gameplay networking.
- Shared packet schema and code generation.
- TODO: any additional gameplay packet participants that are verified later.

## Authority

- The protocol boundary owns packet shape and packet semantics for the gameplay packet families.
- The game server owns authoritative gameplay packet contents for server-to-client state and event output.
- The client owns local packet emission for gameplay requests after input is collected.

## Message or request flow

- Client input, config, pause, and respawn packets flow from client to game server.
- The game server projects the gameplay `state` packet back to the client.
- The game server emits gameplay event and presentation packets back to the client.

## Source-of-truth files

- `shared/packets/gameplay.toml`
- `shared/packets/outputs.toml`
- `client/scripts/generated/networking/packets/packets.gd`
- `services/game-server/internal/game/packets.go`
- `services/game-server/internal/game/runtime/packets_generated.go`

## Service responsibilities

- Client gameplay networking routes gameplay requests and packet updates.
- Game-server networking validates gameplay packet handling and updates gameplay state output.
- Shared schema files define the packet shapes and generated packet outputs used by both sides.
- This stub does not include WebSocket lifecycle, room/lobby packet behavior, or client presentation details.

## Validation and testing

- `services/game-server/internal/networking/gameplay_packets_test.go`

## Related docs

- [Protocol](../!README.md)
- [Game Server Networking Gameplay Adapter](../../services/game-server/networking/gameplay-network-adapter.md)
- [Runtime State Packet Projection](../../services/game-server/simulation/runtime/stubs/state-packet-projection.md)
- [Presentation Event Queue](../../services/game-server/simulation/runtime/stubs/presentation-event-queue.md)
- [Player Input Routing](../../services/game-server/simulation/players/stubs/player-input-routing.md)
- [Player Pause And Suspension](../../services/game-server/simulation/players/stubs/player-pause-and-suspension.md)
- [Player Respawn](../../services/game-server/simulation/players/stubs/player-respawn.md)

## Notes

This is a protocol stub, not the canonical packet schema source.
It does not cover game-server implementation details, player session mutation, respawn mechanics, pause gate implementation, or client presentation behavior.
