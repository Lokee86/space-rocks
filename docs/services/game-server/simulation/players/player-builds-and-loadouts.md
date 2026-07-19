# Player Builds And Loadouts

Parent index: [Game Server Simulation Players](./!INDEX.md)

## Purpose

This document describes the authoritative game-server boundary that turns a routed player inventory snapshot into selectable build options and one immutable match-start player build.

## Flow

```text
player-data HangarInventory
+ player-build Catalog
+ resolved build Rules
-> ComputeEligibility
-> EligibleBuildOptions
-> LoadoutSelection
-> Resolve
-> ResolvedPlayerBuild
-> playerSession
-> Runtime Ship + RuntimeEquipmentState
```

## Ownership

`internal/game/playerbuild` owns:

```text
ship variant definitions
weapon-point and module-slot vocabulary
build restriction rules
owned-option eligibility filtering
machine-readable blocked reasons
fallback selection
server-side loadout validation
match-start build compilation
hardwired effect policy
starting ammunition
shield build policy
immutable build cloning
```

It does not own inventory durability, grant policy, match runtime pickups, current ammo, current cooldowns, UI layout, or room membership.

## Current Catalog Baseline

```text
v_wing
- weight class: standard
- primary_1: hardpoint
- secondary_1: hardpoint
- primary_2 / secondary_2: unavailable
- module slots: shield, armor, engine, utility

pulse
- runtime mapping: basic_cannon
- slot: primary
- size: standard
- delivery: ballistic
- targeting: skill_shot
- effects: direct
- ammunition: infinite
```

The module contract is implemented, but the default production catalog does not yet contain module content.

## Inventory Handoff

`playerbuild.Service` accepts an inventory loader interface implemented by the existing game-server player-inventory runtime client. The resulting `LoadedBuildContext` keeps together the exact inventory snapshot, normalized rules, and authoritative eligible options used for resolution.

Owned ship, weapon, and module instance IDs are validated against that snapshot. Catalog IDs and owned IDs remain separate.

## Eligibility

Eligibility rejects unavailable or unknown inventory records and applies allow/ban filters for:

```text
ship IDs
ship weight classes
weapon IDs and slots
weapon sizes
delivery classes
targeting policies
required effect flags
module IDs, slots, and classes
active-module policy
```

Blocked options include stable reason codes suitable for future client presentation.

## Resolution

Resolution validates the selection again and compiles:

```text
selected owned ship
resolved ship stats
weapon-point layout
runtime Primary / Secondary armory bridge
selected module effects
hardwired declarations and effects
maximum shield policy
starting ammunition
active module behavior declarations
inventory version provenance
```

`ResolvedPlayerBuild` is cloned at storage and read boundaries so callers cannot mutate the session-owned match-start configuration.

## Runtime Integration

A newly created player session receives the current fallback `v_wing`/`pulse` build. Before the first simulation step, `Game.ApplyPlayerBuild` may replace that provisional build with a validated player selection.

After match time begins, build changes are rejected.

Initial spawn and respawn use the session's `ResolvedPlayerBuild`. Mutable runtime ammo, cooldowns, pickup overwrites, health, and shields remain on runtime state rather than the loadout object.

## Code Map

```text
services/game-server/internal/game/playerbuild/
services/game-server/internal/game/player_builds.go
services/game-server/internal/game/session.go
services/game-server/internal/playerinventory/runtime_client.go
services/player-data/protocol/packets.go
```

## Tests

```text
services/game-server/internal/game/playerbuild/*_test.go
services/game-server/internal/game/player_build_integration_test.go
services/game-server/internal/game/player_builds_test.go
```

## Related Docs

- [Player Build And Loadouts](../../../../planning/domains/gameplay/player-build-and-loadouts.md)
- [Inventory And Hangar](../../../../planning/domains/gameplay/inventory-and-hangar.md)
- [Player Session State](player-session-state.md)
- [Player Respawn](player-respawn.md)
