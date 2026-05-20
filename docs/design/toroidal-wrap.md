# Toroidal Wrap Plan

This is a future implementation plan. The current game still behaves as a flat/infinite coordinate space, but recent prep work added `services/game-server/internal/game/space` so future spatial behavior can be centralized.

## Goal

Keep server coordinates bounded while making the world feel visually endless.

Desired behavior:

- server stores positions inside a wrapped playfield
- players can fly continuously without obvious edge transitions
- multiplayer players remain in one shared arena instead of drifting infinitely apart
- spawning, visibility, respawn safety, and collisions work across wrapped edges
- the client can render a camera that visually straddles the edge with no seam

## Current Prep Already Done

Server spatial helpers now live in:

```text
services/game-server/internal/game/space/space.go
```

Current behavior is still flat:

- `Delta` uses normal flat delta
- `Distance` uses flat distance
- `Direction` uses normalized flat delta
- `NormalizePosition` is a no-op

Existing code has started moving coordinate-sensitive logic through this package:

- asteroid spawn direction uses `space.Direction`
- visibility checks use `space.Delta`
- respawn clearance uses `space.Distance`

## Execution Plan

### 1. Add Shared World Bounds

Add constants to:

```text
shared/game_data.toml
```

Proposed values:

```json
WORLD_WRAP_WIDTH: 1062
WORLD_WRAP_HEIGHT: 5250
```

Then regenerate:

```bash
python3 tools/data_sync/main.py -push -constants -go -gds
```

TODO: confirm `1062` width is intentional. It is very narrow compared to `5250`.

### 2. Extend `services/game-server/internal/game/space`

Add:

```go
type Bounds struct {
	Width  float64
	Height float64
}

func DefaultBounds() Bounds
func WrapPosition(pos physics.Vector2, bounds Bounds) physics.Vector2
func ShortestDelta(from physics.Vector2, to physics.Vector2, bounds Bounds) physics.Vector2
func WrappedDistance(from physics.Vector2, to physics.Vector2, bounds Bounds) float64
```

Then update the existing helpers to route through the wrapped implementation when ready:

```go
func Delta(from, to physics.Vector2) physics.Vector2
func Distance(from, to physics.Vector2) float64
func Direction(from, to physics.Vector2) physics.Vector2
func NormalizePosition(position physics.Vector2) physics.Vector2
```

### 3. Add Wrap Math Tests

Extend:

```text
services/game-server/internal/game/space/space_test.go
```

Test cases:

- position wraps right to left
- position wraps left to right
- position wraps top to bottom
- position wraps bottom to top
- shortest delta crosses horizontal edge
- shortest delta crosses vertical edge
- shortest delta stays direct when direct path is shorter
- `NormalizePosition` wraps into bounds once wrapping is enabled

### 4. Wrap Moving Entities Centrally

Prefer central wrapping in:

```text
services/game-server/internal/game/game.go
```

After movement, normalize positions for:

- players
- asteroids
- bullets

Central wrapping is less invasive than putting wrapping directly inside each entity method at first.

### 5. Make Spawning Wrap-Aware

Touch:

```text
services/game-server/internal/game/spawning.go
services/game-server/internal/game/visibility.go
```

Behavior:

- generate spawn position around the camera as today
- call `space.NormalizePosition(spawn)` before storing the asteroid
- ensure `space.Direction(spawn, targetPosition)` uses shortest wrapped delta once wrapping is active
- allow an asteroid spawned past one edge to be stored on the opposite side

### 6. Make Visibility And Despawn Wrap-Aware

Touch:

```text
services/game-server/internal/game/visibility.go
```

The visibility code already uses `space.Delta`. Once `space.Delta` becomes wrap-aware, camera checks should naturally support wrapped edges.

Required behavior:

- asteroid across world edge is still near a camera if it is visually nearby
- bullet across world edge is still near a camera if it is visually nearby
- objects far from all cameras still despawn

### 7. Make Respawn Safety Wrap-Aware

Touch:

```text
services/game-server/internal/game/session.go
```

Respawn clearance already uses `space.Distance`. Once `space.Distance` becomes wrapped, respawn safety should treat objects across the edge as nearby.

Add tests for:

- spawn point near opposite edge is unsafe if an asteroid is close across the wrap
- spawn point near opposite edge is unsafe if a player is close across the wrap
- spawn search still finds a safe point

### 8. Make Collisions Wrap-Aware

Touch:

```text
services/game-server/internal/game/combat.go
```

Build temporary collision bodies in local wrapped space.

For ship/asteroid:

```go
delta := space.Delta(ship.Position(), asteroid.Position())
asteroidBody.Position = ship.Position().Add(delta)
```

For bullet/asteroid:

```go
delta := space.Delta(bullet.Position(), asteroid.Position())
asteroidBody.Position = bullet.Position().Add(delta)
```

This avoids needing duplicate ghost bodies at first.

### 9. Add Server Gameplay Tests

Add focused tests for:

- player crossing right edge wraps to left server coordinate
- player crossing left edge wraps to right server coordinate
- asteroid crossing edge wraps
- bullet crossing edge wraps
- asteroid near opposite edge does not despawn if camera is near boundary
- bullet near opposite edge does not despawn if camera is near boundary
- bullet/asteroid collision works across boundary
- ship/asteroid collision works across boundary
- asteroid spawning near boundary wraps into world bounds

Run:

```bash
cd services/game-server
env GOCACHE=/tmp/space-rocks-go-build go test -buildvcs=false ./...
```

### 10. Add Client Wrap Math

Create:

```text
client/scripts/world_wrap.gd
```

Add helpers matching server behavior:

```gdscript
static func wrap_position(pos: Vector2) -> Vector2
static func shortest_delta(from: Vector2, to: Vector2) -> Vector2
static func visual_position_relative_to(reference: Vector2, target: Vector2) -> Vector2
```

Use generated constants from:

```text
client/scripts/constants/constants.gd
```

### 11. Track Local Visual Continuity

Touch:

```text
client/scripts/game.gd
client/scripts/networking/world_sync.gd
```

Maintain separate server and visual positions:

```gdscript
var local_server_position: Vector2
var local_visual_position: Vector2
```

On state update:

```gdscript
var new_server_position := Vector2(...)
local_visual_position += WorldWrap.shortest_delta(local_server_position, new_server_position)
local_server_position = new_server_position
```

This prevents the camera/background from snapping when the server position wraps.

### 12. Render Objects Relative To Local Player

Touch:

```text
client/scripts/networking/world_sync.gd
```

For remote players, asteroids, and bullets:

```gdscript
visual_position = local_visual_position + WorldWrap.shortest_delta(local_server_position, entity_server_position)
```

This lets the camera visually straddle the invisible edge.

### 13. Preserve Background Continuity

Touch:

```text
client/scripts/game.gd
client/scripts/ui/game_shell.gd
```

Background and camera offset should use the client visual position, not raw wrapped server position.

### 14. Document Final Architecture

Update:

```text
docs/design/architecture.md
docs/notes.md
docs/developer.md
```

Document:

- `services/game-server/internal/game/space` owns gameplay spatial math
- server stores bounded wrapped coordinates
- client renders continuous visual coordinates
- camera can straddle world edges invisibly
- spawning, visibility, respawn safety, and collisions use wrap-aware helpers

### 15. Smoke Test

Manual checks:

- fly straight for several minutes
- cross all four edges without visual snap
- shoot across an edge
- asteroid crosses edge naturally
- asteroid spawns near edge naturally
- collide with asteroid across boundary
- multiplayer clients near opposite server edges see each other nearby
- background does not jump

## Preferred Implementation Order

1. Constants.
2. Extend `services/game-server/internal/game/space`.
3. Add/expand `space` tests.
4. Server movement wrapping in `Game.Step()`.
5. Server visibility/spawning tests.
6. Server collision wrap changes.
7. Client `world_wrap.gd`.
8. Client visual-relative world sync.
9. Client camera/background continuity.
10. Docs and smoke testing.

## Design Caution

The risky part is not the wrap math by itself. The risky part is keeping server authoritative wrapped positions separate from client visual continuity. Treat those as two different coordinate layers.
