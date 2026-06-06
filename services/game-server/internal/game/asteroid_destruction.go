package game

import (
	"github.com/Lokee86/space-rocks/server/internal/constants"
	"github.com/Lokee86/space-rocks/server/internal/game/runtime"
	"github.com/Lokee86/space-rocks/server/internal/game/scoring"
)

func (game *Game) applyProjectileAsteroidDestruction(playerID string, asteroid *runtime.Asteroid) {
	awards := game.scoringPolicy.Evaluate(scoring.Event{
		Kind:         scoring.EventAsteroidDestroyed,
		PlayerID:     playerID,
		TargetID:     asteroid.ID,
		AsteroidSize: asteroid.Size,
	})
	for _, award := range awards {
		game.awardScore(award)
	}

	asteroid.MarkPendingDespawn(constants.CollisionDespawnDelay)
	game.spawnAsteroidFragments(asteroid)
	game.maybeDropPickupFromAsteroidLocked(asteroid)
}
