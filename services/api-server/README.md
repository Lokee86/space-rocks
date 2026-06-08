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
