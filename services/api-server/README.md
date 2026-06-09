# Space Rocks API Server

This service is the Ruby/Rails API-only server for Space Rocks business and backend concerns.

The Go game server still owns real-time simulation, including movement, bullets, collisions, scoring, lives, death, respawn, pause safety, rooms, and websocket state.

This API is no longer just a scaffold. The current baseline includes health, email/password auth, Discord OAuth at the Rails API level, Godot Discord login-session handoff, `/auth/me` validation, and opaque bearer tokens.

## Local Auth Setup

Local Discord OAuth secrets live outside git in `.secrets/api-server.env`.
The `.secrets/` directory is ignored, and secrets must not be committed.
If this repo's `.envrc` is enabled, `direnv` is the preferred local workflow for exporting them into the shell.
Rails must be started from a shell where the Discord OAuth variables are already exported.

Required environment variables:

- `DISCORD_CLIENT_ID`
- `DISCORD_CLIENT_SECRET`
- `DISCORD_REDIRECT_URI`

```bash
bundle install
bundle exec rails db:migrate
bundle exec rails test
bundle exec rails server
```

If this is a brand-new local database, run `bundle exec rails db:create` once before `db:migrate`.

The API server runs locally on port `3000` by default.

Discord smoke path:

1. Ensure the Discord env vars are loaded from `.secrets/` via `direnv` or your shell.
2. Run `bundle exec rails db:migrate`.
3. Start Rails with `bundle exec rails server`.
4. Start Godot.
5. Click `Sign-in`.
6. Complete Discord login in the browser.
7. Return to Godot and confirm the menu shows your display name.
8. Click `Logout`.

Troubleshooting:

- `PendingMigrationError` means run `bundle exec rails db:migrate`.
- Rails reference columns create an index by default, so `add_reference :table, :thing, foreign_key: true` plus `add_index :table, :thing_id` will duplicate the index unless `index: false` is set on the reference.
- Duplicate index names in new migrations usually mean `add_reference` or `t.references` already created the index and a second `add_index` line was added.

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

Discord OAuth is implemented in the Rails API and requires the environment variables listed in Local Auth Setup above.

The Rails API also expects `GAME_SERVER_INTERNAL_TOKEN` for internal calls from the Go game-server. Normal clients must never receive this value. `POST /internal/auth/verify-token` requires this bearer token.

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
Email may be `null`.
Discord OAuth and email/password auth both issue the same opaque bearer token.

### `POST /auth/discord/login_sessions`

Create a login session for the browser Discord handoff.

Returns `login_session_id`, `poll_secret`, `login_url`, and `expires_at`.

### `POST /auth/discord/login_sessions/:id/exchange`

Exchange an authenticated login session for the normal bearer token response.

#### Discord OAuth smoke test

1. Start Rails with the Discord env vars active.
2. Open `http://localhost:3000/auth/discord/start` in a browser.
3. Approve the Discord login.
4. Confirm the callback returns user plus token JSON.
5. Copy the raw token only.
6. Call `GET /auth/me` with `Authorization: Bearer <token>`.

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
- game-server token verification boundaries
