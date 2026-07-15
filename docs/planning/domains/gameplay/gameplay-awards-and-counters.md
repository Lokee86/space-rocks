# Gameplay Awards And Counters
Parent index: [Gameplay Planning](./!INDEX.md)

## Purpose

This doc is the authoritative P4 planning owner for gameplay awards, counters, attribution, assists, combos, multipliers, and streaks.

It defines how authoritative gameplay events become counter mutations, how credit is distributed between players and teams, how award state is ordered before objective and match-end evaluation, and what normalized facts are handed to objectives, rankings, results, statistics, devtools, and player experience. It defines policy boundaries without claiming that the implementation already exists.

## Ownership Boundary

This doc owns planning for:

```text
fixed gameplay-counter catalogue
custom numeric-counter extension seam
counter ownership scopes
authoritative award events
award attribution and distribution
hit, impact, collision, and destruction credit
assist eligibility and contribution history
award-value calculation
combo and multiplier state
streak state
team award-sharing modes
counter mutation operations
counter lifecycle and reset behavior
authoritative award idempotency
counter visibility metadata
normalized counter snapshots and handoffs
```

This doc does not own:

```text
mode objectives or objective completion
participant or team ranking
match-end conditions or MatchDecision locking
combat damage formulas and collision detection
team membership or team assignment
room membership, connection, or reconnect execution
result orchestration or persistence schemas
progression, rewards, achievements, or permanent statistics
client HUD and scoreboard layout
packet schemas or package placement
```

Modes and match rules select award definitions, counter behavior, assist policy, combo policy, streak policy, team distribution, penalties, visibility, and custom-counter definitions as part of resolved rules. The gameplay-awards owner applies those rules to authoritative gameplay events and exposes normalized counter facts. Objectives, rankings, match-end evaluation, results, statistics, and presentation consume those facts rather than reconstructing credit from runtime entities or visible effects.

## Settled Product Model

Gameplay awards, objective progress, final ranking, and result presentation are separate concepts.

A counter may be used by one or more of those systems, but the existence of a `SCORE` counter does not imply that the match ranks participants by score. For example, Score Attack may use `SCORE` as objective progress while ranking successful participants by completion time.

The initial foundation provides:

```text
fixed stable counters
custom numeric counters
player-owned counters
team-owned counters
match-wide counters
objective-owned counters
tunable award definitions
mode-selected distribution and visibility
```

Normal gameplay counters are match- or round-scoped. They reset at the owning lifecycle boundary. General score persistence between rounds is not supported by this system. Any future series, campaign, tournament, or persistent-stat aggregation must consume completed snapshots through its own owner rather than retaining a live round counter across incompatible lifecycle boundaries.

## Fixed Counter Catalogue

The initial fixed catalogue is:

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

These are stable semantic identifiers, not UI labels. Modes may hide unused counters and may select different counters for objectives, ranking, results, or presentation.

The fixed catalogue should transfer cleanly into future permanent statistics without making permanent statistics the runtime counter authority.

### Counter Semantics

`SCORE` is the general gameplay-points counter. It may represent asteroid points, combat points, assist awards, objective awards, or other configured gameplay value.

`KILLS` records eligible final-destruction or elimination credit under the selected mode and combat policy.

`ASSISTS` records eligible assist credit separately from kills and from any `SCORE` award granted for the assist.

`DEATHS` records authoritative death outcomes.

`DAMAGE_DEALT` and `DAMAGE_TAKEN` record authoritative applied damage rather than requested or predicted damage.

`OBJECTIVE_PROGRESS` is available as a general objective-facing counter, but objectives may own more specific counters where one shared numeric value is insufficient.

`RESOURCES_COLLECTED` records eligible match-scoped collection facts. Durable inventory grants remain owned elsewhere.

`COMPLETION_TIME` records an authoritative duration or timestamp-derived value when a mode requires it. It is not inferred from a client's local clock.

## Custom Counter Seam

Modes may define custom numeric counters through an extension seam.

A custom counter definition must identify at least:

```text
stable counter identifier
ownership scope
initial value
allowed mutation behavior
lifecycle/reset boundary
visibility metadata
```

Custom counters are numeric in the first planning model. Boolean state, structured objective state, inventories, lists, and arbitrary documents should remain typed state owned by their appropriate systems rather than being forced into numeric counters.

The custom seam exists because useful seams are easier to remove than to retrofit after objectives, rankings, results, devtools, and statistics depend on a closed catalogue. It must remain concrete and bounded rather than becoming an untyped general-purpose state store.

## Counter Ownership

Counters may be owned by:

```text
player match state
team runtime state
match runtime state
objective runtime state
```

Player counters follow the authoritative participant identity, not the current runtime ship. Destroying and recreating a ship must not erase the player's match counters.

Team counters follow authoritative team membership and team runtime state. The team system supplies membership and relationship facts; it does not independently invent awards.

Match-wide counters belong to the authoritative match runtime and are not duplicated once per player.

Objective-owned counters belong to objective runtime state and are exposed to gameplay awards only through an explicit objective/counter handoff.

One award event may distribute mutations to more than one ownership scope when the resolved policy requires it.

## Award Sources

Authoritative awards may be produced from:

```text
hits
impacts
collisions
damage application
target destruction
player kills
assists
objective interactions
objective completion or failure
survival or time milestones
pickups
resource collection
mode-specific gameplay events
administrative or devtool actions
```

The gameplay system that owns the source event emits normalized authoritative facts. The award system applies configured attribution, calculation, distribution, and mutation. It must not infer gameplay truth from client reports or presentation events.

Award definitions are tunable. Modes may select baseline definitions and apply explicit overrides.

Asteroid score values continue to scale by target size through the existing scoring direction. Exact values remain tuning data rather than hard-coded universal semantics.

## Attribution

Projectile ownership is the primary attribution source for projectile-caused gameplay events.

Configured awards may be granted for:

```text
projectile hits
impacts
collisions
damage contribution
final destruction
```

Final destruction credit goes to the last valid hit under the selected attribution policy.

Last-hit credit does not erase contribution history. Assist evaluation remains separate and may grant credit to additional eligible contributors.

Attribution supports:

```text
direct projectile owner
other owned damaging effect
assist contributors
self-caused event
environmental source
unattributed event
```

The authoritative source event must preserve enough ownership identity to resolve credit even when the projectile, effect, target, or runtime ship is removed during the same simulation update.

## Assist Policy

Assists are mode-enabled rather than universal.

The initial assist policy defaults are:

```text
minimum contribution: 5 percent
contribution window: 5 seconds
multiple assistants: allowed
assist reward: mode-selected
```

The contribution threshold and contribution window are tunable.

Contribution history only needs to be retained for the configured assist window plus a small timing buffer, initially approximately 10 percent of the window. The buffer protects boundary processing without turning contribution history into durable combat history.

Assist contribution should be based on authoritative applied contribution, normally damage, rather than attempted damage or client prediction.

The mode selects the assist reward. The likely standard reward includes an `ASSISTS` increment and a `SCORE` award. Granting `SCORE` does not affect match victory unless that mode's objective or ranking policy actually consumes `SCORE`.

A final credited killer is not also counted as their own assistant for the same destruction event.

Environmental, self-caused, and unattributed destructions may produce no assists unless the selected mode explicitly defines otherwise.

## Contribution History

Contribution history is bounded runtime attribution state.

It should contain only the information required to evaluate the configured assist policy, such as:

```text
eligible contributor identity
contribution amount
last contribution time
source category when required
```

Expired entries are removed after the assist window and retention buffer.

Contribution history is not a permanent statistic, replay ledger, damage log, or result history. Completed awards and final counter values provide downstream facts; observability may record separate diagnostics through its own contracts.

Disconnected or removed participants retain previously earned counters. They stop receiving normal new awards after removal. A source event that became authoritative before removal may complete its already-resolved distribution when the selected attribution policy permits it, but removal must not leave the participant eligible for unrelated future events.

## Award Calculation

Award values are tunable and may use:

```text
fixed values
target type
target size
target difficulty
source event type
mode context
participant or team context
combo multiplier
streak-triggered award definitions
explicit mode overrides
```

Award calculation must produce an explicit distribution before counters are mutated.

A mode may disable an otherwise available award source or replace the baseline calculation. Consumers must not infer award formulas from `mode_id`; the resolved award policy carries the selected definitions.

Negative adjustments and penalties are mode-selected. Standard gameplay defaults to non-negative, generally monotonic accumulation. Modes may explicitly permit deductions, negative values, resets, or other mutation behavior.

## Combo And Multiplier System

Combos and multipliers use one generic mechanic. Streaks remain a separate mechanic.

The initial combo defaults are:

```text
starting multiplier: 1.0x
tier model: discrete
increase per qualifying event: 0.25x
combo window: 0.75 seconds
maximum multiplier: none
reset on timeout: yes
reset on death: yes
```

All values and qualifying events are tunable through resolved policy.

A qualifying event refreshes the combo window and advances one discrete tier. When the window expires, the multiplier returns to `1.0x`.

There is no universal maximum multiplier. Modes may introduce a cap through their selected policy if required, but the baseline system does not impose one.

The multiplier primarily modifies `SCORE`, but the affected counter or award category is configurable. The combo system applies to award calculation; it does not directly decide objectives, ranking, or match end.

There is one combo state per owning player or team. Player-owned combo state is the default and first required behavior. Team-owned combo state may be supported through the same ownership contract when it fits the team runtime cleanly; the first implementation does not need to introduce complicated shared-combo synchronization merely to prove the seam.

The combo policy may define additional reset conditions later. Timeout and death are the required initial resets.

## Streak System

Streaks are generic named consecutive-event trackers separate from combo multipliers.

The initial practical use is PvP kill streaks.

The default streak behavior is:

```text
qualifying event: eligible PvP kill
reset: participant death
multiple active streak definitions: supported
automatic reward triggers: none initially
```

Modes may modify qualifying events and reset conditions.

The architecture should preserve hooks for future streak-triggered behavior such as announcements, award events, drops, achievements, or mode mechanics. Those triggers are not part of the initial implementation requirement.

A streak does not inherently multiply awards. A mode may connect a streak threshold to an award or combo policy explicitly, but streak state and combo state remain independently owned and configurable.

## Team Award Distribution

Team award behavior is mode-configurable.

The default model is:

```text
players receive their individual awards
team values derive from the sum of player counters
```

Modes may instead select:

```text
direct team-owned award
award divided between eligible teammates
full award granted to each eligible teammate
custom explicit distribution
```

Award sharing and team aggregation are separate concerns. Sharing decides who receives mutations from one event. Aggregation decides how player counters contribute to a team value.

Team aggregation defaults to sum as defined by [Teams And Team Rules](teams-and-team-rules.md). A mode may select an explicit alternate aggregation where its objective or ranking semantics require one.

Friendly-fire policy and team relationships are not owned here. The award system consumes normalized same-team and opposing-team facts where an award definition requires them.

## Counter Mutation Operations

The runtime counter system must support:

```text
increment
decrement
set
minimum clamp
maximum clamp
timed accumulation
```

These operations are required for devtools, administration, automated testing, and future mode mechanics.

Normal gameplay defaults to monotonic accumulation where appropriate. Modes must explicitly permit deductions, non-monotonic changes, or values below zero.

Mutation permissions belong to authoritative server policy. Supporting `set` or `decrement` does not imply that clients may request arbitrary counter changes.

Counter mutations should preserve enough source context for authoritative diagnostics and duplicate prevention without turning each counter into an unbounded event ledger.

## Award Processing Order

The authoritative processing order is:

```text
authoritative gameplay event
-> attribution
-> base award calculation
-> assist evaluation and awards
-> combo multiplier application
-> streak-state update and any configured triggered awards
-> penalties or negative adjustments
-> team distribution
-> final award distribution
-> counter mutations
-> objective evaluation
-> ranking-fact update where required
-> match-end evaluation
-> authoritative MatchDecision lock when satisfied
-> EndOfMatchFlow
```

All award distribution and counter mutation caused by the decisive gameplay event completes before the final match lock and match-end flow.

This ordering ensures that the event which ends the match contributes its valid score, kills, assists, damage, objective progress, combo state, and other counters before final facts are frozen.

Triggered awards created during this pipeline must remain bounded and must not create uncontrolled recursive award processing.

## Authoritative Idempotency

Idempotency applies at the authoritative award-event or distribution level, not independently per connected client.

The same authoritative distribution must not mutate the same recipient and counter more than once because of duplicate callbacks, retries, repeated broadcasts, or repeated processing.

A single gameplay event may legitimately distribute awards to multiple players, teams, objectives, or match scopes. That is not a duplicate.

The award system should produce a stable event or distribution identity sufficient for bounded duplicate suppression. Exact identifier shape and retention strategy are implementation-level decisions.

Client broadcast consumes already-applied authoritative mutations or snapshots. Clients do not own award idempotency and must not decide whether an authoritative award should exist.

## Counter Lifecycle

Counters initialize at the beginning of their resolved owning lifecycle.

Typical boundaries are:

```text
match start
round start
objective activation
participant activation
team activation
```

Counters reset when that owning lifecycle ends or restarts unless an owning system explicitly creates a new aggregate from the completed snapshot.

This system does not preserve live `SCORE` or other counters between rounds as a generic feature. Multi-round totals, series standings, campaign accumulation, and persistent statistics must be separate aggregates consuming completed round or match facts.

Death does not reset ordinary player counters unless the selected mode explicitly defines a penalty or reset. Combo and streak reset behavior remains governed by their own policies.

## Visibility Metadata

Counter visibility is mode-defined.

Definitions may identify counters as:

```text
hidden
HUD-visible
scoreboard-visible
results-only
team-visible
player-private
spectator-visible
```

Visibility metadata describes authorized presentation scope. It does not give the client authority over the value and does not define exact UI layout.

A counter may be authoritative and hidden while still being consumed by objectives, ranking, results, devtools, or statistics.

## Results, Statistics, And Persistence Handoff

Exact result and persistence fields are intentionally deferred to their owning plans.

This system must expose a normalized final counter snapshot and any selected attribution facts needed by downstream owners. It must not write persistence-specific payloads or decide which counters become permanent statistics.

Potential consumers include:

```text
MatchDecision inputs
MatchSummary
team and player result projections
permanent player statistics
achievements and milestones
progression and rewards
devtools and diagnostics
```

Each consumer selects the facts it owns. The award system remains the runtime authority for how gameplay events changed match-scoped counters.

## System Handoffs

```text
authoritative gameplay source event
-> normalized source and ownership facts
-> award attribution
-> assist contribution evaluation
-> tunable award calculation
-> combo and streak policy
-> team distribution
-> authoritative idempotent counter mutation
-> normalized player/team/match/objective counter facts
-> objectives and ranking
-> match-end evaluation and locked MatchDecision
-> results, statistics, progression, and presentation consumers
```

### Modes And Match Rules

Modes select the award catalogue entries, custom counters, source eligibility, values, assist policy, combo policy, streak policy, penalties, team distribution, aggregation overrides, mutation permissions, and visibility. Modes consume counter facts for objectives, rankings, and match end without redefining event attribution or runtime counter ownership.

### Combat And Runtime Gameplay

Combat, collision, pickups, objectives, and other gameplay owners emit authoritative source facts. They do not directly mutate arbitrary scoreboards or reconstruct team totals. Projectile and effect ownership must remain available through source normalization.

### Teams And Team Rules

The team system supplies membership and relationship facts plus the default sum aggregation rule. Gameplay awards resolve event distribution and team-owned mutations without becoming a second team-membership authority.

### Lives, Death, Elimination, And Respawn

The lives/death owner supplies authoritative death facts and attribution categories. Gameplay awards applies configured kill, assist, death, penalty, streak-reset, and combo-reset behavior. Neither system should independently reconstruct the other's facts.

### Objectives And Ranking

Objectives consume normalized counter facts as progress inputs where selected. Ranking consumes selected final or live counter facts. Neither system assumes `SCORE` is universally decisive.

### Match Outcomes And Results

Match-end evaluation runs only after the decisive event's award distribution has completed. Result orchestration consumes locked final counter snapshots and selected attribution facts without recalculating awards.

### Player Experience

Client presentation consumes authoritative counter values, combo state, streak state, and visibility metadata. It may animate awards but cannot grant them, alter their order, or decide attribution locally.

### Statistics, Progression, And Achievements

Permanent statistics, progression, rewards, and achievements consume authoritative completed facts through their own seams. They do not own live gameplay counters and must not cause a gameplay award to be applied twice.

## Implementation Direction

The first implementation slice should proceed from current asteroid score behavior into explicit award and counter ownership:

```text
1. Define stable fixed counter identifiers and scoped runtime counter state.
2. Preserve current asteroid size-based SCORE behavior behind an award definition.
3. Normalize projectile-owner, hit, impact, collision, and destruction source facts.
4. Apply last-valid-hit final destruction credit.
5. Add bounded contribution history and mode-enabled assist evaluation.
6. Produce explicit award distributions before counter mutation.
7. Add authoritative distribution-level idempotency.
8. Add tunable combo state with discrete tiers and timeout/death reset.
9. Add generic named streak state with PvP kill-streak support and death reset.
10. Add player/team/match/objective ownership scopes and default team-sum aggregation handoff.
11. Add custom numeric counter definitions and visibility metadata.
12. Route objective and match-end evaluation after final mutations.
13. Expose normalized live and final snapshots to downstream owners.
```

Implementation should preserve existing scoring behavior first. New counter, assist, combo, streak, and team-distribution behavior should land behind explicit seams rather than requiring all future modes to exist immediately.

Exact package placement must respect existing game-server ownership and avoid moving durable participant counters onto runtime entities.

## Testing Direction

Important future checks:

```text
fixed counters use stable identifiers
custom numeric counters cannot replace typed non-counter state
player counters survive runtime ship destruction and respawn
current asteroid SCORE remains size-based
projectile-owner attribution survives source entity removal
hit, impact, and collision awards use configured definitions
last valid hit receives final destruction credit
assist policy is disabled unless selected by the mode
assist threshold defaults to 5 percent
assist window defaults to 5 seconds
multiple assistants may qualify
contribution history expires after window plus buffer
killer is not credited as their own assistant
combo starts at 1.0x
combo advances in discrete 0.25x tiers by default
combo window defaults to 0.75 seconds
combo has no universal maximum
combo resets on timeout and death
combo affects SCORE by default but can target another award category
multiple named streaks can coexist
PvP kill streak resets on death
streaks have no required trigger rewards in the first slice
team awards default to individual mutations with summed team totals
mode-selected alternate team distribution applies exactly once
increment, decrement, set, clamp, and timed accumulation remain server-authoritative
standard gameplay counters remain monotonic unless mode policy allows otherwise
duplicate processing cannot apply one distribution twice
one event may legitimately award multiple recipients
removed players retain prior counters and stop receiving unrelated new awards
all decisive-event mutations complete before MatchDecision lock
clients display awards but never author them
round or match reset does not preserve live SCORE into an unrelated lifecycle
```

## Related Docs

- [Gameplay Planning](./!INDEX.md)
- [Modes And Match Rules](modes-and-match-rules.md)
- [Teams And Team Rules](teams-and-team-rules.md)
- [Lives, Death, Elimination, And Respawn](lives-death-elimination-and-respawn.md)
- [Match Outcomes And Results](match-outcomes-and-results.md)
- [Player Experience Systems](player-experience-systems.md)
- [Progression And Rewards](progression-and-rewards.md)
- [Achievements And Milestones](achievements-and-milestones.md)

## Remaining Implementation-Level Decisions

- Exact type and field names for fixed and custom counter definitions.
- Exact numeric types, precision, overflow handling, and serialization for each counter category.
- Exact stable award-event or distribution identifier and bounded idempotency retention strategy.
- Exact normalized source-event shapes for hits, impacts, collisions, damage, destruction, pickups, and objective events.
- Exact last-valid-hit tie and same-tick ordering behavior where simulation ordering is ambiguous.
- Exact contribution-history storage layout, cleanup cadence, and 10-percent buffer implementation.
- Exact default assist `SCORE` value and tuning data.
- Exact asteroid size-to-score tuning values as content changes.
- Exact combo tier representation, timer precision, and configured counter-target vocabulary.
- Whether team-owned combo state lands in the first implementation or remains a preserved ownership option.
- Exact named-streak definition shape and future trigger-hook interface.
- Exact custom-counter validation, namespacing, and collision rules.
- Exact visibility metadata representation and client projection.
- Exact devtool and administrative authorization for non-standard mutations.
- Exact runtime package boundaries and snapshot/packet shapes.
- Exact result, statistics, persistence, progression, and achievement projections, which remain owned by their later plans.

There are no remaining product-level gameplay-award or counter decisions blocking P4 system planning.

## Core Invariants

```text
Gameplay Awards And Counters is the authoritative runtime award and counter-semantics planning owner.

Gameplay awards, objective progress, ranking, and results remain separate concepts.

The fixed catalogue is stable and custom numeric counters use a bounded extension seam.

Counters may be owned by players, teams, the match, or objectives.

Projectile ownership drives projectile attribution.

Final destruction credit goes to the last valid hit.

Assists are mode-enabled, allow multiple contributors, and default to a 5-percent threshold over a 5-second window.

Contribution history is bounded to the assist window plus a small buffer.

Award values are tunable and current asteroid SCORE remains size-based.

Combos use discrete tiers, default to +0.25x per qualifying event, use a 0.75-second window, have no universal maximum, and reset on timeout or death.

Streaks remain separate from combos; the first use is PvP kill streaks with death reset.

Team awards default to individual awards with team totals derived by sum.

Normal gameplay defaults to monotonic counters, while authoritative devtools, administration, and explicit mode policies may use broader mutation operations.

Idempotency is enforced at the authoritative award-event or distribution level, not per client broadcast.

All valid award distribution and counter mutation completes before objective, ranking, match-end evaluation, and MatchDecision lock.

Disconnected or removed participants retain earned counters but stop receiving unrelated new awards.

Live counters reset with their match, round, participant, team, or objective lifecycle; general cross-round SCORE persistence is not supported.

Clients consume authoritative values and visibility metadata but never author awards or attribution.
```
