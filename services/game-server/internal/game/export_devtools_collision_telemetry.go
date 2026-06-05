package game

import (
	"github.com/Lokee86/space-rocks/server/internal/game/physics"
)

type DevtoolsCollisionPoint struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

type DevtoolsCollisionBody struct {
	Kind   string                  `json:"kind"`
	ID     string                  `json:"id"`
	Shape  string                  `json:"shape"`
	Points []DevtoolsCollisionPoint `json:"points"`
}

func (game *Game) DevtoolsCollisionBodies() []DevtoolsCollisionBody {
	game.mu.Lock()
	defer game.mu.Unlock()

	bodies := make([]DevtoolsCollisionBody, 0, len(game.entities.Players)+len(game.entities.Asteroids)+len(game.entities.Projectiles)+len(game.entities.Pickups))

	for _, player := range game.entities.Players {
		body, ok := player.CollisionBody(game.collisionShapes)
		if !ok {
			continue
		}
		bodies = append(bodies, devtoolsCollisionBody("player", body))
	}
	for _, asteroid := range game.entities.Asteroids {
		body, ok := asteroid.CollisionBody(game.collisionShapes)
		if !ok {
			continue
		}
		bodies = append(bodies, devtoolsCollisionBody("asteroid", body))
	}
	for _, bullet := range game.entities.Projectiles {
		body, ok := bullet.CollisionBody(game.collisionShapes)
		if !ok {
			continue
		}
		bodies = append(bodies, devtoolsCollisionBody("bullet", body))
	}
	for _, pickup := range game.entities.Pickups {
		if pickup == nil {
			continue
		}
		body := pickup.CollisionBody()
		bodies = append(bodies, devtoolsCollisionBody("pickup", body))
	}

	return bodies
}

func devtoolsCollisionBody(kind string, body physics.CollisionBody) DevtoolsCollisionBody {
	points := physics.CollisionBodyOutlinePoints(body)
	telemetryPoints := make([]DevtoolsCollisionPoint, 0, len(points))
	for _, point := range points {
		telemetryPoints = append(telemetryPoints, DevtoolsCollisionPoint{X: point.X, Y: point.Y})
	}

	return DevtoolsCollisionBody{
		Kind:   kind,
		ID:     body.ID,
		Shape:  string(body.Shape.Type),
		Points: telemetryPoints,
	}
}
