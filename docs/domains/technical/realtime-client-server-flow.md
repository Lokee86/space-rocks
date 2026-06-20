# Realtime Client Server Flow

Parent index: [Technical](./!INDEX.md)

## Purpose

This document describes the cross-system realtime client/server flow for Space Rocks.

It covers how the Godot client and Go game server exchange live WebSocket traffic, how packet authority is split, how client intent becomes server-owned state, and how server-owned state returns to client presentation.

## Overview

The realtime client/server flow is the live communication path between the Godot client and the Go game server.

The current transport is JSON text over a WebSocket connection. The game server exposes one realtime route:

```text
GET /ws
```

The client selects a WebSocket target from session mode, opens the connection, optionally sends a WebSocket authentication request, sends room or gameplay intent packets, receives authoritative server packets, and routes those packets into room, gameplay, devtools, telemetry, or presentation flows.

The game server owns the authority behind the connection. The client may request room actions, gameplay input, target changes, pause, respawn, devtools actions, telemetry pings, and viewport configuration. The server decides whether those requests are accepted and projects the resulting authoritative state back to the client.

The realtime flow is not the same thing as gameplay session ownership. A WebSocket connection is only transport readiness. Room membership, active gameplay participation, authenticated account identity, Local Profile identity, and match-result persistence are separate states owned by separate systems.

## Participating systems

* [Client](../../services/client/!INDEX.md) - owns WebSocket connection startup, polling, local packet construction, inbound packet classification, gameplay packet acceptance gates, and presentation routing.
* [Game Server](../../services/game-server/!INDEX.md) - owns the `/ws` route, WebSocket upgrade, per-session transport state, inbound packet routing, room/session adapters, authoritative gameplay routing, and outbound server packets.
* [Protocol](../../protocol/!INDEX.md) - owns communication and message-flow documentation for realtime packets.
* [Data](../../data/!INDEX.md) - owns packet schema source files and generated packet outputs shared by client and game server.
* [Devtools](../../devtools/!INDEX.md) - owns debug-only client/server tooling that uses the normal realtime transport.
* [Player Data](../../services/player-data/!INDEX.md) - owns durable stat/result routing after authoritative match facts leave the live realtime flow.
* [API Server](../../services/api-server/!INDEX.md) - owns authenticated account auth and Rails-backed persistence outside the live realtime simulation path.

## Authority boundaries

The game server owns live authoritative state.

Server-owned authority includes:

```text
room creation and joining
multiplayer admission
room membership
ready state acceptance
game start acceptance
active gameplay player activation
gameplay simulation
movement outcomes
projectile creation
collisions
damage
score
lives
death
respawn validity
pause state
target state
match-over state
room snapshots
match-result summary production
```

The client owns local presentation and intent collection.

Client-owned behavior includes:

```text
menu/session request initiation
WebSocket target selection
socket polling
local input collection
packet send calls
inbound packet classification
gameplay packet acceptance gating
world rendering
HUD presentation
audio/effects presentation
devtools presentation
telemetry display
local navigation after match-end buttons
```

The client does not own authoritative gameplay results. A sent packet is a request or observation, not proof that the server accepted the action.

The packet schema owns packet shapes and generated helpers, but it does not own runtime meaning. Runtime meaning belongs to the service that handles the packet.

```text
shared/packets/*.toml
-> generated packet constants, structs, and client builders
-> runtime routing and authority in client/game-server service code
```

The WebSocket connection itself has no durable persistence authority. Player-data and API-server systems participate only after identity, auth, or match-result data crosses explicit service boundaries.

## State separation

The realtime flow preserves these separate states:

```text
WebSocket connection
!= authenticated account identity
!= room membership
!= active gameplay player
!= player-facing gameplay identity
!= durable profile/account identity
```

A connected client can exist with no room.

A room member can exist before gameplay starts.

An active gameplay player exists only after the game server accepts a start path and activates connected room members into gameplay players.

A player-facing gameplay ID is not the same as a WebSocket session ID, room member ID, account ID, Rails user ID, or Local Profile ID.

## Flow summary

### 1. Client chooses a session route

The client begins from local menu or session intent.

Current boot request types are:

```text
single_player
create_room
join_room
```

The client maps the requested session mode to a WebSocket URL. Current local single-player and multiplayer URLs point to the same `/ws` route, but play mode is not defined by the route path. Mode is expressed through the boot request and enforced by server-side room, session, and admission rules.

### 2. Client opens the WebSocket connection

The client networking layer opens the WebSocket, sets the configured Origin header, polls connection state, receives raw text messages, decodes packet envelopes, and emits decoded packet dictionaries.

When the socket opens, the client sends an `authenticate_request` only if an auth token exists. Single-player boot does not require this authentication result. Multiplayer boot waits for WebSocket auth success unless token verification is unavailable, in which case the pending request is sent so the server can fail admission explicitly.

Connection success means only:

```text
the WebSocket transport is open
```

It does not mean:

```text
the client is authenticated
the client is in a room
the client is ready
the client has an active game player
the client is allowed to affect gameplay
```

### 3. Client sends a boot or room request

After connection/auth gating, the client sends one of the boot packets:

```text
start_single_player_request
create_room_request
join_room_request
```

The game server receives the packet through the WebSocket read loop, decodes the packet envelope, routes it through inbound packet handling, and delegates room behavior to room/session handlers.

Single-player creates a started non-joinable room and activates the connected session into gameplay.

Multiplayer create and join require authenticated account admission. Successful create/join attaches the WebSocket session to a room and sends or broadcasts a room snapshot.

### 4. Server establishes room and active player state

The game server maintains separate session, room, and gameplay routing state.

The WebSocket session carries transient routing fields such as:

```text
session identity
current room
current room ID
current active game player ID
outbound message queue
```

Room membership is owned by the room system. Active gameplay player routing is assigned only when a game starts. Networking stores the active game player ID because inbound gameplay packets need a per-connection player target when routing into the current game instance.

### 5. Client sends live intent packets

During live gameplay, the client sends intent through generated realtime packets.

Current client-to-server packet families include:

```text
auth
telemetry
room and lobby requests
gameplay input
respawn
pause
target selection
target clearing
viewport configuration
devtools commands
```

Gameplay packets require a current room and active game player before the server applies them. Auth and telemetry packets require only the WebSocket session. Lobby packets route to room/session handlers, which apply their own room and admission rules.

Devtools command packets use the same WebSocket transport but are routed through the server devtools boundary before normal gameplay packet handling. They are debug-only requests, not a second gameplay authority layer.

### 6. Server routes inbound packets by family

The server inbound route order is:

```text
devtools packet families
normal packet decode
auth packets
telemetry packets
lobby packets
gameplay packets
```

Devtools routes first because devtools command packets are generated into a separate server devtools packet family.

Normal packets decode into the generated client packet shape before auth, telemetry, lobby, and gameplay handlers are tried.

If a packet cannot be decoded, it is logged and ignored. Decode failure does not by itself close the WebSocket.

### 7. Server advances and projects authoritative state

The game server simulation owns the authoritative runtime state.

On each server tick, the WebSocket write path can send gameplay presentation state when the session has an active game player and the room has a game instance in an eligible state.

The current main gameplay output is:

```text
state
```

That packet is projected per player from the authoritative game instance. It can include player state, player sessions, lifecycle state, bullets, asteroids, pickups, events, targeting state, camera/view state, and match-over classification.

The server stamps outbound gameplay state with server send time before encoding and writing it.

### 8. Server sends one-off and ticker-driven packets

The server sends two broad classes of outbound realtime packets.

Queued one-off responses include:

```text
authenticate_result
room_snapshot
room_error
player_pause_state
telemetry_pong
```

Ticker-driven presentation packets include:

```text
state
debug_status
debug_shape_catalog
```

Room snapshots are sent after room lifecycle changes such as create, join, ready, start, single-player start, return to lobby, leave, and disconnect broadcasts.

Telemetry pong is a same-session diagnostic response. It does not require room membership and does not mutate gameplay state.

Debug status and debug shape catalog packets are devtools-only outputs gated by devtools availability and room/gameplay state.

### 9. Client routes inbound packets

The client decodes raw WebSocket text into packet dictionaries, classifies packets by generated packet type constants, and emits typed networking signals.

Current client inbound routes include:

```text
authenticate_result
room_snapshot
room_state_changed
room_error
state
debug_shape_catalog
debug_status
player_pause_state
telemetry_pong
unknown packet fallback
```

Room packets route into room session handling.

Gameplay state routes into gameplay session handling, but gameplay application is gated. The client begins accepting gameplay packets only after room state reaches `InGame`.

Telemetry pong routes to telemetry consumers and does not pass through normal gameplay state application.

### 10. Client applies presentation consequences

After inbound routing, owning client flows apply presentation state.

Examples:

```text
room snapshots update room-session and lobby read models
state packets update gameplay runtime and world presentation
player pause state updates local pause presentation
debug packets update devtools presentation
telemetry pong updates network timing metrics
match-over room state triggers match-end presentation
```

The client renders from server-observed facts. It does not recalculate authoritative outcomes such as score, collision damage, lives, respawn validity, room match-over, or match results.

### 11. Connection closes or session exits

The client can close gracefully for replay, lobby return, pregame return, main-menu return, or normal session cleanup.

The game server also handles read or write failure by tearing down the WebSocket session. During disconnect or requested leave, the server detaches the session from the current room when needed, clears the session’s room and active player routing fields, and broadcasts a room snapshot if remaining room members should observe the change.

If the room already has a resolved match result before exit, the server attempts to report that result before losing the session reference.

## Inputs and outputs

### Client-to-server inputs

The client can send:

```text
authenticate_request
telemetry_ping
start_single_player_request
create_room_request
join_room_request
leave_room_request
set_ready_request
start_game_request
return_to_lobby_request
input
respawn
pause_request
client_config
set_target_player_request
select_target_at_position_request
clear_target_request
debug command packets
```

These inputs are transient realtime requests or observations. They are not durable facts until accepted and reflected by server-owned state or downstream persistence.

### Server-to-client outputs

The game server can send:

```text
authenticate_result
room_snapshot
room_error
room_state_changed
state
player_pause_state
telemetry_pong
debug_status
debug_shape_catalog
```

These outputs are authoritative readback or diagnostic presentation packets. The client consumes them through the networking dispatcher and routes them to session, gameplay, lobby, devtools, telemetry, or UI flows.

### Durable outputs

The realtime flow does not directly persist player progress.

When a match reaches authoritative completion, the game server can report resolved match facts through the match-result reporting boundary into player-data. That is a separate persistence flow. The realtime packet shown to clients is presentation-safe room/gameplay state, not the durable storage contract.

## Integration points

### Packet schemas

Realtime packet shapes are sourced from packet schema files under:

```text
shared/packets/
```

The generated client and server outputs must stay aligned through the data pipeline. The realtime flow consumes generated packet constants, generated structs, and generated client packet builders, but packet schema data does not decide gameplay meaning.

### Room lifecycle

Room packets are the bridge from WebSocket transport into room authority.

Room lifecycle determines whether a session can create, join, leave, ready, start, return to lobby, or observe match-over state. The networking layer routes packets and stores per-session routing fields; the room system owns the accepted room state.

### Gameplay simulation

Gameplay packets are the bridge from WebSocket transport into authoritative simulation.

The networking layer routes gameplay requests to the current room game instance. The game simulation owns the actual result of input, respawn, pause, target, collision, damage, scoring, death, lifecycle, and state projection.

### Devtools

Devtools use the normal realtime packet and WebSocket path. They do not create a separate debug transport.

Client devtools send generated debug packets. Server networking identifies debug packets before normal gameplay decode and routes them to devtools command handling. Debug effects still route through server-owned devtools and gameplay seams.

### Telemetry

Telemetry ping/pong is diagnostic realtime traffic.

The client sends telemetry pings only when the telemetry flow requests them. The server replies to the same WebSocket session with timing fields. Telemetry does not require room membership and does not mutate gameplay state.

## Out of scope

This domain document does not own:

* direct code maps
* packet field-by-field protocol specification
* generated packet source details
* data-sync command procedures
* WebSocket implementation internals
* room package implementation details
* gameplay simulation phase order
* client world-sync implementation
* HUD or UI widget behavior
* devtools command semantics
* auth token verification internals
* Rails account storage
* Local Profile storage
* player-data persistence internals
* future realtime lane, delta snapshot, quantization, bit-packing, or protobuf design

Those details belong in service, protocol, data, devtools, systems-design, planning, or limits documentation.

## Related docs

* [Technical](./!INDEX.md)
* [Gameplay Session Flow](../player-experience/gameplay-session-flow.md)
* [Client](../../services/client/!INDEX.md)
* [Client Networking Flow](../../services/client/networking-flow/!INDEX.md)
* [Session Boot And Network Target](../../services/client/app-shell-and-session/session-boot-and-network-target.md)
* [Gameplay Runtime](../../services/client/gameplay-runtime/!INDEX.md)
* [World Sync](../../services/client/world-sync/!INDEX.md)
* [Game Server](../../services/game-server/!INDEX.md)
* [Game Server Networking](../../services/game-server/networking/!INDEX.md)
* [Game Server Rooms](../../services/game-server/rooms/!INDEX.md)
* [Game Server Simulation](../../services/game-server/simulation/!INDEX.md)
* [Protocol](../../protocol/!INDEX.md)
* [Packet Schemas](../../data/packet-schemas.md)
* [Data](../../data/!INDEX.md)
* [Devtools](../../devtools/!INDEX.md)
* [Realtime Protocol Architecture](../../planning/protocol/realtime-protocol-architecture.md)
* [Network Observability And Packet Budget](../../planning/domains/technical/network-observability-and-packet-budget.md)

## Notes

Client input is sent to the server, the server advances simulation, and clients render received state. That rule remains current.

WebSocket connection, room membership, and active gameplay participation are separate states. The current implementation still depends on that separation.

The current realtime flow sends full gameplay presentation state on the server tick path. Future realtime protocol work may introduce lanes, deltas, quantization, bit packing, or binary encoding, but those are planning facts until implemented.

Single-player and multiplayer can currently use the same local `/ws` route. That does not collapse their authority model. The boot packet, session mode, auth/admission rule, room joinability, and player-data identity context distinguish the flows.
