# Lobby And Start Rules

Parent index: [Game Server Rooms](./!INDEX.md)

## Purpose

This document describes the game-server room rules for lobby admission, ready state, multiplayer start, single-player start, and return-to-lobby readiness reset.

It covers the authoritative service-side rules under `services/game-server/internal/rooms/`, not client lobby presentation.

## Overview

Lobby and start rules are owned by the game-server rooms package.

The lobby is the only joinable pre-match room state. Clients may request room creation, room join, ready toggles, multiplayer start, single-player start, and return to lobby through realtime lobby packets, but those packets are only requests. The server validates room state, membership, ownership, readiness, joinability, and capacity before mutating room state.

The current multiplayer start path is:

```text
start_game_request
-> networking room adapter
-> RoomManager.StartRoomGame
-> room session resolves to room member player ID
-> Room.StartGameForMember
-> start preconditions
-> pure roomrules.DecideStart policy
-> Lobby -> Starting -> InGame
-> game instance starts
-> match ID advances
-> networking activates connected members as active game players
-> room_snapshot broadcast
```

The current single-player start path is:

```text
start_single_player_request
-> networking room adapter
-> RoomManager.CreateStartedSinglePlayerRoom
-> create non-joinable Lobby room
-> add requesting session as member
-> Room.StartSinglePlayerGame
-> Lobby -> Starting -> InGame
-> game instance starts
-> match ID advances
-> networking attaches local profile ID when supplied
-> networking activates the connected member as an active game player
-> room_snapshot broadcast
```

`Starting` is an implemented room state, but the current transition from `Lobby` through `Starting` to `InGame` is immediate. It still matters because admission and start validation treat `Starting` as closed to join/start requests.

## Code root

```text
services/game-server/internal/rooms/
services/game-server/internal/rooms/roomrules/
```

## Responsibilities

Lobby and start rules own:

* Join validation for existing rooms.
* Room joinability checks.
* Room capacity checks.
* Ready-state mutation while a room is in `Lobby`.
* Start validation for multiplayer rooms.
* Owner-only multiplayer start enforcement.
* Connected-member readiness enforcement before multiplayer start.
* Single-player room creation and immediate start.
* `Lobby -> Starting -> InGame` transition sequencing.
* Match ID advancement when a new match starts.
* Ready-state clearing when a game-over room returns to lobby.
* Room-domain error codes for rejected lobby/start operations.

## Does not own

Lobby and start rules do not own:

* WebSocket transport.
* Packet decoding or encoding.
* Client lobby UI presentation.
* Client-side Start button enablement.
* Authenticated-account admission for multiplayer create/join.
* Local profile storage.
* Active game player routing on WebSocket sessions.
* Gameplay simulation rules.
* Match-over policy.
* Match-result reporting sinks.
* Future lobby countdown presentation or timing policy.
* Future room discovery or matchmaking metadata.

## Domain roles

### Join admission

Join admission is evaluated by `roomrules.DecideJoin`.

A room can be joined only when:

```text
room state == Lobby
room Joinable == true
member count < MaxPlayersPerRoom
```

Rejected join outcomes include:

```text
invalid_room_code    -> requested room code is malformed
room_not_found       -> requested room does not exist
room_in_game         -> room state is Starting or InGame
room_closed          -> room state is Closed
invalid_room_state   -> room is not joinable or in an unknown/non-joinable state
room_full            -> room has reached MaxPlayersPerRoom
```

`RoomManager.JoinRoom` normalizes room codes before lookup and delegates room-state/capacity checks to the room.

### Ready state

Ready state is mutable only while the room is in `Lobby`.

The current ready path is:

```text
set_ready_request
-> RoomManager.SetReady
-> resolve session ID to room member player ID
-> Room.SetReadyInLobby
-> RoomMember.SetReady
-> room_snapshot broadcast
```

Ready mutation rejects when:

```text
room is missing
session is not in the room
room state is not Lobby
```

Ready state is stored on `RoomMember.Ready`. It is included in room snapshots as `members[].ready`.

### Multiplayer start validation

Multiplayer start is owner-gated and readiness-gated.

A multiplayer start is allowed only when:

```text
requesting session is in the room
requesting member is the room owner
room state is Lobby
room has at least one member
every connected member is ready
```

Disconnected members do not block start when they are unready.

The start policy is deliberately split:

```text
Room.validateStartPreconditionsLocked
-> checks aggregate room preconditions

roomrules.DecideStart
-> applies pure start policy from plain StartInput data
```

`roomrules.DecideStart` does not mutate room state. It only returns an allow/reject decision. The room aggregate adapts rejected decisions into `RoomDomainError`.

### Start transition

When multiplayer start is accepted, `Room.StartGameForMember` performs the room transition.

Current transition order:

```text
validate start
mark room Starting
create game instance if missing
start game loop
mark room InGame
begin next match
```

`BeginNextMatch` increments the room match number, sets the current match ID, clears any resolved match summary, and resets match-result-reported state.

The current transition is synchronous. There is no implemented lobby countdown, slow-client wait, or final ready lock in this service path.

### Single-player start

Single-player start uses the room system but bypasses multiplayer owner/readiness policy.

`RoomManager.CreateSinglePlayerRoom` creates a normal room aggregate with:

```text
state = Lobby
joinable = false
one member for the requesting session
```

`Room.StartSinglePlayerGame` then validates only the shared start preconditions:

```text
room state is Lobby
room has at least one member
```

After that it uses the same transition sequence:

```text
Lobby -> Starting -> InGame
create/start game instance
begin next match
```

The networking adapter stores the supplied `local_profile_id` on the room member after the single-player room is created and started.

### Active player activation

The room package starts the game and owns room state, but networking owns active WebSocket game-player routing.

After a successful start, networking calls `activateRoomPlayers`.

Activation:

```text
takes connected room members
finds their live WebSocket sessions
calls gameInstance.AddPlayer()
stores currentGamePlayerID on each session
rebinds the room member player ID through room.SetMemberPlayerIDForSession
increments room active-player count
```

This keeps room membership and active gameplay routing separate:

```text
WebSocket connection != room membership
room membership != active game player
active game player routing == networking session state
```

### Return to lobby readiness reset

Return to lobby is accepted only from `GameOver`.

`Room.ResetToLobby` requires:

```text
requesting player is still a room member
room state is GameOver
```

When accepted, it:

```text
sets all room members ready = false
stops the current game instance if present
clears the room game instance
sets room state to Lobby
```

Networking then deactivates active game player routing for connected sessions and broadcasts a room snapshot.

## Protocols and APIs

Lobby/start rules are reached through the realtime WebSocket lobby packet family.

The packet surface is for client requests and server room-state publication. Clients request room changes; the game server owns authority behind acceptance or rejection. Data crossing this boundary includes room codes, readiness values, local profile IDs for single-player start, room snapshots, and room errors.

Inbound request packet types relevant to this document:

```text
join_room_request
set_ready_request
start_game_request
start_single_player_request
return_to_lobby_request
```

Adjacent inbound packet type:

```text
create_room_request
```

`create_room_request` is routed by networking and creates a lobby room, but authenticated-account admission and packet handling are owned by the networking adapter.

Outbound packet types relevant to this document:

```text
room_snapshot
room_error
```

Room snapshots publish:

```text
room_code
room_state
members[].player_id
members[].ready
members[].connected
local_player_id
owner_id
max_players
match_result
```

Room errors publish:

```text
error_code
message
```

Packet schema source of truth:

```text
shared/packets/lobby.toml
```

Generated Go packet structs and constants:

```text
services/game-server/internal/game/packets.go
```

## Data ownership

Lobby/start rules mutate room runtime state only.

Room-owned state used by this boundary:

```text
Room.ID
Room.State
Room.Joinable
Room.membership.members
Room.membership.ownerID
Room.match.game
Room.match.activePlayers
Room.match.matchNumber
Room.match.currentMatchID
Room.match.resolvedSummary
Room.match.matchResultReported
```

Room-member state used by this boundary:

```text
RoomMember.MemberID
RoomMember.SessionID
RoomMember.PlayerID
RoomMember.AccountID
RoomMember.LocalProfileID
RoomMember.Ready
RoomMember.Connected
```

The room package does not persist lobby or start state to disk. It is live runtime state owned by the game-server process.

`MaxPlayersPerRoom` is sourced from `playerids.MaxPlayers` through the rooms constants file.

## Code map

### Primary implementation files

* `services/game-server/internal/rooms/manager.go` - Room manager create, find, join, ready, leave, cleanup, and error entry points.
* `services/game-server/internal/rooms/lifecycle.go` - Room manager start, single-player start, and return-to-lobby entry points.
* `services/game-server/internal/rooms/room.go` - Room aggregate root and core room fields.
* `services/game-server/internal/rooms/room_join.go` - Joinability validation and member join behavior.
* `services/game-server/internal/rooms/room_lobby.go` - Ready mutation and multiplayer start validation.
* `services/game-server/internal/rooms/room_lifecycle.go` - Room lifecycle transitions, game start, game-over, and reset-to-lobby behavior.
* `services/game-server/internal/rooms/room_membership.go` - Member storage, owner selection, player ID assignment, and all-ready reset.
* `services/game-server/internal/rooms/member.go` - Room member state and ready/connection mutators.
* `services/game-server/internal/rooms/constants.go` - Room states, limits, and room error codes.
* `services/game-server/internal/rooms/room_rule_adapter.go` - Adapter from pure room-rule decisions to room-domain errors.

### Pure rule files

* `services/game-server/internal/rooms/roomrules/decision.go` - Shared allow/reject decision type.
* `services/game-server/internal/rooms/roomrules/join.go` - Pure join policy.
* `services/game-server/internal/rooms/roomrules/start.go` - Pure start policy.

### Networking adapter participants

* `services/game-server/internal/networking/room_handlers.go` - Routes lobby/start packet handlers to room manager methods.
* `services/game-server/internal/networking/inbound/lobby.go` - Classifies generated lobby packet types.
* `services/game-server/internal/networking/player_activation.go` - Activates/deactivates WebSocket sessions as active game players after room lifecycle changes.
* `services/game-server/internal/networking/room_snapshot.go` - Builds and broadcasts authoritative room snapshots.
* `services/game-server/internal/networking/room_error.go` - Encodes rejected room operations as room error packets.

### Generated and source files

* `shared/packets/lobby.toml` - Lobby packet source of truth.
* `services/game-server/internal/game/packets.go` - Generated Go lobby packet structs and packet type constants.

### Important non-ownership boundaries

* `services/game-server/internal/networking/session_admission.go` owns authenticated-account admission for multiplayer create/join.
* `services/game-server/internal/networking/websocket.go` owns WebSocket connection lifetime and disconnect cleanup.
* `services/game-server/internal/game/` owns active gameplay simulation after a room starts.
* `services/game-server/internal/playerdata/` owns match result data shapes and persistence-facing records.
* `client/scripts/lobby/` owns client-side lobby presentation and request sending, not room authority.

## Tests

Relevant room tests:

* `services/game-server/internal/rooms/roomrules/join_test.go`

  * Verifies pure join policy for lobby, starting, in-game, closed, unknown, non-joinable, and full-room states.
* `services/game-server/internal/rooms/roomrules/start_test.go`

  * Verifies pure start policy for owner, non-owner, non-member, connected unready members, disconnected unready members, and invalid room states.
* `services/game-server/internal/rooms/room_join_test.go`

  * Verifies room join validation and room-domain error codes.
* `services/game-server/internal/rooms/room_lobby_test.go`

  * Verifies start validation and ready-state behavior.
* `services/game-server/internal/rooms/room_lifecycle_test.go`

  * Verifies multiplayer start, single-player start, match ID advancement, match-result state reset, game-over summary capture, and return-to-lobby readiness clearing.
* `services/game-server/internal/rooms/manager_test.go`

  * Verifies manager-level join behavior, room-code validation, missing-room rejection, non-joinable rejection, capacity rejection, and state rejection.

Relevant networking tests:

* `services/game-server/internal/networking/player_activation_test.go`

  * Verifies active player activation and member identity preservation behavior.
* `services/game-server/internal/networking/room_snapshot_test.go`

  * Verifies room snapshot projection behavior.
* `services/game-server/internal/networking/room_error_test.go`

  * Verifies room-domain errors are emitted as generated room error packets.
* `services/game-server/internal/networking/gameplay_packets_test.go`

  * Verifies single-player start stores supplied local profile ID on the room member.
* `services/game-server/internal/networking/websocket_test.go`

  * Verifies resolved match results are reported before requested leave or disconnect removes the room member.

Suggested verification command:

```text
go test -buildvcs=false ./services/game-server/internal/rooms ./services/game-server/internal/networking
```

## Related docs

* [Game Server Rooms](./!INDEX.md)
* [Game Server](../!INDEX.md)
* [Room Network Adapter](../networking/room-network-adapter.md)
* [Inbound Packet Routing](../networking/inbound-packet-routing.md)
* [Lobby Session and Presentation](../../client/lobby-flow/lobby-session-and-presentation.md)
* [Gameplay Session Lifecycle](../../client/gameplay-runtime/gameplay-session-lifecycle.md)
* [Lobby Packets](../../../protocol/lobby-packets.md) - incomplete lobby packet protocol documentation.
* [Packet Schema Pipeline](../../../data/packet-schemas.md) - incomplete packet schema pipeline documentation.
* [Player Experience Systems](../../../planning/gameplay/player-experience-systems.md)
* [Multiplayer Session And Lifecycle](../../../planning/domains/platform/stubs/multiplayer-session-and-lifecycle.md) - incomplete multiplayer session lifecycle planning documentation.

## Notes

Room/domain ownership lives in `services/game-server/internal/rooms`, while WebSocket transport and session routing live in `services/game-server/internal/networking`.

The current service path has no implemented lobby countdown. Countdown behavior is planned player-experience work, not current room-service behavior.
