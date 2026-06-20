# Debug Shape Catalog Output

Parent index: [Server](./!INDEX.md)

## Purpose

This document describes the server-side debug shape catalog output.

It covers how the game server builds reusable debug shape definitions from shared collision-shape data, when the catalog is sent to clients, what the packet contains, what it deliberately excludes, and how the output stays separate from gameplay collision authority.

## Overview

The debug shape catalog is a devtools-only server output used by client hitbox overlays.

The server builds a `debug_shape_catalog` packet from the shared collision-shape catalog:

```text
shared/collisions/collision_shapes.json
-> services/game-server/internal/game/physics.LoadCollisionShapeCatalog()
-> services/game-server/internal/devtools.BuildShapeCatalog()
-> services/game-server/internal/networking/outbound.BuildDebugShapeCatalogResponse()
-> debug_shape_catalog websocket packet
```

The packet contains reusable shape definitions keyed by debug shape id. It does not contain live entity state.

Live entity position, rotation, scale, variant, and pickup state continue to come from normal gameplay state packets. The client combines normal gameplay state with the debug shape catalog for presentation-only overlay drawing.

## Debug-only scope

The debug shape catalog is diagnostic metadata.

It may expose:

```text
debug shape ids
entity-kind categories
imported collision shape type
outline points for reusable shape geometry
```

It must not expose or own:

```text
live player collections
live asteroid collections
live bullet collections
live pickup collections
collision results
damage results
pickup collection results
targeting decisions
respawn decisions
score changes
gameplay authority
```

The catalog is useful for comparing client presentation against server collision geometry. It is not a gameplay packet and is not a replacement for normal server state.

## Server authority

The server remains authoritative for collision geometry and collision consequences.

The shape catalog output is derived from the same shared collision data used by game-server collision systems, but it does not decide whether collisions happen. It only publishes reusable outline definitions for devtools presentation.

Current generated shape ids are:

```text
player:<ship_type>
asteroid:<variant>
bullet
pickup:<catalog_key>
```

The current player shape id helper falls back to:

```text
player:v_wing
```

when the ship type is empty.

Asteroid shape ids are indexed by the imported asteroid shape list:

```text
asteroid:0
asteroid:1
asteroid:2
...
```

Bullet shape id is always:

```text
bullet
```

Pickup shape ids are built from the pickup shape catalog key. Current collision-shape data uses pickup class keys such as:

```text
pickup:powerup
pickup:weapon
```

## Catalog packet shape

The generated Go packet struct is:

```go
type DebugShapeCatalogPacket struct {
    Type   string                          `json:"type"`
    Shapes map[string]DebugShapeDefinition `json:"shapes"`
}
```

Each shape definition is:

```go
type DebugShapeDefinition struct {
    ID        string            `json:"id"`
    Kind      string            `json:"kind"`
    ShapeType string            `json:"shape_type"`
    Points    []DebugShapePoint `json:"points"`
}
```

Each point is:

```go
type DebugShapePoint struct {
    X float64 `json:"x"`
    Y float64 `json:"y"`
}
```

The wire packet has this high-level form:

```json
{
  "type": "debug_shape_catalog",
  "shapes": {
    "player:v_wing": {
      "id": "player:v_wing",
      "kind": "player",
      "shape_type": "polygon",
      "points": [
        { "x": 12.0, "y": -11.0 }
      ]
    }
  }
}
```

The `points` array is local reusable geometry. It is not already transformed by live entity position, rotation, or scale.

## Catalog construction

`BuildShapeCatalog()` receives a `physics.CollisionShapeCatalog` and returns:

```go
map[string]DebugShapeDefinition
```

The builder adds:

```text
catalog.Ship      -> player:v_wing
catalog.Bullet    -> bullet
catalog.Asteroids -> asteroid:<index>
catalog.Pickups   -> pickup:<catalog_key>
```

Each imported shape is converted to a runtime `physics.CollisionShape` at scale `1`.

The builder then creates a temporary `physics.CollisionBody` from that reusable shape and calls:

```go
physics.CollisionBodyOutlinePoints(body)
```

The resulting outline points are copied into `DebugShapePoint` values.

Invalid shapes are skipped. A shape is skipped when:

```text
imported shape conversion fails
outline point generation returns no points
```

The catalog builder does not synthesize fallback geometry for invalid data.

## Outline behavior

Outline generation is owned by the game-server physics package.

Current outline support covers:

```text
circle
capsule
rectangle
polygon
```

Circle outlines use a fixed segment count. Capsule outlines derive points from the capsule radius and height. Rectangle and polygon outlines use polygon point output.

The debug shape catalog stores these outline points as reusable local geometry. Client presentation applies live entity transforms later.

## Send lifecycle

The shape catalog is sent by the websocket write loop after gameplay presentation state becomes eligible.

The relevant write-loop sequence is:

```text
write gameplay presentation state
-> write debug shape catalog when eligible and not yet sent for this room id
-> periodically write debug status
```

`writeDebugShapeCatalogMessage()` sends the packet at most once for the current room id tracked by the current websocket write loop.

Eligibility requires:

```text
session.currentRoomID is not empty
the current room id has not already received the catalog from this write loop
outbound.CanSendDebugShapeCatalog(session.room) returns true
```

`CanSendDebugShapeCatalog()` requires:

```text
room is not nil
room.GameInstance() is not nil
devtools.Enabled() is true
room state is InGame or GameOver
```

The send-once behavior is local to the current write loop. It is not an acknowledgement protocol and is not durable delivery.

## Build and runtime gates

Server devtools are enabled in default builds through:

```text
services/game-server/internal/devtools/enabled_default.go
```

Server devtools are disabled when built with the `nodevtools` tag through:

```text
services/game-server/internal/devtools/enabled_nodevtools.go
```

The debug shape catalog output uses the same server-side devtools availability gate as other server devtools outputs.

Client-side UI gates do not control server authority. The client hitbox overlay can be hidden or shown locally, but the server only sends `debug_shape_catalog` when the server-side room, game, and build gates allow it.

## Packet encoding and failure behavior

`BuildDebugShapeCatalogResponse()` loads the collision-shape catalog, builds the debug shape catalog packet, and encodes it through:

```text
services/game-server/internal/protocol/packetcodec/codec.go
```

The current packet codec uses JSON encoding.

Failure behavior:

```text
collision-shape catalog load failure -> log network error and return no packet
packet encode failure -> log network error and return no packet
websocket write failure -> close the write loop for that session
invalid individual shape -> skip that shape and continue building the catalog
```

A failed catalog output does not mutate gameplay state and does not block normal gameplay presentation packets.

## Client presentation boundary

The server sends only reusable shape definitions.

Client-side devtools presentation owns:

```text
reading the debug_shape_catalog packet
storing shape definitions
resolving live gameplay entities to shape ids
combining shape points with gameplay-state transforms
converting server-space points through WorldSync
drawing overlay polylines
```

The current client hitbox overlay uses normal gameplay state fields such as:

```text
server_players
server_asteroids
server_bullets
server_pickups
```

The server catalog does not send those fields.

The overlay does not request a catalog refresh when the checkbox is toggled. The checkbox only controls local overlay visibility.

## Relationship to collision-shape data

The debug shape catalog is downstream of collision-shape data.

Collision geometry is authored in Godot scenes, exported to:

```text
shared/collisions/collision_shapes.json
```

and consumed by the game server.

The debug shape catalog output reloads the shared collision-shape JSON when building the outbound packet. It does not use the already-loaded `Game.collisionShapes` field from the active game instance.

The catalog output should remain a diagnostic projection of shared collision data. It should not become a separate source of truth for collision geometry.

## Relationship to gameplay collision telemetry

The debug shape catalog is different from live collision-body telemetry.

Debug shape catalog output contains reusable shape definitions:

```text
shape id
kind
shape type
local outline points
```

Live collision-body telemetry contains entity-specific collision body observations:

```text
entity kind
entity id
runtime position
runtime outline points
```

The catalog output supports the client hitbox overlay by giving the client reusable geometry. It is not the same thing as exporting the current server collision bodies every frame.

## Code map

Primary server files:

```text
services/game-server/internal/devtools/shape_catalog.go
services/game-server/internal/devtools/shape_ids.go
services/game-server/internal/devtools/packets_generated.go
services/game-server/internal/networking/outbound/debug_shape_catalog_presentation.go
services/game-server/internal/networking/websocket_write.go
```

Build/runtime gate files:

```text
services/game-server/internal/devtools/enabled_default.go
services/game-server/internal/devtools/enabled_nodevtools.go
services/game-server/internal/devtools/disabled.go
```

Collision-shape and outline support:

```text
services/game-server/internal/game/physics/collision_shapes.go
services/game-server/internal/game/physics/collision_outline.go
```

Packet source and generated output:

```text
shared/packets/debug.toml
shared/packets/outputs.toml
services/game-server/internal/devtools/packets_generated.go
```

Shared collision data:

```text
shared/collisions/collision_shapes.json
```

Client consumers:

```text
client/scripts/devtools/hitboxes/debug_shape_catalog_packet_reader.gd
client/scripts/devtools/hitboxes/debug_shape_catalog_store.gd
client/scripts/devtools/hitboxes/debug_shape_id_resolver.gd
client/scripts/gameplay/debug/server_hitbox_overlay_flow.gd
client/scripts/devtools/hitboxes/devtools_server_hitbox_overlay.gd
```

Related tests:

```text
services/game-server/internal/devtools/shape_catalog_test.go
services/game-server/internal/devtools/shape_ids_test.go
services/game-server/internal/devtools/enabled_default_test.go
services/game-server/internal/devtools/disabled_test.go
services/game-server/internal/networking/outbound/debug_shape_catalog_presentation_test.go
client/tests/unit/gameplay/debug/test_server_hitbox_overlay_flow.gd
client/tests/unit/devtools/hitboxes/test_debug_shape_catalog_packet_reader.gd
```

Important non-ownership boundaries:

```text
services/game-server/internal/game/collisions.go
services/game-server/internal/game/pickup_collisions.go
services/game-server/internal/game/targeting.go
services/game-server/internal/game/session.go
services/game-server/internal/game/export_devtools_collision_telemetry.go
client/scripts/world/
client/scripts/gameplay/
```

## Tests and verification

Server devtools shape catalog tests verify that `BuildShapeCatalog()`:

```text
includes expected player, asteroid, bullet, and pickup shape ids
skips invalid shapes
returns non-empty outline points
does not use live entity-id formats as reusable shape ids
does not emit origin-only placeholder points
```

Shape id tests verify:

```text
explicit v_wing player shape id
empty player shape id fallback to v_wing
asteroid variant formatting
bullet shape id formatting
pickup shape id formatting
```

Outbound packet tests verify that `BuildDebugShapeCatalogResponse()`:

```text
builds a JSON response
sets packet type to debug_shape_catalog
includes a non-empty shapes object
does not include debug_collision_bodies
does not include live players
does not include live asteroids
does not include live bullets
does not include live pickups
```

Build-gate tests verify default and `nodevtools` behavior for server devtools availability.

Useful focused verification:

```bash
cd services/game-server
go test -buildvcs=false ./internal/devtools ./internal/networking/outbound
```

Full server verification:

```bash
cd services/game-server
go test -buildvcs=false ./...
```

## Related docs

* [Server Devtools](./!INDEX.md)
* [Devtools](../!INDEX.md)
* [Client Hitbox Overlays](../client/hitbox-overlays.md)
* [Client Packet Routing And Devtools Input](../client/packet-routing-and-devtools-input.md)
* [Game Server Outbound Packet Routing](../../services/game-server/networking/outbound-message-flow.md)
* [Server Collision Shapes](../../services/game-server/simulation/world/collision-shapes.md)
* [Collision Shape Data](../../data/collision-shape-data.md)
* [Packet Schemas](../../data/packet-schemas.md)
* [Protocol](../../protocol/!INDEX.md)

## Notes

The legacy devtools documentation correctly treated server hitbox display as devtools-only presentation and not as gameplay authority. The current implementation has since moved from client-only hitbox templates toward a server-provided reusable shape catalog.

The current packet sends shape definitions once per room id within a websocket write loop. If refresh, acknowledgement, or catalog versioning becomes necessary later, that belongs in protocol or devtools planning until implemented.

Pickup shape ids in the server catalog are based on collision-shape catalog keys. Current collision-shape data uses pickup class keys such as `powerup` and `weapon`, while live pickup state also has a pickup type such as `1_up`. Client overlay drawing depends on resolver ids matching catalog ids.
