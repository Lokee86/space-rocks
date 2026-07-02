## Gameplay Packets

Parent index: [Protocol](./!INDEX.md)

## Purpose

This document describes the current lane-native gameplay realtime packet protocol between the Godot client and the Go game server.

It covers client-originated gameplay requests, server-originated lane gameplay output families, `event_batch`, `player_pause_state`, packet authority, source-of-truth files, runtime routing, and the implementation paths that consume the gameplay packet contract.

## Overview

Gameplay packets are the realtime WebSocket messages used after a client is connected to the game server and, for gameplay mutation, attached to an active game player.

The protocol is server-authoritative:

```text
client sends input or request intent
-> game-server inbound routing classifies the packet
-> active room/game instance receives the packet
-> game simulation mutates authoritative state
-> owning server paths build one-off outputs such as player_pause_state
-> outbound networking writes encoded server packets
-> client receives and applies server-owned lane state
```

The client owns packet emission, local input collection, target-selection intent, viewport config reporting, and presentation after receiving server lane packets. The game server owns acceptance, validation, simulation mutation, pause state, respawn validity, scoring, lives, damage, pickups, spawning, lane packet projection, and presentation event production.

## Canonical realtime protocol

Detailed lane metadata, sequencing, baselines, deltas, control-lane resync packet behavior, and transport lifecycle belong in [Realtime WebSocket Protocol](realtime-websocket-protocol.md).

The current generated recovery packet families are resync_request and resync_required; control is the lane category, not a generated packet family.

This doc summarizes gameplay packet ownership and the high-level packet families only.

## Packet families

Active server-to-client gameplay packet families are:

```text
world_full / world_delta
overlay_full / overlay_delta
session_full / session_delta
event_batch
player_pause_state
resync_request / resync_required
```

Current packet families are lane-native, with `event_batch` carrying presentation events separately from world, overlay, and session lanes. World, overlay, and session lane packets use server-owned numeric wire quantization before delivery. Compact JSON aliases may be applied at the active outbound encode boundary, and the client expands them before normal lane routing. See [Realtime WebSocket Protocol](realtime-websocket-protocol.md) for the quantization details.

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
world entity updates
= id

overlay receiver updates
= self_id

session player updates
= id

session lifecycle updates
= player_id
```

`world_delta`, `overlay_delta`, and `session_delta` are field-delta aware for update arrays. `event_batch` is not a field-delta lane; it remains transient presentation event delivery. `player_pause_state` remains a separate same-session packet and is not part of lane delta delivery.

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
= lane projection, numeric wire quantization, field-delta comparison, sparse delta serialization, and compact alias preparation before packetcodec JSON encoding

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
NetworkClient.poll
-> PacketCodec.decode
-> NetworkClient.packet_received
-> ClientConnectionService
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
= active entity presentation state for ships, asteroids, bullets, pickups

overlay lane
= local-player HUD-facing presentation state such as score, lives, cooldowns, and loadout facts

session lane
= durable match-local player session state and lifecycle-oriented read models

event_batch
= transient presentation events delivered separately from baseline/delta state lanes
```

`player_pause_state` remains a separate same-session packet and should be treated as a current packet family, not as part of lane event or world-state delivery.

`event_batch` is transient event delivery, not a field-delta lane.

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
WebSocket read/write loops
packet-family routing order
server packet JSON encode/decode handoff
session current room and current game player context
lane packet write timing
packet metric logging/write observations
```
Realtime projection owns lane candidate construction, send-plan records, sparse delta omission, compact alias preparation, and current byte-budget planning inputs; networking writes and observes encoded results.

### Game-server simulation

Game-server simulation owns:

```text
input application
respawn behavior
pause mutation
target selection and clearing
lane packet projection inputs
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

Relevant verification areas now include lane packet routing/application, sparse delta omission, quantized wire values, compact alias mapping, lane state application, presentation adapters, and event batch behavior.

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
services/game-server/internal/protocol/realtime/wire_packets.go
services/game-server/internal/protocol/realtime/compact_wire_packet.go
services/game-server/internal/protocol/realtime/quantize/
services/game-server/internal/protocol/realtime/quantize_world.go
services/game-server/internal/protocol/realtime/quantized_records.go
services/game-server/internal/protocol/realtime/active.go
services/game-server/internal/networking/packetmetrics/
services/game-server/internal/game/
```

## Related docs

* [Protocol](./!INDEX.md)
* [Game Server](../services/game-server/!INDEX.md)
* [Client](../services/client/!INDEX.md)
* [Gameplay State Application](../services/client/gameplay-runtime/gameplay-state-application.md)
* [Realtime WebSocket Protocol](realtime-websocket-protocol.md)
* [Lane Packet Projection](../services/game-server/simulation/runtime/lane-packet-projection.md)
* [Packet Schemas](../data/packet-schemas.md)