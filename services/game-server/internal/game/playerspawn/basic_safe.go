package playerspawn

import (
	"fmt"

	"github.com/Lokee86/space-rocks/services/game-server/internal/game/physics"
)

const basicSafeSearchSpacing = 160.0

func PlanBasicSafeSpawnV1(request Request, evaluator SafetyEvaluator) (Plan, error) {
	if request.ProfileID != BasicSafeSpawnProfileID {
		return Plan{}, fmt.Errorf("unsupported player spawn profile %q", request.ProfileID)
	}
	if evaluator == nil {
		return Plan{}, fmt.Errorf("player spawn safety evaluator is required")
	}
	if evaluator.IsSafeSpawn(request.PreferredOrigin, request.PlayerID, request.CollisionShapeID) {
		return Plan{Position: request.PreferredOrigin}, nil
	}

	for ring := 1; ; ring++ {
		for x := -ring; x <= ring; x++ {
			top := request.PreferredOrigin.Add(physics.Vector2{X: float64(x) * basicSafeSearchSpacing, Y: -float64(ring) * basicSafeSearchSpacing})
			if evaluator.IsSafeSpawn(top, request.PlayerID, request.CollisionShapeID) {
				return Plan{Position: top}, nil
			}
			bottom := request.PreferredOrigin.Add(physics.Vector2{X: float64(x) * basicSafeSearchSpacing, Y: float64(ring) * basicSafeSearchSpacing})
			if evaluator.IsSafeSpawn(bottom, request.PlayerID, request.CollisionShapeID) {
				return Plan{Position: bottom}, nil
			}
		}
		for y := -ring + 1; y <= ring-1; y++ {
			left := request.PreferredOrigin.Add(physics.Vector2{X: -float64(ring) * basicSafeSearchSpacing, Y: float64(y) * basicSafeSearchSpacing})
			if evaluator.IsSafeSpawn(left, request.PlayerID, request.CollisionShapeID) {
				return Plan{Position: left}, nil
			}
			right := request.PreferredOrigin.Add(physics.Vector2{X: float64(ring) * basicSafeSearchSpacing, Y: float64(y) * basicSafeSearchSpacing})
			if evaluator.IsSafeSpawn(right, request.PlayerID, request.CollisionShapeID) {
				return Plan{Position: right}, nil
			}
		}
	}
}
