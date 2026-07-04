# Gameplay State Application

Parent index: [Gameplay Runtime](./!INDEX.md)

## Purpose

This document describes the active lane-native client gameplay presentation path.

It covers realtime packet routing, lane state ownership, baseline readiness, presentation adapters, event batch application, and the boundary between gameplay runtime orchestration and world rendering.

## Overview

Gameplay presentation begins after the client networking layer receives a realtime server packet and routes it by packet family.

The active path is:

```text
NetworkClient receives/decodes packet
-> ClientConnectionService receives packet
-> ServerPacketDispatcher / ServerPacketRouter classify packet
-> RealtimeRouter.route_lane_packet(packet)
-> RealtimeRouter applies lane state/readiness
-> ClientConnectionService emits gameplay_packet_received(packet)
-> SessionNetworkController receives gameplay_packet_received
-> GameplaySessionController gates on accepts_gameplay_packets and gameplay readiness
-> PresentationAdapter.fanout_lane_states(...)
-> WorldPresentationAdapter -> WorldSync.apply_world_lane_state(...)
-> OverlayPresentationAdapter -> GameplayHudFlow.apply_overlay_lane_state(...)
-> SessionPresentationAdapter -> GameplayHudFlow.apply_session_lane_state(...)
-> EventPresentationAdapter -> GameplayEventLifecycleFlow / GameplayEventFlow
-> GameplayComposition.restore_alive_presentation_from_realtime_router(...)
-> DevtoolsLaneStateAdapter builds a separate devtools readmodel
```

The client applies lane packets through `RealtimeRouter` and current gameplay runtime adapters rather than a retired aggregate `GameplayStateApplyFlow` path or combined normalized gameplay-state dictionary flow.

Client inbound packets may carry compact aliases and quantized integer wire values. `PacketCodec.decode` and `CompactLanePacket` expand compact aliases and tuple arrays back to readable dictionaries before lane appliers or event presentation flows run, but `CompactLanePacket` does not decode quantized numeric values. `WorldLaneApplier` decodes quantized world values before storing or merging world lane state. `EventBatchApplier` decodes quantized event values before forwarding or storing applied events. `WorldLaneState` stores values exactly as supplied and does not own quantize decoding. Overlay and session values are decoded before presentation/devtools consumption through the realtime quantize helpers and presentation adapters. Event_batch arrives from compact wire output, but it is expanded before client event appliers and presentation flows consume it, so gameplay presentation systems should consume readable long-key event dictionaries rather than compact aliases. Devtools read models decode only the lane values they explicitly pass through `RealtimeQuantize`. Client-facing presentation should expect decoded values with quantized precision loss, not raw simulation precision. Tests should assert the current decode boundary explicitly instead of assuming raw simulation floats. Full packets still replace or initialize complete lane state. Client lane appliers treat missing sparse delta section fields as empty arrays or no-op, missing fields inside present update records as unchanged, and missing `total_asteroids` in `session_delta` as unchanged. A present `total_asteroids: 0` remains meaningful. Quantization does not change gameplay authority, which remains server-owned.

## Code root

```text
client/scripts/protocol/realtime/
client/scripts/session/gameplay_session_controller.gd
client/scripts/session/session_network_controller.gd
```

The realtime client package owns lane state, readiness tracking, and presentation adapters. SessionNetworkController and GameplaySessionController own the handoff after inbound networking has already classified the packet and RealtimeRouter has already applied lane state.

## Responsibilities

The active client gameplay application path owns:

* Realtime packet-family routing after decode.
* Maintaining lane state objects for world, overlay, and session data.
* Tracking required lane baseline sync before gameplay is considered ready.
* Gating presentation fanout behind both `accepts_gameplay_packets` and gameplay readiness.
* Applying world lane state to world sync.
* Applying overlay lane state to HUD/local presentation.
* Applying session lane state to HUD and session-owned presentation.
* Applying event batches through the event batch applier.
* Restoring alive/respawn-facing presentation from current lane state after fanout.
* Keeping devtools gameplay read models separate from primary gameplay presentation.

## Does not own

The lane-native client path does not own:

* WebSocket transport.
* Packet schema source-of-truth files.
* Authoritative simulation outcomes.
* Packet generation.
* Room/lobby authority.
* Server event production.
* Match rules, scoring rules, respawn validity, or pause authority.

## Domain roles

The client lane application surface consumes server lane gameplay packets and turns them into presentation state after `RealtimeRouter` has already applied the inbound lane state.

The client owns transient lane presentation state only. It does not persist authoritative gameplay state.

Generated packet constants and builders come from the packet schema pipeline.

## Protocols and APIs

The client runtime consumes server lane gameplay packets, but packet shape and transport behavior are owned by protocol and data docs.

Authoritative gameplay outcomes are owned by the server.

For packet-family and transport detail, see:

* [Gameplay packets](../../../protocol/gameplay-packets.md)
* [Realtime WebSocket Protocol](../../../protocol/realtime-websocket-protocol.md)
* [Packet Schemas](../../../data/packet-schemas.md)

## Data ownership

The client maintains transient lane presentation state only.

It does not persist authoritative gameplay state.

Generated packet constants and builders come from the packet schema pipeline.

## Presentation adapters

Presentation adapters are the packet-to-runtime boundary for gameplay presentation.

#### Sparse delta handling

World lane sparse compatibility lives in `client/scripts/protocol/realtime/world_lane_applier.gd`. Overlay sparse compatibility lives in `client/scripts/protocol/realtime/overlay_lane_state.gd`. Session sparse compatibility lives in `client/scripts/protocol/realtime/session_lane_state.gd`. Decode helpers live in `client/scripts/protocol/realtime/realtime_quantize.gd`.

Sparse field omission is a server wire-map behavior, and the client compatibility rule is to tolerate missing fields without treating them as deletes or clears. Those appliers treat missing sparse delta section fields as empty arrays or no-op, preserve missing fields inside present update records as unchanged, and preserve meaningful zero values when they are actually emitted. In `session_delta`, missing `total_asteroids` means unchanged, while a present `total_asteroids: 0` remains meaningful.

Current adapter roles are:

```text
WorldPresentationAdapter
= applies world lane state to WorldSync

OverlayPresentationAdapter
= applies overlay lane state to GameplayHudFlow

SessionPresentationAdapter
= applies session lane state to GameplayHudFlow and related session presentation

EventPresentationAdapter
= applies event batches to event/effects presentation
```

The event path uses `EventBatchApplier` for `event_batch` delivery. Compact aliases such as `eb`, `ev`, `ei`, `bb`, `shd`, and `dmg` are wire details and should not leak into client presentation code.

## Active handoff seams

The current handoff boundaries are:

```text
ClientConnectionService
= owns lane packet routing into RealtimeRouter and emits gameplay_packet_received(packet) after lane state/readiness has already been updated

SessionNetworkController
= forwards gameplay_packet_received(packet) into GameplaySessionController

GameplaySessionController
= owns accepts_gameplay_packets, checks gameplay readiness, fans out lane state, builds the devtools readmodel, and triggers alive-presentation restoration

PresentationAdapter
= fans out presentable lane state into world, overlay, session, and event presentation consumers

GameplayComposition
= routes alive-presentation restoration into shell/runtime seams without owning lane application itself

DevtoolsLaneStateAdapter
= builds devtools readmodel dictionaries separately from primary gameplay presentation fanout
```

## World rendering boundary

World entity rendering is not owned by gameplay application flow.

The active runtime boundary is:

```text
world lane state
-> WorldPresentationAdapter
-> WorldSync.apply_world_lane_state(...)
```

`WorldSync` owns entity-family synchronization, interpolation, and rendering behavior after that point.

## Code map

Primary runtime path:

* `client/scripts/session/session_network_controller.gd` - inbound routing handoff from networking.
* `client/scripts/session/gameplay_session_controller.gd` - gameplay packet acceptance and presentation application.
* `client/scripts/protocol/realtime/` - lane states, readiness, adapters, and appliers.
* `client/scripts/world/world_sync.gd` - world entity sync/render boundary.
* `client/scripts/shell/gameplay_hud_flow.gd` - HUD-facing presentation consumers.
* `client/scripts/gameplay/events/` - event consumers and presentation flows.
* `client/scripts/gameplay/effects/` - effects consumers fed by gameplay presentation.
* `client/scripts/devtools/` - devtools lane-state consumers if enabled.

Key lane-native files:

* `client/scripts/protocol/realtime/world_lane_state.gd`
* `client/scripts/protocol/realtime/world_lane_applier.gd` - decodes quantized world records and applies full/delta world packets.
* `client/scripts/protocol/realtime/overlay_lane_state.gd`
* `client/scripts/protocol/realtime/overlay_lane_applier.gd` - routes overlay full/delta packets into overlay lane state.
* `client/scripts/protocol/realtime/session_lane_state.gd`
* `client/scripts/protocol/realtime/session_lane_applier.gd` - routes session full/delta packets into session lane state.
* `client/scripts/protocol/realtime/realtime_quantize.gd` - owns client decode helpers for quantized realtime wire values before presentation/devtools consumption. Compact expansion and asteroid ID rehydration happen before `WorldLaneApplier`; quantized x/y/scale decoding happens inside world lane application through the realtime quantize helpers.
* `client/scripts/protocol/realtime/world_presentation_adapter.gd`
* `client/scripts/protocol/realtime/overlay_presentation_adapter.gd`
* `client/scripts/protocol/realtime/session_presentation_adapter.gd`
* `client/scripts/protocol/realtime/event_batch_applier.gd`
* `client/scripts/protocol/realtime/event_presentation_adapter.gd`
* `client/scripts/protocol/realtime/gameplay_readiness.gd`
* `client/scripts/protocol/realtime/realtime_router.gd`
* `client/scripts/protocol/realtime/devtools_lane_state_adapter.gd`

## Tests

Relevant client tests include:

* `client/tests/unit/protocol/realtime/test_lane_protocol_routing.gd`
* `client/tests/unit/protocol/realtime/test_gameplay_readiness.gd`
* `client/tests/unit/protocol/realtime/test_world_lane_applier.gd` - `test_world_delta_treats_missing_sparse_sections_as_empty_noop` covers missing sparse delta section fields as empty/no-op for world lane application.
* `client/tests/unit/protocol/realtime/test_overlay_session_lane_applier.gd` - `test_overlay_delta_treats_missing_sparse_sections_as_empty_noop` and `test_session_delta_treats_missing_sparse_sections_and_total_asteroids_as_empty_noop` cover missing sparse delta section fields as empty/no-op for overlay and session lane application.
* `client/tests/unit/protocol/realtime/test_event_batch_and_resync.gd`
* `client/tests/unit/protocol/realtime/test_lane_native_presentation_adapters.gd`
* `client/tests/unit/protocol/realtime/test_devtools_lane_state_adapter.gd`
* `client/tests/unit/test_gameplay_session_controller.gd`

## Related docs

* [Gameplay Runtime](./!INDEX.md)
* [World Sync](../world-sync/!INDEX.md)
* [Runtime composition](runtime-composition.md)
* [Gameplay session lifecycle](gameplay-session-lifecycle.md)
* [Gameplay packets](../../../protocol/gameplay-packets.md)
* [Realtime WebSocket Protocol](../../../protocol/realtime-websocket-protocol.md)
* [Packet Schemas](../../../data/packet-schemas.md)

## Notes

This document is the canonical client lane-native gameplay application doc.

Current gameplay application follows lane-adapter flow and `event_batch` delivery only.
