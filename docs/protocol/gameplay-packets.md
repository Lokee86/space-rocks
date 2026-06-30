## Gameplay Packets

Parent index: [Protocol](./!INDEX.md)

## Purpose

This document describes the current lane-native gameplay realtime packet protocol between the Godot client and the Go game server.

It covers client-originated gameplay requests, server-originated world/overlay/session lane packets, `event_batch`, `player_pause_state`, packet authority, source-of-truth files, runtime routing, and the implementation paths that consume the gameplay packet contract.

## Overview

Gameplay packets are the realtime WebSocket messages used after a client is connected to the game server and, for gameplay mutation, attached to an active game player.

The protocol is server-authoritative:

```text
client sends input or request intent
-> game-server inbound routing classifies the packet
-> active room/game instance receives the packet
-> game simulation mutates authoritative state
-> outbound networking projects lane packets or pause output
-> client receives and applies server-owned lane state
```

The client owns packet emission, local input collection, target-selection intent, viewport config reporting, and presentation after receiving server lane packets. The game server owns acceptance, validation, simulation mutation, pause state, respawn validity, scoring, lives, damage, pickups, spawning, lane packet projection, and presentation event production.

## Packet families

Active server-to-client gameplay packet families are:

```text
world_full / world_delta
overlay_full / overlay_delta
session_full / session_delta
event_batch
player_pause_state
resync / control / debug where configured
```

The old combined `state` gameplay packet is no longer the active runtime path.

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

game-server realtime projection
= builds authoritative lane packets for each player/session

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

These are still schema-driven gameplay packets, but they are separate from the removed combined gameplay-state output path.

## Client inbound gameplay application

The active client inbound gameplay path is:

```text
NetworkClient.poll
-> PacketCodec.decode
-> NetworkClient.packet_received
-> ClientConnectionService
-> ServerPacketDispatcher
-> SessionNetworkController
-> GameplaySessionController
-> realtime router
-> world / overlay / session lane states
-> GameplayReadiness baseline sync tracking
-> presentation adapters
-> EventBatchApplier
```

The client no longer routes gameplay presentation through `GameplayStatePacketReader`, combined gameplay-state normalization, or a single gameplay-state readiness flag.

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

`player_pause_state` remains a separate same-session packet and should not be treated as part of the old combined gameplay packet removal.

## Event delivery

Gameplay presentation events are now described as `event_batch` delivery rather than embedded `StatePacket.events` delivery.

The important rule is:

```text
projection may inspect/copy pending presentation events
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
packet metrics and budgeting
```

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

Relevant verification areas now include lane packet routing, lane state application, presentation adapters, and event batch behavior.

## Code map

Packet sources and generated outputs:

```text
shared/packets/gameplay.toml
shared/packets/outputs.toml
tools/data_sync/
services/game-server/internal/game/packets.go
services/game-server/internal/game/runtime/packets_generated.go
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
services/game-server/internal/protocol/realtime/
services/game-server/internal/networking/packetmetrics/
services/game-server/internal/game/
```

## Related docs

* [Protocol](./!INDEX.md)
* [Game Server](../services/game-server/!INDEX.md)
* [Client](../services/client/!INDEX.md)
* [Gameplay State Application](../services/client/gameplay-runtime/gameplay-state-application.md)
* [Lane Packet Projection](../services/game-server/simulation/runtime/state-packet-projection.md)

## Notes

Historical combined `StatePacket`, `GameplayStatePacketReader`, and single gameplay-state normalization/apply paths are no longer the active gameplay runtime design.
