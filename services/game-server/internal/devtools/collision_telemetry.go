package devtools

import "github.com/Lokee86/space-rocks/services/game-server/internal/game/physics"

type CollisionPoint struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

type CollisionBody struct {
	Kind   string           `json:"kind"`
	ID     string           `json:"id"`
	Shape  string           `json:"shape"`
	Points []CollisionPoint `json:"points"`
}

func CollisionBodies(target CollisionTelemetryTarget) []CollisionBody {
	if target == nil {
		return nil
	}

	bodiesByKind := target.CollisionBodiesByKind()
	bodies := make([]CollisionBody, 0)
	for _, kind := range []string{"player", "asteroid", "bullet", "pickup"} {
		for _, body := range bodiesByKind[kind] {
			bodies = append(bodies, collisionBody(kind, body))
		}
	}
	return bodies
}

func collisionBody(kind string, body physics.CollisionBody) CollisionBody {
	points := physics.CollisionBodyOutlinePoints(body)
	telemetryPoints := make([]CollisionPoint, 0, len(points))
	for _, point := range points {
		telemetryPoints = append(telemetryPoints, CollisionPoint{X: point.X, Y: point.Y})
	}

	return CollisionBody{
		Kind:   kind,
		ID:     body.ID,
		Shape:  string(body.Shape.Type),
		Points: telemetryPoints,
	}
}
