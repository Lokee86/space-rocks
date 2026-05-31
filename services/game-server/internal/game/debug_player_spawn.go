package game

import (
	"strconv"
	"strings"

	"github.com/Lokee86/space-rocks/server/internal/devtools"
	"github.com/Lokee86/space-rocks/server/internal/game/entities"
	"github.com/Lokee86/space-rocks/server/internal/game/physics"
)

func (game *Game) ensureDebugPlayerSession(playerID string, spawnPosition physics.Vector2) *playerSession {
	if playerID == "" {
		return nil
	}

	session := newPlayerSession(playerID, spawnPosition)
	game.playerSessions[playerID] = session
	return session
}

func (game *Game) applyDebugPlayerShip(playerID string, session *playerSession, spawnPosition physics.Vector2) bool {
	if playerID == "" {
		return false
	}
	if session == nil {
		return false
	}

	session.RespawnCooldown = 0
	player := session.NewShip(spawnPosition)
	game.state.Players[playerID] = player
	game.ensureDebugPlayerCameraView(playerID, player)

	return true
}

func (game *Game) ensureDebugPlayerCameraView(playerID string, player *entities.Ship) {
	if playerID == "" {
		return
	}
	if player == nil {
		return
	}

	cameraView, ok := game.cameraViews[playerID]
	if !ok || cameraView == nil {
		cameraView = &entities.CameraView{}
		game.cameraViews[playerID] = cameraView
	}

	cameraView.X = player.X
	cameraView.Y = player.Y
	cameraView.Config = player.Config
}


func (game *Game) isDebugGameplayPlayerIDOccupied(playerID string) bool {
	normalizedRequestedID, ok := devtools.NormalizeDebugSpawnPlayerID(playerID)
	if !ok {
		return true
	}

	for existingPlayerID := range game.playerSessions {
		normalizedExistingID, normalized := devtools.NormalizeDebugSpawnPlayerID(existingPlayerID)
		if !normalized {
			continue
		}
		if normalizedExistingID == normalizedRequestedID {
			return true
		}
	}

	for existingPlayerID := range game.state.Players {
		normalizedExistingID, normalized := devtools.NormalizeDebugSpawnPlayerID(existingPlayerID)
		if !normalized {
			continue
		}
		if normalizedExistingID == normalizedRequestedID {
			return true
		}
	}

	return false
}

func (game *Game) reserveDebugGameplayPlayerID(playerID string) bool {
	trimmed := strings.TrimSpace(playerID)
	parts := strings.Split(trimmed, "-")
	if len(parts) != 2 {
		return false
	}

	if parts[0] != "player" && parts[0] != "Player" {
		return false
	}

	number, err := strconv.Atoi(parts[1])
	if err != nil || number <= 0 {
		return false
	}

	normalizedID, ok := devtools.NormalizeDebugSpawnPlayerID(playerID)
	if !ok {
		return false
	}
	if normalizedID == "" {
		return false
	}

	if number > game.nextID {
		game.nextID = number
	}

	return true
}
