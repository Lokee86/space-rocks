# Account And Identity Current State
Parent index: [Platform](./!INDEX.md)

## Purpose

This doc records the current account, identity, authentication, admission, and player-data routing behavior for Space Rocks.

It is a current-state domain reference, not a future planning document.

The purpose is to keep identity, authentication, session, room membership, gameplay player identity, and player-data routing boundaries clear across the client, Rails API, Go game-server, rooms, gameplay, and player-data runtime.

## Overview

Space Rocks currently uses three player identity states:

Guest
Local Profile
Authenticated Account

These identity states are intentionally separate.

Guest:
- temporary local identity
- transient player-data route
- local single-player only

Local Profile:
- durable local identity
- local player-data route
- local single-player only

Authenticated Account:
- durable online identity
- Rails/API-backed player-data route
- online multiplayer

Identity state is not the same as:
- WebSocket session
- Room membership
- Gameplay player
- Ship entity
- OAuth provider account
- Display name
- Callsign
- Player-data store

Current production multiplayer create/join requires Authenticated Account identity. If auth verification is unavailable, multiplayer create/join returns `auth_unavailable`.

## Participating systems

- Godot client - owns selected local player context, stored auth token state, menu routing, and WebSocket auth request sending.
- Rails API server - owns Authenticated Account identity, OAuth login, bearer token issuance, `/api/auth/me`, logout, and internal token verification.
- Go game-server - owns WebSocket session identity, auth verification handoff, production multiplayer admission, room request routing, and trusted match-result reporting.
- Rooms - owns room membership, room lifecycle, room member identity attachment, and active game instance ownership.
- Gameplay - owns live game-player IDs, player session state, scoring, lives, deaths, respawn, and match facts.
- services/player-data - owns identity-based player-data routing and normalized profile/stat read/write behavior.
- Embedded SQLite route - backs Local Profile data in the standard no-tag development build.
- Rails/Postgres route - backs Authenticated Account player-data through the API route.
- Guest memory route - backs transient Guest stats.

## Authority boundaries

### Client authority

- The client may select Guest or Local Profile context for local single-player.
- The client may store and present Authenticated Account session state after login.
- The client may send `authenticate_request` after WebSocket connect when a bearer token exists.
- The client does not choose SQLite, Rails/Postgres, or guest memory directly.
- The client does not own durable account identity.
- The client does not mutate persistent stats directly.

### Rails/API authority

- Rails owns Authenticated Account identity.
- Rails owns OAuth login and provider identity records.
- Rails issues and verifies bearer tokens.
- Rails exposes `/api/auth/me` for authenticated client session validation.
- Rails exposes `/internal/auth/verify-token` for game-server token verification.
- Rails owns online account-backed persistence through API-owned routes.

### Go game-server authority

- The game-server owns WebSocket session identity.
- WebSocket sessions start as Guest identity.
- `authenticate_request` may upgrade a WebSocket session to Authenticated Account identity after Rails verification succeeds.
- Multiplayer create/join admission requires Authenticated Account identity in production behavior.
- If the auth verifier is unavailable, multiplayer create/join returns `auth_unavailable`.
- The game-server must not read Rails auth tables directly.
- The game-server must not write SQLite or Rails/Postgres player-data tables directly.

### Room authority

- Rooms own room membership.
- Rooms attach account or local-profile identity references to members when relevant.
- Room membership identity is not the same as Authenticated Account identity, Local Profile identity, or gameplay player identity.

### Gameplay authority

- Gameplay owns active match players, ship state, scoring, lives, deaths, respawn, and match facts.
- Gameplay player IDs must not be replaced by account IDs or local profile IDs.
- Live gameplay state is not account/profile state.

### Player-data authority

- services/player-data owns identity-based player-data route selection.
- Guest routes to transient memory.
- Local Profile routes to the local profile store.
- Authenticated Account routes through the Rails/API-backed store.
- Backing store details stay behind player-data routes.

### Identifier rules

- SessionID is WebSocket/session scoped.
- MemberID is room-membership scoped and is the reserved reconnect seam.
- PlayerID and GamePlayerID are gameplay scoped.
- LocalProfileID is durable local-profile identity.
- account_id is canonical authenticated-account identity.
- Rails user_id is a Rails internal database foreign key.
- OAuth provider user ID is an external provider login/linking fact.
- display_name and callsign are presentation identity.
- Bearer tokens prove identity but are not gameplay identity.
- Do not use bearer tokens, OAuth provider IDs, display names, callsigns, account IDs, or local profile IDs as gameplay player IDs.

## Flow summary

### Guest local single-player

- Single Player menu route.
- Guest selected or defaulted.
- `start_single_player_request` is sent without `local_profile_id`.
- The game-server creates a non-joinable single-player room.
- Gameplay starts.
- Match result routes as Guest.
- services/player-data uses transient guest memory.

### Local Profile single-player

- Single Player menu route.
- Local Profile selected.
- `local_profile_id` is included in `start_single_player_request`.
- The game-server creates a non-joinable single-player room.
- Room member receives `local_profile_id`.
- Gameplay starts.
- Match result includes `local_profile_id`.
- services/player-data routes to Local Profile store.

### Authenticated multiplayer

- Player signs in through Rails auth flow.
- Client stores Space Rocks bearer token.
- WebSocket connects.
- Client sends `authenticate_request`.
- Game-server verifies token through Rails.
- WebSocket session becomes Authenticated Account identity.
- Multiplayer create/join request is sent.
- Admission succeeds.
- Room member receives `account_id`.
- Gameplay starts after room lifecycle.
- Match result includes `account_id`.
- services/player-data routes through Rails/API-backed store.

### Discord login/session flow

- Client begins Discord login session.
- Rails creates OAuth login session and login URL.
- Client opens browser login URL.
- Client polls login-session exchange route.
- Rails returns bearer token and user payload.
- Client stores token and signed-in state.
- Client validates saved token later through `/api/auth/me`.
- Logout clears local token and signed-in state.

### WebSocket auth flow

- WebSocket connects.
- Session starts as Guest identity.
- Client optionally sends `authenticate_request` with bearer token.
- Game-server authclient calls Rails `/internal/auth/verify-token`.
- Rails returns valid account identity or invalid result.
- Game-server stores Authenticated Account identity if valid.
- `authenticate_result` is sent to client.

## Inputs and outputs

Current inputs:
- selected single-player identity kind
- selected `local_profile_id`
- stored bearer token
- requested play mode
- WebSocket `authenticate_request`
- auth verifier availability
- Rails token verification result
- room create/join/start request
- match result facts

Current outputs:
- client signed-in state
- WebSocket session identity
- room admission result
- room member `account_id` when authenticated
- room member `local_profile_id` for selected local profile single-player
- gameplay player identity
- player-data identity
- player-data route
- profile/stat read result
- match-result write result

Current admission matrix:

| Mode | Guest | Local Profile | Authenticated Account |
| --- | --- | --- | --- |
| Local Single-Player | allowed | allowed | intended rejected |
| Online Multiplayer | rejected | rejected | allowed |
| Multiplayer Simulation | rejected by default | rejected by default | allowed |

Current production behavior:
- Online multiplayer create/join requires Authenticated Account.
- Missing auth verifier returns `auth_unavailable`.
- Unauthenticated multiplayer create/join returns `auth_required`.
- Single-player client flow uses Guest or Local Profile.
- Player-data mode validation rejects `single_player + authenticated_account`.

Known implementation limitation:
- The WebSocket start-single-player path does not currently reject an already-authenticated WebSocket session directly.

Current player-data route table:

| Identity | Durable account-shaped data | Player-data route | Backing store |
| --- | --- | --- | --- |
| Guest | no | guest transient route | in-memory unsaved stats |
| Local Profile | yes | local profile route | embedded SQLite in standard no-tag dev build |
| Authenticated Account | yes | authenticated account route | Rails/API and Postgres |

## Out of scope

This doc does not define or plan:
- future OAuth provider expansion
- manual account creation
- account recovery
- account linking
- provider unlinking
- local-to-online migration
- guest-to-online migration
- account merge
- leaderboard eligibility
- anti-cheat trust policy
- social graph policy
- matchmaking policy
- reconnect behavior
- physical Rails/Postgres schema
- physical SQLite schema
- exact UI layout

Future and unresolved policy belongs in planning docs.

## Active issues

- `start_single_player_request` does not currently reject an already-authenticated WebSocket session at the server boundary. The intended identity model is still Guest or Local Profile for local single-player, and player-data mode validation rejects `single_player + authenticated_account`, but the WebSocket start-single-player path does not enforce that rejection directly yet. See [Current System Limits](../../limits/current-system-limits.md#architecture--networking).

## Related docs

- [Platform domain index](./!README.md)
- [Domain index](../!README.md)
- [Client service index](../../services/client/!README.md)
- [Game Server service index](../../services/game-server/!README.md)
- [API Server service index](../../services/api-server/!README.md)
- [Player Data service index](../../services/player-data/!README.md)
- [Account And Identity Systems planning](../../planning/domains/platform/account-and-identity-systems.md)
- [Multiplayer Session And Lifecycle planning](../../planning/domains/platform/stubs/multiplayer-session-and-lifecycle.md)
- [Player Data And Persistence planning](../../planning/domains/platform/stubs/player-data-and-persistence.md)
- [Current System Limits](../../limits/current-system-limits.md)

## Notes

- This doc records current production behavior. A future devtools or development override for no-auth multiplayer would be a separate devtools/planning concern.
- This doc is the current domain reference for account and identity flow.
- Local Profile is not a Rails/API cache.
- Authenticated Account is not a synced Local Profile.
- Display names and callsigns are presentation identity, not durable routing identity.
- Bearer tokens prove identity but are not gameplay identity.
