# Player Pause And Suspension

Parent index: [Game Server Simulation Players](./!INDEX.md)

## Purpose

This document describes the game-server player pause and suspension boundary.

It explains how normal player pause, dev player freeze, and the shared suspension gate affect server-side input, movement, shooting, collision damage, and score eligibility.

## Overview

Player pause and suspension are game-server simulation responsibilities owned by `services/game-server/internal/game`.

The core model is:

```text
player session SuspensionState
= Paused || DevFrozen
```

`Paused` is the normal gameplay pause flag toggled by a client `pause_request`.

`DevFrozen` is the debug player-freeze flag set through server devtools adapters.

Both flags live on the player session, not on the live ship entity. The live ship entity is still affected by suspension through game-owned capability helpers such as:

```text
playerCanReceiveInput
playerCanMove
playerCanShoot
playerCanTakeCollisionDamage
playerCanReceiveScore
```

The current pause flow is:

```text
client pause_request
-> networking inbound gameplay routing
-> Game.HandlePacket
-> togglePlayerPaused
-> session.Suspension.Paused changes
-> active ship input is cleared
-> pause state response is enqueued
```

When a player is paused, the server rejects new movement/fire input, clears existing input, blocks movement, blocks shooting, blocks asteroid collision damage, and blocks score awards.

When a player resumes, the server clears input again and grants short post-resume invulnerability.

Dev freeze uses the same shared suspension checks, but it does not call normal pause or resume. A player can be both paused and dev-frozen at the same time, and both causes must be cleared before the player is active again.

## Code root

```text
services/game-server/internal/game/
```

Primary supporting packages:

```text
services/game-server/internal/game/runtime/
services/game-server/internal/game/motion/
services/game-server/internal/networking/
services/game-server/internal/devtools/
services/game-server/internal/protocol/packetcodec/
```

## Responsibilities

Player pause and suspension owns the game-server side of:

* Storing normal pause state on the player session.
* Storing dev player-freeze state on the player session.
* Combining pause and dev freeze into a single suspension predicate.
* Toggling normal pause from `pause_request`.
* Ignoring pause toggles for missing sessions, missing active ships, and pending-despawn ships.
* Clearing player input when pausing.
* Clearing player velocity when pausing.
* Clearing player input when resuming.
* Granting post-resume invulnerability.
* Emitting direct `player_pause_state` responses after pause requests.
* Blocking input while suspended.
* Blocking movement while suspended.
* Blocking primary and secondary weapon fire while suspended.
* Blocking asteroid collision damage while suspended.
* Blocking score awards while suspended.
* Keeping normal pause and dev freeze independent.
* Providing devtools adapters for setting and reporting player freeze.

## Does not own

Player pause and suspension does not own:

* Client pause menu presentation.
* Client input polling or hotkey mapping.
* WebSocket transport.
* Packet encoding/decoding internals.
* Room membership or room lifecycle.
* Respawn eligibility.
* Death and pending-despawn lifecycle.
* Player score or lives mutation policy.
* Weapon fire policy internals.
* World freeze, asteroid freeze, bullet freeze, spawning freeze, or collision freeze.
* Debug command parsing.
* Devtools UI state.
* Durable account or profile data.

Those systems may call into or observe pause/suspension state, but they own their own boundaries.

## Domain roles

Player pause and suspension participates in the player experience domain by enforcing a server-authoritative local-player paused state.

The client may request pause and may present paused UI, but the game server owns whether the player can keep affecting the match while paused.

The server enforces that a paused player:

```text
does not accept new movement/fire input
does not continue movement from stale input
does not fire weapons
does not take asteroid collision damage
does not receive score from projectile outcomes
```

The server also preserves the debug/development domain boundary by letting devtools freeze a player through a separate suspension cause without pretending that debug freeze is normal pause.

## Suspension state model

The shared suspension state lives in `runtime.SuspensionState`:

```go
type SuspensionState struct {
    Paused    bool
    DevFrozen bool
}
```

The suspension predicate is:

```go
func (state SuspensionState) IsSuspended() bool {
    return state.Paused || state.DevFrozen
}
```

Normal pause writes only `Paused`.

Dev player freeze writes only `DevFrozen`.

This means:

```text
Paused=false, DevFrozen=false -> active
Paused=true,  DevFrozen=false -> suspended
Paused=false, DevFrozen=true  -> suspended
Paused=true,  DevFrozen=true  -> suspended
```

Resume clears only `Paused`. It does not clear `DevFrozen`.

Unfreeze clears only `DevFrozen`. It does not clear `Paused`.

## Normal pause flow

Normal pause is toggled by a `pause_request` packet from the client.

`internal/networking/inbound.HandleGameplayPacket` routes pause requests to the current room game instance:

```text
pause_request
-> gameInstance.HandlePacket(currentGamePlayerID, packet)
-> session.EnqueuePlayerPauseState()
```

Inside `Game.HandlePacket`, pause requests require an active player entity. If the session has no active ship, normal packet handling returns before pause toggling.

`togglePlayerPaused` also guards against:

```text
missing player session
missing active player entity
pending-despawn player
```

When the player is valid, it flips only `session.Suspension.Paused`.

## Pause behavior

When `setPlayerPaused(playerID, true)` succeeds, the server:

```text
sets session.Suspension.Paused = true
clears active ship input
sets active ship velocity to zero
logs "player paused"
```

Clearing velocity prevents the ship from drifting after pause begins.

Clearing input prevents stale movement or fire state from surviving the transition.

Suspended ships are still present in the authoritative player map unless another lifecycle path removes them. Pause does not despawn the ship, decrement lives, alter score, clear targeting, or move the camera.

## Resume behavior

When `setPlayerPaused(playerID, false)` succeeds, the server:

```text
sets session.Suspension.Paused = false
clears active ship input
sets InvulnerabilityRemaining to PlayerResumeInvulnerabilitySeconds
logs "player resumed"
```

The current generated constant is:

```text
PlayerResumeInvulnerabilitySeconds = 1.5
```

The source value comes from:

```text
shared/constants/server_entities.toml
```

Resume invulnerability blocks asteroid collision damage, but it does not block shooting. A resumed player can fire while still temporarily invulnerable.

Resume is ignored when the player is pending despawn. That prevents a dead or dying ship from being returned to active behavior through a pause toggle.

## Dev freeze behavior

Dev player freeze is a separate suspension cause used by server devtools.

The game-owned adapter is:

```go
DevtoolsSetPlayerFrozen(playerID string, enabled bool) bool
```

It sets `session.Suspension.DevFrozen`.

When enabling dev freeze, it also clears input on the active ship if one exists.

Dev freeze does not call `setPlayerPaused`, does not grant resume invulnerability, and does not change `Paused`.

The shared suspension predicate means dev-frozen players are blocked by the same gameplay gates as paused players:

```text
input blocked
movement blocked
shooting blocked
asteroid collision damage blocked
score awards blocked
```

Devtools status reports `PlayerFrozen` from `session.Suspension.DevFrozen`, not from the aggregate `IsSuspended()` predicate.

## Input gate

Inbound `input` packets are accepted only when `playerCanReceiveInput` returns true.

The gate requires:

```text
player is not pending despawn
player session exists
session is not suspended
```

If the player is suspended, `Game.HandlePacket` ignores the input packet and does not call `player.SetInput`.

This prevents new movement or fire input from replacing the cleared input state while paused or dev-frozen.

## Movement gate

Player movement uses `motion.AdvanceShipWithMovePolicy`.

The game server passes:

```go
game.playerCanMove(player.ID, player)
```

The movement gate requires:

```text
player is not pending despawn
player session exists
session is not suspended
```

When `canMove` is false, `motion.StepShipWithMovePolicy` clears ship input and returns without applying rotation, thrust, damping, velocity movement, or invulnerability countdown.

The ship position is still normalized through world wrapping after the movement helper returns, but no movement is applied while suspended.

## Shooting gate

Primary and secondary fire are checked during `stepPlayers`.

The current flow is:

```text
if bullets can move and primary fire input is set and player can shoot:
    fire primary weapon

if bullets can move and secondary fire input is set and player can shoot:
    fire secondary weapon
```

`playerCanShoot` requires:

```text
player is not pending despawn
player session exists
session is not suspended
primary weapon cooldown is zero
```

Both primary and secondary fire currently use the same `playerCanShoot` helper, so suspension blocks both.

World bullet freeze is separate. Even an active, unsuspended player cannot spawn bullets while `worldSimulationOptions.BulletsCanMove()` is false.

## Collision damage gate

Asteroid/player collision damage is blocked by `playerCanTakeCollisionDamage`.

The gate requires:

```text
player is not pending despawn
player session exists
session is not suspended
player is not invulnerable
player damage options allow damage
```

This means asteroid collision damage is blocked for:

```text
paused players
dev-frozen players
post-resume invulnerable players
respawn-invulnerable players
debug-invincible players
pending-despawn players
players without sessions
```

This gate only controls whether asteroid collision damage applies to the player. It does not own collision detection, damage math, damage application, death, lives decrement, or respawn setup.

## Score eligibility gate

Score awards are applied through `game.awardScore`.

Before mutating score, the game server checks `playerCanReceiveScore`.

The gate requires:

```text
player session exists
session is not suspended
player is not invulnerable
```

This prevents paused, dev-frozen, and invulnerable players from receiving score awards while they are not fully active.

The scoring package remains pure. It computes awards from scoring events, but the game-owned adapter decides whether the player is eligible to receive the award.

## Protocols and APIs

Pause has one inbound gameplay packet and one direct outbound response packet.

The inbound request is:

```json
{
  "type": "pause_request"
}
```

The response is:

```json
{
  "type": "player_pause_state",
  "player_id": "<player-id>",
  "paused": true
}
```

The request is for the current websocket session's active game player. Clients do not choose a target player for normal pause.

The server authority behind the packet is the game instance for the current room. Networking resolves the current room and current game player ID, then delegates to `Game.HandlePacket`.

The data crossing the boundary is intentionally small:

```text
request: packet type only
response: player id and normal paused flag
```

The response reports only normal pause state. It does not report dev freeze, aggregate suspension, invulnerability, movement eligibility, or score eligibility.

Pause state is not currently projected into the normal lane-native gameplay readback. It is sent as a direct `player_pause_state` response after pause requests.

Dev player freeze uses devtools command/status packets, not the normal `pause_request` protocol.

## Data ownership

Pause and suspension are in-memory game-server runtime state.

The owning data is:

```text
playerSession.Suspension.Paused
playerSession.Suspension.DevFrozen
```

Pause and suspension read:

```text
game.playerSessions
game.entities.Players
runtime.Ship.PendingDespawn
runtime.Ship.Input
runtime.Ship.Velocity
runtime.Ship.InvulnerabilityRemaining
runtime.Ship.WeaponState
runtime.Ship.DamageOptions
```

Pause and suspension mutate:

```text
playerSession.Suspension.Paused
playerSession.Suspension.DevFrozen
runtime.Ship.Input
runtime.Ship.Velocity
runtime.Ship.InvulnerabilityRemaining
```

Pause and suspension do not persist profile/account data and do not write external storage.

## Invariants

Player pause and suspension must preserve these rules:

* The server is authoritative for whether a player is paused or suspended.
* Normal pause and dev freeze are independent suspension causes.
* Aggregate suspension is true when either pause or dev freeze is true.
* Resume must not clear dev freeze.
* Unfreeze must not clear normal pause.
* Pausing clears input.
* Pausing clears velocity.
* Resuming clears input.
* Resuming grants short invulnerability.
* Resuming does not allow pending-despawn players to become active again.
* Suspended players cannot receive new input.
* Suspended players cannot move.
* Suspended players cannot shoot.
* Suspended players cannot take asteroid collision damage.
* Suspended players cannot receive score awards.
* Pause does not own death, respawn, scoring, weapon policy, or world-freeze behavior.
* Normal pause targets only the current websocket session's active game player.
* Dev player freeze stays behind devtools adapters and does not become normal gameplay protocol.

## Code map

Primary implementation files:

```text
services/game-server/internal/game/pause.go
services/game-server/internal/game/runtime/suspension.go
services/game-server/internal/game/input.go
services/game-server/internal/game/simulation_players.go
services/game-server/internal/game/session.go
```

Movement and weapon gate participants:

```text
services/game-server/internal/game/motion/motion.go
services/game-server/internal/game/combat.go
services/game-server/internal/game/scoring.go
services/game-server/internal/game/weapons/
```

Networking and packet response files:

```text
services/game-server/internal/networking/inbound/gameplay.go
services/game-server/internal/networking/player_pause_state.go
services/game-server/internal/networking/inbound_adapter.go
services/game-server/internal/protocol/packetcodec/
```

Generated and source packet files:

```text
shared/packets/gameplay.toml
shared/packets/outputs.toml
services/game-server/internal/game/packets.go
client/scripts/generated/networking/packets/packets.gd
```

Generated and source constants:

```text
shared/constants/server_entities.toml
services/game-server/internal/constants/constants.go
```

Devtools participant files:

```text
services/game-server/internal/game/export_devtools_toggles.go
services/game-server/internal/game/export_devtools_status.go
services/game-server/internal/devtools/
shared/packets/debug.toml
```

Important non-ownership boundaries:

```text
services/game-server/internal/game/world_simulation_options.go
services/game-server/internal/game/rules/
services/game-server/internal/rooms/
services/game-server/internal/networking/
client/scripts/gameplay/input/
client/scripts/gameplay/menu/
client/scripts/devtools/
```

`world_simulation_options.go` owns world/asteroid/bullet/spawn/collision freeze, not per-player suspension.

`rules` owns match policy decisions, not pause state.

`rooms` owns room membership and match lifecycle, not per-player simulation gates.

`networking` owns transport and routing, not pause eligibility.

The client owns presentation and input collection, not authoritative pause effects.

## Tests and verification

Relevant tests:

```text
services/game-server/tests/game/pause_test.go
services/game-server/tests/game/devtools_test.go
services/game-server/tests/game/collision_test.go
```

Current pause-focused coverage includes:

* suspension state reflects pause and dev freeze.
* paused and frozen suspension requires both causes to be cleared.
* pause requests toggle pause state.
* pause state packets reflect pause request toggles.
* pause clears input and ignores new input.
* pause clears velocity before resume.
* fresh players accept input and move.
* fresh players can shoot.
* paused players do not move or shoot.
* second pause toggle resumes with invulnerability and allows shooting.
* pause toggle is ignored for dead/inactive players.

Current devtools-related coverage includes:

* debug status reports player freeze.
* debug freeze player all-player toggles affect eligible players.
* debug world freeze is separate from player freeze.
* debug invincibility and infinite lives remain separate from suspension.

Useful verification command:

```bash
cd services/game-server
go test -buildvcs=false ./...
```

Focused verification for pause behavior:

```bash
cd services/game-server
go test -buildvcs=false ./tests/game -run 'Pause|Suspension'
```

Focused verification for devtools freeze behavior:

```bash
cd services/game-server
go test -buildvcs=false ./tests/game -run 'Debug.*Freeze|PlayerFrozen'
```

## Related docs

* [Game Server Simulation Players](./!INDEX.md)
* [Game Server Simulation](../!INDEX.md)
* [Game Server](../../!INDEX.md)
* [Game Server Simulation Runtime](../runtime/!INDEX.md)
* [Game Server Simulation Combat](../combat/!INDEX.md)
* [Game Server Simulation Scoring](../scoring/!INDEX.md)
* [Player Input Routing](player-input-routing.md)
* [Player Session State](player-session-state.md)
* [Active Player Avatar State](active-player-avatar-state.md)
* [Player Death And Despawn](player-death-and-despawn.md)
* [Player Respawn](player-respawn.md)
* [Player Counters](player-counters.md)
* [Collision To Damage Flow](../combat/collision-to-damage-flow.md)
* [Weapons And Projectile Fire](../combat/weapons-and-projectile-fire.md)
* [Toroidal Space And Motion](../world/toroidal-space-and-motion.md)
* [Gameplay Network Adapter](../../networking/gameplay-network-adapter.md)
* [Protocol](../../../../protocol/!INDEX.md)
* [Devtools](../../../../devtools/!INDEX.md)
* [Data](../../../../data/!INDEX.md)

## Notes

`player_pause_state` reports only normal pause. A dev-frozen player can be suspended while `player_pause_state.paused` is false.

Suspended players still remain active player entities unless another lifecycle path removes them. Pause is not the same as death, pending despawn, respawn cooldown, or elimination.

The movement helper clears input while movement is blocked, but pausing also clears input immediately at the transition. This duplication is intentional defense against stale input from both the transition and later simulation ticks.

