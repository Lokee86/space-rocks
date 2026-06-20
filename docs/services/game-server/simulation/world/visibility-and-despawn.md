# Visibility And Despawn

Parent index: [Game Server Simulation World](./!README.md)

## Purpose

This document describes the game-server service boundary for world visibility checks and despawn/removal behavior.

It covers server-side decisions that use player camera views to keep bullets and asteroids alive while they are near any player viewport, remove them after they are far from all player viewports, and preserve short pending-despawn windows for collision presentation.

## Overview

The game server owns world visibility and despawn as part of the authoritative simulation.

Visibility is not client rendering visibility. It is server-side world-retention policy. The server uses per-player `runtime.CameraView` records to decide whether a world position is inside any player view, outside all player views, or far enough from all player views to remove an entity.

The current visibility/despawn flow is:

```text
Game.Step()
-> stepPlayers()
-> update camera view positions from active ships
-> stepAsteroidSpawning()
-> spawn offscreen relative to camera views
-> reject spawn positions visible to any camera view
-> stepAsteroids()
-> remove ready pending-despawn asteroids
-> remove asteroids far from all camera views
-> stepBullets()
-> remove ready pending-despawn bullets
-> remove expired bullets
-> remove bullets far from all camera views
-> stepCollisions()
-> mark hit bullets, destroyed asteroids, and killed players pending despawn
```

During match-over locked stepping, the simulation still advances asteroids, bullets, pickups, and radial effects, but does not run normal player movement, asteroid spawning, weapon firing, or collisions. That lets already-existing world entities continue their cleanup path without continuing normal match gameplay.

Visibility checks use toroidal-space deltas. An entity near one edge of the wrapped world can still be near a camera view on the opposite edge.

The main visibility helpers are:

```text
isOnscreenForAnyCamera(position)
isAsteroidFarFromAllCameras(asteroid)
isBulletFarFromAllCameras(bullet)
isInsideCameraView(view, position)
isFarFromCameraView(view, position)
```

`isInsideCameraView` and `isFarFromCameraView` compare a position against a camera-sized rectangle centered on the camera view. They use `space.Delta(view.Position(), position)` so the rectangle comparison is made in shortest wrapped space.

## Code root

```text
services/game-server/internal/game/
services/game-server/internal/game/runtime/
services/game-server/internal/game/motion/
services/game-server/internal/game/space/
services/game-server/tests/game/
shared/constants/
```

## Responsibilities

The visibility and despawn boundary owns:

* Deciding whether a position is onscreen for any server camera view.
* Deciding whether an asteroid is far from all camera views.
* Deciding whether a bullet is far from all camera views.
* Using toroidal-aware deltas for visibility and far-from-camera checks.
* Removing asteroids that are ready after pending despawn.
* Removing asteroids that are far from every camera view.
* Removing bullets that are ready after pending despawn.
* Removing bullets that have expired.
* Removing bullets that are far from every camera view.
* Treating missing camera views as no active world visibility anchors for bullet and asteroid retention.
* Supporting offscreen asteroid spawn placement by rejecting positions visible to any camera.
* Leaving collision-hit bullets, destroyed asteroids, and killed players visible for a short pending-despawn delay before removal.

## Does not own

This boundary does not own:

* Creating or updating camera-view state.
* Client `Camera2D` nodes.
* Client `ViewAnchor` or visual interpolation.
* Client-side culling or rendering.
* Client viewport measurement.
* Asteroid spawn scheduling or spawn plan construction.
* Bullet spawning.
* Collision detection.
* Damage resolution.
* Scoring.
* Pickup lifecycle.
* Player session state.
* Player respawn state.
* Player camera-view fallback state.
* Packet schema source-of-truth.
* WebSocket transport.
* Room lifecycle.

Those responsibilities belong to player camera-view state, client world-sync/presentation docs, asteroid spawning, combat, scoring, pickups, runtime state, protocol, networking, or room docs as appropriate.

## Domain roles

### Server-side visibility rectangles

Each `runtime.CameraView` provides a server-side rectangle:

```text
center = cameraView.Position()
width = cameraView.VisibleWorldWidth()
height = cameraView.VisibleWorldHeight()
```

A position is inside the view when its wrapped delta from the camera center is within half the visible width and half the visible height:

```text
abs(delta.X) <= visible_width * 0.5
abs(delta.Y) <= visible_height * 0.5
```

This check is used to reject asteroid spawn positions that would appear onscreen for any player.

### Far-from-camera cleanup

A bullet or asteroid is far from a camera view when its wrapped delta is outside the visible rectangle plus `constants.AsteroidDespawnMargin`:

```text
abs(delta.X) > visible_width * 0.5 + AsteroidDespawnMargin
or
abs(delta.Y) > visible_height * 0.5 + AsteroidDespawnMargin
```

The current generated despawn margin is:

```text
constants.AsteroidDespawnMargin = 320.0
```

An asteroid or bullet is removed by camera-distance cleanup only when it is far from every camera view.

### No-camera behavior

When `game.cameraViews` is empty, far-from-all-camera checks return true.

That means:

```text
no camera views -> asteroids are treated as far from all cameras
no camera views -> bullets are treated as far from all cameras
no camera views -> asteroid spawn elapsed time is reset
```

Asteroid spawning also requires at least one camera view. Without camera views, spawn timing does not continue accumulating.

### Asteroid removal

Asteroids are stepped in `stepAsteroids`.

The current asteroid cleanup order is:

```text
motion.AdvanceAsteroid(), when asteroids can move
if asteroid.ReadyForRemoval() -> delete
else if asteroid is far from all cameras -> delete
```

A pending-despawn asteroid counts down its `DespawnDelay` through the motion package. While pending despawn, it does not continue moving. Once the delay reaches zero, `ReadyForRemoval()` returns true and the asteroid is deleted from `game.entities.Asteroids`.

Asteroids can also be removed without pending despawn when they drift far from all camera views.

### Bullet removal

Bullets are stepped in `stepBullets`.

The current bullet cleanup order is:

```text
motion.AdvanceBullet(), when bullets can move
if bullet.ReadyForRemoval() -> delete
else if bullet.IsExpired() -> delete
else if bullet is far from all cameras -> delete
```

A pending-despawn bullet counts down its `DespawnDelay` through the motion package. While pending despawn, it does not continue moving and its lifetime is not reduced by `StepBullet`.

A non-pending bullet can be removed by lifetime expiry or by becoming far from every camera view.

### Pending despawn

Pending despawn is a short presentation-preservation window used after collisions and destruction.

Current pending-despawn participants include:

```text
runtime.Ship
runtime.Bullet
runtime.Asteroid
```

Each has:

```text
PendingDespawn
DespawnDelay
IsPendingDespawn()
ReadyForRemoval()
MarkPendingDespawn(delay)
```

`MarkPendingDespawn` sets the pending flag, stores the delay, and clears velocity.

The current generated collision despawn delay is:

```text
constants.CollisionDespawnDelay = 0.05
```

This delay lets hit bullets, destroyed asteroids, and killed player ships remain in projected state briefly enough for clients to receive the final visible state and related presentation events.

### Collision consequences

Visibility and despawn do not decide collision outcomes. Collision and combat code decide when an entity should become pending despawn.

Current collision-driven pending despawn paths include:

```text
bullet hits asteroid
-> bullet.MarkPendingDespawn(CollisionDespawnDelay)

asteroid is destroyed by projectile damage
-> asteroid.MarkPendingDespawn(CollisionDespawnDelay)
-> spawn asteroid fragments
-> maybe drop pickup

fatal player damage
-> preserve camera view position
-> player.MarkPendingDespawn(CollisionDespawnDelay)
-> decrement lives / set respawn cooldown
-> record ship death event
```

Ship-asteroid collision does not currently despawn the asteroid merely because it killed a player.

### State projection during pending despawn

`statePacket` projects current entity maps directly.

That means entities marked pending despawn remain in state packets until their simulation step removes them from the relevant map.

Current projected maps include:

```text
game.entities.Players     -> StatePacket.players
game.entities.Projectiles -> StatePacket.bullets
game.entities.Asteroids   -> StatePacket.asteroids
```

After removal, the entity disappears from future state packets.

### Match-over locked stepping

When the match-over lock is active, `Game.Step` uses a reduced phase order:

```text
stepPlayerSessions()
stepAsteroids()
stepBullets()
stepPickups()
stepRadialEffects()
simulationStepObservers
return
```

This preserves existing entity cleanup and presentation-adjacent ticking, but skips normal gameplay phases that would create new combat or spawn activity.

Asteroid spawning is skipped in this mode even when camera views exist.

## Protocols and APIs

Visibility and despawn do not expose a direct public API or packet surface.

They consume internal runtime state that comes from other service boundaries:

```text
game.cameraViews
game.entities.Asteroids
game.entities.Projectiles
game.entities.Players
runtime.ClientConfig
```

The relevant client-facing input is the generated `client_config` gameplay packet, but that packet is owned by player camera-view state and protocol/data docs. Visibility only consumes the resulting server-side camera-view dimensions.

The externally observable result is indirect:

```text
entity retained -> entity remains in StatePacket
entity removed -> entity disappears from StatePacket
pending despawn -> entity remains briefly, then disappears after delay
```

The server owns authority behind these outcomes. The client may report viewport size, but it does not decide entity retention, spawn visibility rejection, or removal timing.

## Data ownership

Visibility and despawn use transient runtime state only.

Stored runtime inputs include:

```text
game.cameraViews
game.entities.Asteroids
game.entities.Projectiles
game.entities.Players
```

Visibility and despawn do not persist data.

Relevant generated constants come from shared constant source files:

```text
shared/constants/server_entities.toml
```

Current relevant generated constants include:

```text
constants.AsteroidDespawnMargin
constants.CollisionDespawnDelay
```

Generated Go output lives in:

```text
services/game-server/internal/constants/constants.go
```

Generated constants should not be edited manually.

## Code map

Primary implementation files:

* `services/game-server/internal/game/visibility.go` - Owns camera-visibility helpers, offscreen spawn position selection, onscreen rejection, far-from-all-cameras checks, and toroidal-aware camera rectangle comparisons.
* `services/game-server/internal/game/simulation.go` - Owns top-level simulation phase order and match-over locked reduced stepping.
* `services/game-server/internal/game/simulation_asteroids.go` - Steps asteroids, removes ready pending-despawn asteroids, removes far asteroids, and gates asteroid spawning on camera-view presence.
* `services/game-server/internal/game/simulation_bullets.go` - Steps bullets, removes ready pending-despawn bullets, removes expired bullets, and removes far bullets.
* `services/game-server/internal/game/simulation_players.go` - Updates camera-view positions from active player ship positions during player stepping and removes ready pending-despawn players.
* `services/game-server/internal/game/runtime/asteroid.go` - Implements asteroid pending-despawn and removal readiness helpers.
* `services/game-server/internal/game/runtime/bullet.go` - Implements bullet pending-despawn, expiry, and removal readiness helpers.
* `services/game-server/internal/game/runtime/ship.go` - Implements ship pending-despawn and removal readiness helpers.
* `services/game-server/internal/game/runtime/camera.go` - Implements camera-view config, position, viewport-size fallback, and non-wrapped runtime rectangle helpers.
* `services/game-server/internal/game/runtime/state.go` - Defines runtime `CameraView`, `Asteroid`, `Bullet`, `Ship`, and `EntityStore` shapes.
* `services/game-server/internal/game/motion/motion.go` - Counts down pending-despawn delays and stops movement while entities are pending despawn.
* `services/game-server/internal/game/space/space.go` - Owns wrapped delta and world spatial helpers consumed by visibility checks.
* `services/game-server/internal/game/spawning.go` - Uses visibility spawn-position helpers before normalizing and applying asteroid spawn plans.
* `services/game-server/internal/game/combat.go` - Marks hit bullets and killed players pending despawn and preserves player camera-view position on fatal damage.
* `services/game-server/internal/game/asteroid_destruction.go` - Marks destroyed asteroids pending despawn and triggers fragments/pickup-drop consequences.
* `services/game-server/internal/game/state_packet.go` - Projects remaining entities into state packets until removal.

Source-of-truth and generated files:

* `shared/constants/server_entities.toml` - Source for despawn margin and collision despawn delay constants.
* `services/game-server/internal/constants/constants.go` - Generated Go constants consumed by visibility and despawn behavior.

Related tests:

* `services/game-server/tests/game/visibility_test.go` - Verifies cross-edge camera visibility and far bullet/asteroid cleanup.
* `services/game-server/tests/game/collision_test.go` - Verifies delayed despawn for hit bullets, destroyed asteroids, and killed players.
* `services/game-server/tests/game/spawning_test.go` - Verifies camera-view-based asteroid spawning and offscreen placement behavior.
* `services/game-server/tests/game/movement_test.go` - Verifies world movement and wrap behavior that visibility depends on.
* `services/game-server/tests/game/state_packet_lifecycle_test.go` - Verifies player lifecycle projection around active and inactive states.
* `services/game-server/internal/game/simulation_match_over_test.go` - Verifies match-over stepping does not continue normal spawn behavior.
* `services/game-server/tests/game/helpers_test.go` - Provides camera-view, asteroid, bullet, and pending-despawn helpers used by gameplay tests.

Important non-ownership boundaries:

* `services/game-server/internal/game/players.go` owns normal camera-view creation/removal during player add/remove.
* `services/game-server/internal/game/input.go` owns accepting `client_config` packets and applying valid viewport dimensions to sessions, camera views, and active ships.
* `services/game-server/internal/game/player_world_state.go` owns inactive-player world-state fallback from camera views.
* `services/game-server/internal/game/collisions.go` owns collision fact detection, not entity retention.
* `services/game-server/internal/game/damage/` owns pure damage resolution, not removal timing.
* `client/scripts/world/` owns client-side entity node creation/removal and visual presentation after state packets are received.
* `client/scripts/config/client_viewport_config_flow.gd` owns client viewport measurement and sending config packets, not server visibility policy.

## Tests and verification

Focused verification should cover:

```text
go test ./services/game-server/internal/game/... -buildvcs=false
go test ./services/game-server/tests/game/... -buildvcs=false
```

Relevant behavior to preserve:

* asteroid spawn positions are rejected when onscreen for any camera view
* asteroid spawning does not accumulate when there are no camera views
* asteroids across a wrapped world edge remain near a camera view
* bullets across a wrapped world edge remain near a camera view
* asteroids far from all camera views are removed
* bullets far from all camera views are removed
* bullets expire by lifetime even if camera visibility would otherwise retain them
* hit bullets remain briefly during collision despawn delay
* destroyed asteroids remain briefly during collision despawn delay
* killed player ships remain briefly during collision despawn delay
* pending-despawn entities stop moving while their delay counts down
* pending-despawn entities are removed after their delay reaches zero
* match-over locked stepping does not continue normal asteroid spawning
* visibility checks use wrapped deltas rather than raw coordinate distance

## Related docs

* [Game Server Simulation World](./!README.md)
* [Game Server Simulation](../!README.md)
* [Game Server](../../!README.md)
* [Services](../../../!README.md)
* [Player Camera View State](../players/player-camera-view-state.md)
* [Player Death And Despawn](../players/player-death-and-despawn.md)
* [Active Player Avatar State](../players/active-player-avatar-state.md)
* [Player Respawn](../players/player-respawn.md)
* [Collision To Damage Flow](../combat/collision-to-damage-flow.md)
* [Runtime Entity Store](../runtime/runtime-entity-store.md)
* [Simulation Loop And Phase Order](../runtime/simulation-loop-and-phase-order.md)
* [State Packet Projection](../runtime/state-packet-projection.md)
* [Client Viewport Config Flow](../../../client/app-shell-and-session/client-viewport-config-flow.md)
* [World Sync Coordinator](../../../client/world-sync/world-sync-coordinator.md)
* [View Anchor And Visual Coordinates](../../../client/world-sync/view-anchor-and-visual-coordinates.md)
* [Systems Design World](../../../../systems-design/world/!README.md)
* [Protocol](../../../../protocol/!README.md)
* [Data](../../../../data/!README.md)

## Notes

`runtime.CameraView.IsInside` and `runtime.CameraView.IsFarFrom` perform non-wrapped rectangle checks on the runtime type. Current game-level visibility/despawn behavior uses `visibility.go` helpers instead, because those helpers use `space.Delta` for toroidal-aware comparisons.

The current despawn-margin constant is named `AsteroidDespawnMargin`, but the same margin is used for asteroid and bullet far-from-camera cleanup.

Pending despawn is not the same as far-from-camera despawn. Pending despawn is a short delayed removal state after gameplay consequences. Far-from-camera cleanup removes bullets and asteroids because no player view should still need them.
