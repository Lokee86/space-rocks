---
author: brian
created: "2026-07-19"
document_id: 019f7d55-fb2c-7c74-b210-34c7763a0566
document_type: general
policy_exempt: false
summary: This document describes the server-owned game.Control adapter used by internal/devtools.
---
# Game Control Devtools Adapter

Parent index: [Server](./!INDEX.md)

## Purpose

This document describes the server-owned `game.Control` adapter used by `internal/devtools`.

## Overview

`services/game-server/internal/devtools/` owns debug command dispatch, target capability interfaces, target selection, stream-runtime wiring, observer registration policy, status projection, and debug-facing DTOs.

`services/game-server/internal/game/` owns authoritative gameplay state. `game.NewControl` wraps a `*game.Game` and exposes narrow domain-neutral capabilities through focused `control_*.go` files. `internal/game` does not import `internal/devtools`.

The current execution boundary is:

```text
client devtools input
-> existing debug packet
-> networking inbound classification and decode
-> game.NewControl(room.GameInstance())
-> devtools.NewController(...)
-> capability-specific devtools handler
-> game.Control method
-> authoritative game state
-> normal gameplay readback or debug output
```

The dependency direction is intentional:

```text
internal/networking imports internal/devtools and internal/game
internal/devtools defines capability interfaces but does not import root internal/game
internal/game implements those capabilities without importing internal/devtools
```

`game.Control` is an adapter over authoritative state. It is not a second gameplay implementation and does not own debug command policy.

## Client presentation

Client devtools presentation is separate from server authority.

The client can send debug command packets and render debug outputs, but it does not apply authoritative gameplay effects locally. Visible confirmation comes from lane-native readback, debug status packets, debug shape catalog packets, entity sync, or the absence/presence of entities after the authoritative server update.

The server-facing outputs tied to these seams include:

```text
lane-native gameplay readback
debug_status packets
debug_shape_catalog packets
server logs
```

`debug_status` reflects server-owned devtools state such as invincibility, infinite lives, world freeze, granular freeze flags, and player freeze.

`debug_shape_catalog` provides shape definitions for client-side hitbox presentation. Shape catalog output is built from the physics collision shape catalog, not from client-authored gameplay state.

Collision-body telemetry is exposed as a game-owned read seam. It should remain a diagnostic data surface; drawing, overlay lifecycle, and toggle UI belong to client devtools.

## Commands and controls

Current generated debug command packet types include:

```text
toggle_debug_invincible
toggle_debug_infinite_lives
toggle_debug_freeze_world
toggle_debug_freeze_player
debug_kill_player
debug_spawn_entity
debug_spawn_pickup
debug_begin_continuous_bullet_stream
debug_respawn_player
debug_set_score
debug_add_score
debug_set_lives
debug_add_lives
debug_clear_bullets
debug_clear_asteroids
```

The server command path is:

```text
networking read loop
-> DecodeClientPacketEnvelope
-> inbound.RouteClientPacket
-> sr.tooling packet policy, room, and capability preflight
-> packetcodec.Decode into devtools.DebugCommand
-> Controller.HandleCommand
-> command-specific handler
-> Control capability or existing game API
```

Devtools command packets are routed before normal gameplay packet decoding. They do not reach `Game.HandlePacket`.

Player-targeted commands use `target_player_id` or `target_scope`. The current all-player scope value is:

```text
all_players
```

All-player toggle behavior for invincibility, infinite lives, and player freeze uses set-style semantics:

```text
if any eligible target is inactive -> enable all eligible targets
if every eligible target is active -> disable all eligible targets
```

Respawn all-player behavior still applies normal respawn eligibility guards per target. Active players are ignored.

World freeze is room/global and does not use a player selector. Granular freeze commands use `freeze_target` values handled by the devtools toggle handler:

```text
all
asteroids
bullets
spawning
spawns
collisions
```

Unknown freeze targets are logged and ignored without changing freeze flags.

## Status

`control_status.go` and `control_match.go` own the adapter-side read paths for debug status.

`StatusFor` projects authoritative state into `devtools.DebugStatus` without changing packet shape.

`Control.TargetPlayerIDs()` is command fanout only. `Controller.StatusesForAllPlayers()` uses `MatchDecision().Players` as the membership source so the all-player status readout stays session-backed.

## Toggle

`control_toggles.go` owns the adapter methods for world freeze, granular freeze, invincibility, infinite lives, player freeze, and debug kill.

Control methods are domain-neutral and omit `Devtools` prefixes:

```text
WorldFrozen
SetWorldFrozen
ToggleFreezeWorld
ToggleFreezeAsteroids
ToggleFreezeBullets
ToggleFreezeSpawning
ToggleFreezeCollisions
PlayerInvincible
SetPlayerInvincible
InfiniteLives
SetInfiniteLives
PlayerFrozen
SetPlayerFrozen
ApplyPlayerDefeat
```

`ApplyPlayerDefeat` is the adapter name for the unchanged debug damage and fatal-damage behavior.

## Spawn

`control_spawn.go` owns debug bullet, asteroid, and pickup spawn adapter methods.

Control methods are domain-neutral and omit `Devtools` prefixes:

```text
RandomUnitVector
NextBulletID
AddBullet
SpawnBullet
RandomAsteroidSpeed
ApplyAsteroidSpawnPlan
SpawnPickup
```

`SpawnPickup` remains the current authoritative pickup path used by debug pickup spawning.

## Player spawn

`control_player_spawn.go` owns debug player session and ship setup.

Control methods are domain-neutral and omit `Devtools` prefixes:

```text
EnsurePlayerSession
SpawnPlayerShip
PlayerIDOccupied
ReservePlayerID
TargetPlayerIDs
```

`TargetPlayerIDs` is the sorted union of session IDs and active ship IDs used by command fanout.

`PlayerIDOccupied` validates requested IDs with the current player-ID normalization rules. `ReservePlayerID` rejects invalid or occupied IDs and advances `game.nextID` to the reserved number when needed so later normal allocation cannot reuse the reservation.

`SpawnPlayerShip` preserves the current debug-spawn camera-view behavior and uses the supplied camera config only when it is valid.

## Respawn

`control_respawn.go` owns debug respawn support.

Control methods are domain-neutral and omit `Devtools` prefixes:

```text
SafeRespawnPosition
ForceRespawnPlayer
```

Debug respawn keeps the existing player-session and ship creation flow and still rejects active players before forcing a respawn.

## Counter

`control_counters.go` owns debug score and lives adapter methods.

Control methods are domain-neutral and omit `Devtools` prefixes:

```text
SetPlayerScore
AddPlayerScore
SetPlayerLives
AddPlayerLives
```

These methods return only whether the underlying `PlayerCounterChange` found a player; they do not change the counter semantics.

## Clear

`control_clear.go` owns debug clear-bullets and clear-asteroids behavior.

Control methods are domain-neutral and omit `Devtools` prefixes:

```text
ClearBullets
ClearAsteroids
```

These methods operate on authoritative game state. Command policy still lives in `internal/devtools`.

## Stream

`control_streams.go` owns debug continuous-bullet stream adapter methods.

Control methods are domain-neutral and omit `Devtools` prefixes:

```text
BulletsCanMove
SpawnDebugBullet
RegisterSimulationStepObserver
```

The injected stream runtime owns the continuous stream lifecycle for the controller that receives it.

## Collision telemetry

`control_collision_telemetry.go` owns the authoritative collision-body snapshot.

The adapter method uses the current name without a `Devtools` prefix:

```text
CollisionBodiesByKind
```

`CollisionBodiesByKind` returns raw physics bodies grouped by kind. `devtools.CollisionBodies` owns the JSON projection for debug output.

## Build and runtime gates

Server devtools include build-tag gate helpers:

```text
services/game-server/internal/devtools/enabled_default.go
services/game-server/internal/devtools/enabled_nodevtools.go
services/game-server/internal/devtools/disabled.go
```

Default builds return `true` from `devtools.Enabled()`. Builds with the `nodevtools` tag return `false`.

Current outbound debug status and debug shape catalog sending checks `devtools.Enabled()` before sending debug output.

The devtools package also exposes:

```text
ShouldHandleCommand(packetType string) bool
```

That helper combines command-type classification with `devtools.Enabled()` and has default and `nodevtools` tests.

The current inbound router classifies devtools packets in:

```text
```

It uses route-local packet-type classifiers for simple commands, placement commands, and remaining devtools commands before decoding normal gameplay packets. That route also requires a current room and a non-empty current game player ID before dispatching a command to `Controller.HandleCommand`.

When changing devtools gates, keep these surfaces aligned:

```text
devtools.Enabled
devtools.ShouldHandleCommand
networking/tooling policy and capability preflight
networking outbound debug status/catalog sending
```

Runtime gates include:

```text
current room must exist
current game player ID must exist
command type must be recognized by the tooling packet-policy registry
command payload must decode into devtools.DebugCommand
handler-specific target and payload checks must pass
Control capability or public game method must accept the operation
```

## Locking and mutation model

Every externally callable `game.Control` method that reads or mutates aggregate state acquires `Game.mu` itself. This makes the adapter a uniform aggregate-lock boundary for devtools and other callers.

Successful presentation-visible Control mutations publish a replacement immutable presentation frame before releasing `Game.mu`. This lets devtools command handlers and tests observe authoritative lane/readback state immediately without waiting for the next simulation tick. The affected mutation families include counters, clear operations, entity/player spawn and forced respawn, direct debug-bullet stream spawn, and player defeat.

Read-only Control methods, failed or no-op mutations, world/toggle state that is not part of the gameplay presentation frame, reservation/occupancy queries, and observer registration do not republish merely because they were called.

Control methods do not call public `Game` or `Control` methods while holding that lock when those methods would lock again. They use the existing lock-assuming helpers, such as `matchDecisionLocked`, counter helpers, and `spawnPickupLocked`.

`ObserverKey` is the sole lock-free exception: it returns the immutable `*Game` identity used only for observer-registry deduplication and reads no mutable aggregate state.

`Game.Step` invokes simulation observers while `Game.mu` is held. `RegisterSimulationStepObserver` wraps each registered callback and supplies narrow capability closures for bullet movement and debug-bullet spawning only during that callback. The closures call lock-assuming game helpers inside the already-locked simulation phase; ordinary callers only see the locking `BulletsCanMove` and `SpawnDebugBullet` methods. This preserves phase ordering and avoids recursive locking without exposing a lock-assuming Control method.

`RegisterSimulationStepObserver` acquires `Game.mu` before changing the observer list. Observer callbacks should stay narrow and should route gameplay effects through the lock-assuming observer seam only when invoked by `Game.Step`.

Observer-triggered gameplay mutations during `Game.Step` are included in the single end-of-step presentation-frame publication rather than causing an extra publication inside the observer callback.

## Relationship to real gameplay systems

Control capabilities exist to expose real gameplay systems to debug tooling. They should not create replacement systems.

Current examples:

* Debug kill uses the damage resolver and fatal-player damage path.
* Invincibility changes damage options consumed by collision/damage behavior.
* Infinite lives changes session life options consumed by death/lives behavior.
* Player freeze changes suspension state consumed by movement, input, shooting, and collision capability checks.
* World freeze changes `WorldSimulationOptions` consumed by simulation phase gates.
* Score and lives commands use shared player counter mutation.
* Debug asteroid spawn applies a normal asteroid spawn plan through the game aggregate.
* Debug bullet spawn uses the game-owned debug bullet spawn helper.
* Debug pickup spawn uses the game-owned pickup spawn API.
* Debug respawn creates a new ship from the player session and updates camera view state.
* Collision telemetry reads server collision bodies from runtime entities and collision shapes.

The rule is:

```text
devtools may choose when to request a debug action
game-owned systems decide what that action actually does
```

## Code map

Game Control adapter files:

```text
services/game-server/internal/game/control.go
services/game-server/internal/game/control_match.go
services/game-server/internal/game/control_status.go
services/game-server/internal/game/control_toggles.go
services/game-server/internal/game/control_spawn.go
services/game-server/internal/game/control_player_spawn.go
services/game-server/internal/game/control_respawn.go
services/game-server/internal/game/control_counters.go
services/game-server/internal/game/control_clear.go
services/game-server/internal/game/control_streams.go
services/game-server/internal/game/control_collision_telemetry.go
```

Server devtools command files:

```text
services/game-server/internal/devtools/handler.go
services/game-server/internal/devtools/command_types.go
services/game-server/internal/devtools/packets_generated.go
services/game-server/internal/devtools/toggles.go
services/game-server/internal/devtools/spawn_entity.go
services/game-server/internal/devtools/spawn_asteroid.go
services/game-server/internal/devtools/spawn_bullet.go
services/game-server/internal/devtools/spawn_pickup.go
services/game-server/internal/devtools/spawn_player.go
services/game-server/internal/devtools/respawn_player.go
services/game-server/internal/devtools/player_counters.go
services/game-server/internal/devtools/clear_entities.go
services/game-server/internal/devtools/continuous_bullet_stream.go
services/game-server/internal/devtools/collision_telemetry.go
```

## Tests and verification

Focused tests live beside the adapter and devtools handlers they cover. This document does not define new behavior; it records the current boundary and the files that exercise it.

## Related docs

* `docs/devtools/server/debug-command-surface.md`
* `docs/devtools/server/command-routing-and-build-gates.md`
* `docs/devtools/design/devtools-packet-protocol.md`

## Notes

`game.Control` adapts authoritative game state. `internal/devtools` owns command policy, projection, and DTOs. `internal/game` owns gameplay consequences. `internal/networking` owns packet routing and `Controller` construction.
