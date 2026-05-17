package game

import (
	"fmt"
	"math"
	"math/rand"

	"github.com/Lokee86/space-rocks/server/internal/constants"
)

func (game *Game) spawnBullet(ship *Ship) {
	forward := Vector2{X: 0, Y: -1}.rotated(ship.Rotation)

	game.nextBulletID++
	bulletID := fmt.Sprintf("bullet-%d", game.nextBulletID)
	game.state.Projectiles[bulletID] = &Bullet{
		ID:       bulletID,
		OwnerID:  ship.ID,
		X:        ship.X + forward.X*constants.BulletSpawnOffset,
		Y:        ship.Y + forward.Y*constants.BulletSpawnOffset,
		Rotation: ship.Rotation,
		Velocity: Vector2{
			X: forward.X * constants.BulletSpeed,
			Y: forward.Y * constants.BulletSpeed,
		},
		Life: constants.BulletLifetime,
	}
}

func (game *Game) spawnAsteroidBatch(target *Ship) {
	for range constants.AsteroidSpawnBatchSize {
		game.spawnAsteroid(target)
	}
}

func (game *Game) spawnAsteroid(target *Ship) {
	targetPosition := Vector2{X: target.X, Y: target.Y}
	spawn := game.randomAsteroidSpawnPosition(target)
	direction := spawn.directionTo(targetPosition).rotated(randomRange(
		-degreesToRadians(constants.AsteroidAimRandomnessDegrees),
		degreesToRadians(constants.AsteroidAimRandomnessDegrees),
	))
	velocity := direction.multiply(randomAsteroidSpeed())

	game.addAsteroid(spawn, velocity, rand.Intn(4)+1, rand.Intn(4))
}

func (game *Game) addAsteroid(position Vector2, velocity Vector2, size int, variant int) {
	asteroidID := game.nextAsteroidIDString()
	game.state.Asteroids[asteroidID] = &Asteroid{
		ID:       asteroidID,
		X:        position.X,
		Y:        position.Y,
		Velocity: velocity,
		Size:     size,
		Variant:  variant,
	}
}

func (game *Game) nextAsteroidIDString() string {
	for {
		game.nextAsteroidID++
		asteroidID := fmt.Sprintf("asteroid-%d", game.nextAsteroidID)
		if _, exists := game.state.Asteroids[asteroidID]; !exists {
			return asteroidID
		}
	}
}

func (game *Game) spawnAsteroidFragments(asteroid *Asteroid) {
	fragmentSize := asteroid.Size - 1
	if fragmentSize <= 0 {
		return
	}

	position := Vector2{X: asteroid.X, Y: asteroid.Y}
	for i := 0; i < 2; i++ {
		direction := randomUnitVector()
		game.addAsteroid(
			position,
			direction.multiply(randomAsteroidSpeed()),
			fragmentSize,
			rand.Intn(4),
		)
	}
}

func randomAsteroidSpeed() float64 {
	return randomRange(constants.AsteroidMinSpeed, constants.AsteroidMaxSpeed)
}

func randomUnitVector() Vector2 {
	return Vector2{X: 0, Y: -1}.rotated(randomRange(0, math.Pi*2))
}

func degreesToRadians(degrees float64) float64 {
	return degrees * math.Pi / 180
}
