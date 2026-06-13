# HTTP Contracts

`shared/contracts/http/openapi.yaml` owns all HTTP request and response shapes.

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

### `POST /api/internal/player-data/stats`

- Hosted by the Rails API on `:3000`
- Internal service-to-service stats read by `account_id`
- Called by `RailsStore.LoadStats`
- Uses an internal bearer token for game-server/player-data to Rails service calls

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
- `cd services/player-data && go test ./...`
- `cd services/game-server && go test ./cmd/game-server -run PlayerDataProfileHandler`

## Related

- [Project source-of-truth map](../design/source-of-truth-map.md)
