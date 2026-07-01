# Projectiles

Parent index: [Entities](./!INDEX.md)

## Purpose

This document defines the systems-design model for projectile entities in Space Rocks.

It documents projectile identity, runtime state, authority boundaries, movement, lifetime, collision participation, damage metadata, impact effects, client presentation, and invariants. It does not replace service implementation docs, packet docs, data pipeline docs, combat design docs, or devtools documentation.

## Overview

Projectiles are server-authoritative moving combat entities.

A projectile is created by authoritative server behavior, stored in the live runtime entity store, advanced by simulation, projected to clients as world lane bullet records, and removed by server-owned lifetime, visibility, pending-despawn, or collision consequences.

The current runtime implementation type is still named `Bullet`, but conceptually a projectile is broader than a basic bullet. Weapon-backed projectiles carry weapon identity and projectile type metadata, so the same runtime entity family can represent:

```text
basic cannon bullet
torpedo projectile
future weapon-backed projectile types
```

Current implemented projectile types are:

```text
bullet
torpedo
```

The conceptual split is:

```text
Projectile spawn intent
= pure weapon-fire result describing what should be created

Runtime projectile
= live server-owned entity stored in game.entities.Projectiles

Projectile presentation
= client scene node selected from projected projectile state
```

The server owns projectile existence, identity, movement, lifetime, collision outcomes, damage metadata, impact effects, and removal. The client owns rendering, interpolation, audio/visual feedback, and scene selection from server-provided read models.

## Conceptual model

Projectile behavior follows this lifecycle:

```text
authoritative fire or debug spawn
-> projectile spawn intent or direct debug spawn
-> runtime projectile entity
-> motion and toroidal wrapping
-> world lane realtime projection
-> client presentation
-> collision or expiry
-> damage and impact-effect handling
-> pending despawn
-> removal from authoritative entity map
```

Normal player projectiles are created from weapon fire:

```text
client input intent
-> server input state
-> server weapon fire policy
-> projectile spawn intent
-> runtime projectile
```

The weapon fire policy does not create a live entity. It returns an intent. The game aggregate adapts that intent into a runtime projectile, assigns an authoritative projectile id, stores it in `game.entities.Projectiles`, and later projects it through world lane bullet records.

## Projectile identity

Projectile ids are server-assigned runtime ids.

The current game-server spawner allocates projectile ids for both normal weapon-backed projectiles and debug-created bullets. Clients do not choose projectile ids.

A projectile also carries an owner id. The owner id identifies the player that created or owns the projectile for consequence paths such as scoring, hit ownership, or future attribution.

Weapon-backed projectiles also carry:

```text
weapon_id
projectile_type
```

These fields are not the same concept:

```text
weapon_id
= which weapon profile produced the projectile

projectile_type
= what projectile family should be represented and interpreted
```

For example:

```text
weapon_id: basic_cannon
projectile_type: bullet

weapon_id: torpedo
projectile_type: torpedo
```

Current debug or legacy bullet creation can leave weapon and projectile type metadata empty. The client presentation fallback treats an empty projectile type as `bullet`.

## Runtime state

The current runtime projectile state includes:

```text
id
owner id
weapon id
projectile type
impact effect metadata
position
rotation
velocity
life remaining
legacy integer damage
damage spec
pending-despawn flag
despawn delay
```

The authoritative server stores runtime projectiles in:

```text
game.entities.Projectiles
```

Runtime projectile state is match-local. It is not durable account data, inventory data, profile data, or replay data.

The projectile entity carries enough data for later systems to resolve consequences:

```text
motion uses position, velocity, life, pending-despawn state

collision uses position, rotation, and collision body

damage request adapters use damage spec or legacy damage fallback

impact-effect adapters use impact effect metadata

world lane projection uses id, owner id, position, rotation, weapon id, and projectile type
```

## Creation authority

The server owns projectile creation.

The client may request firing through normal input state, but it cannot authoritatively create projectiles or supply projectile identity, damage, velocity, cooldown, ammo, or impact-effect data.

Normal player projectile creation follows this boundary:

```text
InputState.primary_fire / InputState.secondary_fire
-> server simulation
-> weapon fire policy
-> projectile spawn intent
-> runtime projectile insertion
```

The server may reject firing before a projectile exists. Current rejection reasons include missing equipment, unknown weapon profile, active cooldown, empty limited ammo, or runtime player state that does not allow shooting.

Debug projectile creation is separate devtools behavior. It still creates server-owned runtime projectiles and stores them in the authoritative projectile map. It should not be treated as normal gameplay weapon policy.

## Motion and lifetime

Projectile motion is server-owned.

Projectile movement uses the same bounded toroidal world model as ships and asteroids. Each simulation step may advance projectile position from velocity, decrement projectile lifetime, and wrap the projectile into world bounds.

Conceptually:

```text
projectile position += projectile velocity * delta
projectile life -= delta
projectile position wraps into world bounds
```

When a projectile is pending despawn, normal movement stops. The projectile instead counts down its despawn delay until it is ready for removal.

Projectile removal is server-owned. Current removal reasons include:

```text
pending despawn delay completed
lifetime expired
projectile is far from all cameras
```

The client removes projectile scene nodes when a projectile is missing from the latest server state. Client-side absence is not an authority decision; it is a presentation response to server state.

## Toroidal world behavior

Projectiles exist at one bounded authoritative server position.

The server does not duplicate projectiles at world seams. Instead, projectile movement wraps the single authoritative position, and spatial consumers use wrapped deltas when needed.

Current wrapped projectile consumers include:

```text
projectile movement
projectile visibility and far-from-camera cleanup
projectile -> asteroid collision checks
target candidate generation
client visual positioning
devtools collision telemetry
```

The client renders projectiles relative to the active view anchor. It converts bounded server coordinates into continuous visual positions using shortest wrapped deltas from the active anchor. That visual continuity does not change server authority.

## Collision participation

Current normal gameplay collision participation is:

```text
projectile -> asteroid
```

Projectile collision uses server collision bodies. The projectile body is built from the shared bullet collision shape, current projectile position, and current projectile rotation. The asteroid body is placed into projectile-local wrapped space before primitive collision detection runs, so cross-edge collisions work without duplicating runtime entities.

On projectile/asteroid collision, the game-server combat path:

```text
builds a projectile-to-asteroid collision fact
builds a damage request from the projectile and asteroid
resolves damage through the damage system
applies damage result to the asteroid
records damage presentation when applicable
spawns projectile impact effects when metadata requests them
marks the projectile pending despawn
applies asteroid destruction consequences when the asteroid is destroyed
```

Projectiles do not currently own collision consequences themselves. They are inputs to collision, damage, effects, scoring, spawning, and event paths.

## Damage metadata

Projectiles carry damage intent.

A projectile does not resolve damage by itself. Damage resolution happens through the damage system after a server-owned collision or effect path builds a damage request.

Weapon-backed projectiles carry a `DamageSpec` from the weapon profile. The current projectile/asteroid damage adapter prefers that damage spec. If the damage spec is empty, it falls back to the legacy integer damage field as kinetic projectile damage.

Current examples:

```text
basic cannon
-> projectile type bullet
-> kinetic projectile damage
-> no impact effect

torpedo
-> projectile type torpedo
-> explosive projectile damage intent
-> radial impact effect metadata
```

The current torpedo profile is primarily valuable because of its radial impact effect. Direct impact damage is separate from radial damage and should not be collapsed into one concept.

## Impact effects

Impact effects are optional projectile metadata.

A projectile may carry impact-effect metadata that becomes active only when an authoritative impact path consumes it. Current impact-effect kinds are:

```text
none
radial
```

The basic cannon uses no impact effect.

Torpedo uses a radial impact effect. When a torpedo projectile impacts, the game-owned impact adapter spawns a radial effect at the impact position. The radial effect system then owns timing, zone coverage, target filtering, hit-intent generation, and expiration.

The ownership split is:

```text
weapon profile
-> defines impact effect metadata

projectile entity
-> carries metadata until impact

game impact adapter
-> spawns effect from metadata

radial effect system
-> steps effect and emits hit intents

damage system
-> resolves resulting damage
```

Projectiles do not step radial effects. Radial effects do not own projectile lifetime or projectile storage.

## State projection

Projectile state is projected to clients through world lane full/delta packets.

Current projected projectile fields are:

```text
id
owner_id
x
y
rotation
weapon_id
projectile_type
```

Projectile world lane packets do not expose:

```text
velocity
life remaining
damage spec
impact effect metadata
pending-despawn flag
despawn delay
```

Those are server-side runtime facts. The client receives only the read model needed for presentation and target-position support.

A projectile disappearing from world lane bullet records means the authoritative server no longer presents that projectile as live. The client should remove its corresponding scene node.

## Client presentation

The client owns projectile presentation only.

Client world sync renders server projectiles through `ProjectileSync`. It creates local scene nodes for newly seen projectile ids, removes scene nodes missing from the latest server state, applies server-derived target positions and rotations, interpolates rendered nodes, and plays projectile firing presentation when a projectile first appears.

Projectile scene selection uses `projectile_type`:

```text
torpedo -> torpedo scene
bullet  -> bullet scene
empty or unknown -> bullet fallback
```

The client may present projectile visuals, audio, interpolation, and targeting read-model positions. It must not decide that a projectile exists, has hit something, has expired, has damaged a target, or has spawned an impact effect.

## Targeting participation

Projectiles can participate in target read models.

The canonical target kind for current projectile entities is:

```text
bullet
```

This target kind name reflects the current runtime and protocol vocabulary. It does not mean the conceptual entity is limited to only basic cannon bullets.

Server-side target selection can include projectiles as target candidates when they exist and have usable collision bodies. A pending-despawn projectile is treated as inactive for target status. A missing projectile is treated as missing.

Targeting does not make a projectile damageable, collectible, or eligible for combat effects by itself. It is a selection/read-model system, not a combat consequence system.

## Authority rules

Projectile authority follows these rules:

```text
The server owns projectile creation, ids, owner identity, movement, lifetime, collision participation, damage metadata, impact metadata, pending despawn, and removal.

Weapon fire policy owns projectile spawn intent, not live projectile storage.

The game aggregate owns adapting projectile spawn intent into runtime entities.

The motion system owns per-projectile movement and lifetime stepping, not entity-map insertion or deletion.

The collision system owns projectile collision facts, not damage math or scoring.

The damage system owns damage resolution, not projectile movement or projectile removal.

The radial effect system owns radial timing, coverage, filtering, and hit intents after a projectile impact spawns an effect.

The realtime protocol carries projectile read models, not projectile authority.

The client owns projectile presentation only.
```

No client packet may be treated as authority for projectile id, projectile type, projectile position, projectile velocity, projectile damage, impact effect, lifetime, or removal.

## Invariants

Projectile entities must preserve these invariants:

```text
Projectile spawn intent is not a live entity.

Live projectiles exist only when stored in the authoritative server projectile map.

Projectile ids are server-assigned.

Projectile movement is server-authoritative.

Projectile positions use bounded toroidal world coordinates on the server.

Client visual continuity does not change authoritative projectile coordinates.

Projectile damage is intent until a server-owned impact path resolves it.

Projectile impact effects are metadata until a server-owned impact path spawns an effect.

Projectiles do not resolve their own damage.

Projectiles do not award score, spawn fragments, drop pickups, or mutate player lives directly.

Projectile disappearance from client state is observed, not decided, by the client.

Projectile scene selection is presentation; it does not define server behavior.

The current runtime name `Bullet` should not force future projectile concepts to be modeled as basic bullets.
```

## Participating systems

Projectile entities participate in these systems:

```text
Game server simulation
= authoritative projectile storage, movement, collision, impact, and removal

Weapons
= weapon profiles and projectile spawn intent

Toroidal world and motion
= projectile movement, lifetime stepping, and wrapped coordinates

Collision shapes and physics
= projectile collision body and primitive overlap checks

Damage
= authoritative damage resolution after projectile impact

Radial effects
= timed area effects spawned from projectile impact metadata

Scoring, asteroid destruction, and drops
= downstream consequences of projectile-driven asteroid destruction

Realtime protocol
= projectile world lane projection to clients

Client world sync
= projectile node creation, interpolation, scene selection, and removal

Targeting
= projectile target candidates and target read-model status

Data pipeline
= generated constants, packet fields, and collision shapes
```

## Active issues

Current projectile-related limits are tracked in [Current System Limits](../../limits/current-system-limits.md#combat-systems).

Relevant current limits include:

```text
Bullet/pickup collision damage is not enabled.

Torpedo radial currently targets asteroids and enemies only; players, projectiles, and pickups are excluded.

Radial client visuals are not fully implemented in the radial design path.
```

These limits should not be documented as permanent projectile design constraints.

## Related docs

* [Entities](./!INDEX.md)
* [Systems Design](../!INDEX.md)
* [Weapons](../combat/weapons.md)
* [Damage](../combat/damage.md)
* [Radial Effects](../combat/radial-effects.md)
* [Targeting](../combat/targeting.md)
* [Weapons And Projectile Fire](../../services/game-server/simulation/combat/weapons-and-projectile-fire.md)
* [Collision To Damage Flow](../../services/game-server/simulation/combat/collision-to-damage-flow.md)
* [Radial Effects Service Implementation](../../services/game-server/simulation/combat/radial-effects.md)
* [Runtime Entity Store](../../services/game-server/simulation/runtime/runtime-entity-store.md)
* [Lane Packet Projection](../../services/game-server/simulation/runtime/lane-packet-projection.md)
* [Toroidal Space And Motion](../../services/game-server/simulation/world/toroidal-space-and-motion.md)
* [Collision Shapes](../../services/game-server/simulation/world/collision-shapes.md)
* [Entity Sync Owners](../../services/client/world-sync/entity-sync-owners.md)
* [Gameplay Packets](../../protocol/gameplay-packets.md)
* [Packet Schemas](../../data/packet-schemas.md)
* [Constants](../../data/constants.md)
* [Collision Shape Data](../../data/collision-shape-data.md)
* [Current System Limits](../../limits/current-system-limits.md#combat-systems)

## Notes

The implementation still uses `Bullet` in runtime type names, packet field names, some target vocabulary, and some client layer names. Treat `projectile` as the conceptual entity family and `Bullet` as the current runtime/protocol naming residue.

The current client fallback for missing or unknown projectile type is `bullet`. That is presentation fallback behavior, not a rule that every projectile is conceptually a bullet.

Projectile docs should not absorb the full weapon system. Weapons own firing capability and spawn intent. Projectiles own the live entity behavior after creation.

