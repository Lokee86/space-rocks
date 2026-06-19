# Player Data HTTP Hosting

Parent index: [Game Server Integrations](./!README.md)

## Purpose

This document describes how the game server hosts player-data HTTP handlers inside the game-server process.

This is the game-server integration boundary for player-data HTTP access. It covers route mounting, runtime construction, handler adaptation, auth-verifier bridging, and the limits of what the game-server process owns when it exposes player-data HTTP surfaces.

## Overview

The game server owns the HTTP process that currently exposes both realtime game entry points and player-data HTTP routes on the same `net/http` mux.

At startup, the game server:

1. creates the HTTP mux
2. creates the room manager
3. builds a player-data runtime
4. wraps that runtime as a match-result sink
5. builds the auth verifier
6. mounts game-server routes
7. mounts player-data HTTP handlers
8. starts `http.ListenAndServe(":8080", mux)`

The game server does not implement the player-data HTTP behavior directly. It constructs the player-data runtime and installs handlers from the player-data service package. The player-data service owns request validation, local profile behavior, profile lookup behavior, stats loading, local profile mutation, store routing, and JSON response shapes.

The game server's integration responsibility is to host those handlers and supply the process-level dependencies they need.

## Code root

* `services/game-server/`

## Responsibilities

The game-server player-data HTTP hosting integration owns:

* constructing the player-data runtime during game-server startup
* passing player-data runtime configuration from process environment and local store configuration
* mounting player-data HTTP routes on the game-server HTTP mux
* adapting the game-server auth verifier to the player-data HTTP auth verifier interface
* sharing the same runtime with match-result reporting and profile/local-profile HTTP handlers
* failing startup if the player-data runtime or match reporter cannot initialize
* keeping player-data HTTP route hosting separate from room, simulation, and WebSocket ownership

## Does not own

This integration does not own:

* player-data domain rules
* player-data request or response schema design
* local profile persistence internals
* account persistence internals
* guest stat storage internals
* Rails API-server auth policy
* client profile presentation
* match-result stat calculation
* WebSocket packet routing
* room lifecycle rules
* simulation authority

Those responsibilities belong to the player-data service, API-server/auth seams, client service, match-reporting seam, networking seam, room system, or simulation system.

## Domain roles

The integration currently acts as:

* HTTP route host for player-data surfaces.
* Player-data runtime composer.
* Auth-verifier adapter provider.
* Process-level bridge between the game-server HTTP mux and player-data handlers.
* Non-owner of player-data persistence and domain rules.

## Startup flow

The game-server process builds player-data support before it starts serving HTTP traffic.

```text
main()
  configure logging
  create mux
  create room manager
  build player-data runtime
  create player-data sink
  create match reporter
  build auth verifier
  mount health and WebSocket routes
  mount player-data HTTP routes
  listen on :8080
```

`buildPlayerDataRuntime()` creates a `playerdata.Runtime` using `playerdata.NewConfiguredRuntime()`.

The runtime configuration includes:

* `PLAYER_DATA_RAILS_BASE_URL`
* `PLAYER_DATA_RAILS_INTERNAL_TOKEN`
* the game-server-selected SQLite path
* the game-server-selected local store factory

If runtime initialization fails, the game server logs `player-data runtime initialization failed` and exits. If match reporter initialization fails, the game server logs `player-data reporter initialization failed` and exits.

## Protocols and APIs

### Hosted HTTP routes

The game server mounts these player-data HTTP routes:

```text
POST   /api/player-data/profile
GET    /api/player-data/local-profiles
POST   /api/player-data/local-profiles
PUT    /api/player-data/local-profiles/{local_profile_id}
DELETE /api/player-data/local-profiles/{local_profile_id}
GET    /api/player-data/local-profiles/default
PUT    /api/player-data/local-profiles/default
```

The mounted routes are served by handlers from `services/player-data/httpapi/`.

The game-server route table owns where these handlers are reachable. The player-data HTTP handlers own what each request means.

### Profile handler hosting

`POST /api/player-data/profile` is hosted by the game server and implemented by `httpapi.NewProfileHandler()`.

The game-server integration passes two dependencies into the profile handler:

* the shared player-data runtime
* an auth-verifier adapter backed by the game-server networking auth verifier

The profile handler supports these identity kinds:

```text
guest
local_profile
authenticated_account
```

Guest identity resolves to a guest profile with offline activity status.

Local profile identity requires a non-empty `local_profile_id` and resolves as a local pilot profile.

Authenticated account identity requires a valid bearer token and an available auth verifier. The game-server adapter converts the game-server networking verifier result into the player-data HTTP auth result shape.

The profile handler loads stats through the player-data runtime. Missing stats return zero stats rather than failing the request.

### Local profiles handler hosting

The local profiles routes are hosted by the game server and implemented by `httpapi.NewLocalProfilesHandler()`.

The game-server integration only passes the shared player-data runtime into this handler. Local profile behavior is owned by player-data runtime and store interfaces.

The local profiles handler supports:

* listing local profiles
* creating a local profile
* updating a local profile display name
* deleting a local profile
* reading the default local profile
* setting the default local profile

Local profile display names are validated by the player-data HTTP handler. The current accepted pattern is alphanumeric characters, underscore, and hyphen.

Local profile creation can seed the new profile from guest stats when `seed_from_guest_stats` is true. The handler asks the runtime for guest seed stats; it does not read guest storage directly.

## Runtime and store routing

The game server hosts a player-data runtime, but the runtime's store-routing behavior belongs to the player-data service.

The configured runtime routes identities across separate stores:

```text
authenticated_account -> account store
local_profile         -> local store
guest                 -> guest memory store
```

The account store is selected from configuration:

* when `PLAYER_DATA_RAILS_BASE_URL` is set, the runtime uses a Rails-backed store
* when it is not set, the runtime uses an in-memory account store

The local store is selected from the SQLite path and local store factory:

* when no SQLite path is configured, local profile operations are unavailable through a noop local store
* when a SQLite path and local store factory are configured, local profile operations use the configured local store

Guest data uses a guest memory store.

### Auth verifier bridge

The game server uses a small adapter to bridge from the networking auth verifier interface to the player-data HTTP auth verifier interface.

The adapter:

* receives a raw bearer token from the player-data profile handler
* delegates token verification to the game-server networking verifier
* maps the verifier result into player-data HTTP identity fields
* returns invalid/empty results when no verifier exists

This keeps the profile HTTP handler independent from game-server networking package types while still allowing hosted profile requests to reuse the same auth verification dependency as realtime networking.

## Data ownership

This integration does not own player-data persistence. It only hosts the HTTP route surface and wires runtime dependencies.

Data handled through the hosted HTTP routes includes:

* selected identity kind
* local profile ID
* account ID from auth verification
* display name/callsign presentation values
* local profile summaries
* default local profile selection
* player stats loaded through the runtime
* guest stat seeding input during local profile creation

Persistence ownership remains inside the player-data service stores.

## Failure behavior

Startup failures are fail-fast:

* player-data runtime initialization failure stops the game server
* player-data reporter initialization failure stops the game server

Request-time failures are owned by the player-data handlers:

* unsupported methods return `method_not_allowed`
* invalid request payloads return `invalid_request`
* unavailable local profile storage returns `local_profiles_unavailable`
* missing local profiles return `local_profile_not_found`
* missing or invalid authenticated-account credentials return `unauthorized`
* unavailable profile runtime behavior returns `profile_unavailable`

The game-server integration should not translate these errors after the request reaches the player-data handler.

## Code map

Primary game-server integration files:

* `services/game-server/cmd/game-server/main.go`
* `services/game-server/cmd/game-server/player_data_http.go`

Player-data handler files hosted by this integration:

* `services/player-data/httpapi/profile_handler.go`
* `services/player-data/httpapi/local_profiles_handler.go`

Player-data runtime and routing files used by the hosted handlers:

* `services/player-data/playerdata/runtime.go`
* `services/player-data/playerdata/configured_runtime.go`
* `services/player-data/playerdata/store_router.go`

Important adjacent ownership boundaries:

* `services/game-server/internal/networking/` owns WebSocket networking and the game-server auth verifier interface consumed by the adapter.
* `services/game-server/internal/matchreporting/` owns match-result reporter construction and match-result reporting flow.
* `services/player-data/httpapi/` owns player-data HTTP request handling.
* `services/player-data/playerdata/` owns player-data runtime, store routing, and persistence abstractions.

## Tests

Relevant current tests include:

* `services/player-data/httpapi/local_profiles_handler_test.go`
* `services/player-data/playerdata/runtime_test.go`
* `services/player-data/playerdata/configured_runtime_test.go`
* `services/player-data/playerdata/store_router_test.go`
* `services/player-data/playerdata/rails_store_test.go`
* `services/player-data/playerdata/noop_store_test.go`
* `services/player-data/playerdata/embeddedsqlite/sqlite_store_test.go`

The local profiles handler tests verify unavailable local profile storage behavior and guest-stat seeding behavior during local profile creation.

## Related docs

* [Game Server Integrations](./!README.md)
* [Game Server](../../!README.md)
* [Player Data](../../../player-data/!README.md)
* [Auth Verifier Integration](./auth-verifier-integration.md)
* [Match Result Reporting](./match-result-reporting.md)

## Notes

This document is intentionally scoped to game-server hosting of player-data HTTP handlers.

The player-data service should still receive separate documentation for local profile API behavior, profile stats flow, runtime and store routing, match-result sinks, and persistence details. This document should link to those docs rather than duplicate their full ownership.
