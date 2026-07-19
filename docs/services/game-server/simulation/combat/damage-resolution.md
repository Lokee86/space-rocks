---
author: brian
created: "2026-07-19"
document_id: 019f7d55-fb2c-7b96-873a-e8c5be32931b
document_type: general
policy_exempt: false
summary: This document describes the current game-server damage resolution service boundary.
---
# Damage Resolution

Parent index: [Game Server Simulation Combat](./!INDEX.md)

## Purpose

This document describes the current authoritative damage, healing, repair, shield, relationship, modifier, and damage-over-time resolution boundary.

## Overview

The `damage` package resolves immutable request data into a detailed data-only result. Game-owned adapters establish source/target relationships, call the resolver, apply returned health and shield state, emit events, and trigger death, destruction, awards, motion, or presentation consequences.

```text
gameplay source
-> DamageResolutionRequest
-> relationship and policy eligibility
-> modifier resolution
-> shield / health / restoration resolution
-> DamageResult
-> game-owned state application and consequences
```

The standard policy ID is `standard_damage_v1`. Player-versus-player damage is disabled by default. Relationship permissions normally allow enemies, neutrals, and destructibles; area damage may affect self, and authorized debug damage may also affect allies.

## Code root

```text
services/game-server/internal/game/damage/
```

Game adapters live in `services/game-server/internal/game/`.

## Responsibilities

The damage boundary owns:

- source and target vocabulary
- collision, projectile, debug, area, DoT, hazard, and scripted causes
- kinetic, explosive, energy, thermal, radioactive, and true-damage types
- relationship permissions and standard policy
- invulnerability and authorized dev/admin bypass rules
- damage, healing, repair, blocked, ineffective, and discarded-lethal results
- additive and multiplicative modifier application
- shield bypass and shield overflow policy
- health and shield restoration destinations
- destroyed and player-fatal outcomes
- detailed applied, absorbed, restored, remaining, and reason fields
- radial falloff inputs after target selection
- damage-over-time creation, stacking, replacement, refresh, limits, scheduling, and expiry
- deterministic DoT ordering and pause-aware ticking

## Does not own

This boundary does not own:

- collision detection or radial target selection
- runtime entity storage
- team assignment
- player death/lives transitions
- score, awards, or objectives
- knockback integration
- client effects or HUD
- room lifecycle

Those owners consume `DamageResult` and related normalized facts.

## Domain roles

`DamageSource` preserves responsible player and team IDs, original instigator, explicit relationship permissions, and invulnerability-bypass authority. This lets delayed, area, or transferred effects retain attribution without reconstructing ownership later.

`DamageTarget` supplies current and maximum health/shield state plus target modifiers. The resolver never reads live entities directly.

Positive damage normally reaches shield first. A spec can bypass shields. Overflow may pass through to health or be discarded. Restoration uses explicit health, shield, or both destinations and is clamped to target maxima.

An invulnerable target blocks ordinary damage. Bypass is accepted only for an authorized developer/admin source; a bare bypass flag is insufficient.

DoT effects support stack, replace, refresh, and limited-stack policies. Each accepted effect receives a deterministic runtime ID. Pause stops schedule advancement.

## Protocols and APIs

Important internal APIs include:

```go
func ResolveSingle(request DamageResolutionRequest) DamageResult
func ResolveArea(request AreaDamageRequest) AreaDamageResult
func NewDamageOverTimeRuntime() *DamageOverTimeRuntime
func (runtime *DamageOverTimeRuntime) Add(effect ActiveDamageOverTime) DamageOverTimeAddOutcome
func (runtime *DamageOverTimeRuntime) Step(delta float64, paused bool) []DamageOverTimeTick
func (runtime *DamageOverTimeRuntime) RemoveTarget(targetID string) int
```

Game integration builds requests for projectile, collision, radial, debug, and DoT sources. It publishes resolved events before handing lethal player outcomes to the lives runtime and destruction outcomes to the relevant entity owner.

## Data ownership

`DamageResolutionRequest` owns one immutable resolution input: source, target, spec, and request modifiers.

`DamageResult` is the complete output for that resolution and includes:

```text
result kind
base and modified amount
damage type and cause
applied modifiers
health damage and shield absorption
health and shield restoration
ignored or discarded state
destroyed and fatal state
remaining health and shield
created DoT effects
machine-readable reason
```

The game layer owns applying remaining values to the live entity. `DamageOverTimeRuntime` owns scheduled effects between ticks, not the target's health.

## Code map

```text
services/game-server/internal/game/damage/types.go
services/game-server/internal/game/damage/request.go
services/game-server/internal/game/damage/policy.go
services/game-server/internal/game/damage/eligibility.go
services/game-server/internal/game/damage/modifiers.go
services/game-server/internal/game/damage/resolve.go
services/game-server/internal/game/damage/result.go
services/game-server/internal/game/damage/dot.go
services/game-server/internal/game/damage/dot_runtime.go
services/game-server/internal/game/combat_damage_requests.go
services/game-server/internal/game/damage_resolution.go
services/game-server/internal/game/damage_events.go
services/game-server/internal/game/damage_over_time.go
services/game-server/internal/game/radial_damage_requests.go
services/game-server/internal/game/motion/collision_repulsion.go
```

## Tests

Tests cover relationship eligibility, PvP policy, invulnerability, shield bypass and overflow, healing and repair, modifier order, signed requests, lethal-result handling, DoT policies and timing, radial falloff, collision repulsion, event publication, and game integration.

```text
services/game-server/internal/game/damage/*_test.go
services/game-server/internal/game/damage_rules_integration_test.go
services/game-server/internal/game/damage_result_events_test.go
services/game-server/internal/game/effects/radial/falloff_test.go
services/game-server/internal/game/motion/collision_repulsion_test.go
```

## Related docs

- [Game Server Simulation Combat](./!INDEX.md)
- [Collision To Damage Flow](collision-to-damage-flow.md)
- [Radial Effects](radial-effects.md)
- [Lives, Participation, And Spawn](../players/lives-participation-and-spawn.md)
- [Awards And Counter Runtime](../scoring/awards-and-counter-runtime.md)
- [Damage And Healing Rules Planning](../../../../planning/domains/gameplay/damage-and-healing-rules.md)

## Notes

The damage package remains data-oriented. Runtime mutation, death, destruction, awards, knockback, and presentation stay downstream so new damage sources do not duplicate consequence logic.
