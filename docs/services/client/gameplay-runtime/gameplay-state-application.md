# Gameplay State Application

Parent index: [Gameplay Runtime](./!INDEX.md)

## Purpose

This document describes the completed lane-native client gameplay presentation path.

It covers realtime packet routing, lane state ownership, baseline readiness, the applied-state wrapper `RealtimePresentationState`, gameplay composition handoff, local lifecycle reconciliation, event batch application, and the boundary between gameplay runtime orchestration and world rendering.

## Overview

Gameplay presentation begins after the client networking layer receives a realtime server packet and routes it by packet family.

The current ingress path is:

```text
NetworkClient or WebRTCTransport receives and decodes packet
-> ClientConnectionService passes decoded packet into ServerPacketDispatcher
-> ServerPacketDispatcher classifies packet
-> ServerPacketDispatcher emits typed realtime signal
-> ClientInboundCoordinator routes the typed signal to the matching RealtimePacketPipeline entry point
-> RealtimePacketPipeline expands and validates the packet
-> RealtimeRouter applies the packet to lane state
-> RealtimePacketPipeline refreshes RealtimePresentationState
-> RealtimePacketPipeline.gameplay_packet_applied(packet)
-> PresentationBridge.handle_gameplay_packet(packet)
-> PresentationBridge records pending applied notification
-> GameplaySessionController._process(delta)
-> read RealtimePacketPipeline.is_gameplay_ready()
-> propagate readiness into GameplayComposition
-> PresentationBridge.flush_pending()
-> PresentationBridge obtains world, HUD, event, and local lifecycle presentation targets
-> PresentationAdapter.fanout_lane_states(presentation_state, world_sync, gameplay_hud_flow, event_lifecycle_flow, local_lifecycle_flow)
-> GameplayLocalLifecycleFlow reconciles local lifecycle presentation
-> EventPresentationAdapter drains event output
-> PresentationBridge orchestrates devtools gameplay state through gameplay composition
-> GameplayComposition.process(delta, readiness)
-> DevtoolsLaneStateAdapter builds a separate devtools readmodel
```

The client applies lane packets through `RealtimePacketPipeline`. The pipeline owns the active `RealtimeRouter` and invokes it for lane-specific mutation rather than using the retired aggregate `GameplayStateApplyFlow` path or a combined normalized gameplay-state dictionary flow.

`ClientConnectionService` passes decoded transport packets into `ServerPacketDispatcher`; `ClientInboundCoordinator` owns the post-classification realtime consumer binding. `ClientConnectionService` does not choose lane handlers, own lane diagnostics, or re-emit realtime lane signals.

`PresentationBridge` activation enables scheduling for gameplay-packet presentation, while `RealtimePacketPipeline` gameplay readiness gates fanout into presentation targets.

`RealtimePacketPipeline` applies realtime gameplay packets regardless of gameplay-session activation state.

`WorldLaneApplier` now applies `apply_asteroids_lifecycle`, `apply_bullets_lifecycle`, `apply_asteroid_delta`, and `apply_bullet_delta` so lifecycle packets define existence before hot movement updates are merged.

## Code root

```text
client/scripts/protocol/realtime/
client/scripts/session/gameplay_session_controller.gd
client/scripts/session/session_network_controller.gd
```

The lane-native client package owns lane state and readiness tracking. Inbound networking application completes through RealtimePacketPipeline, which refreshes `RealtimePresentationState` before PresentationBridge and session presentation handoff begin.

## Responsibilities

The active client gameplay application path owns:

* Realtime packet-family routing after decode.
* Maintaining lane state objects for world, overlay, and session data.
* Tracking required lane baseline sync before gameplay is considered ready.
* `PresentationBridge` activation enables scheduling for gameplay-packet presentation.
* `RealtimePacketPipeline` gameplay readiness gates fanout into presentation targets.
* Applying realtime gameplay packets regardless of gameplay-session activation state.
* Applying lifecycle and hot movement packets into WorldLaneState.
* Routing current lane state into gameplay composition for world, HUD, session, and event presentation.
* Reconstructing active, pending-respawn, and eliminated local presentation from authoritative world/session state.
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

Gameplay composition consumes `RealtimePresentationState` after `RealtimePacketPipeline` has completed application and refreshed the applied-state wrapper.

The client owns transient lane presentation state only. It does not persist authoritative gameplay state.

Generated packet constants and builders come from the packet schema pipeline.

## Protocols and APIs

The client runtime consumes server lane gameplay packets, but packet shape and transport behavior are owned by protocol and data docs.

Authoritative gameplay outcomes are owned by the server.

For packet-family and transport detail, see:

* [Gameplay packets](../../../protocol/gameplay-packets.md)
* [Realtime WebSocket Protocol](../../../protocol/realtime-websocket-protocol.md)
* [Packet Schemas](../../../data/packet-schemas.md)
* [Realtime WebRTC Gameplay Transport](../../../protocol/realtime-webrtc-gameplay-transport.md)

## Data ownership

The client maintains transient lane presentation state only.

It does not persist authoritative gameplay state.

Generated packet constants and builders come from the packet schema pipeline.

## Presentation adapters

`RealtimePresentationState` is the packet-to-runtime boundary for gameplay presentation.

#### Sparse delta handling

World lane sparse compatibility lives in `client/scripts/protocol/realtime/world_lane_applier.gd`. Overlay sparse compatibility lives in `client/scripts/protocol/realtime/overlay_lane_state.gd`. Session sparse compatibility lives in `client/scripts/protocol/realtime/session_lane_state.gd`. Decode helpers live in `client/scripts/protocol/realtime/realtime_quantize.gd`.

Sparse field omission is a server wire-map behavior, and the client compatibility rule is to tolerate missing fields without treating them as deletes or clears. Those appliers treat missing sparse delta section fields as empty arrays or no-op, preserve missing fields inside present update records as unchanged, and preserve meaningful zero values when they are actually emitted. In `session_delta`, missing `total_asteroids` means unchanged, while a present `total_asteroids: 0` remains meaningful.

Current applied-state fanout roles are:

```text
WorldPresentationAdapter
= applies world lane state to WorldSync

OverlayPresentationAdapter
= applies overlay lane state to GameplayHudFlow

SessionPresentationAdapter
= applies session lane state to GameplayHudFlow and related session presentation

EventPresentationAdapter
= applies event batches to event/effects presentation

GameplayLocalLifecycleFlow
= reconciles active, pending-respawn, and eliminated local presentation from world/session lane state
```

The event path uses `EventBatchApplier` for `event_batch` delivery. Compact aliases such as `eb`, `ev`, `ei`, `bb`, `shd`, and `dmg` are wire details and should not leak into client presentation code.

`PresentationAdapter` decodes session state once, passes the decoded state to `SessionPresentationAdapter`, then reuses that decoded session state when it calls `GameplayLocalLifecycleFlow.apply_lane_state(world_lane_state, decoded_session_state, self_id)`. Local lifecycle reconciliation runs after world, overlay, and session lane presentation and before `EventPresentationAdapter` drains event output.

`GameplayLocalLifecycleFlow` reconstructs the local presentation state from authoritative lane data. It handles active, pending-respawn, and eliminated status without becoming an authority for lifecycle outcomes. Event presentation remains a separate best-effort immediate-effects path.

## Active handoff seams

The current handoff boundaries are:

```text
ClientConnectionService
= coordinates the public networking handoff before ClientInboundCoordinator routes typed realtime dispatcher signals to the matching RealtimePacketPipeline apply method

ServerPacketDispatcher
= classifies inbound packets before gameplay packets reach the realtime pipeline

RealtimePacketPipeline
= owns compact packet expansion, gameplay packet validation, the active RealtimeRouter, gameplay readiness, protocol reset, lane-routing invocation, gameplay_packet_applied(packet), and realtime gameplay packet application regardless of gameplay-session activation state

RealtimeRouter
= owns lane-specific state mutation, baseline and sequence handling, and lane-state storage beneath RealtimePacketPipeline

PresentationBridge
= owns gameplay_packet_applied notification handling, pending/coalescing state, readiness-gated flush, latest-state retrieval, and orchestration of lane presentation, local lifecycle presentation, and devtools-state adaptation through composition

GameplaySessionController
= owns accepts_gameplay_packets, bridge activation/reset/flush scheduling, frame sequencing, control routing, input routing, reset, and session exits

ClientConnectionService
= owns network ingress coordination only; it does not own lane handlers, lane diagnostics, or gameplay packet application

PresentationAdapter
= fans out presentable lane state into world, overlay, session, local lifecycle, and event presentation consumers; decodes session state once for session and local lifecycle consumers

GameplayComposition
= exposes the local lifecycle flow through shell/runtime composition without owning lane application itself

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
* `client/scripts/session/gameplay_session_controller.gd` - PresentationBridge lifecycle, frame scheduling, input/control routing, and session exits.
* `client/scripts/networking/realtime/` - home of `RealtimePacketPipeline`, gameplay packet application, and readiness gating.
* `client/scripts/protocol/realtime/` - lane states, readiness, adapters, and appliers.
* `client/scripts/world/world_sync.gd` - world entity sync/render boundary.
* `client/scripts/shell/gameplay_hud_flow.gd` - HUD-facing presentation consumers.
* `client/scripts/gameplay/events/` - event consumers and presentation flows.
* `client/scripts/gameplay/lifecycle/gameplay_local_lifecycle_flow.gd` - local active, pending-respawn, and eliminated presentation reconciliation.
* `client/scripts/gameplay/effects/` - effects consumers fed by gameplay presentation.
* `client/scripts/devtools/` - devtools lane-state consumers if enabled.

Key lane-native files:

* `client/scripts/networking/realtime/` - home of `RealtimePacketPipeline`.
* `client/scripts/protocol/realtime/world_lane_state.gd`
* `client/scripts/protocol/realtime/world_lane_applier.gd` - decodes quantized world records and applies full/delta world packets.
* `client/scripts/protocol/realtime/world_lane_applier.gd` - applies `apply_asteroids_lifecycle`, `apply_bullets_lifecycle`, `apply_asteroid_delta`, and `apply_bullet_delta` into `WorldLaneState`.
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
* `client/scripts/protocol/realtime/world_lane_applier.gd` - `test_world_delta_treats_missing_sparse_sections_as_empty_noop` covers missing sparse delta section fields as empty/no-op for world lane application.
* `client/tests/unit/protocol/realtime/test_world_lane_applier.gd` - lifecycle coverage for `apply_asteroids_lifecycle` and `apply_bullets_lifecycle`.
* `client/tests/unit/protocol/realtime/test_lane_protocol_routing.gd` - lifecycle routing coverage.
* `client/tests/unit/protocol/realtime/test_overlay_session_lane_applier.gd` - `test_overlay_delta_treats_missing_sparse_sections_as_empty_noop` and `test_session_delta_treats_missing_sparse_sections_and_total_asteroids_as_empty_noop` cover missing sparse delta section fields as empty/no-op for overlay and session lane application.
* `client/tests/unit/protocol/realtime/test_event_batch_and_resync.gd`
* `client/tests/unit/protocol/realtime/test_lane_native_presentation_adapters.gd`
* `client/tests/unit/protocol/realtime/test_devtools_lane_state_adapter.gd`
* `client/tests/unit/test_gameplay_session_controller.gd`
* `client/tests/unit/world/test_world_sync.gd` - lifecycle-created render fanout coverage.

## Related docs

* [Gameplay Runtime](./!INDEX.md)
* [World Sync](../world-sync/!INDEX.md)
* [Runtime composition](runtime-composition.md)
* [Gameplay session lifecycle](gameplay-session-lifecycle.md)
* [Gameplay packets](../../../protocol/gameplay-packets.md)
* [Realtime WebSocket Protocol](../../../protocol/realtime-websocket-protocol.md)
* [Packet Schemas](../../../data/packet-schemas.md)
* [Realtime WebRTC Gameplay Transport](../../../protocol/realtime-webrtc-gameplay-transport.md)
* [Presentation Bridge](presentation-bridge.md)

## Notes

This document is the canonical client lane-native gameplay application doc.

`RealtimePacketPipeline` owns lane application and refreshes `RealtimePresentationState` before any presentation orchestration occurs.

`PresentationBridge` owns deferred gameplay-packet presentation orchestration after pipeline application and before presentation targets are updated. The ready flush passes the local lifecycle flow as the fifth `PresentationAdapter.fanout_lane_states(...)` argument.

`GameplaySessionController` activates, resets, and flushes `PresentationBridge` with the gameplay session lifecycle.
