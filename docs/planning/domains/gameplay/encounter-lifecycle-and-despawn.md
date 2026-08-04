---
author: brian
created: "2026-07-19"
document_id: 019f7d55-fb2c-787f-9e9c-12194dae1781
document_type: general
policy_exempt: false
summary: This doc is the authoritative P4 planning owner for non-player encounter retirement, despawn, and lifecycle evaluation after spawn.
---
# Encounter Lifecycle And Despawn
Parent index: [Gameplay Planning](./!INDEX.md)

## Purpose

This doc is the authoritative P4 planning owner for non-player encounter retirement, despawn, and lifecycle evaluation after spawn.

It defines how spawned encounter entities remain eligible, become softly or hard-retired, and hand off cleanup without claiming that the implementation already exists. It keeps retirement policy separate from encounter scheduling, entity behavior, player spawning, objectives, match-end policy, and scene reset ownership.

## Overview

This plan describes the current direction, ownership boundary, implementation status, remaining work, and open decisions for Encounter Lifecycle And Despawn.

## Ownership Boundary

This doc owns planning for:

```text
non-player encounter lifecycle evaluation after spawn
retirement and despawn trigger policy
soft retirement and hard retirement policy
outside-player and allowed-region retirement policy
population-pressure cleanup fallback
profile/phase retirement policy
scripted, campaign, and transition cleanup requests
origin-profile lifecycle accounting after spawn
retirement priority and deterministic candidate ordering
retirement attribution and cleanup handoffs
```

This doc does not own:

```text
encounter scheduling or spawn admission
encounter composition, waves, or authored content
entity movement, combat, AI, collision, or behavior
player spawning, respawn, or player lifecycle
objective meaning, objective completion, or objective credit
match-end evaluation or result locking
room membership, admission, or reconnect execution
scene reset or broad simulation teardown ownership
client presentation, UI, or input ownership
packet, persistence, or storage schemas
```

[Encounter Spawn Profiles](encounter-spawn-profiles.md) selects lifecycle policy as part of encounter spawn configuration and owns scheduling. This owner evaluates the selected policy after an entity exists. Entity types may declare lifecycle capabilities and exceptions, but they do not become alternate general lifecycle authorities.

## Settled Product Model

The lifecycle owner evaluates retirement triggers for every eligible non-player encounter entity after spawn. The supported trigger families are:

```text
lifetime expiry
outside all relevant players
leaving an allowed region
population pressure
profile/phase retirement
scripted or campaign cleanup
match transition/reset
```

A trigger makes an entity eligible for retirement according to its lifecycle policy. Retirement is not destruction. It does not grant score, kill credit, drops, or objective credit unless an explicit policy configures that outcome.

Whole-simulation pause pauses lifecycle evaluation. There is no separate encounter pause and no separate no-active-player retirement policy. Match results never wait for ordinary encounter retirement.

## Lifecycle Evaluation After Spawn

The lifecycle owner receives authoritative world, player, profile, phase, match-transition, and entity metadata facts. It evaluates the selected policy without taking ownership of the systems that produced those facts.

The baseline evaluation concept is:

```text
active simulation tick
-> lifecycle owner reads entity metadata and current authoritative facts
-> evaluate configured retirement triggers
-> apply entity-type capabilities and exceptions
-> select soft or hard retirement outcome
-> emit cleanup/retirement handoff
-> update profile population accounting after authoritative removal
```

The exact tick order, evaluation cadence, and data structures remain implementation-level decisions. Lifecycle evaluation must not depend on client visibility alone, and it must not silently turn a normal retirement into a destruction or reward event.

## Supported Retirement Triggers

### Lifetime Expiry

A lifecycle policy may give an entity a finite lifetime. When that lifetime expires, the lifecycle owner evaluates retirement. Exact timer semantics, grace periods, and whether soft retirement delays removal are implementation-level decisions.

### Outside All Relevant Players

All active players are relevant for distance retirement. An entity is eligible only when it is outside every relevant player's viewable area plus a tunable extra distance. The contract is an all-players condition, not a nearest-player or single-target condition.

Visible cleanup should preferably be avoided. The final distance and visibility relationship, wrap-aware spatial calculation, and tunable extra distance remain implementation-level decisions.

### Leaving An Allowed Region

A profile or entity lifecycle policy may define an allowed region. Leaving that region can trigger retirement independently of player distance. Region interpretation, wrap-aware boundaries, and exceptions for entities that may roam beyond the region remain implementation-level decisions.

### Population Pressure

Population pressure first stops or delays new spawning through the Encounter Spawn Profile owner. Cleanup is a fallback when pressure remains or when the policy explicitly requires removal.

When cleanup is required, candidates prefer entities farthest from players, followed by profile-defined ordering. Exact candidate eligibility, weighted population accounting, pressure thresholds, and ordering fields remain implementation-level decisions. Population cleanup must not become an excuse to destroy visible or high-priority entities when a valid lower-impact candidate exists.

### Profile Or Phase Retirement

Profile deactivation or phase retirement stops new scheduling for the affected profile or phase. Existing entities continue their normal lifecycle by default. A profile or phase may explicitly request accelerated or immediate cleanup, but that request is distinct from ordinary deactivation.

Surviving entities remain associated with their originating profile and continue to count toward that profile's limits and budget when the profile is reactivated. Reactivation must not erase existing population accounting or silently reassign surviving entities.

### Scripted Or Campaign Cleanup

Scripted and campaign systems may request cleanup through an explicit lifecycle handoff. The request does not make those systems general despawn authorities. Authorization, scope, and whether the request permits hard removal remain part of the request contract and are implementation-level decisions.

### Match Transition Or Reset

Match results do not wait for ordinary retirement. A match transition or reset may hard-remove everything through the transition/reset owner and its explicit cleanup handoff. This is transition cleanup, not a normal encounter retirement outcome.

## Soft And Hard Retirement

Both soft and hard retirement are supported.

```text
soft retirement
-> entity stops or limits future activity according to its type policy
-> entity remains until the configured cleanup handoff completes

hard retirement
-> entity is removed immediately through the authoritative cleanup path
```

Soft retirement is useful for entity types that need an explicit shutdown, terminal animation, child cleanup, or deferred removal. Hard retirement is available for transition/reset, explicit immediate cleanup, or types whose policy permits direct removal.

The entity type may declare whether it supports soft retirement, requires destruction, requires explicit cleanup, or has other lifecycle capabilities/exceptions. The lifecycle owner still decides when the configured trigger applies; entity-specific behavior decides how its declared cleanup contract is executed.

## Population And Cleanup Priority

Population pressure is primarily a scheduling concern. The scheduler should stop or delay spawning before lifecycle cleanup is used. If fallback cleanup is required, the lifecycle owner evaluates eligible candidates using authoritative player positions and profile policy.

The preferred cleanup ordering is:

```text
farthest from all relevant players according to the configured policy
-> profile-defined priority/order
-> deterministic tie-break ordering
```

The exact distance aggregation, weighted population cost comparison, visibility penalty, profile ordering fields, and deterministic tie-break key are implementation-level decisions. Equal cleanup candidates must resolve deterministically and must not depend on map iteration, goroutine timing, or incidental entity insertion order.

## Entity Metadata And Origin Accounting

Each spawned encounter entity carries lifecycle metadata sufficient for post-spawn evaluation:

```text
originating profile
spawn type
lifecycle policy
priority
weighted population cost
```

Metadata may also include entity-type lifecycle capabilities, phase association, region association, and scripted cleanup eligibility when required by the resolved policy. Exact field and package names remain implementation-level decisions.

The originating profile remains the accounting owner for surviving entities. Deactivation stops new scheduling but does not remove existing entities or transfer their cost to another profile. On reactivation, surviving entities still count toward the originating profile's limits and budget.

## Dependent Entities And Cleanup

Dependent behavior is type-defined. Children generally persist under their own lifecycle rather than being implicitly removed because a parent retires.

An entity type may explicitly require destruction or explicit cleanup of dependents. Such requirements are declared as lifecycle capabilities or exceptions and are honored through the type's cleanup contract. The general lifecycle owner must not infer parent-child destruction semantics from generic entity relationships.

Retirement and destruction remain separate outcomes. A type-defined cleanup path may eventually destroy an entity, but ordinary retirement does not automatically produce destruction credit, scoring, drops, or objective progress.

## Attribution, Logging, And Observability

Routine retirement is not normally logged. Unexpected forced-cleanup failures or invalid lifecycle state may be logged at an appropriate operational level. Debug traces are optional and may expose trigger, candidate, profile, phase, and cleanup disposition for diagnosis.

A retirement outcome must remain observable through authoritative state or telemetry where the surrounding system requires it, without turning routine cleanup into noisy operational logging. Exact event names, fields, levels, sampling, and retention remain implementation-level decisions.

## System Handoffs

```text
Encounter Spawn Profile
-> selected lifecycle policy and originating profile metadata
-> authoritative spawned entity
-> lifecycle owner evaluates current triggers and entity capabilities
-> soft-retirement or hard-retirement cleanup handoff
-> runtime entity/type cleanup contract
-> profile population and weighted-budget accounting update
-> optional telemetry/debug trace
```

### Encounter Spawn Profiles

Encounter Spawn Profiles own scheduling, spawn admission, population limits, profile activation, and lifecycle-policy selection. Deactivation stops new scheduling. Existing entities follow this owner’s normal lifecycle policy unless accelerated or immediate cleanup is explicitly requested.

### Entity Runtime, Combat, And Behavior

Runtime entity types own movement, behavior, combat, collision, destruction mechanics, and type-defined cleanup capabilities. The lifecycle owner supplies a retirement decision and invokes the declared cleanup contract; it does not decide how an enemy, asteroid, boss, or other entity behaves during retirement.

### Player Runtime And Viewability

Player runtime supplies all active relevant-player positions and viewable-area facts. Distance retirement requires the entity to be outside every relevant player's viewable area plus tunable extra distance. Player spawning and player lifecycle do not own non-player retirement.

### Modes And Match Rules

Modes select encounter profiles and match policies. Match rules own match-end semantics and result locking. Match results never wait for ordinary encounter retirement. Match transition/reset may request explicit hard removal through the transition owner.

### Objectives And Objective Runtime

Objectives may provide authored or scripted cleanup context and may consume explicitly configured outcomes. They do not own general retirement, and routine retirement grants no objective credit by default.

### Campaign And Scripted Content

Campaign and scripted systems may issue explicit, authorized cleanup requests. They define the content meaning and scope of the request; the lifecycle owner validates and executes the lifecycle policy without becoming an alternate campaign or script runtime.

### Multiplayer Session And Lifecycle

Session/lifecycle supplies authoritative participation and match-transition facts. It does not create a separate no-active-player policy. Whole-simulation pause and broad transition state come from the simulation/lifecycle boundary and pause lifecycle evaluation consistently.

### Devtools And Telemetry

Devtools may inspect lifecycle metadata, retirement eligibility, cleanup disposition, and optional debug traces. Devtools do not bypass lifecycle policy except through an explicit authorized request path. Telemetry consumes authoritative outcomes rather than reconstructing retirement from client visibility.

## Implementation Direction

The first implementation slice should establish a focused lifecycle owner without claiming that the planned behavior already exists:

```text
1. Define the post-spawn encounter lifecycle contract and selected-policy handoff.
2. Preserve originating profile, spawn type, lifecycle policy, priority, and weighted population cost metadata.
3. Evaluate lifetime, all-relevant-player distance, allowed-region, population-pressure, profile/phase, scripted/campaign, and transition/reset triggers.
4. Keep population pressure primarily in scheduling, with lifecycle cleanup as fallback.
5. Support soft and hard retirement through explicit entity-type cleanup capabilities.
6. Preserve surviving entities and origin-profile accounting across profile deactivation/reactivation.
7. Keep dependent-entity behavior type-defined, with children generally using their own lifecycle.
8. Add deterministic candidate ordering and explicit cleanup handoffs.
9. Keep routine retirement quiet while exposing actionable forced-cleanup failures and optional debug traces.
10. Integrate match transition/reset hard removal without making match results wait for retirement.
```

Implementation should keep retirement policy in this focused owner. Scheduling, entity behavior, player lifecycle, objectives, match results, campaign meaning, devtools, and scene reset should provide facts or consume outcomes rather than becoming alternate encounter-lifecycle authorities. This document describes planning direction only; it does not claim that these steps are implemented.

## Testing Direction

Important future checks:

```text
all supported trigger families are representable
lifetime expiry evaluates the selected lifecycle policy
outside-player retirement requires being outside every relevant active player's viewable area plus extra distance
visible cleanup is preferably avoided
allowed-region exit is distinct from player-distance retirement
population pressure first stops or delays spawning
cleanup is a fallback and prefers farthest eligible candidates plus profile-defined ordering
soft and hard retirement use distinct cleanup paths
entity types can declare destruction or explicit-cleanup exceptions
profile deactivation stops new scheduling but existing entities continue normal lifecycle
explicit accelerated/immediate cleanup is distinct from ordinary deactivation
surviving entities retain origin-profile association and count on reactivation
metadata includes originating profile, spawn type, lifecycle policy, priority, and weighted population cost
children generally persist under their own lifecycle
routine retirement grants no score, kill credit, drops, or objective credit by default
forced-cleanup failures and invalid state are observable without routine retirement log noise
simulation pause pauses lifecycle evaluation
there is no separate encounter pause or no-active-player policy
match results do not wait for retirement
transition/reset can hard-remove everything
equal cleanup candidates resolve deterministically
```

## Related Docs

- [Gameplay Planning](./!INDEX.md)
- [Encounter Spawn Profiles](encounter-spawn-profiles.md)
- [Enemies, Bosses, And Encounters](enemies-bosses-and-encounters.md)
- [Modes And Match Rules](modes-and-match-rules.md)
- [Player Spawn Profiles](player-spawn-profiles.md)
- [Gameplay Awards And Counters](gameplay-awards-and-counters.md)
- [Objectives And Objective Runtime](objectives-and-objective-runtime.md)
- [Multiplayer Session And Lifecycle](../platform/multiplayer-session-and-lifecycle.md)
- [Devtools And Telemetry](../../devtools/devtools-and-telemetry.md)

## Remaining Implementation-Level Decisions

- Exact lifecycle owner type, field, package, and registry names.
- Exact lifecycle-policy representation and profile-to-entity handoff shape.
- Exact trigger evaluation cadence, ordering, and authoritative snapshot boundary.
- Exact lifetime timer, expiry, grace-period, and soft-retirement timing semantics.
- Exact viewable-area representation, wrap-aware distance API, and tunable extra-distance fields.
- Exact all-relevant-player distance aggregation and visibility-avoidance algorithm.
- Exact allowed-region shapes, boundaries, wrap behavior, and entity exceptions.
- Exact population-pressure thresholds, scheduler stop/delay contract, and cleanup fallback threshold.
- Exact weighted population accounting timing and treatment of pending removals.
- Exact farthest-from-players metric and profile-defined candidate ordering fields.
- Exact deterministic tie-break key and ordering across equal cleanup candidates.
- Exact soft-retirement state, hard-removal path, grace period, and cleanup acknowledgement contract.
- Exact entity lifecycle capability and exception vocabulary, including destruction and explicit-cleanup requirements.
- Exact dependent-entity and parent/child cleanup contract for each entity type.
- Exact profile/phase retirement request shape and accelerated/immediate cleanup authorization.
- Exact originating-profile metadata field names and reactivation accounting behavior.
- Exact scripted/campaign cleanup authorization, scope, and override contract.
- Exact match-transition/reset handoff and hard-remove-all sequencing.
- Exact score, kill-credit, drops, and objective-credit opt-in representation.
- Exact forced-cleanup failure and invalid-state logging fields, levels, and retention.
- Exact debug trace fields, sampling, and devtool projection.
- Exact telemetry event names, packet/storage boundaries, and package placement.

There are no remaining product-level Encounter Lifecycle/Despawn decisions blocking P4 planning. Remaining work is implementation-level contract, algorithm, threshold, field, package, and tuning definition.

## Core Invariants

```text
Encounter Lifecycle And Despawn is the authoritative planning owner for non-player post-spawn retirement and lifecycle evaluation.

Encounter Spawn Profiles own scheduling and select lifecycle policy; they do not delegate scheduling authority to this owner.

This owner does not own entity combat, behavior, player spawning, objectives, match-end, or scene reset policy.

Supported triggers are lifetime expiry, outside all relevant players, leaving an allowed region, population pressure, profile/phase retirement, scripted/campaign cleanup, and match transition/reset.

All active players are relevant for distance retirement.

Distance retirement requires the entity to be outside every relevant player's viewable area plus tunable extra distance.

Visible cleanup should preferably be avoided.

Population pressure first stops or delays spawning; cleanup is fallback.

Fallback cleanup prefers farthest eligible entities from players plus profile-defined ordering.

Soft and hard retirement are both supported.

Entity types may declare lifecycle capabilities and exceptions, including destruction or explicit-cleanup requirements.

Profile deactivation stops new scheduling; existing entities continue normal lifecycle unless accelerated or immediate cleanup is explicitly requested.

Surviving entities retain their originating profile and count toward that profile's limits and budget on reactivation.

Lifecycle metadata includes originating profile, spawn type, lifecycle policy, priority, and weighted population cost.

Dependent behavior is type-defined; children generally persist under their own lifecycle.

Retirement is not destruction and grants no score, kill credit, drops, or objective credit unless explicitly configured.

Routine retirement is not normally logged; unexpected forced-cleanup failures or invalid state may log, and debug traces are optional.

Whole-simulation pause pauses lifecycle evaluation; there is no separate encounter pause or no-active-player policy.

Match results never wait for retirement.

Transition/reset may hard-remove everything.

Equal cleanup candidates resolve deterministically.

There are no remaining product-level Encounter Lifecycle/Despawn decisions blocking P4 planning.
```

## Notes

Implemented facts must move to canonical current documentation; this plan should retain only unresolved work, sequencing, and open decisions.
