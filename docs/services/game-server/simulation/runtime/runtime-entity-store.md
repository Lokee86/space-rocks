# Runtime Entity Store

Parent index: [Game Server Simulation Runtime](./!INDEX.md)

## Purpose

This document describes the in-memory entity store used by the game-server simulation runtime.

The store groups the live, match-local entity maps that `game.Game` owns during one running simulation instance. It is a runtime data container, not a persistence layer and not a packet schema source.

## Overview

`runtime.EntityStore` is the shared container for the active gameplay entities that the `Game` aggregate reads and mutates under its lock.

The current shape is:

```go
type EntityStore struct {
    Players     map[string]*Ship
    Projectiles map[string]*Bullet
    Asteroids   map[string]*Asteroid
    Enemies     map[string]*Ship
    Pickups     map[string]*pickups.Pickup
}
```

`runtime.NewEntityStore()` initializes each map eagerly so a fresh game instance starts with empty live-entity collections.

`game.Game` owns the store instance itself through its `entities` field. The `runtime` package owns the data shapes stored in the maps. Same-package game code uses the store to keep authoritative match state in memory while a simulation is running.

## Code root

`services/game-server/internal/game/runtime/`

## Responsibilities

The runtime entity store owns:

* The in-memory maps for `Players`, `Projectiles`, `Asteroids`, `Enemies`, and `Pickups`.
* Empty-map initialization through `runtime.NewEntityStore()`.
* Grouping live entity collections under one aggregate-owned store value.
* Supporting same-package mutation by simulation, combat, pickup, and state-projection code.
* Representing match-local entity presence rather than durable records.

## Does not own

The runtime entity store does not own:

* `Game` lifecycle, room lifecycle, or simulation tick cadence.
* Player session durability, score, lives, respawn state, or other durable counters.
* Packet schema source-of-truth files.
* Network encoding or websocket write loops.
* Client rendering, interpolation, HUD, or presentation effects.
* Persistence or account/profile storage.
* Collision policy, combat policy, spawn policy, or scoring policy.
* Domain-event creation or presentation-event queue semantics.

Those responsibilities live in the game aggregate, neighboring runtime docs, or other game-server service boundaries.

## Domain roles

The entity store participates in the server-authoritative gameplay domain as the match-local runtime backing store for live entities.

Rooms and networking do not own the store. They create, start, stop, and observe `Game` instances that contain the store. During a match, simulation code mutates the store and state packet projection reads from it to build per-player snapshots.

The store exists only while a game instance exists. It is not a durable domain object and does not outlive the match-local simulation aggregate.

## Protocols and APIs

The runtime entity store is not a protocol surface.

It is consumed internally by `game.Game`, same-package simulation helpers, and state packet projection. The primary API that creates it is `runtime.NewEntityStore()`, which returns a fully initialized empty store for `Game.New()`.

There is no external network contract attached to the store itself. Outbound packets observe copies of its contents through `Game.StatePacket(playerID)`, but the store does not define packet schema or transport behavior.

## Data ownership

`runtime.EntityStore` is owned by the game aggregate as live in-memory state.

The data ownership split is:

* `game.Game` owns the `entities` field and the lock around it.
* `runtime` owns the `EntityStore`, `Ship`, `Bullet`, `Asteroid`, and related state-shape types.
* Simulation and gameplay helpers mutate the maps through the aggregate boundary.
* `state_packet.go` reads the maps and projects copy-out packet state.

The store holds authoritative live entity references for the current match only:

* `Players` for active player ships.
* `Projectiles` for live bullets and other projectile entities.
* `Asteroids` for active asteroids.
* `Enemies` for non-player hostile ships.
* `Pickups` for live pickups.

Because these are runtime maps, the store is intentionally mutable and in-memory. It is not a source of persistence, replay, or historical audit data.

## Code map

Core runtime files:

```text
services/game-server/internal/game/runtime/state.go
services/game-server/internal/game/game.go
services/game-server/internal/game/state_packet.go
```

Aggregate and simulation files that create, read, or mutate entity maps:

```text
services/game-server/internal/game/players.go
services/game-server/internal/game/simulation_players.go
services/game-server/internal/game/simulation_asteroids.go
services/game-server/internal/game/simulation_bullets.go
services/game-server/internal/game/pickups.go
services/game-server/internal/game/pickup_lifecycle.go
services/game-server/internal/game/pickup_collisions.go
services/game-server/internal/game/combat.go
services/game-server/internal/game/asteroid_destruction.go
services/game-server/internal/game/radial_spawning.go
services/game-server/internal/game/simulation_radial_effects.go
services/game-server/internal/game/player_world_state.go
```

Supporting runtime references:

```text
services/game-server/internal/game/runtime/packets_generated.go
services/game-server/internal/game/packets.go
```

## Tests

Representative tests that cover live-entity behavior and state projection:

```text
services/game-server/tests/game/state_packet_lifecycle_test.go
services/game-server/tests/game/movement_test.go
services/game-server/tests/game/collision_test.go
services/game-server/tests/game/pickups_test.go
services/game-server/tests/game/player_counters_test.go
services/game-server/tests/game/visibility_test.go
```

Package-level tests may also exercise entity-map mutation indirectly through simulation, combat, and projection paths.

## Related docs

- [Game Server Simulation Runtime](./!INDEX.md)
- [Game Aggregate](game-aggregate.md)
- [Simulation Loop And Phase Order](simulation-loop-and-phase-order.md)
- [State Packet Projection](state-packet-projection.md)
- [Presentation Event Queue](presentation-event-queue.md)
- [Game Server Simulation Players](../players/!INDEX.md)
- [Game Server Simulation Combat](../combat/!INDEX.md)
- [Game Server Simulation Pickups](../pickups/!INDEX.md)
- [Game Server Simulation World](../world/!INDEX.md)

## Notes

The store is deliberately simple: one aggregate-owned container for the live entity maps used during a match.

This document stays focused on ownership and map semantics. It does not duplicate player-session, pickup, combat, or state-packet behavior documented elsewhere.
