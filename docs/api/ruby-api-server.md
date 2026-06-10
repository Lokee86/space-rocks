# Ruby API Server Plan

This is the Ruby/Rails API service for Space Rocks. The current baseline already exists under:

```text
services/api-server/
```

Current implemented baseline:

- Rails API-only service exists under `services/api-server/`
- health endpoint exists
- email/password auth exists
- Discord OAuth auth exists at the Rails API level
- opaque bearer access tokens exist
- provider identity schema exists for future OAuth/provider login
- `/auth/me` verification exists

The current API-owned data model is:

- `users`
- `password_credentials`
- `user_identities`
- `access_tokens`

Rails migrations under `services/api-server/db/migrate/` own the API database schema. `shared/` is not the source of truth for API auth schema.

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

## Auth

The Rails API owns the auth data model:

- users
- password credentials
- provider identities
- access tokens

Discord OAuth support currently uses these required environment variables:

- `DISCORD_CLIENT_ID`
- `DISCORD_CLIENT_SECRET`
- `DISCORD_REDIRECT_URI`
- `GAME_SERVER_INTERNAL_TOKEN`

Implemented auth endpoints:

- `POST /auth/register`
- `POST /auth/login`
- `GET /auth/discord/start`
- `GET /auth/discord/callback`
- `POST /auth/discord/login_sessions` - create a login session for browser Discord handoff
- `POST /auth/discord/login_sessions/:id/exchange` - exchange an authenticated login session for the normal bearer token response
- `POST /internal/auth/verify-token` - verify Space Rocks bearer tokens for the Go game server only
- `GET /auth/me`
- `DELETE /auth/logout`

The player stats and internal match-results endpoints are also consumed by `services/player-data` through its Rails adapter for authenticated_account backing. `services/api-server` remains the Rails/Postgres persistence owner for authenticated account stats, and `services/player-data` does not read Rails tables directly.

### Godot Discord Login-Session Flow

Godot now uses a browser-assisted Discord login-session handoff with Rails:

- Godot asks Rails to create a Discord login session.
- Rails returns `login_session_id`, `poll_secret`, `login_url`, and `expires_at`.
- Godot opens `login_url` in the browser.
- The browser completes Discord OAuth.
- The Rails callback marks the login session authenticated.
- Godot polls and exchanges the login session using `login_session_id` and `poll_secret`.
- Rails returns the normal auth response with the Space Rocks bearer token and user payload.
- Godot stores the Space Rocks bearer token and validates it through `GET /auth/me`.
- The Go game server verifies Space Rocks bearer tokens through Rails at `POST /internal/auth/verify-token` and receives only the minimal identity needed for websocket admission.

This flow keeps Discord client secrets out of Godot and avoids manually copying browser JSON tokens.

Existing email/password auth and the existing direct Discord browser smoke behavior remain API-level capabilities.
Single-player remains unauthenticated and does not require Rails auth.
Websocket token authentication is now implemented for the Go game server through the Rails internal verification boundary.
Rails owns OAuth and bearer-token verification, while the Go game server consumes that boundary through authclient and websocket admission. Embedded DB, Local Profile, player-data routing, and player-data SSoT implementation remain later work.

The Go game server should not read Rails auth tables directly.

Email/password auth and Discord OAuth both issue the same opaque bearer access token.
`GET /auth/me` verifies either login path.

If the game server needs auth, it should use an explicit API or internal verification boundary rather than direct table access. Rails/API owns authenticated users, OAuth identities, and online account persistence.

`account_id` is the canonical cross-system UUID for authenticated accounts. Rails `user_id` remains an internal foreign key to `users.id`, and game-facing payloads should use `account_id` while local profiles use `local_profile_id`.

See [cross-mode routing and player data](../design/cross-mode-routing-and-player-data.md) for the cross-mode admission, identity, and player-data routing model.

JWT is still deferred for now, and the schema is structured so it can be added later without reworking the core account tables.

## Shared Schema Boundary

`shared/` is not the source of truth for API auth schema.

Shared packet schemas remain for the real-time game protocol, not API auth persistence.

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
  - API performs auth/token validation; game server verifies tokens at websocket connect through the Rails internal auth boundary.
- API owns matchmaking queues; game server receives selected room/session assignments.

Future/deferred:

- JWT

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
