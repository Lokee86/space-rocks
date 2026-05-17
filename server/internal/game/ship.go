package game

import (
	"math"

	"github.com/Lokee86/space-rocks/server/internal/constants"
	"github.com/Lokee86/space-rocks/server/internal/game/physics"
)

func (ship *Ship) State() ShipState {
	return ShipState{
		ID:       ship.ID,
		X:        ship.X,
		Y:        ship.Y,
		Rotation: ship.Rotation,
	}
}

func (ship *Ship) SetInput(input InputState) {
	ship.Input = input
}

func (ship *Ship) SetConfig(config ClientConfig) {
	ship.Config = config
}

func (ship *Ship) ApplyInput(delta float64) {
	ship.ShootCooldown = max(0, ship.ShootCooldown-delta)

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
	ship.Velocity = ship.Velocity.LimitLength(constants.PlayerMaxSpeed)

	ship.X += ship.Velocity.X * delta
	ship.Y += ship.Velocity.Y * delta
}

func (ship *Ship) WantsToShoot() bool {
	return ship.Input.Shoot
}

func (ship *Ship) CanShoot() bool {
	return ship.ShootCooldown == 0
}

func (ship *Ship) ResetShootCooldown() {
	ship.ShootCooldown = constants.BulletCooldown
}

func (ship *Ship) Position() physics.Vector2 {
	return physics.Vector2{X: ship.X, Y: ship.Y}
}

func (ship *Ship) Forward() physics.Vector2 {
	return physics.Vector2{X: 0, Y: -1}.Rotated(ship.Rotation)
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
