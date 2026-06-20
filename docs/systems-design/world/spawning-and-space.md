# Spawning And Space

Parent index: [World](./!INDEX.md)

## Purpose

This document describes the conceptual relationship between spawning and world space in Space Rocks.

It defines how spawned entities enter the authoritative world, how toroidal space affects placement and safety checks, which systems own spawn decisions, and which invariants must remain stable across player, asteroid, projectile, pickup, and debug spawn paths.

## Overview

Spawning is the act of introducing an authoritative runtime entity into the live match world.

World space is server-authoritative and toroidal. Runtime entities store bounded server coordinates, and systems that compare positions use wrapped spatial math so entities near opposite world edges can still be near each other.

The conceptual flow is:

```text
spawn trigger
-> authority-owned spawn decision
-> world-space placement rule
-> runtime entity creation
-> insertion into server entity store
-> state packet projection
-> client presentation
```

The client may request actions that lead to spawning, such as firing a weapon or requesting respawn, but the client does not decide authoritative spawn validity, final position, entity id, collision safety, asteroid variant, or runtime insertion.

The server owns spawn outcomes. The client observes spawned entities through gameplay state packets and renders them relative to the active view anchor.

## Conceptual model

Spawned world entities share several concepts:

```text
spawn trigger
spawn reason
spawn position
spawn authority
runtime identity
runtime entity state
world-space visibility
client presentation
```

`spawn trigger` is the event or condition that begins the spawn flow. Examples include a player joining a game, a respawn packet, elapsed asteroid spawn time, asteroid destruction, weapon fire, a pickup drop roll, radial projectile impact metadata, or a devtools command.

`spawn reason` classifies why the entity entered the world. Current explicit reason values exist for player initial spawn, player respawn, timed asteroid spawn, asteroid fragment spawn, and debug asteroid spawn.

`spawn position` is the authoritative server-space location where the entity is created. It may come from a preferred player spawn origin, a safe respawn search result, an offscreen camera-relative asteroid position, a source asteroid position, a ship weapon muzzle position, a drop-table result, or a debug placement request.

`spawn authority` is the system allowed to decide and apply the spawn. Normal gameplay spawning is server-owned. Devtools may request debug spawning, but debug requests still apply through game-owned seams.

`runtime identity` is the match-local entity id allocated by the server. It is not durable profile data and does not come from the client.

`runtime entity state` is the server-owned state inserted into the active game entity store.

`world-space visibility` determines whether the spawn should appear inside, outside, near, or far from player camera views. This is a server-side gameplay and retention concept, not client rendering visibility.

`client presentation` is the rendered result of state projection. The client creates, updates, interpolates, and removes scene nodes from server state, but it does not own the spawned entity.

## Authority rules

The game server owns authoritative spawning for live gameplay.

The server decides:

* which spawn requests are accepted
* whether a match state permits spawning
* whether a player may respawn
* whether a spawn position is safe
* where timed asteroids appear
* which asteroid variants are selected
* when asteroid fragments are created
* when projectiles are created from weapon fire
* when pickup drop results become pickup entities
* when debug spawn requests are allowed to mutate world state
* which runtime ids are assigned
* when spawned entities enter the runtime entity store
* when spawned entities disappear from state projection

The client may:

* send input that can lead to server-owned spawn decisions
* send viewport configuration that informs server camera-view behavior
* render spawned entities from state packets
* interpolate or visually anchor entities across wrap boundaries
* display events and effects derived from server state
* use devtools UI to request debug spawn actions when devtools are enabled

The client must not:

* create authoritative gameplay entities locally
* pick authoritative spawn positions
* bypass server respawn safety
* spawn authoritative asteroids, fragments, pickups, projectiles, or radial effects
* choose authoritative asteroid variants
* decide whether a spawned entity exists
* preserve spawned entities after the server stops projecting them

Devtools must not create a parallel spawn authority path. Debug spawning should route through game-owned mutation seams and normal state projection.

## World-space rules

The authoritative world is bounded and toroidal.

Server runtime positions are stored inside the shared world bounds. A position that crosses one edge wraps to the opposite edge. Spawn positions that may fall outside bounds, such as offscreen asteroid positions near a camera at the world edge, are normalized into world space before authoritative asteroid creation.

Spatial comparisons should use wrapped math when the relationship can cross a world edge. This includes:

* asteroid aim direction toward camera views
* camera visibility checks
* far-from-camera cleanup
* respawn safety checks
* collision-local placement
* radial effect coverage
* client visual placement relative to the active view anchor

The server stores one authoritative entity position. It does not create duplicate ghost entities at world edges. Cross-edge behavior is handled through shortest wrapped deltas and wrapped distance.

## Camera-view relationship

Camera views are server-side visibility anchors.

They are derived from player state and client viewport configuration, then used by server world systems to answer questions such as:

```text
Is this position onscreen for any player view?
Is this asteroid far from all camera views?
Is this bullet far from all camera views?
Where can a timed asteroid spawn without appearing onscreen?
```

Timed asteroid spawning requires at least one active camera view. If no camera views exist, asteroid spawn timing is reset rather than continuing to accumulate.

Timed asteroid spawn positions are chosen around a target camera view, then rejected if they are visible to any active camera view. This keeps normal asteroid spawns offscreen for all players, not only offscreen for the camera view that triggered the batch.

Camera views also protect entity retention. Asteroids and bullets are removed when they are far from every camera view. That removal is cleanup, not destruction, and must not award score, spawn fragments, or roll pickup drops.

## Player spawn model

Player spawning creates or recreates an active ship for a player session.

Initial player spawn happens when the game adds a player. The server chooses a preferred initial spawn origin, runs it through safe player spawn placement, creates the player session, creates the active ship, inserts the ship into the player entity map, and attaches a camera view.

Respawn happens after death when the player still has lives and the respawn cooldown has reached zero. The respawn request is client-initiated, but the server decides whether it is valid.

Safe player placement checks the candidate position against:

```text
non-pending asteroids
other non-pending active players
the respawning or spawning ship collision shape
configured respawn buffer
wrapped world distance
```

If the preferred position is unsafe, the server searches outward in square rings until it finds a safe candidate.

A player spawn or respawn must create an active ship from session-owned state. It must not create a new player identity, reset durable runtime counters, or let the client choose the final position.

## Asteroid spawn model

Asteroids enter the world through timed spawning, fragment spawning, or debug spawning.

Timed asteroid spawning is a normal active-match simulation effect. The server accumulates spawn time, checks world simulation gates and camera-view availability, spawns one batch per active camera view when the interval elapses, chooses offscreen positions, plans asteroid velocity and size, selects an asteroid variant, allocates an asteroid id, and stores the runtime asteroid.

Timed asteroid velocity is aimed from spawn position toward the target camera position using wrapped direction, with random aim spread and speed. This allows edge-adjacent spawn positions to aim through the shortest toroidal path rather than across the long side of the map.

Fragment spawning happens after asteroid destruction. A destroyed asteroid may spawn two smaller asteroid entities at the source asteroid position. Fragment asteroids receive new runtime ids and independently selected fragment-spawn variants. They do not inherit the source asteroid id or variant.

Debug asteroid spawning exists for development tooling. It is not normal gameplay spawning. It still applies through a game-owned asteroid spawn plan and mutation seam.

Asteroid spawn planning should remain separate from mutation into the runtime entity store.

## Projectile spawn model

Projectiles are spawned from server-resolved weapon fire.

The client sends input. The server checks player state, movement and fire gates, weapon state, cooldown, ammo policy, equipped weapon data, ship position, ship forward direction, and rotation. If weapon fire succeeds, the weapon system returns projectile spawn intent. The game server allocates a projectile id, creates the runtime projectile, inserts it into the projectile entity map, and updates weapon slot state.

Projectile spawning must use authoritative ship and weapon facts. The client may display weapon input and local presentation, but it must not create authoritative bullets or decide hit results.

Projectile positions are stored in toroidal server space and later advanced through server motion. Projectile state is projected to clients through gameplay state packets.

## Pickup spawn model

Pickups enter the world through server-owned pickup spawning paths.

Current normal pickup drops can occur after asteroid destruction. The server evaluates drop-table results from asteroid source facts, respects max-active pickup limits, creates pickup entities from server pickup definitions, inserts them into the pickup entity store, and records pickup-dropped events for presentation.

The asteroid supplies source facts such as id, size, and position. Drop-table evaluation decides whether a result exists. Pickup entity creation decides whether a concrete pickup enters the world.

Pickup spawning must stay separate from pickup collection and pickup effects. Collection and effects are later server-owned interactions with the already-spawned pickup entity.

## Radial effect spawn model

Radial effects are spawned from projectile impact metadata.

A projectile impact may carry metadata that asks the game-owned adapter to spawn a radial effect. The radial effect system owns timing, zone coverage, and hit-intent generation. It does not own projectile impact validity, entity store ownership, damage application, score mutation, asteroid fragment spawning, pickup drops, or client presentation.

Radial effect spawning should preserve the same world-space authority rule as other spawn paths: the server owns the effect origin, coverage, timing, and hit results; the client observes resulting events or presentation state.

## Visibility and despawn relationship

Spawning and despawn are related but opposite world-retention decisions.

Spawning decides when an entity enters runtime state. Despawn and cleanup decide when an entity leaves runtime state.

Spawned entities can leave the world through different mechanisms:

```text
pending despawn after collision/destruction
lifetime expiry
far-from-camera cleanup
pickup expiration
collection
match or room cleanup
```

These exits have different meanings.

Destruction-driven pending despawn can produce gameplay consequences such as score, fragments, drops, or presentation events. Far-from-camera cleanup is not destruction. It only removes entities that no player view needs to retain.

A systems-design invariant follows from this split: entity removal must not be treated as equivalent to entity destruction.

## Client presentation model

The client renders spawned entities from authoritative state.

Server positions are bounded. Client visual positions are continuous presentation coordinates relative to the active ViewAnchor. The client uses shortest wrapped deltas to keep entities visually close to the current anchor when server coordinates cross edges.

This means:

```text
server position
= authoritative bounded world coordinate

visual position
= client presentation coordinate derived from server position and active anchor
```

The client can instantiate asteroid, projectile, pickup, player, and effect nodes from server state or server events. It can interpolate and place them visually. It cannot make them authoritative.

When the server stops sending an entity in state packets, the client should remove or retire the corresponding presentation node according to that entity family's presentation rules.

## Participating systems

Spawning and space involves these systems:

* [World Authority](world-authority.md) defines the broader authority split for world state.
* [Toroidal Wrap](toroidal-wrap.md) defines conceptual wrap behavior for movement and spatial relationships.
* [Asteroids](../entities/asteroids.md) defines the asteroid entity model that timed and fragment spawning create.
* [Projectiles](../entities/projectiles.md) defines projectile entities created by weapon fire.
* [Ships](../entities/ships.md) defines active player ships created by initial spawn and respawn.
* [Pickup Entities](../entities/pickup-entities.md) defines pickup entities created by drop results and debug tooling.
* [Asteroid Spawning And Variants](../../services/game-server/simulation/world/asteroid-spawning-and-variants.md) documents server implementation for asteroid spawn planning, variant selection, and runtime insertion.
* [Toroidal Space And Motion](../../services/game-server/simulation/world/toroidal-space-and-motion.md) documents server implementation for wrapped space and movement.
* [Visibility And Despawn](../../services/game-server/simulation/world/visibility-and-despawn.md) documents server implementation for camera-view visibility, offscreen spawn rejection, and cleanup.
* [Player Respawn](../../services/game-server/simulation/players/player-respawn.md) documents server implementation for respawn eligibility and safe respawn placement.
* [Player Camera View State](../../services/game-server/simulation/players/player-camera-view-state.md) documents server camera-view ownership.
* [Weapons And Projectile Fire](../../services/game-server/simulation/combat/weapons-and-projectile-fire.md) documents projectile spawn intent from weapon fire.
* [Pickup Drop Integration](../../services/game-server/simulation/pickups/pickup-drop-integration.md) documents server pickup drop spawning from asteroid destruction.
* [Radial Effects](../../services/game-server/simulation/combat/radial-effects.md) documents radial effect runtime integration.
* [Runtime Entity Store](../../services/game-server/simulation/runtime/runtime-entity-store.md) documents authoritative runtime entity storage.
* [Simulation Loop And Phase Order](../../services/game-server/simulation/runtime/simulation-loop-and-phase-order.md) documents when spawn and cleanup phases run.
* [View Anchor And Visual Coordinates](../../services/client/world-sync/view-anchor-and-visual-coordinates.md) documents client visual-coordinate conversion.
* [Entity Sync Owners](../../services/client/world-sync/entity-sync-owners.md) documents client entity node ownership.
* [Gameplay Packets](../../protocol/gameplay-packets.md) documents the packet surface that carries spawned entity state.
* [Asteroid Variant Contract](../../protocol/asteroid-variant-contract.md) documents the asteroid variant index contract.
* [Constants Pipeline](../../data/constants.md) documents world-size and spawn-tuning constants.
* [Asteroid Variants Data](../../data/asteroid-variants-data.md) documents asteroid variant spawn weights and catalog data.
* [Collision Shape Data](../../data/collision-shape-data.md) documents authoritative collision-shape source data.

## Related docs

* [Player Camera View State](../../services/game-server/simulation/players/player-camera-view-state.md)
* [Runtime Entity Store](../../services/game-server/simulation/runtime/runtime-entity-store.md)
* [Simulation Loop And Phase Order](../../services/game-server/simulation/runtime/simulation-loop-and-phase-order.md)
* [Entity Sync Owners](../../services/client/world-sync/entity-sync-owners.md)

## Invariants

Spawning and space must preserve these invariants:

* Spawned gameplay entities are server-authoritative.
* Runtime entity ids are allocated by the server.
* Client-side node creation is presentation, not authority.
* Spawn planning should stay separate from runtime entity-store mutation where a planning seam exists.
* Spawn positions must be expressed in authoritative server world space before runtime insertion.
* Positions stored on moving world entities should remain bounded by toroidal world bounds.
* Wrapped spatial relationships must use shortest toroidal deltas or wrapped distance.
* Cross-edge behavior must not require duplicate ghost entities in authoritative storage.
* Timed asteroid spawning requires active camera views.
* Timed asteroid spawn positions must be offscreen for all active camera views.
* Timed asteroid spawning must not continue after match over.
* Asteroid variants must be selected by spawn-source-specific catalog helpers.
* Fragment asteroids are new entities with new ids.
* Fragment variants are selected independently from the source asteroid.
* Player respawn must be server-validated against lives, cooldown, active-ship state, and safe placement.
* Respawn safety must account for hazards across wrapped world edges.
* Projectile spawning must come from server-resolved weapon fire, not client-created bullets.
* Pickup spawning from drops must follow server drop-table results and pickup definition rules.
* Radial effect spawning must not bypass projectile impact, radial timing, or damage-application ownership boundaries.
* Far-from-camera cleanup is not entity destruction.
* Entity removal must not automatically imply scoring, drops, fragments, or destruction events.
* Devtools spawn requests must route through game-owned mutation seams.
* Shared world bounds must stay aligned between server authority and client visual wrapping.
* Packet schemas carry spawned state; they do not own spawn authority.

## Active issues

Current world-space limits are tracked in [Current System Limits](../../limits/current-system-limits.md#architecture--networking).

The relevant current issue is that vertical despawn behavior is limited by the relationship between world height, visible viewport height, and despawn margin. This is an implementation constraint, not a spawning-and-space invariant.

## Related docs

* [World](./!INDEX.md)
* [Systems Design](../!INDEX.md)
* [World Authority](world-authority.md)
* [Toroidal Wrap](toroidal-wrap.md)
* [Asteroids](../entities/asteroids.md)
* [Projectiles](../entities/projectiles.md)
* [Ships](../entities/ships.md)
* [Pickup Entities](../entities/pickup-entities.md)
* [Asteroid Spawning And Variants](../../services/game-server/simulation/world/asteroid-spawning-and-variants.md)
* [Toroidal Space And Motion](../../services/game-server/simulation/world/toroidal-space-and-motion.md)
* [Visibility And Despawn](../../services/game-server/simulation/world/visibility-and-despawn.md)
* [Player Respawn](../../services/game-server/simulation/players/player-respawn.md)
* [Weapons And Projectile Fire](../../services/game-server/simulation/combat/weapons-and-projectile-fire.md)
* [Pickup Drop Integration](../../services/game-server/simulation/pickups/pickup-drop-integration.md)
* [Radial Effects](../../services/game-server/simulation/combat/radial-effects.md)
* [View Anchor And Visual Coordinates](../../services/client/world-sync/view-anchor-and-visual-coordinates.md)
* [Gameplay Packets](../../protocol/gameplay-packets.md)
* [Asteroid Variant Contract](../../protocol/asteroid-variant-contract.md)
* [Constants Pipeline](../../data/constants.md)
* [Asteroid Variants Data](../../data/asteroid-variants-data.md)
* [Collision Shape Data](../../data/collision-shape-data.md)
* [Current System Limits](../../limits/current-system-limits.md#architecture--networking)

## Notes

The key split is that the server owns bounded authoritative coordinates and spawn decisions, while the client owns continuous visual presentation.

The current implementation has several useful spawn seams already in place, especially asteroid spawn plans, player spawn plans, weapon projectile spawn intent, drop-table pickup results, and game-owned devtools apply hooks. Future spawn work should extend those seams rather than moving spawn authority into packet readers, client presentation, scoring policy, pickup effects, or devtools-only logic.

Spawning should be treated as a world authority concern first and an entity-family concern second. Each entity family can define its own trigger and runtime state, but the authority rule stays the same: the server decides what enters the live world.
