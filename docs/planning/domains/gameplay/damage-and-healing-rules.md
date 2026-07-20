---
author: brian
created: "2026-07-19"
document_id: 019f7d55-fb2c-7fdb-bba6-a62b01ca82d9
document_type: general
policy_exempt: false
summary: This doc is the authoritative P4 planning owner for damage and healing rules.
---
# Damage And Healing Rules
Parent index: [Gameplay Planning](./!INDEX.md)

## Purpose

This doc is the authoritative P4 planning owner for damage and healing rules.

It defines how authoritative damage and healing intent becomes an eligible, resolved result, how shields and health are changed, and which facts are handed to physics, effects, lives/death, awards, attribution, and presentation. It establishes the product model while leaving exact schema, type, and package choices to implementation planning.

The conceptual and implementation reference for the current damage model is [Damage](../../../systems-design/combat/damage.md). That document remains a reference for existing concepts and seams; this document owns the P4 product policy described here.

## Ownership Boundary

This doc owns planning for:

```text
damage and healing authority
source relationship permissions
eligibility and invulnerability policy boundaries
damage source and attribution facts
damage types, causes, and modifier semantics
shield and health resolution
healing and repair resolution
collision damage and survivor repulsion handoff
area, radial, annular, and falloff policy
damage-over-time stacking and death behavior
lethal-result ordering and discard behavior
hit, damage, healing, and blocked-result semantics
retained-source behavior across disconnect
standard damage policy selection
```

This doc does not own:

```text
collision detection or candidate generation
physics integration and movement simulation
weapon aiming, firing, or projectile lifetime
radial coverage timing or world queries
status/effect lifecycle, cleansing, or broad effect ownership
runtime health/shield storage
lives, death, elimination, or respawn transitions
score, awards, kill credit, or final attribution windows
match objectives or match-end policy
packet or storage schemas
client UI, audio, or visual presentation
```

The damage system separates three stages:

```text
eligibility
-> is this source allowed to affect this target under resolved rules?

pure resolution
-> what signed value, shield/health changes, modifiers, and result facts apply?

downstream consequences
-> what runtime, physics, lifecycle, award, attribution, effect, or presentation action follows?
```

Eligibility and downstream consequences must not be hidden inside pure damage math. The resolver receives an eligible damage or restoration intent and returns a result without mutating live entities.

## Settled Product Model

Damage remains server-authoritative. The server owns eligibility, authoritative request construction, resolution, application, hit facts, health/shield mutation, and downstream gameplay consequences. The client may present authoritative events but must not calculate or grant damage, healing, shields, health, death, or invulnerability outcomes.

The default named policy is:

```text
standard_damage_v1
```

Modes and other resolved rules select or override the applicable damage policy. They may change relationship permissions and other explicitly supported policy inputs, but consumers must not infer gameplay rules from a mode name or from visual presentation.

### Supported Sources

The initial product model supports these damage sources:

```text
weapons
enemy attacks
asteroid/player collisions
hazards
radial effects
annular effects
Damage over time (DoT)
scripted or campaign damage
dev/admin sources
```

Sources may target players, enemies, asteroids, other destructibles, or eligible environment targets. The source carries enough identity and cause information for the authoritative system to apply the selected policy without requiring a separate damage path for each content type.

Each damage source carries default relationship permissions for:

```text
self
allies
enemies
neutrals
destructibles/environment
```

Modes may override those permissions. Relationship eligibility is evaluated before resolution. A source may therefore be capable of resolving damage mathematically while a particular target is ineligible under the active source and mode rules.

Enemy friendly fire may be permitted when the source and mode rules allow it. Player same-team friendly fire remains prohibited. When PvP damage is enabled, player damage is inter-team only; FFA participants count as opposing sides for this purpose. PvP enablement remains separate from team structure.

### Source Attribution

Source attribution preserves the facts needed by immediate and delayed consequences:

```text
immediate source entity
responsible player
responsible team
original instigator
```

The immediate source is the entity that directly applies or schedules the effect. The responsible player and team identify ownership for ordinary credit and relationship evaluation. The original instigator is preserved through chained, reflected, radial, or delayed effects so an intermediate entity does not erase the originating cause.

Attribution facts are retained as source context. This doc does not decide kill credit, assists, awards, or the exact environmental attribution window. Those consumers use the retained facts and their own policy boundaries.

Existing projectiles, DoTs, hazards, and radial effects survive owner disconnect. They may continue to resolve under their retained source context and may still credit retained participant records where the awards and attribution rules allow it.

### Damage Types, Causes, And Modifiers

Existing damage types and causes remain the foundation:

```text
Damage types:
kinetic
explosive
energy
thermal
radioactive
true_damage

Damage causes:
collision
projectile
debug
area
dot
```

Damage type describes what flavor of damage is applied. Damage cause describes how or why it happened. The two fields remain separate. New source categories should map into these existing concepts or extend them through an explicit policy decision rather than collapsing them.

The existing modifier model remains authoritative and should be implemented broadly even where current content has no visible effect yet. Modifiers are filtered by damage type and support the existing categories and operations:

```text
categories: outgoing, resistance, vulnerability, generic
operations: add, multiply, set
```

The existing modifier categories, damage-type filtering, and relative calculation order remain the foundation. The planned signed pipeline must preserve sign and restoration semantics instead of clamping every negative value to zero:

```text
base signed amount
-> add modifiers
-> outgoing/generic multiply modifiers
-> resistance modifiers
-> vulnerability modifiers
-> set modifiers
-> preserve sign/restoration semantics
-> resolve and round the applicable magnitude
```

Resistance values reduce the amount remaining. Vulnerability values multiply the amount. Invalid resistance or vulnerability modifiers are ignored under the existing model. The same modifier foundation applies to healing and repair, with restoration-specific result semantics. The exact magnitude/sign algorithm remains an implementation-level decision; valid negative restoration must survive the pipeline rather than being discarded by a damage-only clamp.

### Signed Damage And Restoration

The resolver pipeline accepts signed values:

```text
positive value
-> damage

negative value
-> healing or repair
```

A source decides whether a negative value restores health, shields, or both. Restoration cannot exceed the target's maximum health or maximum shields. A restoration request that has no eligible destination or has no effective change produces a healing/repair result according to the selected event semantics rather than being represented as ordinary damage.

Healing uses the existing modifier model and calculation order, but healing and repair emit distinct result and event semantics from damage. Downstream systems must be able to distinguish damage applied, healing applied, repair applied, blocked/invulnerable hits, and other non-damaging authoritative hit facts without reconstructing meaning from a signed number alone.

### Shields, Health, And Overflow

Shields absorb first by default. Sources may request shield bypass.

For ordinary shield-interacting damage, the resolved amount is applied to shields first. Overflow policy supports:

```text
pass-through to health
shield-gated discard of overflow
```

A bypassing source does not interact with shields and applies its resolved amount directly to health. These policies are source and/or mode inputs, not client behavior.

Collision profiles may use extremely high ordinary shield-gated damage. There is no separate catastrophic-kill or special collision-kill mechanic. If the ordinary resolved amount is sufficient to destroy a target after the selected shield and overflow rules, the target is destroyed through the normal result path.

### Collisions And Survivor Repulsion

Collision detection supplies authoritative collision facts and candidate targets; it does not perform damage math. Collision damage is resolved through the shared damage pipeline.

When colliding entities survive, the physics owner applies the required repulsion or knockback consequence. Survivors must not remain overlapping after a collision damage result merely because their health stayed above zero. Damage rules provide the normalized consequence request or collision result facts; physics owns movement integration, separation, and the final motion state.

### Area, Radial, And Annular Effects

Existing radial and annular effects provide area coverage and falloff behavior. Their owners determine coverage timing and hit intent, then route eligible intents through the shared damage resolver. The resolver does not inspect world maps, collision shapes, range queries, or entity stores.

No line-of-sight mechanics are required for the P4 damage/healing model. Area effects may affect every eligible candidate covered by the existing radial or annular behavior, subject to source relationship permissions and per-target application rules.

### Damage Over Time

DoT is repeated damage resolution using retained source context and the existing `dot` cause. DoT stacks by default. A source may override the stacking behavior, including by replacing, refreshing, limiting, or otherwise explicitly selecting a different policy.

DoT does not persist through death. Death removes or invalidates active damage-over-time application for the destroyed target. Future removal, cleansing, and broader status/effect lifecycle belong to a broader status/effect owner and are not added to this damage owner.

### Blocked, Invulnerable, And Hit Facts

Blocked or invulnerable hits may still emit authoritative hit facts. A blocked result records that an authoritative source attempted an interaction and that eligibility or target state prevented ordinary application. It must not be treated as applied damage or healing.

Normal gameplay sources cannot bypass invulnerability. Explicit dev/admin sources may bypass invulnerability when the source is marked and authorized for that behavior. Dev/admin requests still route through the same eligibility, resolution, application, and event boundaries rather than creating a parallel damage system.

### Lethal Ordering And Environmental Attribution

Damage after a target's first lethal result is discarded. Once a target has received an authoritative lethal result in the relevant step, later damage or restoration requests for that target cannot alter that already-decided lethal outcome.

Mutual kills remain possible across distinct targets in the same authoritative step. Credit and result interpretation for those outcomes belong outside damage math, in the lives/death, awards, attribution, and match-outcome owners.

Recent player damage or forced movement may claim a later environmental death. Exact attribution windows and precedence belong to awards/attribution planning, not to the pure resolver.

## Result And Event Semantics

A damage or healing result preserves enough information for application and later event projection. Conceptually it includes:

```text
source attribution context
target identity
base signed amount
modified signed amount
damage type
damage cause
applied modifiers
eligibility/blocked state
shield bypass and overflow policy
shield absorption or restoration
health damage or restoration
remaining health and shields
effective-change state
destroyed/fatal state where applicable
DoT creation or stacking outcome where applicable
```

A result is not a packet and is not itself a domain event. Game-owned code applies the result to runtime state, then maps authoritative facts into distinct events for gameplay presentation, lifecycle, awards, or effects.

An ignored, blocked, invulnerable, overkill-discarded, or ineffective result must not mutate health or shields. Such results may still emit authoritative hit facts when the source/target interaction was observed and the event policy requires it. Applied damage, healing, and repair use distinct event semantics.

## System Handoffs

```text
source intent and target candidates
-> source attribution and relationship eligibility
-> shared signed damage/healing resolution
-> game-owned health/shield application
-> physics repulsion for surviving colliders
-> DoT/effect scheduling where applicable
-> lives/death and runtime consequences
-> awards/attribution and retained participant credit
-> authoritative presentation events
```

### Modes, Teams, And Rules

Modes select PvP enablement and may override source relationship permissions, shield/overflow policy, or other explicitly supported damage inputs. Teams supply normalized self, ally, enemy, neutral, and FFA opposing-side relationships. Damage owns the relationship evaluation for damage requests without becoming the team or mode authority.

### Combat Sources And Collision

Weapons, enemy attacks, hazards, projectiles, and scripted/campaign systems construct source intent. Collision systems provide collision facts. Radial and annular effect owners provide coverage and hit intent. These systems do not duplicate damage math or silently bypass relationship eligibility.

### Physics

Physics owns collision detection, movement integration, repulsion, and knockback. Damage returns the normalized collision consequence facts needed to request repulsion when colliders survive. Damage does not directly mutate velocity or position.

### Runtime Ship And Health State

Runtime entities own live health, shields, maximums, modifiers, and damageable state at runtime. The damage owner adapts snapshots into pure resolution and game-owned application updates runtime state. Durable player counters remain with player-session/lives owners rather than the runtime ship.

### Status And Effects

Damage owns DoT damage resolution and its default stacking input. A future broader status/effect owner owns effect lifecycle, removal, cleansing, and persistence policy beyond the explicit rule that DoT does not persist through death.

### Lives, Death, Awards, And Results

Lives/death consumes authoritative lethal results and decides death, elimination, respawn, and life transitions. Awards and attribution consume retained source facts, recent-damage/forced-movement facts, and their own configured windows. Match outcomes decide mutual-kill and final-result interpretation. Damage does not award score, choose a killer, spawn pickups, or decide match end.

### Multiplayer Lifecycle

Lifecycle owns connection, disconnect, removal, reconnect, and retained participant records. Existing effects may continue after owner disconnect using retained source attribution. Damage consumers must not erase historical participant facts merely because the source entity is gone.

### Player Experience And Presentation

The client presents server-authored damage, healing, repair, blocked-hit, shield, death, and effect events. It may collect input and show feedback, but it cannot infer or apply authoritative health, shield, damage, healing, or invulnerability state.

## Implementation Direction

The first implementation slice should proceed from the settled policy into small authoritative seams:

```text
1. Resolve standard_damage_v1 as the default named policy.
2. Normalize source attribution and default relationship permissions.
3. Evaluate source/target eligibility using teams, PvP rules, mode overrides, and invulnerability state.
4. Preserve the existing damage types, causes, modifier categories, operations, and calculation order.
5. Extend the shared pure resolver to signed damage, healing, and repair result semantics.
6. Apply shield-first, bypass, and explicit overflow behavior without a catastrophic-kill special case.
7. Route collision survivor repulsion/knockback to the physics owner.
8. Route radial, annular, hazard, projectile, scripted, enemy, and dev/admin requests through the same boundaries.
9. Preserve source attribution through chained and delayed effects and owner disconnect.
10. Apply default DoT stacking and stop DoT on death while leaving cleansing to a future effect owner.
11. Discard post-lethal damage for the target while allowing distinct-target mutual kills in one authoritative step.
12. Emit distinct authoritative result/event facts for applied, blocked, healing, repair, ineffective, and overkill-discarded outcomes.
13. Hand lethal and environmental-attribution candidates to lives/death and awards/attribution owners.
```

Implementation should keep eligibility, pure resolution, and consequences visibly separate. Exact schema/type/package choices remain implementation-level decisions. Existing direct, projectile, collision, radial, DoT, and devtools paths should converge on the shared resolver rather than growing parallel damage math.

## Testing Direction

Important future checks:

```text
damage and healing are server-authoritative
eligibility, pure resolution, and downstream consequences remain separate
standard_damage_v1 is the default named policy
weapons, enemy attacks, collisions, hazards, radial, annular, DoT, scripted, and dev/admin sources are supported
source defaults expose self, ally, enemy, neutral, and destructible/environment permissions
mode overrides are applied explicitly
enemy friendly fire follows source/mode policy
player same-team damage is always prohibited
PvP player damage is inter-team only
FFA participants oppose one another when PvP is enabled
immediate source, responsible player, responsible team, and original instigator survive chained/delayed attribution
owner disconnect does not automatically cancel existing projectiles, DoTs, hazards, or radial effects
existing damage types and causes remain distinct
modifier filtering, categories, operations, and calculation order remain stable
positive values damage and negative values heal/repair through the shared pipeline
sources select health, shields, or both as restoration targets
healing and repair never exceed maximums
healing emits distinct result/event semantics from damage
shields absorb first by default
shield bypass leaves shields unchanged
pass-through and shield-gated overflow are distinguishable
extremely high ordinary collision damage uses normal resolution
no catastrophic-kill mechanic exists
surviving colliders receive physics-owned repulsion/knockback
radial and annular effects preserve coverage/falloff behavior
line of sight is not required
DoT stacks by default and sources can override stacking
DoT does not persist through death
cleansing remains outside the damage owner
critical hits and random damage variation are unsupported
blocked/invulnerable hits can emit authoritative hit facts
normal sources cannot bypass invulnerability
explicit dev/admin sources may bypass invulnerability
post-lethal damage to the same target is discarded
mutual kills across distinct targets remain possible in one authoritative step
credit and result interpretation remain outside damage math
environmental attribution can consume recent damage/forced-movement facts
exact environmental attribution windows remain with awards/attribution planning
```

## Related Docs

- [Gameplay Planning](./!INDEX.md)
- [Teams And Team Rules](teams-and-team-rules.md)
- [Lives, Death, Elimination, And Respawn](lives-death-elimination-and-respawn.md)
- [Gameplay Awards And Counters](gameplay-awards-and-counters.md)
- [Modes And Match Rules](modes-and-match-rules.md)
- [Player Experience Systems](player-experience-systems.md)
- [Damage](../../../systems-design/combat/damage.md)
- [Damage Resolution](../../../services/game-server/simulation/combat/damage-resolution.md)
- [Collision To Damage Flow](../../../services/game-server/simulation/combat/collision-to-damage-flow.md)
- [Radial Effects](../../../services/game-server/simulation/combat/radial-effects.md)
- [Weapons And Projectile Fire](../../../services/game-server/simulation/combat/weapons-and-projectile-fire.md)
- [Player Death And Despawn](../../../services/game-server/simulation/players/player-death-and-despawn.md)
- [Player Counters](../../../services/game-server/simulation/players/player-counters.md)
- [Current System Limits](../../../limits/current-system-limits.md#combat-systems)

## Remaining Implementation-Level Decisions

- Exact policy, source, target, attribution, eligibility, result, and event type/field names.
- Exact schema for default relationship permissions and mode overrides.
- Exact representation of self, ally, enemy, neutral, destructible/environment, and FFA relationships.
- Exact PvP-enabled rule handoff and team relationship lookup contract.
- Exact source ownership and responsibility retention across chained, reflected, radial, and delayed effects.
- Exact representation of responsible player, responsible team, original instigator, and disconnected participant records.
- Exact damage type and cause extensibility contract while preserving existing values.
- Exact modifier storage, filtering, invalid-value handling, and rounding representation.
- Exact signed-value and restoration-target representation.
- Exact shield bypass and overflow policy field names and precedence.
- Exact collision damage profile values, survivor repulsion/knockback request, and physics handoff.
- Exact radial/annular falloff representation and per-target repeat rules.
- Exact DoT stack keys, refresh/replace/limit behavior, tick timing, and death cleanup handoff.
- Exact invulnerability state and explicit dev/admin authorization marker.
- Exact lethal-step ordering and target-state bookkeeping for post-lethal discard.
- Exact result/event projection and which authoritative hit facts are emitted for blocked or ineffective results.
- Exact handoff contract for recent damage and forced-movement environmental attribution.
- Exact award/attribution windows, precedence, and retained-record lifetime.
- Exact runtime health/shield application boundary and package ownership.
- Exact packet, storage, and package shapes chosen at implementation time.

There are no remaining product-level damage or healing decisions blocking P4 system planning.

## Core Invariants

```text
Damage And Healing Rules is the authoritative P4 planning owner for damage and healing policy.

standard_damage_v1 is the default named policy.

Damage is server-authoritative.

Eligibility, pure resolution, and downstream consequences are separate boundaries.

The shared resolver does not mutate live runtime entities.

Supported source categories route through the same authoritative damage/healing model.

Source defaults define relationship permissions, and modes may override them.

Enemy friendly fire is policy-permitted where resolved source/mode rules allow it.

Player same-team friendly fire is prohibited.

PvP player damage is inter-team only; FFA participants are opposing sides when PvP is enabled.

Immediate source, responsible player, responsible team, and original instigator are retained for chained and delayed effects.

Existing damage types, causes, modifiers, and calculation order remain the foundation.

Positive signed values damage; negative signed values heal or repair through the shared pipeline.

Sources decide whether restoration affects health, shields, or both, and restoration never exceeds maximums.

Healing and repair have distinct result/event semantics from damage.

Shields absorb first by default; bypass and overflow are explicit source/mode policy.

There is no catastrophic-kill mechanic; extremely high ordinary damage uses ordinary resolution.

Surviving colliders receive physics-owned repulsion or knockback so they do not remain overlapped.

Radial and annular effects own coverage/falloff; no line-of-sight mechanic is required.

DoT stacks by default, supports source overrides, and does not persist through death.

Critical hits and random damage variation are not supported or planned.

Blocked/invulnerable hits may emit authoritative hit facts.

Normal gameplay sources cannot bypass invulnerability; explicit dev/admin sources may.

Damage after a target's first lethal result is discarded.

Distinct targets may mutually kill one another in the same authoritative step.

Credit, awards, environmental attribution windows, and result interpretation remain outside damage math.

Existing effects may survive owner disconnect and retain participant credit context.
```
