# Targeting Rule

This document defines the canonical targeting rule and the quarantine boundary for `target_player_id`.

## Canonical Target Identity

Canonical target identity is always the pair:

- `target_kind`
- `target_id`

For player targets, the canonical values are:

- `target_kind = "player"`
- `target_id = playerID`

Generic targeting systems should use `target_kind` plus `target_id`.
Player-only systems should use `playerID` directly when they already know they are dealing with a player.

## `target_player_id` Quarantine

`target_player_id` is a legacy player-only compatibility surface. It is quarantined so new gameplay code does not grow a second targeting model.

Allowed scope:

- devtools/debug player-only commands
- generated packet code only when it comes from debug/devtools packet schemas
- tests for those devtools/debug commands

Disallowed scope:

- normal gameplay packets
- `ShipState` / player state readouts
- `runtime.Ship`
- `PlayerTargeting` internal state
- new gameplay systems
- client gameplay targeting logic
- telemetry/readout display

This is a quarantine and migration rule for the next implementation step. New code should not introduce additional `target_player_id` usage outside the allowed scope above.
