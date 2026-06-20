# Room Network Adapter

Parent index: [Game Server Networking](./!README.md)

## Purpose

This document describes the game-server networking adapter that turns room/lobby WebSocket packets into room-domain operations and turns room-domain results back into room snapshot or room error packets.

## Overview

The room network adapter lives at the boundary between the WebSocket session and `services/game-server/internal/rooms`.

The client sends generated lobby packets over `/ws`. The networking inbound router classifies those packets, calls lobby-specific adapter methods, and the session handler delegates room decisions to the room manager or room aggregate. Networking keeps the per-connection state that the room package cannot own: the WebSocket session ID, current room pointer, current room ID, current active game player ID, authenticated account identity, outbound message queue, and the mapping from room members to live WebSocket sessions.

Room ownership remains in `internal/rooms`. Networking adapts transport/session facts into room calls, attaches account or local-profile identity to members, activates or deactivates active game players when a room enters or leaves gameplay, and broadcasts generated `room_snapshot` packets after successful room changes.

The adapter is intentionally not a second room rules layer. Room joinability, readiness, owner/start rules, state transitions, cleanup scheduling, and match lifecycle decisions come from the room domain.

## Code root

`services/game-server/internal/networking/`

## Responsibilities

- Route generated room/lobby packets from the inbound router to WebSocket session handlers.
- Maintain WebSocket-session room state:
  - `currentRoomID`
  - `room`
  - `currentGamePlayerID`
  - authenticated account identity
  - outbound message queue
- Create, join, leave, ready, start, start-single-player, and return-to-lobby through `rooms.RoomManager`.
- Attach live WebSocket sessions to room members for later room snapshot broadcast.
- Attach authenticated account IDs to multiplayer room members when account identity exists.
- Attach local profile IDs to single-player room members when supplied by `StartSinglePlayerRequest`.
- Activate connected room members into active game players after a successful start.
- Deactivate active game player routing when returning to lobby.
- Encode and enqueue generated `RoomSnapshot` and `RoomError` packets.
- Report an already-resolved match result before requested leave or disconnect removes the room member.

## Does not own

- Room membership rules.
- Room join/start/ready policy.
- Room state transitions.
- Room cleanup policy.
- Game simulation rules.
- Gameplay packet handling after active game player routing is established.
- Packet schema source-of-truth definitions.
- Auth token verification internals.
- Player-data persistence or match-result sink internals.
- Client presentation of lobby, room, or match-result state.

## Domain roles

The adapter implements the game-server side of room session routing.

It participates in the player experience flow where a connected client becomes a room member, optionally becomes an active game player, receives room snapshots, and leaves or returns to lobby. The room package owns the authoritative room state. Networking owns the live connection/session bridge needed to deliver that state to connected clients.

Important identity boundaries:

- WebSocket connection does not imply room membership.
- Room membership does not imply active gameplay participation.
- `sessionID` is WebSocket/session identity.
- room member identity is owned by `internal/rooms`.
- player-facing `PlayerID` values are exposed in room snapshots.
- `currentGamePlayerID` is networking-owned active game routing state for the current WebSocket session.
- authenticated account identity is stored on the session and copied onto room members when relevant.
- local profile ID is copied onto the single-player room member when supplied.

## Protocols and APIs

The adapter consumes the generated lobby packet family over the realtime WebSocket protocol.

The client calls this surface by sending generated packet types through `/ws`. Networking owns transport, decode, routing, and outbound queueing. The room domain owns authority behind successful or failed room operations. Data crossing this boundary includes room codes, readiness values, local profile IDs, account identity already attached to the session, and generated room snapshot/error payloads.

Supported inbound room packets:

```text
create_room_request
join_room_request
leave_room_request
set_ready_request
start_game_request
start_single_player_request
return_to_lobby_request
```

Outbound room packets:

```text
room_snapshot
room_error
```

Request behavior:

- `create_room_request`
  - requires authenticated account admission.
  - rejects sessions already in a room.
  - creates a lobby room through `RoomManager.CreateLobbyRoom`.
  - adds the session as a room member.
  - attaches account identity if present.
  - stores the room on the session.
  - sends a room snapshot only to the creating session.

- `join_room_request`
  - requires authenticated account admission.
  - rejects sessions already in a room.
  - normalizes and validates the requested room code through `RoomManager.JoinRoom`.
  - attaches the WebSocket session to the room session registry.
  - attaches account identity if present.
  - broadcasts a room snapshot to connected room members.

- `leave_room_request`
  - reports a resolved match result once before removing the member when a resolved result exists.
  - removes the session/member through `RoomManager.LeaveMember`.
  - detaches the WebSocket session from the room session registry.
  - clears `room`, `currentRoomID`, and `currentGamePlayerID` on the WebSocket session.
  - broadcasts a room snapshot to remaining members when any remain.

- `set_ready_request`
  - requires the session to already be in a room.
  - delegates ready mutation to `RoomManager.SetReady`.
  - broadcasts a room snapshot after success.

- `start_game_request`
  - requires the session to already be in a room.
  - delegates start policy and room transition to `RoomManager.StartRoomGame`.
  - activates connected room members into active game players.
  - broadcasts a room snapshot after success.

- `start_single_player_request`
  - does not require authenticated account admission.
  - rejects sessions already in a room.
  - creates and starts a non-joinable single-player room through `RoomManager.CreateStartedSinglePlayerRoom`.
  - attaches the live WebSocket session.
  - stores `local_profile_id` on the room member when supplied.
  - activates the connected member into an active game player.
  - broadcasts a room snapshot.

- `return_to_lobby_request`
  - requires the session to already be in a room.
  - delegates reset policy to `RoomManager.ReturnRoomToLobby`.
  - clears active game player IDs from connected sessions.
  - resets room active-player count to zero.
  - broadcasts a room snapshot after success.

Admission behavior is intentionally narrow. Multiplayer room create/join requires an authenticated account. Single-player start remains available to guest/local-profile sessions.

## Data ownership

Networking owns transient connection/session data only.

Networking mutates:

- `webSocketSession.currentRoomID`
- `webSocketSession.room`
- `webSocketSession.currentGamePlayerID`
- `webSocketSession.identity`
- `roomSessions.byRoom`
- outbound queued WebSocket payloads

The room package owns durable room runtime state:

- room ID and room code normalization.
- room members.
- owner selection.
- readiness.
- joinability.
- room state.
- active player count.
- room game instance.
- cleanup scheduling.
- resolved match summary.

The packet schema source of truth is under `shared/packets/`. The generated Go packet structs used by this adapter are emitted to `services/game-server/internal/game/packets.go`.

## Code map

Primary implementation files:

- `services/game-server/internal/networking/client_packet_router.go` - Wires decoded client packets into inbound family handlers.
- `services/game-server/internal/networking/inbound/lobby.go` - Classifies room/lobby packet types and calls the lobby session interface.
- `services/game-server/internal/networking/inbound_adapter.go` - Adapts `webSocketSession` methods to inbound packet family interfaces.
- `services/game-server/internal/networking/room_handlers.go` - Handles create, join, leave, ready, start, single-player start, and return-to-lobby requests.
- `services/game-server/internal/networking/room_sessions.go` - Tracks live WebSocket sessions by room/member session ID for broadcast and activation.
- `services/game-server/internal/networking/room_snapshot.go` - Builds and broadcasts generated room snapshot packets.
- `services/game-server/internal/networking/room_error.go` - Encodes and enqueues generated room error packets.
- `services/game-server/internal/networking/player_activation.go` - Activates/deactivates active game player routing for connected room sessions.
- `services/game-server/internal/networking/websocket.go` - Creates sessions, handles connection lifetime, and performs requested/disconnected room exit cleanup.
- `services/game-server/internal/networking/session_admission.go` - Enforces authenticated-account admission for multiplayer room create/join.

Room-domain implementation files used by the adapter:

- `services/game-server/internal/rooms/manager.go` - Room manager creation, lookup, create, join, ready, leave, and cleanup entry points.
- `services/game-server/internal/rooms/lifecycle.go` - Room manager start, single-player start, and return-to-lobby entry points.
- `services/game-server/internal/rooms/leave.go` - Room member leave and active-player removal behavior.
- `services/game-server/internal/rooms/room_join.go` - Joinability validation and member join behavior.
- `services/game-server/internal/rooms/room_lobby.go` - Ready and start precondition policy.
- `services/game-server/internal/rooms/room_lifecycle.go` - Room state transitions, game instance lifecycle, game-over, and reset-to-lobby behavior.
- `services/game-server/internal/rooms/constants.go` - Room states, limits, and room error codes.

Packet source and generated output:

- `shared/packets/lobby.toml` - Lobby packet structs, packet type IDs, and packet builders.
- `shared/packets/outputs.toml` - Generated packet output configuration.
- `services/game-server/internal/game/packets.go` - Generated Go packet structs and packet type constants.

Important non-ownership boundaries:

- `services/game-server/internal/networking/inbound/gameplay.go` owns gameplay packet routing, not room membership routing.
- `services/game-server/internal/networking/inbound/auth.go` owns authenticate packet routing, not room create/join admission policy itself.
- `services/game-server/internal/protocol/packetcodec` owns JSON encode/decode helpers, not room semantics.
- `services/game-server/internal/matchreporting` owns match-result reporter construction and sink integration, not room exit sequencing.
- `services/game-server/internal/playerdata` owns player-data records and persistence-facing structures, not live room state.

## Tests

Relevant tests:

- `services/game-server/internal/networking/gameplay_packets_test.go`
  - Verifies `StartSinglePlayerRequest` stores the supplied local profile ID on the room member.
- `services/game-server/internal/networking/room_sessions_test.go`
  - Verifies account ID attachment when adding a session member.
- `services/game-server/internal/networking/player_activation_test.go`
  - Verifies activation rebinds member player ID, preserves account ID, updates owner ID, and populates match summary identity.
- `services/game-server/internal/networking/room_snapshot_test.go`
  - Verifies room snapshot match-result projection behavior.
- `services/game-server/internal/networking/room_error_test.go`
  - Verifies room errors enqueue generated outbound packets.
- `services/game-server/internal/networking/websocket_test.go`
  - Verifies resolved match results are reported before requested leave or disconnect removes the room member.
- `services/game-server/internal/rooms/manager_test.go`
  - Verifies room manager behavior.
- `services/game-server/internal/rooms/room_join_test.go`
  - Verifies room join behavior.
- `services/game-server/internal/rooms/room_lobby_test.go`
  - Verifies lobby readiness/start behavior.
- `services/game-server/internal/rooms/room_lifecycle_test.go`
  - Verifies room lifecycle transitions.
- `services/game-server/internal/rooms/room_members_test.go`
  - Verifies room member behavior.
- `services/game-server/internal/rooms/room_match_summary_test.go`
  - Verifies match summary behavior used by room snapshot/reporting flow.

Suggested verification command:

```text
go test -buildvcs=false ./services/game-server/internal/networking ./services/game-server/internal/rooms
```

## Related docs

- [Game Server Networking](./!README.md)
- [Game Server Rooms](../rooms/!README.md)
- [Game Server](../!README.md)
- [Protocol](../../../protocol/!README.md)
- [Data](../../../data/!README.md)
- [Services](../../../!README.md)

## Notes

WebSocket connection, room membership, and active game player routing are separate states, and the implementation reflects that boundary.
