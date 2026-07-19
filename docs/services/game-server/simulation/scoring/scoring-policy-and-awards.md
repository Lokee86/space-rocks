---
author: brian
created: "2026-07-19"
document_id: 019f7d55-fb2c-7783-903b-81bb1a9107c7
document_type: general
policy_exempt: false
summary: This document describes asteroid scoring policy and its handoff into authoritative awards and counters.
---
# Asteroid Scoring Policy

Parent index: [Scoring And Awards](./!INDEX.md)

## Purpose

This document describes the pure asteroid-destruction score calculation and its handoff into authoritative counter application.

## Overview

```text
asteroid destruction fact
-> scoring.Event
-> scoring.Policy.Evaluate
-> scoring.Award
-> game award adapter
-> awards counter runtime and player-session score projection
```

The default score is `constants.BaseScore / asteroid_size`. `BaseScore` is currently 120, producing 120, 60, and 40 points for sizes one, two, and three.

## Code root

```text
services/game-server/internal/game/scoring/
```

Application code lives in `services/game-server/internal/game/scoring.go` and `asteroid_destruction.go`.

## Responsibilities

This boundary owns:

- asteroid-destruction scoring event vocabulary
- pure award calculation
- missing-player and nonpositive-size rejection
- base-score constant consumption
- data-only score awards

The game adapter owns applying valid positive awards to authoritative counters.

## Does not own

This boundary does not own:

- generalized counter storage or idempotence
- damage and destruction detection
- player eligibility state
- combos, streaks, or assists
- match outcomes
- client presentation
- persistence

## Domain roles

The policy turns one completed asteroid-destruction fact into a score amount. It never mutates a session or runtime entity. The generalized awards runtime owns event application and scoped counter state.

## Protocols and APIs

```go
func NewDefaultPolicy() Policy
func (policy Policy) Evaluate(event Event) []Award
```

The only current event kind is `asteroid_destroyed`.

## Data ownership

`scoring.Event` carries player ID, target ID, and asteroid size. `scoring.Award` carries player ID, points, and reason. Neither is durable state.

The authoritative match-local score is applied downstream and mirrored to `playerSession.Score` for current projections and results.

## Code map

```text
services/game-server/internal/game/scoring/scoring.go
services/game-server/internal/game/scoring.go
services/game-server/internal/game/asteroid_destruction.go
services/game-server/internal/game/award_runtime.go
services/game-server/internal/game/player_counters.go
shared/constants/server_constants.toml
```

## Tests

Tests cover size-based score, invalid event rejection, destroyed versus nondestroyed asteroid behavior, and player eligibility gates.

```text
services/game-server/tests/scoring/policy_test.go
services/game-server/tests/game/collision_test.go
services/game-server/internal/game/awards_integration_test.go
```

## Related docs

- [Scoring And Awards](./!INDEX.md)
- [Awards And Counter Runtime](awards-and-counter-runtime.md)
- [Player Counters](../players/player-counters.md)
- [Damage Resolution](../combat/damage-resolution.md)

## Notes

This policy remains intentionally narrow. New scoring sources should produce normalized awards without expanding the pure asteroid formula into a general state owner.
