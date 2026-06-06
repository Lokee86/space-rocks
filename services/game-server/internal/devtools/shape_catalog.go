package devtools

import "github.com/Lokee86/space-rocks/server/internal/game/physics"

func BuildShapeCatalog(catalog physics.CollisionShapeCatalog) map[string]DebugShapeDefinition {
	shapes := make(map[string]DebugShapeDefinition)

	addShape := func(id string, kind string, imported physics.ImportedCollisionShape) {
		shape, err := imported.ToCollisionShape(1)
		if err != nil {
			return
		}

		body := physics.CollisionBody{Shape: shape}
		points := physics.CollisionBodyOutlinePoints(body)
		if len(points) == 0 {
			return
		}

		definition := DebugShapeDefinition{
			ID:        id,
			Kind:      kind,
			ShapeType: string(shape.Type),
			Points:    make([]DebugShapePoint, 0, len(points)),
		}
		for _, point := range points {
			definition.Points = append(definition.Points, DebugShapePoint{X: point.X, Y: point.Y})
		}

		shapes[id] = definition
	}

	addShape(PlayerShapeID("v_wing"), "player", catalog.Ship)
	addShape(BulletShapeID(), "bullet", catalog.Bullet)

	for index, imported := range catalog.Asteroids {
		addShape(AsteroidShapeID(index), "asteroid", imported)
	}

	for pickupType, imported := range catalog.Pickups {
		addShape(PickupShapeID(pickupType), "pickup", imported)
	}

	return shapes
}
