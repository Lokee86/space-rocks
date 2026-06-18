# Rooms And Lifecycle
Parent index: [Game Server](../!README.md)

## Purpose

Stub note: this document is incomplete and non-canonical.
TODO: describe the game server's room state and lifecycle ownership.

## Overview

TODO: summarize room creation, lobby state, in-game state, game-over handling, cleanup, and lifecycle transitions.
Stub note: keep this focused on server-owned room behavior.

## Code root

- `services/game-server/`

## Responsibilities

- TODO: describe room state transitions, room membership, lobby flow, active match flow, game-over flow, cleanup, and lifecycle orchestration.

## Does not own

- Client UI state, presentation flow, and interpolation.
- TODO: any other boundaries that belong outside room ownership.

## Domain roles

- TODO: define the room and lifecycle roles that participate in state transitions.
- Stub note: do not invent role names until they are confirmed from code or design docs.

## Protocols and APIs

- TODO: describe the room and lifecycle APIs exposed through networking or internal runtime seams.
- Stub note: this is intentionally incomplete.

## Data ownership

- TODO: identify room state, lifecycle state, and cleanup-owned data.
- Stub note: do not assume persistence or matchmaking details here.

## Code map

- `services/game-server/internal/rooms/`
- `services/game-server/internal/rooms/room.go`
- `services/game-server/internal/rooms/manager.go`
- `services/game-server/internal/rooms/lifecycle.go`
- `services/game-server/internal/rooms/lifecycle_tick.go`
- `services/game-server/internal/rooms/room_lobby.go`
- `services/game-server/internal/rooms/room_lifecycle.go`
- `services/game-server/internal/rooms/room_cleanup.go`
- TODO: add narrower code links when the ownership story is confirmed.

## Tests

- TODO: add room lifecycle, cleanup, and transition test references.
- Stub note: only list verified tests here.

## Related docs

- [Game Server](../!README.md)
- [Documentation procedure](../../../documentation-procedure.md)
- TODO: add room-specific docs when they exist.

## Notes

Stub note: this document is a placeholder for future room and lifecycle documentation.
Do not treat it as canonical source material.
