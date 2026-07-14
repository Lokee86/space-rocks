package spawning

import (
	"fmt"
	"math"

	"github.com/Lokee86/space-rocks/services/game-server/internal/constants"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/asteroids"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/physics"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/rng"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/runtime"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/space"
)

type SpawnEntityType string

const (
	SpawnEntityTypeAsteroid SpawnEntityType = "asteroid"
)

type SpawnReason string

const (
	SpawnReasonTimedAsteroid    SpawnReason = "timed_asteroid"
	SpawnReasonAsteroidFragment SpawnReason = "asteroid_fragment"
	SpawnReasonDebugAsteroid    SpawnReason = "debug_asteroid"
)

type AsteroidSpawnPlan struct {
	EntityType SpawnEntityType
	Reason     SpawnReason
	Position   physics.Vector2
	Velocity   physics.Vector2
	Size       int
	Variant    int
}

type Spawner struct {
	nextAsteroidID int
	nextBulletID   int
	rngSource      *rng.Source
}

func New(source *rng.Source) *Spawner {
	return &Spawner{rngSource: source}
}

func (spawner *Spawner) NextBulletID() string {
	spawner.nextBulletID++
	return fmt.Sprintf("bullet-%d", spawner.nextBulletID)
}

func (spawner *Spawner) NextAsteroidID(existing map[string]*runtime.Asteroid) string {
	for {
		spawner.nextAsteroidID++
		asteroidID := fmt.Sprintf("asteroid-%d", spawner.nextAsteroidID)
		if _, exists := existing[asteroidID]; !exists {
			return asteroidID
		}
	}
}

func (spawner *Spawner) TotalAsteroidsSpawned() int {
	return spawner.nextAsteroidID
}

func (spawner *Spawner) PlanTimedAsteroidSpawn(position physics.Vector2, targetPosition physics.Vector2) AsteroidSpawnPlan {
	direction := space.Direction(position, targetPosition).Rotated(spawner.randomRange(
		-degreesToRadians(constants.AsteroidAimRandomnessDegrees),
		degreesToRadians(constants.AsteroidAimRandomnessDegrees),
	))
	velocity := direction.Multiply(spawner.randomAsteroidSpeed())

	return AsteroidSpawnPlan{
		EntityType: SpawnEntityTypeAsteroid,
		Reason:     SpawnReasonTimedAsteroid,
		Position:   position,
		Velocity:   velocity,
		Size:       spawner.rngSource.Intn(4) + 1,
		Variant:    asteroids.TimedSpawnVariantIndex(spawner.rngSource.Float64()),
	}
}

func (spawner *Spawner) PlanAsteroidFragmentSpawns(asteroid *runtime.Asteroid) []AsteroidSpawnPlan {
	fragmentSize := asteroid.FragmentSize()
	if fragmentSize <= 0 {
		return nil
	}

	position := asteroid.Position()
	plans := make([]AsteroidSpawnPlan, 0, 2)
	for i := 0; i < 2; i++ {
		direction := spawner.randomUnitVector()
		plans = append(plans, AsteroidSpawnPlan{
			EntityType: SpawnEntityTypeAsteroid,
			Reason:     SpawnReasonAsteroidFragment,
			Position:   position,
			Velocity:   direction.Multiply(spawner.randomAsteroidSpeed()),
			Size:       fragmentSize,
			Variant:    asteroids.FragmentSpawnVariantIndex(spawner.rngSource.Float64()),
		})
	}
	return plans
}

func (spawner *Spawner) RandomAsteroidSpeed() float64 {
	return spawner.randomAsteroidSpeed()
}

func (spawner *Spawner) RandomUnitVector() physics.Vector2 {
	return spawner.randomUnitVector()
}

func (spawner *Spawner) DegreesToRadians(degrees float64) float64 {
	return degreesToRadians(degrees)
}

func (spawner *Spawner) randomAsteroidSpeed() float64 {
	return spawner.randomRange(constants.AsteroidMinSpeed, constants.AsteroidMaxSpeed)
}

func (spawner *Spawner) randomUnitVector() physics.Vector2 {
	return physics.Vector2{X: 0, Y: -1}.Rotated(spawner.randomRange(0, math.Pi*2))
}

func degreesToRadians(degrees float64) float64 {
	return degrees * math.Pi / 180
}

func (spawner *Spawner) randomRange(minValue float64, maxValue float64) float64 {
	return minValue + spawner.rngSource.Float64()*(maxValue-minValue)
}
