# Hangar Inventory

Parent index: [Player Data](./!INDEX.md)

## Purpose

This document describes the player-data service's authoritative hangar inventory contract, starter fallback, grant processing, optimistic versioning, repair behavior, and identity-based store routing.

## Overview

```text
PlayerDataIdentity
-> StoreRouter selects guest, local, or account store
-> InventoryManager.Load or ApplyGrant
-> normalize and validate HangarInventory
-> optimistic versioned store
-> normalized player-data result
```

Every valid identity receives a playable inventory. Missing or unusable inventory data falls back to a stable starter `v_wing` and `pulse` inventory. The manager attempts to persist or repair that fallback without making profile loading unusable when the repair write fails.

## Code root

```text
services/player-data/playerdata/
services/player-data/protocol/
```

## Responsibilities

This boundary owns:

- canonical hangar inventory validation and normalization
- stable owned-instance IDs for starter and granted items
- starter inventory synthesis
- inventory schema and inventory version handling
- guest, local-profile, and authenticated-account routing
- optimistic compare-and-store semantics
- corrupt or invalid inventory fallback and repair attempts
- idempotent inventory grants by `grant_id`
- unlock, entitlement, owned-item, rare-drop, ship-part, stackable, and reversal grant kinds
- owned ship, weapon, module, hardwired, unlock, stackable, default ship, and applied-grant fields
- generated player-data command/result contracts
- including inventory and repair flags in profile responses

## Does not own

Player-data does not own:

- build eligibility or loadout rules
- catalog stats and runtime weapon mapping
- match-local ammo or equipment state
- commerce authorization
- reward policy deciding when a grant is deserved
- client inventory UI
- account authentication

It validates and persists trusted grants and returns owned facts to those systems.

## Domain roles

Identity routing is fixed:

| Identity kind | Store |
| --- | --- |
| `guest` | Process-local guest memory |
| `local_profile` | Embedded SQLite |
| `authenticated_account` | Rails/API-backed account store |

The starter inventory has schema version 1, one owned `v_wing`, one owned `pulse`, no modules, both catalog entries unlocked, and the owned V-Wing selected as default.

A load distinguishes persisted inventory from a synthesized fallback and reports whether repair was attempted. Corrupt embedded or account data does not become authoritative merely because it exists; validation occurs before the result is returned as playable.

Grant processing retries optimistic conflicts and returns duplicate success when the `grant_id` has already been applied.

## Protocols and APIs

Generated command types are:

```text
player_data_load_hangar_inventory
player_data_load_hangar_inventory_result
player_data_apply_inventory_grant
player_data_apply_inventory_grant_result
```

Primary service APIs include:

```go
func (manager *InventoryManager) Load(identity protocol.PlayerDataIdentity) (InventoryLoad, error)
func (manager *InventoryManager) ApplyGrant(command protocol.PlayerDataApplyInventoryGrant) (protocol.HangarInventory, bool, error)

func (router *StoreRouter) LoadHangarInventory(identity protocol.PlayerDataIdentity) (protocol.HangarInventory, bool, error)
func (router *StoreRouter) StoreHangarInventory(identity protocol.PlayerDataIdentity, inventory protocol.HangarInventory, expectedVersion int) (protocol.HangarInventory, error)
```

The profile HTTP response includes the inventory plus `inventory_persisted`, `inventory_synthesized_fallback`, and `inventory_repair_attempted`.

## Data ownership

`HangarInventory` owns:

```text
schema_version
inventory_version
player_ref
owned_ships
owned_weapons
owned_modules
unlocked_content
stackable_items
default_owned_ship_id
applied_grant_ids
```

Owned instance IDs and catalog IDs are distinct. Item state can mark an instance unavailable or reversed without deleting historical acquisition identity.

Every successful store increments `inventory_version`. A nonnegative `expectedVersion` must equal the current stored version or the write fails with `ErrInventoryConflict`.

## Code map

```text
shared/player_data/hangar_inventory.toml
shared/packets/player_data.toml
services/player-data/playerdata/inventory_contract.go
services/player-data/playerdata/inventory_manager.go
services/player-data/playerdata/inventory_storage.go
services/player-data/playerdata/store_router.go
services/player-data/playerdata/guest_inventory_store.go
services/player-data/playerdata/embeddedsqlite/inventory_store.go
services/player-data/playerdata/rails_inventory_store.go
services/player-data/playerdata/dispatcher.go
services/player-data/protocol/packets.go
services/player-data/httpapi/profile_handler.go
```

## Tests

Tests cover starter creation, normalization, validation, missing and corrupt fallback, repair, conflict retry, grant idempotence, every storage route, SQLite persistence, Rails adapter errors, dispatcher commands, and profile response integration.

```text
services/player-data/playerdata/inventory_manager_test.go
services/player-data/playerdata/dispatcher_inventory_test.go
services/player-data/playerdata/embeddedsqlite/inventory_store_test.go
services/player-data/playerdata/rails_inventory_store_test.go
services/player-data/httpapi/profile_inventory_test.go
```

## Related docs

- [Player Data](./!INDEX.md)
- [Runtime And Store Routing](runtime-and-store-routing.md)
- [Player Inventory Persistence](../api-server/player-inventory-persistence.md)
- [Player Inventory Client](../game-server/integrations/player-inventory-client.md)
- [Inventory And Build Flow](../../domains/player-experience/inventory-and-build-flow.md)
- [Player Data Schema](../../data/player-data-schema.md)
- [Inventory And Hangar Planning](../../planning/domains/gameplay/inventory-and-hangar.md)

## Notes

The service guarantees a playable read result where possible, but it reports when that result is synthesized so callers and diagnostics can distinguish repaired durability from temporary fallback operation.
