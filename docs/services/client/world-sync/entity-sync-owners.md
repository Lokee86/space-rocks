# Entity Sync Owners

Parent index: [World Sync](./!INDEX.md)

## Purpose

This document describes the client world-sync entity owners that render server-authoritative projectile, asteroid, and pickup state as Godot scene nodes.

It covers scene-node creation, missing-node cleanup, target visual positions, interpolation, wrap-aware visual continuity, presentation constants, and the boundary between client rendering and server gameplay authority.

## Overview

Entity sync owners are focused rendering seams under client world sync.

`WorldSync` receives normalized server world state from gameplay runtime, then delegates projectile, asteroid, and pickup dictionaries to dedicated sync owners. Each owner is responsible for making its entity family visible in the Godot scene while treating server state as authoritative.

The entity sync owners currently covered by this document are:

```text
ProjectileSync
AsteroidSync
PickupSync
```

They share the same general pattern:

```text
1. Remove local nodes that are missing from the latest server state.
2. Create local nodes for newly seen server entities.
3. Update target visual positions from server coordinates.
4. Apply presentation-specific state.
5. Interpolate rendered nodes toward target state during runtime processing.
```

They do not decide whether an entity should exist. They render the server state they are given.

## Code root

* `client/`

## Responsibilities

* Render server-authoritative projectile, asteroid, and pickup state as client scene nodes.
* Create scene nodes for newly seen entities.
* Remove local scene nodes that are absent from the latest server state.
* Track target visual positions for interpolation.
* Interpolate rendered nodes during world-sync runtime processing.
* Preserve visual continuity when server coordinates wrap around world bounds.
* Select projectile scenes through projectile presentation metadata.
* Apply asteroid scale and variant presentation from server state.
* Select pickup presentation scenes by pickup class.
* Apply pickup type presentation to pickup nodes.
* Forward pickup lifespan state to pickup presentation nodes when supported.
* Apply z-index and presentation-layer constants to rendered entities.
* Keep entity-family rendering concerns split by projectile, asteroid, and pickup ownership.

## Does not own

* Server spawning authority.
* Server despawn authority.
* Collision detection or collision outcomes.
* Projectile hit authority.
* Asteroid split/destruction authority.
* Pickup collection authority.
* Pickup gameplay effects.
* Score, lives, respawn, or match-result decisions.
* Packet schema source-of-truth files.
* Packet decoding or gameplay-state normalization.
* Player and ViewAnchor synchronization.
* Target selection orchestration.
* HUD, menu, input, or devtools ownership.

## Domain roles

### Projectile presentation owner

`ProjectileSync` owns client-side projectile node presentation.

It creates projectile nodes, chooses projectile scenes through `ProjectileSceneResolver`, applies target positions and rotations, plays projectile firing presentation, removes missing projectile nodes, and interpolates projectile nodes toward their latest server-derived target state.

### Asteroid presentation owner

`AsteroidSync` owns client-side asteroid node presentation.

It creates asteroid nodes, applies server-provided radius/scale state, applies asteroid variant presentation, tracks target visual positions, removes missing asteroid nodes, and interpolates asteroid nodes across updates.

Asteroid sync also tracks previous server and visual positions so existing asteroids can move continuously across toroidal wrap boundaries instead of snapping when bounded server coordinates wrap.

### Pickup presentation owner

`PickupSync` owns client-side pickup node presentation.

It creates pickup nodes, chooses pickup scene families by pickup class through `PickupPresentationCatalog`, applies pickup type presentation, forwards lifespan state where supported by the node, removes missing pickup nodes, and interpolates pickup nodes across updates.

Pickup sync renders pickup presence and presentation. Pickup gameplay effects remain server-owned.

## Protocols and APIs

### WorldSync delegation

Entity sync owners are called by `WorldSync.apply_state`.

The normalized state has already been read from gameplay packets before it reaches world sync. Entity sync owners consume dictionaries that represent the current server-visible state for each entity family.

Current delegation shape:

```text
WorldSync.apply_state(
    self_id,
    server_players,
    server_bullets,
    server_asteroids,
    server_pickups
)
```

`WorldSync` handles player/render-anchor state first, then delegates projectile, asteroid, and pickup state with the active anchor basis.

### Anchor-relative visual positioning

Projectile, asteroid, and pickup sync owners receive the active anchor state from the player-render API:

```text
anchor visual position
anchor server position
```

The common visual-position calculation is:

```text
entity visual position =
  anchor visual position +
  shortest wrapped delta from anchor server position to entity server position
```

This keeps entity rendering stable when authoritative server positions wrap at world edges.

### Missing-node cleanup

Each entity sync owner removes rendered nodes when the corresponding server entity id is absent from the latest server dictionary.

This keeps local presentation aligned with server state and prevents stale projectiles, destroyed asteroids, or collected/expired pickups from remaining visible.

### Interpolation

World-sync interpolation is triggered by runtime processing:

```text
GameplayRuntimeContext.process(delta)
-> WorldSync.interpolate(delta)
```

`WorldSync.interpolate(delta)` delegates to each entity sync owner after calculating the interpolation weight.

Entity sync owners move rendered nodes toward their latest target visual positions. They do not extrapolate authoritative gameplay state.

## Data ownership

Entity sync owners own transient client presentation state only.

They may own:

* rendered node dictionaries keyed by server entity id
* target visual positions
* previous server positions needed for wrap continuity
* previous visual positions needed for wrap continuity
* initialized-entity flags
* presentation metadata already applied to nodes
* pickup type/class presentation state
* projectile scene selection state
* asteroid variant presentation state

They do not persist state.

They do not own authoritative gameplay state.

They do not own durable player or profile data.

## Projectile sync

`ProjectileSync` handles projectile presentation.

Current responsibilities include:

* creating projectile scene nodes
* choosing projectile scenes through `ProjectileSceneResolver`
* setting projectile initial positions
* applying projectile target positions
* applying projectile rotations
* removing projectiles missing from server state
* interpolating projectile nodes
* playing projectile firing presentation when a new projectile appears
* exposing projectile positions to target-position read models

Projectile sync treats server projectile dictionaries as the source of truth.
Bullet pulse effects are scene-local presentation owned by `client/scripts/entities/bullet.gd`, not world-sync authority.

Projectile scene selection belongs to:

```text
client/scripts/world/projectiles/projectile_scene_resolver.gd
```

Projectile packet-facing field access belongs to:

```text
client/scripts/world/projectile_sync_state.gd
```

## Asteroid sync

`AsteroidSync` handles asteroid presentation.

Current responsibilities include:

* creating asteroid scene nodes
* applying asteroid radius/scale presentation
* applying asteroid variant presentation
* tracking target visual positions
* tracking previous server positions
* tracking previous visual positions
* removing asteroids missing from server state
* interpolating asteroid nodes
* exposing asteroid positions to target-position read models

Asteroid sync preserves visual continuity for existing asteroids by calculating movement from the asteroid’s previous server position to its current server position using wrapped shortest-delta logic.
Asteroid variant-specific client presentation should live in the client-side asteroid variant presentation doc when that doc exists; this section only covers the sync-owner boundary.

New asteroids are positioned from the current active anchor basis. First-seen asteroids may appear offscreen when the server intentionally spawns them outside the immediate view area.

Asteroid packet-facing field access belongs to:

```text
client/scripts/world/asteroid_sync_state.gd
```

## Pickup sync

`PickupSync` handles pickup presentation.

Current responsibilities include:

* creating pickup scene nodes
* selecting pickup scene families by pickup class
* applying pickup type presentation
* forwarding lifespan state to pickup nodes when supported
* tracking target visual positions
* tracking previous server positions
* tracking previous visual positions
* removing pickups missing from server state
* interpolating pickup nodes
* playing pickup spawn presentation when applicable
* exposing pickup positions to target-position read models

Pickup scene-family selection belongs to:

```text
client/scripts/world/pickups/pickup_presentation_catalog.gd
```

Pickup packet-facing field access belongs to:

```text
client/scripts/world/pickup_sync_state.gd
```

Pickup sync does not decide what a pickup does when collected. It only presents pickup state supplied by the server.

## Code map

### Entity sync owners

* `client/scripts/world/projectile_sync.gd`
* `client/scripts/world/asteroid_sync.gd`
* `client/scripts/world/pickup_sync.gd`

### Entity packet-facing state readers

* `client/scripts/world/projectile_sync_state.gd`
* `client/scripts/world/asteroid_sync_state.gd`
* `client/scripts/world/pickup_sync_state.gd`

### Scene and presentation selection

* `client/scripts/entities/bullet.gd`
* `client/scripts/world/projectiles/projectile_scene_resolver.gd`
* `client/scripts/world/pickups/pickup_presentation_catalog.gd`

### Coordinator and world math

* `client/scripts/world/world_sync.gd`
* `client/scripts/world/world_wrap.gd`

### Runtime callers

* `client/scripts/gameplay/runtime/gameplay_world_state_apply_flow.gd`
* `client/scripts/gameplay/runtime/gameplay_runtime_context.gd`

### Target read-model consumers

* `client/scripts/gameplay/targeting/target_position_source.gd`

### Scenes

* `client/scenes/bullet.tscn`
* `client/scenes/projectiles/torpedo.tscn`
* `client/scenes/asteroid.tscn`
* `client/scenes/pickups/powerup_pickup.tscn`
* `client/scenes/pickups/weapon_pickup.tscn`

### Generated data

* `client/scripts/generated/constants/constants.gd`
* `client/scripts/generated/networking/packets/packets.gd`

### Source-of-truth boundaries

* `shared/packets/gameplay.toml`
* `shared/packets/outputs.toml`
* `shared/constants/`

## Tests

Relevant client tests include:

* `client/tests/unit/test_asteroid_sync_state.gd`
* `client/tests/unit/test_projectile_sync_state.gd`
* `client/tests/unit/test_pickup_sync.gd`
* `client/tests/unit/test_pickup_sync_state.gd`
* `client/tests/unit/world/projectiles/test_projectile_scene_resolver.gd`
* `client/tests/unit/world/test_pickup_presentation_catalog.gd`
* `client/tests/unit/test_world_sync.gd`
* `client/tests/unit/test_world_wrap.gd`
* `client/tests/unit/gameplay/test_gameplay_target_candidate_flow.gd`

Use the normal client GUT verification flow when changing entity sync behavior.

## Related docs

* [World Sync](./!INDEX.md)
* [Gameplay Runtime](../gameplay-runtime/!INDEX.md)
* [World sync coordinator](world-sync-coordinator.md)
* [View anchor and visual coordinates](view-anchor-and-visual-coordinates.md)
* [Pickup Presentation](pickup-presentation.md) - Client pickup presentation documentation.
* [Gameplay packets](../../../protocol/gameplay-packets.md) - Gameplay realtime packet documentation.

## Notes

Entity sync owners render authoritative state; they do not create gameplay truth.

Projectile, asteroid, and pickup sync should stay separate. Their presentation rules overlap, but their scene selection, packet-facing state, and visual details are different enough that collapsing them would make world sync harder to maintain.

Target-position read models may consume entity sync data, but target selection and target requests belong to gameplay targeting flows.

Pickup sync owns pickup rendering. Pickup effect rules, collection validity, rewards, and despawn authority belong to server gameplay systems.
