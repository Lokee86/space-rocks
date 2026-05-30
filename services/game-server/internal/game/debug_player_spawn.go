package game

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Lokee86/space-rocks/server/internal/game/debugging"
	"github.com/Lokee86/space-rocks/server/internal/game/entities"
	"github.com/Lokee86/space-rocks/server/internal/game/physics"
	"github.com/Lokee86/space-rocks/server/internal/game/space"
	"github.com/Lokee86/space-rocks/server/internal/playerids"
)

func (game *Game) debugPlayerSpawnPosition(request debugging.SpawnEntityRequest) physics.Vector2 {
	return space.NormalizePosition(request.Position())
}

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

func (game *Game) applyDebugSpawnPlayer(request debugging.SpawnEntityRequest) (string, bool) {
	playerID, ok := game.resolveDebugSpawnPlayerID(request)
	if !ok {
		return "", false
	}

	spawnPosition := game.debugPlayerSpawnPosition(request)
	session := game.ensureDebugPlayerSession(playerID, spawnPosition)
	if !game.applyDebugPlayerShip(playerID, session, spawnPosition) {
		return "", false
	}

	return playerID, true
}

func (game *Game) resolveDebugSpawnPlayerID(request debugging.SpawnEntityRequest) (string, bool) {
	if request.TargetPlayerID != "" {
		normalizedID, ok := normalizeDebugSpawnPlayerID(request.TargetPlayerID)
		if !ok {
			return "", false
		}
		if !game.reserveDebugGameplayPlayerID(normalizedID) {
			return "", false
		}
		return normalizedID, true
	}

	return game.allocateDebugGameplayPlayerID()
}

func formatDebugGamePlayerID(number int) string {
	return fmt.Sprintf("player-%d", number)
}

func parseDebugGamePlayerIDNumber(playerID string) (int, bool) {
	trimmed := strings.TrimSpace(playerID)
	parts := strings.Split(trimmed, "-")
	if len(parts) != 2 {
		return 0, false
	}

	if parts[0] != "player" && parts[0] != "Player" {
		return 0, false
	}

	number, err := strconv.Atoi(parts[1])
	if err != nil || number <= 0 {
		return 0, false
	}

	return number, true
}

func normalizeDebugSpawnPlayerID(playerID string) (string, bool) {
	number, ok := parseDebugGamePlayerIDNumber(playerID)
	if !ok {
		return "", false
	}

	return formatDebugGamePlayerID(number), true
}

func (game *Game) isDebugGameplayPlayerIDOccupied(playerID string) bool {
	normalizedRequestedID, ok := normalizeDebugSpawnPlayerID(playerID)
	if !ok {
		return true
	}

	for existingPlayerID := range game.playerSessions {
		normalizedExistingID, normalized := normalizeDebugSpawnPlayerID(existingPlayerID)
		if !normalized {
			continue
		}
		if normalizedExistingID == normalizedRequestedID {
			return true
		}
	}

	for existingPlayerID := range game.state.Players {
		normalizedExistingID, normalized := normalizeDebugSpawnPlayerID(existingPlayerID)
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
	number, ok := parseDebugGamePlayerIDNumber(playerID)
	if !ok {
		return false
	}

	if number > game.nextID {
		game.nextID = number
	}

	return true
}

func (game *Game) allocateDebugGameplayPlayerID() (string, bool) {
	for number := 1; number <= playerids.MaxPlayers; number++ {
		candidate := formatDebugGamePlayerID(number)
		if game.isDebugGameplayPlayerIDOccupied(candidate) {
			continue
		}
		if !game.reserveDebugGameplayPlayerID(candidate) {
			continue
		}
		return candidate, true
	}

	return "", false
}
