# Player Camera View State

Parent index: [Game Server Simulation Players](./!INDEX.md)

## Purpose

This document describes the game-server service boundary for per-player camera view state.

It covers the server-side `CameraView` model used for viewport-sized spawning, visibility, despawn, and inactive-player position continuity. It does not describe the Godot `Camera2D`, client `ViewAnchor`, or visual interpolation.

## Overview

The game server keeps a camera-view record per simulation player in:

```text
game.cameraViews map[string]*runtime.CameraView
```

A `runtime.CameraView` stores:

```text
X
Y
ClientConfig.VisibleWorldWidth
ClientConfig.VisibleWorldHeight
```

This is not a rendered camera. It is a server-side visibility and spawning anchor. The client owns rendered camera presentation, but the server needs each player’s visible world size and current anchor position so it can spawn asteroids offscreen, prevent spawning inside any current viewport, and remove bullets or asteroids only after they are far from all player views.

The normal camera-view lifecycle is:

```text
Game.AddPlayer()
-> new playerSession
-> session.NewShip(spawnPosition)
-> game.entities.Players[playerID] = ship
-> setPlayerCameraViewLocked(playerID, ship)

Client client_config packet
-> Game.HandlePacket()
-> session.Config = packet.Config
-> cameraView.Config = packet.Config
-> active ship Config = packet.Config, when active

Game.Step()
-> stepPlayers()
-> motion.AdvanceShipWithMovePolicy()
-> cameraView.SetPosition(player.Position())

Fatal player damage
-> cameraView position is preserved at death position
-> active ship is marked pending despawn

Respawn
-> session.NewShip(spawnPosition)
-> game.entities.Players[playerID] = ship
-> setPlayerCameraViewLocked(playerID, ship)

RemovePlayer()
-> deletes active ship, camera view, player session, targets, and pending events
```

The camera view also acts as the fallback position source for server-side player world state when the player has a session but no active non-despawning ship. That lets pending-respawn or eliminated player read models retain the last known camera/death location without making the player targetable, damageable, or collidable.

## Code root

```text
services/game-server/internal/game/
services/game-server/internal/game/runtime/
services/game-server/tests/game/
shared/packets/gameplay.toml
```

## Responsibilities

The player camera-view boundary owns:

* Creating a camera view when an active player ship is created through normal player add, respawn, or devtools spawn/respawn paths.
* Storing one camera view per player ID.
* Storing the server’s current per-player visibility anchor position.
* Storing the most recent valid viewport config received for the player.
* Seeding missing camera config from session config, active ship config, or world-size fallback.
* Updating camera-view position from the active ship during simulation stepping.
* Preserving the camera-view position at fatal player damage time.
* Removing camera-view state when the player is removed from the game.
* Providing per-player viewport anchors for asteroid spawning.
* Providing all-player viewport anchors for bullet and asteroid visibility/despawn checks.
* Providing inactive-player fallback position for `playerWorldStateLocked`.

## Does not own

This boundary does not own:

* Godot `Camera2D` nodes.
* Client `ViewAnchor` movement.
* Client visual coordinates.
* Client interpolation.
* Client background/parallax presentation.
* Spectate camera selection.
* Client viewport-size measurement.
* Packet schema source-of-truth.
* WebSocket transport.
* Room membership or room lifecycle.
* Player session counters.
* Active player ship/avatar ownership.
* Respawn rules.
* Collision detection.
* Targetability, damageability, or collidability rules.
* General toroidal world design.

Those responsibilities belong to client world-sync, client app-shell/session config, realtime protocol/data, rooms, player session state, active avatar state, respawn, targeting, combat, or world docs as appropriate.

## Domain roles

### Server visibility anchor

`CameraView` is the server’s per-player visibility anchor.

World systems use it to decide whether positions are inside any viewport-sized player view or far enough from every view to remove runtime entities.

Current consumers include:

```text
randomAsteroidSpawnPosition()
isOnscreenForAnyCamera()
isAsteroidFarFromAllCameras()
isBulletFarFromAllCameras()
```

Visibility checks use toroidal-space deltas through `space.Delta`, so entities near a wrapped world edge can still be treated as near a camera on the opposite edge.

### Asteroid spawning anchor

Timed asteroid spawning runs only when asteroid spawning is enabled and at least one camera view exists.

The current spawn path is:

```text
stepAsteroidSpawning()
-> for each cameraView
-> spawnAsteroidBatch(cameraView)
-> spawnAsteroid(cameraView)
-> randomAsteroidSpawnPosition(cameraView)
-> space.NormalizePosition(spawn)
-> spawner.PlanTimedAsteroidSpawn(spawn, cameraView.Position())
-> applyAsteroidSpawn(plan)
```

The spawn position is selected offscreen from the target camera view and rejected if it is onscreen for any player camera view. The final stored asteroid position is normalized back inside world bounds.

### Bullet and asteroid cleanup anchor

`stepAsteroids` and `stepBullets` use camera views for cleanup:

```text
asteroid far from all cameras -> delete asteroid
bullet expired -> delete bullet
bullet far from all cameras -> delete bullet
```

If there are no camera views, far-from-all-camera checks return true for bullets and asteroids. Asteroid spawn elapsed time is also reset when there are no camera views.

### Player world-state fallback

`playerWorldStateLocked` chooses a player position in this order:

```text
session.SpawnPosition
active non-pending ship position
cameraView position, if there is no active non-pending ship
```

The fallback does not make the inactive player active. `player.BuildWorldState` derives capability flags from active-ship presence and lives:

```text
active ship -> active, targetable, damageable, collidable
no active ship + lives > 0 -> pending_respawn, not targetable, not damageable, not collidable
no active ship + lives <= 0 -> eliminated, not targetable, not damageable, not collidable
```

The camera-view fallback is positional continuity only.

### Death-position preservation

When fatal player damage is applied, the server captures the active ship position into the camera view before marking the ship pending despawn.

This preserves the last player anchor position while the active ship is pending removal and while respawn/game-over state is resolved.

### Respawn reattachment

Respawn creates a new active ship from the player session and reattaches the camera view to that new ship position through `setPlayerCameraViewLocked`.

`setPlayerCameraViewLocked` preserves an existing valid camera config when possible. This avoids viewport-size reset/flicker across respawn. If the camera view does not already have a valid config, it seeds from session config, then active ship config, then full world size.

## Protocols and APIs

The camera-view boundary consumes the generated gameplay `client_config` packet.

The packet is for reporting the client’s visible viewport size to the server. The Godot client measures its visible viewport and sends:

```text
type = "client_config"
config.visible_world_width
config.visible_world_height
```

The server owns authority behind this surface. The client reports dimensions; the server decides how those dimensions affect spawning, visibility, cleanup, and stored runtime config.

`Game.HandlePacket` accepts a client config only when both dimensions are positive:

```text
packet.Config.VisibleWorldWidth > 0
packet.Config.VisibleWorldHeight > 0
```

When valid, the server updates:

```text
playerSession.Config
cameraView.Config, if a camera view exists
runtime.Ship.Config, if an active ship exists
```

This means a pending-respawn player with a session and camera view can still refresh server camera config even when no active ship exists.

`CameraView` is not projected directly in lane packets. Clients observe active ship state through world lane records, player session state through session lane records, lifecycle status through session lane lifecycle records, bullets through world lane bullet records, asteroids through world lane asteroid records, pickups through world lane pickup records, and events through `event_batch`. The server camera view remains internal simulation state.

Devtools paths can also create or update camera views when forcing player spawn or respawn:

```text
DevtoolsSpawnPlayerShip()
DevtoolsForceRespawnPlayer()
```

Those paths are devtools adapters around game-owned player/camera state. They do not make devtools the owner of camera-view behavior.

## Data ownership

Camera-view data is transient runtime state.

It is stored only in memory on the game instance:

```text
game.cameraViews
```

The server does not persist camera views.

The source-of-truth for the packet field shape is:

```text
shared/packets/gameplay.toml
```

Generated server outputs include:

```text
services/game-server/internal/game/packets.go
services/game-server/internal/game/runtime/packets_generated.go
```

The generated runtime `ClientConfig` shape is:

```text
VisibleWorldWidth
VisibleWorldHeight
```

Generated files should not be edited manually.

## Code map

Primary implementation files:

* `services/game-server/internal/game/game.go` - Owns the `cameraViews` map on `Game` and initializes it in `New`.
* `services/game-server/internal/game/runtime/state.go` - Defines `runtime.CameraView`.
* `services/game-server/internal/game/runtime/camera.go` - Implements camera config, position, viewport-size fallback, and rectangular inside/far helpers.
* `services/game-server/internal/game/players.go` - Creates, seeds, updates, and removes camera views during player add/remove.
* `services/game-server/internal/game/session.go` - Creates new ships for respawn and reattaches camera views through `setPlayerCameraViewLocked`.
* `services/game-server/internal/game/input.go` - Handles `client_config` packets and updates session, camera-view, and active-ship config.
* `services/game-server/internal/game/simulation_players.go` - Updates camera-view position from active ship position after movement.
* `services/game-server/internal/game/combat.go` - Preserves camera-view position when fatal player damage is applied.
* `services/game-server/internal/game/player_world_state.go` - Uses camera-view position as inactive-player fallback position.
* `services/game-server/internal/game/simulation_asteroids.go` - Uses camera-view presence and per-view spawning.
* `services/game-server/internal/game/spawning.go` - Spawns asteroid batches against camera views.
* `services/game-server/internal/game/visibility.go` - Uses camera views for offscreen spawning, onscreen rejection, and far-from-all-cameras cleanup.
* `services/game-server/internal/game/simulation_bullets.go` - Removes bullets that are expired or far from all camera views.
* `services/game-server/internal/game/export_devtools_respawn.go` - Devtools force-respawn adapter that creates/updates camera view state.
* `services/game-server/internal/game/export_devtools_player_spawn.go` - Devtools player-spawn adapter that creates/updates camera view state.

Source-of-truth and generated files:

* `shared/packets/gameplay.toml` - `ClientConfig` and `client_config` packet source.
* `services/game-server/internal/game/packets.go` - Generated `ClientPacket` and `PacketTypeClientConfig`.
* `services/game-server/internal/game/runtime/packets_generated.go` - Generated `ClientConfig`.

Related tests:

* `services/game-server/tests/game/spawning_test.go` - Verifies asteroid spawning uses client camera view and wraps boundary spawn positions into world bounds.
* `services/game-server/tests/game/visibility_test.go` - Verifies cross-edge camera visibility and far bullet/asteroid cleanup.
* `services/game-server/tests/game/helpers_test.go` - Provides camera-view test helper setup.
* `services/game-server/internal/game/player_world_state_test.go` - Verifies inactive player world-state capability flags.
* `services/game-server/internal/game/export_devtools_respawn_test.go` - Verifies devtools force-respawn creates camera view with dummy config.
* `services/game-server/internal/game/export_devtools_player_spawn_test.go` - Verifies devtools spawn creates camera view with dummy config.
* `services/game-server/internal/game/simulation_match_over_test.go` - Verifies match-over stepping does not spawn asteroids even with a camera view present.

Important non-ownership boundaries:

* `client/scripts/config/client_viewport_config_flow.gd` measures and sends viewport config; it does not own server camera behavior.
* `client/scripts/world/world_sync.gd` and `client/scripts/world/player_render/` own rendered view-anchor behavior; they do not own server camera views.
* `services/game-server/internal/game/space/` owns toroidal spatial math used by visibility checks.
* `services/game-server/internal/game/motion/` owns ship, asteroid, and bullet movement; camera views copy active ship position after movement.
* `services/game-server/internal/game/player/` owns player world-state shape and capability derivation, not camera-view storage.

## Tests

Focused verification should cover:

```text
go test ./services/game-server/internal/game/... -buildvcs=false
go test ./services/game-server/tests/game/... -buildvcs=false
```

Relevant behavior to preserve:

* valid `client_config` packets update player session config and camera-view config
* invalid non-positive viewport dimensions are ignored
* camera view is created on player add
* camera view is reattached on respawn
* camera view position follows active ship movement
* death preserves the last active ship position in camera view
* pending-respawn and eliminated player world state may use camera-view position without becoming targetable, damageable, or collidable
* asteroid spawning uses camera views and does not spawn onscreen for any camera
* asteroid spawn positions near world edges are normalized into bounds
* bullets and asteroids across wrapped edges remain near camera views
* bullets and asteroids far from every camera are removed
* no-camera-view state does not continue asteroid spawn elapsed accumulation

## Related docs

* [Game Server Simulation Players](./!INDEX.md)
* [Game Server Simulation](../!INDEX.md)
* [Game Server](../../!INDEX.md)
* [Services](../../../!INDEX.md)
* [Client Viewport Config Flow](../../../client/app-shell-and-session/client-viewport-config-flow.md)
* [View Anchor And Visual Coordinates](../../../client/world-sync/view-anchor-and-visual-coordinates.md)
* [World Sync Coordinator](../../../client/world-sync/world-sync-coordinator.md)
* [Spectate Session And Camera Flow](../../../client/spectate-flow/spectate-session-and-camera-flow.md)
* [Protocol](../../../../protocol/!INDEX.md)
* [Data](../../../../data/!INDEX.md)
* [Systems Design World](../../../../systems-design/world/!INDEX.md)

## Notes

The name `CameraView` is server-side terminology. It should not be read as ownership of the client camera node.

The current server camera view is anchored to active ship position during normal gameplay. Spectate and rendered camera ownership remain client-side presentation concerns.

`runtime.CameraView.IsInside` and `runtime.CameraView.IsFarFrom` perform non-wrapped rectangle checks on the runtime type, but the game-level visibility helpers use `space.Delta` for toroidal-aware camera comparisons. Current gameplay visibility/despawn behavior should use the game-level helpers.
