## Internal API Surface

Parent index: [API Server](./!README.md)

## Purpose

This document describes the current internal API surface implemented by `services/api-server`.

It covers Rails endpoints consumed by sibling services, not public client auth routes, local profile HTTP routes hosted by the game-server data-handler, or realtime websocket packets.

## Overview

The API server currently exposes three implemented internal Rails surfaces:

| Route                                      | Caller                                                         | Purpose                                                                         |
| ------------------------------------------ | -------------------------------------------------------------- | ------------------------------------------------------------------------------- |
| `POST /internal/auth/verify-token`         | `services/game-server/internal/authclient`                     | Verify a user bearer token for authenticated-account websocket admission.       |
| `POST /api/internal/player-data/stats`     | `services/player-data/playerdata.RailsStore.LoadStats`         | Load authenticated-account stats by `account_id`.                               |
| `POST /internal/player-data/match-results` | `services/player-data/playerdata.RailsStore.RecordMatchResult` | Persist trusted authenticated-account match results and update aggregate stats. |

All three routes require the internal bearer token from `GAME_SERVER_INTERNAL_TOKEN`.

The `/api/internal/...` prefix does not make a route public. Current code uses both `/internal/...` and `/api/internal/...` namespaces for service-to-service endpoints. Both are protected by the same internal base controller.

## Code root

* `services/api-server/`

## Responsibilities

The API server owns:

* Internal bearer-token authentication for Rails service-to-service endpoints.
* User bearer-token verification for game-server authenticated-account admission.
* Mapping valid user bearer tokens to minimal account identity.
* Authenticated-account stats reads by `account_id`.
* Authenticated-account match-result persistence.
* Idempotent match-result handling by `result_id`.
* Aggregate authenticated-account stat updates for:

  * `total_score`
  * `high_score`
  * `ship_deaths`
  * `games_played`
  * `wins`
* Rails/Postgres persistence for authenticated-account users, access tokens, stats, and match results.
* OpenAPI request/response contract coverage for the implemented HTTP surface.

## Does not own

The API server does not own:

* Realtime game simulation.
* Websocket sessions, room membership, or match lifecycle.
* Local single-player Guest behavior.
* Local Profile SQLite persistence.
* Game-server data-handler routes such as `POST /api/player-data/profile`.
* Runtime player-data store selection.
* Client-side profile readout.
* Websocket packet schema or packet codec behavior.
* Direct gameplay stat mutation during a live match.

## Domain roles

The internal API surface participates in these roles:

* **Authenticated account authority:** Rails owns authenticated users, account identity, access tokens, and online account persistence.
* **Token verification boundary:** the game-server submits a user bearer token to Rails and receives only minimal identity data needed for admission.
* **Authenticated-account player-data backing store:** `services/player-data` routes authenticated-account reads and writes to Rails through `RailsStore`.
* **Match-result persistence sink:** Rails stores trusted match summaries produced upstream by the game-server/player-data flow.
* **Contract implementation:** Rails controllers implement HTTP shapes owned by `shared/contracts/http/openapi.yaml`.

## Protocols and APIs

This surface exposes internal Rails endpoints for game-server and player-data service-to-service calls. The game-server consumes token verification, and the player-data service consumes account-backed stats and match-result operations. Rails owns authenticated-account and token authority behind these endpoints, while the boundary carries internal bearer authentication plus account IDs and minimal identity or stats data. It does not own client session routing, gameplay admission policy, or presentation behavior.

### Internal request authentication

All current internal Rails endpoints inherit from `Internal::BaseController`.

Internal requests must send:

```http
Authorization: Bearer <GAME_SERVER_INTERNAL_TOKEN>
```

Failure behavior:

* Missing `GAME_SERVER_INTERNAL_TOKEN` returns `401`.
* Missing `Authorization` header returns `401`.
* Non-`Bearer` authorization scheme returns `401`.
* Wrong bearer token returns `401`.

The controller extracts the bearer token and compares it to `GAME_SERVER_INTERNAL_TOKEN` using `ActiveSupport::SecurityUtils.secure_compare` after checking equal length.

Normal clients must never receive the internal token.

### `POST /internal/auth/verify-token`

Caller:

* `services/game-server/internal/authclient.Client.VerifyToken`

Purpose:

* Verify a user bearer token before authenticated-account websocket admission.
* Return minimal identity data for the game-server session.

Request body:

```json
{
  "token": "user-bearer-token"
}
```

Success response for a valid user token:

```json
{
  "valid": true,
  "user": {
    "id": 1,
    "account_id": "account-uuid",
    "display_name": "Ada"
  }
}
```

Success response for a missing, unknown, revoked, or expired user token:

```json
{
  "valid": false
}
```

Important behavior:

* Invalid user tokens return `200` with `valid: false`.
* Invalid internal service authentication returns `401`.
* The response intentionally excludes credential, token digest, expiry, revocation, and email details.
* `Auth::VerifyAccessToken` updates `last_used_at` for a valid access token.

### `POST /api/internal/player-data/stats`

Caller:

* `services/player-data/playerdata.RailsStore.LoadStats`

Purpose:

* Load authenticated-account stats by `account_id`.

Request body:

```json
{
  "account_id": "account-uuid"
}
```

Success response:

```json
{
  "stats": {
    "total_score": 12,
    "high_score": 12,
    "ship_deaths": 3,
    "games_played": 1,
    "wins": 1
  }
}
```

Failure behavior:

* Missing `account_id` returns `422` with `error: "invalid_input"`.
* Unknown `account_id` returns `404` with `error: "unknown_user"`.
* Invalid internal service authentication returns `401`.

Important behavior:

* The controller looks up `User` by `account_id`.
* If the user exists but has no `PlayerStat`, Rails creates a zeroed stats row before returning the response.
* The endpoint is for service-to-service authenticated-account stats reads, not direct client profile readout.

### `POST /internal/player-data/match-results`

Caller:

* `services/player-data/playerdata.RailsStore.RecordMatchResult`

Purpose:

* Persist a trusted authenticated-account match result.
* Update aggregate account stats.

Request body:

```json
{
  "result_id": "result-1",
  "match_id": "match-1",
  "account_id": "account-uuid",
  "score": 12,
  "ship_deaths": 3,
  "won": true
}
```

Success response:

```json
{
  "accepted": true,
  "duplicate": false,
  "stats": {
    "total_score": 12,
    "high_score": 12,
    "ship_deaths": 3,
    "games_played": 1,
    "wins": 1
  }
}
```

Duplicate response:

```json
{
  "accepted": true,
  "duplicate": true,
  "stats": {
    "total_score": 12,
    "high_score": 12,
    "ship_deaths": 3,
    "games_played": 1,
    "wins": 1
  }
}
```

Failure behavior:

* Missing `result_id`, `match_id`, or `account_id` returns `422` with `accepted: false` and `error: "invalid_input"`.
* Unknown `account_id` returns `404` with `accepted: false` and `error: "unknown_user"`.
* Invalid model/service input returns `422` with `accepted: false` and `error: "invalid_input"`.
* Invalid internal service authentication returns `401`.

Important behavior:

* `PlayerStats::ApplyMatchResult` runs inside a database transaction.
* `result_id` is the idempotency key.
* A duplicate `result_id` is accepted without double-counting stats.
* New results create a `PlayerMatchResult` row and update the user aggregate `PlayerStat`.
* The endpoint trusts upstream match facts; it does not recompute gameplay outcomes.

## Data ownership

The internal API surface owns or mutates these Rails-backed records:

* `users`

  * `account_id` is the canonical cross-system authenticated-account UUID.
  * Rails `users.id` remains an internal database foreign key.
* `access_tokens`

  * Used by `Auth::VerifyAccessToken`.
  * Stored as digests, not raw tokens.
* `player_stats`

  * Aggregate authenticated-account stats.
* `player_match_results`

  * Per-match result records keyed by `result_id`.

The internal API surface does not own:

* Local Profile records in embedded SQLite.
* Guest transient stats.
* Player-data runtime packet state.
* Game-server match lifecycle state.
* Client presentation state.

## Code map

Routes and shared controller boundary:

* `services/api-server/config/routes.rb`
* `services/api-server/app/controllers/internal/base_controller.rb`
* `services/api-server/app/controllers/api/internal/base_controller.rb`

Internal auth verification:

* `services/api-server/app/controllers/internal/auth/verify_tokens_controller.rb`
* `services/api-server/app/services/auth/verify_access_token.rb`
* `services/api-server/app/models/access_token.rb`
* `services/game-server/internal/authclient/client.go`
* `services/game-server/internal/authclient/types.go`

Internal player-data stats read:

* `services/api-server/app/controllers/api/internal/player_data/stats_controller.rb`
* `services/api-server/app/services/player_stats/serialize_stats.rb`
* `services/api-server/app/models/player_stat.rb`
* `services/player-data/playerdata/rails_store.go`

Internal match-result write:

* `services/api-server/app/controllers/internal/player_data/match_results_controller.rb`
* `services/api-server/app/services/player_stats/apply_match_result.rb`
* `services/api-server/app/services/player_stats/serialize_stats.rb`
* `services/api-server/app/models/player_match_result.rb`
* `services/api-server/app/models/player_stat.rb`
* `services/player-data/playerdata/rails_store.go`

Contract source:

* `shared/contracts/http/openapi.yaml`

Important non-ownership boundaries:

* `services/game-server/internal/authclient/` consumes token verification but does not read Rails auth tables.
* `services/player-data/playerdata/rails_store.go` consumes stats and match-result endpoints but does not own Rails persistence.
* Game-server-hosted local profile and profile-readout HTTP routes are not Rails API-server routes.

## Tests

Internal auth verification:

* `services/api-server/test/controllers/internal/auth/verify_tokens_controller_test.rb`
* `services/api-server/test/services/auth/verify_access_token_test.rb`
* `services/game-server/internal/authclient/client_test.go`

Internal player-data stats read:

* `services/api-server/test/controllers/api/internal/player_data/stats_controller_test.rb`
* `services/player-data/playerdata/rails_store_test.go`
* `services/player-data/playerdata/configured_runtime_test.go`

Internal match-result write:

* `services/api-server/test/controllers/internal/player_data/match_results_controller_test.rb`
* `services/api-server/test/services/player_stats/apply_match_result_test.rb`
* `services/player-data/playerdata/rails_store_test.go`
* `services/player-data/playerdata/configured_runtime_test.go`

HTTP contract enforcement:

* `services/api-server/test/contracts/openapi_contract_test.rb`
* `services/api-server/test/support/openapi_contract_assertions.rb`

Suggested verification from `services/api-server/`:

```bash
bundle exec rails test test/controllers/internal/auth/verify_tokens_controller_test.rb
bundle exec rails test test/controllers/api/internal/player_data/stats_controller_test.rb
bundle exec rails test test/controllers/internal/player_data/match_results_controller_test.rb
bundle exec rails test test/services/player_stats/apply_match_result_test.rb
bundle exec rails test test/contracts/openapi_contract_test.rb
```

Cross-service caller verification:

```bash
cd services/game-server && go test -buildvcs=false ./internal/authclient
cd services/player-data && go test ./playerdata
```

## Related docs

* [API Server](./!README.md)
* [Auth And OAuth](auth-and-oauth.md)
* [Player Stats And Match Results](player-stats-and-match-results.md)

## Notes

Legacy source material used while rebuilding this document included API server notes, HTTP contract notes, cross-mode routing/player-data notes, player-data routing notes, and the source-of-truth map. Those legacy docs remain migration source material only and should not be linked here as authority.

The current route split between `/internal/...` and `/api/internal/...` is implementation fact, not a documented design preference. Treat both route families as internal service-to-service API surface.
