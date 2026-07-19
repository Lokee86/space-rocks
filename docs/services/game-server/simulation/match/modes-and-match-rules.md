# Modes And Match Rules

Parent index: [Match](./!INDEX.md)

## Purpose

This document describes the current room-mode configuration and immutable resolved-match-rules boundary.

## Overview

A room accepts a small user-facing mode configuration, validates it with the selected team structure, and resolves it into one complete rules object before participation begins.

```text
RoomModeConfig + teams.Config
-> NormalizeRoomModeConfig
-> ValidateRoomModeConfig
-> modes.Resolve
-> ResolvedMatchRules
-> Game.ConfigureMatchRules
-> focused owner runtimes
```

Implemented presets are `arcade_survival` and `score_attack`. Arcade Survival is the default. Score Attack defaults to a target score of 1000 when no positive target is supplied.

## Code root

```text
services/game-server/internal/game/modes/
services/game-server/internal/game/match_rules.go
```

Room selection and projection live under `services/game-server/internal/rooms/` and `services/game-server/internal/networking/`.

## Responsibilities

This boundary owns:

- normalizing and validating room mode options
- validating starting-lives versus infinite-lives configuration
- validating Score Attack target score
- combining mode and team configuration
- selecting lives, award, objective, ranking, end-condition, result, damage, spawn, joining, progression, and freeze policies
- cloning resolved slice-backed rules at boundaries
- configuring focused game runtimes before participation begins
- rejecting unsupported policies or late reconfiguration
- evaluating current match facts through the selected mode

## Does not own

This boundary does not own:

- team assignment execution
- life-counter mutation
- objective condition evaluation
- award-counter mutation
- encounter scheduling
- damage calculation
- room transition execution
- client mode-selection UI
- durable match-result persistence

It selects those policies and passes them to the owning systems.

## Domain roles

Arcade Survival resolves to:

```text
mode: arcade_survival
end condition: no_active_participants
result policy: final_facts
ranking: none
player PvP: disabled
in-game joining: disabled
player spawn: basic_safe_spawn_v1
encounter spawn: playercentric_asteroids_v1
progression eligible: true
freeze gameplay on end: true
```

Score Attack resolves to the same baseline owners plus:

```text
mode: score_attack
objective: score_attack_target_v1
ranking: completion_time
end precedence: target_score_reached, then no_active_participants
result policy: score_attack
```

The first player to record a successful target-score completion order wins. If no active participant remains before anyone succeeds, the match fails.

## Protocols and APIs

Primary APIs are:

```go
func DefaultRoomModeConfig() RoomModeConfig
func NormalizeRoomModeConfig(config RoomModeConfig) RoomModeConfig
func ValidateRoomModeConfig(config RoomModeConfig) error
func Resolve(config RoomModeConfig, teamConfig teams.Config) (ResolvedMatchRules, error)
func EvaluateMatch(resolved ResolvedMatchRules, facts MatchFacts) rules.MatchDecision

func (game *Game) ConfigureMatchRules(resolved modes.ResolvedMatchRules) error
func (game *Game) ResolvedMatchRules() modes.ResolvedMatchRules
```

`ConfigureMatchRules` must run before any participant record exists and before a final state has been locked.

## Data ownership

`RoomModeConfig` is the room-facing selection shape:

```text
preset_id
starting_lives
infinite_lives
target_score
```

`ResolvedMatchRules` is the authoritative match-start product. It contains the selected team config, lives policy, award policy ID, objective policy, ranking metric, ordered match-end conditions, result policy, player-damage flag, spawn profiles, joining policy, progression eligibility, and match-end freeze behavior.

The game stores a clone. Callers receive clones, preventing mutation after configuration.

## Code map

```text
services/game-server/internal/game/modes/types.go
services/game-server/internal/game/modes/resolve.go
services/game-server/internal/game/modes/evaluate.go
services/game-server/internal/game/match_rules.go
services/game-server/internal/game/match_mode_evaluation.go
services/game-server/internal/rooms/room_mode.go
services/game-server/internal/rooms/room_creation.go
services/game-server/internal/networking/room_snapshot.go
shared/packets/lobby.toml
```

## Tests

Tests cover defaults, invalid option combinations, both presets, clone behavior, room creation/projection, pre-participation configuration, Score Attack success ordering, failure when players are exhausted, and integration with lives/objectives/results.

```text
services/game-server/internal/game/modes/*_test.go
services/game-server/internal/game/mode_rules_integration_test.go
services/game-server/internal/rooms/room_mode_test.go
services/game-server/internal/networking/room_mode_snapshot_test.go
```

## Related docs

- [Match](./!INDEX.md)
- [Objective Runtime](objective-runtime.md)
- [Match Outcomes And Results](match-outcomes-and-results.md)
- [Team Configuration And Membership](../../rooms/team-configuration-and-membership.md)
- [Modes And Match Rules Planning](../../../../planning/domains/gameplay/modes-and-match-rules.md)

## Notes

The current surface intentionally exposes only the options supported by implemented owners. Additional modes should extend resolution rather than scattering mode-ID conditionals through combat, lives, spawning, or results code.
