# Devtools Packets
Parent index: [Protocol](../!README.md)

## Purpose

Stub note: this document is incomplete and non-canonical.
TODO: describe the devtools realtime and debug packet protocol.

## Overview

TODO: summarize how devtools packets support debug views, telemetry, and internal control flows.
Stub note: keep this focused on protocol behavior, not UI presentation.

## Participating systems

- Game server devtools networking.
- Client devtools runtime and packet handling.
- TODO: any other internal tools participants that are confirmed later.

## Authority

- TODO: describe which side owns debug state and which side displays or requests it.
- Stub note: do not invent authority details beyond confirmed debug-only flows.

## Message or request flow

- TODO: describe debug packet flow for status, telemetry, shape catalog, and control messages.
- TODO: document any request/response or stream patterns that are actually used.

## Source-of-truth files

- `shared/packets/debug.toml`
- `shared/packets/outputs.toml`
- TODO: add any other verified packet source files if they become relevant.

## Service responsibilities

- TODO: describe game-server and client responsibilities for devtools packet handling.
- Stub note: keep transport details out of this doc unless they are confirmed.

## Validation and testing

- TODO: add packet schema, codec, and debug flow test references.
- Stub note: only list verified tests or checks here.

## Related docs

- [Protocol](../!README.md)
- TODO: add related service and schema docs when they exist.

## Notes

Stub note: this document is a placeholder for future devtools packet documentation.
Do not treat it as canonical source material.
