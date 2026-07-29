# Player Builds And Loadouts

Parent index: [Game Server Simulation Players](./!INDEX.md)

## Purpose

This document describes the authoritative game-server boundary that converts a routed hangar inventory snapshot and resolved mode rules into selectable options and one immutable match-start player build.

## Overview

```text
player-data HangarInventory
+ playerbuild Catalog
+ rules derived from ResolvedMatchRules
-> ComputeEligibility
-> EligibleBuildOptions
-> LoadoutSelection
-> Resolve
-> ResolvedPlayerBuild
-> playerSession
-> runtime ship and RuntimeEquipmentState
```

The build system validates ownership and mode eligibility before compiling runtime stats, weapons, modules, effects, shields, and starting ammunition. The resolved build is cloned at storage and read boundaries and cannot be changed after match simulation begins.

## Code root

```text
services/game-server/internal/game/playerbuild/
services/game-server/internal/game/player_builds.go
```

Session integration lives in `services/game-server/internal/game/session.go`.

## Responsibilities

This boundary owns:

- ship-variant, weapon-point, module-slot, module, and hardwired vocabulary
- catalog validation
- mode-derived build restrictions
- owned-instance eligibility filtering
- stable machine-readable blocked reasons
- deterministic fallback selection
- server-side selection validation
- duplicate-owned-instance prevention within one loadout
- match-start ship-stat compilation
- weapon and module resolution
- passive and active module declarations
- hardwired allowed, disabled, and normalized behavior
- maximum shield and starting-ammunition policy
- immutable resolved-build cloning
- pre-match application to player sessions and provisional ships
- respawn restoration from the resolved build

## Does not own

This boundary does not own:

- inventory durability or grants
- item acquisition
- current runtime ammo or cooldown mutation
- pickups
- client loadout UI
- named saved-loadout persistence
- room membership
- mode selection

It consumes a trusted inventory snapshot and resolved rules from those owners.

## Domain roles

The current production catalog is intentionally compact but exercises every implemented loadout category:

```text
v_wing
- weight: standard baseline

v_wing_scout
- weight: light
- 25% lower maximum health
- higher rotation, thrust, maximum speed, and damping
- currently reuses the V-Wing collision shape and client visual

both ships
- primary_1 and secondary_1: hardpoints
- primary_2 and secondary_2: unavailable
- shield, armor, engine, and utility module slots

pulse
- runtime mapping: basic_cannon
- primary, ballistic, skill_shot, direct, infinite ammunition

torpedo
- runtime mapping: torpedo
- secondary, missile, skill_shot, direct + area, limited ammunition
- default starting ammunition: 3

shield_capacitor: shield_mod, +50 maximum shields, -5% maximum speed
reinforced_hull: armor_mod, +50 maximum health, -10% thrust
engine_overdrive: engine_mod, +15% thrust, +10% maximum speed, -15 maximum health
flight_stabilizer: utility_mod, +15% rotation and damping, -5% maximum speed
```

Secondary weapons and modules are optional in `LoadoutSelection`; `primary_1` remains required. The client loadout surface provides explicit `NONE` choices for optional equipment.

Eligibility can allow or ban ship IDs, weight classes, weapon IDs, runtime slots, sizes, delivery classes, targeting policies, required effect flags, module IDs/classes/slots, active modules, and hardwired behavior. Options are keyed by owned inventory instance, not only catalog ID.

Fallback prefers the inventory's eligible default ship and the selected variant's default primary weapon. Otherwise it chooses the first deterministic eligible instance.

Selections require `primary_1`, validate every selected point and module slot against the chosen ship, and reject one owned weapon or module being equipped twice in the same build.

## Protocols and APIs

Primary APIs include:

```go
func ComputeEligibility(playerID string, inventory protocol.HangarInventory, catalog Catalog, rules Rules) EligibleBuildOptions
func ValidateSelection(selection LoadoutSelection, inventory protocol.HangarInventory, catalog Catalog, rules Rules, options EligibleBuildOptions) error
func Resolve(selection LoadoutSelection, inventory protocol.HangarInventory, catalog Catalog, rules Rules) (ResolvedPlayerBuild, error)
func RulesForMatch(matchRules modes.ResolvedMatchRules) Rules

func (service *Service) LoadOptions(identity protocol.PlayerDataIdentity, context protocol.PlayerDataRequestContext, matchRules modes.ResolvedMatchRules) (LoadedBuildContext, error)
func (service *Service) ResolveSelection(context LoadedBuildContext, selection LoadoutSelection) (ResolvedPlayerBuild, error)

func (game *Game) ApplyPlayerBuild(playerID string, build ResolvedPlayerBuild) error
func (game *Game) PlayerResolvedBuild(playerID string) (ResolvedPlayerBuild, bool)
```

`ApplyPlayerBuild` is rejected after match time begins or after final match state is locked.

## Data ownership

`LoadedBuildContext` retains the exact inventory load result, normalized build rules, and eligibility options used for later resolution.

`ResolvedPlayerBuild` owns immutable match-start configuration and provenance:

```text
player and mode ID
inventory version
selected owned ship and ship catalog ID
weight class and resolved ship stats
weapon-point layout and equipped weapons
equipped modules and hardwired declarations
passive effects and active behavior declarations
shield policy and starting ammo
runtime PlayerArmory bridge
```

The player session owns the applied resolved build. Mutable runtime ammo, cooldowns, health, shields, and pickup changes live on runtime state. Respawn either restores mutable state or resets to the resolved build according to the lives policy.

## Code map

```text
services/game-server/internal/game/playerbuild/catalog.go
services/game-server/internal/game/playerbuild/rules.go
services/game-server/internal/game/playerbuild/eligibility.go
services/game-server/internal/game/playerbuild/selection.go
services/game-server/internal/game/playerbuild/resolve.go
services/game-server/internal/game/playerbuild/service.go
services/game-server/internal/game/player_builds.go
services/game-server/internal/game/session.go
services/game-server/internal/playerinventory/runtime_client.go
services/player-data/protocol/packets.go
```

## Tests

Tests cover catalog validation, build-rule normalization, every major eligibility restriction, blocked reasons, fallback selection, invalid points and slots, duplicate owned instances, passive/active/hardwired behavior, finite ammunition, clone isolation, inventory-loader handoff, session spawn and respawn, provisional ship replacement, and late-change rejection.

```text
services/game-server/internal/game/playerbuild/*_test.go
services/game-server/internal/game/player_build_integration_test.go
services/game-server/internal/game/player_builds_test.go
```

## Related docs

- [Game Server Simulation Players](./!INDEX.md)
- [Lives, Participation, And Spawn](lives-participation-and-spawn.md)
- [Player Inventory Client](../../integrations/player-inventory-client.md)
- [Hangar Inventory](../../../player-data/hangar-inventory.md)
- [Inventory And Build Flow](../../../../domains/player-experience/inventory-and-build-flow.md)
- [Player Build And Loadouts Planning](../../../../planning/domains/gameplay/player-build-and-loadouts.md)
- [Player Build Limits](../../../../limits/player-build-limits.md)

## Notes

The authoritative owner system, player-facing selection, and first selectable ship/weapon/module catalog are implemented. Remaining work is saved-loadout persistence, broader catalog content, distinct ship presentation, richer equipment explanations, and additional runtime presentation.
