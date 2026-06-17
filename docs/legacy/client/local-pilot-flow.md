# Local Pilot Flow

## Purpose

This document is the focused reference for the Local Pilot / Guest selector flow.

- `LocalPilotFlow` owns the local pilot selector menu, its subpanel flow, and local-pilot intent routing.
- Broader menu routing, profile readout, auth, and match results stay outside this flow.
- This page carries the detailed behavior that belongs with the Local Pilot flow instead of broad menu-flow documentation.

## Scene Ownership

- The Local Pilot selector is a client menu surface.
- `LocalPilotFlow` owns the selector, its subpanels, and the routing between the main selector and the local-pilot subpanel flow.
- Scene scripts emit intent only:
  - `select_pilot_readout.gd` emits load/create/edit/delete intent.
  - `enter_pilot_id.gd` emits confirm/cancel callsign intent.
  - `confirm_delete.gd` emits confirm/cancel delete intent.
- Scene scripts do not own data-handler calls, profile persistence, or local identity policy.

## Identity Rules

- Local profile identity uses `local_profile_id` internally, not display name.
- Identity-kind values come from `ProfileIdentityKind` constants on the client.
- Guest is the fallback/default selectable row.
- In single-player, Guest is `ACTIVE`.
- In multiplayer signed-out fallback handling, Guest is `OFFLINE`.

## Create Behavior

- CREATE opens `enter_pilot_id.tscn` in the subpanel transmission screen.
- CREATE locks primary transmission input while the subpanel is active.
- CREATE validates callsign input before creating a profile.
- `enter_pilot_id.tscn` is a reusable callsign-entry subpanel.
- `LocalPilotFlow` configures the prompt through `enter_pilot_id.gd.configure_label()`.
- CREATE uses `ENTER CALLSIGN`.
- CREATE creates a local profile through the data-handler.
- CREATE seeds from Guest stats only when the loaded identity is Guest.
- CREATE creates fresh zero stats when the loaded identity is a non-Guest local profile.
- CREATE refreshes the selector list after a successful create.

## Load/Default Behavior

- LOAD stores the selected identity as the active single-player context.
- LOAD persists the selected local profile/default through the data-handler.
- LOAD updates the callsign label.
- LOAD uses `local_profile_id` internally, not display name.
- Guest is the default selectable row when no local profile is selected.
- The selector keeps Guest available as the fallback/default row rather than treating it as a removable local profile.

## Edit Behavior

- EDIT is available only for local profiles, not Guest.
- EDIT opens `enter_pilot_id.tscn` in edit mode through the subpanel transmission.
- EDIT uses `ENTER NEW CALLSIGN` with the current callsign prefilled.
- EDIT confirm updates the display name through the data-handler.
- EDIT cancel preserves the selected pilot and does not call the API.
- The subpanel emits confirm/cancel intent only; `LocalPilotFlow` owns the create/edit API calls.

## Delete Behavior

- DELETE is available only for local profiles, not Guest.
- DELETE opens the delete confirmation sub-panel.
- DELETE sends the API delete only after confirmation.
- DELETE refreshes the selector after a successful delete.
- DELETE cancel closes the sub-panel and preserves the selected pilot.
- `confirm_delete.tscn` and `confirm_delete.gd` are the delete-confirmation scene pair.

## Back/Subpanel Behavior

- Back clears the active subpanel transmission first.
- If no subpanel is active, Back clears the primary transmission.
- Clearing the subpanel restores primary transmission input.
- If neither transmission target is active, Back returns to Main Menu.
- Back must not log out the online account or clear the selected local pilot.

## Data-Handler API Ownership

- `LocalPilotFlow` owns the actual create, load, edit, and delete calls into the data-handler.
- Scene scripts only emit user intent and present UI state.
- The data-handler remains the persistence boundary for local profile creation, selection, update, and removal.
- Selector behavior, default-row behavior, and identity persistence all route through the `LocalPilotFlow` ownership seam.
- In no-tag/deployment builds, local profile endpoints can return `local_profiles_unavailable`; the flow should treat that as a missing local-profile store rather than a user input error.
- Guest transient stats stay separate from local-profile persistence.
- Authenticated account stats continue to come from Rails-backed storage, not local-profile storage.
