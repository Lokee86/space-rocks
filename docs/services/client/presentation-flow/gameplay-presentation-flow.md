# Gameplay Presentation Flow

Parent index: [Presentation Flow](./!README.md)

## Purpose

This document describes the client gameplay presentation flow.

It covers the focused client presentation seam under `client/scripts/gameplay/presentation/`, including local player transient visuals, remote player off-screen indicators, player hue presentation, and the read-model handoff from world sync into non-HUD gameplay presentation.

## Overview

Gameplay presentation is client-owned visual behavior.

The current implementation is intentionally narrow. It does not own all gameplay rendering, HUD UI, event effects, match-end UI, background presentation, or world entity synchronization. Those concerns have separate service docs.

The current flow is:

```text
GameplayComposition._configure_gameplay_presentation_flow()
-> GameplayPresentationFlow.configure(...)
-> GameplayComposition.process(...)
-> GameplayPresentationFlow.process(delta, has_received_gameplay_state)
-> LocalPlayerPresentationController.process(...)
-> OSIndicatorController.update_indicators(...)
```

`GameplayComposition` creates `GameplayPresentationFlow` during gameplay composition. It passes the HUD, local player node, active camera provider, remote player visual position provider, and remote player hue provider.

`GameplayPresentationFlow` then coordinates two presentation paths:

```text
Local player presentation
-> LocalPlayerPresentationController
-> local player afterburner visual/audio state

Remote player off-screen presentation
-> OSIndicatorController
-> HUD-mounted off-screen indicator nodes
```

Remote player position and hue data comes from world sync read models:

```text
WorldSync.get_remote_player_visual_positions()
WorldSync.get_remote_player_hues()
```

The active camera comes from:

```text
ViewAnchor/Camera2D
```

This keeps gameplay presentation aligned with the same ViewAnchor basis used by world sync, while leaving ViewAnchor coordinate math and render-anchor selection in world-sync documentation.

## Code root

```text
client/scripts/gameplay/presentation/
```

## Responsibilities

Gameplay presentation owns:

* constructing the focused gameplay presentation coordinator
* wiring the HUD reference used for off-screen indicators
* wiring the local player reference used for local transient presentation
* wiring the active camera provider
* wiring remote player visual-position and hue providers
* ticking local player presentation after gameplay state has started
* turning the local player afterburner visual/audio state on and off from local input state
* hiding local afterburner presentation on reset
* calculating remote player off-screen indicator visibility
* creating and removing HUD-mounted off-screen indicator nodes
* applying remote player hue values to off-screen indicator materials
* positioning off-screen indicators along the visible viewport edge
* rotating off-screen indicators toward the remote player screen position
* hiding indicators when remote players are inside the visible area
* removing stale indicator nodes when remote players disappear from the read model
* keeping this flow presentation-only and non-authoritative

## Does not own

Gameplay presentation does not own:

* server-authoritative gameplay state
* player movement authority
* player lifecycle authority
* local player identity
* room state
* match-over decisions
* match result data
* packet parsing
* gameplay state normalization
* world entity synchronization
* player node creation or removal
* ViewAnchor server/visual coordinate conversion
* active render-anchor selection
* camera startup wiring
* background shader scrolling
* HUD layout and gameplay UI state
* gameplay menu behavior
* match-end orchestration
* match results presentation
* gameplay event/effects routing
* game-over audio one-shot gating
* projectile, asteroid, pickup, or player sync ownership
* devtools overlays or telemetry
* durable player data or profile state
* generated constants source-of-truth files

## Domain roles

### Gameplay presentation coordinator

`GameplayPresentationFlow` is the small coordinator for non-HUD gameplay presentation that is not already owned by world sync, event/effects, background, or match-end flows.

It stores:

```text
hud
camera_provider
remote_positions_provider
remote_hues_provider
os_indicator_controller
local_player_presentation_controller
```

It configures the two focused presentation controllers and ticks them during gameplay processing.

### Local player presentation controller

`LocalPlayerPresentationController` owns local player transient presentation that is driven directly from local input state.

Current behavior is limited to afterburner presentation:

```text
Input.is_action_pressed(player.move_forward_action)
-> player.set_afterburner_active(...)
```

This only runs after gameplay state has been received and when the local player node is present and visible.

`LocalPlayerPresentationController.reset()` disables local afterburner presentation by calling:

```text
player.set_afterburner_active(false)
```

The controller does not send movement input. It only presents local visual/audio feedback.

### Off-screen indicator controller

`OSIndicatorController` owns remote player off-screen indicators.

It receives:

```text
active Camera2D
remote target visual positions
remote player hue values
```

For each remote player position, it converts the world/canvas position into screen position, hides the indicator when the player is inside the visible viewport area, and otherwise clamps the indicator near the viewport edge.

Indicators are mounted under the HUD and use:

```text
res://scenes/ui/elements/osindicator.tscn
```

Each indicator's `TextureRect` material receives a hue shift from the remote player hue read model. If a remote hue is unavailable, the fallback hue from generated constants is used.

### Player hue presenter

`PlayerHuePresenter` owns player hue presentation policy.

It is used by player rendering to apply local and remote player hue values and expose remote hue read models for other presentation consumers.

Current hue behavior includes:

```text
local player hue
remote player order
remote player hue cache
fallback remote hue
remote hue similarity helper
```

Remote player hues are exposed through world sync and consumed by `OSIndicatorController`.

### World-sync read-model consumer

Gameplay presentation consumes world-sync read models but does not own them.

The current remote indicator path reads:

```text
WorldSync.get_remote_player_visual_positions()
WorldSync.get_remote_player_hues()
```

`WorldSync` owns how those values are derived from active player-render state. Gameplay presentation only consumes the result for local visual indicators.

### ViewAnchor camera consumer

Gameplay presentation receives an active camera provider from `GameplayComposition`.

The current provider resolves:

```text
ViewAnchor/Camera2D
```

The camera is used only to convert remote player visual positions into screen positions for off-screen indicators.

Camera startup, ViewAnchor ownership, visual coordinate conversion, and render-anchor selection are documented elsewhere.

## Protocols and APIs

This flow does not define a network protocol or HTTP API.

It consumes already-applied presentation read models and local Godot APIs.

### Configuration API

`GameplayPresentationFlow.configure(...)` receives:

```text
hud_ref
player_ref
camera_provider_ref
remote_positions_provider_ref
remote_hues_provider_ref
```

The coordinator passes the HUD to `OSIndicatorController` and the local player node to `LocalPlayerPresentationController`.

### Process API

`GameplayPresentationFlow.process(delta, has_received_state)` is called by gameplay composition during gameplay processing.

Current behavior:

```text
1. Tick local player presentation with has_received_state.
2. Return early if any provider callable is null.
3. Resolve the active camera.
4. Resolve remote player visual positions.
5. Resolve remote player hue values.
6. Update off-screen indicators.
```

The `delta` parameter is currently accepted but unused by the implementation.

### Reset API

`GameplayPresentationFlow.reset()` resets:

```text
OSIndicatorController
LocalPlayerPresentationController
```

Reset removes indicator nodes and disables local afterburner presentation.

### Off-screen indicator positioning

`OSIndicatorController.update_indicators(...)` uses the active camera viewport canvas transform to convert each target visual position into screen position:

```text
camera.get_viewport().get_canvas_transform() * target_position
```

It then checks the visible area using generated padding constants.

If the target is outside the visible area, the indicator is clamped to the viewport edge margin and rotated toward the target direction.

### Local afterburner input read

`LocalPlayerPresentationController` reads the local player's movement action:

```text
player.move_forward_action
```

and then checks:

```text
Input.is_action_pressed(...)
```

This is presentation feedback only. Actual movement intent packet construction remains owned by the player/input flow.

## Data ownership

Gameplay presentation owns transient client presentation state only.

Current transient state includes:

```text
indicator_nodes
local player reference
HUD reference
provider callables
local afterburner visual/audio state on the player node
remote hue cache inside PlayerHuePresenter
remote player order inside PlayerHuePresenter
```

This state is:

```text
client-local
presentation-only
non-authoritative
not persisted
reset between gameplay sessions
derived from input, scene nodes, generated constants, and world-sync read models
```

The source of presentation tuning values is:

```text
shared/constants/client/presentation.toml
```

Generated client output is:

```text
client/scripts/generated/constants/constants.gd
```

Gameplay presentation must not treat generated constants as runtime-owned mutable data.

## Code map

### Primary implementation files

```text
client/scripts/gameplay/presentation/gameplay_presentation_flow.gd
client/scripts/gameplay/presentation/local_player_presentation_controller.gd
client/scripts/gameplay/presentation/os_indicator_controller.gd
client/scripts/gameplay/presentation/player_hue_presenter.gd
```

### Composition and runtime callers

```text
client/scripts/gameplay/gameplay_composition.gd
client/scripts/session/gameplay_session_controller.gd
client/scripts/shell/gameplay_shell_flow.gd
client/scripts/gameplay/runtime/gameplay_process_flow.gd
```

### World-sync collaborators

```text
client/scripts/world/world_sync.gd
client/scripts/world/player_render/player_render_api.gd
client/scripts/world/player_render/player_meaning_api.gd
client/scripts/world/player_render/view_anchor_sync.gd
```

### Local player and scene files

```text
client/scripts/entities/player.gd
client/scenes/player.tscn
client/scenes/ui/elements/osindicator.tscn
client/scenes/game.tscn
```

### Shader and visual assets

```text
client/shaders/player_hue_shift.gdshader
client/scenes/animations/blue_afterburner.tscn
```

### Generated and source data

```text
shared/constants/client/presentation.toml
client/scripts/generated/constants/constants.gd
```

### Related implementation boundaries

```text
client/scripts/world/
client/scripts/presentation/background/
client/scripts/gameplay/events/
client/scripts/gameplay/effects/
client/scripts/gameplay/audio/
client/scripts/shell/gameplay_hud_flow.gd
client/scripts/ui/hud/
client/scripts/gameplay/match_end/
client/scripts/ui/match_results/
client/scripts/shell/gameplay_menu_flow.gd
```

### Important non-ownership boundaries

```text
client/scripts/world/
```

Owns world entity sync, ViewAnchor server/visual mapping, render-anchor selection, toroidal shortest-delta conversion, and remote player visual read models.

```text
client/scripts/presentation/background/
```

Owns background shader scroll and local ViewAnchor-following background presentation.

```text
client/scripts/gameplay/events/
client/scripts/gameplay/effects/
client/scripts/gameplay/audio/
```

Own gameplay event routing, visual effects, local death handoff, and game-over audio one-shot gating.

```text
client/scripts/shell/gameplay_hud_flow.gd
client/scripts/ui/hud/
```

Own HUD state, HUD widgets, loadout display, death/respawn HUD presentation, and gameplay UI controls.

## Tests

Relevant focused tests include:

```text
client/tests/unit/test_player_sync.gd
```

That test covers `PlayerHuePresenter` deterministic hue behavior, local hue avoidance, and filtering current self id from remote hue read models.

Adjacent tests cover the world-sync and player-render systems that provide gameplay presentation read models:

```text
client/tests/unit/test_world_sync.gd
client/tests/unit/world/player_render/test_player_render_api.gd
client/tests/unit/world/player_render/test_view_anchor_sync.gd
client/tests/unit/test_local_visual_sync.gd
client/tests/unit/test_visual_sync_positions.gd
client/tests/unit/test_world_wrap.gd
```

There are no focused unit tests for:

```text
GameplayPresentationFlow
LocalPlayerPresentationController
OSIndicatorController
```

Manual verification should include:

```text
local afterburner appears only after gameplay state starts
local afterburner stops on reset
remote off-screen indicators appear when remote players leave the visible area
remote off-screen indicators hide when remote players are visible onscreen
remote off-screen indicators use remote player hue values
indicator nodes are removed when remote players disappear
indicator placement follows the active ViewAnchor camera basis
```

## Related docs

* [Presentation Flow](./!README.md)
* [Client](../!README.md)
* [World Sync](../world-sync/!README.md)
* [World Sync Coordinator](../world-sync/world-sync-coordinator.md)
* [View Anchor And Visual Coordinates](../world-sync/view-anchor-and-visual-coordinates.md)
* [Background And Viewport Presentation](background-and-viewport-presentation.md)
* [Gameplay Runtime](../gameplay-runtime/!README.md)
* [Runtime Processing](../gameplay-runtime/runtime-processing.md)
* [HUD And Gameplay UI](../hud-and-gameplay-ui.md)
* [Gameplay Event Presentation](../gameplay-event-presentation/!README.md)
* [Gameplay Audio Flow](../gameplay-event-presentation/gameplay-audio-flow.md)
* [Match End Flow](../match-end-flow/!README.md)
* [Gameplay Menu Flow](../gameplay-menu-flow/!README.md)
* [Constants pipeline](../../../data/stubs/constants-pipeline.md)

## Notes

Legacy documentation used “presentation” broadly for many client-owned visual systems. This document intentionally uses the narrower current implementation boundary: `client/scripts/gameplay/presentation/`.

The afterburner path reads local input state directly for immediate local presentation feedback. That does not make this flow authoritative over movement. Movement packet construction and input routing remain outside this document.

Off-screen indicators are mounted under the HUD, but their behavior is owned by `OSIndicatorController`. HUD documentation owns HUD state and widgets; this document owns the indicator presentation seam that uses the HUD as a mount parent.

Remote player positions and hues are read models from world sync and player-render code. This document should not duplicate the ViewAnchor, toroidal wrap, player-render lifecycle, or coordinate-conversion rules already documented under world sync.

The current `PlayerHuePresenter.REMOTE_PLAYER_HUES` constant array exists but the active `remote_hue_for_player()` path derives ordered remote hues from the local player hue and `REMOTE_HUE_STEP`, falling back to generated constants only when no cached/order-derived hue is available.