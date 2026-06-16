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

It does not own match runtime equipment state, mode filtering, final build eligibility, progression reward policy, shop pricing, or UI layout.

## Ownership Boundary

This doc owns planning for:

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

BuildEligibility owns filtering and selectable-option construction.

Inventory and hangar owns durable owned state and the normalized inventory snapshot consumed by BuildEligibility.

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

Core build-selection flow:

```text
ModeRules
+ HangarInventorySnapshot
+ ShipVariant catalog
+ WeaponProfile catalog
+ ModuleProfile catalog
-> BuildEligibility
-> EligibleBuildOptions
-> LoadoutSelection
-> ResolvedPlayerBuild
-> RuntimeShip / RuntimeEquipmentState
```

Inventory is not the final eligibility system.

Inventory records ownership, access, instance state, and broad availability facts.

BuildEligibility masks that inventory state with selected mode rules, build filters, slot rules, ship rules, weapon rules, module rules, and hardwired-equipment policy.

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

Flow:

```text
OwnedShip.hardwired_equipment
-> BuildEligibility reads allowed hardwired state
-> ResolvedPlayerBuild applies allowed hardwired effects
-> RuntimeShip / RuntimeEquipmentState receive resolved effects
```

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
Grant
- grant_kind: unlock
- target_ref: weapon.railgun

Effect:
- add weapon.railgun to unlocked_content if missing
```

```text
Grant
- grant_kind: inventory_item
- target_ref: ship.scout
- amount: 1

Effect:
- create one OwnedShip instance with ship_id = ship.scout
```

```text
Grant
- grant_kind: inventory_item
- target_ref: weapon.railgun
- amount: 1

Effect:
- create one OwnedWeapon instance with weapon_id = weapon.railgun
```

```text
Grant
- grant_kind: inventory_item
- target_ref: module.overcharger
- amount: 1

Effect:
- create one OwnedModule instance with module_id = module.overcharger
```

```text
Grant
- grant_kind: ship_part
- target_ref: item.coolant_fragment
- amount: 3

Effect:
- increment StackableInventoryItem quantity for item.coolant_fragment
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

Inventory can be changed by grants from:

```text
match rewards
objective rewards
mission rewards
challenge rewards
achievement rewards
milestone rewards
persistent rare drops
shop purchases
entitlements
refunds or reversals
account migration
admin grants
devtools test grants
seasonal or event rewards
future reward tracks
```

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

Guest flow:

```text
Guest identity
-> transient inventory/hangar route
-> normal HangarInventory behavior
-> normal BuildEligibility handoff
-> normal LoadoutSelection / ResolvedPlayerBuild flow
```

If guest state is saved into a new Local Profile, the supported profile-creation flow may copy transient guest inventory and other durable-shaped state into the new durable profile.

Guest should not require separate gameplay rules.

## BuildEligibility Handoff

Inventory/hangar provides a normalized snapshot.

BuildEligibility owns filtering, masking, and selectable-option construction.

Inventory/hangar should not decide final match eligibility.

BuildEligibility consumes:

```text
mode rules
hangar inventory snapshot
ship variant catalog
weapon profile catalog
module profile catalog
progression/access snapshot if not already normalized into inventory
```

BuildEligibility produces:

```text
eligible ships
eligible weapons by weapon point
eligible modules by module slot
blocked reasons
fallback loadout
```

Inventory/hangar owns:

```text
owned ship instances
owned weapon instances
owned module instances
hardwired equipment attached to ships
unlock/access state relevant to inventory
stackable inventory items
acquisition metadata
```

BuildEligibility owns:

```text
mode filtering
slot compatibility
weapon point compatibility
module slot compatibility
ship weight/class restrictions
weapon size restrictions
module restrictions
hardwired equipment allow/disable/normalize policy
blocked reason generation
selectable option generation
server-side validation of submitted selections
```

ResolvedPlayerBuild owns:

```text
match-start compiled stats
selected ship setup
selected equipment setup
applied passive effects
allowed hardwired effects
starting equipment state
runtime setup declarations
```

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

Player-data owns identity and route selection.

Inventory/hangar should be represented as logical player-data concepts that can be implemented by:

```text
guest transient memory
local SQLite
Rails/Postgres
```

Local Profile and Authenticated Account should share the same logical inventory/hangar contract even if their physical storage differs.

Future logical schema sources may include:

```text
shared/player_data/hangar_inventory.toml
```

or split files such as:

```text
shared/player_data/inventory.toml
shared/player_data/hangar.toml
```

Exact source file layout is a gametime implementation decision.

Logical contract should cover:

```text
HangarInventory
OwnedShip
OwnedWeapon
OwnedModule
HardwiredEquipment
StackableInventoryItem
UnlockedContentRef
```

Physical tables are not decided in this doc.

## Planned Player-Data Operations

Likely future player-data operations:

```text
LoadHangarInventory(identity)
ApplyGrantAward(identity, award)
UpdateOwnedShipHardwiredEquipment(identity, owned_ship_id, changes)
SaveDefaultOwnedShip(identity, owned_ship_id)
SaveLoadoutReference(identity, loadout_ref)
```

Initial implementation may start with default-backed reads and grow durable grant application later.

The important seam is that gameplay and client code should not choose the persistence backend directly.

## Fallback And Corrupt Data Policy

The inventory/hangar path must tolerate missing, corrupt, incomplete, or unavailable inventory data.

Fallback rule:

```text
Inventory service must be able to return a valid starter/default hangar snapshot.
BuildEligibility must be able to produce a valid fallback build from that snapshot.
```

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
4. Add local SQLite and Rails/Postgres logical parity.
5. Add inventory grant application.
6. Add owned instance IDs for ships, weapons, and modules.
7. Add BuildEligibility input snapshot.
8. Add loadout references to owned instances.
9. Add hardwired equipment attachment to OwnedShip.
10. Add mode-rule hardwired masking in BuildEligibility.
11. Add fallback/corrupt data tests.
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
profile creation can save guest transient state when supported
owned ships are instanced
owned weapons are instanced
owned modules are instanced
duplicate catalog items can exist as separate owned instances
unlock grants do not create ownership
ownership grants create owned instances
purchase grants and reward grants use the same ownership application path
runtime pickups do not persist into HangarInventory
persistent rare drops grant ownership only after collection
BuildEligibility filters unowned equipment out of selectable options
BuildEligibility masks inventory availability through mode rules
hardwired equipment attaches to one OwnedShip
hardwired equipment is not reusable across ships
ResolvedPlayerBuild applies allowed hardwired effects
mode rules can disable or normalize hardwired effects
fallback works for unavailable API data
local and Rails stores satisfy the same logical inventory contract
```

## Current Inputs

```text
HangarInventory
OwnedShip
OwnedWeapon
OwnedModule
hardwired equipment state
stackable inventory state
unlock/access inputs
acquisition inputs
GrantAward / Grant inputs
rare persistent drop inputs
shop purchase grant inputs
entitlement inputs
guest transient state
BuildEligibility input needs
fallback/default hangar needs
```

## Planned Outputs

```text
clear inventory ownership model
instanced owned ship/weapon/module planning
hardwired equipment ownership boundary
unlock versus ownership rules
grant application flow
guest transient inventory behavior
starter/default hangar policy
fallback behavior for missing/corrupt/API-unavailable data
BuildEligibility handoff rules
player-data persistence boundary
implementation sequence
testing direction
```

## Related Docs

* [Systems Plan Index](../systems-plan-index.md)
* [Player Experience Systems](player-experience-systems.md)
* [Player Build And Loadouts](player-build-and-loadouts.md)
* [Progression And Rewards](progression-and-rewards.md)
* [Shop, Commerce, And Economy](shop-commerce-and-economy.md)
* [Player Data And Persistence](../platform/player-data-and-persistence.md)
* [Anti-Cheat And Trust Policy](../platform/anti-cheat-and-trust-policy.md)
* [API Product Surface](../platform/api-product-surface.md)
* [Source Of Truth Map](../../design/source-of-truth-map.md)
* [Player-Data Schema Source Of Truth](../../design/player-data-schema-ssot.md)

## Open Gametime Decisions

```text
exact physical table layout
exact logical schema file layout
exact starter inventory contents
exact starter ship and weapon IDs
exact hardwire install cost
exact hardwire removal policy
whether hardwiring later becomes a currency sink
whether normal equipment install rules later become stricter
whether normal equipment can later be consumed, bound, or reserved
exact duplicate item grant behavior beyond structural support
exact inventory state vocabulary
exact reversal/refund storage behavior
exact hangar API endpoints
exact hangar UI/menu flow
exact inventory repair/initialization write policy
```

## Core Invariants

```text
Inventory/hangar owns durable owned state.
Runtime equipment owns live match state.
Ships, weapons, and modules are instanced owned items.
Normal owned equipment is reusable unless explicitly made hardwired.
Hardwired equipment is attached to one OwnedShip.
Hardwired equipment is not reusable general inventory.
Hardwiring mutates OwnedShip pre-game stats or effects.
Unlocks do not equal ownership.
Unlocks allow access, purchase, eligibility, or acquisition paths.
Ownership comes from grants.
Purchases produce grants.
Non-purchase rewards can also produce ownership grants.
New ownership grants target catalog refs.
Mutation/reversal grants may target owned instance refs.
Inventory exposes owned/access state.
BuildEligibility filters and masks inventory state into selectable options.
LoadoutSelection selects from eligible owned options.
LoadoutSelection does not mutate inventory.
ResolvedPlayerBuild compiles a legal selected build.
Runtime pickups and temporary overwrites are never saved to inventory.
Persistent rare drops become ownership only after collection and grant application.
Guest inventory behaves normally but stores durable-shaped state in transient memory.
Fallback/default hangar state must prevent missing or corrupt data from making the game unplayable.
```
