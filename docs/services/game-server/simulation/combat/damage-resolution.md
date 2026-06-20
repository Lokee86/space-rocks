# Damage Resolution

Parent index: [Game Server Simulation Combat](./!README.md)

## Purpose

This document describes the current game-server damage resolution service boundary.

It covers the pure resolver package in `services/game-server/internal/game/damage/`, the request and result shapes it operates on, the implemented shield and modifier behavior, and the game-owned adapters that build requests and apply results back into runtime state.

## Overview

Damage resolution is a pure game-server service boundary.

The resolver package accepts immutable request data, calculates damage outcomes, and returns a result. It does not mutate runtime entities, scoring state, packets, pickups, radial timing, or client presentation. Game-owned combat code builds the request objects, calls the resolver, applies the returned result, and handles any gameplay consequences.

Current service flow:

```text
game-owned combat or radial adapter
-> build DamageResolutionRequest
-> damage.ResolveSingle
-> pure DamageResult
-> game-owned result application
-> gameplay consequences
```

The package also provides helper boundaries for area damage and damage-over-time timing. Those helpers stay inside the same resolver package and use the same pure request/result shapes.

## Code root

```text
services/game-server/internal/game/damage/
```

Game-owned adapters and application helpers live in:

```text
services/game-server/internal/game/combat_damage_requests.go
services/game-server/internal/game/radial_damage_requests.go
services/game-server/internal/game/combat_damage_application.go
```

## Responsibilities

The damage-resolution boundary owns:

* Resolving a single damage request through `ResolveSingle`.
* Resolving area damage through `ResolveArea`.
* Advancing active damage-over-time effects through `TickDamageOverTime`.
* Applying modifier filtering and modifier math in the pure resolver.
* Applying shield absorption when shield bypass is not requested.
* Producing destroyed and fatal flags from the resolved outcome.
* Producing ignored results for zero-or-negative modified damage or already-dead targets.
* Producing created damage-over-time effects when the spec enables DoT.
* Returning data-only results for the game layer to apply later.

The game-owned adapters own:

* Building projectile, collision, and radial damage requests.
* Supplying the source, target, spec, and modifier data to the resolver.
* Applying `DamageResult` back into runtime asteroid, player, and enemy state.
* Handling the gameplay consequences that follow from a resolved result.

## Does not own

Damage resolution does not own:

* Runtime entity mutation.
* Score mutation or score award policy.
* Packet encoding or outbound packet writes.
* Pickup spawning or pickup collection rules.
* Radial effect scheduling or timing ownership.
* Client rendering, interpolation, audio, or effects.
* Collision detection.
* Weapon firing policy.
* Room, session, or networking lifecycle.

Those concerns belong to game-owned combat flow, simulation systems, or client/networking boundaries.

## Domain roles

Damage resolution participates in the authoritative combat domain by converting a source, target, and damage spec into a deterministic outcome.

Its role is to keep the actual math and resolution rules separate from the combat code that observes collisions and from the game code that mutates runtime state afterward.

Important boundaries:

* `damage` is a pure resolver package.
* `Game` code is responsible for building requests and applying results.
* `DamageResult` is authoritative for the outcome of the resolution step.
* Gameplay consequences are handled by the game layer, not by the resolver.

## Data ownership

### Resolver-owned data

The pure resolver package owns these data shapes:

* `DamageSource`
* `DamageTarget`
* `DamageSpec`
* `DamageResolutionRequest`
* `DamageResult`
* `DamageModifier`
* `DamageModifierCategory`
* `DamageModifierOperation`
* `AppliedDamageModifier`
* `ModifiedDamageAmount`
* `DamageOverTimeSpec`
* `ActiveDamageOverTime`
* `DamageOverTimeTickResult`
* `DamageTargetRef`
* `AreaDamageRequest`
* `AreaDamageResult`

### Request ownership

`DamageResolutionRequest` carries:

* `Source`
* `Target`
* `Spec`
* `Modifiers`

`DamageTarget` carries the target's current health, shield, and modifiers into the pure resolver.

`DamageSpec` carries the damage amount, type, cause, shield-bypass flag, and optional DoT spec.

### Result ownership

`DamageResult` carries the pure outcome of resolution, including:

* source and target entity IDs and entity types
* base and modified amounts
* damage type and cause
* applied modifiers
* health damage applied
* shield absorbed
* ignored state
* destroyed state
* fatal state
* remaining health
* remaining shield
* created damage-over-time effects
* optional reason text

The result is data-only until the game layer applies it.

## Protocol and API surfaces

Damage resolution is not a network protocol boundary.

It is an internal Go API boundary used by game-server combat and radial-effect code.

Public resolver entry points currently include:

```text
ResolveSingle(DamageResolutionRequest) DamageResult
ResolveArea(AreaDamageRequest) AreaDamageResult
TickDamageOverTime(ActiveDamageOverTime, DamageTarget, delta) DamageOverTimeTickResult
```

The game layer builds requests in:

```text
combat_damage_requests.go
radial_damage_requests.go
```

The game layer applies results in:

```text
combat_damage_application.go
```

### Request and result behavior

`ResolveSingle` combines request-level modifiers with target-level modifiers, resolves the modified amount, applies shield handling, and derives the final flags and remaining values.

Current resolution behavior includes:

* merged request and target modifiers
* filtered and validated modifier handling
* shield absorption unless `BypassShield` is true
* remaining health and remaining shield calculation
* destroyed detection when remaining health reaches zero
* fatal detection when a player target is destroyed
* ignored results when modified damage is zero or negative
* ignored results when the target is already dead
* optional creation of DoT effects when the spec enables them

### Area helper behavior

`ResolveArea` is a pure helper that runs `ResolveSingle` for each candidate target when the radius is positive.

It returns one `DamageResult` per candidate and does not mutate runtime entities.

### DoT helper behavior

`TickDamageOverTime` advances a single active DoT effect for a delta.

It returns:

* the copied source and target refs
* tick timing state
* duration timing state
* one or more `DamageResult` values when ticks fire
* an expired flag when the effect has run out

When a tick fires, the helper resolves each tick through `ResolveSingle` with `DamageCauseDot`.

## Code map

### Pure resolver package

```text
services/game-server/internal/game/damage/resolve.go
```

Contains `ResolveSingle`.

```text
services/game-server/internal/game/damage/area.go
```

Contains `AreaDamageRequest`, `AreaDamageResult`, and `ResolveArea`.

```text
services/game-server/internal/game/damage/dot.go
```

Contains `DamageOverTimeSpec`, `DamageTargetRef`, `ActiveDamageOverTime`, `DamageOverTimeTickResult`, and `TickDamageOverTime`.

```text
services/game-server/internal/game/damage/request.go
```

Contains `DamageSource`, `DamageTarget`, `DamageSpec`, and `DamageResolutionRequest`.

```text
services/game-server/internal/game/damage/result.go
```

Contains `DamageResult`.

```text
services/game-server/internal/game/damage/modifiers.go
```

Contains modifier categories, operations, filtering, applied-modifier tracking, and modified-amount calculation.

### Game-owned adapters

```text
services/game-server/internal/game/combat_damage_requests.go
```

Builds single-target damage requests for projectile and collision combat.

```text
services/game-server/internal/game/radial_damage_requests.go
```

Builds single-target damage requests for radial damage application.

```text
services/game-server/internal/game/combat_damage_application.go
```

Applies resolved damage results to runtime asteroids, players, and enemies.

## Tests and verification

Current package tests cover:

* single-target resolution
* shield absorption and bypass behavior
* modifier filtering and modifier math
* ignored-result handling
* destroyed and fatal flags
* area helper behavior
* DoT creation and ticking

Relevant test files live under:

```text
services/game-server/internal/game/damage/*_test.go
```

This documentation change does not require a generated-file update.

## Related docs

* [Game Server Simulation Combat](./!README.md)
* [Collision To Damage Flow](collision-to-damage-flow.md)
* [Weapons And Projectile Fire](weapons-and-projectile-fire.md)
* [Radial Effects](radial-effects.md)
* [Game Server Simulation](../!README.md)
* [Game Server Simulation Scoring](../scoring/!README.md)
* [Game Server](../../!README.md)

## Notes

Damage math stays pure and data-only. The game layer owns when a result becomes a health change, a death event, a despawn, or another gameplay consequence.
