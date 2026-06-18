# Runtime And Store Routing
Parent index: [Player Data](../!README.md)

## Purpose

Stub note: this document is incomplete and non-canonical.
TODO: describe player-data runtime, mode policy, memory/sqlite/postgres routing ownership.

## Overview

TODO: summarize how the player-data service selects and routes runtime store behavior across memory, sqlite, and postgres modes.
Stub note: keep this focused on service runtime and store routing.

## Code root

- `services/player-data/`

## Responsibilities

- TODO: describe runtime construction, mode policy, store routing, and backend selection responsibilities.

## Does not own

- API-server auth policy.
- Game-server simulation authority.
- TODO: any other boundaries that belong outside player-data runtime ownership.

## Domain roles

- TODO: define the runtime and storage roles that participate in player-data mode selection.
- Stub note: do not invent role names until they are confirmed from code or design docs.

## Protocols and APIs

- TODO: describe runtime configuration surfaces and any store-routing APIs used by the player-data service.
- Stub note: this is intentionally incomplete.

## Data ownership

- TODO: identify what player data the runtime owns, routes, or stores in memory, sqlite, or postgres.
- Stub note: do not assume schema or persistence policy here.

## Code map

- `services/player-data/playerdata/`
- `services/player-data/playerdata/runtime.go`
- `services/player-data/playerdata/configured_runtime.go`
- `services/player-data/playerdata/default_runtime.go`
- `services/player-data/playerdata/mode_policy.go`
- `services/player-data/playerdata/store_router.go`
- `services/player-data/playerdata/memory_store.go`
- `services/player-data/playerdata/rails_store.go`
- `services/player-data/playerdata/noop_store.go`
- `services/player-data/playerdata/embeddedsqlite/sqlite_store.go`
- TODO: add narrower code links when the ownership story is confirmed.

## Tests

- `services/player-data/playerdata/runtime_test.go`
- `services/player-data/playerdata/configured_runtime_test.go`
- `services/player-data/playerdata/default_runtime_test.go`
- `services/player-data/playerdata/mode_policy_test.go`
- `services/player-data/playerdata/store_router_test.go`
- `services/player-data/playerdata/memory_store_test.go`
- `services/player-data/playerdata/rails_store_test.go`
- `services/player-data/playerdata/noop_store_test.go`
- `services/player-data/playerdata/embeddedsqlite/sqlite_store_test.go`
- TODO: add any additional verified tests here.

## Related docs

- [Player Data](../!README.md)
- [Documentation procedure](../../../documentation-procedure.md)
- TODO: add runtime-specific docs when they exist.

## Notes

Stub note: this document is a placeholder for future player-data runtime and store routing documentation.
Do not treat it as canonical source material.
