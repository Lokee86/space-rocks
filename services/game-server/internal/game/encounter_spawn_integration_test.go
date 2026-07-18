package game

import (
	"testing"

	"github.com/Lokee86/space-rocks/services/game-server/internal/constants"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/encounterspawn"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/runtime"
)

func addEncounterSpawnTestCamera(game *Game, playerID string, x float64) {
	game.cameraViews[playerID] = &runtime.CameraView{
		X: x,
		Config: runtime.ClientConfig{
			VisibleWorldWidth:  640,
			VisibleWorldHeight: 360,
		},
	}
}

func TestBaselineEncounterSpawnProfileIsConfiguredAndActive(t *testing.T) {
	game := NewWithSeed(1)
	snapshot, ok := game.encounterSpawnRuntime.Snapshot(encounterspawn.ProfilePlayercentricAsteroidsV1)
	if !ok {
		t.Fatal("baseline encounter spawn profile was not configured")
	}
	if snapshot.State != encounterspawn.StateActive || snapshot.RuntimeStopped {
		t.Fatalf("unexpected baseline profile state: %+v", snapshot)
	}
	config := snapshot.Config
	if config.ScheduleKind != encounterspawn.ScheduleContinuous ||
		config.IntervalSeconds != constants.AsteroidSpawnInterval ||
		config.BatchSize != constants.AsteroidSpawnBatchSize ||
		config.RetryCap != baselineEncounterSpawnRetryCap {
		t.Fatalf("unexpected baseline profile config: %+v", config)
	}
}

func TestEncounterSpawnProfileOwnsBaselineTimerAndRemainder(t *testing.T) {
	game := NewWithSeed(2)
	addEncounterSpawnTestCamera(game, "player-1", 100)

	game.stepAsteroidSpawning(constants.AsteroidSpawnInterval + 0.5)
	if got := len(game.entities.Asteroids); got != constants.AsteroidSpawnBatchSize {
		t.Fatalf("first scheduled batch count = %d, want %d", got, constants.AsteroidSpawnBatchSize)
	}
	game.stepAsteroidSpawning(constants.AsteroidSpawnInterval - 0.5)
	if got := len(game.entities.Asteroids); got != constants.AsteroidSpawnBatchSize*2 {
		t.Fatalf("remainder-scheduled asteroid count = %d, want %d", got, constants.AsteroidSpawnBatchSize*2)
	}
	for entityID := range game.entities.Asteroids {
		entry, ok := game.encounterLifecycleRuntime.Snapshot(entityID)
		if !ok || entry.Registration.Origin.ProfileID != baselineAsteroidProfileID {
			t.Fatalf("asteroid %q missing baseline profile origin: %+v", entityID, entry)
		}
	}
}

func TestEncounterSpawnProfilePausesWithSpawnFreeze(t *testing.T) {
	game := NewWithSeed(3)
	addEncounterSpawnTestCamera(game, "player-1", 100)
	game.worldSimulationOptions.FreezeSpawning = true

	game.stepAsteroidSpawning(constants.AsteroidSpawnInterval)
	if len(game.entities.Asteroids) != 0 {
		t.Fatal("spawn freeze produced asteroids")
	}
	snapshot, _ := game.encounterSpawnRuntime.Snapshot(encounterspawn.ProfilePlayercentricAsteroidsV1)
	if snapshot.ElapsedSeconds != 0 {
		t.Fatalf("spawn freeze advanced profile timer to %v", snapshot.ElapsedSeconds)
	}

	game.worldSimulationOptions.FreezeSpawning = false
	game.stepAsteroidSpawning(constants.AsteroidSpawnInterval)
	if got := len(game.entities.Asteroids); got != constants.AsteroidSpawnBatchSize {
		t.Fatalf("unfrozen asteroid count = %d, want %d", got, constants.AsteroidSpawnBatchSize)
	}
}

func TestEncounterSpawnProfileDeactivationPreservesProgressAndEntities(t *testing.T) {
	game := NewWithSeed(4)
	addEncounterSpawnTestCamera(game, "player-1", 100)
	game.stepAsteroidSpawning(1)
	if err := game.encounterSpawnRuntime.Deactivate(encounterspawn.ProfilePlayercentricAsteroidsV1); err != nil {
		t.Fatal(err)
	}
	game.stepAsteroidSpawning(constants.AsteroidSpawnInterval)
	if len(game.entities.Asteroids) != 0 {
		t.Fatal("deactivated profile spawned asteroids")
	}
	if err := game.encounterSpawnRuntime.Activate(encounterspawn.ProfilePlayercentricAsteroidsV1); err != nil {
		t.Fatal(err)
	}
	game.stepAsteroidSpawning(2)
	if got := len(game.entities.Asteroids); got != constants.AsteroidSpawnBatchSize {
		t.Fatalf("reactivated profile count = %d, want %d", got, constants.AsteroidSpawnBatchSize)
	}
	if err := game.encounterSpawnRuntime.Deactivate(encounterspawn.ProfilePlayercentricAsteroidsV1); err != nil {
		t.Fatal(err)
	}
	game.stepAsteroidSpawning(constants.AsteroidSpawnInterval)
	if got := len(game.entities.Asteroids); got != constants.AsteroidSpawnBatchSize {
		t.Fatalf("deactivation removed existing asteroids, count = %d", got)
	}
}

func TestEncounterSpawnProfileResetsBaselineProgressWithoutTargets(t *testing.T) {
	game := NewWithSeed(5)
	addEncounterSpawnTestCamera(game, "player-1", 100)
	game.stepAsteroidSpawning(1)
	delete(game.cameraViews, "player-1")
	game.stepAsteroidSpawning(constants.AsteroidSpawnInterval)
	addEncounterSpawnTestCamera(game, "player-1", 100)
	game.stepAsteroidSpawning(2)
	if len(game.entities.Asteroids) != 0 {
		t.Fatal("targetless profile progress was not reset")
	}
	game.stepAsteroidSpawning(1)
	if got := len(game.entities.Asteroids); got != constants.AsteroidSpawnBatchSize {
		t.Fatalf("post-reset asteroid count = %d, want %d", got, constants.AsteroidSpawnBatchSize)
	}
}

func TestMatchOverStopsEncounterSpawnRuntime(t *testing.T) {
	game := newMatchOverTestGame()
	addEncounterSpawnTestCamera(game, "player-1", 100)
	game.Step(constants.AsteroidSpawnInterval)
	snapshot, _ := game.encounterSpawnRuntime.Snapshot(encounterspawn.ProfilePlayercentricAsteroidsV1)
	if !snapshot.RuntimeStopped {
		t.Fatal("match over did not stop encounter spawn runtime")
	}
}
