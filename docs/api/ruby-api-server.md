# Ruby API Server Plan

This is a future service plan. The Rails API-only scaffold now exists, but no product features are implemented yet:

```text
services/api-server/
```

## Purpose

The API server should own business/backend concerns that are separate from real-time game simulation.

Good fits:

- accounts and authentication
- user profiles
- matchmaking metadata
- room discovery metadata
- leaderboards
- unlocks/cosmetics
- purchases or entitlement checks
- admin/moderation endpoints
- persistence and database-backed workflows

The Go game server should continue to own real-time gameplay simulation.

## Planned Stack

Planned stack:

- Ruby
- Rails API-only

Rails API-only is a good fit because it gives the API service a focused request/response structure without tempting the API layer to import Go game internals.

## Service Boundary

The language/runtime split is intentional.

```text
services/game-server/  Go real-time simulation
services/api-server/   Ruby/Rails business API
```

Rules:

- The API server should not import or own real-time game simulation.
- The game server should not become an account/database/business API.
- Shared data should cross the boundary through explicit APIs, database records, events, or generated schemas.
- Do not duplicate gameplay authority in the API server.
- Do not put secrets in the Godot client.

Shared-schema boundary note:

- `shared/packets/` is the real-time game client/server protocol, not an automatic API contract.
- API-specific shared-schema output is deferred unless explicitly started.
- API contracts should stay separate unless a feature truly needs shared schema.

## Possible Future Communication

Start simple. The services do not need to talk until a feature requires it.

Possible later boundaries:

- API creates or records room metadata; game server owns live room state.
- API stores user/profile/leaderboard data; game server reports match results through an internal endpoint or event.
- API performs auth/token validation; game server verifies tokens at websocket connect.
- API owns matchmaking queues; game server receives selected room/session assignments.

## Repository Notes

The current game server Go module still uses this module path:

```text
github.com/Lokee86/space-rocks/server
```

That module now lives at:

```text
services/game-server/
```

The module path and filesystem path do not need to match.

## Initial Setup Steps Later

When ready to continue the API scaffold:

1. Keep local run/test/build entrypoints in the Rails project configuration inside `services/api-server/`.
2. Keep the API `.env.example` current if config is needed.
3. Keep API commands documented in `README.md` and `docs/developer.md`.
4. Keep API contracts separate from game packet schemas unless they truly need to be shared.

Do not add API dependencies to the Go game server.
