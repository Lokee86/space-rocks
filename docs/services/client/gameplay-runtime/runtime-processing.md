# Runtime Processing

Parent index: [Gameplay Runtime](./!INDEX.md)

## Purpose

This document describes the current per-frame client gameplay runtime processing path.

It covers how gameplay runtime work is ticked after gameplay composition exists, how world interpolation is updated, how HUD runtime updates are called, how devtools/input/spectate processing is sequenced, and what this processing path deliberately does not own.

## Overview

Runtime processing is the client-side frame loop for active gameplay presentation and orchestration.

It is not the authoritative simulation tick. The server owns gameplay simulation, collision outcomes, scoring, lives, death, respawn validity, and match lifecycle. The client runtime processing path only advances local presentation and client-owned runtime helpers between applied lane packets.

The frame path starts in `GameplaySessionController._process(delta)`. The controller reads the current gameplay readiness from `RealtimePacketPipeline.is_gameplay_ready()`, propagates that readiness into `GameplayComposition`, calls `PresentationBridge.flush_pending()`, and then runs `GameplayComposition.process(delta, readiness)`.

The current per-frame order is:

```text
1. GameplaySessionController._process(delta)
2. Read readiness from `RealtimePacketPipeline.is_gameplay_ready()`
3. `GameplayComposition.set_required_lane_baselines_synced(readiness)`
4. `PresentationBridge.flush_pending()`
5. `GameplayComposition.process(delta, readiness)`
```

Gameplay composition consumes the latest applied lane state before normal gameplay processing so presentation work is coalesced for the frame and applied once. When multiple presentation updates are queued before the frame tick, the composed runtime keeps the latest state and applies that consolidated result rather than replaying each intermediate update.

`GameplayComposition.process` currently ticks three client-side lanes:

```text
1. GameplayShellFlow.process(delta)
2. DevToolsSessionFlow.process(delta)
3. GameplayPresentationFlow.process(delta, required_lane_baselines_synced)
```

`GameplayShellFlow` delegates to `GameplayFlowComposer`, which delegates focused per-frame gameplay work to `GameplayProcessFlow`.

`GameplayProcessFlow` owns the inner gameplay processing order:

```text
1. GameplayRuntimeContext.process(delta)
2. ServerHitboxOverlayFlow.process()
3. GameplayRuntimeTickFlow.process(delta)
4. GameplayDevtoolsContext.process(required_lane_baselines_synced)
5. GameplayInputContext.process(required_lane_baselines_synced)
6. GameplaySpectateContext.process()
```

`GameplayRuntimeContext.process(delta)` currently delegates world interpolation to `WorldSync.interpolate(delta)`.

`GameplayRuntimeTickFlow.process(delta)` currently delegates HUD runtime updates to `GameplayHudFlow.update(delta)`.

This structure keeps frame processing as an ordered orchestration seam. It should not become the owner of HUD internals, input policy, devtools behavior, spectate rules, world entity sync details, or server gameplay authority.

## Code root

* `client/scripts/`

## Responsibilities

* Tick client-owned gameplay runtime work once per Godot frame.
* Preserve a stable processing order for gameplay presentation helpers.
* Pass gameplay readiness to flows that must behave differently before required lane baselines are present.
* Tick world interpolation through `GameplayRuntimeContext`.
* Tick server hitbox overlay presentation through `ServerHitboxOverlayFlow`.
* Tick HUD runtime work through `GameplayRuntimeTickFlow`.
* Tick gameplay devtools context processing.
* Tick gameplay input context processing.
* Tick spectate context processing.
* Keep per-frame runtime orchestration separate from lane packet application.
* Keep per-frame runtime orchestration separate from authoritative simulation.

## Does not own

* Server simulation ticks.
* Server gameplay authority.
* Collision, damage, score, lives, respawn validity, or match-over decisions.
* Raw websocket polling or packet decoding.
* Lane packet application.
* World entity node creation, cleanup, and interpolation details.
* HUD widget internals.
* Input mapping or input action ownership.
* Target selection orchestration.
* Devtools command authority.
* Spectate menu and target-selection rules beyond ticking the spectate context.
* Match-end lifecycle decisions.
* Durable player data.

## Domain roles

### Presentation frame loop

Runtime processing advances client presentation between authoritative server state updates.

This includes interpolation, HUD ticking, devtools presentation refresh, input-process hooks, and spectate-process hooks.

### Ordered runtime coordinator

`GameplayProcessFlow` is the narrow ordering seam for focused runtime processors. It does not own the details inside those processors.

### State-aware processing bridge

Some processing lanes receive gameplay readiness so they can avoid acting as if lane-applied world state is available before required baselines have been applied.

### World-sync tick bridge

Runtime processing is the frame-loop caller for world sync interpolation. World sync owns the actual interpolation details.

## Protocols and APIs

### Frame entry path

Gameplay frame processing enters through Godot `_process` on the gameplay session controller:

```text
GameplaySessionController._process(delta)
-> Read readiness from `RealtimePacketPipeline.is_gameplay_ready()`
-> `GameplayComposition.set_required_lane_baselines_synced(readiness)`
-> `PresentationBridge.flush_pending()`
-> `GameplayComposition.process(delta, readiness)`
```

The readiness flag is read from `RealtimePacketPipeline.is_gameplay_ready()` and then passed through `GameplaySessionController`.

### Composition processing path

`GameplayComposition.process` performs top-level gameplay processing fanout:

```text
GameplayShellFlow.process(delta)
DevToolsSessionFlow.process(delta)
GameplayPresentationFlow.process(delta, required_lane_baselines_synced)
```

`DevToolsSessionFlow` is a separate devtools gameplay-session seam. `GameplayPresentationFlow` owns broader local presentation updates such as camera-facing presentation inputs. The gameplay shell owns the inner gameplay runtime processing path.

### Inner gameplay processing path

`GameplayShellFlow.process(delta)` delegates to `GameplayFlowComposer.process(delta, required_lane_baselines_synced)`.

`GameplayFlowComposer.process` delegates to `GameplayProcessFlow.process(delta, required_lane_baselines_synced)`.

The current `GameplayProcessFlow` order is:

```text
runtime_context.process(delta)
server_hitbox_overlay_flow.process()
runtime_tick_flow.process(delta)
devtools_context.process(required_lane_baselines_synced)
input_context.process(required_lane_baselines_synced)
spectate_context.process()
```

This order means world interpolation runs before HUD runtime ticking, devtools processing, input processing, and spectate processing.

### World interpolation API

`GameplayRuntimeContext.process(delta)` calls:

```gdscript
world_sync.interpolate(delta)
```

World sync then interpolates player/render-anchor state, projectiles, asteroids, and pickups.

Runtime processing does not directly interpolate entity nodes.

### HUD runtime API

`GameplayRuntimeTickFlow.process(delta)` calls:

```gdscript
hud_flow.update(delta)
```

HUD flow owns the details of HUD updates. Runtime processing only provides the per-frame call.

### State-aware process APIs

`GameplayDevtoolsContext.process(required_lane_baselines_synced)` and `GameplayInputContext.process(required_lane_baselines_synced)` receive the lane-readiness flag.

This keeps pre-readiness behavior explicit for flows that may depend on required lane-applied gameplay state being available.

## Data ownership

Runtime processing owns only transient frame-processing coordination.

It uses:

* `delta` from Godot `_process`.
* gameplay readiness from `RealtimePacketPipeline.is_gameplay_ready()`.
* references to composed runtime processors.
* client-owned runtime flow instances.
* client-owned presentation state inside downstream flows.

Runtime processing does not persist data.

Runtime processing does not own authoritative state.

Runtime processing does not mutate durable player records.

Runtime processing does not own packet schemas.

## Code map

### Frame entry and top-level processing

* `client/scripts/session/gameplay_session_controller.gd`
* `client/scripts/gameplay/gameplay_composition.gd`
* `client/scripts/shell/gameplay_shell_flow.gd`

### Runtime processing coordinator

* `client/scripts/gameplay/runtime/gameplay_process_flow.gd`
* `client/scripts/gameplay/runtime/gameplay_flow_composer.gd`
* `client/scripts/gameplay/runtime/gameplay_runtime_context.gd`
* `client/scripts/shell/gameplay_runtime_tick_flow.gd`

### Downstream processing lanes

* `client/scripts/world/world_sync.gd`
* `client/scripts/gameplay/debug/server_hitbox_overlay_flow.gd`
* `client/scripts/devtools/context/`
* `client/scripts/gameplay/input/gameplay_input_context.gd`
* `client/scripts/gameplay/spectate/gameplay_spectate_context.gd`
* `client/scripts/gameplay/presentation/gameplay_presentation_flow.gd`
* `client/scripts/shell/gameplay_hud_flow.gd`
* `client/scripts/ui/hud/`

### Non-ownership boundaries

* `client/scripts/gameplay/state/` and `client/scripts/protocol/realtime/` own lane packet application and readiness.
* `client/scripts/world/` owns world entity sync and interpolation details.
* `client/scripts/networking/` owns websocket transport, packet decoding, and packet dispatch.
* `services/game-server/internal/game/` owns authoritative gameplay simulation.
* `docs/devtools/client/` owns detailed client devtools documentation when those docs are filled in.

## Tests

Current related test coverage includes:

* `client/tests/unit/test_gameplay_session_controller.gd`
* `client/tests/unit/gameplay/test_gameplay_flow_composer.gd`
* `client/tests/unit/gameplay/debug/test_server_hitbox_overlay_flow.gd`
* `client/tests/unit/test_gameplay_input_context.gd`
* `client/tests/unit/gameplay/test_gameplay_alive_restore_flow.gd`
* `client/tests/unit/gameplay/test_gameplay_event_lifecycle_flow.gd`
* `client/tests/unit/test_world_sync.gd`
* `client/tests/unit/world/player_render/test_player_render_api.gd`
* `client/tests/unit/world/player_render/test_view_anchor_sync.gd`

Use the normal client GUT test run for verification after runtime-processing changes.

## Related docs

* [Gameplay Runtime](./!INDEX.md)
* [World Sync](../world-sync/!INDEX.md)
* [Gameplay state application](gameplay-state-application.md)
* [Runtime composition](runtime-composition.md)
* [Gameplay session lifecycle](gameplay-session-lifecycle.md)
* [Presentation bridge](presentation-bridge.md)
* [HUD and gameplay UI](../hud-and-gameplay-ui.md) - Client HUD and gameplay UI documentation.
* [Input and targeting](../input-and-targeting.md) - Client input and targeting documentation.

## Notes

`GameplayRuntimeTickFlow` currently lives under `client/scripts/shell/`, even though it participates in gameplay runtime processing.

`GameplayComposition.process` also ticks `DevToolsSessionFlow` and `GameplayPresentationFlow` outside the inner `GameplayProcessFlow` order. This document includes those calls because they are part of the current per-frame gameplay runtime path.

Current runtime ordering is stable: session readiness is sampled first from `RealtimePacketPipeline.is_gameplay_ready()`, readiness is propagated into `GameplayComposition`, `PresentationBridge.flush_pending()` runs before composition processing, and then frame-specific runtime helpers continue.

Server hitbox overlay processing is debug presentation. Runtime processing ticks it, but it does not make hitbox overlay behavior normal gameplay rendering authority.