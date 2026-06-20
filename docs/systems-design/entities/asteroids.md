# Asteroids

Parent index: [Entities](./!INDEX.md)

## Purpose

This document describes the asteroid entity model for Space Rocks.

It defines what an asteroid is conceptually, which systems own asteroid authority, which invariants must remain stable, and how asteroid identity, movement, variants, damage, destruction, fragments, drops, and presentation fit together.

## Overview

Asteroids are transient, match-local world entities.

They are server-authoritative hazards and combat targets. They move through toroidal world space, collide with projectiles and player ships, take damage, split into smaller fragments when destroyed by projectile/radial damage, and may produce pickup drops.

The core model is:

```text
server-owned spawn decision
-> runtime asteroid entity
-> server-owned movement and wrap
-> server-owned collision and damage consequences
-> state packet projection
-> client-owned presentation
```

An asteroid is not durable player data, not a client-created object, and not a protocol authority by itself. The server owns its authoritative runtime state. The client owns only the rendered node, texture, interpolation, and presentation behavior derived from server state.

## Conceptual model

A runtime asteroid is defined by:

```text
identity
position
velocity
size
variant
health
damage modifiers
collision damage
pending-despawn state
```

These concepts are separate.

`identity` is the match-local runtime id, currently shaped like `asteroid-1`, `asteroid-2`, and so on. It identifies one live asteroid in the current game instance. It is not durable and does not survive the match.

`position` and `velocity` describe the asteroid in authoritative server world space.

`size` controls current visual scale, collision scale, score value, and fragment size. Destroyed asteroids produce fragments with `source size - 1`. A size that would become `0` or lower produces no fragments.

`variant` is the server-selected runtime variant index. It controls presentation catalog lookup and is intended to connect to asteroid data such as texture, collision-shape key, stats profile, drop table, and spawn weights. Current runtime behavior uses only part of that data.

`health` makes asteroid destruction damage-based rather than a hardcoded collision flag. Current asteroid health is low, but the entity model already supports nonfatal projectile damage.

`collision damage` is the damage intent an asteroid applies to a player ship on ship/asteroid collision.

`pending-despawn state` is a short delayed-removal window after destruction. It preserves the asteroid in server state briefly enough for clients to observe final state and related events before the asteroid disappears.

## Authority rules

The game server owns asteroid authority.

The server decides:

* when timed asteroids spawn
* where asteroids spawn
* which asteroid ids are allocated
* which size each asteroid has
* which variant index each asteroid receives
* how asteroids move and wrap
* whether an asteroid collides with a projectile or player
* how much damage an asteroid receives
* whether an asteroid is destroyed
* whether destruction awards score
* whether destruction spawns fragments
* whether destruction rolls pickup drops
* when asteroids are removed from runtime storage

The client does not create authoritative asteroids, choose authoritative variants, resolve asteroid damage, award asteroid score, spawn asteroid fragments, roll asteroid drops, or decide asteroid removal.

The client may:

* instantiate asteroid scene nodes for server-provided asteroid ids
* interpolate asteroid positions
* apply server-provided scale
* resolve the server-provided variant index into texture and scene presentation
* remove asteroid nodes when the server stops sending them

Devtools may request debug asteroid spawns, but debug spawning still routes through game-owned asteroid spawn planning and application seams. Devtools must not create a parallel asteroid authority path.

## Lifecycle

The normal asteroid lifecycle is:

```text
spawn plan
-> runtime asteroid creation
-> live movement
-> collision/damage participation
-> optional destruction consequences
-> pending despawn
-> removal from entity store
-> disappearance from state packets
```

Timed asteroid spawning happens during active match simulation. The server accumulates spawn time, checks active camera views, selects offscreen spawn positions, builds asteroid spawn plans, then applies those plans into the runtime entity store.

Fragment spawning happens after asteroid destruction. A destroyed asteroid creates two fragment spawn plans when its fragment size is greater than `0`. Fragment asteroids are new asteroid entities with their own ids, velocities, sizes, and independently selected fragment-spawn variants.

Debug asteroid spawning exists for development tooling. It builds a debug asteroid spawn plan and applies it through the same game-owned mutation seam as normal asteroid spawning.

Removal can happen in two ways:

```text
destruction -> pending despawn delay -> removal
far from all camera views -> immediate removal from runtime store
```

Far-from-camera removal is world-retention cleanup, not asteroid destruction. It should not award score, spawn fragments, or roll drops.

## Movement and spatial behavior

Asteroids move through server-owned toroidal world space.

Each active asteroid advances from its velocity during the asteroid step, then wraps through the world bounds. This lets asteroids cross edges naturally without becoming separate world instances.

Wrapped-space behavior matters for:

* movement across world edges
* offscreen spawn placement near camera views
* far-from-camera cleanup
* projectile/asteroid collision
* ship/asteroid collision
* client visual placement relative to the active view anchor

Collision checks use wrapped-local placement. The asteroid collision body is temporarily placed near the projectile or player being checked using the shortest wrapped delta. This avoids duplicating stored asteroid entities as ghost bodies while still allowing cross-edge collisions.

## Size model

Asteroid size is a gameplay concept, not only a visual concept.

Current size affects:

* state-packet scale
* server collision shape scaling
* client node scale
* score value from destruction
* whether fragments can spawn
* fragment size after destruction
* drop-table source data

The current scale projection is derived from asteroid size and the generated asteroid size scale constant. The client consumes the projected scale; it should not independently infer authoritative gameplay size behavior from scene settings.

Fragment size is always:

```text
source asteroid size - 1
```

If that result is `0` or lower, no fragments spawn.

## Variant model

Asteroid variants are server-assigned runtime indexes backed by shared asteroid variant data.

The semantic source data lives in [Asteroid Variants Data](../../data/asteroid-variants-data.md). The protocol boundary that carries the runtime index is documented in [Asteroid Variant Contract](../../protocol/asteroid-variant-contract.md).

Conceptually, variant data describes:

```text
stable variant id
runtime index
client texture
collision-shape key
stats-profile key
drop-table key
timed spawn weight
fragment spawn weight
debug spawn weight
```

The runtime asteroid stores the variant index, not the stable variant id string.

The server chooses variants through spawn-source-specific helpers:

```text
timed asteroid spawn -> timed spawn weight
fragment spawn       -> fragment spawn weight
debug spawn          -> debug spawn weight
```

The client receives the selected variant index through asteroid state and uses it for presentation lookup.

Variant lookup wrapping exists as a safety behavior in catalog helpers, but valid server-emitted variants should still come from the agreed catalog index set. Wrapping is not version negotiation and should not hide catalog drift.

## Damage and destruction

Asteroids are damageable server entities.

Projectile/asteroid and radial asteroid damage flow through the server damage model. The damage resolver calculates the result; game-owned combat code applies the result to asteroid health and handles downstream consequences.

A projectile/asteroid hit can be nonfatal. In that case, the asteroid health changes, damage presentation may be recorded, and the projectile can despawn, but the asteroid does not award score, split, or roll drops.

Asteroid destruction consequences require a destroyed damage result.

The current projectile destruction consequence path is:

```text
destroyed asteroid
-> evaluate score policy
-> apply score award
-> mark asteroid pending despawn
-> spawn fragments
-> maybe roll pickup drop
```

Radial asteroid damage should reuse the same damage and destruction consequence boundaries. Radial effects may decide that an asteroid was hit, but they do not own asteroid health mutation, scoring, fragment spawning, or pickup drops.

## Ship collision behavior

Ship/asteroid collision damages the player ship. It does not currently damage or destroy the asteroid.

The asteroid is the damage source:

```text
source entity type = asteroid
damage cause       = collision
damage type        = kinetic
amount             = asteroid collision damage
```

The player ship is the damage target. If the result is fatal player damage, player death, lives, respawn cooldown, camera-view preservation, and ship-death presentation events are handled by player and combat ownership seams.

Players that are paused, suspended, temporarily invulnerable, or debug-invincible do not take asteroid collision damage.

## Despawn and projection

Destroyed asteroids are marked pending despawn instead of being removed immediately.

Pending despawn means:

* the asteroid no longer moves
* its despawn delay counts down
* collision checks skip it
* it remains in state packets until removal
* it is removed when the delay reaches zero

This short delay lets clients receive final asteroid state and related presentation events.

Once an asteroid is removed from `game.entities.Asteroids`, it disappears from future state packets. The client removes the corresponding asteroid node when it is no longer present in server asteroid state.

## Drops and scoring

Asteroid destruction can produce score and pickup drops, but those outcomes are not owned by the asteroid entity itself.

Score is evaluated by the scoring policy from asteroid destruction facts such as player id, target id, and asteroid size. Game-owned score application decides whether the player can actually receive the award.

Pickup drops are evaluated from drop-table data after destruction. The current asteroid drop flow uses the basic asteroid drop table path. Drop-table data and evaluation are documented in [Drop Tables](../../data/drop-tables.md), while pickup entity behavior is documented separately.

The asteroid entity supplies source facts such as id, size, and position. It does not own drop-table rules, pickup spawning policy, pickup collection, pickup effects, or pickup presentation.

## Presentation

Asteroid presentation is client-owned and state-driven.

The server projects asteroid state through gameplay state packets. Current asteroid state includes:

```text
id
x
y
size
health
scale
variant
```

The client world-sync layer creates or updates asteroid nodes from that state. Asteroid presentation uses:

* server position
* server-projected scale
* server-provided variant index
* client-side interpolation
* client asteroid scene texture and polygon presentation

The client scene's collision polygon is presentation-local. Authoritative collision uses server-loaded collision shape data and server collision checks.

The client should not infer asteroid destruction, score, drops, or hit validity from rendering state. It observes those outcomes from server state and server presentation events.

## Participating systems

Asteroid entity behavior participates with these systems:

* [Asteroid Spawning And Variants](../../services/game-server/simulation/world/asteroid-spawning-and-variants.md) owns server implementation for timed, fragment, and debug asteroid spawning.
* [Visibility And Despawn](../../services/game-server/simulation/world/visibility-and-despawn.md) owns server world-retention cleanup and pending-despawn behavior.
* [Toroidal Space And Motion](../../services/game-server/simulation/world/toroidal-space-and-motion.md) owns wrapped world movement and spatial behavior.
* [Collision To Damage Flow](../../services/game-server/simulation/combat/collision-to-damage-flow.md) owns projectile/asteroid and ship/asteroid collision-to-damage implementation.
* [Damage](../combat/damage.md) owns the conceptual damage model used for asteroid health and destruction.
* [Pickups](../combat/pickups.md) owns pickup concepts that may be produced by asteroid destruction.
* [Radial Effects](../combat/radial-effects.md) owns radial hit-intent concepts that can affect asteroids.
* [Asteroid Variants Data](../../data/asteroid-variants-data.md) owns asteroid variant source data and catalog expectations.
* [Drop Tables](../../data/drop-tables.md) owns asteroid drop-table source data and generated output.
* [Collision Shape Data](../../data/collision-shape-data.md) owns collision shape source and generated data.
* [Gameplay Packets](../../protocol/gameplay-packets.md) owns the packet surface that carries asteroid state.
* [Asteroid Variant Contract](../../protocol/asteroid-variant-contract.md) owns the protocol contract for asteroid variant indexes.
* [Client Asteroid Variant Presentation](../../services/client/world-sync/asteroid-variant-presentation.md) owns client-side texture and scene presentation for asteroid variants.
* [Client Entity Sync Owners](../../services/client/world-sync/entity-sync-owners.md) owns client world-sync entity node ownership.

## Invariants

Asteroids must preserve these invariants:

* Asteroids are server-authoritative runtime entities.
* Asteroids are match-local and non-durable.
* Clients observe asteroid existence through server state.
* Clients do not create authoritative asteroids.
* Asteroid ids identify runtime entities, not catalog variants.
* Asteroid variant indexes are runtime integers, not stable id strings.
* Server spawn paths choose variants through asteroid catalog helpers.
* Spawn planning stays separate from mutation into the runtime entity store.
* Asteroid movement uses server-owned toroidal world behavior.
* Collision checks use server-loaded collision bodies.
* Client collision polygons are presentation, not authority.
* Projectile/asteroid collision can produce nonfatal damage.
* Asteroid destruction consequences require a destroyed damage result.
* Nonfatal asteroid damage must not award score, spawn fragments, or roll drops.
* Ship/asteroid collision damages the player, not the asteroid.
* Destroyed asteroids enter pending despawn before removal.
* Pending-despawn asteroids do not continue moving.
* Pending-despawn asteroids are skipped by collision checks.
* Far-from-camera cleanup is not destruction and must not award score, spawn fragments, or roll drops.
* Fragment asteroids are new entities with new ids.
* Fragment variants are selected independently rather than inherited from the source asteroid.
* Drop-table evaluation and pickup spawning stay outside the asteroid entity model.
* Score policy and score mutation stay outside the asteroid entity model.
* Client presentation must not re-decide asteroid gameplay outcomes.

## Active issues

Current asteroid-related limits are tracked in [Current System Limits](../../limits/current-system-limits.md#combat-systems).

Relevant current limits include:

* all current asteroid variants use the same collision-shape key
* all current asteroid variants use the same stats profile
* all current asteroid variants use the same drop-table key
* only `basicasteroids` drop tables exist today
* there is no minimum drop count policy yet
* radial asteroid presentation and broader radial outcomes are still limited

These are current implementation limits, not asteroid entity invariants.

## Related docs

* [Entities](./!INDEX.md)
* [Systems Design](../!INDEX.md)
* [Asteroid Spawning And Variants](../../services/game-server/simulation/world/asteroid-spawning-and-variants.md)
* [Visibility And Despawn](../../services/game-server/simulation/world/visibility-and-despawn.md)
* [Toroidal Space And Motion](../../services/game-server/simulation/world/toroidal-space-and-motion.md)
* [Collision To Damage Flow](../../services/game-server/simulation/combat/collision-to-damage-flow.md)
* [Runtime Entity Store](../../services/game-server/simulation/runtime/runtime-entity-store.md)
* [State Packet Projection](../../services/game-server/simulation/runtime/state-packet-projection.md)
* [Damage](../combat/damage.md)
* [Pickups](../combat/pickups.md)
* [Radial Effects](../combat/radial-effects.md)
* [Asteroid Variants Data](../../data/asteroid-variants-data.md)
* [Drop Tables](../../data/drop-tables.md)
* [Collision Shape Data](../../data/collision-shape-data.md)
* [Gameplay Packets](../../protocol/gameplay-packets.md)
* [Asteroid Variant Contract](../../protocol/asteroid-variant-contract.md)
* [Client Asteroid Variant Presentation](../../services/client/world-sync/asteroid-variant-presentation.md)
* [Client Entity Sync Owners](../../services/client/world-sync/entity-sync-owners.md)
* [Current System Limits](../../limits/current-system-limits.md#combat-systems)

## Notes

Asteroids currently carry enough structure for future richer behavior: health, damage modifiers, collision damage, variants, stats-profile data, drop-table data, and collision-shape data already have seams in or near the entity model.

Current implementation does not fully use every variant data field at runtime. In particular, current server collision-body lookup and pickup drop integration are narrower than the full asteroid variant data shape.

Future asteroid work should extend the existing authority split rather than moving asteroid decisions into the client, packet schema, weapon code, radial-effect timing, scoring policy, or pickup effects.
