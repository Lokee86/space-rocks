# Player Build And Loadouts
Parent index: [Gameplay Planning](./!README.md)

## Purpose

This doc plans the player-build seam for ship variants, weapon-point rules, module slots, loadout selection, and match-start build resolution.

## Ownership Boundary

This doc owns:

- ShipVariant model
- weapon point rules
- weapon classification
- module slots
- BuildEligibility
- EligibleBuildOptions
- LoadoutSelection
- ResolvedPlayerBuild
- starting ammo planning
- pickup interaction planning
- hardwired module relationships
- shield planning
- ship variant implementation planning
- cleanup direction

This doc does not own:

- owned inventory state
- hangar persistence
- mode rules
- progression reward policy
- runtime equipment state
- UI layout

Inventory and hangar owns owned state.

BuildEligibility consumes inventory snapshots.

ResolvedPlayerBuild compiles match-start setup.

## Core Model

`ShipVariant + LoadoutSelection -> ResolvedPlayerBuild -> runtime ship/session setup`

Core build flow:

`ModeRules + HangarInventory + Progression/Unlocks -> BuildEligibility -> EligibleBuildOptions -> LoadoutSelection -> ResolvedPlayerBuild -> RuntimeShip / RuntimeEquipmentState`

Core rule:

- Ineligible ships, weapons, modules, and loadouts should not be selectable in normal UI flow.
- Server validation remains the authoritative safety net.
- Server validation should not be the primary UX path for rejecting normal loadout choices.

Modes can restrict build selection before loadout selection, but they do not own the inventory or hangar model.

## Ship Variants

- ShipVariant owns chassis identity, `weight_class`, movement/handling, survivability, collision shape, weapon point layout, module slot availability, slot capability, and default/fallback equipment.
- `weight_class` values are `light`, `standard`, and `heavy`.
- `weight_class` is a chassis and loadout compatibility classification, not a physics mass value.
- `weight_class` belongs to `ShipVariant`, not `ShipStats`.
- Build validation may use ship `weight_class` together with weapon `size` and module `size` rules.
- Ship `weight_class` is separate from weapon and module `size`.
- `v_wing` uses `standard`.
- `scout` uses `light`.
- `heavy` uses `heavy`.

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

These are the four persistent module slots used by the planned build model.
They are not weapon points.

## Module Model

- `ModuleProfile` owns module identity, slot, class/category, activation type, stat modifiers, penalties/tradeoffs, runtime behavior if active, and compatibility tags.
- Passive modules modify `ResolvedPlayerBuild`.
- Active modules add runtime equipment behavior.
- Loadout modules are selected pre-match and validated with the loadout.
- Hardwired modules are not selected in loadout; they belong to future `OwnedShip`/inventory/progression state.
- `ResolvedPlayerBuild` eventually combines ship variant, selected loadout modules, selected weapons, mode rules, and hardwired owned-ship effects.

## Module Slot Meaning

- `shield_mod`: max shield, shield regen, shield recharge behavior, shield pulse later
- `armor_mod`: max health, resistance, collision durability, possible mass/speed penalty
- `engine_mod`: thrust, max speed, rotation, damping, boost/dash later
- `utility_mod`: pickup radius, ammo reserves, cooldown modifiers, targeting assist, scanner/radar, drone capacity

## BuildEligibility

`BuildEligibility` is the seam that computes selectable build options before the player confirms a loadout.
Mode rules can restrict build selection before loadout selection.

It combines:

- mode rules
- player hangar/inventory
- progression/unlocks later
- ship variants
- weapon profiles
- module profiles

Responsibilities:

- filter eligible owned ships
- filter eligible weapons per weapon point
- filter eligible modules per module slot
- explain why options are blocked
- provide selectable options to the client
- validate submitted `LoadoutSelection` server-side

`BuildEligibility` is the planned source of selectable build options.
It does not make the client authoritative.
It does not move eligibility discovery into `ResolvedPlayerBuild`.

## EligibleBuildOptions

`EligibleBuildOptions` is the authoritative selectable option set produced before player selection.

Likely fields:

- `mode_id`
- `player_id`
- `eligible_ships`
- `selected_ship_id` or default selection
- `weapon_options_by_point`
- `module_options_by_slot`
- `blocked_reasons`
- `fallback_loadout`

Blocked options may be shown disabled or omitted by the client.
Disabled reasons should be machine-readable enough for UI display later.

### Mode Rule Restriction Categories

Mode rules can narrow or reshape eligibility before the player selects a loadout.

Ship restriction categories:

- allowed ship types
- banned ship types
- allowed `weight_class` values
- required ship traits later

Weapon restriction categories:

- allowed weapon slots
- allowed weapon sizes
- allowed delivery classes
- allowed targeting policies
- allowed effect flags
- banned weapon IDs
- banned weapon categories

Module restriction categories:

- allowed module slots
- allowed module classes
- passive-only / active-allowed
- banned module IDs
- hardwired module behavior allowed, disabled, or normalized

Ammo and equipment restriction categories:

- starting ammo limits
- infinite ammo allowed or blocked
- pickup overwrite allowed or blocked later

Mode examples only:

- `survival_arcade`: broad/default eligibility
- `score_attack`: broad/default eligibility, maybe blocks debug or experimental modules
- `ranked_pvp` later: blocks progression power modules, may normalize hardwired modules
- `campaign mission` later: may require or block certain ship classes or utility modules

## Loadout Selection

- LoadoutSelection is the selected build only.
- `selected_owned_ship_id`
- `selected_ship_type` fallback only if owned ships do not exist yet
- `selected_weapons_by_point`
- `selected_modules_by_slot`
- `selected_starting_ammo` choices if needed
- LoadoutSelection does not carry the full option universe.
- LoadoutSelection does not perform eligibility filtering.
- Selection happens before resolution.
- ResolvedPlayerBuild validates and combines the selected build, mode rules, and later availability after selection.

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

- ResolvedPlayerBuild is the immutable match-start setup seam.
- It receives an eligible `LoadoutSelection`, verifies it again, and compiles the result into runtime setup.
- ResolvedPlayerBuild should not store live mutable state.

Likely fields:

- `player_id`
- `mode_id`
- `selected_owned_ship_id`, if present
- `ship_type`
- `weight_class`
- `resolved_ship_stats`
- `collision_shape_id`
- `weapon_point_layout`
- `equipped_weapons`
- `equipped_modules`
- `applied_passive_effects`
- `active_module_declarations`
- `shield_policy`
- `starting_equipment_state`
- validation/source summary

Fields that belong in runtime state instead:

- current health
- current shield
- current ammo
- cooldown remaining
- temporary pickup overwrites
- active module cooldown timers
- active buff/debuff timers

Runtime output split:

- `RuntimeShip` owns current health, current shield, position, velocity, rotation, death/respawn state.
- `RuntimeEquipmentState` owns current ammo, cooldowns, active module charges/cooldowns, pickup overwrites, and temporary softpoint weapons.

Respawn restores from `ResolvedPlayerBuild` and mode/runtime rules.

## Owned Ships And Hardwired Modules

- Hardwiring is not part of loadout.
- Hardwired modules and owned ship instances belong to a later equipment/inventory/progression layer.
- Loadout may later select an `OwnedShip`, but does not install or remove hardwired modules.
- OwnedShip -> BuildEligibility -> LoadoutSelection -> ResolvedPlayerBuild is the later flow.
- Mode rules can allow, disable, or normalize hardwired effects.
- `ResolvedPlayerBuild` applies hardwired effects only after owned ships exist and only if mode rules allow them.

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

## Related Docs

- [Inventory And Hangar](inventory-and-hangar.md)
- [Modes And Match Rules](modes-and-match-rules.md)
- [Progression And Rewards](progression-and-rewards.md)

## Open Gametime Decisions

- Exact ship-variant catalog shape.
- Exact build-option field names for future UI.
- Exact owned-ship selection shape when inventory-backed ship selection lands.
- Exact shield stat and shield-regen policy placement.
- Exact cleanup ordering for old ship-side weapon fields.

## Core Invariants

- Inventory owns owned state.
- BuildEligibility reads inventory snapshots and mode rules.
- LoadoutSelection is the player choice, not the runtime state.
- ResolvedPlayerBuild is the only match-start build object consumed by runtime setup.
- Runtime equipment owns ammo, cooldowns, and pickup-overwrite state.
- Hardwired modules stay separate from pre-match loadout selection.
- `weight_class` stays a loadout-compatibility classification, not physics mass.
- `primary_1` remains required at match start.
