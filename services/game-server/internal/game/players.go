package game

import (
	"fmt"

	"github.com/Lokee86/space-rocks/server/internal/game/entities"
	"github.com/Lokee86/space-rocks/server/internal/logging"
)

func (game *Game) AddPlayer() string {
	game.mu.Lock()
	defer game.mu.Unlock()

	playerIndex := game.nextID
	game.nextID++

	playerID := fmt.Sprintf("player-%d", game.nextID)
	spawnPlan := game.planInitialPlayerSpawn(playerIndex, playerID)
	spawnPosition := spawnPlan.Position
	session := newPlayerSession(playerID, spawnPosition)
	player := session.NewShip(spawnPosition)
	game.playerSessions[playerID] = session
	game.state.Players[playerID] = player
	game.cameraViews[playerID] = &entities.CameraView{
		X:      player.X,
		Y:      player.Y,
		Config: player.Config,
	}
	game.pendingPresentationEvents[playerID] = nil
	logging.Game.Debug("player added",
		logging.FieldPlayerID, playerID,
		"x", spawnPosition.X,
		"y", spawnPosition.Y,
		"lives", session.Lives,
	)

	return playerID
}

func (game *Game) RemovePlayer(playerID string) {
	game.mu.Lock()
	defer game.mu.Unlock()

	delete(game.state.Players, playerID)
	delete(game.cameraViews, playerID)
	delete(game.playerSessions, playerID)
	game.clearTargetsForMissingPlayersLocked()
	delete(game.pendingPresentationEvents, playerID)
	logging.Game.Debug("player removed", logging.FieldPlayerID, playerID)
}

func (game *Game) playerLives(playerID string) int {
	if session, ok := game.playerSessions[playerID]; ok {
		return session.Lives
	}
	if player, ok := game.state.Players[playerID]; ok {
		return player.Lives
	}

	return 0
}
