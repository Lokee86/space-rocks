# Encounter Spawn And Lifecycle

Parent index: [Encounters](./!INDEX.md)

## Purpose

This document describes the current game-server boundaries for encounter spawn-profile scheduling and post-spawn entity lifecycle management.

## Overview

Encounter spawning decides when a configured profile may attempt a batch. Encounter lifecycle records what created each admitted entity, accounts for its weighted population, evaluates retirement triggers, and hands a soft-retire or hard-remove decision to the entity owner.

```text
resolved encounter profile selection
-> encounterspawn.Runtime
-> priority-ordered spawn opportunity
-> game placement and entity creation
-> immutable OriginMetadata registration
-> encounterlifecycle.Runtime
-> trigger evaluation
-> soft retire or hard remove
-> profile population accounting release
```

The current production profile is `playercentric_asteroids_v1`. The contracts also model continuous, wave, event, objective, and scripted schedules; the runtime currently emits scheduled opportunities for continuous profiles.

## Code root

```text
services/game-server/internal/game/encounterspawn/
services/game-server/internal/game/encounterlifecycle/
```

Game integration lives in:

```text
services/game-server/internal/game/encounter_spawn.go
services/game-server/internal/game/encounter_lifecycle.go
```

## Responsibilities

Encounter spawn owns:

- profile configuration and validation
- configured, active, and deactivated profile state
- pause-aware and stop-aware schedule advancement
- priority-ordered continuous spawn opportunities
- batch size and interval policy
- shared, per-profile, and per-spawn-type weighted limits
- retry-cap declaration
- activation, deactivation, stop, resume, and progress reset

Encounter lifecycle owns:

- immutable origin profile, spawn type, policy, priority, cost, and optional phase
- entity cleanup capabilities
- elapsed lifetime while simulation is active
- weighted population totals by profile
- retirement trigger and disposition validation
- one-way transition from active to retirement begun
- accounting release when the entity is removed
- deterministic snapshot and entity-ID ordering

## Does not own

These packages do not own:

- concrete entity construction
- spatial placement or collision safety
- asteroid variant selection
- damage, death, or destruction consequences
- visual despawn effects
- objective definition or mission scripting
- room transitions

The game layer performs those actions using the opportunity and lifecycle decisions.

## Domain roles

Spawn profiles are schedule and budget policy. They can be activated or stopped without deleting their configuration.

Every lifecycle-managed entity carries origin metadata so later cleanup can be attributed to the correct profile and weighted budget. Lifecycle decisions validate both the configured trigger and the entity's capabilities. An entity that requires explicit cleanup must soft-retire; an entity that requires destruction must hard-remove. Transition/reset cleanup is always hard removal.

Supported retirement triggers are:

```text
lifetime_expiry
outside_all_relevant_players
allowed_region_exit
population_pressure
profile_phase_cleanup
scripted_cleanup
transition_reset
```

## Protocols and APIs

Important internal APIs include:

```go
func (runtime *encounterspawn.Runtime) Configure(config encounterspawn.Config) error
func (runtime *encounterspawn.Runtime) Activate(profileID encounterspawn.ProfileID) error
func (runtime *encounterspawn.Runtime) Step(delta float64, simulationPaused bool) ([]encounterspawn.Opportunity, error)

func (runtime *encounterlifecycle.Runtime) Register(entityID string, registration encounterlifecycle.Registration) error
func (runtime *encounterlifecycle.Runtime) Advance(delta float64, simulationPaused bool) error
func (runtime *encounterlifecycle.Runtime) BeginRetirement(entityID string, result encounterlifecycle.EvaluationResult) error
func (runtime *encounterlifecycle.Runtime) Remove(entityID string) (encounterlifecycle.Entry, bool)
func encounterlifecycle.Decide(request encounterlifecycle.DecisionRequest) (encounterlifecycle.Decision, error)
```

These are internal simulation APIs. No client packet can create a spawn opportunity or retirement decision directly.

## Data ownership

`encounterspawn.Config` owns profile schedule and budget declarations. Runtime snapshots clone map-backed limits so callers cannot mutate configured state.

`encounterlifecycle.OriginMetadata` is immutable accounting provenance. `Runtime` owns active entries, elapsed lifetime, retirement state, and per-profile weighted totals.

A lifecycle entry remains accounted until `Remove` succeeds. Beginning retirement does not release the weighted population early.

## Code map

```text
services/game-server/internal/game/encounterspawn/contract.go
services/game-server/internal/game/encounterspawn/runtime.go
services/game-server/internal/game/encounterlifecycle/contract.go
services/game-server/internal/game/encounterlifecycle/policy.go
services/game-server/internal/game/encounterlifecycle/evaluate.go
services/game-server/internal/game/encounterlifecycle/decision.go
services/game-server/internal/game/encounterlifecycle/runtime.go
services/game-server/internal/game/encounter_spawn.go
services/game-server/internal/game/encounter_lifecycle.go
services/game-server/internal/game/simulation_asteroids.go
```

## Tests

Tests cover profile validation and cloning, activation state, cadence and priority, pause/stop behavior, origin validation, capabilities, every trigger family, retirement idempotence, lifetime advancement, population accounting, and game integration.

```text
services/game-server/internal/game/encounterspawn/*_test.go
services/game-server/internal/game/encounterlifecycle/*_test.go
services/game-server/internal/game/encounter_spawn_*_test.go
services/game-server/internal/game/encounter_lifecycle_*_test.go
```

## Related docs

- [Encounters](./!INDEX.md)
- [Asteroid Spawning And Variants](../world/asteroid-spawning-and-variants.md)
- [Modes And Match Rules](../match/modes-and-match-rules.md)
- [Encounter Spawn Profiles Planning](../../../../planning/domains/gameplay/encounter-spawn-profiles.md)
- [Encounter Lifecycle And Despawn Planning](../../../../planning/domains/gameplay/encounter-lifecycle-and-despawn.md)

## Notes

Only continuous scheduling is currently executed by the generic spawn runtime. Other schedule kinds are established vocabulary and validation seams for later mode, objective, event, or script owners.
