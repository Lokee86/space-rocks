# Damage System

## Status Summary

- Damage is server-authoritative.
- Damage is pure resolution and does not mutate runtime entities.
- The root `internal/game` package builds damage targets and applies damage results.
- Damage does not know weapons, pickups, world maps, scoring, packets, or client rendering.
- The main damage seam lives in `services/game-server/internal/game/damage/`.

## Ownership

- `services/game-server/internal/game/damage/` owns damage data shapes, resolution, modifiers, shield handling, area damage, and damage-over-time helpers.
- `internal/game` owns the game-package adapters that build requests and apply results.
- `combat_damage_requests.go` adapts runtime collisions into damage requests.
- `combat_damage_application.go` applies damage results back onto runtime entities.
- Damage does not own game scoring, pickup logic, packet routing, or client presentation.

## Data Sources

- `services/game-server/internal/game/damage/`
- `services/game-server/internal/game/combat_damage_requests.go`
- `services/game-server/internal/game/combat_damage_application.go`
- `services/game-server/internal/game/export_devtools_toggles.go`
- `services/game-server/internal/game/events_test.go`

## Server Runtime Model

- Damage runs as pure result calculation.
- Runtime entities are passed in by the caller, not looked up inside the damage package.
- The damage package returns results for the caller to apply.
- `internal/game` builds damage targets, feeds them into the damage package, and writes `RemainingHealth` and `RemainingShield` back to runtime entities.
- `ResolveSingle` is the main server-side runtime entry point.
- `ResolveArea` and `TickDamageOverTime` are pure helpers that also stay inside the damage seam.

## Damage Request Model

- `DamageResolutionRequest` is the canonical request shape for new damage code.
- `DamageSource` carries `EntityID`, `EntityType`, and `Cause`.
- `DamageTarget` carries `EntityID`, `EntityType`, `Health`, `Shield`, and target-specific `Modifiers`.
- `DamageSpec` carries the intent: `Amount`, `Kind`, `Cause`, `BypassShield`, and nested `DoT`.
- `DamageKind` is damage flavor.
- `DamageCause` is delivery or source cause.
- `DamageCause` examples: `collision`, `projectile`, `debug`, `area`, `dot`.
- `DamageKind` examples: `kinetic`, `explosive`, `energy`, `fire`, `poison`, `true_damage`.
- `DamageKind` answers "what kind of damage is this?"
- `DamageCause` answers "how or why did this damage happen?"
- `DamageTarget` is the caller-provided runtime snapshot for one target, not a live entity handle.
- `DamageTarget` includes health, shield, and target-specific modifiers.

## Damage Result Model

- `DamageResult` is presentation-ready result data from the damage package.
- `DamageResult` is not itself a packet.
- `DamageResult` includes base amount, modified amount, applied modifiers, shield absorption, health damage, remaining health, remaining shield, ignored, destroyed, fatal, and created DoT effects.
- `DamageResult` includes enough data for runtime application and later presentation mapping.
- `DamageResult` preserves target and source identity so callers can route consequences and telemetry.
- `DamageResult` is still pure data until the caller applies it.

## Modifier Model

- Damage modifiers are filtered before they are applied.
- A modifier with an empty `Kind` applies globally.
- A modifier with a matching `Kind` applies to that damage kind.
- A modifier with a different non-empty `Kind` is ignored.
- The current modifier categories are `outgoing`, `resistance`, `vulnerability`, and `generic`.
- The current modifier operations are `add`, `multiply`, and `set`.
- Modifier application order is stable and intentional:
  - add modifiers apply first
  - multiply modifiers apply second
  - set modifiers apply last
  - the final amount is clamped below zero before rounding
- This ordering keeps the resolution math predictable and lets the caller reason about each layer separately.

## Shield Handling

- Shield handling happens after modifier math produces the final damage amount.
- If `BypassShield` is false, shield absorbs damage first.
- Any remaining damage then applies to health.
- If `BypassShield` is true, all modified damage applies directly to health.
- The damage package computes the result, but game code writes `RemainingHealth` and `RemainingShield` back to runtime entities.

## Area Damage

- `ResolveArea` is a pure resolver over the candidates supplied by game code.
- It does not inspect world maps, collision shapes, or runtime stores.
- The caller provides the candidate list, and the damage package calls `ResolveSingle` for each supplied candidate.
- That keeps spatial search and target collection in `internal/game` while damage only handles resolution.

## Damage Over Time

- Damage over time lives in the damage package because it is another form of damage resolution.
- It is not a generic status-effect system yet, and there is no `status_effects` package.
- `TickDamageOverTime` resolves each tick by calling `ResolveSingle` with `DamageCauseDot`.
- The DoT helper keeps damage math, shield handling, and modifier application in one seam.
- Future broader status effects can build on this shape later without changing the current ownership split.

## Combat Integration

- Combat owns the collision facts that decide when damage should be resolved.
- Combat builds the damage request, calls the damage seam, and then applies the returned result.
- After damage resolution, game-owned code continues with scoring, despawn, fragment spawning, pickup drops, and any other combat consequences.
- The damage package stays responsible for resolution only; it does not own combat flow.

## Devtools Integration

- Devtools must route through the same damage seam as live combat.
- Devtools should not create a parallel debug-only damage system or duplicate damage math.
- Debug damage requests may differ in source and target data, but they still resolve through `services/game-server/internal/game/damage/`.

## Client Presentation Future

- Future client rendering should receive damage results or damage events from server output.
- The client should not calculate damage locally.
- Future presentation events could include `damage_applied`, `shield_absorbed`, `damage_immune`, `dot_started`, `dot_tick`, and `damage_area_applied`.
- Those names are future presentation concepts only; they are not the current packet schema.
- Generated packet files should not be manually edited.

## Testing And Verification

- Damage package tests should cover request/result shapes, modifier ordering, shield handling, area resolution, and DoT ticking.
- Game-server verification should include `go test ./...` from `services/game-server` when a broader integration check is needed.
- Focused package tests are useful for small seam changes.
- Generated packet files should not be manually edited during verification or implementation.

## Future Work

- Future work may add weapon-driven damage shaping and pickup modifiers.
- Future work may add client render events for damage presentation.
- Future work may add area falloff rules.
- Future work may extend DoT into broader status effects.
- Future work may add richer presentation and telemetry around damage outcomes.
