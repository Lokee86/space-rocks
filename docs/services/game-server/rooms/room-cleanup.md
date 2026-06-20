# Room Cleanup

Parent index: [Game Server Rooms](./!README.md)

## Purpose

This document describes the current game-server room cleanup boundary.

It covers empty-room cleanup scheduling, cleanup timer/version behavior, room removal from the manager map, game stop behavior during cleanup, and the difference between room cleanup, WebSocket session exit, and process-level room-manager teardown.

## Overview

Room cleanup is owned by the game-server rooms package.

The cleanup system removes rooms from the in-memory room manager after they no longer have members or active gameplay players. Cleanup is delayed by a grace timer so room removal does not happen directly inside leave or disconnect handling.

Current cleanup flow:

```text
member leaves or disconnects
-> networking asks rooms to remove the member/player
-> rooms checks whether the room is empty
-> empty room schedules delayed cleanup
-> cleanup timer fires
-> manager re-checks room existence, active players, members, and cleanup version
-> eligible room stops its game, if present
-> manager deletes the room from its room map
```

Cleanup is intentionally defensive. A scheduled cleanup callback does not delete a room just because a timer fired. It re-checks current room state under the room manager before removal.

The cleanup boundary does not decide why a client left, does not close WebSocket connections, does not broadcast room snapshots, and does not report match results. Those behaviors belong to the WebSocket session lifecycle, networking room handlers, and match-result reporting integration.

## Code root

```text
services/game-server/internal/rooms/
```

Related call sites live in:

```text
services/game-server/internal/networking/
services/game-server/cmd/game-server/
```

## Responsibilities

Room cleanup owns:

* deciding whether a room is cleanup-eligible
* scheduling delayed cleanup for empty rooms
* storing each room cleanup timer handle
* incrementing cleanup versions when cleanup is scheduled
* stopping older cleanup timers when a new cleanup timer is scheduled for the same room
* rejecting stale cleanup callbacks by version
* re-checking room existence before cleanup
* re-checking active player count before cleanup
* re-checking member count before cleanup
* stopping a room game instance before deleting the room
* deleting cleaned rooms from the room manager map
* stopping cleanup timers during room-manager teardown
* deleting all rooms during room-manager teardown
* logging room cleanup lifecycle events through `logging.Rooms`

## Does not own

Room cleanup does not own:

* WebSocket upgrade, read, write, or close behavior.
* WebSocket session state.
* Requested leave packet validation.
* Disconnect detection.
* Room snapshot broadcasting after a member leaves.
* Match result resolution.
* Match result reporting before room exit.
* Lobby join, ready, owner, or start rules.
* Return-to-lobby rules.
* Gameplay simulation mechanics.
* Player death, respawn, score, lives, or match-over policy.
* Process signal handling.
* HTTP graceful shutdown.
* Durable player-data persistence.
* Client-side cleanup or scene removal.

Networking may trigger cleanup by calling room leave APIs, but rooms own the cleanup policy.

Process shutdown may call `RoomManager.StopAll()`, but process shutdown does not own empty-room cleanup rules.

## Domain roles

Room cleanup participates in the game server technical runtime lifecycle.

Its role is to keep the in-memory room manager from retaining unused room state after all members and active players have left.

The cleanup boundary affects:

```text
room manager memory
room game simulation lifetime
room lookup results
future join attempts for removed room codes
room lifecycle diagnostics
```

It does not produce a player-facing transition. A room that has been cleaned up is simply no longer found by the room manager. Later attempts to join the cleaned room code fail as missing-room behavior.

## Protocols and APIs

Room cleanup has no direct external API.

Clients do not send a cleanup packet. Cleanup is reached indirectly through room exit flows and process/test teardown flows.

Relevant client-visible packet routes that can lead to cleanup are:

```text
leave_room_request
WebSocket disconnect
```

Requested leave routes through networking to:

```go
session.leaveRequestedRoom()
```

Disconnect routes through networking to:

```go
session.leaveDisconnectedRoom()
```

Both paths call:

```go
RoomManager.LeaveMember(roomID, sessionID, playerID)
```

`LeaveMember` removes room membership and active gameplay participation, then schedules cleanup when the room is empty.

The room cleanup runtime surface is internal Go code:

```go
RoomManager.ScheduleCleanupIfEmpty(roomID)
RoomManager.StopAll()
Room.ScheduleCleanupTimer(cleanupDelay, cleanupCallback)
Room.StopCleanupTimer()
Room.ShouldCleanup()
Room.CleanupVersionMatches(cleanupVersion)
Room.StopGameIfPresent()
```

## Cleanup eligibility

A room is cleanup-eligible only when it is non-nil and empty.

Current room-level check:

```text
Room.ShouldCleanup()
  room != nil
  room.IsEmpty()
```

Current emptiness check:

```text
Room.IsEmpty()
  active player count == 0
  member count == 0
```

Cleanup eligibility is not currently gated by room state.

Empty rooms are cleanup-eligible across these lifecycle states:

```text
Lobby
InGame
GameOver
```

Non-empty rooms are rejected when either condition is true:

```text
member count > 0
active player count > 0
```

This means an empty lobby can be removed, an empty in-game room can be removed, and an empty game-over room can be removed. A room with remaining lobby members or active gameplay players is preserved.

## Leave and disconnect flow

Room cleanup is normally reached after a member leaves a room.

Current requested-leave flow:

```text
leave_room_request
-> session.leaveRequestedRoom()
-> report resolved match result before room exit, if needed
-> rooms.LeaveMember(roomID, sessionID, currentGamePlayerID)
-> detach room session from room-session registry
-> clear session room fields
-> broadcast room snapshot if members remain
```

Current disconnect flow:

```text
WebSocket runtime exits
-> session.leaveDisconnectedRoom()
-> report resolved match result before room exit, if needed
-> rooms.LeaveMember(roomID, sessionID, currentGamePlayerID)
-> detach room session from room-session registry
-> clear session room fields
-> broadcast room snapshot if members remain
```

The rooms package cleanup step is inside `LeaveMember`:

```text
LeaveMember()
  LeaveRoom()
  remove active game player, when playerID and game exist
  decrement room active-player count when needed
  check room.ShouldCleanup()
  ScheduleCleanupIfEmpty(roomID)
  return LeaveMemberResult
```

`LeaveMemberResult.CleanupScheduled` reports whether the room was empty at leave time.

`LeaveMemberResult.ShouldBroadcastSnapshot` is true only when members remain. Empty-room cleanup does not broadcast a snapshot because there are no remaining room members to receive it.

## Cleanup scheduling

The default cleanup delay is:

```go
RoomCleanupGraceTime = 30 * time.Second
```

The room manager stores the delay as:

```go
RoomManager.cleanupDelay
```

Production construction uses:

```go
NewRoomManager()
NewRoomManagerWithCleanupDelay(RoomCleanupGraceTime)
```

Tests can use:

```go
NewRoomManagerWithCleanupDelay(...)
```

to shorten cleanup timing.

Scheduling flow:

```text
RoomManager.ScheduleCleanupIfEmpty(roomID)
  lock manager
  find room
  return if room missing
  return if room.ShouldCleanup() is false
  scheduleCleanupLocked(roomID, room)
```

`scheduleCleanupLocked` delegates timer ownership to the room:

```text
Room.ScheduleCleanupTimer(cleanupDelay, callback)
  lock room
  increment cleanup version
  stop existing cleanup timer, if present
  install time.AfterFunc(cleanupDelay, callback)
  return cleanup version
```

The callback later calls back into the room manager with the room ID and cleanup version captured at schedule time.

## Cleanup versioning

Cleanup versioning prevents stale timer callbacks from deleting rooms after a newer cleanup schedule has replaced them.

Each room owns:

```go
type roomCleanup struct {
    timer   *time.Timer
    version int
}
```

When cleanup is scheduled:

```text
room.cleanup.version increments
new timer captures that version
previous timer is stopped, if present
```

When cleanup runs, the manager checks:

```text
room.CleanupVersionMatches(cleanupVersion)
```

If the callback version does not match the current room cleanup version, cleanup is skipped as stale.

A cleanup version mismatch can happen when cleanup is scheduled more than once for the same room and an older callback still reaches the manager.

## Cleanup callback checks

When the cleanup timer fires, `RoomManager.cleanupEmptyRoom()` performs defensive checks before deleting the room.

Current cleanup callback flow:

```text
cleanupEmptyRoom(roomID, cleanupVersion)
  lock manager
  find room
  if missing:
    log skipped; room already removed
    return

  read active player count
  if active players > 0:
    log skipped; room active
    return

  if room.ShouldCleanup() is false:
    log skipped; room has members
    return

  if cleanup version is stale:
    log skipped; stale cleanup
    return

  room.StopGameIfPresent()
  delete room from manager map
  log room cleaned up
```

Deletion happens only after all checks pass.

The callback does not set `RoomStateClosed`. Current cleanup removes the room from the manager map instead of transitioning room state.

## Game stop behavior

Cleanup stops a room game instance before deleting the room.

Current cleanup deletion step:

```text
room.StopGameIfPresent()
delete(manager.rooms, roomID)
```

`StopGameIfPresent()` reads the room's current game instance under the room lock, then calls `Game.Stop()` outside that lock when a game exists.

`Game.Stop()` is the simulation stop primitive. It closes the simulation stop channel once and allows the simulation goroutine to return when it observes the stop signal.

Cleanup does not wait for a simulation goroutine acknowledgment before deleting the room from the manager map.

Return-to-lobby also stops and clears a game instance, but that is match lifecycle behavior, not room cleanup.

## Room-manager teardown

`RoomManager.StopAll()` is the broad teardown primitive for all rooms known to the manager.

Current `StopAll()` flow:

```text
RoomManager.StopAll()
  lock manager
  for each room:
    room.StopCleanupTimer()
    log room stopped
    if room has game instance:
      game.Stop()
    delete room from manager map
```

`StopAll()` is used by:

```text
services/game-server/cmd/game-server/main.go
tests that construct room managers
```

The executable registers:

```go
defer rooms.StopAll()
```

after constructing the room manager.

`StopAll()` differs from delayed empty-room cleanup:

```text
empty-room cleanup:
  applies to one room
  requires no members and no active players
  runs after cleanup delay
  checks cleanup version

StopAll:
  applies to every room
  does not require rooms to be empty
  stops cleanup timers
  stops game instances
  removes all rooms immediately
```

`StopAll()` does not run requested-leave behavior, does not detach WebSocket sessions, does not broadcast snapshots, and does not report match results.

## Observability

Room cleanup logs through:

```go
logging.Rooms
```

Current room cleanup lifecycle events include:

```text
room cleanup scheduled
room cleanup skipped; room already removed
room cleanup skipped; room active
room cleanup skipped; room has members
room cleanup skipped; stale cleanup
room cleaned up
room stopped
```

Common cleanup diagnostic fields include:

```text
room_id
cleanup_delay
cleanup_version
current_cleanup_version
active_players
members
```

Room cleanup logs are lifecycle diagnostics. They should stay focused on scheduling, skipping, deleting, and teardown state.

Broader logging policy belongs to game-server observability documentation.

## Data ownership

Room cleanup owns no durable data.

It mutates in-memory runtime state:

```text
RoomManager.rooms
Room.cleanup.timer
Room.cleanup.version
Game stop channel, through Game.Stop()
```

It does not persist state to player-data storage, Rails auth storage, local profile storage, match history, or telemetry storage.

Room cleanup may affect later room lookup results because a cleaned room code is removed from the manager map.

## Code map

Primary implementation files:

```text
services/game-server/internal/rooms/manager.go
services/game-server/internal/rooms/room_cleanup.go
services/game-server/internal/rooms/room_cleanup_state.go
services/game-server/internal/rooms/leave.go
services/game-server/internal/rooms/room_members.go
services/game-server/internal/rooms/room.go
services/game-server/internal/rooms/constants.go
```

Related room lifecycle files:

```text
services/game-server/internal/rooms/lifecycle.go
services/game-server/internal/rooms/room_lifecycle.go
services/game-server/internal/rooms/room_match.go
services/game-server/internal/rooms/room_match_access.go
```

Networking call sites:

```text
services/game-server/internal/networking/websocket.go
services/game-server/internal/networking/room_handlers.go
services/game-server/internal/networking/room_sessions.go
services/game-server/internal/networking/rooms.go
```

Process and simulation call sites:

```text
services/game-server/cmd/game-server/main.go
services/game-server/internal/game/game.go
services/game-server/internal/game/simulation.go
```

Related tests:

```text
services/game-server/internal/rooms/room_ownership_test.go
services/game-server/tests/rooms/room_test.go
services/game-server/tests/rooms/manager_test.go
services/game-server/tests/networking/rooms_test.go
```

Important non-ownership boundaries:

```text
services/game-server/internal/networking/
services/game-server/internal/game/
services/game-server/internal/game/rules/
services/game-server/internal/matchreporting/
services/player-data/
client/
```

`internal/networking` owns WebSocket session exit, room-session registry detachment, packet handling, and snapshot broadcasts.

`internal/game` owns authoritative simulation and the `Game.Stop()` primitive.

`internal/game/rules` owns match/mode decision policy.

`internal/matchreporting` and player-data integration own match result reporting.

`services/player-data` owns player-data storage behavior.

`client/` owns client-side scene, world, and UI cleanup.

## Tests and verification

Current cleanup coverage is split across room, room-manager, and networking tests.

Relevant room tests verify:

```text
Room.ShouldCleanup() accepts empty Lobby, InGame, and GameOver rooms
Room.ShouldCleanup() rejects rooms with members
Room.ShouldCleanup() rejects rooms with active players
nil room does not cleanup
cleanup versions increment and match expected values
```

Relevant room-manager tests verify:

```text
LeaveMember removes room members
LeaveMember removes active game players when a game player ID is supplied
LeaveMember schedules cleanup when the room becomes empty
LeaveMember does not schedule cleanup while a member remains
LeaveMember result reports whether cleanup was scheduled
```

Relevant networking tests verify:

```text
leave_room_request removes a member and broadcasts to remaining members
leave_room_request schedules empty-room cleanup
joining a cleaned room fails with room_not_found
leaving clears session room association
disconnect leaves lobby membership and broadcasts when members remain
```

Useful verification command:

```text
cd services/game-server && go test -buildvcs=false ./internal/rooms ./tests/rooms ./tests/networking
```

A broader game-server verification command is:

```text
cd services/game-server && go test -buildvcs=false ./...
```

## Related docs

* [Game Server Rooms](./!README.md)
* [Game Server](../!README.md)
* [WebSocket Session Lifecycle](../networking/websocket-session-lifecycle.md)
* [Room Network Adapter](../networking/room-network-adapter.md)
* [Game Server Process](../process/!README.md)
* [Service Shutdown](../process/service-shutdown.md)
* [Game Server Observability](../observability/!README.md)
* [Logging And Diagnostics](../observability/logging-and-diagnostics.md)
* [Game Server Simulation](../simulation/!README.md)
* [Match Result Reporting](../integrations/match-result-reporting.md)

## Notes

Legacy architecture notes correctly identified `services/game-server/internal/rooms` as owning room lifecycle and cleanup policy, with `roomCleanup` owning the cleanup timer and cleanup version. The current code keeps that ownership split.

Legacy logging notes described cleanup scheduled, skipped, and completed diagnostics. Current observability documentation now owns the broader logging policy; this document only describes the cleanup events relevant to room cleanup behavior.

This file currently lives in a `stubs/` folder. It should become canonical only when moved into the owning rooms folder and indexed as a direct room service document.
