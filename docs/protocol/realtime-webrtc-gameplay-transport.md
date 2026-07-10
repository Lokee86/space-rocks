## Realtime WebRTC Gameplay Transport

Parent index: [Protocol](./!INDEX.md)

## Purpose

This document defines the current WebRTC transport boundary for active realtime gameplay packets.

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

The current realtime stack uses both WebSocket and WebRTC.

WebSocket owns:

```text
auth packets
room packets
lobby packets
telemetry packets
WebRTC signaling packets
queued one-off server packets
```

WebRTC owns:

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
Lower-sequence `asteroid_delta` and `bullet_delta` packets are rejected by client hot-lane sequence guards. Same-sequence packets are valid when they are chunks of one hot-lane update sequence. Sequence gaps are valid because hot packets can be dropped.
```


## Active Gameplay Readiness

WebSocket connection readiness is not active gameplay transport readiness.

Active gameplay output requires the WebRTC gameplay transport to exist and be ready. The current client boot flow waits for WebRTC gameplay transport readiness before dispatching gameplay boot requests.

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
-> RealtimeTransportSession packet callback
-> ClientConnectionService
-> ServerPacketDispatcher
-> RealtimePacketPipeline.apply_packet(packet)
-> RealtimePacketPipeline expands and validates the packet
-> RealtimeRouter.route_lane_packet(packet)
-> RealtimePresentationState refreshed
-> RealtimePacketPipeline.gameplay_packet_applied(packet)
-> PresentationBridge.handle_gameplay_packet(packet)
```

RealtimeTransportSession owns transport lifecycle and signal handoff only. `ServerPacketDispatcher` classifies gameplay packets to `RealtimePacketPipeline`. `RealtimePacketPipeline` applies the packet, refreshes `RealtimePresentationState`, and emits `gameplay_packet_applied(packet)`. `PresentationBridge` consumes that semantic notification, with later coalesced flush when ready.

Dedicated asteroid and bullet hot movement packets do not create independent rendered state. They merge into the same world presentation state used by gameplay rendering. Lower-sequence `asteroid_delta` and `bullet_delta` packets are rejected by client hot-lane sequence guards. Same-sequence packets are valid when they are chunks of one hot-lane update sequence. Sequence gaps are valid because hot packets can be dropped.
Asteroid and bullet lifecycle packets flow through the same WebRTC active gameplay path as the other gameplay lanes.

Cross-lane ordering is not guaranteed between reliable lifecycle lanes and unreliable hot lanes. Clients must tolerate hot updates arriving before lifecycle create packets and after lifecycle delete packets.

## Client Transport Ownership

ClientConnectionService remains the public connection coordinator, but it does not directly own WebRTC peer and DataChannel lifecycle mechanics.

RealtimeTransportSession owns the active WebRTCTransport reference, transport construction through transport_factory, transport signal wiring, start, poll, close, clearing the active transport after close, and replacement after reconnect.

WebRTCTransport owns the peer connection, DataChannel setup, bounded receive draining, packet decoding, packet emission, outbound DataChannel writes, and transport-level diagnostics.

### Start, Poll, and Close Contract

RealtimeTransportSession.start() returns without action when an active transport is already assigned. Otherwise it creates a transport through transport_factory, wires offer, ICE, failure, ready, and packet signals, then calls WebRTCTransport.start().

RealtimeTransportSession.poll() delegates only while an active transport exists.

RealtimeTransportSession.close() closes the active transport once and clears the stored transport reference. Polling after close is therefore a no-op until a later start creates a replacement transport.

Tests or alternate composition paths that need normal startup must provide transport_factory rather than preassigning transport. A preassigned transport represents an already-active session and intentionally prevents start() from creating or starting another transport.

### Reconnect and Replacement

Connection teardown closes the previous session transport. A later connection uses transport_factory to create a fresh WebRTCTransport, rewires its signals, and starts it. Packets emitted by the replacement transport re-enter the same dispatcher and gameplay application path as packets from the original transport.

Transport replacement does not replace gameplay protocol state ownership. RealtimePacketPipeline separately owns the active RealtimeRouter, gameplay readiness, packet application, and protocol reset.

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

## Control and Resync Boundary

There is no physical `sr.control` WebRTC gameplay DataChannel in the current implementation.

The current recovery/control packet families are logical packet families:

```text
resync_request
resync_required
```

Do not document `control` as a current physical gameplay channel unless a real physical channel is implemented.

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
