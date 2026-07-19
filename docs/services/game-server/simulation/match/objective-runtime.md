# Objective Runtime

Parent index: [Match](./!INDEX.md)

## Purpose

This document describes the current schema-driven objective definition and runtime boundary.

## Overview

Objective definitions declare scope, success and optional failure conditions, lifecycle, visibility, timers, attribution, and associations. Runtime instances consume normalized facts and emit normalized events and viewer-filtered snapshots.

```text
validated Definition
-> Runtime.RegisterDefinition
-> CreateInstance
-> ApplyFacts / Step
-> objective events
-> viewer-filtered snapshots
-> mode match evaluation and results
```

The current Score Attack mode creates one active player-scoped numeric objective per participant, driven by the authoritative `SCORE` counter.

## Code root

```text
services/game-server/internal/game/objectives/
services/game-server/internal/game/objective_runtime.go
services/game-server/internal/game/control_objectives.go
```

## Responsibilities

This boundary owns:

- objective definition and instance identity
- player, team, match, collection, and definition-specific scopes
- manual, boolean, numeric, set, sequence, and maintain conditions
- normalized fact operations
- allowed attribution kinds
- discovery, activation, completion, failure, cancellation, and retirement state
- optional timers and explicit expiry outcomes
- progress decrease, reset, and overflow policy
- association-key validation
- definition retirement and active-instance retirement
- deterministic events and snapshots
- owner and discovery visibility filtering
- pause-aware timer and maintain-condition advancement

## Does not own

This boundary does not own:

- deciding which gameplay event creates a fact
- mutating score or other source counters
- match-end policy
- rewards or progression
- client objective UI layout
- mission or campaign composition
- persistence

Gameplay adapters translate authoritative events into objective facts. Modes decide whether objective outcomes contribute to match completion.

## Domain roles

Facts are the only normal input to active condition evaluation. Supported operations include signal, increment, set, reset, add-member, and remove-member.

Attribution can be constrained to one-hit, in-game, or in-encounter participation. Team ownership is supplied by authoritative team membership; the objective runtime does not assign teams.

Definitions are immutable once registered. Retiring a definition blocks new instances and transitions its nonterminal instances to `retired` with `definition_retired` as the reason.

## Protocols and APIs

Important APIs include:

```go
func (runtime *Runtime) RegisterDefinition(definition Definition) error
func (runtime *Runtime) CreateInstance(definitionID DefinitionID, registration Registration) (InstanceID, []Event, error)
func (runtime *Runtime) ApplyFacts(instanceID InstanceID, facts []Fact) ([]Event, error)
func (runtime *Runtime) ApplyFactsToScope(scope Scope, ownerID string, facts []Fact) ([]Event, error)
func (runtime *Runtime) Step(delta float64, simulationPaused bool) []Event
func (runtime *Runtime) RetireDefinition(id DefinitionID) ([]Event, error)
func (runtime *Runtime) Snapshot(id InstanceID, viewer Viewer) (Snapshot, bool)
func (runtime *Runtime) Snapshots(viewer Viewer) []Snapshot
```

Game integration publishes objective events into the authoritative presentation event flow and projects snapshots for later consumers.

## Data ownership

Definitions own the condition and lifecycle schema. Instances copy the definition and own current status, discovery, success/failure condition state, timer remaining, failure reason, owner, and associations.

Snapshots include definition and instance IDs, scope, owner, associations, status, discovery, progress, target, timer, failure reason, sequence index, set members, and last attribution. Maps and slices are cloned at boundaries.

## Code map

```text
services/game-server/internal/game/objectives/contracts.go
services/game-server/internal/game/objectives/definition.go
services/game-server/internal/game/objectives/condition.go
services/game-server/internal/game/objectives/runtime.go
services/game-server/internal/game/objectives/lifecycle.go
services/game-server/internal/game/objectives/transitions.go
services/game-server/internal/game/objectives/snapshot.go
services/game-server/internal/game/objective_runtime.go
services/game-server/internal/game/control_objectives.go
services/game-server/internal/game/match_rules.go
```

## Tests

Tests cover every condition family, validation, discovery and visibility, timers, attribution, decrease/reset behavior, definition retirement, deterministic snapshots, event publication, and Score Attack integration.

```text
services/game-server/internal/game/objectives/*_test.go
services/game-server/internal/game/objective_runtime_integration_test.go
```

## Related docs

- [Match](./!INDEX.md)
- [Modes And Match Rules](modes-and-match-rules.md)
- [Match Outcomes And Results](match-outcomes-and-results.md)
- [Awards And Counter Runtime](../scoring/awards-and-counter-runtime.md)
- [Objectives And Objective Runtime Planning](../../../../planning/domains/gameplay/objectives-and-objective-runtime.md)

## Notes

Objectives can emit terminal facts, but they do not directly end a match. The selected mode consumes those facts and locks the authoritative match decision.
