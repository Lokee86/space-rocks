package game

import (
	"github.com/Lokee86/space-rocks/server/internal/game/entities"
	"github.com/Lokee86/space-rocks/server/internal/game/physics"
	"github.com/Lokee86/space-rocks/server/internal/game/spawning"
)

func (game *Game) DevtoolsRandomUnitVector() physics.Vector2 {
	return game.spawner.RandomUnitVector()
}

func (game *Game) DevtoolsNextBulletID() string {
	return game.spawner.NextBulletID()
}

func (game *Game) DevtoolsAddBullet(bullet *entities.Bullet) bool {
	if bullet == nil {
		return false
	}
	game.state.Projectiles[bullet.ID] = bullet
	return true
}

func (game *Game) DevtoolsRandomAsteroidSpeed() float64 {
	return game.spawner.RandomAsteroidSpeed()
}

func (game *Game) DevtoolsApplyAsteroidSpawnPlan(plan spawning.AsteroidSpawnPlan) *entities.Asteroid {
	return game.applyAsteroidSpawn(plan)
}

func (game *Game) DevtoolsEnsurePlayerSession(playerID string, spawnPosition physics.Vector2) bool {
	return game.ensureDebugPlayerSession(playerID, spawnPosition) != nil
}

func (game *Game) DevtoolsSpawnPlayerShip(playerID string, spawnPosition physics.Vector2) bool {
	session, ok := game.playerSessions[playerID]
	if !ok || session == nil {
		return false
	}
	return game.applyDebugPlayerShip(playerID, session, spawnPosition)
}

func (game *Game) DevtoolsPlayerIDOccupied(playerID string) bool {
	return game.isDebugGameplayPlayerIDOccupied(playerID)
}

func (game *Game) DevtoolsReservePlayerID(playerID string) bool {
	return game.reserveDebugGameplayPlayerID(playerID)
}
