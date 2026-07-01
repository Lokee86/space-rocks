# Player Death And Despawn

Parent index: [Game Server Simulation Players](./!INDEX.md)

## Purpose

This document describes the current game-server player death and despawn boundary.

It explains how fatal player damage becomes a pending-despawn ship, how lives and ship-death counters change, how respawn cooldown is prepared, how presentation events are emitted, and how the active player avatar is later removed from the runtime entity store.

## Overview

Player death and despawn are authoritative game-server simulation behavior.

The current death flow is:

```text
fatal player damage result
-> applyFatalPlayerDamage
-> capture camera view at death position
-> mark active ship pending despawn
-> increment session ship death count
-> decrement lives when life options allow
-> set respawn cooldown when lives remain
-> log death or game over
-> record ship_death presentation event
-> later remove ready pending-despawn ship
```

The death flow affects two different player-owned states:

```text
active runtime ship
= the current avatar in game.entities.Players

player session
= durable in-match player state such as lives, score, respawn cooldown, selected ship, weapons, damage options, life options, and ship death count
```

Death does not remove the player session. It removes only the active avatar after a short despawn delay. The session remains so the player can respawn when eligible, remain visible to match/lifecycle projection, or be counted as eliminated when no lives remain.

Despawn is delayed through `constants.CollisionDespawnDelay`, currently `0.05`. Respawn is delayed through `constants.PlayerRespawnDelay`, currently `3.0`, when the session still has lives remaining.

## Code root

```text
services/game-server/internal/game/
```

Primary supporting packages:

```text
services/game-server/internal/game/runtime/
services/game-server/internal/game/damage/
services/game-server/internal/game/events/
services/game-server/internal/game/effects/radial/
services/game-server/internal/game/rules/
services/game-server/internal/constants/
```

## Responsibilities

Player death and despawn own the game-server side of:

* Converting fatal player damage into authoritative death consequences.
* Capturing the dead ship position into the player camera view.
* Marking the active player ship pending despawn.
* Clearing dead ship velocity and input through pending-despawn behavior.
* Incrementing the session `ShipDeaths` counter.
* Decrementing session lives when finite lives are enabled.
* Preserving lives when infinite-lives mode is enabled.
* Setting respawn cooldown when the player still has lives.
* Distinguishing death-with-lives from game-over-without-lives in logs.
* Recording `ship_death` events for state-packet presentation.
* Removing ready pending-despawn player ships from `game.entities.Players`.
* Keeping player session state alive after active avatar removal.
* Feeding lifecycle projection with active, pending-respawn, or eliminated status.

## Does not own

Player death and despawn do not own:

* Pure damage math.
* Collision detection.
* Radial effect timing, coverage, or hit selection.
* Player respawn placement.
* Respawn request handling.
* Room membership.
* Match result persistence.
* Client death animation, audio, overlay, or respawn UI.
* WebSocket transport.
* Packet codec generation.
* Durable account/profile statistics persistence.
* Disconnect cleanup.

`RemovePlayer` is not death/despawn behavior. It deletes the active ship, camera view, player session, targets, and pending event queue for a leaving/disconnected player. Death keeps the session and only removes the active avatar after the despawn delay.

## Domain roles

Player death and despawn participate in the player lifecycle and combat domains by enforcing server authority over:

* fatal player damage consequences
* lives decrement
* ship death count
* respawn cooldown setup
* active-avatar removal
* lifecycle status projection
* match-over input facts
* death presentation events

The client observes death through world/session lane readback and `event_batch`. It does not decide that a player died, that a life was lost, that a respawn cooldown started, or that the player was eliminated.

## Death entry points

Current code routes fatal player damage into `applyFatalPlayerDamage` from these paths:

```text
ship/asteroid collision fatal result
radial hit fatal result
devtools kill fatal result
```

Ship/asteroid collision enters through `handleShipAsteroidCollisions`.

That path:

```text
detect player/asteroid collision
-> build playerAsteroidDamageRequest
-> damage.ResolveSingle
-> applyDamageResultToPlayer
-> record damage_applied event when useful
-> if result is fatal player damage, call applyFatalPlayerDamage
```

Radial effects enter through `applyRadialHitToPlayer`.

That path:

```text
radial hit selects player target
-> build radial player damage request
-> damage.ResolveSingle
-> applyDamageResultToPlayer
-> record damage_applied event when useful
-> if result is fatal, call applyFatalPlayerDamage
```

Devtools kill enters through `DevtoolsKillPlayer`.

That path builds a debug damage request against the target player, applies the result to health and shields, and calls `applyFatalPlayerDamage` when the result is fatal.

## Collision death eligibility

Ship/asteroid collision death first passes through collision-damage eligibility.

A player can take asteroid collision damage only when:

```text
the player ship is not pending despawn
the player session exists
the player session is not suspended
the player is not temporarily invulnerable
the player damage options allow damage
```

Suspension includes pause and dev freeze.

Damage options include debug invincibility.

This means paused players, dev-frozen players, temporarily invulnerable players, and debug-invincible players do not die from asteroid collision damage.

This eligibility check is specific to ship/asteroid collision damage. Radial damage and devtools kill use their own entry paths.

## Fatal player damage sequence

`applyFatalPlayerDamage` owns the shared death consequence sequence.

The function first captures the current ship position:

```text
position := player.Position()
```

It then updates or creates the player's camera view at that position. This keeps visibility and state projection anchored after the active ship stops being valid as a live avatar.

The player ship is then marked pending despawn:

```text
player.MarkPendingDespawn(constants.CollisionDespawnDelay)
```

`MarkPendingDespawn` sets:

```text
PendingDespawn = true
DespawnDelay = delay
Velocity = zero vector
Input = empty input state
```

After this point, the ship is no longer considered an active live player avatar.

## Session consequences

When a player session exists, fatal player damage mutates session state:

```text
session.ShipDeaths++
if session.LifeOptions.CanLoseLives() && session.Lives > 0:
    decrement lives by 1
if session.Lives > 0:
    session.RespawnCooldown = PlayerRespawnDelay
```

Lives are decremented through the player counter seam:

```text
game.addPlayerLivesLocked(playerID, -1)
```

The counter seam clamps player counters and keeps mutation centralized.

If infinite lives are enabled, `LifeOptions.CanLoseLives()` returns false. The death still increments `ShipDeaths`, marks the ship pending despawn, records a death event, and applies the respawn delay because lives remain.

## Death versus game over

The death path treats remaining lives as the local distinction between a normal death and a player game-over state.

If lives remain:

```text
log "player died"
record remaining lives
record respawn delay
```

If no lives remain:

```text
log "player game over"
record score
record death position
```

The `ship_death` event is recorded in both cases. The event includes the remaining lives and respawn delay, so consumers can distinguish death-with-respawn from death-with-elimination.

Match-over evaluation is separate. It is handled through `game/rules` using session and active-ship facts.

## Despawn state

Pending-despawn ships stay in `game.entities.Players` until their despawn delay expires.

While pending despawn:

* movement stepping only reduces `DespawnDelay`
* velocity and input are already cleared
* input is rejected
* movement is rejected
* shooting is rejected
* collision damage is rejected
* pickup collection ignores the player
* targeting and radial candidate collection skip the player
* match/lifecycle projection does not count the ship as active

The active avatar is removed by `removeReadyPlayers`.

That path scans `game.entities.Players` and deletes any ship whose `ReadyForRemoval()` returns true:

```text
PendingDespawn == true
DespawnDelay <= 0
```

This removes the active runtime ship only. It does not remove the session, camera view, pending event queue, or room membership.

## Simulation phase position

Player death can be produced during the collision and radial-effect phases.

The active-match simulation order is:

```text
step player sessions
-> step player weapons
-> step players
-> remove ready players
-> step asteroid spawning
-> step asteroids
-> step bullets
-> step pickups
-> step collisions
-> step radial effects
-> simulation step observers
```

`removeReadyPlayers` runs before the current tick's collision and radial-effect phases. A player killed later in the same tick remains in the entity store until a later active-match step removes ready pending-despawn players.

When the match is already over at the start of `Game.Step`, the simulation takes the match-over branch. That branch steps cleanup-safe non-player areas and does not run player stepping, collision checks, or `removeReadyPlayers`.

## Camera view after death

Fatal player damage preserves the dead ship position in `game.cameraViews`.

If a camera view already exists, its `X` and `Y` are updated to the death position.

If no camera view exists, a new `runtime.CameraView` is created with:

```text
X      = death position x
Y      = death position y
Config = player.Config
```

This allows state projection and visibility to continue using a stable player view after the active ship becomes pending despawn or is removed.

## Lifecycle projection

Player lifecycle is projected from session and active-ship facts, not from client-side inference.

A player is active only when the player has a non-pending active ship:

```text
ship exists
ship is not nil
ship is not pending despawn
```

Lifecycle status resolves to:

```text
active
pending_respawn
eliminated
```

The rules are:

```text
active           = player has a non-pending active ship
pending_respawn  = player has no active ship but has remaining lives
eliminated       = player has no active ship and no remaining lives
```

A pending-despawn ship is not treated as active for match decision or player world-state purposes, even though it may still exist in `game.entities.Players` until removal.

## Event output

Fatal player damage records an `events.EventShipDeath` domain event.

The event includes:

```text
type = ship_death
player_id
remaining lives
respawn delay
death x
death y
```

`eventStateForDomainEvent` projects that domain event into packet event state:

```text
Type         = PacketTypeShipDeath
PlayerID     = event.PlayerID
Lives        = event.Lives
RespawnDelay = event.RespawnDelay
X            = event.X
Y            = event.Y
```

`recordDomainEvent` broadcasts the packet event into every player session's pending presentation event queue.

The event is delivered through the next `event_batch` write for each player. Successful lane packet delivery clears that player's pending presentation events after the packet is written.

A fatal damage path may also emit a `damage_applied` event before the `ship_death` event when damage actually changed health or shield.

## Protocols and APIs

Player death and despawn have no direct HTTP API.

Clients observe death/despawn indirectly through normal game-server lane output:

```text
world lane ship records
session lane player records
session lane lifecycle records
event_batch
overlay lane receiver-local lives/readout
```

Relevant packet facts:

```text
world lane ship records
= current runtime ship states, including ships still present during pending-despawn delay

session lane player records
= session read model with lives, score, respawn cooldown, spawn position, and weapon ids

session lane lifecycle records
= active, pending_respawn, or eliminated status

event_batch
= queued presentation events such as damage_applied and ship_death

overlay lane receiver-local lives/readout
= local viewer's current lives
```

The realtime packet contract is generated. The implementation should not hand-edit generated packet definitions.

## Data ownership

Player death and despawn read:

```text
game.entities.Players
game.playerSessions
game.cameraViews
runtime ship position
runtime ship health and shields
runtime ship damage options
runtime ship pending-despawn state
player session lives
player session life options
player session score
player session suspension state
damage results
```

Player death and despawn mutate:

```text
runtime ship pending-despawn state
runtime ship despawn delay
runtime ship velocity
runtime ship input
player session ship death count
player session lives
player session respawn cooldown
camera view position
pending presentation event queues
```

Player death and despawn delete:

```text
game.entities.Players[playerID]
```

only when the pending-despawn ship is ready for removal.

They do not persist external profile data or durable account stats.

## Invariants

Player death and despawn must preserve these rules:

* The server owns player death decisions.
* Fatal player damage consequences route through `applyFatalPlayerDamage`.
* The damage package does not mutate lives, respawn cooldown, camera views, sessions, or events.
* A dead player session remains after active avatar despawn.
* Death removes the active avatar, not the player session.
* `RemovePlayer` is disconnect/leave cleanup, not normal death flow.
* Pending-despawn players are not active for lifecycle, targeting, collision damage, input, movement, shooting, or pickup collection.
* Lives decrement only when life options allow finite lives to be lost.
* Ship death count increments on fatal player damage.
* Respawn cooldown is set only when lives remain after death.
* The camera view is preserved at the death position.
* `ship_death` events are server-authored presentation facts.
* Clients must not infer lifecycle only from presence or absence in `world lane ship records`.
* Match-over status is evaluated through `game/rules`, not by death/despawn code directly.

## Code map

Primary implementation files:

```text
services/game-server/internal/game/combat.go
services/game-server/internal/game/simulation_players.go
services/game-server/internal/game/session.go
services/game-server/internal/game/player_counters.go
services/game-server/internal/game/player_world_state.go
services/game-server/internal/game/player_session_state.go
services/game-server/internal/protocol/realtime/records.go
services/game-server/internal/game/events.go
services/game-server/internal/game/match.go
services/game-server/internal/game/simulation.go
```

Supporting runtime files:

```text
services/game-server/internal/game/runtime/ship.go
services/game-server/internal/game/runtime/state.go
services/game-server/internal/game/runtime/life_options.go
services/game-server/internal/game/runtime/damage_options.go
services/game-server/internal/game/runtime/suspension.go
services/game-server/internal/game/runtime/camera.go
```

Supporting damage and event files:

```text
services/game-server/internal/game/damage/result.go
services/game-server/internal/game/damage/resolve.go
services/game-server/internal/game/combat_damage_requests.go
services/game-server/internal/game/combat_damage_application.go
services/game-server/internal/game/damage_events.go
services/game-server/internal/game/events/events.go
```

Radial and devtools entry points:

```text
services/game-server/internal/game/simulation_radial_effects.go
services/game-server/internal/game/radial_damage_requests.go
services/game-server/internal/game/radial_candidates.go
services/game-server/internal/game/export_devtools_toggles.go
services/game-server/internal/game/export_devtools_respawn.go
```

Match/lifecycle support:

```text
services/game-server/internal/game/rules/match.go
services/game-server/internal/game/player/state.go
```

Generated packet contract:

```text
services/game-server/internal/game/packets.go
```

Important non-ownership boundaries:

```text
services/game-server/internal/game/damage/
services/game-server/internal/game/effects/radial/
services/game-server/internal/game/physics/
services/game-server/internal/networking/
client/
```

`damage` owns result calculation, not session mutation.

`effects/radial` owns radial timing and hit selection, not player death consequences.

`physics` owns primitive collision math, not fatal consequence application.

`networking` owns transport, not lifecycle decisions.

`client` owns presentation only.

## Tests and verification

Relevant game integration tests:

```text
services/game-server/tests/game/collision_test.go
services/game-server/tests/game/respawn_test.go
services/game-server/tests/game/devtools_test.go
services/game-server/internal/game/simulation_match_over_test.go
services/game-server/internal/game/player_world_state_test.go
services/game-server/internal/game/events_test.go
```

Current coverage includes:

* ship/asteroid collision marking the player pending despawn
* delayed removal after despawn delay
* `ship_death` event broadcast to all player views
* `damage_applied` event emission before death
* nonfatal ship collision reducing health without death
* lives not changing on nonfatal damage
* asteroid collision leaving the asteroid active after player death
* cross-boundary ship/asteroid collision death
* paused player skipping asteroid collision damage
* invulnerable player skipping asteroid collision damage
* death after invulnerability expires
* debug kill marking pending despawn
* debug kill reducing lives
* debug kill broadcasting death events
* frozen world/collision devtools gates blocking collision-produced ship death
* player world-state projection for active, pending-respawn, and eliminated states
* match-over stepping avoiding spawning and unsafe cleanup work

Useful verification command:

```bash
cd services/game-server
go test -buildvcs=false ./...
```

Focused verification for player death/despawn behavior:

```bash
cd services/game-server
go test -buildvcs=false ./tests/game -run 'ShipAsteroid|Debug.*Kill|Respawn'
```

Focused verification for lifecycle and event projection:

```bash
cd services/game-server
go test -buildvcs=false ./internal/game -run 'PlayerWorldState|Event|MatchOver'
```

## Related docs

* [Game Server Simulation Players](./!INDEX.md)
* [Game Server Simulation](../!INDEX.md)
* [Game Server](../../!INDEX.md)
* [Collision To Damage Flow](../combat/collision-to-damage-flow.md)
* [Damage Resolution](../combat/damage-resolution.md)
* [Radial Effects](../combat/radial-effects.md)
* [Player Session State](player-session-state.md)
* [Active Player Avatar State](active-player-avatar-state.md)
* [Player Camera View State](player-camera-view-state.md)
* [Player Counters](player-counters.md)
* [Player Respawn](player-respawn.md)
* [Player Pause And Suspension](player-pause-and-suspension.md)
* [Player Lifecycle Overview](player-lifecycle-overview.md)
* [Lane Packet Projection](../runtime/lane-packet-projection.md)
* [Presentation Event Queue](../runtime/presentation-event-queue.md)

## Notes

The legacy docs correctly identified the key boundary: the Go game server owns lives, death, respawn gating, death events, and match facts; the client owns presentation only.

The current implementation uses both `player_lifecycle` and `player_sessions` because active ship presence alone is not enough to represent pending-respawn or eliminated players.

A pending-despawn ship can still appear in `world lane ship records` while it remains in the runtime entity store during the despawn delay. Consumers should use lifecycle/session facts for player status instead of treating ship presence as the only lifecycle source.

