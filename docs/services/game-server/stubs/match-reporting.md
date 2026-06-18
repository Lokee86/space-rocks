# Match Reporting
Parent index: [Game Server](../!README.md)

## Purpose

Stub note: this document is incomplete and non-canonical.
TODO: describe game-server match summary and result reporting ownership.

## Overview

TODO: summarize how runtime match outcomes are reported or mapped for downstream use.
Stub note: keep this focused on server-side reporting responsibility.

## Code root

- `services/game-server/`

## Responsibilities

- TODO: describe match summary mapping, runtime result reporting, and any outbound reporting responsibilities owned by the game server.

## Does not own

- Player-facing scoreboard UI.
- External analytics or persistence policy.
- TODO: any other boundaries that belong outside reporting ownership.

## Domain roles

- TODO: define the reporting roles that participate in runtime match summary handling.
- Stub note: do not invent role names until they are confirmed from code or design docs.

## Protocols and APIs

- TODO: describe any packet, event, or internal adapter surfaces used for match reporting.
- Stub note: this is intentionally incomplete.

## Data ownership

- TODO: identify what match result or summary data the game server owns and what it merely forwards.
- Stub note: do not assume storage or analytics details here.

## Code map

- `services/game-server/internal/matchreporting/`
- `services/game-server/internal/matchreporting/mapper.go`
- `services/game-server/internal/matchreporting/runtime_reporter.go`
- TODO: add narrower code links when the reporting story is confirmed.

## Tests

- `services/game-server/internal/matchreporting/mapper_test.go`
- `services/game-server/internal/matchreporting/runtime_reporter_test.go`
- TODO: add any additional verified tests here.

## Related docs

- [Game Server](../!README.md)
- [Documentation procedure](../../../documentation-procedure.md)
- TODO: add reporting-specific docs when they exist.

## Notes

Stub note: this document is a placeholder for future match reporting documentation.
Do not treat it as canonical source material.
