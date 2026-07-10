## Space Rocks Devtools Extraction Skill

Use this skill when refactoring Space Rocks server devtools ownership out of `internal/game`.

This skill applies to the cleanup item that removes `Devtools*` methods from the Go `Game` struct and moves devtools command/status/spawn/telemetry ownership behind DI in `services/game-server/internal/devtools`.

Do not use this skill for the later devtools network-lane migration. That is a separate task.

## Goal

Move devtools ownership out of `internal/game` without changing gameplay behavior, packet schemas, packet transport, or realtime scheduling.

The target end state is:

```text
internal/devtools
  owns Controller, target capability interfaces, command dispatch, debug DTO shaping,
  target selection, stream runtime wiring, and observer registry wiring

internal/game
  owns authoritative state and exposes it through Control

internal/networking
  decodes and routes existing packets as before, but calls a devtools.Controller
  instead of calling devtools package functions with a game.Control target
```

## Hard rules

* Do not change devtools packet transport in this pass.
* Do not add a WebRTC/UDP devtools lane in this pass.
* Do not edit the realtime scheduler.
* Do not change packet schemas.
* Do not redesign locking.
* Do not change gameplay behavior.
* Do not leave public `func (game *Game) Devtools...` methods behind.
* Production files in `services/game-server/internal/devtools` must not import `github.com/Lokee86/space-rocks/server/internal/game`.
* Tests may import `internal/game` for integration coverage.
* Keep files small and focused.
* Preserve current behavior unless a test proves current behavior cannot be preserved.

## Required shape

Create devtools-owned target interfaces in `services/game-server/internal/devtools/target.go`.

Prefer small capability interfaces:

```go
type StatusTarget interface { ... }
type ToggleTarget interface { ... }
type PlayerTarget interface { ... }
type SpawnTarget interface { ... }
type RespawnTarget interface { ... }
type CounterTarget interface { ... }
type ClearTarget interface { ... }
type CollisionTelemetryTarget interface { ... }
type StreamTarget interface { ... }
```

Compose them:

```go
type Target interface {
    StatusTarget
    ToggleTarget
    PlayerTarget
    SpawnTarget
    RespawnTarget
    CounterTarget
    ClearTarget
    CollisionTelemetryTarget
    StreamTarget
}
```

Use domain-neutral method names where practical:

```go
WorldFrozen() bool
ToggleFreezeWorld() bool
PlayerInvincible(playerID string) (bool, bool)
SetPlayerInvincible(playerID string, enabled bool) bool
KillPlayer(sourcePlayerID string, targetPlayerID string) bool
ClearBullets() int
ClearAsteroids() int
```

Avoid keeping the smell under a new name:

```go
DevtoolsWorldFrozen()
DevtoolsKillPlayer()
DevtoolsSpawnPlayerShip()
```

## Game adapter

Add a game-owned adapter in `services/game-server/internal/game`.

Suggested files:

```text
control.go
control_*.go
```

Suggested shape:

```go
type Control struct {
    game *Game
}

func NewControl(game *Game) *Control {
    return &Control{game: game}
}
```

The adapter may touch private `Game` fields because it lives in `internal/game`.

The adapter should preserve existing lock behavior. Do not use this refactor to redesign locking.

All-player command fanout must use `Control.TargetPlayerIDs()`.

## Devtools controller

Add `services/game-server/internal/devtools/controller.go`.

Suggested shape:

```go
type Dependencies struct {
    Target Target
    Streams *streamruntime.Runtime
    ObserverRegistry *ObserverRegistry
}

type Controller struct {
    target Target
    streams *streamruntime.Runtime
    observerRegistry *ObserverRegistry
}

func NewController(deps Dependencies) *Controller
func (controller *Controller) HandleCommand(playerID string, command DebugCommand) bool
func (controller *Controller) StatusFor(playerID string) DebugStatus
func (controller *Controller) StatusesForAllPlayers() map[string]DebugStatus
```

`Controller.StatusesForAllPlayers()` must preserve the pre-refactor session-backed membership from `MatchDecision().Players`.

Player-ID reservation must:

- accept only valid positive `player-N` IDs
- reject occupied IDs
- advance `game.nextID` so later normal player allocation cannot reuse the reservation

The controller-injected stream runtime owns the begin, step, and clear operations for that same controller.

Convert package-level command handling from:

```go
func HandleCommand(target *game.Game, playerID string, command DebugCommand) bool
```

to:

```go
func (controller *Controller) HandleCommand(playerID string, command DebugCommand) bool
```

## Continuous bullet stream

Do not leave continuous stream runtime keyed directly by `*game.Game` inside `internal/devtools`.

Use an opaque observer key instead:

```go
type StreamTarget interface {
    ObserverKey() any
    BulletsCanMove() bool
    SpawnDebugBullet(ownerPlayerID string, origin physics.Vector2, direction physics.Vector2) bool
    RegisterSimulationStepObserver(observer func(float64))
}
```

The game adapter may return the underlying `*Game` from `ObserverKey()` as `any`.

Devtools must not import `internal/game` to key this registry.

The refactor must preserve the old synchronization behavior:

- clear-bullet and clear-asteroid operations lock the game aggregate
- collision-body snapshot collection locks the game aggregate
- simulation observer registration must not introduce a new nested lock around the existing observer append

The refactor must preserve the old camera-config behavior for debug player spawning.

## Migration order

1. Add devtools target interfaces.
2. Add `game.Control` adapter.
3. Add `devtools.Controller`.
4. Convert `handler.go` to controller methods.
5. Convert continuous bullet stream observer registration to use an opaque target key.
6. Convert status handling off `*game.Game`.
7. Convert target-player resolution off `*game.Game`.
8. Convert toggles off `*game.Game`.
9. Convert counters off `*game.Game`.
10. Convert clear-entity commands off `*game.Game`.
11. Convert spawn/respawn commands off `*game.Game`.
12. Convert collision telemetry formatting so devtools owns debug DTOs.
13. Update networking call sites to construct/use `devtools.Controller`.
14. Move or rewrite tests.
15. Delete old `Game.Devtools*` methods and `export_devtools_*` files.
16. Add guard coverage.

## Files expected to change

Likely remove or replace:

```text
services/game-server/internal/game/export_devtools.go
services/game-server/internal/game/export_devtools_status.go
services/game-server/internal/game/export_devtools_toggles.go
services/game-server/internal/game/export_devtools_spawn.go
services/game-server/internal/game/export_devtools_respawn.go
services/game-server/internal/game/export_devtools_player_spawn.go
services/game-server/internal/game/export_devtools_player_counters.go
services/game-server/internal/game/export_devtools_clear_entities.go
services/game-server/internal/game/export_devtools_streams.go
services/game-server/internal/game/export_devtools_collision_telemetry.go
```

Likely update:

```text
services/game-server/internal/devtools/handler.go
services/game-server/internal/devtools/status.go
services/game-server/internal/devtools/toggles.go
services/game-server/internal/devtools/target_player_ids.go
services/game-server/internal/devtools/spawn_player.go
services/game-server/internal/devtools/spawn_bullet.go
services/game-server/internal/devtools/spawn_asteroid.go
services/game-server/internal/devtools/spawn_pickup.go
services/game-server/internal/devtools/spawn_entity.go
services/game-server/internal/devtools/respawn_player.go
services/game-server/internal/devtools/respawn_handler.go
services/game-server/internal/devtools/player_counters.go
services/game-server/internal/devtools/clear_entities.go
services/game-server/internal/devtools/continuous_bullet_stream.go
services/game-server/internal/networking/inbound/devtools.go
services/game-server/internal/networking/outbound/debug_status_presentation.go
services/game-server/internal/networking/outbound/debug_shape_catalog_presentation.go
```

Likely add:

```text
services/game-server/internal/devtools/target.go
services/game-server/internal/devtools/controller.go
services/game-server/internal/devtools/observer_registry.go
services/game-server/internal/game/control.go
services/game-server/internal/game/control_*.go
```

## DTO ownership

Move debug DTO ownership to `internal/devtools` where possible.

Good devtools-owned names:

```text
StatusSnapshot
CollisionPoint
CollisionBody
```

Avoid making `internal/game` import `internal/devtools` only to construct debug presentation objects.

The game adapter can return raw authoritative or physics data. Devtools should shape that data into debug packets.

## Networking rule

Keep transport unchanged.

Existing WebSocket devtools packets should still be decoded and routed as before.

Only change the execution target:

```go
control := game.NewControl(room.GameInstance())
controller := devtools.NewController(devtools.Dependencies{
    Target: control,
})
controller.HandleCommand(playerID, command)
```

Do not add `sr.devtools`.
Do not move debug status/catalog to WebRTC.
Do not touch channel negotiation.

## Tests and guard checks

Add or preserve coverage for:

```text
- no production internal/devtools file imports internal/game
- no legacy game export files remain
- no func (game *Game) Devtools... remains
- a Control anchor exists with focused control_*.go files
- controller and target-interface files exist
- devtools commands still execute
- debug status payload stays compatible
- debug shape catalog payload stays compatible
- collision telemetry JSON stays compatible
- continuous bullet stream still registers one observer per game target
- target-player commands still resolve the same player IDs
```

## Verification command

Use this after the refactor:

```bash
cd /mnt/d/\!bin/space-rocks
{
  cd services/game-server
  go test ./internal/game/... ./internal/devtools/... ./internal/networking/...
  echo
  grep -R "func (game \*Game) Devtools" internal/game || true
  echo
  grep -R "\"github.com/Lokee86/space-rocks/server/internal/game\"" internal/devtools --include='*.go' || true
} 2>&1 | tee /dev/tty | clip.exe
```

Expected grep result after completion:

```text
- no Game.Devtools* methods
- no production internal/devtools imports of root internal/game
```

Test files under `internal/devtools` may still import `internal/game` if they are integration-style tests.

## Stop conditions

Stop and report if:

```text
- the task requires changing devtools packet transport
- the task requires changing realtime lane scheduling
- the task requires changing packet schemas
- preserving behavior requires a locking redesign
- the change grows into unrelated cleanup
- production internal/devtools still needs to import internal/game after interface extraction
```

Report:

```text
- changed files
- deleted files
- remaining Game.Devtools references, if any
- remaining production devtools imports of internal/game, if any
- tests run
```
