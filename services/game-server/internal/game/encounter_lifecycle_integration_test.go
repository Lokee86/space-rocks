package game

import (
	"testing"

	"github.com/Lokee86/space-rocks/services/game-server/internal/constants"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/encounterlifecycle"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/physics"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/runtime"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/space"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/spawning"
)

func TestAsteroidLifecycleAdoptsLegacyAsteroidAndKeepsItWithoutViews(t *testing.T) {
	game := NewWithSeed(10)
	asteroid := runtime.NewAsteroid("legacy-asteroid", physics.Vector2{X: 7000}, physics.Vector2{}, 2, 0)
	game.entities.Asteroids[asteroid.ID] = asteroid

	game.stepAsteroids(0, space.DefaultBounds())

	if _, exists := game.entities.Asteroids[asteroid.ID]; !exists {
		t.Fatal("zero-view lifecycle evaluation removed asteroid")
	}
	entry, registered := game.encounterLifecycleRuntime.Snapshot(asteroid.ID)
	if !registered {
		t.Fatal("legacy asteroid was not adopted into lifecycle ownership")
	}
	if entry.Registration.Origin.ProfileID != baselineAsteroidProfileID {
		t.Fatalf("legacy asteroid profile = %q", entry.Registration.Origin.ProfileID)
	}
}

func TestAsteroidLifecycleDistanceRequiresEveryViewToBeFar(t *testing.T) {
	game := NewWithSeed(11)
	asteroid := game.applyAsteroidSpawn(spawning.AsteroidSpawnPlan{
		Position: physics.Vector2{X: 500},
		Size:     1,
	})
	game.cameraViews["far"] = &runtime.CameraView{
		X:      0,
		Config: runtime.ClientConfig{VisibleWorldWidth: 100, VisibleWorldHeight: 100},
	}
	game.cameraViews["near"] = &runtime.CameraView{
		X:      480,
		Config: runtime.ClientConfig{VisibleWorldWidth: 100, VisibleWorldHeight: 100},
	}

	game.stepAsteroids(0, space.DefaultBounds())
	if _, exists := game.entities.Asteroids[asteroid.ID]; !exists {
		t.Fatal("asteroid was removed while one relevant view remained near")
	}

	game.cameraViews["near"].X = 1000
	game.stepAsteroids(0, space.DefaultBounds())
	if _, exists := game.entities.Asteroids[asteroid.ID]; exists {
		t.Fatal("asteroid remained after every relevant view was far")
	}
}

func TestAsteroidLifecycleDistanceUsesWrappedWorldDistance(t *testing.T) {
	game := NewWithSeed(12)
	asteroid := game.applyAsteroidSpawn(spawning.AsteroidSpawnPlan{
		Position: physics.Vector2{X: 50},
		Size:     1,
	})
	game.cameraViews["wrapped-near"] = &runtime.CameraView{
		X: constants.WorldWidth - 50,
		Config: runtime.ClientConfig{
			VisibleWorldWidth:  100,
			VisibleWorldHeight: 100,
		},
	}

	game.stepAsteroids(0, space.DefaultBounds())
	if _, exists := game.entities.Asteroids[asteroid.ID]; !exists {
		t.Fatal("asteroid near a view across the world seam was removed")
	}
}

func TestWorldFreezePausesAsteroidLifecycleLifetime(t *testing.T) {
	game := NewWithSeed(13)
	asteroid := game.applyAsteroidSpawn(spawning.AsteroidSpawnPlan{Size: 1})
	game.encounterLifecycleRuntime.Remove(asteroid.ID)
	registration := baselineAsteroidLifecycleRegistration(asteroid)
	registration.Policy = encounterlifecycle.Policy{
		LifetimeExpiry: encounterlifecycle.TriggerPolicy{
			Enabled:     true,
			Disposition: encounterlifecycle.DispositionHardRemove,
		},
		LifetimeSeconds: 1,
	}
	if err := game.encounterLifecycleRuntime.Register(asteroid.ID, registration); err != nil {
		t.Fatal(err)
	}

	game.worldSimulationOptions.SetFreezeWorld(true)
	game.Step(2)
	entry, exists := game.encounterLifecycleRuntime.Snapshot(asteroid.ID)
	if !exists || entry.ElapsedLifetimeSeconds != 0 {
		t.Fatalf("frozen lifecycle entry = %+v, exists=%v", entry, exists)
	}
	if _, exists := game.entities.Asteroids[asteroid.ID]; !exists {
		t.Fatal("frozen lifecycle evaluation removed asteroid")
	}

	game.worldSimulationOptions.SetFreezeWorld(false)
	game.Step(1)
	if _, exists := game.entities.Asteroids[asteroid.ID]; exists {
		t.Fatal("lifetime-expired asteroid remained after simulation resumed")
	}
}

func TestTransitionResetHardRemovesAllAsteroidsThroughLifecycle(t *testing.T) {
	game := NewWithSeed(15)
	game.applyAsteroidSpawn(spawning.AsteroidSpawnPlan{Size: 1})
	game.applyAsteroidSpawn(spawning.AsteroidSpawnPlan{Size: 3})

	removed := game.removeAllAsteroidsForLifecycleTrigger(encounterlifecycle.TriggerTransitionReset)
	if removed != 2 {
		t.Fatalf("transition/reset removed %d asteroids, want 2", removed)
	}
	if len(game.entities.Asteroids) != 0 {
		t.Fatal("transition/reset left asteroids in the entity store")
	}
	if len(game.encounterLifecycleRuntime.EntityIDs()) != 0 {
		t.Fatal("transition/reset left lifecycle registrations")
	}
	if len(game.encounterLifecycleRuntime.ProfileWeightedPopulationTotals()) != 0 {
		t.Fatal("transition/reset left profile population accounting")
	}
}

func TestAsteroidSoftRetirementUsesDeferredCleanupHandoff(t *testing.T) {
	game := NewWithSeed(14)
	asteroid := game.applyAsteroidSpawn(spawning.AsteroidSpawnPlan{
		Position: physics.Vector2{X: 1000},
		Size:     2,
	})
	game.encounterLifecycleRuntime.Remove(asteroid.ID)
	registration := baselineAsteroidLifecycleRegistration(asteroid)
	registration.Policy.OutsideAllRelevantPlayers.Disposition = encounterlifecycle.DispositionSoftRetire
	registration.Capabilities = encounterlifecycle.EntityCapabilities{SupportsSoftRetire: true}
	if err := game.encounterLifecycleRuntime.Register(asteroid.ID, registration); err != nil {
		t.Fatal(err)
	}
	game.cameraViews["far"] = &runtime.CameraView{
		X:      0,
		Config: runtime.ClientConfig{VisibleWorldWidth: 100, VisibleWorldHeight: 100},
	}

	game.stepAsteroids(0, space.DefaultBounds())
	if _, exists := game.entities.Asteroids[asteroid.ID]; !exists {
		t.Fatal("soft retirement hard-removed asteroid")
	}
	if !asteroid.PendingDespawn {
		t.Fatal("soft retirement did not enter deferred cleanup")
	}
	entry, exists := game.encounterLifecycleRuntime.Snapshot(asteroid.ID)
	if !exists || entry.RetirementState != encounterlifecycle.RetirementStateBegun ||
		entry.Retirement.Disposition != encounterlifecycle.DispositionSoftRetire {
		t.Fatalf("unexpected soft retirement entry: %+v, exists=%v", entry, exists)
	}
	if game.encounterLifecycleRuntime.ProfileWeightedPopulationTotals()[baselineAsteroidProfileID] != 2 {
		t.Fatal("soft-retiring asteroid stopped counting before authoritative removal")
	}

	game.stepAsteroids(constants.CollisionDespawnDelay, space.DefaultBounds())
	if _, exists := game.entities.Asteroids[asteroid.ID]; exists {
		t.Fatal("soft-retired asteroid remained after deferred cleanup completed")
	}
}
