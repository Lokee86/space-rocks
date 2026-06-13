# Client Mouse Input Reference

Mouse input should route through semantic actions, not scattered raw click checks across gameplay scripts.

Related references:

- [Canonical targeting reference](../server/targeting.md)
- [Devtools toggles and player-target controls](../devtools/toggles.md)
- [Match end and gameplay UI ownership](match-end-and-gameplay-ui.md)

## Semantic Actions

Current semantic mouse/input concepts:

- select target
- deselect target
- spawn/place entity
- cancel pending mouse action

Raw mouse button checks should stay at input binding/translation edges, then map into these semantic actions.

## Priority Rules

Pending placement/action has priority over generic target behavior.

- Pending placement/action owns left click until it is completed or canceled.
- Right click or Escape cancels pending action before generic deselect behavior runs.

## Gameplay UI Input Protection

- `UserInterface` is the CanvasLayer in `client/scenes/game.tscn`.
- `GameplayUserInterface` is the gameplay-session UI root.
- HUD, Match Results, and overlay `GameMenu` live under `GameplayUserInterface`.
- Gameplay input should not consume pressed mouse-button events when the hovered `Control` is `GameplayUserInterface` or one of its descendants.
- The policy should not protect the whole `UserInterface`, because app/menu/lobby screens are separate ownership.
- Godot handles topmost `Control` click delivery; the gameplay input policy only gates gameplay input.
- See [Match End And Gameplay UI](match-end-and-gameplay-ui.md) for the full ownership map.

## Target Selection and Server Confirmation

Target selection is requested from client input, then confirmed by canonical server-driven gameplay state.

- Client selection actions send selection/deselection intent.
- Canonical target identity is read from authoritative gameplay state updates (`target_kind` + `target_id`), not from local click assumptions.

## Targeting Ownership

Targeting sits above `MouseActionFlow`.

Current client targeting flow:

`InputEvent` -> `GameplayInputContext` -> `MouseActionFlow` -> `GameplayTargetingContext` -> candidate source / picker / packet send

- `MouseActionFlow` remains the lowest-level mouse/input action coordinator.
- `GameplayTargetingContext` owns target selection orchestration.
- `GameplayTargetCandidateFlow` builds target candidates, including players, pickups, asteroids, and bullets.
- `TargetPositionSource` owns targetable position read models.
- `WorldSync` only exposes `target_source()` for targeting and no longer owns direct target-position methods.
- For pickup-specific targeting details, see [pickup system design](../design/pickups.md).
