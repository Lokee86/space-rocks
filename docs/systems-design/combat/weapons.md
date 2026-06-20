# Weapons

Parent index: [Combat](./!README.md)

## Purpose

This document defines the systems-design model for weapons in Space Rocks.

It documents the conceptual combat boundary for weapon identity, equipment slots, firing authority, projectile spawn intent, ammunition state, damage intent, and impact effects. It does not replace service implementation docs, packet docs, data pipeline docs, or future loadout planning.

## Overview

Weapons are server-authoritative combat capabilities.

A weapon is not just a visual effect or input binding. It is a combat profile that describes how a firing request becomes a projectile spawn intent, what damage intent that projectile carries, and what optional impact effect should be attached for later resolution.

The current weapon model separates these concepts:

```text
Weapon identity
= stable conceptual weapon id

Weapon slot
= where the weapon is equipped for runtime firing

Equipped weapon
= weapon id plus ammo policy in a slot

Weapon state
= mutable per-slot cooldown and ammo

Weapon profile
= immutable firing, projectile, damage, and impact-effect metadata

Projectile spawn intent
= pure result of successful fire policy

Runtime projectile
= authoritative server entity created from the spawn intent
```

The current implemented roster is intentionally small:

```text
basic_cannon
torpedo
```

The current implemented slots are:

```text
primary
secondary
```

The current ammo policies are:

```text
infinite
limited
```

The basic cannon is the default primary weapon. Torpedo is the current secondary weapon and is acquired through runtime pickup effects.

## Conceptual model

Weapon behavior is built around a profile-plus-state split.

The profile describes what a weapon is allowed to produce. Runtime state describes whether a specific equipped slot can produce it right now.

```text
Equipped weapon + slot state + ship pose
-> fire policy
-> updated slot state + projectile spawn intent
```

A successful fire does not directly resolve damage. It creates an authoritative projectile intent. That projectile later moves, collides, applies damage through the damage seam, and may trigger impact effects.

This keeps weapons focused on combat capability and firing rules, not every consequence that may follow from a projectile.

## Weapon identity

Weapon IDs identify conceptual weapons.

Current IDs:

```text
basic_cannon
torpedo
```

Weapon identity must stay separate from:

```text
client scene paths
projectile visuals
HUD labels
pickup presentation
durable inventory instance ids
future owned-weapon ids
```

A future owned weapon instance may reference a weapon profile id, but the owned instance is not the same concept as the weapon id. The weapon id names the combat capability. Inventory and hangar systems own durable ownership when that layer exists.

## Slots and equipment

Current runtime weapon slots are primary and secondary.

Primary and secondary are combat equipment slots, not UI layout concepts. The client may display those slots, but the slot authority belongs to the server-side weapon and runtime equipment model.

The current default player armory is:

```text
primary: basic_cannon, infinite ammo
secondary: empty
```

When a player ship is created or respawned, session armory is copied into active ship weapons. The active ship then owns mutable weapon state for its current runtime life.

This creates an important split:

```text
PlayerArmory
= session/default equipment

ShipWeapons
= active runtime equipment copied onto the ship

WeaponState
= active runtime cooldown and ammo state
```

Runtime pickup effects may mutate active ship weapons and weapon state. They do not create durable inventory ownership.

## Firing authority

The server owns weapon firing outcomes.

The client may collect player input and present weapon state, but it does not decide whether a weapon fired, what projectile was created, what damage it carries, or what cooldown/ammo values are authoritative.

The conceptual firing path is:

```text
client input intent
-> server input state
-> server simulation step
-> weapon fire policy
-> projectile spawn intent
-> runtime projectile
-> state projection
-> client presentation
```

The fire decision is deterministic from authoritative server state.

A weapon should not fire when:

```text
no weapon is equipped
the weapon profile is unknown
the slot is still cooling down
the weapon uses limited ammo and ammo is empty
the player is not allowed to shoot by runtime state
```

On successful fire:

```text
cooldown is set from the weapon profile
limited ammo is decremented
infinite ammo is not decremented
a projectile spawn intent is returned
```

## Projectile spawn intent

Projectile spawn intent is the boundary between weapon policy and runtime entities.

A weapon profile provides projectile data such as:

```text
projectile type
projectile speed
projectile lifetime
spawn offset
damage intent
impact effect metadata
```

The fire policy combines that profile data with the firing ship position, forward vector, and rotation to produce a projectile spawn intent.

The spawn intent is not a live entity. Runtime simulation owns converting it into a projectile stored in the authoritative entity map.

This separation keeps weapon policy pure and makes projectile creation an explicit server-owned adaptation step.

## Damage intent

Weapons carry damage intent, not damage resolution authority.

A weapon profile may include a damage spec. That damage spec describes intended damage amount, type, and cause. It does not mutate targets by itself.

Projectile collision later routes through the damage system:

```text
weapon profile damage spec
-> projectile spawn intent
-> runtime projectile
-> collision consequence path
-> damage resolution
-> runtime target mutation
```

This keeps damage math independent from weapon firing. Weapons can define that a projectile should carry kinetic, explosive, area, or other damage intent, but the damage system owns how that intent affects a target.

## Impact effects

Impact effects are optional metadata carried by weapon-backed projectiles.

Current impact effect kinds are:

```text
none
radial
```

The basic cannon carries no impact effect.

Torpedo carries a radial impact effect. When the torpedo projectile impacts, game-owned collision/impact handling spawns a radial effect from projectile metadata. The radial effect system then owns timing, zone coverage, target filtering, hit intent generation, and expiration.

The weapon does not step the radial effect. The radial effect does not know which weapon created it except through the source data supplied at spawn time.

The conceptual boundary is:

```text
weapon profile
-> impact-effect metadata

projectile impact handling
-> creates effect instance

radial effect system
-> emits hit intents

damage system
-> resolves damage
```

## Current weapons

### Basic cannon

The basic cannon is the current default primary weapon.

Conceptually, it is the baseline direct-fire weapon:

```text
slot: primary
projectile: bullet
default ammo policy: infinite
damage type: kinetic
damage cause: projectile
impact effect: none
```

It exists as the always-available weapon capability for the current default ship flow.

### Torpedo

Torpedo is the current secondary weapon.

Conceptually, it is an impact-triggered area weapon:

```text
slot: secondary
projectile: torpedo
pickup ammo policy: limited
direct impact damage type: explosive
impact effect: radial
```

The current torpedo pickup equips torpedo into the secondary slot and adds one ammo. Repeated torpedo pickups add ammo rather than replacing the current ammo count.

Torpedo direct impact damage is currently not the main damage source. Its combat value comes from the radial effect spawned after projectile impact.

## Pickup interaction

Weapon pickups mutate runtime equipment state.

The current weapon pickup behavior is:

```text
torpedo pickup
-> equip torpedo in secondary slot
-> use limited ammo policy
-> add 1 ammo to secondary slot
```

Pickup-driven equipment is runtime state. It is not durable inventory, saved loadout state, or permanent profile ownership.

The design rule is:

```text
runtime pickups may change current combat capability
durable inventory owns permanent equipment ownership later
loadout planning owns pre-match selection later
```

## Client presentation

The client may present weapons, cooldowns, ammo, projectile visuals, pickup visuals, audio, and feedback.

The client must not own:

```text
weapon fire validation
cooldown authority
ammo authority
projectile creation
damage values
impact effect creation
pickup effect authority
```

Client-side weapon UI is a read model of server-owned state.

## Authority rules

The weapon authority rules are:

```text
The server owns firing outcomes.

The weapon profile owns firing, projectile, damage-intent, ammo-policy compatibility, and impact-effect metadata.

Runtime weapon state owns current cooldown and ammo.

The game aggregate owns adapting successful fire into runtime projectile creation.

Collision owns detecting projectile impact.

Damage owns resolving damage.

Radial effects own radial timing, coverage, filtering, hit intents, and expiration.

Pickups own collection/effect intent; the game aggregate owns applying runtime pickup effects.

The client owns input collection and presentation only.
```

No client message may be treated as authority for projectile IDs, projectile damage, impact effects, cooldown completion, or ammo counts.

## Invariants

Weapon identity stays stable and separate from visual presentation.

Weapon profiles stay separate from runtime possession and durable ownership.

Weapon fire produces intent; it does not directly mutate targets.

Runtime projectile creation happens on the authoritative server.

Damage resolution stays outside the weapon package and outside client presentation.

Impact effects are metadata until an authoritative impact path spawns an effect.

Radial effects remain their own system. Weapons may request radial behavior through impact metadata, but weapons do not step radial zones or select radial hits.

Pickup-granted weapons mutate runtime equipment state, not saved loadout state.

Infinite ammo and limited ammo are policies. Current infinite primary ammo should not be generalized into a permanent rule for all equipped weapons.

## Participating systems

Weapon behavior participates in these systems:

```text
Game server simulation
= authoritative firing, runtime equipment, projectile creation, collision consequences, pickup application

Realtime protocol
= client input intent and server state projection

Client
= input collection, HUD display, projectile/pickup presentation, audio/feedback

Data pipeline
= generated constants and packet fields

Damage
= authoritative damage resolution

Radial effects
= timed area-effect behavior after projectile impact

Pickups
= runtime equipment mutation through pickup effects

Future player build/loadouts
= planned pre-match equipment selection, weapon points, classification, starting ammo, and runtime equipment state
```

## Service implementation

The current authoritative implementation lives in the game-server service.

The relevant service implementation document is [Weapons And Projectile Fire](../../services/game-server/simulation/combat/weapons-and-projectile-fire.md).

This systems-design document owns conceptual rules and invariants. The service document owns code paths, runtime packet fields, tests, generated files, and implementation details.

## Data and tuning

Weapon tuning currently comes from generated constants sourced from `shared/constants/weapons.toml`.

The data layer owns source-of-truth files, generated outputs, and validation workflow. The weapon design only depends on the rule that tuning data feeds weapon profiles and should not be duplicated into presentation-only code.

Generated constants and packet fields should be changed through the data pipeline, not by editing generated service/client outputs manually.

## Planning boundary

Future player-build work is expected to expand the model from the current primary/secondary equipment shape into weapon points, loadout selection, starting ammo, weapon classification, runtime equipment state, inventory-backed ownership, and eligibility filtering.

Those planned concepts should not be documented as current runtime behavior until implemented.

The current durable design direction is still compatible with that future model:

```text
weapon profile remains the combat capability
owned weapon remains durable inventory state
loadout selection remains pre-match choice
runtime equipment state remains current ammo/cooldown/pickup state
```

## Related docs

* [Combat](./!README.md)
* [Systems Design](../!README.md)
* [Weapons And Projectile Fire](../../services/game-server/simulation/combat/weapons-and-projectile-fire.md)
* [Damage Resolution](../../services/game-server/simulation/combat/damage-resolution.md)
* [Collision To Damage Flow](../../services/game-server/simulation/combat/collision-to-damage-flow.md)
* [Radial Effects](../../services/game-server/simulation/combat/radial-effects.md)
* [Pickup Effects](../../services/game-server/simulation/pickups/pickup-effects.md)
* [HUD And Gameplay UI](../../services/client/hud-and-gameplay-ui.md)
* [Realtime Protocol](../../protocol/!README.md)
* [Constants](../../data/constants.md)
* [Data Sync And SSoT Pipeline](../../data/data-sync-and-ssot-pipeline.md)
* [Player Build And Loadouts](../../planning/domains/gameplay/player-build-and-loadouts.md)
* [Player Build Limits](../../limits/player-build-limits.md)

## Notes

The current runtime projectile type is still named `Bullet`, but weapon-backed projectiles carry weapon identity and projectile type metadata. Conceptually, “projectile” is the broader combat entity; `Bullet` is the current runtime implementation name.

The current weapon roster is not intended to be a complete final roster. It is the implemented baseline for the weapon seam.
