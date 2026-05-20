package game

import (
	"github.com/Lokee86/space-rocks/server/internal/constants"
	"github.com/Lokee86/space-rocks/server/internal/game/entities"
	"github.com/Lokee86/space-rocks/server/internal/game/physics"
	"github.com/Lokee86/space-rocks/server/internal/logging"
)

func (game *Game) handleBulletAsteroidCollisions() {
	hitBullets := map[string]bool{}
	hitAsteroids := map[string]*entities.Asteroid{}
	scoreAwards := []ScoreAward{}

	for bulletID, bullet := range game.state.Projectiles {
		if hitBullets[bulletID] {
			continue
		}
		if bullet.IsPendingDespawn() {
			continue
		}

		bulletBody, ok := bullet.CollisionBody(game.collisionShapes)
		if !ok {
			continue
		}

		for asteroidID, asteroid := range game.state.Asteroids {
			if _, ok := hitAsteroids[asteroidID]; ok {
				continue
			}
			if asteroid.IsPendingDespawn() {
				continue
			}

			asteroidBody, ok := asteroid.CollisionBody(game.collisionShapes)
			if !ok {
				continue
			}

			if _, ok := physics.DetectCollision(bulletBody, asteroidBody); !ok {
				continue
			}

			hitBullets[bulletID] = true
			hitAsteroids[asteroidID] = asteroid
			scoreAwards = append(scoreAwards, NewAsteroidHitScoreAward(bullet.OwnerID, asteroid))
			impactPosition := bullet.Position()
			game.broadcastEvent(EventState{
				Type: PacketTypeBulletBlast,
				X:    impactPosition.X,
				Y:    impactPosition.Y,
			})
			break
		}
	}

	for _, award := range scoreAwards {
		game.awardScore(award)
	}

	for bulletID := range hitBullets {
		bullet := game.state.Projectiles[bulletID]
		bullet.MarkPendingDespawn(constants.CollisionDespawnDelay)
	}

	for asteroidID := range hitAsteroids {
		asteroid := game.state.Asteroids[asteroidID]
		asteroid.MarkPendingDespawn(constants.CollisionDespawnDelay)
	}

	for _, asteroid := range hitAsteroids {
		game.spawnAsteroidFragments(asteroid)
	}
}

func (game *Game) handleShipAsteroidCollisions() {
	hitPlayers := map[string]*entities.Ship{}

	for playerID, player := range game.state.Players {
		if player.IsPendingDespawn() {
			continue
		}
		if player.Paused || player.IsInvulnerable() || !player.DevTools.CanTakeDamage() {
			continue
		}

		playerBody, ok := player.CollisionBody(game.collisionShapes)
		if !ok {
			continue
		}

		for _, asteroid := range game.state.Asteroids {
			if asteroid.IsPendingDespawn() {
				continue
			}

			asteroidBody, ok := asteroid.CollisionBody(game.collisionShapes)
			if !ok {
				continue
			}

			if _, ok := physics.DetectCollision(playerBody, asteroidBody); !ok {
				continue
			}

			hitPlayers[playerID] = player
			break
		}
	}

	for playerID, player := range hitPlayers {
		position := player.Position()
		player.MarkPendingDespawn(constants.CollisionDespawnDelay)
		lives := 0
		respawnDelay := 0.0
		if session, ok := game.playerSessions[playerID]; ok {
			session.Score = player.Score
			session.RecordDeath(player.DevTools)
			player.Lives = session.Lives
			lives = session.Lives
			respawnDelay = session.RespawnCooldown
		}
		if lives <= 0 {
			logging.Game.Info("player game over",
				logging.FieldPlayerID, playerID,
				"score", player.Score,
				"x", position.X,
				"y", position.Y,
			)
		} else {
			logging.Game.Info("player died",
				logging.FieldPlayerID, playerID,
				"lives", lives,
				"respawn_delay", respawnDelay,
				"x", position.X,
				"y", position.Y,
			)
		}
		game.broadcastEvent(EventState{
			Type:         PacketTypeShipDeath,
			PlayerID:     playerID,
			Lives:        lives,
			RespawnDelay: respawnDelay,
			X:            position.X,
			Y:            position.Y,
		})
	}

}

func (game *Game) broadcastEvent(event EventState) {
	for playerID := range game.state.Players {
		game.pendingEvents[playerID] = append(game.pendingEvents[playerID], event)
	}
}
