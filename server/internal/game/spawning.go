package game

import (
	"fmt"
	"math"
	"math/rand"

	"github.com/Lokee86/space-rocks/server/internal/constants"
	"github.com/Lokee86/space-rocks/server/internal/game/entities"
	"github.com/Lokee86/space-rocks/server/internal/game/physics"
	"github.com/Lokee86/space-rocks/server/internal/logging"
)

func (game *Game) spawnBullet(ship *entities.Ship) {
	forward := ship.Forward()
	spawnPosition := ship.Position().Add(forward.Multiply(constants.BulletSpawnOffset))
	velocity := forward.Multiply(constants.BulletSpeed)

	game.nextBulletID++
	bulletID := fmt.Sprintf("bullet-%d", game.nextBulletID)
	game.state.Projectiles[bulletID] = entities.NewBullet(
		bulletID,
		ship.ID,
		spawnPosition,
		ship.Rotation,
		velocity,
	)
}

func (game *Game) spawnAsteroidBatch(view *entities.CameraView) {
	for range constants.AsteroidSpawnBatchSize {
		game.spawnAsteroid(view)
	}
}

func (game *Game) spawnAsteroid(view *entities.CameraView) {
	targetPosition := view.Position()
	spawn := game.randomAsteroidSpawnPosition(view)
	direction := spawn.DirectionTo(targetPosition).Rotated(randomRange(
		-degreesToRadians(constants.AsteroidAimRandomnessDegrees),
		degreesToRadians(constants.AsteroidAimRandomnessDegrees),
	))
	velocity := direction.Multiply(randomAsteroidSpeed())

	game.addAsteroid(spawn, velocity, rand.Intn(4)+1, rand.Intn(4))
}

func (game *Game) addAsteroid(position physics.Vector2, velocity physics.Vector2, size int, variant int) {
	asteroidID := game.nextAsteroidIDString()
	game.state.Asteroids[asteroidID] = entities.NewAsteroid(asteroidID, position, velocity, size, variant)
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

func (game *Game) spawnAsteroidFragments(asteroid *entities.Asteroid) {
	fragmentSize := asteroid.FragmentSize()
	if fragmentSize <= 0 {
		return
	}

	position := asteroid.Position()
	logging.Game.Debug("asteroid split",
		"asteroid_id", asteroid.ID,
		"source_size", asteroid.Size,
		"fragment_size", fragmentSize,
		"x", position.X,
		"y", position.Y,
	)
	for i := 0; i < 2; i++ {
		direction := randomUnitVector()
		game.addAsteroid(
			position,
			direction.Multiply(randomAsteroidSpeed()),
			fragmentSize,
			rand.Intn(4),
		)
	}
}

func randomAsteroidSpeed() float64 {
	return randomRange(constants.AsteroidMinSpeed, constants.AsteroidMaxSpeed)
}

func randomUnitVector() physics.Vector2 {
	return physics.Vector2{X: 0, Y: -1}.Rotated(randomRange(0, math.Pi*2))
}

func degreesToRadians(degrees float64) float64 {
	return degrees * math.Pi / 180
}
