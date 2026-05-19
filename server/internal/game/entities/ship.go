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
		Score:    ship.Score,
		Lives:    ship.Lives,
		Paused:   ship.Paused,
	}
}

func (ship *Ship) SetInput(input InputState) {
	ship.Input = input
}

func (ship *Ship) ClearInput() {
	ship.Input = InputState{}
}

func (ship *Ship) SetConfig(config ClientConfig) {
	ship.Config = config
}

func (ship *Ship) Pause() {
	ship.Paused = true
	ship.ClearInput()
}

func (ship *Ship) Resume(invulnerabilitySeconds float64) {
	ship.Paused = false
	ship.ClearInput()
	ship.InvulnerabilityRemaining = invulnerabilitySeconds
}

func (ship *Ship) ApplyInput(delta float64) {
	if ship.PendingDespawn {
		ship.DespawnDelay -= delta
		return
	}
	if ship.Paused {
		ship.ClearInput()
		return
	}

	ship.InvulnerabilityRemaining = max(0, ship.InvulnerabilityRemaining-delta)

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
	return !ship.Paused && !ship.IsInvulnerable() && ship.Input.Shoot
}

func (ship *Ship) CanShoot() bool {
	return !ship.Paused && !ship.IsInvulnerable() && ship.ShootCooldown == 0
}

func (ship *Ship) ResetShootCooldown() {
	ship.ShootCooldown = constants.BulletCooldown
}

func (ship *Ship) AddScore(score int) {
	ship.Score += score
}

func (ship *Ship) Position() physics.Vector2 {
	return physics.Vector2{X: ship.X, Y: ship.Y}
}

func (ship *Ship) IsPendingDespawn() bool {
	return ship.PendingDespawn
}

func (ship *Ship) IsInvulnerable() bool {
	return ship.InvulnerabilityRemaining > 0
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
