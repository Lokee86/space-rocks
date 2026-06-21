# Player Data Routing Flow

Parent index: [Platform](./!INDEX.md)

## Purpose

This document describes the cross-system player-data routing flow for Space Rocks.

It explains how Guest, Local Profile, and Authenticated Account data requests move across the client, game-server host process, player-data runtime, API server, and backing stores without giving any caller direct authority over storage selection.

## Overview

Player-data routing is the platform flow that decides where account-shaped player data is read from or written to after identity and play mode are known.

The current routing model has three identity paths:

| Identity              | Durable account-shaped data | Route                       | Backing behavior                                                                |
| --------------------- | --------------------------- | --------------------------- | ------------------------------------------------------------------------------- |
| Guest                 | no                          | guest route                 | process-local transient memory                                                  |
| Local Profile         | yes                         | local profile route         | embedded SQLite in standard no-tag development builds                           |
| Authenticated Account | yes                         | authenticated account route | Rails/API-backed Postgres through the player-data Rails adapter when configured |

The client chooses a player context such as Guest, Local Profile, or Authenticated Account. It does not choose guest memory, SQLite, Rails, or Postgres.

The game server currently hosts the player-data runtime and HTTP handlers in-process. That hosting detail does not make the game server the owner of player-data storage. The game server supplies process dependencies, receives client requests, reports authoritative match results, and forwards player-data work through the player-data runtime boundary.

The player-data runtime owns route selection. It validates identity and play-mode combinations where the packet path carries play-mode context, selects the account, local, or guest route by `identity_kind`, and returns normalized stats to callers.

The API server owns authenticated-account persistence. Embedded SQLite owns local-profile persistence. Guest memory owns transient Guest stats for the runtime instance.

## Participating systems

* Client - owns selected player context, local pilot UI state, bearer-token possession, profile readout presentation, and local pilot API consumption.
* Game Server - owns the HTTP host process, WebSocket session identity, multiplayer admission, room membership identity attachment, match-result reporting, and player-data runtime composition.
* Player Data - owns runtime route selection, identity-based store routing, local profile management behavior, guest stat memory, local store delegation, Rails adapter calls, normalized stats output, and duplicate match-result handling.
* API Server - owns Authenticated Account identity, token verification, OAuth-backed login, authenticated-account stats reads, authenticated-account match-result persistence, and Rails/Postgres physical storage.
* Rooms - owns room membership, local profile or account ID attachment to members, and the match summary source used by match-result reporting.
* Gameplay - owns live simulation facts such as score, deaths, winner resolution inputs, and match completion state.
* Data Pipeline - owns logical player-data schema sources, packet schema generation, and HTTP contract sources that participating services implement.

## Authority boundaries

### Client authority

The client may:

* select Guest for local play
* select a Local Profile for local single-player
* store and send an Authenticated Account bearer token after login
* request profile readout for the active context
* call local profile list, create, update, delete, and default-selection routes
* present returned profile and stats data

The client must not:

* choose SQLite, Rails/Postgres, or guest memory directly
* mutate persistent stats directly
* count gameplay facts as authoritative stats
* call Rails internal player-data routes directly
* treat display name or callsign as routing identity

### Game-server authority

The game server may:

* host player-data HTTP routes on the current development server process
* adapt the game-server auth verifier for profile read requests
* attach `account_id` or `local_profile_id` to room members
* build authoritative match summaries after game-over
* report match results into the player-data runtime
* fail startup when the player-data runtime or reporter cannot initialize

The game server must not:

* read or write SQLite tables directly
* read or write Rails/Postgres player-data tables directly
* choose the backing store for a player-data request outside the player-data runtime
* replace gameplay player IDs with account IDs or local profile IDs
* make profile persistence a room, networking, or simulation responsibility

### Player-data authority

Player Data owns:

* route selection by `identity_kind`
* mode and identity validation for packet-dispatched stats reads and match-result writes
* local profile list/create/update/delete/default behavior
* local profile unavailability behavior when no local store is configured
* Guest transient stats
* Local Profile local-store delegation
* Authenticated Account Rails adapter calls
* normalized stat payloads
* duplicate match-result behavior at the selected route

Player Data does not own:

* OAuth login
* bearer-token issuance
* multiplayer admission
* room lifecycle
* live gameplay state
* authoritative match outcome calculation
* Rails/Postgres schema ownership
* client presentation

### API-server authority

The API server owns:

* Authenticated Account identity
* OAuth provider login and account records
* bearer-token issuance and verification
* internal token-protected authenticated-account player-data endpoints
* Rails/Postgres stats and match-result persistence
* public current-user stats reads

The API server does not own Guest or Local Profile persistence.

### Backing-store authority

Guest memory owns transient stats for Guest identity inside the runtime instance.

The local-profile store owns Local Profile records, default selection, local stats, and local match-result rows when embedded SQLite is configured.

Rails/Postgres owns Authenticated Account aggregate stats and match-result rows.

Backing-store details stay behind the player-data route. Other systems interact through HTTP handlers, runtime methods, generated packet commands, or API endpoints.

## Flow summary

### Guest profile read

```text
Client selected Guest
-> POST /api/player-data/profile
-> game-server hosted player-data profile handler
-> player-data runtime LoadStats(identity_kind=guest)
-> guest memory route
-> normalized profile response
-> client profile readout
```

Guest profile reads return Guest presentation identity and transient stats. Missing stats become zero stats in the profile response.

### Local Profile profile read

```text
Client selected Local Profile
-> POST /api/player-data/profile with local_profile_id
-> game-server hosted player-data profile handler
-> player-data runtime LoadStats(identity_kind=local_profile)
-> local profile route
-> embedded SQLite store when configured
-> normalized profile response
-> client profile readout
```

Local Profile reads require a non-empty `local_profile_id`. The returned profile uses local activity status and the normalized stats shape. The route does not make the client or game server aware of SQLite table layout.

### Authenticated Account profile read

```text
Client has bearer token
-> POST /api/player-data/profile with identity_kind=authenticated_account
-> game-server hosted player-data profile handler
-> game-server auth verifier adapter validates token
-> account_id is resolved
-> player-data runtime LoadStats(identity_kind=authenticated_account, account_id)
-> authenticated account route
-> Rails/API-backed stats read when configured
-> normalized profile response
-> client profile readout
```

Authenticated Account profile reads require a valid bearer token and an auth verifier. The route uses `account_id` as the cross-system account identity. Rails `user_id` remains internal to the API server.

### Local profile management

```text
Client local pilot UI
-> hosted local profiles HTTP route
-> player-data local profiles handler
-> player-data runtime local profile method
-> local profile store interface
-> embedded SQLite store when configured
-> JSON response
-> client local pilot UI
```

This flow covers local profile list, create, update display name, delete, get default, and set default.

Local profile creation may seed the new profile from Guest stats when requested. The handler asks the player-data runtime for seed stats; it does not read the guest store directly.

Deleting a local profile also removes its local stats and local match-result rows through the local store. If the deleted profile was the stored default, the default resets to Guest.

### Match-result write

```text
Gameplay reaches game over
-> room builds authoritative match summary
-> game-server match reporter maps each player summary to a player-data command
-> player_data_record_match_result packet
-> player-data runtime sink
-> player-data dispatcher validates play mode and identity
-> store router selects guest, local, or account route
-> selected backing route records or rejects the result
-> normalized stats result packet
```

Match-result writes use game-server-resolved facts. Player-data does not calculate score, ship deaths, winner state, match-over eligibility, or room results.

The match-result command carries:

* `result_id`
* `match_id`
* `identity`
* `play_mode`
* `score`
* `ship_deaths`
* `won`

`result_id` is the idempotency key. Duplicate behavior is owned by the selected backing route.

### Authenticated multiplayer admission and routing

```text
Client signs in through API server
-> client stores bearer token
-> WebSocket connects to game server
-> client sends authenticate_request
-> game server verifies token through API server
-> session becomes Authenticated Account
-> multiplayer create/join is allowed
-> room member receives account_id
-> match result later routes to authenticated account
```

Production multiplayer create and join require Authenticated Account identity. If auth verification is unavailable, the game server returns `auth_unavailable`. If the session is unauthenticated, it returns `auth_required`.

### Local single-player routing

```text
Client starts local single-player as Guest or Local Profile
-> start_single_player_request optionally carries local_profile_id
-> game server creates a non-joinable single-player room
-> room member receives local_profile_id when provided
-> gameplay runs
-> match result later routes to Guest or Local Profile
```

Guest local single-player writes to transient memory. Local Profile single-player writes to the local profile route when local storage is configured.

## Inputs and outputs

Current routing inputs:

* selected identity kind
* selected `local_profile_id`
* bearer token for Authenticated Account profile reads and WebSocket authentication
* resolved `account_id` after token verification
* requested play mode
* room membership identity attachment
* match ID
* game player ID
* score
* ship deaths
* winner flag
* local profile display name for local profile management
* guest-stat seed request for local profile creation
* runtime configuration for Rails base URL, internal token, SQLite path, and local store factory

Current routing outputs:

* normalized profile response
* normalized stats payload
* local profile summaries
* default local profile response
* local profile creation, update, and deletion results
* player-data route selection
* match-result accepted/rejected status
* duplicate match-result status
* room admission errors for multiplayer auth failures
* storage unavailability errors for missing local profile support
* profile unavailability errors for failed profile reads

Current play-mode and identity matrix:

| Play mode                | Guest    | Local Profile | Authenticated Account                   |
| ------------------------ | -------- | ------------- | --------------------------------------- |
| `single_player`          | allowed  | allowed       | rejected by player-data mode validation |
| `multiplayer`            | rejected | rejected      | allowed                                 |
| `multiplayer_simulation` | rejected | rejected      | allowed                                 |

Current player-data storage matrix:

| Identity kind           | Required routing identity | Runtime route | Backing behavior                                                                                |
| ----------------------- | ------------------------- | ------------- | ----------------------------------------------------------------------------------------------- |
| `guest`                 | none                      | guest store   | transient runtime memory                                                                        |
| `local_profile`         | `local_profile_id`        | local store   | embedded SQLite in standard no-tag development builds                                           |
| `authenticated_account` | `account_id`              | account store | Rails/API-backed store when configured; in-memory account fallback when Rails is not configured |

## Out of scope

This document does not define:

* exact HTTP request/response schemas
* generated packet field ownership
* Rails/Postgres table layout
* embedded SQLite table layout
* client UI layout
* OAuth provider login flow details
* future local-to-online migration
* guest-to-account migration
* leaderboard eligibility
* progression, unlocks, inventory, currency, or loadout persistence
* game integrity policy
* public match history
* reconnect behavior
* deployment topology for a future separate player-data server

Those details belong in protocol, data, service, planning, or limits documentation.

## Active issues

* `start_single_player_request` does not currently reject an already-authenticated WebSocket session at the game-server boundary. Player-data mode validation rejects `single_player + authenticated_account`, but the WebSocket start-single-player path does not enforce that identity rejection directly yet. See [Current System Limits](../../limits/current-system-limits.md#architecture--networking).

## Related docs

* [Platform](./!INDEX.md)
* [Domains](../!INDEX.md)
* [Player Data](../../services/player-data/!INDEX.md)
* [Game Server](../../services/game-server/!INDEX.md)
* [Client](../../services/client/!INDEX.md)
* [API Server](../../services/api-server/!INDEX.md)
* [Protocol](../../protocol/!INDEX.md)
* [Data](../../data/!INDEX.md)
* [Account And Identity Current State](account-and-identity-current-state.md)
* [Player Data Runtime And Store Routing](../../services/player-data/runtime-and-store-routing.md)
* [Profile Stats Flow](../../services/player-data/profile-stats-flow.md)
* [Match Result Sinks](../../services/player-data/match-result-sinks.md)
* [Local Profiles HTTP API](../../services/player-data/local-profiles-http-api.md)
* [Game Server Player Data HTTP Hosting](../../services/game-server/integrations/player-data-http-hosting.md)
* [Game Server Match Result Reporting](../../services/game-server/integrations/match-result-reporting.md)
* [API Server Player Stats And Match Results](../../services/api-server/player-stats-and-match-results.md)
* [Player Data Schema](../../data/player-data-schema.md)
* [HTTP Contract Enforcement](../../protocol/http-contract-enforcement.md)
* [Current System Limits](../../limits/current-system-limits.md)

## Notes

* The current player-data runtime is hosted in-process by the game-server executable, but the routing boundary remains owned by the player-data service.
* `identity_kind`, `account_id`, and `local_profile_id` are routing identity. Display names and callsigns are presentation identity.
* Guest stats are transient and local to the runtime instance.
* Local Profile is local-authoritative account-shaped data, not Guest with saves and not a Rails cache.
* Authenticated Account is Rails/API-backed account-shaped data, not a synced Local Profile.
* The game server reports trusted match facts into player-data; player-data routes and persists those facts but does not recompute them.
