# Inventory And Build Flow

Parent index: [Player Experience](./!INDEX.md)

## Purpose

This document describes the current cross-system flow from player identity to durable hangar inventory and an authoritative match-start build.

## Overview

```text
player identity
-> player-data store route
-> versioned HangarInventory
-> game-server inventory adapter
-> mode-specific build eligibility
-> player loadout selection or deterministic fallback
-> immutable ResolvedPlayerBuild
-> spawn and respawn runtime state
```

The durable inventory and match-local resolved build are different authorities. Inventory answers what the player owns. The build system answers what those owned instances may become in the selected match.

## Participating systems

- API server: persists authenticated-account inventory with optimistic versioning.
- Player-data: owns the hangar schema, starter fallback, grants, repair, and identity routing.
- Game-server inventory integration: loads generated player-data inventory results.
- Modes and match rules: supply build restrictions and the mode identity.
- Player-build service: computes eligibility, validates selection, and resolves the match-start build.
- Player session: owns the applied immutable build for the match.
- Lives and respawn: choose whether mutable equipment persists or resets to the resolved build.
- Client: opens the existing game-styled loadout readout in pregame and lobby surfaces, submits owned-instance selections, and presents resolved hull, shields, weapons, ammunition, cooldowns, and module identities in the HUD.

## Authority boundaries

### Inventory authority

Player-data owns owned ship, weapon, module, unlock, stackable, default-ship, acquisition, state, grant-idempotence, and inventory-version facts.

The game server does not create ownership because a player selects an item. It can only select an eligible owned instance from the loaded snapshot.

### Persistence authority

Guest inventory is process-local. Local-profile inventory is durable in embedded SQLite. Authenticated-account inventory is durable through the API server.

Every persisted write advances `inventory_version`. Concurrent stores use expected-version comparison.

### Build authority

The game server owns catalog interpretation, mode restrictions, eligibility, fallback, selection validation, and resolved runtime configuration.

A catalog ID such as `v_wing` or `pulse` is not an owned identity. Loadouts refer to stable owned-instance IDs and are validated against the exact inventory snapshot used to compute options.

### Runtime authority

`ResolvedPlayerBuild` is immutable match-start configuration. Current ammunition, cooldowns, health, shields, and pickup changes are runtime state.

Build changes are accepted only before simulation begins. Respawn applies the mode's loadout persistence policy and always uses the session-owned resolved build as the reset baseline.

## Flow summary

### 1. Resolve identity and load inventory

Player-data routes `guest`, `local_profile`, or `authenticated_account` identity to the correct store. Missing inventory creates a stable starter inventory. Invalid or corrupt inventory produces a playable fallback and an attempted repair.

### 2. Capture inventory provenance

The game-server adapter returns the complete load result. The build service retains the inventory snapshot, inventory version, fallback/repair facts, normalized build rules, and computed eligibility in one `LoadedBuildContext`.

### 3. Compute eligible options

The build catalog is filtered by ownership, item state, ship support, and mode restrictions. Blocked instances receive machine-readable reasons. Unsupported weapon points and module slots are not exposed merely because an owned item exists.

### 4. Select or fall back

A client selection must use eligible owned-instance IDs. The client requests authoritative options before room creation through a reusable preflight websocket connection, or reads the personalized option state attached to a lobby snapshot. The server supplies a deterministic fallback: eligible default ship, preferred primary weapon, then stable sorted alternatives. Invalid submissions preserve the last accepted build and return a machine-readable error.

### 5. Resolve the build

Resolution revalidates the snapshot and selection, compiles ship stats and equipment, maps current weapons into the runtime armory bridge, applies passive and hardwired policy, and creates starting shield/ammunition state.

### 6. Apply before the match

Room activation clones the accepted `ResolvedPlayerBuild`, substitutes the authoritative runtime player ID, installs the build on the player session, and only then creates the ship. This avoids spawning a default ship and mutating it after simulation begins. Changes are rejected once the room leaves the lobby state.

### 7. Spawn and respawn

Initial spawn uses resolved maximum shields and starting equipment. Respawn either restores mutable loadout state or resets from `ResolvedPlayerBuild`, according to the selected lives policy.

## Inputs and outputs

Inputs:

- trusted player identity
- play-mode and trace context
- versioned hangar inventory
- build catalog
- resolved match rules
- optional loadout selection

Outputs:

- inventory persistence/fallback/repair facts
- eligible ships, weapons, and modules
- blocked option reason codes
- deterministic fallback selection
- immutable resolved player build
- runtime ship stats, armory, shield maximum, and starting ammo

## Out of scope

This document does not define:

- richer loadout-editor content presentation beyond the current baseline selectors
- named saved-loadout persistence
- commerce or reward entitlement policy
- broader ship, weapon, and module content
- active-module input and cooldown mechanics
- inventory trading
- progression unlock design

## Related docs

- [Player Experience](./!INDEX.md)
- [Hangar Inventory](../../services/player-data/hangar-inventory.md)
- [Player Inventory Persistence](../../services/api-server/player-inventory-persistence.md)
- [Player Inventory Client](../../services/game-server/integrations/player-inventory-client.md)
- [Player Builds And Loadouts](../../services/game-server/simulation/players/player-builds-and-loadouts.md)
- [Lives, Participation, And Spawn](../../services/game-server/simulation/players/lives-participation-and-spawn.md)
- [Inventory And Hangar Planning](../../planning/domains/gameplay/inventory-and-hangar.md)
- [Player Build And Loadouts Planning](../../planning/domains/gameplay/player-build-and-loadouts.md)

## Notes

The authority chain is now implemented from identity and inventory loading through pregame/lobby selection, atomic match-start application, realtime projection, HUD presentation, and lobby return. The intentionally narrow V1 catalog remains the primary limitation: `v_wing`, the `pulse` catalog weapon mapped to `basic_cannon`, and no production module entries.
