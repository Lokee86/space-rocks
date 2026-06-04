package game

import (
	"fmt"

	"github.com/Lokee86/space-rocks/server/internal/constants"
	"github.com/Lokee86/space-rocks/server/internal/game/runtime"
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
	game.entities.Players[playerID] = player
	game.setPlayerCameraViewLocked(playerID, player)
	game.pendingPresentationEvents[playerID] = nil
	logging.Game.Debug("player added",
		logging.FieldPlayerID, playerID,
		"x", spawnPosition.X,
		"y", spawnPosition.Y,
		"lives", session.Lives,
	)

	return playerID
}

func (game *Game) setPlayerCameraViewLocked(playerID string, player *runtime.Ship) {
	if playerID == "" || player == nil {
		return
	}

	cameraView, ok := game.cameraViews[playerID]
	if !ok || cameraView == nil {
		cameraView = &runtime.CameraView{}
		game.cameraViews[playerID] = cameraView
	}

	cameraView.X = player.X
	cameraView.Y = player.Y

	// Prefer an existing valid config to avoid flicker. Otherwise seed from session/player.
	cameraConfig := cameraView.Config
	if cameraConfig.VisibleWorldWidth <= 0 || cameraConfig.VisibleWorldHeight <= 0 {
		if session, ok := game.playerSessions[playerID]; ok && session != nil {
			if session.Config.VisibleWorldWidth > 0 && session.Config.VisibleWorldHeight > 0 {
				cameraConfig = session.Config
			}
		}
		if cameraConfig.VisibleWorldWidth <= 0 || cameraConfig.VisibleWorldHeight <= 0 {
			if player.Config.VisibleWorldWidth > 0 && player.Config.VisibleWorldHeight > 0 {
				cameraConfig = player.Config
			}
		}
		if cameraConfig.VisibleWorldWidth <= 0 || cameraConfig.VisibleWorldHeight <= 0 {
			cameraConfig.VisibleWorldWidth = constants.WorldWidth
			cameraConfig.VisibleWorldHeight = constants.WorldHeight
		}
	}
	cameraView.Config = cameraConfig
}

func (game *Game) RemovePlayer(playerID string) {
	game.mu.Lock()
	defer game.mu.Unlock()

	delete(game.entities.Players, playerID)
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
	if player, ok := game.entities.Players[playerID]; ok {
		return player.Lives
	}

	return 0
}
