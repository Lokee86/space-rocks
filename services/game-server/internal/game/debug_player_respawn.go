package game

import (
	"log"

	"github.com/Lokee86/space-rocks/server/internal/game/debugging"
)

func (game *Game) applyDebugRespawnPlayer(request debugging.RespawnPlayerRequest) bool {
	if request.TargetPlayerID == "" {
		return false
	}

	playerID, ok := normalizeDebugSpawnPlayerID(request.TargetPlayerID)
	if !ok {
		return false
	}

	session, ok := game.playerSessions[playerID]
	if !ok {
		return false
	}

	spawnPosition := game.safeRespawnPosition(session)
	session.RespawnCooldown = 0
	player := session.NewShip(spawnPosition)
	game.state.Players[playerID] = player

	cameraView := game.cameraViews[playerID]
	cameraView.X = player.X
	cameraView.Y = player.Y
	cameraView.Config = player.Config
	game.cameraViews[playerID] = cameraView

	log.Printf("debug force respawn applied for player %s at (%.2f, %.2f)", playerID, player.X, player.Y)
	return true
}
