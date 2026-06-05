package game

import (
	"github.com/Lokee86/space-rocks/server/internal/game/entities/pickups"
	"github.com/Lokee86/space-rocks/server/internal/game/runtime"
	"github.com/Lokee86/space-rocks/server/internal/game/physics"
	"github.com/Lokee86/space-rocks/server/internal/game/space"
)

type ProjectileAsteroidCollision struct {
	ProjectileID   string
	AsteroidID     string
	ImpactPosition physics.Vector2
}

type PlayerAsteroidCollision struct {
	PlayerID       string
	AsteroidID     string
	ImpactPosition physics.Vector2
}

type PlayerPickupCollision struct {
	PlayerID       string
	PickupID       string
	ImpactPosition physics.Vector2
}

func detectProjectileAsteroidCollision(
	bullet *runtime.Bullet,
	asteroid *runtime.Asteroid,
	catalog physics.CollisionShapeCatalog,
) (ProjectileAsteroidCollision, bool) {
	bulletBody, ok := bullet.CollisionBody(catalog)
	if !ok {
		return ProjectileAsteroidCollision{}, false
	}

	asteroidBody, ok := asteroid.CollisionBody(catalog)
	if !ok {
		return ProjectileAsteroidCollision{}, false
	}
	delta := space.Delta(bullet.Position(), asteroid.Position())
	asteroidBody.Position = bullet.Position().Add(delta)

	if _, ok := physics.DetectCollision(bulletBody, asteroidBody); !ok {
		return ProjectileAsteroidCollision{}, false
	}

	return ProjectileAsteroidCollision{
		ProjectileID:   bullet.ID,
		AsteroidID:     asteroid.ID,
		ImpactPosition: bullet.Position(),
	}, true
}

func detectPlayerAsteroidCollision(
	playerID string,
	player *runtime.Ship,
	asteroid *runtime.Asteroid,
	catalog physics.CollisionShapeCatalog,
) (PlayerAsteroidCollision, bool) {
	playerBody, ok := player.CollisionBody(catalog)
	if !ok {
		return PlayerAsteroidCollision{}, false
	}

	asteroidBody, ok := asteroid.CollisionBody(catalog)
	if !ok {
		return PlayerAsteroidCollision{}, false
	}
	delta := space.Delta(player.Position(), asteroid.Position())
	asteroidBody.Position = player.Position().Add(delta)

	if _, ok := physics.DetectCollision(playerBody, asteroidBody); !ok {
		return PlayerAsteroidCollision{}, false
	}

	return PlayerAsteroidCollision{
		PlayerID:       playerID,
		AsteroidID:     asteroid.ID,
		ImpactPosition: player.Position(),
	}, true
}

func detectPlayerPickupCollision(
	playerID string,
	player *runtime.Ship,
	pickup *pickups.Pickup,
	catalog physics.CollisionShapeCatalog,
) (PlayerPickupCollision, bool) {
	playerBody, ok := player.CollisionBody(catalog)
	if !ok {
		return PlayerPickupCollision{}, false
	}

	pickupBody := pickup.CollisionBody()
	delta := space.Delta(player.Position(), pickup.Position())
	pickupBody.Position = player.Position().Add(delta)

	if _, ok := physics.DetectCollision(playerBody, pickupBody); !ok {
		return PlayerPickupCollision{}, false
	}

	return PlayerPickupCollision{
		PlayerID:       playerID,
		PickupID:       pickup.ID,
		ImpactPosition: player.Position(),
	}, true
}
