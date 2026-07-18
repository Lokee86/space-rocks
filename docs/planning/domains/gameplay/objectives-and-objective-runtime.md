---
author: brian
created: "2026-07-19"
document_id: 019f7d55-fb2c-777e-8790-042b42dd48f0
document_type: general
policy_exempt: false
summary: This doc is the authoritative P4 planning owner for the shared Objective Foundation.
---
# Objectives And Objective Runtime
Parent index: [Gameplay Planning](./!INDEX.md)

## Purpose

This doc is the authoritative P4 planning owner for the shared Objective Foundation.

It defines the schema/definition-driven objective state-machine factory and runtime: how objective definitions describe conditions and transitions, how objective instances consume authoritative facts, what objective-local state is retained, and which lifecycle, progress, visibility, and resolution facts are reported. It defines a reusable foundation without making objectives the owner of the systems that generate facts or the consumers that interpret them.

## Ownership Boundary

This doc owns planning for:

```text
objective definition and schema contracts
objective state-machine construction and evaluation
objective-local instance state and statistics required by conditions
objective condition and progress consumption
objective lifecycle transitions and failure reasons
objective discovery and definition-owned visibility state
objective success, failure, cancellation, and retirement facts
objective progress and timer event reporting
visibility-aware objective snapshots
objective lifecycle and progress events
multiple-instance identity and definition association fields
objective-foundation devtools actions
```

This doc does not own:

```text
progress generation or progress pipelines
progress aggregation
attribution generation or attribution policy production
scheduling, activation policy, or orchestration
rewards or award calculation
persistence or durable account state
match state or match-end decisions
campaign structure or story sequencing
team membership, joining, or forfeiture
result interpretation or result inclusion policy
achievements or other meta-progression
client UI, presentation, or animation
packet schemas, storage schemas, or package names
```

Separate progress, aggregation, attribution, gameplay, timer, and consumer systems supply facts. The Objective Foundation validates those facts against a definition, advances objective-local state, evaluates the definition's state machine, and reports normalized results. Consumers own the policy that determines when objectives are created, activated, composed, interpreted, or handed off.

## Settled Product Model

The Objective Foundation is a schema/definition-driven state-machine factory/runtime. A definition describes condition and progress inputs, permitted lifecycle transitions, failure capability, visibility behavior, timer behavior, and any definition-specific association or scope rules. The runtime creates objective instances from those definitions and evaluates them from authoritative supplied facts.

Objectives track only the local state and statistics needed by their own conditions. They do not generate progress, aggregate player values, produce attribution, schedule evaluations, grant rewards, persist durable state, decide match outcomes, or become a universal orchestration system.

The required composition hierarchy is:

```text
atomic Single Objective
    -> Multiple Objectives composed of atomic objectives
        -> Meta-Objectives composed of Multiple Objectives
```

This is a composition model, not a fixed initial type list and not a requirement for language-level inheritance. A consumer may use the smallest composition level that fits its policy. The foundation provides the state and evaluation abstractions needed by each level while keeping composition policy with the consumer.

Objective ownership/scope may be:

```text
player
team
match
collection
definition-specific
```

Personal objective instances begin at zero in the first implementation. Shared progress belongs to an aggregate object owned outside the Objective Foundation. Team progress is derived from player progress updates by a separate aggregation/progress owner; the objective consumes the resulting team fact and does not independently aggregate team membership or player values.

Definitions are effectively immutable. No version field is required. If a change in an underlying system makes a definition unattainable, that definition and all active instances are retired, new instances are blocked, and a replacement definition is added with a new definition ID. Retired instances do not finish.

Multiple instances of one definition may be active at the same time. Each instance has an `objective_instance_id` and definition-owned association fields sufficient to identify its owner, scope, source, or consumer context. Instance identity is distinct from definition identity; it is not a replacement-generation mechanism.

The initial baseline remains Arcade Survival with no objective. That is a mode-policy reference, not Objective Foundation ownership of mode selection or baseline rules.

## Definitions, Composition, And Conditions

An objective definition is the authoritative schema for one state-machine shape. It may define a boolean condition, numeric progress, a timer, staged or sequence conditions, collection/set requirements, survival or maintain-condition behavior, or another schema-defined form. This is an extensible set of forms, not a closed catalogue.

Progress increase, decrease, reset, regression, overflow, and related behavior are per-definition or consumer-owned decisions. Progress is fed from planned progress systems. The Objective Foundation consumes the supplied progress/stat facts and applies the definition's condition semantics; it does not invent a second progress pipeline.

Attribution policies consumed by objectives may include:

```text
one-hit attribution
in-game attribution
in-encounter attribution
```

Attribution production remains a separate system. The objective receives the applicable attribution fact and evaluates it according to its definition; it does not establish the credited source or reconstruct attribution from client events.

Definitions may describe local statistics such as observed contributions, current condition values, milestones, set members, or timer state when those values are required for evaluation. They must not silently become owners of the source system's aggregate, attribution, reward, or persistence data.

## Lifecycle, Discovery, And Visibility

The conceptual lifecycle states are:

```text
undiscovered
discovered
inactive
active
completed
failed
cancelled
retired
```

Not every objective uses every state. Definitions and consumers determine which states are reachable. Discovery is distinct from activation: a discoverable objective may move from `undiscovered` to `discovered` before it becomes active. A non-discoverable objective skips the discovery states rather than undergoing discovery.

Visibility is definition-owned. The runtime records discovery and visibility state and produces visibility-aware snapshots, but presentation does not decide what exists or what becomes visible. A snapshot must not expose state that the definition marks undiscoverable or unauthorized.

Failure capability and failure conditions are definition-owned. Some objectives are non-failable. Every failed objective has a failure reason. Timer expiry emits an expiry event first and defaults to:

```text
status: failed
failure_reason: timer_expired
```

unless the definition specifies another valid transition. If success and failure both become true in the same evaluation boundary, success wins.

Cancellation and retirement are distinct from failure and completion. A consumer may cancel an instance under its own policy. Retirement makes the definition unavailable for new instances and retires all active instances when the definition becomes unattainable; retired instances do not finish.

Reset, timer, repeatability, overflow, and visibility behavior are definition/consumer-owned. Disconnect and reconnect preserve objective instance state. The Objective Foundation does not reset an instance because a participant disconnected.

## Progress, Facts, And Timers

Objective evaluation consumes condition, progress, statistic, gameplay, attribution, aggregation, and timer facts supplied by their owning systems. A fact may represent a player update, an aggregate result, an encounter outcome, a timer transition, or another authoritative input named by the definition.

The foundation reports objective-local progress and condition changes, but the source systems retain ownership of fact production. A shared progress object is external state. The objective may reference it through a definition-specific input contract without copying authority into the objective runtime.

Timers are objective-owned. They continue through ordinary simulation pause and stop only when simulation time itself is paused because all players are disconnected. Timer expiry emits an event before the default failure transition is applied. Exact clock and scheduling mechanics remain implementation-level decisions, but timer semantics must use authoritative simulation time.

Objective updates emit optimistically. Individual emission is appropriate for cold or narrow updates; batched emission is appropriate for hot or broad updates. The emission strategy is a transport/consumer concern around the objective event contract, not a second objective state authority.

Idempotency and deduplication are provided by the already planned separate event/progress infrastructure. The Objective Foundation consumes those guarantees and must not redefine that infrastructure. Its responsibility is to evaluate accepted facts and avoid creating a contradictory local transition from a fact that the supplied infrastructure has already deduplicated.

## System Handoffs

```text
schema/definition source
-> immutable objective definition
-> objective instance factory
-> supplied progress, statistic, attribution, gameplay, aggregation, and timer facts
-> objective-local state-machine evaluation
-> lifecycle/progress/success/failure/cancellation/retirement/discovery events
-> visibility-aware objective snapshots
-> consumer-owned activation, composition, scheduling, rewards, results, or presentation
```

### Progress, Aggregation, And Attribution

Progress systems generate progress facts. Aggregation systems derive shared values, including team progress from player progress updates. Attribution systems produce one-hit, in-game, in-encounter, or other resolved attribution facts. The Objective Foundation consumes these facts and applies definition conditions without becoming an alternate source or aggregation authority.

### Gameplay And Statistics

Gameplay owners produce authoritative events and statistics. Objectives may consume those facts through a definition-defined condition input, while runtime entities, combat, counters, and gameplay statistics remain owned by their respective systems.

### Timers And Simulation Time

The timer/simulation-time seam supplies authoritative time progression and the simulation-paused condition. Objective timers own their local timer state and report expiry. Ordinary presentation pause does not stop an objective timer; simulation-time pause caused by all players being disconnected does.

### Consumers, Composition, And Orchestration

Consumers own activation, dependencies, sequencing, branching, optionality, repeatability, scheduling, and orchestration. They may compose Single Objectives into Multiple Objectives and Multiple Objectives into Meta-Objectives. The foundation supplies APIs and state abstractions for those consumers; it does not make those policies universal.

### Teams And Multiplayer Lifecycle

Teams owns membership, team relationships, joining, and forfeiture. Lifecycle owns connection, disconnection, reconnect, and participation transitions. The Objective Foundation consumes player/team facts and preserves instance state across reconnect; it does not assign teams, admit players, or decide forfeiture.

### Rewards And Progression

Rewards and progression systems own award calculation, durable progression, and persistence. They may consume objective events or snapshots, but objective state does not grant rewards or become account, campaign, season, challenge, or achievement state by default. Achievements remain a separate system.

### Match Rules And Results

Match rules own match state and match-end decisions. Match result inclusion and interpretation are mode-owned. A mode may select objective facts for results, but the Objective Foundation does not request match-state changes, end a match, or interpret result meaning.

### Player Experience

Player experience consumes authorized visibility-aware snapshots and objective events for presentation. It may display discovery, progress, timers, outcomes, and retirement, but it cannot advance, resolve, reveal, cancel, or reset objectives authoritatively.

## Devtools Boundary

The initial Objective Foundation devtools surface supports:

```text
force progress
set progress
activate
discover
complete
fail with reason
cancel
retire
reset
```

These actions exercise the foundation's state-machine seam and are validated against definition and instance state. Devtools do not acquire administrative ownership, bypass consumer policy, or become a separate authority for objectives.

## Implementation Status

The first server implementation slice is complete.

Implemented:

```text
schema-driven immutable objective definitions
atomic objective instances with independent definition and instance identity
player, team, match, collection, and definition-specific scopes
boolean, numeric, set, sequence, maintain-condition, manual, and timer inputs
supplied fact consumption with one-hit, in-game, and in-encounter attribution filters
undiscovered, discovered, inactive, active, completed, failed, cancelled, and retired lifecycle states
success-over-failure resolution at one evaluation boundary
timer-expiry event ordering and default timer_expired failure
objective timers that ignore ordinary world freeze and pause with no connected players
visibility-aware owner/public/discovery snapshots
multiple active instances of one definition
unattainable-definition retirement and new-instance blocking
Single, Multiple, and Meta composition contracts without inheritance
objective event observers and deterministic snapshots
award-counter facts feeding matching objective scopes
state preservation across player removal/disconnection
force progress, set progress, activate, discover, complete, fail, cancel, retire, and reset devtools actions
```

The foundation does not register objectives for the Arcade Survival baseline. Activation, scheduling, dependency graphs, branching, repeatability policy, rewards, match-end interpretation, result inclusion, campaign sequencing, persistence, packet schemas, and presentation remain consumer-owned future work around this seam.

## Implementation Direction

The implemented first slice proceeds from definitions and supplied facts into local objective state:

```text
1. Define the schema/definition contract for condition, progress, timer, lifecycle, failure, discovery, visibility, and association behavior.
2. Define the atomic Single Objective state-machine factory and instance identity.
3. Define composition abstractions for Multiple Objectives and Meta-Objectives without requiring inheritance.
4. Consume planned progress, statistic, aggregation, attribution, gameplay, and timer facts.
5. Track only objective-local state/statistics required by the definition.
6. Implement the conceptual lifecycle, required failure reasons, timer-expiry event, and success-over-failure tie rule.
7. Preserve state across disconnect/reconnect and begin personal instances at zero.
8. Retire unattainable definitions and all active instances, block new instances, and require replacement definitions to use new IDs.
9. Emit lifecycle, progress, discovery, visibility-aware snapshot, and resolution events using the existing event/progress infrastructure guarantees.
10. Add the initial foundation devtools actions without adding admin ownership.
```

Consumers should build activation, sequencing, branching, optionality, repeatability, scheduling, rewards, results, and presentation around this seam. Implementation should keep the foundation focused on schema-driven state-machine evaluation rather than creating a universal objective, campaign, progression, or persistence runtime.

## Testing Direction

Important future checks:

```text
definitions construct objective instances through a schema-driven state-machine factory
Single Objectives compose into Multiple Objectives, and Multiple Objectives compose into Meta-Objectives
composition does not require language-level inheritance
personal, team, match, collection, and definition-specific scopes are representable
personal instances begin at zero in the first implementation
shared progress is external aggregate state
team progress consumes aggregation derived from player progress updates
boolean, numeric, timed, staged/sequence, collection/set, survival/maintain-condition, and schema-defined forms are supported
progress increase, decrease, reset, regression, and overflow follow definition/consumer policy
objectives consume supplied facts and do not generate a separate progress pipeline
one-hit, in-game, and in-encounter attribution facts are consumed without objective-owned attribution production
undiscovered, discovered, inactive, active, completed, failed, cancelled, and retired states are distinguishable
not every objective requires every lifecycle state
definition-owned visibility prevents undiscovered state from leaking in snapshots
every failure has a failure reason
timer expiry emits before defaulting to failed with failure_reason timer_expired
definition-specific timer transitions can replace the default failure transition
success wins when success and failure become true in one evaluation boundary
non-failable objectives cannot fail through ordinary condition evaluation
timers continue through ordinary simulation pause
timers stop only when simulation time is paused because all players are disconnected
disconnect/reconnect preserves objective instance state
definitions are effectively immutable and have no version field
unattainable definitions retire all active instances, block new instances, and require a replacement definition ID
retired instances do not finish
multiple instances of one definition can be active simultaneously
objective_instance_id and definition-owned association fields distinguish instances
existing event/progress infrastructure supplies idempotency and deduplication
objective runtime does not redefine event/progress deduplication
hot/broad updates can batch while cold/narrow updates can emit individually
initial devtools actions are limited to the stated foundation surface
devtools do not add admin ownership
match result inclusion and interpretation remain mode-owned
achievements remain separate
Arcade Survival remains a no-objective mode baseline
objective state does not own progress generation, aggregation, attribution, scheduling, rewards, persistence, match state, match end, campaign structure, team membership, joining, forfeiture, results, or presentation
```

## Related Docs

- [Gameplay Planning](./!INDEX.md)
- [Modes And Match Rules](modes-and-match-rules.md)
- [Gameplay Awards And Counters](gameplay-awards-and-counters.md)
- [Teams And Team Rules](teams-and-team-rules.md)
- [Lives, Death, Elimination, And Respawn](lives-death-elimination-and-respawn.md)
- [Levels, Missions, And Content Structure](levels-missions-and-content-structure.md)
- [Match Outcomes And Results](match-outcomes-and-results.md)
- [Progression And Rewards](progression-and-rewards.md)
- [Achievements And Milestones](achievements-and-milestones.md)
- [Player Experience Systems](player-experience-systems.md)
- [Multiplayer Session And Lifecycle](../platform/multiplayer-session-and-lifecycle.md)

## Remaining Implementation-Level Decisions

- Exact schema, definition, instance, state, event, snapshot, and devtools field names.
- Exact condition-input and supplied-fact contracts for progress, statistics, aggregation, attribution, gameplay, and timers.
- Exact composition representation for Single Objectives, Multiple Objectives, and Meta-Objectives.
- Exact scope and definition-owned association-field representation.
- Exact local statistics retained for each schema-defined condition form.
- Exact lifecycle transition validation and state persistence representation.
- Exact failure-reason vocabulary beyond `timer_expired`.
- Exact definition-owned visibility and discovery projection rules.
- Exact timer clock, simulation-pause detection, expiry ordering, reset, and overflow mechanics.
- Exact event emission and batching thresholds for cold/narrow versus hot/broad updates.
- Exact integration contract with the existing event/progress idempotency and deduplication infrastructure.
- Exact retirement and replacement-definition registration flow.
- Exact consumer APIs for activation, dependencies, sequencing, branching, optionality, repeatability, scheduling, and orchestration.
- Exact devtools command/request shape and validation errors.
- Exact mode-owned result selection and interpretation handoff.
- Exact package, packet, storage, and persistence boundaries chosen at implementation time.

There are no remaining product-level Objective Foundation decisions blocking P4 planning.

## Core Invariants

```text
The Objective Foundation is a schema/definition-driven state-machine factory/runtime.

Objectives consume supplied condition, progress, statistic, aggregation, attribution, gameplay, and timer facts.

Objectives track only objective-local state/statistics required by their definitions.

Progress generation, aggregation, attribution production, scheduling, rewards, persistence, match state, match end, campaign structure, team membership, joining, forfeiture, result interpretation, achievements, and presentation remain outside Objective Foundation ownership.

Single Objectives are atomic; Multiple Objectives compose atomic objectives; Meta-Objectives compose Multiple Objectives.

Composition is not a fixed initial type list or a language-level inheritance requirement.

Player, team, match, collection, and definition-specific scope are supported.

Team progress is derived from player progress updates by a separate aggregation/progress owner.

Definitions own condition forms, progress behavior, failure capability, timer behavior, reset, repeatability, overflow, and visibility policy where applicable.

The conceptual lifecycle includes undiscovered, discovered, inactive, active, completed, failed, cancelled, and retired; not every objective uses every state.

Every failure has a failure reason.

Timer expiry emits an event and defaults to failed/timer_expired unless the definition specifies another valid transition.

Success wins over failure when both become true in one evaluation boundary.

Timers continue through ordinary simulation pause and stop only while simulation time is paused because all players are disconnected.

Personal instances begin at zero in the first implementation, and disconnect/reconnect preserves instance state.

Definitions are effectively immutable and have no version field.

Unattainable definitions and all active instances retire; new instances are blocked; replacement definitions use new IDs; retired instances do not finish.

Multiple instances of one definition may be active simultaneously and are distinguished by objective_instance_id plus definition-owned association fields.

Idempotency and deduplication come from separate planned event/progress infrastructure; the Objective Foundation consumes, rather than redefines, those guarantees.

Objective updates may emit individually or in batches according to cold/narrow versus hot/broad update characteristics.

Initial devtools exercise foundation state without adding administrative ownership.

Arcade Survival remains the no-objective baseline by mode policy.

No product-level Objective Foundation decisions remain blocking P4 planning.
```