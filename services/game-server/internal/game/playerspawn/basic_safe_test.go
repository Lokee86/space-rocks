package playerspawn

import (
	"testing"

	"github.com/Lokee86/space-rocks/services/game-server/internal/game/physics"
)

type testEvaluator struct {
	unsafe map[physics.Vector2]bool
}

func (evaluator testEvaluator) IsSafeSpawn(position physics.Vector2, _ string, _ string) bool {
	return !evaluator.unsafe[position]
}

func TestBasicSafeSpawnAcceptsOrigin(t *testing.T) {
	origin := physics.Vector2{X: 100, Y: 200}
	plan, err := PlanBasicSafeSpawnV1(Request{ProfileID: BasicSafeSpawnProfileID, PreferredOrigin: origin}, testEvaluator{})
	if err != nil || plan.Position != origin {
		t.Fatalf("plan = %+v, err = %v", plan, err)
	}
}

func TestBasicSafeSpawnUsesDeterministicOutwardFallback(t *testing.T) {
	origin := physics.Vector2{X: 100, Y: 200}
	plan, err := PlanBasicSafeSpawnV1(Request{
		ProfileID:       BasicSafeSpawnProfileID,
		PreferredOrigin: origin,
	}, testEvaluator{unsafe: map[physics.Vector2]bool{
		origin:           true,
		{X: -60, Y: 40}:  true,
		{X: -60, Y: 360}: true,
		{X: 100, Y: 40}:  true,
		{X: 100, Y: 360}: true,
		{X: 260, Y: 40}:  true,
		{X: 260, Y: 360}: true,
		{X: -60, Y: 200}: true,
	}})
	if err != nil {
		t.Fatal(err)
	}
	if plan.Position != (physics.Vector2{X: 260, Y: 200}) {
		t.Fatalf("fallback position = %+v, want (260, 200)", plan.Position)
	}
}
