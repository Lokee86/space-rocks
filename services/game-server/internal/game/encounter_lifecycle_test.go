package game

import (
	"testing"

	"github.com/Lokee86/space-rocks/services/game-server/internal/constants"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/encounterlifecycle"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/physics"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/space"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/spawning"
)

func TestApplyAsteroidSpawnRegistersBaselineLifecycleMetadata(t *testing.T) {
	game := NewWithSeed(1)
	asteroid := game.applyAsteroidSpawn(spawning.AsteroidSpawnPlan{
		Position: physics.Vector2{X: 10, Y: 20},
		Size:     3,
	})

	entry, ok := game.encounterLifecycleRuntime.Snapshot(asteroid.ID)
	if !ok {
		t.Fatalf("asteroid %q was not registered", asteroid.ID)
	}
	origin := entry.Registration.Origin
	if origin.ProfileID != baselineAsteroidProfileID ||
		origin.SpawnType != baselineAsteroidSpawnType ||
		origin.LifecyclePolicyID != baselineAsteroidLifecyclePolicyID ||
		origin.Priority != baselineAsteroidPriority ||
		origin.WeightedPopulationCost != 3 {
		t.Fatalf("unexpected asteroid origin metadata: %+v", origin)
	}
	policy := entry.Registration.Policy
	if !policy.OutsideAllRelevantPlayers.Enabled ||
		policy.OutsideAllRelevantPlayers.Disposition != encounterlifecycle.DispositionHardRemove ||
		policy.ExtraPlayerDistance != constants.AsteroidDespawnMargin {
		t.Fatalf("unexpected asteroid lifecycle policy: %+v", policy)
	}
	if !entry.Registration.Capabilities.SupportsHardRemove {
		t.Fatal("baseline asteroid registration lacks hard-remove capability")
	}
}

func TestAsteroidSpawnAndRemovalUpdateProfileAccounting(t *testing.T) {
	game := NewWithSeed(2)
	asteroid := game.applyAsteroidSpawn(spawning.AsteroidSpawnPlan{Size: 4})
	totals := game.encounterLifecycleRuntime.ProfileWeightedPopulationTotals()
	if totals[baselineAsteroidProfileID] != 4 {
		t.Fatalf("initial asteroid population total = %v, want 4", totals[baselineAsteroidProfileID])
	}

	if !game.removeAsteroidAuthoritatively(asteroid.ID) {
		t.Fatal("authoritative asteroid removal failed")
	}
	if _, exists := game.entities.Asteroids[asteroid.ID]; exists {
		t.Fatal("removed asteroid remained in entity store")
	}
	if _, exists := game.encounterLifecycleRuntime.Snapshot(asteroid.ID); exists {
		t.Fatal("removed asteroid remained in lifecycle runtime")
	}
	if _, exists := game.encounterLifecycleRuntime.ProfileWeightedPopulationTotals()[baselineAsteroidProfileID]; exists {
		t.Fatal("baseline asteroid accounting remained after removal")
	}
	if game.removeAsteroidAuthoritatively(asteroid.ID) {
		t.Fatal("removing missing asteroid unexpectedly succeeded")
	}
}

func TestReadyAsteroidRemovalUsesAuthoritativeLifecycleRemoval(t *testing.T) {
	game := NewWithSeed(3)
	asteroid := game.applyAsteroidSpawn(spawning.AsteroidSpawnPlan{Size: 1})
	asteroid.MarkPendingDespawn(0)

	game.stepAsteroids(0, space.DefaultBounds())

	if _, exists := game.entities.Asteroids[asteroid.ID]; exists {
		t.Fatal("ready asteroid remained in entity store")
	}
	if _, exists := game.encounterLifecycleRuntime.Snapshot(asteroid.ID); exists {
		t.Fatal("ready asteroid remained in lifecycle runtime")
	}
}
