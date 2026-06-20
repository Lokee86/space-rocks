# Gameplay Session Flow

Parent index: [Player Experience](./!INDEX.md)

## Purpose

This document describes the cross-system gameplay session flow for Space Rocks.

It covers how a player-facing play session moves from menu intent into live gameplay, how room and gameplay state cross service boundaries, how the client presents the session, how match-over state resolves, and how results leave the live session flow.

## Overview

A gameplay session is the player-facing flow that starts when the player chooses a play mode and ends when the session returns to lobby, pregame, main menu, or replay boot.

The flow spans:

* client menu and session boot
* realtime WebSocket connection
* optional authenticated account admission
* room creation or room entry
* server-owned room lifecycle
* server-owned gameplay simulation
* client gameplay presentation
* authoritative match-over detection
* match-result presentation and reporting

The client does not directly enter gameplay by changing scenes. It records the requested session mode, opens or reuses a WebSocket connection, sends a boot request, receives server room state, then begins accepting gameplay packets only after the server room reaches `InGame`.

The game server owns the live room and match lifecycle. It creates or joins rooms, starts games, activates connected room members into active game players, advances simulation, detects match completion, builds resolved match summaries, and broadcasts room snapshots.

The client owns presentation and local lifecycle orchestration. It renders lobby state, gameplay state, HUD, world entities, input, menus, match-end presentation, and result-window routing from server-observed facts.

Player-data and API-server systems participate only where identity and result persistence cross the gameplay-session boundary. They do not own live gameplay simulation or room lifecycle.

## Participating systems

* [Client](../../services/client/!INDEX.md) - owns menu intent, session boot, transient room-session cache, gameplay presentation, input collection, match-end UI, and route execution after player intent.
* [Game Server](../../services/game-server/!INDEX.md) - owns realtime WebSocket handling, room lifecycle, room membership, gameplay simulation, match-over authority, room snapshots, and match-result reporting.
* [Player Data](../../services/player-data/!INDEX.md) - owns identity-based stats and match-result store routing after the game server reports resolved match facts.
* [API Server](../../services/api-server/!INDEX.md) - owns authenticated-account auth and Rails/Postgres persistence behind account-backed player-data flows.
* [Protocol](../../protocol/!INDEX.md) - owns the realtime packet contract used by session boot, room state, gameplay state, auth, and match-result presentation payloads.
* [Data](../../data/!INDEX.md) - owns generated packet and constants source material consumed by client and game-server session flows.

## Authority boundaries

The gameplay session uses separate authority layers.

A WebSocket connection is not room membership. A connected client can be unauthenticated, unaffiliated with a room, or waiting to send a pending boot request.

Room membership is not active gameplay participation. A player can be a room member in the lobby before any active game player exists.

Active gameplay participation begins only after a successful server-side start flow. The game server activates connected room members into active game players after the room lifecycle accepts the start request.

The game server owns authoritative room state:

```text
Lobby
-> Starting
-> InGame
-> GameOver
-> Lobby
```

The game aggregate owns authoritative simulation facts:

```text
movement
projectiles
collisions
damage
score
lives
death
respawn
match-over decision
state packet projection
```

The client owns only local presentation and routing consequences:

```text
menu display
lobby display
gameplay packet acceptance gate
world rendering
HUD
gameplay menu
match results window
local reset and navigation after player intent
```

The API server owns account authentication and account persistence. It does not own live rooms, WebSocket gameplay, match lifecycle, collisions, score during a match, lives, death, respawn, or authoritative state packets.

Player-data owns match-result/stat storage routing. It does not own live room state, gameplay packet routing, or match-over decisions.

## Flow summary

### 1. Player chooses a session path

The gameplay session begins from client menu intent.

Current boot request paths are:

```text
single_player
create_room
join_room
```

Single-player records a local profile id when one is selected. Multiplayer create and join paths use the multiplayer session mode and are later gated by WebSocket auth.

The client selects a WebSocket target from session mode. Current local development URLs for single-player and multiplayer both point to the same `/ws` game-server route, but the route path does not define play mode. Mode is expressed through the boot request and server-side policy.

### 2. Client opens the realtime connection

The client connection service opens the WebSocket connection and routes decoded packets into classified client signals.

When the connection opens:

* pending single-player boot requests are sent immediately
* pending multiplayer boot requests wait for successful WebSocket auth
* when token verification is unavailable, multiplayer boot requests may still be sent so the server can fail admission explicitly

The client sends viewport configuration after the boot request is sent, not merely when the socket opens.

### 3. Server admits or rejects the room request

The game server receives generated realtime packets over `/ws`.

Single-player start:

```text
start_single_player_request
-> create non-joinable single-player room
-> attach live session
-> attach local_profile_id when supplied
-> start game immediately
-> activate connected member into active game player
-> broadcast room snapshot
```

Multiplayer create:

```text
create_room_request
-> require authenticated account
-> create lobby room
-> add creating session as room member
-> attach account identity when present
-> send room snapshot to creating session
```

Multiplayer join:

```text
join_room_request
-> require authenticated account
-> validate room code and joinability
-> add session as room member
-> attach account identity when present
-> broadcast room snapshot
```

The server returns room errors for rejected room operations. The client presents those errors through room-entry or lobby presentation surfaces.

### 4. Lobby state is presented for multiplayer

Room snapshots carry presentation-safe room state.

The client caches room-session facts and applies lobby-specific fields into a transient lobby read model:

```text
room_code
room_state
local_player_id
owner_id
max_players
members
match_result
```

The multiplayer lobby is shown only when the active session mode is multiplayer and the room state is `Lobby`.

The client can present owner, readiness, connected state, and whether the local owner can press Start. That is presentation logic only. The server remains authoritative for readiness mutation and start acceptance.

### 5. Server starts gameplay

For multiplayer, the owner sends a start request from the lobby. The room lifecycle validates start rules, including ownership and readiness, then moves through:

```text
Lobby
-> Starting
-> InGame
```

For single-player, the room is created as non-joinable and started immediately.

When a match starts, the room creates or reuses a game instance, starts the simulation loop, increments the match number, sets a current match ID, clears any previous resolved match summary, and clears the reported-result flag.

The server activates connected room members into active game players after start succeeds. That activation is the bridge from room membership into gameplay packet routing.

### 6. Client begins gameplay presentation

The client does not process gameplay state packets as active gameplay until server-observed room state reaches `InGame`.

Room snapshots and room-state-change packets update the client room-session cache. After that update, the client session network controller opens the gameplay packet gate when the current room state is `InGame`.

Current client transition:

```text
room snapshot or room-state-change
-> room-session cache updates current room state
-> room state is InGame
-> gameplay-session lifecycle begins accepting gameplay packets
-> gameplay state packets apply to runtime and world presentation
```

The client then presents gameplay through world sync, HUD, input, audio/effects, devtools presentation, gameplay menu, respawn, spectate, and match-end flows.

### 7. Gameplay packets and state continue during the match

During live gameplay, the client sends input and gameplay requests through the realtime protocol. The game server routes those packets only when the session is attached to a room and has an active game player ID.

The game server advances simulation and emits authoritative gameplay presentation state. The client renders from those packets and does not recalculate authoritative outcomes.

Current authoritative gameplay facts include:

```text
player movement state
projectiles
asteroids
pickups
score
lives
death
respawn lifecycle
player lifecycle
target state
camera/view state
presentation events
match-over classification
```

The client may locally present death, respawn, game-over HUD state, and effects from server packets and events, but those presentations do not become match authority.

### 8. Local elimination and room match-over diverge

Local elimination and authoritative room match-over are separate states.

Local elimination means the local player has reached zero lives. The client can update local HUD/menu presentation and play game-over audio from this event.

Authoritative room match-over means the server room state is `GameOver`. Only room match-over shows match results and locks the gameplay HUD for the completed room.

This distinction prevents the client from presenting final room results just because one player has been eliminated.

### 9. Server resolves match-over

The game server room lifecycle observes the game aggregate’s match decision. When the game is complete, the room transitions from `InGame` to `GameOver`.

During that transition, the room builds and stores a resolved match summary if one has not already been stored.

The resolved summary is based on authoritative game facts, including:

```text
match_id
mode
game_player_id
score
ship_deaths
winner flag
identity context for reporting
```

Room snapshots expose a presentation-safe match result summary to clients. That snapshot result intentionally excludes account IDs and local profile IDs.

### 10. Client presents match results

After room state reaches `GameOver`, the client match-end flow reads the current room state and cached match-result payload through providers from the room-session cache.

The client then:

```text
hide and lock HUD for match over
enable match-over gameplay menu overlay
request game-over audio
read cached match_result
convert result players into presentation rows
show the match results window
```

The match results window can emit intent for replay, lobby return, pregame return, or main-menu quit. Those are client route intents. They do not alter server authority by themselves.

### 11. Results are reported to player-data

The game server reports resolved match results once through its match-result reporting boundary.

Current reporting flow:

```text
room game-over lifecycle
-> resolved match summary
-> once-only room reporting gate
-> game-server match-result reporter
-> player-data command per player
-> player-data store routing
```

The player-data runtime routes the result by identity and mode:

```text
authenticated account -> API-server-backed account persistence
local profile -> local profile storage
guest -> guest behavior
```

The game server marks the result reported only after the reporter succeeds. If reporting fails, the stored resolved summary remains available for a later retry path.

### 12. Player exits or repeats the session

After match over, player intent can route through one of several paths.

Return to lobby:

```text
client sends return_to_lobby_request
-> server validates room can return from GameOver
-> room clears ready states
-> room stops and clears game instance
-> room returns to Lobby
-> server deactivates active game players
-> server broadcasts room snapshot
```

Replay:

```text
client closes gracefully
-> client resets gameplay/session state
-> client clears pending boot state
-> client emits replay request
-> app boot flow starts a new session request
```

Return to pregame:

```text
client begins graceful close
-> client resets gameplay/session state
-> client clears pending boot state
-> app shows the relevant pregame flow
```

Quit to main menu:

```text
client begins graceful close
-> client resets gameplay/session state
-> client clears session and boot state
-> client shows main menu
```

Requested leave and disconnect cleanup also attempt to report an already-resolved match result before the server removes the room member.

## Inputs and outputs

## Client inputs

The gameplay session begins from local player intent:

```text
single-player play request
multiplayer create request
multiplayer join request
lobby ready toggle
lobby start request
gameplay input
pause/menu input
respawn request
targeting request
match-results button intent
```

## Client outputs

The client sends generated realtime packets such as:

```text
start_single_player_request
create_room_request
join_room_request
set_ready_request
start_game_request
input
pause_request
respawn
targeting requests
return_to_lobby_request
leave_room_request
```

The client also sends bearer-token authentication packets when multiplayer auth is available.

## Server inputs

The game server receives:

```text
WebSocket connection state
authentication packets
room/lobby packets
gameplay packets
client viewport config
disconnect or close events
```

The server converts those inputs into room operations, gameplay input, simulation state, or error responses.

## Server outputs

The game server sends:

```text
authenticate_result
room_snapshot
room_error
room_state_changed
gameplay state packets
player pause state packets
debug/status packets when relevant
telemetry responses when relevant
```

Room snapshots are the main cross-session state envelope. Gameplay state packets are the main live simulation presentation envelope.

## Durable outputs

The durable output of a completed gameplay session is the reported match result. That output flows through player-data and, when authenticated-account backed, into the API-server persistence path.

The presentation-safe result shown to the client is not the durable persistence contract. It is a room snapshot projection for UI.

## Session modes and identity

The current gameplay session distinguishes session mode from identity.

Session mode answers how the player is trying to play:

```text
single_player
multiplayer
```

Identity answers how player-data should route durable stats and results:

```text
guest
local_profile
authenticated_account
```

Single-player may use Guest or Local Profile identity and does not require signed-in account auth.

Multiplayer room create and join require Authenticated Account admission.

Gameplay player IDs remain player-facing gameplay/session identifiers. They are not replaced by account IDs, Rails user IDs, local profile IDs, WebSocket session IDs, or room member IDs.

## Out of scope

This domain doc does not own:

* direct code maps
* packet field-by-field protocol specification
* WebSocket transport implementation
* room package implementation details
* gameplay simulation phase order
* client world-sync implementation
* HUD widget implementation
* match-results UI rendering implementation
* player-data store internals
* Rails auth implementation details
* future matchmaking design
* future progression or leaderboard design

Those details belong in service, protocol, data, systems-design, planning, or limits documentation.

## Related docs

* [Player Experience](./!INDEX.md)
* [Client](../../services/client/!INDEX.md)
* [Game Server](../../services/game-server/!INDEX.md)
* [Player Data](../../services/player-data/!INDEX.md)
* [API Server](../../services/api-server/!INDEX.md)
* [Protocol](../../protocol/!INDEX.md)
* [Data](../../data/!INDEX.md)
* [Session Boot And Network Target](../../services/client/app-shell-and-session/session-boot-and-network-target.md)
* [Room Session State](../../services/client/app-shell-and-session/room-session-state.md)
* [Lobby Flow](../../services/client/lobby-flow/!INDEX.md)
* [Gameplay Runtime](../../services/client/gameplay-runtime/!INDEX.md)
* [Gameplay Session Lifecycle](../../services/client/gameplay-runtime/gameplay-session-lifecycle.md)
* [Match End Flow](../../services/client/match-end-flow/!INDEX.md)
* [Game Server Networking](../../services/game-server/networking/!INDEX.md)
* [Room Network Adapter](../../services/game-server/networking/room-network-adapter.md)
* [Auth Routing](../../services/game-server/networking/auth-routing.md)
* [Game Server Rooms](../../services/game-server/rooms/!INDEX.md)
* [Room Match Lifecycle](../../services/game-server/rooms/room-match-lifecycle.md)
* [Game Server Simulation](../../services/game-server/simulation/!INDEX.md)
* [Match Result Reporting](../../services/game-server/integrations/match-result-reporting.md)
* [Current System Limits](../../limits/current-system-limits.md)

## Notes

WebSocket connection, room membership, and active gameplay participation are separate states. That separation remains central to the current gameplay session flow.

The current local development WebSocket targets for single-player and multiplayer may be the same URL. This does not make single-player and multiplayer the same session flow. The server-side room request, admission rules, room joinability, and player-data identity context distinguish them.

Room match-over and local player elimination must stay separate in documentation and implementation. Local elimination is a player presentation state; room match-over is the authoritative session completion state.
