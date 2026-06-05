# Devtools Telemetry Reference

Telemetry in this document means live debug readouts, not analytics.

Related references:

- [Canonical targeting reference](../server/targeting.md)
- [Devtools targeting controls](toggles.md)

## Current Devtools Window Telemetry

The devtools window currently exposes two raw telemetry readouts from gameplay state:

- `LocalPlayerTelemetry`: supports `State Packet` and `Session Packet`.
- `TargetTelemetry`: supports `State Packet` and `Session Packet`.

These readouts are generic packet/state inspection surfaces. They should not hand-map score, lives, health, shields, or other fields into custom per-stat UI logic.
`State Packet` is the live ship/avatar read model, while `Session Packet` is the durable player/session read model. Score and lives are read from `player_sessions`, not `players`.
The entity-state telemetry source can also display canonical non-player targets such as asteroids, bullets, enemies, and pickups.
When pickup telemetry is shown from `StatePacket.pickups`, the `health` field is current health only. This does not mean bullet/pickup collision damage is enabled yet.

## Display Behavior

- If the selected source has no matching data, telemetry renders unavailable/empty output.
- `Session Packet` for non-player targets renders unavailable/empty output.
- When state is present, raw dictionary values render as key/value lines.
- `TargetTelemetry` still shows `target_kind` and `target_id` above the raw dictionary body when a target is selected.

## Separation of Responsibilities

- HUD is player-facing.
- Devtools window owns raw inspection and controls.
- A future world telemetry overlay is separate from HUD and should provide glanceable metrics/counts (for example performance/network summaries), not raw packet dumps.

## Current Verified Baseline

- `LocalPlayerTelemetry` works.
- `TargetTelemetry` works.
- GUT baseline was green at 100 tests and 199 asserts.
