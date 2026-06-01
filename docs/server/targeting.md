# Targeting System Reference

This document defines targeting terms used by gameplay state, devtools, and telemetry.

Related references:

- [Client mouse input seam](../client/mouse-input.md)
- [Server devtools targeting behavior](devtools.md)
- [Devtools telemetry readouts](../devtools/telemetry.md)

## Canonical Gameplay Target

The canonical gameplay target is the shared target stored in player/gameplay state.

- `target_kind`: target type discriminator
- `target_id`: identifier within that target type

Supported `target_kind` values:

- `player`
- `enemy`
- `asteroid`
- `bullet`

This canonical target is the common source used by client selection, devtools target resolution, and debug telemetry readouts.

## Per-Tool Devtools Target

A per-tool devtools target is command-specific targeting input used by an individual debug tool or action.

- It may use direct tool inputs, or
- it may resolve from the canonical gameplay target (for tools that support `Game Target`)

Per-tool target behavior is owned by that specific devtools command, not by the canonical target itself.

## Game Target Dropdown Option

`Game Target` is a devtools option that resolves from canonical gameplay target fields (`target_kind` + `target_id`).

For player-only devtools commands:

- `Game Target` may resolve only when `target_kind == "player"`
- resolved `target_id` is used as `target_player_id`

If canonical target is non-player (`enemy`, `asteroid`, or `bullet`), `Game Target` is not a valid `target_player_id` source for player-only commands.

## Telemetry Target Readout

Telemetry target readout is display-only status derived from canonical target data.

- Non-player canonical targets are valid for telemetry/readout purposes.
- Non-player canonical targets are not valid player-command targets.

Raw `LocalPlayerTelemetry` and `TargetTelemetry` in the devtools window are part of this readout surface and are separate from HUD behavior.
