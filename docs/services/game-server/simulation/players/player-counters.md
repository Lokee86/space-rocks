# Player Counters

Parent index: [Game Server Simulation Players](./!INDEX.md)

## Purpose

This document describes game-server player counter ownership.

It covers the current score and lives mutation seams, how those counters are stored on player sessions, which gameplay systems call them, how counter changes are projected to state packets, and what this boundary does not own.

## Overview

Player counters are authoritative game-server simulation state.

The current counter boundary covers:

```text
score
lives
```

Both counters are stored on `playerSession`, not on the live `runtime.Ship` avatar. The live ship owns movement, health, shields, weapons, input, targeting, invulnerability, and despawn state. The player session owns durable per-match player state that must survive live ship removal and respawn.

The core mutation seam is:

```text
services/game-server/internal/game/player_counters.go
```

The seam supports set and add operations for score and lives:

```text
SetPlayerScore
AddPlayerScore
SetPlayerLives
AddPlayerLives
```

Each operation returns a `PlayerCounterChange` describing whether the player session was found, the previous value, the updated value, and the delta.

Counter values are clamped at zero. Negative set values become zero, and negative add operations cannot reduce a counter below zero.

## Code root

```text
services/game-server/internal/game/
```

Supporting packages and generated outputs:

```text
services/game-server/internal/game/scoring/
services/game-server/internal/game/pickups/
services/game-server/internal/devtools/
services/game-server/internal/constants/
services/game-server/internal/rooms/
shared/constants/
shared/packets/
```

## Responsibilities

Player counters own:

* Authoritative per-match score storage.
* Authoritative per-match lives storage.
* Score and lives set/add mutation helpers.
* Counter clamping to prevent negative score or lives values.
* Thread-safe public mutation entry points for callers outside locked game sections.
* Locked mutation helpers for same-package simulation code already holding game authority.
* Returning `PlayerCounterChange` results for callers that need mutation feedback.
* Keeping score and lives on `playerSession`.
* Projecting score and lives through player session state.
* Supplying score and deaths to match-result summary facts.
* Providing narrow game-owned devtools adapters for score/lives mutation commands.

## Does not own

Player counters do not own:

* Scoring policy.
* Asteroid destruction scoring rules.
* Player death detection.
* Player despawn timing.
* Respawn eligibility or respawn placement.
* Player health, shields, invulnerability, or damage modifiers.
* Pickup collection rules.
* Pickup entity lifecycle.
* Weapon pickup equipment or ammo mutation.
* Match-over policy.
* Room result reporting.
* Player-data persistence.
* Client HUD or match result presentation.
* Packet schema source-of-truth files.
* WebSocket transport or packet encoding.
* Devtools command routing.

Those systems may read or call the counter seam, but they own their own decisions and side effects.

## Domain roles

Player counters participate in the player, combat, pickup, and match-result domains by carrying the durable score and lives values that outlive an active ship.

They do not own the gameplay events or rules that decide when those counters change.

## Counter storage model

The current authoritative storage is:

```go
type playerSession struct {
    Score int
    Lives int
}
```

A new player session starts with:

```text
Score: 0
Lives: constants.PlayerStartingLives
```

`PlayerStartingLives` is generated from shared constants and is currently sourced from:

```text
shared/constants/server_entities.toml
```

The generated Go constant currently lives in:

```text
services/game-server/internal/constants/constants.go
```

Live ships do not store score or lives. This is intentional because players can be pending respawn or eliminated while their session still exists. During those states, the player may be absent from `game.entities.Players` but still present in `game.playerSessions`.

## Mutation model

The public mutation surface is:

```go
func (game *Game) SetPlayerScore(playerID string, score int) PlayerCounterChange
func (game *Game) AddPlayerScore(playerID string, amount int) PlayerCounterChange
func (game *Game) SetPlayerLives(playerID string, lives int) PlayerCounterChange
func (game *Game) AddPlayerLives(playerID string, amount int) PlayerCounterChange
```

These public methods lock `game.mu` and then delegate to locked helpers.

The same-package locked helpers are:

```go
setPlayerScoreLocked
addPlayerScoreLocked
setPlayerLivesLocked
addPlayerLivesLocked
```

Simulation code that already runs inside locked game-state mutation paths uses the locked helpers directly.

All set and add operations route through the same clamp helper:

```go
func clampPlayerCounter(value int) int {
    if value < 0 {
        return 0
    }

    return value
}
```

Set operations write the clamped target value.

Add operations read the current session value, add the requested amount, then route through the corresponding set operation so clamping and change reporting remain consistent.

## Change result

Every mutation returns:

```go
type PlayerCounterChange struct {
    PlayerID string
    Found    bool
    Before   int
    After    int
    Delta    int
}
```

`Found` is true only when `game.playerSessions[playerID]` exists.

When the session is missing, the returned change contains the requested `PlayerID` and `Found: false`. No mutation is applied.

When the session is found:

```text
Before = previous counter value
After  = clamped updated counter value
Delta  = After - Before
```

Callers use this result differently depending on the flow:

* devtools score/lives commands use `Found` to decide whether the command affected a player
* pickup effects use `Found` and `After` before emitting `pickup_effect_applied`
* score award logging uses `After` as the new score value

`PlayerCounterChange` itself does not emit events, write packets, persist stats, or log by default. Those side effects stay with the caller that owns the gameplay context.

## Score flow

Score policy is separate from score mutation.

Pure scoring policy lives in:

```text
services/game-server/internal/game/scoring/
```

The scoring package receives scoring facts and returns awards. It does not mutate game state.

The current implemented score event is:

```text
asteroid_destroyed
```

The current default policy awards:

```text
constants.BaseScore / asteroid_size
```

`BaseScore` is generated from shared constants and is currently sourced from:

```text
shared/constants/server_constants.toml
```

The game-owned score application flow is:

```text
projectile destroys asteroid
-> game.applyProjectileAsteroidDestruction
-> scoring.Policy.Evaluate
-> game.awardScore
-> playerCanReceiveScore gate
-> addPlayerScoreLocked
-> score stored on playerSession
```

`game.awardScore` rejects awards when:

```text
award points <= 0
player has no active ship
player cannot receive score
```

`playerCanReceiveScore` requires the player session to exist and rejects suspended or invulnerable players. This keeps score mutation outside the pure scoring policy package while still letting the game layer enforce current gameplay gates.

## Lives flow

Lives mutate through the same player counter seam.

The main gameplay callers are:

```text
fatal player damage
pickup add_lives effect
devtools score/lives commands
tests and setup helpers
```

On fatal player damage, the game-owned combat flow decrements lives only when the session can lose lives:

```text
player takes fatal damage
-> applyFatalPlayerDamage
-> session.ShipDeaths++
-> if LifeOptions.CanLoseLives() and Lives > 0
   -> addPlayerLivesLocked(playerID, -1)
-> if Lives > 0
   -> set respawn cooldown
-> record ship_death event
```

Infinite lives is not a counter mutation. It is a session life option:

```text
runtime.LifeOptions.InfiniteLives
```

When infinite lives is enabled, the player can still die and despawn, but the lives counter is not decremented.

Pickup effects can add lives through:

```text
applyPickupEffectIntentLocked
-> addPlayerLivesLocked
-> pickup_effect_applied event with lives_after
```

The current `1_up` pickup resolves to `add_lives` amount `1`.

## Protocols and APIs

Player counters are projected through gameplay state packets.

The current packet-facing surfaces are:

```text
StatePacket.lives
StatePacket.player_sessions[player_id].score
StatePacket.player_sessions[player_id].lives
StatePacket.events[].lives
StatePacket.events[].lives_after
RoomPlayerMatchSummary.score
```

`StatePacket.lives` is the requesting player’s lives value.

`StatePacket.player_sessions` carries durable per-player session read-model state for all known player sessions, including score and lives.

`StatePacket.players` is active live ship state only. It should not be used as the source of truth for score or lives.

Death events use `events[].lives` to report the player’s remaining lives after the death consequence is applied.

Pickup effect events use `events[].lives_after` to report the updated lives value after an `add_lives` effect succeeds.

Room match results use `PlayerMatchFacts`, which reads score and ship deaths from player sessions. Room code later combines those game-owned facts with room membership identity when building match result summaries.

The counter seam is an internal Go API, not a client-callable protocol.

### Gameplay callers

Gameplay code calls locked helpers while already inside authoritative simulation mutation flow.

Examples:

```text
game.awardScore
game.applyFatalPlayerDamage
game.applyPickupEffectIntentLocked
```

These callers own why the counter should change.

The counter seam owns how the counter changes.

### Devtools callers

Devtools command routing lives outside the normal gameplay packet path.

Devtools calls narrow game-owned adapters:

```go
DevtoolsSetPlayerScore
DevtoolsAddPlayerScore
DevtoolsSetPlayerLives
DevtoolsAddPlayerLives
```

Those adapters delegate directly to the same public counter methods used by non-devtools setup and tests.

The devtools package resolves command targets and decides whether a debug command targets one player or all players. The game counter seam only mutates the requested player sessions.

### Match-result callers

Room match-result reporting does not mutate counters.

It reads game-owned match facts through:

```go
Game.PlayerMatchFacts()
```

Those facts include:

```text
game_player_id
score
ship_deaths
```

Room code then adds room/member identity data outside the game simulation boundary.

## Data ownership

Player counters mutate only in-memory game-server match state:

```text
game.playerSessions[player_id].Score
game.playerSessions[player_id].Lives
```

Counter mutation does not directly persist to player-data storage.

Persistent match reporting is downstream of room match-result summary construction and player-data integration. The player counter boundary only supplies the per-match score value that those systems consume later.

Counter-related constants come from generated server constants:

```text
constants.PlayerStartingLives
constants.BaseScore
```

Source files:

```text
shared/constants/server_entities.toml
shared/constants/server_constants.toml
```

Generated file:

```text
services/game-server/internal/constants/constants.go
```

Packet-facing state shapes are generated from shared packet source files.

Relevant generated file:

```text
services/game-server/internal/game/packets.go
```

Relevant packet source:

```text
shared/packets/gameplay.toml
```

Generated files should not be edited manually.

## Invariants

Player counters must preserve these rules:

* Score and lives are authoritative server state.
* Score and lives are stored on player sessions, not live ship/avatar entities.
* Counter values must not become negative.
* Set and add operations must use the same clamping behavior.
* Public counter methods must lock game state before mutation.
* Locked helper methods are only for callers already inside game-owned locked mutation flow.
* Missing player sessions must not create implicit sessions.
* Missing player sessions return `Found: false`.
* Scoring policy computes awards but does not mutate score.
* Game-owned score adapters apply scoring awards.
* Fatal player damage owns whether lives should decrement.
* Infinite lives prevents life loss but does not prevent death/despawn.
* Pickup `add_lives` effects mutate session lives, not ship health.
* Counter mutation does not directly persist player-data records.
* Client presentation must read counters from server state, not infer them locally.

## Code map

Primary implementation files:

```text
services/game-server/internal/game/player_counters.go
services/game-server/internal/game/session.go
services/game-server/internal/game/scoring.go
services/game-server/internal/game/asteroid_destruction.go
services/game-server/internal/game/combat.go
services/game-server/internal/game/pickup_effects.go
services/game-server/internal/game/player_session_state.go
services/game-server/internal/game/state_packet.go
services/game-server/internal/game/match.go
services/game-server/internal/game/match_facts.go
```

Pure policy and supporting packages:

```text
services/game-server/internal/game/scoring/scoring.go
services/game-server/internal/game/runtime/life_options.go
services/game-server/internal/game/pickups/collection.go
```

Devtools adapters and command handlers:

```text
services/game-server/internal/game/export_devtools_player_counters.go
services/game-server/internal/devtools/player_counters.go
```

Room and integration consumers:

```text
services/game-server/internal/rooms/room_match_summary.go
```

Generated and source-of-truth files:

```text
shared/constants/server_entities.toml
shared/constants/server_constants.toml
shared/packets/gameplay.toml
services/game-server/internal/constants/constants.go
services/game-server/internal/game/packets.go
```

Related tests:

```text
services/game-server/tests/game/player_counters_test.go
services/game-server/tests/game/collision_test.go
services/game-server/tests/game/devtools_test.go
services/game-server/internal/game/match_facts_test.go
services/game-server/internal/game/player_world_state_test.go
services/game-server/internal/game/pickup_effects_test.go
services/game-server/internal/rooms/room_match_summary_test.go
```

Important non-ownership boundaries:

```text
services/game-server/internal/game/scoring/
services/game-server/internal/game/damage/
services/game-server/internal/game/pickups/
services/game-server/internal/game/weapons/
services/game-server/internal/devtools/
services/game-server/internal/rooms/
services/game-server/internal/networking/
services/player-data/
client/
tools/data_sync/
```

`scoring` owns pure award policy, not score mutation.

`damage` owns damage result calculation, not lives mutation.

`pickups` owns pickup collection/effect intent classification, not direct session mutation.

`weapons` owns weapon state and firing policy, not player score/lives counters.

`devtools` owns command routing and target selection, not authoritative counter storage.

`rooms` owns match summary and player-data handoff, not simulation counter mutation.

`networking` owns transport and packet routing, not counter authority.

`client` owns presentation only.

`tools/data_sync` owns generation from shared constants and packet schemas.

## Tests and verification

Focused counter tests cover:

* setting score to an exact value
* clamping negative score to zero
* adding positive score
* adding negative score
* clamping score below zero
* setting lives to an exact value
* clamping negative lives to zero
* adding positive lives
* adding negative lives
* clamping lives below zero
* projecting counter seam updates through state packets

Collision tests cover score and lives integration behavior:

* destroyed asteroids award score
* non-destroying asteroid hits do not award score
* score is based on asteroid size
* paused players do not receive score from bullet/asteroid collision
* invulnerable players do not receive score from bullet/asteroid collision
* fatal player collision reduces lives
* nonfatal player collision does not reduce lives
* frozen collisions do not award score or reduce lives

Devtools tests cover:

* debug set score for all players
* debug add score for all players
* debug set lives for all players
* debug add lives for all players
* direct counter seam updates appearing in state packets
* infinite lives allowing death without reducing lives

Match-result tests cover:

* `PlayerMatchFacts` returning score and ship death counters
* match facts not carrying account or local-profile identity fields
* room match summary adding identity data outside the game simulation boundary

Useful verification command:

```bash
cd services/game-server
go test -buildvcs=false ./...
```

Focused verification for player counters:

```bash
cd services/game-server
go test -buildvcs=false ./tests/game -run 'PlayerCounter|SetPlayerScore|SetPlayerLives|Debug.*Score|Debug.*Lives'
```

## Related docs

* [Game Server Simulation Players](./!README.md)
* [Game Server Simulation](../!README.md)
* [Game Server](../../!README.md)
* [Game Server Simulation Scoring](../scoring/!README.md)
* [Collision To Damage Flow](../combat/collision-to-damage-flow.md)
* [Damage Resolution](../combat/damage-resolution.md)
* [Pickup Collection](../pickups/pickup-collection.md)
* [Pickup Effects](../pickups/pickup-effects.md)
* [Match Result Reporting](../../integrations/match-result-reporting.md)
* [Room Match Lifecycle](../../rooms/room-match-lifecycle.md)
* [Player Session State](player-session-state.md)
* [Player Death And Despawn](player-death-and-despawn.md)
* [Player Respawn](player-respawn.md)
* [Player Pause And Suspension](player-pause-and-suspension.md)
* [Gameplay packets](../../../../protocol/stubs/gameplay-packets.md) - Stub: gameplay realtime packet documentation.
* [Devtools packets](../../../../protocol/stubs/devtools-packets.md) - Stub: devtools packet documentation.
* [Data](../../../../data/!README.md)
* [Devtools](../../../../devtools/!README.md)

## Notes

The current counter seam has a lower bound of zero and no upper bound.

Score is currently awarded only through the implemented asteroid-destruction scoring path.

`StatePacket.lives` is a convenience projection for the packet recipient’s lives. Multi-player counter readback should use `StatePacket.player_sessions`.

`PlayerCounterChange` is intentionally data-only. Event emission, logging, packet projection, persistence, and match summary construction stay with the caller or downstream boundary that owns that context.
