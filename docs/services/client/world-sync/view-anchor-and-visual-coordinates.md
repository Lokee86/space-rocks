# View Anchor And Visual Coordinates

Parent index: [World Sync](./!INDEX.md)

## Purpose

This document describes the client ViewAnchor and visual-coordinate implementation.

It documents how the Godot client converts server-bounded world positions into continuous visual positions, how ViewAnchor acts as the active render origin, how render-anchor selection supports local-player and spectate presentation, and how world sync exposes coordinate conversion helpers to gameplay presentation, targeting, pointer, camera, and debug flows.

## Overview

The game server owns authoritative world positions. Those positions are bounded by the server world dimensions and may wrap from one edge of the world to the opposite edge.

The client owns visual continuity. It should not present a player, projectile, asteroid, pickup, event effect, camera, or targeting read model as if it snapped across the screen just because the server coordinate wrapped. Instead, world sync maps bounded server coordinates into continuous visual coordinates relative to the current ViewAnchor.

The core rule is:

```text
Server position = authoritative bounded world coordinate.
Visual position = client presentation coordinate relative to the active render anchor.
```

ViewAnchor is the client-side render origin. Normal gameplay anchors presentation to the local player. Spectate and view-target behavior may anchor presentation to a different player. Entity sync owners then place other world objects by calculating the shortest wrapped delta from the anchor's server position to each entity's server position and adding that delta to the anchor's visual position.

The current implementation keeps this responsibility under world sync:

```text
WorldSync
-> PlayerRenderApi
   -> ViewAnchorSync
   -> PlayerMeaningApi
-> WorldWrap
```

`WorldSync` owns the public coordinate conversion API used by other gameplay presentation flows.

`PlayerRenderApi` owns the active player-render API boundary. It chooses and applies the active render anchor, delegates ViewAnchor updates, and delegates player meaning updates.

`ViewAnchorSync` owns the active ViewAnchor server/visual mapping and wraps the quarantined legacy local-visual-sync implementation.

`PlayerMeaningApi` wraps the quarantined legacy player-sync implementation and exposes player node, remote player, hue, and view-target facts through the active player-render API.

`WorldWrap` owns shortest-delta math for toroidal presentation.

## Code root

* `client/`

## Responsibilities

* Keep server-bounded world positions separate from client visual positions.
* Maintain ViewAnchor as the active world render origin.
* Track the active render anchor's server position and visual position.
* Convert server positions to visual positions relative to the active render anchor.
* Convert visual positions back to server positions for pointer and targeting requests.
* Preserve visual continuity when server positions wrap across world edges.
* Use generated world-size constants for toroidal shortest-delta math.
* Support local-player anchoring during normal gameplay.
* Support view-target anchoring for spectate and camera focus behavior.
* Expose coordinate conversion helpers from `WorldSync`.
* Keep player-render legacy code behind the active `client/scripts/world/player_render/` API.
* Provide stable presentation coordinates for gameplay events and effects.
* Provide server-space pointer coordinates for input and targeting flows.
* Keep camera and world presentation aligned to ViewAnchor rather than directly coupling all consumers to the local player node.

## Does not own

* Server-authoritative world positions.
* Server-side toroidal simulation.
* Player movement authority.
* Collision outcomes.
* Room, match, death, respawn, or spectate authority.
* Packet schema source-of-truth files.
* Raw gameplay packet parsing.
* Gameplay state normalization before world state reaches `WorldSync`.
* Target selection orchestration.
* Input request sending.
* HUD or menu behavior.
* Entity-family sync details for projectiles, asteroids, or pickups.
* Legacy player-render internals.
* Durable player data or profile state.

## Domain roles

### Server coordinate

A server coordinate is an authoritative bounded position from gameplay state.

Server coordinates are the source facts sent by the authoritative game server. They should be used for protocol, simulation, and outbound targeting requests.

### Visual coordinate

A visual coordinate is a client presentation position.

Visual coordinates may move outside the bounded server world range so presentation can remain continuous across wrap boundaries.

### ViewAnchor

ViewAnchor is the active client render origin for gameplay world presentation.

The ViewAnchor stores enough state to map server positions into visual positions. It should be treated as the root reference for world presentation, camera behavior, background movement, visual effects, and target picking that needs stable onscreen coordinates.

### Render anchor

The render anchor is the player whose server/visual position currently drives ViewAnchor.

Normal gameplay uses the local player as the render anchor. Spectate or explicit view-target behavior may use another player.

### Player meaning

Player meaning is the presentation-layer interpretation of player state around the active anchor: which player is local, which players are remote, which player is the current view target, and which player nodes or remote facts are exposed to consumers.

## Protocols and APIs

### Coordinate ownership boundary

Gameplay runtime forwards normalized world state into `WorldSync`.

`WorldSync` then applies that state to player rendering and entity sync. ViewAnchor and visual coordinates are handled inside world sync, not in packet readers or runtime packet fanout.

```text
GameplayStatePacketReader
-> GameplayStateApplyFlow
-> GameplayWorldStateApplyFlow
-> WorldSync
-> PlayerRenderApi
-> ViewAnchorSync
```

### Server-to-visual conversion

`WorldSync.visual_position_for_server_position(server_position)` converts authoritative server-space event or entity positions into client visual-space positions.

This is used by gameplay presentation code that receives server-space positions but must spawn or place visual effects at the correct continuous client location.

The conversion rule is:

```text
visual_position =
  anchor_visual_position
  + shortest_wrapped_delta(anchor_server_position, server_position)
```

The anchor values come from the active ViewAnchor/player-render state.

### Visual-to-server conversion

`WorldSync.server_position_for_visual_position(visual_position)` converts a visual-space pointer or target position back into bounded server space.

This is used by pointer and targeting flows when client presentation needs to send a target request to the server.

The conversion rule is the inverse of the server-to-visual mapping:

```text
server_position =
  anchor_server_position
  + visual_delta_from_anchor
  wrapped back into server world bounds
```

The server remains authoritative over whether the resulting request is valid.

### Shortest wrapped delta

`WorldWrap.shortest_delta(from, to)` calculates the shortest visual movement between two bounded server positions across toroidal world edges.

It uses generated world-size constants from:

```text
client/scripts/generated/constants/constants.gd
```

The underlying source-of-truth constants are generated from shared data, not from the world-sync docs.

### Active anchor selection

`PlayerRenderApi` chooses the active render anchor from available server player state.

Normal behavior uses `self_id`.

If a view target is set and that player exists in the current server player state, the view target can become the active anchor for presentation.

The selected anchor updates `ViewAnchorSync`, and player meaning is applied relative to that anchor.

### View target support

World sync exposes view-target helpers for spectate and camera focus behavior:

```gdscript
focus_camera_on_player(player_id)
set_view_target_player(player_id)
clear_view_target_player()
```

These calls route through the active player-render API rather than directly manipulating legacy player sync.

### Player render API boundary

Active code should use:

```text
client/scripts/world/player_render/player_render_api.gd
client/scripts/world/player_render/view_anchor_sync.gd
client/scripts/world/player_render/player_meaning_api.gd
```

Active code should not import quarantined legacy player-render files directly.

The legacy implementation currently remains under:

```text
client/legacy/player_render/
```

That legacy directory is implementation support behind the active API boundary, not the service documentation target.

## Data ownership

ViewAnchor and visual-coordinate code owns transient client presentation state only.

Current presentation state includes:

* current anchor server position
* current anchor visual position
* current anchor rotation where needed for presentation
* active view-target player id when one is set
* player node and remote-player facts exposed through the active player-render API
* temporary coordinate conversion results
* visual positions exposed to event, targeting, input, camera, spectate, and debug consumers

This state is not durable.

It is not authoritative.

It is cleared or replaced as gameplay state and session lifecycle require.

## World wrap behavior

The world is toroidal. A bounded server position can cross from one world edge to the opposite edge without that representing a long visual movement.

The client therefore treats wrapped movement as a shortest-delta presentation problem.

Example:

```text
World width: 1000
Anchor server x: 990
Entity server x: 10
Naive delta: -980
Shortest wrapped delta: +20
```

The visual entity should appear slightly to the right of the anchor, not far to the left across the entire world.

This applies to players and to any entity or visual effect whose position is converted relative to ViewAnchor.

Entity-specific continuity rules for projectiles, asteroids, and pickups are owned by [Entity Sync Owners](entity-sync-owners.md). This document owns the anchor and coordinate model those sync owners rely on.

## Presentation consumers

### World sync

`WorldSync` owns the public conversion helpers and applies the active anchor basis to entity sync owners.

### Gameplay events and effects

Gameplay event presentation receives server-space positions from server events and uses world-sync conversion to place visual effects at continuous visual positions.

### Input and targeting

Pointer and targeting flows work with visual mouse positions during local presentation, then convert those positions back into server space before building server requests.

Target selection itself is not owned here.

### Spectate

Spectate can change which player presentation follows by setting a view target. World sync keeps this as a presentation anchor concern rather than a server authority concern.

### Debug and telemetry

Debug overlays and telemetry that display world positions should respect the same coordinate boundary: server values are authoritative facts; visual values are presentation positions.

## Code map

### Active world-sync API

* `client/scripts/world/world_sync.gd`
* `client/scripts/world/world_wrap.gd`

### Active player-render API

* `client/scripts/world/player_render/!INDEX.md`
* `client/scripts/world/player_render/player_render_api.gd`
* `client/scripts/world/player_render/view_anchor_sync.gd`
* `client/scripts/world/player_render/player_meaning_api.gd`

### Legacy implementation behind active API

* `client/legacy/player_render/local_visual_sync.gd`
* `client/legacy/player_render/visual_sync_positions.gd`
* `client/legacy/player_render/player_sync.gd`
* `client/legacy/player_render/player_sync_targets.gd`
* `client/legacy/player_render/player_sync_interpolation.gd`
* `client/legacy/player_render/player_sync_lifecycle.gd`
* `client/legacy/player_render/player_sync_presentation.gd`
* `client/legacy/player_render/API.md`

### Runtime callers

* `client/scripts/gameplay/runtime/gameplay_world_state_apply_flow.gd`
* `client/scripts/gameplay/runtime/gameplay_runtime_context.gd`
* `client/scripts/gameplay/state/gameplay_state_apply_flow.gd`

### Gameplay consumers

* `client/scripts/gameplay/events/gameplay_event_controller.gd`
* `client/scripts/gameplay/events/gameplay_event_flow.gd`
* `client/scripts/gameplay/effects/gameplay_effects.gd`
* `client/scripts/gameplay/input/gameplay_pointer_position_provider.gd`
* `client/scripts/gameplay/input/target_visual_picker.gd`
* `client/scripts/gameplay/targeting/gameplay_targeting_context.gd`
* `client/scripts/gameplay/targeting/gameplay_target_candidate_flow.gd`
* `client/scripts/gameplay/targeting/target_request_flow.gd`
* `client/scripts/gameplay/targeting/target_position_source.gd`
* `client/scripts/gameplay/spectate/gameplay_spectate_context.gd`
* `client/scripts/gameplay/spectate/gameplay_spectate_flow.gd`
* `client/scripts/gameplay/spectate/spectate_session_flow.gd`

### Presentation consumers

* `client/scripts/presentation/background/background_controller.gd`
* `client/scripts/presentation/background/background_flow.gd`

### Generated data

* `client/scripts/generated/constants/constants.gd`
* `shared/constants/server_constants.toml`

## Tests

Relevant tests include:

* `client/tests/unit/test_world_sync.gd`
* `client/tests/unit/test_world_wrap.gd`
* `client/tests/unit/test_local_visual_sync.gd`
* `client/tests/unit/test_visual_sync_positions.gd`
* `client/tests/unit/test_player_sync.gd`
* `client/tests/unit/test_player_sync_state.gd`
* `client/tests/unit/world/player_render/test_player_render_api.gd`
* `client/tests/unit/world/player_render/test_view_anchor_sync.gd`
* `client/tests/unit/gameplay/test_gameplay_event_controller.gd`
* `client/tests/unit/gameplay/test_gameplay_target_candidate_flow.gd`
* `client/tests/unit/test_target_request_flow.gd`
* `client/tests/unit/test_target_visual_picker.gd`
* `client/tests/unit/test_mouse_action_flow.gd`

The standard client verification path is the Godot headless GUT test suite for `client/tests`.

## Related docs

* [World Sync](./!INDEX.md)
* [World Sync Coordinator](world-sync-coordinator.md)
* [Entity Sync Owners](entity-sync-owners.md)
* [Gameplay Runtime](../gameplay-runtime/!INDEX.md)
* [Gameplay State Application](../gameplay-runtime/gameplay-state-application.md)
* [Input and targeting](../input-and-targeting.md) - Client input and targeting documentation.
* [Toroidal wrap](../../../systems-design/world/stubs/toroidal-wrap.md) - Stub: toroidal world design documentation.
* [World authority](../../../systems-design/world/stubs/world-authority.md) - Stub: world authority design documentation.
* [Constants pipeline](../../../data/stubs/constants-pipeline.md) - Stub: constants source-of-truth and generated output documentation.
* [Gameplay packets](../../../protocol/stubs/gameplay-packets.md) - Stub: gameplay realtime packet documentation.

## Notes

Legacy toroidal-wrap documentation correctly identified the key boundary: the server owns bounded authoritative positions, while the client owns continuous visual presentation.

`client/scripts/world/player_render/` is the active seam over quarantined legacy player-render code. This service documentation should describe the active seam and only mention legacy files as implementation support behind that seam.

ViewAnchor is a presentation anchor, not a gameplay authority. Changing the active ViewAnchor changes what the client follows and how it converts coordinates; it does not change authoritative player identity, room state, match state, or server simulation.

Targeting consumes coordinate conversion and read models from world sync, but target selection and outbound target requests belong to client input/targeting documentation.
