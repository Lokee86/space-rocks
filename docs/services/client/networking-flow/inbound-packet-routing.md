---
author: brian
created: "2026-07-19"
document_id: 019f7d55-fb2c-70aa-8252-7980089bdbb6
document_type: general
policy_exempt: false
summary: This document describes the current client inbound packet routing path.
---
# Inbound Packet Routing

Parent index: [Networking Flow](./!INDEX.md)

## Purpose

This document describes the current client inbound packet routing path.

It covers how decoded server packet dictionaries move from the client WebSocket transport and WebRTC DataChannel transport into packet classification, dispatcher signals, application-facing connection-service signals, and downstream session, room, gameplay, auth, devtools, and telemetry consumers.

## Overview

Inbound packet routing begins after the WebSocket transport or WebRTC DataChannel transport has already decoded raw text into a packet dictionary.
WebSocket remains the client route for session, control, room, lobby, auth, WebRTC signaling/control, and queued non-tooling packets. Runtime telemetry and developer readouts use the dedicated reliable `sr.tooling` WebRTC DataChannel. In particular, `webrtc_answer`, `webrtc_ice_candidate`, `webrtc_ready`, `webrtc_smoke`, and `webrtc_failed` are signaling/control packets received through WebSocket ingress. Lane-specific WebRTC gameplay DataChannels are the route for active realtime gameplay packets: `world`, `overlay`, `session`, `event`, ship lifecycle, asteroid lifecycle, and bullet lifecycle packets use the `sr.world`, `sr.overlay`, `sr.session`, `sr.event`, `sr.ships.lifecycle`, `sr.asteroids.lifecycle`, and `sr.bullets.lifecycle` ordered/reliable lifecycle channels, while ship, asteroid, and bullet delta packets use the `sr.ships`, `sr.asteroids`, and `sr.bullets` unordered/unreliable hot-update lanes. WebRTC connectivity is established by ICE, not by a WebRTC URL, and deployment must ensure the advertised ICE address can reach the game server directly.

`NetworkClient` owns raw WebSocket polling, text receive, JSON decode, envelope validation, and `packet_received` emission for WebSocket packets. `RealtimeTransportSession` owns the WebRTC transport-session lifecycle and lane-aware dispatch callback. WebSocket packets and gameplay-lane WebRTC packets enter `ServerPacketDispatcher`; `sr.tooling` packets enter `ToolingPacketRouter` directly. `ClientConnectionService` owns the application-facing non-realtime dispatcher bindings, tooling-router signal bridge, and public facade. `ClientInboundCoordinator` owns only the WebRTC control bindings for `webrtc_answer_received`, `webrtc_ice_candidate_received`, `webrtc_ready_received`, `webrtc_smoke_received`, and `webrtc_failed_received`. `RealtimePacketPipeline` owns all gameplay lane dispatcher bindings and typed gameplay entry points.

Compact `event_batch` packets are expanded by the client compact packet expansion layer before event appliers receive them, so downstream gameplay code sees readable long-key event dictionaries rather than compact aliases. Nested `ev` or `events` entries are expanded as part of that same transport step, and event dedupe still keys off `event_id` after expansion. Runtime wire packets keep compact aliases on the wire; domain logs may still show raw x/y before projection.

Current flow:

```text
NetworkClient.packet_received(packet)
-> ClientConnectionService._on_packet_received(packet)
-> ServerPacketDispatcher.dispatch(packet)
-> ClientConnectionService application-facing event
or ClientInboundCoordinator WebRTC control handling through RealtimeTransportSession

WebRTCTransport receives DataChannel text
-> PacketCodec.decode expands compact aliases and validates the packet envelope
-> WebRTCTransport.packet_received(packet)
-> RealtimeTransportSession._on_packet_received(packet)
-> RealtimeTransportSession dispatch callback(packet)
-> ServerPacketDispatcher.dispatch(packet)
-> typed dispatcher signal
-> RealtimePacketPipeline typed entry point for the packet family
-> expand and validate recognized packet
-> no active protocol match: buffer by non-empty match_id without state mutation
-> active protocol match: require matching match_id
-> RealtimeRouter.route_lane_packet(packet)
-> lifecycle packet: LifecycleLaneGate apply / queue / reject / resync on capacity loss
-> accepted lifecycle packet: WorldLaneApplier validates and mutates WorldLaneState
-> RealtimePacketPipeline.gameplay_packet_applied(packet)
-> PresentationBridge.handle_gameplay_packet(packet)

WebRTCTransport receives sr.tooling text
-> PacketCodec.decode validates the tooling envelope
-> RealtimeTransportSession dispatch callback(packet, tooling)
-> ClientConnectionService._on_tooling_packet_received(packet)
-> ToolingPacketRouter.dispatch(packet)
-> telemetry, measurement, debug-readout, command-result, or tooling-error signal
```

The routing path is signal-based and lane-aware. It does not mutate server authority, does not parse payload-specific gameplay data, and does not apply presentation state directly. Its job is to classify packet family by generated packet type constants, forward lane packets through the realtime router, and hand the dictionary to the owning client subsystem. Inbound realtime lane packets may already contain quantized numeric wire values. The client routes them by lane and packet family, does not own authoritative quantization decisions, and uses `client/scripts/protocol/realtime/realtime_quantize.gd` when it needs to decode quantized realtime lane values. Lifecycle packets enter `LifecycleLaneGate`: they apply immediately only when their explicit world baseline matches the active synced world baseline, otherwise they queue or reject. `WorldLaneApplier` validates and mutates lifecycle state only after gate acceptance. Hot lanes update known entities only, but a lifecycle notification does not prove that an entity was created or deleted. Hot ship, asteroid, and bullet packets are routed on unordered/unreliable lanes. Their sequence values must be finite, non-negative, integer-valued numerics; missing, fractional, negative, non-finite, string, and boolean values are rejected before hot-lane state mutation. The client accepts distinct valid chunk indices with consistent `chunk_count` values for each hot lane sequence. Duplicate indices, inconsistent or malformed chunk metadata, and lower sequences are rejected; sequence gaps remain valid, and ship, asteroid, and bullet tracking are independent.

## Code root

* `client/scripts/networking/`
* `client/scripts/session/`

## Responsibilities

* Bind application-facing non-realtime `ServerPacketDispatcher` signals to `ClientConnectionService`, gameplay lane signals to `RealtimePacketPipeline`, and tooling packet signals to `ToolingPacketRouter`.
* Re-emit stable application-facing dispatcher and tooling-router signals from `ClientConnectionService`.
* Track websocket auth result state from `authenticate_result` packets.
* Preserve the auth-state, room, debug, player-pause, telemetry, and unknown-packet facade signals while keeping their physical transports explicit.
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

`WebRTCTransport` receives DataChannel text, calls `PacketCodec.decode`, and emits `packet_received` with its lane. `RealtimeTransportSession` forwards active gameplay lanes to `ServerPacketDispatcher.dispatch(packet)` and forwards the tooling lane to `ClientConnectionService._on_tooling_packet_received(packet)`, which delegates to `ToolingPacketRouter`. `ClientInboundCoordinator` does not consume WebRTC gameplay or tooling packets and does not treat them as signaling; it consumes WebSocket-produced dispatcher signals for `webrtc_answer`, `webrtc_ice_candidate`, `webrtc_ready`, `webrtc_smoke`, and `webrtc_failed`.

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
webrtc_answer
webrtc_ice_candidate
webrtc_ready
webrtc_smoke
webrtc_failed
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
player_pause_state
```

### Dispatcher signal fanout

`ServerPacketDispatcher` owns the ordered classification chain and typed signal emission.

Current dispatcher outputs are:

```text
room_snapshot_received(packet)
authenticate_result_received(packet)
room_state_changed(packet)
room_error_received(packet)
webrtc_answer_received(packet)
webrtc_ice_candidate_received(packet)
webrtc_ready_received(packet)
webrtc_smoke_received(packet)
webrtc_failed_received(packet)
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
player_pause_state_received(packet)
unknown_packet_received(packet)
```

The dispatcher does not know which application subsystem will consume each signal. It only converts packet-type classification into a signal and does not apply lane state.

### Resync control routing

`resync_required` is an acknowledgment of a client recovery request. The realtime router changes `ResyncState` only while the lane is still pending; it does not mark a completed lane unsynced, emit another request, or regress a recovery completed over WebRTC. Reason values `missing_baseline`, `wrong_baseline`, and `stale_or_invalid_sequence` are preserved by their corresponding mappings. A final accepted full clears the pending tracker state and `ResyncState`.

Inbound `resync_request` remains a compatibility route. It may mark the supplied lane unsynced and set its supplied/default reason, but it never creates an outbound request loop. The delayed-ack behavior is required because WebSocket acknowledgment and WebRTC recovery full delivery have no shared order.

### Connection-service signal bridge

`ClientConnectionService` owns the public connection facade, collaborator composition, polling, reset coordination, outbound API, and stable application-facing signal relay, including application-facing non-realtime dispatcher bindings and `ToolingPacketRouter` bindings. It also owns a typed `AuthSessionController` reference for the websocket authentication handoff. `ClientInboundCoordinator` owns only the five WebRTC control dispatcher bindings: `webrtc_answer_received`, `webrtc_ice_candidate_received`, `webrtc_ready_received`, `webrtc_smoke_received`, and `webrtc_failed_received`. `RealtimePacketPipeline` owns all gameplay lane dispatcher bindings plus realtime expansion, validation, application, readiness, reset, presentation-state refresh, and applied notification. Raw WebRTC control packets and realtime gameplay packets are not re-emitted through `ClientConnectionService`.

RealtimePacketPipeline emits a structured network diagnostic event when a lane packet is routed for the first time by packet type:

```text
event: lane_packet_routed
level: info
category: network
fields:
  packet_type
  readiness
```

The once-per-packet-type guard remains diagnostic-only. It does not affect routing or lane state.

Under hot-lane stress, multiple `ship_delta`, `asteroid_delta`, or `bullet_delta` packets may arrive for the same lane sequence in one poll window. The routing path treats them as separate packets and must not coalesce or drop them solely because they share a sequence; the owning hot-lane state accepts only distinct valid chunk indices whose `chunk_count` matches that lane sequence. Duplicate indices, inconsistent counts, malformed chunk metadata, and lower sequences are rejected, while distinct chunks may arrive in any order and sequence gaps remain valid. Ship, asteroid, and bullet tracking are independent.

Presentation remains frame-coalesced and is intentionally unchanged in this stage.

## Client receive hardening

Before an authoritative match is active, the client buffers at most 4 match buckets. Each bucket is limited to 128 packets and 256 KiB of estimated expanded JSON and expires after 5000 ms; the oldest bucket is evicted at capacity. A selected match with lost or expired buffered state fails closed and requests world, overlay, and session recovery.

`world_full`, ship lifecycle, asteroid lifecycle, and bullet lifecycle assembly are each limited to 128 chunks, 16384 cumulative records, 2 MiB of estimated expanded JSON, and 5000 ms. Limit, expiry, malformed metadata, interrupted, duplicate, mismatched, and non-contiguous failures reset the incomplete assembly, apply no partial state, and request authoritative recovery. Baseline IDs are non-empty strings; sequence/chunk values are finite, non-negative integer-valued numerics; final-chunk metadata is boolean and must agree with index/count. Valid stale deltas remain silent. These are defensive client receive limits, separate from the server's approximately 1200-byte candidate construction cap.

### Websocket auth result cache

`ClientConnectionService` handles `authenticate_result` specially because websocket auth state is connection-level state. Its `auth_session_controller` field is a typed `AuthSessionController` reference.

On `authenticate_result`, the connection service updates:

```text
websocket_auth_authenticated
websocket_auth_user_id
websocket_auth_display_name
```

`websocket_auth_user_id` is an integer cache. `NO_WEBSOCKET_AUTH_USER_ID` (`-1`) represents no accepted websocket identity; the cache does not use `null`.

An `authenticate_result` contributes an identity only when `authenticated` is `true` and `user_id` is an actual integer. Missing, string, float, or unauthenticated values use `NO_WEBSOCKET_AUTH_USER_ID`.

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
realtime_transport_ready
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

Realtime gameplay lane packets enter `RealtimePacketPipeline`, which expands and validates recognized packets before its protocol match gate. WebSocket room snapshots and WebRTC gameplay packets have no cross-transport ordering guarantee, so packets with a non-empty `match_id` are buffered by match before authoritative activation and do not mutate lane or presentation state. The authoritative `InGame` snapshot resets gameplay presentation/composition and calls `begin_realtime_match(match_id)`; the pipeline clears all pending buckets, replays only the matching bucket, and discards unrelated matches. Missing IDs remain rejected, and mismatched IDs remain rejected once active. `GameOver` retains the active match; returning to `Lobby`, reset, or connection teardown clears pending and protocol state. The routed notification means routing/state refresh completed; it does not prove that a particular lifecycle packet mutated state.

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

Lane-native realtime gameplay packets route through the semantic handoff only after protocol match activation and pipeline state refresh. `PresentationBridge` activation/readiness is separate: matching packets may route while presentation is inactive, whereas pre-activation packets remain buffered. Lifecycle packets may have applied immediately, been queued for a matching world baseline, or been rejected before this notification.

Current handoff:

```text
RealtimePacketPipeline.gameplay_packet_applied(packet)
-> PresentationBridge.handle_gameplay_packet(packet)
-> gameplay composition / runtime presentation flows
```

`SessionNetworkController` remains the bridge from networking events into room and gameplay session controllers. It still preserves the control routing for connection, room, auth, debug, and player-pause packets.

Gameplay packet application continues through gameplay runtime documentation after this point. Inbound routing only delivers the packet.

### Debug packet handoff

Debug shape catalog and debug status packets arrive through `sr.tooling` and `ToolingPacketRouter`, then route through the same gameplay session controller because current devtools presentation is composed inside gameplay session context.

Current handoff:

```text
ToolingPacketRouter.debug_shape_catalog_received
-> ClientConnectionService.debug_shape_catalog_received
-> SessionNetworkController._on_debug_shape_catalog_received
-> GameplaySessionController.handle_debug_shape_catalog_packet

ToolingPacketRouter.debug_status_received
-> ClientConnectionService.debug_status_received
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

`telemetry_snapshot` and `telemetry_pong` packets arrive through `sr.tooling`, are classified by `ToolingPacketRouter`, and are re-emitted by the connection service.

`WorldTelemetryContext` connects directly to:

```text
ClientConnectionService.telemetry_pong_received
```

and applies the pong packet to network telemetry metrics. The same context consumes `ClientConnectionService.telemetry_snapshot_received` and merges authoritative server runtime fields into the visible telemetry snapshot.

Telemetry handling is diagnostic. It does not require room membership, does not mutate gameplay state, and does not route through normal gameplay packet application.

### Unknown packet fallback

If the packet envelope is valid but no current router predicate matches the packet type, the dispatcher emits:

```text
unknown_packet_received(packet)
```

The unknown-packet event is owned by `SessionNetworkController` and is emitted canonically with the active connection or room operation trace. It is not forwarded through a removed shell-status helper.

Unknown packets are not applied to gameplay, room, auth, or telemetry state.

### Lifecycle routing note

Lifecycle packets use the same dispatcher and pipeline classification path as other gameplay packets:

```text
typed lifecycle signal
-> RealtimePacketPipeline.apply_ships_lifecycle / apply_asteroids_lifecycle / apply_bullets_lifecycle
-> RealtimeRouter.route_lane_packet(packet)
-> LifecycleLaneGate apply / queue / reject / resync on capacity loss
-> WorldLaneApplier lifecycle validation and WorldLaneState mutation only after acceptance
```

After a completed matching `world_full` is applied and recorded, `RealtimeRouter` drains pending lifecycle packets for that world baseline, ordered within each lifecycle lane. There is no ordering contract between the two lifecycle lanes.

Reliable/ordered delivery orders messages only within one DataChannel. Cross-lane ordering is not guaranteed between `sr.world` and either lifecycle channel, between the two lifecycle channels, or between lifecycle and unreliable hot lanes. Clients must tolerate hot updates arriving before lifecycle create packets and after lifecycle delete packets; lifecycle packets may also wait for a matching world baseline.

## Related Docs

* [Presentation Bridge](../gameplay-runtime/presentation-bridge.md)

## Protocols and APIs

### Inbound routing surface

The inbound routing surface is the client-side handling path for decoded server packets.

The surface is consumed by client session controllers and direct consumers such as telemetry. The game server owns authority behind the packets. Data crossing this boundary is a decoded packet dictionary whose `type` field has already passed envelope validation.

Inbound routing explicitly does not own the packet schema, the raw transport, or the domain consequences of applying a packet.

### Routing sequence

WebSocket service and control ingress:

```
  NetworkClient.packet_received(packet)
  -> ClientConnectionService._on_packet_received(packet)
  -> ServerPacketDispatcher.dispatch(packet)
  -> ServerPacketRouter checks packet type
  -> ClientConnectionService application-facing event
     or ClientInboundCoordinator WebRTC control handling
```

WebRTC active gameplay ingress:

```
  WebRTCTransport.packet_received(packet, gameplay_lane)
  -> RealtimeTransportSession gameplay dispatch
  -> ServerPacketDispatcher.dispatch(packet)
  -> ServerPacketRouter checks packet type
  -> RealtimePacketPipeline typed entry point for the packet family
  -> RealtimeRouter.route_lane_packet(packet)
  -> RealtimePacketPipeline.gameplay_packet_applied(packet)
  -> PresentationBridge.handle_gameplay_packet(packet)
```

WebRTC tooling ingress:

```
  WebRTCTransport.packet_received(packet, tooling)
  -> RealtimeTransportSession tooling dispatch
  -> ClientConnectionService._on_tooling_packet_received(packet)
  -> ToolingPacketRouter.dispatch(packet)
  -> stable ClientConnectionService tooling signal
```

Packet parse failures do not enter either routing path. They are emitted separately as:

```
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

world_full
-> world_full_received(packet)
-> RealtimePacketPipeline.apply_world_full(packet)
-> RealtimeRouter.route_lane_packet(packet)
-> RealtimePacketPipeline.gameplay_packet_applied(packet)
-> PresentationBridge.handle_gameplay_packet(packet)
-> GameplayComposition / runtime presentation flows

world_delta
-> world_delta_received(packet)
-> RealtimePacketPipeline.apply_world_delta(packet)
-> RealtimeRouter.route_lane_packet(packet)
-> RealtimePacketPipeline.gameplay_packet_applied(packet)
-> PresentationBridge.handle_gameplay_packet(packet)
-> GameplayComposition / runtime presentation flows

asteroid_delta
-> asteroid_delta_received(packet)
-> RealtimePacketPipeline.apply_asteroid_delta(packet)
-> RealtimeRouter.route_lane_packet(packet)
-> RealtimePacketPipeline.gameplay_packet_applied(packet)
-> PresentationBridge.handle_gameplay_packet(packet)
-> GameplayComposition / runtime presentation flows

bullet_delta
-> bullet_delta_received(packet)
-> RealtimePacketPipeline.apply_bullet_delta(packet)
-> RealtimeRouter.route_lane_packet(packet)
-> RealtimePacketPipeline.gameplay_packet_applied(packet)
-> PresentationBridge.handle_gameplay_packet(packet)
-> GameplayComposition / runtime presentation flows

asteroids_lifecycle
-> asteroids_lifecycle_received(packet)
-> RealtimePacketPipeline.apply_asteroids_lifecycle(packet)
-> RealtimeRouter.route_lane_packet(packet)
-> LifecycleLaneGate apply / queue / reject / resync on capacity loss
-> WorldLaneApplier validates and mutates only after acceptance
-> RealtimePacketPipeline.gameplay_packet_applied(packet)
-> PresentationBridge.handle_gameplay_packet(packet)
-> GameplayComposition / runtime presentation flows

bullets_lifecycle
-> bullets_lifecycle_received(packet)
-> RealtimePacketPipeline.apply_bullets_lifecycle(packet)
-> RealtimeRouter.route_lane_packet(packet)
-> LifecycleLaneGate apply / queue / reject / resync on capacity loss
-> WorldLaneApplier validates and mutates only after acceptance
-> RealtimePacketPipeline.gameplay_packet_applied(packet)
-> PresentationBridge.handle_gameplay_packet(packet)
-> GameplayComposition / runtime presentation flows

overlay_full
-> overlay_full_received(packet)
-> RealtimePacketPipeline.apply_overlay_full(packet)
-> RealtimeRouter.route_lane_packet(packet)
-> RealtimePacketPipeline.gameplay_packet_applied(packet)
-> PresentationBridge.handle_gameplay_packet(packet)
overlay_delta
-> overlay_delta_received(packet)
-> RealtimePacketPipeline.apply_overlay_delta(packet)
-> RealtimeRouter.route_lane_packet(packet)
-> RealtimePacketPipeline.gameplay_packet_applied(packet)
-> PresentationBridge.handle_gameplay_packet(packet)
-> GameplayComposition / runtime presentation flows

session_full
-> session_full_received(packet)
-> RealtimePacketPipeline.apply_session_full(packet)
-> RealtimeRouter.route_lane_packet(packet)
-> RealtimePacketPipeline.gameplay_packet_applied(packet)
-> PresentationBridge.handle_gameplay_packet(packet)
-> GameplayComposition / runtime presentation flows

session_delta
-> session_delta_received(packet)
-> RealtimePacketPipeline.apply_session_delta(packet)
-> RealtimeRouter.route_lane_packet(packet)
-> RealtimePacketPipeline.gameplay_packet_applied(packet)
-> PresentationBridge.handle_gameplay_packet(packet)
-> GameplayComposition / runtime presentation flows

event_batch
-> event_batch_received(packet)
-> RealtimePacketPipeline.apply_event_batch(packet)
-> RealtimeRouter.route_lane_packet(packet)
-> RealtimePacketPipeline.gameplay_packet_applied(packet)
-> PresentationBridge.handle_gameplay_packet(packet)
-> GameplayComposition / runtime presentation flows

resync_request
-> resync_request_received(packet)
-> RealtimePacketPipeline.apply_resync_request(packet)
-> RealtimeRouter.route_lane_packet(packet)
-> RealtimePacketPipeline.gameplay_packet_applied(packet)
-> PresentationBridge.handle_gameplay_packet(packet)
-> GameplayComposition / runtime presentation flows

resync_required
-> resync_required_received(packet)
-> RealtimePacketPipeline.apply_resync_required(packet)
-> RealtimeRouter.route_lane_packet(packet)
-> RealtimePacketPipeline.gameplay_packet_applied(packet)
-> PresentationBridge.handle_gameplay_packet(packet)
-> GameplayComposition / runtime presentation flows

debug_shape_catalog
-> ToolingPacketRouter.debug_shape_catalog_received
-> ClientConnectionService.debug_shape_catalog_received
-> GameplaySessionController.handle_debug_shape_catalog_packet

debug_status
-> ToolingPacketRouter.debug_status_received
-> ClientConnectionService.debug_status_received
-> GameplaySessionController.handle_debug_status_packet

player_pause_state
-> player_pause_state_received
-> GameplaySessionController.handle_player_pause_state

telemetry_snapshot
-> ToolingPacketRouter.telemetry_snapshot_received
-> ClientConnectionService.telemetry_snapshot_received
-> WorldTelemetryContext._on_telemetry_snapshot_received

telemetry_pong
-> ToolingPacketRouter.telemetry_pong_received
-> ClientConnectionService.telemetry_pong_received
-> WorldTelemetryContext._on_telemetry_pong_received

unmatched packet type
-> unknown_packet_received
-> SessionNetworkController owns canonical unknown-route emission with the active connection/operation trace
```

### Auth gate interaction

Inbound routing participates in multiplayer boot gating only by delivering connection and auth signals.

Current multiplayer boot behavior is owned by `SessionNetworkController` and `ShellBootFlow`:

Logger mechanics belong to [Client Logging](../client-logging.md). Routing failures and unknown routes are emitted by their owning connection/session flow through `ClientLogger.emit_canonical(...)` with the active operation trace.

```text
connected + pending multiplayer request + websocket auth already authenticated
-> send pending request

connected + pending multiplayer request + websocket auth not authenticated
-> wait for authenticate_result

authenticate_result authenticated=true
-> send pending request
```
