# Runtime Processing

Parent index: [Gameplay Runtime](./!INDEX.md)

## Purpose

This document describes the current per-frame client gameplay runtime processing path.

It covers how gameplay runtime work is ticked after gameplay composition exists, how world interpolation is updated, how HUD runtime updates are called, how devtools/input/spectate processing is sequenced, and what this processing path deliberately does not own.

## Overview

Runtime processing is the client-side frame loop for active gameplay presentation and orchestration.

It is not the authoritative simulation tick. The server owns gameplay simulation, collision outcomes, scoring, lives, death, respawn validity, and match lifecycle. The client runtime processing path only advances local presentation and client-owned runtime helpers between authoritative state packets.

The frame path starts in `GameplaySessionController._process(delta)`. The controller asks `GameplayStateFlow` whether any gameplay state has been received, then calls `GameplayComposition.process(delta, has_received_state)`.

`GameplayComposition.process` currently ticks three client-side lanes:

```text
1. GameplayShellFlow.process(delta)
2. DevToolsSessionFlow.process(delta)
3. GameplayPresentationFlow.process(delta, has_received_gameplay_state)
```

`GameplayShellFlow` delegates to `GameplayFlowComposer`, which delegates focused per-frame gameplay work to `GameplayProcessFlow`.

`GameplayProcessFlow` owns the inner gameplay processing order:

```text
1. GameplayRuntimeContext.process(delta)
2. ServerHitboxOverlayFlow.process()
3. GameplayRuntimeTickFlow.process(delta)
4. GameplayDevtoolsContext.process(has_received_state)
5. GameplayInputContext.process(has_received_state)
6. GameplaySpectateContext.process()
```

`GameplayRuntimeContext.process(delta)` currently delegates world interpolation to `WorldSync.interpolate(delta)`. This is the bridge from gameplay runtime processing into world-sync rendering.

`GameplayRuntimeTickFlow.process(delta)` currently delegates HUD runtime updates to `GameplayHudFlow.update(delta)`.

This structure keeps frame processing as an ordered orchestration seam. It should not become the owner of HUD internals, input policy, devtools behavior, spectate rules, world entity sync details, or server gameplay authority.

## Code root

* `client/scripts/`

## Responsibilities

* Tick client-owned gameplay runtime work once per Godot frame.
* Preserve a stable processing order for gameplay presentation helpers.
* Pass `has_received_state` to flows that must behave differently before the first gameplay state packet.
* Tick world interpolation through `GameplayRuntimeContext`.
* Tick server hitbox overlay presentation through `ServerHitboxOverlayFlow`.
* Tick HUD runtime work through `GameplayRuntimeTickFlow`.
* Tick gameplay devtools context processing.
* Tick gameplay input context processing.
* Tick spectate context processing.
* Keep per-frame runtime orchestration separate from state-packet application.
* Keep per-frame runtime orchestration separate from authoritative simulation.

## Does not own

* Server simulation ticks.
* Server gameplay authority.
* Collision, damage, score, lives, respawn validity, or match-over decisions.
* Raw websocket polling or packet decoding.
* Gameplay state packet normalization.
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

Some processing lanes receive `has_received_state` so they can avoid acting as if gameplay state is available before the first authoritative state packet has been applied.

### World-sync tick bridge

Runtime processing is the frame-loop caller for world sync interpolation. World sync owns the actual interpolation details.

## Protocols and APIs

### Frame entry path

Gameplay frame processing enters through Godot `_process` on the gameplay session controller:

```text
GameplaySessionController._process(delta)
-> GameplayComposition.process(delta, has_received_state)
```

`has_received_state` is read from `GameplayStateFlow.has_received_state()`.

### Composition processing path

`GameplayComposition.process` performs top-level gameplay processing fanout:

```text
GameplayShellFlow.process(delta)
DevToolsSessionFlow.process(delta)
GameplayPresentationFlow.process(delta, has_received_gameplay_state)
```

`DevToolsSessionFlow` is a separate devtools gameplay-session seam. `GameplayPresentationFlow` owns broader local presentation updates such as camera-facing presentation inputs. The gameplay shell owns the inner gameplay runtime processing path.

### Inner gameplay processing path

`GameplayShellFlow.process(delta)` delegates to `GameplayFlowComposer.process(delta, has_received_state)`.

`GameplayFlowComposer.process` delegates to `GameplayProcessFlow.process(delta, has_received_state)`.

The current `GameplayProcessFlow` order is:

```text
runtime_context.process(delta)
server_hitbox_overlay_flow.process()
runtime_tick_flow.process(delta)
devtools_context.process(has_received_state)
input_context.process(has_received_state)
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

`GameplayDevtoolsContext.process(has_received_state)` and `GameplayInputContext.process(has_received_state)` receive the gameplay-state flag.

This keeps pre-first-state behavior explicit for flows that may depend on authoritative gameplay state being available.

## Data ownership

Runtime processing owns only transient frame-processing coordination.

It uses:

* `delta` from Godot `_process`.
* `has_received_state` from `GameplayStateFlow`.
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

* `client/scripts/gameplay/state/` owns gameplay state reading and state application.
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
* [HUD and gameplay UI](../hud-and-gameplay-ui.md) - Client HUD and gameplay UI documentation.
* [Input and targeting](../input-and-targeting.md) - Client input and targeting documentation.

## Notes

`GameplayRuntimeTickFlow` currently lives under `client/scripts/shell/`, even though it participates in gameplay runtime processing.

`GameplayComposition.process` also ticks `DevToolsSessionFlow` and `GameplayPresentationFlow` outside the inner `GameplayProcessFlow` order. This document includes those calls because they are part of the current per-frame gameplay runtime path.

Server hitbox overlay processing is debug presentation. Runtime processing ticks it, but it does not make hitbox overlay behavior normal gameplay rendering authority.
