# Match Outcomes And Results
Parent index: [Gameplay Planning](./!INDEX.md)

## Purpose

This doc plans `EndOfMatchFlow`, the authoritative match-end orchestration seam for Space Rocks.

The goal is to make match end a single explicit flow that freezes final gameplay state, finalizes scoring and result facts, emits one `MatchSummary`, and hands the dissected summary data to the systems that need it.

This doc is about match-end orchestration and result handoff. It is not only a result-data shape.

## Ownership Boundary

This doc owns:

```text
EndOfMatchFlow orchestration
one-time match-end execution
runtime freeze coordination
final score/result locking
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

Modes and match rules own match-end policy, scoring policy, objective policy, and result policy.

Progression and rewards owns reward evaluation, XP, currency, unlocks, and `GrantAward` construction.

Achievements and milestones own achievement definitions, evaluation, and their fact-processing pipeline.

Levels, missions, and content structure owns mission/challenge catalog structure and challenge behavior.

Player-data owns persistence routing and storage.

## Core Architecture

The planned end-of-match flow is:

```text
match-end condition reached
-> EndOfMatchFlow duplicate guard runs
-> runtime freeze/end seams are applied
-> scoring/result policy is finalized
-> player participation is finalized
-> objective, mission, and challenge resolutions are aggregated
-> one MatchSummary is emitted
-> MatchSummaryDispatcher dissects MatchSummary
-> downstream slices are sent to persistence, progression, achievement facts, and presentation
```

`EndOfMatchFlow` is the orchestration seam.

`MatchSummary` is the one emitted end-of-match summary object.

`MatchSummaryDispatcher` is the small splitter/dispatcher seam that derives downstream slices from `MatchSummary`.

Downstream systems own interpretation and policy. The dispatcher only extracts and routes relevant summary data.

## EndOfMatchFlow

`EndOfMatchFlow` runs once per match.

It should guard against duplicate execution so repeated game-over ticks, repeated room snapshots, or client reconnects do not rebuild or mutate the final result.

Recommended execution order:

```text
1. Match-end condition is detected from resolved match rules.
2. EndOfMatchFlow duplicate guard runs.
3. Runtime freeze/end seams are applied.
4. Scoring/result policy is finalized.
5. Player participation is finalized.
6. Objective, mission, and challenge resolutions are aggregated.
7. MatchSummary is emitted.
8. MatchSummaryDispatcher dissects and dispatches summary slices.
9. Presentation-safe result data reaches client flow.
```

The first implementation should preserve current behavior while moving match-end work behind this explicit seam.

## Runtime Freeze

At match end, gameplay/world state should effectively freeze.

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

The important rule is that final match facts must not drift after `MatchSummary` emission.

## MatchSummary

`MatchSummary` is the single emitted summary object for a completed match.

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
match-end policy
objective policy
scoring policy
result policy
progression eligibility policy inputs
```

`EndOfMatchFlow` consumes resolved match rules and finalizes results according to those rules.

Baseline mode expectations:

```text
survival_arcade
- ends on all players eliminated
- records score and deaths
- completed, not necessarily won
```

```text
score_attack
- ends on target score reached or all players eliminated
- records target_score
- records success/failure
- records score and deaths
```

`EndOfMatchFlow` should not define mode behavior. It finalizes and summarizes the facts produced by resolved mode policy.

## Objectives And Missions

Objectives and missions may cause or contribute to match end.

Objective and mission behavior belongs to modes, match rules, and content planning. `EndOfMatchFlow` owns final aggregation into `MatchSummary`.

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
EndOfMatchFlow
-> MatchSummary
-> MatchSummaryDispatcher
-> existing MatchResultSummary-compatible persistence slice
```

Early implementation should not break current single-player or multiplayer result reporting.

## Implementation Planning

Recommended implementation sequence:

```text
1. Add EndOfMatchFlow as the one-time match-end orchestration seam.
2. Keep current match-result behavior working through EndOfMatchFlow.
3. Introduce the planned MatchSummary concept.
4. Move current MatchResultSummary construction behind the MatchSummary path.
5. Add MatchSummaryDispatcher.
6. Route persistence data through the dispatcher.
7. Add presentation-safe result projection through the dispatcher.
8. Add achievement fact-pipeline extraction through the dispatcher.
9. Add progression extraction through the dispatcher.
10. Add objective and mission resolution aggregation.
11. Add challenge resolution aggregation with aggregation-scope support.
12. Preserve current result UI behavior while allowing richer result data later.
```

Early slices should prove orchestration and handoff before full challenge, mission, achievement, or progression mechanics exist.

The first useful slice can keep current result data but route it through `EndOfMatchFlow`.

## Testing Direction

Important future tests:

```text
EndOfMatchFlow runs once.
Repeated game-over snapshots do not rebuild final results.
Runtime mutation freezes after match end.
Respawning freezes after match end.
Late join freezes after match end.
Disconnect handling still works after match end.
Rejoin handling still works after match end.
Current persistence reporting still works.
MatchSummaryDispatcher sends persistence slice.
MatchSummaryDispatcher sends presentation-safe slice.
MatchSummaryDispatcher sends achievement fact-pipeline slice.
MatchSummaryDispatcher sends progression slice.
Presentation output excludes durable identity internals.
Score Attack result records target score and success/failure.
Survival Arcade preserves current result behavior.
Objective resolutions aggregate into MatchSummary.
Mission resolutions aggregate into MatchSummary.
Challenge resolutions aggregate by challenge_id.
Challenge resolutions can aggregate by player_ref.
Challenge resolutions can aggregate by team_ref later.
Challenge resolutions can aggregate by mode or content refs where needed.
Immediate and end-of-match challenge resolutions can both appear in MatchSummary.
```

## Related Docs

* [Modes And Match Rules](modes-and-match-rules.md)
* [Progression And Rewards](progression-and-rewards.md)
* [Progression And Rewards](progression-and-rewards.md)
* [Levels, Missions, And Content Structure](levels-missions-and-content-structure.md)
* [Player Data And Persistence](../../services/player-data/!INDEX.md)
* [Client Match End Flow](../../../services/client/match-end-flow/!INDEX.md)

## Open Gametime Decisions

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
EndOfMatchFlow is the authoritative match-end orchestration seam.

EndOfMatchFlow runs once per match.

Gameplay/world state freezes at match end.

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

EndOfMatchFlow finalizes and emits results; it does not define reward, achievement, persistence, or UI policy.
```
