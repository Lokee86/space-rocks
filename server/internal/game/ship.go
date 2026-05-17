package game

import (
	"math"

	"github.com/Lokee86/space-rocks/server/internal/constants"
)

func (ship *Ship) State() ShipState {
	return ShipState{
		ID:       ship.ID,
		X:        ship.X,
		Y:        ship.Y,
		Rotation: ship.Rotation,
	}
}

func (ship *Ship) applyInput(delta float64) {
	rotationInput := axis(ship.Input.Left, ship.Input.Right)
	thrustInput := axis(ship.Input.Back, ship.Input.Forward)

	ship.Rotation += rotationInput * constants.PlayerRotationSpeed * delta

	if thrustInput != 0 {
		ship.Velocity.X += math.Sin(ship.Rotation) * constants.PlayerThrustForce * thrustInput * delta
		ship.Velocity.Y += -math.Cos(ship.Rotation) * constants.PlayerThrustForce * thrustInput * delta
	}

	damping := math.Pow(constants.PlayerDamping, delta/(1.0/60.0))
	ship.Velocity.X *= damping
	ship.Velocity.Y *= damping
	ship.Velocity = ship.Velocity.limitLength(constants.PlayerMaxSpeed)

	ship.X += ship.Velocity.X * delta
	ship.Y += ship.Velocity.Y * delta
}

func axis(negative bool, positive bool) float64 {
	var value float64
	if negative {
		value -= 1
	}
	if positive {
		value += 1
	}

	return max(-1, min(value, 1))
}
