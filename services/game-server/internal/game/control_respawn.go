package game

import (
	"github.com/Lokee86/space-rocks/server/internal/game/physics"
	runtimepkg "github.com/Lokee86/space-rocks/server/internal/game/runtime"
)

func (target *Control) SafeRespawnPosition(playerID string) (physics.Vector2, bool) {
	session, ok := target.game.playerSessions[playerID]
	if !ok || session == nil {
		return physics.Vector2{}, false
	}
	return target.game.safeRespawnPosition(session), true
}

func (target *Control) ForceRespawnPlayer(playerID string, position physics.Vector2, cameraConfig runtimepkg.ClientConfig) bool {
	session, ok := target.game.playerSessions[playerID]
	if !ok || session == nil {
		return false
	}

	session.RespawnCooldown = 0
	player := session.NewShip(position)
	target.game.entities.Players[playerID] = player

	cameraView := target.game.cameraViews[playerID]
	if cameraView == nil {
		cameraView = &runtimepkg.CameraView{}
		target.game.cameraViews[playerID] = cameraView
	}
	cameraView.X = player.X
	cameraView.Y = player.Y
	cameraView.Config = cameraConfig

	return true
}
