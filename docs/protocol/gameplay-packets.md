---
author: brian
created: "2026-07-19"
document_id: 019f7d55-fb2c-7fd0-be25-9116c2d9a612
document_type: general
policy_exempt: false
summary: This document describes the current lane-native gameplay realtime packet protocol between the Godot client and the Go game server.
---
## Gameplay Packets

Parent index: [Protocol](./!INDEX.md)

## Purpose

This document describes the current lane-native gameplay realtime packet protocol between the Godot client and the Go game server.

It covers client-originated gameplay requests, server-originated lane gameplay output families, `event_batch`, `player_pause_state`, packet authority, source-of-truth files, runtime routing, and the implementation paths that consume the gameplay packet contract.

## Overview

Gameplay packets are the realtime packets used after a client is connected to the game server and, for gameplay mutation, attached to an active game player. Client-originated gameplay intent currently travels over the WebSocket session/control path, while server-originated active gameplay output travels over WebRTC gameplay DataChannels: `sr.world`, `sr.overlay`, `sr.session`, `sr.event`, `sr.ships.lifecycle`, `sr.asteroids.lifecycle`, and `sr.bullets.lifecycle` remain ordered/reliable, while `sr.ships`, `sr.asteroids`, and `sr.bullets` are unordered/unreliable hot-update lanes.

The protocol is server-authoritative:

```text
client sends input or request intent
-> game-server inbound routing classifies the packet
-> active room/game instance receives the packet
-> game simulation mutates authoritative state
-> owning server paths build one-off outputs such as player_pause_state
-> outbound networking delivers queued one-off packets over WebSocket and active gameplay lane packets over lane-specific WebRTC gameplay DataChannels
-> client receives and applies server-owned lane state
```

The client owns packet emission, local input collection, target-selection intent, viewport config reporting, and presentation after receiving server lane packets. The game server owns acceptance, validation, simulation mutation, pause state, respawn validity, scoring, lives, damage, pickups, spawning, lane-native realtime projection, and presentation event production.

## Canonical realtime protocol

Detailed lane metadata, sequencing, baselines, deltas, control-lane resync packet behavior, and transport lifecycle belong in [Realtime WebSocket Protocol](realtime-websocket-protocol.md).

The current generated recovery packet families are resync_request and resync_required; control is the lane category, not a generated packet family.

This doc summarizes gameplay packet ownership and the high-level packet families only. `resync_request` is a client-to-server WebSocket control packet; `resync_required` is a server-to-client WebSocket acknowledgment. They are not active server WebRTC lane outputs. Recovery fulls are ordinary required world, overlay, or session candidates on the existing reliable WebRTC lanes.

The client request carries `type`, `lane`, `baseline_id`, `sequence`, and `reason` and follows `BaselineTracker -> RealtimeRouter -> RealtimePacketPipeline -> ClientConnectionService -> ClientPacketSender -> generated resync_request_packet -> NetworkClient -> WebSocket`. The server validates active room/game/player context, queues the typed request with room and receiver context, writes `resync_required`, and invalidates the requested baseline only after a successful write. The next planner pass emits the lane's full recovery candidate. Recovery is limited to world, overlay, and session baselines; it does not provide hot-lane or lifecycle resync, retries, resend, reconnect recovery, or durable queues.

Lifecycle packet behavior at this boundary is: `ships_lifecycle`, `asteroids_lifecycle`, and `bullets_lifecycle` each maintain an independent strict lifecycle sequence; each packet must include explicit `lane`, `sequence`, `baseline_id`, `snapshot_id`, `snapshot_kind`, and `server_sent_msec` metadata; and `baseline_id` names the world baseline used to build the candidate. Hard-cap splitting may produce distinct chunks for one logical lifecycle sequence. The client assembles and validates the complete chunk series before `LifecycleLaneGate` sequence/baseline validation and mutation. Malformed, duplicate, mismatched, or interrupted series fail closed and request world resync. Sequence gaps are valid. Detailed gate bounds, drain ordering, and reset behavior remain canonical in the realtime protocol document.

## Packet families

Active server-to-client gameplay packet families are:

```text
world_full / world_delta
ship_delta
player_locator
asteroid_delta
bullet_delta
ships_lifecycle
asteroids_lifecycle
bullets_lifecycle
overlay_full / overlay_delta
session_full / session_delta
event_batch
player_pause_state
```

Current packet families are lane-native, with `event_batch` carrying transient presentation-event delivery separately from world, overlay, and session state lanes. `player_locator` carries a low-cadence replace-all snapshot of coarse active-player positions for clients whose recipient-specific interest set no longer includes every detailed ship. World, overlay, and session lane packets use server-owned numeric wire quantization before delivery. `event_batch` now also goes through compact output encoding, and the client expands compact aliases and tuple-packed records before normal lane routing. `ship_delta`, `asteroid_delta`, and `bullet_delta` are hot movement packets on unordered/unreliable lanes: they move existing detailed entities, they do not create new entities client-side, and lifecycle creates/deletes are handled by `ships_lifecycle`, `asteroids_lifecycle`, and `bullets_lifecycle`. `player_locator` also uses the unordered/unreliable `sr.ships` transport, but it updates a dedicated coarse locator read model rather than detailed ship state. `ship_delta`, `asteroid_delta`, and `bullet_delta` are high-priority, hot-supersedable movement candidates. They are not required lifecycle packets, and they may be dropped or replaced by newer movement state. When a hot movement update list would exceed the encoded packet cap, `ship_delta`, `asteroid_delta`, and `bullet_delta` may be emitted as multiple same-sequence chunks. These chunks still only move existing entities; lifecycle creates/deletes remain on the dedicated lifecycle lanes. See [Realtime WebSocket Protocol](realtime-websocket-protocol.md) for the quantization details.

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

ship_delta ship movement updates
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


`world_delta`, `overlay_delta`, and `session_delta` are field-delta aware for update arrays. `event_batch` is not a field-delta lane; it remains transient presentation-event delivery. Known event records are explicitly shaped from `EventState` into sparse wire records for known event types, and compact event aliases and tuple-packed event records are transport details that expand before gameplay presentation code consumes them. `ship_delta`, `asteroid_delta`, and `bullet_delta` are supersedable hot updates on unordered/unreliable lanes, while `ships_lifecycle`, `asteroids_lifecycle`, and `bullets_lifecycle` own lifecycle creates/deletes. `player_pause_state` remains a separate same-session packet and is not part of lane delta delivery.

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
= lane projection, numeric wire quantization, field-delta comparison, subtractive ship/asteroid/bullet movement splitting, sparse delta serialization, and compact alias preparation before packetcodec JSON encoding

client gameplay runtime
= routes lane packets into lane states, baseline readiness, presentation adapters, and event application
```

The client does not own authoritative confirmation. A client request is confirmed only when reflected by server output such as lane packets, `player_pause_state`, room snapshots, or presentation events. For hot `ship_delta`, `asteroid_delta`, and `bullet_delta` packets, `sequence` is an int; valid hot sequence values are finite, non-negative integer-valued numeric values. Missing, fractional, negative, non-finite, string, and boolean values are protocol-invalid. Sequence gaps remain valid, and same-sequence chunks are accepted only for distinct, valid `chunk_index` values whose `chunk_count` matches the count established for that lane sequence. Duplicate indices, inconsistent counts, malformed chunk metadata, and lower sequences are rejected; distinct chunks may arrive in any order, and ship, asteroid, and bullet tracking are independent.

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
set_view_target_request
clear_view_target_request
```

The view-target packets are spectate/control intent. They select the recipient camera anchor used by server network interest and do not mutate canonical gameplay targeting.

These are still schema-driven gameplay packets, and they route alongside the current lane-native output families.

## Client inbound gameplay application

The active client inbound gameplay path is:

```text
WebRTCTransport receives DataChannel text for active gameplay lane packets
-> PacketCodec.decode
-> RealtimeTransportSession._on_packet_received(packet)
-> ServerPacketDispatcher / ServerPacketRouter classify packet
-> ClientInboundCoordinator
-> RealtimePacketPipeline typed apply method for the packet family
-> RealtimePacketPipeline expands and validates the packet
-> RealtimeRouter.route_lane_packet(packet)
-> RealtimePresentationState refreshed
-> RealtimePacketPipeline.gameplay_packet_applied(packet)
-> PresentationBridge.handle_gameplay_packet(packet)
-> EventBatchApplier
```
Realtime lane packets are owned by the realtime inbound pipeline and are not re-emitted through the connection-service shell. `ClientInboundCoordinator` does not mutate gameplay state itself, but it owns dispatcher-consumer bindings for every realtime gameplay family. Those bindings invoke the typed `RealtimePacketPipeline` apply methods for world, ship, player-locator, asteroid, bullet, lifecycle, overlay, session, event, and resync packets. `RealtimePacketPipeline` and its owned `RealtimeRouter` own packet and lane-state application before `PresentationBridge` handles the semantic applied notification. Lifecycle packets route through `LifecycleLaneGate` for apply/queue/reject decisions, then `WorldLaneApplier` validates the accepted payload and mutates `WorldLaneState`. Presentation flow continues through gameplay composition and `event_batch` application.

Lifecycle packets are applied through RealtimePacketPipeline and its owned RealtimeRouter before presentation handoff, so entity existence and identity are established before session and presentation handling.

For accepted realtime gameplay packets, the ordering is transport receives, classification, pipeline application and validation, RealtimePresentationState refresh, gameplay_packet_applied, PresentationBridge.handle_gameplay_packet(packet), then later frame-coalesced presentation fanout when ready.

## Lane ownership

Current packet-family ownership is:

```text
world lane
= recipient-relevant ships, pickups, player/world presentation state, and full/bootstrap world snapshots

ships lane
= detailed `ship_delta` movement for recipient-relevant ships
= coarse `player_locator` snapshots for all active player identities at lower cadence

ships lifecycle lane
= recipient-relevant ship creates/deletes and reliable non-transform ship state

asteroids lane
= regular asteroid movement updates on unordered/unreliable hot lane sr.asteroids

asteroids lifecycle lane
= asteroid creates/deletes and initial asteroid presentation identity, including variant/size/scale

bullets lane
= regular bullet movement updates on unordered/unreliable hot lane sr.bullets

bullets lifecycle lane
= bullet/projectile creates/deletes and initial projectile identity, including owner, weapon_id, projectile_type, and torpedo identity

overlay lane
= local-player HUD-facing presentation state such as score, lives, cooldowns, and loadout facts

session lane
= durable match-local player session state and lifecycle-oriented read models

event_batch
= transient presentation events delivered separately from baseline/delta state lanes
```

`player_pause_state` remains a separate same-session packet and should be treated as a current packet family, not as part of lane event or world-state delivery.

Lifecycle defines detailed entity presentation existence. Hot lanes update known detailed entities only. `player_locator` is a separate coarse read model: it does not create ship nodes, does not define durable player existence, and remains available when detailed ship presentation leaves recipient interest.

Hot movement packets must never create entities. Unknown hot asteroid updates are ignored. Unknown hot bullet updates are buffered only where the client explicitly supports waiting for lifecycle create; hot updates after delete are ignored and must not resurrect removed entities.

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
Realtime projection owns lane candidate construction, send-plan records, sparse delta omission, compact alias preparation, hot ship/asteroid/bullet movement splitting, and current byte-budget planning inputs; networking delivers encoded active gameplay lane packets over ordered/reliable lanes for `sr.world`, `sr.overlay`, `sr.session`, `sr.event`, `sr.ships.lifecycle`, `sr.asteroids.lifecycle`, and `sr.bullets.lifecycle`, and unordered/unreliable hot-update lanes for `sr.ships`, `sr.asteroids`, and `sr.bullets`, and emits the active debug wire logs plus non-empty per-tick write summaries after successful writes.

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

Physical realtime-wire validation remains separate:

```bash
data-sync -validate -realtime-wire
data-sync -diff -realtime-wire -go -gds -json -docs
data-sync -push -realtime-wire -go -gds -json -docs
data-sync -check -realtime-wire -go -gds -json -docs
```

Relevant verification areas now include lane-native packet routing/application, player-locator routing and increasing-sequence replacement, lifecycle packet routing/application, recipient-interest entry/exit behavior, `test_lifecycle_lane_gate.gd` coverage for baseline and sequence policy, `client/tests/unit/networking/realtime/test_realtime_packet_pipeline.gd` reset coverage, server lifecycle wire metadata coverage in `wire_packets_test.go`, sparse delta omission, quantized wire values, generated descriptor-driven compact encoding/decoding, tuple-packed record expansion, lane state application, presentation adapters, lifecycle existence handling, and event_batch behavior.

## Code map

Packet sources and generated outputs:

```text
shared/packets/gameplay.toml
shared/packets/outputs.toml
shared/packets/realtime_wire.toml
shared/packets/generated/realtime_wire.json
docs/protocol/generated/realtime-wire-reference.md
tools/data_sync/
services/game-server/internal/game/packets.go
services/game-server/internal/game/runtime/packets_generated.go
services/game-server/internal/protocol/realtime/packets_generated.go
services/game-server/internal/protocol/realtimewire/generated.go
client/scripts/generated/networking/packets/packets.gd
client/scripts/generated/networking/realtime_wire_generated.gd
```

Client inbound lane-native gameplay application:

```text
client/scripts/networking/webrtc/webrtc_transport.gd
client/scripts/networking/inbound/client_inbound_coordinator.gd
client/scripts/networking/inbound/server_packet_dispatcher.gd
client/scripts/networking/inbound/server_packet_router.gd
client/scripts/networking/realtime/realtime_packet_pipeline.gd
client/scripts/protocol/realtime/realtime_router.gd
client/scripts/protocol/realtime/compact_lane_packet.gd
client/scripts/protocol/realtime/compact_wire_descriptor_index.gd
client/scripts/protocol/realtime/compact_wire_descriptor_ids.gd
client/scripts/protocol/realtime/compact_wire_descriptor_records.gd
client/scripts/protocol/realtime/compact_wire_descriptor_decoder.gd
client/scripts/protocol/realtime/realtime_quantize.gd
client/scripts/protocol/realtime/lifecycle_lane_gate.gd
client/scripts/protocol/realtime/baseline_tracker.gd
client/scripts/protocol/realtime/player_locator_applier.gd
client/scripts/protocol/realtime/player_locator_state.gd
client/scripts/protocol/realtime/world_lane_applier.gd
client/scripts/networking/realtime/realtime_presentation_state.gd
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
services/game-server/internal/protocol/realtime/planner.go
services/game-server/internal/protocol/realtime/lane_candidate_world.go
services/game-server/internal/protocol/realtime/lane_candidate_lifecycle.go
services/game-server/internal/protocol/realtime/network_interest.go
services/game-server/internal/protocol/realtime/player_locator.go
services/game-server/internal/protocol/realtime/lane_candidate_overlay.go
services/game-server/internal/protocol/realtime/lane_candidate_session.go
services/game-server/internal/protocol/realtime/lane_candidate_event.go
services/game-server/internal/protocol/realtime/candidate_types.go
services/game-server/internal/protocol/realtime/candidate_policy.go
services/game-server/internal/protocol/realtime/candidate_diagnostics.go
services/game-server/internal/protocol/realtime/wire_packets.go
services/game-server/internal/protocol/realtime/wire_reflect.go
services/game-server/internal/protocol/realtime/compact_wire_packet.go
services/game-server/internal/protocol/realtime/compact_wire_descriptor.go
services/game-server/internal/protocol/realtime/quantize/
services/game-server/internal/protocol/realtime/quantize_world.go
services/game-server/internal/protocol/realtime/quantize_overlay.go
services/game-server/internal/protocol/realtime/quantize_session.go
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
* [Realtime WebRTC Gameplay Transport](realtime-webrtc-gameplay-transport.md)
* [Realtime Compact Wire Mapping](../services/game-server/networking/realtime-compact-wire-mapping.md)
* [Generated Realtime Wire Reference](./generated/realtime-wire-reference.md)
* [Game Server Network Interest](../services/game-server/networking/network-interest.md)
* [Client Gameplay Presentation Flow](../services/client/presentation-flow/gameplay-presentation-flow.md)

## Notes

This doc stays at the gameplay packet family and ownership boundary. Detailed lane metadata, wire behavior, and transport sequencing remain canonical in `realtime-websocket-protocol.md`.
