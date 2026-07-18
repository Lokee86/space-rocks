package game

import (
	"github.com/Lokee86/space-rocks/services/game-server/internal/constants"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/runtime"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/scoring"
)

func (game *Game) applyProjectileAsteroidDestruction(playerID string, asteroid *runtime.Asteroid) {
	game.damageOverTime().RemoveTarget(asteroid.ID)
	awards := game.scoringPolicy.Evaluate(scoring.Event{
		Kind:         scoring.EventAsteroidDestroyed,
		PlayerID:     playerID,
		TargetID:     asteroid.ID,
		AsteroidSize: asteroid.Size,
	})
	game.applyDestructionAwardsLocked(asteroid.ID, awards)

	asteroid.MarkPendingDespawn(constants.CollisionDespawnDelay)
	game.spawnAsteroidFragments(asteroid)
	game.maybeDropPickupFromAsteroidLocked(asteroid)
}
