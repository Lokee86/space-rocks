# Networking And Sessions
Parent index: [Game Server](../!README.md)

## Purpose

Stub note: this document is incomplete and non-canonical.
TODO: describe game-server websocket, packet, session, and room session ownership.

## Overview

TODO: summarize how the game server owns websocket sessions, packet routing, admission, identity, and room session behavior.
Stub note: keep this focused on server-side networking responsibility.

## Code root

- `services/game-server/`

## Responsibilities

- TODO: describe websocket session handling, packet routing, admission checks, session identity, room session ownership, and related lifecycle coordination.

## Does not own

- Client websocket transport implementation.
- Client packet codec and inbound/outbound packet dispatch.
- TODO: any other boundaries that belong outside server networking ownership.

## Domain roles

- TODO: define the networking and session roles that participate in admission and packet handling.
- Stub note: do not invent role names until they are confirmed from code or design docs.

## Protocols and APIs

- TODO: describe websocket entry points, inbound packet routing, session admission, and identity APIs.
- Stub note: this is intentionally incomplete.

## Data ownership

- TODO: identify session identity, room session state, and any networking-owned runtime state.
- Stub note: do not assume persistence or account-system details here.

## Code map

- `services/game-server/internal/networking/`
- `services/game-server/internal/networking/websocket.go`
- `services/game-server/internal/networking/websocket_session.go`
- `services/game-server/internal/networking/client_packet_router.go`
- `services/game-server/internal/networking/session_admission.go`
- `services/game-server/internal/networking/session_identity.go`
- `services/game-server/internal/networking/session_auth.go`
- `services/game-server/internal/networking/room_sessions.go`
- `services/game-server/internal/networking/rooms.go`
- `services/game-server/internal/networking/inbound/`
- `services/game-server/internal/networking/outbound/`
- TODO: add narrower code links when the ownership story is confirmed.

## Tests

- TODO: add networking, session, routing, and admission test references.
- Stub note: only list verified tests here.

## Related docs

- [Game Server](../!README.md)
- [Documentation procedure](../../../documentation-procedure.md)
- TODO: add networking-specific docs when they exist.

## Notes

Stub note: this document is a placeholder for future networking and session documentation.
Do not treat it as canonical source material.
