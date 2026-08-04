---
author: brian
created: "2026-07-19"
document_id: 019f7d55-fb2c-7180-81fa-db7af9247ca9
document_type: general
policy_exempt: false
summary: This doc plans EndOfMatchFlow, the result-orchestration seam that begins after the authoritative game/match layer locks a MatchDecision.
---
# Match Outcomes And Results
Parent index: [Gameplay Planning](./!INDEX.md)

## Purpose

This doc plans `EndOfMatchFlow`, the result-orchestration seam that begins after the authoritative game/match layer locks a `MatchDecision`.

The goal is to make match end a single explicit flow that consumes the locked decision, preserves final gameplay facts, emits one `MatchSummary`, and hands the dissected summary data to the systems that need it.

This doc is about match-end orchestration and result handoff. It is not only a result-data shape.

## Overview

This plan describes the current direction, ownership boundary, implementation status, remaining work, and open decisions for Match Outcomes And Results.

## Ownership Boundary

This doc owns:

```text
EndOfMatchFlow orchestration
one-time result orchestration after MatchDecision lock
runtime freeze handoff coordination
final MatchSummary/result-emission locking
MatchSummary emission
MatchSummaryDispatcher boundary
player participation finalization
objective resolution aggregation
mission resolution aggregation
challenge resolution aggregation
presentation-safe result handoff
client-impacting end-of-match sequencing
```

This doc does not own:

```text
mode rule definitions
scoring formulas
reward formulas
GrantAward construction
achievement definitions or evaluation
challenge definition behavior
challenge status vocabulary
player-data routing or storage
client UI layout
packet or API schema details
```

Modes and match rules own gameplay award selection, objective policy selection/composition, ranking, match-end, result, and related resolved policies. Detailed objective behavior and authoritative objective snapshots belong to [Objectives And Objective Runtime](objectives-and-objective-runtime.md). Teams and team rules owns team aggregation defaults and team-result requirements. The authoritative game/match layer evaluates final `MatchFacts` and locks `MatchDecision` once.

Modes and match rules provide the resolved match terminal status and participant/team result policy that `EndOfMatchFlow` hands off. This doc preserves the universal outcome vocabulary and the separation between terminal status, participant/team outcome, placement, and participation disposition without becoming an alternate owner of those policies.

Detailed award and counter semantics, including attribution, distribution, mutation, visibility, and idempotency, belong to [Gameplay Awards And Counters](gameplay-awards-and-counters.md). Its clean final award/counter snapshot is the authoritative input to result orchestration. `EndOfMatchFlow` consumes that snapshot and does not reconstruct awards, counters, attribution, or distribution from runtime entities.

`EndOfMatchFlow` aggregates the selected authoritative objective facts and final objective snapshot supplied by Objectives And Objective Runtime. It never reconstructs objective resolution, progress, contributors, or completion from runtime entities or client events.

Progression and rewards owns reward evaluation, XP, currency, unlocks, and `GrantAward` construction.

Achievements and milestones own achievement definitions, evaluation, and their fact-processing pipeline.

Levels, missions, and content structure owns mission/challenge catalog structure and challenge behavior.

Player-data owns persistence routing and storage.

## Core Architecture

The authoritative boundary is:

```text
game simulation
-> final gameplay mutations for tick complete
-> MatchFacts evaluated
-> MatchDecision locks once
-> gameplay mutation stops by default
-> EndOfMatchFlow consumes locked decision
```

Result orchestration then continues:

```text
locked MatchDecision
-> EndOfMatchFlow duplicate guard runs
-> player participation is finalized
-> objective, mission, and challenge resolutions are aggregated
-> one MatchSummary is emitted
-> MatchSummaryDispatcher dissects MatchSummary
-> downstream slices are sent to persistence, progression, achievement facts, and presentation
```

`EndOfMatchFlow` is the orchestration seam.

It does not discover match end, independently apply winner/loser, tie, draw, or forfeiture policy, or reconstruct the decision from room state. It consumes the authoritative locked `MatchDecision` and the final facts referenced by that decision.

`MatchSummary` is the one emitted end-of-match summary object.

`MatchSummaryDispatcher` is the small splitter/dispatcher seam that derives downstream slices from `MatchSummary`.

Downstream systems own interpretation and policy. The dispatcher only extracts and routes relevant summary data.

## EndOfMatchFlow

`EndOfMatchFlow` runs once per match.

It should guard against duplicate execution so repeated game-over ticks, repeated room snapshots, or client reconnects do not rebuild or mutate the final result.

Recommended execution order:

```text
1. Game simulation completes all gameplay mutations for the tick.
2. Resolved match rules evaluate normalized `MatchFacts`.
3. The authoritative game/match layer locks one `MatchDecision`.
4. Gameplay mutation stops by default.
5. EndOfMatchFlow consumes the locked decision and its final facts.
6. EndOfMatchFlow duplicate guard prevents repeated orchestration.
7. Historical participant, objective, mission, and challenge facts are aggregated.
8. MatchSummary is emitted.
9. MatchSummaryDispatcher dissects and dispatches summary slices.
10. Presentation-safe result data reaches client flow.
```

The first implementation should preserve current behavior while moving match-end work behind this explicit seam.

## Runtime Freeze

When `MatchDecision` locks, gameplay/world mutation should effectively freeze before result orchestration begins by default. Any explicit mode-specific bypass or override is already part of the resolved authoritative decision; `EndOfMatchFlow` does not choose it.

Frozen behavior includes:

```text
respawning
late join
spawning
score mutation
objective mutation
challenge mutation where end-locked
damage progression
pickup collection
world stepping
```

Continuing behavior includes:

```text
disconnect handling
rejoin handling
room/session lifecycle
result delivery
client navigation
post-match cleanup
```

`EndOfMatchFlow` should coordinate existing freeze and lifecycle seams where possible. It should not introduce a large new freeze mechanism unless implementation proves one is missing.

The important rule is that no gameplay or result mutation may change the locked decision or its final facts. `MatchSummary` is a projection of that stable boundary.

## MatchSummary

`MatchSummary` is the single emitted summary object for an ended match.

It contains authoritative final match facts.

Likely planning sections:

```text
match identity
resolved mode/result summary
participant/player summaries
objective resolutions
mission resolutions
challenge resolution aggregates
participation summary
presentation-source facts
```

Planned result facts include:

```text
mode_id
session_context
match_terminal_status
end_reason
participation disposition
winning_player_refs
winning_team_refs
participant outcomes
team outcomes
individual placements
participant result records
completion time
ranking inputs
final score
target values
```

`mode_id` is gameplay identity. `session_context` records hosting/session context such as single-player or multiplayer and must not be overloaded into mode identity.

Participant and team outcomes use the universal vocabulary:

```text
won
lost
draw
completed
failed
aborted
```

Match terminal status is a separate concept from participant/team outcome. It may be `completed`, `failed`, `cancelled`, `invalid`, or `administratively terminated`, as defined by [Modes And Match Rules](modes-and-match-rules.md). A participant may carry both a team outcome and an individual placement. Ties and multiple winners are valid, including co-op victories, team victories, and tied outcomes.

Forfeiture means leaving a match. It is participation disposition/end-reason context, not an automatic `won` or `lost` outcome. Normal metrics and the resolved mode result policy determine winners and losers after forfeiture. Competitive modes may abort when participation becomes invalid. When nobody satisfies the victory condition, the default result is `draw`; a mode may explicitly override that default.

`MatchSummary` should not contain downstream-specific derived sections such as:

```text
progression_inputs
achievement_facts
storage-specific data
packet-specific data
API-specific data
```

Progression inputs and achievement facts are derived by `MatchSummaryDispatcher`.

Storage-specific data is derived by `MatchSummaryDispatcher`.

Presentation-safe result data is derived by `MatchSummaryDispatcher`.

## Participant Identity

`MatchSummary` should identify participants through normalized player references.

Planning concept:

```text
player_ref
```

A player reference may represent a guest/session player, local profile player, or authenticated account player, but `MatchSummary` should not be conceptually built around storage-specific identity fields.

`MatchSummaryDispatcher` adapts player references for each destination:

```text
persistence slice
-> identity form required by player-data

progression slice
-> progression-eligible player reference

achievement fact slice
-> achievement-owning player reference

presentation slice
-> display-safe player identity
```

Presentation-safe output must not expose durable identity internals.

Current implementation may keep compatibility fields internally while the planned model moves toward normalized participant references.

Every participant receives a result record, including eliminated, departed/disconnected, forfeited, late-joining, and partial-match participants. Participants who took part before disconnecting remain represented in `MatchSummary`, with their accumulated historical facts and resolved outcome available even though disconnect removes them from active match-rule evaluation. Departed/disconnected players do not need a special player-facing result label.

## MatchSummaryDispatcher

`MatchSummaryDispatcher` dissects `MatchSummary` and dispatches the relevant slices.

Planned outputs:

```text
persistence / player-data slice
progression and rewards slice
achievement fact-pipeline slice
presentation-safe client slice
```

The dispatcher does not own reward formulas, achievement evaluation, persistence routing, or UI layout.

Correct relationship:

```text
MatchSummary
-> MatchSummaryDispatcher
-> persistence slice
-> player-data / persistence
```

```text
MatchSummary
-> MatchSummaryDispatcher
-> progression slice
-> Progression And Rewards
-> GrantAward construction
```

```text
MatchSummary
-> MatchSummaryDispatcher
-> achievement fact slice
-> achievement/milestone fact pipeline
```

```text
MatchSummary
-> MatchSummaryDispatcher
-> presentation-safe result slice
-> client result flow
```

The dispatcher should remain small. Its job is extraction and routing, not gameplay policy.

## Modes And Match Rules Relationship

Modes and match rules own:

```text
gameplay award policy
objective policy selection/composition
ranking policy
match-end policy
result policy
team-system selection and mode-specific team overrides
progression eligibility policy inputs
```

The authoritative game/match layer evaluates final `MatchFacts` under those resolved policies and locks `MatchDecision`. The resolved handoff includes the match terminal status, participant/team outcomes, placements, tie or multiple-winner results, and participation disposition context. `EndOfMatchFlow` consumes that locked decision; it does not discover match end or independently apply winner/loser, forfeiture, draw, or abort policy.

Baseline mode expectations:

```text
arcade_survival
- baseline rules with no mode-specific objective
- ends when no active participants remain
- records final match facts, including score and deaths where applicable
- no required competitive ranking
```

```text
score_attack
- objective is to reach target score
- ends immediately when the first valid participant reaches target
- also ends when no active participants remain before anyone succeeds
- ranks successful participants by completion time, lower is better
- records winner/success/failure, completion time, final score, deaths, and target score
```

FFA preserves standard individual player results. Team modes produce standard team results plus individual player results as required by [Teams And Team Rules](teams-and-team-rules.md). A participant may have both a team outcome and an individual placement. Result orchestration preserves the locked scopes, ties, and multiple winners instead of inferring teams or winners from score rows.

`EndOfMatchFlow` should not define mode behavior. It summarizes the locked decision and final facts produced by resolved mode policy.

The default no-victory result is a draw, and competitive participation invalidation may produce an abort, only when those policies are supplied by the resolved mode decision. `EndOfMatchFlow` records and presents those outcomes without independently applying either rule.

## Objectives And Missions

Objectives and missions may cause or contribute to match end.

Detailed objective behavior belongs to [Objectives And Objective Runtime](objectives-and-objective-runtime.md). Modes and match rules select and compose objective policy, while content planning supplies mission/content references. `EndOfMatchFlow` owns final aggregation of the selected authoritative objective facts into `MatchSummary`.

Planned aggregation:

```text
objective resolution facts
mission resolution facts
participant contribution facts where available
mode result facts
```

Mission completion facts may later feed progression and achievement facts through `MatchSummaryDispatcher`.

Missions do not directly grant rewards from this system.

## Challenge Resolution Aggregation

Challenges usually do not cause match end.

Challenges may resolve in different ways:

```text
immediately during the match
at mission completion
at match end
through accumulated progress finalized at match end
```

`MatchSummary` should aggregate challenge resolutions for the match.

This system does not define challenge behavior or challenge status vocabulary. The challenge/content system owns what a challenge resolution means.

`MatchSummary` records challenge resolution aggregates by challenge identity and relevant aggregation scope.

The aggregation seam should allow challenge results to be grouped by dimensions such as:

```text
challenge_id
mode_id
team_ref
player_ref
objective_ref
mission_ref
match-level or content ref
event/period ref if relevant
```

Exact aggregation dimensions and field names are gametime implementation decisions.

The planning requirement is that challenge aggregation must not be limited to one flat match-level result. Some challenges may need match-wide aggregation. Some may need per-player aggregation. Future team or mode-specific challenges may need team or mode aggregation.

Planned shape concept:

```text
ChallengeResolutionAggregate
- challenge_id
- aggregation_scope
- resolution
- source refs
- participant refs when relevant
- summary values
- metadata optional
```

`aggregation_scope` is the important seam, not an exact field contract.

Example scopes:

```text
match-wide challenge result
mode-specific challenge result
team challenge result
player challenge result
mission-attached challenge result
objective-attached challenge result
```

The dispatcher can derive progression inputs, achievement facts, and presentation rows from these challenge aggregates.

## Progression And Rewards Relationship

Progression inputs are not stored directly in `MatchSummary`.

`MatchSummary` contains final match facts. `MatchSummaryDispatcher` derives the progression slice.

Progression and rewards consumes trusted summary-derived inputs and owns:

```text
reward evaluation
XP awards
currency awards
unlock awards
rare persistent reward grants
GrantAward construction
idempotent grant IDs
player-data grant handoff
```

Correct flow:

```text
MatchSummary
-> MatchSummaryDispatcher
-> progression slice
-> Progression And Rewards
-> GrantAward
```

## Achievements And Milestones Relationship

Achievement facts are not stored directly in `MatchSummary`.

Achievements use the planned achievement-specific fact pipeline.

`MatchSummaryDispatcher` derives end-of-match achievement facts from `MatchSummary` and emits them into the achievement fact pipeline.

Possible end-of-match facts include:

```text
match completed
match won
score finalized
objective completed
mission completed
challenge completed
```

Exact achievement fact shapes belong to the achievement/milestone system.

This doc owns only the match-end handoff point.

## Player-Data And Persistence Relationship

Player-data and persistence are fed from the dissected `MatchSummary`.

The current `MatchResultSummary` can remain as the persistence-facing compatibility slice.

Correct flow:

```text
MatchSummary
-> MatchSummaryDispatcher
-> persistence slice / MatchResultSummary-compatible data
-> player-data runtime
-> persistence route
```

Player-data owns identity routing, store selection, and physical persistence.

This doc does not define database schema, local profile storage, account storage, or player-data transport details.

## Client Presentation Relationship

The client receives a presentation-safe result projection derived from `MatchSummary`.

Current presentation can remain small:

```text
player
deaths
score
```

Planned presentation may later include:

```text
mode result
success/failure
target score
completion time
ranking inputs / rank
player or team result scope
objective results
mission results
challenge resolutions
```

The result projection must exclude durable identity internals and storage-routing data.

`EndOfMatchFlow` affects the client because it defines when result presentation becomes valid and what final result facts are available. This doc does not define UI layout, scene hierarchy, button behavior, or packet shapes.

## Current Implementation Relationship

The current implementation already has a useful seed:

```text
room reaches game over
-> resolved MatchResultSummary is built
-> room stores the resolved summary once
-> room snapshot exposes match_result
-> client result flow presents result rows
-> player-data receives match-result reporting
```

The planned direction is to preserve that behavior while moving it behind the broader end-of-match flow:

```text
authoritative locked MatchDecision
-> EndOfMatchFlow
-> MatchSummary
-> MatchSummaryDispatcher
-> existing MatchResultSummary-compatible persistence slice
```

Early implementation should not break current single-player or multiplayer result reporting.

## Implementation Planning

Recommended implementation sequence:

```text
1. Consume the one-time locked MatchDecision from the authoritative game/match layer.
2. Add EndOfMatchFlow as the one-time result-orchestration seam.
3. Keep current match-result behavior working through EndOfMatchFlow.
4. Introduce the planned MatchSummary concept with participant and team outcome scopes.
5. Move current MatchResultSummary construction behind the MatchSummary path.
6. Add MatchSummaryDispatcher.
7. Route persistence and presentation-safe data through the dispatcher.
8. Add achievement and progression extraction through the dispatcher.
9. Add objective, mission, and challenge resolution aggregation with aggregation-scope support.
10. Preserve current result UI behavior while allowing richer result data later.
```

Early slices should prove orchestration and handoff before full challenge, mission, achievement, or progression mechanics exist.

The first useful slice can keep current result data but route it through `EndOfMatchFlow`.

## Testing Direction

Important future tests:

```text
EndOfMatchFlow runs once.
Repeated game-over snapshots do not rebuild final results.
MatchDecision locks once after final gameplay mutations for the tick.
Gameplay mutation stops by default after MatchDecision locks; any mode-resolved post-end bypass or override is explicit.
The locked MatchDecision and final authoritative result facts remain immutable after MatchDecision locks.
Respawning freezes by default after MatchDecision locks, subject to an explicit mode-resolved post-end override.
Late join freezes by default after MatchDecision locks, subject to an explicit mode-resolved post-end override.
Disconnect handling still works after match end.
Rejoin handling still works after match end.
Current persistence reporting still works.
MatchSummaryDispatcher sends persistence slice.
MatchSummaryDispatcher sends presentation-safe slice.
MatchSummaryDispatcher sends achievement fact-pipeline slice.
MatchSummaryDispatcher sends progression slice.
Presentation output excludes durable identity internals.
Score Attack terminates immediately when the first valid participant reaches target score.
Score Attack ranks successful participants by completion time, lower first.
Score Attack result records winner/success/failure, completion time, final score, deaths, and target score.
Arcade Survival records baseline final facts without a special objective or required ranking.
FFA results preserve player result scope.
Co-op, Custom Teams, and Auto-balanced Teams preserve resolved team and participant outcome scopes.
Disconnected participants retain accumulated facts and outcomes in MatchSummary.
Disconnected participants require no special player-facing result label.
Every participant receives a result record, including eliminated, departed/disconnected, forfeited, late-joining, and partial-match participants.
Participant and team outcomes use only won, lost, draw, completed, failed, or aborted.
Match terminal status remains separate from participant/team outcome and supports completed, failed, cancelled, invalid, or administratively terminated.
Participants can carry both a team outcome and an individual placement.
Ties and multiple winners are preserved for co-op, team, and tied outcomes.
Forfeiture is recorded as participation disposition/end-reason context and does not automatically assign won or lost.
Normal metrics and resolved mode result policy determine winners and losers after forfeiture.
Competitive modes can abort when participation becomes invalid.
When nobody satisfies the victory condition, the default result is draw unless the mode explicitly overrides it.
EndOfMatchFlow consumes locked MatchDecision without independently applying result or outcome policy.
session_context remains separate from gameplay mode_id.
Objective resolutions aggregate into MatchSummary.
Mission resolutions aggregate into MatchSummary.
Challenge resolutions aggregate by challenge_id.
Challenge resolutions can aggregate by player_ref.
Challenge resolutions can aggregate by team_ref.
Challenge resolutions can aggregate by mode or content refs where needed.
Immediate and end-of-match challenge resolutions can both appear in MatchSummary.
```

## Related Docs

* [Modes And Match Rules](modes-and-match-rules.md)
* [Gameplay Awards And Counters](gameplay-awards-and-counters.md)
* [Objectives And Objective Runtime](objectives-and-objective-runtime.md)
* [Teams And Team Rules](teams-and-team-rules.md)
* [Lives, Death, Elimination, And Respawn](lives-death-elimination-and-respawn.md)
* [Progression And Rewards](progression-and-rewards.md)
* [Levels, Missions, And Content Structure](levels-missions-and-content-structure.md)
* [Player Data And Persistence](../../services/player-data/!INDEX.md)
* [Client Match End Flow](../../../services/client/match-end-flow/!INDEX.md)

## Open Gametime Decisions

The universal outcome vocabulary, terminal-status separation, tie and multiple-winner support, forfeiture treatment, default draw, complete participant result coverage, and `EndOfMatchFlow` handoff authority are settled product decisions. Remaining decisions are implementation-level:

* Exact package placement for `EndOfMatchFlow`.
* Exact package placement for `MatchSummaryDispatcher`.
* Exact `MatchSummary` field names.
* Exact normalized player reference shape.
* Exact challenge aggregation scope representation.
* Exact objective and mission resolution summary shape.
* Exact handling for failed downstream dispatch.
* Exact retry/idempotency behavior for persistence, progression, and achievement dispatch.
* Exact client result projection shape.
* Exact migration path from current `MatchResultSummary` to the broader `MatchSummary` path.

## Core Invariants

```text
The authoritative game/match layer evaluates MatchFacts and locks MatchDecision once.

EndOfMatchFlow is the one-time result-orchestration seam that consumes the locked MatchDecision.

EndOfMatchFlow runs once per match.

EndOfMatchFlow does not discover match end or independently decide winners.

Gameplay/world state freezes by default when MatchDecision locks; an explicit mode-resolved bypass or override may alter post-end continuation.

The locked MatchDecision and final authoritative result facts remain immutable after MatchDecision locks.

Disconnect, rejoin, room/session lifecycle, result delivery, and cleanup can continue after match end.

MatchSummary is the one emitted end-of-match summary object.

MatchSummaryDispatcher dissects MatchSummary for downstream systems.

Progression inputs are derived from MatchSummary, not stored on it.

Achievement facts are derived from MatchSummary, not stored on it.

Player-data receives a dissected persistence slice.

Current MatchResultSummary can remain as the persistence-facing compatibility slice.

Presentation output is derived and presentation-safe.

Challenge resolutions aggregate into MatchSummary.

Challenge aggregation must support more than flat match-level summaries.

Challenge status meaning belongs to the challenge/content system.

Mode rules define match-end and result policy.
The universal participant/team outcome vocabulary is won, lost, draw, completed, failed, or aborted.
Match terminal status is separate from participant/team outcome.
Match terminal status may be completed, failed, cancelled, invalid, or administratively terminated.
Participants may have both a team outcome and an individual placement.
Ties and multiple winners are valid.
Forfeiture is participation disposition/end-reason context, not an automatic winner/loser outcome.
Normal metrics and resolved mode result policy determine winners and losers after forfeiture.
Competitive modes may abort when participation becomes invalid.
The default no-victory result is draw unless the mode explicitly overrides it.
Every participant receives a result record, including eliminated, departed/disconnected, forfeited, late-joining, and partial-match participants.
MatchSummary retains historical facts for participants who disconnected after participating.

session_context remains separate from gameplay mode_id.

EndOfMatchFlow consumes and emits locked results; it does not independently apply match-end, winner/loser, tie, forfeiture, draw, or abort policy, and does not define gameplay, reward, achievement, persistence, or UI policy.
```

## Notes

Implemented facts must move to canonical current documentation; this plan should retain only unresolved work, sequencing, and open decisions.
