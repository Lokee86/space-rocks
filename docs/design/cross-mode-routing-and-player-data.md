# Cross-Mode Routing And Player Data

## Purpose

This document defines the server-side routing architecture for play modes, identity states, admission, and player-data destinations.

The main thesis is:

- Gameplay emits player-relevant intent.
- Admission decides whether a session may enter a mode.
- Identity describes who the session represents.
- Player-data routing decides which player-data service boundary handles durable data.

## Terminology

- Play mode means the semantic session/game route, not simply whether a process is running locally.
- Identity state means the session's account or profile posture.
- Player-data destination means where durable player-shaped data is owned and persisted.

## Play Modes

Initial play modes:

- Local Single-Player
- Online Multiplayer
- Multiplayer Simulation

Local Single-Player:

- local-only gameplay
- no authenticated account
- no Rails dependency
- Guest or Local Profile only

Online Multiplayer:

- account-authenticated route
- Authenticated Account required
- player data routes to Rails/API

Implementation status: the online-multiplayer auth/admission seam is now in place with Rails internal token verification, Go authclient, websocket session identity, and websocket auth packets. This does not yet implement Local Profile, embedded DB, player-data routing, loadout persistence, unlocks, achievements, or profile sync.

Multiplayer Simulation:

- controlled local/test environment for online-style multiplayer/account behavior
- semantically follows online multiplayer routing
- not the same as local single-player

`local` is not enough information; a locally running server can run either local single-player or multiplayer simulation.

## Server Capability Profiles

Server capability profile means the set of compiled and available backend capabilities.

Initial server capability profiles:

- Local-capable server profile
- Online-authoritative server profile

Local-capable server profile:

- may allow Guest
- may allow Local Profile
- may include embedded DB
- may also exercise Authenticated Account through multiplayer simulation

Online-authoritative server profile:

- requires Authenticated Account
- excludes embedded player-data DB path
- rejects Guest and Local Profile gameplay admission
- routes durable player data to Rails/API

This is not about debug, admin, or tester privileges.

Debug and development builds may expose multiple capabilities at once for testing, but debug privileges are overlays and are not part of this routing architecture.

## Admission Routing

Admission is the seam that decides whether a session or identity may enter a requested play mode.

Admission inputs:

- requested play mode
- identity state
- server capability profile
- room/session policy

Admission outputs:

- allow
- reject with reason
- possibly allow with degraded or no-durable behavior only where explicitly defined

| Mode | Guest | Local Profile | Authenticated Account |
| --- | --- | --- | --- |
| Local Single-Player | allowed | allowed | rejected |
| Online Multiplayer | rejected | rejected | allowed |
| Multiplayer Simulation | rejected by default | rejected by default | allowed |

Multiplayer Simulation may add explicit controlled exceptions later, but the default should mimic online multiplayer.

Admission ownership notes:

- Admission owns mode/identity validity and auth-required decisions.
- Admission does not own profile persistence, match scoring, live simulation, token verification mechanics, or embedded DB writes.

## Websocket And Auth Boundary

Current state:

- Godot connects to the Go websocket.
- The game server creates session-local player/session state.
- The game server now has websocket account verification through Rails token verification, Go authclient, session identity, and websocket auth packets.

Target lifecycle:

- websocket upgrade
- create session
- optional auth handshake
- mode request
- admission decision
- room/game flow

Local-capable lifecycle:

- session may default to Guest
- client may start local single-player as Guest
- client may select Local Profile
- Local Profile flow depends on `services/player-data-server`.
- client may authenticate account for multiplayer simulation or online route

Local single-player with Guest:

- client connects to game-server
- no durable player-data service route is required
- game-server runs match normally

Local single-player with Local Profile:

- client uses `services/player-data-server` to create, list, and select Local Profile
- client starts local game-server session with selected `LocalProfileID` or local profile session reference
- game-server asks `services/player-data-server` for profile and loadout data needed for the match
- game-server reports trusted match result back to `services/player-data-server`
- `services/player-data-server` persists through SQLite

The game-server websocket should not become a general local profile management API.
Profile management UI should go through `services/player-data-server`, not game-server.

Online-authoritative lifecycle:

- token required during upgrade or first auth message
- game-server verifies token through Rails/API
- session gets Authenticated Account identity
- only online multiplayer mode is admitted

Do not log bearer tokens.
Do not persist bearer tokens in game state.
Do not use bearer tokens as gameplay identity.
The game-server must not read Rails auth tables directly.

## Identity States

Initial identity states:

- Guest
- Local Profile
- Authenticated Account

Guest:

- local-only
- temporary/session identity
- no durable account-shaped profile
- no online persistence

Local Profile:

- local-only
- durable
- account-shaped
- stored through the local player-data service backed by SQLite
- mirrors online account/profile concepts such as profile, loadout, unlocks, progression, stats, and settings if relevant

Authenticated Account:

- online/API-backed
- durable
- account-shaped
- Rails-owned identity and persistence
- required for online multiplayer
- rejected by local single-player

Implementation status: this routing model exists as the auth/admission boundary, but it is not the same as Local Profile or player-data implementation.

## Non-Goals

- Do not make gameplay simulation depend on Rails.
- Do not make local single-player require OAuth/auth.
- Do not allow authenticated accounts in local single-player.
- Do not allow online multiplayer with Guest or Local Profile identities.
- Do not make the Go game-server read Rails tables directly.
- Do not let gameplay code choose embedded DB versus Rails/API directly.
- Do not make game-server own Local Profile persistence.
- Do not make game-server directly write SQLite player-data tables.
- Do not make game-server directly write Postgres account/player-data tables.
- Do not make game-server websocket handle general local profile CRUD.
- Do not let client bypass the player-data service to write Local Profile data.
- Do not treat debug/dev/admin/tester privileges as identity states.
- Do not persist live simulation state as account/profile data.

## Core Invariants

- Local Single-Player allows Guest and Local Profile.
- Local Single-Player rejects Authenticated Account.
- Online Multiplayer requires Authenticated Account.
- Online Multiplayer rejects Guest and Local Profile.
- Local Profile is durable, account-shaped, local-only, and stored through the local player-data service backed by SQLite.
- Authenticated Account is durable, account-shaped, online/API-backed.
- Guest is temporary/local-only and has no durable account-shaped data.

## Identifier Separation

- SessionID is websocket/session scoped.
- GamePlayerID is temporary match/simulation scoped.
- LocalProfileID is durable embedded profile identity.
- account_id is the canonical cross-system UUID identity for authenticated accounts.
- Rails `user_id` remains an internal foreign key to `users.id`.
- Never replace GamePlayerID with account_id or LocalProfileID.

## Data Destinations

Initial data destinations:

- No durable route
- Local player-data service backed by SQLite
- Online player-data service backed by Postgres

No durable route means the data is session-only or scratch-only and is not persisted as account-shaped data.
Data destination means the service route that owns the data operation, not just a database.
Backing store details are hidden behind the owning service.
SQLite belongs to `services/player-data-server`.
Postgres belongs to `services/api-server`.
The game-server should not directly write either account-shaped player-data database.

## Local Profile, Player-Data Server, And SQLite

The local player-data service exists to support Local Profile.

Local Profile is durable, account-shaped, and local-only.

SQLite is owned by `services/player-data-server`, not `services/game-server`.

The embedded or local database is not a gameplay concern.

The game-server must not know SQLite table names, DB file paths, or local profile migration details.

The local player-data service owns local versions of:

- profiles
- loadouts
- progression/unlocks
- local stats
- local match summaries if desired
- schema versioning later

The local player-data service must not store:

- Discord access tokens
- Rails bearer tokens as local profile identity
- online account secrets
- online leaderboard eligibility

The local player-data service is not a cache for Rails/API.

Local Profile data is local-authoritative and should not be treated as online-trusted by default.

## Authenticated Account And Rails/API

Rails/API is the durable backend for Authenticated Account.
`services/api-server` is the online player-data service for Authenticated Account.

Rails/API owns:

- authenticated users
- OAuth identities
- online profiles
- online loadouts/progression
- online player-data writes for profile
- online player-data writes for loadout
- online player-data writes for progression
- online player-data writes for unlocks
- online player-data writes for match history
- online player-data writes for leaderboards
- leaderboards
- account-owned match history
- future moderation/admin account data

The Go game-server interacts with Rails/API only through explicit API clients or endpoints.

Conceptual future token verification endpoint example:

```http
POST /internal/auth/verify-token
```

Token verification returns an Authenticated Account identity context to the game server.

The game server stores identity context, not the Rails token as gameplay identity.
The client uses `services/api-server` for authenticated account profile, loadout, and progression UI.

## Player-Data Service Boundary

Player-data is not owned by the game-server.

The game-server consumes player-data service APIs.

The client consumes player-data service APIs for local profile and loadout UI.

The backing store is hidden behind the service.

Planned service split:

- `services/game-server` owns simulation, rooms, match lifecycle, gameplay events, and trusted match results.
- `services/player-data-server` owns Local Profile persistence backed by SQLite.
- `services/api-server` owns Authenticated Account persistence backed by Postgres.

Symmetry:

- Local Profile path: `client`/`game-server` -> `services/player-data-server` -> SQLite.
- Authenticated Account path: `client`/`game-server` -> `services/api-server` -> Postgres.

This split is planned from the first Local Profile implementation, not as a later extraction.

### Live Grant Transport

Live progression grants may use internal HTTP from the game-server to the owning player-data service as the first viable path.

For Authenticated Account, the target service is `services/api-server`.

For future Local Profile, the target service is `services/player-data-server`.

A server-to-server websocket is not required for the first version.

Durable queues and outbox workers are future hardening options, not the starting point.

Live grant writes must be idempotent using a `grant_id` or `event_id`.

Retries must not double-credit rewards.

### Progression Ownership

Player-data services own progression persistence, not live gameplay authority.

The game-server owns gameplay facts, match results, and progression-producing events.

Not every player-data write should wait until match end.

Summary-style stats can be finalized at match resolution, while valuable durable rewards should be persisted live or near-live so they are not coupled to the end-of-match summary path.

Examples of match-summary stats:

- total score
- high score
- ship deaths
- games played
- wins

Examples of live durable grants:

- currency
- ship parts
- rare drops
- unlock tokens
- account-affecting rewards

Gameplay emits authoritative domain events, but player-data services own persistence of the durable result.

The game-server should not update Rails/SQLite tables directly.

Game-server-owned facts include:

- `MatchCompleted`
- `AsteroidDestroyed`
- `BossDefeated`
- `ScoreEarned`
- `SurvivalTimeReached`
- `PickupCollected`
- `DamageDealt`

Player-data-owned persisted data includes:

- progression state
- unlock records
- persistent stats
- achievement/progress markers
- match summary history
- loadout availability

The player-data service should not decide combat or gameplay rules such as whether an asteroid kill grants immediate score.

Account and local-profile progression policies may be applied by the player-data service when processing trusted match results, but gameplay rules remain in gameplay and rules systems.

### V1 Persistent Stats

The initial persistent stats payload is summary-only and can be committed at match resolution.

V1 stat fields:

- `total_score`
- `high_score`
- `ship_deaths`
- `games_played`
- `wins`

For V1 multiplayer, the winner is the authenticated player with the highest match score.

This V1 stats payload does not include currency, ship parts, unlocks, loadouts, achievements, or match history yet.

### Stats Event Pipeline

For V1 stats, the flow is:

- game-server emits domain events during gameplay
- a match or session summary accumulates per-player facts
- match resolution decides the final score and V1 winner
- the game-server reports the summary to the player-data service later
- the player-data service persists stats

Likely event inputs include:

- `ScoreEarned`
- `ShipDeath`
- `MatchCompleted`
- `PlayerJoined`
- `PlayerFinished`

Gameplay code should not directly mutate persistent player stats.

For V1, match summary reporting is the commit point for stats.

Live durable rewards use a separate progression-grant style path instead of the stats summary path.

## Shared Player-Data Schema SSoT

Local Profile and Authenticated Account share the same account-shaped player-data concepts.

The logical schema for those concepts should come from `shared/player_data` via the future data-sync player-data domain.

Physical storage may differ between embedded DB and Rails/Postgres.

Gameplay-facing code should depend on playerdata contracts, not storage-specific tables.

Live simulation state is excluded.

See [player-data schema source of truth](player-data-schema-ssot.md).

## Player-Data Service Contract Expectations

Conceptual player-data service contract example:

```text
LoadProfile
SaveProfile
LoadLoadout
SaveLoadout
LoadProgression
SaveProgression
RecordMatchResult
```

Service-route wording:

- Guest has no durable account-shaped service route.
- Local Profile uses `services/player-data-server`.
- Authenticated Account uses `services/api-server`.

Services may implement the contract differently, but they must satisfy the shared logical schema.

The game-server and client should consume service APIs, not storage implementations.

Included data categories:

- profile
- loadout
- progression/unlocks
- match result summary
- server-owned settings if they affect authoritative gameplay or cross-session consistency

Excluded data categories:

- live asteroid state
- live bullet state
- per-tick movement
- collision internals
- temporary pickups
- room simulation state
- frame telemetry
- pure client presentation preferences unless intentionally server-owned

## Player-Data Routing

- Guest routes to no durable account-shaped data.
- Local Profile routes to `services/player-data-server`, backed by SQLite.
- Authenticated Account routes to `services/api-server`, backed by Postgres.

| Identity | Durable account-shaped data | Service route | Backing store |
| --- | --- | --- | --- |
| Guest | no | none/no durable route | none |
| Local Profile | yes | `services/player-data-server` | SQLite |
| Authenticated Account | yes | `services/api-server` | Postgres |

Clarifications:

- Local Profile is not "guest with saves."
- Authenticated Account is not "local profile synced online."
- Local Profile and Authenticated Account share logical player-data contracts, but use different service implementations.

## Impossible States

These states should not be admitted by the routing architecture:

- Local Single-Player + Authenticated Account
- Online Multiplayer + Guest
- Online Multiplayer + Local Profile
- Online-authoritative server + player-data service owned by the wrong backing store
- Online-authoritative server + local profile gameplay admission
- Online-authoritative server + guest gameplay admission
- game-server directly writing Local Profile SQLite tables
- game-server directly writing Authenticated Account Postgres tables
- client directly mutating SQLite outside player-data-server
- player-data-server owning live gameplay simulation
- api-server owning local SQLite Local Profile persistence
- gameplay code directly selecting Rails/API versus embedded DB
- Go game-server directly reading Rails auth tables

## Failure And Error Model

Possible future error codes:

- `auth_required`
- `invalid_token`
- `token_verification_unavailable`
- `identity_not_allowed_for_mode`
- `local_profile_required`
- `local_profile_not_found`
- `local_profile_unavailable`
- `online_account_required`
- `embedded_data_unavailable`
- `data_route_unavailable`

Failure behavior notes:

- API unavailable during online admission should reject or make the server unavailable.
- Embedded DB unavailable should make Local Profile unavailable; Guest may still work if allowed.
- `NoDurableStore` receiving durable writes must be explicit no-op or explicit reject, not accidental fake success.

## Observability And Security

Log the mode, identity, admission decision, and data route.

Never log bearer tokens, Discord access tokens, OAuth codes, raw OAuth state, or client secrets.

## Phased Rollout

1. Documentation and invariants.
2. Pure server vocabulary: play mode, identity, capability profile.
3. Admission package and routing matrix tests.
4. Room mode field.
5. Session identity field.
6. Admission wiring, initially behavior-preserving where needed.
7. Create `services/player-data-server` scaffold.
8. Define the player-data service contract from the shared logical schema.
9. Local Profile CRUD in `services/player-data-server`.
10. SQLite-backed persistence and migrations in `services/player-data-server`.
11. Game-server consuming `services/player-data-server` APIs for loadout and profile reads plus match-result writes.
12. Client consuming `services/player-data-server` APIs for local profile UI.
13. Rails token verification endpoint and Go auth client.
14. Websocket auth handshake.
15. Enforce online multiplayer admission.
16. Authenticated Account path in `services/api-server` as the online equivalent.
17. Store contract tests.

The first code milestone after docs should still be pure vocabulary and admission unless the team deliberately starts the new service first.

## Deferred Work

- local profile creation/list/select/rename/delete
- local profile schema migration/versioning
- account linking or local-to-online migration
- online leaderboards
- anti-cheat/trust policy
- client token storage
- store contract tests
