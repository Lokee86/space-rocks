# World Sync Coordinator

Parent index: [World Sync](./!INDEX.md)

## Purpose

This document describes the client `WorldSync` coordinator.

It explains how the client applies normalized server world state to world presentation seams, how `WorldSync` delegates entity-family synchronization, how interpolation is coordinated, and how targeting and presentation code access world-sync read models.

## Overview

`WorldSync` is the client-side coordinator for rendering server-authoritative world state.

It does not parse raw packets and does not decide gameplay outcomes. Gameplay runtime passes already-normalized world-state dictionaries into `WorldSync.apply_state`, and `WorldSync` delegates the actual player, projectile, asteroid, and pickup presentation work to focused sync owners.

The current runtime path is:

```text
GameplayStateApplyFlow
-> GameplayWorldStateApplyFlow
-> WorldSync.apply_state
-> PlayerRenderApi
-> ProjectileSync
-> AsteroidSync
-> PickupSync
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
* Apply normalized server world state to the client world.
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
* Gameplay-state packet normalization.
* Detailed player-render internals.
* Detailed projectile, asteroid, or pickup sync internals.
* Target selection orchestration.
* Input handling.
* HUD behavior.
* Pickup gameplay effects.
* Persistent player data.

## Domain roles

### World-state coordinator

`WorldSync` is the coordinator for applying server world state to client presentation.

It receives the world-state subset selected by gameplay runtime and delegates each entity family to its owner. It is the boundary between normalized gameplay state and rendered world entities.

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

### World-state application input

`WorldSync.apply_state(...)` receives normalized state from gameplay runtime:

```gdscript
func apply_state(
    self_id: String,
    server_players: Dictionary,
    server_bullets: Dictionary,
    server_asteroids: Dictionary,
    server_pickups: Dictionary = {}
) -> void:
```

These values are passed by `GameplayWorldStateApplyFlow` from the normalized gameplay state dictionary.

### Apply order

The current `WorldSync.apply_state` order is:

```text
1. Store current self_id.
2. Update TargetPositionSource current self id.
3. Remove missing players.
4. Remove missing projectiles.
5. Remove missing asteroids.
6. Remove missing pickups.
7. Apply player/render-anchor state.
8. Apply projectile state using the active anchor basis.
9. Apply asteroid state using the active anchor basis.
10. Apply pickup state using the active anchor basis.
```

The non-player entity sync owners receive the active visual and server anchor basis from `PlayerRenderApi`:

```gdscript
player_render_api.visual_position()
player_render_api.server_position()
```

This keeps projectile, asteroid, and pickup rendering aligned with the current ViewAnchor/render-anchor state.

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
* `view_anchor`
* `local_player`

`WorldSync` does not persist state.

`WorldSync` does not own authoritative world data. It receives server state, applies presentation changes, and exposes read models for client presentation and targeting consumers.

Entity-specific node maps, target positions, and interpolation state belong inside the relevant sync owners. Player anchor and player meaning state belong behind `PlayerRenderApi`.

## Code map

### Primary implementation

* `client/scripts/world/world_sync.gd`

### Runtime callers

* `client/scripts/gameplay/runtime/gameplay_world_state_apply_flow.gd`
* `client/scripts/gameplay/runtime/gameplay_runtime_context.gd`
* `client/scripts/gameplay/state/gameplay_state_apply_flow.gd`
* `client/scripts/gameplay/state/gameplay_state_packet_reader.gd`

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

* `client/scripts/gameplay/runtime/` owns gameplay runtime composition and state fanout into world sync.
* `client/scripts/gameplay/state/` owns gameplay packet normalization and state application ordering before world sync.
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
* `client/tests/unit/gameplay/test_gameplay_target_candidate_flow.gd`
* `client/tests/unit/gameplay/test_gameplay_flow_composer.gd`

Expected verification should confirm:

* `WorldSync.configure` creates and wires all sync owners.
* `WorldSync.apply_state` removes missing entities before applying new state.
* `WorldSync.apply_state` delegates player state before projectile, asteroid, and pickup state.
* `WorldSync.apply_state` passes the active anchor visual/server basis into non-player sync owners.
* `WorldSync.interpolate` delegates interpolation to each sync owner.
* `WorldSync.target_source()` returns a configured target-position source.
* Reset behavior clears intended presentation state.

## Related docs

* [World Sync](./!INDEX.md)
* [Gameplay Runtime](../gameplay-runtime/!INDEX.md)
* [View Anchor And Visual Coordinates](view-anchor-and-visual-coordinates.md)
* [Entity Sync Owners](entity-sync-owners.md)
* [Gameplay packets](../../../protocol/stubs/gameplay-packets.md) - Stub: gameplay realtime packet documentation.
* [Toroidal wrap](../../../systems-design/world/stubs/toroidal-wrap.md) - Stub: toroidal world design documentation.
* [Input and targeting](../input-and-targeting.md) - Client input and targeting documentation.

## Notes

This document intentionally stays at the `WorldSync` coordinator boundary.

Detailed ViewAnchor, render-anchor, toroidal wrap, and coordinate-conversion behavior belongs in [View Anchor And Visual Coordinates](view-anchor-and-visual-coordinates.md).

Detailed projectile, asteroid, and pickup node synchronization belongs in [Entity Sync Owners](entity-sync-owners.md).

`WorldSync.reset()` currently resets `PlayerRenderApi`, `AsteroidSync`, and `PickupSync`, then clears the view target. It does not explicitly call `ProjectileSync.reset()` in the current implementation. That should be treated as current implementation behavior unless changed.

`TargetPositionSource.player_positions()` currently reports remote player `server_position` as the same value as `visual_position`; the local player entry uses separate values from `PlayerRenderApi`. Keep targeting documentation aware of that read-model shape.
