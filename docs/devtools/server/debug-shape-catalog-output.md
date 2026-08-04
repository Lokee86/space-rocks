---
author: brian
created: "2026-07-19"
document_id: 019f7d55-fb2c-7981-8fb8-01e66ddf9fd9
document_type: general
policy_exempt: false
summary: 'This document describes the server-side debugshapecatalog readout: how reusable collision-shape definitions are built, requested through sr.tooling, correlated, and consumed by the existing client hitbox presentation.'
---
# Debug Shape Catalog Output

Parent index: [Server](./!INDEX.md)

## Purpose

This document describes the server-side `debug_shape_catalog` readout: how reusable collision-shape definitions are built, requested through `sr.tooling`, correlated, and consumed by the existing client hitbox presentation.

## Overview

This document describes the debug-only Debug Shape Catalog Output surface, its authority boundary, controls, telemetry, runtime gates, implementation owners, and tests.

## Current route

The shape catalog is a one-shot privileged request/response interaction on the reliable ordered `sr.tooling` WebRTC DataChannel.

```text
client enters an active room and sr.tooling becomes ready
-> debug_shape_catalog_request(request_id)
-> tooling packet policy preflight
-> tooling.read capability and room attachment checks
-> outbound.BuildDebugShapeCatalogPacket(...)
-> debug_shape_catalog(request_id, shapes) over sr.tooling
-> ToolingPacketRouter
-> ClientConnectionService.debug_shape_catalog_received
-> existing gameplay composition and hitbox catalog owners
```

The legacy WebSocket write-loop push is removed. The server no longer tracks a connection-local “catalog sent room id,” and the client WebSocket dispatcher no longer classifies `debug_shape_catalog`.

The client requests the catalog once for each active room after both conditions are true:

```text
sr.tooling is ready
room state is in_game or game_over
```

Returning to a non-game room state clears the client’s per-room request marker. Replacing the realtime transport also clears the marker, so a recovered `sr.tooling` channel requests the catalog again for the still-active room. A later active room receives a new request.

## Request and response contract

`debug_shape_catalog_request` is defined in `shared/packets/tooling.toml` and carries:

```text
type
a non-empty request_id
```

It requires an attached room and `tooling.read`.

The successful response echoes the same `request_id`:

```text
type = debug_shape_catalog
request_id
shapes
```

If the room/runtime cannot provide a catalog, the tooling router returns a correlated `tooling_error` with:

```text
error_code = debug_shape_catalog_unavailable
```

The request is one-shot. There is no server subscription, periodic resend, or WebSocket send-once state.

## Catalog construction

The server derives the packet from the shared collision-shape catalog:

```text
shared/collisions/collision_shapes.json
-> physics.LoadCollisionShapeCatalog()
-> devtools.BuildShapeCatalog()
-> outbound.BuildDebugShapeCatalogPacket()
-> devtools.DebugShapeCatalogPacket
```

The packet contains reusable definitions keyed by debug shape ID. It does not contain live entity collections, collision results, damage results, targeting decisions, or gameplay authority.

Current shape ID forms are:

```text
player:<ship_type>
asteroid:<variant>
bullet
pickup:<catalog_key>
```

An empty player ship type falls back to `player:v_wing`.

`BuildShapeCatalog()` converts each imported collision shape at scale `1`, derives reusable local outline points through `physics.CollisionBodyOutlinePoints`, and skips invalid shapes or shapes with no outline points.

Live position, rotation, scale, variant, and pickup state continue to come from lane-native realtime state. The client combines those transforms with the reusable catalog for presentation-only drawing.

## Packet shape

The readout packet remains defined in `shared/packets/debug.toml`:

```go
type DebugShapeCatalogPacket struct {
    Type      string                          `json:"type"`
    RequestID string                          `json:"request_id"`
    Shapes    map[string]DebugShapeDefinition `json:"shapes"`
}
```

Each definition contains:

```text
id
kind
shape_type
points[{x, y}]
```

Example:

```json
{
  "type": "debug_shape_catalog",
  "request_id": "catalog-request-1",
  "shapes": {
    "player:v_wing": {
      "id": "player:v_wing",
      "kind": "player",
      "shape_type": "polygon",
      "points": [{"x": 12.0, "y": -11.0}]
    }
  }
}
```

Points are local reusable geometry. They are not transformed into world space by the server readout.

## Eligibility and failure behavior

`CanSendDebugShapeCatalog` requires:

```text
room is not nil
room has a game instance
devtools.Enabled() is true
room state is InGame or GameOver
```

Tooling preflight additionally requires:

```text
client-to-server packet direction
attached room
tooling.read capability
valid request_id
valid packet payload
```

Failure behavior is:

```text
collision-shape catalog load failure
-> canonical runtime asset load failure event
-> no catalog packet
-> correlated tooling_error

invalid individual shape
-> skip that shape
-> continue building remaining definitions

sr.tooling send/serialization failure
-> tooling transport/session failure handling
-> no gameplay mutation
```

`nodevtools` builds make catalog output unavailable through `devtools.Enabled() == false`.

## Client presentation boundary

The migration preserves the existing application ownership after the tooling router:

```text
ToolingPacketRouter.debug_shape_catalog_received
-> ClientConnectionService._on_debug_shape_catalog_received
-> ClientConnectionService.debug_shape_catalog_received
-> SessionNetworkController
-> GameplaySessionController
-> GameplayComposition
-> server hitbox overlay catalog store and drawing flow
```

Client devtools own:

```text
reading the packet
storing definitions
resolving live entity state to shape IDs
applying realtime transforms
converting through WorldSync
drawing diagnostic outlines
```

The hitbox checkbox only controls local presentation. It does not grant authorization and does not mutate gameplay.

## Authority boundary

The shape catalog is diagnostic metadata derived from the same shared geometry used by server collision systems. It does not decide whether a collision occurred and must not become a separate collision source of truth.

The catalog deliberately excludes:

```text
live players
live asteroids
live bullets
live pickups
collision bodies or results
damage or pickup outcomes
score, respawn, and targeting decisions
```

## Code map

```text
shared/packets/tooling.toml
shared/packets/debug.toml
shared/packets/outputs.toml
shared/collisions/collision_shapes.json
services/game-server/internal/networking/tooling/router.go
services/game-server/internal/networking/tooling/packet_contract.go
services/game-server/internal/networking/outbound/debug_shape_catalog_presentation.go
services/game-server/internal/devtools/shape_catalog.go
services/game-server/internal/devtools/shape_ids.go
services/game-server/internal/devtools/packets_generated.go
services/game-server/internal/game/physics/collision_shapes.go
services/game-server/internal/game/physics/collision_outline.go
client/scripts/networking/inbound/tooling_packet_router.gd
client/scripts/networking/client_connection_service.gd
client/scripts/devtools/hitboxes/debug_shape_catalog_packet_reader.gd
client/scripts/devtools/hitboxes/debug_shape_catalog_store.gd
client/scripts/gameplay/debug/server_hitbox_overlay_flow.gd
```

## Verification

Relevant coverage includes:

```text
services/game-server/internal/networking/outbound/debug_shape_catalog_presentation_test.go
services/game-server/internal/networking/tooling/router_test.go
services/game-server/internal/devtools/shape_catalog_test.go
services/game-server/internal/devtools/shape_ids_test.go
client/tests/unit/networking/test_tooling_packet_router.gd
client/tests/unit/test_client_connection_service.gd
client/tests/unit/devtools/hitboxes/test_debug_shape_catalog_packet_reader.gd
client/tests/unit/gameplay/debug/test_server_hitbox_overlay_flow.gd
```

Tests verify request correlation, non-empty typed catalog construction, tooling-router response delivery, preservation of the existing client signal/application path, shape ID generation, invalid-shape skipping, and client read/store behavior.

## Related docs

* [Server Devtools](./!INDEX.md)
* [Tooling Channel Migration Contract](../design/tooling-channel-migration-contract.md)
* [Devtools Packet Protocol](../design/devtools-packet-protocol.md)
* [Client Hitbox Overlays](../client/hitbox-overlays.md)
* [Client Packet Routing And Devtools Input](../client/packet-routing-and-devtools-input.md)
* [Collision Shape Data](../../data/collision-shape-data.md)

## Notes

Changes to this boundary should update its canonical owner, code map or source map, verification evidence, and related documentation in the same change.
