# Client Menu Flow

Parent index: [Client](./!README.md)

## Purpose

This document describes the high-level client menu flow. It covers how the app enters the menu stack, how top-level routes move between main menu, sign-in, pregame, lobby, gameplay, match results, and shutdown, and which implementation surfaces own those transitions.

This document is canonical Service documentation for the client menu flow. Pregame-specific profile, local pilot, and transmission behavior is documented separately in [Pregame Menu Flow](pregame-menu-flow/!README.md).

## Overview

The client menu flow is the navigation layer around playable sessions. It does not own gameplay simulation, world synchronization, profile persistence, or HTTP contracts. It coordinates user-facing route changes and delegates detailed behavior to focused flows.

The high-level menu flow connects:

* app startup into the main menu
* main menu choices into single-player, multiplayer, sign-in, or quit
* multiplayer entry into lobby creation, joining, readiness, and game start
* gameplay pause/menu actions back into gameplay, lobby, pregame, or main menu
* match end actions into replay, lobby, pregame, or main menu

## Code root

* `client/scenes/ui`
* `client/scripts/ui`
* `client/scripts/shell`
* `client/scripts/session`
* `client/scripts/lobby`
* `client/scripts/main_menu`

## Responsibilities

* Coordinate high-level client route changes across the menu stack.
* Route app startup into the main menu.
* Route main menu choices into single-player, multiplayer, sign-in, or quit.
* Route multiplayer entry into lobby creation, joining, readiness, and game start.
* Route gameplay pause/menu actions back into gameplay, lobby, pregame, or main menu.
* Route match-end actions into replay, lobby, pregame, or main menu.
* Keep high-level navigation separate from pregame, gameplay, lobby, and auth details.
* Delegate detailed behavior to focused flows instead of duplicating route-local rules.

## Does not own

* Gameplay simulation.
* World synchronization.
* Profile persistence.
* HTTP contract details.
* Pregame profile, local pilot, or transmission internals.
* Gameplay menu legality rules beyond route-level navigation.
* Match result scoring logic.
* Room packet contracts.
* Auth session implementation details.

## Domain roles

### Route coordinator

`menu_flow_controller.gd` owns the high-level route coordination between the client menu surfaces.

### Route vocabulary

`menu_route.gd` defines the route vocabulary used by the menu flow layer.

### Pregame flow

`pregame_menu_flow.gd` owns pregame menu state and mode-specific presentation. The high-level menu flow only treats pregame as a distinct route.

### Multiplayer entry flow

`multiplayer_entry_flow.gd` owns multiplayer entry from the menu stack toward room and lobby flow.

### Sign-in flow

`sign_in_flow.gd` owns the sign-in window flow.

### Gameplay menu flow

`gameplay_menu_flow.gd` owns Escape and gameplay menu behavior after a gameplay session has started.

### Match results flow

`match_results_flow.gd` owns match result window population and result-window actions.

### Shared menu UI primitives

`button_long.tscn`, `button_square.tscn`, `window_7.tscn`, and `window_8.tscn` are reusable menu chrome scenes used by multiple menu surfaces.

They provide shared UI primitives for menu presentation, but they do not own route behavior, menu state, auth, lobby behavior, match results, or gameplay flow.

## Protocols and APIs

The menu flow layer uses local scene and script routing rather than a separate router service.

It coordinates:

* menu route values from `menu_route.gd`
* gameplay-session lifecycle signals from shell and session controllers
* menu UI interactions from main menu, pregame menu, game menu, and match-results surfaces
* multiplayer entry and lobby transitions from the UI flow layer

This document does not define transport or packet schemas. It only describes how the client menu layer routes between existing implementation seams.

## Data ownership

The client owns transient menu-navigation state only.

That includes:

* current top-level route selection
* whether the client is in main menu, sign-in, pregame, lobby, gameplay menu, or match-results navigation
* route-local presentation state owned by the focused flow

It does not own durable account data, profile persistence, room membership authority, or gameplay state.

## Code map

### Menu flow controllers and routes

* `client/scripts/ui/menu_flow/menu_flow_controller.gd`
* `client/scripts/ui/menu_flow/menu_route.gd`
* `client/scripts/ui/menu_flow/pregame_menu_flow.gd`
* `client/scripts/ui/menu_flow/multiplayer_entry_flow.gd`
* `client/scripts/ui/menu_flow/transmission_flow.gd`
* `client/scripts/ui/menu_flow/local_pilot_flow.gd`

### Menu UI surfaces

* `client/scripts/ui/menus/main_menu.gd`
* `client/scripts/ui/menus/pregame_menu.gd`
* `client/scripts/ui/menus/game_menu.gd`

### Shell and session transition scripts

* `client/scripts/shell/app_entry.gd`
* `client/scripts/shell/gameplay_shell_flow.gd`
* `client/scripts/shell/gameplay_menu_flow.gd`
* `client/scripts/shell/gameplay_hud_flow.gd`
* `client/scripts/shell/gameplay_runtime_tick_flow.gd`
* `client/scripts/session/client_session_context.gd`
* `client/scripts/session/session_network_controller.gd`
* `client/scripts/session/room_session_controller.gd`
* `client/scripts/session/gameplay_session_controller.gd`
* `client/scripts/main_menu/main_menu_session_controller.gd`

### Lobby and multiplayer transition scripts

* `client/scripts/lobby/lobby_flow.gd`
* `client/scripts/lobby/lobby_shell_flow.gd`
* `client/scripts/lobby/lobby_return_flow.gd`
* `client/scripts/lobby/lobby_session_state.gd`
* `client/scripts/ui/lobby/multiplayer_dialog.gd`
* `client/scripts/ui/lobby/join_dialog.gd`
* `client/scripts/ui/lobby/join_dialog_flow.gd`
* `client/scripts/ui/lobby/multiplayer_lobby.gd`

### Auth and account menu surfaces

* `client/scripts/auth/auth_session.gd`
* `client/scripts/auth/auth_session_controller.gd`
* `client/scripts/ui/sign_in/login_window.gd`
* `client/scripts/ui/sign_in/sign_in_flow.gd`

### Match result and gameplay menu surfaces

* `client/scripts/ui/match_results/match_result_window.gd`
* `client/scripts/ui/match_results/match_results_flow.gd`
* `client/scripts/gameplay/match_end/match_end_flow.gd`

### Primary menu scenes

* `client/scenes/ui/elements/button_long.tscn`
* `client/scenes/ui/elements/button_square.tscn`
* `client/scenes/ui/elements/windows/window_7.tscn`
* `client/scenes/ui/elements/windows/window_8.tscn`
* `client/scenes/ui/main_menu.tscn`
* `client/scenes/ui/pregame_menu.tscn`
* `client/scenes/ui/dialogs/login_window.tscn`
* `client/scenes/ui/dialogs/multiplayer_dialog.tscn`
* `client/scenes/ui/dialogs/join_dialog.tscn`
* `client/scenes/ui/dialogs/multiplayer_lobby.tscn`
* `client/scenes/ui/dialogs/game_menu.tscn`
* `client/scenes/ui/dialogs/match_result_window.tscn`
* `client/scenes/game.tscn`

## Tests

Relevant tests are under:

* `client/tests/unit/ui/menu_flow/`
* `client/tests/unit/ui/menus/`
* `client/tests/unit/ui/sign_in/`
* `client/tests/unit/ui/lobby/`
* `client/tests/unit/ui/match_results/`
* `client/tests/unit/lobby/`
* `client/tests/unit/shell/`
* `client/tests/unit/boot/`

Use these tests to verify route behavior, menu presentation behavior, shell return behavior, sign-in flow behavior, lobby entry behavior, and gameplay menu behavior.

## Related docs

* [Auth Session Flow](auth-session-flow.md)
* [HUD and Gameplay UI](hud-and-gameplay-ui.md)
* [Input and Targeting](input-and-targeting.md)
* [Pregame Menu Flow](pregame-menu-flow/!README.md)
* [Pregame Local Pilot Flow](pregame-menu-flow/local-pilot-flow.md)
* [Pregame Profile Flow](pregame-menu-flow/profile-flow.md)
* [Gameplay Runtime](gameplay-runtime/!README.md)
* [Gameplay Session Lifecycle](gameplay-runtime/gameplay-session-lifecycle.md)
* [World Sync](world-sync/!README.md)
* [Gameplay Menu Flow](gameplay-menu-flow/!README.md)
* [Match End Flow](match-end-flow/!README.md)

## Notes

This document stays at the high-level client menu boundary and does not duplicate pregame-specific implementation details.
