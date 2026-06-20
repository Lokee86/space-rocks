# Variants

Parent index: [Entities](./!README.md)

## Purpose

This document defines the conceptual model, authority rules, and invariants for entity variants in Space Rocks.

It explains what a variant means at the entity-design level, how variant identity differs from runtime entity state, and which systems are allowed to choose or consume variants.

## Overview

An entity variant is a stable identity or catalog entry that changes how an entity is configured, selected, or presented without changing the entity’s base role.

Variants are not a generic client-side skin system. A variant may affect presentation, spawn eligibility, collision lookup, stats lookup, drop-table lookup, or future loadout compatibility, but each of those effects must be owned by the correct authoritative system.

The current implementation has two different variant states:

```text
Asteroid variants
= implemented catalog-backed runtime variant indexes.

Ship variants
= partial server-side ship type/stat/collision seam, with only default v_wing currently used.
```

Asteroid variants are the current complete entity-variant example. The server selects an asteroid variant index when an asteroid is created, stores it on the runtime asteroid, exports it through asteroid state, and the client consumes it for presentation.

Ship variants are not fully implemented as selectable player-build content. The server currently carries a `ship_type` seam and resolves stats from that identity, but only the default `v_wing` ship type is active in normal gameplay.

Projectiles and pickups currently use their own entity-type, profile, weapon, pickup-class, and effect identifiers. They should not be folded into the variant model unless they later need a distinct catalog-backed variant layer.

## Conceptual model

A variant has three conceptual layers:

```text
Stable catalog identity
-> authoritative runtime selection
-> service-specific consumption
```

Stable catalog identity is the durable meaning of the variant. For asteroid variants, this is the catalog entry such as `asteroid_1`. For future ship variants, this will be a ship or chassis identity such as `v_wing`.

Authoritative runtime selection is the actual value chosen for live gameplay. For asteroids, this is the zero-based integer index stored on the runtime asteroid and sent through `AsteroidState.variant`. For ships, this is the server-owned `ship_type` / `ShipTypeID` value carried by player session and ship state.

Service-specific consumption is what each system does with the selected variant. The server may use it for spawn selection, stats, collision, drops, or build resolution. The client may use it for presentation. Protocol and data docs own the exact packet and source-of-truth contracts.

## Current asteroid variant model

Asteroid variants are catalog-backed.

The current asteroid catalog contains eight variants:

```text
asteroid_1 -> index 0
asteroid_2 -> index 1
asteroid_3 -> index 2
asteroid_4 -> index 3
asteroid_5 -> index 4
asteroid_6 -> index 5
asteroid_7 -> index 6
asteroid_8 -> index 7
```

The stable catalog id is not the runtime protocol value. The runtime protocol value is the zero-based index.

The server chooses asteroid variants for:

```text
timed asteroid spawns
asteroid fragment spawns
debug asteroid spawns
```

Each spawn source has its own spawn-weight field. A positive weight makes the variant eligible for that spawn source. A zero or negative weight excludes it. All current asteroid variants have equal weights for all current spawn sources.

Asteroid fragments receive newly selected fragment variants. They do not inherit the source asteroid’s variant.

## Current ship variant seam

The current ship variant model is a partial seam, not a full selectable variant system.

The game server currently carries:

```text
ShipTypeID
ShipStats
ShipStatModifiers
CollisionShapeID
ship_type packet field
```

The default ship type is:

```text
v_wing
```

Player sessions preserve `ShipTypeID`, resolve stats from it, and use those stats when spawning or respawning ships. The ship state exported to clients includes `ship_type`.

Full ship variants are not yet implemented. Current limits include:

```text
only v_wing is used
no full selectable ship variant catalog
no loadout-backed ship selection
no keyed multi-ship collision catalog
no client ship scene mapping from ship_type
```

The existing ship seam should be treated as the current ownership path that future ship variants will grow through, not as a finished player-build system.

## Authority rules

The game server is authoritative for runtime entity variant assignment during gameplay.

For asteroids, the server owns:

```text
when an asteroid exists
which variant index the asteroid receives
which spawn-source helper selects that variant
which variant index is stored on runtime.Asteroid
which variant index is exported in AsteroidState
```

The client owns presentation consumption only. It may map a received variant index to texture or scene presentation. It must not choose authoritative asteroid variants or reinterpret variant identity as gameplay authority.

For ships, the server owns:

```text
ship type identity
resolved ship stats
collision shape id used by server collision behavior
ship_type values exported through state packets
```

Future player-build, loadout, inventory, or hangar systems may participate in choosing an eligible ship variant before match start, but the final runtime ship identity must still be server-validated before it affects gameplay.

The data catalog owns stable variant meaning. The protocol owns the field names and transport shape. Service implementations consume those values within their own responsibility boundaries.

## Invariants

Entity variants must preserve these rules:

* A variant must not make the client authoritative over gameplay state.
* Stable variant ids and runtime protocol values must not be confused.
* Runtime variant values must be interpreted through the correct catalog for that entity type.
* Asteroid runtime indexes remain zero-based.
* Asteroid spawn code must select variants through catalog helpers, not hardcoded random pools.
* Client asteroid rendering must consume the server-provided variant index.
* Client lookup wrapping is a safety behavior, not version negotiation.
* Variant catalog changes that alter runtime index meaning are protocol-affecting changes.
* Ship collision behavior remains server-owned.
* Full ship selection must flow through future build/loadout validation, not direct client request authority.
* Variant-specific presentation must not imply variant-specific gameplay unless the server implements and owns that gameplay behavior.
* Data, protocol, and service docs own exact implementation contracts; this document owns the conceptual boundary.

## Participating systems

The current asteroid variant flow participates in:

```text
shared asteroid variant data
game-server asteroid spawning
game-server runtime asteroid state
realtime gameplay state packets
client world sync
client asteroid presentation
collision-shape data
drop-table data
devtools debug asteroid spawning
```

The current ship variant seam participates in:

```text
game-server player sessions
game-server runtime ship state
game-server resolved ship stats
game-server collision shape lookup
realtime gameplay state packets
client state consumption
client devtools hitbox/debug presentation
future player-build and loadout planning
```

Projectiles, pickups, weapons, modules, and future enemies may eventually grow variant-like catalogs, but they should do so only when the base type/profile identity is no longer enough to describe the needed behavior.

## Service implementation touchpoints

Asteroid variant behavior is implemented through the game server and client, with source and protocol ownership split out to data and protocol docs.

Primary asteroid service touchpoints:

```text
services/game-server/internal/game/asteroids/variants.go
services/game-server/internal/game/spawning/spawner.go
services/game-server/internal/game/runtime/asteroid.go
client/scripts/generated/asteroids/asteroid_variants.gd
client/scripts/world/asteroid_sync.gd
client/scripts/entities/asteroid.gd
```

Primary ship seam touchpoints:

```text
services/game-server/internal/game/runtime/state.go
services/game-server/internal/game/runtime/ship.go
services/game-server/internal/game/runtime/ship_stats.go
services/game-server/internal/game/session.go
shared/packets/gameplay.toml
```

These implementation touchpoints are listed to clarify the conceptual boundary. Detailed ownership, tests, and code maps belong in the related service, data, and protocol docs.

## Related docs

* [Entities](./!README.md)
* [Asteroids](asteroids.md)
* [Ships](ships.md)
* [Asteroid Variants Data](../../data/asteroid-variants-data.md)
* [Asteroid Variant Contract](../../protocol/asteroid-variant-contract.md)
* [Server Asteroid Spawning And Variants](../../services/game-server/simulation/world/asteroid-spawning-and-variants.md)
* [Client Asteroid Variant Presentation](../../services/client/world-sync/asteroid-variant-presentation.md)
* [Player Build And Loadouts](../../planning/domains/gameplay/player-build-and-loadouts.md)
* [Player Build Limits](../../limits/player-build-limits.md)

## Notes

Current asteroid variant data includes `collision_shape`, `stats_profile`, and `drop_table` fields. Current runtime behavior does not fully apply each of those fields as separate gameplay-owned variant effects.

Current server collision-body construction uses the runtime asteroid variant index against the loaded collision-shape catalog. It does not resolve the asteroid variant catalog `collision_shape` string as a separate runtime key.

Current pickup drop integration does not select drop tables from the runtime asteroid variant. It uses the basic asteroid drop-table path directly.

Current ship variant support should be described as a seam, not a completed selectable variant system.
