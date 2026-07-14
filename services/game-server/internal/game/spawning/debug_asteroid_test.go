package spawning

import (
	"math"
	"testing"

	"github.com/Lokee86/space-rocks/services/game-server/internal/game/physics"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/rng"
)

func TestPlanDebugAsteroidSpawnNormalizesExplicitDirection(t *testing.T) {
	spawner := New(rng.New(12345))
	position := physics.Vector2{X: 120, Y: 340}
	requestedDirection := physics.Vector2{X: 3, Y: 4}

	plan := spawner.PlanDebugAsteroidSpawn(position, requestedDirection, true)

	assertVectorApproxEqual(t, plan.Position, position)
	assertVectorApproxEqual(t, plan.Velocity.Normalized(), requestedDirection.Normalized())
}

func TestPlanDebugAsteroidSpawnRetainsDebugIdentityAndPosition(t *testing.T) {
	spawner := New(rng.New(54321))
	position := physics.Vector2{X: -42, Y: 88}

	plan := spawner.PlanDebugAsteroidSpawn(position, physics.Vector2{X: 1, Y: 0}, true)

	assertVectorApproxEqual(t, plan.Position, position)
	if plan.EntityType != SpawnEntityTypeAsteroid {
		t.Fatalf("EntityType = %q, want %q", plan.EntityType, SpawnEntityTypeAsteroid)
	}
	if plan.Reason != SpawnReasonDebugAsteroid {
		t.Fatalf("Reason = %q, want %q", plan.Reason, SpawnReasonDebugAsteroid)
	}
}

func TestPlanDebugAsteroidSpawnIsSeedDeterministicForSizeAndVariant(t *testing.T) {
	const seed int64 = 777777
	spawnerA := New(rng.New(seed))
	spawnerB := New(rng.New(seed))

	positions := []physics.Vector2{
		{X: 1, Y: 2},
		{X: 3, Y: 4},
		{X: 5, Y: 6},
		{X: 7, Y: 8},
		{X: 9, Y: 10},
	}
	requestedDirection := physics.Vector2{X: 3, Y: 4}

	for i, position := range positions {
		planA := spawnerA.PlanDebugAsteroidSpawn(position, requestedDirection, true)
		planB := spawnerB.PlanDebugAsteroidSpawn(position, requestedDirection, true)

		if planA.Size != planB.Size {
			t.Fatalf("call %d: Size = %d vs %d", i, planA.Size, planB.Size)
		}
		if planA.Variant != planB.Variant {
			t.Fatalf("call %d: Variant = %d vs %d", i, planA.Variant, planB.Variant)
		}
		if planA.Size < 1 || planA.Size > 4 {
			t.Fatalf("call %d: Size = %d, want 1..4", i, planA.Size)
		}
	}
}

func assertVectorApproxEqual(t *testing.T, got, want physics.Vector2) {
	t.Helper()
	if !floatApproxEqual(got.X, want.X) || !floatApproxEqual(got.Y, want.Y) {
		t.Fatalf("vector = %+v, want %+v", got, want)
	}
}

func floatApproxEqual(got, want float64) bool {
	const epsilon = 1e-9
	return math.Abs(got-want) <= epsilon
}
