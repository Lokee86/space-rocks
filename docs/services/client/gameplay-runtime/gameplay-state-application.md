# Gameplay State Application

Parent index: [Gameplay Runtime](./!INDEX.md)

## Purpose

This document describes the active lane-native client gameplay presentation path.

It covers realtime packet routing, lane state ownership, baseline readiness, presentation adapters, event batch application, and the boundary between gameplay runtime orchestration and world rendering.

## Overview

Gameplay presentation begins after the client networking layer receives a realtime server packet and routes it by packet family.

The active path is:

```text
NetworkClient.packet_received
-> ClientConnectionService
-> ServerPacketDispatcher
-> SessionNetworkController
-> GameplaySessionController
-> realtime router
-> world / overlay / session lane states
-> GameplayReadiness baseline sync tracking
-> presentation adapters
-> runtime consumers
```

The client no longer normalizes one combined gameplay dictionary through `GameplayStatePacketReader` or `GameplayStateApplyFlow`.

Instead, gameplay presentation is split by lane:

```text
world lane
= entity presentation state for ships, asteroids, bullets, pickups

overlay lane
= local-player HUD presentation state such as score, lives, loadout, cooldowns

session lane
= durable match-local player session state and lifecycle read models

event batch
= transient presentation events applied through the event batch applier
```

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

## Baseline readiness

Gameplay readiness is baseline-sync based, not “received one combined gameplay state packet”.

Required readiness currently means the client has received the required baseline packets for:

```text
world
overlay
session
```

Once those baselines are synced, gameplay presentation and input readiness can proceed through the active runtime flow.

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

The event path uses `EventBatchApplier` rather than a combined state packet reader.

## World rendering boundary

World entity rendering is not owned by a combined gameplay-state fanout layer.

The active runtime boundary is:

```text
world lane state
-> WorldPresentationAdapter
-> WorldSync.apply_world_lane_state(...)
```

`WorldSync` owns entity-family synchronization, interpolation, and rendering behavior after that point.

## Code map

Primary runtime path:

* `client/scripts/session/session_network_controller.gd`
* `client/scripts/session/gameplay_session_controller.gd`
* `client/scripts/protocol/realtime/`
* `client/scripts/world/world_sync.gd`
* `client/scripts/shell/gameplay_hud_flow.gd`
* `client/scripts/gameplay/events/`

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

## Related docs

* [Gameplay Runtime](./!INDEX.md)
* [World Sync](../world-sync/!INDEX.md)
* [Runtime composition](runtime-composition.md)
* [Gameplay session lifecycle](gameplay-session-lifecycle.md)
* [Gameplay packets](../../../protocol/gameplay-packets.md)

## Notes

This document describes the active lane-native gameplay presentation path.

Any remaining references to `GameplayStatePacketReader`, combined gameplay-state fanout, or single-packet gameplay readiness are historical only and should not be treated as the active runtime design.
