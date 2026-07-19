# Player Inventory Persistence

Parent index: [API Server](./!INDEX.md)

## Purpose

This document describes authenticated-account hangar inventory persistence in the API server.

## Overview

The API server stores one versioned JSON inventory per user and exposes internal load/store endpoints to the player-data service.

```text
player-data RailsStore
-> internal token authorization
-> account_id lookup
-> PlayerInventory load or store
-> user-row lock + expected-version check
-> versioned JSON response
```

The API server is the durability owner for authenticated-account inventory only. Guest and local-profile inventory never route here.

## Code root

```text
services/api-server/app/models/player_inventory.rb
services/api-server/app/controllers/api/internal/player_data/inventories_controller.rb
```

## Responsibilities

This boundary owns:

- one inventory row per user
- JSON-object validation
- positive integer inventory versions
- user-locked create/update transactions
- optional expected-version comparison
- version increment on every accepted store
- unknown-user, invalid-input, not-found, and conflict responses
- internal service authorization inherited from the API internal base controller

## Does not own

The API server does not own:

- hangar schema semantics
- starter inventory creation
- inventory repair policy
- grant application
- loadout eligibility
- local-profile SQLite data
- guest memory data
- client inventory presentation

Player-data owns schema validation and mutation before calling this persistence surface.

## Domain roles

`PlayerInventory.store_for_user!` locks the user row, reads the current inventory version, compares an optional expected version, increments the version, writes that version into the JSON object, and creates or updates the single inventory record.

A mismatch raises `VersionConflict` and returns HTTP 409 with `inventory_version_conflict`. A missing inventory load returns `{ found: false }`; an unknown account returns 404.

## Protocols and APIs

Internal endpoints are:

```text
POST /api/internal/player-data/inventory/load
POST /api/internal/player-data/inventory/store
```

Load input contains `account_id`.

Store input contains:

```text
account_id
inventory JSON object
optional expected_version
```

Successful responses include the inventory and authoritative `inventory_version`.

## Data ownership

The `player_inventories` table owns:

```text
user_id, unique
inventory, JSON object
inventory_version, positive integer
timestamps
```

The JSON document is opaque to Rails beyond object validation and version insertion. Player-data remains responsible for the canonical hangar schema.

## Code map

```text
services/api-server/app/models/player_inventory.rb
services/api-server/app/models/user.rb
services/api-server/app/controllers/api/internal/player_data/inventories_controller.rb
services/api-server/config/routes.rb
services/api-server/db/migrate/20260608001100_create_player_inventories.rb
services/api-server/db/schema.rb
```

## Tests

Tests cover model validation, one-row-per-user ownership, versioned create/update, expected-version conflicts, endpoint authorization and validation, missing inventory, unknown users, and successful load/store responses.

```text
services/api-server/test/models/player_inventory_test.rb
services/api-server/test/controllers/api/internal/player_data/inventories_controller_test.rb
```

## Related docs

- [API Server](./!INDEX.md)
- [Internal API Surface](internal-api-surface.md)
- [Hangar Inventory](../player-data/hangar-inventory.md)
- [Inventory And Build Flow](../../domains/player-experience/inventory-and-build-flow.md)

## Notes

Rails persists the player-data-owned document and enforces concurrency. It deliberately does not duplicate the Go service's inventory rules.
