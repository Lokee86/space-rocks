package game

import (
	"github.com/Lokee86/space-rocks/server/internal/constants"
	"github.com/Lokee86/space-rocks/server/internal/game/damage"
	"github.com/Lokee86/space-rocks/server/internal/game/entities"
	"github.com/Lokee86/space-rocks/server/internal/game/events"
	"github.com/Lokee86/space-rocks/server/internal/game/scoring"
	"github.com/Lokee86/space-rocks/server/internal/logging"
)

func (game *Game) handleBulletAsteroidCollisions() {
	hitBullets := map[string]bool{}
	hitAsteroids := map[string]*entities.Asteroid{}
	scoreAwards := []scoring.Award{}

	for bulletID, bullet := range game.state.Projectiles {
		if hitBullets[bulletID] {
			continue
		}
		if bullet.IsPendingDespawn() {
			continue
		}

		for asteroidID, asteroid := range game.state.Asteroids {
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

			damageRequest := projectileAsteroidDamageRequest(collision)
			damageResult := damage.Resolve(damageRequest)
			if !damageResult.Destroyed {
				continue
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

func projectileAsteroidDamageRequest(collision ProjectileAsteroidCollision) damage.DamageRequest {
	return damage.DamageRequest{
		TargetEntityID:   collision.AsteroidID,
		TargetEntityType: damage.EntityTypeAsteroid,
		SourceEntityID:   collision.ProjectileID,
		SourceEntityType: damage.EntityTypeProjectile,
		Amount:           1,
		Type:             damage.DamageTypeProjectile,
		Flags: damage.DamageFlags{
			Lethal: true,
		},
	}
}

func (game *Game) recordProjectileAsteroidHit(
	collision ProjectileAsteroidCollision,
	bulletID string,
	bullet *entities.Bullet,
	asteroidID string,
	asteroid *entities.Asteroid,
	hitBullets map[string]bool,
	hitAsteroids map[string]*entities.Asteroid,
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
	hitAsteroids map[string]*entities.Asteroid,
	scoreAwards []scoring.Award,
) {
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
		if !game.playerCanTakeCollisionDamage(playerID, player) {
			continue
		}

		for asteroidID, asteroid := range game.state.Asteroids {
			if asteroid.IsPendingDespawn() {
				continue
			}

			collision, ok := detectPlayerAsteroidCollision(playerID, player, asteroid, game.collisionShapes)
			if !ok {
				continue
			}

			damageRequest := playerAsteroidDamageRequest(collision, asteroidID)
			damageResult := damage.Resolve(damageRequest)
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

func (game *Game) applyPlayerFatalAsteroidHit(playerID string, player *entities.Ship) {
	position := player.Position()
	player.MarkPendingDespawn(constants.CollisionDespawnDelay)
	lives := 0
	respawnDelay := 0.0
	if session, ok := game.playerSessions[playerID]; ok {
		session.Score = player.Score
		session.RecordDeath()
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
	game.recordDomainEvent(events.Event{
		Type:         events.EventShipDeath,
		PlayerID:     playerID,
		Lives:        lives,
		RespawnDelay: respawnDelay,
		X:            position.X,
		Y:            position.Y,
	})
}

func playerAsteroidDamageRequest(collision PlayerAsteroidCollision, asteroidID string) damage.DamageRequest {
	return damage.DamageRequest{
		TargetEntityID:   collision.PlayerID,
		TargetEntityType: damage.EntityTypePlayer,
		SourceEntityID:   asteroidID,
		SourceEntityType: damage.EntityTypeAsteroid,
		Amount:           1,
		Type:             damage.DamageTypeCollision,
		Flags: damage.DamageFlags{
			Lethal: true,
		},
	}
}
