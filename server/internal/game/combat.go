package game

import "github.com/Lokee86/space-rocks/server/internal/constants"

func (game *Game) handleBulletAsteroidCollisions() {
	hitBullets := map[string]bool{}
	hitAsteroids := map[string]*Asteroid{}

	for bulletID, bullet := range game.state.Projectiles {
		if hitBullets[bulletID] {
			continue
		}
		if bullet.PendingDespawn {
			continue
		}

		bulletBody, ok := bullet.collisionBody(game.collisionShapes)
		if !ok {
			continue
		}

		for asteroidID, asteroid := range game.state.Asteroids {
			if _, ok := hitAsteroids[asteroidID]; ok {
				continue
			}
			if asteroid.PendingDespawn {
				continue
			}

			asteroidBody, ok := asteroid.collisionBody(game.collisionShapes)
			if !ok {
				continue
			}

			if _, ok := DetectCollision(bulletBody, asteroidBody); !ok {
				continue
			}

			hitBullets[bulletID] = true
			hitAsteroids[asteroidID] = asteroid
			game.broadcastEvent(EventState{
				Type: "bullet_blast",
				X:    bullet.X,
				Y:    bullet.Y,
			})
			break
		}
	}

	for bulletID := range hitBullets {
		bullet := game.state.Projectiles[bulletID]
		bullet.PendingDespawn = true
		bullet.DespawnDelay = constants.CollisionDespawnDelay
		bullet.Velocity = Vector2{}
	}

	for asteroidID := range hitAsteroids {
		asteroid := game.state.Asteroids[asteroidID]
		asteroid.PendingDespawn = true
		asteroid.DespawnDelay = constants.CollisionDespawnDelay
		asteroid.Velocity = Vector2{}
	}

	for _, asteroid := range hitAsteroids {
		game.spawnAsteroidFragments(asteroid)
	}
}

func (game *Game) broadcastEvent(event EventState) {
	for playerID := range game.state.Players {
		game.pendingEvents[playerID] = append(game.pendingEvents[playerID], event)
	}
}
