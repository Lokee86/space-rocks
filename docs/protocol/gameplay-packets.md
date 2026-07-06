## Gameplay Packets

Parent index: [Protocol](./!INDEX.md)

## Purpose

This document describes the current lane-native gameplay realtime packet protocol between the Godot client and the Go game server.

It covers client-originated gameplay requests, server-originated lane gameplay output families, `event_batch`, `player_pause_state`, packet authority, source-of-truth files, runtime routing, and the implementation paths that consume the gameplay packet contract.

## Overview

Gameplay packets are the realtime packets used after a client is connected to the game server and, for gameplay mutation, attached to an active game player. Client-originated gameplay intent currently travels over the WebSocket session/control path, while server-originated active gameplay lane output travels over reliable/ordered WebRTC gameplay DataChannels.

The protocol is server-authoritative:

```text
client sends input or request intent
-> game-server inbound routing classifies the packet
-> active room/game instance receives the packet
-> game simulation mutates authoritative state
-> owning server paths build one-off outputs such as player_pause_state
-> outbound networking delivers queued one-off packets over WebSocket and active gameplay lane packets over reliable/ordered WebRTC gameplay DataChannels
-> client receives and applies server-owned lane state
```

The client owns packet emission, local input collection, target-selection intent, viewport config reporting, and presentation after receiving server lane packets. The game server owns acceptance, validation, simulation mutation, pause state, respawn validity, scoring, lives, damage, pickups, spawning, lane-native realtime projection, and presentation event production.

## Canonical realtime protocol

Detailed lane metadata, sequencing, baselines, deltas, control-lane resync packet behavior, and transport lifecycle belong in [Realtime WebSocket Protocol](realtime-websocket-protocol.md).

The current generated recovery packet families are resync_request and resync_required; control is the lane category, not a generated packet family.

This doc summarizes gameplay packet ownership and the high-level packet families only.

## Packet families

Active server-to-client gameplay packet families are:

```text
world_full / world_delta
asteroid_delta
bullet_delta
overlay_full / overlay_delta
session_full / session_delta
event_batch
player_pause_state
resync_request / resync_required
```

Current packet families are lane-native, with `event_batch` carrying transient presentation-event delivery separately from world, overlay, and session state lanes. World, overlay, and session lane packets use server-owned numeric wire quantization before delivery. `event_batch` now also goes through compact output encoding, and the client expands compact aliases and tuple-packed records before normal lane routing. See [Realtime WebSocket Protocol](realtime-websocket-protocol.md) for the quantization details.

Current lane delta behavior:

```text
create arrays
= full records

update arrays
= identity key plus changed fields only

delete arrays
= IDs
```

Empty delta sections are omitted from emitted `world_delta`, `overlay_delta`, and `session_delta` packets. Missing delta sections mean empty or no-op, not delete or clear. Missing fields inside present update records mean unchanged. Meaningful false and zero values inside present records remain meaningful. `session_delta` omits `total_asteroids` when unchanged; `total_asteroids: 0` remains meaningful when present.

Current update identity keys are:

```text
world ship and pickup updates
= id

asteroid_delta asteroid movement updates
= id

bullet_delta bullet movement updates
= id

overlay receiver updates
= self_id

session player updates
= id

session lifecycle updates
= player_id
```

`world_delta`, `overlay_delta`, and `session_delta` are field-delta aware for update arrays. `event_batch` is not a field-delta lane; it remains transient presentation-event delivery. Known event records are explicitly shaped from `EventState` into sparse wire records for known event types, and compact event aliases and tuple-packed event records are transport details that expand before gameplay presentation code consumes them. `player_pause_state` remains a separate same-session packet and is not part of lane delta delivery.

Detailed lane metadata, baseline, sequencing, and field-delta semantics belong in [Realtime WebSocket Protocol](realtime-websocket-protocol.md).


## Protocol authority

Packet schema authority lives in:

```text
shared/packets/gameplay.toml
shared/packets/outputs.toml
docs/data/packet-schemas.md
```

Generated packet code is output only and should not be edited by hand.

Runtime behavior authority is split:

```text
client outbound flow
= builds and sends generated gameplay packet dictionaries

game-server inbound routing
= classifies packet type and forwards to the active authoritative game instance

game-server realtime projection
= lane projection, numeric wire quantization, field-delta comparison, subtractive asteroid/bullet movement splitting, sparse delta serialization, and compact alias preparation before packetcodec JSON encoding

client gameplay runtime
= routes lane packets into lane states, baseline readiness, presentation adapters, and event application
```

The client does not own authoritative confirmation. A client request is confirmed only when reflected by server output such as lane packets, `player_pause_state`, room snapshots, or presentation events.

## Client-to-server gameplay packets

Client-originated gameplay packets remain request/intention packets such as:

```text
input
client_config
respawn
pause_request
set_target_player_request
select_target_at_position_request
clear_target_request
```

These are still schema-driven gameplay packets, and they route alongside the current lane-native output families.

## Client inbound gameplay application

The active client inbound gameplay path is:

```text
WebRTCTransport receives DataChannel text for active gameplay lane packets
-> PacketCodec.decode
-> ClientConnectionService._handle_webrtc_transport_packet
-> ServerPacketDispatcher / ServerPacketRouter classify packet
-> ClientConnectionService routes lane packets through RealtimeRouter.route_lane_packet(packet)
-> RealtimeRouter applies lane state/readiness
-> ClientConnectionService emits gameplay_packet_received(packet)
-> SessionNetworkController
-> GameplaySessionController.handle_gameplay_packet
-> presentation adapters
-> EventBatchApplier
```

`RealtimeRouter` applies inbound lane state before `GameplaySessionController` handles the packet for acceptance and presentation fanout. Presentation flow continues through the current lane adapters and `event_batch` application.

## Lane ownership

Current packet-family ownership is:

```text
world lane
= ships, pickups, asteroid/bullet lifecycle creates/deletes

asteroids lane
= regular asteroid movement updates

bullets lane
= regular bullet movement updates

overlay lane
= local-player HUD-facing presentation state such as score, lives, cooldowns, and loadout facts

session lane
= durable match-local player session state and lifecycle-oriented read models

event_batch
= transient presentation events delivered separately from baseline/delta state lanes
```

`player_pause_state` remains a separate same-session packet and should be treated as a current packet family, not as part of lane event or world-state delivery.

`event_batch` is transient presentation-event delivery, not a field-delta lane. It is not authoritative state, and it does not use baselines, deltas, or chunks.

## Event delivery

The important rule is:

```text
projection may inspect or copy pending presentation events
active send/write path is the drain point
```

Projection and shadow/inspection paths must not implicitly drain the event lane.

## Service responsibilities

### Client

The client owns:

```text
input collection
outbound gameplay packet construction
outbound packet send attempts
inbound packet classification after decode
lane state maintenance
baseline readiness tracking
world sync and presentation adapter fanout
HUD, audio, effects, match-end presentation, and devtools readouts
```

The client does not own gameplay authority, lane packet contents, respawn validity, score/lives authority, or server event production.

### Game-server networking

Game-server networking owns:

```text
WebSocket read loop, queued WebSocket writes, and active WebRTC gameplay delivery handoff
packet-family routing order
server packet JSON encode/decode handoff
session current room and current game player context
lane packet write timing
encoded packet write observations, debug packet wire logs, and non-empty per-tick write summaries
```
Realtime projection owns lane candidate construction, send-plan records, sparse delta omission, compact alias preparation, hot asteroid/bullet movement splitting, and current byte-budget planning inputs; networking delivers encoded active gameplay lane packets over reliable/ordered WebRTC gameplay DataChannels and emits the active debug wire logs plus non-empty per-tick write summaries after successful writes.

### Game-server simulation

Game-server simulation owns:

```text
input application
respawn behavior
pause mutation
target selection and clearing
lane-native realtime projection inputs
player session state
active avatar state
lifecycle classification
projectile, asteroid, pickup, and event projection inputs
presentation event queueing
```

## Validation and testing

Packet schema validation remains:

```bash
data-sync -validate -packets
data-sync -diff -packets -go -gds
data-sync -push -packets -go -gds
data-sync -check -packets -go -gds
```

Relevant verification areas now include lane-native packet routing/application, sparse delta omission, quantized wire values, compact alias mapping, tuple-packed record expansion, lane state application, presentation adapters, and event_batch behavior.

## Code map

Packet sources and generated outputs:

```text
shared/packets/gameplay.toml
shared/packets/outputs.toml
tools/data_sync/
services/game-server/internal/game/packets.go
services/game-server/internal/game/runtime/packets_generated.go
services/game-server/internal/protocol/realtime/packets_generated.go
client/scripts/generated/networking/packets/packets.gd
```

Client inbound lane-native gameplay application:

```text
client/scripts/networking/webrtc/webrtc_transport.gd
client/scripts/networking/client_connection_service.gd
client/scripts/networking/inbound/server_packet_dispatcher.gd
client/scripts/networking/inbound/server_packet_router.gd
client/scripts/session/session_network_controller.gd
client/scripts/session/gameplay_session_controller.gd
client/scripts/protocol/realtime/
client/scripts/world/world_sync.gd
client/scripts/gameplay/events/
client/scripts/gameplay/effects/
```

Game-server outbound gameplay projection:

```text
services/game-server/internal/networking/websocket_write.go
services/game-server/internal/networking/webrtc_transport.go
services/game-server/internal/protocol/realtime/wire_packets.go
services/game-server/internal/protocol/realtime/compact_wire_packet.go
services/game-server/internal/protocol/realtime/quantize/
services/game-server/internal/protocol/realtime/quantize_world.go
services/game-server/internal/protocol/realtime/quantized_records.go
services/game-server/internal/protocol/realtime/active.go
services/game-server/internal/networking/packetmetrics/
services/game-server/internal/game/
```
`packetmetrics/` remains a helper/support seam here and should not be read as current `realtime lane metric` log output.

## Related docs

* [Protocol](./!INDEX.md)
* [Game Server](../services/game-server/!INDEX.md)
* [Client](../services/client/!INDEX.md)
* [Gameplay State Application](../services/client/gameplay-runtime/gameplay-state-application.md)
* [Realtime WebSocket Protocol](realtime-websocket-protocol.md)
* [Lane Packet Projection](../services/game-server/simulation/runtime/lane-packet-projection.md)
* [Packet Schemas](../data/packet-schemas.md)

## Notes

This doc stays at the gameplay packet family and ownership boundary. Detailed lane metadata, wire behavior, and transport sequencing remain canonical in `realtime-websocket-protocol.md`.

