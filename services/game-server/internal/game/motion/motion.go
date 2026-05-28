package motion

import (
	"math"

	"github.com/Lokee86/space-rocks/server/internal/game/entities"
	"github.com/Lokee86/space-rocks/server/internal/game/physics"
	"github.com/Lokee86/space-rocks/server/internal/game/space"
)

func StepShip(ship *entities.Ship, delta float64) {
	StepShipWithMovePolicy(ship, delta, true)
}

func StepShipWithMovePolicy(ship *entities.Ship, delta float64, canMove bool) {
	if ship.PendingDespawn {
		ship.DespawnDelay -= delta
		return
	}
	if !canMove {
		ship.ClearInput()
		return
	}

	ship.InvulnerabilityRemaining = max(0, ship.InvulnerabilityRemaining-delta)

	ship.ShootCooldown = max(0, ship.ShootCooldown-delta)

	rotationInput := axis(ship.Input.Left, ship.Input.Right)
	thrustInput := axis(ship.Input.Back, ship.Input.Forward)

	ship.Rotation += rotationInput * ship.Stats.RotationSpeed * delta

	if thrustInput != 0 {
		ship.Velocity.X += math.Sin(ship.Rotation) * ship.Stats.ThrustForce * thrustInput * delta
		ship.Velocity.Y += -math.Cos(ship.Rotation) * ship.Stats.ThrustForce * thrustInput * delta
	}

	damping := math.Pow(ship.Stats.Damping, delta/(1.0/60.0))
	ship.Velocity.X *= damping
	ship.Velocity.Y *= damping
	ship.Velocity = ship.Velocity.LimitLength(ship.Stats.MaxSpeed)

	ship.X += ship.Velocity.X * delta
	ship.Y += ship.Velocity.Y * delta
}

func AdvanceShip(ship *entities.Ship, delta float64, bounds space.Bounds) {
	StepShip(ship, delta)
	wrapped := normalizePosition(ship.Position(), bounds)
	ship.X = wrapped.X
	ship.Y = wrapped.Y
}

func AdvanceShipWithMovePolicy(ship *entities.Ship, delta float64, bounds space.Bounds, canMove bool) {
	StepShipWithMovePolicy(ship, delta, canMove)
	wrapped := normalizePosition(ship.Position(), bounds)
	ship.X = wrapped.X
	ship.Y = wrapped.Y
}

func StepAsteroid(asteroid *entities.Asteroid, delta float64) {
	if asteroid.PendingDespawn {
		asteroid.DespawnDelay -= delta
		return
	}

	asteroid.X += asteroid.Velocity.X * delta
	asteroid.Y += asteroid.Velocity.Y * delta
}

func AdvanceAsteroid(asteroid *entities.Asteroid, delta float64, bounds space.Bounds) {
	StepAsteroid(asteroid, delta)
	wrapped := normalizePosition(asteroid.Position(), bounds)
	asteroid.X = wrapped.X
	asteroid.Y = wrapped.Y
}

func StepBullet(bullet *entities.Bullet, delta float64) {
	if bullet.PendingDespawn {
		bullet.DespawnDelay -= delta
		return
	}

	bullet.X += bullet.Velocity.X * delta
	bullet.Y += bullet.Velocity.Y * delta
	bullet.Life -= delta
}

func AdvanceBullet(bullet *entities.Bullet, delta float64, bounds space.Bounds) {
	StepBullet(bullet, delta)
	wrapped := normalizePosition(bullet.Position(), bounds)
	bullet.X = wrapped.X
	bullet.Y = wrapped.Y
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

func normalizePosition(position physics.Vector2, bounds space.Bounds) physics.Vector2 {
	return space.WrapPosition(position, bounds)
}
