---
author: brian
created: "2026-07-19"
document_id: 019f7d55-fb2c-78db-afe0-ef4f1ed3c27d
document_type: general
policy_exempt: false
summary: This document describes the client WorldSync coordinator.
---
# World Sync Coordinator

Parent index: [World Sync](./!INDEX.md)

## Purpose

This document describes the client `WorldSync` coordinator.

It explains how the client applies world lane state to world presentation seams, how `WorldSync` delegates entity-family synchronization, how interpolation is coordinated, and how targeting and presentation code access world-sync read models.

## Overview

`WorldSync` is the client-side coordinator for rendering server-authoritative world state.

`WorldSync` is a concrete Godot `class_name` contract used by gameplay lifecycle and presentation composition. `WorldLaneState` is the typed world-lane input contract at this boundary; callers pass accumulated lane state directly rather than relying on dynamic method compatibility.

It does not parse raw packets and does not decide gameplay outcomes. RealtimeRouter mutates world lane state, the RealtimePacketPipeline refreshes `RealtimePresentationState`, and WorldPresentationAdapter receives `world_lane_state` from that wrapper before forwarding it into WorldSync. `WorldSync` then delegates the actual player, projectile, asteroid, and pickup presentation work to focused sync owners.

The current runtime path is:

```text
RealtimeRouter.route_lane_packet(packet)
-> lane appliers mutate world_lane_state
-> RealtimePacketPipeline.refresh_presentation_state(...)
-> RealtimePresentationState.world_lane_state
-> WorldPresentationAdapter.apply_world_lane_state(...)
-> WorldSync.apply_world_lane_state(world_lane_state)
```

`WorldSync` also owns the composition of the target-position read model used by gameplay targeting flows. It exposes `target_source()`, but targeting orchestration stays outside world sync.

Per-frame interpolation is triggered by gameplay runtime:

```text
GameplayRuntimeContext.process(delta)
-> WorldSync.interpolate(delta)
```

`WorldSync` then ticks interpolation for player rendering, projectiles, asteroids, and pickups using the generated player interpolation constant.

## Code root

* `client/`

## Responsibilities

* Configure the world-sync presentation seams for a gameplay session.
* Create and configure:

  * `PlayerRenderApi`
  * `ProjectileSync`
  * `AsteroidSync`
  * `PickupSync`
  * `TargetPositionSource`
* Store the current `self_id` for world-sync read-model access.
* Apply lane-applied server world state to the client world.
* Remove missing player, projectile, asteroid, and pickup nodes before applying new state.
* Delegate player presentation to `PlayerRenderApi`.
* Delegate projectile presentation to `ProjectileSync`.
* Delegate asteroid presentation to `AsteroidSync`.
* Delegate pickup presentation to `PickupSync`.
* Pass active render-anchor visual/server positions into non-player entity sync owners.
* Set world entity layer z-index values from generated constants.
* Coordinate interpolation for player rendering, projectiles, asteroids, and pickups.
* Expose remote player visual positions and hues to presentation consumers.
* Expose player-node and remote-player-node lookups to gameplay consumers.
* Expose camera/view-target helpers used by spectate and presentation flows.
* Expose server/visual coordinate conversion helpers through the player-render API.
* Expose `TargetPositionSource` as a read-model seam for targeting flows.
* Reset world presentation state during gameplay-session teardown.

## Does not own

* Server-authoritative simulation.
* Gameplay outcome decisions.
* Collision outcomes.
* Spawn/despawn authority.
* Packet schema source-of-truth files.
* Raw WebSocket transport.
* Packet decoding.
* Realtime world lane packet application and world state readiness.
* Detailed player-render internals.
* Detailed projectile, asteroid, or pickup sync internals.
* Target selection orchestration.
* Input handling.
* HUD behavior.
* Pickup gameplay effects.
* Persistent player data.

## Domain roles

### World-state coordinator

`WorldSync` is the coordinator for applying world lane state to client presentation.

It receives `world_lane_state` after the realtime protocol seam has already applied the inbound world packets and delegates each entity family to its owner. It is the boundary between lane-applied world state and rendered world entities.

### Entity-sync delegation seam

`WorldSync` keeps entity-family behavior out of the coordinator.

The coordinator decides the update order and passes the active render-anchor basis to entity sync owners. The entity sync owners decide how their own nodes are created, updated, cleaned up, and interpolated.

### Runtime interpolation coordinator

`WorldSync.interpolate(delta)` coordinates visual interpolation for the rendered world.

The gameplay runtime owns when interpolation is ticked. `WorldSync` owns which world presentation seams receive that interpolation tick.

### Target-position read-model provider

`WorldSync` configures `TargetPositionSource` with the active player-render API and entity sync owners.

Targeting flows can request target-position data through this read-model seam without reaching directly into world entity maps.

## Protocols and APIs

### Configuration

`WorldSync.configure(...)` receives the gameplay scene owner, local player node, ViewAnchor node, world entity containers, and optional pause-state tracker.

It creates the focused sync owners:

```text
AsteroidSync
ProjectileSync
PickupSync
PlayerRenderApi
TargetPositionSource
```

It also sets layer ordering on the entity containers with generated constants:

```text
Constants.ASTEROID_Z_INDEX
Constants.PICKUP_Z_INDEX
Constants.BULLET_Z_INDEX
```

### Active WorldLaneState application input

The active WorldLaneState input path is:

```text
RealtimeRouter.route_lane_packet(packet)
-> world, lifecycle, and hot lane appliers update world_lane_state
-> WorldPresentationAdapter.apply_world_lane_state(world_sync, world_lane_state, self_id)
-> WorldSync.set_current_self_id(self_id)
-> WorldSync.apply_world_lane_state(world_lane_state)
```

`world_lane_state` currently carries:

```text
world_lane_state.ships
world_lane_state.bullets
world_lane_state.asteroids
world_lane_state.pickups
```

`world_lane_state.bullets` and `world_lane_state.asteroids` may be populated by lifecycle lanes and updated by hot lanes.

### Apply order

The current `WorldSync.apply_world_lane_state` order is:

```text
1. Store current world_lane_state.
2. Remove missing players from world_lane_state.ships.
3. Apply player/render-anchor state from world_lane_state.ships.
4. Remove missing projectiles from world_lane_state.bullets.
5. Apply projectile state from world_lane_state.bullets using the active anchor basis.
6. Remove missing asteroids from world_lane_state.asteroids.
7. Apply asteroid state from world_lane_state.asteroids using the active anchor basis.
8. Remove missing pickups from world_lane_state.pickups.
9. Apply pickup state from world_lane_state.pickups using the active anchor basis.
```

Incremental dirty bullets and asteroids can create render nodes when lifecycle has accepted the entity.

Hot updates alone must not create missing projectiles or asteroids.

```gdscript
player_render_api.visual_position()
player_render_api.server_position()
```

This keeps projectile, asteroid, and pickup rendering aligned with the current ViewAnchor/render-anchor state.

### Legacy compatibility path

`WorldSync.apply_state(...)` still exists in the current implementation as compatibility or internal support for aggregate dictionaries:

```gdscript
func apply_state(
    self_id: String,
    server_players: Dictionary,
    server_bullets: Dictionary,
    server_asteroids: Dictionary,
    server_pickups: Dictionary = {}
) -> void:
```

It should not be treated as the active lane-native world path in service documentation.

### Interpolation

`WorldSync.interpolate(delta)` derives an interpolation weight from generated constants:

```gdscript
var weight := 1.0 - exp(-Constants.PLAYER_INTERPOLATION_SPEED * delta)
```

It then delegates interpolation to:

```text
player_render_api.interpolate(weight, current_self_id)
projectile_sync.interpolate(weight)
asteroid_sync.interpolate(weight)
pickup_sync.interpolate(weight)
```

World sync does not decide when the frame tick occurs. Gameplay runtime calls world sync through `GameplayRuntimeContext.process(delta)`.

### Read-model APIs

`WorldSync` exposes player presentation read models:

```gdscript
get_remote_player_visual_positions()
get_remote_player_hues()
remote_player_nodes()
player_nodes()
```

These methods route through `PlayerRenderApi`.

### View-target APIs

`WorldSync` exposes view-target helpers:

```gdscript
focus_camera_on_player(player_id)
set_view_target_player(player_id)
clear_view_target_player()
```

These route through `PlayerRenderApi`. Detailed ViewAnchor and visual-coordinate behavior belongs in `view-anchor-and-visual-coordinates.md`.

### Coordinate conversion APIs

`WorldSync` exposes coordinate conversion helpers:

```gdscript
visual_position_for_server_position(server_position)
server_position_for_visual_position(visual_position)
```

These route through `PlayerRenderApi`.

Gameplay events use server-to-visual conversion when spawning local presentation effects from server event positions. Input and targeting flows use visual-to-server conversion when translating pointer positions into server-space requests.

### Target source API

`WorldSync.target_source()` returns the configured `TargetPositionSource`.

`TargetPositionSource` is configured with:

```text
PlayerRenderApi
AsteroidSync
ProjectileSync
PickupSync
```

The target source exposes read-model access for:

```text
player positions
asteroid positions
projectile positions
pickup positions
```

Targeting flows own candidate selection and request behavior. World sync only exposes the position data seam.

## Data ownership

`WorldSync` owns transient client presentation coordination state only.

Current coordinator state includes:

* `current_self_id`
* `player_render_api`
* `projectile_sync`
* `asteroid_sync`
* `pickup_sync`
* `target_position_source`
* `world_lane_state`
* `view_anchor`
* `local_player`

`WorldSync` does not persist state.

`WorldSync` does not own authoritative world data. The server owns authoritative simulation, realtime protocol owns world, lifecycle, and hot lane packet application, and `WorldSync` applies the resulting accumulated lane state to presentation and exposes read models for client presentation and targeting consumers.

Entity-specific node maps, target positions, and interpolation state belong inside the relevant sync owners. Player anchor and player meaning state belong behind `PlayerRenderApi`.

## Code map

### Primary implementation

* `client/scripts/world/world_sync.gd`

### Active world lane callers

* `client/scripts/protocol/realtime/world_presentation_adapter.gd`
* `client/scripts/protocol/realtime/world_lane_state.gd`
* `client/scripts/protocol/realtime/realtime_router.gd`
* `client/scripts/gameplay/runtime/gameplay_runtime_context.gd`

### Legacy or compatibility callers


### Delegated sync owners

* `client/scripts/world/player_render/player_render_api.gd`
* `client/scripts/world/player_render/player_meaning_api.gd`
* `client/scripts/world/player_render/view_anchor_sync.gd`
* `client/scripts/world/projectile_sync.gd`
* `client/scripts/world/asteroid_sync.gd`
* `client/scripts/world/pickup_sync.gd`

### Target read model

* `client/scripts/gameplay/targeting/target_position_source.gd`
* `client/scripts/gameplay/targeting/gameplay_targeting_context.gd`
* `client/scripts/gameplay/targeting/gameplay_target_candidate_flow.gd`

### Generated/source data

* `client/scripts/generated/constants/constants.gd`
* `client/scripts/generated/networking/packets/packets.gd`
* `shared/constants/`
* `shared/packets/gameplay.toml`
* `shared/packets/outputs.toml`

### Non-ownership boundaries

* `client/scripts/protocol/realtime/` owns lane packet application and world state readiness before world sync.
* `client/scripts/gameplay/runtime/` owns gameplay runtime composition and interpolation tick entry.
* `client/scripts/gameplay/targeting/` owns target selection and targeting request behavior.
* `client/scripts/gameplay/input/` owns gameplay input handling.
* `client/scripts/shell/gameplay_hud_flow.gd` owns runtime HUD presentation.
* `client/scripts/ui/hud/` owns HUD widget presentation.
* `services/game-server/internal/game/` owns authoritative world simulation and gameplay outcomes.

## Tests

World-sync coordinator behavior is covered or should be covered by tests around:

* `client/tests/unit/test_world_sync.gd`
* `client/tests/unit/test_world_wrap.gd`
* `client/tests/unit/world/player_render/test_player_render_api.gd`
* `client/tests/unit/world/player_render/test_view_anchor_sync.gd`
* `client/tests/unit/test_pickup_sync.gd`
* `client/tests/unit/test_asteroid_sync_state.gd`
* `client/tests/unit/test_projectile_sync_state.gd`
* `client/tests/unit/world/test_projectile_sync.gd`
* `client/tests/unit/gameplay/test_gameplay_target_candidate_flow.gd`
* `client/tests/unit/gameplay/test_gameplay_flow_composer.gd`

Expected verification should confirm:

* `WorldSync.configure` creates and wires all sync owners.
* `WorldSync.apply_world_lane_state` delegates player state before projectile, asteroid, and pickup state.
* `WorldSync.apply_world_lane_state` passes the active anchor visual/server basis into non-player sync owners.
* `WorldSync.interpolate` delegates interpolation to each sync owner.
* `WorldSync.target_source()` returns a configured target-position source.
* `WorldSync.reset()` clears `current_self_id` through `set_current_self_id("")`, which also clears `TargetPositionSource` identity state.
* `WorldSync.reset()` resets `PlayerRenderApi`, `ProjectileSync`, `AsteroidSync`, and `PickupSync`, clears `world_lane_state`, and clears the view target.

## Related docs

* [World Sync](./!INDEX.md)
* [Gameplay Runtime](../gameplay-runtime/!INDEX.md)
* [View Anchor And Visual Coordinates](view-anchor-and-visual-coordinates.md)
* [Entity Sync Owners](entity-sync-owners.md)
* [Gameplay packets](../../../protocol/gameplay-packets.md) - gameplay realtime packet documentation.
* [Toroidal wrap](../../../systems-design/world/toroidal-wrap.md) - toroidal world design documentation.
* [Input and targeting](../input-and-targeting.md) - Client input and targeting documentation.

## Notes

This document intentionally stays at the `WorldSync` coordinator boundary.

Detailed ViewAnchor, render-anchor, toroidal wrap, and coordinate-conversion behavior belongs in [View Anchor And Visual Coordinates](view-anchor-and-visual-coordinates.md).

Detailed projectile, asteroid, and pickup node synchronization belongs in [Entity Sync Owners](entity-sync-owners.md).

`WorldSync.reset()` clears the current self identity by calling `set_current_self_id("")`, resets `PlayerRenderApi`, `ProjectileSync`, `AsteroidSync`, and `PickupSync`, clears `world_lane_state`, and clears the view target. Clearing the self identity through the setter also clears the identity held by `TargetPositionSource`.

`TargetPositionSource.player_positions()` currently reports remote player `server_position` as the same value as `visual_position`; the local player entry uses separate values from `PlayerRenderApi`. Keep targeting documentation aware of that read-model shape.
