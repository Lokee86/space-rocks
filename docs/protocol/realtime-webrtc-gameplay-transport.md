## Realtime WebRTC Gameplay Transport

Parent index: [Protocol](./!INDEX.md)

## Purpose

This document defines the current WebRTC transport boundary for active realtime gameplay packets.

It is the canonical protocol doc for physical gameplay DataChannels, lane-to-channel mapping, active gameplay readiness, server send routing, client receive routing, and the current reliable/ordered delivery policy.

## Ownership

This document owns:

```text
physical WebRTC gameplay DataChannel names
physical WebRTC gameplay DataChannel numeric IDs
logical realtime lane to physical channel mapping
current active gameplay transport split from WebSocket
current reliable/ordered gameplay channel policy
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
future unreliable/unordered delivery design
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
```

The current channel policy is:

```text
sr.world, sr.overlay, sr.session, and sr.event are negotiated ordered/reliable channels.
sr.asteroids and sr.bullets are negotiated unordered/unreliable channels with maxRetransmits=0.
sr.asteroids carries supersedable asteroid_updates only.
sr.bullets carries supersedable bullet_updates only.
lifecycle creates/deletes remain on sr.world.
late asteroid_delta and bullet_delta packets are rejected by monotonic sequence on the client.
sequence gaps are valid because hot packets can be dropped.
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

The server active gameplay path builds lane candidates, encodes each selected candidate, and writes each encoded packet to the matching WebRTC gameplay channel.

The server send boundary is:

```text
BuildActiveRealtimeResultForGame
-> selected realtime lane candidates
-> encode selected candidate packet
-> SendEncodedLaneJSON(candidate.Lane, encodedPacket)
-> physical WebRTC gameplay DataChannel
```

Active metadata advancement, event draining, and baseline persistence must only happen after the active WebRTC write path succeeds.

## Client Receive Boundary

The client active gameplay receive path is:

```text
WebRTCTransport receives DataChannel text
-> PacketCodec.decode expands compact aliases and validates the packet envelope
-> ClientConnectionService dispatches non-smoke WebRTC packets through ServerPacketDispatcher
-> RealtimeRouter applies lane packets
-> lane state is fanned out to gameplay presentation
```

Dedicated asteroid and bullet hot movement packets do not create independent rendered state. They merge into the same world presentation state used by gameplay rendering. Late asteroid_delta and bullet_delta packets are rejected on the client by monotonic sequence checks, and sequence gaps are valid because hot packets can be dropped.

## Hot Movement Split

Regular asteroid movement updates are emitted as:

```text
asteroid_delta on sr.asteroids
```

Regular bullet movement updates are emitted as:

```text
bullet_delta on sr.bullets
```

`world_delta` remains responsible for ships, pickups, and lifecycle/bootstrap/resync-safe asteroid and bullet creates/deletes. World serializer compatibility may still accept asteroid or bullet update sections, but regular active asteroid and bullet movement is split to the dedicated hot movement lanes.

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
payload fragmentation
hot-lane unreliable/unordered gameplay delivery
```

Compact JSON aliases, sparse delta omission, numeric quantization, tuple packing, and dedicated asteroid/bullet hot movement packets are implemented before the final WebRTC write boundary.

## Related Docs

* [Realtime WebSocket Protocol](realtime-websocket-protocol.md)
* [Gameplay Packets](gameplay-packets.md)
* [Realtime Compact Wire Mapping](../services/game-server/networking/realtime-compact-wire-mapping.md)
* [Outbound Message Flow](../services/game-server/networking/outbound-message-flow.md)
* [Lane Packet Projection](../services/game-server/simulation/runtime/lane-packet-projection.md)
