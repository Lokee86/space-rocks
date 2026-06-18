# Realtime Websocket Protocol
Parent index: [Protocol](../!README.md)

## Purpose

Stub note: this document is incomplete and non-canonical.
TODO: describe the client and game-server websocket protocol.

## Overview

TODO: summarize websocket connection, framing, and server/client messaging behavior.
Stub note: keep this focused on realtime websocket protocol flow.

## Participating systems

- Client networking.
- Game-server networking.
- TODO: any other websocket participants that are confirmed later.

## Authority

- TODO: describe which side owns realtime state and which side exchanges packets over the websocket.
- Stub note: do not invent authority details beyond the known server-authoritative direction.

## Message or request flow

- TODO: describe connect, read, write, close, and gameplay packet exchange flow.
- TODO: document any session or reconnection behavior that is actually used.

## Source-of-truth files

- `client/scripts/networking/packet_codec/`
- `client/scripts/networking/network_client.gd`
- `services/game-server/internal/protocol/packetcodec/`
- `services/game-server/internal/networking/websocket.go`
- `services/game-server/internal/networking/websocket_read.go`
- `services/game-server/internal/networking/websocket_write.go`
- TODO: add any other verified websocket source files if they become relevant.

## Service responsibilities

- TODO: describe client and game-server websocket responsibilities.
- Stub note: keep transport details out of this doc unless they are confirmed.

## Validation and testing

- TODO: add websocket protocol and packet codec test references.
- Stub note: only list verified tests or checks here.

## Related docs

- [Protocol](../!README.md)
- TODO: add websocket-specific docs when they exist.

## Notes

Stub note: this document is a placeholder for future realtime websocket protocol documentation.
Do not treat it as canonical source material.
