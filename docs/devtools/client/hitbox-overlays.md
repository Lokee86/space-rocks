# Hitbox Overlays

Parent index: [Client](./!INDEX.md)

## Purpose

This document describes the current client-side server hitbox overlay implementation.

It covers how the Godot client presents server-derived collision outlines for development diagnostics, how the overlay is controlled, what data feeds it, and where the client boundary ends.

## Overview

The server hitbox overlay is a devtools-only visual outline layer for inspecting authoritative server collision geometry against client presentation.

The overlay is mounted in the main game scene:

```text
client/scenes/game.tscn
-> ServerHitboxOverlay
```

The mounted scene is:

```text
client/scenes/devtools/server_hitbox_overlay.tscn
```

The overlay is hidden by default. It is enabled through the devtools window checkbox labeled:

```text
Show Server Collision Telemetry
```

When enabled, the client draws outline polylines for supported server-state entities using:

* the latest normalized gameplay state
* the latest server-provided debug shape catalog
* the current `WorldSync` server-to-visual coordinate conversion

The overlay currently supports draw entries for:

* players
* asteroids
* bullets
* pickups

The overlay does not perform collision detection. It does not decide hit results, damage, pickup collection, despawn, scoring, targeting, or gameplay validity. It is a presentation-only diagnostic view over server-owned facts.

## Debug-only scope

The hitbox overlay is a client devtools surface.

It is not:

* player-facing HUD
* production gameplay UI
* a gameplay collision system
* a local prediction system
* a replacement for server collision bodies
* a packet command surface

The overlay exists to help compare rendered client entities against server collision geometry during development.

Normal gameplay entities do not draw their own debug outlines. The visible outline layer lives under the devtools scene and script path, while normal world/runtime code only provides the read-only state and coordinate conversion needed to render the diagnostic overlay.

## Server authority

The server remains authoritative for collision geometry and collision consequences.

The server sends a `debug_shape_catalog` packet when the session enters an eligible room state and server devtools are enabled. The packet contains reusable shape definitions, not live entity positions.

The server-side shape catalog is built from the loaded collision shape catalog:

```text
shared/collisions/collision_shapes.json
-> services/game-server/internal/game/physics
-> services/game-server/internal/devtools.BuildShapeCatalog
-> debug_shape_catalog packet
```

The `debug_shape_catalog` packet includes shape definitions keyed by debug shape id. Each shape definition contains:

```text
id
kind
shape_type
points
```

The shape catalog does not contain live players, asteroids, bullets, pickups, collision results, or debug collision bodies. Live entity placement comes from normal gameplay state packets.

The client combines the shape catalog with gameplay state for display only.

## Client presentation

`DevtoolsServerHitboxOverlay` owns the actual drawing.

Its draw input is an array of dictionaries with:

```text
kind
id
points
```

The overlay only draws entries when it is visible. Disabling the overlay clears cached entries and queues a redraw.

The overlay draws each valid `PackedVector2Array` as a closed polyline. Invalid entries, missing point arrays, empty outlines, and malformed point data are ignored.

The overlay has a high `z_index` so outlines render above normal world presentation.

## Shape resolution

`ServerHitboxOverlayFlow` resolves shape ids from current normalized gameplay state.

Current client shape id rules:

```text
player -> player:<ship_type>
asteroid -> asteroid:<variant>
bullet -> bullet
pickup -> pickup:<type>
```

Player shape resolution falls back to:

```text
player:v_wing
```

Asteroid shape resolution falls back to:

```text
asteroid:0
```

if the requested asteroid variant shape is missing.

If no matching shape definition is available, the entity is skipped for overlay drawing. Missing shape data does not affect gameplay state application or normal rendering.

## Data flow

The runtime flow is:

```text
server loads collision shape catalog
-> server builds debug shape catalog packet
-> client routes debug_shape_catalog packet
-> ServerHitboxOverlayFlow stores shape definitions
-> client receives normal gameplay state packets
-> GameplayStatePacketReader normalizes state
-> ServerHitboxOverlayFlow stores latest gameplay state
-> GameplayProcessFlow ticks overlay flow
-> overlay flow resolves entity shape ids
-> overlay flow transforms local shape points by entity position, rotation, and scale
-> WorldSync converts server positions to visual positions
-> DevtoolsServerHitboxOverlay draws closed outlines
```

The overlay uses normal gameplay state fields after client normalization:

```text
server_players
server_asteroids
server_bullets
server_pickups
```

It does not read live Godot entity collision nodes.

## Commands and controls

The overlay is controlled through the devtools window, not through a direct numbered dev toggle.

The control path is:

```text
DevtoolsWindow.ShowServerHitboxesCheckBox
-> show_server_hitboxes_changed
-> DevtoolsWindowController
-> DevtoolsWindowActionContext
-> DevtoolsOverlayContext.set_server_hitboxes_enabled()
-> DevtoolsServerHitboxOverlay.set_enabled()
```

The checkbox state is cached by `DevtoolsWindowController` so the UI can restore the current selected state when the devtools window is recreated.

The overlay does not send a request packet when toggled. The client only changes local visibility. The server shape catalog is sent as part of server outbound devtools presentation when eligible.

## Build and runtime gates

Client-side gates:

* `client/scripts/devtools/dev_tools_build_flags.gd` erases numbered dev-toggle input actions when `public_build` is true.
* The overlay scene is present in `game.tscn`, but starts hidden.
* The overlay only draws when explicitly enabled.
* The devtools window is the current UI control for toggling the overlay.

Server-side gates:

* `services/game-server/internal/devtools/enabled_default.go` enables devtools in default builds.
* `services/game-server/internal/devtools/enabled_nodevtools.go` disables devtools when the server is built with the `nodevtools` tag.
* `outbound.CanSendDebugShapeCatalog` only allows the shape catalog when devtools are enabled and the room is `InGame` or `GameOver`.
* `writeDebugShapeCatalogMessage` tracks the last room id and sends the catalog once per room per websocket write loop.

## Relationship to WorldSync

The overlay relies on `WorldSync` for coordinate conversion.

`ServerHitboxOverlayFlow` builds outline points in server-space first. It then calls:

```text
world_sync.visual_position_for_server_position(server_position)
```

for each transformed point.

This keeps overlay drawing aligned with the same ViewAnchor and toroidal visual coordinate basis used by normal world presentation. The overlay does not own camera anchoring, toroidal wrapping, interpolation, or render-origin policy.

## Reset behavior

The overlay flow resets when gameplay composition resets.

Reset clears:

* latest debug collision body cache
* latest gameplay state
* stored shape catalog entries
* visible overlay draw entries

The overlay itself clears entries when disabled.

This prevents stale outlines from persisting across sessions, room transitions, or gameplay resets.

## Code map

Primary client files:

```text
client/scenes/devtools/server_hitbox_overlay.tscn
client/scenes/devtools/devtools_window.tscn
client/scenes/game.tscn
client/scripts/devtools/hitboxes/devtools_server_hitbox_overlay.gd
client/scripts/devtools/hitboxes/debug_shape_catalog_packet_reader.gd
client/scripts/devtools/hitboxes/debug_shape_catalog_store.gd
client/scripts/devtools/hitboxes/debug_shape_id_resolver.gd
client/scripts/gameplay/debug/server_hitbox_overlay_flow.gd
```

Client wiring files:

```text
client/scripts/devtools/devtools_window.gd
client/scripts/devtools/devtools_window_controller.gd
client/scripts/devtools/context/devtools_overlay_context.gd
client/scripts/devtools/context/devtools_window_action_context.gd
client/scripts/devtools/gameplay_devtools_context.gd
client/scripts/gameplay/input/gameplay_input_context.gd
client/scripts/gameplay/runtime/gameplay_flow_composer.gd
client/scripts/gameplay/runtime/gameplay_process_flow.gd
```

Client packet routing files:

```text
client/scripts/networking/inbound/server_packet_router.gd
client/scripts/networking/inbound/server_packet_dispatcher.gd
client/scripts/networking/client_connection_service.gd
client/scripts/session/session_network_controller.gd
client/scripts/session/gameplay_session_controller.gd
client/scripts/shell/gameplay_shell_flow.gd
client/scripts/gameplay/gameplay_composition.gd
```

Server catalog and outbound files:

```text
services/game-server/internal/devtools/shape_catalog.go
services/game-server/internal/devtools/shape_ids.go
services/game-server/internal/devtools/packets_generated.go
services/game-server/internal/networking/outbound/debug_shape_catalog_presentation.go
services/game-server/internal/networking/websocket_write.go
```

Source data:

```text
shared/collisions/collision_shapes.json
```

Related tests:

```text
client/tests/unit/gameplay/debug/test_server_hitbox_overlay_flow.gd
client/tests/unit/gameplay/test_gameplay_flow_composer.gd
services/game-server/internal/devtools/shape_catalog_test.go
services/game-server/internal/networking/outbound/debug_shape_catalog_presentation_test.go
```

## Tests and verification

Client tests verify that `ServerHitboxOverlayFlow`:

* draws a player outline when both gameplay state and matching catalog shape data exist
* produces no draw entries when catalog shape data is missing

Client gameplay composition tests verify that the server hitbox overlay flow is reset with gameplay composition reset.

Server tests verify that `BuildShapeCatalog`:

* includes expected player, asteroid, bullet, and pickup shape ids
* skips invalid shapes
* returns outline points rather than origin-only placeholders

Server outbound tests verify that `debug_shape_catalog` responses:

* encode as JSON
* include a non-empty `shapes` object
* do not include live entity collections
* do not include `debug_collision_bodies`

## Related docs

* [Client Devtools](./!INDEX.md)
* [Devtools](../!INDEX.md)
* [Server Devtools](../server/!INDEX.md)
* [Client Gameplay Runtime](../../services/client/gameplay-runtime/!INDEX.md)
* [Client World Sync](../../services/client/world-sync/!INDEX.md)
* [Client Networking Flow](../../services/client/networking-flow/!INDEX.md)
* [Server Collision Shapes](../../services/game-server/simulation/world/collision-shapes.md)
* [Collision Shape Data](../../data/collision-shape-data.md)
* [Protocol](../../protocol/!INDEX.md)

## Notes

The UI label currently says “Show Server Collision Telemetry,” while the implementation names use “server hitbox overlay” and “hitbox” terminology.

The overlay reconstructs visible outlines from shape definitions and live state. It does not receive per-entity precomputed outline geometry from the server.

Pickup overlay drawing depends on the client-resolved pickup shape id matching a shape id in the debug shape catalog. If no matching pickup shape exists, the pickup is skipped for overlay drawing without affecting normal pickup presentation or gameplay behavior.
