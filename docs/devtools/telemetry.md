# Devtools Telemetry Reference

Telemetry in this document means live debug readouts, not analytics.

Related references:

- [Canonical targeting reference](../server/targeting.md)
- [Devtools targeting controls](toggles.md)

## Current Devtools Window Telemetry

The devtools window currently exposes two raw telemetry readouts from gameplay state:

- `LocalPlayerTelemetry`: displays the raw local player state dictionary from the latest gameplay state packet.
- `TargetTelemetry`: displays `target_kind`, `target_id`, and the raw resolved target state dictionary.

These readouts are generic packet/state inspection surfaces. They should not hand-map score, lives, health, shields, or other fields into custom per-stat UI logic.

## Display Behavior

- If required state is missing or unresolved, telemetry renders unavailable/empty output.
- When state is present, raw dictionary values render as key/value lines.

## Separation of Responsibilities

- HUD is player-facing.
- Devtools window owns raw inspection and controls.
- A future world telemetry overlay is separate from HUD and should provide glanceable metrics/counts (for example performance/network summaries), not raw packet dumps.

## Current Verified Baseline

- `LocalPlayerTelemetry` works.
- `TargetTelemetry` works.
- GUT baseline was green at 100 tests and 199 asserts.
