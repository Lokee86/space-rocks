# Runtime And Store Routing

Parent index: [Player Data](./!INDEX.md)

## Purpose

This document describes the player-data service runtime, mode policy, store-routing seam, and backing-store selection behavior.

It exists to keep player-data routing owned by `services/player-data` instead of by the game server, client, API server, or gameplay simulation.

## Overview

`services/player-data` owns the runtime boundary that accepts player-data commands, validates packet-level mode and identity combinations, routes reads and writes by identity kind, and delegates storage behavior to the configured store.

The current runtime is hosted in-process by the game-server process. That hosting detail does not make player-data a game-server subsystem. The game server builds the runtime, mounts HTTP handlers, and reports match results through a sink, but it does not choose the backing store for a player request and does not read or write SQLite, Rails, or Postgres tables directly.

The runtime routes three identity kinds:

| Identity kind           | Runtime route | Backing behavior                                                                                                         |
| ----------------------- | ------------- | ------------------------------------------------------------------------------------------------------------------------ |
| `guest`                 | guest store   | runtime-local transient memory                                                                                           |
| `local_profile`         | local store   | embedded SQLite in the standard no-tag development build, or unavailable/noop behavior when no local store is configured |
| `authenticated_account` | account store | Rails/API-backed store when configured, otherwise in-memory account store                                                |

The store router is the main seam. Callers provide player identity and request context; player-data selects the store from identity kind.

```text
player-data command or HTTP handler
-> playerdata.Runtime
-> Dispatcher or direct runtime method
-> StoreRouter
-> account store | local store | guest store
```

## Code root

* `services/player-data/`

## Responsibilities

The player-data runtime owns:

* runtime construction from explicit store configuration
* packet dispatch for player-data runtime commands
* packet-level mode and identity validation
* identity-kind based store routing
* direct stat loading through `Runtime.LoadStats`
* local profile management calls exposed through runtime methods
* guest stat seeding for local profile creation
* the account/local/guest store interface boundary
* local profile unavailability behavior when no local-profile-capable store is configured
* Rails adapter calls for authenticated-account stats reads and match-result writes
* embedded SQLite local profile storage when the SQLite build path is enabled
* duplicate match-result handling at the store layer
* normalized player-data stats returned to callers

## Does not own

The player-data runtime does not own:

* game-server room lifecycle
* game-server match-over decisions
* authoritative score, death, win, or match facts
* WebSocket admission
* OAuth login
* bearer token issuance
* Rails auth tables
* Rails/Postgres physical schema
* client profile presentation
* client callsign display rules
* live gameplay state
* asteroid, projectile, pickup, ship, room, or simulation entities
* leaderboard eligibility
* online account trust policy

The game server remains the gameplay authority. Rails/API remains the authenticated-account identity and online persistence authority. Player-data owns routing and store-facing player-data behavior.

## Domain roles

Player-data participates in the service boundary that routes player-data runtime and store access.

Its service role is:

* runtime/store-routing boundary for player-data reads and writes
* mode and identity gate for packet-dispatched and direct runtime calls
* guest, local, and account route selector based on identity kind
* local-profile persistence boundary for local profile management operations
* Rails adapter caller for authenticated-account reads and match-result writes
* non-owner of gameplay authority, match outcomes, or simulation state

The game server may host the runtime in-process, but player-data still owns the routing and persistence seam rather than gameplay authority.

## Runtime construction

The core runtime is constructed by `playerdata.NewRuntime`.

```text
NewRuntime(Config{Store: store})
```

The runtime requires a non-nil `Store`. It installs a dispatcher over that store and exposes both encoded-payload handling and direct Go methods.

`playerdata.NewConfiguredRuntime` builds the normal routed runtime:

```text
account store:
  PLAYER_DATA_RAILS_BASE_URL set -> RailsStore
  PLAYER_DATA_RAILS_BASE_URL empty -> MemoryStore

local store:
  SQLitePath empty -> NoopStore
  SQLitePath set -> LocalStoreFactory(SQLitePath)

guest store:
  GuestMemoryStore
```

The configured runtime then wraps those stores in `StoreRouter`.

The core `playerdata` package does not import embedded SQLite directly. SQLite construction is injected through `LocalStoreFactory`. This keeps the core runtime compilable without the embedded SQLite package and keeps build-tag behavior at the process composition boundary.

`playerdata.NewRuntimeFromEnv` reads only:

* `PLAYER_DATA_RAILS_BASE_URL`
* `PLAYER_DATA_RAILS_INTERNAL_TOKEN`

It does not configure the local SQLite path. The game-server composition root supplies the SQLite path and local store factory when embedded local storage is available.

## Game-server hosting

The game server currently hosts the player-data runtime in-process.

At startup, `services/game-server/cmd/game-server/main.go`:

1. creates the HTTP mux
2. creates the room manager
3. builds the player-data runtime
4. wraps it in a runtime sink
5. creates the match-result reporter
6. builds the auth verifier
7. mounts game routes
8. mounts player-data HTTP routes
9. listens on `:8080`

`buildPlayerDataRuntime()` passes this configuration into `playerdata.NewConfiguredRuntime`:

* Rails base URL from `PLAYER_DATA_RAILS_BASE_URL`
* Rails internal token from `PLAYER_DATA_RAILS_INTERNAL_TOKEN`
* local SQLite path from the game-server build-specific local-store file
* local store factory from the game-server build-specific local-store file

In the standard no-tag development build, the game-server local-store factory creates `playerdata/embeddedsqlite.Store`, initializes its schema, and returns it as the local store.

In `noembeddedsqlite` builds, the game-server passes an empty SQLite path and no local store factory. That makes local profile management unavailable through runtime local-profile methods.

## Store routing

`StoreRouter` owns identity-based store selection.

```text
authenticated_account -> accountStore
local_profile         -> localStore
guest                 -> guestStore
```

The router implements the common `Store` interface:

```text
LoadStats(identity)
RecordMatchResult(command)
```

It also forwards local profile management calls only when the configured local store implements `LocalProfileStore`.

If the local store does not implement `LocalProfileStore`, local profile management returns `ErrLocalProfileUnavailable`.

## Mode policy

`ValidateModeIdentity` defines the current play-mode and identity-kind matrix for packet-dispatched player-data commands.

| Play mode                | Guest    | Local Profile | Authenticated Account |
| ------------------------ | -------- | ------------- | --------------------- |
| `single_player`          | allowed  | allowed       | rejected              |
| `multiplayer`            | rejected | rejected      | allowed               |
| `multiplayer_simulation` | rejected | rejected      | allowed               |

Unknown play modes, missing identity kinds, and unknown identity kinds are rejected.

The dispatcher applies this validation to encoded `player_data_load_stats` and `player_data_record_match_result` packets before it calls the store.

Direct runtime methods such as `LoadStats`, `ListLocalProfiles`, and `CreateLocalProfile` do not receive a play-mode context. They rely on the caller or HTTP handler to resolve the identity and request shape before calling the runtime.

## Runtime surfaces

### Encoded packet surface

The encoded runtime surface is:

```text
Runtime.Handle(payload []byte) ([]byte, error)
```

It is used by the in-process runtime sink:

```text
RuntimeSink.HandlePlayerDataCommand(payload)
```

The game-server match reporter sends encoded player-data commands through this sink.

Current generated packet types are:

```text
player_data_load_stats
player_data_load_stats_result
player_data_record_match_result
player_data_record_match_result_result
```

The packet surface carries:

* player-data identity
* play-mode request context
* match result identifiers
* authoritative score
* authoritative ship deaths
* winner flag
* returned aggregate stats
* accepted/duplicate/found status
* error code and message when rejected

This packet surface does not own gameplay authority. It receives already-resolved player-data facts and routes them to the selected store.

### Direct runtime methods

The runtime also exposes direct Go methods used by hosted HTTP handlers and composition code:

```text
LoadStats(identity)
LocalProfileSeedStats(seedFromGuestStats)
ListLocalProfiles()
CreateLocalProfile(localProfileID, displayName, stats)
DeleteLocalProfile(localProfileID)
UpdateLocalProfileDisplayName(localProfileID, displayName)
GetDefaultLocalProfile()
SetDefaultLocalProfile(identityKind, localProfileID)
```

These methods keep HTTP handlers from reaching into concrete stores.

### HTTP handler consumption

The player-data HTTP handlers live under `services/player-data/httpapi/`, but they are currently hosted by the game-server process.

Current hosted routes are:

```text
POST   /api/player-data/profile
GET    /api/player-data/local-profiles
POST   /api/player-data/local-profiles
PUT    /api/player-data/local-profiles/{local_profile_id}
DELETE /api/player-data/local-profiles/{local_profile_id}
GET    /api/player-data/local-profiles/default
PUT    /api/player-data/local-profiles/default
```

The HTTP handlers call runtime methods. They do not choose SQLite, Rails, or guest memory directly.

## Store behavior

### Guest store

`GuestMemoryStore` owns guest stats for the runtime instance.

Guest behavior:

* requires `identity_kind = guest`
* stores one transient aggregate stats object per runtime
* tracks processed result IDs for duplicate detection
* returns stats as found for guest reads
* never persists account-shaped data

Guest stats are useful for local transient play and guest-to-local-profile seeding. They are not online-trusted and are not durable.

### Account store

The account route uses `RailsStore` when `PLAYER_DATA_RAILS_BASE_URL` is configured.

`RailsStore` behavior:

* requires `identity_kind = authenticated_account`
* requires `account_id`
* requires `PLAYER_DATA_RAILS_INTERNAL_TOKEN` for Rails internal calls
* reads stats from `POST /api/internal/player-data/stats`
* writes match results to `POST /internal/player-data/match-results`
* sends the internal token as a bearer token for internal Rails paths

When no Rails base URL is configured, the account route uses `MemoryStore`. That fallback is an in-memory runtime behavior, not Rails/Postgres persistence.

### Local store

The local route uses the configured local store.

In the standard no-tag development build, the game-server injects the embedded SQLite store. The store initializes these current tables:

```text
local_profiles
local_profile_default
local_player_stats
local_player_match_results
```

The embedded SQLite local store owns:

* local profile list
* local profile creation
* local profile deletion
* local profile display-name update
* default local profile read/write
* local profile stat reads
* local match-result writes
* duplicate local match-result detection by `result_id`

When local SQLite is not configured, the local route uses `NoopStore`. `NoopStore` implements stat load/write methods but does not implement local profile management. As a result, local profile management returns `local_profiles_unavailable` through the HTTP handler path.

## Local profile management

Local profile management is routed through the runtime and local store interface.

The runtime does not allow handlers to access the concrete SQLite store. It checks whether the configured local store implements `LocalProfileStore`.

`LocalProfileStore` owns:

```text
ListLocalProfiles
CreateLocalProfile
DeleteLocalProfile
UpdateLocalProfileDisplayName
GetDefaultLocalProfile
SetDefaultLocalProfile
```

Local profile create can seed the new local profile from guest stats. The HTTP handler asks the runtime for guest seed stats with `LocalProfileSeedStats(seed_from_guest_stats)`. The handler does not read the guest store directly.

Default local profile selection stores identity kind and local profile ID. Display name is presentation data returned from the local profile row when the selected identity is local profile. If no default row exists, or if the selected local profile no longer exists, the SQLite store returns the guest default.

Deleting a local profile removes its local stats and local match-result rows. If the deleted profile is the stored default, the SQLite store resets the default to guest.

## Data ownership

The runtime routes and mutates player-data aggregates. It does not own live gameplay state.

Current stats shape:

```text
total_score
high_score
ship_deaths
games_played
wins
```

Match-result write inputs:

```text
result_id
match_id
identity
play_mode
score
ship_deaths
won
```

Identity routing inputs:

```text
identity_kind
account_id
local_profile_id
```

Local profile management inputs:

```text
local_profile_id
display_name
identity_kind
seed_from_guest_stats
```

Data ownership by route:

| Route                 | Data owned                                                          | Persistence                            |
| --------------------- | ------------------------------------------------------------------- | -------------------------------------- |
| Guest                 | transient aggregate stats and processed result IDs                  | runtime memory                         |
| Local Profile         | local profiles, local default, local stats, local match-result rows | embedded SQLite when configured        |
| Authenticated Account | account stats and match results through Rails adapter               | Rails/API and Postgres when configured |

SQLite belongs to `services/player-data`. Postgres belongs to `services/api-server`. The game server does not know table names or mutate either database directly.

## Failure behavior

Runtime construction fails when:

* `NewRuntime` receives no store
* `NewConfiguredRuntime` receives `SQLitePath` without `LocalStoreFactory`
* Rails store construction receives an empty Rails base URL
* injected local store construction fails
* injected SQLite schema initialization fails at the game-server composition root

Packet dispatch returns encoded error result packets for:

* invalid play-mode and identity combinations
* store errors during stat load
* store errors during match-result recording

The dispatcher returns a direct error for unknown packet types or malformed packet payloads.

Local profile HTTP behavior maps local store failures to stable response errors:

* unavailable local profile store -> `local_profiles_unavailable`
* missing local profile -> `local_profile_not_found`
* invalid request -> `invalid_request`
* unsupported method -> `method_not_allowed`

Profile HTTP behavior maps unavailable runtime or store errors to `profile_unavailable`.

## Code map

Primary player-data runtime files:

* `services/player-data/playerdata/runtime.go`
* `services/player-data/playerdata/configured_runtime.go`
* `services/player-data/playerdata/default_runtime.go`
* `services/player-data/playerdata/runtime_sink.go`
* `services/player-data/playerdata/dispatcher.go`
* `services/player-data/playerdata/mode_policy.go`
* `services/player-data/playerdata/store.go`
* `services/player-data/playerdata/store_router.go`
* `services/player-data/playerdata/identity.go`

Store implementations:

* `services/player-data/playerdata/memory_store.go`
* `services/player-data/playerdata/guest_memory_store.go`
* `services/player-data/playerdata/noop_store.go`
* `services/player-data/playerdata/rails_store.go`
* `services/player-data/playerdata/embeddedsqlite/sqlite_store.go`

Runtime protocol and codec files:

* `services/player-data/codec/codec.go`
* `services/player-data/codec/envelope.go`
* `services/player-data/protocol/packets.go`
* `shared/packets/player_data.toml`
* `shared/player_data/stats.toml`
* `shared/player_data/match_result.toml`

HTTP handlers that consume runtime methods:

* `services/player-data/httpapi/profile_handler.go`
* `services/player-data/httpapi/local_profiles_handler.go`

Game-server composition and integration files:

* `services/game-server/cmd/game-server/main.go`
* `services/game-server/cmd/game-server/player_data_http.go`
* `services/game-server/cmd/game-server/player_data_local_store_dev.go`
* `services/game-server/cmd/game-server/player_data_local_store_noembeddedsqlite.go`
* `services/game-server/internal/matchreporting/runtime_reporter.go`
* `services/game-server/internal/matchreporting/mapper.go`

Important non-ownership boundaries:

* `services/game-server/internal/rooms/` owns room and match lifecycle.
* `services/game-server/internal/game/` owns live simulation.
* `services/game-server/internal/matchreporting/` owns mapping game-server match summaries into player-data commands.
* `services/api-server/app/controllers/internal/player_data/` owns Rails internal player-data persistence endpoints.
* `services/api-server/app/services/player_stats/` owns Rails account stat mutation.

## Tests

Primary player-data runtime tests:

* `services/player-data/playerdata/runtime_test.go`
* `services/player-data/playerdata/configured_runtime_test.go`
* `services/player-data/playerdata/configured_runtime_embedded_sqlite_test.go`
* `services/player-data/playerdata/default_runtime_test.go`
* `services/player-data/playerdata/dispatcher_test.go`
* `services/player-data/playerdata/mode_policy_test.go`
* `services/player-data/playerdata/store_router_test.go`
* `services/player-data/playerdata/identity_test.go`

Store tests:

* `services/player-data/playerdata/memory_store_test.go`
* `services/player-data/playerdata/guest_memory_store_test.go`
* `services/player-data/playerdata/noop_store_test.go`
* `services/player-data/playerdata/rails_store_test.go`
* `services/player-data/playerdata/embeddedsqlite/sqlite_store_test.go`

HTTP tests:

* `services/player-data/httpapi/local_profiles_handler_test.go`

Game-server integration tests related to this runtime boundary:

* `services/game-server/internal/matchreporting/runtime_reporter_test.go`
* `services/game-server/internal/matchreporting/mapper_test.go`

Current verified behavior includes:

* configured runtime defaults to account memory, local noop, and guest memory when no external stores are configured
* configured runtime uses Rails for authenticated account when Rails base URL is configured
* configured runtime keeps account, local, and guest stats separate
* configured runtime routes local profile commands through the injected local store when SQLite path and local factory are configured
* embedded SQLite local profile stats persist across runtime reconstruction in the standard no-tag build
* mode policy accepts and rejects the current play-mode and identity matrix
* store router dispatches reads and writes by identity kind
* guest memory detects duplicate result IDs
* Rails store sends internal bearer-token requests to Rails internal player-data endpoints
* local profile handler returns unavailable behavior when the local profile store is unavailable

## Related docs

* [Player Data](./!INDEX.md)
* [Game Server Player Data HTTP Hosting](../game-server/integrations/player-data-http-hosting.md)
* [Game Server Match Result Reporting](../game-server/integrations/match-result-reporting.md)
* [API Server Player Stats And Match Results](../api-server/player-stats-and-match-results.md)
* [Platform Account And Identity Current State](../../domains/platform/account-and-identity-current-state.md)
* [Data](../../data/!INDEX.md)
* [Protocol](../../protocol/!INDEX.md)

## Notes

The player-data runtime is currently hosted in-process by the game server, but its store-routing boundary is intentionally extractable. A future separate player-data server can replace the in-process transport without moving store selection into gameplay or room code.

`display_name` and callsign values are presentation identity. Store routing uses `identity_kind`, `account_id`, and `local_profile_id`.

The embedded SQLite local stats table currently stores `total_score`, `high_score`, `ship_deaths`, and `games_played`. The local SQLite store returns `wins` as zero from local profile stat reads.
