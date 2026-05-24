package motion

import "github.com/Lokee86/space-rocks/server/internal/game/entities"

func StepShip(ship *entities.Ship, delta float64) {
	ship.ApplyInput(delta)
}

func StepAsteroid(asteroid *entities.Asteroid, delta float64) {
	asteroid.Step(delta)
}

func StepBullet(bullet *entities.Bullet, delta float64) {
	bullet.Step(delta)
}
