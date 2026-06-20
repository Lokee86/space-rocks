# Service Startup

Parent index: [Game Server Process](./!INDEX.md)

## Purpose

This document describes the game-server process startup boundary.

It covers the executable composition root, startup ordering, environment-backed dependency construction, HTTP mux creation, route mounting summary, listen behavior, and the current in-process player-data dependency that piggy-backs on the game-server process.

## Overview

The game server starts from `services/game-server/cmd/game-server/main.go`.

The process currently owns one HTTP server on `:8080`. That server exposes the health route, the WebSocket gameplay route, and player-data HTTP routes on the same `net/http` mux.

Startup is not only game-server-local setup. The game server also constructs a `services/player-data` runtime before it begins listening. That runtime is a sibling Go module dependency, not game-server internals. The game-server process hosts it in-process for now and passes it to:

* player-data HTTP handlers mounted on the game-server mux
* the match-result reporter sink used by room and WebSocket lifecycle paths

The player-data runtime is therefore a startup dependency of the game-server process. If the runtime cannot be constructed, the game server logs the failure and exits before binding `:8080`.

Current high-level startup flow:

```text
main()
  configure logging
  create HTTP mux
  create room manager
  defer room manager StopAll
  build in-process player-data runtime
  wrap player-data runtime as a sink
  build match-result reporter
  build optional auth verifier
  mount health route
  mount WebSocket route
  mount player-data HTTP routes
  listen on :8080
```

## Code root

`services/game-server/cmd/game-server/`

Supporting service roots:

* `services/game-server/`
* `services/player-data/`
* `services/api-server/`

## Responsibilities

The game-server startup boundary owns:

* configuring game-server logging from environment variables
* creating the process HTTP mux
* creating the room manager
* registering process shutdown cleanup through `defer rooms.StopAll()`
* constructing the in-process player-data runtime
* selecting player-data local-store startup behavior from build tags and store configuration
* wrapping the player-data runtime as a match-result sink
* constructing the match-result reporter used by room lifecycle reporting
* constructing the optional API-server auth verifier from environment configuration
* mounting the health route
* mounting the WebSocket route with room manager, auth verifier, and match reporter dependencies
* mounting game-server-hosted player-data HTTP routes
* logging the server start event
* starting `http.ListenAndServe(":8080", mux)`
* exiting when fatal startup or listen errors occur

## Does not own

The startup boundary does not own:

* WebSocket session behavior after route entry
* inbound packet routing
* outbound packet routing
* room lifecycle rules
* room cleanup policy beyond registering `StopAll()` for process exit
* simulation mechanics
* match-result summary calculation
* player-data request validation
* player-data store routing rules
* player-data persistence internals
* Rails auth and account persistence
* client connection behavior
* graceful shutdown orchestration beyond the current deferred room cleanup
* route documentation detail that belongs in route-composition documentation

## Domain roles

The game-server startup boundary currently acts as:

* process composition root for the game-server executable
* HTTP mux owner for the local game-server process
* dependency constructor for rooms, auth, match reporting, and player-data hosting
* in-process host for the player-data runtime
* fail-fast gate for player-data runtime initialization
* optional API-server auth-verifier consumer
* process-level owner of the `:8080` listen address

The game server remains the realtime gameplay authority. Player-data remains the identity and store-routing authority for player-shaped durable data. API-server remains the authenticated-account auth and Rails/Postgres persistence authority.

## Startup sequence

### Logging configuration

Startup begins by configuring logging:

```go
logging.Configure(os.Getenv(logging.EnvGlobalLevel))
```

`LOG_LEVEL` supplies the global default. Category overrides are read inside the logging package:

```text
LOG_GAME
LOG_NETWORK
LOG_ROOMS
LOG_SERVER
```

The default level is `warn`. The `server starting` event is logged through `logging.Server.Info`, so it is visible only when server or global logging allows info-level output.

### HTTP mux creation

Startup creates one `http.ServeMux`:

```go
mux := http.NewServeMux()
```

All process routes are mounted on this mux before the process starts listening.

### Room manager creation

Startup creates one room manager through the networking boundary:

```go
rooms := networking.NewRoomManager()
defer rooms.StopAll()
```

The room manager is passed into the WebSocket handler. The deferred `StopAll()` call is the current process-exit cleanup hook for room cleanup timers and running game instances.

Detailed shutdown behavior belongs in service-shutdown documentation.

### Player-data runtime construction

Startup builds the player-data runtime before constructing the match reporter or mounting player-data HTTP routes:

```go
playerDataRuntime, err := buildPlayerDataRuntime()
```

`buildPlayerDataRuntime()` calls `playerdata.NewConfiguredRuntime()` with:

```text
PLAYER_DATA_RAILS_BASE_URL
PLAYER_DATA_RAILS_INTERNAL_TOKEN
playerDataLocalStorePath()
playerDataLocalStoreFactory()
```

This is the current player-data piggy-back point. The game-server process hosts the player-data runtime in-process, but the runtime implementation belongs to `services/player-data`.

The player-data module is a sibling Go module dependency of the game-server module:

```text
services/game-server/go.mod
  require github.com/Lokee86/space-rocks/player-data v0.0.0
  replace github.com/Lokee86/space-rocks/player-data => ../player-data
```

The game-server startup code should not bypass this service boundary by directly writing SQLite, Rails, Postgres, or player-data tables.

### Player-data local store selection

Local-profile startup behavior depends on the game-server build tag path.

Default development build:

```text
//go:build !noembeddedsqlite
```

The game server supplies:

```text
SQLitePath: services/player-data/data/player-data.sqlite3
LocalStoreFactory: embedded SQLite store factory
```

The embedded SQLite factory creates the store and initializes schema during startup. If that fails, player-data runtime construction fails and the game-server process exits.

Restricted or deployment-style build:

```text
//go:build noembeddedsqlite
```

The game server supplies:

```text
SQLitePath: ""
LocalStoreFactory: nil
```

With no SQLite path, `services/player-data` uses a noop local store. The game-server process can still start, but local-profile operations are unavailable at request time through the player-data handler behavior.

### Player-data account store selection

Account-routed player-data behavior is selected by `services/player-data` runtime configuration:

```text
PLAYER_DATA_RAILS_BASE_URL set   -> Rails-backed account store
PLAYER_DATA_RAILS_BASE_URL empty -> in-memory account store
```

`PLAYER_DATA_RAILS_INTERNAL_TOKEN` is passed to the Rails-backed store when Rails backing is configured.

Missing Rails player-data configuration does not stop game-server startup. It changes the account-store backing selected by the player-data runtime.

### Player-data sink and match reporter construction

After runtime construction, startup wraps the runtime as a player-data sink:

```go
playerDataSink := newPlayerDataSink(playerDataRuntime)
```

The match reporter is then constructed from that sink:

```go
reporter, err := matchreporting.NewRuntimeReporter(playerDataSink)
```

The reporter is passed into the WebSocket route. Later, room and WebSocket lifecycle paths use it to report resolved match results into the player-data runtime.

If reporter construction fails, startup logs `player-data reporter initialization failed` and exits.

### Auth verifier construction

Startup builds the auth verifier after the player-data runtime and reporter:

```go
authVerifier := buildAuthVerifierFromEnv()
```

The verifier is configured by:

```text
API_SERVER_BASE_URL
GAME_SERVER_INTERNAL_TOKEN
```

If either value is missing, startup returns a nil verifier and continues. If verifier construction fails, startup logs `auth verifier initialization failed` and continues with no verifier.

This makes API-server auth verification optional at process startup, but multiplayer create/join requests still fail closed later when verification is unavailable.

The same verifier is passed into:

* WebSocket auth/session handling
* the player-data profile HTTP auth-verifier adapter

## Player-data piggy-back dependency

The current player-data arrangement is intentionally co-located but still separate by service boundary.

Current dependency shape:

```text
game-server process
  imports services/player-data module
  constructs playerdata.Runtime during startup
  mounts player-data HTTP handlers on game-server mux
  sends match-result commands into player-data RuntimeSink
```

This means:

* There is no separate player-data server process required for the current local game-server startup path.
* The game server must successfully construct a player-data runtime before it starts listening.
* The same runtime instance is shared by hosted player-data HTTP routes and match-result reporting.
* The game server owns hosting and dependency injection, not player-data behavior.
* Player-data owns runtime dispatch, identity-based store routing, local-profile behavior, guest behavior, Rails adapter behavior, and store persistence.
* Future extraction to a separate player-data process should replace the in-process transport, not move player-data ownership into game-server internals.

Failure implications:

```text
player-data runtime init fails -> game-server startup fails
player-data reporter init fails -> game-server startup fails
Rails player-data env missing -> game-server starts with non-Rails account backing
embedded SQLite disabled -> game-server starts with noop local-profile store
embedded SQLite init fails in default build -> game-server startup fails
```

## Protocols and APIs

The startup boundary mounts process-level runtime surfaces. Detailed behavior belongs to the owning networking, integration, and player-data docs.

### Health route

```text
GET /health
```

The health route is implemented directly in `main.go` and returns:

```text
OK
```

It is a process health check only. It does not verify room state, player-data persistence health, API-server reachability, or WebSocket behavior.

### WebSocket route

```text
GET /ws
```

The WebSocket route is mounted with:

```text
room manager
auth verifier
match-result reporter
```

The route exists so clients can establish realtime sessions. The startup boundary only constructs and injects dependencies. WebSocket upgrade behavior, packet routing, session identity, lobby flow, gameplay input, and outbound state packets belong to game-server networking docs.

### Player-data HTTP routes

Startup mounts these player-data routes on the same game-server mux:

```text
POST   /api/player-data/profile
GET    /api/player-data/local-profiles
POST   /api/player-data/local-profiles
PUT    /api/player-data/local-profiles/{local_profile_id}
DELETE /api/player-data/local-profiles/{local_profile_id}
GET    /api/player-data/local-profiles/default
PUT    /api/player-data/local-profiles/default
```

The startup boundary owns only route mounting and dependency injection.

The player-data HTTP handlers own request parsing, validation, response shape, local profile operations, profile lookup, guest stat seeding, and request-time errors.

### Listen address

The process listens with:

```go
http.ListenAndServe(":8080", mux)
```

The address is currently hard-coded in `main.go`.

If `ListenAndServe` returns an error, startup logs `server stopped` with the address and exits with status `1`.

## Data ownership

Startup constructs access to data-owning systems, but it does not own their stored data.

### Game-server-owned runtime state

The game-server process owns:

* the HTTP mux
* room manager runtime state
* WebSocket runtime state after sessions connect
* process-local logging configuration
* the injected match-result reporter dependency

### Player-data-owned runtime state

The in-process player-data runtime owns:

* account/local/guest store routing
* guest memory store
* local profile store behavior
* authenticated-account Rails adapter behavior
* player-data packet dispatch
* profile/stat loading behavior
* match-result store mutation behavior

### API-server-owned durable state

API-server owns authenticated-account auth and Rails/Postgres persistence.

The game server can consume API-server through:

* auth verifier calls for token verification
* player-data Rails adapter calls made by `services/player-data`

The game server startup boundary does not directly access Rails tables or Postgres records.

### Local SQLite state

In the default embedded SQLite build, local-profile data is stored through `services/player-data` at:

```text
services/player-data/data/player-data.sqlite3
```

The game-server startup boundary supplies the path and factory. The player-data SQLite store owns schema initialization, reads, writes, and deletes.

## Failure behavior

Startup is fail-fast for dependencies that are required before serving process routes:

```text
player-data runtime initialization failure -> log and exit
player-data reporter initialization failure -> log and exit
ListenAndServe failure -> log and exit
```

Startup is tolerant for optional auth verifier configuration:

```text
missing API_SERVER_BASE_URL -> nil verifier, process continues
missing GAME_SERVER_INTERNAL_TOKEN -> nil verifier, process continues
auth verifier construction error -> log, nil verifier, process continues
```

Startup is tolerant for missing Rails player-data backing:

```text
missing PLAYER_DATA_RAILS_BASE_URL -> player-data uses memory account store
```

Startup behavior should not be confused with request-time behavior. A process can start successfully while later requests fail because auth verification, local profiles, or backing stores are unavailable for that request.

## Code map

Primary startup files:

* `services/game-server/cmd/game-server/main.go`
* `services/game-server/cmd/game-server/auth_config.go`
* `services/game-server/cmd/game-server/player_data_http.go`
* `services/game-server/cmd/game-server/player_data_local_store_dev.go`
* `services/game-server/cmd/game-server/player_data_local_store_noembeddedsqlite.go`

Game-server dependencies constructed during startup:

* `services/game-server/internal/logging/logger.go`
* `services/game-server/internal/networking/rooms.go`
* `services/game-server/internal/networking/websocket.go`
* `services/game-server/internal/authclient/client.go`
* `services/game-server/internal/authclient/types.go`
* `services/game-server/internal/matchreporting/runtime_reporter.go`

Player-data runtime dependencies constructed or consumed during startup:

* `services/player-data/playerdata/configured_runtime.go`
* `services/player-data/playerdata/runtime.go`
* `services/player-data/playerdata/store_router.go`
* `services/player-data/playerdata/rails_store.go`
* `services/player-data/playerdata/guest_memory_store.go`
* `services/player-data/playerdata/noop_store.go`
* `services/player-data/playerdata/embeddedsqlite/sqlite_store.go`
* `services/player-data/httpapi/profile_handler.go`
* `services/player-data/httpapi/local_profiles_handler.go`

Module dependency files:

* `services/game-server/go.mod`
* `services/player-data/go.mod`

Important non-ownership boundaries:

* `services/game-server/cmd/game-server/` owns process composition.
* `services/game-server/internal/networking/` owns WebSocket behavior after route entry.
* `services/game-server/internal/rooms/` owns room state and room cleanup behavior.
* `services/game-server/internal/matchreporting/` owns game-server-to-player-data match-result reporting.
* `services/player-data/playerdata/` owns player-data runtime and store routing.
* `services/player-data/httpapi/` owns player-data HTTP handler behavior.
* `services/api-server/` owns Rails auth and authenticated-account persistence.

## Tests

There are no direct `cmd/game-server` startup composition tests identified for this boundary.

Relevant lower-level tests include:

* `services/game-server/internal/authclient/client_test.go`
* `services/game-server/internal/matchreporting/runtime_reporter_test.go`
* `services/game-server/internal/networking/session_auth_test.go`
* `services/game-server/tests/networking/auth_test.go`
* `services/game-server/tests/networking/auth_admission_test.go`
* `services/player-data/playerdata/configured_runtime_test.go`
* `services/player-data/playerdata/configured_runtime_embedded_sqlite_test.go`
* `services/player-data/playerdata/runtime_test.go`
* `services/player-data/playerdata/store_router_test.go`
* `services/player-data/playerdata/rails_store_test.go`
* `services/player-data/playerdata/noop_store_test.go`
* `services/player-data/playerdata/guest_memory_store_test.go`
* `services/player-data/playerdata/embeddedsqlite/sqlite_store_test.go`
* `services/player-data/httpapi/local_profiles_handler_test.go`

Useful verification commands:

```bash
cd services/game-server && go test -buildvcs=false ./...
cd services/player-data && go test ./...
cd services/player-data && go test -tags noembeddedsqlite ./...
```

## Related docs

* [Game Server Process](./!INDEX.md)
* [Game Server](../!INDEX.md)
* [Game Server Integrations](../integrations/!INDEX.md)
* [Player Data HTTP Hosting](../integrations/player-data-http-hosting.md)
* [Auth Verifier Integration](../integrations/auth-verifier-integration.md)
* [Match Result Reporting](../integrations/match-result-reporting.md)
* [Game Server Networking](../networking/!INDEX.md)
* [Game Server Rooms](../rooms/!INDEX.md)
* [Player Data](../../player-data/!INDEX.md)
* [API Server](../../api-server/!INDEX.md)

## Notes

This document is scoped to startup composition. Route-by-route ownership belongs in route-composition documentation. Process cleanup and shutdown behavior belong in service-shutdown documentation.

The game server currently hosts player-data in-process while `services/player-data` remains a separate service boundary.

The player-data piggy-back model should not be treated as permission to move player-data persistence into game-server internals. The current implementation co-locates runtime execution, not ownership.
