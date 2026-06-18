# Runtime And Health

Parent index: [API Server](!README.md)

## Purpose

This document describes the API-server runtime and health surface for `services/api-server/`.

It covers the Rails API-only runtime, health-check endpoints, PostgreSQL connection configuration, Puma port behavior, and CI entrypoint checks that keep the service bootable and deployment-ready.

## Overview

The API server is a Rails API-only service.

It exposes two implemented health-related HTTP surfaces:

* `GET /health` is handled by `HealthController#show` and returns a small JSON payload with service status.
* `GET /up` is the standard Rails health route exposed for boot checks, load balancers, and uptime monitoring.

The service uses PostgreSQL through `config/database.yml`. Runtime database settings can be overridden with `SPACE_ROCKS_API_DATABASE_USERNAME`, `SPACE_ROCKS_API_DATABASE_PASSWORD`, `SPACE_ROCKS_API_DATABASE_HOST`, and `SPACE_ROCKS_API_DATABASE_PORT`.

Puma listens on port `3000` by default and can be overridden with `PORT`.

CI runs the service bootstrap, style, security, and test checks from `config/ci.rb`.

## Code root

* `services/api-server/`

## Responsibilities

* Own the Rails API-only runtime configuration for the API server.
* Own `GET /health` as the service-specific health endpoint.
* Expose `GET /up` as the Rails boot and load-balancer health route.
* Own PostgreSQL connection configuration for the API service.
* Support environment-based database overrides for local, test, and production deployments.
* Own Puma port configuration and `PORT` override behavior.
* Define the CI step sequence used to validate boot, style, security, and tests.
* Keep runtime and health concerns separate from auth and player-stats domain behavior.

## Does not own

* Authenticated-account identity or token verification.
* OAuth login flow or bearer-token issuance.
* Player stats or match-result persistence.
* Game-server simulation, match lifecycle, or websocket transport.
* Client-side health presentation.
* Database schema ownership for auth or player-data tables.

## Domain roles

The API server runtime participates in these roles:

* **Service availability boundary:** provides lightweight HTTP endpoints that let deployment systems verify the service is running.
* **Rails boot boundary:** exposes the standard Rails health check expected by load balancers and process supervisors.
* **Infrastructure configuration owner:** supplies runtime settings for the database adapter, Puma listener port, and CI entrypoint.

These roles are operational rather than gameplay-specific.

## Protocols and APIs

### `GET /health`

Caller:

* service monitors
* deployment checks
* manual runtime verification

Behavior:

* Handled by `HealthController#show`.
* Returns JSON with `status: "ok"` and `service: "space-rocks-api"`.

### `GET /up`

Caller:

* Rails boot checks
* load balancers
* uptime monitors

Behavior:

* Provided by the standard Rails health route in `config/routes.rb`.
* Returns success when the app boots without exceptions.

### Runtime configuration

Relevant configuration files:

* `config/application.rb` sets the app to API-only mode.
* `config/database.yml` defines PostgreSQL connection settings and environment overrides.
* `config/puma.rb` sets the default listener port and `PORT` override.
* `config/ci.rb` defines the service CI entrypoint.

## Data ownership

The runtime and health surface owns service-level configuration, not application business data.

Owned configuration inputs include:

* PostgreSQL adapter settings from `config/database.yml`
* `SPACE_ROCKS_API_DATABASE_USERNAME`
* `SPACE_ROCKS_API_DATABASE_PASSWORD`
* `SPACE_ROCKS_API_DATABASE_HOST`
* `SPACE_ROCKS_API_DATABASE_PORT`
* `PORT`

The runtime and health surface does not own auth records, player stats, match results, or other persisted domain data.

## Code map

### Routes and controller

* `services/api-server/config/routes.rb`
* `services/api-server/app/controllers/health_controller.rb`

### Runtime configuration

* `services/api-server/config/application.rb`
* `services/api-server/config/database.yml`
* `services/api-server/config/puma.rb`
* `services/api-server/config/ci.rb`

### Tests

* `services/api-server/test/controllers/health_controller_test.rb`

## Tests

* `GET /health` is covered by `services/api-server/test/controllers/health_controller_test.rb`.
* `assert_openapi_response!` in the controller test verifies the documented response shape stays aligned with the contract assertions used elsewhere in the service.

## Related docs

* [API Server](./!README.md)
* [Auth And OAuth](auth-and-oauth.md)
* [Internal API Surface](internal-api-surface.md)
* [Player Stats And Match Results](player-stats-and-match-results.md)

## Notes

`GET /health` is intentionally small and static so it can serve as a simple service-specific check without depending on auth, player data, or gameplay state.

`GET /up` remains the conventional Rails health endpoint for infrastructure tooling.
