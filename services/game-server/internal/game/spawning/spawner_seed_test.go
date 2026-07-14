package spawning

import (
	"testing"

	"github.com/Lokee86/space-rocks/services/game-server/internal/game/physics"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/rng"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/runtime"
)

func TestSpawnerSeededTimedSpawnMatchesForEqualSeeds(t *testing.T) {
	left := New(rng.New(12345))
	right := New(rng.New(12345))
	position := physics.Vector2{X: 12, Y: -8}
	target := physics.Vector2{X: -3, Y: 17}

	leftPlan := left.PlanTimedAsteroidSpawn(position, target)
	rightPlan := right.PlanTimedAsteroidSpawn(position, target)

	if leftPlan.Position != rightPlan.Position {
		t.Fatalf("timed spawn position mismatch: got %+v, want %+v", leftPlan.Position, rightPlan.Position)
	}
	if leftPlan.Size != rightPlan.Size {
		t.Fatalf("timed spawn size mismatch: got %d, want %d", leftPlan.Size, rightPlan.Size)
	}
	if leftPlan.Velocity != rightPlan.Velocity {
		t.Fatalf("timed spawn velocity mismatch: got %+v, want %+v", leftPlan.Velocity, rightPlan.Velocity)
	}
}

func TestSpawnerSeededFragmentsMatchForEqualSeeds(t *testing.T) {
	left := New(rng.New(777))
	right := New(rng.New(777))
	asteroid := runtime.NewAsteroid("asteroid-1", physics.Vector2{X: 90, Y: 45}, physics.Vector2{}, 3, 9)

	leftPlans := left.PlanAsteroidFragmentSpawns(asteroid)
	rightPlans := right.PlanAsteroidFragmentSpawns(asteroid)

	if len(leftPlans) != len(rightPlans) {
		t.Fatalf("fragment spawn count mismatch: got %d, want %d", len(leftPlans), len(rightPlans))
	}
	for i := range leftPlans {
		if leftPlans[i].Position != rightPlans[i].Position {
			t.Fatalf("fragment position[%d] mismatch: got %+v, want %+v", i, leftPlans[i].Position, rightPlans[i].Position)
		}
		if leftPlans[i].Size != rightPlans[i].Size {
			t.Fatalf("fragment size[%d] mismatch: got %d, want %d", i, leftPlans[i].Size, rightPlans[i].Size)
		}
		if leftPlans[i].Velocity != rightPlans[i].Velocity {
			t.Fatalf("fragment velocity[%d] mismatch: got %+v, want %+v", i, leftPlans[i].Velocity, rightPlans[i].Velocity)
		}
	}
}

func TestSpawnerSeededPublicRandomMethodsMatchForEqualSeeds(t *testing.T) {
	left := New(rng.New(2468))
	right := New(rng.New(2468))

	for i := 0; i < 3; i++ {
		if got, want := left.RandomAsteroidSpeed(), right.RandomAsteroidSpeed(); got != want {
			t.Fatalf("RandomAsteroidSpeed sequence mismatch at step %d: got %v, want %v", i, got, want)
		}
	}
	for i := 0; i < 3; i++ {
		if got, want := left.RandomUnitVector(), right.RandomUnitVector(); got != want {
			t.Fatalf("RandomUnitVector sequence mismatch at step %d: got %+v, want %+v", i, got, want)
		}
	}
}

func TestSpawnerSeededTimedSpawnChangesForDifferentSeeds(t *testing.T) {
	left := New(rng.New(13579))
	right := New(rng.New(97531))
	position := physics.Vector2{X: 12, Y: -8}
	target := physics.Vector2{X: -3, Y: 17}

	leftPlan := left.PlanTimedAsteroidSpawn(position, target)
	rightPlan := right.PlanTimedAsteroidSpawn(position, target)

	if leftPlan.Size == rightPlan.Size && leftPlan.Velocity == rightPlan.Velocity {
		t.Fatal("expected timed-spawn owned randomness to differ for different seeds")
	}
}
