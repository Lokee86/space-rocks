# Objectives And Objective Runtime
Parent index: [Gameplay Planning](./!INDEX.md)

## Purpose

This doc is the authoritative P4 planning owner for match-objective definitions and objective runtime semantics.

It defines how a resolved mode and content package describe objectives, how authoritative runtime state tracks objective lifecycle and progress, how objective facts become visible, and how objective outcomes hand off to match integration, results, and future content orchestration. It defines policy boundaries without claiming that the described system or implementation already exists.

## Ownership Boundary

This doc owns planning for:

```text
match-objective definitions
objective runtime lifecycle
objective status and transition semantics
objective scope and participation
objective progress, target, and timer semantics
objective sequencing, dependency, and optionality
objective visibility, reveal, addition, replacement, and removal
objective-specific facts and normalized snapshots
objective requests and typed request validation boundaries
objective resolution idempotency
objective-to-award and counter handoffs
objective-to-match-state and match-end evaluation requests
late-join and reconnect objective-state handoff
objective result requirements and resolution facts
```

This doc does not own:

```text
mode selection or general match-rule resolution
combat simulation, collision detection, or gameplay entity ownership
award calculation or gameplay counter mutation
team membership or team relationship semantics
lives, death, elimination, or respawn policy
match locking, match-end policy, or result orchestration
account, campaign, season, challenge, or achievement progression
persistence or durable account state
client UI layout, animation, or presentation effects
packet schemas, storage schemas, or package names
```

Modes and match rules select the objective model, objective definitions, and related policy as part of resolved rules. The objective runtime validates and evaluates those definitions, maintains authoritative objective state, and emits normalized facts or typed requests. Owning systems execute their own policy rather than becoming alternate objective authorities.

## Settled Product Model

Arcade Survival has no objective. It remains the baseline mode and must not acquire implicit objective completion, failure, progress, or match-end behavior merely because the objective runtime exists.

The objective model supports multiple simultaneous objectives. Multiple active objectives are required for story and campaign missions and are not treated as an exceptional extension. A mission may expose a primary objective alongside optional, hidden, or bonus objectives, and multiple objectives may progress or resolve during the same simulation tick.

Supported objective classifications and lifecycle features include:

```text
primary
optional
bonus
hidden
revealed
sequential
simultaneous
```

Objectives may be activated, completed, failed, suspended, removed, replaced, hidden, or revealed according to resolved mode and content policy. Sequential objectives may depend on another objective's resolution. Optional and bonus objectives do not automatically become match-ending requirements. Hidden objectives may exist authoritatively before their client-visible reveal.

Success, failure, and partial-completion facts or states are supported. Partial completion is objective-defined: it may represent an intermediate milestone, a useful result state, a contribution threshold, or another explicitly resolved condition. Partial completion does not universally imply success, failure, activation of a successor, or match end.

Objective scope is mode-defined and may be:

```text
player
team
match-wide
```

A player-scoped objective has separate authoritative state for each eligible player. A team-scoped objective has authoritative team state supplied by team rules. A match-wide objective has one shared state. Scope is part of the definition and is not inferred from the display recipient or projection.

Progress behavior is objective-defined. Progress may increase, decrease, reset, be contested, or be stolen. A client-visible progress bar is only a projection of authoritative objective state; it does not imply monotonic progress. An objective may define a target where one is meaningful, but not every objective requires numeric progress or a target.

Timed objectives are supported. A timer may measure a duration, impose a deadline, or control a phase of objective eligibility. Timer expiration produces an authoritative objective event that the objective definition interprets as success, failure, partial completion, pause, reset, or another supported transition.

Objective completion or failure may request a match-state update or match-end evaluation. Objective logic never directly locks or ends the match. Matches may continue after primary objective success or failure, including for niche debug or administrative use. Match rules and outcome orchestration decide whether a request changes match state, evaluates an end condition, or has no immediate match-ending consequence.

Tie and simultaneous-completion handling belongs to a mode-specific sub-policy where relevant. The objective runtime preserves the authoritative event and resolution facts needed by that sub-policy; it does not invent one universal tie rule for every mode.

## Objective Definitions And Runtime State

The planning model separates immutable definitions from mutable authoritative state.

`ObjectiveDefinition` is immutable for the lifetime of the definition instance. It describes identity, classification, scope, condition and progress semantics, timing, visibility defaults, dependencies, and the typed requests or handoffs permitted by the selected mode/content policy. Replacing a definition creates a new definition instance or a new resolved objective generation; it does not mutate the meaning of historical resolution facts.

`ObjectiveState` is mutable authoritative runtime state. It records the current lifecycle status, progress, target where applicable, timer state, reveal state, dependency state, contributors, and resolution facts for one objective instance and scope. It is owned by the objective runtime rather than by a client, runtime entity, award counter, team display row, or persistence system.

A definition may be reused as content, but each match objective instance receives an authoritative identity and lifecycle. Objective identity must distinguish the definition/content identity from the match-instance identity when that distinction is needed for replacement, replay, result, or idempotency handling.

The runtime should normalize condition evaluation rather than allowing each objective implementation to mutate arbitrary match state. Objective-specific condition adapters may consume authoritative events, award/counter facts, team facts, lifecycle facts, entity facts, timers, and mode context. They return objective facts or typed requests through the objective seam.

## Objective Lifecycle And Sequencing

An objective lifecycle may include these conceptual states:

```text
pending
active
partially-complete
succeeded
failed
suspended
removed
replaced
```

The exact status vocabulary remains implementation-level, but the runtime must distinguish active evaluation from a resolved terminal outcome and from an objective that is no longer present. A removed or replaced objective may retain a historical final snapshot even when it is no longer visible in the current objective set.

Sequential objectives are content- and mode-defined. A successor may activate after a predecessor succeeds, fails, partially completes, or reaches another explicitly named state. Optional objectives may activate independently or alongside a primary sequence. Hidden objectives may evaluate authoritatively before reveal when the definition permits it, but their hidden facts must not leak through unauthorized projections.

Dynamic orchestration may add, remove, replace, hide, or reveal objectives during a match. Addition and replacement must preserve unambiguous objective identity and generation facts. Removal is not silently equivalent to success or failure; the resolution reason or lifecycle reason identifies what happened. A mode may define whether a removed or replaced objective contributes to results, progression handoff, or later sequencing.

## Progress, Conditions, And Timers

Objective conditions are authoritative and objective-defined. A condition may consume an event stream, normalized gameplay facts, completed counter mutations, entity facts, lifecycle facts, timer transitions, or another objective's state. Conditions must state the scope and eligibility rules they use rather than relying on the client-visible participant list.

Progress is a semantic value, not necessarily a counter with one universal operation. An objective may define:

```text
incremental progress
regressive progress
a resettable meter
contested progress
stealable progress
threshold progress
boolean state
non-numeric milestone state
```

When progress is numeric, the definition specifies the valid target and operation. When progress is contested or stealable, the runtime preserves the authoritative owner, contesting parties, contribution, and transition facts needed by the selected mode. Progress mutation must be validated against objective state and scope; clients cannot submit arbitrary progress values.

Timers are authoritative objective state. The runtime records start time, relevant pause/resume or reset facts, deadline or duration where applicable, and resolution time. The simulation clock and tick boundary used for timer evaluation must be consistent with match rules. Presentation may display a countdown or elapsed timer but cannot decide expiration.

## Typed Objective Requests And Facts

Objectives may cause entity spawn, removal, or modification only through typed `ObjectiveRequest` values. Objective logic does not directly mutate game entities. A request identifies the requested operation, objective instance, source transition or event identity, scope, target or entity reference where applicable, and the policy context needed by the owning gameplay system.

The request catalogue is mode/content-defined and must remain typed rather than becoming an arbitrary command or generic mutation escape hatch. Examples may include requests to:

```text
spawn an objective-owned encounter or entity set
remove an eligible objective-owned entity or entity set
modify an explicitly supported entity property
request an objective phase or match-state update
request match-end evaluation
```

The objective runtime validates request shape, objective authorization, lifecycle status, target scope, and idempotency. The owning gameplay system validates and executes the request according to its own rules. Invalid, stale, duplicate, or post-lock requests are rejected authoritatively and do not partially mutate entities.

`ObjectiveFacts` is the normalized authoritative fact stream produced by objective evaluation. It includes common facts and may carry typed objective-specific facts. A final objective snapshot is the complete authoritative view required by results, match integration, and authorized projections.

Common objective facts include:

```text
objective identity and definition/generation identity
status and lifecycle transition
scope and scoped owner
visibility and reveal state
start time and resolution time
progress and target where applicable
timer start, deadline, duration, or elapsed value where applicable
resolution reason
responsible player or team where applicable
contributors and contribution facts where applicable
source event or transition identity
```

Typed facts may describe contest ownership, stolen progress, milestone data, entity-request outcomes, dependency transitions, or mode-specific tie results. Typed facts must extend the normalized contract without forcing unrelated consumers to understand every objective type.

## Tick Resolution And Idempotency

Objective transitions resolve once at the end of the simulation tick, after authoritative gameplay events and completed award/counter mutations. This ordering ensures that objective conditions consume the authoritative result of the tick rather than a partial client projection or an intermediate counter state.

The intended boundary is:

```text
authoritative gameplay simulation/events
-> award attribution and completed counter mutations
-> objective condition/progress evaluation
-> objective transitions and typed request creation
-> typed request validation/execution by owning systems
-> match-state update requests
-> match-end evaluation
-> replication/result handoff
```

The exact internal ordering of request execution and match-state application remains an implementation-level decision, but objective requests are processed before match-end evaluation. Requests submitted after match lock are rejected. Objective logic never bypasses the lock by directly changing room or match lifecycle state.

Transitions are idempotent. A completed or failed objective cannot resolve twice. Duplicate source events, duplicate request delivery, repeated evaluation, reconnect replay, and late client messages must not create a second completion, failure, award handoff, entity mutation, or match-state request. Idempotency applies to the objective transition and its derived requests/facts, not only to client presentation.

## Visibility And Client Projections

The client receives visibility-aware projections of authoritative objective state. A projection may include, according to mode and viewer authorization:

```text
identity and description
status
progress and target where permitted
timer or deadline
primary, optional, or bonus classification
scope
reveal state
addition, replacement, or removal events
```

Visibility is separate from ownership and mutation. A hidden objective may remain authoritative and may affect mode evaluation without exposing its identity, description, status, progress, timer, contributors, or resolution reason. Projection code must not reveal hidden-objective existence through list size, target values, event ordering, error text, or an indirect award/result field unless the resolved policy permits it.

Revealing an objective is an authoritative visibility transition. Addition, replacement, and removal are also authoritative lifecycle changes and must be projected in a way that lets clients converge on the current objective set without reconstructing history from UI events.

Late joiners receive current match-scoped objective state that is visible to them under the resolved policy. They do not need the full historical event stream to reconstruct current state. When reconnect is implemented, reconnect restores prior player-specific objective state rather than resetting it, unless an explicit mode policy says otherwise. Hidden and unauthorized player-specific state remains hidden on join and reconnect.

## Awards, Counters, And Objective Progress

Objective progress and gameplay counters remain distinct. The awards/counters system owns gameplay award calculation and counter mutation. Objective runtime owns the meaning of objective progress and decides whether a completed counter mutation satisfies, advances, contests, resets, or otherwise affects an objective.

An objective may emit an award-source fact or consume an award/counter fact. It must not directly mutate `SCORE`, `KILLS`, or another gameplay counter. Conversely, `OBJECTIVE_PROGRESS` may record configured progress but does not grant the awards/counters system authority to decide objective completion.

Objective completion, failure, partial completion, contributors, and responsible participants may become inputs to award policy. Any award is resolved through the authoritative awards pipeline and remains separate from objective resolution. Objective state is not ranking, match-end, result, progression, achievement, or durable account state.

## Match Integration And Result Handoff

Objective completion and failure may emit a match-state update request or a request for match-end evaluation. These are requests, not direct lifecycle mutations. Match rules and outcome orchestration decide whether the match continues, changes phase, becomes eligible for ending, or ends according to resolved policy.

A match may continue after primary objective success or failure. This supports modes with follow-on objectives, post-objective play, campaign presentation, debug scenarios, and administrative control. Primary objective resolution is not a universal match lock or end trigger.

Objective runtime emits complete authoritative resolution facts and a final objective snapshot. Mode/result policy selects which facts enter player-facing results and which are handed to persistence or progression consumers. Result consumers must not reconstruct completion, ordering, contributors, or responsible players from client events or from current runtime entities after the match is locked.

The objective snapshot may include historical objectives that were removed or replaced when the selected result policy requires them. The current visible objective projection and the final result snapshot are different products with different authorization and retention needs.

## Reusable Seam Without A Universal Meta-System

The objective lifecycle, condition, counter, timer, visibility, and normalized-fact primitives may later be reused by campaign tracking, daily or weekly challenges, and season journeys where their ownership contracts fit.

That reuse is a seam, not a requirement that every consumer use one runtime, one definition language, or one persistence model. Achievements remain a separate meta-system. Account, campaign, and season progress and durable persistence remain separate meta-system ownership. A match objective may provide an eligible handoff to those systems, but it does not own their durable state, account identity, cross-match aggregation, challenge scheduling, achievement rules, or persistence writes.

Campaign/story orchestration may dynamically add, remove, replace, hide, or reveal match objectives. It owns content sequencing and campaign context; the objective runtime owns the authoritative lifecycle and facts for the active match objective instances.

## System Handoffs

```text
resolved mode/content rules
-> immutable ObjectiveDefinition instances
-> authoritative ObjectiveState per objective scope
-> gameplay events, team/lifecycle facts, entity facts, timers, and completed award/counter facts
-> objective condition/progress evaluation
-> ObjectiveFacts and final objective snapshot
-> idempotent typed ObjectiveRequest values
-> owning gameplay-system validation/execution
-> match-state update request or match-end evaluation request
-> MatchDecision / MatchSummary and selected result facts
-> authorized player projections and progression/meta-system handoffs
```

### Modes And Match Rules

Modes select whether objectives exist, objective definitions, scope, classifications, dependencies, progress/timer semantics, visibility, tie handling, allowed requests, continuation after resolution, match-state request policy, and result requirements. Modes do not delegate match locking or ending to objective logic and do not infer objective semantics from client presentation.

### Gameplay Awards And Counters

Awards and counters own authoritative award calculation, attribution, and counter mutation. Objective runtime consumes completed award/counter facts and may emit objective-sourced award inputs. Neither system becomes an alternate authority for the other system's progress or counter meaning.

### Teams And Team Rules

Teams owns membership, team assignment, relationships, and normalized team facts. Objective runtime consumes those facts for team-scoped objective evaluation and contributors. It does not assign teams, aggregate team membership independently, or redefine team elimination.

### Lives, Death, Elimination, And Respawn

Lives/death/respawn owns ship death, life accounting, elimination, respawn, and recovery transitions. Objective runtime consumes the relevant authoritative lifecycle facts and may define an objective condition or request based on them, but it does not grant lives, respawn players, or decide elimination directly.

### Content And Campaign Orchestration

Content and campaign systems own mission composition, story sequencing, unlock context, and when a new content-defined objective should be added, removed, replaced, hidden, or revealed. Objective runtime validates and evaluates the active objective instance and emits normalized facts without becoming the campaign or account progression owner.

### Gameplay Entity Owners

Combat, spawning, encounter, entity, and runtime owners own entity creation, removal, modification, simulation, and validation. Objective runtime emits typed requests only. Entity owners reject invalid, stale, unauthorized, duplicate, or post-lock requests and return authoritative request outcomes or facts.

### Match Outcomes And Results

Match rules and outcome orchestration own match-state transitions, match-end evaluation, match lock, `MatchDecision`, `EndOfMatchFlow`, `MatchSummary`, and result emission. Objective requests are processed before match-end evaluation. Objective runtime supplies complete facts; outcome policy selects whether objective states affect ending and results.

### Progression, Challenges, And Achievements

Progression, challenges, daily/weekly systems, season journeys, and achievements own their own eligibility, aggregation, durable state, and persistence. They may consume selected objective result facts. Achievements remain separate, and no match objective becomes an account or season tracker by default.

### Player Experience

Player experience consumes authorized objective projections, including identity/description, status, progress/target, timers, classification, scope, and lifecycle changes. It may collect objective-related input where a mode permits it, but it cannot advance, resolve, reveal, add, remove, or replace objectives authoritatively. Hidden-objective presentation must not leak unauthorized facts.

## Implementation Direction

The first implementation slice should proceed from resolved definitions to authoritative, end-of-tick objective facts:

```text
1. Define immutable ObjectiveDefinition and scoped mutable ObjectiveState shapes.
2. Define objective identity, status, visibility, progress, target, timer, and resolution facts.
3. Define condition/progress adapters for authoritative gameplay, team, lifecycle, entity, timer, and award/counter facts.
4. Queue typed ObjectiveRequest values and validate them before handing execution to owning gameplay systems.
5. Resolve objective transitions once at the end of the simulation tick after completed award/counter mutations.
6. Make transitions and derived requests/facts idempotent; reject duplicate terminal resolution and post-lock requests.
7. Process objective requests before match-end evaluation and preserve continuation after primary objective resolution.
8. Expose visibility-aware current-state projections for active, hidden, revealed, added, replaced, and removed objectives.
9. Provide current match-scoped state to late joiners and preserve player-specific state across reconnect when reconnect exists.
10. Emit complete final objective facts for mode/result selection without making the snapshot persistence-specific.
11. Preserve reusable lifecycle/condition/counter/timer seams without requiring one universal meta-runtime or persistence model.
```

Implementation should keep objective policy in a focused gameplay owner. Game-loop coordination, networking, entity systems, awards, teams, lives, match outcomes, content orchestration, progression, achievements, persistence, and UI should route facts or execute their own policy rather than becoming alternate objective authorities.

## Testing Direction

Important future checks:

```text
Arcade Survival has no objective and remains the baseline
multiple simultaneous objectives are supported
story/campaign missions can require multiple simultaneous objectives
sequential, optional, hidden, bonus, and revealed objectives are distinct
success, failure, and partial-completion states/facts are representable
objective scope can be player, team, or match-wide
progress may increase, decrease, reset, be contested, or be stolen when defined
numeric targets are optional and objective-defined
timed objectives use authoritative start/deadline/duration and tick evaluation
immutable ObjectiveDefinition is distinct from mutable ObjectiveState
objective identity distinguishes active instances and replacement generations where needed
objective conditions consume authoritative facts, not client projections
transitions resolve once at end of tick after gameplay events and completed award/counter mutations
completed and failed objectives cannot resolve twice
queued typed ObjectiveRequest values are validated and idempotent
objective requests cannot directly mutate entities
entity owners validate and execute objective requests
requests after match lock are rejected
objective requests are processed before match-end evaluation
objective completion/failure may request match-state update or match-end evaluation without directly locking or ending the match
matches may continue after primary objective success/failure
mode-specific tie policy handles simultaneous completion where relevant
common ObjectiveFacts include identity, status, scope, visibility, times, progress/target, reason, responsibility, and contributors where applicable
typed objective-specific facts extend common facts without forcing every consumer to understand them
hidden objectives do not leak identity, progress, timers, contributors, existence, or resolution through unauthorized projections
reveal, addition, replacement, and removal converge current client state
late joiners receive current visible match-scoped objective state
reconnect restores prior player-specific objective state when reconnect is implemented
objective progress remains distinct from gameplay counters, ranking, match end, results, and durable progression
objective-sourced awards use the awards pipeline and do not make awards the objective authority
final objective snapshot contains complete authoritative resolution facts
mode/result policy selects player-facing result facts and progression/persistence handoffs
achievements remain separate from match-objective ownership
campaign, challenge, season, account, and durable persistence owners remain separate
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

- Exact `ObjectiveDefinition`, `ObjectiveState`, `ObjectiveFacts`, final-snapshot, and `ObjectiveRequest` type and field names.
- Exact objective identity, definition identity, generation, and replacement identity rules.
- Exact status vocabulary for pending, active, partial, succeeded, failed, suspended, removed, and replaced states.
- Exact condition representation and adapter contract for gameplay, awards/counters, teams, lifecycle, entities, timers, and other objectives.
- Exact progress numeric types, target representation, decrement/reset/contest/steal operations, and bounds.
- Exact timer clock, tick ordering, pause/resume, reset, deadline, and expiration details.
- Exact request catalogue, authorization rules, validation errors, execution result facts, and retry behavior.
- Exact event adapters and source-event identity composition.
- Exact end-of-tick scheduling and conflict precedence when multiple objective transitions or requests interact.
- Exact simultaneous-completion and tie-policy contract for modes that require one.
- Exact visibility authorization, hidden-objective projection, reveal event, and current-state replication shape.
- Exact late-join snapshot and reconnect restoration contract for match-scoped and player-specific state.
- Exact result, progression, challenge, campaign, season, and persistence projection fields selected by their owners.
- Exact package, runtime-storage, packet, and persistence boundaries chosen at implementation time.

There are no remaining product-level match-objective decisions blocking P4 system planning.

## Core Invariants

```text
Objectives And Objective Runtime is the authoritative planning owner for match-objective definitions, lifecycle, progress, visibility, requests, facts, and match integration.

Arcade Survival has no objective and remains the baseline.

Multiple simultaneous objectives are supported and required for story/campaign missions.

Sequential, optional, hidden, bonus, and dynamically revealed objectives are supported.

Success, failure, and partial-completion facts/states are supported.

Objective scope is mode-defined and may be player, team, or match-wide.

Progress behavior is objective-defined and may decrease, reset, be contested, or be stolen.

Timed objectives use authoritative runtime state and tick evaluation.

ObjectiveDefinition is immutable; ObjectiveState is mutable authoritative runtime state.

Objective transitions resolve once at the end of the simulation tick after authoritative gameplay events and completed award/counter mutations.

Transitions and derived requests/facts are idempotent; completed or failed objectives cannot resolve twice.

ObjectiveRequest is queued and typed. Objective logic never directly mutates game entities, room lifecycle, rewards, persistence, or UI.

Entity and gameplay owners validate and execute objective requests.

Objective requests are processed before match-end evaluation, and requests after match lock are rejected.

Objective completion or failure may request match-state update or match-end evaluation but never directly locks or ends the match.

Matches may continue after primary objective success or failure.

Tie and simultaneous-completion handling belongs to a mode-specific sub-policy where relevant.

ObjectiveFacts and the final objective snapshot contain complete authoritative resolution facts, including common facts and typed objective-specific facts.

Client projections are visibility-aware; hidden objectives cannot leak through direct or indirect presentation facts.

Late joiners receive current visible match-scoped objective state, and reconnect restores prior player-specific state when reconnect is implemented.

Objective progress is distinct from gameplay counters, ranking, match end, results, and durable progression.

Awards/counters, teams, lives/respawn, entity systems, content orchestration, outcomes/results, progression/challenges, achievements, persistence, and player experience retain their ownership boundaries.

Reusable objective lifecycle/condition/counter/timer primitives do not require every campaign, challenge, achievement, season, or progression consumer to use one runtime or persistence model.

No product-level match-objective decisions remain blocking P4 system planning.
```
