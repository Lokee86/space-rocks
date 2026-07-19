---
author: brian
created: "2026-07-19"
document_id: 019f7d55-fb2c-7de6-a90c-eb87be6faa55
document_type: general
policy_exempt: false
summary: This file captures the active constraints on player builds, ship variants, loadouts, weapons, shields, and presentation.
---
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

- The authoritative server-side `BuildEligibility`, `EligibleBuildOptions`, `LoadoutSelection`, and `ResolvedPlayerBuild` paths are implemented.
- The client does not yet provide a full pregame loadout editor or saved-loadout selector.
- Hardpoint and module-slot validation is implemented, but the current real catalog exposes only the baseline `v_wing` and `pulse` content.
- Softpoints remain runtime pickup capacity rather than pre-match selection points.
- Starting ammunition is compiled into `ResolvedPlayerBuild`; current baseline `pulse` uses infinite ammunition.
- Named saved-loadout persistence is not implemented.

### Weapons And Ship Stats

- The runtime weapon bridge remains Primary and Secondary even though the build contract models `primary_1`, `primary_2`, `secondary_1`, and `secondary_2`.
- The current `v_wing` exposes `primary_1` and `secondary_1` hardpoints; the second points are unavailable.
- The build catalog supports weapon size, delivery class, targeting policy, effect flags, ammo policy, and mode restriction filters.
- Only `pulse` is present in the default owned-equipment catalog; other weapons remain runtime/pickup content until deliberately added.
- Any remaining ship-side bullet cooldown, speed, lifetime, spawn-offset, or damage fields are legacy ownership drift against the weapon-profile model.

### Shields

- Damage resolution supports shield absorption.
- `ResolvedPlayerBuild` now carries maximum shields and starting shield policy into spawn and respawn.
- The current default catalog contains no shield module and `v_wing` resolves to zero shields.
- Shield module contracts and stat adjustments exist, but real shield equipment content, regeneration policy, and client presentation remain unimplemented.

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
