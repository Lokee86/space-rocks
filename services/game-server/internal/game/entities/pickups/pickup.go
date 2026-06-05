package pickups

import "github.com/Lokee86/space-rocks/server/internal/game/physics"

type Pickup struct {
	ID     string
	Type   PickupType
	X      float64
	Y      float64
	Radius float64
}

func (pickup *Pickup) Position() physics.Vector2 {
	return physics.Vector2{X: pickup.X, Y: pickup.Y}
}

func (pickup *Pickup) CollisionBody() physics.CollisionBody {
	return physics.CollisionBody{
		ID:       pickup.ID,
		Position: pickup.Position(),
		Shape: physics.CollisionShape{
			Type:   "circle",
			Radius: pickup.Radius,
		},
	}
}
