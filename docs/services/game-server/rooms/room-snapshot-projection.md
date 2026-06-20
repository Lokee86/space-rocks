# Room Snapshot Projection

Parent index: [Game Server Rooms](./!README.md)

## Purpose

This document describes the current game-server room snapshot projection boundary.

It covers how room state, room membership, local player identity, ownership, capacity, and resolved match-result summaries are projected into generated `room_snapshot` packets.

## Overview

Room snapshot projection is the read-model adapter between authoritative room state and client-facing lobby/session presentation.

The authoritative room state lives in `services/game-server/internal/rooms/`. The packet projection and WebSocket enqueue/broadcast behavior currently live in `services/game-server/internal/networking/room_snapshot.go`.

Current projection flow:

```text
room state and membership
-> networking.BuildRoomSnapshot(room, localSessionID)
-> generated game.RoomSnapshot
-> packetcodec.Encode
-> per-session outbound queue
-> WebSocket write loop
-> client room/lobby/session handlers
```

Room snapshots are sent after successful room lifecycle changes such as create, join, ready, start, single-player start, return to lobby, leave, disconnect cleanup, and room game-over detection.

The snapshot is presentation-safe. It exposes player-facing room facts such as `player_id`, `local_player_id`, and `owner_id`. It intentionally does not expose server-internal room member identity, WebSocket session identity, account IDs, or local profile IDs.

A snapshot can also include a resolved match-result summary when the room has one. That summary is used by the client match-results presentation path, not by durable player-data persistence.

## Code root

Primary projection adapter:

```text
services/game-server/internal/networking/room_snapshot.go
```

Authoritative room-state inputs:

```text
services/game-server/internal/rooms/
```

Generated packet shape:

```text
services/game-server/internal/game/packets.go
```

Packet source of truth:

```text
shared/packets/lobby.toml
```

## Responsibilities

Room snapshot projection owns:

* Reading room member state through `room.MembersSnapshot()`.
* Sorting member snapshots by `SessionID` before projection.
* Projecting room members into generated `game.RoomMemberState` values.
* Exposing only player-facing member fields:

  * `player_id`
  * `ready`
  * `connected`
* Resolving `local_player_id` from the target WebSocket session ID.
* Projecting room-level fields:

  * `room_code`
  * `room_state`
  * `owner_id`
  * `max_players`
* Projecting resolved match-result summaries when the room has one.
* Encoding generated `RoomSnapshot` packets through `packetcodec`.
* Enqueuing snapshots onto the target WebSocket session.
* Broadcasting snapshots to currently attached room sessions.
* Keeping per-session snapshot output correct when `local_player_id` differs by receiver.

## Does not own

Room snapshot projection does not own:

* Room membership mutation.
* Room joinability rules.
* Lobby readiness rules.
* Owner selection rules.
* Room state transitions.
* Match lifecycle decisions.
* Game-over detection.
* Match-result summary resolution.
* Match-result persistence or player-data reporting.
* WebSocket upgrade behavior.
* WebSocket write-loop mechanics.
* Packet schema source-of-truth definitions.
* Client room, lobby, or match-results presentation.
* Durable account or local-profile identity storage.

## Domain roles

Room snapshot projection participates in the game-server room session flow.

Its role is to expose the current room read model to connected clients after authoritative room changes. The room aggregate remains the source of truth. Networking performs packet projection because it knows the receiving WebSocket session and therefore can compute receiver-local fields such as `local_player_id`.

Important identity boundaries:

* `PlayerID` is player-facing and appears in room snapshots.
* `OwnerID` is player-facing and appears in room snapshots.
* `local_player_id` is receiver-specific and is derived from the WebSocket session receiving the snapshot.
* `SessionID` is server-side WebSocket/session identity and is not exposed in snapshot members.
* `MemberID` is server-internal room membership identity and is not exposed in normal room snapshots.
* `AccountID` and `LocalProfileID` may be attached to room members for match reporting, but they are not exposed in room snapshot presentation payloads.
* `currentGamePlayerID` belongs to networking session routing and is not a room snapshot identity field.

## Protocols and APIs

Room snapshot projection produces the generated realtime WebSocket packet type:

```text
room_snapshot
```

The packet is consumed by the Godot client after WebSocket decode and server-packet routing. The server owns authority behind the packet contents. The client treats the packet as authoritative room readback for lobby state, room state transitions, local player identity, owner gating, and match-over result presentation.

The data crossing this boundary is room presentation state:

```text
type
room_code
room_state
members[]
local_player_id
owner_id
max_players
match_result
```

The packet does not carry gameplay simulation state. Gameplay state is projected separately through the ticked `state` packet.

### Snapshot fields

Current generated Go shape:

```text
RoomSnapshot
  type
  room_code
  room_state
  members
  local_player_id
  owner_id
  max_players
  match_result
```

Current member shape:

```text
RoomMemberState
  player_id
  ready
  connected
```

Current match-result presentation shape:

```text
RoomMatchResultSummary
  match_id
  mode
  players[]

RoomPlayerMatchSummary
  game_player_id
  score
  ship_deaths
  won
```

### Build behavior

`BuildRoomSnapshot(room, localSessionID)`:

1. Reads the room member snapshot from `room.MembersSnapshot()`.
2. Sorts members by `SessionID`.
3. Copies each member into `game.RoomMemberState`.
4. Resolves `local_player_id` through `room.PlayerIDForSession(localSessionID)`.
5. Reads `owner_id` through `room.OwnerID()`.
6. Uses `rooms.MaxPlayersPerRoom` for `max_players`.
7. Reads the resolved match summary through `room.ResolvedMatchSummary()`.
8. Returns a generated `game.RoomSnapshot`.

The member projection deliberately copies only `PlayerID`, `Ready`, and `Connected`.

### Match-result projection

`buildRoomMatchResultSummary(room)` returns an empty generated `RoomMatchResultSummary` when:

* the room is nil
* the room has no resolved match summary

When a resolved summary exists, the projection copies:

* `match_id`
* `mode`
* each player `game_player_id`
* each player `score`
* each player `ship_deaths`
* each player `won`

The room's resolved match summary may include persistence-facing identity such as account ID or local profile ID. Snapshot projection strips those fields. Client match-results presentation receives only the presentation-safe summary.

### Per-session enqueue behavior

`EnqueueRoomSnapshot(room)` builds a snapshot for one WebSocket session and encodes it through `packetcodec`.

Because `local_player_id` depends on the receiving session, each session must build its own snapshot payload. Broadcast cannot safely reuse one encoded snapshot for every receiver.

Encode failures are logged through the network logger and the snapshot is dropped for that session.

### Broadcast behavior

`BroadcastRoomSnapshot(room)`:

1. Reads and sorts room members by `SessionID`.
2. Builds a list of member session IDs.
3. Resolves currently attached WebSocket sessions through `snapshotRoomSessions`.
4. Calls `session.EnqueueRoomSnapshot(room)` for each attached session.

Broadcast only targets attached live WebSocket sessions. It does not write directly to sockets and does not create a durable delivery queue.

## Data ownership

Room snapshot projection reads room data but does not own room data.

### Room-owned data

`services/game-server/internal/rooms/` owns:

* room ID
* room state
* room joinability
* room members
* member ready state
* member connected state
* owner ID
* player ID assignment
* active player count
* game instance reference
* current match ID
* resolved match summary
* cleanup state

### Networking-owned data

`services/game-server/internal/networking/` owns:

* live WebSocket sessions
* per-session outbound queues
* current room pointer on each session
* current room ID on each session
* current active game player ID on each session
* room-to-session attachment registry
* room snapshot packet projection and enqueueing

### Packet-owned data

The `room_snapshot` packet schema is sourced from:

```text
shared/packets/lobby.toml
```

Generated server packet structs are emitted to:

```text
services/game-server/internal/game/packets.go
```

The generated struct is the code-level packet shape, not the source of truth.

### Player-data-owned data

Player-data owns durable match-result storage and profile/stat mutation.

Room snapshot match-result projection is presentation-only. Durable reporting uses the resolved room summary through the match-result reporting integration, not through the `room_snapshot` packet.

## Projection model

Room snapshot projection is intentionally shallow. It should copy current room facts into a client-facing packet shape without adding room rules or client presentation logic.

Correct projection behavior:

```text
room member -> RoomMemberState
room state -> room_state
room owner -> owner_id
receiver session -> local_player_id
resolved summary -> presentation-safe match_result
```

Incorrect projection behavior:

```text
decide if a room can start
decide if a room can be joined
mutate readiness
mutate owner
recalculate game-over
persist match results
expose account/local-profile IDs
expose MemberID or SessionID
derive client UI routes
```

## Code map

Primary implementation files:

* `services/game-server/internal/networking/room_snapshot.go` - Builds, encodes, enqueues, and broadcasts room snapshots.
* `services/game-server/internal/networking/room_sessions.go` - Tracks attached WebSocket sessions by room and session ID for broadcast.
* `services/game-server/internal/networking/room_handlers.go` - Calls room operations and broadcasts snapshots after successful room changes.
* `services/game-server/internal/networking/websocket_gameplay_tick.go` - Broadcasts snapshots when the room game-over lifecycle advances.
* `services/game-server/internal/networking/websocket.go` - Handles requested/disconnected room exit and snapshot broadcast to remaining members.

Room input files:

* `services/game-server/internal/rooms/room.go` - Defines the room aggregate.
* `services/game-server/internal/rooms/member.go` - Defines room member fields.
* `services/game-server/internal/rooms/room_members.go` - Exposes member snapshots, owner ID, and player ID lookup.
* `services/game-server/internal/rooms/room_membership.go` - Owns room membership storage and member snapshot copying.
* `services/game-server/internal/rooms/room_match.go` - Stores current match ID, resolved summary, and reported state.
* `services/game-server/internal/rooms/room_match_access.go` - Exposes resolved match summary access.
* `services/game-server/internal/rooms/room_match_summary.go` - Builds resolved match summaries from game/player facts.
* `services/game-server/internal/rooms/room_lifecycle.go` - Builds the resolved summary when entering game over and clears it when a new match begins.
* `services/game-server/internal/rooms/lifecycle_tick.go` - Advances game-over lifecycle and triggers snapshot broadcast.

Packet and codec files:

* `shared/packets/lobby.toml` - Source of truth for room snapshot packet structs.
* `shared/packets/outputs.toml` - Packet generation output configuration.
* `services/game-server/internal/game/packets.go` - Generated Go packet structs and packet type constants.
* `services/game-server/internal/protocol/packetcodec/codec.go` - JSON packet encode/decode helper.

Related client consumers:

* `client/scripts/networking/inbound/server_packet_dispatcher.gd` - Emits `room_snapshot_received`.
* `client/scripts/session/session_network_controller.gd` - Routes room snapshots into the room session controller.
* `client/scripts/session/room_session_controller.gd` - Applies snapshots, caches room state, and caches match-result data.
* `client/scripts/lobby/lobby_flow.gd` - Applies room snapshot values to lobby session state.
* `client/scripts/lobby/lobby_packet_reader.gd` - Reads room snapshot fields.
* `client/scripts/lobby/lobby_session_state.gd` - Stores client-side lobby room state.
* `client/scripts/gameplay/match_end/match_end_flow.gd` - Uses cached match-result rows for match-over presentation.

Important non-ownership boundaries:

* `services/game-server/internal/rooms/` owns authoritative room state and match summaries.
* `services/game-server/internal/networking/` owns WebSocket session state, packet projection, and outbound queueing.
* `services/game-server/internal/matchreporting/` owns durable match-result reporting into player-data.
* `services/player-data/` owns player-data runtime and persistence routing.
* `client/scripts/lobby/` owns client lobby presentation state.
* `client/scripts/gameplay/match_end/` owns client match-end presentation orchestration.

## Tests

Primary tests:

* `services/game-server/internal/networking/room_snapshot_test.go`

  * Verifies empty match-result projection when no resolved summary exists.
  * Verifies resolved match-result projection includes match ID, mode, score, ship deaths, and win flag.

* `services/game-server/tests/networking/room_snapshot_test.go`

  * Verifies room code, room state, capacity, local player ID, owner ID, member list, ready state, and connected state projection.

Related tests:

* `services/game-server/internal/rooms/room_match_summary_test.go`

  * Verifies resolved match-summary behavior that snapshot projection reads.

* `services/game-server/internal/rooms/room_lifecycle_test.go`

  * Verifies game-over and return-to-lobby lifecycle behavior.

* `services/game-server/internal/rooms/lifecycle_tick_test.go`

  * Verifies room game-over lifecycle ticking and snapshot broadcast callback behavior.

* `services/game-server/internal/networking/room_sessions_test.go`

  * Verifies room session attachment behavior used by snapshot broadcast.

* `services/game-server/internal/networking/player_activation_test.go`

  * Verifies active game player routing and identity preservation that affects later snapshot/reporting behavior.

Suggested verification command:

```text
go test -buildvcs=false ./services/game-server/internal/networking ./services/game-server/internal/rooms ./services/game-server/tests/networking
```

## Related docs

* [Game Server Rooms](./!README.md)
* [Game Server](../!README.md)
* [Game Server Networking](../networking/!README.md)
* [Room Network Adapter](../networking/room-network-adapter.md)
* [Outbound Packet Routing](../networking/outbound-message-flow.md)
* [Match Result Reporting](../integrations/match-result-reporting.md)
* [Player Data](../../player-data/!README.md)
* [Protocol](../../../protocol/!README.md)
* [Data](../../../data/!README.md)

## Notes

Legacy documentation supplied two useful current facts for this boundary: room match-result presentation flows through `BuildRoomSnapshot`, and server-internal `MemberID` should not be exposed in normal room snapshot packets.

A room returning to lobby does not itself clear the resolved match summary. Starting the next match through `BeginNextMatch` clears the previous resolved summary. Snapshot projection therefore includes a match result whenever the room currently has a resolved summary, not only while the room state is `game_over`.

`BroadcastRoomSnapshot` builds one snapshot per receiving session because `local_player_id` is receiver-specific.
