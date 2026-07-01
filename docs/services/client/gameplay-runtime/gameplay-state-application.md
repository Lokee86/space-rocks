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
-> ClientConnectionService routes lane packets through RealtimeRouter.route_lane_packet(packet)
-> RealtimeRouter applies lane state/readiness
-> ClientConnectionService emits gameplay_packet_received(packet)
-> SessionNetworkController receives gameplay_packet_received
-> GameplaySessionController.handle_gameplay_packet performs acceptance/presentation fanout
-> presentation adapters
-> runtime consumers
```

The client applies lane packets through `RealtimeRouter` and current gameplay runtime adapters rather than a combined dictionary flow.

Presentation state consumes server-owned quantized wire values as received. Omitted delta fields still mean unchanged, not cleared. Client presentation and devtools comparisons should expect quantized values rather than raw simulation precision. Quantization does not change gameplay authority, which remains server-owned.

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
* Applying world lane state to world sync.
* Applying overlay lane state to HUD/local presentation.
* Applying session lane state to HUD and session-owned presentation.
* Applying event batches through the event batch applier.
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

The client lane application surface consumes server lane gameplay packets and turns them into presentation state after RealtimeRouter has already applied the inbound lane state.

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

The event path uses `EventBatchApplier` for `event_batch` delivery.

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
* `client/scripts/protocol/realtime/overlay_lane_state.gd`
* `client/scripts/protocol/realtime/session_lane_state.gd`
* `client/scripts/protocol/realtime/world_presentation_adapter.gd`
* `client/scripts/protocol/realtime/overlay_presentation_adapter.gd`
* `client/scripts/protocol/realtime/session_presentation_adapter.gd`
* `client/scripts/protocol/realtime/event_batch_applier.gd`
* `client/scripts/protocol/realtime/event_presentation_adapter.gd`
* `client/scripts/protocol/realtime/gameplay_readiness.gd`
* `client/scripts/protocol/realtime/realtime_router.gd`

## Tests

Relevant client tests include:

* `client/tests/unit/protocol/realtime/test_lane_protocol_routing.gd`
* `client/tests/unit/protocol/realtime/test_gameplay_readiness.gd`
* `client/tests/unit/protocol/realtime/test_world_lane_applier.gd`
* `client/tests/unit/protocol/realtime/test_overlay_session_lane_applier.gd`
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

This document describes the active lane-native gameplay presentation path.

Current gameplay application follows lane-adapter flow and event_batch delivery only.

