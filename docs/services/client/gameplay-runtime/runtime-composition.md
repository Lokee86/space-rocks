# Runtime Composition

Parent index: [Gameplay Runtime](!README.md)

## Purpose

This document describes the current client gameplay runtime composition.

It documents how the Godot client builds the gameplay runtime from focused flows, how composition keeps packet application, world sync, HUD, input, devtools, spectate, match-end, and event presentation behind narrow seams, and where runtime composition stops owning behavior.

## Overview

The client gameplay runtime is presentation orchestration. It does not simulate authoritative gameplay.

Runtime composition starts after the client has entered a gameplay-capable session and the gameplay scene has been mounted. The composition layer wires existing scene references, runtime services, gameplay flows, and signal routes into a single runtime surface that can receive normalized state, process per-frame presentation work, and reset cleanly when the gameplay session exits.

The main composition chain is:

```text
GameplaySessionController
-> GameplayComposition
-> GameplayShellFlow
-> GameplayRuntimeContext
-> GameplayFlowComposer
```

`GameplaySessionController` owns the session-facing lifecycle and calls into gameplay composition.

`GameplayComposition` is the top-level gameplay composition object. It wires the gameplay shell, HUD flow, gameplay menu flow, match-end flow, match-results flow, spectate flow, devtools session flow, and gameplay presentation flow.

`GameplayShellFlow` owns the mounted gameplay shell. It creates the gameplay runtime context, configures world sync and respawn dependencies, creates the flow composer, tracks first-state application, and emits the gameplay-start signal when the first gameplay state is applied.

`GameplayRuntimeContext` is the runtime holder for world sync, respawn, input and presentation collaborators that need to be shared across gameplay flows.

`GameplayFlowComposer` creates and connects the focused flows used by runtime state application and per-frame processing.

The important boundary is that composition wires flows together, but does not collapse their behavior into one controller.

## Code root

* `client/`

## Responsibilities

* Compose client gameplay runtime objects after the gameplay scene is mounted.
* Keep gameplay runtime collaborators behind focused seams.
* Connect gameplay shell lifecycle signals to the session-facing composition layer.
* Create and configure `GameplayRuntimeContext`.
* Create and configure `GameplayFlowComposer`.
* Wire world sync, HUD runtime flow, input context, devtools context, spectate context, event lifecycle flow, targeting context, alive-restore flow, server hitbox overlay flow, and gameplay process flow.
* Provide a single runtime surface for applying normalized gameplay state.
* Provide a single runtime surface for applying player pause state.
* Provide a single runtime surface for debug status and debug shape catalog packets.
* Provide a single runtime surface for per-frame gameplay presentation processing.
* Track whether the first gameplay state has been received.
* Emit gameplay-start lifecycle once, after the first gameplay state is applied.
* Reset composed runtime state during gameplay-session teardown.
* Keep runtime composition separate from entity sync, packet schema ownership, gameplay input behavior, HUD widget behavior, and match-end policy.

## Does not own

* Server simulation authority.
* Match rules or gameplay outcomes.
* Collision, damage, scoring, lives, respawn validity, or match-over authority.
* Raw WebSocket transport.
* Packet schema source-of-truth files.
* Packet decoding before gameplay packet dispatch.
* Gameplay packet normalization details.
* World entity node synchronization.
* ViewAnchor or continuous visual-coordinate math.
* HUD widget internals.
* Gameplay input rules.
* Target selection rules.
* Devtools command authority.
* Match result authority.
* Persistent player data.
* Profile or local pilot storage.

## Domain roles

### Gameplay composition

`GameplayComposition` is the top-level client runtime composition seam. It owns the wiring between the session controller and the gameplay shell.

It receives scene-level dependencies, creates the major gameplay flows, forwards normalized state into the shell, and exposes reset/process entry points back to the session layer.

### Gameplay shell

`GameplayShellFlow` owns the runtime shell inside the mounted gameplay scene.

It is responsible for creating the runtime context and flow composer. It is also the place where first-state lifecycle is tracked before gameplay-start is emitted.

### Runtime context

`GameplayRuntimeContext` groups shared runtime collaborators. It prevents lower-level flows from needing to rediscover scene nodes or duplicate runtime dependencies.

The context is also the world-sync processing holder. During runtime processing, it delegates interpolation to world sync.

### Flow composer

`GameplayFlowComposer` is the detailed composition seam for focused gameplay flows.

It keeps the runtime made of small owners instead of turning the shell or composition class into a large multipurpose controller.

Current composed concerns include:

```text
event lifecycle
alive/respawn restoration
targeting context
pointer position provider
input context
devtools context
gameplay state application
server hitbox overlay
runtime HUD tick
spectate context
gameplay process flow
```

## Protocols and APIs

### Runtime construction path

The current runtime construction path is:

```text
GameplaySessionController
-> GameplayComposition
-> GameplayShellFlow
-> GameplayRuntimeContext
-> GameplayFlowComposer
```

The session controller owns the outer lifecycle. Composition owns gameplay runtime wiring. The shell owns runtime-context and composer creation.

### Gameplay state application entry

Runtime composition receives normalized gameplay state through:

```text
GameplayComposition.apply_gameplay_state
-> GameplayShellFlow.apply_gameplay_state
-> GameplayFlowComposer.apply_gameplay_state
```

Composition does not normalize the raw packet. That belongs to gameplay state application docs.

Composition does not directly synchronize world entities. World state is eventually routed through `GameplayWorldStateApplyFlow` into `WorldSync`.

### Player pause state entry

Player pause state is forwarded through composition into the shell and pause-state flow:

```text
GameplayComposition.apply_player_pause_state_packet
-> GameplayShellFlow.apply_player_pause_state_packet
```

Composition owns routing, not pause-state parsing details.

### Debug packet entry

Debug status and debug shape catalog packets route through composition to the relevant runtime collaborators.

Debug status belongs to devtools context.

Debug shape catalog data belongs to the server hitbox overlay flow.

Composition wires these routes but does not own debug behavior itself.

### Runtime processing entry

Per-frame processing flows through:

```text
GameplayComposition.process
-> GameplayShellFlow.process
-> GameplayFlowComposer.process
```

The flow composer delegates to runtime processing owners. Runtime processing details belong in `runtime-processing.md`.

### Reset entry

Gameplay-session teardown routes reset through composition and the shell so that composed runtime state is cleared consistently.

Composition reset should clear runtime presentation state without inventing server-side outcomes or durable player-data changes.

## Data ownership

Runtime composition owns transient wiring state only.

It may hold references to:

* gameplay shell
* runtime context
* flow composer
* HUD flow
* gameplay menu flow
* match-end flow
* match-results flow
* spectate flow
* devtools session flow
* gameplay presentation flow
* input context
* targeting context
* event lifecycle flow
* alive-restore flow
* server hitbox overlay flow
* runtime HUD tick flow
* gameplay process flow

It may track:

* whether a first gameplay state has been received
* whether gameplay-start has already been emitted for the mounted runtime

It does not own authoritative gameplay state.

It does not persist runtime state.

It does not own durable profile, account, or player progression data.

## Code map

### Main composition files

* `client/scripts/gameplay/gameplay_composition.gd`
* `client/scripts/shell/gameplay_shell_flow.gd`
* `client/scripts/gameplay/runtime/gameplay_runtime_context.gd`
* `client/scripts/gameplay/runtime/gameplay_flow_composer.gd`

### State and runtime collaborators

* `client/scripts/gameplay/state/gameplay_state_apply_flow.gd`
* `client/scripts/gameplay/runtime/gameplay_world_state_apply_flow.gd`
* `client/scripts/gameplay/runtime/gameplay_process_flow.gd`
* `client/scripts/shell/gameplay_runtime_tick_flow.gd`
* `client/scripts/gameplay/state/gameplay_pause_state_flow.gd`

### Presentation collaborators

* `client/scripts/gameplay/events/gameplay_event_lifecycle_flow.gd`
* `client/scripts/gameplay/events/gameplay_event_flow.gd`
* `client/scripts/gameplay/events/gameplay_event_controller.gd`
* `client/scripts/gameplay/respawn/gameplay_alive_restore_flow.gd`
* `client/scripts/gameplay/presentation/gameplay_presentation_flow.gd`
* `client/scripts/gameplay/spectate/spectate_session_flow.gd`
* `client/scripts/gameplay/match_end/match_end_flow.gd`
* `client/scripts/gameplay/debug/server_hitbox_overlay_flow.gd`

### External boundaries

* `client/scripts/session/gameplay_session_controller.gd`
* `client/scripts/session/session_network_controller.gd`
* `client/scripts/world/world_sync.gd`
* `client/scripts/networking/client_connection_service.gd`

## Tests

Runtime-composition-relevant tests include:

* `client/tests/unit/gameplay/test_gameplay_flow_composer.gd`
* `client/tests/unit/test_gameplay_session_controller.gd`
* `client/tests/unit/test_session_network_controller.gd`
* `client/tests/unit/test_gameplay_state_apply_flow.gd`
* `client/tests/unit/gameplay/test_gameplay_alive_restore_flow.gd`
* `client/tests/unit/gameplay/test_gameplay_event_lifecycle_flow.gd`
* `client/tests/unit/gameplay/test_gameplay_event_controller.gd`
* `client/tests/unit/gameplay/debug/test_server_hitbox_overlay_flow.gd`
* `client/tests/unit/test_gameplay_input_context.gd`

Use the normal Godot headless GUT client test run for verification.

## Related docs

* [Gameplay Runtime](./!README.md)
* [Gameplay State Application](gameplay-state-application.md)
* [Gameplay Session Lifecycle](gameplay-session-lifecycle.md)
* [Runtime Processing](runtime-processing.md)
* [World Sync](../world-sync/!README.md)
* [HUD and gameplay UI](../hud-and-gameplay-ui.md) - Client HUD and gameplay UI documentation.
* [Input and targeting](../input-and-targeting.md) - Client input and targeting documentation.
* [Match End Flow](../match-end-flow/!README.md) - Client match-end orchestration and match-results presentation documentation.
* [Gameplay Menu Flow](../gameplay-menu-flow/!README.md) - Client gameplay menu and match-over overlay menu documentation.

## Notes

Composition should stay a wiring seam. When a section starts describing detailed packet normalization, per-frame processing order, target selection, HUD widget behavior, or world entity interpolation, that content belongs in the more specific client service document.

`GameplayRuntimeContext` and `GameplayFlowComposer` are the main guardrails against runtime composition becoming a multipurpose gameplay controller.

`GameplayWorldStateApplyFlow` is composed by the gameplay runtime, but detailed entity synchronization belongs to world sync documentation.
