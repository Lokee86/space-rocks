## Gameplay Packets

Parent index: [Protocol](./!INDEX.md)

## Purpose

This document describes the current gameplay realtime packet protocol between the Godot client and the Go game server.

It covers client-originated gameplay requests, server-originated gameplay state, pause-state output, embedded presentation events, packet authority, source-of-truth files, runtime routing, validation, and the implementation paths that consume the gameplay packet contract.

## Overview

Gameplay packets are the realtime WebSocket messages used after a client is connected to the game server and, for gameplay mutation, attached to an active game player.

The protocol is server-authoritative:

```text
client sends input or request intent
-> game-server inbound routing classifies the packet
-> active room/game instance receives the packet
-> game simulation mutates authoritative state
-> outbound networking projects state or pause output
-> client receives and applies server-owned state
```

The client owns packet emission, local input collection, target-selection intent, viewport config reporting, and presentation after receiving server state. The game server owns acceptance, validation, simulation mutation, target state, pause state, respawn validity, scoring, lives, damage, pickups, spawning, state packet projection, and presentation event production.

The gameplay packet protocol does not include lobby room snapshots, room membership requests, auth results, devtools commands, player-data runtime packets, or HTTP contracts. Those packet families may share the same WebSocket transport or packet schema pipeline, but they are separate protocol surfaces.

## Participating systems

```text
client/
```

Collects local gameplay input, target intent, viewport config, respawn intent, pause intent, and applies inbound gameplay state to presentation systems.

```text
services/game-server/internal/networking/
```

Owns WebSocket read/write loops, session context, packet-family routing, outbound state write timing, and pause-state packet enqueueing.

```text
services/game-server/internal/rooms/
```

Owns room lifecycle, active game instance access, room state, room game-over state, and room membership context used by gameplay routing.

```text
services/game-server/internal/game/
```

Owns authoritative gameplay mutation, state packet projection, player session state, player lifecycle classification, input application, respawn, pause, target selection, event queueing, pickups, combat, scoring, spawning, and runtime state.

```text
shared/packets/
```

Owns the editable packet schema sources and generated packet output routing.

```text
tools/data_sync/
```

Generates Go and GDScript packet outputs from the shared packet schema source files.

## Protocol authority

Packet shape authority lives in:

```text
shared/packets/gameplay.toml
shared/packets/outputs.toml
```

Generated packet code is output only and should not be edited by hand.

Runtime behavior authority is split:

```text
client outbound flow
= builds and sends generated gameplay packet dictionaries

game-server inbound routing
= classifies packet type and forwards to the active authoritative game instance

game-server game simulation
= accepts, rejects, mutates, or ignores gameplay requests

game-server state projection
= builds authoritative state packets for each player

client gameplay runtime
= normalizes and applies server-owned state to presentation
```

The client does not own authoritative confirmation. A client request is confirmed only when reflected by server output such as `state`, `player_pause_state`, room snapshots, or presentation events.

## Packet source and generated outputs

Primary source files:

```text
shared/packets/gameplay.toml
shared/packets/outputs.toml
```

Current generated outputs used by this protocol:

```text
services/game-server/internal/game/packets.go
services/game-server/internal/game/runtime/packets_generated.go
client/scripts/generated/networking/packets/packets.gd
```

`shared/packets/gameplay.toml` defines gameplay packet structs, packet type values, and client packet builders.

`shared/packets/outputs.toml` decides which packet types, structs, and builders are emitted to the game-server and client generated files.

## Client-to-server gameplay packets

### `input`

`input` carries the local player's current movement and fire input state.

Current payload shape:

```text
type = "input"

input.forward
input.back
input.right
input.left
input.primary_fire
input.secondary_fire
```

Client source:

```text
GameplayInputFlow
-> player.get_input_packet()
-> ClientConnectionService.send_input_packet(packet)
-> NetworkClient.send_raw_packet(packet)
```

Server path:

```text
inbound.HandleGameplayPacket
-> Game.HandlePacket(playerID, packet)
-> player.SetInput(packet.Input)
```

The server applies input only when the player can receive input. Pending-despawn players and suspended players cannot receive input.

### `client_config`

`client_config` carries the client viewport dimensions used by the server-side camera view and session config.

Current payload shape:

```text
type = "client_config"

config.visible_world_width
config.visible_world_height
```

Client source:

```text
ClientViewportConfigFlow
-> Packets.client_config_packet(width, height)
-> ClientConnectionService.send_packet(packet)
```

Server path:

```text
inbound.HandleGameplayPacket
-> Game.HandlePacket(playerID, packet)
-> player session config update
-> camera view config update
-> active player config update when an active player exists
```

The server ignores non-positive viewport dimensions.

### `respawn`

`respawn` requests a player respawn.

Current payload shape:

```text
type = "respawn"
```

Server path:

```text
inbound.HandleGameplayPacket
-> Game.HandlePacket(playerID, packet)
-> game.respawnPlayer(playerID)
```

The packet is a request only. Respawn validity, spawn position, remaining lives, cooldown state, and active ship creation are server-owned.

### `pause_request`

`pause_request` toggles the local player's pause state.

Current payload shape:

```text
type = "pause_request"
```

Server path:

```text
inbound.HandleGameplayPacket
-> Game.HandlePacket(playerID, packet)
-> game.togglePlayerPaused(playerID)
-> session.EnqueuePlayerPauseState()
```

The server responds on the same session with a `player_pause_state` packet when a pause-state packet can be built.

### `set_target_player_request`

`set_target_player_request` requests target selection by target identity.

Current payload shape:

```text
type = "set_target_player_request"
target_kind
target_id
```

Server path:

```text
inbound.HandleGameplayPacket
-> Game.SetPlayerTarget(currentGamePlayerID, packet.TargetID)
```

Current implementation forwards `target_id` into the player-target path. Normal targeting identity is moving toward `target_kind` plus `target_id`, but the current set-player route still calls the player-target API.

### `select_target_at_position_request`

`select_target_at_position_request` requests target selection at a client-visible world position, optionally carrying target identity context.

Current payload shape:

```text
type = "select_target_at_position_request"
x
y
target_kind
target_id
```

Server path:

```text
inbound.HandleGameplayPacket
-> Game.SelectTargetAtPosition(
     currentGamePlayerID,
     packet.X,
     packet.Y,
     TargetRef{Kind: packet.TargetKind, ID: packet.TargetID},
   )
```

The client sends candidate intent. The server validates whether the selected position and target reference resolve to an authoritative target.

### `clear_target_request`

`clear_target_request` requests clearing the local player's current target.

Current payload shape:

```text
type = "clear_target_request"
```

Server path:

```text
inbound.HandleGameplayPacket
-> Game.ClearTarget(currentGamePlayerID)
```

## Server-to-client gameplay packets

### `state`

`state` is the main authoritative gameplay state packet.

Current packet shape:

```text
type
self_id
lives
players
player_sessions
player_lifecycle
bullets
asteroids
pickups
total_asteroids
events
server_sent_msec
```

Server path:

```text
networking write tick
-> outbound.BuildGameplayPresentationStateResponse
-> room.GameInstance().StatePacket(playerID)
-> Game.statePacket(playerID)
-> packetcodec.Encode
-> WebSocket write
```

The state packet is built per receiving player. `self_id`, top-level `lives`, and `events` are receiver-specific. Most world maps are shared authoritative read models for the current game instance.

The server sends gameplay presentation state only when:

```text
session.currentGamePlayerID is not empty
room exists
room has a game instance
room state is InGame or GameOver
```

The packet is eligible while the room is `GameOver` so match-over presentation can still receive final state/event information.

### `player_pause_state`

`player_pause_state` is a same-session pause-state response produced after a valid pause request.

Current packet shape:

```text
type = "player_pause_state"
player_id
paused
```

Server path:

```text
pause_request
-> Game.HandlePacket
-> session.EnqueuePlayerPauseState
-> Game.PlayerPauseStatePacket(playerID)
-> packetcodec.Encode
-> session outbound queue
```

The packet reports the server-owned pause state for the active player on that session. It is not a client-side confirmation generated by the UI.

### Embedded `events`

Gameplay presentation events are embedded in `StatePacket.events`, and stable event identity is required before `event_batch` cutover. Future `event_batch` delivery must not drain events on projection, scheduling, shadow, or encode; it drains only after active socket write/enqueue success.

Current event payload shape:

```text
type
player_id
lives
respawn_delay
x
y
pickup_id
pickup_type
source_type
source_id
table_id
lives_after
effect_type
amount
```

Current event types include:

```text
bullet_blast
ship_death
pickup_dropped
pickup_collected
pickup_effect_applied
pickup_expired
radial_effect_started
damage_applied
damage_over_time_started
damage_over_time_tick
```

The server stores pending packet-facing events in:

```text
pendingPresentationEvents[playerID]
```

Event IDs are assigned when pending presentation events enter the pending event queue, not during projection. `Game.StatePacket(playerID)` copies the requesting player's pending events into `StatePacket.events`; shadow projection may peek and copy pending events, but it must never drain them. Active event drain happens only after the active send path selects the records and they are successfully encoded for active send.

The event lane is transient presentation state. It is not a durable event log, not a guaranteed-delivery queue, and not the domain event source of truth.

## Diagnostic packets defined with gameplay schema

`telemetry_ping` and `telemetry_pong` are defined through the gameplay packet schema, but their runtime role is diagnostic transport telemetry, not gameplay mutation.

Current telemetry request shape:

```text
type = "telemetry_ping"
sequence
client_sent_msec
```

Current telemetry response shape:

```text
type = "telemetry_pong"
sequence
client_sent_msec
server_received_msec
server_sent_msec
```

Telemetry routing does not require room membership, does not require an active game player, and does not mutate gameplay state.

## State packet field roles

### `self_id`

`self_id` is the receiving player's active game-player identity.

### `lives`

Top-level `lives` is the receiving player's current session lives. It is a convenience receiver-specific projection.

Durable match-local lives also appear in `player_sessions`.

### `players`

`players` is active ship/avatar state only.

Current `ShipState` fields include:

```text
id
ship_type
x
y
rotation
health
shields
thrusting
target_kind
target_id
primary_weapon_id
primary_ammo_policy
primary_cooldown_remaining
primary_ammo_remaining
secondary_weapon_id
secondary_ammo_policy
secondary_cooldown_remaining
secondary_ammo_remaining
```

Pending-respawn and eliminated players can be absent from `players`.

### `player_sessions`

`player_sessions` is match-local durable player session state.

Current `PlayerSessionState` fields include:

```text
id
ship_type
score
lives
respawn_cooldown
primary_weapon_id
primary_ammo_policy
secondary_weapon_id
secondary_ammo_policy
spawn_x
spawn_y
```

This is not account/profile/player-data persistence.

### `player_lifecycle`

`player_lifecycle` maps player ID to lifecycle status.

Current statuses:

```text
active
pending_respawn
eliminated
```

Clients must not infer lifecycle from `players` alone. `players` is active avatar state; `player_lifecycle` is the lifecycle read model.

### `bullets`

`bullets` is active projectile state.

Current `BulletState` fields include:

```text
id
owner_id
x
y
rotation
weapon_id
projectile_type
```

### `asteroids`

`asteroids` is active asteroid state.

Current `AsteroidState` fields include:

```text
id
x
y
size
health
scale
variant
```

`variant` is the runtime asteroid variant index.

### `pickups`

`pickups` is active pickup state.

Current `PickupState` fields include:

```text
id
type
pickup_class
x
y
health
age_seconds
lifespan_seconds
```

`pickup_class` selects the generic client scene family. `type` is the gameplay identity and client badge/icon selector.

Scene paths are client-owned and are not sent in gameplay packets.

### `total_asteroids`

`total_asteroids` is the cumulative spawned asteroid count from the current game spawner. It is not the active asteroid map length.

### `server_sent_msec`

`server_sent_msec` is stamped by outbound networking immediately before packet encoding.

It is not set by `Game.statePacket`.

## Message flow

### Outbound client request flow

```text
client gameplay/input/UI caller
-> generated Packets helper or packet-family wrapper
-> ClientConnectionService
-> ClientPacketSender
-> NetworkClient.send_raw_packet
-> PacketCodec.encode
-> WebSocketPeer.send_text
-> game-server WebSocket read loop
```

The client send path is best-effort and non-queued. If the socket is not open, if the packet sender is unavailable, or if JSON encoding fails, the packet is not sent.

### Inbound game-server request flow

```text
webSocketSession read loop
-> inbound.DecodeClientPacketEnvelope
-> inbound.RouteClientPacket
-> devtools routing first
-> packetcodec.Decode into game.ClientPacket
-> auth routing
-> telemetry routing
-> lobby routing
-> gameplay routing
-> current room game instance
```

Gameplay routing receives only packets not consumed by earlier families.

Current gameplay route table:

```text
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

### Server state output flow

```text
WebSocket write tick
-> outbound.CanSendGameplayPresentationState(room)
-> outbound.BuildGameplayPresentationStateResponse(room, playerID, roomID, remoteAddr)
-> Game.StatePacket(playerID)
-> stamp server_sent_msec
-> packetcodec.Encode
-> WebSocket write
```

### Client inbound state flow

```text
NetworkClient.poll
-> PacketCodec.decode
-> NetworkClient.packet_received
-> ClientConnectionService
-> ServerPacketDispatcher
-> gameplay_state_received
-> SessionNetworkController
-> GameplaySessionController
-> GameplayStateFlow
-> GameplayStatePacketReader
-> GameplayStateApplyFlow
-> HUD, world sync, alive/respawn restore, event presentation, devtools state
```

The client normalizes gameplay state before applying it. Broad presentation code should consume normalized gameplay state rather than scattering raw packet field access. The old `has_received_gameplay_state` / `has_received_state` readiness concept becomes required lane baseline sync, and the docs should refer to both `GameplayStateFlow.has_received_gameplay_state` and `GameplayShellFlow.has_received_state` when naming client readiness.

## Session-state requirements

Different gameplay packet paths require different session context.

```text
input, respawn, client_config
require current room and current game player to apply
are consumed without applying when room/player is missing

target request packets
require current room and current game player
return unhandled when room/player is missing

pause_request
requires current room and current game player
routes to Game.HandlePacket and then enqueues player_pause_state
```

The WebSocket connection itself does not imply room membership.

Room membership does not imply an active game player.

`currentGamePlayerID` is networking-owned active gameplay routing state for the current session.

## Service responsibilities

### Client

The client owns:

```text
input collection
outbound gameplay packet construction
outbound packet send attempts
target-selection request construction
viewport config reporting
respawn and pause request emission
inbound packet classification after decode
gameplay packet acceptance gating in GameplaySessionController
state packet normalization
world sync and presentation fanout
HUD, audio, effects, match-end presentation, and devtools readouts
```

The client does not own:

```text
gameplay authority
state packet contents
target validation
respawn validity
pause authority
score/lives authority
damage/collision outcomes
pickup authority
asteroid spawning or variants authority
server event production
durable match result persistence
```

### Game-server networking

Game-server networking owns:

```text
WebSocket read/write loops
packet envelope decode
packet-family routing order
server packet JSON encode/decode handoff
session current room and current game player context
gameplay adapter handoff
pause-state packet enqueueing
outbound state write timing
server_sent_msec stamping
packet-size diagnostics
```

Game-server networking does not own:

```text
simulation rules
state packet schema source
state packet projection internals
target validation
respawn mechanics
pause implementation
event production
client presentation
```

### Game-server rooms

Rooms own:

```text
room lifecycle
room state
room game instance access
room GameOver state
membership and active-player context
```

Rooms do not own gameplay packet schema or client presentation.

### Game-server simulation

Game-server simulation owns:

```text
input application
client config application to session/camera/player state
respawn behavior
pause mutation
target selection and clearing
state packet projection
player session state
active avatar state
lifecycle classification
projectile, asteroid, pickup, and event projection
presentation event queueing
```

Simulation does not own WebSocket transport, raw JSON encoding, generated packet pipeline behavior, or client rendering.

### Data-sync and shared schema

The packet schema pipeline owns:

```text
packet type wire strings
packet structs and generated field names
selected generated Go outputs
selected generated GDScript constants
selected generated GDScript builders
generated output routing
packet schema validation
packet drift checks
```

It does not own runtime gameplay semantics.

## Compatibility expectations

Gameplay packet changes must preserve these expectations:

```text
packet type strings are stable wire identifiers
schema edits happen in shared/packets/gameplay.toml
generated outputs are refreshed through data-sync
server and client generated outputs are kept in the same change
service tests cover runtime behavior affected by the packet change
client packet readers are updated when state packet fields change
state packet wrappers or adapters deliberately preserve new state fields
client presentation does not infer lifecycle from active player presence
room/lobby snapshot data is not added to gameplay state unless it is truly gameplay state
match result data does not belong in the ticked gameplay state packet
devtools command packets stay separate from normal gameplay packets
```

Adding a new gameplay packet requires both schema work and runtime routing work. Adding only a generated packet type does not make the packet meaningful.

## Validation and testing

Packet schema validation:

```bash
data-sync -validate -packets
data-sync -diff -packets -go -gds
data-sync -push -packets -go -gds
data-sync -check -packets -go -gds
```

Game-server verification for gameplay routing and simulation behavior:

```bash
cd services/game-server && go test -buildvcs=false ./internal/networking/...
cd services/game-server && go test -buildvcs=false ./internal/networking/inbound/...
cd services/game-server && go test -buildvcs=false ./internal/networking/outbound/...
cd services/game-server && go test -buildvcs=false ./internal/game/...
cd services/game-server && go test -buildvcs=false ./tests/game/...
```

Client verification for packet decode, state readers, state application, input, target requests, and session routing:

```bash
godot --headless --path client -s res://addons/gut/gut_cmdln.gd -gdir=res://tests/unit -ginclude_subdirs -gexit
```

Relevant current tests include:

```text
services/game-server/internal/networking/gameplay_packets_test.go
services/game-server/internal/networking/outbound/gameplay_presentation_test.go
services/game-server/internal/game/events_test.go
services/game-server/tests/game/state_packet_lifecycle_test.go
services/game-server/tests/game/packets_generated_test.go
client/tests/unit/test_packet_codec.gd
client/tests/unit/test_gameplay_state_packet_reader.gd
client/tests/unit/test_gameplay_state_apply_flow.gd
client/tests/unit/test_gameplay_input_context.gd
client/tests/unit/test_target_request_flow.gd
client/tests/unit/test_gameplay_session_controller.gd
client/tests/unit/test_session_network_controller.gd
```

## Code map

### Packet sources and generated outputs

```text
shared/packets/gameplay.toml
shared/packets/outputs.toml
tools/data_sync/
services/game-server/internal/game/packets.go
services/game-server/internal/game/runtime/packets_generated.go
client/scripts/generated/networking/packets/packets.gd
```

### Client outbound packet construction and send path

```text
client/scripts/networking/client_connection_service.gd
client/scripts/networking/network_client.gd
client/scripts/networking/outbound/client_packet_sender.gd
client/scripts/networking/outbound/gameplay_client_packets.gd
client/scripts/networking/packets/packet_codec.gd
client/scripts/entities/player.gd
client/scripts/gameplay/input/gameplay_input_flow.gd
client/scripts/gameplay/targeting/target_request_flow.gd
client/scripts/config/client_viewport_config_flow.gd
```

### Client inbound packet routing and gameplay application

```text
client/scripts/networking/inbound/server_packet_dispatcher.gd
client/scripts/networking/inbound/server_packet_router.gd
client/scripts/session/session_network_controller.gd
client/scripts/session/gameplay_session_controller.gd
client/scripts/gameplay/state/gameplay_state_flow.gd
client/scripts/gameplay/state/gameplay_state_packet_reader.gd
client/scripts/gameplay/state/gameplay_state_apply_flow.gd
client/scripts/gameplay/runtime/gameplay_world_state_apply_flow.gd
client/scripts/world/world_sync.gd
client/scripts/gameplay/events/
client/scripts/gameplay/effects/
```

### Game-server inbound routing and gameplay mutation

```text
services/game-server/internal/networking/websocket_read.go
services/game-server/internal/networking/client_packet_router.go
services/game-server/internal/networking/inbound/router.go
services/game-server/internal/networking/inbound/gameplay.go
services/game-server/internal/networking/player_pause_state.go
services/game-server/internal/game/input.go
services/game-server/internal/game/pause.go
services/game-server/internal/game/player_targeting.go
services/game-server/internal/game/targeting.go
services/game-server/internal/game/respawn.go
```

### Game-server state and event output

```text
services/game-server/internal/networking/websocket_write.go
services/game-server/internal/networking/outbound/gameplay_presentation.go
services/game-server/internal/networking/outbound/gameplay_state_metrics.go
services/game-server/internal/game/state_packet.go
services/game-server/internal/game/player_session_state.go
services/game-server/internal/game/events.go
services/game-server/internal/game/events/
services/game-server/internal/game/pickups.go
services/game-server/internal/game/runtime/
services/game-server/internal/protocol/packetcodec/codec.go
```

### Important non-ownership boundaries

```text
services/game-server/internal/rooms/
```

Owns room lifecycle and game instance access, not gameplay packet shape.

```text
services/game-server/internal/devtools/
```

Owns devtools command packet handling, not normal gameplay packet routing.

```text
shared/packets/debug.toml
```

Owns devtools packet schema, not normal gameplay packet schema.

```text
shared/packets/lobby.toml
```

Owns lobby/auth/room packet schema, not gameplay state projection.

```text
shared/contracts/http/openapi.yaml
```

Owns HTTP request/response contracts, not realtime gameplay packets.

```text
services/player-data/
```

Owns player-data routing and persistence, not gameplay state packets.

## Related docs

* [Protocol](./!INDEX.md)
* [Packet Schemas](../data/packet-schemas.md)
* [Data Sync and SSoT Pipeline](../data/data-sync-and-ssot-pipeline.md)
* [Source of Truth Map](../data/source-of-truth-map.md)
* [Game Server](../services/game-server/!INDEX.md)
* [Game Server Networking](../services/game-server/networking/!INDEX.md)
* [Inbound Packet Routing](../services/game-server/networking/inbound-packet-routing.md)
* [Gameplay Network Adapter](../services/game-server/networking/gameplay-network-adapter.md)
* [Outbound Message Flow](../services/game-server/networking/outbound-message-flow.md)
* [State Packet Projection](../services/game-server/simulation/runtime/state-packet-projection.md)
* [Presentation Event Queue](../services/game-server/simulation/runtime/presentation-event-queue.md)
* [Player Input Routing](../services/game-server/simulation/players/player-input-routing.md)
* [Player Pause And Suspension](../services/game-server/simulation/players/player-pause-and-suspension.md)
* [Player Respawn](../services/game-server/simulation/players/player-respawn.md)
* [Client](../services/client/!INDEX.md)
* [Client Networking Flow](../services/client/networking-flow/!INDEX.md)
* [Outbound Packet Sending](../services/client/networking-flow/outbound-packet-sending.md)
* [Client Inbound Packet Routing](../services/client/networking-flow/inbound-packet-routing.md)
* [Gameplay State Application](../services/client/gameplay-runtime/gameplay-state-application.md)
* [World Sync](../services/client/world-sync/!INDEX.md)
* [Gameplay Events And Effects](../services/client/gameplay-event-presentation/gameplay-events-and-effects.md)
* [Realtime WebSocket Protocol](realtime-websocket-protocol.md)
* [Lobby Packets](lobby-packets.md)

## Notes

`telemetry_ping` and `telemetry_pong` live in the gameplay packet schema today, but their runtime behavior is diagnostic and transport-adjacent. They should not be treated as gameplay mutation packets.

`StatePacket.events` carries packet-facing presentation events. It is not a domain event queue and does not guarantee delivery across disconnects, encode failures, write failures, or late joins.

The current server inbound route silently drops packets that decode but are not consumed by a packet-family handler. That is current behavior, not a compatibility promise.

New state packet fields must be copied through any wrapper, adapter, debug, or presentation path that re-emits state. Do not assume a new `StatePacket` field automatically survives every outbound path.


