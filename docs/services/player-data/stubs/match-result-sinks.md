# Match Result Sinks
Parent index: [Player Data](../!README.md)

## Purpose

Stub note: this document is incomplete and non-canonical.
TODO: describe match result routing from game server into durable and transient storage.

## Overview

TODO: summarize how match results arrive from the game server and are routed into storage sinks.
Stub note: keep this focused on player-data sink behavior.

## Code root

- `services/player-data/`

## Responsibilities

- TODO: describe match result routing, sink selection, and durable versus transient storage responsibilities.

## Does not own

- Game-server match simulation or reporting policy.
- API-server auth or OpenAPI enforcement.
- TODO: any other boundaries that belong outside player-data sink ownership.

## Domain roles

- TODO: define the match result sink roles that participate in routing and storage.
- Stub note: do not invent role names until they are confirmed from code or design docs.

## Protocols and APIs

- TODO: describe the match-result intake surfaces and any storage-facing APIs used by the player-data service.
- Stub note: this is intentionally incomplete.

## Data ownership

- TODO: identify what match result data is persisted, cached, or forwarded by the player-data service.
- Stub note: do not assume schema or retention policy here.

## Code map

- `services/player-data/playerdata/`
- `services/player-data/playerdata/runtime_sink.go`
- `services/player-data/playerdata/store_router.go`
- `services/player-data/playerdata/guest_memory_store.go`
- `services/player-data/playerdata/memory_store.go`
- `services/player-data/playerdata/rails_store.go`
- `services/player-data/playerdata/embeddedsqlite/sqlite_store.go`
- TODO: add narrower code links when the ownership story is confirmed.

## Tests

- TODO: add match result sink test references if they are confirmed.
- Stub note: only list verified tests here.

## Related docs

- [Player Data](../!README.md)
- [Documentation procedure](../../../documentation-procedure.md)
- TODO: add match-result-specific docs when they exist.

## Notes

Stub note: this document is a placeholder for future match result sink documentation.
Do not treat it as canonical source material.
