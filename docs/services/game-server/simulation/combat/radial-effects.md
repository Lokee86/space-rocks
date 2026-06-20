# Radial Effects

Parent index: [Game Server Simulation Combat](./!INDEX.md)

## Purpose

This document describes the current game-server radial effects service boundary.

It covers the implemented `services/game-server/internal/game/effects/radial/` package, the game-owned adapters that spawn radial effects and build candidates, the step flow that produces radial hits, and the handoff from radial hits into pure damage resolution.

## Overview

Radial effects are a game-server simulation boundary for timing, zone progression, target filtering, and hit intent generation.

The radial package accepts a spawn request, creates a stored effect, advances its zones over time, tests eligible candidates, and returns hit intents. Game-owned combat code creates the effect from projectile impact metadata, builds candidates from runtime entities, converts radial hits into damage requests, resolves damage through `damage.ResolveSingle`, and applies the results back into gameplay state.

Current service flow:

```text
projectile impact metadata
-> game-owned radial spawn adapter
-> radial.NewEffect
-> radial effect store
-> stepRadialEffects
-> radial.Step
-> radial.Hit intents
-> radial_damage_requests.go
-> damage.ResolveSingle
-> game-owned damage application
-> gameplay consequences
```

## Code root

```text
services/game-server/internal/game/effects/radial/
```

Game-owned adapters and integration points live in:

```text
services/game-server/internal/game/radial_spawning.go
services/game-server/internal/game/radial_candidates.go
services/game-server/internal/game/simulation_radial_effects.go
services/game-server/internal/game/radial_damage_requests.go
services/game-server/internal/game/combat_damage_application.go
```

## Responsibilities

The radial effects boundary owns:

* Creating stored radial effects from spawn requests.
* Tracking active radial effects in the effect store.
* Building zone timing and zone coverage data from `Spec`.
* Filtering targets with `TargetFilter`.
* Advancing effect age and zone tick timing in `Step`.
* Producing `Hit` intents for eligible targets.
* Supporting annular-wave and expanding-fill coverage modes.
* Supporting simultaneous and sequential zone expiration modes.
* Carrying source, zone, target, and damage-spec data in the hit intent.
* Removing expired effects from the store.

The game-owned adapters own:

* Spawning radial effects from projectile impact metadata.
* Building candidates from runtime asteroids, enemies, players, projectiles, and pickups.
* Converting radial hits into damage requests.
* Resolving damage through `damage.ResolveSingle`.
* Applying resolved damage results to runtime entities.
* Handling the gameplay consequences that follow a destroyed asteroid or fatal player result.

## Does not own

Radial effects do not own:

* Runtime entity maps.
* Damage math.
* Collision detection.
* Weapon firing policy.
* Score mutation or score award policy.
* Packet encoding or outbound packet writes.
* Client rendering, interpolation, audio, or effects.
* Room or session lifecycle.

Those concerns belong to game-owned combat and runtime systems, or to networking/client boundaries.

## Domain roles

Radial effects participate in the authoritative combat domain by providing timed area-style hit generation around a projectile impact.

Their role is to keep the area timing and candidate filtering separate from damage math and from the game code that mutates runtime state after a hit.

Important boundaries:

* `radial` is the pure timing and hit-intent package.
* `Game` code is responsible for spawning the effect, building candidates, and consuming hits.
* `damage.ResolveSingle` remains the owner of damage math.
* Gameplay consequences are handled by the game layer, not by the radial package.

## Runtime flow

`Game.Step` calls `stepRadialEffects` during the simulation step for active matches.

Current flow:

```text
stepRadialEffects
-> collect radial candidates from runtime entities
-> iterate active radial effects
-> radial.Step(effect, delta, candidates)
-> apply each radial hit
-> resolve single-target damage
-> apply damage to runtime entity
-> record damage-applied event when useful
-> run destroyed/fatal game-owned consequences
-> remove expired effects
```

Radial effects are spawned from projectile impact metadata in the game layer:

```text
bullet impact
-> spawnRadialEffectFromBullet
-> radial.NewEffect
-> radialEffects.Add
```

Candidate building is also game-owned:

```text
runtime asteroids
runtime enemies
runtime players
runtime projectiles
runtime pickups
-> radialCandidates()
```

`radial.Step` is the package-level decision point that advances the effect and emits hit intents for allowed candidates in range.

## Data ownership

### Radial package-owned data

The pure radial package owns these data shapes:

* `Spec`
* `CoverageMode`
* `ExpirationMode`
* `TargetFilter`
* `TargetKind`
* `Candidate`
* `Hit`
* `StepResult`
* `Zone`
* `Effect`
* `SpawnRequest`
* `Store`

### Hit intent ownership

`Hit` is the package's current hit intent shape.

It carries:

* effect ID
* source ID
* source player ID
* zone index when applicable
* target ID
* target kind
* target position
* damage spec

The hit intent is data-only. The game layer decides how to respond to the hit.

### Effect ownership

`Effect` stores:

* effect identity
* source identity
* source player identity
* origin
* spec
* age
* zones

`Store` owns the active effect map and provides add, remove, enumerate, and length operations.

## Protocol and API surfaces

Radial effects are not a network protocol boundary.

They are an internal Go API boundary consumed by game-server combat and simulation code.

Public radial entry points currently include:

```text
radial.NewEffect(SpawnRequest) Effect
radial.Step(*Effect, delta, []Candidate) StepResult
radial.NewStore() Store
Store.Add(effect)
Store.All()
Store.Remove(id)
Store.Len()
```

The game layer calls:

```text
spawnRadialEffectFromBullet
radialCandidates
stepRadialEffects
applyRadialHit
radialDamageRequestFromHitAndAsteroid
radialDamageRequestFromHitAndEnemy
radialDamageRequestFromHitAndPlayer
```

The radial package does not know about packets, networking sessions, or client presentation.

## Code map

### Pure radial package

```text
services/game-server/internal/game/effects/radial/effect.go
```

Defines `SpawnRequest`, `Effect`, and `NewEffect`.

```text
services/game-server/internal/game/effects/radial/step.go
```

Defines `Step` and the current hit-intent emission flow.

```text
services/game-server/internal/game/effects/radial/zone.go
```

Defines `Zone` and zone construction.

```text
services/game-server/internal/game/effects/radial/spec.go
```

Defines `Spec`.

```text
services/game-server/internal/game/effects/radial/targets.go
```

Defines `TargetKind` and `TargetFilter`.

```text
services/game-server/internal/game/effects/radial/hit.go
```

Defines `Candidate`, `Hit`, and `StepResult`.

```text
services/game-server/internal/game/effects/radial/store.go
```

Defines the active effect store.

```text
services/game-server/internal/game/effects/radial/coverage.go
```

Defines radial overlap helpers used by `Step`.

### Game-owned adapters

```text
services/game-server/internal/game/radial_spawning.go
```

Spawns radial effects from projectile impact metadata and stores them.

```text
services/game-server/internal/game/radial_candidates.go
```

Builds radial candidates from runtime entities.

```text
services/game-server/internal/game/simulation_radial_effects.go
```

Steps active radial effects, applies hits, and removes expired effects.

```text
services/game-server/internal/game/radial_damage_requests.go
```

Converts radial hits into damage requests.

```text
services/game-server/internal/game/combat_damage_application.go
```

Applies resolved damage results to runtime entities.

## Tests and verification

Current package coverage includes:

* zone construction
* coverage and overlap behavior
* target filtering
* store add/remove behavior
* step behavior for annular and expanding-fill coverage modes
* projectile-impact spawning behavior

Relevant test files live under:

```text
services/game-server/internal/game/effects/radial/*_test.go
services/game-server/internal/game/radial_*_test.go
services/game-server/internal/game/radial_projectile_impact_test.go
```

This documentation change does not require generated-file updates.

## Related docs

* [Game Server Simulation Combat](./!INDEX.md)
* [Damage Resolution](damage-resolution.md)
* [Collision To Damage Flow](collision-to-damage-flow.md)
* [Weapons And Projectile Fire](weapons-and-projectile-fire.md)
* [Game Server Simulation](../!INDEX.md)
* [Game Server Simulation Scoring](../scoring/!INDEX.md)
* [Game Server](../../!INDEX.md)

## Notes

Radial effects are a timing-and-hit-intent boundary. They do not own the damage calculation itself, and they do not directly own the downstream gameplay consequences that follow from a resolved hit.
