## Player Session State

Parent index: [Game Server Simulation Players](./!INDEX.md)

## Purpose

This document describes the game-server simulation’s per-player session state.

Player session state is the server-owned record that persists player match state across active ship creation, death, despawn, respawn, pause gates, and lane-native realtime projection.

## Overview

The game server stores per-player simulation session state in `Game.playerSessions`, keyed by player ID. A `playerSession` is separate from the active ship entity in `game.entities.Players`.

That separation matters because the active ship can be pending despawn, removed, or recreated while the player’s match-local state remains available. The session stores the player’s identity, default ship/build data, camera/client config, targeting state, score, lives, death count, respawn cooldown, pause/suspension state, damage/life options, and default armory.

A new session is created when the simulation adds a player. The session is immediately projected into a `runtime.Ship` for the initial active avatar. Later, respawn uses the same session to create a new ship after cooldown and life checks pass.

Session state is also projected into session lane player records as `PlayerSessionState`. That lane projection lets clients see player-level state that is not guaranteed to exist on the active ship entity, such as remaining lives, score, respawn cooldown, spawn position, and session armory identifiers.

## Code root

`services/game-server/internal/game/`

## Responsibilities

* Own the `playerSession` records stored in `Game.playerSessions`.
* Create default player session state when a player joins the simulation.
* Preserve match-local player state separately from the active ship entity.
* Store the player ID, ship type, resolved ship stats, spawn position, client config, targeting state, score, lives, ship death count, respawn cooldown, suspension state, damage options, life options, and player armory.
* Project session state into a new `runtime.Ship` for initial spawn and respawn.
* Step per-session cooldown state during the simulation tick.
* Gate respawn eligibility with session lives and respawn cooldown.
* Provide session-backed counter mutation for score and lives.
* Provide session-backed pause/suspension checks for input, movement, shooting, collision damage, and score receipt.
* Project session state into session lane player records.

## Does not own

* Room membership, room identity, lobby ownership, or ready state.
* WebSocket transport behavior.
* Client-side player UI, HUD rendering, respawn overlays, or pause menu presentation.
* Durable account, profile, or progression persistence.
* Cross-match player data storage.
* Active ship physics, movement integration, collision shapes, or projectile behavior.
* Pickup entity lifecycle.
* Match result persistence outside the simulation state used to produce summaries.

## Domain roles

Player session state supports the player lifecycle inside the authoritative game-server simulation.

It acts as the match-local continuity layer between these runtime states:

```text
player joins simulation
-> playerSession is created
-> active runtime.Ship is created from the session
-> active ship receives input and simulation updates
-> fatal damage marks the active ship pending despawn
-> session counters and cooldown are updated
-> active ship is removed
-> respawn request checks the session
-> new active runtime.Ship is created from the session
```

The session is not the domain identity for a player across the platform. It is only the simulation-local state record for one running game instance.

## Protocols and APIs

Player session state participates in the realtime state protocol.

The client does not own or mutate arbitrary session fields. Client packets can indirectly affect session state through specific server-handled packet types and simulation events.

`client_config` packets update the session’s `runtime.ClientConfig` when the visible world size is valid. The same config can also be applied to the active ship and camera view when those runtime objects exist.

`respawn` packets request a respawn. The server uses the session to decide whether respawn is allowed, using remaining lives and cooldown, then creates a new active ship from the stored session state.

Session lane packets include player records. Each entry is a projected `PlayerSessionState` containing player-level fields that clients need even when a ship entity is absent or between lifecycle states.

The projected session packet includes:

```text
id
ship_type
score
lives
respawn_cooldown
primary_weapon_id
primary_ammo_policy
secondary_weapon_id
secondary_ammo_policy
spawn_x
spawn_y
```

This projection is not a persistence contract. It is a realtime presentation/state-sync surface owned by the game server.

## Data ownership

Player session state is in-memory simulation state.

The game server owns the session map for the lifetime of the `Game` instance. Session records are created when players are added and deleted when players are removed.

The current session defaults are resolved from runtime constants and runtime defaults:

* `runtime.DefaultShipTypeID`
* `runtime.ResolveShipStats`
* `constants.WorldWidth`
* `constants.WorldHeight`
* `constants.PlayerStartingLives`
* `weapons.DefaultPlayerArmory`

Session score and lives are clamped through player counter helpers. Fatal player damage increments `ShipDeaths`, may reduce lives, and sets `RespawnCooldown` when lives remain. Pickup effects can modify session lives through the same counter helpers.

Current pickup weapon equip behavior updates the active ship weapon state. The session armory remains the source used when a new ship is created from the session.

## Code map

Primary implementation files:

* `services/game-server/internal/game/game.go` - `Game` owns and initializes the `playerSessions` map.
* `services/game-server/internal/game/session.go` - Defines `playerSession`, session defaults, cooldown stepping, respawn eligibility, and ship creation from session state.
* `services/game-server/internal/game/players.go` - Creates and removes sessions through `AddPlayer` and `RemovePlayer`.
* `services/game-server/internal/game/player_session_state.go` - Projects `playerSession` records into `PlayerSessionState`.
* `services/game-server/internal/protocol/realtime/records.go` - Includes session projections in outbound lane packets.
* `services/game-server/internal/game/packets.go` - Defines generated packet structs, including `PlayerSessionState` and the lane packet types.
* `services/game-server/internal/game/input.go` - Applies valid client config packets to the session and routes respawn requests.
* `services/game-server/internal/game/pause.go` - Stores pause/suspension state on the session and gates player actions from that state.
* `services/game-server/internal/game/player_counters.go` - Mutates session score and lives.
* `services/game-server/internal/game/combat.go` - Updates session deaths, lives, and respawn cooldown after fatal player damage.
* `services/game-server/internal/game/targeting.go` - Stores targeting state on the session and applies it to active ships.
* `services/game-server/internal/game/player_targeting.go` - Defines session-owned player targeting projection and active-ship application.

Related generated/source files:

* `services/game-server/internal/game/packets.go` - Generated by `tools/data_sync/main.py`.
* `services/game-server/internal/constants/` - Provides world and player lifecycle constants used for session defaults and respawn timing.
* `services/game-server/internal/game/runtime/` - Provides ship, config, stats, suspension, damage, and life option types.
* `services/game-server/internal/game/weapons/` - Provides default armory and equipped weapon state.

Important non-ownership boundaries:

* `game.entities.Players` owns active ship entities, not persistent player session records.
* Room code owns room membership and room lifecycle.
* Networking code owns packet transport.
* Client code owns presentation of session lane records.

## Tests

Relevant verification is covered by game-server package tests around player lifecycle, respawn, counters, pause/suspension, targeting, and lane-native realtime projection.

Run the game-server tests after changing this area:

```bash
go test -buildvcs=false ./services/game-server/internal/game/...
```

Also run generated-data checks when packet shapes or generated packet definitions change:

```bash
data-sync -check -packets -go -gds
```

## Related docs

* [Game Server Simulation Players](./!INDEX.md)
* [Game Server Simulation](../!INDEX.md)
* [Game Server](../../!INDEX.md)
* [Game Server Simulation Targeting](../targeting/!INDEX.md)
* [Game Server Simulation Runtime](../runtime/!INDEX.md)
* [Realtime Protocol](../../../../protocol/!INDEX.md)
* [Data Pipeline](../../../../data/!INDEX.md)

## Notes

Legacy targeting documentation confirms that canonical gameplay targeting is stored as shared player/gameplay state. Current implementation stores that target reference on the player session and applies it to the active ship when one exists.

Session state is match-local. Anything that must survive beyond the running game instance belongs in player-data or another durable platform service, not in `playerSession`.
