package game

import (
	"testing"

	"github.com/Lokee86/space-rocks/server/internal/constants"
)

func TestHandleBulletAsteroidCollisionsDelaysHitDespawns(t *testing.T) {
	game := New()
	game.collisionShapes = CollisionShapeCatalog{
		Bullet: ImportedCollisionShape{
			Type:   "capsule",
			Radius: 3,
			Height: 24,
		},
		Asteroids: []ImportedCollisionShape{
			{
				Type: "polygon",
				Points: [][]float64{
					{-40, -40},
					{40, -40},
					{40, 40},
					{-40, 40},
				},
			},
		},
	}
	game.state.Projectiles["bullet-1"] = &Bullet{
		ID: "bullet-1",
		X:  100,
		Y:  100,
	}
	game.state.Asteroids["asteroid-1"] = &Asteroid{
		ID:   "asteroid-1",
		X:    100,
		Y:    100,
		Size: 1,
	}
	game.state.Asteroids["asteroid-2"] = &Asteroid{
		ID:   "asteroid-2",
		X:    1000,
		Y:    1000,
		Size: 1,
	}

	game.handleBulletAsteroidCollisions()

	bullet, ok := game.state.Projectiles["bullet-1"]
	if !ok {
		t.Fatal("expected hit bullet to remain during despawn delay")
	}
	if !bullet.PendingDespawn {
		t.Fatal("expected hit bullet to be marked for delayed despawn")
	}

	asteroid, ok := game.state.Asteroids["asteroid-1"]
	if !ok {
		t.Fatal("expected hit asteroid to remain during despawn delay")
	}
	if !asteroid.PendingDespawn {
		t.Fatal("expected hit asteroid to be marked for delayed despawn")
	}
	if _, ok := game.state.Asteroids["asteroid-2"]; !ok {
		t.Fatal("expected untouched asteroid to remain")
	}

	game.Step(constants.CollisionDespawnDelay)

	if _, ok := game.state.Projectiles["bullet-1"]; ok {
		t.Fatal("expected hit bullet to be removed after despawn delay")
	}
	if _, ok := game.state.Asteroids["asteroid-1"]; ok {
		t.Fatal("expected hit asteroid to be removed after despawn delay")
	}
}
