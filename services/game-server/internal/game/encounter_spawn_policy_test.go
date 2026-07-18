package game

import (
	"reflect"
	"testing"

	"github.com/Lokee86/space-rocks/services/game-server/internal/constants"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/encounterlifecycle"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/encounterspawn"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/physics"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/runtime"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/spawning"
)

func setEncounterSpawnTestConfig(t *testing.T, game *Game, config encounterspawn.Config) {
	t.Helper()
	runtime := encounterspawn.NewRuntime()
	if err := runtime.Configure(config); err != nil {
		t.Fatal(err)
	}
	game.encounterSpawnRuntime = runtime
}

func TestEncounterSpawnAdmissionUsesWeightedLimits(t *testing.T) {
	game := NewWithSeed(6)
	config := baselineEncounterSpawnConfig()
	config.SharedWeightedBudget = 5
	config.ProfileWeightedLimit = 4
	config.SpawnTypeWeightedLimits = map[string]encounterspawn.WeightedPopulation{"asteroid": 3}
	setEncounterSpawnTestConfig(t, game, config)

	if !game.canAdmitEncounterSpawn(config.ID, "asteroid", 3) {
		t.Fatal("empty profile rejected an admissible weighted spawn")
	}
	game.applyAsteroidSpawnForProfile(spawning.AsteroidSpawnPlan{Size: 3}, config.ID)
	if game.canAdmitEncounterSpawn(config.ID, "asteroid", 1) {
		t.Fatal("spawn-type weighted limit admitted excess population")
	}
	if game.canAdmitEncounterSpawn(config.ID, "enemy", 2) {
		t.Fatal("profile weighted limit admitted excess population")
	}
}

func TestEncounterSpawnAdmissionUsesSharedPopulationAcrossProfiles(t *testing.T) {
	game := NewWithSeed(7)
	config := baselineEncounterSpawnConfig()
	config.SharedWeightedBudget = 4
	setEncounterSpawnTestConfig(t, game, config)
	game.applyAsteroidSpawnForProfile(spawning.AsteroidSpawnPlan{Size: 3}, encounterspawn.ProfileID("other_profile"))

	if game.canAdmitEncounterSpawn(config.ID, "asteroid", 2) {
		t.Fatal("shared weighted budget ignored another profile population")
	}
}

func TestEncounterSpawnSafetyRetriesAreBounded(t *testing.T) {
	game := NewWithSeed(8)
	view := &runtime.CameraView{
		Config: runtime.ClientConfig{
			VisibleWorldWidth:  constants.WorldWidth * 2,
			VisibleWorldHeight: constants.WorldHeight * 2,
		},
	}
	game.cameraViews["player-1"] = view

	spawned := game.spawnAsteroidBatchForProfile(view, encounterspawn.ProfilePlayercentricAsteroidsV1, 3, 0)
	if spawned != 0 || len(game.entities.Asteroids) != 0 {
		t.Fatalf("unsafe bounded spawn produced %d asteroids", spawned)
	}
}

func TestEncounterSpawnTargetsAreDeterministicAcrossCameraInsertionOrder(t *testing.T) {
	left := NewWithSeed(9)
	right := NewWithSeed(9)
	addEncounterSpawnTestCamera(left, "b", 1000)
	addEncounterSpawnTestCamera(left, "a", 100)
	addEncounterSpawnTestCamera(right, "a", 100)
	addEncounterSpawnTestCamera(right, "b", 1000)

	left.stepAsteroidSpawning(constants.AsteroidSpawnInterval)
	right.stepAsteroidSpawning(constants.AsteroidSpawnInterval)
	if !reflect.DeepEqual(left.entities.Asteroids, right.entities.Asteroids) {
		t.Fatalf("camera insertion order changed deterministic spawn output:\nleft=%+v\nright=%+v", left.entities.Asteroids, right.entities.Asteroids)
	}
}

func TestAsteroidFragmentsRetainOriginatingEncounterProfile(t *testing.T) {
	game := NewWithSeed(10)
	profileID := encounterspawn.ProfileID("custom_profile")
	parent := game.applyAsteroidSpawnForProfile(spawning.AsteroidSpawnPlan{
		Position: physics.Vector2{X: 10, Y: 20},
		Size:     2,
	}, profileID)
	game.spawnAsteroidFragments(parent)

	if len(game.entities.Asteroids) != 3 {
		t.Fatalf("asteroid count = %d, want parent plus two fragments", len(game.entities.Asteroids))
	}
	for entityID := range game.entities.Asteroids {
		if entityID == parent.ID {
			continue
		}
		entry, ok := game.encounterLifecycleRuntime.Snapshot(entityID)
		if !ok || entry.Registration.Origin.ProfileID != encounterlifecycle.ProfileID(profileID) {
			t.Fatalf("fragment %q lost profile origin: %+v", entityID, entry)
		}
	}
}
