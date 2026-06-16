# Inventory And Hangar

## Purpose

This doc plans the durable inventory and hangar architecture for player-held ships, weapons, modules, hardwired equipment, unlock/access state, acquisition state, and the handoff into build eligibility.

The inventory and hangar layer answers:

```text
What does this player own?
What content has this player unlocked or gained access to?
Which owned ship instances exist?
Which owned equipment instances exist?
Which hardwired equipment is attached to each owned ship?
What durable or transient inventory state should BuildEligibility consume?
```

This doc focuses on ownership state and hangar flow.

It does not own match runtime equipment state, final build eligibility, progression reward policy, shop pricing, or UI layout.

## Ownership Boundary

This doc owns:

```text
HangarInventory
OwnedShip
OwnedWeapon
OwnedModule
hardwired equipment attached to OwnedShip
stackable inventory items
unlock/access state relevant to inventory and builds
acquisition metadata
grant application into owned inventory
guest transient inventory behavior
fallback/default hangar behavior
handoff to BuildEligibility
```

This doc does not own:

```text
XP, rank, or reward formula policy
GrantAward construction policy
shop prices or currency sinks
real-money purchases
match result facts
mode rule definitions
loadout eligibility filtering
ResolvedPlayerBuild compilation
runtime ammo
runtime cooldowns
runtime pickup overwrites
softpoint pickup weapons
current health or shield
temporary buffs/debuffs
UI hangar layout
physical database table details
```

Progression and rewards owns reward evaluation and `GrantAward` construction.

Commerce owns pricing, purchases, receipts, refunds, and shop catalog policy.

Player-data owns identity routing and persistence routes.

BuildEligibility consumes the normalized inventory snapshot.

## Core Architecture

Core acquisition flow:

```text
Trusted Source Event
-> Progression / Commerce / Entitlement Policy
-> GrantAward
-> Player-Data Runtime Route
-> Inventory Grant Application
-> HangarInventory
```

Inventory is not the final eligibility system.

Inventory records ownership, access, instance state, and broad availability facts.
It does not resolve build outcomes.

## Persistent Versus Runtime State

Persistent or durable inventory/hangar state:

```text
owned ship instances
owned weapon instances
owned module instances
hardwired equipment attached to owned ship instances
stackable inventory items
unlock/access state relevant to builds or purchases
acquisition metadata
grant/application metadata needed for dedupe or audit
default owned ship reference
saved loadout references later
```

Runtime-only state excluded from inventory/hangar persistence:

```text
softpoint pickup weapons
temporary hardpoint overwrites
same-weapon pickup ammo increases
dedicated ammo pickups
current ammo
current cooldowns
current health
current shield
temporary buffs/debuffs
runtime-only rare drops
normal pickup effects
active match equipment timers
```

Runtime weapon pickups mutate runtime equipment state only.

Runtime pickups do not create ownership unless the pickup is explicitly promoted into a persistent rare-drop collection flow that emits a durable grant.

## HangarInventory Model

`HangarInventory` is the normalized player inventory and hangar state used by inventory operations and BuildEligibility.

Planned shape:

```text
HangarInventory
- player_ref
- owned_ships[]
- owned_weapons[]
- owned_modules[]
- unlocked_content[]
- stackable_items[]
- default_owned_ship_id optional
- saved_loadout_refs later
- metadata optional
```

`HangarInventory` should be treated as player-owned state, not match state.

It may be backed by guest transient memory, local SQLite, or Rails/Postgres depending on identity and player-data routing.

The fallback/default hangar must keep missing or corrupt data from making the game unplayable.

## Owned Item Identity

All owned ships, weapons, and modules are instanced.

Catalog identity and owned instance identity are separate.

Examples:

```text
owned_ship_id = owned instance id
ship_id = catalog/profile id

owned_weapon_id = owned instance id
weapon_id = catalog/profile id

owned_module_id = owned instance id
module_id = catalog/profile id
```

Owned instance IDs should be compact generated IDs.

They should not be long semantic strings.

Catalog refs remain readable content refs such as:

```text
ship.v_wing
ship.scout
weapon.railgun
module.overcharger
```

A player may own multiple instances of the same ship, weapon, or module catalog item.

V0 does not need to generate many duplicates, but the model should allow them structurally.

## OwnedShip Model

Ships are owned instances.

Recommended shape:

```text
OwnedShip
- owned_ship_id
- ship_id
- acquired_at
- acquisition_ref
- hardwired_equipment[]
- state
- metadata optional
```

`ship_id` points to a ship catalog/profile entry.

`owned_ship_id` identifies the player's specific ship instance.

`acquisition_ref` links the owned ship back to the grant, purchase, achievement, entitlement, migration, admin grant, or other acquisition source that created it.

`state` is a logical state field for normal, unavailable, disabled, reversed, or other future status if needed.

Exact state vocabulary is a gametime persistence decision.

## OwnedWeapon Model

Weapons are owned instances.

Recommended shape:

```text
OwnedWeapon
- owned_weapon_id
- weapon_id
- acquired_at
- acquisition_ref
- state
- metadata optional
```

Owned weapons are reusable equipment instances unless a later design explicitly changes that rule.

The same owned weapon instance may be referenced by multiple saved loadouts.

Loadouts do not globally reserve a weapon instance.

Only the selected loadout matters at match start.

Runtime weapon pickups do not create `OwnedWeapon` records.

Persistent weapon drops create ownership only through durable grants.

## OwnedModule Model

Modules are owned instances.

Recommended shape:

```text
OwnedModule
- owned_module_id
- module_id
- acquired_at
- acquisition_ref
- state
- metadata optional
```

Owned modules are reusable equipment instances unless a later design explicitly changes that rule.

The same owned module instance may be referenced by multiple saved loadouts or normal install contexts.

Loadouts do not globally reserve a module instance.

Hardwired equipment is the exception and is not reusable general owned equipment once hardwired.

## Hardwired Equipment

Hardwired equipment is attached ship state.

It is not normal reusable inventory equipment.

When equipment becomes hardwired, it is conceptually removed from general reusable ownership and attached to one specific `OwnedShip`.

Hardwired equipment cannot be installed on multiple ships at the same time.

Planned shape inside `OwnedShip`:

```text
HardwiredEquipment
- hardwired_id
- source_equipment_ref
- equipment_id
- hardwired_slot optional
- installed_at
- acquisition_ref or install_ref
- stat_modifiers
- behavior_refs optional
- metadata optional
```

Hardwiring mutates `OwnedShip` pre-game stats or effects.

Mode rules may later:

```text
allow hardwired effects
disable hardwired effects
normalize hardwired effects
block ships with specific hardwired equipment
block hardwired equipment categories
```

Commerce may later make hardwire install/removal a currency sink.

That is a gametime economy decision, not required for the ownership seam.

## Stackable Inventory Items

Some grants may create stackable inventory state rather than owned equipment instances.

Examples:

```text
ship parts
crafting materials
repair charges
consumables
event tokens
upgrade materials later
```

Recommended shape:

```text
StackableInventoryItem
- item_ref
- quantity
- updated_at
- metadata optional
```

Stackable grants require idempotency receipts so replayed grants do not double-apply quantities.

Exact physical receipt storage belongs to player-data persistence.

## Unlocks, Access, And Ownership

Unlocks do not equal ownership.

Unlocks allow purchase, access, eligibility, or appearance in acquisition paths.

Ownership comes from grants.

Purchases produce grants.

Non-purchase rewards can also produce ownership grants.

Examples:

```text
unlock ship.scout
-> scout is now allowed to appear as purchasable, earnable, selectable, or otherwise accessible content depending on rules

inventory_item ship.scout
-> player receives an OwnedShip instance

unlock weapon.railgun
-> railgun is accessible for purchase or grant paths

inventory_item weapon.railgun
-> player receives an OwnedWeapon instance

inventory_item module.overcharger
-> player receives an OwnedModule instance
```

BuildEligibility should not read raw progression internals directly where a normalized inventory/access snapshot can be used instead.

Inventory/hangar should expose build-relevant unlock/access state together with owned state.

## Grant Application

Inventory/hangar applies inventory-related grants after player-data routing selects the correct identity and store path.

Inventory/hangar does not decide whether the player deserved the grant.

Grant policy belongs to progression, rewards, commerce, entitlement, admin, migration, or refund systems.

Inventory/hangar applies valid routed grants idempotently.

Relevant grant kinds include:

```text
unlock
inventory_item
ship_part
rare_drop
entitlement
reversal
```

Examples:

```text
Grant unlock weapon.railgun
-> add weapon.railgun to unlocked_content if missing

Grant inventory_item ship.scout x1
-> create one OwnedShip instance with ship_id = ship.scout

Grant inventory_item weapon.railgun x1
-> create one OwnedWeapon instance with weapon_id = weapon.railgun

Grant inventory_item module.overcharger x1
-> create one OwnedModule instance with module_id = module.overcharger

Grant ship_part item.coolant_fragment x3
-> increment StackableInventoryItem quantity for item.coolant_fragment
```

New ownership grants target catalog refs.

Mutation or reversal grants may target owned instance refs.

Examples:

```text
inventory_item ship.scout
-> create new OwnedShip instance

inventory_item weapon.railgun
-> create new OwnedWeapon instance

reversal owned_weapon.<owned_weapon_id>
-> remove, disable, reverse, or otherwise mark that owned weapon instance according to persistence policy
```

Exact reversal behavior is a gametime decision.

## Acquisition Sources

Inventory changes come from grants routed through the owning systems, including match, objective, mission, challenge, achievement, milestone, persistent rare-drop collection, shop, entitlement, refund, migration, admin, devtools, seasonal, and future reward-track paths.

Direct inventory mutation should be avoided outside controlled initialization, repair, migration, or devtools paths.

Normal acquisition should resolve through `GrantAward`.

## Persistent Rare Drops

Rare drops are not owned until collected.

A rare drop that is merely rolled or spawned does not create inventory state.

Persistent rare-drop flow:

```text
rare persistent drop appears in runtime
-> player collects drop
-> trusted collection event
-> GrantAward
-> inventory grant application
-> owned inventory state
```

Runtime-only rare drops are excluded from inventory persistence.

Normal pickups, temporary weapon pickups, ammunition pickups, health pickups, shield pickups, buffs, and softpoint weapons do not enter inventory.

## Starter Inventory

Every non-corrupt profile-shaped identity needs a valid starter/default hangar.

Starter inventory should be sufficient to produce a playable fallback build.

Baseline starter inventory should include:

```text
one default OwnedShip for the baseline ship
default primary weapon access or ownership
no hardwired equipment
no optional modules unless intentionally granted
```

Exact starter inventory contents are a gametime balancing decision.

The architecture requirement is that missing, corrupt, incomplete, or unavailable inventory/API data must not make the game unplayable.

Fallback behavior should produce or synthesize a safe default hangar snapshot.

## Guest Inventory Behavior

Guest profiles behave like non-guest profiles for gameplay flow.

The difference is storage durability.

Guest inventory, hangar, progression, and other durable-shaped state live in transient memory until and unless they are saved into a new profile through an explicit supported flow.

Guest flow keeps the same logical inventory behavior through transient storage.

If guest state is saved into a new Local Profile, the supported profile-creation flow may copy transient guest inventory and other durable-shaped state into the new durable profile.

Guest should not require separate gameplay rules.

## BuildEligibility Handoff

Inventory/hangar provides a normalized snapshot.

BuildEligibility consumes that snapshot, applies mode and compatibility rules, and produces eligible options in its own seam.

Inventory/hangar should not decide final match eligibility or compile a resolved build.

## Loadout Relationship

LoadoutSelection selects from eligible owned options.

It should reference owned instances where owned instances exist.

Likely selection fields:

```text
selected_owned_ship_id
selected_owned_weapons_by_point
selected_owned_modules_by_slot
selected_starting_ammo choices if needed
```

LoadoutSelection does not contain the whole inventory.

LoadoutSelection does not install or remove hardwired equipment.

LoadoutSelection does not mutate ownership.

LoadoutSelection does not reserve equipment globally across saved loadouts.

## Hardwired Equipment And Loadouts

Hardwired equipment is not selected in normal loadout slots.

Hardwired equipment belongs to `OwnedShip`.

When a loadout selects an owned ship, that ship may carry hardwired equipment.

BuildEligibility may mask, disable, normalize, or block that hardwired equipment based on mode rules.

ResolvedPlayerBuild applies the allowed hardwired effects after the owned ship and selected loadout are validated.

## Player-Data Persistence Boundary

Player-data owns the storage route and identity mapping.

Inventory and hangar owns the logical inventory shape and grant application result.

This doc stays at the contract boundary and does not specify physical storage mechanics.

## Planned Player-Data Operations

```text
accept routed grants
load normalized inventory snapshots
persist owned state
persist stackable item counts
persist unlock/access state
apply guest transient storage when appropriate
support fallback/default hangar repair behavior
```

## Fallback And Corrupt Data Policy

The inventory/hangar path must tolerate missing, corrupt, incomplete, or unavailable inventory data.

Fallback should be safe and minimal.

It should not silently grant permanent durable ownership unless the selected player-data route explicitly performs an initialization write.

Safe fallback may be synthesized for runtime use when persistence is unavailable.

Durable repair/initialization behavior is a gametime persistence decision.

## Implementation Planning

Recommended implementation sequence:

```text
1. Define logical hangar/inventory contracts.
2. Add default-backed HangarInventory loading.
3. Add guest transient hangar state.
4. Add inventory grant application.
5. Add owned instance IDs for ships, weapons, and modules.
6. Add normalized inventory snapshot handoff.
7. Add hardwired equipment attachment to OwnedShip.
8. Add fallback/corrupt data tests.
```

Early slices should favor seams over complete mechanics.

The first useful slice can return a default hangar snapshot and feed BuildEligibility without implementing full acquisition.

## Testing Direction

Important future tests:

```text
starter inventory produces a valid fallback build
missing inventory data produces a safe fallback snapshot
corrupt inventory data does not make the game unplayable
guest inventory behaves like normal inventory but remains transient
owned ships, weapons, and modules are instanced
unlock grants do not create ownership
ownership grants create owned instances
purchase grants and reward grants use the same ownership application path
runtime pickups do not persist into HangarInventory
persistent rare drops grant ownership only after collection
BuildEligibility consumes a normalized inventory snapshot
hardwired equipment attaches to one OwnedShip
fallback works for unavailable API data
local and Rails stores satisfy the same logical inventory contract
```

## Related Docs

- [Progression And Rewards](progression-and-rewards.md)
- [Player Build And Loadouts](player-build-and-loadouts.md)
- [Player Data And Persistence](../platform/player-data-and-persistence.md)

## Open Gametime Decisions

- Exact logical schema file layout.
- Exact starter inventory contents.
- Exact hardwire install cost.
- Exact hardwire removal policy.
- Whether hardwiring later becomes a currency sink.
- Exact duplicate item grant behavior beyond structural support.
- Exact inventory repair/initialization write policy.

## Core Invariants

- Inventory and hangar own durable owned state.
- Runtime equipment owns live match state.
- Ships, weapons, and modules are instanced owned items.
- Normal owned equipment is reusable unless explicitly made hardwired.
- Hardwired equipment is attached to one OwnedShip.
- Hardwired equipment is not reusable general inventory.
- Unlocks do not equal ownership.
- Ownership comes from grants.
- Purchases produce grants.
- Non-purchase rewards can also produce ownership grants.
- Inventory exposes owned/access state.
- BuildEligibility consumes a normalized inventory snapshot.
- Runtime pickups and temporary overwrites are never saved to inventory.
- Persistent rare drops become ownership only after collection and grant application.
- Guest inventory behaves normally but stores durable-shaped state in transient memory.
- Fallback/default hangar state must prevent missing or corrupt data from making the game unplayable.
