# Route Composition

Parent index: [Game Server Process](./!INDEX.md)

## Purpose

This document describes the game-server process route composition boundary.

It covers the HTTP mux routes registered by the game-server executable, the dependencies injected into those routes, and the current player-data HTTP piggy-back behavior.

## Overview

The game-server process composes one `net/http` mux and serves it on:

```text
:8080
```

Current route composition is owned by:

```text
services/game-server/cmd/game-server/main.go
```

At startup, the process:

```text
configure logging
create HTTP mux
create room manager
build player-data runtime
wrap player-data runtime as a match-result sink
build match-result reporter
build auth verifier
mount core game-server routes
mount player-data HTTP routes
start ListenAndServe(":8080", mux)
```

The route table is intentionally thin. It decides which HTTP paths are reachable from the game-server process and which runtime dependencies each mounted handler receives. It does not own WebSocket packet routing, room rules, simulation rules, auth internals, player-data request semantics, or player-data persistence.

The player-data HTTP routes currently piggy-back on the game-server process. They are mounted on the same mux and served from the same `:8080` listener as `/health` and `/ws`. The game server hosts those handlers, but the player-data service owns their request handling, validation, runtime routing, and backing-store selection.

## Code root

```text
services/game-server/cmd/game-server/
```

## Responsibilities

Route composition owns:

* creating the process-level HTTP mux
* registering the health route
* registering the realtime WebSocket route
* registering game-server-hosted player-data HTTP routes
* constructing the dependencies required before handlers are mounted
* passing the room manager into the WebSocket handler
* passing the auth verifier into WebSocket authentication and player-data profile reads
* passing the match-result reporter into WebSocket session lifecycle handling
* passing the shared player-data runtime into hosted player-data HTTP handlers
* keeping route mounting in the process entrypoint instead of scattering public HTTP routes across runtime packages
* starting the HTTP server with the composed mux

## Does not own

Route composition does not own:

* WebSocket upgrade internals
* WebSocket read/write loop behavior
* inbound packet-family routing
* outbound packet projection
* room lifecycle rules
* room admission policy beyond dependency injection
* simulation or gameplay rules
* match-result summary construction
* match-result persistence
* auth token verification mechanics
* Rails auth or account persistence
* player-data HTTP request validation
* player-data response schemas
* player-data store routing
* local profile SQLite persistence
* guest stat storage
* client-side route selection or WebSocket target selection
* packet source-of-truth files or generated packet code

## Domain roles

The game-server process route composer acts as:

* HTTP process root for the game server.
* Route table owner for the current game-server listener.
* Dependency composition root for route handlers.
* Host for player-data HTTP handlers inside the game-server process.
* Bridge between game-server process dependencies and service-owned handlers.
* Non-owner of route handler internals.

## Protocols and APIs

The game-server process currently exposes these HTTP routes from one mux.

| Method | Route | Handler owner | Purpose |
| --- | --- | --- | --- |
| `GET` | `/health` | game-server process | Process health response. |
| `GET` | `/ws` | game-server networking | Realtime WebSocket entrypoint. |
| `POST` | `/api/player-data/profile` | player-data HTTP API hosted by game-server | Profile readout for guest, local profile, or authenticated account identity. |
| `GET` | `/api/player-data/local-profiles` | player-data HTTP API hosted by game-server | List local profiles. |
| `POST` | `/api/player-data/local-profiles` | player-data HTTP API hosted by game-server | Create a local profile. |
| `PUT` | `/api/player-data/local-profiles/{local_profile_id}` | player-data HTTP API hosted by game-server | Update a local profile display name. |
| `DELETE` | `/api/player-data/local-profiles/{local_profile_id}` | player-data HTTP API hosted by game-server | Delete a local profile. |
| `GET` | `/api/player-data/local-profiles/default` | player-data HTTP API hosted by game-server | Read the default local profile selection. |
| `PUT` | `/api/player-data/local-profiles/default` | player-data HTTP API hosted by game-server | Set the default local profile selection. |

### Health route

The health route is mounted directly in `main.go`:

```text
GET /health
```

It writes:

```text
OK
```

The current health handler does not check room manager state, player-data runtime state, Rails availability, SQLite availability, or WebSocket readiness. It is a minimal process response.

### WebSocket route

The realtime route is mounted as:

```text
GET /ws
```

The process route composer passes these dependencies into the networking handler:

```text
room manager
auth verifier
match-result reporter
```

The networking package owns the WebSocket upgrade, origin policy, session creation, read loop, write loop, lifecycle ticker, room detachment on disconnect, and handoff into inbound packet routing.

Single-player and multiplayer currently share the same `/ws` route. Mode behavior is decided by packets, session state, identity, room admission, and room/game rules, not by separate WebSocket paths.

### Player-data HTTP routes

The player-data HTTP routes are mounted by the game-server process but implemented by `services/player-data/httpapi`.

The process route composer builds one player-data runtime and passes it into:

```text
newPlayerDataProfileHTTPHandler(...)
newPlayerDataLocalProfilesHTTPHandler(...)
```

The profile handler also receives an auth-verifier adapter so authenticated-account profile reads can verify bearer tokens through the same game-server auth verifier used by WebSocket authentication.

These routes are reachable through the game-server HTTP listener, but the game-server process does not own their domain behavior. The player-data service owns profile reads, local profile CRUD/default behavior, request validation, response payloads, mode/identity validation, and store routing.

## Dependency composition

Route composition depends on startup-time dependency construction.

Current dependency flow:

```text
mux := http.NewServeMux()

rooms := networking.NewRoomManager()
defer rooms.StopAll()

playerDataRuntime := buildPlayerDataRuntime()
playerDataSink := newPlayerDataSink(playerDataRuntime)
reporter := matchreporting.NewRuntimeReporter(playerDataSink)

authVerifier := buildAuthVerifierFromEnv()

mount /health
mount /ws with rooms, authVerifier, reporter
mount player-data HTTP routes with playerDataRuntime and authVerifier
```

### Room manager dependency

The route composer creates one room manager for the process and passes it into the WebSocket handler.

The room manager is process-wide for accepted WebSocket sessions. It owns room lookup and room aggregate access through the networking and rooms packages, not through `main.go`.

### Player-data runtime dependency

The route composer builds one player-data runtime before routes are mounted.

That runtime is shared by:

* match-result reporting
* profile HTTP reads
* local profile HTTP routes

This is the current player-data piggy-back/dependency point. Player-data behavior is available through the game-server process only because the game-server process constructs the player-data runtime and mounts player-data HTTP handlers on its mux.

The player-data runtime is configured with:

```text
PLAYER_DATA_RAILS_BASE_URL
PLAYER_DATA_RAILS_INTERNAL_TOKEN
playerDataLocalStorePath()
playerDataLocalStoreFactory()
```

In the standard no-tag development build, the local store factory creates the embedded SQLite local store and initializes its schema. In `noembeddedsqlite` builds, the local store path is empty and the runtime uses unavailable/noop local-profile behavior.

If player-data runtime construction fails, the game-server logs:

```text
player-data runtime initialization failed
```

and exits before any route is served.

### Match-result reporter dependency

The route composer wraps the player-data runtime in a player-data sink and builds a `matchreporting.RuntimeReporter`.

That reporter is passed into the WebSocket handler. Networking and rooms use it later when a resolved match result must be reported before or during room lifecycle transitions.

If reporter construction fails, the game-server logs:

```text
player-data reporter initialization failed
```

and exits before any route is served.

### Auth verifier dependency

The route composer builds the auth verifier from:

```text
API_SERVER_BASE_URL
GAME_SERVER_INTERNAL_TOKEN
```

If either value is missing, no verifier is installed.

If verifier construction fails, the game-server logs:

```text
auth verifier initialization failed
```

and continues with no verifier.

A nil verifier affects downstream route behavior:

* WebSocket authentication returns token-verification unavailable.
* Multiplayer create/join admission fails closed with `auth_unavailable`.
* Authenticated-account profile reads through `/api/player-data/profile` fail as unauthorized.
* Guest and local-profile player-data HTTP behavior can still run when their required player-data runtime support exists.

## Data ownership

Route composition owns no durable data.

It creates and wires runtime dependencies that later access or mutate data through their owning boundaries.

Data that crosses the route-composition boundary includes:

```text
HTTP method and path
HTTP request and response objects
WebSocket upgrade request
player-data profile HTTP requests
player-data local profile HTTP requests
auth verifier dependency
room manager dependency
match-result reporter dependency
player-data runtime dependency
```

The game-server process route table does not choose player-data backing stores per request. Store selection belongs to `services/player-data/playerdata`.

The game-server process route table also does not persist route configuration, player profiles, room state, auth state, match results, or gameplay state.

## Failure behavior

Startup failures:

* player-data runtime initialization failure stops the process
* match-result reporter initialization failure stops the process
* `http.ListenAndServe(":8080", mux)` failure logs `server stopped` and exits
* auth verifier initialization failure logs an error and continues with no verifier

Request-time failures:

* WebSocket upgrade failures are handled by the networking WebSocket handler.
* WebSocket packet decode and routing failures are handled by networking and inbound packet routing.
* Player-data HTTP request failures are handled by player-data HTTP handlers.
* Authenticated-account profile verification failures are handled by the player-data profile handler through the game-server auth-verifier adapter.
* Room and gameplay request rejections are handled below the WebSocket route by networking, rooms, and game packages.

Route composition should not translate or reinterpret handler-owned failures after a request reaches the mounted handler.

## Code map

Primary process files:

* `services/game-server/cmd/game-server/main.go` - Creates the mux, constructs process dependencies, mounts routes, starts HTTP serving, and defines the health handler.
* `services/game-server/cmd/game-server/auth_config.go` - Builds the optional game-server auth verifier from environment configuration.
* `services/game-server/cmd/game-server/player_data_http.go` - Builds the player-data runtime, player-data sink, hosted player-data HTTP handlers, and auth-verifier adapter.
* `services/game-server/cmd/game-server/player_data_local_store_dev.go` - Standard build local-store path and embedded SQLite local-store factory.
* `services/game-server/cmd/game-server/player_data_local_store_noembeddedsqlite.go` - Restricted build local-store disablement.

Mounted WebSocket implementation:

* `services/game-server/internal/networking/websocket.go` - WebSocket handler construction and connection runtime.
* `services/game-server/internal/networking/websocket_origin.go` - WebSocket origin allowlist.
* `services/game-server/internal/networking/websocket_session.go` - Per-connection session construction.
* `services/game-server/internal/networking/client_packet_router.go` - Inbound packet routing handoff after WebSocket reads.

Player-data handlers hosted by the route table:

* `services/player-data/httpapi/profile_handler.go`
* `services/player-data/httpapi/local_profiles_handler.go`

Player-data runtime dependencies:

* `services/player-data/playerdata/configured_runtime.go`
* `services/player-data/playerdata/runtime.go`
* `services/player-data/playerdata/store_router.go`
* `services/player-data/playerdata/guest_memory_store.go`
* `services/player-data/playerdata/noop_store.go`
* `services/player-data/playerdata/rails_store.go`
* `services/player-data/playerdata/embeddedsqlite/sqlite_store.go`

Match-result reporting dependencies:

* `services/game-server/internal/matchreporting/runtime_reporter.go`
* `services/game-server/internal/matchreporting/mapper.go`
* `services/game-server/internal/rooms/match_result_reporter.go`

Auth dependencies:

* `services/game-server/internal/authclient/client.go`
* `services/game-server/internal/authclient/types.go`
* `services/game-server/internal/networking/session_auth.go`
* `services/game-server/internal/networking/session_admission.go`
* `services/game-server/internal/networking/session_identity.go`

Important non-ownership boundaries:

* `services/game-server/internal/networking/` owns WebSocket transport, session runtime, and packet handoff.
* `services/game-server/internal/rooms/` owns room membership, room lifecycle, and match lifecycle.
* `services/game-server/internal/game/` owns authoritative simulation and generated gameplay packet projection.
* `services/game-server/internal/matchreporting/` owns game-server to player-data match-result command mapping.
* `services/player-data/httpapi/` owns hosted player-data HTTP request behavior.
* `services/player-data/playerdata/` owns player-data runtime behavior and store routing.
* `services/api-server/` owns Rails auth authority and authenticated-account persistence.

## Tests

There are no dedicated route-registration tests under `services/game-server/cmd/game-server/` in the current tree.

Relevant adjacent tests include:

* `services/game-server/internal/authclient/client_test.go`
* `services/game-server/internal/networking/websocket_test.go`
* `services/game-server/internal/networking/session_auth_test.go`
* `services/game-server/internal/networking/session_identity_test.go`
* `services/game-server/internal/networking/room_error_test.go`
* `services/game-server/internal/matchreporting/runtime_reporter_test.go`
* `services/game-server/internal/matchreporting/mapper_test.go`
* `services/player-data/httpapi/local_profiles_handler_test.go`
* `services/player-data/playerdata/configured_runtime_test.go`
* `services/player-data/playerdata/runtime_test.go`
* `services/player-data/playerdata/store_router_test.go`
* `services/player-data/playerdata/noop_store_test.go`
* `services/player-data/playerdata/rails_store_test.go`
* `services/player-data/playerdata/embeddedsqlite/sqlite_store_test.go`

Suggested verification for this boundary is split by service module:

```text
cd services/game-server && go test -buildvcs=false ./cmd/game-server ./internal/authclient ./internal/networking ./internal/matchreporting ./internal/rooms
cd services/player-data && go test ./httpapi ./playerdata/...
```

## Related docs

* [Game Server Process](./!INDEX.md)
* [Game Server](../!INDEX.md)
* [Game Server Networking](../networking/!INDEX.md)
* [WebSocket Session Lifecycle](../networking/websocket-session-lifecycle.md)
* [Inbound Packet Routing](../networking/inbound-packet-routing.md)
* [Outbound Message Flow](../networking/outbound-message-flow.md)
* [Auth Verifier Integration](../integrations/auth-verifier-integration.md)
* [Player Data HTTP Hosting](../integrations/player-data-http-hosting.md)
* [Match Result Reporting](../integrations/match-result-reporting.md)
* [Game Server Observability](../observability/!INDEX.md)
* [Player Data](../../player-data/!INDEX.md)
* [API Server](../../api-server/!INDEX.md)
* [HTTP Contract Enforcement](../../../protocol/http-contract-enforcement.md)

## Notes

The current game-server route table is process-local and hard-coded in `main.go`. There is no route registry package and no environment-configured listen address.

The player-data HTTP surface is intentionally documented here only as a route-composition dependency. Detailed hosted-handler behavior belongs in player-data HTTP hosting and player-data service docs.

The game-server entrypoint should register routes, configure dependencies, and start the process. Reusable simulation, packet routing, player-data behavior, and service integrations should remain outside `main.go`.
