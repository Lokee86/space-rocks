# Input And Targeting

Parent index: [Client](./!README.md)

## Purpose

This document describes how the client service owns local gameplay input translation, mouse action coordination, UI click protection, and client-side targeting presentation.

It documents client implementation responsibility only. It does not define server authority, weapon rules, devtools command policy, or the full cross-system targeting domain.

## Overview

Client input and targeting translate local player intent into gameplay-safe client actions.

The client reads Godot input actions, blocks gameplay input when gameplay UI owns the pointer interaction, coordinates mouse actions, builds target candidates from local presentation state, and sends intent to the realtime server. The server remains authoritative for gameplay results and canonical target state.

The client may predict or present candidate selection locally, but durable target identity comes back from authoritative gameplay state through shared target fields.

Current target identity uses:

```text
target_kind
target_id
```

Supported canonical target kinds are:

```text
player
enemy
pickup
asteroid
bullet
```

The client’s responsibility is to make input predictable and safe without letting local UI state become gameplay authority.

## Code root

* `client/`

## Responsibilities

The client input and targeting implementation owns:

* Reading local Godot input actions.
* Translating raw input into semantic gameplay actions.
* Sending movement, firing, respawn, pause, menu, and targeting intent packets.
* Coordinating pending mouse actions before generic target selection.
* Building target candidates from currently rendered or synchronized world state.
* Sending target selection and deselection intent to the server.
* Reading canonical target identity from server-driven gameplay state.
* Preventing gameplay input from also consuming clicks intended for gameplay UI.
* Keeping gameplay-session UI protection separate from app/menu/lobby screen ownership.

## Does not own

The client input and targeting implementation does not own:

* Authoritative movement simulation.
* Authoritative hit detection.
* Authoritative weapon behavior.
* Authoritative target validity.
* Devtools command-specific target interpretation.
* Durable target state.
* Match rules.
* Loadout or weapon eligibility.
* Server-side target resolution.
* App/menu/lobby routing.

Those belong to server, protocol, data, devtools, domain, or planning documentation as appropriate.

## Domain roles

### Gameplay input translation

`GameplayInputContext` translates raw Godot input events into semantic gameplay actions.

### Mouse action priority

`MouseActionFlow` gives pending mouse actions priority over generic target selection so one click does not both finish a pending action and select a target.

### Gameplay UI input protection

`HudInputPolicy` keeps gameplay input from consuming clicks that belong to `GameplayUserInterface` or its descendants.

### Targeting orchestration

`GameplayTargetingContext` owns target-selection orchestration and request routing.

### Target candidate construction

`GameplayTargetCandidateFlow` builds candidate targets from the currently rendered or synchronized client state.

### Canonical target readback

`TargetPositionSource` exposes the canonical target read model that gameplay input and targeting flows read back from authoritative state.

### Spectate boundary

`GameplaySpectateContext` remains a separate boundary from gameplay targeting. It may reuse synchronized player positions, but it does not mutate canonical gameplay target state.

## Protocols and APIs

The client-facing packet boundary is intentionally narrow.

Outbound intent flows include movement, firing, respawn, pause, menu, and target-intent packets sent through the client networking layer. These are requests, not authority.

Authoritative state readback comes from gameplay state and target read models. The client reads canonical target identity and targetable positions from server-driven state rather than treating local clicks as durable truth.

This document does not define the full packet schema. It only describes how input and targeting code uses the outbound intent path and the authoritative readback path.

## Data ownership

The client owns transient input state and presentation-side target candidates.

The client does not persist targeting data.

Durable or authoritative state is owned elsewhere:

```text
Realtime server
= authoritative target state, movement, hit resolution, validation

Realtime protocol
= packet fields and wire shape for input and target messages

Shared generated packet constants
= client/server field names and packet helpers

Client service
= local input translation, candidate selection, and presentation response
```

Raw keyboard and mouse checks should stay at translation edges.

Gameplay code should prefer semantic actions such as:

```text
move_forward
move_backward
turn_left
turn_right
shoot
open menu
select target
deselect target
spawn/place entity
cancel pending mouse action
respawn
switch spectate target
```

Godot project input bindings are configuration. Client scripts consume those bindings and translate them into service behavior.

The important rule is:

```text
Raw input -> semantic client action -> packet or presentation update
```

Raw click checks should not be scattered across unrelated gameplay scripts.

## Gameplay input packet flow

The normal gameplay movement/fire path is:

```text
Godot input action
-> local player input state
-> input packet
-> NetworkClient
-> realtime server
```

The client sends gameplay input only when the session is active enough to accept gameplay input. Paused gameplay, disconnected network state, and blocking UI state should prevent gameplay input packets from being sent.

The server remains authoritative after input is sent. The client updates presentation from incoming state packets rather than treating local input as confirmed gameplay state.

## Mouse action priority

Mouse input has priority rules.

Pending placement or pending mouse actions take precedence over generic target behavior.

```text
Pending action active
-> left click resolves pending action
-> right click or Escape cancels pending action
-> generic target selection does not run first
```

Generic target selection runs only when no higher-priority pending action owns the mouse input.

This prevents one click from both resolving a specific action and selecting a gameplay target.

## Gameplay UI input protection

Gameplay input must not consume pressed mouse-button events when the pointer interaction belongs to gameplay-session UI.

The protected UI root is:

```text
GameplayUserInterface
```

`GameplayUserInterface` owns gameplay-session UI such as:

```text
HUD
GameMenu overlays
Match Results
gameplay-session modals
```

The broader `UserInterface` CanvasLayer should not be protected as a whole. App/menu/lobby screens are separate UI ownership and should not be folded into gameplay input protection.

Godot remains responsible for topmost `Control` click delivery. The client gameplay input guard only prevents gameplay logic from also consuming a click that belongs to `GameplayUserInterface` or its descendants.

## Targeting flow

The current client targeting ownership chain is:

```text
InputEvent
-> GameplayInputContext
-> MouseActionFlow
-> GameplayTargetingContext
-> GameplayTargetCandidateFlow
-> TargetPositionSource
-> packet send
-> authoritative state confirmation
```

Responsibilities by seam:

```text
GameplayInputContext
= translates raw input events into gameplay-safe semantic actions

MouseActionFlow
= coordinates pending mouse actions and generic mouse action priority

GameplayTargetingContext
= owns target selection orchestration

GameplayTargetCandidateFlow
= builds selectable candidates from available presentation/sync state

TargetPositionSource
= exposes targetable position read models

WorldSync
= exposes target source data needed by targeting without owning targeting policy
```

`WorldSync` should expose the target source needed by the targeting layer. It should not accumulate direct target-position helper methods for each target type when those belong in the targeting read model.

## Target candidate sources

Target candidates may be built from synchronized or rendered state for:

```text
players
enemies
pickups
asteroids
bullets
```

The client may use local presentation state to determine what the player can click or inspect, but candidate presence is not authority.

A client-side candidate means:

```text
The client can request this target.
```

It does not mean:

```text
The target is valid, alive, interactable, or accepted by the server.
```

## Canonical target state

Canonical gameplay target state is server-driven.

The client reads canonical target identity from gameplay state fields:

```text
target_kind
target_id
```

Selection and deselection from the client are requests. The server confirms the canonical target through state updates.

The client should not treat the local clicked candidate as durable target state until the authoritative state confirms it.

## Devtools target boundary

Canonical gameplay target state may be reused by devtools readouts or commands, but devtools target behavior is not owned by normal client input.

A devtools command may use a separate per-tool target, or it may resolve from the canonical gameplay target when the command supports that behavior.

Player-only devtools commands must not treat non-player canonical targets as valid player command targets. For those commands, `target_kind` must be compatible with the command’s expected target type.

## Spectate input boundary

Spectate input is related to camera targeting, but it is not the same as gameplay targeting.

Spectate target selection chooses a presentation/camera subject for a dead or observing player. It does not set canonical gameplay target state and should not send gameplay target selection intent.

The gameplay target system and the spectate camera system may both use synchronized player positions, but they should remain separate ownership seams.

## Code map

Current implementation paths:

* `client/project.godot`
* `client/scripts/gameplay/input/`
* `client/scripts/gameplay/targeting/`
* `client/scripts/gameplay/input/gameplay_input_context.gd`
* `client/scripts/gameplay/input/gameplay_input_flow.gd`
* `client/scripts/gameplay/input/gameplay_pause_input_flow.gd`
* `client/scripts/gameplay/input/hud_input_policy.gd`
* `client/scripts/gameplay/input/mouse_action_flow.gd`
* `client/scripts/gameplay/input/target_visual_candidate.gd`
* `client/scripts/gameplay/input/target_visual_picker.gd`
* `client/scripts/gameplay/targeting/gameplay_targeting_context.gd`
* `client/scripts/gameplay/targeting/gameplay_target_candidate_flow.gd`
* `client/scripts/gameplay/targeting/target_position_source.gd`
* `client/scripts/gameplay/runtime/gameplay_runtime_context.gd`
* `client/scripts/session/gameplay_session_controller.gd`
* `client/scripts/world/world_sync.gd`
* `client/scripts/networking/outbound/client_packet_sender.gd`
* `client/scripts/networking/outbound/gameplay_client_packets.gd`
* `client/scripts/networking/outbound/`
* `client/scripts/networking/client_connection_service.gd`
* `client/scripts/generated/networking/packets/packets.gd`
* `shared/packets/`

## Tests

Relevant tests include:

* `client/tests/unit/gameplay/input/test_hud_input_policy.gd`
* `client/tests/unit/test_gameplay_input_context.gd`
* `client/tests/unit/test_target_request_flow.gd`
* `client/tests/unit/test_target_visual_picker.gd`
* `client/tests/unit/gameplay/test_gameplay_target_candidate_flow.gd`
* `client/tests/unit/test_devtools_target_resolver.gd`
* `client/tests/unit/test_devtools_player_target_model.gd`
* `client/tests/unit/boot/test_session_network_target.gd`

Documentation/test coverage note:

* No additional dedicated spectate-boundary test file was found in the current tree.
* No dedicated canonical target readback test file was found beyond the existing input and targeting coverage above.

## Related docs

* [Client](./!README.md)
* [Services](../!README.md)
* [Game Server](../game-server/!README.md)
* [Realtime Protocol](../../protocol/!README.md)
* [Data](../../data/!README.md)
* [Devtools](../../devtools/!README.md)

## Notes

This document is current canonical service documentation for client input and targeting. The key client invariant is that local input can request gameplay intent, but server state confirms gameplay truth.
