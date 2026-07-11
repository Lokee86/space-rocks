# HUD And Gameplay UI

Parent index: [Client](./!INDEX.md)

## Purpose

This document describes the current client HUD and gameplay-session UI implementation.

It documents how the Godot client mounts gameplay UI, updates HUD presentation from lane-native presentation fanout and gameplay events, protects gameplay UI from gameplay mouse input, presents local death and respawn state, and renders weapon or loadout HUD state.

## Overview

The client owns HUD and gameplay UI as presentation only.

The authoritative facts shown by the HUD come from server-driven lane-applied gameplay state, room state, and gameplay events. The client reads those facts, converts them into local presentation state, and updates Godot scenes and controls. It does not decide score, lives, match-over state, respawn validity, weapon state, cooldown truth, or match results.

The main client scene is:

```text
client/scenes/game.tscn
```

`game.tscn` contains two important UI roots:

```text
UserInterface
GameplayUserInterface
```

`UserInterface` is the top-level `CanvasLayer` for app-level UI such as the main menu, pregame menu, login window, join dialog, and multiplayer lobby.

`GameplayUserInterface` is the gameplay-session UI root. HUD, gameplay menu overlays, match results, and gameplay-session modals belong under this root. `GameplayUserInterface` uses `mouse_filter = IGNORE` so it does not block sibling app or menu screens by itself.

Local elimination presentation is mode-specific: multiplayer may immediately show the local eliminated HUD/menu state, while single-player waits for authoritative room `GameOver`. The room match-over path then hides and locks the HUD, configures match-over UI, requests game-over audio through its owning flow, and presents results.

The HUD scene is:

```text
client/scenes/ui/hud.tscn
```

It contains the visible gameplay HUD controls for score, lives, local death or respawn text, game-over presentation, the embedded live gameplay menu path, and the loadout display container.

Runtime HUD behavior is coordinated by `GameplayHudFlow`. `ClientConnectionService` composes the networking collaborators and exposes the stable `RealtimePacketPipeline` seam. `ClientInboundCoordinator` owns the dispatcher-consumer bindings for realtime packet families; the pipeline owns realtime packet application and readiness, with downstream presentation flowing through `PresentationBridge`. `ClientConnectionService` delegates gameplay presentation orchestration to `GameplaySessionController`, which owns gameplay-session and input acceptance state, player-pause gating, `PresentationBridge` activation/reset/frame flush scheduling, gameplay composition lifecycle, frame sequencing, debug/control routing, and session exits. `PresentationBridge` is the orchestration seam for applied notification handling, pending and coalescing, latest-state retrieval, lane presentation orchestration, local lifecycle presentation, and devtools-state adaptation. Durable local death, respawn, and elimination presentation is reconstructed by `GameplayLocalLifecycleFlow` from authoritative world and session lanes; `event_batch` remains a separate best-effort immediate-effects path. `event_batch` reaches HUD-side event presentation only after compact wire aliases are expanded into readable long-key event dictionaries, and HUD code should not depend on compact aliases.

## Code root

```text
client/
```

## Responsibilities

The client HUD and gameplay UI implementation owns:

* Mounting gameplay-session UI under `GameplayUserInterface`.
* Keeping app, menu, and lobby screens outside `GameplayUserInterface`.
* Showing normal gameplay HUD presentation after gameplay state starts.
* Hiding the room id label during active gameplay HUD display.
* Applying score from overlay or session lane presentation inputs.
* Applying authoritative lives from overlay or session lane state.
* Presenting local death state.
* Presenting the respawn countdown.
* Exposing whether the client can request respawn through HUD presentation state.
* Showing the "Press R to Respawn" prompt only after the respawn countdown reaches zero.
* Clearing stale death presentation when the local player is restored to active state.
* Hiding and locking the HUD after authoritative room match-over.
* Preventing repeated `GameOver` snapshots from reopening normal HUD presentation.
* Hosting the embedded live gameplay menu path that uses nodes inside `hud.tscn`.
* Hosting the match-over overlay parent through `GameplayUserInterface`.
* Protecting mouse clicks over gameplay-session UI from also becoming gameplay input.
* Rendering displayable loadout weapons in the HUD loadout container.
* Rendering limited-ammo labels for displayable weapons.
* Rendering cooldown overlays and ready effects for displayable weapons.
* Clearing HUD presentation and loadout display state on gameplay reset.

## Does not own

The client HUD and gameplay UI implementation does not own:

* Authoritative score calculation.
* Authoritative lives, death, respawn, or elimination decisions.
* Authoritative match-over decisions.
* Authoritative match results.
* Room lifecycle.
* Packet schemas.
* WebSocket transport behavior.
* Server simulation state.
* Weapon or loadout rules.
* Weapon fire validation.
* Cooldown authority.
* Persistent player profile, account, or match-result storage.
* App-level menu, login, pregame, lobby, or route ownership.
* Devtools telemetry overlay ownership.
* Game-over audio playback or one-shot audio gating.
* Match result row calculation or persistence.

## Domain roles

### Gameplay UI root

`GameplayUserInterface` is the scene root for gameplay-session UI.

It owns the place where gameplay HUD, match results, gameplay menu overlays, and gameplay-session modal UI are mounted. It does not own app-level UI.

### HUD presentation surface

`HUD` is the visible gameplay HUD scene.

It presents score, lives, local death or respawn state, game-over text, the embedded live gameplay menu path, and loadout display controls.

### HUD flow

`GameplayHudFlow` owns HUD visibility mechanics and local HUD presentation state.

It stores local presentation flags such as:

```text
hidden_for_match_over
is_dead
is_game_over
can_respawn
current_score
respawn_countdown_remaining
```

These are client presentation facts only. They do not become authoritative gameplay state.

### Lane-native HUD presenter

Overlay and session presentation adapters feed the HUD directly.

Current active path:

```text
OverlayPresentationAdapter
-> GameplayHudFlow.apply_overlay_lane_state(...)

SessionPresentationAdapter
-> GameplayHudFlow.apply_session_lane_state(...)
```

The HUD's current active lane inputs are:

```text
overlay_lane_state.self_id
overlay_lane_state.lives
overlay_lane_state.score
overlay_lane_state.primary_weapon_id
overlay_lane_state.secondary_weapon_id
overlay_lane_state.primary_ammo_policy
overlay_lane_state.secondary_ammo_policy
overlay_lane_state.primary_cooldown_remaining
overlay_lane_state.secondary_cooldown_remaining
overlay_lane_state.primary_ammo_remaining
overlay_lane_state.secondary_ammo_remaining

session_lane_state.player_sessions
session_lane_state.player_lifecycle
```

`GameplayHudFlow.apply_overlay_lane_state(...)` owns receiver-local overlay facts such as score, lives, and loadout or cooldown presentation.

`GameplayHudFlow.apply_session_lane_state(...)` owns player-session and lifecycle-driven HUD readback such as score fallback or session-owned local player facts.

### Local death and respawn presenter

`GameplayDeathFlow` reacts to local self-death events from `event_batch` presentation. That path consumes readable event dictionaries after compact expansion, not compact wire aliases, and remains useful for immediate death animation, sound, and fast self-death response.

The event path is best-effort and is not the durable source for dead HUD, respawn availability, authoritative lives, or eliminated state. Reconstructable local lifecycle presentation comes from `GameplayLocalLifecycleFlow` after `PresentationAdapter` applies authoritative world and decoded session lane state.

If the event path reports zero lives, `GameplayDeathFlow` delegates final local elimination to `MatchEndFlow.handle_local_player_eliminated(lives)`, passing the integer lives value rather than the event dictionary. The lifecycle path independently reconstructs eliminated presentation from authoritative session state. Both paths may call the handler, but only multiplayer accepts immediate local-elimination HUD/menu consequences; single-player waits for room `GameOver`.

`GameplayRespawnFlow` uses `GameplayHudFlow.can_request_respawn()` before sending a respawn request. The respawn path currently emits three structured network diagnostics through [Client Logging](./client-logging.md): `respawn_request_send_started` and `respawn_awaiting_confirmation_marked` at info level, and `respawn_request_blocked` at info level when a local guard blocks the request. These diagnostics are one-shot or guarded by local flow state and do not change respawn permission, gameplay readiness, or server authority.

`GameplayLocalLifecycleFlow` is the local lifecycle presentation owner. It receives world lane state, decoded session lane state, and `self_id` from `PresentationAdapter` and owns reconstructable local active, pending-respawn, and eliminated presentation.

`PlayerLifecycle.status_for(...)` is the canonical lifecycle decoder. It accepts a lifecycle value as a string, a `{state: ...}` record, or a `{status: ...}` record.

For `pending_respawn`, the flow reads the local `player_sessions` record, applies authoritative lives, reads decoded `respawn_cooldown`, stops transient local effects on entry, and calls `GameplayHudFlow.set_dead(cooldown)`. A cooldown of `0.0` is valid and makes respawn immediately available. Unchanged pending-respawn fanout does not restart the countdown; a changed authoritative cooldown refreshes it.

For `eliminated`, it applies authoritative lives and delegates to `MatchEndFlow.handle_local_player_eliminated(lives)`. The handler is presentation-effective only in multiplayer. In single-player, authoritative room `GameOver` owns the later HUD/menu match-over presentation and results handoff. Active restoration requires authoritative `active` lifecycle plus the local ship in world state before stale dead presentation and respawn confirmation are cleared.

### Match-over participant

`MatchEndFlow` owns match-end presentation orchestration.

For authoritative room match-over, it asks the HUD to hide through:

```text
GameplayHudFlow.hide_for_match_over()
```

That sets the match-over visibility lock. While the lock is active, gameplay lane packets cannot re-show the HUD through normal `show_gameplay()` calls.

`MatchEndFlow.reset()` clears the lock, but it does not re-show the HUD. Normal gameplay lane state must start again before the HUD is shown.

### Loadout display presenter

`LoadoutDisplayFlow` owns HUD weapon display nodes under `%LoadoutContainer`.

It reads loadout and cooldown state from overlay-lane HUD inputs and creates a display only for weapons registered by `WeaponDisplayRegistry`.

`LoadoutDisplayFlow` instantiates weapon display scene nodes from `weapon_display.tscn`.

Current behavior:

* `torpedo` is displayable.
* `basic_cannon` is not displayable.
* Empty or unknown weapon ids clear the slot display.
* Limited-ammo weapons show an ammo label.
* Non-limited ammo policies hide the ammo label.
* Cooldown state is shown through `CooldownOverlay`.
* Ready transitions can play ring, sweep, and flash effects.

`client/scenes/ui/weapon_displays/weapon_display.tscn` is the scene backing `WeaponDisplay`.

`WeaponDisplay` owns per-slot icon, ammo, cooldown, and ready-effect presentation.

### Input protection

`HudInputPolicy` protects gameplay-session UI from gameplay input.

`GameplaySessionController._input()` checks devtools input first. It then asks `/root/HudInputPolicy` whether a pressed mouse-button event is over `GameplayUserInterface` or one of its descendants. If so, gameplay input is not allowed to also consume that click.

This policy protects gameplay UI only. It does not protect the whole `UserInterface` canvas layer because app, menu, and lobby screens have separate ownership.

## Protocols and APIs

### HUD lane-state input

HUD presentation is updated from lane-native presentation fanout.

Current active input path:

```text
OverlayPresentationAdapter
-> GameplayHudFlow.apply_overlay_lane_state(...)

SessionPresentationAdapter
-> GameplayHudFlow.apply_session_lane_state(...)
```

Current lane-state inputs used by HUD presentation include:

```text
overlay_lane_state.self_id
overlay_lane_state.lives
overlay_lane_state.score
overlay_lane_state.primary_weapon_id
overlay_lane_state.secondary_weapon_id
overlay_lane_state.primary_ammo_policy
overlay_lane_state.secondary_ammo_policy
overlay_lane_state.primary_cooldown_remaining
overlay_lane_state.secondary_cooldown_remaining
overlay_lane_state.primary_ammo_remaining
overlay_lane_state.secondary_ammo_remaining

session_lane_state.player_sessions
session_lane_state.player_lifecycle
```

HUD-specific usage is intentionally narrow:

```text
overlay_lane_state.lives -> apply_lives(lives)
overlay_lane_state.score -> apply_score(score)
overlay lane loadout or cooldown fields -> loadout_display_flow.apply_player_state(...)
session_lane_state.player_sessions[self_id] -> session-backed local player readback when needed
```

### Local death event input

Immediate local death presentation is driven by `event_batch` output through `GameplayEventLifecycleFlow` and `GameplayDeathFlow`; durable lifecycle presentation is driven by authoritative world and session lanes through `GameplayLocalLifecycleFlow`.

The local self-death path uses:

```text
lives
respawn_delay
```

When `lives > 0`, HUD presentation moves into dead or respawn state.

When `lives == 0`, final local elimination is delegated to `MatchEndFlow.handle_local_player_eliminated(lives)` with the integer lives value. The handler performs immediate HUD/menu orchestration only in multiplayer; single-player waits for authoritative room `GameOver`.

### Local lifecycle and immediate event inputs

Durable local lifecycle reconstruction and immediate event effects are separate paths:

Current path:

```text
authoritative session/world lanes
-> PresentationAdapter
-> GameplayLocalLifecycleFlow.apply_lane_state(world_lane_state, decoded_session_lane_state, self_id)
-> reconstruct active, pending_respawn, or eliminated HUD presentation

event_batch
-> GameplayEventLifecycleFlow
-> GameplayDeathFlow
-> immediate dead presentation or integer-lives elimination handoff
```

`GameplaySessionController` verifies pipeline readiness before this fanout and orchestrates the handoff without owning lifecycle or respawn recovery policy. `GameplayHudFlow.apply_session_lane_state(...)` still applies session-owned score and lives data, but it no longer owns `set_dead(...)` from `respawn_cooldown`; `GameplayLocalLifecycleFlow` owns that durable pending-respawn reconstruction.

### Room match-over input

Room match-over presentation is driven by room state, not local HUD inference.

`RoomSessionController` caches latest room state from room snapshots and room state changes. `GameplaySessionController` provides that room state to `GameplayComposition`, which provides it to `MatchEndFlow`.

When the current room state is `GameOver`, `MatchEndFlow` handles room match-over once and asks HUD presentation to hide and lock. This is the required single-player match-end path and the final-results path in multiplayer, separate from multiplayer local elimination.

### Match results input

Match results are not part of gameplay lane presentation.

`RoomSessionController` caches match results from room snapshots when the snapshot contains a match result with a non-empty match id. `MatchEndFlow` reads that cached result through a provider and passes presentation rows to `MatchResultsFlow`.

HUD does not own match result data or result-window rendering.

### Respawn request gate

HUD presentation state participates in respawn request gating.

The flow is:

```text
Gameplay input
-> GameplayRuntimeContext.request_respawn()
-> GameplayRespawnFlow.request_respawn()
-> GameplayHudFlow.can_request_respawn()
-> connection_service.send_respawn_request()
```

`GameplayHudFlow.can_request_respawn()` returns true only when the local player is dead, the room is not game-over, and the respawn prompt is available.

### Mouse input gate

The gameplay UI input policy accepts only pressed mouse-button events over the gameplay UI root or descendants.

The current preferred method is:

```text
should_gameplay_ui_receive_mouse_event(event, gameplay_ui_root, viewport)
```

A narrower HUD-only fallback still exists:

```text
should_hud_receive_mouse_event(event, hud, viewport)
```

The gameplay-session root method is the current owner because gameplay UI now includes HUD, match results, gameplay menu overlays, and gameplay-session modals.

### HTTP APIs

HUD and gameplay UI do not expose HTTP APIs.

## Data ownership

The HUD owns only local, resettable presentation state.

Current HUD-owned local state includes:

```text
hidden_for_match_over
is_dead
is_game_over
can_respawn
current_score
respawn_countdown_remaining
respawn_timer_template
display_nodes
displayed_weapon_ids
previous_cooldown_remaining
ready_effect_played_for_cooldown
```

This state is not persisted.

The HUD does not store account data, profile data, match results, room state, packet history, or authoritative gameplay facts.

The loadout display reads generated packet field names and generated client constants, but it does not own either source.

## Code map

### Scene roots

* `client/scenes/game.tscn` - Main client scene, `UserInterface`, `GameplayUserInterface`, and mounted HUD instance.
* `client/scenes/ui/hud.tscn` - HUD scene, score or lives labels, local death or respawn UI, game-over container, embedded game menu, and loadout display container.
* `client/scenes/ui/dialogs/game_menu.tscn` - Gameplay menu scene used by live gameplay and match-over overlay paths.
* `client/scenes/ui/dialogs/match_result_window.tscn` - Match results scene mounted under gameplay UI by the match results flow.
* `client/scenes/ui/weapon_displays/weapon_display.tscn` - Scene backing `WeaponDisplay`.

### Gameplay composition and session routing

* `client/scripts/shell/app_entry.gd` - Wires scene nodes into the gameplay session controller.
* `client/scripts/session/gameplay_session_controller.gd` - Owns gameplay-session and input acceptance state, player-pause gating, `PresentationBridge` activation/reset/frame flush scheduling, gameplay composition lifecycle, frame sequencing, debug/control routing, and session exits.
* `client/scripts/gameplay/gameplay_composition.gd` - Constructs HUD, menu, match-end, match-results, shell, spectate, devtools, and presentation flows.
* `client/scripts/shell/gameplay_shell_flow.gd` - Delegates gameplay state, processing, input, reset, and menu lifecycle through focused gameplay flows.
* `client/scripts/gameplay/runtime/gameplay_flow_composer.gd` - Wires runtime ticking, input, devtools, spectate, events, local lifecycle, and match-end dependencies.
* `client/scripts/gameplay/runtime/gameplay_process_flow.gd` - Processes runtime interpolation, server hitbox overlay, HUD ticking, devtools, gameplay input, and spectate processing.

### HUD flow and presentation state

* `client/scripts/shell/gameplay_hud_flow.gd` - Main HUD presentation flow for score, lives, local death, respawn countdown, game-over presentation, loadout display, reset, and match-over visibility lock.
* `client/scripts/shell/gameplay_runtime_tick_flow.gd` - Ticks HUD countdown presentation each frame.
* `client/scripts/protocol/realtime/presentation_bridge.gd` - Owns semantic applied-packet subscription, pending and coalesced realtime presentation, readiness-gated frame flushing, lane fanout orchestration, local lifecycle flow handoff, and devtools-state adaptation.
* `client/scripts/protocol/realtime/overlay_presentation_adapter.gd` - Feeds overlay lane state into `GameplayHudFlow.apply_overlay_lane_state(...)`.
* `client/scripts/protocol/realtime/session_presentation_adapter.gd` - Feeds session lane state into `GameplayHudFlow.apply_session_lane_state(...)`.
* `client/scripts/gameplay/events/gameplay_event_lifecycle_flow.gd` - Wires `event_batch` output into event and death presentation flows.
* `client/scripts/gameplay/events/gameplay_death_flow.gd` - Handles local self-death presentation and delegates final elimination to match-end flow.
* `client/scripts/gameplay/lifecycle/gameplay_local_lifecycle_flow.gd` - Reconstructs active, pending-respawn, and eliminated local HUD presentation from authoritative world/session state.
* `client/scripts/gameplay/respawn/gameplay_respawn_flow.gd` - Gates respawn requests through `GameplayHudFlow.can_request_respawn()`.

### Match-end and gameplay menu collaborators

* `client/scripts/gameplay/match_end/match_end_flow.gd` - Presentation orchestration for local elimination and room match-over; asks HUD to hide or lock on authoritative room match-over.
* `client/scripts/shell/gameplay_menu_flow.gd` - Owns gameplay menu behavior, embedded HUD menu path, and match-over overlay menu instance.
* `client/scripts/ui/match_results/match_results_flow.gd` - Owns result-window mounting, clearing, and result button intent forwarding.
* `client/scripts/session/room_session_controller.gd` - Provides latest room state and cached match result to gameplay presentation flows.

### HUD widgets

* `client/scripts/ui/hud/loadout_display_flow.gd` - Creates, updates, and clears loadout display widgets.
* `client/scripts/ui/hud/weapon_display_registry.gd` - Maps displayable weapon ids to HUD display scene definitions and cooldown totals.
* `client/scripts/ui/hud/weapon_display.gd` - Applies weapon icon, ammo, cooldown, and ready-effect presentation for one weapon display.
* `client/scripts/ui/hud/cooldown_overlay.gd` - Draws cooldown countdown wedge, label, and cooldown-finished signal.
* `client/scripts/ui/hud/ring_highlight.gd` - Draws animated ready ring highlight.
* `client/scripts/ui/hud/ready_sweep_highlight.gd` - Plays shader-driven ready sweep highlight.

### Input protection

* `client/scripts/gameplay/input/hud_input_policy.gd` - Determines whether a pressed mouse-button event is over gameplay UI or HUD controls and should block gameplay input.

### Generated inputs

* `client/scripts/generated/networking/packets/packets.gd` - Generated packet field constants consumed by HUD, death, match-end, and loadout presentation flows.
* `client/scripts/generated/constants/constants.gd` - Generated client constants, including session mode and cooldown values consumed by presentation flows.

### Non-owning boundaries

* `client/scripts/world/` - Owns rendered world entities and interpolation, not HUD controls.
* `client/scripts/devtools/` - Owns devtools windows, telemetry, labels, and overlays, not player-facing HUD.
* `client/scripts/networking/` - Owns WebSocket transport and packet routing, not HUD presentation policy.
* `client/scripts/ui/menu_flow/` - Owns app-level menu routing, not gameplay-session HUD presentation.

## Tests

### HUD and loadout display tests

* `client/tests/unit/ui/hud/test_loadout_display_flow.gd`
* `client/tests/unit/ui/hud/test_weapon_display_registry.gd`
* `client/tests/unit/ui/hud/test_weapon_display.gd`
* `client/tests/unit/ui/hud/test_cooldown_overlay.gd`

These tests verify displayable weapon registration, display creation and clearing, ammo label behavior, cooldown overlay behavior, ready effects, and cooldown-finished signaling.

`test_weapon_display.gd` covers the scene-backed `WeaponDisplay` presentation path.

### Input protection tests

* `client/tests/unit/gameplay/input/test_hud_input_policy.gd`

These tests verify gameplay UI root and descendant hover detection, non-pressed event rejection, null safety, and rejection of controls outside the gameplay UI root.

### State and lifecycle tests

* `client/tests/unit/gameplay/events/test_gameplay_death_flow.gd`
* `client/tests/unit/gameplay/match_end/test_match_end_flow.gd`
* `client/tests/unit/gameplay/lifecycle/`
* `client/tests/unit/gameplay/test_gameplay_flow_composer.gd`
* `client/tests/unit/protocol/realtime/test_lane_native_presentation_adapters.gd`

These tests verify lane-native fanout into HUD, local death handling, local lifecycle reconstruction, match-end HUD hiding, match results presentation handoff, and gameplay flow composition.

### Session and menu collaboration tests

* `client/tests/unit/test_gameplay_session_controller.gd`
* `client/tests/unit/shell/test_gameplay_menu_flow.gd`
* `client/tests/unit/ui/menus/test_game_menu.gd`
* `client/tests/unit/ui/match_results/test_match_results_flow.gd`
* `client/tests/unit/ui/match_results/test_match_result_window.gd`

These tests verify gameplay session lifecycle, gameplay menu behavior, and match-results UI collaborators that mount under or interact with gameplay-session UI.

## Related docs

* [Client](./!INDEX.md)
* [Gameplay Runtime](gameplay-runtime/!INDEX.md)
* [World Sync](world-sync/!INDEX.md)
* [Input and targeting](input-and-targeting.md) - Client input and targeting documentation.
* [Match End Flow](match-end-flow/!INDEX.md) - Client match-end orchestration and match-results presentation documentation.
* [Gameplay Menu Flow](./gameplay-menu-flow/!INDEX.md) - Client gameplay menu and match-over overlay menu documentation.
* [Presentation Bridge](./gameplay-runtime/presentation-bridge.md) - Realtime gameplay presentation bridge ownership and lifecycle documentation.
* [Pickup Presentation](./world-sync/pickup-presentation.md) - Client pickup presentation documentation.
* [Realtime websocket protocol](../../protocol/realtime-websocket-protocol.md) - Realtime WebSocket protocol documentation.
* [Gameplay packets](../../protocol/gameplay-packets.md) - Gameplay packet documentation.
* [Client devtools](../../devtools/client/!INDEX.md)

## Notes

The gameplay-session UI split between `UserInterface` and `GameplayUserInterface`, gameplay UI mouse-input protection, and the rule that match-over packets must not reopen the HUD after authoritative room match-over are current service behavior.

Runtime HUD behavior currently lives in `client/scripts/shell/gameplay_hud_flow.gd`, while HUD widget scripts live in `client/scripts/ui/hud/`.

`HUD` currently has `mouse_filter = PASS` in the scene, while `GameplayUserInterface` has `mouse_filter = IGNORE`. Gameplay input protection is therefore handled by `HudInputPolicy` in `GameplaySessionController`, not by making the whole gameplay UI root consume input.

The HUD scene still contains a `GameOverSound` node, but audio playback and one-shot gating are owned by the gameplay event, effects, and audio path. HUD documentation should not treat that node as audio ownership.
