package game

import (
	"fmt"
	"sort"

	"github.com/Lokee86/space-rocks/services/game-server/internal/constants"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/encounterlifecycle"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/encounterspawn"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/runtime"
)

const baselineEncounterSpawnRetryCap = 64

func baselineEncounterSpawnConfig() encounterspawn.Config {
	return encounterspawn.Config{
		ID:                   encounterspawn.ProfilePlayercentricAsteroidsV1,
		ScheduleKind:         encounterspawn.ScheduleContinuous,
		IntervalSeconds:      constants.AsteroidSpawnInterval,
		BatchSize:            constants.AsteroidSpawnBatchSize,
		Priority:             0,
		SharedWeightedBudget: 0,
		ProfileWeightedLimit: 0,
		RetryCap:             baselineEncounterSpawnRetryCap,
		InitiallyActive:      true,
	}
}

func newEncounterSpawnRuntime() (*encounterspawn.Runtime, error) {
	runtime := encounterspawn.NewRuntime()
	if err := runtime.Configure(baselineEncounterSpawnConfig()); err != nil {
		return nil, err
	}
	return runtime, nil
}

func (game *Game) encounterSpawn() *encounterspawn.Runtime {
	if game.encounterSpawnRuntime == nil {
		runtime, err := newEncounterSpawnRuntime()
		if err != nil {
			panic(fmt.Errorf("failed to initialize encounter spawn runtime: %w", err))
		}
		game.encounterSpawnRuntime = runtime
	}
	return game.encounterSpawnRuntime
}

func (game *Game) EncounterSpawnProfileSnapshot(profileID encounterspawn.ProfileID) (encounterspawn.Snapshot, bool) {
	game.mu.Lock()
	defer game.mu.Unlock()
	return game.encounterSpawn().Snapshot(profileID)
}

func (game *Game) sortedCameraViews() []*runtime.CameraView {
	playerIDs := make([]string, 0, len(game.cameraViews))
	for playerID, view := range game.cameraViews {
		if view != nil {
			playerIDs = append(playerIDs, playerID)
		}
	}
	sort.Strings(playerIDs)
	views := make([]*runtime.CameraView, 0, len(playerIDs))
	for _, playerID := range playerIDs {
		views = append(views, game.cameraViews[playerID])
	}
	return views
}

func (game *Game) canAdmitEncounterSpawn(profileID encounterspawn.ProfileID, spawnType string, cost encounterspawn.WeightedPopulation) bool {
	if cost <= 0 {
		return false
	}
	snapshot, ok := game.encounterSpawn().Snapshot(profileID)
	if !ok || snapshot.State != encounterspawn.StateActive || snapshot.RuntimeStopped {
		return false
	}

	profileTotals := game.encounterLifecycle().ProfileWeightedPopulationTotals()
	sharedTotal := encounterspawn.WeightedPopulation(0)
	for _, total := range profileTotals {
		sharedTotal += encounterspawn.WeightedPopulation(total)
	}
	if limit := snapshot.Config.SharedWeightedBudget; limit > 0 && sharedTotal+cost > limit {
		return false
	}

	lifecycleProfileID := encounterlifecycle.ProfileID(profileID)
	profileTotal := encounterspawn.WeightedPopulation(profileTotals[lifecycleProfileID])
	if limit := snapshot.Config.ProfileWeightedLimit; limit > 0 && profileTotal+cost > limit {
		return false
	}

	limit := snapshot.Config.SpawnTypeWeightedLimits[spawnType]
	if limit <= 0 {
		return true
	}
	spawnTypeTotal := encounterspawn.WeightedPopulation(0)
	for _, entityID := range game.encounterLifecycle().EntityIDs() {
		entry, exists := game.encounterLifecycle().Snapshot(entityID)
		if !exists || entry.Registration.Origin.ProfileID != lifecycleProfileID || string(entry.Registration.Origin.SpawnType) != spawnType {
			continue
		}
		spawnTypeTotal += encounterspawn.WeightedPopulation(entry.Registration.Origin.WeightedPopulationCost)
	}
	return spawnTypeTotal+cost <= limit
}
