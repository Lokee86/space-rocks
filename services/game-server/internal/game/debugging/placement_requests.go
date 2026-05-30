package debugging

import "github.com/Lokee86/space-rocks/server/internal/game/physics"

const (
	EntityTypePlayer   = "player"
	EntityTypeAsteroid = "asteroid"
	EntityTypeBullet   = "bullet"
)

type SpawnEntityRequest struct {
	EntityType     string
	X              float64
	Y              float64
	HasDirection   bool
	DirectionX     float64
	DirectionY     float64
	TargetPlayerID string
}

type RespawnPlayerRequest struct {
	TargetPlayerID string
	X              float64
	Y              float64
}

func (request SpawnEntityRequest) Position() physics.Vector2 {
	return physics.Vector2{X: request.X, Y: request.Y}
}

func (request SpawnEntityRequest) DirectionOr(fallback physics.Vector2) physics.Vector2 {
	if !request.HasDirection {
		return fallback.Normalized()
	}

	requestedDirection := physics.Vector2{X: request.DirectionX, Y: request.DirectionY}
	if requestedDirection.Length() == 0 {
		return fallback.Normalized()
	}
	return requestedDirection.Normalized()
}
