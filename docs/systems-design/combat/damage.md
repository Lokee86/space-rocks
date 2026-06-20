# Damage

Parent index: [Combat](./!README.md)

## Purpose

This document describes the conceptual damage model for Space Rocks combat.

It defines the authority rules, resolver boundary, outcome model, and invariants that should remain stable even as weapons, radial effects, enemies, shields, status effects, and presentation evolve.

## Overview

Damage is server-authoritative combat resolution.

The core model is:

```text
source + target snapshot + damage spec
-> pure damage resolution
-> damage result
-> game-owned application
-> gameplay consequences
-> optional presentation events
```

Damage answers one question:

```text
Given this source, this target snapshot, this damage intent, and these modifiers, what is the resolved outcome?
```

It does not decide why the target was eligible, how the hit was detected, how the result is applied to runtime storage, whether score is awarded, whether pickups drop, whether fragments spawn, whether a player respawns, or how the client presents the result.

Those decisions belong to the game-server combat, player, scoring, spawning, pickup, radial-effect, event, and client presentation boundaries.

## Conceptual model

Damage resolution is built from four core concepts:

```text
DamageSource
= the entity identity, entity type, and cause behind the damage

DamageTarget
= a caller-provided snapshot of the target's current damageable state

DamageSpec
= the amount and kind of damage being attempted

DamageResult
= the resolved outcome returned by the damage model
```

The target is a snapshot, not a live entity handle. The resolver receives health, shields, and modifiers from the caller, calculates an outcome, and returns the remaining health and shields in the result.

The game layer decides whether to write those remaining values back to runtime entities.

## Damage type and damage cause

Damage type and damage cause are separate concepts.

```text
DamageType
= what flavor of damage this is

DamageCause
= how or why the damage happened
```

Current damage types include:

```text
kinetic
explosive
energy
thermal
radioactive
true_damage
```

Current damage causes include:

```text
collision
projectile
debug
area
dot
```

A projectile can carry explosive damage. A radial hit can apply area-caused explosive damage. A collision can apply kinetic damage. These concepts should not be collapsed into one field.

## Authority rules

The server owns damage authority.

The client may render damage-related effects, HUD changes, death presentation, or future hit feedback, but it does not calculate authoritative damage, shields, death, score, respawn, asteroid destruction, pickup drops, or match outcomes.

The damage resolver owns pure calculation only.

The game-server simulation owns:

* collision facts
* radial hit consumption
* damage request construction
* runtime health and shield mutation
* fatal player consequences
* asteroid destruction consequences
* score application
* pickup drop evaluation
* fragment spawning
* presentation event recording

Devtools must route through the same damage model as live combat. Debug kill or invincibility behavior may use different source/cause data, but it must not create a parallel damage resolver or duplicate damage math.

## Resolution lifecycle

Current direct combat flow:

```text
collision fact
-> game-owned damage request adapter
-> damage resolver
-> game-owned result application
-> game-owned consequences
```

Current radial flow:

```text
projectile impact metadata
-> radial effect spawn
-> radial effect step
-> hit intent
-> game-owned radial damage request adapter
-> damage resolver
-> game-owned result application
-> game-owned consequences
```

Current devtools kill flow:

```text
debug command
-> game-owned debug damage request
-> damage resolver
-> game-owned result application
-> fatal player handling when needed
```

The resolver is shared. The entry paths differ.

## Result model

A damage result preserves enough information for the game layer to apply consequences and for later presentation mapping.

Conceptually, a result contains:

* source identity
* target identity
* base amount
* modified amount
* damage type
* damage cause
* applied modifiers
* health damage
* shield absorption
* ignored state
* destroyed state
* fatal state
* remaining health
* remaining shield
* created damage-over-time effects when requested

A result is not a packet.

A result is not itself a domain event.

A result becomes runtime state only when game-owned code applies it.

A result becomes presentation only when game-owned event code maps it into an event that the client later receives.

## Ignored damage

Damage can resolve to an ignored result.

Current ignored cases include:

* the modified damage amount is zero or negative
* the target already has zero or negative health

Ignored results are not applied to runtime health or shields.

Ignored results should not produce damage presentation events.

## Shield and health model

Damage modifiers are resolved before shield handling.

If shield bypass is not requested, shields absorb modified damage first. Any remaining damage applies to health.

If shield bypass is requested, modified damage applies directly to health and shields remain unchanged.

Destroyed means the target's remaining health reached zero.

Fatal is a narrower result. Current fatal damage means a player target was destroyed. Asteroids, enemies, and other target types can be destroyed without producing a fatal player result.

## Modifier model

Damage modifiers are filtered by damage type.

A modifier with no damage type applies globally.

A modifier with a matching damage type applies to that type.

A modifier with a different non-empty damage type does not apply.

Current modifier categories are:

```text
outgoing
resistance
vulnerability
generic
```

Current modifier operations are:

```text
add
multiply
set
```

The current modifier order is intentional:

```text
base amount
-> add modifiers
-> outgoing/generic multiply modifiers
-> resistance modifiers
-> vulnerability modifiers
-> set modifiers
-> clamp below zero
-> round to integer
```

Resistance values represent damage reduction. A resistance value of `0.25` means the damage keeps `75%` of the current amount.

Vulnerability values represent damage multipliers. A vulnerability value of `1.25` means the damage becomes `125%` of the current amount.

Invalid resistance or vulnerability modifiers are ignored.

## Area and radial damage

Area damage and radial effects are related but not the same boundary.

Area damage is a pure damage helper over caller-supplied candidates. The resolver does not inspect world maps, collision shapes, range queries, entity stores, or target eligibility. The caller supplies the candidate targets.

Radial effects own timed coverage and hit-intent generation. They decide which candidate targets are hit by an active radial effect. They do not own damage math.

The game layer converts radial hits into damage requests and sends those requests through the same resolver as direct projectile and collision damage.

## Damage over time

Damage over time is modeled as repeated damage resolution.

The current damage model can create active damage-over-time effects from a damage spec and can resolve ticks through the same single-target resolver using the `dot` cause.

This is not a complete status-effect system. It is the damage side of future over-time behavior. Broader status effects should build around this seam rather than moving damage math into a generic status package prematurely.

## Combat consequences

Damage resolution stops at the result.

Projectile damage against an asteroid may later cause:

* asteroid health mutation
* projectile pending despawn
* asteroid pending despawn
* score award evaluation
* asteroid fragment spawning
* pickup drop evaluation
* damage and blast presentation events

Collision damage against a player may later cause:

* player health and shield mutation
* player pending despawn
* camera-view preservation at death position
* ship death count increment
* lives decrement
* respawn cooldown setup
* ship death presentation event
* match lifecycle changes through separate rules

Radial damage may reuse those same consequence paths after a hit has been converted into a damage request and resolved.

None of those consequence paths belong inside the damage resolver.

## Events and presentation

Damage presentation is event-driven.

A `damage_applied` event is recorded only when a resolved result is not ignored and actually changes health or shields.

Damage-over-time event mapping exists for started and tick events, but active DoT gameplay ownership is not fully wired as a complete gameplay system.

The packet-facing damage event projection currently carries a reduced presentation shape. It does not expose every field from the damage result, such as full target identity, cause, base amount, shield absorption, or remaining health.

The client consumes server events as presentation facts. It should not infer authoritative damage outcomes from visual effects.

## Participating systems

The damage model participates with these systems:

* [Game Server Simulation Combat](../../services/game-server/simulation/combat/!README.md) owns the implementation-side combat docs.
* [Damage Resolution](../../services/game-server/simulation/combat/damage-resolution.md) owns the game-server resolver implementation boundary.
* [Collision To Damage Flow](../../services/game-server/simulation/combat/collision-to-damage-flow.md) owns collision-driven damage application and consequences.
* [Radial Effects](../../services/game-server/simulation/combat/radial-effects.md) owns radial timing, coverage, and hit-intent behavior.
* [Weapons And Projectile Fire](../../services/game-server/simulation/combat/weapons-and-projectile-fire.md) owns weapon profiles, projectile spawn intent, and damage intent carried by projectiles.
* [Presentation Event Queue](../../services/game-server/simulation/runtime/presentation-event-queue.md) owns server event projection into client-visible event packets.
* [Gameplay Events And Effects](../../services/client/gameplay-event-presentation/gameplay-events-and-effects.md) owns client-side presentation of supported server events.

## Invariants

Damage must preserve these invariants:

* Damage is server-authoritative.
* Damage resolution is pure calculation.
* The resolver does not mutate runtime entities.
* Runtime entities are adapted into damage targets by the caller.
* Damage results are applied by game-owned code.
* Damage type and damage cause remain separate concepts.
* Collision detection does not perform damage math.
* Weapons carry damage intent but do not resolve impact damage.
* Radial effects produce hit intents but do not resolve damage.
* Devtools damage routes through the same damage resolver as live combat.
* Client presentation observes server results and does not recalculate them.
* Fatal player damage consequences stay outside the damage resolver.
* Asteroid destruction consequences stay outside the damage resolver.
* Score, pickup drops, fragments, respawn, packets, and audio are downstream consequences, not damage-resolver responsibilities.
* Ignored damage must not mutate health, shields, or presentation state.
* Pending-respawn and eliminated players must not become damageable merely because their identity is still present in session read models.

## Active issues

Current combat limits for damage and related effects are tracked in [Current System Limits](../../limits/current-system-limits.md#combat-systems).

Relevant current limits include incomplete client rendering for damage events, incomplete active DoT gameplay ownership, non-implemented presentation concepts such as shield absorption feedback, and incomplete enemy death consequences.

Future damage and effect presentation work is tracked in [Domain Backlog](../../planning/domain-backlog.md#damage).

## Related docs

* [Combat](./!README.md)
* [Game Server Simulation Combat](../../services/game-server/simulation/combat/!README.md)
* [Damage Resolution](../../services/game-server/simulation/combat/damage-resolution.md)
* [Collision To Damage Flow](../../services/game-server/simulation/combat/collision-to-damage-flow.md)
* [Radial Effects](../../services/game-server/simulation/combat/radial-effects.md)
* [Weapons And Projectile Fire](../../services/game-server/simulation/combat/weapons-and-projectile-fire.md)
* [Presentation Event Queue](../../services/game-server/simulation/runtime/presentation-event-queue.md)
* [Player Death And Despawn](../../services/game-server/simulation/players/player-death-and-despawn.md)
* [Player Counters](../../services/game-server/simulation/players/player-counters.md)
* [Gameplay Events And Effects](../../services/client/gameplay-event-presentation/gameplay-events-and-effects.md)
* [Realtime Client Server Flow](../../domains/technical/realtime-client-server-flow.md)
* [Gameplay Session Flow](../../domains/player-experience/gameplay-session-flow.md)
* [Current System Limits](../../limits/current-system-limits.md#combat-systems)
* [Domain Backlog](../../planning/domain-backlog.md#damage)

## Notes

`DamageResult` is deliberately rich enough to support both runtime application and later presentation, but it should not become a catch-all event or telemetry boundary.

The current model already has seams for shields, modifiers, area damage, enemy targets, and damage over time. Future work should extend those seams rather than moving damage ownership into weapons, radial effects, client presentation, or devtools.

Area falloff, richer status effects, PvP/team damage rules, and detailed damage telemetry are future behavior. They should preserve the same authority split: eligibility and target selection outside the resolver, math inside the resolver, consequences after the resolver.
