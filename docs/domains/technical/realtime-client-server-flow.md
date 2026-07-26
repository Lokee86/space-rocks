---
author: brian
created: "2026-07-19"
document_id: 019f7d55-fb2c-710c-a88e-90912b4c2221
document_type: general
policy_exempt: false
summary: This document describes the cross-system realtime client/server flow for Space Rocks.
---
# Realtime Client Server Flow

Parent index: [Technical](./!INDEX.md)

## Purpose

This document describes the cross-system realtime client/server flow for Space Rocks.

It covers how the Godot client and Go game server exchange live WebSocket and WebRTC DataChannel traffic, how packet authority is split, how client intent becomes server-owned state, and how server-owned state returns to client presentation.

## Overview

The realtime client/server flow is the live communication path between the Godot client and the Go game server.

The current transport is JSON text over a WebSocket connection for session/control/auth/lobby/signaling packets, and JSON text over lane-specific WebRTC DataChannels for active realtime gameplay packets. The game server exposes one realtime route:

```text
GET /ws
```

The client selects a WebSocket target from session mode, opens the connection, optionally sends a WebSocket authentication request, sends room or gameplay intent packets, receives authoritative server packets, and routes those packets into room, gameplay, devtools, telemetry, or presentation flows.

The game server owns the authority behind the connection. The client may request room actions, gameplay input, target changes, pause, respawn, devtools actions, telemetry pings, and viewport configuration. The server decides whether those requests are accepted and projects the resulting authoritative state back to the client.


The realtime flow is not the same thing as gameplay session ownership. A WebSocket connection is only transport readiness. Room membership, active gameplay participation, authenticated account identity, Local Profile identity, and match-result persistence are separate states owned by separate systems.

## Participating systems

* [Client](../../services/client/!INDEX.md) - owns WebSocket connection startup, WebRTC transport setup, polling, local packet construction, inbound packet classification, gameplay packet acceptance gates, and presentation routing.
* [Game Server](../../services/game-server/!INDEX.md) - owns the `/ws` route, WebSocket upgrade, per-session WebSocket and WebRTC transport state, inbound packet routing, room/session adapters, authoritative gameplay routing, queued WebSocket packets, and active WebRTC gameplay lane packets.
* [Protocol](../../protocol/!INDEX.md) - owns communication and message-flow documentation for realtime packets.
* [Data](../../data/!INDEX.md) - owns packet schema source files and generated packet outputs shared by client and game server.
* [Devtools](../../devtools/!INDEX.md) - owns debug-only client/server tooling that uses the normal realtime transport.
* [Player Data](../../services/player-data/!INDEX.md) - owns durable stat/result routing after authoritative match facts leave the live realtime flow.
* [API Server](../../services/api-server/!INDEX.md) - owns authenticated account auth and Rails-backed persistence outside the live realtime simulation path.

## Client receive hardening

The client applies defensive pre-match buffering limits before authoritative activation: at most 4 match buckets, 128 packets and 256 KiB of estimated expanded JSON per match, with a 5000 ms bucket lifetime. The oldest bucket is evicted when the bucket cap is reached. If the selected authoritative match has lost or expired buffered state, the client fails closed and requests world, overlay, and session recovery. These are client receive limits; they do not change the server's approximately 1200-byte candidate construction cap.

`world_full`, asteroid lifecycle, and bullet lifecycle assemblies each allow at most 128 chunks, 16384 cumulative records, 2 MiB of estimated expanded JSON, and 5000 ms of assembly lifetime. Limit, expiry, malformed metadata, interrupted, duplicate, mismatched, and non-contiguous failures reset the incomplete assembly, apply no partial state, and request authoritative recovery.

Baseline IDs must be non-empty strings. Sequence and chunk values must be finite, non-negative, integer-valued numerics; final-chunk metadata must be an actual boolean consistent with index/count. Valid stale deltas remain silently rejected.

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

Logical packet schemas own packet type strings, structs, fields, JSON names, generated constants, and generated helpers. The separate physical contract in `shared/packets/realtime_wire.toml` owns compact aliases, packet metadata, record encodings, tuple and sparse layouts, quantization assignments, ID codecs/selectors, event layouts, and decode compatibility alternatives. Runtime protocol code projects authoritative state and applies the generated contract; runtime meaning belongs to the service that handles the packet.

```text
shared/packets/*.toml
-> generated packet constants, structs, and client builders
-> runtime routing and authority in client/game-server service code

services/game-server/internal/protocol/realtime/wire_packets.go
-> active lane wire-map emission

shared/packets/realtime_wire.toml
-> physical compact-wire source of truth

services/game-server/internal/protocol/realtimewire/generated.go
services/game-server/internal/protocol/realtime/compact_wire_packet.go
services/game-server/internal/protocol/realtime/compact_wire_descriptor.go
client/scripts/generated/networking/realtime_wire_generated.gd
client/scripts/protocol/realtime/compact_lane_packet.gd
client/scripts/protocol/realtime/compact_wire_descriptor_decoder.gd
-> generated descriptor data and generic encode/decode application

services/game-server/internal/protocol/realtime/quantize/
client/scripts/protocol/realtime/realtime_quantize.gd
-> quantization algorithms; field-path policy assignments come from generated realtime-wire descriptors
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

The client networking layer opens the WebSocket, sets the configured Origin header, polls connection state, receives raw WebSocket text for session/control/auth/lobby/signaling packets, and receives raw WebRTC DataChannel text for active realtime gameplay packets before decoding packet envelopes and emitting decoded packet dictionaries.

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
WebSocket:
  auth
  legacy telemetry ping
  room and lobby requests
  gameplay input
  respawn
  pause
  target selection
  target clearing
  viewport configuration

sr.tooling:
  runtime measurement
  runtime developer commands
  developer-readout subscriptions and requests
  tooling telemetry requests
```

Gameplay packets require a current room and active game player before the server applies them. Auth and legacy telemetry ping require only the WebSocket session. Lobby packets route to room/session handlers, which apply their own room and admission rules.

Runtime developer commands and developer-readout requests use `sr.tooling`. The tooling router applies packet policy, room attachment, capability checks, and request correlation before dispatching to devtools or readout owners. They are debug-only interactions, not a second gameplay authority layer.

### 6. Server routes inbound packets by transport and family

The WebSocket inbound route order is:

```text
normal packet envelope decode
auth packets
legacy telemetry packets
lobby packets
gameplay/control packets
```

The separate `sr.tooling` route handles measurement, runtime developer commands, developer-readout subscriptions/requests, and tooling telemetry packets before they can reach any owning controller or provider.

Normal WebSocket packets decode into the generated client packet shape before auth, telemetry, lobby, and gameplay handlers are tried.

If a packet cannot be decoded, it is logged and ignored. Decode failure does not by itself close the WebSocket.

### 7. Server advances and projects authoritative state

The game server simulation owns the authoritative runtime state.

The 60 Hz session write-loop tick can build and encode active gameplay lane packets. Ship, asteroid, and bullet hot movement independently use the same chunk-pressure cadence tiers: one chunk at 60 Hz, two chunks at 30 Hz, three chunks at 20 Hz, and four or more chunks at the 15 Hz floor. On an eligible tick, every chunk for the logical sequence is sent in one unordered same-tick burst. Movement and lifecycle changes on other lanes do not bypass cadence, and reliable world/lifecycle traffic remains eligible while a hot lane is suppressed.

The current active gameplay output uses lane-native packet families: `world_full`/`world_delta`, `ship_delta`, `asteroid_delta`, `bullet_delta`, `ships_lifecycle`, `asteroids_lifecycle`, `bullets_lifecycle`, `overlay_full`/`overlay_delta`, `session_full`/`session_delta`, and `event_batch`. Generated record descriptors tuple-pack selected world, hot-movement, session, and known-event records. Lifecycle creates and deletes use the contract's primary map/readable-ID encoding; historical tuple and numeric-ID lifecycle forms remain declared decode-only alternatives. Active realtime gameplay delivery uses WebRTC DataChannel text over ordered/reliable lanes for `sr.world`, `sr.overlay`, `sr.session`, `sr.event`, `sr.ships.lifecycle`, `sr.asteroids.lifecycle`, and `sr.bullets.lifecycle`, and unordered/unreliable hot lanes for `sr.ships`, `sr.asteroids`, and `sr.bullets` after signaling succeeds; WebSocket remains the session/control/auth/lobby/signaling path. Ship creates, deletes, and non-transform state updates are subtractively removed from active `world_delta` and emitted reliably on `sr.ships.lifecycle`. Asteroid and bullet creates/deletes are emitted reliably on `sr.asteroids.lifecycle` and `sr.bullets.lifecycle`. Dedicated unordered hot lanes carry movement updates only. Regular ship, asteroid, and bullet movement updates are subtractively removed from `sr.world` and emitted as `ship_delta`, `asteroid_delta`, and `bullet_delta`; their lifecycle records no longer remain in `world_delta` as active ownership. The roughly 1,200-byte `HardCapBytes` construction limit applies to `world_full`, `ships_lifecycle`, `asteroids_lifecycle`, `bullets_lifecycle`, `ship_delta`, `asteroid_delta`, and `bullet_delta` candidates. Server candidate expansion exact-encodes compact payloads while constructing chunks, preserves logical identity metadata, and returns an explicit error when an individual record cannot fit. Tuple packing is a wire optimization, not a domain model change.


World lane carries full/bootstrap ship state, pickups, player/world presentation, and non-entity world changes. Ships.lifecycle carries ship creates/deletes plus reliable non-transform updates such as health, shields, ship type, and target state. Asteroids.lifecycle carries asteroid creates/deletes. Bullets.lifecycle carries bullet/projectile creates/deletes. Ships, asteroids, and bullets carry movement updates only. Overlay lane carries receiver-specific HUD-facing values. Session lane carries player/session/lifecycle/asteroid-count presentation state. For world, ship, asteroid, bullet, overlay, and session state lanes, numeric wire quantization, field deltas, sparse delta omission, and compact JSON aliases are current active behavior. `event_batch` is transient presentation-event delivery, not a state delta lane. It uses compact output encoding and tuple-packed known event records, but remains batched. Known event `x`/`y` and `ship_death` `respawn_delay` are quantized during event wire shaping. `event_batch` does not use baselines, deltas, state snapshots, or chunking.

The server stamps outbound gameplay lane packets with server send time before encoding them and delivering them over the current lane policy: ordered/reliable for `sr.world`, `sr.overlay`, `sr.session`, `sr.event`, `sr.ships.lifecycle`, `sr.asteroids.lifecycle`, and `sr.bullets.lifecycle`, and unordered/unreliable for `sr.ships`, `sr.asteroids`, and `sr.bullets`.

The client atomically assembles `world_full`, `ships_lifecycle`, `asteroids_lifecycle`, and `bullets_lifecycle` chunks before applying them. Lifecycle assembly completes before `LifecycleLaneGate` validates the assembled packet's lane-local sequence. Partial, duplicate, mismatched, interrupted, or otherwise invalid series fail closed, do not partially mutate lane state, and request one world resync; router/pipeline reset clears partial assembly. Hot `ship_delta`, `asteroid_delta`, and `bullet_delta` sequence values must be finite, non-negative, integer-valued numerics; missing, fractional, negative, non-finite, string, and boolean values are rejected before hot-lane state mutation. Each hot lane also assembles all declared same-sequence chunks before mutating world state. A newer sequence discards an incomplete older sequence without partial application; duplicate and inconsistent chunks are rejected, while sequence gaps remain valid. Overlay, session, and event behavior is unchanged: those lanes are not subject to this full/lifecycle assembly path, and `event_batch` remains transient presentation delivery without baseline or chunk metadata.

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
world_full
world_delta
ship_delta
asteroid_delta
bullet_delta
ships_lifecycle
asteroids_lifecycle
bullets_lifecycle
overlay_full
overlay_delta
session_full
session_delta
event_batch
debug_status
debug_shape_catalog
```

Room snapshots are sent after room lifecycle changes such as create, join, ready, start, single-player start, return to lobby, leave, and disconnect broadcasts.

Telemetry pong is a same-session diagnostic response. It does not require room membership and does not mutate gameplay lane state.

Debug status and debug shape catalog packets are devtools-only outputs gated by devtools availability and room/gameplay lane state.

### 9. Client routes inbound packets

The client decodes raw WebSocket text for session/control/auth/lobby/signaling packets and raw WebRTC DataChannel text for active realtime gameplay packets into packet dictionaries, classifies packets by generated packet type constants, and emits typed networking signals.

Current client inbound routes include:

```text
authenticate_result
room_snapshot
room_state_changed
room_error
world_full
world_delta
asteroid_delta
bullet_delta
asteroids_lifecycle
bullets_lifecycle
overlay_full
overlay_delta
session_full
session_delta
event_batch
debug_shape_catalog
debug_status
player_pause_state
telemetry_pong
unknown packet fallback
```

Room packets route into room session handling.

Gameplay state routes into `RealtimePacketPipeline`, which expands and validates recognized packets before applying the match-identity boundary established by the authoritative room snapshot. Before an active match exists, non-empty `match_id` packets are buffered by match without lane or presentation mutation; missing IDs are rejected. An authoritative `InGame` snapshot resets gameplay presentation/composition, calls `begin_realtime_match(match_id)`, clears all pending match buckets, and replays only packets for that match through normal lane routing. Once active, mismatched IDs are rejected. Room state reaching `InGame` separately activates client gameplay input, player-pause forwarding, and `PresentationBridge` scheduling; required world/overlay/session readiness gates presentation fanout, not the active protocol routing path.

The cross-system lifecycle correctness flow is:

```text
server lifecycle candidate carries its required world baseline dependency
-> world and lifecycle packets travel on separate WebRTC DataChannels
-> arrival order may differ
-> client waits when the referenced world baseline is not active
-> lifecycle create/delete applies only after matching world state is active
```

`ships_lifecycle`, `asteroids_lifecycle`, and `bullets_lifecycle` use strict independent lane-local sequences. WebSocket room snapshots and WebRTC gameplay packets have no cross-transport ordering guarantee. Recognized packets with non-empty match IDs that arrive before protocol match activation are buffered by match; activation replays only the authoritative match and discards unrelated buckets. After match activation, valid lifecycle packets whose referenced world baseline is not yet active are queued for matching world-baseline activation; invalid or stale lifecycle packets are rejected. Reliable/ordered delivery is per DataChannel and does not order `sr.world` against any lifecycle channel or establish lifecycle-to-hot-lane ordering. Clients must tolerate ship, asteroid, and bullet hot updates arriving before lifecycle create packets and after lifecycle delete packets.

Telemetry pong routes to telemetry consumers and does not pass through normal gameplay lane state application.

### 10. Client applies presentation consequences

After inbound routing, owning client flows apply presentation state.

Examples:

```text
room snapshots update room-session and lobby read models
gameplay lane packets update gameplay runtime and world presentation
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
world_full
world_delta
asteroids_lifecycle
bullets_lifecycle
asteroid_delta
bullet_delta
overlay_full
overlay_delta
session_full
session_delta
event_batch
player_pause_state
telemetry_pong
debug_status
debug_shape_catalog
```

These outputs are authoritative readback or diagnostic presentation packets. The client consumes them through the networking dispatcher and routes them to session, gameplay, lobby, devtools, telemetry, or UI flows.

### Durable outputs

The realtime flow does not directly persist player progress.

When a match reaches authoritative completion, the game server can report resolved match facts through the match-result reporting boundary into player-data. That is a separate persistence flow. The realtime packet shown to clients is presentation-safe room/gameplay lane state, not the durable storage contract.

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

The client sends telemetry pings only when the telemetry flow requests them. The server replies to the same WebSocket session with timing fields. Telemetry does not require room membership and does not mutate gameplay lane state.

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
* record/entity-level prioritization, deeper packet-budget behavior beyond current candidate-level send-plan selection, binary/bit-packed representation, protobuf/custom binary representation, or future transport evolution beyond the current WebSocket control/signaling path plus ordered/reliable lanes for `sr.world`, `sr.overlay`, `sr.session`, `sr.event`, `sr.ships.lifecycle`, `sr.asteroids.lifecycle`, and `sr.bullets.lifecycle`, and unordered/unreliable hot-update lanes for `sr.ships`, `sr.asteroids`, and `sr.bullets`

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
* [Realtime WebRTC Gameplay Transport](../../protocol/realtime-webrtc-gameplay-transport.md)

## Notes

Client input is sent to the server, the server advances simulation, and clients render received state. That rule remains current.

WebSocket connection, room membership, and active gameplay participation are separate states. The current implementation still depends on that separation.

Lane-native packets are current active realtime behavior. World, ship, asteroid, bullet, overlay, and session state lanes currently use deltas, numeric wire quantization, sparse delta omission, and compact JSON aliases. Regular ship, asteroid, and bullet movement updates are split into dedicated hot movement lane packets instead of remaining in `sr.world`. `event_batch` remains compact sparse quantized presentation-event delivery. The server now emits compact tuple wire shape for selected hot records, and the client accepts distinct valid chunk indices with consistent `chunk_count` values for each hot `ship_delta`, `asteroid_delta`, and `bullet_delta` lane sequence while rejecting duplicate indices, inconsistent or malformed metadata, and lower sequences; gaps remain valid and tracking is independent per lane. The client expands those tuples before appliers run. Remaining future work includes record/entity-level prioritization, deeper packet-budget behavior beyond current candidate-level send-plan selection and current chunker-owned hot-lane hard-size guarding, binary/bit-packed representation, protobuf/custom binary representation, and transport evolution beyond the current WebSocket control/signaling path plus ordered/reliable lanes for `sr.world`, `sr.overlay`, `sr.session`, `sr.event`, `sr.ships.lifecycle`, `sr.asteroids.lifecycle`, and `sr.bullets.lifecycle`, and unordered/unreliable hot-update lanes for `sr.ships`, `sr.asteroids`, and `sr.bullets`.

Single-player and multiplayer can currently use the same local `/ws` route. That does not collapse their authority model. The boot packet, session mode, auth/admission rule, room joinability, and player-data identity context distinguish the flows.
