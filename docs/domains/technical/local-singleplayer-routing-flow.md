# Local Single-Player Flow

Parent index: [Technical](./!README.md)

## Purpose

This document describes the current technical domain flow for local single-player startup, routing, room creation, gameplay entry, result handoff, and teardown.

It focuses on the cross-system path that turns a local single-player menu request into a server-backed gameplay session. It does not document client menu implementation details, room package internals, packet field ownership, gameplay simulation rules, or player-data storage internals.

## Overview

Local single-player is a server-backed local session flow.

It is not an offline client simulation. The Godot client still connects to the Go game server over the realtime WebSocket route, sends a generated `start_single_player_request`, receives room snapshots and gameplay state from the server, and renders server-authoritative gameplay packets.

The current technical distinction is:

```text
local single-player
= single-player session mode
= Guest or Local Profile identity
= no authenticated-account admission requirement
= non-joinable server room
= immediate game start
= normal server-authoritative gameplay
```

The current local development WebSocket URLs for single-player and multiplayer may point to the same `/ws` route. That route does not define session mode. Session mode is expressed through the client boot request, the generated packet type, room creation policy, and player-data identity routing.

Local single-player currently allows:

```text
Guest
Local Profile
```

Local single-player is intended to reject:

```text
Authenticated Account
```

The active implementation keeps single-player boot independent of WebSocket auth. If the client has a stored bearer token, the client connection service may still send an auth packet after connect, but the single-player boot request does not wait for auth success and does not require Rails/API availability.

## Participating systems

* [Client](../../services/client/!README.md) - owns menu intent, active single-player context, session boot, WebSocket target selection, outbound start packet sending, room-state caching, gameplay packet acceptance, presentation, replay, and navigation after match end.
* [Game Server](../../services/game-server/!README.md) - owns WebSocket session handling, room creation, single-player room start, active game-player activation, authoritative gameplay simulation, room snapshots, match-over state, and match-result reporting.
* [Player Data](../../services/player-data/!README.md) - owns Guest and Local Profile stat routing after match results are reported and owns Local Profile persistence behind the local profile routes.
* [API Server](../../services/api-server/!README.md) - owns Authenticated Account auth and Rails-backed persistence, but local single-player startup does not depend on API-server auth admission.
* [Protocol](../../protocol/!README.md) - owns the generated realtime packet contract used by the client and game server.
* [Data](../../data/!README.md) - owns packet and constants source material that generate client and server packet helpers.

## Authority boundaries

### Client authority

The client owns local intent and presentation.

The client may:

* route Main Menu Single Player into the single-player pregame surface
* apply the saved Guest or Local Profile default
* read the active single-player context before boot
* include `local_profile_id` in the single-player start request when Local Profile is active
* send `start_single_player_request`
* open the single-player WebSocket target
* begin accepting gameplay packets after the server-observed room state reaches `InGame`
* render gameplay, HUD, match results, and route button intent after match end

The client must not:

* create authoritative gameplay state locally
* treat a scene transition as gameplay admission
* decide room state
* decide match-over state
* persist Local Profile stats directly
* use `local_profile_id` as a gameplay player ID
* treat a shared `/ws` URL as proof that single-player and multiplayer are the same flow

### WebSocket and boot authority

The WebSocket route is transport, not play-mode authority.

Current client boot state records:

```text
requested session mode
pending boot request type
pending local_profile_id
selected WebSocket URL
```

Single-player boot sends immediately after connection. Multiplayer create/join waits for WebSocket auth success or a token-verification-unavailable result so server-side admission can fail explicitly.

The single-player path does not wait for auth.

### Game-server authority

The game server owns the room and gameplay consequences of `start_single_player_request`.

For local single-player, the game server:

* rejects the request if the WebSocket session is already in a room
* creates a non-joinable room
* adds the requesting session as the only room member
* starts the room immediately
* moves the room through `Lobby -> Starting -> InGame`
* creates and starts the game instance
* begins the next match
* stores `local_profile_id` on the room member when the packet supplies one
* activates the connected room member into an active game player
* broadcasts a room snapshot

The game server does not require authenticated-account admission for this path.

### Room authority

Rooms own room membership and room lifecycle.

For local single-player, the room is created as:

```text
state = Lobby
joinable = false
member count = 1
```

The room then starts immediately and becomes `InGame`.

Room membership remains separate from active gameplay participation. The room member exists before the active game player is created. Active game-player routing is established by networking after the room start succeeds.

### Gameplay authority

Gameplay owns live match state after the room enters `InGame`.

Gameplay owns:

```text
movement
projectiles
collisions
damage
score
lives
death
respawn
player lifecycle
match-over decisions
state packet projection
```

The client renders these facts from server packets. It does not recalculate authoritative outcomes.

### Player-data authority

Player-data owns stat routing after match facts leave gameplay and room lifecycle.

For local single-player:

```text
Guest -> transient Guest route
Local Profile -> local profile route
```

The game server reports trusted match facts. Player-data routes and records those facts based on identity. Player-data does not calculate score, deaths, match-over state, or winner flags.

### API-server boundary

The API server is not required for local single-player startup.

Authenticated Account identity belongs to the API-server/Rails path and is used for authenticated multiplayer and account-backed persistence. It is not the intended identity for local single-player.

## Flow summary

### 1. Enter single-player pregame

```text
Main Menu
-> Single Player
-> Pregame Menu in single-player mode
-> apply saved Guest or Local Profile default
-> update callsign display
```

The client enters the shared pregame menu in single-player mode. The pregame mode controls which buttons are active and which actions are routed.

Single-player mode currently routes Play Endless and Select Pilot. Multiplayer create/join actions are not part of the single-player mode path.

### 2. Resolve the active local identity

Before boot, the client resolves the active single-player context.

Current identity contexts:

```text
Guest
= identity_kind guest
= no local_profile_id

Local Profile
= identity_kind local_profile
= selected local_profile_id
```

The client sends only the durable routing key needed by the server:

```text
local_profile_id
```

The client does not send callsign as gameplay identity, and the game server does not need the local profile display name to start the match.

### 3. Start local single-player from pregame

```text
Play Endless
-> read active single-player context
-> clear menu UI for gameplay
-> request single-player session boot
```

If the active context is Local Profile, the client passes `local_profile_id` into the session boot request. If the active context is Guest, the request carries an empty local profile ID.

### 4. Select the WebSocket target and connect

```text
request single-player
-> requested_mode = single_player
-> pending boot request = single_player
-> WebSocket URL = SINGLE_PLAYER_WS_URL
-> connect to game server
```

The current generated single-player and multiplayer WebSocket URLs both target the local Go game-server route during development.

This does not collapse the modes. The selected URL only chooses transport. The packet type and server-side room policy decide the flow.

### 5. Send the boot request after connection

```text
WebSocket connected
-> optional authenticate_request if a token exists
-> pending single-player request sends immediately
-> client sends start_single_player_request
-> client sends viewport config after boot request
```

Single-player boot does not wait for `authenticate_result`.

The generated packet shape includes:

```text
type = start_single_player_request
local_profile_id = selected local profile id or empty string
```

### 6. Create and start the server room

```text
start_single_player_request
-> reject if session already has a room
-> create non-joinable single-player room
-> add requesting session as member
-> start single-player game
-> store local_profile_id on member when supplied
-> activate connected member as active game player
-> broadcast room_snapshot
```

The room starts immediately. There is no multiplayer lobby, owner gate, ready gate, room-code entry, or join surface in this flow.

### 7. Client opens gameplay packet acceptance

```text
room_snapshot
-> room-session cache updates
-> requested mode becomes active mode
-> multiplayer lobby mount is skipped because active mode is single_player
-> room state is InGame
-> gameplay packet gate opens
```

The client does not process gameplay state as active gameplay until it observes server room state `InGame`.

After the gate opens, gameplay presentation consumes server state packets through the normal gameplay runtime and world-sync paths.

### 8. Gameplay runs through the normal authoritative path

During local single-player, gameplay still uses the server-authoritative realtime path:

```text
client input
-> WebSocket packet
-> game-server gameplay routing
-> authoritative simulation step
-> state packet
-> client presentation
```

The same simulation authority used by multiplayer owns local single-player movement, bullets, asteroids, pickups, score, lives, death, respawn, player lifecycle, and match-over classification.

### 9. Match over and results handoff

When the server determines the match is over:

```text
gameplay match decision
-> room transitions to GameOver
-> room stores resolved match summary
-> room snapshot exposes presentation-safe match_result
-> client presents match results
-> game-server reports trusted facts to player-data
```

For Guest single-player, match results route to Guest behavior.

For Local Profile single-player, the room member’s `local_profile_id` becomes the identity reference used by match reporting and player-data routing.

The client result payload does not expose account IDs or local profile IDs.

### 10. Replay, pregame return, or quit

After match over, the player can route out of the session through client-owned intent surfaces.

Current local single-player exits include:

```text
Replay
-> close current connection gracefully
-> reset gameplay/session state
-> request a new single-player boot

Return to pregame
-> begin graceful close
-> reset gameplay/session state
-> clear session and pending boot state
-> show single-player pregame

Quit to main menu
-> begin graceful close
-> reset gameplay/session state
-> clear session and pending boot state
-> show main menu
```

Replay starts a new local single-player session. It does not reuse the ended room as a multiplayer lobby.

## Inputs and outputs

## Inputs

Local single-player flow inputs include:

```text
main menu Single Player intent
active pregame mode
saved local profile default
identity_kind
local_profile_id
Play Endless intent
generated single-player WebSocket URL
WebSocket connected event
start_single_player_request
room lifecycle state
gameplay input packets
authoritative match facts
match-results button intent
```

## Outputs

Local single-player flow outputs include:

```text
requested session mode = single_player
pending boot request = single_player
start_single_player_request packet
non-joinable server room
active game player routing
room_snapshot
gameplay state packets
presentation-safe match_result
player-data match-result command
Guest or Local Profile stat update
client route intent after match end
```

## Data crossing the flow

The local single-player boot packet carries:

```text
type
local_profile_id
```

The server room snapshot publishes presentation state such as:

```text
room_code
room_state
members
local_player_id
owner_id
max_players
match_result
```

The gameplay state path publishes live gameplay presentation state such as:

```text
self_id
players
player_sessions
player_lifecycle
bullets
asteroids
pickups
events
lives
score
```

The match-result persistence path uses authoritative server facts such as:

```text
match_id
result_id
identity_kind
local_profile_id
score
ship_deaths
won
play_mode
```

## Invariants

* Local single-player is server-backed, not client-authoritative offline simulation.
* The WebSocket URL does not define play mode.
* `start_single_player_request` defines the local single-player room entry request.
* Single-player boot does not require authenticated-account admission.
* Guest and Local Profile are the intended local single-player identities.
* Local Profile identity is carried by `local_profile_id`, not callsign or display name.
* Gameplay player IDs remain separate from `local_profile_id`, `account_id`, WebSocket session IDs, and room member IDs.
* Single-player creates a non-joinable room and starts immediately.
* Multiplayer lobby readiness and owner start rules do not apply to local single-player start.
* Gameplay state remains server-authoritative.
* Match-result persistence starts from server-resolved match facts, not client-side counters.

## Active issues

* The client currently expects the Go game server to already be running. Local server launch from the Godot client is not implemented. See [Current System Limits](../../limits/current-system-limits.md#architecture--networking).
* `start_single_player_request` does not currently reject an already-authenticated WebSocket session directly at the game-server boundary. The intended identity model remains Guest or Local Profile for local single-player, and player-data mode validation rejects `single_player + authenticated_account`. See [Current System Limits](../../limits/current-system-limits.md#architecture--networking).
* Several single-player pregame buttons are currently disabled, including Campaign, Loadout, Provisioner, Buy Scrap, and Rankings. See [Current System Limits](../../limits/current-system-limits.md#client-menu-flow).

## Out of scope

This document does not define:

* exact packet source schema
* generated packet code details
* client menu scene implementation
* local pilot selector implementation
* WebSocket transport implementation
* room package code map
* gameplay simulation phase order
* Local Profile HTTP endpoint contracts
* SQLite schema
* Guest stat memory implementation
* Authenticated Account login flow
* multiplayer create/join admission details
* future offline mode
* future local server launch
* future campaign setup
* future loadout setup
* future progression, rewards, inventory, or leaderboard rules

Those belong in service, protocol, data, planning, limits, or player-experience domain documentation.

## Related docs

* [Technical](./!README.md)
* [Domains](../!README.md)
* [Player Experience Gameplay Session Flow](../player-experience/gameplay-session-flow.md)
* [Player Experience Local Pilot Profile Flow](../player-experience/local-pilot-profile-flow.md)
* [Platform Account And Identity Current State](../platform/account-and-identity-current-state.md)
* [Platform Player Data Routing Flow](../platform/player-data-routing-flow.md)
* [Client](../../services/client/!README.md)
* [Client Session Boot And Network Target](../../services/client/app-shell-and-session/session-boot-and-network-target.md)
* [Client Pregame Mode And Actions](../../services/client/pregame-menu-flow/pregame-mode-and-actions.md)
* [Client Local Pilot Flow](../../services/client/pregame-menu-flow/local-pilot-flow.md)
* [Game Server](../../services/game-server/!README.md)
* [Game Server Room Network Adapter](../../services/game-server/networking/room-network-adapter.md)
* [Game Server Lobby And Start Rules](../../services/game-server/rooms/lobby-and-start-rules.md)
* [Player Data](../../services/player-data/!README.md)
* [Protocol](../../protocol/!README.md)
* [Data](../../data/!README.md)
* [Current System Limits](../../limits/current-system-limits.md)

## Notes

The old local-singleplayer routing stub used the right topic boundary but did not contain current-system facts. This document replaces that placeholder shape with the current cross-system technical flow.

The useful legacy distinction remains valid: `local` is not enough information. A locally running server can host local single-player, multiplayer simulation, or authenticated multiplayer behavior. The current flow is identified by session mode, packet type, room policy, and player-data identity context.

Local single-player currently shares infrastructure with multiplayer in several places. That is intentional for the authoritative simulation path and does not make Local Profile, Guest, and Authenticated Account interchangeable.
