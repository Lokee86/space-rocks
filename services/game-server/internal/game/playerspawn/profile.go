package playerspawn

import "github.com/Lokee86/space-rocks/services/game-server/internal/game/physics"

const BasicSafeSpawnProfileID = "basic_safe_spawn_v1"

type Request struct {
	ProfileID        string
	PlayerID         string
	SpawnReason      string
	PreferredOrigin  physics.Vector2
	CollisionShapeID string
}

type Plan struct {
	Position physics.Vector2
}

type SafetyEvaluator interface {
	IsSafeSpawn(position physics.Vector2, playerID string, collisionShapeID string) bool
}
