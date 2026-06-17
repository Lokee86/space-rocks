# HTTP Contracts

`shared/contracts/http/openapi.yaml` owns all HTTP request and response shapes.

OpenAPI remains the SSoT for request and response shapes.

Rails controllers implement that contract, and Rails integration tests enforce it using `openapi_first`.

This is Level 2 enforcement:

- test-time request/response validation
- runtime OpenAPI middleware is not active yet

## Player-data HTTP Endpoints

### `POST /api/player-data/profile`

- Hosted by the game-server on `:8080` during the in-process player-data runtime phase
- Client-facing profile read facade
- Returns normalized profile payload
- Authenticated reads use the user bearer token to prove identity to the game-server; guest reads remain unauthenticated
- Does not call Rails stats directly
- Does not use `PLAYER_DATA_RAILS_BEARER_TOKEN`

### `GET /api/player-data/local-profiles`

- Hosted by the game-server data-handler on `:8080`
- Lists local profiles by `local_profile_id` and `display_name`
- `display_name` is presentation, not identity
- In the standard no-tag development build, this route is backed by the embedded SQLite local-profile store
- In `-tags noembeddedsqlite` deployment/restricted builds, the embedded SQLite package and `modernc.org/sqlite` dependency are not present, and local profile management returns `local_profiles_unavailable`

### `POST /api/player-data/local-profiles`

- Hosted by the game-server data-handler on `:8080`
- Creates a local profile
- Request includes `display_name` and `seed_from_guest_stats`
- `display_name` is presentation, not identity
- `local_profile_id` is generated server-side
- Guest stats seeding stays separate from local profile storage

### `DELETE /api/player-data/local-profiles/{local_profile_id}`

- Hosted by the game-server data-handler on `:8080`
- Deletes the local profile, local stats, and local match results
- Deleting the default local profile resets the default to Guest
- Requires `local_profile_id` as the path key

### `PUT /api/player-data/local-profiles/{local_profile_id}`

- Hosted by the game-server data-handler on `:8080`
- Updates the local profile display name only
- Does not reset stats
- Leaves `local_profile_id` unchanged
- Request includes `display_name`

### `GET /api/player-data/local-profiles/default`

- Hosted by the game-server data-handler on `:8080`
- Returns the persisted local pilot default
- Returns `local_profiles_unavailable` in `-tags noembeddedsqlite` deployment/restricted builds

### `PUT /api/player-data/local-profiles/default`

- Hosted by the game-server data-handler on `:8080`
- Persists Guest or a local profile as the default
- Guest uses `identity_kind = guest`
- Local profile uses `identity_kind = local_profile` plus `local_profile_id`
- Returns `local_profiles_unavailable` in `-tags noembeddedsqlite` deployment/restricted builds

### `POST /api/internal/player-data/stats`

- Hosted by the Rails API on `:3000`
- Internal service-to-service stats read by `account_id`
- Called by `RailsStore.LoadStats`
- Uses an internal bearer token for game-server/player-data to Rails service calls
- Authenticated account stats remain Rails-backed

### `GET /api/player/stats`

- Existing public Rails endpoint
- Still available for public Rails API behavior
- Not the profile readout source after the data-handler reroute

### Token Roles

- User bearer token proves identity to the game-server profile endpoint for authenticated reads
- Internal bearer token is used for game-server/player-data to Rails service calls
- Static user bearer tokens such as `PLAYER_DATA_RAILS_BEARER_TOKEN` are not used

## What This Does Not Do

- does not generate Rails controllers
- does not replace Rails strong params
- does not own Rails database schema
- does not generate TypeScript yet
- does not cover WebSocket packet schema
- does not own player-data packet schemas

## Update Rule

HTTP request/response shape changes must update `shared/contracts/http/openapi.yaml` and the relevant contract tests in the same change.

Player-data runtime packet shape changes must update the relevant `shared/packets/player_data.toml` entries and the generated outputs in the same change.

## Verification

- `cd services/api-server && bundle exec rails test test/contracts/openapi_contract_test.rb`
- `cd services/api-server && bundle exec rails test test/controllers/api/internal/player_data/stats_controller_test.rb`
- `cd services/player-data && go test ./...` should include `modernc.org/sqlite`
- `cd services/player-data && go test -tags noembeddedsqlite ./...`
- `cd services/player-data && go list -tags noembeddedsqlite -deps ./... | grep modernc.org/sqlite` should find nothing in deployment/restricted builds
- `cd services/game-server && go test -buildvcs=false ./cmd/game-server` should include `modernc.org/sqlite`
- `cd services/game-server && go test -tags noembeddedsqlite -buildvcs=false ./cmd/game-server`
- `cd services/game-server && go list -tags noembeddedsqlite -deps ./cmd/game-server | grep modernc.org/sqlite` should find nothing in deployment/restricted builds

Local profile management endpoints return `local_profiles_unavailable` when embedded local storage is unavailable.

## Related

- [Project source-of-truth map](../design/source-of-truth-map.md)
