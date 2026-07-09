package game

import (
	"strings"

	"github.com/Lokee86/space-rocks/server/internal/game/physics"
	runtimepkg "github.com/Lokee86/space-rocks/server/internal/game/runtime"
)

func (target *Control) EnsurePlayerSession(playerID string, spawnPosition physics.Vector2) bool {
	if playerID == "" {
		return false
	}
	target.game.playerSessions[playerID] = newPlayerSession(playerID, spawnPosition)
	return true
}

func (target *Control) SpawnPlayerShip(playerID string, spawnPosition physics.Vector2, cameraConfig runtimepkg.ClientConfig) bool {
	session, ok := target.game.playerSessions[playerID]
	if !ok || session == nil {
		return false
	}
	session.RespawnCooldown = 0
	player := session.NewShip(spawnPosition)
	target.game.entities.Players[playerID] = player
	target.game.setPlayerCameraViewLocked(playerID, player)
	cameraView := target.game.cameraViews[playerID]
	if cameraView != nil {
		cameraView.Config = cameraConfig
	}
	return true
}

func (target *Control) PlayerIDOccupied(playerID string) bool {
	normalizedRequestedID, ok := normalizeControlPlayerID(playerID)
	if !ok {
		return true
	}

	for existingPlayerID := range target.game.playerSessions {
		normalizedExistingID, normalized := normalizeControlPlayerID(existingPlayerID)
		if !normalized {
			continue
		}
		if normalizedExistingID == normalizedRequestedID {
			return true
		}
	}

	for existingPlayerID := range target.game.entities.Players {
		normalizedExistingID, normalized := normalizeControlPlayerID(existingPlayerID)
		if !normalized {
			continue
		}
		if normalizedExistingID == normalizedRequestedID {
			return true
		}
	}

	return false
}

func (target *Control) ReservePlayerID(playerID string) bool {
	if target.PlayerIDOccupied(playerID) {
		return false
	}
	return true
}

func normalizeControlPlayerID(playerID string) (string, bool) {
	normalized := strings.TrimSpace(playerID)
	if normalized == "" {
		return "", false
	}
	return strings.ToLower(normalized), true
}
