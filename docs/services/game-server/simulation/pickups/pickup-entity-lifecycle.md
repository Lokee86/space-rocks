# Pickup Entity Lifecycle

Parent index: [Game Server Simulation Pickups](./!INDEX.md)

## Purpose

This document describes the current game-server pickup entity lifecycle boundary.

It covers pickup creation, runtime storage, authoritative aging, expiry, state projection, removal, and lifecycle event recording.

## Overview

Pickups are server-authoritative runtime entities owned by the game server simulation.

A pickup enters the world when game-owned code calls the pickup spawn API with a known pickup type and position. The spawn path validates the pickup type against server definitions, allocates a server pickup ID, creates the runtime entity, initializes age to zero, copies health and lifespan from the definition, and stores the entity in `game.entities.Pickups`.

Current lifecycle flow:

```text
spawn source
-> Game.spawnPickupLocked
-> entities.Pickups[pickup_id]
-> Game.Step
-> stepPickups(delta)
-> StatePacket.pickups while active
-> collection removal or expiry removal
-> event lane when applicable
```

Pickup age and expiry are advanced by the server during simulation. The client receives pickup age and lifespan through `StatePacket.pickups` and derives remaining lifetime locally for presentation. The server does not send a `remaining_lifespan` field.

A pickup leaves the authoritative entity map in one of two current gameplay paths:

```text
player/pickup collision
-> removePickupLocked
-> pickup_collected event
-> pickup effect application path
```

```text
age_seconds >= lifespan_seconds
-> pickup_expired event
-> delete from entities.Pickups
```

Pickup collection and pickup effects are separate service boundaries. This document only covers the entity lifecycle side: the pickup exists, ages, is projected, and is removed.

## Code root

```text
services/game-server/internal/game/
services/game-server/internal/game/entities/pickups/
```

## Responsibilities

The pickup entity lifecycle boundary owns:

* Validating pickup type definitions during spawn.
* Allocating authoritative pickup IDs.
* Creating runtime pickup entities.
* Storing active pickups in the game entity store.
* Initializing pickup position, health, age, and lifespan.
* Advancing pickup age during simulation.
* Expiring pickups whose positive lifespan has been reached.
* Removing expired pickups from the active entity map.
* Recording `pickup_expired` domain events before expiry removal.
* Projecting active pickup entities into `StatePacket.pickups`.
* Providing class-based collision bodies for pickup collision consumers.
* Keeping pickup scene paths out of server definitions.

The game-owned lifecycle adapter owns the mutation of `game.entities.Pickups`. The pickup entity package owns the pickup struct, definition lookup, class lookup, position read model, and collision-body construction.

## Does not own

Pickup entity lifecycle does not own:

* Drop-table chance evaluation.
* Drop-table source rules.
* Pickup collection intent resolution.
* Pickup effect rules.
* Player lives, weapons, ammo, inventory, or profile mutation.
* Client pickup rendering.
* Client pickup lifespan blinking.
* Pickup audio or particle effects.
* Devtools pickup selector UI.
* Realtime packet schema source-of-truth files.
* Collision-shape source-of-truth data.
* Bullet/pickup damage rules.

Those concerns belong to the drop-table, pickup collection, pickup effects, client presentation, data pipeline, or combat boundaries.

## Domain roles

Pickup entity lifecycle participates in the authoritative pickup domain as the owner of active pickup existence.

Its domain role is to answer:

```text
Does this pickup currently exist?
Where is it?
What type and class is it?
How old is it?
When should it expire?
What state should clients receive while it exists?
```

It does not answer what a pickup does after collection. That is handled by the collection and effect-intent seams.

Current pickup lifecycle is intentionally source-agnostic after creation. Devtools spawn, drop-table integration, and any later normal spawn system should all converge on the same authoritative spawn path instead of creating parallel pickup entity rules.

## Protocols and APIs

### Internal spawn API

Pickup creation is exposed through:

```text
Game.SpawnPickup(pickupType, position)
```

The public method locks the game and delegates to:

```text
Game.spawnPickupLocked(pickupType, position)
```

The locked spawn path:

1. Looks up the pickup definition with `pickups.DefinitionFor`.
2. Rejects unknown pickup types with an error.
3. Increments `game.nextPickupID`.
4. Creates an ID using the `pickup_<number>` format.
5. Creates the pickup entity with server-defined type, health, and lifespan.
6. Stores the pickup in `game.entities.Pickups`.
7. Returns the created entity.

Current spawn sources include:

* drop-table integration after asteroid destruction
* devtools pickup spawn commands

Drop-table integration records `pickup_dropped` after a successful spawn. Devtools spawn currently creates the authoritative pickup and logs the debug action, but does not record a `pickup_dropped` gameplay event.

### Internal remove API

Pickup removal is exposed through:

```text
Game.RemovePickup(id)
```

The public method locks the game and delegates to:

```text
Game.removePickupLocked(id)
```

The locked remove path deletes the pickup from `game.entities.Pickups` and returns whether an entity existed.

Removal itself does not choose an event. The caller owns the semantic reason:

* collection flow records `pickup_collected`
* lifecycle expiry records `pickup_expired`
* drop flow records `pickup_dropped` at creation time, not removal time

### Simulation lifecycle

`Game.Step(delta)` calls `stepPickups(delta)` during normal simulation and during the match-over branch.

Normal simulation order currently places pickup stepping before collision resolution:

```text
stepPlayerWeapons
stepPlayers
removeReadyPlayers
stepAsteroidSpawning
stepAsteroids
stepBullets
stepPickups
stepCollisions
stepRadialEffects
```

When the match is over, the server still advances asteroids, bullets, pickups, and radial effects, but skips normal gameplay spawning and collision handling.

`stepPickups(delta)` increments each pickup's age by the simulation delta. If `LifespanSeconds` is positive and `AgeSeconds` is greater than or equal to `LifespanSeconds`, it records `pickup_expired` and removes the pickup from the active entity map.

A non-positive lifespan currently means the pickup does not expire through `stepPickups`.

### State packet surface

Active pickups are projected into `StatePacket.pickups`.

The current packet-facing pickup fields are:

```text
id
type
pickup_class
x
y
health
age_seconds
lifespan_seconds
```

`type` is the gameplay pickup identity.

`pickup_class` is derived from the server definition and selects the class-level presentation/collision family, such as `powerup` or `weapon`.

`age_seconds` and `lifespan_seconds` are authoritative server lifecycle values. The client derives remaining lifetime from them.

### Event surface

Pickup lifecycle currently uses the server gameplay event lane for expiry.

`pickup_expired` includes:

```text
pickup_id
pickup_type
x
y
```

The event is recorded before the pickup is deleted from `game.entities.Pickups`.

Adjacent pickup events are produced by other pickup boundaries:

```text
pickup_dropped          -> drop integration after successful spawn
pickup_collected        -> collection flow after player/pickup collision
pickup_effect_applied   -> effect application after successful gameplay mutation
```

## Data ownership

### Runtime entity state

Pickup runtime entities carry:

```text
ID
Type
X
Y
Health
AgeSeconds
LifespanSeconds
```

Runtime pickup entities live in:

```text
game.entities.Pickups
```

The entity store type is defined by the game runtime package and stores pickups as a map keyed by pickup ID.

### Definition state

Pickup definitions carry:

```text
Type
Class
Health
LifespanSeconds
```

Current implemented pickup definitions include:

```text
1_up
torpedo
```

Current generated constants give both implemented pickup types positive health and a 12.0 second lifespan.

### Collision state

Pickup collision bodies are built from pickup class, not pickup type.

Current class shape keys include:

```text
powerup
weapon
```

The pickup entity package asks the collision shape catalog for a pickup shape using the pickup class. This keeps server collision lookup aligned with generic pickup scene families instead of per-pickup type names.

### Packet state

Pickup packet state is generated from the shared gameplay packet schema and projected by the game package.

The server sends active pickup state only. Removed pickups disappear from later `StatePacket.pickups` maps, with semantic removal feedback carried by events when applicable.

## Code map

### Pickup entity model

```text
services/game-server/internal/game/entities/pickups/types.go
```

Defines pickup class and type identifiers.

```text
services/game-server/internal/game/entities/pickups/definitions.go
```

Maps implemented pickup types to generated server constants.

```text
services/game-server/internal/game/entities/pickups/pickup.go
```

Defines the runtime pickup entity, position read model, class lookup, and collision-body construction.

### Game-owned lifecycle adapter

```text
services/game-server/internal/game/pickups.go
```

Owns public and locked pickup spawn/remove helpers and pickup state projection.

```text
services/game-server/internal/game/pickup_lifecycle.go
```

Owns authoritative pickup aging, expiry detection, expiry event recording, and expiry removal.

```text
services/game-server/internal/game/runtime/state.go
```

Defines the entity store that contains active pickups.

```text
services/game-server/internal/game/simulation.go
```

Calls `stepPickups(delta)` during both normal simulation and match-over simulation.

```text
services/game-server/internal/game/state_packet.go
```

Projects active pickup state into outbound gameplay state packets.

### Event adapter

```text
services/game-server/internal/game/events/events.go
```

Defines the pickup-related domain event vocabulary.

```text
services/game-server/internal/game/events.go
```

Converts pickup lifecycle and adjacent pickup domain events into packet-facing event state.

### Adjacent lifecycle callers

```text
services/game-server/internal/game/pickup_drops.go
```

Creates pickups from drop-table results and records `pickup_dropped`.

```text
services/game-server/internal/game/pickup_collisions.go
```

Removes pickups after player/pickup collision and routes collection into the pickup collection seam.

```text
services/game-server/internal/devtools/spawn_pickup.go
```

Creates authoritative pickups from debug spawn commands.

### Adjacent pickup behavior

```text
services/game-server/internal/game/pickups/collection.go
```

Resolves collection results and effect intents after lifecycle removal.

```text
services/game-server/internal/game/pickup_effects.go
```

Applies pickup effect intents to player sessions or active ship weapon state.

### Generated and source data

```text
shared/constants/pickups.toml
shared/constants/weapon_pickups.toml
shared/packets/gameplay.toml
shared/collisions/collision_shapes.json
services/game-server/internal/constants/powerups.go
services/game-server/internal/constants/weapon_pickups.go
services/game-server/internal/game/runtime/packets_generated.go
services/game-server/internal/game/packets.go
```

## Tests

Relevant pickup lifecycle and adjacent coverage includes:

```text
services/game-server/internal/game/entities/pickups/definitions_test.go
```

Verifies pickup definitions for implemented pickup types and rejection of unknown pickup types.

```text
services/game-server/internal/game/entities/pickups/pickup_test.go
```

Verifies pickup collision bodies use class-level shape keys.

```text
services/game-server/internal/game/pickup_drops_test.go
```

Verifies drop integration creates pickups, respects active pickup caps, and projects spawned pickups into state packets.

```text
services/game-server/internal/game/events_test.go
```

Verifies `pickup_expired` and other pickup domain events convert into packet-facing event state.

```text
services/game-server/internal/game/pickups/collection_test.go
```

Verifies adjacent collection result and effect-intent behavior.

```text
services/game-server/internal/game/pickup_effects_test.go
```

Verifies adjacent pickup effect application behavior for weapon pickup ammo.

Use normal game-server test verification when changing pickup lifecycle behavior:

```text
go test ./...
```

from `services/game-server`.

## Related docs

* [Pickup Collection](pickup-collection.md)
* [Pickup Effects](pickup-effects.md)
* [Pickup Drop Integration](pickup-drop-integration.md)
* [Game Server Simulation Pickups](./!INDEX.md)
* [Game Server Simulation](../!INDEX.md)
* [Game Server](../../!INDEX.md)
* [Client Pickup Presentation](../../../client/world-sync/pickup-presentation.md)
* [Gameplay packets](../../../../protocol/stubs/gameplay-packets.md) - Stub: gameplay realtime packet documentation.
* [Constants pipeline](../../../../data/stubs/constants-pipeline.md) - Stub: generated constant data documentation.
* [Packet schema pipeline](../../../../data/stubs/packet-schema-pipeline.md) - Stub: packet schema data documentation.
* [Collision shape data](../../../../data/stubs/collision-shape-data.md) - Stub: collision-shape source-of-truth documentation.
* [Pickup entities](../../../../systems-design/entities/stubs/pickup-entities.md) - Stub: pickup entity design documentation.

## Notes

Pickup health is currently part of the runtime and packet-facing pickup state, but the lifecycle path does not currently reduce pickup health. Player/pickup collection removes the pickup directly after collision.

The client derives end-of-life presentation from `age_seconds` and `lifespan_seconds`. The server remains authoritative for actual expiry.

Projectile/pickup collision damage is not part of the current pickup lifecycle implementation.
