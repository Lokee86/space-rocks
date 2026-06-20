# Service Shutdown

Parent index: [Game Server Process](./!INDEX.md)

## Purpose

This document describes the game-server process shutdown boundary.

It covers the currently implemented process-level cleanup hook, the `RoomManager.StopAll()` behavior it delegates to, and the shutdown responsibilities that belong to rooms, networking, simulation, or host process management instead.

## Overview

The game-server executable starts from `services/game-server/cmd/game-server/main.go`.

During startup, `main()`:

1. configures logging
2. creates an HTTP mux
3. constructs a room manager through `networking.NewRoomManager()`
4. registers `defer rooms.StopAll()`
5. builds player-data runtime dependencies
6. builds the match result reporter
7. builds the auth verifier
8. registers health, WebSocket, and player-data HTTP routes
9. blocks in `http.ListenAndServe(":8080", mux)`

The current executable does not install OS signal handling and does not create an `http.Server` value with a graceful `Shutdown` call. The only process-level cleanup hook currently present in `main()` is the deferred `rooms.StopAll()` call.

Because startup and server error paths call `os.Exit(1)`, those paths bypass deferred cleanup. As implemented, `RoomManager.StopAll()` is the cleanup primitive for room-owned runtime state, but the executable does not yet provide a complete graceful shutdown sequence for HTTP draining, WebSocket closure, signal handling, or dependency teardown.

## Code root

`services/game-server/cmd/game-server/`

Primary shutdown behavior delegates into:

`services/game-server/internal/rooms/`

## Responsibilities

This boundary owns:

* installing the process-level room cleanup hook in the executable entrypoint
* invoking `RoomManager.StopAll()` when the process returns through normal defer execution
* keeping process shutdown separate from room gameplay rules
* keeping process shutdown separate from WebSocket session-specific disconnect handling
* documenting what the executable currently does and does not shut down explicitly

`RoomManager.StopAll()` owns the in-memory room cleanup primitive used by process shutdown and tests.

It:

* locks the room manager
* iterates all known rooms
* stops each room cleanup timer
* logs that the room was stopped
* stops each room game instance when one exists
* deletes each room from the manager map

`Game.Stop()` owns the simulation stop primitive.

It:

* closes the game simulation stop channel exactly once
* allows the simulation goroutine to return on its next stop-channel select
* relies on `sync.Once` to make repeated stop calls safe

## Does not own

This document and process boundary do not own:

* OS signal handling.
* HTTP server graceful shutdown.
* WebSocket close-frame delivery to clients.
* Enumerating and closing active WebSocket sessions.
* Room membership leave rules.
* Empty-room cleanup scheduling after disconnect.
* Match result resolution.
* Match result reporting during process termination.
* Simulation mechanics.
* Player-data persistence internals.
* Auth verifier internals.
* Logging policy beyond process-level shutdown observations.

Room-level cleanup belongs under [Game Server Rooms](../rooms/!INDEX.md).

WebSocket disconnect behavior belongs under [Game Server Networking](../networking/!INDEX.md).

## Domain roles

Process shutdown participates in the technical runtime lifecycle of the game server.

Its domain role is narrow:

* keep room-owned in-memory runtime state from surviving a normal executable return
* stop active game simulations through the room manager
* cancel pending room cleanup timers during room-manager teardown
* avoid encoding gameplay state transitions into process shutdown

Process shutdown is not a player-facing flow. It does not decide match outcomes, return rooms to lobby, award results, or broadcast final snapshots.

## Protocols and APIs

There is no public shutdown API.

Shutdown is not exposed through:

* `GET /health`
* `GET /ws`
* player-data HTTP routes
* lobby packets
* gameplay packets
* devtools packets

The shutdown surface is internal process control. It is for the executable and host process environment, not for clients.

Clients leave gameplay through WebSocket disconnects or explicit room-leave packets. Those flows are handled by networking session logic, not by process shutdown. WebSocket session teardown closes the connection, leaves the disconnected room, detaches the session, reports resolved match results before room exit when applicable, and broadcasts room snapshots when remaining members exist.

Process-level `RoomManager.StopAll()` does not perform that session flow. It only stops room cleanup timers, stops game simulations, and removes rooms from the manager.

## Shutdown flow

Current executable flow:

```text
main()
  configure logging
  create mux
  create room manager
  defer rooms.StopAll()
  build player-data runtime
  build reporter
  build auth verifier
  register routes
  http.ListenAndServe(":8080", mux)
```

Current room-manager stop flow:

```text
RoomManager.StopAll()
  lock manager
  for each room:
    room.StopCleanupTimer()
    log room stopped
    if room has game instance:
      game.Stop()
    delete room from manager
```

Current game stop flow:

```text
Game.Stop()
  close stopSimulation once

runSimulation()
  select:
    stopSimulation closed -> return
    ticker tick -> Step(delta)
```

## Data ownership

Process shutdown owns no durable data.

It mutates only in-memory runtime state:

* the room manager map
* room cleanup timer handles
* game simulation stop channels

It does not persist shutdown state, flush match results, write profile data, or close player-data stores.

Player-data runtime construction happens in the process entrypoint, but player-data storage ownership remains under the player-data service/runtime. Match result reporting is handled through the match reporting integration and room/session lifecycle, not through `RoomManager.StopAll()`.

## Code map

Primary implementation files:

* `services/game-server/cmd/game-server/main.go` - executable entrypoint, route registration, `defer rooms.StopAll()`, and `http.ListenAndServe`.
* `services/game-server/internal/networking/rooms.go` - networking-facing room manager constructor wrapper.
* `services/game-server/internal/rooms/manager.go` - `RoomManager`, room map ownership, `StopAll()`, and cleanup scheduling.
* `services/game-server/internal/rooms/room_cleanup.go` - room cleanup timer stop, cleanup scheduling, and `StopGameIfPresent()`.
* `services/game-server/internal/game/game.go` - game start/stop lifecycle and stop channel ownership.
* `services/game-server/internal/game/simulation.go` - simulation goroutine loop and stop-channel return behavior.

Related implementation files:

* `services/game-server/internal/networking/websocket.go` - connection-scope cleanup through `defer session.conn.Close()` and `defer session.leaveDisconnectedRoom()`.
* `services/game-server/internal/networking/websocket_read.go` - read-loop exit on WebSocket read error.
* `services/game-server/internal/networking/websocket_write.go` - write-loop exit on read/write close and ticker cleanup.
* `services/game-server/internal/networking/websocket_gameplay_tick.go` - session gameplay lifecycle ticker and done-channel exit.
* `services/game-server/internal/rooms/leave.go` - member leave flow and empty-room cleanup scheduling.
* `services/game-server/internal/rooms/lifecycle.go` - room start, single-player start, and return-to-lobby lifecycle operations.
* `services/game-server/internal/rooms/room_lifecycle.go` - room game start, game-over, reset-to-lobby, and game stop behavior.

Related tests:

* `services/game-server/tests/rooms/manager_test.go` - room manager lifecycle, leaving members, empty-room cleanup scheduling, return-to-lobby game stop behavior.
* `services/game-server/tests/rooms/room_test.go` - room cleanup eligibility and game stop assertions.
* `services/game-server/tests/networking/rooms_test.go` - WebSocket room behavior using `defer manager.StopAll()` for test cleanup.
* `services/game-server/tests/networking/auth_test.go` - authenticated WebSocket paths using manager cleanup.
* `services/game-server/tests/networking/auth_admission_test.go` - auth admission paths using manager cleanup.

Important non-ownership boundaries:

* `RoomManager.StopAll()` does not close WebSocket connections.
* `RoomManager.StopAll()` does not run room-leave packet behavior.
* `RoomManager.StopAll()` does not broadcast room snapshots.
* `RoomManager.StopAll()` does not report match results.
* `Game.Stop()` does not wait for the simulation goroutine to acknowledge shutdown.
* `main()` does not currently handle OS signals or call `http.Server.Shutdown`.

## Tests

There is no dedicated process-level graceful shutdown test.

Current verification is indirect:

* room manager tests use `defer manager.StopAll()` to prevent room and game simulation leakage during tests
* room tests verify that reset-to-lobby stops the old game instance before clearing it
* cleanup tests verify that empty rooms schedule cleanup and that cleanup eligibility is based on room membership and active players
* networking tests use `defer manager.StopAll()` while exercising WebSocket room creation, joining, single-player start, leave/disconnect, auth, and admission flows

The current test surface verifies the reusable room/game cleanup primitives. It does not verify a full executable shutdown path with signal handling, HTTP draining, WebSocket close delivery, or process dependency teardown.

## Related docs

* [Game Server Process](./!INDEX.md)
* [Game Server Rooms](../rooms/!INDEX.md)
* [Game Server Networking](../networking/!INDEX.md)
* [Game Server Observability](../observability/!INDEX.md)
* [Game Server Integrations](../integrations/!INDEX.md)
* [Protocol](../../../protocol/!INDEX.md)
* [Services](../../!INDEX.md)

## Notes

Legacy docs do not currently provide useful authoritative detail for game-server process shutdown. The relevant current behavior is in the executable entrypoint, room manager, room cleanup helpers, WebSocket session cleanup, and game simulation stop primitive.

`RoomStateClosed` exists as a room state constant, but current `RoomManager.StopAll()` does not transition rooms into that state. It stops timers, stops games, and removes rooms from the manager.

The current executable cleanup hook should not be described as graceful shutdown. It is a room-manager cleanup primitive registered with `defer`, while the process still lacks signal handling and HTTP server shutdown orchestration.
