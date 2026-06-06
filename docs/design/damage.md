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
- Runtime ships and asteroids may carry `DamageModifiers` in game state.
- The damage package returns results for the caller to apply.
- `internal/game` builds damage targets, feeds them into the damage package, and writes `RemainingHealth` and `RemainingShield` back to runtime entities.
- `combat_damage_requests.go` is the adapter between runtime entities and `DamageTarget`.
- `ResolveSingle` is the main server-side runtime entry point.
- `ResolveArea` and `TickDamageOverTime` are pure helpers that also stay inside the damage seam.

## Damage Request Model

- `DamageResolutionRequest` is the canonical request shape for new damage code.
- `DamageSource` carries `EntityID`, `EntityType`, and `Cause`.
- `DamageTarget` carries `EntityID`, `EntityType`, `Health`, `Shield`, and target-specific `Modifiers`.
- `DamageSpec` carries the intent: `Amount`, `Type`, `Cause`, `BypassShield`, and nested `DoT`.
- `DamageType` is damage flavor.
- `DamageCause` is delivery or source cause.
- `DamageCause` examples: `collision`, `projectile`, `debug`, `area`, `dot`.
- `DamageType` examples: `kinetic`, `explosive`, `energy`, `thermal`, `radioactive`, `true_damage`.
- `DamageType` answers "what type of damage is this?"
- `DamageCause` answers "how or why did this damage happen?"
- `DamageTarget` is the caller-provided runtime snapshot for one target, not a live entity handle.
- `DamageTarget` includes health, shield, and target-specific modifiers.
- Runtime entity modifiers are copied into `DamageTarget.Modifiers` by the game-package request builders.
- Modifiers are per `DamageType`, not flat resistance profiles.
- The damage package consumes the modifiers but does not own runtime entity storage.

## Damage Result Model

- `DamageResult` is presentation-ready result data from the damage package.
- `DamageResult` is not itself a packet.
- `DamageResult` includes base amount, modified amount, applied modifiers, shield absorption, health damage, remaining health, remaining shield, ignored, destroyed, fatal, and created DoT effects.
- `DamageResult` includes enough data for runtime application and later presentation mapping.
- `DamageResult` preserves target and source identity so callers can route consequences and telemetry.
- `DamageResult` is still pure data until the caller applies it.

## Modifier Model

- Modifier values are floats.
- Damage modifiers are keyed by `DamageType`.
- A modifier with an empty `DamageType` applies globally.
- A modifier with a matching `DamageType` applies to that damage type.
- A modifier with a different non-empty `DamageType` is ignored.
- The current modifier categories are `outgoing`, `resistance`, `vulnerability`, and `generic`.
- The current modifier operations are `add`, `multiply`, and `set`.
- Outgoing and generic modifiers stay on the normal add/multiply/set path.
- Resistance semantics:
  - resistance values are resistance amounts
  - valid range is `0 <= value < 1`
  - `0.25` means 25% resistance
  - applied as `damage *= (1 - value)`
  - multiple resistances stack by multiplying the remaining damage
- Vulnerability semantics:
  - vulnerability values are damage multipliers
  - valid range is `value > 1`
  - `1.25` means +25% incoming damage
  - applied as `damage *= value`
- Invalid resistance and vulnerability modifiers are ignored.
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
- Future presentation events could include `shield_absorbed`, `damage_immune`, and `damage_area_applied`.
- `damage_applied`, `damage_over_time_started`, and `damage_over_time_tick` now exist as implemented domain-event names, but client rendering is still not implemented from them here.
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

## Damage Events And Presentation

- `DamageResult` is not a domain event.
- Applying a `DamageResult` is game-owned state mutation, not a domain event.
- Game-owned damage application emits damage domain events when useful.
- Implemented domain-event names include `damage_applied`, `damage_over_time_started`, and `damage_over_time_tick`.
- `damage_applied` is wired for current live combat damage paths.
- `damage_over_time_started` and `damage_over_time_tick` have adapters/mapping, but active DoT gameplay ownership is not fully wired here unless the code says otherwise.
- Possible future presentation events still include `shield_absorbed`, `damage_immune`, and `damage_area_applied`.
- Those names are presentation concepts unless they are already wired elsewhere.
