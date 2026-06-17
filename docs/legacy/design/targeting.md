# Targeting Rule

This document defines the targeting quarantine rule that guides the next implementation step.

## Canonical Target Identity

Canonical target identity is always the pair:

- `target_kind`
- `target_id`

For player targets, the canonical values are:

- `target_kind = "player"`
- `target_id = playerID`

Generic targeting systems must use `target_kind` plus `target_id`.
Player-only systems should use `playerID` directly when they already know they are handling a player.

## `target_player_id` Quarantine

`target_player_id` is a legacy player-only compatibility surface.
It exists only as a migration bridge so new gameplay code does not grow a second targeting model.

Allowed scope:

- devtools/debug player-only command schemas
- devtools/debug handlers
- generated debug packet code
- tests for those devtools/debug commands

Disallowed scope:

- normal gameplay packets
- `ShipState` / player state readouts
- `runtime.Ship`
- `PlayerTargeting` internal state
- new gameplay systems
- client gameplay targeting logic
- telemetry/readout display

New gameplay systems must not introduce `target_player_id`.
The intended direction is to keep generic targeting on `target_kind` + `target_id` and keep player-only code on direct `playerID` access when that is already the owning context.
