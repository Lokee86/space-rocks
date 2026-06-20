## Pregame Mode And Actions

Parent index: [Pregame Menu Flow](./!README.md)

## Purpose

This document describes the client implementation responsibility for pregame mode presentation and menu action routing.

## Overview

`PregameMenu` owns the current single-player and multiplayer pregame mode presentation.

`PregameModePresenter` applies mode-specific labels, visible button text, and disabled future actions for the current mode. `PregameMenuMode` provides the mode values used by the presenter and menu action routing.

Single-player and multiplayer share the same pregame surface, but the current mode changes the visible labels and which menu actions are routed when the user presses the available buttons.

The pregame menu also hands off transmission-panel-related actions to the transmission flow and the focused pregame subflows when those actions are requested. This document keeps those handoffs at a high level and does not restate transmission-panel mounting details.

## Code root

```text
client/
```

Primary implementation areas:

```text
client/scripts/ui/menus/
client/scripts/ui/menu_flow/
client/scenes/ui/
```

## Responsibilities

The client pregame mode and action routing flow owns:

* presenting the current pregame mode as single-player or multiplayer
* applying mode-specific button label swaps
* disabling future buttons that are not currently available
* routing single-player Play Endless actions
* routing multiplayer Create, Join, and Logout actions
* routing profile and local-pilot requests to the correct pregame subflows
* keeping mode presentation aligned with the active pregame mode

## Does not own

The client pregame mode and action routing flow does not own:

* transmission panel mounting details
* local pilot selection policy
* profile readout shaping
* lobby transport
* auth session authority
* gameplay runtime
* server state

## Mode behavior

### Single-player mode

Single-player mode shows the single-player labels, keeps the Play Endless action available, and routes the select-pilot action to the local pilot flow.

The single-player view keeps the multiplayer-specific actions hidden or disabled and leaves future buttons disabled.

### Multiplayer mode

Multiplayer mode shows the multiplayer labels, keeps the Create and Join actions available, and routes Logout through the multiplayer pregame path.

The multiplayer view hides the single-player-specific labels and leaves future buttons disabled.

## Action routing

`PregameMenu` routes button presses through the current mode:

* Play Endless routes from single-player mode.
* Create routes from multiplayer mode.
* Join routes from multiplayer mode.
* Logout routes from multiplayer mode.
* Select Pilot routes from single-player mode.
* Profile routes independently of mode as a pregame profile request.

The menu keeps routing state aligned with `PregameMenuMode` so a button only emits the action that is valid for the active mode.

## Code map

Primary implementation files:

```text
client/scripts/ui/menus/pregame_menu.gd
client/scripts/ui/menus/pregame_mode_presenter.gd
client/scripts/ui/menu_flow/pregame_menu_mode.gd
```

## Tests

No focused test is documented yet for this flow.

The closest verification boundary is the pregame menu implementation in:

```text
client/scripts/ui/menus/pregame_menu.gd
```

## Related docs

* [Pregame Menu Flow](./!README.md)
* [Transmission Panel Flow](transmission-panel-flow.md)
* [Local Pilot Flow](local-pilot-flow.md)
* [Profile Flow](profile-flow.md)
* [Client Menu Flow](../menu-flow.md)
* [Lobby Flow](../lobby-flow/!README.md)
* [App Shell And Session](../app-shell-and-session/!README.md)
* [Client](../!README.md)
* [Services](../../!README.md)

## Notes

This document captures current pregame mode and action routing behavior only. It does not describe future menu policy or procedural setup steps.
