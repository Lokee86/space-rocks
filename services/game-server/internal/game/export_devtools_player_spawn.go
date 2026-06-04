package game

import (
	"sort"
	"strconv"
	"strings"

	"github.com/Lokee86/space-rocks/server/internal/game/runtime"
	"github.com/Lokee86/space-rocks/server/internal/game/physics"
)

func (game *Game) DevtoolsEnsurePlayerSession(playerID string, spawnPosition physics.Vector2) bool {
	return game.ensureDevtoolsPlayerSession(playerID, spawnPosition) != nil
}

func (game *Game) DevtoolsSpawnPlayerShip(playerID string, spawnPosition physics.Vector2, cameraConfig runtime.ClientConfig) bool {
	session, ok := game.playerSessions[playerID]
	if !ok || session == nil {
		return false
	}
	return game.applyDevtoolsPlayerShip(playerID, session, spawnPosition, cameraConfig)
}

func (game *Game) DevtoolsPlayerIDOccupied(playerID string) bool {
	return game.isDevtoolsPlayerIDOccupied(playerID)
}

func (game *Game) DevtoolsReservePlayerID(playerID string) bool {
	return game.reserveDevtoolsPlayerID(playerID)
}

func (game *Game) DevtoolsTargetPlayerIDs() []string {
	playerIDs := make(map[string]struct{}, len(game.playerSessions)+len(game.entities.Players))
	for playerID := range game.playerSessions {
		if playerID == "" {
			continue
		}
		playerIDs[playerID] = struct{}{}
	}
	for playerID := range game.entities.Players {
		if playerID == "" {
			continue
		}
		playerIDs[playerID] = struct{}{}
	}

	ids := make([]string, 0, len(playerIDs))
	for playerID := range playerIDs {
		ids = append(ids, playerID)
	}
	sort.Strings(ids)
	return ids
}

func (game *Game) ensureDevtoolsPlayerSession(playerID string, spawnPosition physics.Vector2) *playerSession {
	if playerID == "" {
		return nil
	}

	session := newPlayerSession(playerID, spawnPosition)
	game.playerSessions[playerID] = session
	return session
}

func (game *Game) applyDevtoolsPlayerShip(playerID string, session *playerSession, spawnPosition physics.Vector2, cameraConfig runtime.ClientConfig) bool {
	if playerID == "" {
		return false
	}
	if session == nil {
		return false
	}

	session.RespawnCooldown = 0
	player := session.NewShip(spawnPosition)
	game.entities.Players[playerID] = player
	game.ensureDevtoolsPlayerCameraView(playerID, player, cameraConfig)

	return true
}

func (game *Game) ensureDevtoolsPlayerCameraView(playerID string, player *runtime.Ship, cameraConfig runtime.ClientConfig) {
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
	if cameraConfig.VisibleWorldWidth > 0 && cameraConfig.VisibleWorldHeight > 0 {
		cameraView.Config = cameraConfig
	}
}

func (game *Game) isDevtoolsPlayerIDOccupied(playerID string) bool {
	normalizedRequestedID, ok := normalizeDevtoolsSpawnPlayerID(playerID)
	if !ok {
		return true
	}

	for existingPlayerID := range game.playerSessions {
		normalizedExistingID, normalized := normalizeDevtoolsSpawnPlayerID(existingPlayerID)
		if !normalized {
			continue
		}
		if normalizedExistingID == normalizedRequestedID {
			return true
		}
	}

	for existingPlayerID := range game.entities.Players {
		normalizedExistingID, normalized := normalizeDevtoolsSpawnPlayerID(existingPlayerID)
		if !normalized {
			continue
		}
		if normalizedExistingID == normalizedRequestedID {
			return true
		}
	}

	return false
}

func (game *Game) reserveDevtoolsPlayerID(playerID string) bool {
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

	normalizedID, ok := normalizeDevtoolsSpawnPlayerID(playerID)
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

func normalizeDevtoolsSpawnPlayerID(playerID string) (string, bool) {
	trimmed := strings.TrimSpace(playerID)
	parts := strings.Split(trimmed, "-")
	if len(parts) != 2 {
		return "", false
	}
	if parts[0] != "player" && parts[0] != "Player" {
		return "", false
	}
	number, err := strconv.Atoi(parts[1])
	if err != nil || number <= 0 {
		return "", false
	}
	return "player-" + strconv.Itoa(number), true
}
