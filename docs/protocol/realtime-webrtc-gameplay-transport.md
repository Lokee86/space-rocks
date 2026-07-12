## Realtime WebRTC Gameplay Transport

Parent index: [Protocol](./!INDEX.md)

## Purpose

This document defines the current transport boundary between gameplay data delivery and signaling/control.

It is the canonical protocol doc for physical gameplay DataChannels, lane-to-channel mapping, active gameplay readiness, server send routing, client receive routing, and the current mixed WebRTC gameplay channel policy.

## Ownership

This document owns:

```text
physical WebRTC gameplay DataChannel names
physical WebRTC gameplay DataChannel numeric IDs
logical realtime lane to physical channel mapping
current active gameplay transport split from WebSocket
current gameplay channel reliability policy
active gameplay transport readiness expectations
server send path at the transport boundary
client receive path at the transport boundary
control/resync transport boundary
```

## Does Not Belong

This document does not own:

```text
room lifecycle rules
simulation rules
client rendering behavior
devtools command consequences
packet schema generation rules
compact field alias details
future binary/protobuf packet format design
future transport designs beyond the current unordered/unreliable asteroid and bullet hot lanes
record-level prioritization policy
```

## Transport Split

The current realtime stack uses two separate paths:

WebSocket owns signaling/control and queued non-gameplay server packets:

```text
auth packets
room packets
lobby packets
telemetry packets
WebRTC signaling packets
queued one-off server packets
```

WebRTC owns active gameplay data packets:

```text
active realtime gameplay output packets
```

Active realtime gameplay output is not WebSocket-owned. There is no WebSocket fallback for active gameplay output packets.

## Physical Gameplay Channels

Current active gameplay output uses negotiated WebRTC DataChannels with lane-specific reliability policy.

```text
logical lane | physical channel | negotiated id | packet families
world        | sr.world         | 1             | world_full, world_delta
overlay      | sr.overlay       | 2             | overlay_full, overlay_delta
session      | sr.session       | 3             | session_full, session_delta
event        | sr.event         | 4             | event_batch
asteroids    | sr.asteroids     | 5             | asteroid_delta
bullets      | sr.bullets       | 6             | bullet_delta
asteroids.lifecycle | sr.asteroids.lifecycle | 7 | asteroids_lifecycle
bullets.lifecycle   | sr.bullets.lifecycle   | 8 | bullets_lifecycle
```

The current channel policy is:

```text
sr.world, sr.overlay, sr.session, sr.event, sr.asteroids.lifecycle, and sr.bullets.lifecycle are negotiated ordered/reliable channels.
sr.asteroids and sr.bullets are negotiated unordered/unreliable channels with maxRetransmits=0.
sr.asteroids carries supersedable asteroid_updates only.
sr.bullets carries supersedable bullet_updates only.
Entity lifecycle ownership is split by entity family. The world lane owns player, pickup, world, and full/bootstrap presentation state. Asteroid lifecycle packets use sr.asteroids.lifecycle. Bullet/projectile lifecycle packets use sr.bullets.lifecycle. Hot asteroid and bullet lanes are unreliable movement/update lanes only and must not create entities implicitly.
Lower-sequence `asteroid_delta` and `bullet_delta` packets are rejected by client hot-lane sequence guards. Same-sequence packets are valid only for distinct `chunk_index` values of the same `chunk_count`; duplicate chunk indices are rejected. Sequence gaps remain valid because hot packets can be dropped.
```

Ordered/reliable delivery is scoped to one DataChannel only. It does not order `sr.world` against `sr.asteroids.lifecycle` or `sr.bullets.lifecycle`, and it does not order a lifecycle channel against its corresponding unordered hot channel. The world/lifecycle race is therefore expected: a lifecycle packet may arrive before the matching `world_full` or after a hot update. The lifecycle packet's explicit world `baseline_id` determines whether the client can apply it now or must keep it pending.


## Active Gameplay Readiness

WebSocket connection readiness is not active gameplay transport readiness.

Active gameplay output requires the WebRTC gameplay transport to exist and be ready. The current client boot flow waits for WebRTC gameplay transport readiness before dispatching gameplay boot requests, while gameplay runtime flows remain responsible for presentation activation after packet application.

A session is not eligible for active gameplay output unless the server has:

```text
a current room
an active game instance
a current active game player id
a ready WebRTC transport
the selected candidate lane mapped to an open gameplay DataChannel
```

## Server Send Boundary

The server active gameplay path builds lane candidates, encodes each selected candidate, and writes each encoded packet to the matching WebRTC active gameplay channel.

The server send boundary is:

```text
BuildActiveRealtimeResultForGame
-> realtime lane candidates, including expanded asteroid/bullet hot chunks when needed
-> selected realtime lane candidates
-> encoded lane packet list
-> SendEncodedLaneJSON(candidate.Lane, encodedPacket) for each encoded packet
-> physical WebRTC active gameplay DataChannel
```

Active metadata advancement, event draining, and baseline persistence must only happen after the active WebRTC write path succeeds.

## Client Receive Boundary

The client active gameplay receive path is:

```text
WebRTCTransport receives DataChannel text
-> PacketCodec.decode(packet_text)
-> WebRTCTransport.packet_received(packet)
-> RealtimeTransportSession._on_packet_received(packet)
-> RealtimeTransportSession dispatch callback(packet)
-> ServerPacketDispatcher.dispatch(packet)
-> typed dispatcher signal
-> ClientInboundCoordinator
-> matching typed RealtimePacketPipeline.apply_* method
-> RealtimeRouter.route_lane_packet(packet)
-> lifecycle packet: LifecycleLaneGate apply / queue / reject / resync on capacity loss
-> lifecycle apply: WorldLaneApplier validates and mutates WorldLaneState
-> RealtimePresentationState refreshed
-> RealtimePacketPipeline.gameplay_packet_applied(packet)
-> PresentationBridge.handle_gameplay_packet(packet)
```

The typed pipeline entry point matches the packet family: `world_full` → `apply_world_full`, `world_delta` → `apply_world_delta`, `asteroid_delta` → `apply_asteroid_delta`, `bullet_delta` → `apply_bullet_delta`, `asteroids_lifecycle` → `apply_asteroids_lifecycle`, `bullets_lifecycle` → `apply_bullets_lifecycle`, `overlay_full` → `apply_overlay_full`, `overlay_delta` → `apply_overlay_delta`, `session_full` → `apply_session_full`, `session_delta` → `apply_session_delta`, `event_batch` → `apply_event_batch`, `resync_request` → `apply_resync_request`, and `resync_required` → `apply_resync_required`. Lifecycle routing submits explicit lane, sequence, and world-baseline metadata to `LifecycleLaneGate`; a completed matching `world_full` records the active world baseline and drains pending packets for that baseline, sorted within each lifecycle lane.

`RealtimeTransportSession` owns the WebRTC transport lifecycle and transport-originated callbacks. `ClientConnectionService` configures the transport-session dispatch callback but does not inspect or relay each gameplay packet. `ServerPacketDispatcher` emits the typed gameplay signal to `ClientInboundCoordinator`, which invokes the matching typed `RealtimePacketPipeline` apply method. The pipeline expands and validates the packet, refreshes `RealtimePresentationState`, and emits `gameplay_packet_applied(packet)`. `PresentationBridge` consumes that semantic notification with later coalesced flush when ready.

## WebRTC signaling and control routing

WebSocket packets carry the signaling and control route for the current transport session:

```text
WebSocket packet
-> NetworkClient
-> ClientConnectionService
-> ServerPacketDispatcher
-> ClientInboundCoordinator
-> RealtimeTransportSession
```

The WebSocket signaling/control route does not pass through the gameplay pipeline. `ClientInboundCoordinator` forwards answer, remote ICE, ready, smoke, and failure handling to `RealtimeTransportSession` or its coordinator-owned diagnostic/readiness handlers.

Inbound control ownership is:

```text
answer -> RealtimeTransportSession.handle_answer
remote ICE -> RealtimeTransportSession.handle_remote_ice
failure -> RealtimeTransportSession.handle_remote_failure
smoke -> diagnostic consumption
ready -> semantic realtime_transport_ready
```

Outbound signaling ownership is:

```text
local offer and local ICE originate from RealtimeTransportSession
callbacks use the existing ClientConnectionService.send_webrtc_* outbound facade
final packet sending remains ClientPacketSender and NetworkClient
```

Readiness is split between transport-local and server-confirmed states:

```text
WebRTCTransport.ready
-> local DataChannel readiness
-> smoke diagnostics start

server webrtc_ready packet
-> ServerPacketDispatcher
-> ClientInboundCoordinator.handle_webrtc_ready
-> ClientInboundCoordinator.realtime_transport_ready
-> ClientConnectionService.realtime_transport_ready
-> SessionNetworkController._on_realtime_transport_ready
-> _webrtc_gameplay_ready = true
-> _try_send_pending_boot_request()
```

`ClientConnectionService` owns logical connection composition, polling, reset coordination, outbound facade, and semantic ready relay. `ClientInboundCoordinator` owns inbound WebRTC control routing. `RealtimeTransportSession` owns WebRTC transport lifecycle and transport-originated callbacks. `WebRTCTransport` owns peer/DataChannel mechanics.

On reconnect, disconnect clears the coordinator transport-session target, the replacement session is assigned on reconnect, and later control packets target only the replacement session.

## Client Transport Ownership

ClientConnectionService remains the public connection coordinator, but it does not directly own WebRTC peer and DataChannel lifecycle mechanics.

RealtimeTransportSession owns the active WebRTCTransport reference, transport construction through transport_factory, transport signal wiring, start, poll, close, clearing the active transport after close, and replacement after reconnect.

WebRTCTransport owns the peer connection, DataChannel setup, bounded receive draining, packet decoding, packet emission, outbound DataChannel writes, and transport-level diagnostics. Receive polling uses repeated one-packet-per-lane round-robin passes; reliable asteroid and bullet lifecycle lanes are serviced before the general group, and lifecycle and general group start positions rotate between polls. The existing limits remain `MAX_PACKETS_PER_POLL = 48` and `MAX_PACKETS_PER_LANE_PER_POLL = 12`.

### Start, Poll, and Close Contract

RealtimeTransportSession.start() returns without action when an active transport is already assigned. Otherwise it creates a transport through transport_factory, wires offer, ICE, failure, ready, and packet signals, then calls WebRTCTransport.start().

RealtimeTransportSession.poll() delegates only while an active transport exists.

RealtimeTransportSession.close() closes the active transport once and clears the stored transport reference. Polling after close is therefore a no-op until a later start creates a replacement transport.

Tests or alternate composition paths that need normal startup must provide transport_factory rather than preassigning transport. A preassigned transport represents an already-active session and intentionally prevents start() from creating or starting another transport.

### Reconnect and Replacement

Connection teardown closes the previous session transport. A later connection uses transport_factory to create a fresh WebRTCTransport, rewires its signals, and starts it. Packets emitted by the replacement transport re-enter the same dispatcher and gameplay application path as packets from the original transport.

Transport replacement does not replace gameplay protocol state ownership. RealtimePacketPipeline separately owns the active RealtimeRouter, gameplay readiness, packet application, and protocol reset. `RealtimePacketPipeline.reset()` replaces the router and clears lifecycle pending packets, pending duplicate tracking, and latest applied lifecycle sequences. Lifecycle queue capacity loss uses the existing world-lane resync path; transport replacement itself does not provide reconnect recovery.

## Hot Movement Split

Regular asteroid movement updates are emitted as:

```text
asteroid_delta on sr.asteroids
```

Regular bullet movement updates are emitted as:

```text
bullet_delta on sr.bullets
```

`world_delta` remains responsible for ships, pickups, and full/bootstrap/resync-safe presentation state. World serializer compatibility may still accept asteroid or bullet update sections, but regular active asteroid and bullet movement is split to the dedicated hot movement lanes.

### Hot Movement Cadence

Each eligible active build advances an independent per-session 60 Hz `HotLaneTick`. Asteroid movement emits at 60 Hz when unchunked and 30 Hz when chunking is required; bullet movement emits at 60 Hz for one chunk, 30 Hz for two chunks, and 20 Hz for three or more chunks. Forced sends bypass cadence suppression.

## Control and Resync Boundary

The current signaling/control path remains WebSocket-owned.

There is no physical `sr.control` WebRTC gameplay DataChannel in the current implementation.

The current recovery/control packet families are logical packet families:

```text
resync_request
resync_required
```

Do not document `control` as a current physical gameplay channel unless a real physical channel is implemented.

Resync control packets remain WebSocket-only: `resync_request` is client-to-server and `resync_required` is the server acknowledgment. Neither is a `RealtimeLaneCandidate` or selected by the realtime scheduler. The server writes the acknowledgment over WebSocket, while the next normal required full candidate travels over the ordered/reliable WebRTC lane for world, overlay, or session. These transports have no shared delivery order, so the client applies `resync_required` only while recovery is pending and ignores a delayed acknowledgment after the full has completed. Transport reset clears protocol state for the current session; it is not reconnect recovery, resend, retry, or session resume, and there is no `sr.control` DataChannel.

## Current Packet Encoding

The current active gameplay transport is JSON text over WebRTC DataChannels.

The implementation does not currently provide:

```text
binary packet encoding
protobuf packet encoding
compression
schema negotiation
runtime version negotiation
record-level packet-budget enforcement
general-purpose fragmentation for all lane families
```

Focused hot-lane chunking is implemented for `asteroid_delta` and `bullet_delta` before the WebRTC send boundary. It emits multiple JSON messages on `sr.asteroids` or `sr.bullets` when a hot movement update list would exceed the hard encoded packet cap.

Unordered/unreliable hot-lane delivery is implemented for sr.asteroids and sr.bullets.

Compact JSON aliases, sparse delta omission, numeric quantization, tuple packing, and dedicated asteroid/bullet hot movement packets are implemented before the final WebRTC write boundary.

## Related Docs

* [Realtime WebSocket Protocol](realtime-websocket-protocol.md)
* [Gameplay Packets](gameplay-packets.md)
* [Realtime Compact Wire Mapping](../services/game-server/networking/realtime-compact-wire-mapping.md)
* [Outbound Message Flow](../services/game-server/networking/outbound-message-flow.md)
* [Lane Packet Projection](../services/game-server/simulation/runtime/lane-packet-projection.md)

Lifecycle transport implementation and verification references include `client/scripts/protocol/realtime/lifecycle_lane_gate.gd`, `client/scripts/protocol/realtime/realtime_router.gd`, `client/scripts/networking/realtime/realtime_packet_pipeline.gd`, `client/tests/unit/protocol/realtime/test_lifecycle_lane_gate.gd`, `client/tests/unit/networking/realtime/test_realtime_packet_pipeline.gd` reset coverage, and the server lifecycle wire metadata tests in `services/game-server/internal/protocol/realtime/wire_packets_test.go`.
