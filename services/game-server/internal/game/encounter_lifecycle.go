package game

import (
	"fmt"
	"math"

	"github.com/Lokee86/space-rocks/services/game-server/internal/constants"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/encounterlifecycle"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/encounterspawn"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/runtime"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/space"
)

const (
	baselineAsteroidProfileID         encounterlifecycle.ProfileID         = encounterlifecycle.ProfileID(encounterspawn.ProfilePlayercentricAsteroidsV1)
	baselineAsteroidSpawnType         encounterlifecycle.SpawnType         = "asteroid"
	baselineAsteroidLifecyclePolicyID encounterlifecycle.LifecyclePolicyID = "baseline_asteroid"
	baselineAsteroidPriority          encounterlifecycle.Priority          = 0
)

func baselineAsteroidLifecycleRegistration(asteroid *runtime.Asteroid) encounterlifecycle.Registration {
	return baselineAsteroidLifecycleRegistrationForProfile(asteroid, encounterspawn.ProfilePlayercentricAsteroidsV1)
}

func baselineAsteroidLifecycleRegistrationForProfile(asteroid *runtime.Asteroid, profileID encounterspawn.ProfileID) encounterlifecycle.Registration {
	populationCost := asteroid.Size
	if populationCost < 1 {
		populationCost = 1
	}
	return encounterlifecycle.Registration{
		Origin: encounterlifecycle.OriginMetadata{
			ProfileID:              encounterlifecycle.ProfileID(profileID),
			SpawnType:              baselineAsteroidSpawnType,
			LifecyclePolicyID:      baselineAsteroidLifecyclePolicyID,
			Priority:               baselineAsteroidPriority,
			WeightedPopulationCost: encounterlifecycle.WeightedPopulationCost(populationCost),
		},
		Policy: encounterlifecycle.Policy{
			OutsideAllRelevantPlayers: encounterlifecycle.TriggerPolicy{
				Enabled:     true,
				Disposition: encounterlifecycle.DispositionHardRemove,
			},
			ExtraPlayerDistance: constants.AsteroidDespawnMargin,
		},
		Capabilities: encounterlifecycle.EntityCapabilities{
			SupportsHardRemove: true,
		},
	}
}

func (game *Game) encounterLifecycle() *encounterlifecycle.Runtime {
	if game.encounterLifecycleRuntime == nil {
		game.encounterLifecycleRuntime = encounterlifecycle.NewRuntime()
	}
	return game.encounterLifecycleRuntime
}

func (game *Game) registerAsteroidLifecycle(asteroid *runtime.Asteroid) {
	game.registerAsteroidLifecycleForProfile(asteroid, encounterspawn.ProfilePlayercentricAsteroidsV1)
}

func (game *Game) registerAsteroidLifecycleForProfile(asteroid *runtime.Asteroid, profileID encounterspawn.ProfileID) {
	if err := game.encounterLifecycle().Register(asteroid.ID, baselineAsteroidLifecycleRegistrationForProfile(asteroid, profileID)); err != nil {
		panic(fmt.Errorf("failed to register asteroid %q with encounter lifecycle: %w", asteroid.ID, err))
	}
}

func (game *Game) ensureAsteroidLifecycleRegistered(asteroid *runtime.Asteroid) encounterlifecycle.Entry {
	if entry, ok := game.encounterLifecycle().Snapshot(asteroid.ID); ok {
		return entry
	}
	game.registerAsteroidLifecycle(asteroid)
	entry, _ := game.encounterLifecycle().Snapshot(asteroid.ID)
	return entry
}

func (game *Game) advanceEncounterLifecycle(delta float64) {
	if err := game.encounterLifecycle().Advance(delta, game.worldSimulationOptions.IsWorldFrozen()); err != nil {
		panic(fmt.Errorf("failed to advance encounter lifecycle: %w", err))
	}
}

func (game *Game) evaluateAsteroidLifecycle(asteroid *runtime.Asteroid) {
	entry := game.ensureAsteroidLifecycleRegistered(asteroid)
	if entry.RetirementState != encounterlifecycle.RetirementStateActive {
		return
	}

	result, triggered, err := encounterlifecycle.Evaluate(encounterlifecycle.EvaluationRequest{
		Origin:       entry.Registration.Origin,
		Capabilities: entry.Registration.Capabilities,
		Policy:       entry.Registration.Policy,
		Facts: encounterlifecycle.EvaluationFacts{
			ElapsedLifetimeSeconds: entry.ElapsedLifetimeSeconds,
			RelevantPlayers:        game.asteroidRelevantPlayerDistances(asteroid),
			InsideAllowedRegion:    true,
			SimulationPaused:       game.worldSimulationOptions.IsWorldFrozen(),
		},
	})
	if err != nil {
		panic(fmt.Errorf("failed to evaluate asteroid %q lifecycle: %w", asteroid.ID, err))
	}
	if triggered {
		game.applyAsteroidLifecycleDecision(asteroid, result)
	}
}

func (game *Game) asteroidRelevantPlayerDistances(asteroid *runtime.Asteroid) []encounterlifecycle.RelevantPlayerDistance {
	players := make([]encounterlifecycle.RelevantPlayerDistance, 0, len(game.cameraViews))
	for _, view := range game.cameraViews {
		players = append(players, encounterlifecycle.RelevantPlayerDistance{
			Distance:       space.Distance(view.Position(), asteroid.Position()),
			ViewableRadius: math.Hypot(view.VisibleWorldWidth()*0.5, view.VisibleWorldHeight()*0.5),
		})
	}
	return players
}

func (game *Game) applyAsteroidLifecycleDecision(asteroid *runtime.Asteroid, result encounterlifecycle.EvaluationResult) {
	if err := game.encounterLifecycle().BeginRetirement(asteroid.ID, result); err != nil {
		panic(fmt.Errorf("failed to begin asteroid %q retirement: %w", asteroid.ID, err))
	}

	switch result.Disposition {
	case encounterlifecycle.DispositionHardRemove:
		game.removeAsteroidAuthoritatively(asteroid.ID)
	case encounterlifecycle.DispositionSoftRetire:
		asteroid.MarkPendingDespawn(constants.CollisionDespawnDelay)
	default:
		panic(fmt.Errorf("unsupported asteroid lifecycle disposition %q", result.Disposition))
	}
}

func (game *Game) removeAllAsteroidsForLifecycleTrigger(trigger encounterlifecycle.Trigger) int {
	for _, asteroid := range game.entities.Asteroids {
		game.ensureAsteroidLifecycleRegistered(asteroid)
	}

	removed := 0
	for _, entityID := range game.encounterLifecycle().EntityIDs() {
		asteroid, exists := game.entities.Asteroids[entityID]
		if !exists {
			continue
		}
		entry, _ := game.encounterLifecycle().Snapshot(entityID)
		if entry.RetirementState == encounterlifecycle.RetirementStateActive {
			game.applyAsteroidLifecycleDecision(asteroid, encounterlifecycle.EvaluationResult{
				Trigger:     trigger,
				Disposition: encounterlifecycle.DispositionHardRemove,
			})
		} else {
			game.removeAsteroidAuthoritatively(entityID)
		}
		removed++
	}
	return removed
}

func (game *Game) removeAsteroidAuthoritatively(entityID string) bool {
	if _, exists := game.entities.Asteroids[entityID]; !exists {
		return false
	}
	delete(game.entities.Asteroids, entityID)
	game.encounterLifecycle().Remove(entityID)
	game.damageOverTime().RemoveTarget(entityID)
	return true
}
