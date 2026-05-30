package entities

import (
	"github.com/Lokee86/space-rocks/server/internal/constants"
	"github.com/Lokee86/space-rocks/server/internal/game/physics"
)

func NewAsteroid(id string, position physics.Vector2, velocity physics.Vector2, size int, variant int) *Asteroid {
	return &Asteroid{
		ID:              id,
		X:               position.X,
		Y:               position.Y,
		Velocity:        velocity,
		Size:            size,
		Variant:         variant,
		Health:          constants.AsteroidHealth,
		CollisionDamage: constants.AsteroidCollisionDamage,
	}
}

func (asteroid *Asteroid) State() AsteroidState {
	return AsteroidState{
		ID:      asteroid.ID,
		X:       asteroid.X,
		Y:       asteroid.Y,
		Size:    asteroid.Size,
		Scale:   float64(asteroid.Size) * constants.AsteroidSizeScale,
		Variant: asteroid.Variant,
	}
}

func (asteroid *Asteroid) Position() physics.Vector2 {
	return physics.Vector2{X: asteroid.X, Y: asteroid.Y}
}

func (asteroid *Asteroid) IsPendingDespawn() bool {
	return asteroid.PendingDespawn
}

func (asteroid *Asteroid) ReadyForRemoval() bool {
	return asteroid.PendingDespawn && asteroid.DespawnDelay <= 0
}

func (asteroid *Asteroid) MarkPendingDespawn(delay float64) {
	asteroid.PendingDespawn = true
	asteroid.DespawnDelay = delay
	asteroid.Velocity = physics.Vector2{}
}

func (asteroid *Asteroid) FragmentSize() int {
	return asteroid.Size - 1
}

func (asteroid *Asteroid) CollisionBody(catalog physics.CollisionShapeCatalog) (physics.CollisionBody, bool) {
	shape, err := catalog.AsteroidShape(asteroid.Variant, asteroid.Size)
	if err != nil {
		return physics.CollisionBody{}, false
	}

	return physics.CollisionBody{
		ID:       asteroid.ID,
		Position: asteroid.Position(),
		Shape:    shape,
	}, true
}
