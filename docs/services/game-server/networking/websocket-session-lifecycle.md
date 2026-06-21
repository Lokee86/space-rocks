# WebSocket Session Lifecycle

Parent index: [Game Server Networking](./!INDEX.md)

## Purpose

This document describes the current game-server WebSocket session lifecycle.

It covers WebSocket upgrade handling, per-connection session creation, read/write/lifecycle goroutine coordination, room cleanup on disconnect, and the transient session state used by networking adapters.

## Overview

The game server exposes one realtime WebSocket endpoint:

```text
GET /ws
```

The process root registers that route, then the networking package owns the upgrade and connection runtime.

A successful WebSocket upgrade creates one `webSocketSession`. That session is the per-connection runtime object for:

```text
session identity
current room pointer
current room ID
current active game player ID
outbound message queue
auth verifier access
match result reporter access
```

The WebSocket connection itself is session-only. It does not imply room membership, authenticated identity, or an active gameplay player.

Related player and room boundaries:

* [Room Membership And Identity](../rooms/room-membership-and-identity.md)
* [Player Lifecycle Overview](../simulation/players/player-lifecycle-overview.md)

WebSocket connection identity is not room membership, authenticated identity, or simulation player state.

Current connection lifecycle:

```text
GET /ws
-> Gorilla WebSocket upgrade
-> newWebSocketSession(...)
-> handleConnection(...)
-> start read loop goroutine
-> start room/game-over lifecycle ticker goroutine
-> run write loop on connection goroutine
-> write loop exits on read close or write failure
-> signal lifecycle ticker to stop
-> leave disconnected room if needed
-> close WebSocket connection
```

The session lifecycle boundary owns transport/session coordination. It does not own room rules, match rules, gameplay simulation, packet schema, client presentation, Rails auth internals, or player-data persistence.

## Code root

```text
services/game-server/internal/networking/
```

Related process registration lives in:

```text
services/game-server/cmd/game-server/main.go
```

## Responsibilities

The WebSocket session lifecycle owns:

* registering the game-server WebSocket handler behind the process route setup
* checking allowed WebSocket origins during upgrade
* upgrading `GET /ws` HTTP requests into Gorilla WebSocket connections
* creating one `webSocketSession` per accepted connection
* assigning server-internal session IDs
* initializing each session with Guest identity
* retaining the room manager, auth verifier, and match result reporter for session handlers
* owning the per-session outbound queue
* coordinating read, write, and room/game-over lifecycle goroutines
* reading raw WebSocket text messages
* handing inbound raw messages to packet routing after envelope decode
* writing queued and ticker-driven outbound WebSocket text messages
* advancing room game-over lifecycle from the session lifecycle ticker
* reporting resolved match results before session-driven room exit when needed
* leaving/detaching the session from its current room on disconnect
* clearing session room and active game-player state after room exit
* closing the WebSocket connection when the session runtime exits
* logging transport-adjacent connect, disconnect, read-close, write-close, and decode failures

## Does not own

The WebSocket session lifecycle does not own:

* HTTP server process startup or shutdown
* health endpoint behavior
* player-data HTTP routes
* packet schema source-of-truth files
* packet generation
* full inbound packet-family routing
* full outbound packet projection
* room creation, join, leave, ready, or start rules
* room owner selection
* room cleanup policy beyond detaching a disconnected session
* gameplay simulation authority
* game-over decision policy
* match result payload construction rules
* auth token issuance
* auth token verification internals
* Rails auth storage
* Local Profile CRUD
* embedded SQLite or Rails player-data persistence
* client-side WebSocket lifecycle
* client packet construction
* client packet routing
* retry, reconnect, acknowledgement, or durable delivery semantics

This doc stays at the transport/session boundary. Room membership and simulation player state are owned elsewhere.

## Runtime surface

The runtime surface is the server-side WebSocket transport for realtime game packets.

The caller is the Godot client networking layer. The game server owns authority behind accepted room, gameplay, auth-result, telemetry, and devtools consequences.

Data crossing this surface includes:

```text
HTTP upgrade request
Origin header
remote address
WebSocket text messages
raw packet bytes
server-internal session ID
session identity
current room ID
current active game player ID
encoded outbound packet bytes
```

The current wire format is JSON text over WebSocket. Encoding and decoding mechanics are delegated to the packet codec and packet-family routing seams.

## Endpoint and origin policy

The game-server process registers:

```text
GET /ws
```

The route is registered by `main.go` and handled by:

```text
networking.WebSocketHandlerWithAuthAndReporter(...)
```

The WebSocket handler uses Gorilla WebSocket upgrade behavior with a custom origin check.

Allowed origins currently include:

```text
empty Origin header
https://space-rocks-client.local
http://localhost:8080
http://127.0.0.1:8080
http://[::1]:8080
```

If upgrade fails, the server logs:

```text
websocket upgrade failed
```

and does not create a session.

Single-player and multiplayer currently use the same `/ws` endpoint. The route does not decide session mode. Single-player versus multiplayer behavior is enforced by packet handling, session identity, and room admission policy.

## Session creation

`newWebSocketSession()` creates the per-connection session object after the WebSocket upgrade succeeds.

Initial session fields include:

```text
conn                -> upgraded Gorilla WebSocket connection
sessionID           -> "session-" plus an atomic sequence number
currentRoomID       -> empty
currentGamePlayerID -> empty
room                -> nil
rooms               -> shared room manager
outbound            -> buffered channel with capacity 16
identity            -> Guest session identity
authVerifier        -> configured token verifier, possibly nil
matchResultReporter -> configured reporter or noop reporter
```

The session starts as Guest even when the client intends to authenticate. Authentication is a later packet-level handshake through `authenticate_request`.

The session ID is server-internal. It is used for room membership attachment and session routing. It is not the player-facing gameplay ID.

## Connection lifecycle

`handleConnection()` owns the top-level connection runtime.

Current setup:

```text
defer session.conn.Close()
defer session.leaveDisconnectedRoom()

readErr := make(chan error, 1)
gameplayLifecycleDone := make(chan struct{})
defer close(gameplayLifecycleDone)

go readClientInput(session, remoteAddr, readErr)
go tickSessionGameplayLifecycle(session, gameplayLifecycleDone)

writeServerMessages(session, remoteAddr, readErr)
```

The write loop runs on the connection goroutine. It is the blocking foreground runtime for the accepted WebSocket session.

When the write loop returns, deferred teardown runs in this order:

```text
close gameplayLifecycleDone
leave disconnected room
close WebSocket connection
```

The lifecycle ticker receives the done signal before room-exit cleanup runs.

## Read loop

`readClientInput()` owns raw inbound WebSocket reads.

Current read flow:

```text
session.conn.ReadMessage()
-> inbound.DecodeClientPacketEnvelope(msg)
-> handleClientPacket(session, remoteAddr, msg, envelope)
```

If `ReadMessage()` returns an error, the read loop sends the error into `readErr` and returns. The write loop receives that error, logs the close/failure, and exits.

If envelope decode fails, the server logs:

```text
websocket packet envelope decode failed
```

and continues reading the next message. The invalid message does not enter packet-family routing.

The read loop does not directly apply room or gameplay behavior. It only reads raw WebSocket messages, performs minimal envelope decode, and hands accepted messages to the inbound packet routing boundary.

## Inbound packet handoff

After a message envelope decodes, `handleClientPacket()` creates an inbound session adapter around the active `webSocketSession`.

The adapter exposes narrow session operations to the inbound router, including:

```text
CurrentRoomID
CurrentRoom
CurrentGamePlayerID
SessionID
OutboundMessages
LogLobbyPacketReceived
HandleAuthenticateRequest
HandleCreateRoomRequest
HandleJoinRoomRequest
HandleLeaveRoomRequest
HandleSetReadyRequest
HandleStartGameRequest
HandleStartSinglePlayerRequest
HandleReturnToLobbyRequest
EnqueuePlayerPauseState
```

This keeps packet-family routing separate from the full session object.

The detailed packet-family order belongs to inbound packet routing documentation. The lifecycle boundary only owns when raw WebSocket messages enter that routing path.

## Write loop

`writeServerMessages()` owns outbound delivery for the active WebSocket session.

It selects over three inputs:

```text
read close/error from readErr
queued outbound messages from session.outbound
server tick events
```

Queued outbound messages are already encoded byte payloads. They are written as WebSocket text messages through:

```text
outbound.WriteServerMessage(session.conn, message, onWriteClose)
```

The write loop also runs a ticker at:

```text
constants.ServerTickRate
```

Ticker-driven writes include:

```text
gameplay presentation state
debug shape catalog, when eligible
debug status, when eligible
```

Gameplay presentation writes require:

```text
session.currentGamePlayerID is not empty
outbound.CanSendGameplayPresentationState(session.room)
```

Debug status and debug shape catalog writes additionally require devtools eligibility inside the outbound helper package.

If a WebSocket write fails, the write-close logger runs and the write loop returns. Returning from the write loop starts connection teardown.

## Outbound queue

Each session owns:

```text
outbound chan []byte
```

The queue is created with capacity:

```text
16
```

Session handlers enqueue already encoded payloads through `session.enqueue(payload)`. Current producers include room errors, room snapshots, auth results, player pause state, and telemetry pong responses.

The queue is an in-memory handoff to the write loop. It is not durable storage and does not provide retry or acknowledgement semantics.

## Gameplay lifecycle ticker

`tickSessionGameplayLifecycle()` is a separate per-session ticker goroutine.

It exists because some room lifecycle transitions are evaluated while gameplay presentation is active.

On each server tick, it checks:

```text
session.currentGamePlayerID is not empty
outbound.CanSendGameplayPresentationState(session.room)
```

When eligible, it calls:

```text
rooms.TickRoomGameOverLifecycle(session.room, BroadcastRoomSnapshot)
```

If that call advances the room game-over lifecycle, networking logs the transition and calls:

```text
rooms.ReportResolvedMatchResultOnce(session.room, session.matchResultReporter)
```

The ticker stops when `gameplayLifecycleDone` is closed during connection teardown.

The ticker does not own match decision policy. Match-over decisions and room state transitions are owned by `services/game-server/internal/rooms/` and `services/game-server/internal/game/rules/`.

## Room attachment and active player state

A WebSocket session can exist without a room.

Room attachment happens later through room request handlers such as create, join, or start single-player. When a room accepts the session, networking stores:

```text
session.room
session.currentRoomID
```

and attaches the session to the room-session registry.

Starting gameplay activates connected room sessions by assigning game-player IDs:

```text
session.currentGamePlayerID = playerID
room.SetMemberPlayerIDForSession(session.sessionID, playerID)
```

Returning a room to lobby clears active game-player IDs for attached sessions.

Important identity separation:

```text
sessionID
server-internal WebSocket/session identity

currentRoomID
room routing state for the current session

currentGamePlayerID
networking-owned active game routing state

PlayerID
player-facing gameplay identity exposed in room/game packets

account_id
authenticated-account routing identity, not gameplay identity

local_profile_id
local durable profile identity, not WebSocket session identity
```

The WebSocket connection itself does not imply room membership. Room membership does not imply an active gameplay player.

## Disconnect and room exit

When a connection runtime exits, `leaveDisconnectedRoom()` runs before the WebSocket connection is closed.

Disconnect cleanup flow:

```text
report resolved match result before room exit, if needed
if session has no current room, return
rooms.LeaveMember(roomID, sessionID, currentGamePlayerID)
detach room session from room-session registry
clear session.room
clear session.currentRoomID
clear session.currentGamePlayerID
if room still has members, broadcast room snapshot
```

Requested leave uses the same core cleanup shape through `leaveRequestedRoom()`, but differs in one important behavior: if the session is not in a room, requested leave enqueues a `not_in_room` room error. Disconnect cleanup silently returns when the session is not in a room.

Both requested leave and disconnect report resolved match results before removing the member when the room has a reportable match result.

## Match result reporting on exit

The session lifecycle has a narrow responsibility for reporting already resolved match results before room exit.

The helper:

```text
session.reportResolvedMatchBeforeRoomExit(reason)
```

checks whether the session still has a room and calls:

```text
rooms.ReportResolvedMatchResultOnce(session.room, session.matchResultReporter)
```

The session lifecycle does not build the match summary or decide who won. It only protects against losing the last session reference before an already resolved match result is reported.

Current exit reasons include:

```text
requested room leave
disconnected
```

## Close and failure behavior

### Upgrade failure

If upgrade fails, the server logs `websocket upgrade failed` and does not create a session.

### Read close

If the read loop exits, the write loop receives the read error.

Expected WebSocket close codes are logged at debug level as:

```text
websocket read closed
```

Unexpected read failures are logged at warn level as:

```text
websocket read failed
```

Expected close codes currently include:

```text
CloseNormalClosure
CloseGoingAway
CloseNoStatusReceived
```

### Write close

If a WebSocket write fails, the write loop exits.

Expected write closes are logged at debug level as:

```text
websocket write closed
```

Unexpected write failures are logged at error level as:

```text
websocket write failed
```

### Packet decode failure

Envelope decode failure logs a warning and keeps the session connected.

Normal packet decode failure is handled by inbound packet routing. It logs a warning and does not close the WebSocket by itself.

### Auth failure

Authentication failure does not close the WebSocket. The session remains connected as Guest unless another flow ends the connection.

### Room/action rejection

Room and gameplay request rejection does not close the WebSocket by itself. Rejected room actions enqueue room errors where the relevant handler owns that behavior.

## Observability

The session lifecycle emits networking and room diagnostics for:

```text
websocket upgrade failed
websocket connected
websocket disconnected
websocket read closed
websocket read failed
websocket write closed
websocket write failed
websocket packet envelope decode failed
room member left
broadcasting room snapshot after member left
reported resolved match result before room exit
room game-over lifecycle advanced; reporting match result
```

Transport diagnostics include remote address where available.

Room/gameplay diagnostics may include:

```text
room_id
player_id
session_id
current_room_id
remaining_members
reason
```

Broader logging policy belongs in game-server observability documentation.

## Data ownership

The WebSocket session lifecycle owns only transient per-connection state.

It mutates:

```text
webSocketSession.currentRoomID
webSocketSession.currentGamePlayerID
webSocketSession.room
webSocketSession.identity
webSocketSession.outbound
room-session registry attachment
```

It does not own durable player data, Rails auth data, Local Profile storage, account stats, match history storage, or packet schema data.

The session may carry identity and room references that downstream systems use for player-data routing, but store selection and persistence are owned by `services/player-data` and integration boundaries.

## Code map

### Process registration

* `services/game-server/cmd/game-server/main.go` - Registers `GET /ws` and injects the room manager, auth verifier, and match result reporter into the WebSocket handler.
* `services/game-server/cmd/game-server/auth_config.go` - Builds the auth verifier used by WebSocket session auth handling.
* `services/game-server/cmd/game-server/player_data_runtime.go` - Builds the player-data runtime used by match result reporting and player-data HTTP routes.

### WebSocket lifecycle files

* `services/game-server/internal/networking/websocket.go` - WebSocket handler construction, upgrade, connection runtime, room-exit cleanup, and match-result-before-exit helper.
* `services/game-server/internal/networking/websocket_origin.go` - WebSocket origin allowlist.
* `services/game-server/internal/networking/websocket_session.go` - Per-connection session state and session construction.
* `services/game-server/internal/networking/websocket_read.go` - Raw WebSocket read loop and envelope-decode handoff.
* `services/game-server/internal/networking/websocket_write.go` - Write loop, outbound queue consumption, gameplay presentation writes, debug status writes, and debug shape catalog writes.
* `services/game-server/internal/networking/websocket_gameplay_tick.go` - Per-session room game-over lifecycle ticker.
* `services/game-server/internal/networking/websocket_close_logging.go` - Expected and unexpected read/write close logging.

### Session adapter and downstream handlers

* `services/game-server/internal/networking/client_packet_router.go` - Builds the inbound router around the active session adapter.
* `services/game-server/internal/networking/inbound_adapter.go` - Exposes narrow session operations to inbound packet-family handlers.
* `services/game-server/internal/networking/room_handlers.go` - Room request handlers that mutate session room fields.
* `services/game-server/internal/networking/room_sessions.go` - Room-session attachment registry used for broadcasts and player activation.
* `services/game-server/internal/networking/player_activation.go` - Assigns and clears active game-player IDs for attached sessions.
* `services/game-server/internal/networking/session_auth.go` - Auth packet handling and session identity mutation.
* `services/game-server/internal/networking/session_identity.go` - Guest and authenticated-account session identity model.
* `services/game-server/internal/networking/session_admission.go` - Authenticated-account admission guard for multiplayer create/join.
* `services/game-server/internal/networking/room_snapshot.go` - Room snapshot build, enqueue, and broadcast helpers.
* `services/game-server/internal/networking/room_error.go` - Room error enqueue helper and generic session outbound enqueue.
* `services/game-server/internal/networking/player_pause_state.go` - Same-session pause-state enqueue helper.

### Outbound write helpers

* `services/game-server/internal/networking/outbound/server_message_writer.go` - Writes encoded payloads as WebSocket text messages.
* `services/game-server/internal/networking/outbound/gameplay_presentation.go` - Builds encoded gameplay state presentation packets.
* `services/game-server/internal/networking/outbound/gameplay_state_metrics.go` - Logs large gameplay packets and slow writes.
* `services/game-server/internal/networking/outbound/debug_status_presentation.go` - Builds encoded debug status packets.
* `services/game-server/internal/networking/outbound/debug_shape_catalog_presentation.go` - Builds encoded debug shape catalog packets.

### Packet and source boundaries

* `services/game-server/internal/protocol/packetcodec/codec.go` - JSON encode/decode wrapper used by session routing.
* `services/game-server/internal/game/packets.go` - Generated game/lobby/auth/telemetry packet structs and constants.
* `services/game-server/internal/game/runtime/packets_generated.go` - Generated runtime packet shapes.
* `services/game-server/internal/devtools/packets_generated.go` - Generated devtools packet shapes.
* `shared/packets/gameplay.toml` - Gameplay packet source definitions.
* `shared/packets/lobby.toml` - Lobby, auth, and room packet source definitions.
* `shared/packets/debug.toml` - Devtools packet source definitions.
* `shared/packets/outputs.toml` - Packet generation output routing.

### Important non-ownership boundaries

* `services/game-server/internal/rooms/` owns room membership, room lifecycle, room cleanup policy, match lifecycle, and match result resolution state.
* `services/game-server/internal/game/` owns authoritative simulation state and gameplay packet projection.
* `services/game-server/internal/game/rules/` owns match/mode policy decisions.
* `services/game-server/internal/devtools/` owns devtools command effects and debug status inputs.
* `services/player-data/` owns player-data runtime routing and backing-store selection.
* `services/api-server/` owns Rails auth and authenticated-account persistence.
* `client/scripts/networking/` owns client-side WebSocket connection lifecycle and packet send/receive behavior.

## Tests and verification

Relevant tests include:

* `services/game-server/internal/networking/websocket_test.go`
* `services/game-server/internal/networking/session_identity_test.go`
* `services/game-server/internal/networking/session_auth_test.go`
* `services/game-server/internal/networking/player_activation_test.go`
* `services/game-server/internal/networking/room_sessions_test.go`
* `services/game-server/internal/networking/room_snapshot_test.go`
* `services/game-server/internal/networking/room_error_test.go`
* `services/game-server/internal/networking/gameplay_packets_test.go`
* `services/game-server/internal/networking/outbound/gameplay_presentation_test.go`
* `services/game-server/internal/networking/outbound/debug_status_presentation_test.go`
* `services/game-server/internal/networking/outbound/debug_shape_catalog_presentation_test.go`
* `services/game-server/tests/networking/room_snapshot_test.go`
* `services/game-server/tests/game/pause_test.go`

`websocket_test.go` currently verifies that requested leave reports resolved match results before removing the room member and that disconnected leave skips already reported match results while clearing session room state.

A focused verification command for this boundary is:

```text
cd services/game-server && go test -buildvcs=false ./internal/networking ./internal/networking/outbound ./internal/rooms ./internal/game/rules ./cmd/game-server
```

## Related docs

* [Game Server Networking](./!INDEX.md)
* [Game Server](../!INDEX.md)
* [Game Server Process](../process/!INDEX.md)
* [Game Server Rooms](../rooms/!INDEX.md)
* [Game Server Simulation](../simulation/!INDEX.md)
* [Game Server Integrations](../integrations/!INDEX.md)
* [Game Server Observability](../observability/!INDEX.md)
* [Inbound Packet Routing](./inbound-packet-routing.md)
* [Outbound Packet Routing](./outbound-message-flow.md)
* [Room Network Adapter](./room-network-adapter.md)
* [Gameplay Network Adapter](./gameplay-network-adapter.md)
* [Auth Routing](./auth-routing.md)
* [Realtime WebSocket Protocol](../../../protocol/realtime-websocket-protocol.md)
* [Packet Schema Pipeline](../../../data/packet-schemas.md)

## Notes

The current session object is shared by the read loop, write loop, lifecycle ticker, and session handlers. There is no separate session actor abstraction. This document describes the current implementation shape, not a general concurrency model.

The WebSocket lifecycle intentionally stays separate from room/game authority. Adding new room or gameplay behavior should usually extend the packet routing, room, or game seams rather than adding rules directly to WebSocket upgrade or read/write loop code.

The `/ws` endpoint is shared by local single-player and multiplayer. New mode distinctions should remain explicit session/packet/admission behavior unless the protocol architecture changes.
