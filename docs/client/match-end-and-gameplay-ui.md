# Match End And Gameplay UI

This document defines the client-side ownership split for match-end presentation and the gameplay-session UI hierarchy.

The goal is simple: gameplay can end, the HUD can be hidden, and match-over UI can appear without blurring ownership between app/menu screens and gameplay-session screens.

## Purpose

- Keep match-end orchestration in one gameplay-owned seam.
- Keep app/menu/lobby screens separate from gameplay-session UI.
- Prevent gameplay packets from reopening HUD or modal gameplay UI at the wrong time.
- Preserve clear route ownership so UI intent can bubble outward without scene scripts performing navigation directly.

## Scene Ownership

- `client/scenes/game.tscn` owns the top-level client scene tree for the live game client.
- `UserInterface` is the CanvasLayer node in `client/scenes/game.tscn`.
- `GameplayUserInterface` is the gameplay-session UI `Control` root.
- `UserInterface` owns app/menu/lobby screens.
- `GameplayUserInterface` owns gameplay-session UI.

## Gameplay UI Roots

- `GameplayUserInterface` is the mount root for gameplay-session UI.
- `HUD`, Match Results, overlay `GameMenu`, and gameplay-session modals live under `GameplayUserInterface`.
- `GameplayUserInterface` should use `mouse_filter = IGNORE` so it does not block sibling app/menu screens.
- `UserInterface` remains the parent root for Main Menu, Pregame Menu, LoginWindow, JoinDialog, and MultiplayerLobby.
- Gameplay-session UI must not be mounted under `UserInterface` by convenience.

## MatchEndFlow

- `MatchEndFlow` is the client-side orchestration seam for match-end presentation.
- It does not decide authoritative match results.
- It reacts to client-observed facts and routes presentation requests to the owning flows.
- It keeps local elimination and room match-over separate.
- It requests HUD hiding/locking from `GameplayHudFlow` and does not own raw HUD visibility itself.
- It requests match-over overlay behavior from `GameplayMenuFlow`.
- It requests result-window presentation from `MatchResultsFlow`.
- It must guard room match-over presentation so repeated `GameOver` snapshots do not remount result UI.
- It does not own scoring, winner calculation, result persistence, audio playback/gating, or final route execution.

## Local Elimination Versus Room Match Over

- MatchEndFlow must keep two states separate:

1. Local elimination: the local player has reached `lives == 0`.
2. Room match-over: the authoritative room state is `GameOver`.

- Local elimination may update HUD/menu state and request game-over audio.
- Local elimination must not show Match Results.
- Room match-over hides and locks HUD presentation, enables match-over overlay menu mode, requests game-over audio, and shows Match Results.
- Repeated `GameOver` snapshots must not remount the result window.

## MatchResultsFlow

- `MatchResultsFlow` owns `match_result_window.tscn`.
- It owns mounting the result window under the gameplay UI mount root.
- It owns clearing the mounted result window.
- It owns forwarding button intent from the result window outward.
- It does not own navigation, route changes, or persistence.
- It should receive rows from the match-end orchestration seam and present them, not derive wider game policy.

## GameplayMenuFlow

- `GameplayMenuFlow` remains the permanent owner of gameplay menu behavior.
- It owns embedded HUD menu behavior for normal gameplay.
- It owns the overlay `GameMenu` path used during match-over.
- It decides which game menu instance is active and when it is shown or hidden.
- It does not own app/menu/lobby routing.
- It does not own match-end orchestration or result-window ownership.

## HUD Visibility

- `GameplayHudFlow` owns HUD visibility mechanics.
- `MatchEndFlow` requests match-over hiding through `GameplayHudFlow`.
- `GameplayHudFlow.hide_for_match_over()` should hide the HUD and set a match-over visibility lock.
- `GameplayHudFlow.clear_match_over_visibility_lock()` should release the lock without showing the HUD.
- Gameplay state packets should not be able to re-show the HUD while the match-over lock is active.
- HUD should show again only when normal gameplay state starts and `GameplayHudFlow.show_gameplay()` is allowed.

## Audio Ownership

- Audio playback, delay, and one-shot sound gating remain owned by the gameplay event/effects/audio path.
- `MatchEndFlow` may request game-over audio through the gameplay orchestration seam.
- `MatchEndFlow` does not play sounds itself.
- `MatchEndFlow` does not own gating rules for repeated audio playback.

## Input Protection

- Gameplay input must not consume clicks over `GameplayUserInterface`.
- The input policy should protect clicks over `GameplayUserInterface` descendants from gameplay input.
- Godot decides which topmost `Control` receives clicks.
- The input policy only prevents gameplay from also consuming those clicks.
- `UserInterface` should not be protected as a whole.
- App/menu/lobby screens remain outside gameplay-session click protection.

## Route Ownership

- Result button routes should bubble outward from the result window.
- `MenuButton` should express intent to leave match-over presentation, not perform navigation itself.
- `QuitButton` should express intent to leave to the main menu, not perform navigation itself.
- `LobbyReplayButton` should express replay or lobby intent, depending on mode.
- Result-window and GameMenu intents bubble outward through composition/session controllers.
- AppEntry and session-level owners execute route changes.
- `MatchResultsFlow` and `GameplayMenuFlow` should forward intent, not perform outer navigation.

## Match Results Data Path

The authoritative ownership chain is `room.MarkGameOver / ResolvedMatchSummary -> BuildRoomSnapshot.MatchResult -> room_snapshot.match_result -> RoomSessionController.latest_match_result -> MatchEndFlow provider -> MatchResultsFlow`.

`MatchEndFlow` remains presentation orchestration only.

`MatchResultsFlow` owns mounting, clearing, button-intent forwarding, and row rendering.

The result payload is presentation-safe and excludes `account_id` and `local_profile_id`.

The current Match Results columns are `PLAYER`, `DEATHS`, and `SCORE`. `kills` is not currently displayed or tracked there.

## What Not To Do

- Do not mount gameplay-session UI under `UserInterface`.
- Do not mount app/menu/lobby screens under `GameplayUserInterface`.
- Do not let gameplay packets re-open the HUD after authoritative room match-over.
- Do not re-show the HUD from `MatchEndFlow.reset()`.
- Do not make `MatchEndFlow` own raw HUD visibility.
- Do not make `MatchResultsFlow` own route changes or persistence.
- Do not make `GameplayMenuFlow` own match-end orchestration.
- Do not treat local elimination as room match-over.
- Do not add result data to the ticked gameplay `StatePacket`.
- Do not bypass the `GameplayUserInterface` input guard by relying on the whole `UserInterface` CanvasLayer.
