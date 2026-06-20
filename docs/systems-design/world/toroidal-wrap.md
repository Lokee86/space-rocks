# Toroidal Wrap

Parent index: [World](./!INDEX.md)

## Purpose

This document defines the conceptual toroidal world-wrap model for Space Rocks.

It explains how bounded authoritative server coordinates, shortest wrapped spatial relationships, and continuous client visual coordinates work together so the world behaves as one wraparound playfield rather than as a flat rectangle with hard edges.

## Overview

Space Rocks uses a toroidal playfield.

A toroidal world has no terminal edge. Crossing the right edge returns an entity to the left edge, crossing the left edge returns it to the right edge, crossing the bottom edge returns it to the top edge, and crossing the top edge returns it to the bottom edge.

The design goal is:

```text
server authority stays bounded
client presentation stays continuous
spatial relationships use the shortest wrapped path
```

The server stores one authoritative bounded position per entity. It does not duplicate entities as ghost copies at world seams. When an entity moves outside the world bounds, the server normalizes that position back into the bounded coordinate range.

The client receives bounded server positions, but it renders them as continuous visual positions relative to the active ViewAnchor. This prevents the camera, background, local player, remote players, projectiles, asteroids, pickups, and effects from visibly snapping when the authoritative server coordinate wraps.

## Conceptual model

Toroidal wrap has two coordinate layers:

```text
authoritative server coordinate
= bounded gameplay position inside the world dimensions

client visual coordinate
= continuous presentation position relative to the active render anchor
```

The server coordinate is the gameplay fact. It is used for simulation, spawning, collision, respawn safety, radial coverage, visibility, despawn, targeting validation, state packets, and authoritative outcomes.

The client visual coordinate is a presentation fact. It is used for rendering, camera continuity, background continuity, local pointer conversion, target picking presentation, event effects, and interpolation.

The two layers must remain related by shortest wrapped delta:

```text
visual_position =
  anchor_visual_position
  + shortest_wrapped_delta(anchor_server_position, entity_server_position)
```

This means two entities near opposite edges of the bounded server world can be visually close if the toroidal path between them is short.

Example:

```text
World width: 1000
Anchor server x: 990
Entity server x: 10

Naive delta: -980
Shortest wrapped delta: +20
```

The entity should appear slightly to the right of the anchor, not almost a full world width to the left.

## World bounds

World dimensions are shared constants.

The current source values live in:

```text
shared/constants/server_constants.toml
```

Current world dimensions are:

```text
world_width  = 17200.0
world_height = 9200.0
```

These values generate to both server and client outputs:

```text
services/game-server/internal/constants/constants.go
client/scripts/generated/constants/constants.gd
```

The server uses generated Go constants for authoritative simulation bounds:

```text
constants.WorldWidth
constants.WorldHeight
```

The client uses generated GDScript constants for visual wrap math:

```text
Constants.WORLD_WIDTH
Constants.WORLD_HEIGHT
```

The shared world constants must stay aligned across Go and GDScript. If the server and client use different world dimensions, authoritative wrap and visual wrap will disagree.

## Authority rules

The game server owns authoritative wrap behavior.

The server is responsible for:

* storing bounded authoritative positions
* normalizing moved positions into world bounds
* evaluating shortest wrapped distance, direction, and delta for gameplay systems
* publishing bounded positions through gameplay state packets
* resolving gameplay outcomes using wrapped spatial relationships

The client owns presentation continuity.

The client is responsible for:

* converting bounded server positions into continuous visual positions
* keeping the active ViewAnchor stable across wrap edges
* rendering entities, effects, camera, and background relative to the same anchor basis
* converting visual pointer or target positions back into bounded server coordinates before sending requests

The client must not use visual continuity as gameplay authority. A visual coordinate can leave the bounded server range; that does not make it an authoritative gameplay position.

## Server wrap model

The server keeps one position per runtime entity.

Ships, asteroids, and projectiles move through simulation and then wrap back into bounded space. Conceptually:

```text
step entity movement
wrap final position into world bounds
store wrapped position
```

The server-side spatial model also owns shortest wrapped relationships:

```text
Delta(from, to)
Distance(from, to)
Direction(from, to)
NormalizePosition(position)
```

These relationships are used by systems that need toroidal behavior without owning the wrap model themselves.

Current server consumers include:

* player movement
* asteroid movement
* projectile movement
* asteroid spawn aiming
* visibility and despawn checks
* projectile/asteroid collision checks
* player/asteroid collision checks
* player/pickup collision checks
* radial effect coverage
* respawn safety

The important design rule is that collision and proximity systems should use wrapped-local placement or wrapped distance, not duplicate runtime entities at seams.

## Client visual model

The client renders continuous coordinates.

The active ViewAnchor is the render origin for world presentation. Normal gameplay usually anchors to the local player. Spectate or explicit view-target behavior can use another active player as the render anchor.

The client tracks both:

```text
anchor_server_position
anchor_visual_position
```

When the anchor crosses a server wrap boundary, its visual position advances by the shortest wrapped delta from the previous bounded server position to the new bounded server position. This keeps the camera and background continuous.

Other world entities are then rendered relative to that same anchor basis.

Conceptually:

```text
entity_visual_position =
  anchor_visual_position
  + shortest_wrapped_delta(anchor_server_position, entity_server_position)
```

Players, asteroids, projectiles, pickups, gameplay effects, target read models, and debug overlays should all use the same coordinate boundary: server coordinates are authoritative facts, and visual coordinates are presentation positions.

## Shortest wrapped path

Toroidal wrap is not just position normalization.

Systems that compare two positions must use the shortest wrapped path when the relationship crosses an edge.

Position normalization answers:

```text
Where does this coordinate live inside the bounded world?
```

Shortest wrapped delta answers:

```text
What is the shortest vector from one bounded coordinate to another in a toroidal world?
```

Both are required.

Without normalization, authoritative positions can drift outside the bounded world.

Without shortest wrapped deltas, systems near opposite edges incorrectly treat nearby toroidal positions as far apart.

## Invariants

Toroidal wrap must preserve these invariants:

* The server owns authoritative gameplay positions.
* Authoritative gameplay positions are bounded by world width and world height after movement normalization.
* The server stores one runtime entity position, not duplicated ghost positions at wrap seams.
* Server spatial relationships use shortest wrapped deltas when edge-crossing proximity matters.
* Collision checks across wrap boundaries use wrapped-local placement or equivalent shortest-path logic.
* Respawn safety treats threats across a wrap boundary as nearby when the toroidal distance is short.
* Asteroid spawn aiming uses wrapped direction so edge-adjacent targets are aimed at through the shortest path.
* Visibility and despawn use wrapped deltas from camera views.
* The client owns continuous visual coordinates, not authoritative gameplay coordinates.
* ViewAnchor is the active render basis for camera, background, entity presentation, and coordinate conversion.
* Camera position and render-anchor position must not diverge.
* Client target and pointer requests must convert visual positions back into bounded server coordinates before sending requests.
* Shared world-size constants must remain aligned between generated Go and generated GDScript outputs.

## Participating systems

The main participating systems are:

```text
Game server simulation
= authoritative positions, movement, wrap normalization, spatial relationships, collision, respawn, visibility, and state packets

Client world sync
= continuous visual coordinates, ViewAnchor, entity sync, interpolation, and presentation coordinate conversion

Data/constants pipeline
= shared world width and height source of truth

Realtime protocol
= bounded authoritative positions sent from server to client
```

The server and client have different roles by design. The server answers what is true. The client answers how that truth is presented smoothly.

## Implementation references

The following implementation references are non-exhaustive orientation points for where this systems-design rule is realized. They are not a code map or ownership map.

The systems-design rule is implemented primarily by the server toroidal space and motion boundary and the client ViewAnchor/visual-coordinate boundary.

Server implementation reference:

```text
services/game-server/internal/game/space/space.go
services/game-server/internal/game/motion/motion.go
services/game-server/internal/game/simulation.go
services/game-server/internal/game/simulation_players.go
services/game-server/internal/game/simulation_asteroids.go
services/game-server/internal/game/simulation_bullets.go
services/game-server/internal/game/visibility.go
services/game-server/internal/game/spawning.go
services/game-server/internal/game/spawning/spawner.go
services/game-server/internal/game/session.go
services/game-server/internal/game/collisions.go
services/game-server/internal/game/effects/radial/step.go
```

Client implementation reference:

```text
client/scripts/world/world_wrap.gd
client/scripts/world/world_sync.gd
client/scripts/world/asteroid_sync.gd
client/scripts/world/projectile_sync.gd
client/scripts/world/pickup_sync.gd
client/scripts/world/player_render/player_render_api.gd
client/scripts/world/player_render/view_anchor_sync.gd
client/legacy/player_render/local_visual_sync.gd
client/legacy/player_render/visual_sync_positions.gd
```

Data reference:

```text
shared/constants/server_constants.toml
services/game-server/internal/constants/constants.go
client/scripts/generated/constants/constants.gd
```

These references are not a full ownership map. Detailed implementation ownership belongs in the relevant service and data docs.

## Verification expectations

Toroidal wrap should be verified at both math and gameplay levels.

Server verification should cover:

* wrapping right, left, top, and bottom edges
* wrapping positions more than one world size outside bounds
* shortest wrapped horizontal and vertical deltas
* wrapped distance
* wrapped direction
* player movement across edges
* asteroid movement across edges
* projectile movement across edges
* asteroid spawning near boundaries
* visibility and despawn near boundaries
* collision checks across boundaries
* respawn safety across boundaries

Client verification should cover:

* world wrap helper math
* continuous local visual position tracking
* ViewAnchor continuity across server wrap boundaries
* visual-relative placement of players, asteroids, projectiles, and pickups
* server-to-visual conversion
* visual-to-server conversion
* camera and background continuity

Manual smoke verification should include crossing every world edge and checking that camera, background, projectiles, asteroids, collision behavior, respawn safety, multiplayer relative positions, and event effects remain coherent.

## Related docs

* [World](./!INDEX.md)
* [Spawning And Space](spawning-and-space.md)
* [World Authority](world-authority.md)
* [Server Toroidal Space And Motion](../../services/game-server/simulation/world/toroidal-space-and-motion.md)
* [Server Visibility And Despawn](../../services/game-server/simulation/world/visibility-and-despawn.md)
* [Server Asteroid Spawning And Variants](../../services/game-server/simulation/world/asteroid-spawning-and-variants.md)
* [Server Collision To Damage Flow](../../services/game-server/simulation/combat/collision-to-damage-flow.md)
* [Server Player Respawn](../../services/game-server/simulation/players/player-respawn.md)
* [Client View Anchor And Visual Coordinates](../../services/client/world-sync/view-anchor-and-visual-coordinates.md)
* [Client Entity Sync Owners](../../services/client/world-sync/entity-sync-owners.md)
* [Constants Pipeline](../../data/constants.md)
* [Gameplay Packets](../../protocol/gameplay-packets.md)

## Notes

The core split is bounded authoritative server coordinates and continuous client visual coordinates. This document preserves that systems-design rule while leaving service-level implementation detail to the server and client docs.

`client/legacy/player_render/` still contains implementation support behind the active player-render API. Current docs should describe the active ViewAnchor/player-render seam and only reference implementation backing where necessary.

Toroidal wrap is a world model, not a camera trick. The camera and background benefit from continuous visual coordinates, but authoritative gameplay behavior also depends on wrapped distance, direction, visibility, collision, spawning, and respawn safety.
