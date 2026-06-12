# Client Menu Flow

This document defines the canonical client menu-flow design for the final Multiplayer V1.1 client slice.

This is not a separate Phase 6. It is the final client slice for Multiplayer V1.1.

## Implementation Status

Phase 1 / foundation slice is complete and green.
Phase 2 / single-player pregame action slice is complete and green.

Completed Phase 1:

- Main Menu is now a top-level route launcher.
- Main Menu keeps login indicator and logout button.
- Single Player routes to `pregame_menu.tscn` in single-player mode.
- Multiplayer routes to `pregame_menu.tscn` in multiplayer mode.
- PregameMenu script is attached to `pregame_menu.tscn`.
- PregameModePresenter applies Single Player vs Multiplayer labels, visibility, and disabled states.
- Pregame Back returns to Main Menu.
- Old Main Menu multiplayer dialog/sign-in behavior is removed.

Completed Phase 2:

- Play Endless from Pregame starts the old single-player flow.
- PregameMenu clears when gameplay starts.
- Main Menu stays hidden during gameplay.
- Disabled Single Player future buttons remain disabled.
- Pregame Back still returns to Main Menu.

## Remaining Client Slice Plan

1. Sign In screen
2. Multiplayer pre-lobby create/join/logout routing
3. Local Pilot / Guest selection
4. Profile readout transmission
5. Match Results window
6. Stats refresh / final smoke

## Rollout Tracker

- [x] Main Menu route launcher
- [x] Pregame Menu scene mounted by `MenuFlowController`
- [x] Single Player mode presentation
- [x] Multiplayer mode presentation
- [x] Pregame Back returns to Main Menu
- [x] Play Endless from Pregame
- [ ] Sign In screen
- [ ] Multiplayer Create/Join/Logout from Pregame
- [ ] Local Pilot / Guest selector
- [ ] Profile readout transmission
- [ ] Match Results window
- [ ] Stats refresh / final smoke

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
- `TransmissionFlow` owns `ScreenDisplay` mount and clear behavior.
- `ProfileFlow` owns profile queries and view models.
- `LocalPilotFlow` owns guest/local pilot selection.
- `MultiplayerFlow` owns sign-in, create, join, and logout routing.
- `MatchResultsFlow` owns match result population and button behavior.

`pregame_menu.gd` must not own API calls, profile parsing, local profile persistence, room create/join logic, or match row building.

## Main Menu

- Login indicator and logout button remain.
- Single Player routes to `pregame_menu.tscn` in single-player mode.
- Multiplayer routes to `pregame_menu.tscn` in multiplayer mode.
- Quit remains unchanged.
- Options is deferred.

## Pregame Menu

Pregame Menu is a wiring shell only.

It should connect the active mode, route intent, and back behavior, but it should not own feature policy.

## Single Player Mode

- Play Endless is implemented and uses the old Main Menu single-player start behavior.
- Play Endless clears menu UI for gameplay through the menu-flow seam.
- Campaign, Loadout, Provisioner, Buy Scrap, and Rankings are disabled.
- Profile opens `profile_readout.tscn` in `TransmissionScreen/ScreenDisplay`.
- Select Pilot controls the Guest / Local Pilot / New Pilot flow.
- Callsign defaults to Guest.

## Multiplayer Mode

- Signed out opens the Sign In screen.
- Sign In screen behavior:
  - Manual is disabled.
  - Google is disabled.
  - Discord is enabled.
  - Cancel returns to Main Menu.
- Signed in pre-lobby behavior:
  - Create Game is available.
  - Join Game is available.
  - Profile is available.
  - Logout is available.
- Create Game uses the old create behavior.
- Join Game temporarily uses `join_dialog.tscn`.
- Lobby remains unchanged.

## Back Behavior

- If a transmission is open, Back closes it and leaves `ScreenDisplay` blank.
- Otherwise Back returns to Main Menu.
- Back must not log out the online account or clear the selected local pilot.

## Profile Behavior

- `profile_readout.tscn` loads into `TransmissionScreen/ScreenDisplay`.
- Profile displays the currently shown callsign context.
- Profile data is fetched by the profile flow/controller, not by scene scripts.

## Match Results

- `match_result_window.tscn` and `player_score_row.tscn` are used.
- This replaces only room game-over flow.
- Active game, personal death, and personal game-over behavior remain unchanged.
- `LobbyReplayButton` says Replay in single-player and Lobby in multiplayer.
- Multiplayer Lobby uses the existing return-to-lobby flow.
- `MenuButton` leaves the room and returns to Pregame Menu.
- `QuitButton` leaves the room and returns to Main Menu.
