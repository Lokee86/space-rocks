package game

import (
	"github.com/Lokee86/space-rocks/server/internal/game/runtime"
	"github.com/Lokee86/space-rocks/server/internal/game/physics"
)

func (game *Game) DevtoolsSafeRespawnPosition(playerID string) (physics.Vector2, bool) {
	session, ok := game.playerSessions[playerID]
	if !ok || session == nil {
		return physics.Vector2{}, false
	}
	return game.safeRespawnPosition(session), true
}

func (game *Game) DevtoolsForceRespawnPlayer(playerID string, position physics.Vector2, cameraConfig runtime.ClientConfig) bool {
	session, ok := game.playerSessions[playerID]
	if !ok || session == nil {
		return false
	}

	session.RespawnCooldown = 0
	player := session.NewShip(position)
	game.state.Players[playerID] = player

	cameraView := game.cameraViews[playerID]
	if cameraView == nil {
		cameraView = &runtime.CameraView{}
		game.cameraViews[playerID] = cameraView
	}
	cameraView.X = player.X
	cameraView.Y = player.Y
	cameraView.Config = cameraConfig

	return true
}
