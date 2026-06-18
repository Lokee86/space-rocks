# Pickup Presentation

Parent index: [World Sync](./!README.md)

## Purpose

This document describes how the Godot client presents pickup entities received from server-authoritative gameplay state.

It covers pickup scene-family selection, pickup badge/icon selection, client-side lifespan presentation, pickup node synchronization, collection-effect presentation, and the boundary between pickup presentation and server gameplay authority.

## Overview

Pickups are authoritative server entities. The server decides when pickups exist, where they are, when they expire, when they are collected, and what effect a collection applies.

The client presents those authoritative facts as Godot scene nodes.

Pickup presentation has two separate client paths:

```text
StatePacket.pickups
-> WorldSync
-> PickupSync
-> pickup scene node
```

```text
server event: pickup_collected
-> GameplayEventController
-> GameplayEffects
-> pickup_collect effect scene
```

`PickupSync` owns persistent pickup node presentation while the pickup exists in server state. It selects a pickup scene family from `pickup_class`, applies the visible badge/icon from `type`, forwards lifespan state to the pickup node, tracks target visual positions, removes stale nodes, and interpolates pickup nodes between authoritative state updates.

Collection effects are not owned by `PickupSync`. When the server emits `pickup_collected`, gameplay event presentation converts the event's server-space position into visual space and spawns a short-lived collection effect. This allows collected-sound and particle cleanup to outlive the pickup node that may be removed from world sync on the next state application.

## Code root

* `client/`

## Responsibilities

* Present server-authoritative pickup state as Godot scene nodes.
* Select pickup scene family from `pickup_class`.
* Apply pickup badge/icon visibility from pickup `type`.
* Keep pickup scene paths client-side.
* Keep pickup type/class presentation state local and transient.
* Place pickup nodes in visual coordinates relative to the active render anchor.
* Preserve pickup visual continuity across toroidal wrap boundaries.
* Track pickup target visual positions for interpolation.
* Remove local pickup nodes when their ids are absent from latest server state.
* Forward pickup age and lifespan state into pickup scene presentation.
* Play pickup spawn sound through the gameplay audio flow when a pickup node is first created.
* Expose pickup visual/server position entries for targeting read models.
* Spawn collection particles and sound from `pickup_collected` events through gameplay effects.
* Keep pickup presentation separate from pickup gameplay effects and collection authority.

## Does not own

* Pickup spawn authority.
* Pickup despawn authority.
* Pickup expiry authority.
* Pickup collection validity.
* Pickup collision outcomes.
* Pickup gameplay effect rules.
* Durable player stat, profile, or inventory mutation.
* Packet schema source-of-truth files.
* Server pickup definitions.
* Collision-shape source-of-truth data.
* HUD reward presentation.
* Devtools pickup-spawn authority.
* Drop-table rules.
* General projectile or asteroid presentation.

## Domain roles

### Pickup sync owner

`PickupSync` owns the world-sync presentation state for pickups that currently exist in `StatePacket.pickups`.

It creates pickup scene nodes, chooses the scene family through `PickupPresentationCatalog`, applies pickup type presentation to the node, forwards lifespan state, tracks server and visual positions, interpolates nodes, removes stale nodes, and exposes pickup entries for target-position consumers.

### Pickup presentation catalog

`PickupPresentationCatalog` owns the client-side mapping from pickup class to Godot scene family.

Current scene-family mapping:

```text
powerup -> res://scenes/pickups/powerup_pickup.tscn
weapon  -> res://scenes/pickups/weapon_pickup.tscn
```

The catalog also exposes available pickup types by inspecting `Badge` children in pickup scenes. Devtools uses that list for pickup selection, but devtools spawn authority remains separate.

### Pickup scene node

`client/scripts/entities/pickup.gd` owns scene-local pickup presentation behavior.

It applies badge visibility, reads collision radius from the scene shape for presentation/read-model use, stores lifespan state, runs local pulse/glow animation, runs end-of-life blinking, and delegates spawn-sound playback to the gameplay audio flow.

### Collection effect presenter

`GameplayEventController` and `GameplayEffects` own short-lived collection-effect presentation from server events.

`pickup_collected` spawns `client/scenes/pickups/pickup_collect.tscn` at the event position after converting server coordinates to visual coordinates. This path is separate from the persistent pickup node because the pickup node may already be removed by world sync.

## Protocols and APIs

### State packet pickup presentation

Pickup node presentation is driven by the pickup dictionary passed through world-sync state application.

Current application path:

```text
GameplayStatePacketReader
-> GameplayStateApplyFlow
-> GameplayWorldStateApplyFlow
-> WorldSync.apply_state(...)
-> PickupSync.remove_missing(...)
-> PickupSync.apply(...)
```

`WorldSync.apply_state(...)` receives normalized world-state dictionaries and delegates `server_pickups` to `PickupSync` after player/render-anchor state is applied.

### Packet-facing pickup fields

`PickupSyncState` reads packet-facing pickup fields from generated packet constants.

Current pickup presentation fields include:

```text
id
type
pickup_class
x
y
age_seconds
lifespan_seconds
```

`type` is the pickup identity used for badge/icon selection.

`pickup_class` is the scene-family selector.

`x` and `y` are authoritative server coordinates.

`age_seconds` and `lifespan_seconds` drive local end-of-life presentation.

### Scene-family selection

`PickupSync` asks `PickupPresentationCatalog.scene_for_class(pickup_class)` for the scene to instantiate.

Unknown pickup classes return `null` and do not create a pickup node.

The client must not receive or trust scene paths from gameplay packets. Scene paths stay client-side.

### Badge/icon selection

Pickup scenes expose a `Badge` node with child icons whose names match pickup `type` strings.

`pickup.gd` applies pickup presentation by hiding all `Badge` children, then showing the child whose name matches the pickup type:

```text
Badge/<pickup_type>
```

Current verified pickup type children include:

```text
1_up
torpedo
```

The client should not add a separate `icon_id` while `type` already names the badge child.

### Visual positioning

Pickup server positions are converted to visual positions under world sync.

For first-seen pickups, `PickupSync` positions the node relative to the active render anchor:

```text
pickup visual position =
  anchor visual position
  + shortest wrapped delta from anchor server position to pickup server position
```

For existing pickups, `PickupSync` advances visual position from the previous pickup server position to the latest pickup server position using shortest wrapped delta. This keeps pickup movement continuous if authoritative server coordinates wrap at world edges.

### Missing-node cleanup

`PickupSync.remove_missing(server_pickups)` removes any local pickup node whose id is not present in the latest server pickup dictionary.

This is presentation cleanup only. It does not decide whether the pickup was collected, expired, or otherwise removed. The server made that decision before the client received the state.

### Lifespan presentation

`PickupSync.apply(...)` forwards pickup age and lifespan to pickup nodes that implement:

```gdscript
apply_lifespan_state(age_seconds, lifespan_seconds)
```

`pickup.gd` derives remaining lifespan locally:

```text
remaining = lifespan_seconds - age_seconds
```

When remaining lifetime enters the configured end-of-life warning window, the pickup node blinks locally. The server still owns actual expiry.

### Spawn sound

When `PickupSync` creates a pickup node, it calls `play_spawn_sound(audio_flow)` if the node implements that method.

The pickup node then delegates sound playback to `GameplayAudioFlow`.

Spawn sound is scene-local presentation for node creation. It is not server authority and does not imply pickup collection.

### Collection effect event

Collection effects are driven by server events, not by pickup-node cleanup.

Current collection-effect path:

```text
server event: pickup_collected
-> GameplayEventController.apply_pickup_collected(...)
-> WorldSync.visual_position_for_server_position(...)
-> GameplayEffects.spawn_pickup_collected(...)
-> res://scenes/pickups/pickup_collect.tscn
```

`pickup_effect_applied` is currently received by the event controller but does not spawn pickup collection particles in the current implementation.

This separation keeps pickup-consumption presentation distinct from gameplay mutation feedback.

### Target read model

`PickupSync.pickup_position_entries()` exposes pickup entries to target-position consumers.

Current entry shape:

```text
visual_position
server_position
pickup_type
pickup_class
node
```

Targeting may use these entries as client presentation/read-model data. The server remains authoritative over whether a pickup target request is valid.

## Data ownership

Pickup presentation owns transient client state only.

Current local state includes:

* pickup nodes keyed by server pickup id
* pickup types keyed by server pickup id
* pickup classes keyed by server pickup id
* initialized-pickup flags
* target visual positions
* last known pickup server positions
* last known pickup visual positions
* pickup lifespan state stored on pickup nodes
* pickup pulse/glow/blink state stored on pickup nodes
* short-lived collection effect nodes

This state is not durable.

It is not authoritative.

It is reset or replaced as gameplay state and session lifecycle require.

## Scene requirements

Current pickup scenes are expected to provide the scene shape consumed by `pickup.gd`.

Required presentation nodes:

```text
GlowSprite2D
Badge
CollisionShape2D
AudioStreamPlayer2D
```

Current scene roots:

```text
PowerupPickup
WeaponPickup
```

Current pickup scenes:

```text
client/scenes/pickups/powerup_pickup.tscn
client/scenes/pickups/weapon_pickup.tscn
```

Each pickup scene's `Badge` node should contain child icons named exactly like packet pickup type strings.

Examples:

```text
Badge/1_up
Badge/torpedo
```

`CollisionShape2D` is hidden in the scene and used for read-model/presentation radius access. It is not client collision authority.

## Code map

### World sync path

* `client/scripts/world/world_sync.gd`
* `client/scripts/world/pickup_sync.gd`
* `client/scripts/world/pickup_sync_state.gd`
* `client/scripts/world/world_wrap.gd`

### Pickup presentation catalog

* `client/scripts/world/pickups/pickup_presentation_catalog.gd`

### Pickup scene behavior

* `client/scripts/entities/pickup.gd`
* `client/scenes/pickups/powerup_pickup.tscn`
* `client/scenes/pickups/weapon_pickup.tscn`

### Collection effect presentation

* `client/scripts/gameplay/events/gameplay_event_controller.gd`
* `client/scripts/gameplay/events/gameplay_event_flow.gd`
* `client/scripts/gameplay/effects/gameplay_effects.gd`
* `client/scripts/gameplay/audio/gameplay_audio_flow.gd`
* `client/scenes/pickups/pickup_collect.tscn`

### Runtime callers

* `client/scripts/gameplay/runtime/gameplay_world_state_apply_flow.gd`
* `client/scripts/gameplay/runtime/gameplay_runtime_context.gd`
* `client/scripts/gameplay/state/gameplay_state_apply_flow.gd`

### Targeting consumers

* `client/scripts/gameplay/targeting/target_position_source.gd`
* `client/scripts/gameplay/targeting/gameplay_target_candidate_flow.gd`

### Generated data

* `client/scripts/generated/networking/packets/packets.gd`
* `client/scripts/generated/constants/constants.gd`

### Source-of-truth boundaries

* `shared/packets/gameplay.toml`
* `shared/constants/pickups.toml`
* `shared/constants/weapon_pickups.toml`
* `shared/constants/client/presentation.toml`
* `shared/collisions/collision_shapes.json`

## Tests

Relevant tests include:

* `client/tests/unit/test_pickup_sync.gd`
* `client/tests/unit/test_pickup_sync_state.gd`
* `client/tests/unit/test_pickup.gd`
* `client/tests/unit/world/test_pickup_presentation_catalog.gd`
* `client/tests/unit/test_world_sync.gd`
* `client/tests/unit/test_world_wrap.gd`
* `client/tests/unit/gameplay/test_gameplay_event_controller.gd`
* `client/tests/unit/gameplay/effects/test_gameplay_effects.gd`
* `client/tests/unit/gameplay/test_gameplay_target_candidate_flow.gd`

Use the normal client GUT verification flow when changing pickup presentation behavior.

## Related docs

* [World Sync](./!README.md)
* [World Sync Coordinator](./world-sync-coordinator.md)
* [Entity Sync Owners](./entity-sync-owners.md)
* [View Anchor And Visual Coordinates](./view-anchor-and-visual-coordinates.md)
* [Gameplay Runtime](../gameplay-runtime/!README.md)
* [Gameplay State Application](../gameplay-runtime/gameplay-state-application.md)
* [Input and Targeting](../input-and-targeting.md)
* [HUD and Gameplay UI](../hud-and-gameplay-ui.md)
* [Gameplay packets](../../../protocol/stubs/gameplay-packets.md) - Stub: gameplay realtime packet documentation.
* [Pickup entities](../../../systems-design/entities/stubs/pickup-entities.md) - Stub: pickup entity design documentation.
* [Collision shape data](../../../data/stubs/collision-shape-data.md) - Stub: collision-shape source-of-truth documentation.

## Notes

Pickup presentation should stay under world sync because live pickup nodes are rendered from server world state and anchored to ViewAnchor visual coordinates.

Collection effects are documented here only because they are part of pickup presentation behavior. Their implementation owner is gameplay event/effects presentation, not `PickupSync`.

Do not collapse pickup, projectile, and asteroid presentation into one generic rendering implementation. Their sync pattern overlaps, but their packet-facing fields, scene selection, and presentation details are separate.
