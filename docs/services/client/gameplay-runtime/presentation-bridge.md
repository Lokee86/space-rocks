# Presentation Bridge

Parent index: [Gameplay Runtime](./!INDEX.md)

`PresentationBridge` owns the boundary between applied realtime gameplay state and frame-coalesced client presentation.

It is configured by `GameplaySessionController`, driven by semantic applied-packet notifications from `RealtimePacketPipeline`, and flushed once per gameplay frame before normal gameplay composition processing.

## Ownership

`PresentationBridge` owns:

- Activation and reset state for the current gameplay session.
- Pending gameplay-presentation state.
- Coalescing multiple applied gameplay packets into one frame flush.
- Retaining pending presentation while required gameplay lane baselines are not ready.
- Reading the latest `RealtimePresentationState` at flush time.
- Resolving world, HUD, and event presentation targets through `GameplayComposition`.
- Calling `PresentationAdapter` for lane-native presentation fanout.
- Building the devtools gameplay read model through `DevtoolsLaneStateAdapter`.
- Applying devtools gameplay state through `GameplayComposition`.
- Triggering alive-presentation restoration through `GameplayComposition`.
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
logger Callable
```

`RealtimePacketPipeline` provides applied gameplay notifications, gameplay readiness, and the current `RealtimePresentationState`.

`PresentationAdapter` performs stateless lane-native fanout.

`GameplayComposition` provides the concrete runtime presentation targets and focused presentation entry points.

## Applied Packet Handoff

The completed inbound gameplay path is:

```text
ClientConnectionService
-> ServerPacketDispatcher
-> RealtimePacketPipeline.apply_packet(packet)
-> RealtimeRouter applies the packet
-> RealtimePresentationState is refreshed
-> RealtimePacketPipeline.gameplay_packet_applied(packet)
-> PresentationBridge.handle_gameplay_packet(packet)
```

The bridge receives a semantic notification only after the gameplay packet has been applied.

`SessionNetworkController` and `GameplaySessionController` do not relay generic gameplay packet notifications in this path.

## Lifecycle

### Configuration

```text
GameplaySessionController.configure(...)
-> construct PresentationAdapter
-> construct PresentationBridge
-> configure PresentationBridge with pipeline, adapter, composition, and logger
```

The configured bridge remains inactive until gameplay packet acceptance begins.

### Activation

```text
GameplaySessionController.begin_accepting_gameplay_packets()
-> accepts_gameplay_packets = true
-> PresentationBridge.activate()
```

Once active, applied gameplay notifications may mark presentation pending.

### Applied Packet Notification

```text
RealtimePacketPipeline.gameplay_packet_applied(packet)
-> PresentationBridge.handle_gameplay_packet(packet)
-> collect relevant diagnostics
-> mark presentation pending
```

Current gameplay readiness does not prevent the bridge from recording pending presentation.

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
-> resolve presentation targets
-> PresentationAdapter.fanout_lane_states(...)
-> build and apply devtools gameplay state
-> restore alive presentation from realtime state
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

The bridge coalesces applied gameplay notifications until the next frame flush.

```text
packet A applied
-> pending = true

packet B applied
-> pending remains true

packet C applied
-> pending remains true

next gameplay frame
-> one flush of the latest coherent RealtimePresentationState
```

The bridge does not retain packet dictionaries as visual snapshots.

It reads the newest applied state from `RealtimePacketPipeline` when the frame flush occurs. This keeps lane presentation coherent when several lane packets are applied between rendered frames.

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

* World presentation through `GameplayComposition.gameplay_shell_flow.runtime_context.world_sync`.
* Overlay and session HUD presentation through `GameplayComposition.gameplay_hud_flow`.
* Applied event presentation through `GameplayComposition.get_event_lifecycle_flow()`.
* Devtools gameplay presentation through `GameplayComposition.apply_devtools_gameplay_state(...)`.
* Alive-state restoration through `GameplayComposition.restore_alive_presentation_from_realtime_state(...)`.

The bridge owns the order of these calls, but the target systems own their implementation.

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

`RealtimePacketPipeline` remains authoritative for applied client realtime state and gameplay readiness. Presentation consumers receive state only after pipeline application has completed.