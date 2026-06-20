# Room Manager

Parent index: [Game Server Rooms](./!INDEX.md)

## Purpose

This document describes the game-server room manager boundary.

It covers top-level room registry ownership, room code creation, room lookup, manager-level create/join/leave/ready/start/reset entry points, cleanup scheduling, and the boundary between room state and WebSocket networking.

## Overview

The room manager lives in `services/game-server/internal/rooms`.

`RoomManager` is the in-memory registry for active rooms inside one game-server process. It owns the top-level map from room ID/code to `*Room`, the cleanup grace delay used when rooms become empty, and manager-level entry points consumed by the networking layer.

The room manager does not own WebSocket transport. Networking decodes generated lobby packets, tracks live connection/session state, and calls room manager methods. The room manager owns the authoritative room decisions behind those requests.

Current room manager shape:

```text
RoomManager
  mutex
  rooms map[string]*Room
  cleanupDelay
```

Current room shape managed by the manager:

```text
Room
  ID
  State
  Joinable
  roomMembership
  roomMatch
  roomCleanup
  mutex
```

The manager creates two broad kinds of rooms:

* lobby rooms for multiplayer create/join flow
* non-joinable single-player rooms for local/single-player start flow

Room codes are generated server-side. A valid room code is six characters long and uses the room-code alphabet from `rooms/constants.go`:

```text
ABCDEFGHJKLMNPQRSTUVWXYZ23456789
```

Room code input is normalized to uppercase and trimmed before lookup. Invalid room-code shape is rejected before the room map is consulted.

The manager protects the room map with its own mutex. Individual room state is protected by each `Room` mutex. Most manager methods look up the room under the manager lock, release that lock, and then delegate room-specific mutation to the room aggregate. Cleanup and process shutdown paths coordinate map removal and room cleanup from the manager boundary.

## Code root

`services/game-server/internal/rooms/`

Supporting roots:

* `services/game-server/internal/networking/`
* `services/game-server/internal/game/`
* `services/game-server/internal/logging/`

## Responsibilities

The room manager owns:

* creating the process-local room registry
* configuring room cleanup delay
* storing active rooms by room ID/code
* finding rooms by normalized room ID
* creating lobby rooms with generated room codes
* creating non-joinable single-player rooms
* creating and immediately starting single-player rooms
* validating room-code shape for join requests
* joining sessions to existing rooms through room-domain join rules
* leaving rooms by session ID
* removing active game players during member leave when a game player ID is supplied
* scheduling cleanup when rooms become empty
* deleting empty rooms after the cleanup grace delay when the cleanup version is still current
* setting lobby readiness through room-domain rules
* starting multiplayer room games through room-domain start rules
* returning game-over rooms to the lobby
* stopping cleanup timers and game instances during process-level `StopAll`
* returning room-domain errors as stable code/message pairs for networking to convert into room error packets
* emitting room lifecycle diagnostics through `logging.Rooms`

## Does not own

The room manager does not own:

* WebSocket upgrade behavior
* WebSocket read/write loops
* packet decoding or encoding
* generated lobby packet schemas
* room snapshot packet construction
* room error packet construction
* live WebSocket session registry by room
* per-connection session identity
* authenticated-account token verification
* local profile or account persistence
* player-data store routing
* match-result persistence sinks
* game simulation mechanics
* collision, scoring, spawn, damage, or respawn rules
* client lobby presentation
* client match-result presentation
* process startup beyond being constructed by startup
* process shutdown beyond exposing `StopAll`

## Domain roles

The room manager is the game-server authority for the active-room registry.

It participates in the player experience flow where a client connection becomes a room member, the room enters gameplay, the room reaches game over, and the room either returns to lobby or is cleaned up.

Important state boundaries:

* A WebSocket connection is not automatically a room member.
* A room member is not automatically an active game player.
* Active game players are created when networking activates room members after a successful room start.
* The room manager owns room lookup and room-level operations.
* The room aggregate owns room state, membership state, match state, and cleanup state.
* Networking owns the live session pointer, current room pointer, current room ID, and current active game player ID for each WebSocket connection.
* The game instance owns simulation state after the room starts a match.
* Player-data owns durable profile and match-result persistence.

Identity boundaries:

* `sessionID` is WebSocket/session identity supplied by networking.
* `MemberID` is internal room membership identity.
* `PlayerID` is the player-facing room/game identity such as `Player-1`.
* `AccountID` and `LocalProfileID` are copied onto room members by networking when relevant.
* `currentGamePlayerID` is networking-owned active gameplay routing state, not room membership identity.

## Protocols and APIs

The room manager is not an external network protocol surface. It is an internal Go service API consumed by game-server networking.

The client-facing surface is the realtime WebSocket lobby packet family. Networking owns packet decode, transport state, and outbound packet delivery. The room manager owns the room authority behind successful or failed operations. Data crossing this internal boundary includes session IDs, room IDs, room codes, readiness values, active game player IDs for leave cleanup, room pointers, member state, room state, and room-domain error codes.

The room manager explicitly does not own packet shape, packet serialization, WebSocket lifetime, or client presentation.

### Construction and lookup

```text
NewRoomManager()
NewRoomManagerWithCleanupDelay(cleanupDelay)
Find(roomID)
RoomCount()
```

`NewRoomManager()` constructs a manager with `RoomCleanupGraceTime`.

`NewRoomManagerWithCleanupDelay()` is used by tests and any caller that needs a custom cleanup delay.

`Find()` normalizes an empty room ID to `DefaultRoomID`, looks up the room under the manager lock, and returns the room pointer plus an existence flag.

`RoomCount()` returns the current number of rooms tracked by the manager.

### Room creation

```text
CreateLobbyRoom()
CreateSinglePlayerRoom(sessionID)
CreateStartedSinglePlayerRoom(sessionID)
```

`CreateLobbyRoom()` generates a unique room code, creates a lobby room, stores it in the manager map, and leaves membership empty. Networking adds the creating session as a member afterward so it can also attach account identity and live WebSocket session state.

`CreateSinglePlayerRoom()` generates a unique room code, creates a lobby-state room, marks it non-joinable, adds the requesting session as the only member, and stores the room in the manager map.

`CreateStartedSinglePlayerRoom()` creates a single-player room and immediately starts its game instance through the room lifecycle path.

Room code generation retries up to 16 times to avoid collisions in the manager map. If generation fails or uniqueness cannot be achieved, creation returns an error.

### Join and readiness

```text
JoinRoom(sessionID, roomCode)
SetReady(roomID, sessionID, ready)
```

`JoinRoom()` normalizes and validates the room code before lookup. Invalid shape returns `invalid_room_code`. Missing room returns `room_not_found`.

After lookup, join policy is delegated to the room aggregate. Join is allowed only when the room is in lobby state, is joinable, and is below `MaxPlayersPerRoom`.

`SetReady()` requires the room to exist, requires the room to be in lobby state, resolves the session to a room member, and then delegates the readiness mutation to the room aggregate.

### Match lifecycle

```text
StartRoomGame(roomID, sessionID)
ReturnRoomToLobby(roomID, sessionID)
```

`StartRoomGame()` resolves the session to a room member and delegates start policy to the room aggregate. The room aggregate enforces owner-only start, connected-member readiness, valid lobby state, game instance creation, game start, transition to in-game state, and match ID creation.

`ReturnRoomToLobby()` resolves the session to a room member and delegates reset policy to the room aggregate. The room aggregate only allows return from `GameOver`, clears ready state, stops and clears the game instance, and moves the room back to `Lobby`.

Networking is responsible for activating or deactivating live WebSocket sessions around these room lifecycle transitions.

### Leave and cleanup

```text
LeaveRoom(roomID, sessionID)
LeaveMember(roomID, sessionID, playerID)
ScheduleCleanupIfEmpty(roomID)
StopAll()
```

`LeaveRoom()` removes a room member by session ID and returns the room, room ID, session ID, and remaining member count.

`LeaveMember()` extends `LeaveRoom()` for networking. If an active game player ID is supplied and the room has a game instance, it removes that player from the game and decrements the room active-player count. It schedules cleanup when the room becomes empty and reports whether a room snapshot should still be broadcast to remaining members.

`ScheduleCleanupIfEmpty()` schedules cleanup only when the room exists and `room.ShouldCleanup()` is true. Cleanup uses a version number stored in `roomCleanup` so stale timers cannot delete rooms after a later cleanup scheduling event has superseded them.

`StopAll()` stops cleanup timers, stops any running game instances, logs room stop events, and removes all rooms from the manager map. Startup registers this method as the current process-exit cleanup hook.

## Data ownership

The room manager owns in-memory process runtime state only.

Manager-owned data:

* `rooms map[string]*Room`
* manager mutex
* cleanup delay

Room-owned data under manager supervision:

* room ID
* room state
* joinability flag
* room membership state
* room owner ID
* member readiness
* member connected state
* member account/local-profile identity fields
* room game instance pointer
* active player count
* match number
* current match ID
* resolved match result summary
* match-result reported flag
* cleanup timer
* cleanup version

The room manager does not persist room data. Restarting the game-server process loses active room state.

The manager does not own WebSocket session registry data. Live session attachment is stored in `services/game-server/internal/networking/room_sessions.go`.

The manager does not own durable player data. Account/local-profile persistence and match-result mutation belong to player-data and the match-result reporting integration.

The manager does not own packet source data. Lobby packet definitions live under `shared/packets/`, and generated Go packet structs are consumed by networking.

## Error behavior

Room manager and room-domain operations return `RoomDomainError` for expected room failures.

Current room error codes include:

```text
room_not_found
room_closed
room_in_game
room_full
already_in_room
not_in_room
invalid_room_code
not_ready
not_room_owner
invalid_room_state
auth_required
invalid_token
token_verification_unavailable
```

The room manager directly returns some errors, such as invalid room code and missing room. Room aggregate and room rule adapters return policy errors such as room full, room in game, not ready, or not room owner.

Networking converts these code/message pairs into generated `room_error` packets.

Authentication-related room errors are defined in the room constants file for packet consistency, but authentication admission is enforced in networking before create/join reaches the room manager.

## Cleanup behavior

Room cleanup is manager-owned at the registry level and room-owned at the timer/version level.

Cleanup sequence:

```text
member leaves
  LeaveMember removes room member
  active game player is removed when playerID is supplied
  active player count may be decremented
  room.ShouldCleanup() is checked
  manager schedules cleanup if empty
  cleanup timer captures cleanup version
  timer fires after cleanup delay
  manager re-checks room existence
  manager re-checks active player count
  manager re-checks member count
  manager re-checks cleanup version
  game instance is stopped if present
  room is deleted from manager map
```

Cleanup is skipped when:

* the room was already removed
* the room has active players
* the room has members
* the cleanup version is stale

The current cleanup grace time is:

```text
30 seconds
```

`StopAll()` is separate from delayed cleanup. It immediately stops timers, stops game instances, logs each stopped room, and removes rooms from the map.

## Code map

Primary implementation files:

* `services/game-server/internal/rooms/manager.go` - Room manager type, room map ownership, create/join/leave/ready, cleanup scheduling, cleanup execution, and `StopAll`.
* `services/game-server/internal/rooms/lifecycle.go` - Manager-level start, single-player start, and return-to-lobby entry points.
* `services/game-server/internal/rooms/leave.go` - Leave-member result shape, active game player removal, cleanup scheduling, and broadcast hinting.
* `services/game-server/internal/rooms/code.go` - Room ID/code normalization, validation, and generation.
* `services/game-server/internal/rooms/constants.go` - Room states, limits, cleanup grace time, room code settings, and room error codes.
* `services/game-server/internal/rooms/room.go` - Room aggregate shape.
* `services/game-server/internal/rooms/room_join.go` - Joinable state and join validation/mutation.
* `services/game-server/internal/rooms/room_lobby.go` - Ready state and start precondition policy adapter.
* `services/game-server/internal/rooms/room_lifecycle.go` - Room lifecycle transitions, game instance lifecycle, game-over, and reset-to-lobby behavior.
* `services/game-server/internal/rooms/room_members.go` - Room member access, session-to-player lookup, owner/member operations, and identity attachment.
* `services/game-server/internal/rooms/room_membership.go` - Membership map, owner selection, player ID allocation, and member snapshots.
* `services/game-server/internal/rooms/room_match.go` - Game instance pointer, active-player count, match ID sequencing, resolved summary, and report state.
* `services/game-server/internal/rooms/room_match_access.go` - Room match accessors.
* `services/game-server/internal/rooms/room_cleanup.go` - Cleanup timer access, cleanup scheduling, cleanup version checks, and game stop helper.
* `services/game-server/internal/rooms/room_cleanup_state.go` - Cleanup timer/version storage.
* `services/game-server/internal/rooms/room_rule_adapter.go` - Pure room rule decision to room-domain error adapter.
* `services/game-server/internal/rooms/roomrules/` - Pure join/start policy decisions.

Networking callers:

* `services/game-server/internal/networking/rooms.go` - Networking wrapper constructors for room manager creation.
* `services/game-server/internal/networking/websocket.go` - Injects the room manager into WebSocket sessions and handles leave-on-disconnect.
* `services/game-server/internal/networking/websocket_session.go` - Stores the room manager pointer on each WebSocket session.
* `services/game-server/internal/networking/room_handlers.go` - Calls room manager methods for create, join, leave, ready, start, single-player start, and return to lobby.
* `services/game-server/internal/networking/room_sessions.go` - Owns live WebSocket session attachment by room/member session ID.
* `services/game-server/internal/networking/player_activation.go` - Activates/deactivates live room members as game players after manager-controlled lifecycle transitions.
* `services/game-server/internal/networking/room_snapshot.go` - Projects room state to generated room snapshot packets.
* `services/game-server/internal/networking/room_error.go` - Projects room-domain errors to generated room error packets.

Process caller:

* `services/game-server/cmd/game-server/main.go` - Constructs the room manager through networking and defers `StopAll()`.

Related systems:

* `services/game-server/internal/game/` - Owns simulation after a room starts a game instance.
* `services/game-server/internal/logging/logger.go` - Provides `logging.Rooms`.
* `services/game-server/internal/matchreporting/` - Owns reporter construction for match-result persistence.
* `services/player-data/` - Owns durable player-data and match-result storage.

Important non-ownership boundaries:

* `services/game-server/internal/networking/` owns WebSocket transport and live connection state.
* `services/game-server/internal/protocol/packetcodec/` owns JSON packet encode/decode.
* `shared/packets/` owns packet source definitions.
* `services/game-server/internal/game/` owns simulation mechanics.
* `services/player-data/` owns player-data persistence and store routing.
* `services/api-server/` owns Rails auth and authenticated-account persistence.

## Tests

Room manager and room behavior are covered by package-level and service-level tests.

Primary room manager tests:

* `services/game-server/internal/rooms/manager_test.go`
* `services/game-server/tests/rooms/manager_test.go`

Related room package tests:

* `services/game-server/internal/rooms/room_join_test.go`
* `services/game-server/internal/rooms/room_lobby_test.go`
* `services/game-server/internal/rooms/room_lifecycle_test.go`
* `services/game-server/internal/rooms/room_members_test.go`
* `services/game-server/internal/rooms/room_ownership_test.go`
* `services/game-server/internal/rooms/room_match_summary_test.go`
* `services/game-server/internal/rooms/lifecycle_tick_test.go`
* `services/game-server/internal/rooms/player_ids_test.go`
* `services/game-server/internal/rooms/roomrules/join_test.go`
* `services/game-server/internal/rooms/roomrules/start_test.go`

Related networking tests:

* `services/game-server/internal/networking/room_sessions_test.go`
* `services/game-server/internal/networking/player_activation_test.go`
* `services/game-server/internal/networking/room_snapshot_test.go`
* `services/game-server/internal/networking/room_error_test.go`
* `services/game-server/internal/networking/websocket_test.go`
* `services/game-server/internal/networking/gameplay_packets_test.go`
* `services/game-server/tests/networking/rooms_test.go`

Useful verification command:

```text
go test -buildvcs=false ./services/game-server/internal/rooms ./services/game-server/internal/networking ./services/game-server/tests/rooms ./services/game-server/tests/networking
```

## Related docs

* [Game Server Rooms](./!INDEX.md)
* [Game Server](../!INDEX.md)
* [Game Server Process](../process/!INDEX.md)
* [Service Startup](../process/service-startup.md)
* [Service Shutdown](../process/service-shutdown.md)
* [Game Server Networking](../networking/!INDEX.md)
* [Room Network Adapter](../networking/room-network-adapter.md)
* [Game Server Integrations](../integrations/!INDEX.md)
* [Match Result Reporting](../integrations/match-result-reporting.md)
* [Game Server Observability](../observability/!INDEX.md)
* [Logging And Diagnostics](../observability/logging-and-diagnostics.md)
* [Protocol](../../../protocol/!INDEX.md)
* [Services index](../../!INDEX.md)

## Notes

Legacy documentation confirmed the current split between networking and rooms: networking owns WebSocket transport and session routing, while `internal/rooms` owns room state, membership, lifecycle, cleanup policy, and game-instance lifecycle attachment.

The room manager is process-local. It is not a distributed room registry, not durable storage, and not a matchmaking service.

The room manager currently generates room codes directly. If public room discovery or matchmaking is added later, that should be documented as a separate platform/domain or service boundary rather than being folded into this manager by default.
