---
author: brian
created: "2026-07-19"
document_id: 019f7d55-fb2c-7f77-a0b9-af51346c2043
document_type: general
policy_exempt: false
summary: This doc is the authoritative P4 planning owner for non-player encounter spawn profile selection, scheduling, population policy, targeting, safety, and lifecycle handoffs.
---
# Encounter Spawn Profiles
Parent index: [Gameplay Planning](./!INDEX.md)

## Purpose

This doc is the authoritative P4 planning owner for non-player encounter spawn profile selection, scheduling, population policy, targeting, safety, and lifecycle handoffs.

It defines how modes select one or more encounter spawn profiles, how validated profile configuration becomes runtime scheduling policy, and how encounter requests become authoritative spawn outcomes. It defines planning boundaries without claiming that the implementation already exists.

## Overview

This plan describes the current direction, ownership boundary, implementation status, remaining work, and open decisions for Encounter Spawn Profiles.

## Ownership Boundary

This doc owns planning for:

```text
non-player encounter spawn profile identifiers and selection
profile-declared spawn configuration
spawn rosters and encounter composition
continuous, wave, event, objective, and scripted scheduling
population controls and shared weighted population budgets
profile and encounter-type population limits
dynamic encounter scaling
spatial targeting and spawn placement policy
encounter safety validation
match-seeded deterministic RNG
runtime profile state and profile activation
profile deactivation, swapping, and scheduling lifecycle
spawn failure handling, retry, and drop policy
encounter lifecycle and despawn-policy selection
campaign, objective, admin/devtool, and scripted spawn request validation
encounter-spawn telemetry and logging requirements
```

This doc does not own:

```text
player spawning, respawn, or player spawn safety
mode identity, objective meaning, or match-end policy
team assignment or team membership authority
combat, movement, collision, or asteroid-fragmentation behavior
entity-specific behavior after an encounter is spawned
room membership, admission, or reconnect execution
client presentation, UI layout, or input ownership
packet, persistence, or storage schemas
```

Modes select Encounter Spawn Profiles and provide only profile-declared, validated configuration. They do not submit arbitrary scheduler internals or become alternate encounter-spawn authorities. Encounter Spawn Profiles consume authoritative world, player, team, objective, and match facts while keeping encounter policy in this owner.

## Settled Product Model

The baseline existing profile is:

```text
playercentric_asteroids_v1
```

It remains the initial Encounter Spawn Profile for the existing player-centric asteroid behavior. Asteroid fragmentation remains asteroid behavior rather than encounter scheduling: a spawned asteroid's behavior may create fragments under its own rules, but fragmentation is not a profile scheduler event.

Multiple Encounter Spawn Profiles may run concurrently. A profile may be activated, deactivated, or swapped during a match. Activation and deactivation are runtime policy transitions, not room-creation-only configuration changes.

Profiles own the complete non-player encounter scheduling contract, including spawn rosters, schedule types, population limits, scaling, targeting, safety, deterministic randomness, runtime state, failure handling, and lifecycle/despawn policy selection.

## Profile Selection And Configuration

The mode resolves the active profile set and supplies only options declared by those profiles. Configuration is validated before it becomes runtime profile state. A client, campaign, objective, admin, or devtool request cannot replace the resolved profile or bypass its validation, safety, or population limits.

A profile may declare:

```text
spawn roster and encounter types
schedule capabilities and schedule parameters
per-profile population limits
per-type population limits
scaling inputs and bounds
targeting modes and spatial constraints
safety requirements
weighted population costs
retry and failure limits
lifecycle/despawn policy options
priority and contention behavior
```

The resolved configuration must identify the profile, its validated options, and the runtime activation state. Consumers use the resolved contract rather than interpreting profile-specific fields themselves.

## Scheduling Model

Profiles may schedule encounters through:

```text
continuous production
waves
timed or match events
objective-driven triggers
campaign transitions
admin/devtool requests
scripted content requests
```

A profile may combine schedule types, and multiple Encounter Spawn Profiles may schedule concurrently. Scheduling evaluates current profile state, active population, shared budget availability, target context, and validated limits before producing a spawn attempt.

Match end, cutscenes, and campaign transitions stop profile scheduling. These transitions do not invent countdown-specific or no-active-player spawn rules. Simulation pause pauses everything, including profile timers, pending schedule progression, retries, and spawn evaluation; it does not create a separate encounter pause policy.

Deactivation stops new scheduling for that profile. Existing spawned entities remain in the world and follow the normal encounter-lifecycle/despawn policy selected by the profile and evaluated by [Encounter Lifecycle And Despawn](encounter-lifecycle-and-despawn.md). Reactivation resumes the profile's validated scheduling policy from its runtime state according to the implementation-defined profile-to-lifecycle handoff contract; deactivation must not silently delete already spawned entities.

Profile swapping applies the same boundary: the outgoing profile stops new scheduling, existing entities follow their selected normal lifecycle/despawn policy through [Encounter Lifecycle And Despawn](encounter-lifecycle-and-despawn.md), and the incoming profile starts only after its configuration and activation state are validated.

## Population Budgets And Contention

All active profiles share a match-wide weighted population budget/cap. Profiles may additionally define stricter per-profile and per-encounter-type limits. A candidate spawn must satisfy both the shared budget and every applicable local limit.

Weighted cost is roughly based on:

```text
packet size
update frequency
compute cost
```

The exact cost model is implementation-level, but the budget must represent that encounter types with greater replication or simulation burden consume more capacity than equivalent entity counts suggest.

Budget contention is priority-based. When multiple valid scheduling opportunities compete for insufficient shared capacity, the resolved profile and request priority determine which opportunity is admitted. Contention must not be resolved by accidental iteration order or by allowing one profile to bypass the shared cap.

Existing population, pending removals, and accepted spawn reservations are included in budget and limit evaluation according to the eventual runtime accounting contract. A failed or dropped attempt must not reserve population indefinitely.

## Scaling

Profiles may scale encounter production or composition using:

```text
player count
team count
elapsed match time
difficulty
objective state
existing encounter population
```

Player-performance scaling is not part of the settled product model. Score, accuracy, damage dealt, kill rate, skill estimate, or other player-performance measures must not be implicit scaling inputs unless a later product decision explicitly adds that capability.

Join and leave response is indirectly mode-owned through profile configuration. A mode may select profile scaling behavior that reacts to the authoritative player/team facts it provides. Immediate adjustment is sufficient; gradual adjustment is not required. Profile state must re-evaluate validated scaling inputs after relevant join or leave changes without making lifecycle or team systems alternate scaling authorities.

## Targeting, Placement, And Safety

Profiles support spatial targeting based on:

```text
player position or player group
team position or team group
world position or world region
zone
objective position or objective state
scripted position
```

Player-, team-, world-, zone-, objective-, and scripted-position targeting are all supported. Targeting selects or constrains encounter placement; it does not grant a caller authority to choose an unsafe location or bypass profile validation.

Encounter safety belongs to Encounter Spawn Profiles. It covers valid placement regions, wrap-aware spatial checks, dangerous-object and encounter overlap rules, target-distance constraints, and any profile-declared safety fallback. Player spawning remains separate and is owned by [Player Spawn Profiles](player-spawn-profiles.md); encounter safety must not become an implicit player-placement authority or reuse player spawn policy without an explicit contract.

Safety checks use authoritative current world state and wrap-aware spatial rules. A profile may define different safety policy for roster entries or schedule types, but every accepted encounter request passes through the selected profile's safety and population controls.

## Runtime Profile State And Lifecycle

Each active profile has runtime state sufficient to continue and observe its validated policy, including schedule progress, activation state, deterministic RNG state or derivation, pending requests, retry state, population accounting, and relevant scaling inputs.

Profile lifecycle states should distinguish at least:

```text
configured
active
paused by simulation pause
stopped by match/cutscene/campaign transition
deactivated
swapped or retired
```

These states describe scheduling authority, not the existence of already spawned entities. Existing entities are governed by encounter lifecycle and the selected despawn policy after their profile stops scheduling.

Encounter retirement/despawn is a focused post-spawn sub-seam owned by [Encounter Lifecycle And Despawn](encounter-lifecycle-and-despawn.md). This document selects the lifecycle policy and hands the spawned entity and originating profile metadata to that owner for retirement, despawn, and cleanup evaluation.

## External Spawn Requests

Campaign, objective, admin/devtool, and scripted spawn requests are supported. They remain requests into the selected profile rather than alternate spawn authorities.

Every external request passes through:

```text
profile selection and request capability validation
profile-declared configuration validation
schedule/request priority resolution
shared and local population limits
spatial targeting and encounter safety
match-seeded deterministic RNG where randomness is required
normal failure, retry, logging, and lifecycle handling
```

A request may declare its source, target context, requested roster entry, and priority only where the selected profile permits those fields. Campaign and scripted content may provide authored positions or timing, but authored input does not bypass safety, limits, or profile validation. Admin/devtool authority affects who may request an action; it does not remove the encounter profile's runtime constraints.

## Determinism And Randomness

Encounter randomness is deterministic from the match seed and the profile/request context. The implementation must derive stable outcomes without depending on process-global randomness, wall-clock timing, map iteration order, or incidental concurrent execution order.

Deterministic RNG applies to roster selection, wave composition, placement choice, and other profile-owned random decisions. It must remain compatible with profile activation, concurrent profiles, retries, and contention so replay/debug reasoning can distinguish an intentional policy decision from scheduling-order noise.

## Failure And Logging

Standard failed spawns retry up to a configured cap, then drop. Drops are logged. Retries may use debug-level logs, while exhausted attempts and other actionable failures must remain observable at an appropriate operational level.

A failed attempt must report enough context to diagnose profile policy and contention without requiring clients to reconstruct the decision:

```text
profile ID
request or schedule source
encounter type
targeting context
failure category
retry count and configured cap
population/budget contention result
final disposition
```

Failure handling must not return an unsafe encounter merely because placement or budget evaluation failed. A dropped request does not silently bypass the profile or become an unbounded retry loop.

## System Handoffs

```text
resolved mode rules
-> active Encounter Spawn Profile IDs and validated profile options
-> profile activation/runtime state
-> schedule or validated external request
-> authoritative player, team, world, zone, objective, and match facts
-> scaling and priority evaluation
-> shared weighted budget and local population checks
-> deterministic roster and target selection
-> encounter safety validation
-> authoritative spawn attempt
-> selected lifecycle policy and post-spawn retirement/despawn evaluation
-> encounter-spawn telemetry and failure logging
```

### Modes And Match Rules

Modes select Encounter Spawn Profiles and provide only profile-declared validated configuration. They may indirectly control join/leave response through profile configuration, but they do not own roster policy, scheduling, scaling implementation, targeting, safety, budgets, retries, or lifecycle decisions.

### Player Spawn Profiles

Player Spawn Profiles own player placement and player spawn safety. Encounter Spawn Profiles own non-player encounter placement and encounter safety. The two owners may share wrap-aware spatial primitives and world facts, but neither becomes the other's authority.

### Teams And Team Rules

Teams expose authoritative team count, membership, and spatial facts where available. Encounter Spawn Profiles may use team count and team position/group targeting as configured scaling or targeting inputs. Team rules do not schedule encounters or select encounter placement.

### Objectives, Campaign, And Scripted Content

Objectives and campaign/scripted systems provide validated context or requests. The Encounter Spawn Profile validates capability, safety, population, and scheduling before accepting them. Objective or campaign meaning and transition policy remain with those systems; they do not become alternate spawn authorities.

### Multiplayer Session And Lifecycle

Lifecycle supplies authoritative player join, leave, active-participation, and match-transition facts. Encounter Spawn Profiles re-evaluate configured scaling and scheduling in response. Lifecycle does not own encounter population policy or spawn retry behavior.

### Runtime Encounters And Entity Behavior

The runtime owns the spawned entity's live behavior after an authoritative spawn outcome. Entity-specific movement, combat, collision, asteroid fragmentation, and other behavior remain with the entity/runtime owners. The profile owns only the scheduling and spawn-policy decisions leading to the entity's creation and selected lifecycle handoff; [Encounter Lifecycle And Despawn](encounter-lifecycle-and-despawn.md) evaluates post-spawn retirement and cleanup.

### Match End And Presentation

Match-end, cutscene, and campaign-transition owners stop profile scheduling through an explicit lifecycle handoff. Presentation and devtools consume authoritative profile, spawn, population, and failure telemetry; they do not infer or recreate scheduling decisions.

## Implementation Direction

The first implementation slice should establish the Encounter Spawn Profile seam while preserving the baseline existing behavior:

```text
1. Define the Encounter Spawn Profile identifier and playercentric_asteroids_v1 contract.
2. Define validated profile configuration and active-profile runtime state.
3. Route existing non-player spawn scheduling through the profile owner.
4. Support concurrent profile activation, deactivation, and validated swapping.
5. Add continuous, wave, event, objective, and scripted scheduling seams.
6. Add shared weighted population budget plus optional per-profile and per-type limits.
7. Add priority-based contention and deterministic match-seeded RNG.
8. Add configured scaling from player/team count, elapsed time, difficulty, objective state, and existing population.
9. Add player-, team-, world-, zone-, objective-, and scripted-position targeting.
10. Apply encounter safety separately from player-spawn safety.
11. Add bounded retry, drop, telemetry, and logging behavior.
12. Hand accepted entities to normal runtime encounter lifecycle and selected despawn policy.
13. Preserve asteroid fragmentation as asteroid behavior rather than scheduler behavior.
```

Implementation should keep encounter policy in this focused owner. Modes, lifecycle, objectives, campaign content, devtools, runtime entities, and presentation should provide facts or consume outcomes rather than becoming alternate encounter-spawn authorities. This document describes planning direction only; it does not claim that these steps are implemented.

## Testing Direction

Important future checks:

```text
playercentric_asteroids_v1 is the baseline existing profile
modes select profiles and only profile-declared configuration is accepted
multiple profiles can run concurrently
profiles can activate, deactivate, and swap during a match
deactivation stops new scheduling but preserves existing entities
existing entities follow selected normal lifecycle/despawn policy
continuous, wave, event, objective, and scripted scheduling are representable
simulation pause pauses all profile scheduling and state progression
match end, cutscenes, and campaign transitions stop scheduling
scaling supports player count, team count, elapsed time, difficulty, objective state, and existing population
player-performance scaling is not used
join/leave response can adjust immediately through profile configuration
deterministic match-seeded RNG is independent of wall clock and incidental iteration order
player-, team-, world-, zone-, objective-, and scripted-position targeting are supported
encounter safety is enforced separately from player spawning
campaign, objective, admin/devtool, and scripted requests pass profile validation, safety, and limits
shared weighted population budget applies across all active profiles
optional per-profile and per-type limits can be stricter than the shared cap
weighted cost accounts roughly for packet size, update frequency, and compute cost
budget contention is priority-based rather than iteration-order based
standard failures retry up to the configured cap and then drop
exhausted drops are logged and retries may be debug-level
asteroid fragmentation remains asteroid behavior
```

## Related Docs

- [Gameplay Planning](./!INDEX.md)
- [Modes And Match Rules](modes-and-match-rules.md)
- [Player Spawn Profiles](player-spawn-profiles.md)
- [Teams And Team Rules](teams-and-team-rules.md)
- [Enemies, Bosses, And Encounters](enemies-bosses-and-encounters.md)
- [Encounter Lifecycle And Despawn](encounter-lifecycle-and-despawn.md)
- [Objectives And Objective Runtime](objectives-and-objective-runtime.md)
- [Levels, Missions, And Content Structure](levels-missions-and-content-structure.md)
- [Multiplayer Session And Lifecycle](../platform/multiplayer-session-and-lifecycle.md)
- [Devtools And Telemetry](../../devtools/devtools-and-telemetry.md)

## Remaining Implementation-Level Decisions

- Exact profile type, field, package, and registry names.
- Exact resolved-rule representation for active profile sets and profile options.
- Exact schedule state, activation, deactivation, and swapping transition model.
- Exact roster, encounter-type, and schedule configuration shapes.
- Exact scaling formulas, bounds, update cadence, and difficulty representation.
- Exact weighted cost formula and budget accounting timing.
- Exact local population-limit and priority vocabulary.
- Exact contention tie-break rules after priority is applied.
- Exact match-seeded RNG derivation and state persistence across retries and profile swaps.
- Exact targeting data shapes for player, team, world, zone, objective, and scripted positions.
- Exact wrap-aware spatial APIs and encounter safety categories.
- Exact reservation semantics for population and weighted budget.
- Exact retry timing, failure categories, cap configuration, and drop telemetry fields.
- Exact runtime handoff, profile-to-entity lifecycle metadata, and lifecycle-policy selection contract.
- Exact campaign, objective, admin/devtool, and scripted authorization and request shapes.
- Exact telemetry event names, field types, sampling, and retention.
- Exact packet, persistence, and package boundaries chosen at implementation time.

There are no remaining product-level Encounter Spawn Profile or Encounter Lifecycle/Despawn decisions blocking P4 system planning. Lifecycle/retirement/despawn product planning is complete in [Encounter Lifecycle And Despawn](encounter-lifecycle-and-despawn.md).

## Core Invariants

```text
Encounter Spawn Profiles is the authoritative planning owner for all non-player encounter spawning rules and policy.

Modes select profiles and provide only profile-declared validated configuration.

playercentric_asteroids_v1 is the baseline existing profile.

Multiple profiles may run concurrently and may activate, deactivate, or swap during a match.

Profiles own rosters, scheduling, population controls, scaling, targeting, safety, deterministic RNG, runtime state, failure handling, and lifecycle/despawn policy selection.

Scaling uses player/team count, elapsed time, difficulty, objective state, and existing population; player-performance scaling is excluded.

Join/leave response is indirectly mode-owned through profile configuration; immediate adjustment is sufficient.

Player-, team-, world-, zone-, objective-, and scripted-position targeting are supported.

Encounter safety belongs here; player spawning remains separate.

Simulation pause pauses everything. Match end, cutscenes, and campaign transitions stop profile scheduling.

Deactivation stops new scheduling while existing entities follow normal encounter lifecycle/despawn policy until retirement.

Campaign, objective, admin/devtool, and scripted requests pass through profile validation, safety, and population limits.

Standard failures retry up to a configured cap, then drop and log the drop.

All active profiles share a match-wide weighted population budget/cap, with optional stricter per-profile and per-type limits.

Budget contention is priority-based, and weighted cost roughly reflects packet size, update frequency, and compute cost.

Asteroid fragmentation remains asteroid behavior rather than encounter scheduling.

Encounter Spawn Profiles select lifecycle/despawn policy; [Encounter Lifecycle And Despawn](encounter-lifecycle-and-despawn.md) is the authoritative post-spawn retirement, despawn, and cleanup evaluation owner.

There are no remaining product-level Encounter Spawn Profile decisions blocking P4 system planning.
```

## Notes

Implemented facts must move to canonical current documentation; this plan should retain only unresolved work, sequencing, and open decisions.
