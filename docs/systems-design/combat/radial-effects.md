# Radial Effects

Parent index: [Combat](./!INDEX.md)

## Purpose

This document describes the systems-design model for radial effects in Space Rocks combat.

It defines the conceptual ownership, authority rules, and invariants for timed radial or area-style effects without duplicating the game-server implementation documentation.

## Overview

Radial effects are server-authoritative combat effects that apply timed coverage around an origin point.

A radial effect is not just an animation and not just damage math. It is a combat model for turning an authoritative cause, usually a projectile impact, into one or more timed hit intents against eligible entities inside radial coverage.

Current implemented use:

```text
torpedo projectile impact
-> radial effect spawned at impact position
-> radial zones advance over server simulation time
-> eligible targets inside active coverage produce hit intents
-> hit intents become damage requests
-> damage resolution returns results
-> game-owned combat code applies consequences
-> client receives presentation events
```

The current torpedo radial bloom uses annular-wave coverage, simultaneous expiration, explosive area damage, and an asteroid/enemy target filter. Direct torpedo impact damage and torpedo radial bloom damage are separate combat facts.

## Conceptual model

A radial effect has:

```text
origin
source identity
source player identity
coverage mode
expiration mode
target filter
zone count
zone width
zone spawn timing
tick timing
total duration
zone lifetime
damage intent
```

The effect origin is the server-space point where coverage starts. For torpedoes, that origin is the projectile impact position.

The effect does not immediately mutate runtime entities. It creates active coverage over time. Candidates inside active coverage can produce hit intents. Those hit intents are then adapted into normal damage resolution.

Radial effects currently support two coverage concepts:

```text
annular wave
= one or more radial bands expanding outward from the origin

expanding fill
= a filled radius that grows outward from the origin
```

They also support two expiration concepts:

```text
simultaneous expiration
= all zones expire at the total effect time

sequential expiration
= each zone expires based on its own start time plus zone lifetime
```

Current torpedo behavior uses annular wave plus simultaneous expiration. The specific torpedo values are tuning data, not permanent systems-design constants.

## Authority rules

The server owns radial effect authority.

The client may present a radial effect start event, but it does not decide that the effect exists, which entities are hit, how much damage is applied, or whether any target is destroyed.

The authoritative ownership split is:

```text
weapon profile
-> may declare projectile impact effect metadata

game-owned projectile impact handling
-> decides when impact metadata spawns a radial effect

radial effect model
-> owns effect timing, coverage, target filtering, and hit-intent generation

damage resolver
-> owns damage math

game-owned combat application
-> owns runtime mutation and consequences

presentation event flow
-> owns client-visible explosion/effect presentation
```

The radial effect model must not own runtime entity maps, scoring, asteroid fragmentation, pickup drops, player death handling, room lifecycle, packet transport, or client rendering.

## Invariants

Radial effects preserve these combat invariants:

* Radial effects are server-authoritative.
* Client radial visuals are presentation only.
* A radial effect emits hit intents, not direct gameplay mutations.
* Damage math remains owned by the damage-resolution seam.
* Runtime consequences remain owned by game-server combat application code.
* Target inclusion must be explicit through the target filter.
* A target kind that is not enabled by the filter must not be hit by that radial effect.
* Direct projectile impact damage and radial bloom damage are separate facts.
* Radial timing is based on simulation effect age, not render frames.
* Repeated radial ticks may apply repeated damage if a target remains in active coverage.
* Radial spatial checks must use world-space rules that respect toroidal wrapping.
* Candidate size may participate in coverage overlap; coverage is not limited to target center points.
* Radial effects must remain separate from devtools-only shortcuts or debug-only gameplay logic.

## Coverage and timing

Radial effects are zone-based.

A zone has:

```text
index
inner radius
outer radius
start time
expiration time
next tick time
```

For annular-wave effects, each active zone covers a ring between its inner and outer radius. A candidate overlaps the zone if the candidate body overlaps that radial band.

For expanding-fill effects, the current fill radius grows outward as zones become active. A candidate overlaps the effect if its body overlaps the filled radius.

Zone timing determines when coverage can produce hits:

```text
effect age below zone start
-> zone does not participate

effect age at or after zone start
-> zone can participate

effect age at or after zone expiration
-> zone no longer participates

effect age below next tick time
-> zone waits

effect age at or after next tick time
-> zone may emit hits, then advances next tick time
```

This model allows a radial effect to behave like a bloom, pulse, shockwave, blast ring, or filled expanding area without changing the downstream damage contract.

## Target model

Radial effects evaluate candidates, not raw runtime entities.

Conceptually, a candidate contains:

```text
id
target kind
position
radius
```

Current target kinds are:

```text
asteroid
enemy
player
projectile
pickup
```

The target filter decides which target kinds are eligible for a given effect.

Current torpedo radial filtering includes asteroids and enemies only. Players, projectiles, and pickups are valid model concepts, but they are not currently enabled for the torpedo radial effect.

Candidate construction belongs to the game-server runtime layer because the radial model should not know about runtime maps or concrete entity storage.

## Damage and consequence model

A radial hit is a combat intent.

It carries enough information for the game-server combat layer to build a normal damage request:

```text
effect identity
source identity
source player identity
zone index
target identity
target kind
target position
damage spec
```

Damage is then resolved through the normal single-target damage path.

For current radial damage, the damage source uses projectile identity and area cause. If a radial damage spec does not provide a cause, game-owned adaptation normalizes it to area damage.

Consequences remain outside the radial model:

```text
asteroid destroyed
-> game-owned asteroid destruction, scoring, fragments, and drops

enemy destroyed
-> enemy death consequences where wired

player fatal damage
-> game-owned player fatal damage flow
```

The radial model must not duplicate consequence handling.

## Presentation model

Radial effect presentation starts from a server event.

When a projectile impact spawns a radial effect, the server records a radial-effect-started presentation event. The client consumes that event from gameplay state and may spawn a local visual effect.

Current client behavior presents `radial_effect_started` as the torpedo explosion scene.

Presentation facts are intentionally weaker than combat facts:

```text
combat fact
= server spawned and stepped a radial effect

presentation fact
= client may show an explosion/effect at the server-provided position
```

The presentation event does not grant the client authority to calculate hits, apply damage, or infer combat outcomes locally.

## Participating systems

Radial effects participate in these systems:

* [Game Server Simulation Combat](../../services/game-server/simulation/combat/!INDEX.md) owns runtime combat integration.
* [Radial Effects](../../services/game-server/simulation/combat/radial-effects.md) documents the game-server service boundary and implementation flow.
* [Weapons And Projectile Fire](../../services/game-server/simulation/combat/weapons-and-projectile-fire.md) documents projectile impact metadata and the torpedo impact-effect source.
* [Collision To Damage Flow](../../services/game-server/simulation/combat/collision-to-damage-flow.md) documents how projectile impacts connect to damage and radial impact effects.
* [Damage Resolution](../../services/game-server/simulation/combat/damage-resolution.md) documents the damage seam used by radial hit adaptation.
* [Simulation Loop And Phase Order](../../services/game-server/simulation/runtime/simulation-loop-and-phase-order.md) documents when radial effects step during the server tick.
* [Gameplay Events And Effects](../../services/client/gameplay-event-presentation/gameplay-events-and-effects.md) documents client-side radial presentation.
* [Current System Limits](../../limits/current-system-limits.md) tracks current radial-effect limitations.

## Active issues

* Broader radial client visuals are not fully implemented as a complete radial design path. Current client presentation handles the torpedo explosion event.
* Torpedo radial targeting currently includes asteroids and enemies only.
* Radial knockback is not implemented.
* Radial status effects are not implemented.
* Enemy death consequences are not fully wired yet.

See [Current System Limits](../../limits/current-system-limits.md#combat-systems).

## Related docs

* [Combat](./!INDEX.md)
* [Game Server Simulation Combat](../../services/game-server/simulation/combat/!INDEX.md)
* [Radial Effects](../../services/game-server/simulation/combat/radial-effects.md)
* [Weapons And Projectile Fire](../../services/game-server/simulation/combat/weapons-and-projectile-fire.md)
* [Collision To Damage Flow](../../services/game-server/simulation/combat/collision-to-damage-flow.md)
* [Damage Resolution](../../services/game-server/simulation/combat/damage-resolution.md)
* [Gameplay Events And Effects](../../services/client/gameplay-event-presentation/gameplay-events-and-effects.md)
* [Current System Limits](../../limits/current-system-limits.md)

## Notes

The radial-effect concept is intentionally broader than the current torpedo bloom. The current implementation proves the boundary with torpedo impact effects, but the design should continue to support other radial combat effects without moving timing, filtering, or hit-intent ownership into weapons, damage resolution, client presentation, or devtools.

Torpedo radial zone count, width, timing, and damage are tuning data. They should not be treated as permanent combat-design invariants.
