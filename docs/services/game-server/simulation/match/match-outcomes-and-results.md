# Match Outcomes And Results

Parent index: [Match](./!INDEX.md)

## Purpose

This document describes the current authoritative match-decision, result construction, and once-only summary boundary.

## Overview

Mode evaluation creates a locked `rules.MatchDecision` from authoritative match facts. The match-results package combines that decision with retained participant facts, team structure, objectives, missions, and challenge aggregates to produce one immutable `MatchSummary`.

```text
mode MatchFacts
-> modes.EvaluateMatch
-> locked rules.MatchDecision
+ retained participant/result facts
-> matchresults.ResolveDecision
-> EndOfMatchFlow.Run
-> immutable MatchSummary
-> room snapshot and reporting dispatcher
```

The result flow preserves removed participants. Active session removal therefore does not erase score, deaths, team, disposition, or outcome eligibility.

## Code root

```text
services/game-server/internal/matchresults/
services/game-server/internal/game/match_mode_evaluation.go
services/game-server/internal/rooms/
```

## Responsibilities

This boundary owns:

- mode-owned terminal status, end reason, winners, and participant decisions
- participant result ordering and placement
- won, lost, draw, completed, failed, and aborted outcomes
- participated, departed, forfeited, and late-join dispositions
- team aggregation and team outcomes
- single-player and multiplayer fallback result behavior
- objective, mission, and challenge result inclusion
- immutable cloning of result maps and slices
- once-only summary construction
- returning the same locked summary on repeated calls
- room storage, presentation projection, and reporting handoff

## Does not own

This boundary does not own:

- live score or death-counter mutation
- mode rule selection
- room transport
- client result-window layout
- player-data aggregate-stat mutation
- API database persistence
- achievement or progression policy

It consumes normalized facts from those owners.

## Domain roles

When a mode supplies participant outcomes, placements, completion times, targets, and winners, those locked values take precedence over generic score ranking.

Without a locked mode outcome, single-player participants complete, while multiplayer participants are ranked by score. A unique highest score wins; tied highest scores produce draws. Non-FFA team results aggregate player scores. Co-op teams complete together; competitive teams receive score-based outcomes unless the locked decision supplies them.

Score placements are stable and deterministic: participants are sorted by descending score and then player ID, with equal scores sharing placement.

`EndOfMatchFlow.Run` is idempotent. Its first successful call stores the summary; later calls return clones of the same result and report that no new summary was created.

## Protocols and APIs

Primary APIs include:

```go
func ResolveDecision(input BuildInput) (MatchDecision, error)
func NewEndOfMatchFlow() *EndOfMatchFlow
func (flow *EndOfMatchFlow) Run(input BuildInput) (MatchSummary, bool, error)
func (flow *EndOfMatchFlow) Summary() (MatchSummary, bool)
```

Rooms call the flow when the authoritative game decision becomes terminal. Room snapshots expose a presentation-safe projection. The match-result dispatcher and game-server reporting integration send durable facts separately.

## Data ownership

`BuildInput` carries match and trace identity, mode, session context, team structure, locked mode decision, end reason, retained participants, objectives, missions, and challenge aggregates.

`MatchSummary` owns the locked final projection:

```text
terminal status and end reason
participant outcomes, placements, dispositions, score, deaths, completion time, target
team outcomes, placements, and score
winning player and team references
objective, mission, and challenge resolutions
```

The room owns the stored summary for the current match. Player-data owns downstream aggregate stat mutation.

## Code map

```text
services/game-server/internal/matchresults/types.go
services/game-server/internal/matchresults/decision.go
services/game-server/internal/matchresults/team_decision.go
services/game-server/internal/matchresults/flow.go
services/game-server/internal/matchresults/dispatcher.go
services/game-server/internal/matchresults/clone.go
services/game-server/internal/game/match_mode_evaluation.go
services/game-server/internal/game/final_match_state_test.go
services/game-server/internal/rooms/room_match_summary.go
services/game-server/internal/rooms/room_match.go
services/game-server/internal/rooms/room_lifecycle.go
```

## Tests

Tests cover participant sorting and ties, locked mode decisions, team outcomes, clone isolation, once-only flow behavior, room summary storage, dispatcher behavior, Score Attack completion data, removed participant history, and presentation projection.

```text
services/game-server/internal/matchresults/*_test.go
services/game-server/internal/rooms/room_match_summary_test.go
services/game-server/internal/rooms/room_end_of_match_flow_test.go
services/game-server/tests/game/match_decision_test.go
```

## Related docs

- [Match](./!INDEX.md)
- [Modes And Match Rules](modes-and-match-rules.md)
- [Objective Runtime](objective-runtime.md)
- [Match End And Results Flow](../../../../domains/player-experience/match-end-and-results-flow.md)
- [Match Result Reporting](../../integrations/match-result-reporting.md)
- [Match Outcomes And Results Planning](../../../../planning/domains/gameplay/match-outcomes-and-results.md)

## Notes

The locked summary is the shared source for client presentation and durable reporting, but those paths receive different projections. Durable identity never needs to be exposed in the client result window.
