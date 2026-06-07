# Weapon System

## Status Summary

- Weapon support is server-authoritative.
- `services/game-server/internal/game/weapons` owns weapon IDs, slots, equipped weapon state, armory defaults, weapon profiles, fire policy, and projectile spawn intent construction.
- Weapon fire is resolved through a pure `weapons.Fire` policy function that returns updated slot state and projectile spawn intent.
- Weapon profiles can carry impact effects, and `Torpedo` currently uses a radial impact effect.
- Root `internal/game` adapts weapon fire results into runtime projectile creation.

## Ownership

- `services/game-server/internal/game/weapons` owns weapon IDs, weapon slots, equipped weapon data shapes, armory defaults, weapon profiles, fire policy, and projectile spawn intent construction.
- Root `internal/game` adapts weapon fire results into runtime projectile creation.
- `services/game-server/internal/game/weapons` does not own runtime entity maps, inserting projectiles into `game.entities.Projectiles`, collision detection, damage resolution, scoring, radial ticking, packets, or client rendering.
- Client code should only own presentation, input collection, and visual feedback.
- Shared tuning and schema data remain the source of truth for weapon-related values.

## Data Sources

- `shared/constants/weapons.toml`
- `services/game-server/internal/constants/weapons.go`
- `client/scripts/generated/constants/constants.gd`
- `shared/packets/`
- `tools/data_sync/config.toml`
- `services/game-server/internal/game/weapons/types.go`
- `services/game-server/internal/game/weapons/profiles.go`
- `services/game-server/internal/game/weapons/fire.go`

## Weapon Identity And Slots

- `weapons.ID` is the stable identity type for weapons.
- Current weapon IDs are `BasicCannon` and `Torpedo`.
- `weapons.Slot` is the stable slot type for weapon ownership and equip placement.
- Current slots are `Primary` and `Secondary`.
- Primary and secondary are ownership slots, not rendering or UI concepts.
- Weapon identity should stay separate from visual presentation.

## Player Armory Model

- `AmmoPolicy` describes how a weapon is allowed to consume ammunition.
- Current ammo policies are `InfiniteAmmo` and `LimitedAmmo`.
- `Equipped` is the per-slot equipped weapon shape.
- `Equipped` carries weapon identity plus ammo policy.
- `PlayerArmory` is the player-owned armory shape.
- `ShipWeapons` is the live ship-facing weapon shape.
- `DefaultPlayerArmory` gives players `BasicCannon` in `Primary` and leaves `Secondary` empty unless assigned elsewhere.
- The default primary weapon uses `InfiniteAmmo`.
- Ammo policy exists even when the current default primary uses infinite ammo.
- Durable armory ownership should remain distinct from live runtime state.

## Weapon Profile Model

- `Profile` is the weapon profile shape.
- `Profile` binds weapon ID, slot, cooldown, projectile type/speed/lifetime/spawn offset, damage spec, and optional impact effect.
- `ProjectileProfile` carries projectile-specific tuning and spawn parameters.
- `ImpactEffectKind` classifies optional impact behavior.
- Current impact effect kinds are `ImpactEffectNone` and `ImpactEffectRadial`.
- `ImpactEffectSpec` carries impact-effect metadata for a projectile.
- `ImpactEffectSpec` is carried by projectiles, not executed immediately by the weapon profile itself.
- `Lookup` is the current profile registry.
- `Lookup` is code-defined and keyed by weapon ID.
- Missing or unknown profiles cause fire to fail cleanly.
- Profile data should stay separate from per-player possession.
- BasicCannon and Torpedo differ in slot, damage type, projectile type, and impact effect.
- BasicCannon currently uses kinetic projectile damage and no impact effect.
- Torpedo is the secondary weapon profile.
- Torpedo currently uses projectile type `torpedo`.
- Torpedo currently uses explosive direct impact damage.
- Torpedo currently carries radial impact effect metadata.

## Fire Flow

- Player/session weapon ownership is represented by equipped weapon state in the live player/session runtime flow.
- Input reaches the server-side firing path through `internal/game`.
- `FireRequest` carries the equipped weapon, current slot state, world position, forward vector, and rotation.
- `SlotState` carries per-slot firing state, including cooldown and ammo counts.
- `weapons.Fire` is the pure fire-policy function.
- `FireResult` reports whether firing succeeded, the updated slot state, and any projectile spawn intent.
- Fire decision order is strict:
  - empty equipped weapon does not fire
  - unknown profile does not fire
  - cooldown blocks fire
  - limited ammo with zero ammo blocks fire
  - valid fire returns a projectile spawn intent and updated slot state
- `weapons.Fire` is pure and does not mutate `Game`, sessions, runtime entities, packets, or client state.
- On a successful fire, cooldown is set from the weapon profile in the returned slot state.
- Limited ammo is decremented only after a successful fire in the returned slot state.
- The server remains the source of truth for weapon firing outcomes.
- Client handling should stay limited to intent collection and presentation.
- `BasicCannon` and `Torpedo` both follow the same fire path, but they produce different projectile metadata and impact behavior through their profiles.
- The seam is ready for additional weapon profiles that follow the same request/profile/spawn pattern.

## Projectile Spawn Model

- `ProjectileSpawn` is the projectile spawn intent/data shape returned by `weapons.Fire`.
- `ProjectileSpawn` fields are `WeaponID`, `ProjectileType`, `Position`, `Rotation`, `Velocity`, `Lifetime`, `Damage`, and `ImpactEffect`.
- Projectile position is computed from the firing position plus the normalized forward vector multiplied by the profile spawn offset.
- Projectile velocity is the normalized forward vector multiplied by the profile projectile speed.
- `ProjectileSpawn` is an intent, not a live entity.
- Root `internal/game` and spawning code own creating and storing the runtime projectile from that intent.
- The current runtime shape stores that spawn data on the bullet-like projectile entity for later collision handling.
- Projectile damage and impact effect are copied into runtime projectile state for later collision handling.
- Spawn ownership should stay on the authoritative server side.
- Client code should only observe the resulting state and effects.
- `runtime.NewBulletFromWeaponSpawn` copies the weapon ID, projectile type, impact effect, rotation, velocity, lifetime, and damage spec into runtime bullet state.
- `combat.go` later consumes projectile metadata from the runtime bullet when collision happens.
- Direct damage and impact effects happen later, not inside `weapons.Fire`.

## Damage Integration

- Weapon profiles carry `damage.DamageSpec` as damage intent.
- Weapon damage specs carry amount, type, and cause intent.
- Direct projectile damage is not resolved by the weapons package.
- Combat and game adapters convert runtime collision facts into `DamageResolutionRequest`.
- `combat_damage_requests.go` is the adapter path from runtime collision facts to the damage seam.
- `combat_damage_application.go` applies `DamageResult` back onto runtime entities after resolution.
- `damage.ResolveSingle` remains the authoritative damage resolver.
- Current basic cannon projectile damage is kinetic.
- Current torpedo direct impact damage is explosive.
- Weapon damage should route through the damage seam described in [docs/design/damage.md](damage.md).
- Damage shaping and resolution should stay outside weapon presentation code.

## Impact Effect Integration

- Impact effects are optional projectile metadata.
- Current impact effect kinds are `none` and `radial`.
- Torpedo uses `ImpactEffectRadial`.
- Collision and combat detect projectile impact, then root `internal/game` spawns the radial effect from the projectile metadata.
- The weapons package does not step radial effects.
- The radial package does not know weapons.
- Radial effect details live in [docs/design/radial-effects.md](radial-effects.md).
- Hit feedback should stay separate from damage resolution.

## Current Weapons

- Basic cannon is the current primary profile.
- Torpedo is the current secondary profile.
- Both profiles are code-defined in `services/game-server/internal/game/weapons/profiles.go` and use generated constants.
- The current roster is intentionally small and authoritative.
- Do not treat this outline as a complete roster.

## Constants And Tuning

- Weapon tuning lives in `shared/constants/weapons.toml`.
- Weapon constants are split by weapon ID under `constants.server.weapons.*`.
- Generated Go constants live in `services/game-server/internal/constants/weapons.go`.
- `basic_cannon` tuning currently feeds `BasicCannonProjectileSpeed`, `BasicCannonProjectileLifetime`, `BasicCannonCooldown`, `BasicCannonProjectileSpawnOffset`, and `BasicCannonDamage`.
- `torpedo` tuning currently feeds `TorpedoProjectileSpeed`, `TorpedoProjectileLifetime`, `TorpedoCooldown`, `TorpedoProjectileSpawnOffset`, `TorpedoImpactDamage`, `TorpedoRadialDamage`, `TorpedoRadialZoneCount`, `TorpedoRadialZoneWidth`, `TorpedoRadialZoneSpawnSeconds`, `TorpedoRadialTickSeconds`, `TorpedoRadialTotalSeconds`, and `TorpedoRadialZoneLifetimeSeconds`.
- Torpedo direct impact damage comes from the torpedo impact damage constant.
- Torpedo radial damage comes from the torpedo radial damage constant.
- Torpedo radial zone and timing values are generated tuning data, not architectural constants.
- Keep this section focused on ownership and tunability, not a full numeric dump.

## Client Presentation

- The client may render weapon visuals, fire feedback, equipped weapon display, and impact presentation.
- The client does not own authoritative weapon rules, cooldown, ammo, damage, or projectile creation.
- This doc does not claim completed UI or equip presentation unless that work is implemented elsewhere.

## Testing And Verification

- Focused tests live under `services/game-server/internal/game/weapons`.
- Broader integration uses `go test ./...` from `services/game-server`.
- Packet or client checks belong only where weapon state crosses those seams.

## Future Work

- Add additional weapon profiles if new weapons are implemented.
- Add client equip/presentation flows when those systems are implemented.
- Add focused tests for new fire or profile rules as the weapon seam evolves.
