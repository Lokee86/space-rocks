# Client Menu Flow

Parent index: [Client Legacy](./!README.md)

This document defines the current client menu-flow ownership, route structure, and implemented menu behavior.

It covers main menu routing, pregame routing, sign-in entry, profile readout entry, local pilot selection, match results entry, and stats refresh behavior.

## Current Baseline

- Main Menu is a top-level route launcher with login indicator and logout button.
- Single Player routes to `pregame_menu.tscn` in single-player mode.
- Multiplayer routes through `MultiplayerEntryFlow` and opens `LoginWindow` when signed out.
- Signed in Multiplayer opens `pregame_menu.tscn` in multiplayer mode.
- PregameMenu is attached to `pregame_menu.tscn` and uses `PregameModePresenter` for labels, visibility, and disabled states.
- Play Endless from Pregame starts the single-player gameplay flow and clears menu UI.
- Discord login is implemented through `LoginWindow`.
- Multiplayer Create, Join, and Logout route through the existing multiplayer flow.
- Profile readout opens through the transmission seam.
- Match Results presentation is implemented.
- Local Pilot / Guest selector is implemented.
- Stats refresh is implemented.

## High-Level Scene Flow

Main Menu -> Pregame Menu -> Sign In / Join Dialog / Profile Transmission / Match Results

The scene route names should identify scenes only.

- `MAIN_MENU`
- `PREGAME_MENU`
- `SIGN_IN_SCREEN`
- `JOIN_DIALOG`
- `MATCH_RESULTS`

Mode and state belong outside route names.

- Single Player mode
- Multiplayer mode
- transmission open/closed state

## Ownership

- Scene scripts emit intent and expose display methods.
- `MenuFlowController` owns scene routing.
- `PregameMenuFlow` owns mode wiring and contextual back behavior.
- `PregameModePresenter` owns button visibility, labels, and disabled states.
- `TransmissionFlow` owns primary and subpanel `ScreenDisplay` mounting and clearing.
- `TransmissionFlow` resolves its display roots by unique scene node names, not hardcoded scene paths.
- `ProfileFlow` owns profile queries and view models.
- `ProfileIdentityKind` owns client identity-kind constants.
- `LocalPilotFlow` owns the local pilot menu/sub-menu flow.
- `MultiplayerFlow` owns sign-in, create, join, and logout routing.
- `MatchEndFlow` owns match-end orchestration.
- `MatchResultsFlow` owns result-window presentation and button intent forwarding.
- `GameplayMenuFlow` remains the permanent Esc/gameplay menu owner, including overlay match-over behavior.
- AppEntry and session-level owners execute the actual route changes that result-window and GameMenu intents request.

`pregame_menu.gd` must not own API calls, profile parsing, local profile persistence, room create/join logic, or match row building.

## Main Menu

- Login indicator and logout button remain.
- Single Player routes to `pregame_menu.tscn` in single-player mode.
- Multiplayer routes to `pregame_menu.tscn` in multiplayer mode.
- Quit remains unchanged.
- Options is unavailable; see [current system limits](../limits/current-system-limits.md).

## Pregame Menu

Pregame Menu is a wiring shell only.

It should connect the active mode, route intent, and back behavior, but it should not own feature policy.

## Single Player Mode

- Play Endless starts the current single-player gameplay flow.
- Play Endless clears menu UI for gameplay through the menu-flow seam.
- Unavailable single-player pregame actions are tracked in [current system limits](../limits/current-system-limits.md).
- Profile opens `profile_readout.tscn` in `TransmissionScreen/ScreenDisplay`.
- Select Pilot opens the Local Pilot selector in the primary `TransmissionScreen`.
- Callsign defaults to Guest.

### Local Pilot

- The Local Pilot selector uses the primary `TransmissionScreen` and the subpanel transmission flow.
- The detailed create/load/edit/delete/default-selector behavior lives in [local-pilot-flow.md](./local-pilot-flow.md).
- Guest remains the fallback/default selectable row.

## Multiplayer Mode

- Signed out opens the Sign In screen.
- Sign In screen behavior:
  - Manual and Google are unavailable; Discord is the implemented sign-in path.
  - Cancel returns to Main Menu.
- Signed in pre-lobby behavior:
  - Create Game is available.
  - Join Game is available.
  - Profile is available.
  - Logout is available.
- Create Game uses the current multiplayer create-room path.
- Join Game opens `join_dialog.tscn`.
- Lobby behavior remains owned by the lobby flow after create or join succeeds.

## Back Behavior

- Back clears the active subpanel transmission first.
- If no subpanel is active, Back clears the primary transmission.
- Clearing the subpanel restores primary transmission input.
- If neither transmission target is active, Back returns to Main Menu.
- Back must not log out the online account or clear the selected local pilot.

## Profile Behavior

- `profile_readout.tscn` loads into `TransmissionScreen/ScreenDisplay`.
- Profile displays the currently shown callsign context.
- Single-player Guest is `ACTIVE`.
- Multiplayer signed-out Guest fallback is `OFFLINE`.
- Profile data is fetched by the profile flow/controller, not by scene scripts.

## Match Results

- `match_result_window.tscn` and `player_score_row.tscn` are used.
- `room_snapshot.match_result` is the data source.
- `RoomSessionController` caches the payload.
- `MatchEndFlow` passes rows to `MatchResultsFlow`.
- `PLAYER / DEATHS / SCORE` are the current columns.
- `kills` is not shown.
- `account_id` / `local_profile_id` stay out of the UI payload.
- Result button route execution remains session/AppEntry-owned.
- `GameplayMenuFlow` remains the permanent Esc/gameplay menu owner.
- See [docs/client/match-end-and-gameplay-ui.md](match-end-and-gameplay-ui.md) for the full match-end and gameplay UI ownership map.