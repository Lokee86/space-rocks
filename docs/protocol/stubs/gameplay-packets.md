# Gameplay Packets
Parent index: [Protocol](../!README.md)

## Purpose

Stub note: this document is incomplete and non-canonical.
TODO: describe gameplay realtime packet documentation.

## Overview

TODO: summarize the gameplay packet flow between client and game server.
Stub note: keep this focused on gameplay packet behavior, not UI presentation.

## Participating systems

- Client gameplay networking.
- Game-server gameplay networking.
- TODO: any other packet participants that are confirmed later.

## Authority

- TODO: describe which side owns gameplay state and which side transports gameplay packets.
- Stub note: do not invent authority details beyond the known server-authoritative direction.

## Message or request flow

- TODO: describe gameplay input, state update, respawn, pause, and other realtime packet flow.
- TODO: document any request/response or publish/update patterns that are actually used.

## Source-of-truth files

- `shared/packets/gameplay.toml`
- `shared/packets/outputs.toml`
- `client/scripts/generated/networking/packets/packets.gd`
- `services/game-server/internal/game/packets.go`
- `services/game-server/internal/game/runtime/packets_generated.go`
- `services/game-server/internal/networking/gameplay_packets_test.go`
- TODO: add any other verified packet source files if they become relevant.

## Service responsibilities

- TODO: describe client and game-server responsibilities for gameplay packet handling.
- Stub note: keep transport details out of this doc unless they are confirmed.

## Validation and testing

- `services/game-server/internal/networking/gameplay_packets_test.go`
- TODO: add packet schema, codec, and client-side test references if they are confirmed.
- Stub note: only list verified tests or checks here.

## Related docs

- [Protocol](../!README.md)
- TODO: add gameplay-specific docs when they exist.

## Notes

Stub note: this document is a placeholder for future gameplay packet documentation.
Do not treat it as canonical source material.
