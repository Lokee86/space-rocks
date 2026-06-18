# Player Data Integration
Parent index: [Game Server](../!README.md)

## Purpose

Stub note: this document is incomplete and non-canonical.
TODO: describe the game server's integration with the player-data service.

## Overview

TODO: summarize how the game server consumes player-data information and how that integration affects runtime behavior.
Stub note: keep this focused on service integration rather than player-data ownership itself.

## Code root

- `services/game-server/`

## Responsibilities

- TODO: describe player-data lookup, summary resolution, and any runtime integration points that depend on player-data service responses.

## Does not own

- Player-data service storage and persistence policy.
- Account system ownership.
- TODO: any other boundaries that belong outside game-server integration.

## Domain roles

- TODO: define the integration roles that read or surface player-data information.
- Stub note: do not invent role names until they are confirmed from code or design docs.

## Protocols and APIs

- TODO: describe the player-data integration APIs, request shapes, or internal adapters used by the game server.
- Stub note: this is intentionally incomplete.

## Data ownership

- TODO: identify what player-data fields the game server reads, caches, or forwards.
- Stub note: do not assume persistence or schema details here.

## Code map

- `services/game-server/internal/playerdata/`
- `services/game-server/internal/playerdata/resolve.go`
- `services/game-server/internal/playerdata/summary.go`
- `services/game-server/internal/authclient/`
- `services/game-server/internal/authclient/client.go`
- `services/game-server/internal/authclient/types.go`
- TODO: add narrower code links when the integration story is confirmed.

## Tests

- TODO: add player-data integration and resolver test references.
- Stub note: only list verified tests here.

## Related docs

- [Game Server](../!README.md)
- [Documentation procedure](../../../documentation-procedure.md)
- TODO: add player-data-specific docs when they exist.

## Notes

Stub note: this document is a placeholder for future player-data integration documentation.
Do not treat it as canonical source material.
