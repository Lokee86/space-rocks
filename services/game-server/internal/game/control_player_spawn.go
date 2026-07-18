package game

import (
	"strconv"
	"strings"

	"github.com/Lokee86/space-rocks/services/game-server/internal/game/lives"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/physics"
	playerstate "github.com/Lokee86/space-rocks/services/game-server/internal/game/player"
	runtimepkg "github.com/Lokee86/space-rocks/services/game-server/internal/game/runtime"
)

func (target *Control) EnsurePlayerSession(playerID string, spawnPosition physics.Vector2) bool {
	target.game.mu.Lock()
	defer target.game.mu.Unlock()
	if playerID == "" {
		return false
	}
	if err := target.game.lifeRuntime.RegisterParticipant(lives.ParticipantRegistration{PlayerID: playerID}); err != nil {
		return false
	}
	target.game.playerSessions[playerID] = newPlayerSession(playerID, spawnPosition)
	target.game.publishPresentationFrameLocked()
	return true
}

func (target *Control) SpawnPlayerShip(playerID string, spawnPosition physics.Vector2, cameraConfig runtimepkg.ClientConfig) bool {
	target.game.mu.Lock()
	defer target.game.mu.Unlock()
	session, ok := target.game.playerSessions[playerID]
	if !ok || session == nil {
		return false
	}
	status, registered := target.game.lifeRuntime.Status(playerID)
	if !registered || status != playerstate.StatusActive {
		return false
	}
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
	target.game.publishPresentationFrameLocked()
	return true
}

func (target *Control) EnableBotPlayer(playerID string) bool {
	target.game.mu.Lock()
	defer target.game.mu.Unlock()
	return target.game.enableBotPlayerLocked(playerID)
}

func (target *Control) PlayerIDOccupied(playerID string) bool {
	target.game.mu.Lock()
	defer target.game.mu.Unlock()
	return target.playerIDOccupiedLocked(playerID)
}

func (target *Control) playerIDOccupiedLocked(playerID string) bool {
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
	target.game.mu.Lock()
	defer target.game.mu.Unlock()
	normalizedPlayerID, ok := normalizeControlPlayerID(playerID)
	if !ok {
		return false
	}
	if target.playerIDOccupiedLocked(normalizedPlayerID) {
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
