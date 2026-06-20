## Asteroid Variant Presentation

Parent index: [World Sync](./!README.md)

## Purpose

This document describes the client-side asteroid variant presentation slice.

## Overview

`AsteroidSync` consumes server-provided asteroid variant ids from world-sync state and hands them to the asteroid scene node.

The generated asteroid variant catalog provides client-side texture-path lookup and variant-count wrapping. The asteroid scene then applies the selected texture and collision polygon presentation for the chosen variant.

This document stays on the client presentation side of the boundary. The shared variant data and server variant catalog remain separate ownership.

## Code root

```text
client/
```

Primary implementation areas:

```text
client/scripts/world/
client/scripts/entities/
client/scripts/generated/asteroids/
client/scenes/
client/tests/unit/entities/
```

## Responsibilities

The client asteroid variant presentation flow owns:

* consuming server-provided asteroid variant ids from `AsteroidSync`
* handing asteroid variant ids into asteroid scene nodes
* selecting asteroid textures through the generated asteroid variant catalog
* applying collision polygon presentation in the asteroid scene
* wrapping variant indexes when the generated catalog is indexed beyond its bounds
* verifying generated-variant consumption through client tests

## Does not own

The client asteroid variant presentation flow does not own:

* the asteroid variant source of truth
* server asteroid spawning or variant assignment
* packet schema ownership
* client world-sync entity creation generally
* gameplay collision outcomes
* server collision-shape authority

## Flow behavior

### Server variant-id consumption

`AsteroidSync` reads the server variant id from world-sync asteroid state and stores it with the asteroid presentation state.

The asteroid sync owner is responsible for handing that variant id to the asteroid scene the first time the node is initialized.

### Texture selection

`client/scripts/generated/asteroids/asteroid_variants.gd` exposes generated catalog lookups for asteroid textures.

`client/scripts/entities/asteroid.gd` uses the generated catalog to choose the texture path for the active variant and loads that texture onto the asteroid sprite.

### Collision presentation

The asteroid scene contains a base collision polygon plus variant collision polygon nodes under `CollisionVariants`.

`client/scripts/entities/asteroid.gd` applies the selected variant polygon to the scene and uses the variant collision shape presentation as a local client rendering concern.

### Variant wrapping

The generated asteroid variant catalog wraps indexes so client lookups remain stable when the supplied variant index falls outside the catalog size.

This wrapping is a client presentation convenience and should not be treated as source-of-truth behavior.

### Client tests

Client tests cover generated asteroid variant consumption and wrapping behavior.

## Code map

Primary implementation files:

```text
client/scripts/world/asteroid_sync.gd
client/scripts/entities/asteroid.gd
client/scripts/generated/asteroids/asteroid_variants.gd
client/scenes/asteroid.tscn
client/tests/unit/entities/test_asteroid_variants.gd
```

## Tests

The focused client test path documented for this flow is:

```text
client/tests/unit/entities/test_asteroid_variants.gd
```

It covers generated asteroid variant consumption and wrapping behavior.

## Related docs

* [Entity Sync Owners](entity-sync-owners.md)
* [Asteroid Variants Data](../../../data/stubs/asteroid-variants-data.md)
* [Gameplay Packets Stub](../../../protocol/stubs/gameplay-packets.md)
* [World Sync](./!README.md)
* [Client](../!README.md)
* [Services](../../!README.md)

## Notes

This document captures the client presentation side of asteroid variants only. It does not replace the asteroid variant data source doc or the server asteroid spawning and variant docs.
