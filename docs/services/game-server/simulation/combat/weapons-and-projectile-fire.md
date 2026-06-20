# Weapons And Projectile Fire

Parent index: [Game Server Simulation Combat](./!INDEX.md)

## Purpose

This document describes the game-server service boundary for authoritative weapon firing and projectile creation.

It covers the current runtime implementation in the Go game server. It does not define the full future loadout system, client weapon presentation, or cross-system weapon design.

## Overview

Weapon fire is server-authoritative.

The client sends input intent through the realtime input packet. The game server stores that input on the player ship, steps weapon cooldown state during simulation, decides whether the active player can shoot, evaluates the requested weapon slot through the `weapons` package, and creates a runtime projectile only when the weapon policy allows firing.

The current weapon implementation separates several concerns:

```text
PlayerArmory
= session/default equipment owned by the player session

ShipWeapons
= live weapon equipment copied onto the active runtime ship

WeaponState
= live cooldown and ammo state for primary and secondary slots

Profile
= immutable weapon tuning and projectile/damage/effect metadata

ProjectileSpawn
= pure spawn intent returned by weapon fire policy

runtime.Bullet
= stored projectile entity used by movement, collision, state projection, and impact effects
```

The normal player-fire path is:

```text
Client input packet
-> runtime.InputState.PrimaryFire / SecondaryFire
-> Game.Step
-> stepPlayerWeapons
-> stepPlayers
-> firePlayerPrimaryWeapon / firePlayerSecondaryWeapon
-> weapons.Fire
-> runtime.NewBulletFromWeaponSpawn
-> game.entities.Projectiles
-> state packet projection
-> collision and damage flow
```

`weapons.Fire` is the pure decision point. It does not mutate `Game`, runtime entity maps, packets, damage targets, scoring, pickups, radial effects, or client state. It returns an updated slot state and a projectile spawn intent. The `Game` package adapts that result into runtime projectile storage.

## Code root

```text
services/game-server/internal/game/weapons/
services/game-server/internal/game/player_weapons.go
services/game-server/internal/game/simulation_weapons.go
services/game-server/internal/game/simulation_players.go
services/game-server/internal/game/runtime/bullet.go
services/game-server/internal/game/runtime/state.go
services/game-server/internal/game/runtime/packets_generated.go
services/game-server/internal/game/combat.go
services/game-server/internal/game/combat_damage_requests.go
services/game-server/internal/game/radial_spawning.go
shared/constants/weapons.toml
shared/packets/gameplay.toml
```

## Responsibilities

The game-server weapon and projectile-fire boundary owns:

* Weapon IDs used by the current runtime.
* Primary and secondary weapon slots.
* Equipped weapon state shapes.
* Default player armory.
* Per-slot cooldown and ammo state.
* Weapon profile lookup.
* Weapon fire policy.
* Projectile spawn intent construction.
* Weapon-backed runtime projectile creation through the `Game` adapter.
* Projection of weapon and projectile metadata into gameplay state packets.
* Copying projectile damage and impact-effect metadata onto runtime projectile entities.
* Current pickup-driven torpedo equip behavior through the pickup effect adapter.

## Does not own

This boundary does not own:

* Client-side input bindings or presentation.
* WebSocket transport.
* Packet codec generation.
* Full future loadout validation.
* Durable inventory or hangar ownership.
* Collision detection.
* Damage math.
* Asteroid destruction consequences.
* Scoring.
* Pickup drop rules.
* Radial effect stepping.
* Devtools-only projectile spawning policy.
* Client projectile rendering.

Those responsibilities belong to client, protocol, data, player-data, simulation combat, simulation pickups, simulation world, devtools, or planning docs as appropriate.

## Current weapon model

Current weapon identities are:

```text
basic_cannon
torpedo
```

Current weapon slots are:

```text
primary
secondary
```

Current ammo policies are:

```text
infinite
limited
```

`DefaultPlayerArmory` equips `basic_cannon` in the primary slot with infinite ammo. The default secondary slot is empty.

The active runtime ship receives weapon equipment from the player session when a ship is created or respawned:

```text
playerSession.PlayerArmory
-> runtime.Ship.ShipWeapons
```

`WeaponState` lives on the runtime ship and stores mutable cooldown and ammo state for the active ship instance. It is separate from the session armory.

## Current weapons

### Basic cannon

`basic_cannon` is the default primary weapon.

It currently uses:

```text
slot: primary
projectile_type: bullet
ammo_policy: infinite by default
damage_type: kinetic
damage_cause: projectile
impact_effect: none
```

The basic cannon projectile speed, lifetime, cooldown, spawn offset, and damage are generated from `shared/constants/weapons.toml`.

### Torpedo

`torpedo` is the current secondary weapon profile.

It currently uses:

```text
slot: secondary
projectile_type: torpedo
ammo_policy: limited when granted by pickup
impact_damage: explosive projectile damage
impact_effect: radial
```

Current torpedo pickup collection equips `torpedo` into the secondary slot and adds one ammo. Re-collecting a torpedo pickup adds ammo to the secondary slot rather than resetting the ammo count.

The current torpedo impact damage constant is zero. Torpedo damage is therefore primarily delivered by the radial impact effect after projectile impact, not by direct impact damage.

## Fire policy

`weapons.Fire` receives:

```text
Equipped
SlotState
Position
Forward
Rotation
```

The fire decision order is:

1. Empty equipped weapon does not fire.
2. Unknown weapon profile does not fire.
3. Cooldown greater than zero does not fire.
4. Limited ammo with zero or negative ammo does not fire.
5. Valid fire returns a projectile spawn intent and updated slot state.

On successful fire:

* cooldown is set from the weapon profile
* limited ammo is decremented
* infinite ammo is not decremented
* projectile position is computed from firing position plus normalized forward vector times spawn offset
* projectile velocity is computed from normalized forward vector times projectile speed
* projectile lifetime, damage spec, weapon ID, projectile type, rotation, and impact effect are copied from profile/request data into the spawn intent

The fire function does not create a projectile entity. It returns `ProjectileSpawn` for the game-owned adapter to consume.

## Game integration

`Game.Step` calls `stepPlayerWeapons` before `stepPlayers` during normal non-match-over simulation.

`stepPlayerWeapons` decrements cooldowns for every active player ship and clamps cooldown to zero.

`stepPlayers` handles movement first, then checks fire input:

```text
PrimaryFire
-> firePlayerPrimaryWeapon

SecondaryFire
-> firePlayerSecondaryWeapon
```

Both paths call `weapons.Fire` with the matching equipped weapon and slot state.

When fire succeeds, the game server:

1. allocates a projectile ID from the spawner
2. calls `runtime.NewBulletFromWeaponSpawn`
3. stores the returned projectile in `game.entities.Projectiles`
4. writes the returned slot state back to the player ship
5. refreshes depleted limited-ammo equipment from the session armory when needed

Current shooting eligibility is checked before either primary or secondary fire reaches the weapon package. The current gate requires the player to exist, not be pending despawn, not be suspended, and have primary cooldown at zero. The secondary slot then still applies its own cooldown and ammo checks inside `weapons.Fire`.

## Projectile runtime model

Runtime projectiles are stored in:

```text
game.entities.Projectiles
```

The runtime projectile type is still named `Bullet`, but weapon-backed projectiles carry weapon metadata:

```text
WeaponID
ProjectileType
ImpactEffect
Damage
DamageSpec
```

`runtime.NewBulletFromWeaponSpawn` copies the weapon spawn intent into the runtime projectile entity.

Projectile state projected to clients includes:

```text
id
owner_id
x
y
rotation
weapon_id
projectile_type
```

Projectile damage spec and impact effect metadata are server-side runtime data. They are not projected as normal projectile state.

## Projectile movement and removal

Projectile movement is handled outside the weapons package.

`stepBullets` advances projectiles through the motion package when bullets are not frozen by world simulation options. It removes projectiles when they are ready for removal, expired, or far from all cameras.

Projectile lifetime is carried from the weapon profile into the runtime projectile. Expired projectiles are deleted from the projectile map.

## Collision, damage, and impact effects

Weapons do not resolve damage.

When a projectile collides with an asteroid, `combat.go` builds a projectile-to-asteroid damage request from the runtime projectile and target asteroid. `damage.ResolveSingle` performs damage resolution, and game-owned adapters apply the result back to the asteroid.

Projectile damage requests prefer `bullet.DamageSpec`. If the damage spec is empty, the adapter falls back to legacy kinetic projectile damage using the projectile’s integer `Damage` field.

Projectile impact effects are also consumed outside the weapons package. If a projectile has `ImpactEffectRadial`, game-owned impact handling spawns a radial effect at the impact position. The radial package owns timing, coverage, target filtering, and hit-intent generation after the effect is spawned.

The torpedo path currently combines these pieces:

```text
torpedo profile
-> projectile spawn intent with radial impact effect
-> runtime projectile
-> projectile/asteroid collision
-> direct impact damage request
-> radial effect spawn
-> radial stepping
-> radial damage requests
```

## Pickup integration

Weapon pickup collection is resolved through the pickup collection and pickup effect flow.

The pickup collection package resolves a `torpedo` pickup into an effect intent:

```text
effect_type: equip_weapon
weapon_id: torpedo
slot: secondary
ammo: 1
```

The game-owned pickup effect adapter applies that intent to the active runtime player ship. It equips the weapon in the requested slot with limited ammo and adds the pickup ammo to that slot’s runtime ammo count.

This is runtime equipment mutation. It is not durable inventory or future loadout ownership.

## Protocol and state surface

The client sends firing intent as part of the normal input packet:

```text
primary_fire
secondary_fire
```

The server does not accept client-created projectile IDs, damage values, cooldown values, ammo values, or impact-effect metadata.

The state packet sends weapon and projectile readback through generated packet fields:

```text
ShipState.primary_weapon_id
ShipState.primary_ammo_policy
ShipState.primary_cooldown_remaining
ShipState.primary_ammo_remaining
ShipState.secondary_weapon_id
ShipState.secondary_ammo_policy
ShipState.secondary_cooldown_remaining
ShipState.secondary_ammo_remaining

BulletState.weapon_id
BulletState.projectile_type
```

This surface lets the client present current equipment, cooldown/ammo state, and projectile visuals without owning authoritative firing results.

## Data ownership

Weapon tuning currently comes from:

```text
shared/constants/weapons.toml
```

Generated Go constants live in:

```text
services/game-server/internal/constants/weapons.go
```

The current constants include basic cannon projectile speed, lifetime, cooldown, spawn offset, and damage. They also include torpedo projectile tuning, torpedo cooldown, torpedo impact damage, torpedo radial damage, and torpedo radial timing/shape values.

Packet field shape comes from:

```text
shared/packets/gameplay.toml
```

Generated game-server runtime packet types live in:

```text
services/game-server/internal/game/runtime/packets_generated.go
```

Generated files should not be edited manually.

## Code map

Primary implementation files:

* `services/game-server/internal/game/weapons/types.go` - Weapon IDs, slots, ammo policies, equipped shapes, player armory, and default armory.
* `services/game-server/internal/game/weapons/state.go` - Per-slot cooldown and ammo state stepping.
* `services/game-server/internal/game/weapons/profiles.go` - Current weapon profiles and impact-effect metadata.
* `services/game-server/internal/game/weapons/fire.go` - Pure fire policy and projectile spawn intent construction.
* `services/game-server/internal/game/player_weapons.go` - Game adapter from weapon fire result to runtime projectile insertion.
* `services/game-server/internal/game/simulation_weapons.go` - Per-tick weapon cooldown stepping.
* `services/game-server/internal/game/simulation_players.go` - Runtime player input handling and fire invocation.
* `services/game-server/internal/game/runtime/bullet.go` - Runtime projectile construction and state projection.
* `services/game-server/internal/game/runtime/state.go` - Runtime ship/projectile fields.
* `services/game-server/internal/game/runtime/ship.go` - Ship state projection for weapon fields.
* `services/game-server/internal/game/player_session_state.go` - Session state projection for armory fields.
* `services/game-server/internal/game/combat.go` - Projectile collision consequence path.
* `services/game-server/internal/game/combat_damage_requests.go` - Projectile damage request adapter.
* `services/game-server/internal/game/radial_spawning.go` - Projectile impact effect adapter.
* `services/game-server/internal/game/pickups/collection.go` - Pickup collection effect intent for torpedo equip.
* `services/game-server/internal/game/pickup_effects.go` - Runtime application of weapon pickup effects.

Source-of-truth and generated files:

* `shared/constants/weapons.toml`
* `services/game-server/internal/constants/weapons.go`
* `shared/packets/gameplay.toml`
* `services/game-server/internal/game/runtime/packets_generated.go`

Important non-ownership boundaries:

* `services/game-server/internal/game/damage/` owns damage resolution, not weapons.
* `services/game-server/internal/game/effects/radial/` owns radial effect stepping, not weapons.
* `services/game-server/internal/game/physics/` owns collision detection, not weapons.
* `client/` owns input collection and presentation, not authoritative weapon results.
* `services/game-server/internal/game/spawning.go` still contains debug bullet spawning through `runtime.NewBullet`; normal player weapon fire uses `runtime.NewBulletFromWeaponSpawn`.

## Tests and verification

Relevant focused tests include:

* `services/game-server/internal/game/weapons/types_test.go`
* `services/game-server/internal/game/weapons/state_test.go`
* `services/game-server/internal/game/weapons/profiles_test.go`
* `services/game-server/internal/game/weapons/fire_test.go`
* `services/game-server/internal/game/player_weapons_test.go`
* `services/game-server/internal/game/pickup_effects_test.go`
* `services/game-server/internal/game/radial_projectile_impact_test.go`
* `services/game-server/internal/game/runtime/entity_health_test.go`

Broader verification should include the game-server Go test suite when weapon changes touch runtime integration, packet projection, pickup effects, collision, damage, or radial impact behavior.

## Related docs

* [Game Server Simulation Combat](./!INDEX.md)
* [Game Server Simulation](../!INDEX.md)
* [Game Server](../../!INDEX.md)
* [Services](../../../!INDEX.md)
* [Realtime Protocol](../../../../protocol/!INDEX.md)
* [Data](../../../../data/!INDEX.md)
* [Systems Design Combat](../../../../systems-design/combat/!INDEX.md)
* [Player Build Limits](../../../../limits/player-build-limits.md)
* [Player Build And Loadouts](../../../../planning/domains/gameplay/player-build-and-loadouts.md)

## Notes

This document describes the current service implementation. Future player-build and loadout work is expected to expand weapon points, ammo ownership, loadout validation, and runtime equipment state, but those planning facts are not current game-server behavior yet.
