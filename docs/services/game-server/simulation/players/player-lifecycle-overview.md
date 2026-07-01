# Player Lifecycle Overview

Parent index: [Game Server Simulation Players](./!INDEX.md)

## Purpose

This document describes the current game-server player lifecycle sequence at overview level.

It maps how a gameplay player enters the simulation, owns session state, receives an active runtime ship, moves through active, pending-respawn, and eliminated states, projects lifecycle state to clients, and leaves the simulation.

## Overview

Player lifecycle is owned by the game-server simulation after room and networking code have already decided that a connected room member should become an active game participant.

The lifecycle overview is:

```text
room/networking activation
-> Game.AddPlayer
-> playerSession
-> runtime.Ship
-> camera view
-> simulation stepping
-> world lane realtime projection
-> fatal damage
-> pending despawn
-> active ship removal
-> pending respawn or eliminated
-> respawn request
-> runtime.Ship recreation
-> camera view reattachment
-> room match-over observation
-> room return-to-lobby or cleanup removal
```

The game server intentionally separates player-facing lane readback into world and session lanes:

```text
playerSession
= durable per-match player state owned by the simulation

runtime.Ship
= active avatar/world entity while the player currently has a live ship

player_lifecycle
= packet-facing lifecycle classification derived from match rules
```

`playerSession` is the durable per-match record for score, lives, respawn cooldown, spawn position, client config, targeting, suspension state, damage/life options, ship type, ship stats, and player armory.

`runtime.Ship` is the active avatar entity in `game.entities.Players`. It owns live movement state, current input, rotation, velocity, health, shields, active weapon state, target fields, invulnerability, and pending-despawn state.

`world lane ship records` are active ship state only. They must not be used as the lifecycle source of truth. Pending-respawn and eliminated players can be absent from world lane ship records while still appearing in session lane player records and session lane lifecycle records.

## Code root

```text
services/game-server/internal/game/
services/game-server/internal/game/runtime/
services/game-server/internal/game/rules/
```

## Responsibilities

The player lifecycle overview boundary owns the sequence map across narrower player simulation docs.

It describes how these responsibilities fit together:

* Creating simulation players through `Game.AddPlayer`.
* Creating simulation-local `playerSession` records.
* Creating active `runtime.Ship` avatar entities from sessions.
* Initializing and updating player camera views.
* Advancing session cooldown state during simulation.
* Advancing active player movement, weapon input, camera position, and pending-despawn removal during simulation.
* Routing respawn requests to respawn eligibility and safe-placement logic.
* Moving fatal-damage consequences into pending despawn, lives mutation, death count mutation, respawn cooldown setup, death events, and lifecycle classification.
* Projecting active ship state, durable session state, and lifecycle status into world/session lane records.
* Allowing room lifecycle code to observe match-over state through game-owned match decisions.
* Removing player state when the game aggregate is explicitly told to remove a player.

This document does not replace the narrower player docs. It exists to show the sequence across them.

## Does not own

This overview does not own:

* Room membership, lobby ownership, ready checks, or match state transitions.
* WebSocket connection lifecycle.
* Per-connection `currentGamePlayerID` assignment.
* Room snapshot projection.
* Client gameplay state application.
* Client respawn UI.
* Client spectate camera behavior.
* Client match-end UI.
* Durable account/profile persistence.
* Player-data reporting.
* Detailed scoring rules.
* Detailed weapon fire policy.
* Detailed combat damage resolution.
* Detailed pickup effect rules.
* Devtools command handling.

Those concerns belong to rooms, networking, client, player-data, combat, pickup, weapon, protocol, data, or devtools documentation.

## Domain roles

Player lifecycle overview participates in the player, room, networking, combat, respawn, and projection domains by describing how a connected room member becomes an active simulation player and later moves through death, respawn, elimination, and removal.

It does not own the mechanics or policy decisions that make those transitions happen.

## Lifecycle sequence

### 1. Room and networking activation

Room and networking code decide when a connected room member should become an active game participant.

The current activation path lives outside the game simulation:

```text
successful room start
-> networking activateRoomPlayers
-> gameInstance.AddPlayer
-> websocket session currentGamePlayerID
-> room member player_id
-> room active player count
```

The game package does not own room membership or WebSocket session identity. It only creates a game player when `AddPlayer` is called.

### 2. Player creation

`Game.AddPlayer` creates both the durable session and the active ship.

Creation order:

```text
nextID
-> player ID
-> planInitialPlayerSpawn
-> newPlayerSession
-> session.NewShip
-> game.playerSessions[playerID]
-> game.entities.Players[playerID]
-> setPlayerCameraViewLocked
-> pendingPresentationEvents[playerID]
```

Initial spawn uses a `PlayerSpawnPlan` with reason `initial_player`. The preferred initial position is based on player index and then passed through the safe player spawn placement check.

The new session starts with generated server constants for starting lives and default visible world size. It also receives the default ship type, resolved ship stats, empty targeting state, and default player armory.

### 3. Active runtime ship

The active runtime ship is created from the session with:

```text
playerSession.NewShip(position)
```

The ship copies session-owned data into active avatar state:

```text
session.ID
session.ShipTypeID
session.Stats
session.Config
session.DamageOptions
session.PlayerArmory.Primary
session.PlayerArmory.Secondary
session.Targeting
```

The ship starts with max health from resolved ship stats. It then becomes the active world avatar stored in:

```text
game.entities.Players[playerID]
```

Only active ships are movement, collision, damage, targeting, and weapon-fire participants.

### 4. Camera view attachment

When a player is created or respawned, the game attaches or refreshes the player's camera view with:

```text
setPlayerCameraViewLocked(playerID, player)
```

The camera view stores a server-side read model for the player's current view center and visible world config. It follows the active ship during simulation.

If the active ship later dies and enters pending despawn, the camera view is preserved at the death position. If the active ship is removed while the player still has a session, fallback player world state can still use the camera view position.

### 5. Simulation stepping

`Game.Step(delta)` drives player lifecycle during the game simulation.

Normal non-match-over player order:

```text
stepPlayerSessions
stepPlayerWeapons
stepPlayers
removeReadyPlayers
...
stepCollisions
```

`stepPlayerSessions` decrements respawn cooldowns on durable sessions.

`stepPlayerWeapons` advances mutable weapon cooldown state on active ships.

`stepPlayers` advances active ship movement through the motion seam, updates camera views, skips pending-despawn ships for input/fire behavior, and routes primary and secondary fire when input and shooting gates allow it.

`removeReadyPlayers` deletes active ships whose pending-despawn delay has finished.

After match over, `Game.Step` still advances session cooldowns and cleanup-safe entity stepping, but skips normal player movement, player weapon stepping, player despawn removal, asteroid spawning, collisions, and pickup collection.

### 6. Input routing while active

`Game.HandlePacket(playerID, packet)` is the simulation-local packet dispatch entry point after networking has already resolved the game player context.

Current input-related packet behavior:

```text
respawn
-> respawnPlayer(playerID)
-> return

client_config
-> update session config
-> update camera view config
-> update active ship config when active

input
-> active ship required
-> playerCanReceiveInput
-> player.SetInput

pause_request
-> active ship required
-> togglePlayerPaused
```

Respawn requests are special because they can be valid while the player does not have an active ship. Normal input, pause, and active ship config mutation require an active `runtime.Ship`.

### 7. Fatal damage and pending despawn

Fatal player damage currently reaches player lifecycle from combat-owned collision and damage application.

Current fatal player path:

```text
ship/asteroid collision
-> playerCanTakeCollisionDamage
-> playerAsteroidDamageRequest
-> damage.ResolveSingle
-> applyDamageResultToPlayer
-> applyFatalPlayerDamage
```

`applyFatalPlayerDamage` performs player lifecycle consequences:

```text
store death position in camera view
mark active ship pending despawn
increment session ship deaths
decrement lives when life options allow it
set respawn cooldown when lives remain
record ship death event
```

The active ship is not deleted immediately. It is marked pending despawn with the configured collision despawn delay. During this state it cannot receive input, move, shoot, take collision damage, or participate as a normal active player.

Once the pending-despawn delay reaches zero, `removeReadyPlayers` removes the ship from `game.entities.Players`.

### 8. Pending respawn

A player is pending respawn when:

```text
playerSession exists
active non-pending ship does not exist
session.Lives > 0
```

The session remains in `game.playerSessions`. The player can still appear in `session lane player records` and `session lane lifecycle records`, but should be absent from `world lane ship records` once the pending-despawn ship has been removed.

The session's `RespawnCooldown` is decremented by `stepPlayerSessions`.

Respawn eligibility is defined by:

```text
session.CanRespawn()
= session.Lives > 0 && session.RespawnCooldown == 0
```

### 9. Respawn request

Respawn is requested through a gameplay packet with type `respawn`.

Respawn handling rejects the request when:

```text
session is missing
session.CanRespawn() is false
an active ship already exists for the player
```

When the request succeeds, the game:

```text
planPlayerRespawn(session)
-> safeRespawnPosition(session)
-> session.NewShip(spawnPosition)
-> game.entities.Players[playerID]
-> setPlayerCameraViewLocked(playerID, player)
```

Respawn placement starts from the session's stored spawn position and searches for a safe position when the original position is blocked. Safety checks use the player's collision shape, active asteroids, active players, wrapped distance, and the configured respawn buffer.

Respawn recreates an active `runtime.Ship` from durable session state. It does not create a new session.

### 10. Eliminated

A player is eliminated when:

```text
playerSession exists
active non-pending ship does not exist
session.Lives <= 0
```

Eliminated players remain part of the per-match player session state until the game or room lifecycle removes the game instance or explicitly removes the player.

Eliminated players are not active, targetable, damageable, or collidable. Their session state can still be used for scoreboard, match result, and world lane realtime projection.

### 11. Match-over observation

The game package exposes match-over state through:

```text
Game.MatchDecision()
Game.IsGameOver()
```

The game builds a match snapshot from sessions and active ship presence:

```text
session.ID
has active non-pending ship
has remaining lives
```

The pure rules package classifies each player as:

```text
active
pending_respawn
eliminated
```

The current match is over only when all players in a non-empty match snapshot are eliminated.

Room lifecycle code observes the game match decision. The room owns the transition from `InGame` to `GameOver`; the game owns the facts used by that decision.

### 12. Removal

`Game.RemovePlayer(playerID)` removes all game-owned player state for that player:

```text
delete game.entities.Players[playerID]
delete game.cameraViews[playerID]
delete game.playerSessions[playerID]
clear targets for missing players
delete game.pendingPresentationEvents[playerID]
```

This is full simulation removal, not normal death. Normal death removes or despawns the active ship but keeps the session.

Room return-to-lobby and room cleanup stop and clear the game instance rather than routing every player through normal death.

## Lifecycle states

Current packet-facing lifecycle states are:

```text
active
pending_respawn
eliminated
```

These states are derived from current match facts, not stored directly as a field on `playerSession`.

### Active

```text
session exists
active non-pending runtime.Ship exists
```

Active players have a live avatar in `game.entities.Players`.

They can be included in `world lane ship records`, can be projected as active in `session lane lifecycle records`, and can participate in movement, collision, damage, targeting, and weapon fire when no suspension or invulnerability gate blocks the specific action.

### Pending respawn

```text
session exists
no active non-pending runtime.Ship exists
lives > 0
```

Pending-respawn players remain in durable session state but are not active world avatars.

They can be shown in `session lane player records` and `session lane lifecycle records`. They should not be inferred from `world lane ship records`.

### Eliminated

```text
session exists
no active non-pending runtime.Ship exists
lives <= 0
```

Eliminated players remain available for match facts and result projection. They are not active world participants.

## Protocols and APIs

Player lifecycle is projected through three related world lane packet areas.

### Active ship state

```text
world lane ship records
```

This map is built from `game.entities.Players`.

It contains active runtime ship state. It is not the full player lifecycle read model.

### Durable session state

```text
session lane player records
```

This map is built from `game.playerSessions`.

It includes durable per-match values such as:

```text
id
ship_type
score
lives
respawn_cooldown
spawn_x
spawn_y
primary_weapon_id
primary_ammo_policy
secondary_weapon_id
secondary_ammo_policy
```

### Lifecycle status

```text
session lane lifecycle records
```

This map is built from `MatchDecision.Players`.

It projects each player ID to one lifecycle status string:

```text
active
pending_respawn
eliminated
```

This is the packet-facing lifecycle status that clients should use when deciding whether a player is active, pending respawn, or eliminated.

## Data ownership

### Session-owned data

`playerSession` owns durable per-match player data:

```text
ID
ShipTypeID
Stats
SpawnPosition
Config
Targeting
Score
Lives
ShipDeaths
RespawnCooldown
Suspension
DamageOptions
LifeOptions
PlayerArmory
```

Session data persists across active ship death and respawn.

### Active ship-owned data

`runtime.Ship` owns active avatar state:

```text
ID
ShipTypeID
Stats
X/Y
Rotation
Velocity
Input
Config
ShipWeapons
WeaponState
TargetKind
TargetID
Health
Shields
DamageModifiers
DamageOptions
InvulnerabilityRemaining
PendingDespawn
DespawnDelay
```

Active ship state exists only while the player currently has a live or pending-despawn avatar in the entity store.

### Camera-view-owned data

`runtime.CameraView` owns the server-side view read model:

```text
X
Y
Config
```

It tracks active ship position while active and preserves a useful view position when the active ship is gone.

### Rule-owned lifecycle classification

`services/game-server/internal/game/rules` owns pure lifecycle classification from plain match snapshot facts.

It does not mutate the game. It receives only:

```text
player ID
has active ship
has remaining lives
```

and returns match decision data.

## Code map

### Lifecycle creation and removal

```text
services/game-server/internal/game/players.go
```

Owns `Game.AddPlayer`, active ship/session/camera creation, full player removal, and player lives lookup.

```text
services/game-server/internal/networking/player_activation.go
```

Activates connected room members into game players and stores the resulting game player ID on networking session and room-member state.

### Session and respawn

```text
services/game-server/internal/game/session.go
```

Defines `playerSession`, session defaults, cooldown stepping, respawn eligibility, ship recreation, initial spawn planning, respawn planning, and safe respawn placement.

```text
services/game-server/internal/game/spawn_types.go
```

Defines player spawn plan vocabulary and reasons used by initial spawn and respawn.

### Active player simulation

```text
services/game-server/internal/game/simulation.go
```

Owns top-level simulation order and match-over branch behavior.

```text
services/game-server/internal/game/simulation_players.go
```

Owns player session stepping, active player movement/fire stepping, camera view updates, and pending-despawn removal.

```text
services/game-server/internal/game/runtime/state.go
```

Defines `runtime.Ship`, `runtime.CameraView`, and the game entity store maps.

```text
services/game-server/internal/game/runtime/ship.go
```

Defines active ship state projection, input/config mutation, pending-despawn behavior, position read model, invulnerability check, and collision body construction.

```text
services/game-server/internal/game/runtime/camera.go
```

Defines camera view config, position, visibility, and far-from-view helpers.

### Input, pause, and suspension gates

```text
services/game-server/internal/game/input.go
```

Routes simulation-local gameplay packets for respawn, client config, input, and pause.

```text
services/game-server/internal/game/pause.go
```

Owns pause mutation and player action gates for input, movement, shooting, collision damage, and score eligibility.

```text
services/game-server/internal/game/runtime/suspension.go
```

Defines runtime suspension state used by the player session.

### Death and lifecycle consequences

```text
services/game-server/internal/game/combat.go
```

Routes ship/asteroid fatal damage into player lifecycle consequences, pending despawn, lives mutation, respawn cooldown setup, death count mutation, logging, and ship death event recording.

```text
services/game-server/internal/game/combat_damage_application.go
```

Applies damage results to active player health and shields.

```text
services/game-server/internal/game/player_counters.go
```

Owns score and lives counter mutation helpers used by gameplay and devtools adapters.

```text
services/game-server/internal/game/events.go
services/game-server/internal/game/events/events.go
```

Convert ship death domain events into packet-facing event state.

### Lifecycle read models and projection

```text
services/game-server/internal/game/match.go
```

Builds match snapshots, exposes match decisions, and exposes player match facts.

```text
services/game-server/internal/game/rules/match.go
```

Classifies players as active, pending respawn, or eliminated and determines whether the match is over.

```text
services/game-server/internal/protocol/realtime/records.go
```

Projects active ships, player sessions, lifecycle status, entities, and presentation events into gameplay world lane packets.

```text
services/game-server/internal/game/player_session_state.go
```

Projects durable session state into `PlayerSessionState`.

```text
services/game-server/internal/game/player_world_state.go
```

Builds a server-side player world read model from session, active ship, and camera view state.

```text
services/game-server/internal/game/player/state.go
```

Defines player world status and capability flags for active, pending-respawn, and eliminated read models.

### Generated and source data

```text
shared/constants/server_entities.toml
services/game-server/internal/constants/constants.go
shared/packets/gameplay.toml
services/game-server/internal/game/packets.go
services/game-server/internal/game/runtime/packets_generated.go
```

`shared/constants/server_entities.toml` defines starting lives and respawn delay. Generated constants are consumed by player session and death/respawn code. Gameplay packet source defines packet-facing player session and lifecycle fields.

### Important non-ownership boundaries

```text
services/game-server/internal/rooms/
```

Owns room lifecycle, match state transitions, return-to-lobby behavior, and game instance ownership.

```text
services/game-server/internal/networking/
```

Owns WebSocket transport, session routing, active game player ID assignment, gameplay tick calls, and packet delivery.

```text
client/
```

Owns presentation, input collection, local UI, respawn UI, spectate UI, world rendering, and match-end display.

## Tests

Relevant current test coverage includes:

```text
services/game-server/internal/game/player_world_state_test.go
```

Verifies active, pending-respawn, and eliminated player world-state classification.

```text
services/game-server/internal/game/player/state_test.go
```

Verifies player world-state status and capability flags and snake_case JSON projection.

```text
services/game-server/internal/game/simulation_match_over_test.go
```

Verifies post-match-over simulation behavior skips normal asteroid spawning and remains cleanup-safe.

```text
services/game-server/internal/game/match_facts_test.go
```

Verifies player match facts used by room match result construction.

```text
services/game-server/internal/game/events_test.go
```

Verifies ship death event projection, including lives and respawn delay data.

```text
services/game-server/internal/game/targeting_test.go
```

Includes coverage for pending-respawn player target lookup, inactive target preservation, and target restoration on respawn.

```text
services/game-server/internal/networking/player_activation_test.go
```

Verifies room member activation into game player IDs and preservation of account identity.

```text
services/game-server/internal/rooms/room_match_summary_test.go
```

Verifies match summary construction from game-owned player facts.

Suggested verification command from `services/game-server`:

```text
go test -buildvcs=false ./internal/game ./internal/game/player ./internal/game/rules ./internal/networking ./internal/rooms
```

## Related docs

* [Game Server Simulation Players](./!INDEX.md)
* [Game Server Simulation](../!INDEX.md)
* [Game Server](../../!INDEX.md)
* [Player Session State](player-session-state.md)
* [Active Player Avatar State](active-player-avatar-state.md)
* [Player Counters](player-counters.md)
* [Player Death And Despawn](player-death-and-despawn.md)
* [Player Respawn](player-respawn.md)
* [Player Input Routing](player-input-routing.md)
* [Player Pause And Suspension](player-pause-and-suspension.md)
* [Player Camera View State](player-camera-view-state.md)
* [Room Match Lifecycle](../../rooms/room-match-lifecycle.md)
* [Room Membership And Identity](../../rooms/room-membership-and-identity.md)
* [Room Snapshot Projection](../../rooms/room-snapshot-projection.md)
* [Game Server Networking](../../networking/!INDEX.md)
* [Gameplay State Application](../../../client/gameplay-runtime/gameplay-state-application.md)
* [Gameplay packets](../../../../protocol/gameplay-packets.md) - gameplay realtime packet documentation.
* [Constants pipeline](../../../../data/data-sync-and-ssot-pipeline.md) - generated constant data documentation.

## Notes

Legacy documentation supplied one still-current lifecycle rule: do not infer player lifecycle from active ship presence alone. `world lane ship records` is active avatar state, while `session lane lifecycle records` and `session lane player records` preserve lifecycle and durable session information for inactive players.

Current player lifecycle state is partly duplicated across `rules.PlayerParticipationStatus` and `player.Status`. The former drives match decision and `session lane lifecycle records`; the latter drives the server-side player world read model. Both currently use the same status strings.

`PlayerID` values currently use lowercase `player-<n>` from `Game.AddPlayer`. Some older documentation examples use capitalized `Player-<n>` as illustrative values. The active code path is lowercase.
