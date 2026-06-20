# Hitbox And Shape Debugging

Parent index: [Server](./!INDEX.md)

## Purpose

This document describes the server side of hitbox and shape debugging for Space Rocks devtools.

It covers how the game server exposes collision-shape metadata for client overlays, how shape identifiers are built, what runtime gates control the output, which gameplay collision systems remain authoritative, and where the server boundary ends.

## Overview

Server hitbox and shape debugging is a devtools-only diagnostic surface over authoritative server collision geometry.

The current active client-facing server output is:

```text
debug_shape_catalog
```

That packet contains reusable shape definitions. It does not contain live entity positions, live collision bodies, collision results, damage events, pickup collection results, or gameplay mutations.

The server builds the debug shape catalog from the shared collision-shape handoff:

```text
shared/collisions/collision_shapes.json
-> physics.LoadCollisionShapeCatalog()
-> devtools.BuildShapeCatalog(...)
-> DebugShapeCatalogPacket
-> packetcodec.Encode(...)
-> websocket write loop
```

The client combines that catalog with normal gameplay state packets to draw the server hitbox overlay. Live entity placement comes from the ordinary `state` packet, not from the shape catalog.

The server also has a game-owned collision body telemetry adapter:

```text
Game.DevtoolsCollisionBodies()
```

That adapter converts current runtime collision bodies into outline telemetry for players, asteroids, bullets, and pickups. It is a narrow game-owned export seam, not a replacement collision system. Current outbound packet tests require `debug_collision_bodies` to be absent from gameplay presentation, debug status, and debug shape catalog packets.

## Debug-only scope

Hitbox and shape debugging exists for development diagnostics.

It is not:

* production gameplay UI
* a gameplay collision resolver
* a client prediction mechanism
* a gameplay authority transfer to the client
* a packet command surface
* a data authoring workflow
* a replacement for server collision bodies
* a separate debug-only collision implementation

The debug shape catalog is read-only presentation metadata. It lets the client draw outlines that are derived from server collision geometry, but it does not allow the client to decide hits, damage, pickup collection, target validity, respawn safety, or scoring.

## Server authority

The game server remains authoritative for collision geometry as used by gameplay systems.

Runtime collision bodies are built from:

```text
physics.CollisionShapeCatalog
runtime entity state
entity CollisionBody(...) methods
```

Current entity body builders are:

```text
runtime.Ship.CollisionBody(...)
runtime.Bullet.CollisionBody(...)
runtime.Asteroid.CollisionBody(...)
pickups.Pickup.CollisionBody(...)
```

Those bodies are consumed by server-owned systems for:

```text
projectile -> asteroid collision checks
player -> asteroid collision checks
player -> pickup collection checks
server-side target click validation
safe spawn and respawn placement
radial candidate radius derivation
devtools collision telemetry
```

The devtools shape catalog reuses the same imported collision-shape catalog and the same outline helper used by server collision telemetry:

```text
physics.CollisionBodyOutlinePoints(...)
```

That helper produces outline points for supported shapes:

```text
circle
capsule
rectangle
polygon
```

The shape catalog path does not own collision consequences. Damage, scoring, pickup effects, target selection, respawn decisions, and lifecycle changes remain in their owning gameplay systems.

## Shape catalog output

The server builds a `DebugShapeCatalogPacket` with this shape:

```text
type
shapes
```

`type` is:

```text
debug_shape_catalog
```

`shapes` is a map keyed by debug shape id. Each `DebugShapeDefinition` contains:

```text
id
kind
shape_type
points
```

Each point contains:

```text
x
y
```

The current debug shape id rules are:

```text
player:<ship_type>
asteroid:<variant>
bullet
pickup:<collision_shape_key>
```

Current server helpers are:

```text
PlayerShapeID(shipType)
AsteroidShapeID(variant)
BulletShapeID()
PickupShapeID(pickupShapeKey)
```

`PlayerShapeID("")` falls back to:

```text
player:v_wing
```

`BuildShapeCatalog` currently adds:

```text
player:v_wing
bullet
asteroid:<index> for each imported asteroid shape
pickup:<key> for each imported pickup shape entry
```

Invalid imported shapes are skipped instead of producing empty or placeholder shape definitions.

## Shape catalog source data

The server loads collision-shape data from:

```text
shared/collisions/collision_shapes.json
```

The current shared catalog has these top-level entries:

```text
ship
bullet
asteroids
pickups
```

Current shape usage is:

```text
ship     -> polygon
bullet   -> capsule
asteroid -> polygon list
pickups  -> circle shapes keyed by pickup collision class
```

The debug shape catalog output uses the loaded catalog as diagnostic source data. It does not author or regenerate the collision-shape file.

Collision-shape authoring and export are owned by the data pipeline and Godot scene export workflow, not by server devtools.

## Client presentation

The client presentation surface is the server hitbox overlay.

Current client flow:

```text
server sends debug_shape_catalog
-> client routes debug_shape_catalog through inbound packet routing
-> ServerHitboxOverlayFlow stores shape definitions
-> client receives normal gameplay state packets
-> ServerHitboxOverlayFlow resolves shape ids for live entities
-> ServerHitboxOverlayFlow transforms local shape points by entity state
-> WorldSync converts server positions to visual positions
-> DevtoolsServerHitboxOverlay draws closed outlines
```

The server output only supplies reusable shape definitions. The client uses normal gameplay state for live entity values such as:

```text
position
rotation
scale
variant
ship_type
pickup type
```

The overlay is presentation-only. It does not send a server packet when toggled and it does not mutate gameplay.

## Commands or controls

There is no server command that requests, refreshes, or toggles the shape catalog.

The server sends the `debug_shape_catalog` packet automatically from the WebSocket write loop when all gates pass.

The visible overlay is controlled on the client through the devtools window checkbox:

```text
Show Server Collision Telemetry
```

That checkbox only changes client-side overlay visibility. It does not cause a shape-catalog request packet and does not change server collision state.

## Outbound send behavior

The debug shape catalog is sent by:

```text
writeDebugShapeCatalogMessage(...)
```

The write loop attempts the catalog send after a successful gameplay presentation write.

The send is limited by the write loop’s room tracking:

```text
lastDebugShapeCatalogRoomID
```

For a given WebSocket write-loop context, the catalog is sent at most once per room ID. This is not an acknowledgement protocol. Re-entering a different room ID makes the catalog eligible again in that write-loop context.

The outbound helper:

```text
BuildDebugShapeCatalogResponse(...)
```

loads the shared collision-shape catalog, builds the devtools shape catalog, encodes it through `packetcodec`, and returns the encoded WebSocket payload.

If loading or encoding fails, the server logs the failure and does not send the packet.

## Telemetry

The active outbound telemetry payload for hitbox overlay shape data is:

```text
debug_shape_catalog
```

It contains shape definitions only.

It intentionally does not include:

```text
players
asteroids
bullets
pickups
debug_collision_bodies
collision results
contact points
damage events
pickup collection events
```

Current live entity positions come from normal gameplay state packets.

The game-owned collision body telemetry adapter returns per-entity outline data shaped as:

```text
kind
id
shape
points
```

That adapter uses current runtime entities and the game instance’s loaded collision-shape catalog. It skips entities whose collision bodies cannot be built.

## Build and runtime gates

Server devtools are enabled in default builds:

```text
services/game-server/internal/devtools/enabled_default.go
```

Server devtools are disabled when built with the `nodevtools` tag:

```text
services/game-server/internal/devtools/enabled_nodevtools.go
```

The shape catalog outbound path is eligible only when:

```text
room != nil
room.GameInstance() != nil
devtools.Enabled()
room.State == InGame or GameOver
session.currentRoomID != ""
the current room id has not already received the catalog from this write loop
```

The shape catalog is not sent in lobby-only room states and is not sent from `nodevtools` builds.

## Relationship to real gameplay implementation

Server hitbox and shape debugging must route through real gameplay and physics seams.

The debug shape catalog uses:

```text
physics.LoadCollisionShapeCatalog()
physics.ImportedCollisionShape.ToCollisionShape(...)
physics.CollisionBodyOutlinePoints(...)
```

The collision body telemetry adapter uses each runtime entity’s normal `CollisionBody(...)` method.

This preserves the ownership split:

```text
collision-shape data -> data/export pipeline
shape loading/conversion -> physics
runtime bodies -> entity methods
collision consequences -> gameplay consumers
debug metadata -> devtools
packet delivery -> networking outbound
visual outlines -> client devtools
```

Devtools must not add alternate collision constants or duplicate collision rules outside the normal shape-loading and body-building paths.

## Failure behavior

Shape catalog output is best-effort diagnostics.

If `physics.LoadCollisionShapeCatalog()` fails during debug shape catalog output, the server logs:

```text
debug shape catalog load failed
```

and skips the packet.

If packet encoding fails, the server logs:

```text
debug shape catalog packet encode failed
```

and skips the packet.

If an imported shape cannot convert into a runtime collision shape, `BuildShapeCatalog` skips that shape entry.

If the client receives gameplay state for an entity but does not have a matching debug shape definition, the client skips drawing that entity’s overlay entry. Normal gameplay presentation and authoritative simulation continue.

## Code map

Primary server devtools files:

```text
services/game-server/internal/devtools/shape_catalog.go
services/game-server/internal/devtools/shape_ids.go
services/game-server/internal/devtools/packets_generated.go
services/game-server/internal/devtools/enabled_default.go
services/game-server/internal/devtools/enabled_nodevtools.go
```

Outbound networking files:

```text
services/game-server/internal/networking/websocket_write.go
services/game-server/internal/networking/outbound/debug_shape_catalog_presentation.go
services/game-server/internal/networking/outbound/server_message_writer.go
```

Game-owned collision telemetry seam:

```text
services/game-server/internal/game/export_devtools_collision_telemetry.go
```

Physics and collision-shape support:

```text
services/game-server/internal/game/physics/collision_shapes.go
services/game-server/internal/game/physics/collision.go
services/game-server/internal/game/physics/collision_outline.go
services/game-server/internal/game/physics/vector.go
```

Runtime collision body builders:

```text
services/game-server/internal/game/runtime/ship.go
services/game-server/internal/game/runtime/bullet.go
services/game-server/internal/game/runtime/asteroid.go
services/game-server/internal/game/entities/pickups/pickup.go
```

Packet schema and generated output:

```text
shared/packets/debug.toml
shared/packets/outputs.toml
services/game-server/internal/devtools/packets_generated.go
```

Collision-shape source data:

```text
shared/collisions/collision_shapes.json
```

Important non-ownership boundaries:

```text
client/scripts/devtools/
client/scripts/gameplay/debug/
client/scripts/world/
client/tools/export_collision_shapes.gd
services/game-server/internal/game/damage/
services/game-server/internal/game/scoring/
services/game-server/internal/game/entities/pickups/
services/game-server/internal/game/effects/radial/
tools/data_sync/
```

## Tests and verification

Focused server tests:

```text
services/game-server/internal/devtools/shape_catalog_test.go
services/game-server/internal/devtools/shape_ids_test.go
services/game-server/internal/networking/outbound/debug_shape_catalog_presentation_test.go
services/game-server/internal/game/export_devtools_collision_telemetry_test.go
services/game-server/internal/devtools/enabled_default_test.go
services/game-server/internal/devtools/disabled_test.go
```

Shape catalog tests verify that:

* expected player, asteroid, bullet, and pickup shape ids are included
* invalid shapes are skipped
* entries use debug shape ids rather than live entity id formats
* outline points are populated rather than origin-only placeholders

Outbound tests verify that `debug_shape_catalog` responses:

* encode as JSON
* use packet type `debug_shape_catalog`
* include a non-empty `shapes` object
* do not include live entity collections
* do not include `debug_collision_bodies`

Collision telemetry tests verify that `Game.DevtoolsCollisionBodies()`:

* uses runtime server collision body builders
* returns lowercase JSON keys
* includes outline points for supported entity bodies
* skips entities whose collision bodies cannot be built

Useful focused verification:

```bash
cd services/game-server
go test -buildvcs=false ./internal/devtools ./internal/networking/outbound ./internal/game -run 'Shape|DebugShape|DevtoolsCollision'
```

Nodevtools gate verification:

```bash
cd services/game-server
go test -buildvcs=false -tags nodevtools ./internal/devtools/...
```

Full game-server verification:

```bash
cd services/game-server
go test -buildvcs=false ./...
```

Packet drift verification when debug packet schemas change:

```bash
data-sync -check -packets -go -gds
```

## Related docs

* [Server Devtools](./!INDEX.md)
* [Devtools](../!INDEX.md)
* [Client Hitbox Overlays](../client/hitbox-overlays.md)
* [Client Devtools](../client/!INDEX.md)
* [Game Server Networking](../../services/game-server/networking/!INDEX.md)
* [Outbound Packet Routing](../../services/game-server/networking/outbound-message-flow.md)
* [Server Collision Shapes](../../services/game-server/simulation/world/collision-shapes.md)
* [Server Physics](../../services/game-server/simulation/world/physics.md)
* [Collision Shape Data](../../data/collision-shape-data.md)
* [Packet Schemas](../../data/packet-schemas.md)
* [Protocol](../../protocol/!INDEX.md)
* [Player Build Limits](../../limits/player-build-limits.md)

## Notes

The UI label currently says “Show Server Collision Telemetry,” while the current client implementation names the visual layer as a server hitbox overlay.

The current outbound overlay path reconstructs visible outlines from reusable shape definitions plus normal gameplay state. It does not receive live precomputed per-entity collision body outlines from the server.

Pickup debug shape ids are based on collision-shape catalog keys, while pickup gameplay state can expose pickup type. A pickup overlay entry is drawn only when the client-resolved pickup shape id matches a catalog shape id. Missing overlay shape data does not affect pickup simulation, pickup rendering, or pickup collection behavior.

The collision-shape catalog currently has one ship shape. Unknown or empty ship shape IDs fall back to the default `v_wing` shape in server collision-shape support.
