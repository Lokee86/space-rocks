# Inbound Packet Routing

Parent index: [Networking Flow](./!INDEX.md)

## Purpose

This document describes the current client inbound packet routing path.

It covers how decoded server packet dictionaries move from the client WebSocket transport and WebRTC DataChannel transport into packet classification, dispatcher signals, connection-service signals, and downstream session, room, gameplay, auth, devtools, and telemetry consumers.

## Overview

Inbound packet routing begins after the WebSocket transport or WebRTC DataChannel transport has already decoded raw text into a packet dictionary.
WebSocket remains the client route for session, control, room, lobby, auth, telemetry, signaling, and queued non-gameplay packets. Lane-specific WebRTC gameplay DataChannels are the route for active realtime gameplay packets and diagnostic smoke packets: `sr.world`, `sr.overlay`, `sr.session`, `sr.event`, `sr.asteroids.lifecycle`, and `sr.bullets.lifecycle` are ordered/reliable lifecycle channels, while `sr.asteroids` and `sr.bullets` are unordered/unreliable hot-update lanes. WebRTC connectivity is established by ICE, not by a WebRTC URL, and deployment must ensure the advertised ICE address can reach the game server directly.

`NetworkClient` owns raw WebSocket polling, text receive, JSON decode, envelope validation, and `packet_received` emission for WebSocket packets. `WebRTCTransport` owns DataChannel text receive, packet decode, and `packet_received` emission for WebRTC packets. After those transport signals fire, `ServerPacketRouter` identifies packet types, `ServerPacketDispatcher` emits typed packet signals, `ClientConnectionService` coordinates the public networking handoff, and `RealtimePacketPipeline` owns realtime gameplay packet expansion, validation, application, readiness, reset, the active `RealtimeRouter`, and the post-application notification that feeds `PresentationBridge`.

Compact `event_batch` packets are expanded by the client compact packet expansion layer before event appliers receive them, so downstream gameplay code sees readable long-key event dictionaries rather than compact aliases. Nested `ev` or `events` entries are expanded as part of that same transport step, and event dedupe still keys off `event_id` after expansion. Runtime wire packets keep compact aliases on the wire; domain logs may still show raw x/y before projection.

Current flow:

```text
NetworkClient.packet_received(packet)
-> ClientConnectionService._on_packet_received(packet)
-> ServerPacketDispatcher.dispatch(packet)
-> ServerPacketRouter packet-type checks
-> ClientConnectionService re-emits the typed WebSocket service signal or handles lane packets

WebRTCTransport receives DataChannel text
-> PacketCodec.decode expands compact aliases and validates the packet envelope
-> WebRTCTransport.packet_received(packet)
-> ClientConnectionService._handle_webrtc_transport_packet(packet)
-> ServerPacketDispatcher.dispatch(packet)
-> ServerPacketDispatcher emits a typed dispatcher signal
-> ClientConnectionService._route_gameplay_packet(packet)
-> RealtimePacketPipeline.apply_packet(packet)
-> RealtimeRouter.route_lane_packet(packet)
-> RealtimePacketPipeline.gameplay_packet_applied(packet)
-> PresentationBridge.handle_gameplay_packet(packet)
```

The routing path is signal-based and lane-aware. It does not mutate server authority, does not parse payload-specific gameplay data, and does not apply presentation state directly. Its job is to classify packet family by generated packet type constants, forward lane packets through the realtime router, and hand the dictionary to the owning client subsystem. Inbound realtime lane packets may already contain quantized numeric wire values. The client routes them by lane and packet family, does not own authoritative quantization decisions, and uses `client/scripts/protocol/realtime/realtime_quantize.gd` when it needs to decode quantized realtime lane values. Lifecycle defines existence. Hot lanes update known entities only. Hot asteroid and bullet packets are routed on unordered/unreliable lanes. The client rejects lower sequence values so late packets cannot roll positions backward. Same-sequence packets are valid for chunked `asteroid_delta` or `bullet_delta` output and may apply independently. Sequence gaps are valid because hot packets can be dropped.

## Code root

* `client/scripts/networking/`
* `client/scripts/session/`

## Responsibilities

* Receive decoded packet dictionaries from `NetworkClient` and `WebRTCTransport`.
* Dispatch inbound packets by generated packet type constants.
* Re-emit typed packet signals from the connection service.
* Track websocket auth result state from `authenticate_result` packets.
* Route room packets to room session handling.
* Route gameplay packets to gameplay session handling.
* Route debug shape catalog and debug status packets to gameplay/devtools handling.
* Route player pause packets to gameplay session handling.
* Route telemetry pong packets to telemetry consumers.
* Emit an unknown-packet signal for recognized envelopes with unhandled packet types.
* Keep packet-family routing separate from raw WebSocket transport and payload-specific packet readers.

## Does not own

* Raw WebSocket connection lifecycle.
* Raw WebRTC peer and DataChannel lifecycle.
* WebSocket polling.
* Packet JSON parsing.
* Packet encode/decode result types.
* Packet schema source-of-truth files.
* Generated packet constant ownership.
* Server packet production.
* Server room authority.
* Server gameplay authority.
* Gameplay state application.
* Payload-specific packet reader behavior.
* World sync or entity rendering.
* HUD or menu presentation.
* Outbound packet construction.
* Outbound send timing.
* Auth token verification.
* Rails account identity.
* Persistent player data.
* Devtools command authority.

## Domain roles

### Decoded packet handoff

`NetworkClient` receives raw WebSocket text and hands it to `PacketCodec.decode`. `PacketCodec.decode` accepts compact realtime aliases by expanding them to readable long-key dictionaries before envelope validation and dispatch; see [Realtime Compact Wire Mapping](../../game-server/networking/realtime-compact-wire-mapping.md). Legacy long-key packets remain accepted, and packets with neither `type` nor compact `t` still fail envelope validation. For `event_batch`, that expansion happens before `event_batch_applier.gd` or gameplay event appliers see the packet payload, and compact event aliases stay a transport detail rather than leaking into gameplay presentation code.

`WebRTCTransport` receives DataChannel text, calls `PacketCodec.decode` before emitting `packet_received`, and the decoded dictionary then flows through `ClientConnectionService._handle_webrtc_transport_packet(packet)` for diagnostic smoke handling or gameplay dispatch.

`ClientConnectionService._on_packet_received(packet)` receives the WebSocket dictionary and delegates to `ServerPacketDispatcher.dispatch(packet)`.

Inbound routing therefore assumes:

```text
packet is a Dictionary
packet has a non-empty String type field
payload envelope validation has already happened
```

Payload-specific validation still belongs to later packet readers or consumers.

### Packet type classification

`ServerPacketRouter` is the pure classification helper.

It reads the packet type through:

```gdscript
packet.get(Packets.FIELD_TYPE, "")
```

and compares the value against generated packet type constants from:

```text
client/scripts/generated/networking/packets/packets.gd
```

Current classified inbound packet types are:

```text
room_snapshot
authenticate_result
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
resync_request
resync_required
debug_shape_catalog
debug_status
player_pause_state
telemetry_pong
```

### Dispatcher signal fanout

`ServerPacketDispatcher` owns the ordered classification chain and typed signal emission.

Current dispatcher outputs are:

```text
room_snapshot_received(packet)
authenticate_result_received(packet)
room_state_changed(packet)
room_error_received(packet)
world_full_received(packet)
world_delta_received(packet)
asteroid_delta_received(packet)
bullet_delta_received(packet)
asteroids_lifecycle_received(packet)
bullets_lifecycle_received(packet)
overlay_full_received(packet)
overlay_delta_received(packet)
session_full_received(packet)
session_delta_received(packet)
event_batch_received(packet)
resync_request_received(packet)
resync_required_received(packet)
debug_shape_catalog_received(packet)
debug_status_received(packet)
player_pause_state_received(packet)
telemetry_pong_received(packet)
unknown_packet_received(packet)
```

The dispatcher does not know which application subsystem will consume each signal. It only converts packet-type classification into a signal and does not apply lane state.

### Connection-service signal bridge

`ClientConnectionService` owns the public networking facade, dispatcher wiring, transport coordination, and control-routing compatibility methods. It does not own the active RealtimeRouter or realtime gameplay readiness.

`RealtimePacketPipeline` owns the active RealtimeRouter, compact packet expansion, gameplay packet validation, lane-routing invocation, gameplay readiness, protocol-state reset, and post-application notification.

`RealtimeRouter` owns lane-specific state mutation, baseline tracking, sequence handling, and lane-state storage beneath `RealtimePacketPipeline`.

`ClientConnectionService` connects dispatcher signals to local handlers, then re-emits service-level signals with the same packet dictionary.

Room, auth, debug, player-pause, and telemetry packets are re-emitted through service-level signals so callers can stay on the connection-service facade.

Realtime gameplay packets take the direct realtime path inside `ClientConnectionService`:

```text
ClientConnectionService
-> ServerPacketDispatcher.dispatch(packet)
-> RealtimePacketPipeline.apply_packet(packet)
-> RealtimeRouter
-> gameplay_packet_applied(packet)
-> PresentationBridge.handle_gameplay_packet(packet)
```

`ClientConnectionService` delegates realtime gameplay packets to `RealtimePacketPipeline`, which invokes its owned `RealtimeRouter`.

`ClientConnectionService` exposes the stable `RealtimePacketPipeline` through `get_realtime_packet_pipeline()`. Session consumers use `RealtimePacketPipeline.is_gameplay_ready()` and `RealtimePacketPipeline.get_presentation_state()`. `RealtimeRouter` and `GameplayReadiness` remain pipeline-internal implementation details. No session, presentation, or connection-service consumer may retain or inspect the router directly. `reset_realtime_protocol_state()` remains the connection-level reset entry point when that method still exists.

The connection service routes gameplay packets through the semantic pipeline/application handoff, while `RealtimePacketPipeline.gameplay_packet_applied(packet)` and `PresentationBridge.handle_gameplay_packet(packet)` carry presentation delivery.

ClientConnectionService still emits a structured network diagnostic event when a lane packet is routed for the first time by packet type:

```text
event: lane_packet_routed
level: info
category: network
fields:
  packet_type
  readiness
```

The once-per-packet-type guard remains diagnostic-only. It does not affect routing or lane state.

Under hot-lane stress, multiple `asteroid_delta` or `bullet_delta` packets may arrive for the same lane sequence in one poll window. The routing path should treat those as separate packets and must not coalesce or drop them solely because they share a sequence.

Presentation remains frame-coalesced and is intentionally unchanged in this stage.

### Websocket auth result cache

`ClientConnectionService` handles `authenticate_result` specially because websocket auth state is connection-level state.

On `authenticate_result`, the connection service updates:

```text
websocket_auth_authenticated
websocket_auth_user_id
websocket_auth_display_name
```

and emits:

```text
websocket_auth_result_received(packet)
```

The connection service does not verify the token. It only records the result returned by the game server.

### Session network handoff

`SessionNetworkController` connects to connection-service signals in three groups:

```text
connect_connection_signals()
connect_room_signals()
connect_gameplay_signals()
```

Connection-level signals handle:

```text
connected
closed
packet_parse_failed
unknown_packet_received
websocket_auth_result_received
```

Room signals handle:

```text
room_snapshot_received
room_state_changed
room_error_received
```

Gameplay signals handle:

```text
debug_shape_catalog_received
debug_status_received
player_pause_state_received
```

`SessionNetworkController` is the bridge from networking events into room and gameplay session controllers. It does not classify packet types itself.

For gameplay application, the controller now follows the semantic presentation handoff instead of connecting separately to each lane-specific packet signal.

Lane-specific service signals still exist for consumers, tests, and diagnostics, while normal gameplay session handoff continues after `PresentationBridge.handle_gameplay_packet(packet)`.

### Room packet handoff

Room packets route through `SessionNetworkController` into `RoomSessionController`.

Current handoff:

```text
room_snapshot_received
-> SessionNetworkController._on_room_snapshot_received
-> RoomSessionController.handle_room_snapshot

room_state_changed
-> SessionNetworkController._on_room_state_changed
-> RoomSessionController.handle_room_state_changed

room_error_received
-> SessionNetworkController._on_room_error_received
-> RoomSessionController.handle_room_error
```

`RoomSessionController` owns lobby/session consequences such as applying room snapshots, tracking latest room state, caching match result data from snapshots, showing room errors, and sending client config after room entry when needed.

Inbound packet routing does not own those consequences.

### Gameplay packet handoff

Lane-native realtime gameplay packets route through the completed semantic application path after the realtime router has already applied lane state.

Current handoff:

```text
RealtimePacketPipeline.gameplay_packet_applied(packet)
-> PresentationBridge.handle_gameplay_packet(packet)
-> gameplay composition / runtime presentation flows
```

`SessionNetworkController` remains the bridge from networking events into room and gameplay session controllers. It still preserves the control routing for connection, room, auth, debug, and player-pause packets.

Gameplay packet application continues through gameplay runtime documentation after this point. Inbound routing only delivers the packet.

### Debug packet handoff

Debug shape catalog and debug status packets route through the same gameplay session controller because current devtools presentation is composed inside gameplay session context.

Current handoff:

```text
debug_shape_catalog_received
-> SessionNetworkController._on_debug_shape_catalog_received
-> GameplaySessionController.handle_debug_shape_catalog_packet

debug_status_received
-> SessionNetworkController._on_debug_status_received
-> GameplaySessionController.handle_debug_status_packet
```

Debug packet routing does not grant authority to the client. Server-side devtools authority remains server-owned.

### Player pause packet handoff

Player pause packets route through gameplay session handling:

```text
player_pause_state_received
-> SessionNetworkController._on_player_pause_state_received
-> GameplaySessionController.handle_player_pause_state
```

`GameplaySessionController.handle_player_pause_state` gates pause packets with `accepts_gameplay_packets` before forwarding them to gameplay composition.

### Telemetry packet handoff

`telemetry_pong` packets are routed by the dispatcher and re-emitted by the connection service.

`WorldTelemetryContext` connects directly to:

```text
ClientConnectionService.telemetry_pong_received
```

and applies the pong packet to network telemetry metrics.

Telemetry pong handling is diagnostic. It does not require room membership, does not mutate gameplay state, and does not route through normal gameplay packet application.

### Unknown packet fallback

If the packet envelope is valid but no current router predicate matches the packet type, the dispatcher emits:

```text
unknown_packet_received(packet)
```

`SessionNetworkController` currently logs the unknown-packet event through its configured logger.

Unknown packets are not applied to gameplay, room, auth, or telemetry state.

### Lifecycle routing note

Lifecycle packets use the same routing path as other gameplay packets, but they are applied by `WorldLaneApplier` lifecycle methods before the presentation bridge fans them to gameplay consumers.

Cross-lane ordering is not guaranteed between reliable lifecycle lanes and unreliable hot lanes. Clients must tolerate hot updates arriving before lifecycle create packets and after lifecycle delete packets.

## Related Docs

* [Presentation Bridge](../gameplay-runtime/presentation-bridge.md)

## Protocols and APIs

### Inbound routing surface

The inbound routing surface is the client-side handling path for decoded server packets.

The surface is consumed by client session controllers and direct consumers such as telemetry. The game server owns authority behind the packets. Data crossing this boundary is a decoded packet dictionary whose `type` field has already passed envelope validation.

Inbound routing explicitly does not own the packet schema, the raw transport, or the domain consequences of applying a packet.

### Routing sequence

Normal decoded packet sequence:

```text
NetworkClient.packet_received(packet)
-> ClientConnectionService._on_packet_received(packet)
-> ServerPacketDispatcher.dispatch(packet)
-> ServerPacketRouter checks packet type
-> typed dispatcher signal emitted
-> ClientConnectionService._route_gameplay_packet(packet)
-> RealtimePacketPipeline.apply_packet(packet)
-> RealtimeRouter.route_lane_packet(packet)
-> RealtimePacketPipeline.gameplay_packet_applied(packet)
-> PresentationBridge.handle_gameplay_packet(packet)
-> owning session/controller handles the packet
```

Packet parse failures do not enter this routing path. They are emitted separately as:

```text
packet_parse_failed(text)
```

### Current inbound packet routes

```text
authenticate_result
-> websocket_auth_result_received
-> SessionNetworkController websocket auth gate

room_snapshot
-> room_snapshot_received
-> RoomSessionController.handle_room_snapshot

room_state_changed
-> room_state_changed
-> RoomSessionController.handle_room_state_changed

room_error
-> room_error_received
-> RoomSessionController.handle_room_error

world_full/world_delta/asteroid_delta/bullet_delta/overlay_full/overlay_delta/session_full/session_delta/event_batch/resync_request/resync_required
-> dispatcher lane signal
-> ClientConnectionService._route_gameplay_packet
-> lane-specific service signal
-> PresentationBridge.handle_gameplay_packet
-> GameplayComposition / runtime presentation flows

debug_shape_catalog
-> debug_shape_catalog_received
-> GameplaySessionController.handle_debug_shape_catalog_packet

debug_status
-> debug_status_received
-> GameplaySessionController.handle_debug_status_packet

player_pause_state
-> player_pause_state_received
-> GameplaySessionController.handle_player_pause_state

telemetry_pong
-> telemetry_pong_received
-> WorldTelemetryContext._on_telemetry_pong_received

unmatched packet type
-> unknown_packet_received
-> SessionNetworkController logs unknown packet through its configured logger callable
```

### Auth gate interaction

Inbound routing participates in multiplayer boot gating only by delivering connection and auth signals.

Current multiplayer boot behavior is owned by `SessionNetworkController` and `ShellBootFlow`:

Logger mechanics belong to [Client Logging](../client-logging.md). `AppEntry` wires the session network controller logger to `_log_shell_status()`, which forwards to `ClientLogger.shell_info(...)`.

```text
connected + pending multiplayer request + websocket auth already authenticated
-> send pending request

connected + pending multiplayer request + websocket auth not authenticated
-> wait for authenticate_result

authenticate_result authenticated=true
-> send pending request
