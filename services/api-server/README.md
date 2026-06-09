# Space Rocks API Server

This service is the Ruby/Rails API-only server for Space Rocks business and backend concerns.

The Go game server still owns real-time simulation, including movement, bullets, collisions, scoring, lives, death, respawn, pause safety, rooms, and websocket state.

This API is no longer just a scaffold. The current baseline includes health and email/password auth.

## Local Setup

```bash
bundle install
bundle exec rails db:create
bundle exec rails test
bundle exec rails server
```

The API server runs locally on port `3000` by default.

## Health Check

`GET /health`

Returns a static JSON response:

```json
{
  "status": "ok",
  "service": "space-rocks-api"
}
```

## Auth

The Rails API owns the auth persistence layer at a high level:

- users
- password credentials
- provider identities
- access tokens

The auth endpoints issue opaque bearer tokens for API access. Tokens are stored hashed in the database.
Both email/password auth and Discord OAuth issue the same opaque bearer access token.

Discord OAuth requires these environment variables:

- `DISCORD_CLIENT_ID`
- `DISCORD_CLIENT_SECRET`
- `DISCORD_REDIRECT_URI`

### `POST /auth/register`

Create a new user with an email/password login.

Request body:

```json
{
  "display_name": "Test Pilot",
  "email": "test@example.com",
  "password": "password123"
}
```

Returns the created user plus a token.

### `POST /auth/login`

Log in with an existing email/password credential.

Request body:

```json
{
  "email": "test@example.com",
  "password": "password123"
}
```

Returns the current user plus a new token.

### `GET /auth/discord/start`

Begin the Discord OAuth flow by redirecting the browser to Discord.

### `GET /auth/discord/callback`

Handle the browser-driven Discord OAuth callback after Discord redirects back with `code` and `state`.

Returns the current user plus a new token on success.

### `GET /auth/me`

Return the current authenticated user.

Protected endpoint. Send:

```http
Authorization: Bearer <token>
```

Returns the user payload for a valid token.
This works for bearer tokens issued by either email/password auth or Discord OAuth.

### `DELETE /auth/logout`

Revoke the current bearer token.

Protected endpoint. Send:

```http
Authorization: Bearer <token>
```

Returns no content on success. The same token should fail on `GET /auth/me` after logout.

## Bruno Smoke Tests

Use the Bruno collection rooted at `bruno-api/` for local API smoke testing.

Local environment variables:

- `base_url`
- `email`
- `password`
- `display_name`
- `auth_token`

Suggested smoke-test order:

1. `Health`
2. `Register` or `Login`
3. Copy the returned token into `auth_token`
4. `Me`
5. `Logout`
6. `Me` should fail with the same token after logout

The collection is for manual smoke testing only and should not use real secrets or real auth tokens.

Future/deferred:

- JWT
- game-server auth verification boundaries
