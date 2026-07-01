## Collision Body Telemetry

Parent index: [Server](./!INDEX.md)

## Purpose

This document describes the server-side collision body telemetry support used by devtools.

It covers how the game server exposes authoritative collision outline data for debugging, how that data stays tied to real runtime collision bodies, which packet surfaces currently exist, and where the boundary ends.

## Overview

Collision body telemetry is a debug-only server support path for inspecting authoritative collision geometry.

The server builds collision bodies from the same runtime state and collision-shape catalog used by gameplay systems. The devtools adapter converts those bodies into outline snapshots that can be inspected or presented by debug tooling.

The current server-side collision telemetry path is:

```text
Game.collisionShapes
-> runtime entity CollisionBody(...) methods
-> Game.DevtoolsCollisionBodies()
-> physics.CollisionBodyOutlinePoints(...)
-> DevtoolsCollisionBody values
```

A telemetry body contains:

```text
kind
id
shape
points
```

Current supported entity kinds are:

```text
player
asteroid
bullet
pickup
```

This telemetry is for diagnostics only. It does not participate in collision detection, damage, targeting, pickup collection, scoring, despawn, respawn, or match outcome decisions.

The visible client server-hitbox overlay currently uses a related but separate packetized path:

```text
shared/collisions/collision_shapes.json
-> physics.LoadCollisionShapeCatalog
-> devtools.BuildShapeCatalog
-> debug_shape_catalog packet
-> client combines reusable shape definitions with normal gameplay state
```

`Game.DevtoolsCollisionBodies()` produces live server-side body outline snapshots, but the current outbound protocol does not emit a standalone `debug_collision_bodies` packet. Server outbound tests intentionally verify that `debug_shape_catalog` and normal gameplay presentation packets do not include `debug_collision_bodies`.

## Debug-only scope

Collision body telemetry is development tooling.

It may be used to:

* inspect server-owned collision geometry
* compare authoritative collision bodies against client presentation
* verify entity body construction paths
* diagnose shape catalog, rotation, scale, and outline projection issues
* support devtools-only presentation or future debug packet output

It must not become:

* player-facing HUD
* gameplay collision authority separate from normal collision code
* a local client prediction surface
* a second entity model
* a scoring, damage, pickup, respawn, or targeting system
* a production analytics or observability pipeline

The important rule is that telemetry observes the same server collision body construction path used by gameplay support. It must not invent a separate debug-only geometry model.

## Server authority

The server owns authoritative collision bodies.

Runtime entities expose body builders:

```text
runtime.Ship.CollisionBody(...)
runtime.Bullet.CollisionBody(...)
runtime.Asteroid.CollisionBody(...)
pickups.Pickup.CollisionBody(...)
```

Those methods combine:

```text
entity id
entity position
entity rotation when relevant
shape from CollisionShapeCatalog
```

The collision-shape catalog is loaded from:

```text
shared/collisions/collision_shapes.json
```

and stored on each `Game` instance as:

```text
collisionShapes physics.CollisionShapeCatalog
```

`Game.DevtoolsCollisionBodies()` locks the game aggregate, iterates current runtime entity maps, asks each entity to build a real collision body, and skips entities whose bodies cannot be built.

The adapter does not mutate runtime state. It only converts available bodies into devtools telemetry values.

The server remains authoritative for:

```text
which entities exist
which collision shape catalog is loaded
how runtime entity bodies are constructed
which bodies are skipped when shapes are unavailable
how outline points are derived
```

The client may draw the result for debugging, but the client does not become authoritative for collision geometry.

## Client presentation

Client presentation is separate from server telemetry ownership.

The current client hitbox overlay is controlled by the devtools window checkbox labeled:

```text
Show Server Collision Telemetry
```

The checkbox does not request live collision bodies from the server. It toggles local overlay visibility.

Current overlay drawing uses:

```text
debug_shape_catalog packet
world lane full/delta packets
WorldSync visual coordinate conversion
```

The server sends reusable shape definitions through `debug_shape_catalog`. The client combines those shape definitions with live entity state from world lane packet readback.

This means current client presentation does not consume `Game.DevtoolsCollisionBodies()` directly. The live body telemetry adapter remains a server-side support seam and test-covered implementation path, while the packetized client overlay uses shape catalog plus gameplay state.

The client must treat all collision telemetry and shape catalog data as presentation input only. It must not use overlay geometry to decide hits, target validity, pickup collection, damage, or respawn safety.

## Commands or controls

There is no current client command that requests server collision body telemetry.

Current controls related to hitbox presentation are client-side:

```text
DevtoolsWindow.ShowServerHitboxesCheckBox
-> show_server_hitboxes_changed
-> DevtoolsWindowController
-> DevtoolsWindowActionContext
-> DevtoolsOverlayContext.set_server_hitboxes_enabled()
-> DevtoolsServerHitboxOverlay.set_enabled()
```

The server-side shape catalog is sent opportunistically by the WebSocket write loop when eligible. It is not requested by a direct command packet.

Current packetized diagnostic surfaces related to collision display are:

```text
debug_shape_catalog
debug_status
```

There is no generated `debug_collision_bodies` packet in the current shared debug packet source.

If live collision body telemetry becomes packetized later, it should use the same generated packet/data-sync path as other debug packets and should remain gated behind server devtools enablement.

## Telemetry behavior

`Game.DevtoolsCollisionBodies()` returns a snapshot of currently buildable collision bodies.

The output type is:

```go
type DevtoolsCollisionBody struct {
    Kind   string
    ID     string
    Shape  string
    Points []DevtoolsCollisionPoint
}
```

Each point uses:

```go
type DevtoolsCollisionPoint struct {
    X float64
    Y float64
}
```

The JSON field names are lowercase:

```text
kind
id
shape
points
x
y
```

The adapter builds bodies in this order:

```text
players
asteroids
projectiles
pickups
```

For each entity:

1. The entity attempts to build a `physics.CollisionBody` from the game collision-shape catalog.
2. If body construction fails, the entity is skipped.
3. `physics.CollisionBodyOutlinePoints` projects the body into outline points.
4. The adapter records the entity kind, entity id, shape type, and outline points.

Supported outline projection behavior comes from the physics package:

```text
circle    -> 24 outline points
capsule   -> cap/segment outline points
rectangle -> polygon outline points
polygon   -> polygon outline points
```

The outline points are server-space points. They are not collision results, contact manifolds, or hit events.

## Shape catalog relationship

Collision body telemetry and debug shape catalog output share lower-level physics support but have different runtime meanings.

Collision body telemetry:

```text
live entity id
live entity position
live entity rotation
resolved runtime shape
server-space outline points
```

Debug shape catalog output:

```text
reusable shape id
entity kind
shape type
local outline points
no live entity ids
no live entity positions
```

The current client overlay uses the shape catalog path, not live body telemetry. That keeps the packetized server payload small and lets the client transform local shape points using normal gameplay state.

`debug_shape_catalog` is built from the loaded collision-shape catalog through:

```text
devtools.BuildShapeCatalog(...)
```

Current shape ids include:

```text
player:v_wing
asteroid:<variant>
bullet
pickup:<pickup-shape-key>
```

The shape catalog packet intentionally excludes live entity maps and `debug_collision_bodies`.

## Build and runtime gates

Server devtools are enabled in default builds and disabled when the game server is built with the `nodevtools` tag.

Relevant gate files:

```text
services/game-server/internal/devtools/enabled_default.go
services/game-server/internal/devtools/enabled_nodevtools.go
services/game-server/internal/devtools/disabled.go
```

`Game.DevtoolsCollisionBodies()` itself is a game-owned export adapter. It does not check the devtools build flag internally. Any caller that exposes the output outside the game package must apply the appropriate devtools gate.

The currently packetized shape catalog output is gated by:

```text
room exists
room has a game instance
server devtools are enabled
room state is InGame or GameOver
```

The WebSocket write loop sends the debug shape catalog once per room id per write loop after gameplay presentation has started for that session.

Current `nodevtools` behavior disables devtools command handling and prevents gated debug output such as the shape catalog from being eligible.

## Data and protocol surface

Collision body telemetry reads generated shared collision-shape data indirectly through the game collision-shape catalog:

```text
shared/collisions/collision_shapes.json
```

It does not own the source data or export pipeline.

The current generated debug packet source is:

```text
shared/packets/debug.toml
```

Current generated server devtools packets include `DebugShapeCatalogPacket` and `DebugStatusPacket`, but do not include a live collision-body telemetry packet.

Generated packet output is configured through:

```text
shared/packets/outputs.toml
```

The current outbound server behavior is:

```text
world lane packet readback
-> does not include debug_collision_bodies

debug_shape_catalog packet
-> includes reusable shape definitions
-> does not include live entity collections
-> does not include debug_collision_bodies

debug_status packet
-> includes debug toggle/status values
-> does not include collision bodies
```

If a future packet exposes `DevtoolsCollisionBody` values, that packet should be documented under protocol and data docs as a debug-only packet surface.

## Invariants

Collision body telemetry must preserve these rules:

```text
telemetry observes server-owned runtime bodies
telemetry uses the same entity CollisionBody methods as gameplay support
telemetry does not invent debug-only geometry
telemetry skips entities whose collision bodies cannot be built
telemetry does not mutate game state
telemetry does not decide collision consequences
telemetry does not bypass devtools build/runtime gates when exposed
telemetry output is diagnostic presentation data only
client overlays must not become collision authority
packetized debug output must use generated packet contracts
```

## Code map

Primary server telemetry adapter:

```text
services/game-server/internal/game/export_devtools_collision_telemetry.go
```

Runtime collision body builders:

```text
services/game-server/internal/game/runtime/ship.go
services/game-server/internal/game/runtime/bullet.go
services/game-server/internal/game/runtime/asteroid.go
services/game-server/internal/game/entities/pickups/pickup.go
```

Physics support:

```text
services/game-server/internal/game/physics/collision.go
services/game-server/internal/game/physics/collision_shapes.go
services/game-server/internal/game/physics/collision_outline.go
services/game-server/internal/game/physics/vector.go
```

Game aggregate catalog storage:

```text
services/game-server/internal/game/game.go
```

Related packetized shape catalog path:

```text
services/game-server/internal/devtools/shape_catalog.go
services/game-server/internal/devtools/shape_ids.go
services/game-server/internal/devtools/packets_generated.go
services/game-server/internal/networking/outbound/debug_shape_catalog_presentation.go
services/game-server/internal/networking/websocket_write.go
```

Server devtools gates:

```text
services/game-server/internal/devtools/enabled_default.go
services/game-server/internal/devtools/enabled_nodevtools.go
services/game-server/internal/devtools/disabled.go
```

Packet source and generated-output configuration:

```text
shared/packets/debug.toml
shared/packets/outputs.toml
```

Collision-shape source data:

```text
shared/collisions/collision_shapes.json
```

Related client presentation files:

```text
client/scenes/devtools/server_hitbox_overlay.tscn
client/scripts/gameplay/debug/server_hitbox_overlay_flow.gd
client/scripts/devtools/hitboxes/devtools_server_hitbox_overlay.gd
client/scripts/devtools/hitboxes/debug_shape_catalog_packet_reader.gd
client/scripts/devtools/hitboxes/debug_shape_catalog_store.gd
client/scripts/devtools/hitboxes/debug_shape_id_resolver.gd
```

Important non-ownership boundaries:

```text
services/game-server/internal/game/collisions.go
services/game-server/internal/game/combat.go
services/game-server/internal/game/pickup_collisions.go
services/game-server/internal/game/targeting.go
services/game-server/internal/game/session.go
services/game-server/internal/game/radial_candidates.go
services/game-server/internal/networking/
client/
docs/data/
docs/protocol/
```

Gameplay collision consumers own gameplay consequences. Networking owns packet transport. Client devtools owns visual presentation. Data and protocol docs own generated packet and source-of-truth documentation.

## Tests and verification

Focused server tests:

```text
services/game-server/internal/game/export_devtools_collision_telemetry_test.go
```

Those tests verify that:

* collision body telemetry uses server collision body builders
* player, bullet, and pickup bodies produce expected kind/id/shape output
* outline points are generated from resolved runtime bodies
* entities whose bodies cannot be built are skipped
* JSON output uses lowercase field names

Related physics tests:

```text
services/game-server/internal/game/physics/collision_outline_test.go
services/game-server/internal/game/physics/collision_shapes_test.go
services/game-server/internal/game/physics/collision_test.go
```

Related shape catalog and outbound tests:

```text
services/game-server/internal/devtools/shape_catalog_test.go
services/game-server/internal/networking/outbound/debug_shape_catalog_presentation_test.go
services/game-server/internal/networking/outbound/gameplay_presentation_test.go
services/game-server/internal/networking/outbound/debug_status_presentation_test.go
```

Those tests verify that packetized debug shape catalog output includes reusable shape definitions and excludes live collision body payloads.

Useful verification command:

```bash
cd services/game-server
go test -buildvcs=false ./...
```

Focused verification for this boundary:

```bash
cd services/game-server
go test -buildvcs=false ./internal/game ./internal/game/physics ./internal/devtools ./internal/networking/outbound -run 'CollisionBody|CollisionOutline|ShapeCatalog|DebugShapeCatalog|GameplayPresentation|DebugStatus'
```

Run packet generation checks when changing `shared/packets/debug.toml` or `shared/packets/outputs.toml`.

Run collision-shape data checks when changing Godot collision nodes, the collision-shape exporter, or `shared/collisions/collision_shapes.json`.

## Related docs

* [Server Devtools](./!INDEX.md)
* [Devtools](../!INDEX.md)
* [Client Hitbox Overlays](../client/hitbox-overlays.md)
* [Client Telemetry Overlays](../client/telemetry-overlays.md)
* [Game Server Collision Shapes](../../services/game-server/simulation/world/collision-shapes.md)
* [Game Server Physics](../../services/game-server/simulation/world/physics.md)
* [Game Server Telemetry And Packet Routing](../../services/game-server/networking/telemetry-packet-routing.md)
* [Collision Shape Data](../../data/collision-shape-data.md)
* [Packet Schemas](../../data/packet-schemas.md)
* [Protocol](../../protocol/!INDEX.md)

## Notes

The UI label uses “Server Collision Telemetry,” while several implementation paths use “hitbox” or “shape catalog” terminology. In current implementation, “server hitbox overlay” is the client presentation surface, `debug_shape_catalog` is the packetized reusable-shape surface, and `Game.DevtoolsCollisionBodies()` is the server-side live body snapshot adapter.

The current packetized overlay path reconstructs outlines on the client from reusable shape definitions and normal gameplay state. That is separate from sending precomputed live collision-body outlines from the server.

Collision body telemetry should remain boring and literal: it should expose what the server collision body path already knows, not create a more convenient but less authoritative debug geometry model.
