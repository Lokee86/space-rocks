package spawning

import (
	"fmt"
	"math"
	"math/rand"

	"github.com/Lokee86/space-rocks/server/internal/constants"
	"github.com/Lokee86/space-rocks/server/internal/game/entities"
	"github.com/Lokee86/space-rocks/server/internal/game/physics"
	"github.com/Lokee86/space-rocks/server/internal/game/space"
)

type SpawnEntityType string

const (
	SpawnEntityTypeAsteroid SpawnEntityType = "asteroid"
)

type SpawnReason string

const (
	SpawnReasonTimedAsteroid    SpawnReason = "timed_asteroid"
	SpawnReasonAsteroidFragment SpawnReason = "asteroid_fragment"
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
}

func New() *Spawner {
	return &Spawner{}
}

func (spawner *Spawner) NextBulletID() string {
	spawner.nextBulletID++
	return fmt.Sprintf("bullet-%d", spawner.nextBulletID)
}

func (spawner *Spawner) BuildBullet(ship *entities.Ship) *entities.Bullet {
	forward := ship.Forward()
	spawnPosition := ship.Position().Add(forward.Multiply(ship.Stats.BulletSpawnOffset))
	velocity := forward.Multiply(ship.Stats.BulletSpeed)
	bulletID := spawner.NextBulletID()

	return entities.NewBullet(
		bulletID,
		ship.ID,
		spawnPosition,
		ship.Rotation,
		velocity,
		ship.Stats.BulletLifetime,
	)
}

func (spawner *Spawner) NextAsteroidID(existing map[string]*entities.Asteroid) string {
	for {
		spawner.nextAsteroidID++
		asteroidID := fmt.Sprintf("asteroid-%d", spawner.nextAsteroidID)
		if _, exists := existing[asteroidID]; !exists {
			return asteroidID
		}
	}
}

func (spawner *Spawner) PlanTimedAsteroidSpawn(position physics.Vector2, targetPosition physics.Vector2) AsteroidSpawnPlan {
	direction := space.Direction(position, targetPosition).Rotated(randomRange(
		-degreesToRadians(constants.AsteroidAimRandomnessDegrees),
		degreesToRadians(constants.AsteroidAimRandomnessDegrees),
	))
	velocity := direction.Multiply(randomAsteroidSpeed())

	return AsteroidSpawnPlan{
		EntityType: SpawnEntityTypeAsteroid,
		Reason:     SpawnReasonTimedAsteroid,
		Position:   position,
		Velocity:   velocity,
		Size:       rand.Intn(4) + 1,
		Variant:    rand.Intn(4),
	}
}

func (spawner *Spawner) PlanAsteroidFragmentSpawns(asteroid *entities.Asteroid) []AsteroidSpawnPlan {
	fragmentSize := asteroid.FragmentSize()
	if fragmentSize <= 0 {
		return nil
	}

	position := asteroid.Position()
	plans := make([]AsteroidSpawnPlan, 0, 2)
	for i := 0; i < 2; i++ {
		direction := randomUnitVector()
		plans = append(plans, AsteroidSpawnPlan{
			EntityType: SpawnEntityTypeAsteroid,
			Reason:     SpawnReasonAsteroidFragment,
			Position:   position,
			Velocity:   direction.Multiply(randomAsteroidSpeed()),
			Size:       fragmentSize,
			Variant:    rand.Intn(constants.AsteroidVariants),
		})
	}
	return plans
}

func (spawner *Spawner) RandomAsteroidSpeed() float64 {
	return randomAsteroidSpeed()
}

func (spawner *Spawner) RandomUnitVector() physics.Vector2 {
	return randomUnitVector()
}

func (spawner *Spawner) DegreesToRadians(degrees float64) float64 {
	return degreesToRadians(degrees)
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

func randomRange(minValue float64, maxValue float64) float64 {
	return minValue + rand.Float64()*(maxValue-minValue)
}
