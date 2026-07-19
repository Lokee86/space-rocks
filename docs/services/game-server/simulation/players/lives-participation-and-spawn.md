# Lives, Participation, And Spawn

Parent index: [Game Server Simulation Players](./!INDEX.md)

## Purpose

This document describes the current authoritative game-server boundary for lives, death history, respawn eligibility, AFK forfeiture, and player spawn placement.

## Overview

The simulation owns a dedicated lives runtime instead of storing lifecycle authority on the active ship. Player sessions and runtime ships are consumers of that state.

```text
resolved lives policy
-> register participant
-> active / pending_respawn / eliminated / removed state
-> authoritative death transition and history
-> respawn eligibility and restoration
-> basic_safe_spawn_v1 placement
-> runtime ship recreation
```

Implemented life models are finite per-player lives, shared team pools, and infinite lives. The baseline uses finite per-player lives, manual respawn, generated starting-life and respawn-delay constants, and `basic_safe_spawn_v1`.

## Code root

```text
services/game-server/internal/game/lives/
services/game-server/internal/game/participation/
services/game-server/internal/game/playerspawn/
```

Game integration lives in `services/game-server/internal/game/`.

## Responsibilities

This boundary owns:

- validating and cloning lives policies
- participant and team-pool registration
- personal, shared-team, and infinite-life accounting
- active, pending-respawn, eliminated, and removed transitions
- authoritative death facts and retained death history
- respawn counts, cooldowns, and eligibility
- life grants and targeted recovery from death exhaustion
- restoration policy for health, shields, cooldowns, temporary effects, and loadout state
- active and pending-respawn AFK tracking
- `afk_forfeit` expiry requests
- deterministic safe-spawn search around a preferred origin
- life and team-pool snapshots for projection and results

## Does not own

This boundary does not own:

- damage calculation or lethal-source eligibility
- room connection and removal execution
- team assignment
- mode victory evaluation
- inventory durability or loadout selection
- client respawn UI
- spawn protection
- scoring or award policy

It emits normalized lifecycle facts and requests actions from those owners.

## Domain roles

The lives runtime is the authority for whether a participant still has a normal path back into active play. Destroying a runtime ship does not erase participant counters or history.

For shared team pools, each qualifying death consumes one pool life. Active teammates remain active when the pool reaches zero; teammates already waiting to respawn become death-exhausted and eliminated. A later targeted life grant can restore an eligible exhausted participant to pending respawn.

The AFK runtime tracks active and pending-respawn participants continuously. Any accepted player action resets the timer. The baseline timeout is 35 seconds; expiry creates an `afk_forfeit` removal request while retained match facts remain available.

Player spawn placement is delegated to `basic_safe_spawn_v1`. It tests the preferred origin first, then searches deterministic square rings at 160-unit spacing until the safety evaluator accepts a location. The current profile does not grant spawn protection.

## Protocols and APIs

Important internal APIs include:

```go
func NewRuntime(policy lives.Policy) (*lives.Runtime, error)
func (runtime *lives.Runtime) RegisterParticipant(registration lives.ParticipantRegistration) error
func (runtime *lives.Runtime) ParticipantSnapshot(playerID string) (lives.ParticipantState, bool)
func (runtime *lives.Runtime) TeamPoolSnapshot(teamID teams.ID) (lives.TeamPoolState, bool)
func (runtime *lives.Runtime) DeathHistory(playerID string) ([]lives.DeathFact, bool)

func participation.NewRuntime(policy participation.AFKPolicy) (*participation.Runtime, error)
func (runtime *participation.Runtime) RecordAction(playerID string) bool
func (runtime *participation.Runtime) Step(delta float64, status ...) []participation.ExpiryRequest

func playerspawn.PlanBasicSafeSpawnV1(request playerspawn.Request, evaluator playerspawn.SafetyEvaluator) (playerspawn.Plan, error)
```

Game-owned control and simulation methods translate these pure/runtime contracts into ship destruction, participant removal, respawn, projection, and presentation events.

## Data ownership

`lives.Runtime` owns participant lifecycle counters and death history. `playerSession` mirrors current values needed for runtime projection, but the active `runtime.Ship` is not the durable owner.

The restoration policy currently supports:

```text
full or no health restoration
full or no shield restoration
short-cooldown reset threshold, baseline 10 seconds
temporary-effect remove or persist
loadout persist or reset
```

`playerspawn.Request` carries a profile ID, player ID, spawn reason, preferred origin, and collision-shape ID. The returned plan owns only the chosen position; ship creation remains game-owned.

## Code map

```text
services/game-server/internal/game/lives/contract.go
services/game-server/internal/game/lives/policy.go
services/game-server/internal/game/lives/runtime.go
services/game-server/internal/game/lives/death.go
services/game-server/internal/game/lives/respawn.go
services/game-server/internal/game/lives/team_pool.go
services/game-server/internal/game/participation/afk.go
services/game-server/internal/game/playerspawn/profile.go
services/game-server/internal/game/playerspawn/basic_safe.go
services/game-server/internal/game/control_respawn.go
services/game-server/internal/game/control_player_spawn.go
services/game-server/internal/game/player_lives_projection.go
services/game-server/internal/game/session.go
```

## Tests

Coverage includes all life models, death attribution/history, shared-pool exhaustion and grants, respawn restoration, AFK expiry, deterministic placement, participant removal, projections, and registration rollback.

```text
services/game-server/internal/game/lives/*_test.go
services/game-server/internal/game/participation/afk_test.go
services/game-server/internal/game/playerspawn/basic_safe_test.go
services/game-server/internal/game/control_respawn_test.go
services/game-server/internal/game/respawn_restoration_test.go
services/game-server/internal/game/shared_lives_projection_test.go
```

## Related docs

- [Game Server Simulation Players](./!INDEX.md)
- [Player Death And Despawn](player-death-and-despawn.md)
- [Player Respawn](player-respawn.md)
- [Player Session State](player-session-state.md)
- [Player Builds And Loadouts](player-builds-and-loadouts.md)
- [Lives, Death, Elimination, And Respawn Planning](../../../../planning/domains/gameplay/lives-death-elimination-and-respawn.md)
- [Participation And Joining Planning](../../../../planning/domains/gameplay/participation-and-joining.md)
- [Player Spawn Profiles Planning](../../../../planning/domains/gameplay/player-spawn-profiles.md)

## Notes

The current implementation deliberately separates lifecycle authority, placement, and runtime ship state. Recovery rules, non-manual triggers, spawn protection, and richer reconnect semantics remain future mode/lifecycle work.
