# Gameplay Menu Flow

Parent index: [Gameplay Menu Flow](./!INDEX.md)

## Purpose

This document describes the client gameplay menu flow.

It documents how the Godot client opens, configures, hides, and routes the gameplay menu during active gameplay, local game-over presentation, authoritative match-over presentation, and spectating transitions.

## Overview

`GameplayMenuFlow` is the client-owned orchestration seam for gameplay-session menu behavior.

It is not the high-level app menu router. It is active only after gameplay has started and the gameplay runtime has been composed. It owns the rules for the gameplay menu surface that opens from `OpenMenu`, the embedded HUD menu path used during live gameplay, and the overlay menu path used after authoritative match-over.

The primary implementation files are:

```text
client/scripts/shell/gameplay_menu_flow.gd
client/scripts/ui/menus/game_menu.gd
client/scenes/ui/dialogs/game_menu.tscn
```

`GameplayMenuFlow` reads local session presentation context, configures a `GameMenu` UI instance for the current session state, forwards menu button intent through signals, and sends pause requests when opening or closing a live pause menu.

There are two menu mounting paths:

```text
Live gameplay path:
HUD
  -> CenterContainer/GameOverContainer/MarginContainer2/GameMenu

Match-over overlay path:
GameplayUserInterface or HUD fallback
  -> dynamically instantiated game_menu.tscn
```

The live path uses nodes already embedded in `hud.tscn`. The match-over overlay path instantiates `game_menu.tscn` under the configured overlay parent so match-over controls can stay available after the HUD is hidden for authoritative room match-over.

`GameMenu` is the rendered UI component. It configures button labels and primary-button availability from session mode, game-over state, room state, and spectate-target availability.

Current primary-button behavior:

```text
single-player, not game over   -> Resume, enabled
single-player, game over       -> Resume, disabled
multiplayer, not game over     -> Resume, enabled
multiplayer, game over, spectate targets available -> Spectate, enabled
multiplayer, game over, no spectate targets        -> Waiting, disabled
```

The secondary menu button emits a return-to-menu intent. Route execution stays outside the menu widget.

## Code root

```text
client/
```

## Responsibilities

The gameplay menu flow owns:

* Configuring gameplay menu dependencies after gameplay composition is created.
* Finding the embedded live gameplay menu path inside the HUD.
* Logging missing expected live gameplay menu paths.
* Opening and closing the live gameplay pause menu.
* Tracking whether the gameplay menu is visible.
* Tracking local `is_gameplay_paused` presentation state.
* Tracking local `is_game_over` presentation state.
* Blocking live pause menu opening before the first gameplay state has been received.
* Handling deferred `OpenMenu` input that was pressed before initial gameplay state arrived.
* Sending pause requests when opening, closing, or resuming from the live pause menu.
* Deactivating player afterburner when the gameplay menu opens.
* Configuring the menu for current session mode.
* Configuring the menu for game-over state.
* Enabling the match-over overlay menu path when match-end orchestration requests it.
* Instantiating the overlay `GameMenu` under the configured overlay parent.
* Hiding the overlay menu when match-over overlay mode is disabled.
* Forwarding menu intents through gameplay-menu signals.
* Forwarding return-to-pregame intent with the active session mode.
* Forwarding return-to-lobby intent.
* Forwarding quit-to-main-menu intent.
* Forwarding spectate intent.
* Showing the spectating menu path when already in spectating mode.
* Showing the `CycleView` control when spectate is requested.
* Reading spectate-target availability from `SpectateMenuState`.
* Reading room state through a configured provider.
* Resetting transient menu visibility and presentation flags.

## Does not own

The gameplay menu flow does not own:

* High-level app menu routing.
* Main menu, sign-in, pregame, lobby, or profile flow internals.
* Authoritative pause state.
* Authoritative room state.
* Authoritative match-over decisions.
* Match-end orchestration.
* Match result data.
* Match result window mounting or population.
* Match replay behavior.
* Lobby return request implementation.
* WebSocket transport.
* Packet schema ownership.
* Gameplay simulation.
* Player movement or weapon input rules.
* HUD score, lives, respawn, or loadout presentation.
* World sync or camera target ownership.
* Spectate target selection rules beyond asking whether targets exist.
* Account, profile, or persistence data.

## Domain roles

### Gameplay menu flow

`GameplayMenuFlow` is the orchestration object for gameplay-session menu presentation.

It is constructed by `GameplayComposition` and configured with:

```text
HUD reference
connection service
player reference
session context
overlay parent
room state provider
spectate menu state
lifecycle route callbacks
```

It owns menu visibility and routes menu signals outward. It does not execute app-level navigation itself.

### Game menu UI

`GameMenu` is the rendered menu control.

It owns:

* Locating the primary action button.
* Locating the menu button.
* Showing the correct primary button label.
* Enabling or disabling the primary button.
* Emitting menu-intent signals when buttons are pressed.

`GameMenu` does not know how to pause gameplay, return to lobby, return to pregame, close a network connection, or start spectating. It only emits intent.

### Live pause menu

The live pause menu is the embedded `GameMenu` under the HUD path:

```text
CenterContainer/GameOverContainer/MarginContainer2/GameMenu
```

This path is used when `uses_match_over_overlay_menu` is false.

Opening the live pause menu:

```text
OpenMenu input
-> GameplayPauseInputFlow
-> GameplayMenuFlow.handle_open_menu_pressed()
-> GameplayMenuFlow.open_live_pause_from_request()
-> GameplayMenuFlow.show_menu()
-> GameplayMenuFlow.show_live_pause_menu()
```

When the live pause menu opens, `GameplayMenuFlow` sends a pause request through the connection service if the room is not game-over.

Closing or resuming from the live pause menu also sends a pause request when a connection service exists.

### Match-over overlay menu

The match-over overlay menu is enabled by match-end orchestration through:

```text
GameplayMenuFlow.set_match_over_overlay_enabled(true)
```

When enabled, `GameplayMenuFlow` instantiates `game_menu.tscn` under the configured overlay parent and uses that menu as the active menu surface.

`GameplayComposition` configures the overlay parent as:

```text
GameplayUserInterface
```

If `GameplayUserInterface` is unavailable, it falls back to the HUD.

The overlay path exists because authoritative match-over hides the HUD. The player still needs a gameplay-session menu surface after the HUD is hidden.

### Session mode participant

`GameplayMenuFlow` reads session mode from `session_context.active_mode`.

If no active mode is available, it falls back to:

```text
Constants.SESSION_MODE_SINGLE_PLAYER
```

The session mode is passed to `GameMenu.configure_for_state()` so the UI can choose the correct button behavior.

### Room state participant

`GameplayMenuFlow` can read current room state through a configured provider.

Room state is not owned by the menu flow. It is only passed into menu configuration as state context.

### Spectate participant

`GameplayMenuFlow` receives a `SpectateMenuState` reference from `SpectateSessionFlow`.

It asks:

```text
spectate_menu_state.has_spectate_targets()
```

That result decides whether multiplayer game-over can show an enabled Spectate primary action or a disabled Waiting primary action.

When the user chooses Spectate, `GameplayMenuFlow` emits `spectate_requested`. Spectate execution is handled by the spectate context and world sync collaborators.

## Protocols and APIs

### Input action

The gameplay menu is opened by the generated/input-mapped action:

```text
OpenMenu
```

`GameplayPauseInputFlow` is the immediate processor for pause/menu input. It calls into `GameplayMenuFlow`.

Before the first gameplay state is received, `GameplayPauseInputFlow` stores a pending open-menu request instead of opening the menu immediately. After gameplay state arrives, it opens the menu once through `open_live_pause_from_request(true)`.

### Gameplay menu flow API

Primary configuration methods:

```text
configure(hud_ref, connection_service_ref, player_ref, session_context_ref)
configure_spectate_menu_state(spectate_menu_state_ref)
configure_overlay_parent(parent)
configure_lifecycle_routes(quit_route, return_to_lobby_route)
configure_room_state_provider(provider)
```

Primary state methods:

```text
reset()
set_game_over()
set_alive()
refresh_game_over_menu_state()
set_match_over_overlay_enabled(enabled)
```

Primary menu methods:

```text
handle_open_menu_pressed(has_initial_spawn)
open_live_pause_from_request(has_initial_spawn)
show_menu()
close_menu()
hide_menu()
is_menu_visible()
```

### Signals emitted by GameplayMenuFlow

`GameplayMenuFlow` emits:

```text
quit_to_main_menu_requested
return_to_pregame_requested(session_mode)
return_to_lobby_requested
spectate_requested
```

These are intent signals. The menu flow does not perform the final route itself.

### Signals consumed from GameMenu

`GameplayMenuFlow` connects to `GameMenu` signals when the rendered menu supports them:

```text
resume_requested
menu_requested
lobby_requested
spectate_requested
```

The flow also contains a guarded connection path for `quit_requested` if a compatible menu exposes it, but the current `GameMenu` scene/script uses the menu button path for return-to-menu behavior.

### Pause request API

The gameplay menu flow uses the connection service for pause requests:

```text
connection_service.send_pause_request()
```

It sends pause requests when:

* Opening the live pause menu.
* Closing the live pause menu through `OpenMenu`.
* Pressing Resume.

Pause requests are not sent while `is_game_over` is true.

### GameMenu configuration API

`GameMenu` is configured through:

```text
configure_for_state(session_mode, game_over, room_state, has_spectate_targets)
```

Current logic:

```text
if session_mode is multiplayer:
    if game_over and has_spectate_targets:
        primary action = Spectate
        enabled = true
    elif game_over:
        primary action = Waiting
        enabled = false
    else:
        primary action = Resume
        enabled = true
else:
    primary action = Resume
    enabled = not game_over
```

The current `_room_state` argument is accepted but not used by `GameMenu`.

### HTTP APIs

The gameplay menu flow exposes no HTTP APIs.

### Realtime packets

The gameplay menu flow does not parse realtime packets directly.

It consumes:

* Session mode from client session context.
* Room state through a configured provider.
* Pause request sending through the connection service.
* Spectate availability through `SpectateMenuState`.

Packet parsing, room snapshots, gameplay state, and player lifecycle state are owned by other flows.

## Data ownership

The gameplay menu flow owns only transient client presentation state.

Current local fields include:

```text
uses_match_over_overlay_menu
is_gameplay_paused
is_game_over
hud
game_over_container
game_over_margin_container
cycle_view
game_menu
overlay_parent
overlay_game_menu
connection_service
player
spectate_menu_state
session_context
room_state_provider
```

This state is resettable and not persisted.

`GameMenu` owns transient UI state:

```text
primary_action
primary_action_button
menu_button
```

No gameplay menu state is written to account data, profile data, local profile storage, match result storage, or server persistence.

## Code map

### Primary implementation

* `client/scripts/shell/gameplay_menu_flow.gd` - Gameplay menu orchestration, live pause menu path, match-over overlay path, menu visibility, pause request calls, and outward route intent signals.
* `client/scripts/ui/menus/game_menu.gd` - Rendered gameplay menu control, primary action selection, button label switching, button enabled state, and menu-intent signals.
* `client/scenes/ui/dialogs/game_menu.tscn` - Gameplay menu scene used by both embedded and overlay menu paths.

### Composition and lifecycle collaborators

* `client/scripts/gameplay/gameplay_composition.gd` - Creates `GameplayMenuFlow`, configures HUD, connection service, player, session context, overlay parent, room state provider, match-end flow, match-results flow, and spectate flow wiring.
* `client/scripts/shell/gameplay_shell_flow.gd` - Connects gameplay menu route signals into gameplay shell lifecycle signals.
* `client/scripts/session/gameplay_session_controller.gd` - Owns gameplay session reset, route execution around gameplay, graceful close behavior, lobby return request sending, and return-to-pregame/replay signal emission.
* `client/scripts/session/room_session_controller.gd` - Provides latest room state to gameplay flows through providers.
* `client/scripts/shell/app_entry.gd` - Wires scene-level UI roots and gameplay session controller dependencies.

### Input collaborators

* `client/scripts/gameplay/input/gameplay_input_context.gd` - Composes gameplay input, pause input, mouse targeting, respawn, devtools, and spectate input routes.
* `client/scripts/gameplay/input/gameplay_pause_input_flow.gd` - Processes `OpenMenu`, defers pre-spawn menu requests, and calls `GameplayMenuFlow`.
* `client/scripts/gameplay/input/gameplay_input_flow.gd` - Uses the menu flow as part of gameplay input processing.
* `client/scripts/gameplay/input/hud_input_policy.gd` - Prevents gameplay UI mouse events from also becoming gameplay input.

### Match-end and result collaborators

* `client/scripts/gameplay/match_end/match_end_flow.gd` - Enables match-over overlay menu mode, sets game-over menu state, and coordinates room match-over presentation.
* `client/scripts/ui/match_results/match_results_flow.gd` - Owns match result window mounting and result button intent forwarding.
* `client/scripts/ui/match_results/match_result_window.gd` - Owns match result window buttons and result row display.
* `client/scenes/ui/dialogs/match_result_window.tscn` - Match result window scene mounted separately from `GameMenu`.

### Spectate collaborators

* `client/scripts/gameplay/spectate/spectate_session_flow.gd` - Applies gameplay state to spectate menu state and configures that state into `GameplayMenuFlow`.
* `client/scripts/gameplay/spectate/spectate_menu_state.gd` - Owns spectate target availability derived from player lifecycle state.
* `client/scripts/gameplay/spectate/gameplay_spectate_context.gd` - Connects `GameplayMenuFlow.spectate_requested` to spectate execution.
* `client/scripts/gameplay/spectate/gameplay_spectate_flow.gd` - Begins spectating, opens the spectating menu, cycles targets, and updates world sync view target.

### Scene paths

* `client/scenes/ui/hud.tscn` - Contains the embedded live gameplay menu path under `GameOverContainer`.
* `client/scenes/game.tscn` - Contains `GameplayUserInterface`, the preferred mount parent for match-over overlay gameplay UI.

### Generated inputs

* `client/scripts/generated/constants/constants.gd` - Provides session mode constants, room-state constants, and gameplay menu primary action constants.

### Non-owning boundaries

* `client/scripts/ui/menu_flow/menu_flow_controller.gd` - Owns app-level menu route coordination, not gameplay-session menu rules.
* `client/scripts/ui/menu_flow/menu_route.gd` - Owns app-level route vocabulary, not gameplay menu state.
* `client/scripts/ui/menu_flow/pregame_menu_flow.gd` - Owns pregame route state, not active gameplay menu behavior.
* `client/scripts/lobby/` - Owns lobby UI/session behavior, not gameplay menu presentation.
* `client/scripts/networking/` - Owns transport and packet routing, not gameplay menu policy.
* `client/scripts/world/` - Owns rendered world state and camera target behavior, not menu UI.

## Tests

Gameplay menu flow tests:

* `client/tests/unit/shell/test_gameplay_menu_flow.gd`

These tests verify:

* Return-to-pregame intent includes the active session mode.
* Match-over overlay mode uses a dynamically mounted overlay `GameMenu`.
* Normal live pause behavior continues to use the embedded HUD menu when overlay mode is disabled.

Gameplay menu UI tests:

* `client/tests/unit/ui/menus/test_game_menu.gd`

These tests verify:

* The menu button emits `menu_requested`.
* Multiplayer game-over does not set the primary action to Lobby.
* Multiplayer game-over with spectate targets sets Spectate and enables the primary button.
* Multiplayer game-over without spectate targets sets Waiting and disables the primary button.
* Multiplayer non-game-over uses Resume and enables the primary button.

Related lifecycle and collaborator tests:

* `client/tests/unit/gameplay/match_end/test_match_end_flow.gd`
* `client/tests/unit/test_gameplay_session_controller.gd`
* `client/tests/unit/gameplay/test_gameplay_flow_composer.gd`
* `client/tests/unit/ui/match_results/test_match_results_flow.gd`
* `client/tests/unit/ui/match_results/test_match_result_window.gd`

These tests cover match-end coordination, gameplay session route execution, flow composition, and match-results collaborators that interact with gameplay menu state.

## Related docs

* [Client](../!INDEX.md)
* [Client menu flow](../menu-flow.md)
* [HUD and gameplay UI](../hud-and-gameplay-ui.md)
* [Input and targeting](../input-and-targeting.md)
* [Gameplay Runtime](../gameplay-runtime/!INDEX.md)
* [Match End Flow](../match-end-flow/!INDEX.md)
* [Match end orchestration](../match-end-flow/match-end-orchestration.md)
* [Match results presentation](../match-end-flow/match-results-presentation.md)
* [Pregame Menu Flow](../pregame-menu-flow/!INDEX.md)
* [World Sync](../world-sync/!INDEX.md)

## Notes

The gameplay menu flow is intentionally separate from the high-level client menu flow. `menu-flow.md` documents broad app navigation, while this document owns the active gameplay-session menu seam.

The current `GameMenu` scene has one primary action button and one menu button. Multiple primary-action labels exist under the same button, and `GameMenu` switches label visibility instead of replacing button text directly.

The `room_state` argument currently passes through `GameplayMenuFlow` into `GameMenu.configure_for_state()`, but `GameMenu` does not use it yet. Session mode, game-over state, and spectate-target availability drive the current button behavior.

The live gameplay menu uses historical HUD node names such as `GameOverContainer`. In the current implementation, that container is reused for live pause-menu presentation and should not be interpreted as authoritative match-over ownership.

Match-over overlay mode exists because authoritative room match-over hides and locks the HUD through match-end orchestration. The overlay menu keeps menu actions available without reopening normal HUD presentation.