package game

import (
	"strconv"
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
	cameraView := target.game.cameraViews[playerID]
	if cameraView == nil {
		cameraView = &runtimepkg.CameraView{}
		target.game.cameraViews[playerID] = cameraView
	}
	cameraView.X = player.X
	cameraView.Y = player.Y
	if cameraConfig.VisibleWorldWidth > 0 && cameraConfig.VisibleWorldHeight > 0 {
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
	normalizedPlayerID, ok := normalizeControlPlayerID(playerID)
	if !ok {
		return false
	}
	if target.PlayerIDOccupied(normalizedPlayerID) {
		return false
	}
	nextID := normalizedControlPlayerNumber(normalizedPlayerID)
	if nextID > target.game.nextID {
		target.game.nextID = nextID
	}
	return true
}

func normalizeControlPlayerID(playerID string) (string, bool) {
	normalized := strings.TrimSpace(playerID)
	if len(normalized) < 8 {
		return "", false
	}
	if normalized[:7] != "player-" && normalized[:7] != "Player-" {
		return "", false
	}
	number, err := strconv.Atoi(normalized[7:])
	if err != nil || number <= 0 {
		return "", false
	}
	return "player-" + strconv.Itoa(number), true
}

func normalizedControlPlayerNumber(playerID string) int {
	_, digits, _ := strings.Cut(playerID, "-")
	number, _ := strconv.Atoi(digits)
	return number
}
