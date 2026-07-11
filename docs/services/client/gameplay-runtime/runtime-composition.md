# Runtime Composition

Parent index: [Gameplay Runtime](./!INDEX.md)

## Purpose

This document describes the current client gameplay runtime composition.

It documents how the Godot client builds the gameplay runtime from focused flows, how composition keeps lane-native presentation delegation, world sync, HUD, input, devtools, spectate, match-end, and event presentation behind narrow seams, and where runtime composition stops owning behavior.

## Overview

The client gameplay runtime is presentation orchestration. It does not simulate authoritative gameplay.

Runtime composition starts after the client has entered a gameplay-capable session and the gameplay scene has been mounted. The composition layer wires existing scene references, runtime services, gameplay flows, and signal routes so the composed runtime can receive lane-native presentation handoffs, process per-frame presentation work, and reset cleanly when the gameplay session exits.

The main composition chain is:

```text
GameplaySessionController
-> GameplayComposition
-> GameplayShellFlow
-> GameplayRuntimeContext
-> GameplayFlowComposer
```

`GameplaySessionController` owns the session-facing lifecycle and calls into gameplay composition.

`GameplayComposition` is the top-level client runtime composition seam. It wires the gameplay shell, HUD flow, gameplay menu flow, match-end flow, match-results flow, spectate flow, devtools session flow, and gameplay presentation flow, and it provides the concrete runtime presentation targets and focused entry points used by `PresentationBridge`.

`GameplaySessionController` constructs `PresentationAdapter`, constructs `PresentationBridge`, constructs `GameplayComposition`, and configures `PresentationBridge` with `RealtimePacketPipeline`, `PresentationAdapter`, `GameplayComposition`, `GameplayComposition.world_sync`, and logger.

`GameplaySessionController` connects `RealtimePacketPipeline.gameplay_packet_applied` to `PresentationBridge.handle_gameplay_packet`. `GameplayComposition` supplies the presentation targets and focused entry points that the bridge uses; it is not the direct signal consumer.

`PresentationBridge` is the dedicated orchestration seam. `GameplaySessionController` activates, resets, and flushes it with the gameplay session lifecycle.

`GameplayShellFlow` owns the mounted gameplay shell. It creates the gameplay runtime context, configures world sync and respawn dependencies, captures and threads `world_sync` directly into `GameplayFlowComposer`, creates the flow composer, stores required lane baseline sync, and preserves a stable runtime pipeline identity for the composed gameplay frame path.

## Code root

* `client/`

## Responsibilities

* Compose client gameplay runtime objects after the gameplay scene is mounted.
* Keep gameplay runtime collaborators behind focused seams.
* Connect gameplay shell lifecycle signals to the session-facing composition layer.
* Create and configure `GameplayRuntimeContext`.
* Create and configure `GameplayFlowComposer`.
* Wire world sync, HUD runtime flow, input context, devtools context, spectate context, event lifecycle flow, targeting context, local lifecycle flow, server hitbox overlay flow, and gameplay process flow.
* Provide a stable pipeline identity for the composed gameplay runtime.
* Provide the concrete runtime presentation targets and focused entry points used by `PresentationBridge`.
* Preserve current per-frame gameplay presentation ordering.
* Reset composed runtime state during gameplay-session teardown.
* Keep runtime composition separate from entity sync, packet schema ownership, gameplay input behavior, HUD widget behavior, match-end policy, and gameplay packet relay.

## Does not own

* Server simulation authority.
* Match rules or gameplay outcomes.
* Collision, damage, scoring, lives, respawn validity, or match-over authority.
* Raw WebSocket transport.
* Packet schema source-of-truth files.
* Packet decoding before gameplay packet dispatch.
* Lane packet normalization details.
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

`GameplayComposition` is the top-level client runtime composition seam. It owns the wiring between the session controller and the gameplay shell, and it provides the concrete runtime presentation targets and focused entry points used by `PresentationBridge`.

It receives scene-level dependencies, creates the major gameplay flows, forwards lane-native presentation delegation into the shell, and exposes reset/process entry points back to the session layer.

### Gameplay shell

`GameplayShellFlow` owns the runtime shell inside the mounted gameplay scene.

It is responsible for creating the runtime context and flow composer, and for capturing and threading `world_sync` directly into the flow composer. It also stores whether required lane baselines are synced before delegating runtime input/process work.

### Runtime context

`GameplayRuntimeContext` groups shared runtime collaborators. It prevents lower-level flows from needing to rediscover scene nodes or duplicate runtime dependencies.

The context is also the world-sync processing holder. During runtime processing, it delegates interpolation to world sync.

### Flow composer

`GameplayFlowComposer` is the detailed composition seam for focused gameplay flows.

It keeps the runtime made of small owners instead of turning the shell or composition class into a large multipurpose controller.

Current composed concerns include:

```text
event lifecycle
local lifecycle reconciliation
targeting context
pointer position provider
input context
devtools context
server hitbox overlay
runtime HUD tick
spectate context
gameplay process flow
```

`GameplayFlowComposer` preloads, constructs, configures, owns, exposes, and resets `GameplayLocalLifecycleFlow`.

It configures the local lifecycle flow with the directly injected `world_sync`, `GameplayRuntimeContext.respawn_flow`, `GameplayHudFlow`, `MatchEndFlow`, and the local `Player`. It retains `GameplayRuntimeContext` for respawn and per-frame runtime responsibilities.

`GameplayRuntimeContext` creates and owns `world_sync`. `GameplayShellFlow` captures that reference immediately after runtime-context world configuration and threads it directly into `GameplayFlowComposer`. `GameplayComposition` then captures `GameplayShellFlow.world_sync` and injects it directly into `GameplayPresentationFlow`, `DevToolsSessionFlow`, and `PresentationBridge` through `GameplaySessionController`.

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

### Lane-native presentation entries

Runtime composition participates after `RealtimePacketPipeline` emits `gameplay_packet_applied`. `GameplaySessionController` drives `PresentationBridge` activation, frame flushing, and reset. The signal route is `RealtimePacketPipeline.gameplay_packet_applied` -> `PresentationBridge.handle_gameplay_packet`; `GameplayComposition` supplies the concrete presentation targets and focused entry points rather than consuming the signal directly.

Current lane-native delegation surfaces are:

```text
GameplayComposition.get_local_lifecycle_flow()
-> GameplayShellFlow.get_local_lifecycle_flow()
-> GameplayFlowComposer.get_local_lifecycle_flow()
-> GameplayLocalLifecycleFlow
```

```text
GameplayComposition.apply_devtools_gameplay_state(state)
-> GameplayShellFlow.apply_devtools_gameplay_state(state)
-> GameplayFlowComposer.apply_devtools_gameplay_state(state)
```

`PresentationBridge` orchestrates calls to the composition-owned presentation targets.

GameplayComposition and its focused flows own those targets. `GameplayComposition.get_local_lifecycle_flow()` delegates through `GameplayShellFlow` to `GameplayFlowComposer` so `PresentationBridge` can provide the local lifecycle flow to `PresentationAdapter` during each ready flush.

Composition-owned presentation targets currently include:

```text
HUD flow
gameplay menu flow
match-end flow
match-results flow
spectate flow
devtools session flow
gameplay presentation flow
local lifecycle flow
```

`GameplayLocalLifecycleFlow` owns local active, pending-respawn, and eliminated presentation reconciliation from authoritative world and session lane state. It is a presentation consumer, not an authority for gameplay outcomes.

The devtools gameplay-state path exists for devtools and server-hitbox readmodels only. It is not the primary gameplay world/session/overlay application path.

### Player pause state entry

Player pause state is forwarded through composition into the shell and pause-state flow:

```text
GameplayComposition.apply_player_pause_state_packet
-> GameplayShellFlow.apply_player_pause_state_packet
-> GameplayPauseStateFlow
-> PlayerPauseStatePacketReader
-> PlayerPauseStateTracker
```

Composition owns routing, not pause-state parsing details.

The composition routes player pause packets, but it does not own server pause authority.

`PlayerPauseStatePacketReader` identifies `player_pause_state` packets and normalizes `player_id` / `paused`.

`PlayerPauseStateTracker` stores transient per-player pause flags and resets with runtime teardown.

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

`GameplaySessionController` controls `PresentationBridge` activation, reset, and flush scheduling. It does not own gameplay packet fanout or relay.

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
* local lifecycle flow
* server hitbox overlay flow
* runtime HUD tick flow
* gameplay process flow
* world_sync

It may track:

* whether required lane baselines are currently synced for runtime-facing helpers

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

* `client/scripts/gameplay/runtime/gameplay_process_flow.gd`
* `client/scripts/shell/gameplay_runtime_tick_flow.gd`
* `client/scripts/gameplay/state/gameplay_pause_state_flow.gd`
* `client/scripts/gameplay/state/player_pause_state_packet_reader.gd`
* `client/scripts/gameplay/state/player_pause_state_tracker.gd`

### Presentation collaborators

* `client/scripts/gameplay/events/gameplay_event_lifecycle_flow.gd`
* `client/scripts/gameplay/events/gameplay_event_flow.gd`
* `client/scripts/gameplay/events/gameplay_event_controller.gd`
* `client/scripts/gameplay/lifecycle/gameplay_local_lifecycle_flow.gd`
* `client/scripts/gameplay/presentation/gameplay_presentation_flow.gd`
* `client/scripts/protocol/realtime/presentation_bridge.gd`
* `client/scripts/gameplay/spectate/spectate_session_flow.gd`
* `client/scripts/gameplay/match_end/match_end_flow.gd`
* `client/scripts/gameplay/debug/server_hitbox_overlay_flow.gd`

### External boundaries

* `client/scripts/session/gameplay_session_controller.gd`
* `client/scripts/session/session_network_controller.gd`
* `client/scripts/world/world_sync.gd`
* `client/scripts/networking/client_connection_service.gd`
* `client/scripts/protocol/realtime/`

## Tests

Runtime-composition-relevant tests include:

* `client/tests/unit/gameplay/test_gameplay_flow_composer.gd`
* `client/tests/unit/test_gameplay_session_controller.gd`
* `client/tests/unit/test_session_network_controller.gd`
* `client/tests/unit/test_player_pause_state_packet_reader.gd`
* `client/tests/unit/test_player_pause_state_tracker.gd`
* `client/tests/unit/gameplay/lifecycle/`
* `client/tests/unit/gameplay/test_gameplay_event_lifecycle_flow.gd`
* `client/tests/unit/gameplay/test_gameplay_event_controller.gd`
* `client/tests/unit/gameplay/debug/test_server_hitbox_overlay_flow.gd`
* `client/tests/unit/test_gameplay_input_context.gd`

Use the normal Godot headless GUT client test run for verification.

## Related docs

* [Gameplay Runtime](./!INDEX.md)
* [Gameplay State Application](gameplay-state-application.md)
* [Gameplay Session Lifecycle](gameplay-session-lifecycle.md)
* [Runtime Processing](runtime-processing.md)
* [Presentation Bridge](presentation-bridge.md)
* [World Sync](../world-sync/!INDEX.md)
* [HUD and gameplay UI](../hud-and-gameplay-ui.md) - Client HUD and gameplay UI documentation.
* [Input and targeting](../input-and-targeting.md) - Client input and targeting documentation.
* [Match End Flow](../match-end-flow/!INDEX.md) - Client match-end orchestration and match-results presentation documentation.
* [Gameplay Menu Flow](../gameplay-menu-flow/!INDEX.md) - Client gameplay menu and match-over overlay menu documentation.

## Notes

Composition should stay a wiring seam. When a section starts describing detailed lane packet application, per-frame processing order, target selection, HUD widget behavior, or world entity interpolation, that content belongs in the more specific client service document.

Consumers must not rediscover `world_sync` by navigating through `GameplayComposition -> GameplayShellFlow -> GameplayRuntimeContext`; dependency threading is explicit while creation and ownership remain in `GameplayRuntimeContext`.

`GameplayRuntimeContext` and `GameplayFlowComposer` are the main guardrails against runtime composition becoming a multipurpose gameplay controller.

Lane-native world/session/overlay application belongs to the gameplay-state-application and realtime protocol seams, while detailed entity synchronization belongs to world sync documentation.