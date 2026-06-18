# Lobby Packets
Parent index: [Protocol](../!README.md)

## Purpose

Stub note: this document is incomplete and non-canonical.
TODO: describe the lobby and room realtime packet protocol.

## Overview

TODO: summarize lobby packet flow for room joining, room state updates, and lobby-to-room transitions.
Stub note: keep this focused on protocol behavior, not implementation details.

## Participating systems

- Game server networking and room runtime.
- Client lobby and gameplay networking.
- TODO: any other packet participants that are confirmed later.

## Authority

- TODO: describe which side owns room state and which side renders or requests updates.
- Stub note: do not invent authority details beyond known server-authoritative direction.

## Message or request flow

- TODO: describe lobby join, leave, room snapshot, and state update packet flow.
- TODO: document any request/response or publish/update patterns that are actually used.

## Source-of-truth files

- `shared/packets/lobby.toml`
- `shared/packets/outputs.toml`
- TODO: add any other verified packet source files if they become relevant.

## Service responsibilities

- TODO: describe game-server and client responsibilities for lobby packet handling.
- Stub note: keep transport details out of this doc unless they are confirmed.

## Validation and testing

- TODO: add packet schema, codec, and round-trip test references.
- Stub note: only list verified tests or checks here.

## Related docs

- [Protocol](../!README.md)
- [Documentation procedure](../../documentation-procedure.md)
- TODO: add related service and schema docs when they exist.

## Notes

Stub note: this document is a placeholder for future lobby packet documentation.
Do not treat it as canonical source material.
