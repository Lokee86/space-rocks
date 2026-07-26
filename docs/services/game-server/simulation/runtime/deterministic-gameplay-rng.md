---
author: brian
created: "2026-07-19"
document_id: 019f7d55-fb2c-7ba4-806e-ab5be9611826
document_type: general
policy_exempt: false
summary: This document describes the authoritative game-server runtime RNG seam used for gameplay randomness.
---
# Deterministic Gameplay RNG Runtime

Parent index: [Game Server Simulation Runtime](./!INDEX.md)

## Purpose

This document describes the authoritative game-server runtime RNG seam used for gameplay randomness.

It covers ownership of the single game-local random source, the supported construction paths, the current seeded call sites, the synchronization contract around random access, the current determinism evidence, and the limits of what this seam does not guarantee.

## Ownership

The `game.Game` aggregate owns deterministic gameplay randomness for one match instance.

Current ownership shape:

```text
one Game
-> one concrete rng.Source
-> shared with spawning.Spawner
-> accessed only while Game.mu or Control locking is held
```

The `Game` constructor creates one concrete `services/game-server/internal/game/rng.Source` and passes that same source instance to `spawning.New(source)`.

That is the only production gameplay RNG source for the aggregate. The source is not copied into separate production RNG objects for game logic and spawning; the same source is shared by reference so the gameplay call stream stays ordered and attributable to one match-local seed.

## Code root

```text
services/game-server/internal/game/
services/game-server/internal/game/rng/
services/game-server/internal/game/spawning/
services/game-server/internal/devtools/
```

## Construction and seeding

The runtime supports three seed-related construction paths:

### `New()` production path

`game.New()` builds a fresh aggregate with production seeding:

```text
New()
-> newGame(rng.NewProduction())
```

`rng.NewProduction()` seeds from `crypto/rand` and falls back to `time.Now().UnixNano()` if cryptographic randomness cannot be read.

Current implementation shape:

```text
binary.Read(crypto/rand.Reader, little-endian int64)
-> on failure, use time.Now().UnixNano()
-> call rng.New(seed)
```

### `NewWithSeed(seed)` deterministic path

`game.NewWithSeed(seed)` creates the same aggregate shape but installs a caller-provided seed.

This is the supported path for scripted scenarios, synthetic scenarios, and deterministic tests that need repeatable gameplay random choices.

### `SimulationSeed()` readback path

`game.SimulationSeed()` returns the seed stored inside the aggregate's concrete RNG source.

That makes the current simulation seed observable for diagnostics and test assertions without exposing the underlying `math/rand.Rand` directly.

## Synchronization and access contract

`rng.Source` has no internal mutex.

That is intentional: the source is treated as part of the `Game` aggregate's synchronized state, not as a standalone concurrent service.

Current access rule:

```text
random reads and writes happen only while Game.mu is held
or through Control methods that lock Game.mu first
```

This is why random access appears in both game-owned helpers and `Control` methods, but not as a lock-free shared utility.

Current locking entry points include:

- game-owned mutation helpers called under `Game.mu`
- `Control.RandomUnitVector()`
- `Control.RandomAsteroidSpeed()`
- `Control.PlanDebugAsteroidSpawn()`
- `Control.ApplyAsteroidSpawnPlan()`
- `Control.SpawnBullet()`
- `Control.SpawnDebugBullet()`

The lock discipline keeps the ordered RNG call stream aligned with authoritative mutation order.

## Architecture guard

The current architecture guard is:

- do not use process-global `math/rand` in authoritative game code
- do not use process-global `math/rand` in devtools production code that participates in authoritative gameplay
- use `services/game-server/internal/game/rng` for production gameplay randomness
- use test-local deterministic sources only inside tests and fixtures when building expected values

Pitlord rule `game-server-no-process-global-math-rand` enforces the production import restriction. This keeps the gameplay random stream anchored to the match seed instead of ambient process state.

## Current seeded call sites

The current seeded gameplay and devtools call sites are:

### Offscreen asteroid spawn position

`game.randomOffscreenPosition()` uses the shared game RNG to choose:

- which side of the camera rectangle to spawn from
- which coordinate along that side to use

That drives the current offscreen asteroid placement behavior.

### Timed asteroid spawn planning

`spawning.Spawner.PlanTimedAsteroidSpawn()` uses the shared source for:

- spawn size selection
- spawn-speed selection
- weighted timed-spawn asteroid variant selection
- random angular offset around the target direction through `randomRange`

### Fragment asteroid spawn planning

`spawning.Spawner.PlanAsteroidFragmentSpawns()` uses the shared source for:

- fragment direction selection
- fragment speed selection
- weighted fragment-spawn asteroid variant selection

### Debug asteroid spawn planning

`spawning.Spawner.PlanDebugAsteroidSpawn()` uses the shared source for:

- fallback direction when the requested direction is missing or zero
- weighted debug-spawn asteroid variant selection
- debug asteroid speed selection
- debug asteroid size selection

### Pickup drop rolls

`game.maybeDropPickupFromAsteroidLocked()` samples `game.rngSource.Float64()` into the roll values passed to the drop-table evaluator.

The drop-table package remains the pure evaluator; the game aggregate owns the seeded roll stream that feeds it.

### Debug bullet fallback direction

`services/game-server/internal/devtools/spawn_bullet.go` falls back to `target.RandomUnitVector()` when a debug bullet request does not provide a direction.

That fallback is still sourced from the game-owned RNG path because `Target.RandomUnitVector()` locks `Game.mu` and forwards to `game.spawner.RandomUnitVector()`.

## Pure asteroid variant resolvers

The asteroid variant selection helpers are pure functions.

Current helpers:

- `asteroids.TimedSpawnVariantIndex(roll)`
- `asteroids.FragmentSpawnVariantIndex(roll)`
- `asteroids.DebugSpawnVariantIndex(roll)`

They all receive a normalized roll value and compute the chosen variant index from the weighted catalog data.

Important properties:

- they do not touch RNG state
- they do not read game state
- they do not mutate anything
- they depend only on the supplied roll and the current weighted catalog data

This keeps random sampling separate from catalog resolution.

## Determinism evidence

Current tests and call sites show the following evidence:

- `internal/game/game_seed_test.go` verifies `NewWithSeed(seed)` preserves the seed and that `SimulationSeed()` matches the stored source seed.
- `internal/game/visibility_test.go` verifies offscreen asteroid positions follow the seeded RNG sequence.
- `internal/game/pickup_drops_test.go` verifies pickup drop evaluation stays deterministic for equal seeds.
- `internal/game/spawning/spawner_seed_test.go` verifies seeded spawner helpers produce matching random sequences for equal seeds.
- `internal/game/asteroids/variants_test.go` verifies weighted variant resolution is stable for known normalized rolls.

Current deterministic gameplay evidence means this:

```text
same build
+ same initial state
+ same ordered inputs/steps
+ same seed
-> reproduces the current gameplay random choices
```

This is useful for seeded scripted scenarios and repeatable integration tests.

## Determinism limits

This document does not claim full simulation determinism.

It is not:

- a replay system
- an input recorder
- a cross-build compatibility guarantee
- a guarantee against nondeterministic ordering elsewhere in the simulation
- a promise that every subsystem outside the RNG seam is deterministic

Examples of things outside this document's guarantee:

- unordered iteration in unrelated code paths
- different binaries or compiler outputs
- different initial state
- different event ordering
- other sources of nondeterminism that are not routed through the seeded gameplay RNG seam

The current guarantee is intentionally narrower: seeded gameplay randomness is reproducible when the surrounding simulation inputs and ordering are also reproduced.

## P3C use

For P3C-style scripted or synthetic scenarios, callers can inject a seed by using `NewWithSeed(seed)` or by otherwise wiring a deterministic seed into the match setup flow.

That makes seeded scenario runs reproducible without needing a replay system.

Logging or diagnostic bundle inclusion for these runs is deferred to observability integration and is not part of this RNG seam yet.

## Code map

### Game aggregate

- `services/game-server/internal/game/game.go`
  - owns the aggregate seed field through `rngSource`
  - constructs `rng.NewProduction()` in `New()`
  - constructs `rng.New(seed)` in `NewWithSeed(seed)`
  - passes the same source to `spawning.New(source)`
  - exposes `SimulationSeed()`

### RNG source

- `services/game-server/internal/game/rng/source.go`
  - stores the seed
  - wraps `math/rand.Rand`
  - provides `New(seed)`, `NewProduction()`, `Seed()`, `Intn()`, and `Float64()`

### Spawning

- `services/game-server/internal/game/spawning/spawner.go`
  - consumes the shared source for asteroid spawn size, speed, direction, and variant rolls
- `services/game-server/internal/game/spawning/debug_asteroid.go`
  - uses the shared source for fallback direction, debug size, and debug variant selection
- `services/game-server/internal/game/spawning.go`
  - uses the shared source for offscreen asteroid spawn position selection
- `services/game-server/internal/devtools/spawn_bullet.go`
  - uses `Control.RandomUnitVector()` as the fallback direction source for seeded debug bullet spawning

### Drops

- `services/game-server/internal/game/pickup_drops.go`
  - samples drop-table roll values from the shared source

### Devtools-facing control seams

- `services/game-server/internal/game/control_spawn.go`
- `services/game-server/internal/game/control_streams.go`

These files route production debug commands through the locked game-owned RNG path rather than process-global randomness, and they support the seeded debug bullet fallback flow documented above.

## Related docs

- [Game Aggregate](./game-aggregate.md)
- [Simulation Loop And Phase Order](./simulation-loop-and-phase-order.md)
- [Asteroid Spawning And Variants](../world/asteroid-spawning-and-variants.md)
- [Pickup Drop Integration](../pickups/pickup-drop-integration.md)
- [Visibility And Despawn](../world/visibility-and-despawn.md)

## Tests and verification

Current repo coverage that exercises this seam:

- `services/game-server/internal/game/game_seed_test.go`
- `services/game-server/internal/game/visibility_test.go`
- `services/game-server/internal/game/pickup_drops_test.go`
- `services/game-server/internal/game/spawning/spawner_seed_test.go`
- `services/game-server/internal/game/asteroids/variants_test.go`

This document is aligned to implemented behavior only. It does not describe any future observability bundle packaging or replay plumbing.

## Non-goals

This seam does not own:

- full simulation replay support
- authoritative determinism across different builds
- input capture or input playback
- process-global RNG state
- client-side randomness
- observability bundle assembly
- broad testing strategy outside the RNG seam
- non-RNG gameplay ownership such as scoring, collisions, physics, or lifecycle policy

Those belong to their respective simulation or observability boundaries.
