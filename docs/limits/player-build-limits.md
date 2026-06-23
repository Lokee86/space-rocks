# Player Build Limits
Parent index: [Current Limits](./!INDEX.md)

## Purpose

This file captures the active constraints on player builds, ship variants, loadouts, weapons, shields, and presentation.

## Overview

It serves as the current-limits companion for player-build work. The entries below describe present-day caps, missing wiring, and fallback behavior that still shape gameplay and client presentation.

## Issue list or backlog

### Ship Variants

- Only the default ship type `v_wing` is currently used.
- Full selectable ship variants are not implemented.
- The server has only one imported ship collision shape in `shared/collisions/collision_shapes.json`.
- The collision shape ID seam exists, but a keyed multi-ship collision catalog is not implemented.

### Loadouts

- There is no full pre-match `LoadoutSelection` system yet.
- There is no `ResolvedPlayerBuild` path yet.
- Hardpoints and softpoints are planning concepts, not implemented loadout validation.
- Module slot selection is not implemented.
- Starting ammunition is not yet owned by a loadout model.

### Weapons And Ship Stats

- The current weapon equip model is only Primary and Secondary.
- Full hardpoint/softpoint loadout validation is not implemented.
- Full weapon classification fields are not implemented.
- Any remaining ship-side bullet cooldown, speed, lifetime, spawn-offset, or damage fields are legacy ownership drift against the weapon-profile model.

### Shields

- Damage resolution supports shield absorption.
- Full player-build shield ownership through ship variants or loadouts is not implemented.
- Ship-variant max-shield setup is not implemented as part of the player-build model.
- Loadout-driven shield modules are not implemented.
- Treat current shield support as damage/runtime support, not a complete player-build system.

### Client Presentation

- `ship_type` exists in `ShipState`.
- The client receives `ship_type`, but current player rendering does not select a different ship scene from it.

## Affected docs/systems

- [Current System Limits](current-system-limits.md)
- [Development Roadmap](../planning/development-roadmap.md)
- [Player Build And Loadouts](../planning/domains/gameplay/player-build-and-loadouts.md)
- [Domain Backlog](../planning/domain-backlog.md)

## Status

Active current-limits document. The entries below describe present-day player-build constraints and incomplete behavior in the live system.

## Related docs

- [Current System Limits](current-system-limits.md)
- [Development Roadmap](../planning/development-roadmap.md)
- [Player Build And Loadouts](../planning/domains/gameplay/player-build-and-loadouts.md)
- [Domain Backlog](../planning/domain-backlog.md)

## Notes

Keep this file focused on current player-build constraints. Use planning docs for future loadout or ship-variant design work.
