# Stub: Asteroid Spawning And Variants

Parent index: [Game Server Simulation World](../!README.md)

## Purpose

This stub documents the current server implementation for asteroid spawning, variant selection, runtime storage, and collision-shape lookup.

## Overview

The server owns asteroid variant catalog behavior and the spawn paths that choose variants for new asteroids.

The current implementation keeps the variant catalog in `services/game-server/internal/game/asteroids/variants.go`, uses weighted selection for spawn paths, stores the selected variant on runtime asteroids, and exports the variant id through asteroid runtime state.

## Responsibilities

Server-side asteroid spawning and variant ownership includes:

* consuming the asteroid variant catalog
* selecting variants with weighted random choice
* assigning variants for timed asteroid spawns
* assigning variants for fragment asteroid spawns
* assigning variants for debug asteroid spawns
* storing the selected variant on runtime asteroid state
* exporting asteroid variant ids from asteroid runtime state
* looking up collision shapes by asteroid size and variant

Timed and fragment spawn planning is owned by the spawner layer. Variant catalog selection is owned by the asteroid variant helpers.

## Does not own

This stub does not document:

* client presentation behavior
* packet schema details
* data-sync or source-of-truth mechanics beyond the fact that server asteroid state exports variant ids
* client-side asteroid variant presentation

## Code root

```text
services/game-server/
```

## Code map

### Variant catalog and selection

* `services/game-server/internal/game/asteroids/variants.go`
* `services/game-server/internal/game/asteroids/variants_test.go`

### Spawn planning

* `services/game-server/internal/game/spawning/spawner.go`

### Runtime asteroid storage and shape lookup

* `services/game-server/internal/game/runtime/asteroid.go`

## Related docs

* [Game Server Simulation World](../!README.md)
* [Game Server Simulation](../../!README.md)

## Notes

This stub is intentionally limited to current server-side asteroid spawning and variant responsibilities.
