# Player Build And Loadouts

## Purpose

This document defines the planning direction for player builds, ship variants, weapons, and loadout composition.

## Core Model

`ShipVariant + LoadoutSelection -> ResolvedPlayerBuild -> runtime ship/session setup`

## Ship Variants

- ShipVariant owns chassis identity, movement/handling, survivability, collision shape, weapon point layout, module slot availability, slot capability, and default/fallback equipment.

## Weapon Points

- `primary_1`
- `primary_2`
- `secondary_1`
- `secondary_2`
- `hardpoint` means pre-match equip capable.
- `softpoint` means pickup-only runtime capacity.
- `none` means unavailable.
- Every valid ship has `primary_1` as a hardpoint.
- LoadoutSelection fills hardpoints only.
- Softpoints are filled only by runtime pickups.
- `primary_1` cannot be empty at match start.

## Weapon Classification

- `slot`
- `size`
- `delivery_class`
- `targeting_policy`
- `effect_flags`
- ammo policy
- WeaponProfile owns firing, projectile, damage, ammo, impact behavior, and weapon classification.
- `size` values are `light`, `standard`, and `heavy`.
- `delivery_class` values are `ballistic`, `missile`, `beam`, `mine`, `drone`, and `self`.
- `targeting_policy` values are `skill_shot`, `auto_target`, `target_lock`, `self_target`, and `area_placed`.
- `effect_flags` values are `direct`, `area`, `radial`, and `over_time`.
- `effect_flags` are composable and at least one must be present.

## Module Slots

- `shield_mod`
- `armor_mod`
- `engine_mod`
- `utility_mod`

## Loadout Selection

- LoadoutSelection chooses equipment.
- ResolvedPlayerBuild validates and combines selected ship, selected equipment, mode rules, and later availability.

## Loadout Ammunition

- Loadouts track starting ammunition, not live ammunition.
- Runtime equipment state owns current ammo, cooldowns, temporary pickup changes, and temporary overwrite state.
- Infinite ammo should not be assumed for permanently equipped weapons.

## Pickup Interaction

- Weapon pickups mutate runtime weapon state, not saved loadout.
- Same-weapon pickups increase ammunition.
- Dedicated ammunition pickups increase ammo without granting a weapon.
- Weapon pickups fill compatible empty softpoints or hardpoints first.
- If no empty compatible point exists, pickup behavior may temporarily overwrite a filled weapon point.
- Softpoint pickup weapons persist for the run.
- Hardpoint overwrites are temporary for the run.

## ResolvedPlayerBuild

- ResolvedPlayerBuild is the planned validation and composition seam that combines the selected ship, selected equipment, mode rules, and later availability into runtime ship/session setup.

## Owned Ships And Hardwired Modules

- Hardwiring modules is not part of loadout.
- Hardwired modules and owned ship instances belong to a later equipment/inventory/progression layer.
- Loadout may later select an owned ship, but does not install or remove hardwired modules.

## Shield Support Planning

- Full shield support belongs in the player-build implementation slice.
- Damage resolution already supports shield absorption.
- Player-build work should define max shield through resolved build data.
- Respawn should restore shield from resolved build data.
- HUD/state presentation should expose shield as part of the completed build support.

## Ship Variant Implementation Planning

- Real ship definitions
- Keyed multi-ship collision shape catalog
- Client ship scene mapping from `ship_type`
- Ship selection through a future loadout/readiness path
- Tests for real variants

Collision catalog planning rule:

- The current single ship collision shape should eventually become a keyed ship catalog.
- `ShipShapeByID` should remain the lookup seam.
- Malformed or unknown IDs should preserve safe fallback behavior.

Client scene mapping planning rule:

- Player rendering should eventually map `ship_type` to scene path or preload.
- Unknown ship types should use the default scene.

Selection planning rule:

- Acquisition, ownership, unlocks, purchases, and persistence must stay outside the realtime `Ship` entity.
- Future selection should flow through player build/loadout readiness, not direct client collision authority.

## Cleanup Direction

- ShipStats should not own weapon projectile tuning.
- Stale ship-side bullet cooldown, speed, lifetime, spawn-offset, and damage fields should be removed or rerouted through weapon profiles.
