package motion

import (
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/physics"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/runtime"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/space"
)

func ApplyShipAsteroidRepulsion(ship *runtime.Ship, asteroid *runtime.Asteroid, shipImpulse float64, asteroidImpulse float64) {
	if ship == nil || asteroid == nil || shipImpulse < 0 || asteroidImpulse < 0 {
		return
	}
	direction := space.Direction(asteroid.Position(), ship.Position())
	if direction.LengthSquared() == 0 {
		direction = physics.Vector2{Y: -1}
	}
	ship.Velocity = ship.Velocity.Add(direction.Multiply(shipImpulse)).LimitLength(ship.Stats.MaxSpeed)
	asteroid.Velocity = asteroid.Velocity.Add(direction.Multiply(-asteroidImpulse))
}
