# Client Menu Flow

This document defines the canonical client menu-flow design for the final Multiplayer V1.1 client slice.

Phase 5 is complete. Phase 6 / Match Results is complete and green.

## Implementation Status

Phase 1 / foundation slice is complete and green.
Phase 2 / single-player pregame action slice is complete and green.
Phase 3 / Sign In screen slice is complete and green.
Phase 4 / Multiplayer pre-lobby actions slice is complete and green.
Phase 5 / Profile readout transmission slice is complete and green.

Completed Phase 1:

- Main Menu is now a top-level route launcher.
- Main Menu keeps login indicator and logout button.
- Single Player routes to `pregame_menu.tscn` in single-player mode.
- Multiplayer routes through `MultiplayerEntryFlow`.
- Signed out Multiplayer opens `LoginWindow`.
- Signed in Multiplayer opens `pregame_menu.tscn` in multiplayer mode.
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

Completed Phase 3:

- Main Menu Multiplayer is auth-aware through `MultiplayerEntryFlow`.
- Signed-out Multiplayer opens `LoginWindow`.
- Discord login works from `LoginWindow` using the existing auth flow.
- Back from `LoginWindow` returns to Main Menu.
- Manual and Google login remain disabled.
- Signed-in Multiplayer opens Pregame Menu in Multiplayer mode.
- Successful Discord auth clears `LoginWindow` and routes to Multiplayer Pregame.

Completed Phase 4:

- Signed-in Multiplayer Pregame Create uses the existing create-room path.
- Create clears Pregame UI before room/lobby transition.
- Join opens `join_dialog.tscn`.
- `JoinDialog` validates empty room code and stays open.
- JoinDialog Cancel returns to Multiplayer Pregame.
- Valid Join clears UI and uses the existing join-room path.
- Logout from Multiplayer Pregame returns to Main Menu signed out.
- Lobby Leave now sends leave-room, clears Lobby, and returns to Multiplayer Pregame without logging out.

Completed Phase 5:

- Profile button opens `profile_readout.tscn` through the transmission seam.
- Profile readout mounts under `TransmissionScreen/ScreenDisplay`.
- Profile readout fills callsign, activity status, and stat labels from the profile flow.
- Single-player guest profile reads are supported.
- Multiplayer authenticated account profile reads are supported.

Completed Phase 6:

- Match Results window is complete and green.
- `room_snapshot.match_result` is the data source.
- `RoomSessionController` caches the payload.
- `MatchEndFlow` passes rows to `MatchResultsFlow`.
- Result rows render as `PLAYER / DEATHS / SCORE`.
- `kills` is not shown.
- Result button route execution remains session/AppEntry-owned.

## Remaining Client Slice Plan

1. Local Pilot / Guest selector
2. Stats refresh / final smoke

## Rollout Tracker

- [x] Main Menu route launcher
- [x] Pregame Menu scene mounted by `MenuFlowController`
- [x] Single Player mode presentation
- [x] Multiplayer mode presentation
- [x] Pregame Back returns to Main Menu
- [x] Play Endless from Pregame
- [x] Sign In screen
- [x] Multiplayer Create/Join/Logout from Pregame
- [x] Lobby Leave returns to Multiplayer Pregame
- [x] Profile readout transmission
- [x] Match Results window
- [ ] Local Pilot / Guest selector
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
- `room_snapshot.match_result` is the data source.
- `RoomSessionController` caches the payload.
- `MatchEndFlow` passes rows to `MatchResultsFlow`.
- `PLAYER / DEATHS / SCORE` are the current columns.
- `kills` is not shown.
- `account_id` / `local_profile_id` stay out of the UI payload.
- Result button route execution remains session/AppEntry-owned.
- `GameplayMenuFlow` remains the permanent Esc/gameplay menu owner.
- See [docs/client/match-end-and-gameplay-ui.md](match-end-and-gameplay-ui.md) for the full match-end and gameplay UI ownership map.
