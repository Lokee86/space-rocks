# Match Results Presentation

Parent index: [Match End Flow](./!INDEX.md)

## Purpose

This document describes the current client match-results presentation implementation.

It documents how the Godot client mounts the match result window, renders per-player result rows, configures the Replay/Lobby button by session mode, and forwards result-window button intent back to match-end/session routing.

## Overview

Match results presentation is a client UI seam under the match-end flow.

The authoritative match result is produced outside this UI layer. The client receives or caches the match result through room/session flow, `MatchEndFlow` converts the cached result into display rows, and `MatchResultsFlow` mounts a result window under the gameplay UI root.

The presentation flow is:

```text
authoritative room match-over
-> RoomSessionController cached match result
-> MatchEndFlow result row extraction
-> MatchResultsFlow.show_results(session_mode, rows)
-> MatchResultWindow.apply_rows(rows)
-> PlayerScoreRow.apply_row(row)
```

The result window currently displays:

```text
PLAYER | DEATHS | SCORE
```

It does not currently display kills, account identity, profile identity, currency rewards, unlocks, achievements, or progression summaries.

The result-window buttons emit intent only. They do not execute app-level routes directly. The intent path is:

```text
MatchResultWindow button
-> MatchResultsFlow signal
-> MatchEndFlow signal
-> GameplayComposition signal
-> GameplaySessionController route/session action
```

## Code root

```text
client/scripts/ui/match_results/
```

## Responsibilities

The match-results presentation implementation owns:

- Instantiating `match_result_window.tscn`.
- Mounting the result window under the configured gameplay UI mount parent.
- Moving the mounted result window to the front.
- Clearing any existing result window before mounting a new one.
- Freeing the mounted result window on clear/reset.
- Connecting result-window button signals to presentation-flow handlers.
- Configuring the primary result button label for single-player or multiplayer mode.
- Applying result rows to the visible score container.
- Creating one `PlayerScoreRow` scene per supplied row.
- Rendering each row’s player id, ship deaths, and score.
- Forwarding Replay, Lobby, Menu, and Quit intent outward through signals.

## Does not own

The match-results presentation implementation does not own:

- Authoritative match-over decisions.
- Authoritative match result generation.
- Score calculation.
- Death counting.
- Win/loss calculation.
- Room state.
- Room snapshots.
- WebSocket transport.
- Packet schemas.
- Match result caching.
- Match-end orchestration.
- HUD hiding or match-over HUD locking.
- Gameplay menu overlay behavior.
- App-level routing.
- Replay session creation.
- Return-to-lobby request execution.
- Connection shutdown.
- Persistent match result storage.
- Player profile, account, currency, unlock, or achievement storage.

## Domain roles

### MatchResultsFlow

`MatchResultsFlow` is the result-window presentation orchestrator.

It is configured with a `mount_parent`, usually the gameplay UI root selected by `GameplayComposition`. When `show_results(session_mode, rows)` is called, it clears the prior window, instantiates `match_result_window.tscn`, mounts it, connects window signals, configures it for the session mode, applies result rows, and returns the mounted window.

`MatchResultsFlow` emits outward route-intent signals:

```text
replay_requested
return_to_lobby_requested
return_to_pregame_requested
quit_to_main_menu_requested
```

The `LobbyReplayButton` has mode-specific intent:

```text
single_player -> replay_requested
multiplayer   -> return_to_lobby_requested
```

### MatchResultWindow

`MatchResultWindow` is the visible result-window UI controller.

It owns the scene-local button signal wiring, button label mode, row clearing, and row application. It emits scene-local signals:

```text
lobby_replay_requested
menu_requested
quit_requested
```

It does not know whether the mounted result window belongs to single-player replay, multiplayer lobby return, pregame routing, or main-menu routing. That decision stays outside the scene.

### PlayerScoreRow

`PlayerScoreRow` renders one row in the result table.

It reads a dictionary and displays:

```text
player_id or game_player_id
ship_deaths
score
```

If both `player_id` and `game_player_id` are absent, it displays `Player`.

The row accepts a `won` field in the row dictionary only because upstream result rows currently preserve it. The current row scene does not display win/loss state.

### MatchEndFlow collaboration

`MatchEndFlow` owns the handoff into match-results presentation.

On authoritative room match-over, `MatchEndFlow` calls:

```text
match_results_flow.show_results(_current_session_mode(), _current_match_result_rows())
```

`MatchEndFlow` also connects the outward `MatchResultsFlow` signals and re-emits them through the match-end seam.

### GameplayComposition collaboration

`GameplayComposition` constructs `MatchResultsFlow`, configures its mount parent, gives it to `MatchEndFlow`, clears it during gameplay reset, and forwards match-end route-intent signals toward `GameplaySessionController`.

The preferred mount parent is:

```text
GameplayUserInterface
```

If that node is unavailable, composition falls back to:

```text
HUD
```

## Protocols and APIs

### Presentation API

`MatchResultsFlow` exposes:

```text
configure(mount_parent_ref: Node) -> void
show_results(session_mode: String, rows: Array = []) -> Control
clear() -> void
```

`show_results()` is safe to call repeatedly. It clears any previously mounted window before mounting the new one.

If no mount parent is configured, `show_results()` returns `null` and does not mount a result window.

### Session mode input

The presentation flow receives session mode as a string.

Current mode behavior:

```text
"multiplayer" -> show Lobby label and emit return_to_lobby_requested from the primary button
other values  -> show Replay label and emit replay_requested from the primary button
```

The session mode is presentation input only. The result window does not own the active session context.

### Row input

The presentation flow receives result rows as dictionaries.

Current row fields consumed by display are:

```text
player_id
game_player_id
ship_deaths
score
```

Current upstream rows from `MatchEndFlow` use:

```text
game_player_id
score
ship_deaths
won
```

Display fallback behavior:

```text
player_id fallback -> game_player_id fallback -> "Player"
ship_deaths fallback -> 0
score fallback -> 0
```

### Result-window signals

`MatchResultWindow` emits:

```text
lobby_replay_requested
menu_requested
quit_requested
```

`MatchResultsFlow` translates those into:

```text
lobby_replay_requested + single_player -> replay_requested
lobby_replay_requested + multiplayer   -> return_to_lobby_requested
menu_requested                         -> return_to_pregame_requested
quit_requested                         -> quit_to_main_menu_requested
```

### HTTP APIs

Match-results presentation does not expose or call HTTP APIs.

### Realtime packets

Match-results presentation does not read realtime packets directly.

Realtime room snapshots and room state changes are handled before this presentation seam. The result presentation layer receives already-extracted row dictionaries.

## Data ownership

Match-results presentation owns only transient UI state.

`MatchResultsFlow` owns:

```text
mount_parent
window
current_session_mode
```

`MatchResultWindow` owns scene-local row children under `%ScoreContainer`.

`PlayerScoreRow` owns scene-local label text after `apply_row()` is called.

None of this state is persisted. None of it is authoritative.

The presentation layer does not store match history, account id, local profile id, match id, result source packet, room state, or player progression state.

## Code map

### Primary implementation

- `client/scripts/ui/match_results/match_results_flow.gd` - Result-window mounting, clearing, mode storage, button-intent translation, and outward signals.
- `client/scripts/ui/match_results/match_result_window.gd` - Result-window scene controller, button signal wiring, mode label toggling, row clearing, and row creation.
- `client/scripts/ui/match_results/player_score_row.gd` - Per-player result row label population.

### Scenes

- `client/scenes/ui/dialogs/match_result_window.tscn` - Match result window scene, result table header, score container, and Lobby/Replay, Menu, and Quit buttons.
- `client/scenes/ui/elements/player_score_row.tscn` - Per-player row scene with player id, deaths, and score labels.

### Collaborators

- `client/scripts/gameplay/match_end/match_end_flow.gd` - Owns authoritative room match-over presentation orchestration and passes rows into `MatchResultsFlow`.
- `client/scripts/gameplay/gameplay_composition.gd` - Constructs and configures `MatchResultsFlow`, wires it into `MatchEndFlow`, and clears it on reset.
- `client/scripts/session/room_session_controller.gd` - Caches room state and latest non-empty match result from room snapshots.
- `client/scripts/session/gameplay_session_controller.gd` - Executes route/session actions after result intent bubbles outward.

### Generated/source inputs

- `client/scripts/generated/constants/constants.gd` - Provides session mode constants used by upstream match-end/session logic.
- `client/scripts/generated/networking/packets/packets.gd` - Provides room snapshot and match-result field names used before rows reach presentation.

### Non-owning boundaries

- `client/scripts/gameplay/match_end/` owns match-end orchestration, not result-window rendering.
- `client/scripts/shell/gameplay_menu_flow.gd` owns gameplay menu and match-over overlay menu behavior, not result-window rendering.
- `client/scripts/shell/gameplay_hud_flow.gd` owns HUD visibility and match-over HUD locking, not result-window rendering.
- `client/scripts/networking/` owns packet routing and WebSocket behavior, not result-window presentation.
- `client/scripts/ui/menu_flow/` owns broader app menu routing, not result-window route execution.

## Tests

### Match results flow tests

- `client/tests/unit/ui/match_results/test_match_results_flow.gd`

These tests verify:

- `show_results()` mounts a result window.
- Calling `show_results()` twice clears the old window and leaves one mounted window.
- Single-player primary button intent emits `replay_requested`.
- Multiplayer primary button intent emits `return_to_lobby_requested`.
- Menu intent emits `return_to_pregame_requested`.
- Quit intent emits `quit_to_main_menu_requested`.

### Match result window tests

- `client/tests/unit/ui/match_results/test_match_result_window.gd`

These tests verify:

- Applying rows creates player score rows.
- Player id, deaths, and score labels are populated.
- No kills label is rendered.
- Lobby/Replay, Menu, and Quit buttons emit their scene-local signals.

### Match-end collaboration tests

- `client/tests/unit/gameplay/match_end/test_match_end_flow.gd`

These tests verify:

- Local elimination does not show match results directly.
- Authoritative room match-over passes session mode and rows to `MatchResultsFlow`.
- Empty or missing result data still opens the result window with empty rows.
- Repeated room match-over refreshes do not repeatedly show results.

## Related docs

- [Match End Flow](./!INDEX.md)
- [match-end-orchestration.md](match-end-orchestration.md) - Client match-end orchestration and handoff into result presentation.
- [Gameplay Menu Flow](../gameplay-menu-flow/!INDEX.md) - Client gameplay menu and match-over overlay menu documentation.
- [HUD And Gameplay UI](../hud-and-gameplay-ui.md) - Gameplay UI roots, HUD behavior, and match-over HUD visibility locking.
- [Gameplay Runtime](../gameplay-runtime/!INDEX.md) - Client gameplay runtime composition and state application docs.
- [Client](../!INDEX.md) - Client service documentation index.

## Notes

Legacy documentation grouped match-end orchestration, match-results presentation, and gameplay menu behavior together. Current documentation splits them because the implementation has separate owners.

`MatchResultsFlow.clear()` removes the mounted window from its parent before queueing it for deletion. This keeps repeated result presentation from stacking duplicate windows.

`MatchResultWindow.clear_rows()` only queues existing `PlayerScoreRow` children for deletion. Static header, separator, spacer, and button nodes remain part of the scene.

The current visible result table intentionally omits kills. Tests assert that no `GameKillsLabel` exists in the rendered row.

The `won` value is currently preserved in rows passed from match-end orchestration but is not displayed by the result row scene.