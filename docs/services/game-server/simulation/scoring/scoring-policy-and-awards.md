# Scoring Policy And Awards

Parent index: [Scoring](./!README.md)

## Purpose

This document describes the game-server scoring policy and award boundary.

It covers the pure scoring package, the currently implemented scoring event, how score awards are calculated, and how game-owned code applies those awards to authoritative player counters.

## Overview

Scoring policy is a pure game-server simulation boundary.

The scoring package receives scoring facts as `scoring.Event` values and returns `scoring.Award` values. It does not mutate player sessions, inspect the game aggregate, check pause state, check invulnerability, emit packets, log, persist data, or update score counters.

Current scoring flow:

```text
asteroid destruction consequence
-> build scoring.Event
-> scoring.Policy.Evaluate
-> return []scoring.Award
-> game.awardScore
-> playerCanReceiveScore gate
-> addPlayerScoreLocked
-> score stored on playerSession
```

Only asteroid-destruction scoring is implemented. The default policy awards:

```text
constants.BaseScore / asteroid_size
```

`constants.BaseScore` is currently `120`, sourced from `shared/constants/server_constants.toml`.

## Code root

```text
services/game-server/internal/game/scoring/
```

Game-owned application code lives in:

```text
services/game-server/internal/game/scoring.go
services/game-server/internal/game/asteroid_destruction.go
services/game-server/internal/game/player_counters.go
```

## Responsibilities

The scoring policy boundary owns:

* The scoring event vocabulary.
* The scoring award data shape.
* Pure award calculation.
* Default policy construction.
* Rejecting unsupported scoring event kinds.
* Rejecting asteroid-destruction events without a player ID.
* Rejecting asteroid-destruction events with a non-positive asteroid size.
* Returning one or more award values for valid scoring facts.

The game-owned scoring application seam owns:

* Building scoring events from gameplay consequences.
* Applying awards to player counters.
* Rejecting awards with non-positive point values.
* Requiring an active player ship before score can be applied.
* Checking whether a player can currently receive score.
* Mutating the authoritative score counter.
* Logging successful score awards.

## Does not own

Scoring policy does not own:

* Player session mutation.
* Score counter storage.
* Player lives.
* Player death or respawn.
* Pause or suspension rules.
* Invulnerability rules.
* Collision detection.
* Damage resolution.
* Asteroid despawn timing.
* Asteroid fragment spawning.
* Pickup drop evaluation.
* Match-result summaries.
* Packet projection.
* Client HUD or match-results presentation.
* Player-data persistence.

Those concerns are owned by game-owned simulation adapters, player counter seams, combat flow, rooms, networking, or downstream persistence boundaries.

## Domain roles

Scoring policy participates in authoritative combat consequences.

Its role is narrow: convert a completed scoring fact into data-only awards. The gameplay layer decides when a fact exists, and the player counter layer decides how score is stored.

The current implemented scoring fact is:

```text
asteroid_destroyed
```

That fact can be produced when projectile or radial damage destroys an asteroid. Nonfatal asteroid hits do not create a scoring event and do not award score.

## Policy model

The current scoring package defines:

```go
type EventKind string

const (
    EventAsteroidDestroyed EventKind = "asteroid_destroyed"
)

type Event struct {
    Kind         EventKind
    PlayerID     string
    TargetID     string
    AsteroidSize int
}

type Award struct {
    PlayerID string
    Points   int
    Reason   EventKind
}

type Policy struct {
    BaseAsteroidDestroyedPoints int
}
```

`NewDefaultPolicy` creates a policy using:

```text
constants.BaseScore
```

`Policy.Evaluate` dispatches by event kind. Unknown event kinds return no awards.

For `asteroid_destroyed`, the policy returns no awards when:

```text
PlayerID == ""
AsteroidSize <= 0
```

For a valid asteroid-destruction event, it returns one award:

```text
PlayerID = event.PlayerID
Points   = policy.BaseAsteroidDestroyedPoints / event.AsteroidSize
Reason   = asteroid_destroyed
```

Current score examples:

```text
size 1 asteroid -> 120 points
size 2 asteroid -> 60 points
size 3 asteroid -> 40 points
```

`TargetID` is carried on the scoring event, but the current scoring calculation does not inspect it.

## Award application

`Game` owns award application.

The current destruction adapter builds the scoring event in:

```text
services/game-server/internal/game/asteroid_destruction.go
```

Flow:

```text
applyProjectileAsteroidDestruction(playerID, asteroid)
-> scoring.Event{
     Kind: asteroid_destroyed,
     PlayerID: playerID,
     TargetID: asteroid.ID,
     AsteroidSize: asteroid.Size,
   }
-> game.scoringPolicy.Evaluate
-> for each award: game.awardScore
```

After award evaluation, the same destruction path continues with:

```text
asteroid.MarkPendingDespawn
spawnAsteroidFragments
maybeDropPickupFromAsteroidLocked
```

Scoring does not own those consequences.

`game.awardScore` applies the award only when:

```text
award.Points > 0
game.entities.Players[award.PlayerID] exists
playerCanReceiveScore returns true
```

`playerCanReceiveScore` requires an existing player session and rejects suspended or invulnerable players.

Successful score application calls:

```text
addPlayerScoreLocked
```

The score is stored on `playerSession.Score`, not on the active `runtime.Ship`.

## Protocols and APIs

Scoring policy is not a network protocol boundary.

It is an internal Go API consumed by game-owned simulation code. The surface exists to separate pure award calculation from authoritative runtime mutation.

The internal API is:

```go
func NewDefaultPolicy() Policy
func (policy Policy) Evaluate(event Event) []Award
```

The caller supplies a scoring event. The policy returns data-only awards. The game aggregate owns authority behind applying those awards, and player counters own score storage.

No client can call the scoring policy directly. Clients observe score only through server-projected state packets and match-result summaries.

Packet-facing score projection belongs to player session state and match-result reporting, not this scoring policy package.

## Data ownership

Scoring policy owns no durable runtime state.

It reads generated server constants through:

```text
services/game-server/internal/constants/constants.go
```

The source constant is:

```text
shared/constants/server_constants.toml
```

Relevant source value:

```toml
[constants.server.scoring]
base_score = 120
```

Score mutation is owned by:

```text
game.playerSessions[player_id].Score
```

That counter is mutated through the player counter seam in:

```text
services/game-server/internal/game/player_counters.go
```

Scoring policy does not persist match stats or profile progression. Any durable persistence is downstream of match result and player-data integration.

## Code map

Primary scoring policy file:

```text
services/game-server/internal/game/scoring/scoring.go
```

Game-owned scoring application files:

```text
services/game-server/internal/game/game.go
services/game-server/internal/game/asteroid_destruction.go
services/game-server/internal/game/scoring.go
services/game-server/internal/game/player_counters.go
services/game-server/internal/game/pause.go
```

Gameplay producers:

```text
services/game-server/internal/game/combat.go
services/game-server/internal/game/simulation_radial_effects.go
```

Generated and source files:

```text
shared/constants/server_constants.toml
services/game-server/internal/constants/constants.go
```

Related tests:

```text
services/game-server/tests/scoring/policy_test.go
services/game-server/tests/game/collision_test.go
```

Important non-ownership boundaries:

```text
services/game-server/internal/game/damage/
services/game-server/internal/game/drops/
services/game-server/internal/game/spawning/
services/game-server/internal/game/player_counters.go
services/game-server/internal/rooms/
services/player-data/
client/
```

## Tests and verification

Focused scoring policy tests cover:

* Asteroid destruction awarding base score divided by asteroid size.
* Missing player ID returning no award.
* Non-positive asteroid size returning no award.
* Unknown event kind returning no award.

Combat integration tests cover:

* Destroyed asteroids awarding score.
* Non-destroying asteroid hits not awarding score.
* Score being based on asteroid size.
* Paused players not receiving score.
* Invulnerable players not receiving score.

Useful verification command:

```bash
cd services/game-server
go test -buildvcs=false ./...
```

Focused scoring verification:

```bash
cd services/game-server
go test -buildvcs=false ./tests/scoring ./tests/game -run 'AsteroidDestroyed|Score|PausedPlayerDoesNotScore|InvulnerablePlayerDoesNotScore'
```

## Related docs

* [Game Server Simulation Scoring](./!README.md)
* [Game Server Simulation](../!README.md)
* [Game Server](../../!README.md)
* [Player Counters](../players/player-counters.md)
* [Collision To Damage Flow](../combat/collision-to-damage-flow.md)
* [Damage Resolution](../combat/damage-resolution.md)
* [Radial Effects](../combat/radial-effects.md)
* [Pickup Drop Integration](../pickups/pickup-drop-integration.md)
* [Game Aggregate](../runtime/game-aggregate.md)
* [Data](../../../../data/!README.md)

## Notes

The current scoring package is intentionally smaller than the player counter documentation. This document owns score-award policy only; player counter storage, mutation, packet projection, devtools score commands, and match-result readback belong in [Player Counters](../players/player-counters.md).

Score awards currently have no upper bound beyond the integer value produced by policy and the player counter seam. The counter seam clamps only the lower bound at zero.
