# API Server Plan

This is a future service plan. The API server is not implemented yet beyond the repository placeholder:

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

- Node.js
- TypeScript
- NestJS

NestJS is a good fit because it gives the API service a strong module/controller/service structure without tempting the API layer to import Go game internals.

## Service Boundary

The language/runtime split is intentional.

```text
services/game-server/  Go real-time simulation
services/api-server/   Node/TypeScript business API
```

Rules:

- The API server should not import or own real-time game simulation.
- The game server should not become an account/database/business API.
- Shared data should cross the boundary through explicit APIs, database records, events, or generated schemas.
- Do not duplicate gameplay authority in the API server.
- Do not put secrets in the Godot client.

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

When ready to scaffold the API service:

1. Initialize a NestJS project inside `services/api-server/`.
2. Add local run/test/build scripts to `services/api-server/package.json`.
3. Add an API `.env.example` if config is needed.
4. Document API commands in `README.md` and `docs/developer.md`.
5. Keep API contracts separate from game packet schemas unless they truly need to be shared.

Do not add API dependencies to the Go game server.
