## Spectate Session And Camera Flow

Parent index: [Spectate Flow](./!README.md)

## Purpose

This document describes the current client spectate session and camera handoff flow.

## Overview

The client spectate flow owns the availability of spectate targets, the cycling of those targets, the handoff into spectate mode, and the camera/view-target handoff that follows from the selected spectate target.

`GameplaySpectateContext` wires the spectate flow to gameplay menu state and world sync. `SpectateSessionFlow` keeps the spectate menu state updated from gameplay state. `GameplaySpectateFlow` starts spectating, advances to the next target, opens the spectate menu when spectating is active, and hands off the selected target to world sync and camera focus.

Spectate target selection is separate from gameplay targeting. Spectate uses player lifecycle availability and view-target handoff, while gameplay targeting continues to own combat or interaction target rules.

## Code root

```text
client/
```

Primary implementation areas:

```text
client/scripts/gameplay/spectate/
client/scripts/gameplay/
client/scripts/ui/menu_flow/
```

## Responsibilities

The client spectate session and camera flow owns:

* collecting spectate-available player ids from gameplay state
* excluding the local player from spectate target availability
* excluding dead and game-over players where lifecycle state marks them unavailable
* beginning spectate mode from the current available target
* cycling to the next available spectate target
* syncing spectate menu state from gameplay state
* opening spectate menu state when spectating is active
* handing the chosen target to world sync as the active view target
* handing the chosen target to the camera focus path
* keeping spectate selection separate from gameplay targeting rules

## Does not own

The client spectate session and camera flow does not own:

* server player authority
* gameplay targeting rules
* world-sync entity synchronization generally
* gameplay menu presentation generally
* combat target selection
* server-side lifecycle policy
* devtools behavior

## Flow behavior

### Spectate target availability

`SpectateMenuState` reads `self_id` and `player_lifecycle` from gameplay state and builds the available spectate target list from that lifecycle data.

The current implementation excludes:

* the local player
* players marked `Dead`
* players marked `GameOver`

If no available targets remain, spectate cannot begin.

### Beginning spectate mode

`GameplaySpectateFlow.begin_spectating()` asks the menu state for the current target and then hands that target to world sync.

If world sync accepts the target focus, the flow marks spectate mode active.

### Cycling targets

`GameplaySpectateFlow.request_cycle_target()` advances to the next available spectate target only while spectating is already active.

The flow updates world sync with the new view target and then requests camera focus for the same player id.

### Menu-state integration

`SpectateSessionFlow.apply_gameplay_state()` keeps the spectate menu state aligned with incoming gameplay state.

`GameplaySpectateContext.configure()` wires the spectate request signal from the menu flow so the spectate flow can begin spectating from the current menu state.

### World-sync and camera handoff

The spectate flow hands the chosen target to world sync through the view-target path and then hands the same target to the camera focus path.

That handoff keeps spectate presentation aligned with the selected player instead of treating spectate selection as a gameplay target request.

## Code map

Primary implementation files:

```text
client/scripts/gameplay/spectate/gameplay_spectate_context.gd
client/scripts/gameplay/spectate/gameplay_spectate_flow.gd
client/scripts/gameplay/spectate/spectate_menu_state.gd
client/scripts/gameplay/spectate/spectate_session_flow.gd
```

## Tests

No focused test is documented yet for this flow.

The closest verification boundary is the spectate flow implementation in:

```text
client/scripts/gameplay/spectate/
```

## Related docs

* [Spectate Flow](./!README.md)
* [Gameplay Runtime](../gameplay-runtime/!README.md)
* [World Sync](../world-sync/!README.md)
* [Input And Targeting](../input-and-targeting.md)
* [Gameplay Menu Flow](../gameplay-menu-flow/!README.md)
* [Client](../!README.md)
* [Services](../../!README.md)

## Notes

This document captures current spectate session and camera handoff behavior. It stays separate from gameplay targeting and does not describe future menu or spectate policy.
