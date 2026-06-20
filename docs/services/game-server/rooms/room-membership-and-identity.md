## Room Membership And Identity

Parent index: [Game Server Rooms](./!INDEX.md)

## Purpose

This document describes the game-server room membership and identity boundary.

It covers room members, player-facing room IDs, internal member IDs, room owner identity, readiness, connected state, account/local-profile attachment, and the bridge from room membership to active game player identity.

## Overview

Room membership is owned by `services/game-server/internal/rooms`.

A WebSocket session does not automatically mean room membership. A session becomes a room member only after a room create, join, or single-player start path adds a `RoomMember` to a `Room`.

Room membership also does not automatically mean active gameplay participation. Before a match starts, a member has lobby membership identity and readiness state. When a room successfully starts, networking activates connected room members by creating active game players in the room game instance and rebinding each connected member to the game player ID returned by `game.AddPlayer()`.

Current identity layers are:

```text
webSocketSession.sessionID
  connection/session identity owned by networking

RoomMember.MemberID
  server-internal room-member identity, generated as UUID v4

RoomMember.PlayerID
  room-facing player identity used for owner, ready, room snapshot members,
  local_player_id, and active game routing once activated

webSocketSession.currentGamePlayerID
  networking-owned active game routing value for the current session

RoomMember.AccountID
  authenticated-account attachment copied from networking session identity

RoomMember.LocalProfileID
  local-profile attachment supplied by single-player start
```

Room snapshots expose `player_id`, `ready`, `connected`, `local_player_id`, and `owner_id`. They do not expose `MemberID`, `SessionID`, `AccountID`, or `LocalProfileID`.

The membership boundary is intentionally narrow. It owns who is in the room and which player-facing identity represents each member. It does not own WebSocket transport, authentication verification, player-data persistence, or game simulation mechanics.

Related player boundaries:

* [Player Session State](../simulation/players/player-session-state.md)
* [Active Player Avatar State](../simulation/players/active-player-avatar-state.md)

Player session state and active avatar state are owned under the players simulation docs; room membership only bridges to them.

## Code root

`services/game-server/internal/rooms/`

Supporting boundaries:

* `services/game-server/internal/networking/`
* `services/game-server/internal/game/`
* `services/game-server/internal/playerids/`
* `services/game-server/internal/playerdata/`
* `shared/packets/`

## Responsibilities

Room membership owns:

* Creating `RoomMember` values for joining sessions.
* Generating internal `MemberID` values.
* Storing member `SessionID` values for room/session lookup.
* Assigning provisional room player IDs when members join.
* Storing members by current `PlayerID`.
* Looking up a member player ID from a WebSocket session ID.
* Rebinding a member's `PlayerID` when active game player activation creates a game player ID.
* Updating owner ID when the owner leaves or is rebound.
* Selecting the first member as the initial owner.
* Reassigning ownership when the current owner is removed.
* Tracking lobby readiness per member.
* Tracking connected state on the member record.
* Attaching authenticated-account IDs to members.
* Attaching local-profile IDs to members.
* Returning value-copy member snapshots for room snapshot projection and activation.
* Reporting member count and fullness through the room aggregate.
* Supporting match summary identity attachment by retaining account/local-profile IDs on the member.

## Does not own

Room membership does not own:

* WebSocket upgrade, read loops, write loops, or outbound queues.
* The live room-session registry used for broadcasting to connected WebSocket sessions.
* Authentication token verification.
* Multiplayer create/join admission policy.
* Local profile loading, profile creation, or profile persistence.
* Authenticated account persistence.
* Match-result persistence.
* Room code generation or normalization.
* Joinability rules beyond member-count facts supplied to room rules.
* Start-game decision rules beyond member facts supplied to room rules.
* Active game player creation.
* Game simulation, player spawning, lives, scoring, or death.
* Packet schema source-of-truth files.
* Client lobby or room presentation.

## Domain roles

Room membership participates in the player experience flow where a connected client becomes a room member, receives a room-facing identity, optionally becomes an active game player, and later leaves or returns to lobby.

The authoritative room membership state is inside the room aggregate. Networking adapts WebSocket/session facts into room calls, but the room package owns membership state once a member is added.

Important identity rules:

```text
connection does not imply room membership
room membership does not imply active game participation
active game participation begins only after successful start activation
MemberID is internal room-member identity
SessionID is networking/session identity stored on the room member for lookup
PlayerID is the room/player-facing identity used by snapshots and room rules
currentGamePlayerID is networking-owned active game routing state
AccountID and LocalProfileID are identity attachments, not room-auth authorities
```

Current implementation uses two player ID allocation sources:

* lobby/member join assigns readable provisional IDs through `playerids.Format`, currently `Player-1`, `Player-2`, and so on.
* active game activation calls `game.AddPlayer()`, currently producing game player IDs such as `player-1`, then `room.SetMemberPlayerIDForSession` rekeys the member to that game player ID.

Consumers should use the ID values supplied by the current packet or room API. They should not infer identity semantics from casing or assume a WebSocket session ID is a player ID.

## Protocols and APIs

Room membership is consumed through room-domain APIs and projected into generated lobby packets by networking.

The protocol surface exists so clients can create, join, observe, ready, start, and leave rooms over the realtime WebSocket connection. Clients consume generated `room_snapshot` packets. The room package owns authority for membership state; networking owns packet decode, handler routing, session fields, and outbound delivery. Data crossing the boundary includes session IDs from networking, room member facts, account/local-profile attachments, and snapshot-safe player-facing member state. The surface does not expose internal member IDs, account IDs, local profile IDs, or WebSocket session IDs to clients.

Room-domain membership APIs:

```text
Room.AddMember
Room.AddMemberSessionID
Room.PlayerIDForSession
Room.SetMemberAccountIDForSession
Room.SetMemberLocalProfileIDForSession
Room.SetMemberPlayerIDForSession
Room.OwnerID
Room.RemoveMember
Room.MemberCount
Room.IsFull
Room.IsEmpty
Room.MembersSnapshot
```

Manager APIs that consume membership identity:

```text
RoomManager.JoinRoom
RoomManager.LeaveRoom
RoomManager.LeaveMember
RoomManager.SetReady
RoomManager.StartRoomGame
RoomManager.CreateSinglePlayerRoom
RoomManager.CreateStartedSinglePlayerRoom
RoomManager.ReturnRoomToLobby
```

Generated inbound packet paths that can mutate or consume membership:

```text
create_room_request
join_room_request
leave_room_request
set_ready_request
start_game_request
start_single_player_request
return_to_lobby_request
```

Generated outbound snapshot fields derived from membership:

```text
RoomSnapshot.members[].player_id
RoomSnapshot.members[].ready
RoomSnapshot.members[].connected
RoomSnapshot.local_player_id
RoomSnapshot.owner_id
```

Membership identity is also used when building resolved match summaries. `room.buildMatchResultSummaryLocked` matches game player facts back to room members by `GamePlayerID` / member `PlayerID` and copies `AccountID` or `LocalProfileID` into the player-data match summary.

## Data ownership

Room membership data is transient runtime state. It is not persisted directly by the room package.

The room package owns these in-memory member fields:

```text
RoomMember.MemberID
RoomMember.SessionID
RoomMember.PlayerID
RoomMember.AccountID
RoomMember.LocalProfileID
RoomMember.Ready
RoomMember.Connected
```

The room membership aggregate owns:

```text
roomMembership.members
roomMembership.ownerID
```

Networking owns:

```text
webSocketSession.sessionID
webSocketSession.currentRoomID
webSocketSession.currentGamePlayerID
webSocketSession.identity
roomSessions.byRoom
```

The game package owns active game player/session state after activation:

```text
game.Game player IDs
game player sessions
game active ship state
game match facts
```

Player-data owns durable profile/account stat mutation. Room membership only retains `AccountID` or `LocalProfileID` long enough to connect a resolved match summary to the correct player-data identity.

Packet data ownership:

* `shared/packets/lobby.toml` owns lobby/room packet schema.
* Generated Go packet structs are emitted under `services/game-server/internal/game/packets.go`.
* Room membership supplies runtime facts to snapshot construction; it does not own packet generation.

## Code map

Primary implementation files:

* `services/game-server/internal/rooms/member.go` - Defines `RoomMember`, member construction, readiness mutation, and connected-state mutation.
* `services/game-server/internal/rooms/member_ids.go` - Generates UUID v4 member IDs.
* `services/game-server/internal/rooms/room_membership.go` - Owns the member map, owner ID, member add/remove, session lookup, player ID rebinding, snapshots, and player ID allocation.
* `services/game-server/internal/rooms/room_members.go` - Exposes room aggregate membership APIs with room locking.
* `services/game-server/internal/rooms/player_ids.go` - Adapts room provisional player ID formatting through the shared player ID formatter.
* `services/game-server/internal/playerids/playerids.go` - Defines shared player ID formatting and max player count.
* `services/game-server/internal/rooms/room.go` - Composes `roomMembership` into the `Room` aggregate.
* `services/game-server/internal/rooms/room_lobby.go` - Projects member readiness and connected facts into start-rule decisions.
* `services/game-server/internal/rooms/room_join.go` - Validates joinability and adds members on join.
* `services/game-server/internal/rooms/leave.go` - Removes room members and removes active game players when leaving.
* `services/game-server/internal/rooms/lifecycle.go` - Resolves session IDs into member player IDs for start and return-to-lobby manager calls.
* `services/game-server/internal/rooms/room_lifecycle.go` - Uses membership lookup for start, reset-to-lobby, ready reset, and game-over match summary resolution.
* `services/game-server/internal/rooms/room_match_summary.go` - Copies account/local-profile identity from room members into player-data match summaries.

Networking integration files:

* `services/game-server/internal/networking/room_handlers.go` - Handles room packets and attaches account/local-profile identity to members.
* `services/game-server/internal/networking/room_sessions.go` - Tracks live WebSocket sessions by room and session ID for broadcast and activation.
* `services/game-server/internal/networking/player_activation.go` - Activates connected members into game players and rebinds member `PlayerID`.
* `services/game-server/internal/networking/room_snapshot.go` - Projects safe room membership fields into generated room snapshots.
* `services/game-server/internal/networking/websocket.go` - Removes members on requested leave or disconnect.
* `services/game-server/internal/networking/session_identity.go` - Defines guest and authenticated-account session identity.

Related source/generated files:

* `shared/packets/lobby.toml` - Source of truth for room/lobby packet structs and packet types.
* `services/game-server/internal/game/packets.go` - Generated Go packet structs for room snapshots and room errors.
* `services/game-server/internal/playerdata/types.go` - Match summary identity shape consumed by match-result reporting.

Important non-ownership boundaries:

* `services/game-server/internal/networking/` owns WebSocket sessions, not room membership authority.
* `services/game-server/internal/game/` owns active game player state, not lobby membership.
* `services/game-server/internal/rooms/roomrules/` owns pure room rule decisions, not member storage.
* `services/game-server/internal/playerdata/` owns player-data contracts, not live room membership.
* `services/player-data/` owns profile/stat persistence, not room membership.
* `services/api-server/` owns authenticated-account auth and Rails/Postgres persistence, not room member state.

## Tests

Relevant room tests:

* `services/game-server/internal/rooms/room_members_test.go`

  * Verifies first member owner selection.
  * Verifies owner reassignment when the owner leaves.
  * Verifies non-owner removal preserves owner.
  * Verifies session-to-player lookup.
  * Verifies account ID attachment.
  * Verifies local profile ID attachment.
  * Verifies missing-session identity attachment fails.
  * Verifies member snapshots are value copies.
  * Verifies member count and fullness use membership state.

* `services/game-server/internal/rooms/player_ids_test.go`

  * Verifies provisional player ID formatting.
  * Verifies next available player ID selection.
  * Verifies member add assigns player IDs.
  * Verifies owner/start/reset lookups use player IDs.

* `services/game-server/internal/rooms/room_lobby_test.go`

  * Verifies connected unready members block start.
  * Verifies disconnected unready members do not block start.

* `services/game-server/internal/rooms/room_match_summary_test.go`

  * Verifies account/local-profile IDs are copied into resolved match summaries where applicable.

Relevant networking tests:

* `services/game-server/internal/networking/player_activation_test.go`

  * Verifies activation rebinds member player ID, preserves account ID, updates owner ID, and carries identity into match summary.

* `services/game-server/internal/networking/room_sessions_test.go`

  * Verifies account ID attachment during session member add.

* `services/game-server/internal/networking/gameplay_packets_test.go`

  * Verifies single-player start stores supplied local profile ID on the room member.

* `services/game-server/internal/networking/room_snapshot_test.go`

  * Verifies room snapshot projection behavior.

* `services/game-server/internal/networking/websocket_test.go`

  * Verifies room leave/disconnect behavior that removes members and reports resolved match results before removal.

Suggested verification command:

```text
go test -buildvcs=false ./services/game-server/internal/rooms ./services/game-server/internal/networking
```

## Related docs

* [Game Server Rooms](./!INDEX.md)
* [Game Server Networking](../networking/!INDEX.md)
* [Room Network Adapter](../networking/room-network-adapter.md)
* [WebSocket Session Lifecycle](../networking/websocket-session-lifecycle.md)
* [Game Server](../!INDEX.md)
* [Player Data](../../player-data/!INDEX.md)
* [API Server](../../api-server/!INDEX.md)
* [Protocol](../../../protocol/!INDEX.md)
* [Data](../../../data/!INDEX.md)

## Notes

Legacy room architecture notes correctly identified the important separation between WebSocket connection, room membership, and active game player routing. That separation is current.

Legacy notes also identified `MemberID` as internal membership identity and not normal room snapshot data. That remains current: `MemberID` exists on `RoomMember`, but snapshot projection omits it.

The current implementation uses readable string player IDs, not UUID player-facing IDs. Lobby membership initially uses provisional `Player-N` values from `playerids.Format`; active game activation currently rebinds connected members to game player IDs returned by `game.AddPlayer()`, such as `player-1`.
