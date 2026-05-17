package entities

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
	if ship.PendingDespawn {
		ship.DespawnDelay -= delta
		return
	}

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

func (ship *Ship) IsPendingDespawn() bool {
	return ship.PendingDespawn
}

func (ship *Ship) ReadyForRemoval() bool {
	return ship.PendingDespawn && ship.DespawnDelay <= 0
}

func (ship *Ship) MarkPendingDespawn(delay float64) {
	ship.PendingDespawn = true
	ship.DespawnDelay = delay
	ship.Velocity = physics.Vector2{}
	ship.Input = InputState{}
}

func (ship *Ship) Forward() physics.Vector2 {
	return physics.Vector2{X: 0, Y: -1}.Rotated(ship.Rotation)
}

func (ship *Ship) CollisionBody(catalog physics.CollisionShapeCatalog) (physics.CollisionBody, bool) {
	shape, err := catalog.ShipShape()
	if err != nil {
		return physics.CollisionBody{}, false
	}

	return physics.CollisionBody{
		ID:       ship.ID,
		Position: ship.Position(),
		Rotation: ship.Rotation,
		Shape:    shape,
	}, true
}

func (ship *Ship) IsInsideView(position physics.Vector2) bool {
	width := ship.VisibleWorldWidth()
	height := ship.VisibleWorldHeight()
	left := ship.X - width*0.5
	right := ship.X + width*0.5
	top := ship.Y - height*0.5
	bottom := ship.Y + height*0.5

	return position.X >= left &&
		position.X <= right &&
		position.Y >= top &&
		position.Y <= bottom
}

func (ship *Ship) IsFarFromView(position physics.Vector2) bool {
	width := ship.VisibleWorldWidth()
	height := ship.VisibleWorldHeight()
	left := ship.X - width*0.5 - constants.AsteroidDespawnMargin
	right := ship.X + width*0.5 + constants.AsteroidDespawnMargin
	top := ship.Y - height*0.5 - constants.AsteroidDespawnMargin
	bottom := ship.Y + height*0.5 + constants.AsteroidDespawnMargin

	return position.X < left ||
		position.X > right ||
		position.Y < top ||
		position.Y > bottom
}

func (ship *Ship) VisibleWorldWidth() float64 {
	if ship.Config.VisibleWorldWidth > 0 {
		return ship.Config.VisibleWorldWidth
	}

	return constants.WorldWidth
}

func (ship *Ship) VisibleWorldHeight() float64 {
	if ship.Config.VisibleWorldHeight > 0 {
		return ship.Config.VisibleWorldHeight
	}

	return constants.WorldHeight
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
