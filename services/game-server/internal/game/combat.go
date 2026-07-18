package game

import (
	"github.com/Lokee86/space-rocks/services/game-server/internal/constants"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/damage"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/events"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/lives"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/runtime"
	"github.com/Lokee86/space-rocks/services/game-server/internal/logging"
	observability "github.com/Lokee86/space-rocks/shared/go/observabilityevent"
)

func (game *Game) handleBulletAsteroidCollisions() {
	hitBullets := map[string]bool{}
	hitAsteroids := map[string]*runtime.Asteroid{}
	hitAsteroidOwners := map[string]string{}

	for _, bulletID := range game.collisionProjectileIDsSorted() {
		bullet := game.entities.Projectiles[bulletID]
		if bullet == nil {
			continue
		}
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

		for _, asteroidRef := range game.asteroidCollisionCandidates(bulletBody) {
			asteroidID := asteroidRef.ID
			asteroid := game.entities.Asteroids[asteroidID]
			if asteroid == nil {
				continue
			}
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
			if event, ok := damageAppliedEventForResult(damageResult, collision.ImpactPosition.X, collision.ImpactPosition.Y); ok {
				game.recordDomainEvent(event)
			}
			game.spawnRadialEffectFromBullet(bullet, bullet.OwnerID, collision.ImpactPosition)
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
				hitAsteroidOwners,
			)
			break
		}
	}

	game.applyProjectileAsteroidHitConsequences(hitBullets, hitAsteroids, hitAsteroidOwners)
}

func (game *Game) recordProjectileAsteroidHit(
	collision ProjectileAsteroidCollision,
	bulletID string,
	bullet *runtime.Bullet,
	asteroidID string,
	asteroid *runtime.Asteroid,
	hitBullets map[string]bool,
	hitAsteroids map[string]*runtime.Asteroid,
	hitAsteroidOwners map[string]string,
) {
	hitBullets[bulletID] = true
	hitAsteroids[asteroidID] = asteroid
	hitAsteroidOwners[asteroidID] = bullet.OwnerID
	game.recordDomainEvent(events.Event{
		Type: events.EventBulletBlast,
		X:    collision.ImpactPosition.X,
		Y:    collision.ImpactPosition.Y,
	})
}

func (game *Game) applyProjectileAsteroidHitConsequences(
	hitBullets map[string]bool,
	hitAsteroids map[string]*runtime.Asteroid,
	hitAsteroidOwners map[string]string,
) {
	for bulletID := range hitBullets {
		bullet := game.entities.Projectiles[bulletID]
		bullet.MarkPendingDespawn(constants.CollisionDespawnDelay)
	}

	for asteroidID := range hitAsteroids {
		asteroid := game.entities.Asteroids[asteroidID]
		game.applyProjectileAsteroidDestruction(hitAsteroidOwners[asteroidID], asteroid)
	}
}

func (game *Game) handleShipAsteroidCollisions() {
	hitPlayers := map[string]*runtime.Ship{}

	for _, playerID := range game.collisionPlayerIDsSorted() {
		player := game.entities.Players[playerID]
		if player == nil {
			continue
		}
		if player.IsPendingDespawn() {
			continue
		}
		if !game.playerCanTakeCollisionDamage(playerID, player) {
			continue
		}

		playerBody, ok := player.CollisionBody(game.collisionShapes)
		if !ok {
			continue
		}

		for _, asteroidRef := range game.asteroidCollisionCandidates(playerBody) {
			asteroidID := asteroidRef.ID
			asteroid := game.entities.Asteroids[asteroidID]
			if asteroid == nil {
				continue
			}
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
			if event, ok := damageAppliedEventForResult(damageResult, player.Position().X, player.Position().Y); ok {
				game.recordDomainEvent(event)
			}
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
	game.applyFatalPlayerDamageWithInput(playerID, player, lives.DeathInput{
		CauseCode:   "collision",
		Attribution: lives.AttributionEnvironmental,
	})
}

func (game *Game) applyFatalPlayerDamage(playerID string, player *runtime.Ship, reasons ...string) {
	reason := "collision"
	if len(reasons) > 0 && reasons[0] != "" {
		reason = reasons[0]
	}
	game.applyFatalPlayerDamageWithInput(playerID, player, lives.DeathInput{CauseCode: reason, Attribution: lives.AttributionUnattributed})
}

func (game *Game) applyFatalPlayerDamageWithInput(playerID string, player *runtime.Ship, input lives.DeathInput) {
	position := player.Position()
	input.PlayerID = playerID
	input.DestroyedShipID = player.ID
	input.MatchID = game.matchID
	input.ModeID = game.modeID
	if session, ok := game.playerSessions[playerID]; ok {
		input.TeamID = session.TeamID
	}
	lives := 0
	respawnDelay := 0.0
	if _, ok := game.playerSessions[playerID]; ok {
		if session := game.playerSessions[playerID]; session != nil {
			session.CaptureBetweenLifeState(player)
		}
		deathFact := game.lifeRuntime.ApplyDeath(input)
		if !deathFact.Accepted {
			return
		}
		if state, ok := game.lifeRuntime.ParticipantSnapshot(playerID); ok {
			lives = game.projectedPlayerLives(playerID, state)
		}
		respawnDelay = deathFact.RespawnDelay
	}
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
	if lives > 0 && game.matchID != "" && game.matchTraceID != "" {
		logging.Emit(observability.Request{
			Event: observability.EventNamePlayerDied,
			Context: observability.Context{
				TraceID:  game.matchTraceID,
				MatchID:  game.matchID,
				PlayerID: playerID,
			},
			Fields: observability.Fields{
				"reason_code":   input.CauseCode,
				"lives":         lives,
				"respawn_delay": respawnDelay,
				"x":             position.X,
				"y":             position.Y,
			},
		})
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
