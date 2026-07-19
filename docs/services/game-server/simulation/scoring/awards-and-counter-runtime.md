# Awards And Counter Runtime

Parent index: [Scoring](./!INDEX.md)

## Purpose

This document describes the current generalized awards, counters, contribution, combo, and streak runtime.

## Overview

Gameplay adapters convert authoritative consequences into idempotent counter mutations. The awards runtime validates the event, applies all mutations, tracks optional contribution/combo/streak state, and returns detailed changes for synchronization with current player-session counters.

```text
damage / destruction / death / objective fact
-> award adapter and policy
-> event ID + []awards.Mutation
-> awards.Runtime.ApplyEvent
-> []awards.Change
-> player-session compatibility counters and result facts
```

The older pure scoring policy still calculates asteroid score values. The awards runtime is the broader owner for normalized counters and event application.

## Code root

```text
services/game-server/internal/game/awards/
services/game-server/internal/game/award_runtime.go
```

## Responsibilities

This boundary owns:

- fixed counter vocabulary
- validated custom counter registration
- player, team, match, and objective counter scopes
- counter visibility policy
- increment, decrement, set, minimum, maximum, and timed-accumulation mutations
- idempotent event IDs
- detailed before/after/delta change results
- deterministic counter snapshots
- derived team totals from authoritative membership
- contribution retention for assist eligibility
- combo state and multiplier policy
- named streak state
- owner removal and contribution cleanup
- standard award-policy normalization

Fixed counters are:

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

## Does not own

This boundary does not own:

- deciding whether damage or destruction occurred
- team membership
- objective condition evaluation
- match ranking or victory
- player-data persistence
- client scoreboard layout
- durable progression or achievements

It records normalized match-local values for those consumers.

## Domain roles

Custom counter IDs use uppercase identifier form and cannot shadow fixed IDs. Visibility can be hidden, HUD, scoreboard, results-only, team-only, player-private, or spectator-visible.

`ApplyEvent` validates every mutation before changing state. An event ID is marked processed only after successful application. Repeating the same ID returns a duplicate result without applying changes again.

The standard policy ID is `standard_awards_v1`. Baseline assists are disabled, but the policy seam carries the intended five-second contribution window and five-percent threshold. Combo state is enabled for `SCORE`, and the baseline named kill streak is `pvp_kills`.

## Protocols and APIs

Important APIs include:

```go
func NewRuntime() *Runtime
func (runtime *Runtime) RegisterCustomCounter(id CounterID) error
func (runtime *Runtime) SetCounterVisibility(id CounterID, visibility Visibility) error
func (runtime *Runtime) ApplyEvent(eventID string, mutations []Mutation) (EventResult, error)
func (runtime *Runtime) Counter(owner Owner, id CounterID) (float64, bool)
func (runtime *Runtime) Snapshot() []CounterSnapshot
func (runtime *Runtime) DerivedTeamTotals(memberships map[string]string, ids []CounterID) []CounterSnapshot
func (runtime *Runtime) RemoveOwner(owner Owner)
```

Game-owned award adapters create stable event IDs from authoritative consequences and mirror relevant counter changes into existing session projections.

## Data ownership

The awards runtime owns match-local counter values, processed event IDs, combo state, streak state, and contribution records.

`playerSession.Score` and related existing fields remain compatibility/projection state while consumers migrate to the generalized counter runtime. They do not replace the awards runtime's idempotence and scoped counter ownership.

Team totals are derived from player values and supplied membership; they are not a second independently mutable copy unless a mode explicitly writes team-scoped counters.

## Code map

```text
services/game-server/internal/game/awards/contracts.go
services/game-server/internal/game/awards/policy.go
services/game-server/internal/game/awards/runtime.go
services/game-server/internal/game/awards/progress.go
services/game-server/internal/game/awards/contributions.go
services/game-server/internal/game/award_runtime.go
services/game-server/internal/game/damage_awards.go
services/game-server/internal/game/destruction_awards.go
services/game-server/internal/game/death_awards.go
services/game-server/internal/game/player_counters.go
services/game-server/internal/game/scoring.go
```

## Tests

Tests cover fixed/custom counter validation, every mutation operation, event idempotence, visibility, deterministic snapshots, team totals, contribution retention, assist thresholds, combo and streak behavior, owner removal, and game integration.

```text
services/game-server/internal/game/awards/*_test.go
services/game-server/internal/game/awards_integration_test.go
```

## Related docs

- [Scoring](./!INDEX.md)
- [Scoring Policy And Awards](scoring-policy-and-awards.md)
- [Player Counters](../players/player-counters.md)
- [Damage Resolution](../combat/damage-resolution.md)
- [Objective Runtime](../match/objective-runtime.md)
- [Gameplay Awards And Counters Planning](../../../../planning/domains/gameplay/gameplay-awards-and-counters.md)

## Notes

The current runtime establishes the generalized owner seam while preserving existing score/session projections. Broader client presentation and durable progression remain separate work.
