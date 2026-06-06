package game

import (
	"github.com/Lokee86/space-rocks/server/internal/constants"
	"github.com/Lokee86/space-rocks/server/internal/game/damage"
	"github.com/Lokee86/space-rocks/server/internal/game/runtime"
	"github.com/Lokee86/space-rocks/server/internal/game/events"
	"github.com/Lokee86/space-rocks/server/internal/game/scoring"
	"github.com/Lokee86/space-rocks/server/internal/logging"
)

func (game *Game) handleBulletAsteroidCollisions() {
	hitBullets := map[string]bool{}
	hitAsteroids := map[string]*runtime.Asteroid{}
	scoreAwards := []scoring.Award{}

	for bulletID, bullet := range game.entities.Projectiles {
		if hitBullets[bulletID] {
			continue
		}
		if bullet.IsPendingDespawn() {
			continue
		}

		for asteroidID, asteroid := range game.entities.Asteroids {
			if _, ok := hitAsteroids[asteroidID]; ok {
				continue
			}
			if asteroid.IsPendingDespawn() {
				continue
			}

			collision, ok := detectProjectileAsteroidCollision(bullet, asteroid, game.collisionShapes)
			if !ok {
				continue
			}

			damageRequest := projectileAsteroidDamageRequest(collision, bullet, asteroid)
			damageResult := damage.ResolveSingle(damageRequest)
			applyDamageResultToAsteroid(asteroid, damageResult)
			hitBullets[bulletID] = true
			if !damageResult.Destroyed {
				break
			}

			game.recordProjectileAsteroidHit(
				collision,
				bulletID,
				bullet,
				asteroidID,
				asteroid,
				hitBullets,
				hitAsteroids,
				&scoreAwards,
			)
			break
		}
	}

	game.applyProjectileAsteroidHitConsequences(hitBullets, hitAsteroids, scoreAwards)
}

func (game *Game) recordProjectileAsteroidHit(
	collision ProjectileAsteroidCollision,
	bulletID string,
	bullet *runtime.Bullet,
	asteroidID string,
	asteroid *runtime.Asteroid,
	hitBullets map[string]bool,
	hitAsteroids map[string]*runtime.Asteroid,
	scoreAwards *[]scoring.Award,
) {
	hitBullets[bulletID] = true
	hitAsteroids[asteroidID] = asteroid
	awards := game.scoringPolicy.Evaluate(scoring.Event{
		Kind:         scoring.EventAsteroidDestroyed,
		PlayerID:     bullet.OwnerID,
		TargetID:     asteroid.ID,
		AsteroidSize: asteroid.Size,
	})
	*scoreAwards = append(*scoreAwards, awards...)
	game.recordDomainEvent(events.Event{
		Type: events.EventBulletBlast,
		X:    collision.ImpactPosition.X,
		Y:    collision.ImpactPosition.Y,
	})
}

func (game *Game) applyProjectileAsteroidHitConsequences(
	hitBullets map[string]bool,
	hitAsteroids map[string]*runtime.Asteroid,
	scoreAwards []scoring.Award,
) {
	for _, award := range scoreAwards {
		game.awardScore(award)
	}

	for bulletID := range hitBullets {
		bullet := game.entities.Projectiles[bulletID]
		bullet.MarkPendingDespawn(constants.CollisionDespawnDelay)
	}

	for asteroidID := range hitAsteroids {
		asteroid := game.entities.Asteroids[asteroidID]
		asteroid.MarkPendingDespawn(constants.CollisionDespawnDelay)
	}

	for _, asteroid := range hitAsteroids {
		game.spawnAsteroidFragments(asteroid)
		game.maybeDropPickupFromAsteroidLocked(asteroid)
	}
}

func (game *Game) handleShipAsteroidCollisions() {
	hitPlayers := map[string]*runtime.Ship{}

	for playerID, player := range game.entities.Players {
		if player.IsPendingDespawn() {
			continue
		}
		if !game.playerCanTakeCollisionDamage(playerID, player) {
			continue
		}

		for asteroidID, asteroid := range game.entities.Asteroids {
			if asteroid.IsPendingDespawn() {
				continue
			}

			collision, ok := detectPlayerAsteroidCollision(playerID, player, asteroid, game.collisionShapes)
			if !ok {
				continue
			}

			damageRequest := playerAsteroidDamageRequest(collision, asteroidID, player, asteroid)
			damageResult := damage.ResolveSingle(damageRequest)
			applyDamageResultToPlayer(player, damageResult)
			if !damageResult.Fatal || damageResult.TargetEntityType != damage.EntityTypePlayer {
				continue
			}

			hitPlayers[playerID] = player
			break
		}
	}

	for playerID, player := range hitPlayers {
		game.applyPlayerFatalAsteroidHit(playerID, player)
	}

}

func (game *Game) applyPlayerFatalAsteroidHit(playerID string, player *runtime.Ship) {
	game.applyFatalPlayerDamage(playerID, player)
}

func (game *Game) applyFatalPlayerDamage(playerID string, player *runtime.Ship) {
	position := player.Position()
	if cameraView, ok := game.cameraViews[playerID]; ok && cameraView != nil {
		cameraView.X = position.X
		cameraView.Y = position.Y
	} else {
		game.cameraViews[playerID] = &runtime.CameraView{
			X:      position.X,
			Y:      position.Y,
			Config: player.Config,
		}
	}
	player.MarkPendingDespawn(constants.CollisionDespawnDelay)
	lives := 0
	score := 0
	respawnDelay := 0.0
	if session, ok := game.playerSessions[playerID]; ok {
		score = session.Score
		if session.LifeOptions.CanLoseLives() && session.Lives > 0 {
			game.addPlayerLivesLocked(playerID, -1)
		}
		if session.Lives > 0 {
			session.RespawnCooldown = constants.PlayerRespawnDelay
		}
		lives = session.Lives
		respawnDelay = session.RespawnCooldown
	}
	if lives <= 0 {
		logging.Game.Info("player game over",
			logging.FieldPlayerID, playerID,
			"score", score,
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
	game.recordDomainEvent(events.Event{
		Type:         events.EventShipDeath,
		PlayerID:     playerID,
		Lives:        lives,
		RespawnDelay: respawnDelay,
		X:            position.X,
		Y:            position.Y,
	})
}
