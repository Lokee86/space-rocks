# Space Rocks API Server

This service is the Ruby/Rails API-only server for Space Rocks business and backend concerns.

The Go game server still owns real-time simulation, including movement, bullets, collisions, scoring, lives, death, respawn, pause safety, rooms, and websocket state.

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

The auth endpoints issue opaque bearer tokens for API access. Tokens are stored hashed in the database.

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

### `GET /auth/me`

Return the current authenticated user.

Protected endpoint. Send:

```http
Authorization: Bearer <token>
```

Returns the user payload for a valid token.

### `DELETE /auth/logout`

Revoke the current bearer token.

Protected endpoint. Send:

```http
Authorization: Bearer <token>
```

Returns no content on success. The same token should fail on `GET /auth/me` after logout.
