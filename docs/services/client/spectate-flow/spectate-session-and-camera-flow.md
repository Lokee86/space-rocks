---
author: brian
created: "2026-07-19"
document_id: 019f7d55-fb2c-7ec0-aac2-3d38739f63b5
document_type: general
policy_exempt: false
summary: This document describes the current client spectate session, server view-target request, and camera handoff flow.
---
## Spectate Session And Camera Flow

Parent index: [Spectate Flow](./!INDEX.md)

## Purpose

This document describes the current client spectate session and camera handoff flow.

It covers target availability, target cycling, local ViewAnchor selection, server view-target control requests, and the retry behavior needed when recipient-specific network interest has not yet delivered the selected target's detailed ship presentation.

## Overview

The client spectate flow owns the availability of spectate targets, the cycling of those targets, the handoff into spectate mode, and the camera/view-target handoff that follows from the selected spectate target.

`GameplaySpectateContext` wires the spectate flow to gameplay menu state, world sync, and the connection service. `SpectateSessionFlow` keeps the spectate menu state updated from gameplay lifecycle state. `GameplaySpectateFlow` starts spectating, advances to the next target, sends the selected view target to the server, and repeatedly attempts local camera focus until the selected target's detailed ship presentation is available.

Spectate target selection is separate from gameplay targeting. Spectate uses player lifecycle availability and view-target handoff. It sends `set_view_target_request` or `clear_view_target_request`, not canonical combat-target selection packets.

## Code root

```text
client/scripts/gameplay/spectate/
```

## Responsibilities

The client spectate session and camera flow owns:

- Collecting spectate-available player IDs from gameplay lifecycle state.
- Excluding the local player and lifecycle-ineligible players.
- Beginning spectate mode from the current available target.
- Cycling to the next available target.
- Storing the selected spectate target independently from immediate render-node availability.
- Setting the selected target on world sync as the local view target.
- Attempting local camera focus immediately and again during each spectate process tick.
- Sending `set_view_target_request` to the server when the selected target changes.
- Sending `clear_view_target_request` when spectate state resets.
- Opening the spectate menu while spectating is active.
- Keeping spectate selection separate from gameplay targeting rules.

## Does not own

The client spectate flow does not own:

- Server player or lifecycle authority.
- Server network-interest policy.
- Authoritative validation of the requested view target.
- Gameplay target state.
- General world entity synchronization.
- WebSocket transport mechanics.
- Detailed player-node creation or lifecycle.
- Camera interpolation and ViewAnchor coordinate math.
- Gameplay menu presentation generally.

## Domain roles

### Spectate target availability

`SpectateMenuState` reads `self_id` and `player_lifecycle` from gameplay state and builds the available spectate target list from lifecycle truth.

The current implementation excludes:

- the local player
- players marked `Dead`
- players marked `GameOver`

Target availability does not require a currently rendered detailed ship node. This matters after network interest is enabled: the lifecycle/session read model can identify a valid player before the selected target's detailed ship presentation has entered the recipient's interest set.

### Beginning spectate mode

`GameplaySpectateFlow.begin_spectating()` asks the menu state for the current target and passes it to `_set_target(...)`.

A non-empty target immediately becomes the selected spectate target and marks spectating active. The flow does not require `focus_camera_on_player(...)` to succeed on that first call.

### Cycling targets

`GameplaySpectateFlow.request_cycle_target()` advances to the next available target only while spectating is active. The new target replaces the previous target locally and is sent to the server.

### Server interest-anchor handoff

For each selected target, the flow sends:

```text
set_view_target_request(view_target_player_id)
```

through the normal client connection service and WebSocket control path.

The server stores that selection on the WebSocket session. When recipient-specific realtime output is built, the selected target's coarse locator becomes the interest-camera anchor and the selected target ship is always included in detailed ship interest.

This request changes presentation delivery for the receiving session. It does not set canonical gameplay target state and does not grant simulation authority to the client.

### Local world-sync and camera handoff

The flow calls:

```text
world_sync.set_view_target_player(target_player_id)
world_sync.focus_camera_on_player(target_player_id)
```

`set_view_target_player` records the intended render anchor even before the detailed player node exists. `focus_camera_on_player` succeeds only when the detailed remote player node is currently available.

`GameplaySpectateFlow.process()` repeats the set/focus handoff while spectating. This lets camera focus complete after the server includes the target in interest and the client receives the corresponding lifecycle create/full presentation.

### Reset behavior

Reset performs both sides of the view-target teardown:

```text
clear_view_target_request
WorldSync.clear_view_target_player()
```

The clear request is sent only when a target had been selected and the connection service can send packets. Local spectate state is cleared regardless of transport availability.

## Protocols and APIs

### Configuration API

`GameplaySpectateContext.configure(...)` and `GameplaySpectateFlow.configure(...)` receive:

```text
menu_flow
spectate_menu_state
world_sync
connection_service
```

The connection service is optional for local/test composition. Without it, local spectate selection and focus can still run, but the server cannot re-anchor recipient network interest.

### View-target packets

The generated client packet helpers are:

```text
Packets.set_view_target_request_packet(player_id)
Packets.clear_view_target_request_packet()
```

These packets travel over WebSocket control rather than a gameplay WebRTC DataChannel. The server handles them in gameplay inbound routing and stores the current selection on the session.

### Process API

`GameplaySpectateContext.process()` delegates to `GameplaySpectateFlow.process()` each gameplay process tick. The process method is intentionally idempotent: it reapplies the current view-target intention and retries camera focus without cycling or sending another packet.

## Data ownership

The spectate flow owns transient client state:

```text
is_spectating
target_player_id
```

`SpectateMenuState` owns the transient available-target list. `WorldSync` and player-render code own local view-target/render-anchor state. The server WebSocket session owns the recipient's current requested view-target ID used by network-interest projection.

None of this state is durable or persisted.

## Code map

### Primary implementation

```text
client/scripts/gameplay/spectate/gameplay_spectate_context.gd
client/scripts/gameplay/spectate/gameplay_spectate_flow.gd
client/scripts/gameplay/spectate/spectate_menu_state.gd
client/scripts/gameplay/spectate/spectate_session_flow.gd
```

### Collaborators

```text
client/scripts/world/world_sync.gd
client/scripts/world/player_render/player_render_api.gd
client/scripts/networking/client_connection_service.gd
client/scripts/networking/outbound/client_packet_sender.gd
client/scripts/generated/networking/packets/packets.gd
client/scripts/ui/menus/game_menu.gd
client/scripts/gameplay/gameplay_composition.gd
```

### Server request consumers

```text
services/game-server/internal/networking/inbound/gameplay.go
services/game-server/internal/networking/inbound_adapter.go
services/game-server/internal/networking/websocket_session.go
services/game-server/internal/protocol/realtime/network_interest.go
```

### Schema source

```text
shared/packets/gameplay.toml
shared/packets/outputs.toml
```

## Tests

Focused spectate flow coverage lives in:

```text
client/tests/unit/gameplay/spectate/test_gameplay_spectate_flow.gd
```

Adjacent coverage includes:

```text
client/tests/unit/gameplay/spectate/test_spectate_menu_state.gd
client/tests/unit/world/player_render/test_player_render_api.gd
services/game-server/internal/networking/gameplay_packets_test.go
services/game-server/internal/networking/inbound/context_test.go
services/game-server/internal/protocol/realtime/network_interest_test.go
```

Verification should prove:

- Spectating can start before the detailed target node exists.
- The selected target is retained while camera focus is pending.
- `set_view_target_request` is sent when selection begins or changes.
- Process ticks retry focus without repeatedly sending the request.
- Reset sends one clear request and clears local view-target state.
- The server-selected target anchors recipient interest and remains detailed-ship relevant.

## Related docs

- [Spectate Flow](./!INDEX.md)
- [Gameplay Runtime](../gameplay-runtime/!INDEX.md)
- [World Sync](../world-sync/!INDEX.md)
- [View Anchor And Visual Coordinates](../world-sync/view-anchor-and-visual-coordinates.md)
- [Input And Targeting](../input-and-targeting.md)
- [Gameplay Menu Flow](../gameplay-menu-flow/!INDEX.md)
- [Gameplay Packets](../../../protocol/gameplay-packets.md)
- [Game Server Network Interest](../../game-server/networking/network-interest.md)

## Notes

The selected spectate target has two related but separate effects: it requests recipient-specific server interest around that player, and it selects the local ViewAnchor/render subject. The server request can succeed before the client has received the detailed ship lifecycle create needed for local camera focus.
