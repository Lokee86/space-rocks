# Presentation Bridge

Parent index: [Gameplay Runtime](./!INDEX.md)

`PresentationBridge` owns the boundary between refreshed realtime gameplay state and frame-coalesced client presentation.

It is configured by `GameplaySessionController`, driven by packet-ingress notifications from `RealtimePacketPipeline`, and flushed once per gameplay frame before normal gameplay composition processing.

Packet ingress is the only input path for this bridge; it does not consume lane-specific presentation signals directly.

## Ownership

`PresentationBridge` owns:

- Activation and reset state for the current gameplay session.
- Pending gameplay-presentation state.
- Coalescing multiple routed gameplay notifications into one frame flush.
- Retaining pending presentation while required gameplay lane baselines are not ready.
- Reading the latest `RealtimePresentationState` at flush time.
- Using the directly injected `world_sync` for world presentation while resolving HUD, event, local lifecycle, and devtools targets through `GameplayComposition`.
- Obtaining the local lifecycle flow through `GameplayComposition` for each ready presentation flush.
- Calling `PresentationAdapter` for lane-native presentation fanout.
- Building the devtools gameplay read model through `DevtoolsLaneStateAdapter`.
- Applying devtools gameplay state through `GameplayComposition`.
- Gameplay event-batch presentation diagnostics owned by this handoff.

## Does Not Own

`PresentationBridge` does not own:

- WebSocket or WebRTC transport.
- Server packet decoding or classification.
- Gameplay packet application.
- Realtime lane routing.
- Baseline tracking or resynchronization.
- Gameplay readiness calculation.
- Session admission or room-state transitions.
- Player pause packets.
- Debug-status or debug-shape control packets.
- World, HUD, event, or devtools presentation implementation.
- Normal gameplay composition processing.

## Dependencies

`GameplaySessionController` constructs the bridge and configures it with:

```text
RealtimePacketPipeline
PresentationAdapter
GameplayComposition
WorldSync
logger Callable
```

`RealtimePacketPipeline` provides the historically named `gameplay_packet_applied` routed-packet notification, gameplay readiness, and the current `RealtimePresentationState`.

`PresentationAdapter` performs stateless lane-native fanout.

`GameplaySessionController` injects `GameplayComposition.world_sync` directly into `PresentationBridge`; `WorldSync` is the world presentation target. `GameplayComposition` continues to provide HUD, event/local lifecycle, and devtools entry points.

## Applied Packet Handoff

The packet-ingress path is:

```text
ServerPacketDispatcher
-> ClientInboundCoordinator typed realtime binding
-> RealtimePacketPipeline typed apply entry point
-> RealtimeRouter routes the packet; lifecycle packets enter LifecycleLaneGate for immediate apply / queue / reject
-> WorldLaneApplier validates and mutates accepted lifecycle payloads
-> RealtimePresentationState is refreshed
-> RealtimePacketPipeline.gameplay_packet_applied(packet)
-> PresentationBridge.handle_gameplay_packet(packet)
```

The bridge receives a routed-packet notification after `RealtimePacketPipeline` has completed routing and refreshed `RealtimePresentationState`. The notification does not prove that the particular lifecycle packet mutated state: it may have been queued for a matching world baseline or rejected by `LifecycleLaneGate`.

`SessionNetworkController` and `GameplaySessionController` do not relay generic gameplay packet notifications in this path, and `PresentationBridge` is not fed by lane-specific `ClientConnectionService` signals.

## Lifecycle

### Configuration

```text
GameplaySessionController.configure(...)
-> construct PresentationAdapter
-> construct PresentationBridge
-> configure PresentationBridge with pipeline, adapter, composition, world sync, and logger
```

The configured bridge remains inactive until the gameplay session is activated.

### Activation

```text
GameplaySessionController.begin_accepting_gameplay_packets()
-> accepts_gameplay_packets = true
-> PresentationBridge.activate()
```

`begin_accepting_gameplay_packets()` activates the bridge.

Once active, routed gameplay notifications may mark presentation pending.

An inactive bridge does not mark presentation pending.

### Routed Packet Notification

```text
`RealtimePacketPipeline.gameplay_packet_applied(packet)`
-> PresentationBridge.handle_gameplay_packet(packet)
-> collect relevant diagnostics
-> mark presentation pending
```

Current gameplay readiness does not prevent the bridge from recording pending presentation.

For lifecycle packets, this notification may mark presentation pending while the lifecycle packet itself is queued or rejected. A later matching `world_full` can cause `RealtimeRouter` to drain and apply the queued lifecycle packet, and the bridge remains correct because it presents the latest refreshed `RealtimePresentationState` at flush time.

Readiness is evaluated when the pending state is flushed.

### Frame Flush

```text
GameplaySessionController._process(delta)
-> read RealtimePacketPipeline.is_gameplay_ready()
-> propagate readiness into GameplayComposition
-> PresentationBridge.flush_pending()
-> GameplayComposition.process(delta, readiness)
```

`flush_pending()` behaves as follows:

```text
inactive
-> no fanout

nothing pending
-> no fanout

pending and gameplay not ready
-> no fanout
-> pending remains set

pending and gameplay ready
-> read latest RealtimePresentationState
-> use the injected world sync and resolve the remaining targets through GameplayComposition
-> obtain GameplayLocalLifecycleFlow through GameplayComposition
-> PresentationAdapter.fanout_lane_states(..., local_lifecycle_flow)
-> build and apply devtools gameplay state
-> clear pending
```

Pending state is cleared only after a successful fanout.

### Reset

```text
GameplaySessionController.reset()
-> accepts_gameplay_packets = false
-> PresentationBridge.reset()
-> GameplayComposition.reset()
```

Bridge reset deactivates the bridge, clears pending presentation, and clears bridge-owned per-session diagnostic state.

This prevents presentation from one gameplay session from flushing into another.

## Coalescing

The bridge coalesces routed gameplay notifications until the next frame flush.

```text
packet A routed / notification received
-> pending = true

packet B routed / notification received
-> pending remains true

packet C routed / notification received
-> pending remains true

next gameplay frame
-> one flush of the latest coherent RealtimePresentationState
```

The bridge does not retain packet dictionaries as visual snapshots.

It reads the newest `RealtimePresentationState` from `RealtimePacketPipeline` when the frame flush occurs. It does not retain packet dictionaries as resulting visual snapshots. This keeps lane presentation coherent when several lane packets are routed between rendered frames, including when a later matching `world_full` mutates lifecycle state that was previously queued.

## Readiness

Readiness ownership is separated from presentation ownership:

```text
RealtimePacketPipeline.is_gameplay_ready()
= determines whether required gameplay lane baselines are available

PresentationBridge
= retains pending presentation until readiness is true

PresentationAdapter
= performs stateless lane fanout
```

`PresentationAdapter` does not own a presentability latch and does not expose `is_presentable()`, `can_fanout()`, `has_fanned_out()`, or `mark_fanned_out()`.

## Presentation Targets

During a successful flush, the bridge resolves:

* World presentation through directly injected `WorldSync` from `GameplayComposition.world_sync`.
* Overlay and session HUD presentation through `GameplayComposition.gameplay_hud_flow`.
* Applied event presentation through `GameplayComposition.get_event_lifecycle_flow()`.
* Local lifecycle presentation through `GameplayComposition.get_local_lifecycle_flow()`, delegated through the shell and flow composer.
* Devtools gameplay presentation through `GameplayComposition.apply_devtools_gameplay_state(...)`.

The bridge owns the order of these calls, but the target systems own their implementation.

`GameplayLocalLifecycleFlow` is passed as the fifth argument to `PresentationAdapter.fanout_lane_states(...)` during every ready flush. The bridge does not reconstruct local lifecycle state itself.

## Ordering Invariant

Presentation flush occurs before normal gameplay composition processing in the same frame.

```text
apply all received gameplay packets
-> coalesce presentation pending state
-> begin gameplay frame
-> flush latest presentation state
-> process gameplay composition
```

This ordering is covered by focused bridge and gameplay-session-controller unit tests.

## Related Docs

* [Gameplay state application](./gameplay-state-application.md)
* [Gameplay session lifecycle](./gameplay-session-lifecycle.md)
* [Runtime composition](./runtime-composition.md)
* [Runtime processing](./runtime-processing.md)
* [Inbound packet routing](../networking-flow/inbound-packet-routing.md)
* [HUD and gameplay UI](../hud-and-gameplay-ui.md)

## Notes

The bridge is an orchestration seam, not an alternate realtime state store.

`RealtimePacketPipeline` remains authoritative for refreshed client realtime state and gameplay readiness. Presentation consumers receive the latest state after pipeline routing has completed; `gameplay_packet_applied` is a routed notification, not a per-packet mutation receipt.