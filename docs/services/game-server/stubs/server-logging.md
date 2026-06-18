# Server Logging
Parent index: [Game Server](../!README.md)

## Purpose

Stub note: this document is incomplete and non-canonical.
TODO: describe the game-server logging responsibility.

## Overview

TODO: summarize what the game server logs and how logging fits runtime and debugging needs.
Stub note: keep this focused on server-side logging responsibility, not observability policy.

## Code root

- `services/game-server/`

## Responsibilities

- TODO: describe runtime logging, error logging, and any structured logging responsibilities owned by the game server.

## Does not own

- Client presentation or HUD messaging.
- Central logging policy or infrastructure.
- TODO: any other boundaries that belong outside game-server logging ownership.

## Domain roles

- TODO: define the logging roles that emit or consume log context.
- Stub note: do not invent role names until they are confirmed from code or design docs.

## Protocols and APIs

- TODO: describe any logger interfaces, log sinks, or structured log surfaces used by the game server.
- Stub note: this is intentionally incomplete.

## Data ownership

- TODO: identify what log context or structured fields the game server owns.
- Stub note: do not assume retention or transport details here.

## Code map

- `services/game-server/internal/logging/`
- `services/game-server/internal/logging/logger.go`
- `services/game-server/internal/networking/websocket_close_logging.go`
- TODO: add narrower code links when the logging story is confirmed.

## Tests

- TODO: add logging test references if and when they are confirmed.
- Stub note: only list verified tests here.

## Related docs

- [Game Server](../!README.md)
- TODO: add logging-specific docs when they exist.

## Notes

Stub note: this document is a placeholder for future server logging documentation.
Do not treat it as canonical source material.
