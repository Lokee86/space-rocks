# Inbound Packet Routing

Parent index: [Game Server Networking](./!README.md)

## Purpose

This document describes the current game-server inbound packet routing path.

It covers how raw WebSocket messages are decoded, classified, and routed from a `webSocketSession` into auth, telemetry, lobby, gameplay, and devtools handlers.

## Overview

Inbound packet routing begins after the WebSocket connection has already been upgraded and a `webSocketSession` exists.

The read loop receives raw WebSocket messages, decodes only the packet envelope first, then routes the message through the inbound packet router. The router handles devtools packet families before decoding the message into the normal generated `game.ClientPacket` shape. After normal decode, auth, telemetry, lobby, and gameplay handlers are tried in order.

Current flow:

```text
webSocketSession connection
-> readClientInput
-> conn.ReadMessage()
-> inbound.DecodeClientPacketEnvelope(raw message)
-> handleClientPacket(session, remoteAddr, raw message, envelope)
-> newInboundSessionAdapter(session)
-> inbound.RouteClientPacket(...)
-> packet-family handler
-> session, room, game, auth, telemetry, or devtools boundary
```

The routing path owns packet classification and adapter handoff only. It does not own room rules, gameplay simulation authority, auth token verification logic, packet schema source-of-truth files, or outbound presentation packet production.

## Code root

```text
services/game-server/internal/networking/
services/game-server/internal/networking/inbound/
```

## Responsibilities

- Read raw WebSocket messages from an established session.
- Decode the minimal packet envelope before full packet routing.
- Log and skip messages whose envelope cannot be decoded.
- Construct an inbound session adapter around `webSocketSession`.
- Route devtools command packets before normal `game.ClientPacket` decode.
- Decode normal client packets through the server packet codec.
- Route authenticate packets to session auth handling.
- Route telemetry ping packets to same-session telemetry pong handling.
- Route lobby packets to room/session handlers.
- Route gameplay packets to the current room game instance.
- Consume some recognized packets when the session lacks the required room or active game player.
- Keep packet-family routing separate from WebSocket lifecycle, room authority, gameplay authority, and outbound message writing.

## Does not own

- WebSocket upgrade policy.
- WebSocket write-loop behavior.
- Outbound state packet projection.
- Room membership rules.
- Room lifecycle rules.
- Game simulation rules.
- Input application semantics inside the game instance.
- Respawn validity.
- Target selection authority inside the game instance.
- Auth token verification implementation.
- Rails account identity storage.
- Player-data persistence.
- Packet schema source-of-truth files.
- Generated packet code.
- Client packet construction.
- Client-side packet routing.
- Devtools command behavior after command routing.
- Logging policy outside routing-adjacent diagnostics.

## Domain roles

### Envelope decode

The read loop calls:

```text
inbound.DecodeClientPacketEnvelope(raw message)
```

before routing the message. The envelope currently contains only:

```text
type
```

This first decode lets routing identify packet families that do not use the normal generated `game.ClientPacket` shape, especially devtools commands.

If envelope decode fails, networking logs:

```text
websocket packet envelope decode failed
```

and continues reading the next WebSocket message. The failed message does not enter packet-family routing.

### Router construction

`handleClientPacket` builds an `inbound.ClientPacketRouter` with closures around the active `webSocketSession`.

The concrete session is hidden behind `inboundSessionAdapter`, which exposes only the operations needed by inbound packet-family handlers:

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

This keeps the inbound package mostly pure and avoids passing the full networking session object into every packet-family handler.

### Routing order

Current route order is:

```text
simple devtools packets
placement devtools packets
remaining devtools packets
normal game.ClientPacket decode
auth packets
telemetry packets
lobby packets
gameplay packets
```

Devtools packets route first because devtools command packet types and payloads are generated under the devtools package, not the normal game packet family.

Normal packets are decoded into:

```text
game.ClientPacket
```

only after devtools handlers have had the chance to consume the message.

### Devtools packet routing

Devtools packet routing checks `envelope.Type` before normal packet decode.

Current devtools groups are:

```text
simple devtools packets
placement devtools packets
remaining devtools packets
```

Simple devtools packet types include toggles, player counter mutation commands, kill-player, and entity clearing. Placement devtools packets include spawn entity and spawn pickup commands. Remaining devtools packets include continuous bullet stream and debug respawn commands.

All devtools groups delegate to the same command handling path:

```text
handleDevtoolsCommandPacket
-> packetcodec.Decode(raw message, devtools.DebugCommand)
-> devtools.HandleCommand(room.GameInstance(), currentGamePlayerID, command)
```

Devtools packet routing requires both a current room and a current active game player. If either is missing, the devtools packet is consumed but no command is applied. This prevents devtools command packets from falling through into normal game packet routing.

Inbound routing does not own the devtools command effects. Those are owned by `services/game-server/internal/devtools/` and narrow game-owned devtools export seams under `services/game-server/internal/game/`.

### Auth packet routing

Auth routing handles:

```text
authenticate_request
```

The handler delegates to:

```text
session.HandleAuthenticateRequest(token)
```

through the inbound adapter.

Auth packet routing does not verify the token itself. Verification is owned by the session auth path and its configured `TokenVerifier`.

Auth packets do not require room membership or an active game player.

### Telemetry packet routing

Telemetry routing handles:

```text
telemetry_ping
```

The server responds by encoding a same-session:

```text
telemetry_pong
```

with:

```text
sequence
client_sent_msec
server_received_msec
server_sent_msec
```

The response is written to the current session outbound channel only.

Telemetry routing does not require room membership, does not require an active game player, and does not mutate gameplay state.

### Lobby packet routing

Lobby routing handles:

```text
create_room_request
join_room_request
leave_room_request
set_ready_request
start_game_request
start_single_player_request
return_to_lobby_request
```

The inbound lobby handler only classifies packet type and delegates to session methods through the adapter.

The session methods own the actual room consequences, including auth admission checks for multiplayer create/join, room creation, room join, leave, ready state, game start, single-player room creation, room return-to-lobby behavior, player activation/deactivation, room snapshots, and room errors.

### Gameplay packet routing

Gameplay routing handles packet types that affect the current active game instance.

Current direct gameplay routes include:

```text
input
respawn
client_config
set_target_player_request
select_target_at_position_request
clear_target_request
pause_request
```

For normal gameplay packets:

```text
input
respawn
client_config
```

the handler requires a current room and a current game player. If either is missing, the packet is consumed and ignored. If both exist, the handler delegates to:

```text
room.GameInstance().HandlePacket(currentGamePlayerID, packet)
```

For target and pause packets, the handler first requires a current room and active game player, then routes by packet type:

```text
set_target_player_request
-> Game.SetPlayerTarget(currentGamePlayerID, packet.TargetID)

select_target_at_position_request
-> Game.SelectTargetAtPosition(currentGamePlayerID, packet.X, packet.Y, TargetRef)

clear_target_request
-> Game.ClearTarget(currentGamePlayerID)

pause_request
-> Game.HandlePacket(currentGamePlayerID, packet)
-> session.EnqueuePlayerPauseState()
```

Gameplay routing does not own the semantic result of these requests. It only forwards the request to the authoritative game instance and performs the pause-state outbound enqueue required by the current networking/session boundary.

### Unknown or unhandled packets

If the envelope decodes but no packet-family handler consumes the packet, the routing function returns without applying the message.

There is no final unknown-packet outbound response in the current server inbound path.

Decode failures for the normal `game.ClientPacket` shape are logged as:

```text
websocket packet decode failed
```

and the packet is not routed further.

## Protocols and APIs

### Inbound routing surface

The inbound routing surface is the server-side handling path for client-originated WebSocket messages.

The caller is the WebSocket read loop. The consumers are auth, telemetry, room/session handlers, game simulation handlers, and devtools command handlers. The game server owns authority behind all accepted room, gameplay, and devtools consequences.

Data crossing this surface is raw WebSocket text, then a minimal decoded envelope, then either a generated `game.ClientPacket` or a generated `devtools.DebugCommand`.

Inbound routing explicitly does not own packet schema, durable player identity, room authority, gameplay rules, or client presentation.

### Current route table

```text
toggle_debug_invincible
toggle_debug_infinite_lives
toggle_debug_freeze_world
toggle_debug_freeze_player
debug_kill_player
debug_set_score
debug_add_score
debug_set_lives
debug_add_lives
debug_clear_bullets
debug_clear_asteroids
-> devtools.DebugCommand
-> devtools.HandleCommand

debug_spawn_entity
debug_spawn_pickup
-> devtools.DebugCommand
-> devtools.HandleCommand

debug_begin_continuous_bullet_stream
debug_respawn_player
-> devtools.DebugCommand
-> devtools.HandleCommand

authenticate_request
-> session.handleAuthenticateRequest

telemetry_ping
-> same-session telemetry_pong

create_room_request
-> session.handleCreateRoomRequest

join_room_request
-> session.handleJoinRoomRequest

leave_room_request
-> session.handleLeaveRoomRequest

set_ready_request
-> session.handleSetReadyRequest

start_game_request
-> session.handleStartGameRequest

start_single_player_request
-> session.handleStartSinglePlayerRequest

return_to_lobby_request
-> session.handleReturnToLobbyRequest

input
respawn
client_config
-> Game.HandlePacket

set_target_player_request
-> Game.SetPlayerTarget

select_target_at_position_request
-> Game.SelectTargetAtPosition

clear_target_request
-> Game.ClearTarget

pause_request
-> Game.HandlePacket
-> session.EnqueuePlayerPauseState
```

### Session-state requirements

Different packet families have different session-state requirements.

```text
authenticate_request
requires websocket session only

telemetry_ping
requires websocket session only

lobby packets
require whatever the delegated room/session handler requires

devtools command packets
require current room and active game player

input, respawn, client_config
require current room and active game player to apply
consume without applying when room/player is missing

target and pause gameplay packets
require current room and active game player
fall through unhandled when room/player is missing
```

The websocket connection itself does not imply room membership, and room membership does not imply an active game player. `currentGamePlayerID` is the networking-owned active game routing state used to target gameplay requests at the current game instance.

## Data ownership

Inbound packet routing owns no durable data.

Transient data handled by this boundary includes:

```text
raw websocket message bytes
client packet envelope
game.ClientPacket
devtools.DebugCommand
remote address for logging
current session routing fields
same-session outbound telemetry pong payload
```

Delegated handlers may mutate session, room, auth, or game state, but those mutations are owned by their downstream boundaries.

Packet type constants and packet payload structs are generated from shared packet source files.

Current relevant source files:

```text
shared/packets/gameplay.toml
shared/packets/lobby.toml
shared/packets/debug.toml
shared/packets/outputs.toml
```

Current relevant generated server files:

```text
services/game-server/internal/game/packets.go
services/game-server/internal/game/runtime/packets_generated.go
services/game-server/internal/devtools/packets_generated.go
```

## Code map

### WebSocket read and route handoff

- `services/game-server/internal/networking/websocket.go` - WebSocket connection setup and lifecycle coordination.
- `services/game-server/internal/networking/websocket_session.go` - Per-connection session fields used by inbound routing.
- `services/game-server/internal/networking/websocket_read.go` - WebSocket read loop, envelope decode, and handoff to client packet routing.
- `services/game-server/internal/networking/client_packet_router.go` - Adapter construction and inbound router wiring.
- `services/game-server/internal/networking/inbound_adapter.go` - Narrow adapter from `webSocketSession` to inbound handler interfaces.

### Inbound routing package

- `services/game-server/internal/networking/inbound/client_packet_envelope.go` - Minimal packet envelope decode.
- `services/game-server/internal/networking/inbound/router.go` - Ordered packet-family routing.
- `services/game-server/internal/networking/inbound/auth.go` - Authenticate packet classification.
- `services/game-server/internal/networking/inbound/telemetry.go` - Telemetry ping/pong handling.
- `services/game-server/internal/networking/inbound/lobby.go` - Lobby packet classification and session method delegation.
- `services/game-server/internal/networking/inbound/gameplay.go` - Gameplay, target, pause, respawn, input, and client config routing.
- `services/game-server/internal/networking/inbound/devtools.go` - Devtools packet type classification and command handoff.

### Downstream networking/session handlers

- `services/game-server/internal/networking/session_auth.go` - Auth request handling and authenticate result enqueue.
- `services/game-server/internal/networking/room_handlers.go` - Room and lobby request handling.
- `services/game-server/internal/networking/player_activation.go` - Active game player activation and deactivation.
- `services/game-server/internal/networking/player_pause_state.go` - Pause-state outbound enqueue.
- `services/game-server/internal/networking/room_snapshot.go` - Room snapshot projection and enqueue/broadcast helpers.
- `services/game-server/internal/networking/room_sessions.go` - Room session attachment and detachment.

### Downstream domain boundaries

- `services/game-server/internal/rooms/` - Room membership, room lifecycle, match lifecycle, and room rule authority.
- `services/game-server/internal/game/input.go` - Authoritative handling for input, respawn, pause, and client config packets.
- `services/game-server/internal/game/player_targeting.go` - Authoritative target selection and target clearing.
- `services/game-server/internal/devtools/` - Server-authoritative devtools command behavior.
- `services/game-server/internal/protocol/packetcodec/codec.go` - Server JSON encode/decode wrapper used by inbound routing.

### Generated and source boundaries

- `services/game-server/internal/game/packets.go` - Generated normal game/lobby/auth/telemetry packet constants and structs.
- `services/game-server/internal/devtools/packets_generated.go` - Generated devtools packet constants and structs.
- `shared/packets/gameplay.toml` - Gameplay packet source definitions.
- `shared/packets/lobby.toml` - Lobby and auth packet source definitions.
- `shared/packets/debug.toml` - Devtools packet source definitions.
- `shared/packets/outputs.toml` - Packet generation output routing.

### Non-ownership boundaries

- `services/game-server/internal/networking/outbound/` owns outbound gameplay and debug presentation write helpers.
- `client/scripts/networking/` owns client-side packet construction, send flow, receive flow, and client packet dispatch.
- `docs/protocol/` owns protocol-level packet behavior.
- `docs/data/` owns packet source-of-truth and generation pipeline documentation.

## Tests

Relevant tests include:

- `services/game-server/internal/networking/gameplay_packets_test.go`
- `services/game-server/internal/networking/session_auth_test.go`
- `services/game-server/internal/networking/player_activation_test.go`
- `services/game-server/internal/networking/room_sessions_test.go`
- `services/game-server/internal/networking/room_snapshot_test.go`
- `services/game-server/internal/networking/room_error_test.go`
- `services/game-server/internal/networking/websocket_test.go`
- `services/game-server/internal/devtools/command_types_test.go`
- `services/game-server/internal/devtools/disabled_test.go`
- `services/game-server/internal/devtools/enabled_default_test.go`

`gameplay_packets_test.go` currently verifies that `client_config` routes through gameplay packet handling into the game instance, and that `start_single_player_request` routes through lobby handling while preserving the local profile ID on the room member.

`session_auth_test.go` verifies auth request handling stores the Rails account identity returned by the token verifier.

Direct unit coverage for `inbound.RouteClientPacket` ordering is currently thin. Existing coverage is mostly through packet-family handlers and downstream networking/session behavior.

## Related docs

- [Game Server Networking](./!README.md)
- [Game Server](../!README.md)
- [Game Server Rooms](../rooms/!README.md)
- [Game Server Simulation](../simulation/!README.md)
- [Game Server Integrations](../integrations/!README.md)
- [WebSocket Session Lifecycle](./websocket-session-lifecycle.md) - WebSocket upgrade, session lifecycle, and read/write loop ownership.
- [Room Network Adapter](./room-network-adapter.md) - Room/session adapter behavior behind lobby packet routing.
- [Gameplay Network Adapter](./gameplay-network-adapter.md) - Gameplay adapter behavior behind input, respawn, pause, target, and client config routing.
- [Auth Routing](./auth-routing.md) - Authenticate packet routing and auth verifier handoff.
- [Telemetry Packet Routing](./telemetry-packet-routing.md) - Telemetry ping/pong routing.
- [Outbound Message Flow](./outbound-message-flow.md) - Outbound server message writing and presentation packet flow.
- [Realtime Websocket Protocol](../../../protocol/stubs/realtime-websocket-protocol.md) - Realtime WebSocket protocol documentation.
- [Gameplay Packets](../../../protocol/stubs/gameplay-packets.md) - Gameplay packet documentation.
- [Lobby Packets](../../../protocol/stubs/lobby-packets.md) - Lobby packet documentation.
- [Devtools Packets](../../../protocol/stubs/devtools-packets.md) - Devtools packet documentation.
- [Packet Schema Pipeline](../../../data/stubs/packet-schema-pipeline.md) - Shared packet schema and generated output documentation.

## Notes

`services/game-server/internal/networking/inbound` is the pure inbound packet family handler boundary.

The current router silently drops valid-envelope packets that are decoded but not handled by any packet family. That is current behavior, not a protocol guarantee.

Devtools routing intentionally occurs before normal `game.ClientPacket` decode. New devtools packet families should preserve that separation unless the generated packet ownership changes.
