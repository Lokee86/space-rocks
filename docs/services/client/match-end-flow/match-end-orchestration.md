# Match End Orchestration

Parent index: [Match End Flow](./!INDEX.md)

## Purpose

This document describes the client-side match-end orchestration flow.

It documents how the Godot client distinguishes local elimination from authoritative room match-over, coordinates presentation changes, forwards result-window intent, and hands final route execution to session-level owners.

## Overview

`MatchEndFlow` is the client presentation orchestration seam for match-end behavior.

It reacts to two different match-end-adjacent states:

```text
local elimination
= the local player has reached zero lives

authoritative room match-over
= the room state is GameOver
```

Those states are related but not equivalent.

`handle_local_player_eliminated(lives)` performs immediate local-elimination orchestration only in multiplayer. It applies lives, shows local eliminated/game-over HUD and menu state, and requests game-over audio; it does not show match results. The best-effort self-death event path and reconstructable authoritative lifecycle path may both call it, but its session-mode guard returns without presentation consequences in single-player.

Authoritative room match-over is the room-level end condition for both modes. Single-player waits for this path. It hides and locks the HUD, enables the match-over gameplay menu overlay, requests game-over audio, reads cached match-result data, and asks `MatchResultsFlow` to present result rows. In multiplayer it remains separate from local elimination, which may occur while other players remain active.

`MatchEndFlow` does not calculate match winners, scores, deaths, persistence, or room state. Those facts come from server-owned room state and match-result payloads. The client only adapts those facts into presentation-safe UI rows and forwards user intent outward.


## Code root

```text
client/
```

## Responsibilities

The match-end orchestration flow owns:

* Distinguishing local elimination from authoritative room match-over.
* Handling local player elimination presentation from immediate self-death events and authoritative eliminated lifecycle state.
* Updating local lives on final multiplayer local elimination.
* Moving HUD and gameplay menu presentation into local multiplayer eliminated/game-over state.
* Requesting game-over audio through the gameplay event flow for accepted multiplayer local elimination.
* Guarding `handle_local_player_eliminated(lives)` by session mode so single-player has no immediate local-elimination presentation consequences.
* Suppressing duplicate local-elimination side effects with `local_player_eliminated_handled`.
* Ignoring local elimination after `room_match_over_handled` so room-wide match-over remains the higher-level outcome path.
* Checking current room state through a configured provider.
* Handling authoritative room match-over exactly once per match-end lifecycle.
* Hiding and locking HUD presentation for authoritative room match-over.
* Enabling match-over gameplay menu overlay behavior.
* Reading cached match-result data through a configured provider.
* Converting match-result player entries into presentation rows.
* Calling `MatchResultsFlow.show_results()` with the active session mode and result rows.
* Forwarding result-window replay, lobby, pregame, and main-menu intent signals outward.
* Clearing match-over presentation locks on reset without re-showing the HUD.
* Clearing the local elimination guard on `reset()` and `handle_alive_restored()` so a later lifecycle can be handled.

## Does not own

The match-end orchestration flow does not own:

* Authoritative match-over decisions.
* Authoritative room lifecycle.
* Winner calculation.
* Score calculation.
* Death count calculation.
* Match-result persistence.
* Match-result packet schema authority.
* WebSocket transport.
* Room snapshot construction.
* HUD widget implementation.
* Raw HUD visibility mechanics.
* Gameplay menu button layout.
* Match result window mounting details.
* Result row rendering.
* Audio playback implementation.
* Audio one-shot gating rules.
* App-level navigation execution.
* Lobby return execution.
* Single-player replay boot execution.
* Main-menu route execution.

## Domain roles

### Match-end orchestrator

`MatchEndFlow` is the client-side match-end presentation orchestrator.

It receives collaborators from gameplay composition:

```text
GameplayHudFlow
GameplayMenuFlow
session context
GameplayEventFlow
MatchResultsFlow
room state provider
match result provider
```

It coordinates those collaborators but does not take over their ownership.

### Local elimination handler

Local elimination enters `MatchEndFlow` through:

```text
GameplayEventLifecycleFlow
-> GameplayDeathFlow
-> MatchEndFlow.handle_local_player_eliminated(lives)

GameplayLocalLifecycleFlow
-> MatchEndFlow.handle_local_player_eliminated(lives)
```

`GameplayDeathFlow` supplies the immediate event response and passes integer `lives` when a `ship_death` event reaches zero lives. `GameplayLocalLifecycleFlow` supplies reconstructable eliminated state from authoritative lifecycle/session data and also passes authoritative integer `lives`.

The local elimination path:

```text
self-death event with lives == 0
-> MatchEndFlow.handle_local_player_eliminated(lives)

authoritative eliminated lifecycle with integer lives
-> MatchEndFlow.handle_local_player_eliminated(lives)

both paths
-> MatchEndFlow checks the active session mode

multiplayer
-> apply authoritative lives to HUD
-> set HUD game-over presentation
-> set gameplay menu game-over presentation
-> request delayed game-over sound

single-player
-> return without presentation consequences
```

In multiplayer this path does not show match results because the room may not have reached authoritative match-over yet. In single-player, the handler returns without immediate HUD, menu, or audio consequences; the client waits for the authoritative room `GameOver` path.

`local_player_eliminated_handled` suppresses duplicate multiplayer local-elimination side effects when the event and authoritative lifecycle paths describe the same outcome. The session-mode guard is evaluated before those consequences are applied. Local elimination is also ignored after `room_match_over_handled`; room-wide match-over remains the separate higher-level outcome path. `reset()` and `handle_alive_restored()` clear the local elimination guard so a later lifecycle can be handled.

### Room match-over handler

Authoritative room match-over enters `MatchEndFlow` through the configured room-state provider.

`refresh_match_end_state()` reads the current room state. When the state is `GameOver`, it calls `handle_room_match_over()`.

The room match-over path:

```text
room state provider returns GameOver
-> guard against repeated handling
-> hide and lock HUD
-> enable match-over gameplay menu overlay
-> set gameplay menu game-over presentation
-> request game-over sound
-> read cached match result
-> pass rows to MatchResultsFlow
```

`room_match_over_handled` prevents repeated `GameOver` snapshots from repeatedly remounting result UI.

### HUD collaborator

`GameplayHudFlow` owns HUD mechanics.

`MatchEndFlow` calls HUD methods such as:

```text
apply_lives()
set_game_over()
hide_for_match_over()
clear_match_over_visibility_lock()
set_alive()
```

For authoritative room match-over, `hide_for_match_over()` is the preferred path because it hides the HUD and prevents later gameplay-state application from re-showing it while the match-over lock is active.

`MatchEndFlow.reset()` clears the match-over visibility lock, but does not show the HUD. Normal gameplay state must start again before HUD presentation returns.

### Gameplay menu collaborator

`GameplayMenuFlow` owns gameplay menu behavior.

`MatchEndFlow` calls menu methods such as:

```text
set_game_over()
set_alive()
set_match_over_overlay_enabled(true)
set_match_over_overlay_enabled(false)
```

The menu flow decides whether the embedded HUD menu or overlay menu is active. `MatchEndFlow` only requests match-over mode.

### Event and audio collaborator

`GameplayEventFlow` owns gameplay event presentation and audio request behavior.

`MatchEndFlow` may call:

```text
play_game_over_sound_after_delay()
```

The match-end flow does not play audio directly and does not own repeated-audio gating.

### Match results collaborator

`MatchResultsFlow` owns result-window presentation.

`MatchEndFlow` calls:

```text
show_results(session_mode, rows)
```

It also connects to result-flow intent signals and re-emits them outward:

```text
replay_requested
return_to_lobby_requested
return_to_pregame_requested
quit_to_main_menu_requested
```

The session-level owner decides what those intents do.

### Session-level route owner

`GameplayComposition` connects match-end intent signals and forwards them to `GameplaySessionController`.

`GameplaySessionController` executes session-level consequences such as:

```text
close connection
send return-to-lobby request
reset gameplay state
clear session context
clear boot flow
show main menu
emit replay or pregame route intent
```

`MatchEndFlow` does not execute those routes directly.

## Protocols and APIs

### Room state input

Room match-over is detected from room state, not local HUD state.

`RoomSessionController` caches latest room state from room snapshots and room-state-change packets. `GameplaySessionController` provides a room-state provider to `GameplayComposition`, which configures `MatchEndFlow`.

The match-end flow checks for:

```text
Constants.ROOM_STATE_GAME_OVER
```

When that state is observed, authoritative room match-over presentation is triggered.

### Local death event input

Local elimination is driven by server event data normalized through gameplay event handling.

The relevant local death event field is:

```text
Packets.FIELD_LIVES
```

If lives are greater than zero, `GameplayDeathFlow` handles local death/respawn presentation without match-end orchestration.

If lives are zero, `GameplayDeathFlow` calls:

```text
MatchEndFlow.handle_local_player_eliminated(lives)
```

The protocol-neutral handler accepts only the integer lives value. It may also be reached by `GameplayLocalLifecycleFlow` when authoritative `player_lifecycle` reports `eliminated`.

`handle_local_player_eliminated(lives)` checks the active session mode. Only multiplayer accepts immediate local-elimination HUD, menu, and delayed-audio consequences. Single-player returns from this handler and waits for room state `GameOver` to enter the authoritative room match-over path.

### Match result input

Match results are read from a configured provider.

The provider currently resolves through `RoomSessionController.current_match_result()`, which caches `room_snapshot.match_result` only when the payload includes a non-empty match id.

`MatchEndFlow` reads:

```text
match_result.players
```

Each player dictionary is converted into a presentation row:

```text
game_player_id
score
ship_deaths
won
```

If `game_player_id` is absent, the flow falls back to `player_id`, then `"Player"`.

If there is no provider, no match result, or no players array, the result rows are empty.

### Result-window intent output

`MatchEndFlow` does not directly receive button presses from the result window.

The signal path is:

```text
MatchResultWindow
-> MatchResultsFlow
-> MatchEndFlow
-> GameplayComposition
-> GameplaySessionController
```

The intent signals are:

```text
replay_requested
return_to_lobby_requested
return_to_pregame_requested
quit_to_main_menu_requested
```

### HTTP APIs

Match-end orchestration does not expose HTTP APIs.

Match result persistence and account/profile-backed result storage belong outside this client flow.

## Data ownership

`MatchEndFlow` owns only local, resettable presentation orchestration state.

Current local state:

```text
room_match_over_handled
local_player_eliminated_handled
```

`room_match_over_handled` prevents repeated room match-over handling during the same lifecycle. `local_player_eliminated_handled` prevents duplicate multiplayer local HUD, menu, and delayed-audio side effects across event and authoritative lifecycle entry paths; it does not make single-player local elimination consequential. Both guards are reset for a later lifecycle; `handle_alive_restored()` specifically clears the local elimination guard.

`MatchEndFlow` also holds references to collaborators and providers:

```text
hud_flow
menu_flow
event_flow
match_results_flow
session_context
match_result_provider
room_state_provider
```

Those references do not make match-end orchestration the owner of the data behind them.

The flow does not persist data. It does not own match-result storage. It does not own room-state storage. It does not mutate authoritative gameplay facts.

## Code map

### Primary implementation

* `client/scripts/gameplay/match_end/match_end_flow.gd` - Client match-end presentation orchestration, local elimination handling, room match-over handling, result-row extraction, and result intent forwarding.

### Composition and lifecycle wiring

* `client/scripts/gameplay/gameplay_composition.gd` - Constructs `MatchEndFlow`, wires HUD/menu/results collaborators, configures providers, connects match-end route intent signals, and clears match-end state on reset.
* `client/scripts/session/gameplay_session_controller.gd` - Provides room-state and match-result providers to gameplay composition, refreshes match-end state, and executes route consequences after match-end intent bubbles outward.
* `client/scripts/session/room_session_controller.gd` - Caches latest room state and match-result payload from room snapshots and room-state-change packets.

### Event and respawn collaborators

* `client/scripts/gameplay/events/gameplay_event_lifecycle_flow.gd` - Wires gameplay event presentation and local death handling into `MatchEndFlow`.
* `client/scripts/gameplay/events/gameplay_death_flow.gd` - Handles local self-death events and delegates final local elimination to `MatchEndFlow`.
* `client/scripts/gameplay/lifecycle/gameplay_local_lifecycle_flow.gd` - Reconstructs eliminated and active local presentation from authoritative world/session state and collaborates with `MatchEndFlow`.

### HUD, menu, and result collaborators

* `client/scripts/shell/gameplay_hud_flow.gd` - Owns HUD visibility, local death, respawn, game-over presentation, and match-over visibility lock mechanics.
* `client/scripts/shell/gameplay_menu_flow.gd` - Owns gameplay menu state, live pause menu behavior, and match-over overlay menu behavior.
* `client/scripts/ui/match_results/match_results_flow.gd` - Owns match result window mounting, clearing, button intent forwarding, and single-player replay versus multiplayer lobby interpretation.
* `client/scripts/ui/match_results/match_result_window.gd` - Owns result-window controls and emits button intent.
* `client/scripts/ui/match_results/player_score_row.gd` - Renders one result row.

### Scenes

* `client/scenes/game.tscn` - Provides `GameplayUserInterface`, the gameplay-session UI root used as the match-results and overlay parent.
* `client/scenes/ui/hud.tscn` - Contains HUD controls and embedded gameplay menu paths used by HUD/menu collaborators.
* `client/scenes/ui/dialogs/match_result_window.tscn` - Result window mounted by `MatchResultsFlow`.
* `client/scenes/ui/elements/player_score_row.tscn` - Result row scene used by the match result window.
* `client/scenes/ui/dialogs/game_menu.tscn` - Gameplay menu scene used by live gameplay and match-over overlay paths.

### Generated inputs

* `client/scripts/generated/constants/constants.gd` - Generated constants for room states and session modes.
* `client/scripts/generated/networking/packets/packets.gd` - Generated packet field constants used for death events, room state, and match result fields.

### Non-owning boundaries

* `client/scripts/networking/` - Owns transport and packet routing, not match-end presentation policy.
* `client/scripts/ui/menu_flow/` - Owns app and pregame menu routing, not gameplay match-end orchestration.
* `client/scripts/lobby/` - Owns lobby UI/state presentation, not match-end result orchestration.
* `client/scripts/world/` - Owns entity rendering and world sync, not match-end UI state.
* `services/game-server/` - Owns authoritative simulation, room match-over decisions, and match result creation.

## Tests

### Match-end orchestration tests

* `client/tests/unit/gameplay/match_end/test_match_end_flow.gd`

These tests verify:

* Multiplayer local elimination updates local HUD/menu game-over state and requests game-over audio.
* Single-player local-elimination calls return without presentation consequences.
* Single-player waits for authoritative room `GameOver` before HUD/menu/audio match-over orchestration and results presentation.
* Multiplayer local elimination does not show final match results while other players remain active.
* Room match-over hides the HUD and passes rows to results presentation.
* Empty match-result providers still show the result window with empty rows.
* Repeated room match-over refreshes do not repeatedly show result windows.

### Local death and lifecycle tests

* `client/tests/unit/gameplay/events/test_gameplay_death_flow.gd`
* `client/tests/unit/gameplay/lifecycle/`

These tests verify immediate local death delegation, authoritative lifecycle reconstruction, duplicate elimination suppression, and valid active restoration.

### Composition and session tests

* `client/tests/unit/gameplay/test_gameplay_flow_composer.gd`
* `client/tests/unit/test_gameplay_session_controller.gd`

These tests cover gameplay flow wiring and session-level routing behavior around gameplay reset and exit paths.

### Collaborator tests

* `client/tests/unit/shell/test_gameplay_menu_flow.gd`
* `client/tests/unit/ui/menus/test_game_menu.gd`
* `client/tests/unit/ui/match_results/test_match_results_flow.gd`
* `client/tests/unit/ui/match_results/test_match_result_window.gd`

These tests verify the menu and result collaborators that `MatchEndFlow` coordinates but does not own.

## Related docs

* [Client](../!INDEX.md)
* [Match end flow](./!INDEX.md)
* [Match results presentation](match-results-presentation.md)
* [Gameplay menu flow](../gameplay-menu-flow/!INDEX.md)
* [HUD and gameplay UI](../hud-and-gameplay-ui.md)
* [Gameplay runtime](../gameplay-runtime/!INDEX.md)
* [Gameplay session lifecycle](../gameplay-runtime/gameplay-session-lifecycle.md)
* [Gameplay state application](../gameplay-runtime/gameplay-state-application.md)
* [Gameplay packets](../../../protocol/gameplay-packets.md) - gameplay realtime packet documentation.
* [Realtime websocket protocol](../../../protocol/realtime-websocket-protocol.md) - realtime websocket protocol documentation.
* [Match end and results flow](../../../domains/player-experience/match-end-and-results-flow.md) - cross-system player-experience match-end flow.

## Notes

This file stays focused on `MatchEndFlow` and its immediate orchestration boundary, while match-results presentation, HUD behavior, gameplay menu behavior, and route execution remain in their own client-service docs.

Multiplayer local elimination and authoritative room match-over must remain separate. Showing match results from local elimination would be incorrect because the local player can be out of lives while other players remain active. Single-player deliberately waits for authoritative room `GameOver`; its local-elimination handler is a no-op for presentation.

`MatchEndFlow.reset()` clears match-end state, the local elimination guard, and the HUD visibility lock, but it must not directly show the HUD again. `handle_alive_restored()` also clears the local elimination guard. HUD presentation returns through normal gameplay state flow when a new active gameplay lifecycle begins.

The current client result rows include `won`, but the current result row renderer displays `PLAYER`, `DEATHS`, and `SCORE`. Result-window rendering details belong in `match-results-presentation.md`.
