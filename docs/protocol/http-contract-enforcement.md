## HTTP Contract Enforcement

Parent index: [Protocol](!README.md)

## Purpose

This document describes how Space Rocks currently defines and enforces HTTP request and response contracts across services.

It covers the shared OpenAPI contract, the Rails API contract test boundary, client HTTP consumption, and the service responsibilities around HTTP request/response shape changes.

## Overview

HTTP request and response shapes are owned by:

```text
shared/contracts/http/openapi.yaml
```

That file is the source of truth for the current HTTP API contract. It defines the request bodies, response bodies, status codes, security expectations, and shared schemas for the Rails API, player-data-facing HTTP endpoints, auth verification, local profile endpoints, and internal match-result submission.

Current enforcement is Level 2:

```text
test-time OpenAPI parsing and request/response validation
```

Runtime OpenAPI middleware is not active. The OpenAPI contract does not generate Rails controllers, does not replace Rails strong params, does not generate Godot client code, does not generate Go server/client code, and does not own database schema.

HTTP contracts are separate from realtime WebSocket packet contracts. Realtime packet shapes remain owned by the shared packet schema pipeline, not by OpenAPI.

## Participating systems

```text
shared/contracts/http/openapi.yaml
```

Defines the HTTP request and response contract.

```text
services/api-server/
```

Implements Rails HTTP endpoints for auth, account-facing player stats, internal token verification, and internal authenticated-account player-data persistence routes.

```text
client/
```

Consumes HTTP endpoints through JSON request helpers. The client sends JSON bodies, accepts JSON responses, and attaches bearer tokens when a caller provides one.

```text
services/game-server/
```

Participates in HTTP contract flow where game-server runtime code verifies authenticated bearer tokens through Rails and reports match results through the player-data path.

The in-process player-data HTTP surface is currently associated with the game-server runtime during local/player-data integration.

```text
services/player-data/
```

Owns player-data routing behavior and persistence selection for guest, local profile, and authenticated account data. Authenticated account persistence routes through Rails instead of reading Rails database tables directly.

## Authority

OpenAPI owns HTTP request and response shape.

Rails controllers own Rails endpoint implementation.

Rails migrations own the Rails API database schema.

Player-data storage code owns player-data persistence behavior.

The game server owns realtime simulation and match result production.

The client owns presentation and HTTP request consumption.

OpenAPI does not own:

```text
Rails database schema
Rails controller generation
Rails strong params
Godot API client generation
Go API client generation
runtime middleware enforcement
WebSocket packet contracts
player-data runtime packet schemas
```

## Message or request flow

Public auth endpoints are Rails API endpoints.

```text
POST   /api/auth/register
POST   /api/auth/login
DELETE /api/auth/logout
GET    /api/auth/me
GET    /api/auth/discord/start
GET    /api/auth/discord/callback
POST   /api/auth/discord/login_sessions
POST   /api/auth/discord/login_sessions/{id}/exchange
```

The client sends JSON HTTP requests through its API HTTP helper. The helper adds JSON accept/content headers, optionally adds an Authorization bearer header, serializes non-GET request bodies as JSON, parses JSON dictionary responses, and converts HTTP or parse failures into request-result failures.

`GET /api/auth/discord/callback` currently has two `200` success shapes: the normal auth response for direct browser OAuth, or a short handoff message after login-session authentication.

Authenticated client requests use the user bearer token.

Internal service-to-service requests use an internal bearer token where required.

Token verification crosses from the game server to Rails through:

```text
POST /internal/auth/verify-token
```

Authenticated account stats and match result persistence use Rails-owned internal routes:

```text
POST /api/internal/player-data/stats
POST /internal/player-data/match-results
```

Player-data profile and local profile HTTP routes are part of the shared HTTP contract:

```text
POST   /api/player-data/profile
GET    /api/player-data/local-profiles
POST   /api/player-data/local-profiles
PUT    /api/player-data/local-profiles/{local_profile_id}
DELETE /api/player-data/local-profiles/{local_profile_id}
GET    /api/player-data/local-profiles/default
PUT    /api/player-data/local-profiles/default
```

The exact Go implementation paths for those player-data HTTP handlers still need confirmation.

## Source-of-truth files

```text
shared/contracts/http/openapi.yaml
```

Primary HTTP contract source.

```text
services/api-server/config/routes.rb
```

Rails route implementation surface.

```text
services/api-server/test/contracts/openapi_contract_test.rb
```

Verifies that the shared OpenAPI definition parses.

```text
services/api-server/test/support/openapi_contract_assertions.rb
```

Defines request, response, and combined OpenAPI contract assertions for Rails integration tests.

```text
services/api-server/test/test_helper.rb
```

Installs the OpenAPI contract assertion helper into Rails integration tests.

```text
services/api-server/Gemfile
```

Keeps `openapi_first` in the development/test dependency group.

```text
client/scripts/api/api_http_client.gd
```

Godot HTTP JSON request helper.

## Service responsibilities

### API server

The API server owns Rails HTTP route implementation for Rails-hosted endpoints.

It owns:

```text
auth routes
Discord OAuth routes
current-user route
public account stats route
internal token verification route
internal authenticated-account stats route
internal match-result persistence route
Rails contract tests
Rails database schema for API-owned persistence
```

It does not own:

```text
game simulation
local-profile durable storage
guest transient storage
OpenAPI generation
runtime OpenAPI middleware enforcement
WebSocket packet schemas
```

### Client

The client consumes HTTP endpoints.

It owns:

```text
JSON request construction
optional bearer token header attachment
JSON response parsing
HTTP error mapping for UI/application flows
```

It does not currently own:

```text
generated API clients
OpenAPI schema validation
provider secrets
server-side auth verification
```

### Game server

The game server owns realtime gameplay simulation and produces match result summaries.

It participates in HTTP contract flow when it:

```text
verifies authenticated bearer tokens through Rails
reports resolved match results through the player-data path
hosts or composes the in-process player-data HTTP facade during local/player-data runtime
```

It does not own Rails auth tables or Rails persistence schema.

### Player data

Player-data runtime owns profile and stats routing between:

```text
guest transient data
local profile storage
authenticated account storage through Rails
```

It should not read Rails tables directly. Authenticated account reads and writes cross explicit HTTP/service boundaries.

## Validation and testing

Rails OpenAPI contract support currently validates through test helpers, not runtime middleware.

Minimum Rails validation:

```text
cd services/api-server && bundle exec rails test test/contracts/openapi_contract_test.rb
```

Broader Rails validation:

```text
cd services/api-server && bundle exec rails test
```

Player-data and game-server validation should be run when HTTP contract changes affect those services:

```text
cd services/player-data && go test ./...
cd services/game-server && go test -buildvcs=false ./...
```

When embedded local profile behavior is affected, also verify the restricted build path:

```text
cd services/player-data && go test -tags noembeddedsqlite ./...
cd services/game-server && go test -tags noembeddedsqlite -buildvcs=false ./cmd/game-server
```

HTTP request or response shape changes must update the OpenAPI contract and the relevant service tests in the same change.

## Code map

Primary contract:

```text
shared/contracts/http/openapi.yaml
```

Rails API implementation and enforcement:

```text
services/api-server/config/routes.rb
services/api-server/Gemfile
services/api-server/test/test_helper.rb
services/api-server/test/contracts/openapi_contract_test.rb
services/api-server/test/support/openapi_contract_assertions.rb
```

Client HTTP consumer:

```text
client/scripts/api/api_http_client.gd
```

Paths still requiring confirmation:

```text
Rails controller tests that call assert_openapi_contract!
Go player-data HTTP handler registration
Go local-profile HTTP handlers
Go RailsStore authenticated-account HTTP adapter
Go game-server token verification client
Go match-result reporting path into player-data
```

Important non-ownership boundaries:

```text
OpenAPI does not own Rails database schema.
OpenAPI does not own WebSocket packet schema.
OpenAPI does not currently generate clients or controllers.
OpenAPI is not currently runtime middleware.
```

## Related docs

* [Protocol](!README.md)
* [Data](../data/!README.md)
* [API Server](../services/api-server/!README.md)
* [Game Server](../services/game-server/!README.md)
* [Player Data](../services/player-data/!README.md)
* [Documentation policy](../documentation-policy.md)
* [Documentation procedure](../documentation-procedure.md)

## Notes

The Go implementation paths and Rails controller-test coverage still need confirmation.

If future work focuses on OpenAPI generation, source updates, generated outputs, or pipeline commands, split that material into a data doc instead of expanding this protocol doc into a data pipeline document.

If future work adds runtime OpenAPI middleware, update the enforcement level and add the runtime implementation path to the code map.
