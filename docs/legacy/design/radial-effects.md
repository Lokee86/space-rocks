# Radial Effects

Parent index: [Design Legacy](./!README.md)

## Status Summary

- Radial effects are server-authoritative runtime effects.
- `services/game-server/internal/game/effects/radial` owns radial specs, coverage/expiration modes, zones, target filters, stores, stepping, and hit intents.
- Root `internal/game` spawns radial effects from projectile impact metadata, builds candidates from runtime maps, steps active effects, and applies returned hits through the damage seam.
- `Torpedo` currently uses a radial impact effect with annular-wave coverage, simultaneous expiration, and asteroid/enemy-only filtering.
- Radial hits are intents/facts; `damage.ResolveSingle` still performs the actual damage math.

## Ownership

- `services/game-server/internal/game/effects/radial` owns radial effect data shapes, spec and timing modes, coverage modes, target filters, zones, the effect store, stepping, and hit intent generation.
- `services/game-server/internal/game/effects/radial` does not own runtime entity maps, projectile collision facts, damage resolution, asteroid destruction, scoring, player death handling, or client rendering.
- Root `internal/game` owns spawning radial effects from projectile impacts, building candidates from runtime entities, applying radial hits to concrete runtime entities, converting hits into damage requests, and consequence handling.
- `services/game-server/internal/game/radial_spawning.go` bridges projectile impacts into radial effect spawn requests.
- `services/game-server/internal/game/radial_candidates.go` builds the runtime candidate list.
- `services/game-server/internal/game/simulation_radial_effects.go` steps stored radial effects and applies their hits.
- `services/game-server/internal/game/radial_damage_requests.go` converts radial hits into damage requests.

## Data Sources

- `services/game-server/internal/game/effects/radial/`
- `services/game-server/internal/game/radial_candidates.go`
- `services/game-server/internal/game/simulation_radial_effects.go`
- `services/game-server/internal/game/radial_damage_requests.go`
- `services/game-server/internal/game/radial_spawning.go`

## Radial Spec Model

- `radial.Spec` is the declarative radial effect shape.
- `radial.Spec` fields are `CoverageMode`, `ExpirationMode`, `TargetFilter`, `ZoneCount`, `ZoneWidth`, `ZoneSpawnSeconds`, `TickSeconds`, `TotalSeconds`, `ZoneLifetimeSeconds`, and `Damage`.
- `Damage` is a `damage.DamageSpec`.
- `radial.Spec` describes timing, coverage, targets, and damage intent.
- `radial.Spec` is not itself runtime mutation.
- Target inclusion is explicit through `TargetFilter`.

## Coverage Modes

- `CoverageAnnularWave` is the banded zone mode.
- In annular wave, active zones hit candidates within each zone's inner/outer radius band.
- In annular wave, zones can tick independently while active.
- `CoverageExpandingFill` is the filled-radius mode.
- In expanding fill, the active radius expands outward and includes the filled area from center to the current radius.

## Expiration Modes

- `ExpirationSimultaneous` makes zones expire at total effect time.
- `ExpirationSequential` makes each zone expire based on its own start time plus zone lifetime.
- Expiration mode controls when a zone can still participate in stepping.

## Zone Timing Model

- Zones are built from `ZoneCount`, `ZoneWidth`, `ZoneSpawnSeconds`, `TotalSeconds`, and `ZoneLifetimeSeconds`.
- Zone fields are `Index`, `InnerRadius`, `OuterRadius`, `StartsAt`, `ExpiresAt`, and `NextTickAt`.
- `Index` identifies the zone in build order.
- `InnerRadius` and `OuterRadius` define the zone's radial band.
- `StartsAt` is when the zone becomes active.
- `ExpiresAt` is when the zone stops participating in stepping.
- `NextTickAt` is the next effect-age threshold at which the zone can tick again.
- A zone only ticks after `StartsAt`.
- Expired zones do not tick.
- A zone only ticks when effect age reaches `NextTickAt`.
- After ticking, `NextTickAt` advances by `TickSeconds`.
- One tick can represent a gameplay interval, not necessarily one render frame.
- Repeated ticks are how bloom damage can apply multiple damage pulses.
- `effect.AgeSeconds` is the running effect clock used by `Step`.
- `StartsAt`, `ExpiresAt`, and `NextTickAt` are compared against `effect.AgeSeconds`, not render frames.
- In simultaneous expiration, all zones stop when `effect.AgeSeconds` reaches `TotalSeconds`.
- In sequential expiration, each zone stops at its own `StartsAt + ZoneLifetimeSeconds`.
- Repeated ticks happen when the effect remains active across multiple tick thresholds, which is why bloom damage can apply multiple pulses.

## Target Filtering

- Target kinds are `asteroid`, `enemy`, `player`, `projectile`, and `pickup`.
- `TargetFilter.Allows` decides whether a target kind is included.
- `TargetFilter` is explicit; kinds are either allowed or disallowed.
- Current torpedo target filter includes asteroids and enemies only.

## Candidate And Hit Model

- `Candidate` fields are `ID`, `Kind`, and `Position`.
- `Hit` fields are `EffectID`, `SourceID`, `SourcePlayerID`, `ZoneIndex`, `TargetID`, `TargetKind`, `TargetPosition`, and `Damage`.
- Root `internal/game` builds candidates from all relevant runtime entity maps.
- The radial package filters candidates by target kind and spatial coverage.
- A hit is an intent or fact, not a mutation.
- `Hit` carries the data needed for later game-owned consequence handling and damage application.

## Step Flow

- End-to-end flow:
  - projectile impact metadata spawns a radial effect
  - `radial.Store` holds active effects
  - Game builds candidates
  - `radial.Step` evaluates active zones against those candidates
  - `Step` emits hits
  - Game applies hits
  - expired effects are removed from the store
- `radial.Step` receives an effect, delta time, and a candidate list.
- `Step` checks effect expiration first.
- `Step` evaluates coverage, filters candidates, emits hit intents, advances zone tick times, advances effect age, and reports expiration through `StepResult`.
- `Step` does not look up runtime entities.
- `Step` does not resolve damage.
- `Step` does not mutate health.
- `Step` does not remove effects from the store.
- `Step` does not spawn effects, build candidates, or apply hits.
- `Store` owns active effect storage.
- Game decides when to remove expired effects from the store.

## Game Integration

- `Game` owns a `radial.Store`.
- Radial effects are spawned from projectile impact metadata.
- `services/game-server/internal/game/radial_spawning.go` turns projectile impact metadata into radial effect spawn requests.
- `Game` builds candidate lists from runtime maps.
- `Game` steps active radial effects during simulation.
- `Game` applies returned hits to concrete entities.
- `Game` removes expired effects from the store.
- `services/game-server/internal/game/simulation_radial_effects.go` is the main simulation bridge for stepping and consequence handling.
- `Game` is the adapter between radial hit intents and concrete runtime mutation.
- Radial behavior stays in the radial package; lifecycle orchestration stays in Game.

## Damage Integration

- Radial hits are converted into `DamageResolutionRequest` values by Game-owned adapters.
- `radialDamageRequestFromHitAndAsteroid` adapts asteroid hits.
- `radialDamageRequestFromHitAndEnemy` adapts enemy hits.
- `radialDamageRequestFromHitAndPlayer` adapts player hits.
- Radial damage source uses projectile identity and area cause.
- `normalizeRadialDamageSpec` supplies area cause when missing.
- `damage.ResolveSingle` handles the actual damage math.
- Asteroid damage can cause destruction, scoring, drop, and fragment consequences through existing Game code.
- Player fatal handling routes through existing player fatal damage flow if player target filtering enables players.
- Weapon and radial code stay out of the damage math itself.
- See [docs/design/damage.md](damage.md) for the damage seam details.

## Current Torpedo Bloom Behavior

- Torpedo bloom is spawned from torpedo projectile impact metadata.
- Torpedo direct impact damage and radial bloom damage are separate.
- Direct impact damage is projectile damage.
- Radial bloom damage is area damage from radial hits.
- Current torpedo radial spec uses `CoverageAnnularWave`.
- Current torpedo radial spec uses `ExpirationSimultaneous`.
- Current torpedo target filter includes asteroids and enemies.
- Current torpedo radial damage is explosive area damage.
- Zone count, zone width, spawn interval, tick interval, total time, and zone lifetime are tunable weapon constants.
- Repeated radial ticks can damage targets multiple times if they remain in active coverage over time.
- This section describes current behavior, not a final balance statement.

## Event And Presentation Model

- Radial effect spawning records `radial_effect_started` for presentation and event flow.
- `services/game-server/internal/game/events.go` maps the domain event into the presentation event stream.
- `services/game-server/internal/game/events/events.go` defines the domain event type.

## Related Limits

- [Current System Limits](../limits/current-system-limits.md)

## Testing And Verification

- `cd services/game-server`
- `env GOCACHE=/tmp/space-rocks-go-build go test -buildvcs=false ./...`
- Focused tests live under `services/game-server/internal/game/effects/radial/*_test.go` and cover zone building, coverage modes, stepping, store behavior, and target filtering.
- `services/game-server/internal/game/simulation_radial_effects.go` and `services/game-server/internal/game/radial_spawning.go` are the key integration seams to re-check when radial spawning or hit handling changes.
- If the broader Go test run is unavailable, run the focused radial package tests first.