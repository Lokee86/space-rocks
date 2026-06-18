# Profile Stats Flow
Parent index: [Player Data](../!README.md)

## Purpose

Stub note: this document is incomplete and non-canonical.
TODO: describe profile stats reads, writes, and summary flow ownership.

## Overview

TODO: summarize how the player-data service reads, updates, and summarizes profile stats.
Stub note: keep this focused on service-owned profile stats flow.

## Code root

- `services/player-data/`

## Responsibilities

- TODO: describe profile stats read/write handling, summary generation, and any profile stats flow responsibilities.

## Does not own

- API-server auth policy.
- Game-server match simulation.
- TODO: any other boundaries that belong outside player-data profile stats ownership.

## Domain roles

- TODO: define the profile stats roles that participate in reads, writes, and summary flow.
- Stub note: do not invent role names until they are confirmed from code or design docs.

## Protocols and APIs

- TODO: describe profile stats HTTP endpoints, service calls, or summary APIs used by the player-data service.
- Stub note: this is intentionally incomplete.

## Data ownership

- TODO: identify what profile stats and summary data the player-data service owns or exposes.
- Stub note: do not assume schema or retention policy here.

## Code map

- `services/player-data/httpapi/`
- `services/player-data/httpapi/profile_handler.go`
- `services/player-data/playerdata/`
- `services/player-data/playerdata/runtime_sink.go`
- TODO: add narrower code links when the ownership story is confirmed.

## Tests

- TODO: add profile stats flow test references if they are confirmed.
- Stub note: only list verified tests here.

## Related docs

- [Player Data](../!README.md)
- TODO: add profile stats docs when they exist.

## Notes

Stub note: this document is a placeholder for future profile stats flow documentation.
Do not treat it as canonical source material.
