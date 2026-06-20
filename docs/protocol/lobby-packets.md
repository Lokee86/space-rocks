## Lobby Packets

Parent index: [Protocol](./!README.md)

## Purpose

This document describes the current realtime lobby packet protocol between the Godot client and the Go game server.

It covers room creation, room join, room leave, readiness, multiplayer start, single-player start, return-to-lobby, room snapshots, room errors, adjacent websocket authentication packets, packet schema ownership, routing order, service responsibilities, and validation.

## Overview

Lobby packets are JSON packets sent over the realtime WebSocket connection.

The client sends lobby request packets to express user intent. The game server treats those packets as requests, validates them against authoritative room/session state, and responds by sending room snapshots or room errors.

Current protocol flow:

```text
client lobby or boot intent
-> generated client packet builder
-> ClientPacketSender
-> NetworkClient JSON/WebSocket send
-> game-server websocket read loop
-> packet envelope decode
-> inbound packet router
-> lobby or auth handler
-> room/networking service behavior
-> generated room_snapshot, authenticate_result, or room_error packet
-> packetcodec JSON encode
-> websocket outbound queue
-> client packet decode
-> server packet dispatcher
-> room/session/lobby presentation flow
```

The protocol is request/readback based. The client does not mutate room state locally as authority. It updates presentation from server-owned `room_snapshot` packets and treats `room_error` packets as rejected request readback.

Room snapshots are the primary room-state publication mechanism. They carry the current room code, room state, member list, local player identity, owner identity, capacity, and optional match-result summary. `room_state_changed` exists in the packet schema and client dispatcher, but the current server room adapter publishes authoritative room state through `room_snapshot`.

## Participating systems

```text
client/scripts/networking/
```

Owns client WebSocket transport, packet encode/decode, outbound packet sending, inbound packet classification, and connection-service signals.

```text
client/scripts/lobby/
```

Owns client lobby snapshot consumption, transient lobby read model, multiplayer lobby presentation coordination, and lobby UI request routing.

```text
client/scripts/session/
```

Owns room-session handling, boot request timing, websocket auth result handling, room packet routing into lobby/gameplay session flows, and cached match-result readback.

```text
services/game-server/internal/networking/
```

Owns WebSocket sessions, packet decode/routing, authentication request handling, lobby request adapters, session room state, outbound queues, room snapshot projection, room error packets, and active game routing after start.

```text
services/game-server/internal/rooms/
```

Owns authoritative room state, membership, owner selection, readiness, joinability, room lifecycle transitions, match summary resolution, and room-domain error codes.

```text
shared/packets/
```

Owns packet schema source files and output routing for generated Go and GDScript packet code.

```text
tools/data_sync/
```

Owns packet schema validation and generation into Go and GDScript outputs.

## Protocol authority

The packet schema owns:

```text
packet type strings
packet JSON field names
generated packet constants
generated Go packet structs
generated GDScript field constants
generated GDScript outbound builders
```

The game server owns authority for:

```text
whether a create/join/start/ready/leave/return request is accepted
room membership state
room owner identity
member ready state
member connected state
room state
room capacity
active game player activation after start
match result summary projection into room snapshots
room error codes and rejection messages
```

The client owns:

```text
local user intent collection
outbound request emission
transient lobby read model
lobby UI presentation
local owner/start-button presentation derivation
websocket connection lifecycle from the client side
```

The client does not own room authority. Client helpers such as `LobbySessionState.can_start_game()` are presentation helpers only; the server still validates any `start_game_request`.

## Packet families

### Client-to-server lobby requests

```text
create_room_request
join_room_request
leave_room_request
set_ready_request
start_game_request
start_single_player_request
return_to_lobby_request
```

### Adjacent websocket auth packets

```text
authenticate_request
authenticate_result
```

Auth packet definitions live in `shared/packets/lobby.toml`, and auth routing happens before lobby routing in the server inbound packet router.

### Server-to-client room packets

```text
room_snapshot
room_error
room_state_changed
```

`room_snapshot` and `room_error` are active server outputs in the current room adapter. `room_state_changed` is generated and client-routable, but current room-state publication is snapshot-based.

## Routing order

The game-server WebSocket read loop first decodes a minimal packet envelope with a `type` field.

After envelope handling, the server inbound router applies this family order:

```text
devtools envelope-routed packet families
full generated ClientPacket decode
auth packets
telemetry packets
lobby packets
gameplay packets
```

Lobby packet handling runs after auth and telemetry handling. This lets an authenticated connection establish websocket account identity before multiplayer room create or join requests are sent.

The client inbound dispatcher classifies server packets by generated type constants and emits signals for:

```text
authenticate_result
room_snapshot
room_state_changed
room_error
gameplay state
debug packets
player pause state
telemetry pong
unknown packet
```

Room-session code consumes room packets. Lobby presentation consumes `room_snapshot` after the room-session controller delegates it into the lobby flow.

## Request behavior

### `authenticate_request`

Purpose:

```text
client asks the game server to verify an existing bearer token for this websocket session
```

Request fields:

```text
type
token
```

Current server behavior:

```text
empty token -> authenticate_result authenticated=false error_code=invalid_token
missing verifier -> authenticate_result authenticated=false error_code=token_verification_unavailable
verification error -> authenticate_result authenticated=false error_code=token_verification_unavailable
invalid token -> authenticate_result authenticated=false error_code=invalid_token
valid token -> store authenticated account identity on session and send authenticated=true
```

Response packet:

```text
authenticate_result
```

Authentication establishes websocket session identity. It does not itself create or join a room.

### `create_room_request`

Purpose:

```text
client asks the server to create a multiplayer lobby room
```

Current request fields:

```text
type
```

The source struct contains `local_profile_id`, but the current generated client builder does not send it and the current server create-room handler does not consume it.

Current server behavior:

```text
requires authenticated account admission
rejects when auth verification is unavailable
rejects unauthenticated account sessions
rejects sessions already in a room
creates a joinable lobby room
adds the websocket session as a room member
stores the room on the websocket session
clears active game-player routing
sends a room_snapshot to the creating session
```

Primary rejection packet:

```text
room_error
```

### `join_room_request`

Purpose:

```text
client asks the server to join an existing multiplayer lobby room by room code
```

Request fields:

```text
type
room_code
```

Current server behavior:

```text
requires authenticated account admission
rejects sessions already in a room
normalizes and validates the room code through room manager behavior
rejects missing, malformed, full, closed, in-game, starting, or otherwise non-joinable rooms
adds the websocket session as a room member
attaches the live websocket session to the room-session registry
attaches authenticated account identity to the room member when present
stores the room on the websocket session
clears active game-player routing
broadcasts a room_snapshot to connected room members
```

Primary response packets:

```text
room_snapshot
room_error
```

### `leave_room_request`

Purpose:

```text
client asks to leave its current room
```

Request fields:

```text
type
```

Current server behavior:

```text
reports a resolved match result once before member removal when a resolved summary exists
rejects sessions that are not in a room
removes the member from room membership
removes active game player state when relevant
detaches the websocket session from the room-session registry
clears room and active-player routing fields on the websocket session
broadcasts a room_snapshot to remaining room members when any remain
```

A leaving client is locally routed back by client lobby return flow. Remaining members observe the updated room state through the broadcast snapshot.

### `set_ready_request`

Purpose:

```text
client asks to change its room-member ready state
```

Request fields:

```text
type
ready
```

Current server behavior:

```text
rejects sessions that are not in a room
delegates ready mutation to RoomManager.SetReady
accepts ready mutation only where room rules allow it
broadcasts a room_snapshot after success
```

Ready state is stored on the room member and returned as `members[].ready` in the next room snapshot.

### `start_game_request`

Purpose:

```text
client asks to start a multiplayer room
```

Request fields:

```text
type
```

Current server behavior:

```text
rejects sessions that are not in a room
delegates start policy and room transition to RoomManager.StartRoomGame
requires owner/start/readiness rules in the room domain
starts the game instance when accepted
activates connected room members as active game players
rebinds room member player IDs to active game player IDs
broadcasts a room_snapshot after success
```

Current room transition:

```text
Lobby -> Starting -> InGame
```

The current transition through `Starting` is immediate. The state still exists in the room model and generated constants.

### `start_single_player_request`

Purpose:

```text
client asks to create and immediately start a single-player room
```

Request fields:

```text
type
local_profile_id
```

Current server behavior:

```text
does not require authenticated account admission
rejects sessions already in a room
creates a non-joinable single-player room
adds the websocket session as the only room member
starts the game
stores local_profile_id on the room member when supplied
attaches the live websocket session to the room-session registry
activates the connected member as an active game player
broadcasts a room_snapshot
```

Single-player uses the room system but bypasses multiplayer create/join account admission and multiplayer owner/readiness start policy.

### `return_to_lobby_request`

Purpose:

```text
client asks to move a game-over room back to lobby
```

Request fields:

```text
type
```

Current server behavior:

```text
rejects sessions that are not in a room
delegates reset policy to RoomManager.ReturnRoomToLobby
accepts only when the room is in GameOver and the requester is still a room member
resets all member ready values to false
stops and clears the current game instance
sets room state to Lobby
deactivates active game player routing for connected sessions
broadcasts a room_snapshot after success
```

Return-to-lobby changes the room back to lobby state. It does not make the client the authority for replay or match result state.

## Response behavior

### `authenticate_result`

Purpose:

```text
server reports websocket auth verification result
```

Fields:

```text
type
authenticated
user_id
display_name
error_code
message
```

The client stores websocket auth state in `ClientConnectionService` and uses `SessionNetworkController` to decide when a pending multiplayer boot request can be sent.

### `room_snapshot`

Purpose:

```text
server publishes the authoritative room read model for one receiving websocket session
```

Fields:

```text
type
room_code
room_state
members
local_player_id
owner_id
max_players
match_result
```

Member fields:

```text
player_id
ready
connected
```

Match-result fields:

```text
match_result.match_id
match_result.mode
match_result.players[].game_player_id
match_result.players[].score
match_result.players[].ship_deaths
match_result.players[].won
```

Snapshot construction is receiver-specific because `local_player_id` depends on the receiving websocket session. Broadcast therefore builds and encodes one snapshot per live recipient instead of reusing one encoded payload for every client.

Room snapshots intentionally do not expose:

```text
RoomMember.MemberID
webSocketSession.sessionID
RoomMember.AccountID
RoomMember.LocalProfileID
currentGamePlayerID
Rails/API auth token
player-data persistence identity
```

The client uses `room_snapshot` to update room-session state, lobby presentation, owner/start presentation helpers, match-result cache, and gameplay-packet acceptance once the room reaches `InGame`.

### `room_error`

Purpose:

```text
server reports that a room/lobby request was rejected
```

Fields:

```text
type
error_code
message
```

Common current error codes include:

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
auth_unavailable
```

`auth_unavailable` is currently emitted by multiplayer admission when no websocket auth verifier is configured. `token_verification_unavailable` is used by auth result handling and also exists in room constants.

### `room_state_changed`

Purpose:

```text
generated room state notification packet
```

Fields:

```text
type
room_code
room_state
```

The client dispatcher and room-session controller can route this packet. Current server room publication uses `room_snapshot` for authoritative room-state readback.

## Packet shapes

Current source structs in `shared/packets/lobby.toml`:

```text
CreateRoomRequest
  type
  local_profile_id

JoinRoomRequest
  type
  room_code

LeaveRoomRequest
  type

SetReadyRequest
  type
  ready

StartGameRequest
  type

StartSinglePlayerRequest
  type
  local_profile_id

ReturnToLobbyRequest
  type

AuthenticateRequest
  type
  token

AuthenticateResult
  type
  authenticated
  user_id
  display_name
  error_code
  message

RoomMemberState
  player_id
  ready
  connected

LobbyMemberFieldAliases
  is_ready
  is_connected
  name
  member_name

RoomPlayerMatchSummary
  game_player_id
  score
  ship_deaths
  won

RoomMatchResultSummary
  match_id
  mode
  players[]

RoomSnapshot
  type
  room_code
  room_state
  members[]
  local_player_id
  owner_id
  max_players
  match_result

RoomStateChanged
  type
  room_code
  room_state

RoomError
  type
  error_code
  message
```

`LobbyMemberFieldAliases` contributes generated field constants used by client compatibility readers. It is not a separate active wire packet.

## Source-of-truth files

Primary packet schema source:

```text
shared/packets/lobby.toml
```

Packet output routing:

```text
shared/packets/outputs.toml
```

Generated client packet constants and builders:

```text
client/scripts/generated/networking/packets/packets.gd
```

Generated game-server packet structs and constants:

```text
services/game-server/internal/game/packets.go
```

Packet schema pipeline:

```text
tools/data_sync/
```

JSON encode/decode helpers:

```text
services/game-server/internal/protocol/packetcodec/codec.go
client/scripts/networking/packets/packet_codec.gd
```

## Generated outputs

The lobby packet family is generated into the `server_game_packets` and `client_packets` outputs.

Server output:

```text
services/game-server/internal/game/packets.go
```

Includes:

```text
PacketTypeCreateRoomRequest
PacketTypeJoinRoomRequest
PacketTypeLeaveRoomRequest
PacketTypeSetReadyRequest
PacketTypeStartGameRequest
PacketTypeStartSinglePlayerRequest
PacketTypeReturnToLobbyRequest
PacketTypeAuthenticateRequest
PacketTypeAuthenticateResult
PacketTypeRoomSnapshot
PacketTypeRoomStateChanged
PacketTypeRoomError

ClientPacket
CreateRoomRequest
JoinRoomRequest
LeaveRoomRequest
SetReadyRequest
StartGameRequest
StartSinglePlayerRequest
ReturnToLobbyRequest
AuthenticateRequest
AuthenticateResult
RoomMemberState
RoomPlayerMatchSummary
RoomMatchResultSummary
RoomSnapshot
RoomStateChanged
RoomError
```

Client output:

```text
client/scripts/generated/networking/packets/packets.gd
```

Includes packet type constants, field constants, and builders for:

```text
create_room_request_packet()
join_room_request_packet(room_code)
leave_room_request_packet()
set_ready_request_packet(ready)
start_game_request_packet()
start_single_player_request_packet(local_profile_id)
return_to_lobby_request_packet()
authenticate_request_packet(token)
```

Generated files are outputs. Packet changes must be made in `shared/packets/lobby.toml` and pushed through `data-sync`.

## Service responsibilities

### Client networking

Client networking owns:

```text
building generated lobby request dictionaries
encoding dictionaries to JSON
sending over WebSocket
decoding inbound JSON packets
validating packet envelope shape
classifying room/auth packets by generated type constants
emitting room/auth signals to session controllers
```

Client networking does not own:

```text
room authority
room rules
member identity assignment
match lifecycle
room snapshot construction
packet schema source files
```

### Client lobby/session flow

Client lobby and session flow owns:

```text
sending create/join/start-single-player boot requests at the right time
waiting for websocket auth before multiplayer boot when required
applying room snapshots to transient room and lobby state
showing multiplayer lobby presentation only for multiplayer Lobby state
routing Ready, Start, and Leave UI intent into outbound lobby requests
caching match-result summary from room snapshots
beginning gameplay packet acceptance when room state reaches InGame
presenting room errors to the user
```

Client lobby/session flow does not own:

```text
server admission
room-code validity
capacity
readiness authority
owner authority
room lifecycle authority
match result authority
durable player data persistence
```

### Game-server networking

Game-server networking owns:

```text
WebSocket upgrade and session lifetime
packet envelope decode
generated ClientPacket decode
auth packet routing
lobby packet routing
multiplayer authenticated-account admission checks
session current room state
session current active game player routing
room-session registry for broadcasts
local profile attachment for single-player room members
account identity attachment for multiplayer room members
room snapshot packet projection
room error packet encoding
outbound queueing
```

Game-server networking does not own:

```text
room membership authority
room rules
room lifecycle decisions
game simulation
player-data persistence
packet schema source files
client presentation
```

### Game-server rooms

Game-server rooms own:

```text
room creation
room lookup
room-code normalization and validation
room joinability
room membership
owner selection
readiness
room state transitions
single-player room creation/start
multiplayer start policy
return-to-lobby policy
game-over room state
resolved match summary
room-domain error codes
```

Game-server rooms do not own:

```text
WebSocket transport
generated packet decode
per-session outbound queues
client UI state
auth token verification internals
packet schema source files
```

### Shared packet data pipeline

The packet schema pipeline owns:

```text
loading lobby TOML packet schemas
validating packet structs, packet types, fields, outputs, and builders
generating Go packet structs/constants
generating GDScript constants/builders
checking generated output drift
```

The packet schema pipeline does not own runtime acceptance or rejection semantics for any lobby request.

## Validation and testing

Packet schema validation:

```text
data-sync -validate -packets
data-sync -check -packets -go -gds
```

Review generated packet output before writing it:

```text
data-sync -diff -packets -go -gds
```

Write generated packet output after schema edits:

```text
data-sync -push -packets -go -gds
```

Relevant game-server tests:

```text
services/game-server/internal/networking/gameplay_packets_test.go
services/game-server/internal/networking/room_error_test.go
services/game-server/internal/networking/room_snapshot_test.go
services/game-server/internal/networking/session_auth_test.go
services/game-server/internal/networking/player_activation_test.go
services/game-server/internal/networking/room_sessions_test.go
services/game-server/internal/networking/websocket_test.go
services/game-server/tests/networking/room_snapshot_test.go
services/game-server/tests/game/packets_generated_test.go
services/game-server/internal/rooms/manager_test.go
services/game-server/internal/rooms/room_join_test.go
services/game-server/internal/rooms/room_lobby_test.go
services/game-server/internal/rooms/room_lifecycle_test.go
services/game-server/internal/rooms/room_members_test.go
```

Relevant client tests:

```text
client/tests/unit/test_lobby_session_state.gd
client/tests/unit/test_lobby_member_view_model.gd
client/tests/unit/test_lobby_status_view_model.gd
client/tests/unit/lobby/test_lobby_shell_flow.gd
client/tests/unit/lobby/test_lobby_return_flow.gd
client/tests/unit/test_room_session_controller.gd
```

Suggested game-server verification:

```text
go test -buildvcs=false ./services/game-server/internal/networking ./services/game-server/internal/rooms ./services/game-server/tests/networking ./services/game-server/tests/game
```

Suggested client verification when packet readers or lobby presentation behavior changes:

```text
godot --headless --path client -s res://addons/gut/gut_cmdln.gd -gdir=res://tests/unit -ginclude_subdirs -gexit
```

## Code map

### Packet source and generation

```text
shared/packets/lobby.toml
shared/packets/outputs.toml
tools/data_sync/
client/scripts/generated/networking/packets/packets.gd
services/game-server/internal/game/packets.go
```

### Shared codec paths

```text
services/game-server/internal/protocol/packetcodec/codec.go
client/scripts/networking/packets/packet_codec.gd
```

### Client outbound paths

```text
client/scripts/networking/outbound/lobby_client_packets.gd
client/scripts/networking/outbound/client_packet_sender.gd
client/scripts/networking/client_connection_service.gd
client/scripts/networking/network_client.gd
```

### Client inbound paths

```text
client/scripts/networking/inbound/server_packet_router.gd
client/scripts/networking/inbound/server_packet_dispatcher.gd
client/scripts/networking/client_connection_service.gd
client/scripts/session/session_network_controller.gd
client/scripts/session/room_session_controller.gd
```

### Client lobby consumers

```text
client/scripts/lobby/lobby_packet_reader.gd
client/scripts/lobby/lobby_session_state.gd
client/scripts/lobby/lobby_flow.gd
client/scripts/lobby/lobby_shell_flow.gd
client/scripts/lobby/lobby_network_actions.gd
client/scripts/lobby/lobby_return_flow.gd
client/scripts/lobby/multiplayer_lobby_presenter.gd
client/scripts/ui/lobby/
client/scenes/ui/dialogs/multiplayer_lobby.tscn
```

### Game-server inbound routing

```text
services/game-server/internal/networking/websocket_read.go
services/game-server/internal/networking/client_packet_router.go
services/game-server/internal/networking/inbound/client_packet_envelope.go
services/game-server/internal/networking/inbound/router.go
services/game-server/internal/networking/inbound/auth.go
services/game-server/internal/networking/inbound/lobby.go
services/game-server/internal/networking/inbound_adapter.go
```

### Game-server room network adapter

```text
services/game-server/internal/networking/room_handlers.go
services/game-server/internal/networking/room_snapshot.go
services/game-server/internal/networking/room_error.go
services/game-server/internal/networking/room_sessions.go
services/game-server/internal/networking/player_activation.go
services/game-server/internal/networking/websocket.go
services/game-server/internal/networking/session_auth.go
services/game-server/internal/networking/session_admission.go
services/game-server/internal/networking/session_identity.go
```

### Game-server room authority

```text
services/game-server/internal/rooms/manager.go
services/game-server/internal/rooms/lifecycle.go
services/game-server/internal/rooms/room.go
services/game-server/internal/rooms/room_join.go
services/game-server/internal/rooms/room_lobby.go
services/game-server/internal/rooms/room_lifecycle.go
services/game-server/internal/rooms/room_membership.go
services/game-server/internal/rooms/member.go
services/game-server/internal/rooms/constants.go
services/game-server/internal/rooms/roomrules/
```

### Important non-ownership boundaries

```text
services/game-server/internal/game/ owns active gameplay simulation after a room starts.
services/game-server/internal/networking/inbound/gameplay.go owns gameplay packet routing, not room routing.
services/game-server/internal/networking/inbound/telemetry.go owns telemetry packet routing, not room state.
services/game-server/internal/networking/inbound/auth.go owns auth packet routing, not room membership.
services/game-server/internal/playerdata/ owns player-data contracts, not live room state.
services/player-data/ owns profile/stat persistence, not room snapshot presentation.
services/api-server/ owns Rails auth and account persistence, not realtime room state.
client/scripts/lobby/ owns presentation read models, not room authority.
```

## Related docs

* [Protocol](./!README.md)
* [Packet Schemas](../data/packet-schemas.md)
* [Data Sync and SSoT Pipeline](../data/data-sync-and-ssot-pipeline.md)
* [Game Server](../services/game-server/!README.md)
* [Game Server Networking](../services/game-server/networking/!README.md)
* [Room Network Adapter](../services/game-server/networking/room-network-adapter.md)
* [Inbound Packet Routing](../services/game-server/networking/inbound-packet-routing.md)
* [Auth Routing](../services/game-server/networking/auth-routing.md)
* [WebSocket Session Lifecycle](../services/game-server/networking/websocket-session-lifecycle.md)
* [Game Server Rooms](../services/game-server/rooms/!README.md)
* [Room Membership And Identity](../services/game-server/rooms/room-membership-and-identity.md)
* [Lobby And Start Rules](../services/game-server/rooms/lobby-and-start-rules.md)
* [Room Snapshot Projection](../services/game-server/rooms/room-snapshot-projection.md)
* [Client](../services/client/!README.md)
* [Client Networking Flow](../services/client/networking-flow/!README.md)
* [Lobby Session and Presentation](../services/client/lobby-flow/lobby-session-and-presentation.md)
* [Room Entry and Join Dialog](../services/client/lobby-flow/room-entry-and-join-dialog.md)
* [Room Session State](../services/client/app-shell-and-session/room-session-state.md)
* [Auth Session Flow](../services/client/auth-session-flow.md)

## Notes

Legacy architecture documentation correctly identified the durable boundary that still matters for this protocol: WebSocket connection, room membership, and active gameplay participation are separate states.

`PlayerID` is the player-facing identity used in room snapshots. `SessionID` and `MemberID` stay server-internal. Account and local-profile identifiers can be attached to room members for reporting/persistence integration, but room snapshots do not expose them.

The current client still has compatibility fallbacks for older lobby member aliases such as `is_ready` and `is_connected`. Current room snapshots use `ready` and `connected`.

`create_room_request` has a schema-level `local_profile_id` field, but the active create-room client builder and server handler do not use it. Current local profile attachment for gameplay uses `start_single_player_request.local_profile_id`.

`room_state_changed` remains generated and client-routable, but current room lifecycle readback should be documented and tested through `room_snapshot` unless a server producer for `room_state_changed` is added.
