---
author: brian
created: "2026-07-19"
document_id: 019f7d55-fb2c-765a-b7a6-a82f5e526f9f
document_type: general
policy_exempt: false
summary: This document describes the client background and viewport presentation flow.
---
# Background And Viewport Presentation

Parent index: [Presentation Flow](./!INDEX.md)

## Purpose

This document describes the client background and viewport presentation flow.

It explains how the Godot client wires the gameplay camera, ViewAnchor, parallax background layers, world-space shader sampling, and viewport-adjacent presentation behavior without owning world-sync coordinate math or server visibility authority.

## Overview

Background and viewport presentation is client-owned visual behavior.

The root scene provides three repeated background texture layers and a `ViewAnchor` node with a child `Camera2D`. `AppEntry` wires those nodes into `BackgroundController`, makes the `ViewAnchor/Camera2D` camera current at startup, and passes the same `ViewAnchor` into gameplay/session/world-sync composition.

The current flow is:

```text
client/scenes/game.tscn
-> Game root
-> ParallaxBackground layers
-> RepeatedBackground / RepeatedForegroundBackground / RepeatedPlanetBackground
-> ViewAnchor
   -> Camera2D

AppEntry._ready()
-> BackgroundController.configure(..., view_anchor, gameplay_camera)
-> GameplaySessionController.configure(..., view_anchor, ...)
-> AppEntry._make_view_anchor_camera_current()
```

`BackgroundController` owns the node-facing background presentation lifecycle. It creates `GameplayBackgroundFlow`, passes the repeated texture nodes, parallax target, and active gameplay camera, and calls `process_frame()` each frame.

`GameplayBackgroundFlow` owns background world-sampling presentation. It reads the parallax target's `global_position` and `Camera2D.zoom`, then writes camera world position, zoom, per-layer parallax, and world-unit offsets to each background layer's `ShaderMaterial`. The shader reconstructs the visible world position around the camera center and samples repeating textures in world units rather than anchoring texture phase or scale to the window.

The background follows the same `ViewAnchor` basis as camera and world presentation. It does not choose the active render anchor. Detailed ViewAnchor, render-anchor, visual coordinate, and toroidal wrap behavior belongs to world-sync documentation.

## Code root

```text
client/scenes/game.tscn
client/scripts/shell/app_entry.gd
client/scripts/presentation/background/
```

## Responsibilities

Background and viewport presentation owns:

* wiring repeated background texture nodes from the root scene
* wiring the background parallax target
* wiring the active gameplay camera into background presentation
* making the `ViewAnchor/Camera2D` camera current during app entry
* advancing per-frame background drift offsets
* sampling the current parallax target position
* preserving the last valid parallax position when the target reference is unavailable
* combining parallax target position with generated background constants
* writing camera world position and zoom to background materials
* writing per-layer parallax factors and world-unit offsets
* sampling repeated textures in camera-centered world coordinates
* keeping texture scale and phase independent of window dimensions
* resetting background offsets when requested
* keeping background presentation aligned with the client render anchor seam

## Does not own

Background and viewport presentation does not own:

* server-authoritative world position
* toroidal wrap math
* ViewAnchor server/visual coordinate conversion
* active render-anchor selection
* local player identity
* spectate authority
* gameplay packet parsing
* gameplay state application
* entity interpolation
* projectile, asteroid, pickup, or player sync ownership
* HUD layout or gameplay UI ownership
* server camera or visibility behavior
* `client_config` packet sending
* graphics settings or durable user preferences
* packet schema source-of-truth files
* background art source files

## Domain roles

### Root scene presentation anchors

`client/scenes/game.tscn` owns the stable node anchors used by the background and viewport presentation flow:

```text
ParallaxBackground
BackgroundLayer
RepeatedBackground
ForegroundBackgroundLayer
RepeatedForegroundBackground
PlanetBackgroundLayer
RepeatedPlanetBackground
ViewAnchor
ViewAnchor/Camera2D
```

The background layers are scene-level presentation nodes. They are not gameplay entities and are not synchronized from server state.

### ViewAnchor camera carrier

`Camera2D` lives under `ViewAnchor`.

`AppEntry._make_view_anchor_camera_current()` resolves `ViewAnchor/Camera2D` and calls `make_current()` during startup. That makes the active gameplay camera follow the same node that world-sync uses as the client render origin.

This file documents the startup camera handoff only. The ViewAnchor's server/visual mapping and active anchor updates are owned by world sync.

### Background controller

`BackgroundController` is the node-owned background presentation controller.

It is created by `AppEntry`, added as a child node, configured with the repeated background `TextureRect` nodes, the `ViewAnchor`, and `ViewAnchor/Camera2D`, and processed by Godot each frame.

Its public surface is intentionally small:

```text
configure(...)
set_parallax_target(...)
reset_background()
```

The controller delegates scroll calculation and shader writes to `GameplayBackgroundFlow`.

### Background flow

`GameplayBackgroundFlow` is the focused background calculation seam.

It stores:

```text
repeated_background
repeated_foreground_background
repeated_planet_background
parallax_target
camera
background_drift_offset
foreground_drift_offset
planet_drift_offset
last_valid_parallax_position
```

Each frame, it:

```text
1. Advances world-unit drift offsets from generated constants.
2. Reads parallax_target.global_position when available.
3. Falls back to last_valid_parallax_position when the target is missing.
4. Reads the current uniform camera zoom.
5. Writes camera_world_position and camera_zoom to every layer.
6. Writes each layer's parallax_factor.
7. Writes each layer's world-unit drift and authored offset as layer_world_offset.
```

### Repeated shader layers

The scene uses repeated background `TextureRect` nodes with `ShaderMaterial` materials.

`GameplayBackgroundFlow` expects each configured `TextureRect` to have a `ShaderMaterial` that accepts:

```text
base_tile_world_size
camera_world_position
camera_zoom
parallax_factor
layer_world_offset
```

The shader converts each fragment's distance from the viewport center into world distance by dividing by `camera_zoom`. It adds the parallax-scaled camera world position and the layer's world-unit offset, then samples the repeating texture using `base_tile_world_size`. Window dimensions therefore only determine how much of the world-space texture is visible; they do not determine texture scale or phase. If a texture node or shader material is missing, the parameter write no-ops for that layer.

### Parallax constants

Background presentation uses generated client constants from:

```text
client/scripts/generated/constants/constants.gd
```

The current background constants are generated from:

```text
shared/constants/client/presentation.toml
```

Relevant generated values include:

```text
BACKGROUND_PARALLAX
FOREGROUND_BACKGROUND_PARALLAX
PLANET_BACKGROUND_PARALLAX
BACKGROUND_DRIFT_PER_FRAME
FOREGROUND_BACKGROUND_DRIFT_PER_FRAME
PLANET_BACKGROUND_DRIFT_PER_FRAME
FOREGROUND_BACKGROUND_OFFSET
PLANET_BACKGROUND_OFFSET
```

## Protocols and APIs

This flow does not define a network protocol or HTTP API.

It consumes scene nodes, generated constants, and Godot presentation APIs:

```text
Node2D.global_position
Camera2D.make_current()
Camera2D.zoom
ShaderMaterial.set_shader_parameter("camera_world_position", value)
ShaderMaterial.set_shader_parameter("camera_zoom", value)
ShaderMaterial.set_shader_parameter("parallax_factor", value)
ShaderMaterial.set_shader_parameter("layer_world_offset", value)
```

Viewport-size reporting is adjacent but separate. The `client_config` packet flow lives in [Client Viewport Config Flow](../app-shell-and-session/client-viewport-config-flow.md). That flow reports visible viewport dimensions to the server. This document only covers local visual camera/background presentation.

## Data ownership

Background and viewport presentation owns transient client presentation state only.

Current transient state includes:

```text
background_drift_offset
foreground_drift_offset
planet_drift_offset
last_valid_parallax_position
active parallax target reference
shader camera world-position values
shader camera zoom values
shader parallax factors
shader layer world offsets
```

This state is:

```text
client-local
non-authoritative
not persisted
safe to reset between sessions
derived from scene state and generated constants
```

The source of presentation tuning values is:

```text
shared/constants/client/presentation.toml
```

Generated client output is:

```text
client/scripts/generated/constants/constants.gd
```

The background flow must not treat generated constants as runtime-owned mutable data.

## Code map

### Primary implementation files

```text
client/scripts/presentation/background/background_controller.gd
client/scripts/presentation/background/background_flow.gd
```

### Scene and composition files

```text
client/scenes/game.tscn
client/scripts/shell/app_entry.gd
```

### Shader and assets

```text
client/shaders/repeating_background.gdshader
client/assets/background.png
client/assets/backgroun-fore.png
client/assets/background-planets.png
```

### Generated and source data

```text
shared/constants/client/presentation.toml
client/scripts/generated/constants/constants.gd
```

### Related implementation boundaries

```text
client/scripts/world/world_sync.gd
client/scripts/world/player_render/player_render_api.gd
client/scripts/world/player_render/view_anchor_sync.gd
client/scripts/session/client_config_controller.gd
client/scripts/config/client_viewport_config_flow.gd
```

### Important non-ownership boundaries

```text
client/scripts/world/
```

Owns ViewAnchor server/visual mapping, render-anchor selection, toroidal shortest-delta conversion, and entity sync.

```text
client/scripts/config/client_viewport_config_flow.gd
```

Owns viewport-size packet sending. It does not own local camera or background presentation.

```text
client/scripts/shell/gameplay_hud_flow.gd
client/scripts/ui/hud/
```

Owns HUD presentation. Background presentation must not become a HUD layout or gameplay UI owner.

## Tests

Focused background-flow tests cover camera world-position and zoom propagation, distinct layer parallax and world offsets, explicit scroll-reference updates, and the neutral zoom fallback when no camera is configured:

```text
client/tests/unit/background/test_background_flow.gd
```

Adjacent tests cover the systems that background and viewport presentation depends on:

```text
client/tests/unit/ui/menu_flow/test_app_entry_menu_flow.gd
client/tests/unit/test_world_sync.gd
client/tests/unit/world/player_render/test_player_render_api.gd
client/tests/unit/world/player_render/test_view_anchor_sync.gd
client/tests/unit/test_local_visual_sync.gd
client/tests/unit/test_visual_sync_positions.gd
client/tests/unit/test_world_wrap.gd
client/tests/unit/test_room_session_controller.gd
client/tests/unit/test_shell_boot_flow.gd
```

Manual verification should include:

```text
camera is current after app boot
background scrolls during gameplay
background follows ViewAnchor movement
mouse-wheel zoom scales ships and all three background layers together
resizing the window changes only the visible crop, not star or planet size
background texture phase remains anchored to the camera world position across window-size changes
zoom remains centered instead of shifting the texture pattern toward a corner
background does not visibly jump across world-wrap edges
foreground and planet layers use their distinct parallax offsets
returning to gameplay after menu/session transitions does not orphan background updates
```

## Related docs

* [Presentation Flow](./!INDEX.md)
* [Client](../!INDEX.md)
* [App Entry Composition](../app-shell-and-session/app-entry-composition.md)
* [Client Viewport Config Flow](../app-shell-and-session/client-viewport-config-flow.md)
* [Gameplay Runtime](../gameplay-runtime/!INDEX.md)
* [World Sync](../world-sync/!INDEX.md)
* [View Anchor And Visual Coordinates](../world-sync/view-anchor-and-visual-coordinates.md)
* [Hud And Gameplay UI](../hud-and-gameplay-ui.md)
* [Constants pipeline](../../../data/data-sync-and-ssot-pipeline.md)

## Notes

Legacy documentation correctly identified the core invariant: camera and background should follow `ViewAnchor`, not the local player node directly.

`ParallaxBackground` and `ParallaxLayer` nodes exist in the scene, but the current implementation drives layer motion through world-space shader inputs rather than relying on Godot parallax layer motion. All three background layers use `client/shaders/repeating_background.gdshader`, with tile dimensions expressed as world units. The layers currently use zero motion scale and oversized texture rectangles that provide screen coverage while the shader performs the infinite repetition.

`GameplayBackgroundFlow.set_scroll_reference()` can write an explicit camera world position without accumulated drift, but the normal runtime path uses `process_frame()` with the configured parallax target.

The background flow keeps `last_valid_parallax_position` so a temporarily missing parallax target does not immediately reset scroll sampling to zero during a frame. Reset behavior is explicit through `clear()`.

Future camera behavior should preserve the same boundary: camera/background consumers follow the active ViewAnchor seam, while world sync owns how that anchor is selected and converted.
