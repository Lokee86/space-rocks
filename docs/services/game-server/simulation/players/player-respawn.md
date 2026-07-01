# Player Respawn

Parent index: [Game Server Simulation Players](./!INDEX.md)

## Purpose

This document describes the game-server player respawn implementation.

It covers authoritative respawn request handling, respawn eligibility, cooldown progression, safe respawn placement, live ship recreation, and camera-view reattachment.

## Overview

Player respawn is owned by the game server simulation. The client may request a respawn, but the server decides whether the player is eligible, where the new ship appears, and what session state is preserved.

Respawn operates on the player session, not on the previous ship entity. When a player dies, fatal player damage removes the active ship through pending-despawn cleanup, updates the session's lives and respawn cooldown, and emits a `ship_death` event with the remaining lives and respawn delay. After the cooldown reaches zero, a `respawn` client packet can recreate the player's active `runtime.Ship` from the stored session state.

The recreated ship keeps session-owned state:

* player ID
* ship type
* resolved ship stats
* client config
* damage options
* selected primary and secondary weapons
* stored targeting state

Respawn does not create a new player session. It only restores an active ship for an existing session that still has lives remaining.

## Code root

`services/game-server/internal/game/`

## Responsibilities

Player respawn owns:

* receiving authoritative respawn requests after networking has routed them to `Game.HandlePacket`
* checking that the player session exists
* checking that the player has remaining lives
* checking that the respawn cooldown has reached zero
* blocking respawn when an active ship already exists for the player
* choosing a safe respawn position near the session spawn position
* recreating the active `runtime.Ship` from session state
* restoring the ship's weapons and targeting state from the session
* inserting the recreated ship into `game.entities.Players`
* reattaching or refreshing the player's server-side camera view
* exposing respawn cooldown through player session state
* contributing to player lifecycle classification as `pending_respawn` while the player has lives but no active ship

## Does not own

Player respawn does not own:

* WebSocket transport or raw packet decoding.
* Room membership or player identity assignment.
* Client-side HUD prompts, countdown rendering, or input binding.
* Player death collision detection.
* Damage resolution.
* Score mutation.
* Extra-life pickup effects.
* Match result persistence.
* Player-data profile persistence.
* Devtools command UI or debug selector behavior.
* Client-side camera rendering.

## Domain roles

Respawn participates in the player lifecycle flow inside a live match.

The authoritative lifecycle states are derived from session and active-ship state:

```text
active
= player session exists
+ non-pending active ship exists

pending_respawn
= player session exists
+ no active ship
+ lives > 0

eliminated
= player session exists
+ no active ship
+ lives <= 0
```

Respawn is the transition from `pending_respawn` back to `active`.

The server keeps pending-respawn players in the session map even when their active ship is absent from `game.entities.Players`. This lets world/session lane readback continue reporting lives, respawn cooldown, spawn position, weapons, and lifecycle status without treating the player as removed from the match.

## Protocols and APIs

The inbound respawn surface is the realtime `respawn` client packet.

```text
client respawn input
-> networking inbound gameplay adapter
-> room.GameInstance().HandlePacket(playerID, packet)
-> Game.respawnPlayer(playerID)
```

The packet carries no respawn authority beyond its packet type. The server uses the networking session's current game player ID as the player identity and ignores client-selected position, lives, cooldown, or ship state.

`Game.HandlePacket` handles `respawn` before active-ship lookup. This is intentional: a valid respawn request normally arrives while the player has no active ship. If the handler required an active ship before recognizing respawn, dead players could not request respawn.

Successful respawn is reflected through normal lane projection rather than a dedicated respawn event:

* `world lane ship records` include the recreated active ship.
* `session lane player records[playerID].respawn_cooldown` is zero.
* `session lane lifecycle records[playerID]` becomes `active`.

Death and respawn availability are surfaced separately:

* `ship_death` events include `lives` and `respawn_delay` and are delivered through `event_batch`.
* `session lane player records` include `respawn_cooldown`.
* `session lane lifecycle records` report `pending_respawn` while the player has lives but no active ship.

## Respawn flow

Fatal player damage sets up the respawn path:

```text
fatal player damage
-> preserve camera view at death position
-> mark active ship pending despawn
-> increment ship death count
-> decrement lives when life options allow it
-> set respawn cooldown if lives remain
-> emit ship_death event with lives and respawn_delay
```

The simulation step advances respawn cooldowns through `stepPlayerSessions(delta)`. Each `playerSession.Step(delta)` reduces `RespawnCooldown` toward zero.

A normal respawn request then follows this flow:

```text
respawn packet
-> Game.HandlePacket
-> respawnPlayer(playerID)
-> find player session
-> require session.CanRespawn()
-> require no active ship for playerID
-> plan safe respawn position
-> create session.NewShip(position)
-> store ship in game.entities.Players[playerID]
-> setPlayerCameraViewLocked(playerID, player)
```

`session.CanRespawn()` is true only when:

```text
session.Lives > 0
session.RespawnCooldown == 0
```

The active-ship guard blocks duplicate respawns. If `game.entities.Players[playerID]` already contains a ship, the request is ignored.

## Cooldown behavior

Respawn cooldown is session-owned state.

The configured values are generated constants:

```text
PlayerStartingLives = 3
PlayerRespawnDelay = 3.0
PlayerRespawnBuffer = 160.0
```

When a fatal player hit leaves lives remaining, the session's `RespawnCooldown` is set to `PlayerRespawnDelay`.

Cooldown decreases during simulation stepping:

```text
if session.RespawnCooldown > 0:
    session.RespawnCooldown = max(0, session.RespawnCooldown - delta)
```

A player can request respawn only after the cooldown reaches zero. Requests before that point are blocked and leave the player without an active ship.

Final death does not create a usable respawn cooldown. When lives reach zero, `CanRespawn()` remains false and later respawn requests are ignored.

## Safe respawn placement

Respawn starts from the session's stored spawn position:

```text
session.SpawnPosition
```

The server first tests that position. If it is unsafe, the server searches outward in square rings using this spacing:

```text
max(64, PlayerRespawnBuffer)
```

Each candidate is checked against:

* non-pending asteroids
* other non-pending active players
* the respawning player's resolved ship collision shape

The respawning player's own player ID is ignored during player clearance checks so stale or replacement state for the same player does not block itself.

A candidate is unsafe when the wrapped world distance to an asteroid or another active player is less than or equal to:

```text
respawn ship radius + blocker radius + PlayerRespawnBuffer
```

The distance check uses toroidal world distance, so hazards near one world edge can block respawn candidates near the opposite edge.

Collision-shape radius is approximated by shape type:

```text
circle    -> radius
capsule   -> half height
rectangle -> half diagonal
polygon   -> farthest point from origin
```

If the server cannot resolve the player's configured ship collision shape, the current implementation treats the candidate as safe. Tests cover this fallback.

## Ship recreation

Respawn recreates a ship by calling `session.NewShip(spawnPosition)`.

The new ship is initialized from session-owned values:

```text
ID            = session.ID
ShipTypeID    = session.ShipTypeID
Stats         = session.Stats
Config        = session.Config
Health        = session.Stats.MaxHealth
DamageOptions = session.DamageOptions
Primary       = session.PlayerArmory.Primary
Secondary     = session.PlayerArmory.Secondary
Targeting     = session.Targeting
```

Respawn does not reset score, lives, ship death count, selected weapons, targeting, ship type, or client config.

The session remains the durable runtime state for the player. The ship is the active avatar state.

## Camera-view behavior

The server keeps a camera view for each player so state projection can continue to provide a useful view anchor while the active ship is absent.

On death, fatal player damage stores the death position in the player's camera view. If no camera view exists, one is created using the dying ship's config.

On respawn, `setPlayerCameraViewLocked(playerID, player)` reattaches the camera view to the recreated ship position.

Camera config is preserved in this order:

```text
existing valid camera view config
-> valid session config
-> valid player config
-> world-size fallback
```

This prevents respawn from resetting a valid client viewport config when the ship is recreated.

## Data ownership

Respawn mutates only game-server runtime state:

* `game.playerSessions[playerID].RespawnCooldown`
* `game.entities.Players[playerID]`
* `game.cameraViews[playerID]`

Respawn reads session-owned state:

* `Lives`
* `SpawnPosition`
* `ShipTypeID`
* `Stats`
* `Config`
* `DamageOptions`
* `PlayerArmory`
* `Targeting`

Respawn does not persist player state outside the game server. Player-data and API-server persistence are outside this boundary.

The constants used by respawn are generated into `services/game-server/internal/constants/constants.go` from shared constants source files. Packet type and lane packet shapes are generated into `services/game-server/internal/game/packets.go` from the shared realtime packet source.

## Code map

Primary implementation files:

* `services/game-server/internal/game/session.go`

  * Defines `playerSession`.
  * Owns `RespawnCooldown`.
  * Owns `CanRespawn()`.
  * Owns `NewShip()`.
  * Owns `respawnPlayer()`.
  * Owns safe respawn planning and clearance checks.

* `services/game-server/internal/game/input.go`

  * Routes `PacketTypeRespawn` to `game.respawnPlayer(playerID)` before requiring an active ship.

* `services/game-server/internal/game/combat.go`

  * Sets respawn cooldown after fatal player damage when lives remain.
  * Records `ship_death` events with `RespawnDelay`.
  * Preserves camera view at death position.

* `services/game-server/internal/game/simulation_players.go`

  * Advances player session cooldown timers.
  * Removes ready-for-removal dead player ships from active player entities.

* `services/game-server/internal/game/players.go`

  * Creates initial player sessions and ships.
  * Owns `setPlayerCameraViewLocked()` used by respawn.
  * Removes player sessions and camera views when a player fully leaves.

* `services/game-server/internal/game/player_world_state.go`

  * Projects active, pending-respawn, and eliminated player state from sessions, ships, and camera views.

* `services/game-server/internal/game/player/state.go`

  * Defines player world-state status values and targetability/damageability/collidability flags.

* `services/game-server/internal/game/match.go`

* `services/game-server/internal/game/rules/match.go`

  * Classify players as active, pending respawn, or eliminated for match lifecycle decisions and lane packets.

Related generated and source files:

* `shared/constants/server_entities.toml`

  * Source values for player lifecycle and respawn constants.

* `services/game-server/internal/constants/constants.go`

  * Generated server constants consumed by respawn.

* `shared/packets/gameplay.toml`

  * Source packet definitions for `respawn`, `PlayerSessionState`, `EventState`, and `lane packet`.

* `services/game-server/internal/game/packets.go`

  * Generated packet constants and packet/state structs.

* `services/game-server/internal/game/spawn_types.go`

  * Defines `PlayerSpawnPlan` and `SpawnReasonPlayerRespawn`.

* `services/game-server/internal/networking/inbound/gameplay.go`

  * Adapts inbound `respawn` packets from the current room/session to the game instance.

* `services/game-server/internal/game/export_devtools_respawn.go`

  * Exposes safe respawn position and force-respawn helpers for devtools.
  * Debug force respawn intentionally bypasses normal client request cooldown gates.

Related tests:

* `services/game-server/tests/game/respawn_test.go`

  * Covers death cooldown, blocked early respawn, no-lives blocking, safe placement, respawn buffer behavior, wrap-boundary hazards, and avoidance of existing players.

* `services/game-server/tests/game/ship_collision_shape_test.go`

  * Covers respawn fallback behavior for unknown session collision shape IDs.

* `services/game-server/internal/game/player/state_test.go`

  * Covers player world-state status and JSON field names for respawn cooldown.

* `services/game-server/internal/game/player_world_state_test.go`

  * Covers pending respawn state without an active ship.

* `services/game-server/tests/game/rules_match_test.go`

* `services/game-server/tests/game/match_decision_test.go`

  * Cover pending-respawn participation in match decisions.

* `services/game-server/internal/game/targeting_test.go`

  * Covers stored targeting state surviving dead-state targeting changes and being mirrored onto the respawned ship.

* `services/game-server/internal/game/export_devtools_respawn_test.go`

  * Covers devtools force-respawn camera-view creation.

Important non-ownership boundaries:

* `networking/inbound` owns packet routing, not respawn rules.
* `combat` owns fatal damage entry points, not respawn request handling.
* `playerSession` owns durable player runtime state.
* `runtime.Ship` owns the active avatar only.
* `rooms` owns room lifecycle and game instance access.
* client gameplay/HUD code owns respawn presentation only.

## Tests

Useful verification from `services/game-server/`:

```text
go test -buildvcs=false ./internal/game/...
go test -buildvcs=false ./tests/game -run 'Respawn|PlayerWorldState|MatchDecision|ShipCollisionShape'
```

Useful behavior checks:

* fatal death decrements lives and sets respawn delay when lives remain
* final death leaves the player unable to respawn
* respawn before cooldown completion is blocked
* respawn after cooldown completion recreates the active ship
* respawn preserves session lives
* added lives persist through death and respawn
* respawn avoids unsafe asteroid positions
* respawn avoids unsafe active-player positions
* respawn safety detects hazards across world wrap boundaries
* pending-respawn players remain in lifecycle/session state without active ships
* targeting stored in the session is applied to the respawned ship
* camera view is restored or refreshed when the ship is recreated

## Related docs

* [Game Server Simulation Players](./!INDEX.md)
* [Game Server Simulation](../!INDEX.md)
* [Gameplay Network Adapter](../../networking/gameplay-network-adapter.md)
* [Collision To Damage Flow](../combat/collision-to-damage-flow.md)
* [Weapons And Projectile Fire](../combat/weapons-and-projectile-fire.md)
* [Logging And Diagnostics](../../observability/logging-and-diagnostics.md)
* [Player Death And Despawn](player-death-and-despawn.md)
* [Active Player Avatar State](active-player-avatar-state.md)
* [Player Camera View State](player-camera-view-state.md)
* [Player Input Routing](player-input-routing.md)
* [Player Counters](player-counters.md)
* [HUD And Gameplay UI](../../../client/hud-and-gameplay-ui.md)

## Notes

Initial player spawn and player respawn share the safe player spawn helper, but respawn uses the existing session spawn position and `SpawnReasonPlayerRespawn`.

The current respawn path has no dedicated successful-respawn event. Clients infer successful respawn from the next authoritative world/session lane update that contains the recreated active ship and lifecycle status `active`.

Debug force respawn is intentionally separate from normal player respawn. It is exposed for devtools and should not be treated as the client gameplay respawn path.
