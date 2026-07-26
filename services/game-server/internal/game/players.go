package game

import (
	"fmt"

	"github.com/Lokee86/space-rocks/services/game-server/internal/game/runtime"
)

func (game *Game) AddPlayer() string {
	game.mu.Lock()
	defer game.mu.Unlock()

	return game.addPlayerLocked()
}

func (game *Game) addPlayerLocked() string {
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
	game.publishPresentationFrameLocked()

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
			cameraConfig = runtime.DefaultCameraConfig()
		}
	}
	cameraView.Config = runtime.ClampCameraConfig(cameraConfig)
}

func (game *Game) RemovePlayer(playerID string) {
	game.mu.Lock()
	defer game.mu.Unlock()

	delete(game.entities.Players, playerID)
	game.inputMu.Lock()
	delete(game.pendingPlayerInputs, playerID)
	game.inputMu.Unlock()
	delete(game.cameraViews, playerID)
	delete(game.playerSessions, playerID)
	delete(game.botControllers, playerID)
	game.clearTargetsForMissingPlayersLocked()
	delete(game.pendingPresentationEvents, playerID)
	game.publishPresentationFrameLocked()
}

func (game *Game) playerLives(playerID string) int {
	if session, ok := game.playerSessions[playerID]; ok {
		return session.Lives
	}

	return 0
}
