# Authoritative Simulation
Parent index: [Game Server](../!README.md)

## Purpose

Stub note: this document is incomplete and non-canonical.
TODO: describe the game server's authoritative simulation ownership.

## Overview

TODO: summarize how the game server owns simulation outcomes.
Stub note: keep this focused on runtime authority, not client presentation.

## Code root

- `services/game-server/`

## Responsibilities

- TODO: describe authoritative movement, bullets, collisions, scoring, lives, death, respawn, pause safety, and simulation state updates.

## Does not own

- Client rendering, UI, audio, and interpolation.
- TODO: any other boundaries that belong outside the simulation owner.

## Domain roles

- TODO: define the simulation roles that participate in authoritative updates.
- Stub note: do not invent role names until they are confirmed from code or design docs.

## Protocols and APIs

- TODO: describe the inbound and outbound packet surfaces used by simulation ownership.
- Stub note: this is intentionally incomplete.

## Data ownership

- TODO: identify simulation-owned state and durable state boundaries.
- Stub note: do not assume persistence details here.

## Code map

- `services/game-server/internal/game/`
- `services/game-server/internal/game/motion/`
- `services/game-server/internal/game/combat.go`
- `services/game-server/internal/game/session.go`
- `services/game-server/internal/game/match.go`
- `services/game-server/internal/game/rules/`
- `services/game-server/internal/game/spawning.go`
- `services/game-server/internal/game/scoring.go`
- `services/game-server/internal/game/runtime/`
- TODO: add narrower code links when the ownership story is confirmed.

## Tests

- TODO: add authoritative simulation tests and smoke coverage references.
- Stub note: only list verified tests here.

## Related docs

- [Game Server](../!README.md)
- TODO: add simulation-specific docs when they exist.

## Notes

Stub note: this document is a placeholder for future authoritative simulation documentation.
Do not treat it as canonical source material.
