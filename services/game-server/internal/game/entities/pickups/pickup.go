package pickups

import "github.com/Lokee86/space-rocks/server/internal/game/physics"

type Pickup struct {
	ID              string
	Type            PickupType
	X               float64
	Y               float64
	Health          int
	AgeSeconds      float64
	LifespanSeconds float64
}

func (pickup *Pickup) Position() physics.Vector2 {
	return physics.Vector2{X: pickup.X, Y: pickup.Y}
}

func (pickup *Pickup) Class() PickupClass {
	definition, ok := DefinitionFor(pickup.Type)
	if !ok {
		return ""
	}

	return definition.Class
}

func (pickup *Pickup) CollisionBody(catalog physics.CollisionShapeCatalog) (physics.CollisionBody, bool) {
	pickupClass := pickup.Class()
	if pickupClass == "" {
		return physics.CollisionBody{}, false
	}

	shape, err := catalog.PickupShape(string(pickupClass))
	if err != nil {
		return physics.CollisionBody{}, false
	}

	return physics.CollisionBody{
		ID:       pickup.ID,
		Position: pickup.Position(),
		Shape:    shape,
	}, true
}
