# Toroidal World Wrap

Space Rocks now uses a bounded toroidal playfield on the server and continuous visual coordinates on the client.

The design goal is simple: server coordinates stay inside one shared arena, while the client renders motion across edges without an obvious seam.

## World Bounds

World size is sourced from:

```text
shared/constants/server_constants.toml
```

Current generated Go constants:

```go
constants.WorldWidth
constants.WorldHeight
```

Current generated GDScript constants:

```gdscript
Constants.WORLD_WIDTH
Constants.WORLD_HEIGHT
```

When world size changes, regenerate both Go and GDScript outputs:

```bash
python3 tools/data_sync/main.py -push -constants -go -gds
```

The client wrap helper depends on generated GDScript world constants. If only Go constants are regenerated, the server and client can disagree about wrap size.

## Server Model

Server simulation stores bounded wrapped positions. Gameplay spatial math is centralized in:

```text
services/game-server/internal/game/space/space.go
```

The package owns:

- `Bounds`
- `DefaultBounds`
- `WrapPosition`
- `ShortestDelta`
- `WrappedDistance`
- `Delta`
- `Distance`
- `Direction`
- `NormalizePosition`

`Game.Step` chooses the active world bounds with `space.DefaultBounds()`, but per-entity step-and-wrap behavior lives in `services/game-server/internal/game/motion`. `motion.AdvanceShip`, `motion.AdvanceAsteroid`, and `motion.AdvanceBullet` each step one entity and wrap it with `space.WrapPosition`. `Game.Step` still owns the entity map loops, world-devtool gates, camera updates, deletion/despawn checks, spawning, collisions, scoring, and lifecycle order.

Server systems that rely on spatial relationships use the `space` package:

- asteroid aim direction uses wrapped `space.Direction`
- visibility/despawn uses wrapped `space.Delta`
- respawn safety uses wrapped `space.Distance`
- collision checks place temporary collision bodies in wrapped-local space before testing

Stored entity coordinates are not duplicated as ghost bodies. For collision checks, the server moves only the temporary collision body near the actor being tested.

## Client Model

The Godot client keeps two coordinate layers:

- server positions: bounded authoritative coordinates from state packets
- visual positions: continuous positions used for rendering and camera/background continuity

Client wrap math lives in:

```text
client/scripts/world/world_wrap.gd
```

`local_visual_sync.gd` tracks:

```gdscript
local_server_position
local_visual_position
```

On each local-player state update, `client/scripts/world/world_sync.gd` routes the authoritative position through `LocalVisualSync`, which advances visual position by the shortest wrapped delta from the previous server position to the new server position. This prevents the local player, camera, and background from snapping when the server coordinate wraps.

Remote players, asteroids, bullets, and server-driven effects render relative to the local player. Entity ownership lives in `client/scripts/world/player_sync.gd`, `client/scripts/world/asteroid_sync.gd`, and `client/scripts/world/bullet_sync.gd`, while `client/scripts/world/world_sync.gd` coordinates update order:

```gdscript
visual_position = local_visual_position + WorldWrap.shortest_delta(
	local_server_position,
	entity_server_position
)
```

The camera is a child of the local player node, so it follows the continuous local visual position. Background scroll follows the local player node as well.

## Tests

Server tests live under:

```text
services/game-server/tests/
```

Relevant areas:

- `tests/space`: wrap math helpers
- `tests/game`: movement wrapping, spawning, visibility/despawn, respawn safety, and cross-edge collisions

Client tests live under:

```text
client/tests/unit/
```

Relevant coverage:

- `test_world_wrap.gd`: client wrap math
- `test_world_sync.gd`: visual-relative coordination and entity sync delegation
- `test_local_visual_sync.gd`: local visual/server position continuity

## Manual Smoke Checklist

Before calling wrap behavior complete after a visual or balance change, manually verify:

- fly across right edge
- fly across left edge
- fly across top edge
- fly across bottom edge
- camera does not snap
- background does not jump
- bullets cross edges naturally
- asteroids cross edges naturally
- asteroids spawn near edges naturally
- visibility/despawn behaves near edges
- bullet/asteroid collision works across edges
- ship/asteroid collision works across edges
- respawn safety detects danger across edges
- multiplayer clients near opposite server edges see each other nearby
- bullet blast and ship death effects spawn near visible entity positions

## Known Follow-Ups

- Full gameplay/network smoke testing is still manual.
- Consider whether current world dimensions are final for gameplay balance.
- Vertical despawn behavior is limited by the relationship between world height, visible viewport height, and despawn margin.
