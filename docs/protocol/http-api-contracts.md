# HTTP API Contracts

Parent index: [Protocol](./!INDEX.md)

## Purpose

This document describes HTTP request/response contract ownership, the OpenAPI source of truth, enforcement expectations, and contract-change rules.

For the current API product surface map, see [API Product Surface](api-product-surface.md).
Future API surface planning belongs in [API Product Surface Planning](../planning/protocol/api-product-surface.md).

## Overview

HTTP request and response shapes are owned by `shared/contracts/http/openapi.yaml`.

That file is the source of truth for current JSON HTTP contracts. It defines path and method combinations, request bodies, response bodies, status codes, security declarations, and shared schemas for the currently implemented HTTP surfaces.

The contract is implemented manually by services. It does not generate Rails controllers, Go handlers, Godot API clients, Go API clients, Rails strong params, database migrations, or runtime middleware.

Current enforcement is test-time enforcement. Rails tests load the OpenAPI definition through `openapi_first`, and controller tests can assert request and response conformance with `assert_openapi_contract!`. Runtime OpenAPI middleware is not active.

HTTP API contracts are separate from realtime WebSocket packet contracts. WebSocket packet shapes are owned by the shared packet schema pipeline under `shared/packets/`, not by OpenAPI.

## Message or Request Flow

HTTP contract flow is intentionally shallow:

```text
shared/contracts/http/openapi.yaml
-> service implementation
-> service-specific behavior docs
```

The contract doc names the HTTP shape. The surface map in [API Product Surface](api-product-surface.md) names the current product surfaces only. Detailed behavior for the player-data HTTP subset belongs in [Player Data HTTP API](./player-data-http-api.md).

## Authority

OpenAPI owns the shape of HTTP messages.

Service implementations own the behavior behind those messages.

The contract does not own service internals, storage layout, or gameplay authority. It only owns HTTP request/response shape and related contract rules.

OpenAPI does not own:

```text
Rails database schema
Rails migrations
embedded SQLite schema
Rails controller generation
Go handler generation
Godot client generation
runtime OpenAPI middleware
WebSocket packet schemas
player-data runtime packets
gameplay simulation authority
```

## Source-of-truth Files

`shared/contracts/http/openapi.yaml` owns the HTTP request and response shape for the current contract set.

The contract is implemented manually by the participating services. Service code must stay aligned with the OpenAPI source, but the source file remains the authority for shape.

Supporting docs:

- [API Product Surface](api-product-surface.md)
- [Planning API Product Surface](../planning/protocol/api-product-surface.md)
- [Player Data HTTP API](./player-data-http-api.md)
- [HTTP Contract Enforcement](./http-contract-enforcement.md)

Detailed behavior belongs in the linked service docs and protocol docs, not here.

## Participating Systems

```text
shared/contracts/http/openapi.yaml
```

Owns HTTP request and response shape.

```text
services/api-server/
```

Implements Rails-hosted HTTP routes that participate in the contract for auth, OAuth handoff, current-user reads, public player stats, internal token verification, authenticated-account stats reads, and authenticated-account match-result persistence.

```text
services/game-server/
```

Hosts the current player-data HTTP facade on the game-server HTTP process and exposes the game-server health surface that participates in the contract set.

```text
services/player-data/
```

Implements profile read, local profile management, runtime store routing, authenticated-account Rails adapter calls, guest transient stats, and local profile storage behavior behind the contract.

```text
client/
```

Consumes HTTP routes through Godot JSON API clients, but does not own the contract authority.

## Service Responsibilities

Contract responsibilities are intentionally narrow:

- The API server owns Rails-hosted auth, OAuth handoff, current-user reads, public player stats, internal token verification, and authenticated-account persistence behavior behind the HTTP shape.
- The game server owns the current host process for player-data HTTP routes and its health/process endpoints.
- The player-data service owns profile readout, local profile management, and authenticated-account Rails adapter behavior behind the HTTP shape.
- The client owns HTTP consumption and presentation-facing handling, not contract authority.

Surface ownership details stay in [API Product Surface](api-product-surface.md), while detailed player-data behavior stays in [Player Data HTTP API](./player-data-http-api.md).

## Validation and Testing

HTTP request or response shape changes must update the OpenAPI source and affected implementation/tests in the same change.

The minimum contract update rule is:

```text
1. Update shared/contracts/http/openapi.yaml.
2. Update the implementing Rails controller, Go handler, or client/API wrapper.
3. Update affected controller, handler, adapter, or client tests.
4. Run OpenAPI parsing and affected service tests.
```

Rails contract validation uses `openapi_first` in tests.

The basic OpenAPI parse test is:

```text
cd services/api-server && bundle exec rails test test/contracts/openapi_contract_test.rb
```

Rails integration tests can validate the current request and response by calling:

```text
assert_openapi_contract!
```

That assertion validates both:

```text
assert_openapi_request!
assert_openapi_response!
```

Go-hosted player-data HTTP routes are currently listed in the OpenAPI contract and tested through Go handler/runtime tests. They are not currently validated by Rails integration tests because they are not Rails routes.

Recommended verification when HTTP contracts change:

```text
cd services/api-server && bundle exec rails test
cd services/player-data && go test ./...
cd services/game-server && go test -buildvcs=false ./...
```

When local profile availability or build-tag behavior changes, also run:

```text
cd services/player-data && go test -tags noembeddedsqlite ./...
cd services/game-server && go test -tags noembeddedsqlite -buildvcs=false ./cmd/game-server
```

When client HTTP call shapes or response handling change, run the affected Godot tests in addition to service tests.

## Compatibility And Updates

HTTP contract changes should preserve compatibility unless the change is intentionally breaking and coordinated across the affected callers.

When a contract changes, keep the OpenAPI source, service implementation, and tests aligned in the same change. Prefer additive updates over replacement updates when the product surface can support them.

WebSocket packet changes must stay in the packet schema pipeline and must not be treated as HTTP contract changes.

## Related Docs

- [API Product Surface](api-product-surface.md)
- [Player Data HTTP API](./player-data-http-api.md)
- [HTTP Contract Enforcement](./http-contract-enforcement.md)
- [Auth And OAuth](../services/api-server/auth-and-oauth.md)
- [Internal API Surface](../services/api-server/internal-api-surface.md)
- [Player Stats And Match Results](../services/api-server/player-stats-and-match-results.md)
- [Runtime And Health](../services/api-server/runtime-and-health.md)
- [Client HTTP API Flow](../services/client/client-http-api-flow.md)
- [Local Profiles HTTP API](../services/player-data/local-profiles-http-api.md)
- [Profile Stats Flow](../services/player-data/profile-stats-flow.md)
- [Route Composition](../services/game-server/process/route-composition.md)

## Notes

The OpenAPI `/health` route describes the Rails API JSON health endpoint. The game-server process also exposes `GET /health`, but that route is documented as a game-server process route rather than the Rails JSON health contract.

The current OpenAPI contract is authoritative for HTTP shape, but only Rails tests currently use OpenAPI assertion helpers directly. Go-hosted player-data routes rely on Go handler/runtime tests and manual alignment with the shared OpenAPI file.

Detailed endpoint behavior belongs in the service docs. This protocol doc stays focused on contract authority, enforcement expectations, and update rules.
