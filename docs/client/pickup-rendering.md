# Client Pickup Rendering

This document describes the current client-side pickup rendering contract.

## Data Flow

- The server sends pickup state in `StatePacket.pickups`.
- `PickupState.type` is the gameplay identity and the icon node name.
- `PickupState.pickup_class` is the scene-family selector.

## Scene Selection

- `pickup_class = "powerup"` uses `res://scenes/pickups/powerup_pickup.tscn`.
- `pickup_class = "weapon"` uses `res://scenes/pickups/weapon_pickup.tscn`.

## Required Scene Shape

Each pickup scene must contain:

- `GlowSprite2D`
- `Badge`
- `CollisionShape2D`
- `AudioStreamPlayer2D`

The root should be the appropriate pickup scene root:

- `PowerupPickup`
- `WeaponPickup`

The `Badge` node must contain children named exactly like packet type strings, for example:

- `1_up`
- `torpedo`

`pickup.gd` hides all `Badge` children and then shows `Badge/<pickup_type>`.

## Ownership

- Server owns type and class authority.
- Client owns scene selection, icon visibility, pulse/glow, lifespan blink, audio scene nodes, interpolation, and target-position presentation.

## Do Not

- Do not add client type-to-class maps.
- Do not send scene paths over gameplay packets.
- Do not add `icon_id` while `type` already names the icon node.
- Do not make `world_sync` own pickup presentation rules.
