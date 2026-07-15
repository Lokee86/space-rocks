# Gameplay Awards And Counters
Parent index: [Gameplay Planning](./!INDEX.md)

## Purpose

This doc is the authoritative P4 planning owner for gameplay awards and counters.

It defines how authoritative gameplay events become attributed awards, counter mutations, assists, combos, multipliers, streaks, and team distributions. It also defines the runtime ownership, visibility, idempotency, and handoff boundaries required for those facts to remain consistent through match resolution.

This document is a planning specification. It does not claim that the described systems or implementation already exist.

## Ownership Boundary

This doc owns planning for:

```text
award catalogue and award-source semantics
counter ownership scope
attribution and contribution history
assist eligibility and assist awards
combo and multiplier policy
streak policy and future trigger hooks
team award distribution
counter mutation operations
counter lifecycle and reset boundaries
award-event idempotency
award and counter visibility policy
runtime ownership of player, team, objective, and match counters
final award/counter snapshot handoff
```

This doc does not own:

```text
combat simulation or collision detection
objective rules or objective completion policy
ranking or leaderboard calculation
match-end policy or match decision ownership
result orchestration or persistence
progression, achievements, or account rewards
HUD, scoreboard, result, or spectator layout
packet schemas or storage schemas
room membership, connection, removal, or reconnect execution
```

The gameplay award system consumes authoritative gameplay outcomes and resolved mode policy. It produces normalized award, counter, attribution, and final-snapshot facts. Objective, ranking, match-end, results, progression, and presentation systems consume those facts through their own seams rather than reimplementing award policy.

## Settled Product Model

The initial fixed counter catalogue is:

```text
SCORE
KILLS
ASSISTS
DEATHS
DAMAGE_DEALT
DAMAGE_TAKEN
OBJECTIVE_PROGRESS
RESOURCES_COLLECTED
COMPLETION_TIME
```

The system supports custom numeric counters through an extension seam. Custom counters must use the same authoritative ownership, mutation, visibility, idempotency, lifecycle, and snapshot rules as catalogue counters. The extension seam must not become an alternate path for arbitrary presentation or persistence state.

A counter may be owned by:

```text
player
team
match-wide scope
objective-specific scope
```

Ownership scope is selected by the resolved mode, objective, or award definition. A counter's scope is not inferred from its display location. A player scoreboard row may display a team or objective counter without making the client its owner.

Gameplay awards represent match-scoped authoritative credit. They are distinct from objective progress, ranking, match-end and result facts, progression rewards, and presentation effects. An award may mutate SCORE or another gameplay counter while a separate owner decides whether that value contributes to an objective, ranking, match outcome, result row, or progression grant.

Counter lifecycle follows normal match or round ownership. Counters reset at the boundary selected by that owner. The architecture does not support or require generic cross-round score persistence.

Removed players retain earned counters and historical attribution, but stop receiving normal new awards after removal. Delayed events are accepted or rejected using authoritative attribution and lifecycle rules rather than client timing.

## Award Catalogue And Sources

Awards may update one or more counters according to resolved mode policy. The fixed catalogue provides the standard vocabulary; it does not require every mode to expose every counter.

Supported award sources include:

```text
destruction
impacts, hits, or collisions
objectives
survival or time milestones
pickups and resources
mode-specific events
```

Destruction awards apply when an authoritative destruction outcome is resolved. Asteroid awards continue scaling by target size. Award values are tunable, and modes may override the calculation. A mode override must be explicit in resolved rules rather than inferred from a mode name.

Impacts, hits, and collisions may award damage, hit, or mode-specific credit even when they do not destroy a target. Combat and collision owners supply authoritative outcomes and candidate sources; the award owner decides the configured award consequences.

Objective-sourced awards consume objective outcome facts. Awards/counters emit authoritative facts; separate progress and aggregation owners produce progress facts, and the Objective Foundation consumes those facts while owning definition condition evaluation and objective-local lifecycle/state. An objective award can mutate a gameplay counter or emit a handoff without making gameplay counters the objective authority.

Survival/time milestones, pickups/resources, and mode-specific events enter through the same authoritative award pipeline. They must identify their source, event identity, owner scope, and resolved policy so that they cannot bypass attribution or idempotency.

Award values and calculations are tunable. Modes may select different values, scaling functions, eligible recipients, target counters, penalties, or distribution policies. The award system owns the common pipeline and validation; mode rules own explicit policy selection.

## Attribution And Contribution

Projectile ownership is authoritative for hit, impact, and collision attribution. The owner of a projectile receives the eligible hit or impact attribution for that projectile under the resolved mode rules.

Final destruction credit goes to the last valid hit. A last hit is valid only when its source, owner, timing, target, and lifecycle status satisfy authoritative attribution rules. Clients must not reconstruct final destruction credit from visible projectile order or effect timing.

Contribution history records the eligible source contributions needed for assists and other configured attribution. It is retained only for the tunable assist window plus approximately a 10% timing buffer. The buffer protects authoritative processing from small timing differences without turning contribution history into indefinite match history.

Contribution records must carry enough context to reject stale, invalid, removed, or duplicated sources. Exact record fields and storage are implementation-level decisions. Historical award and counter facts remain available through the match snapshot even after short-lived contribution history expires.

Attribution must distinguish direct credit, assist credit, self-caused outcomes, environmental or unattributed outcomes, and any mode-specific category required by the selected rules. This system does not own combat cause detection; it normalizes the award consequences of authoritative cause facts.

## Assists

Assists are mode-enabled. When enabled, the initial eligibility threshold is 5% contribution within an initial 5-second contribution window. Multiple assistants are allowed.

The assist award is mode-selected and will likely be SCORE. An assist affects match scoring only when the selected mode uses SCORE. An assist may instead update ASSISTS or another configured counter without implying that every mode has a score consequence.

The mode resolves contribution measurement, threshold interpretation, source eligibility, tie handling, and whether a participant can receive both final destruction credit and assist credit for the same outcome. The initial model uses contribution within the assist window and permits multiple eligible assistants.

Assist distribution is part of the same authoritative award event as the credited destruction outcome. It is not a client-side follow-up and must not be separately inferred from a death or destruction notification.

## Combos And Multipliers

Combos and multipliers are one generic system, separate from streaks. A combo modifies qualifying award values according to a combo state owned by a configured owner.

The initial combo model is:

```text
discrete tiers
starting multiplier: 1.0x
initial increase: +0.25x per qualifying hit
initial qualifying window: 0.75 seconds
no maximum multiplier
reset on timeout
reset on death
one combo state per owner
player-owned by default
```

The target counter is configurable, although SCORE is the primary initial target. A mode may select another numeric counter where the mutation remains meaningful and explicitly resolved.

A player-owned combo is the default. A team-owned combo may be supported only if the runtime seam can provide clean authoritative ownership, ordering, timeout, reset, distribution, and visibility semantics. Team combo must not be made default merely to reuse team aggregation.

Combo modification occurs after base awards and assists have been resolved. Combo state updates and modified values must be represented in the authoritative award event so recipients cannot apply a multiplier twice.

A combo timeout is authoritative. Death resets the relevant combo state by default, and modes may explicitly modify qualifying or reset rules if the generic contract supports that override. Presentation may show the current tier but cannot advance, preserve, or reset the combo locally.

## Streaks

Streaks are a separate generic named system. The initial focus is PvP kill streaks.

A streak tracks a named sequence of qualifying events for its owner. Multiple streaks may be active at once. A streak resets on death by default, and the selected mode may modify qualifying and reset rules.

The first product slice defines no initial bonus, drop, or announcement triggers. The system must nevertheless preserve future trigger hooks for those categories without assigning them current behavior. Streak progression and streak counter mutation are distinct from combo multiplier modification.

A streak may be player-owned by default and may later be team-owned or attached to another supported runtime owner when the mode explicitly selects that scope. Exact named streak catalogue, qualification rules, output fields, and future triggers remain implementation-level decisions unless a mode settles them.

## Award Pipeline

All normal gameplay awards follow one authoritative ordered pipeline:

```text
1. attribution
2. base awards
3. assists
4. combo modification
5. streak updates
6. penalties
7. team distribution
8. counter mutation
9. objective evaluation
10. match-end evaluation
```

Attribution identifies eligible sources and recipients before any counter mutation. Base awards then calculate the direct event value. Assists add eligible secondary recipients. Combo modification changes configured award values using the resolved combo state. Streak updates advance named streak state and expose future trigger hooks without silently becoming combo logic. Penalties apply after the positive award decisions and before distribution. Team distribution resolves all team recipients or team totals. Counter mutation applies the resulting operations exactly once. Objective and match-end evaluation consume the completed award result.

Award distribution completes before final match lock and `EndOfMatchFlow`. A final destruction, milestone, or other legitimate event that falls within the authoritative resolution boundary must be fully attributed, distributed, and mutated before the match becomes locked. Match-end evaluation may decide that the event ends the match, but it must not truncate the award pipeline.

Objective evaluation and match-end evaluation are handoffs. They may consume the final counter state and award facts, but they do not reorder or duplicate the award pipeline.

## Team Award Distribution

Team award behavior is mode-configurable. The default is full individual awards, with team totals derived by summing player counters.

The default means:

```text
eligible players receive their resolved individual awards
team counters aggregate the configured player values by sum
team presentation may display the derived team total
```

Alternative distribution models are supported by the seam but are not default. A mode may distribute an award to a team only, split a team award, apply a team modifier, or use another explicit model. Any alternative must define eligible recipients, counter scope, duplicate prevention, visibility, and result snapshot behavior.

Team distribution must not duplicate individual mutation when a team total is derived by sum. If a mode selects a dedicated team-owned award, it must be represented as a distinct authoritative distribution target rather than inferred from a displayed aggregate.

The team system supplies authoritative membership and team relationship facts. This system consumes those facts to resolve award recipients; it does not assign teams or redefine team membership.

## Counter Mutation And Lifecycle

Counter mutation supports:

```text
increment
decrement
set
min/max
 timed accumulation
```

Timed accumulation is primarily for devtools/admin operations and must remain an explicit authoritative mutation path. Normal gameplay awards should use the operation resolved by the award definition rather than allowing arbitrary client-selected mutation types.

Normal gameplay defaults to monotonic counters. Modes may configure non-monotonic behavior where the counter semantics require it, including explicit decrement, set, or bounded mutation. A mode must declare the operation and any min/max behavior; consumers must not infer counter semantics from the counter name.

`COMPLETION_TIME` is a numeric gameplay counter when selected by a mode, but the mode/objective owner remains responsible for deciding what completion means. `OBJECTIVE_PROGRESS` records configured gameplay progress without making this system the owner of objective rules.

Counters reset with normal match/round ownership. A destroyed runtime ship, respawn, disconnect, reconnect, or presentation transition must not reset durable match-scoped player counters unless the resolved lifecycle or mode policy explicitly defines a match-state transition. Removed players retain earned counters and do not receive normal new awards.

## Visibility And Snapshots

Visibility is mode-defined. Supported visibility classes include:

```text
hidden
HUD
scoreboard
results-only
team-only
player-private
spectator-visible
```

Visibility controls where an authoritative fact may be exposed; it does not change ownership or mutation. A player-private counter remains server-authoritative. A team-only award may be shown to teammates without making team members independent authorities. Spectators receive only facts permitted by the selected mode.

The award system exposes a clean final snapshot for result and presentation handoff. The snapshot contains resolved counters, attribution facts, assist facts, combo and streak state or final facts required by mode, team distributions, and visibility metadata as selected by policy.

Result and persistence handoff remains deferred to result planning. This system does not emit persistence-specific output or decide account progression. `MatchDecision`, `EndOfMatchFlow`, and `MatchSummary` consume the clean final snapshot through their own contracts.

## Event Idempotency

Idempotency belongs at the authoritative award-event, distribution, and broadcast level, not independently per client.

Each legitimate recipient receives a distribution once for a given legitimate award event. Multiple legitimate recipients from one event are valid: for example, a destruction event may have one final destruction recipient and multiple assist recipients, or an explicit team distribution may target a team scope as well as eligible individuals according to mode policy.

An event identity must cover the authoritative event and resolved distribution instance. Replays, duplicate transport delivery, stale client acknowledgements, and duplicate handler execution must not mutate counters twice or broadcast a second effective award. Client deduplication may improve presentation resilience, but it is not the authoritative guarantee.

Idempotency must be applied after recipient resolution and before counter mutation/broadcast completion. A failed or partial distribution must have an explicit authoritative retry or rejection outcome; it must not rely on clients to decide whether a recipient already received credit.

## Runtime Ownership

Runtime ownership follows existing server ownership seams. Exact package names remain implementation-level decisions.

```text
player counters, combo, and streak
-> player match state

team counters and optional team combo/streak
-> team runtime state

objective counters
-> objective runtime state

match-wide counters
-> match runtime state
```

The runtime ship owns live avatar state, not durable player counters, combo, streak, or match history. A destroyed or recreated ship cannot lose player score, kills, assists, deaths, or other retained match facts.

The award owner coordinates policy and mutation through these existing owners. It must not create a parallel counter store in networking, presentation, transport, or a generic shared utility. Team membership comes from Teams And Team Rules. The Objective Foundation owns objective-local state, while separate progress and aggregation owners produce the facts objectives consume. Match runtime owns match-wide values and final resolution context.

## System Handoffs

```text
resolved mode and award rules
-> award catalogue, values, counter scopes, visibility, and mutation policy
-> authoritative combat/objective/resource/time event
-> attribution and contribution resolution
-> base awards, assists, combo, streaks, and penalties
-> team distribution
-> idempotent counter mutation and award broadcast
-> completed counter state
-> objective evaluation and match-end evaluation
-> final match lock
-> clean final snapshot for results and presentation
```

### Combat, Destruction, And Runtime Ship

Combat and collision systems own simulation outcomes, projectile ownership, hit/impact/collision facts, damage, and destruction detection. The award system consumes those outcomes, resolves final valid-hit credit and assists, and applies configured award consequences. Runtime ship remains the live avatar owner and does not become the counter store.

### Modes And Match Rules

Modes select enabled counters, award values, scaling overrides, attribution eligibility, assist policy, combo/streak policy, penalties, team distribution, visibility, mutation operations, lifecycle resets, and match-end participation. They provide explicit resolved rules rather than making downstream consumers infer policy from mode identity.

### Objectives And Objective Runtime

[Objectives And Objective Runtime](objectives-and-objective-runtime.md) owns the Objective Foundation: objective definitions/schema, definition condition evaluation, and objective-local lifecycle/state. Separate progress and aggregation owners produce progress facts for it to consume. It may emit award sources and consume completed counter mutations. `OBJECTIVE_PROGRESS` remains a handoff counter, not objective authority or permission for the award system to decide objective completion.

### Teams And Team Rules

Teams owns structure, membership, relationship facts, and team lifecycle facts. The award system uses those facts for recipient resolution and default team aggregation. It does not assign players, decide team elimination, or replace team-owned runtime state.

### Lives, Death, Elimination, And Respawn

Lives/death/respawn owns death and lifecycle transitions and supplies authoritative death attribution categories. The award system applies configured DEATHS, killer, assist, penalty, combo reset, and streak reset consequences through the award pipeline without owning respawn or elimination.

### Match Outcomes And Results

Match rules and outcome orchestration own match end, final lock, result meaning, and result emission. Award distribution completes before final lock. Results consume the clean final snapshot and do not reconstruct counters, attribution, or distribution from runtime entities.

### Ranking And Progression

Ranking consumes the result or final snapshot through its own policy and decides ranking calculations. Progression, achievements, and account rewards consume eligible result/progression handoffs. Gameplay awards do not persist cross-match score, calculate leaderboards, grant progression, or decide reward eligibility.

### Player Experience And Presentation

HUD, scoreboard, results, and spectator presentation consume mode-authorized visible facts. Presentation may animate awards, combo tiers, streaks, and counter changes, but cannot mutate counters, apply multipliers, select recipients, or deduplicate authoritative credit.

## Implementation Direction

The first implementation slice should proceed from resolved rules through one authoritative award pipeline:

```text
1. Define normalized catalogue, custom-counter extension, scope, visibility, and mutation shapes.
2. Define authoritative award-event identity and distribution idempotency contract.
3. Consume combat, objective, resource, time, and mode-event sources.
4. Resolve projectile ownership, last-valid-hit destruction credit, and contribution history.
5. Apply mode-enabled assists using the initial 5% / 5-second policy.
6. Add generic combo state and separate named streak state through player match ownership.
7. Apply penalties and configurable team distribution, defaulting to individual awards plus summed team totals.
8. Mutate player, team, objective, and match runtime counters exactly once.
9. Expose mode-authorized visibility and a clean final snapshot.
10. Hand completed facts to objective and match-end evaluation before final lock and to results/progression/ranking through their owners.
11. Preserve extension seams for custom counters, team combo/streak, alternate distributions, and future triggers without making them default.
```

Implementation should keep award policy in a focused gameplay owner. Networking, presentation, game-loop coordination, objective code, team code, result code, and progression code should route authoritative facts or execute their own policy rather than becoming alternate award authorities.

## Testing Direction

Important future checks:

```text
fixed catalogue supports SCORE, KILLS, ASSISTS, DEATHS, DAMAGE_DEALT, DAMAGE_TAKEN, OBJECTIVE_PROGRESS, RESOURCES_COLLECTED, and COMPLETION_TIME
custom numeric counters use the same ownership and idempotency seam
player, team, match-wide, and objective-specific ownership are distinct
award sources include destruction, impacts/hits/collisions, objectives, milestones, pickups/resources, and mode events
projectile owner receives eligible hit/impact/collision attribution
last valid hit receives final destruction credit
assist eligibility is mode-enabled and initial 5% contribution within 5 seconds
multiple assistants are allowed and assist award is mode-selected
contribution history expires after the assist window plus approximately 10% buffer
asteroid awards continue scaling by target size
mode award values and calculations override defaults explicitly
pipeline order is attribution, base awards, assists, combo, streaks, penalties, team distribution, mutation, objective, match end
award distribution completes before final match lock and EndOfMatchFlow
combo is separate from streaks and starts at 1.0x with +0.25x qualifying-hit tiers
combo uses the initial 0.75-second window, has no maximum, and resets on timeout and death
one combo exists per owner; player ownership is default
team combo is not enabled unless its ownership seam is cleanly supported
streaks are named, generic, multiple, and reset on death by default
initial streak behavior has no bonus, drop, or announcement trigger while future hooks remain
team default is full individual awards with team totals summed from player counters
alternative team distributions require explicit mode policy and do not duplicate mutation
counter mutation supports increment, decrement, set, min/max, and timed accumulation for devtools/admin
normal gameplay counters are monotonic by default and mode overrides are explicit
counter lifecycle resets with normal match/round ownership and has no generic cross-round persistence
removed players retain earned counters but receive no normal new awards
delayed events use authoritative attribution and lifecycle facts
idempotency is authoritative per award event/distribution/broadcast, not per client
multiple legitimate recipients from one event each receive one valid distribution
visibility classes are mode-defined and do not change ownership
player, team, objective, and match runtime owners retain their respective counter state
final snapshot is clean and not persistence-specific
objective, ranking, match-end, results, progression, and presentation owners retain their boundaries
```

## Related Docs

- [Gameplay Planning](./!INDEX.md)
- [Modes And Match Rules](modes-and-match-rules.md)
- [Objectives And Objective Runtime](objectives-and-objective-runtime.md)
- [Teams And Team Rules](teams-and-team-rules.md)
- [Lives, Death, Elimination, And Respawn](lives-death-elimination-and-respawn.md)
- [Match Outcomes And Results](match-outcomes-and-results.md)
- [Achievements And Milestones](achievements-and-milestones.md)
- [Progression And Rewards](progression-and-rewards.md)
- [Player Experience Systems](player-experience-systems.md)
- [Leaderboards And Rankings](../platform/leaderboards-and-rankings.md)

## Remaining Implementation-Level Decisions

- Exact catalogue identifiers, custom-counter registration, numeric types, and validation rules.
- Exact award-event, attribution, contribution, distribution, mutation, visibility, and final-snapshot field names.
- Exact package boundaries and runtime state types within existing server ownership seams.
- Exact source-event adapters for combat, destruction, objectives, resources, time milestones, and mode events.
- Exact last-valid-hit ordering, source expiry, tie handling, and delayed-event acceptance rules.
- Exact assist contribution measurement, threshold comparison, window timing, and 10% buffer implementation.
- Exact award-value tables, asteroid-size scaling curve, and mode override representation.
- Exact combo qualification, tier representation, timeout scheduling, and configurable target-counter contract.
- Exact conditions under which team-owned combo or alternative team distribution is enabled.
- Exact named streak catalogue, qualification/reset contracts, and future trigger hook shape.
- Exact penalty operation and ordering details for each mode.
- Exact counter mutation bounds and timed-accumulation authorization for devtools/admin.
- Exact authoritative idempotency key composition, retry behavior, and broadcast acknowledgment handling.
- Exact visibility projection and spectator/team/private authorization checks.
- Exact objective and match-end evaluation handoff contracts.
- Exact final snapshot fields and result/progression/ranking projection contracts.
- Exact packet and storage shapes chosen by their owning systems.

There are no remaining product-level awards or counters decisions blocking P4 system planning.

## Core Invariants

```text
Gameplay Awards And Counters is the authoritative planning owner for awards, counters, attribution, assists, combos, streaks, distribution, mutation, visibility, and award-event idempotency.

The fixed catalogue is SCORE, KILLS, ASSISTS, DEATHS, DAMAGE_DEALT, DAMAGE_TAKEN, OBJECTIVE_PROGRESS, RESOURCES_COLLECTED, and COMPLETION_TIME, with custom numeric counters behind an extension seam.

Counters may be player-owned, team-owned, match-wide, or objective-specific.

Projectile ownership supplies hit/impact/collision attribution, and the last valid hit receives final destruction credit.

Assists are mode-enabled; the initial threshold is 5% contribution in a 5-second window, multiple assistants are allowed, and the assist award is mode-selected.

Contribution history lasts only through the assist window plus approximately a 10% timing buffer.

Award values are tunable, asteroid awards scale by target size, and modes may explicitly override calculations.

The award pipeline order is attribution, base awards, assists, combo modification, streak updates, penalties, team distribution, counter mutation, objective evaluation, and match-end evaluation.

Award distribution completes before final match lock and EndOfMatchFlow.

Combos are separate from streaks, use discrete tiers, start at 1.0x, initially rise by +0.25x per qualifying hit, use a 0.75-second window, have no maximum, and reset on timeout and death.

One combo state exists per owner; player-owned combo is default, and team-owned combo is optional only when cleanly supportable.

Streaks are generic named systems, initially focused on PvP kill streaks, reset on death by default, and support multiple active streaks.

The default team behavior gives full individual awards and derives team totals by summing player counters; alternatives are explicit mode policy.

Counter mutation supports increment, decrement, set, min/max, and timed accumulation for devtools/admin. Normal gameplay is monotonic by default.

Counters reset with normal match/round ownership; generic cross-round score persistence is not designed.

Removed players retain earned counters but stop receiving normal new awards.

Idempotency is authoritative at award-event/distribution/broadcast level. Each legitimate recipient is credited once, while multiple legitimate recipients remain valid.

Runtime ships do not own durable player counters, combos, streaks, or match history.

Player, team, objective, and match runtime states own their corresponding facts through existing server seams.

Visibility is mode-defined and may be hidden, HUD, scoreboard, results-only, team-only, player-private, or spectator-visible.

The final snapshot is clean and not persistence-specific; result, ranking, progression, objective, match-end, and presentation systems retain their ownership boundaries.
```
