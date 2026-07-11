package devtools

import (
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/physics"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/space"
)

func debugPlayerSpawnPosition(request SpawnEntityRequest) physics.Vector2 {
	return space.NormalizePosition(request.Position())
}

func resolveDebugSpawnPlayerID(target Target, request SpawnEntityRequest) (string, bool) {
	if target == nil {
		return "", false
	}

	if request.TargetPlayerID != "" {
		normalizedID, ok := NormalizeDebugSpawnPlayerID(request.TargetPlayerID)
		if !ok {
			return "", false
		}
		if !target.ReservePlayerID(normalizedID) {
			return "", false
		}
		return normalizedID, true
	}

	return AllocateDebugGameplayPlayerID(
		target.PlayerIDOccupied,
		target.ReservePlayerID,
	)
}

func applyDebugSpawnPlayer(target Target, request SpawnEntityRequest) (string, physics.Vector2, bool) {
	playerID, ok := resolveDebugSpawnPlayerID(target, request)
	if !ok {
		return "", physics.Vector2{}, false
	}

	spawnPosition := debugPlayerSpawnPosition(request)
	if !target.EnsurePlayerSession(playerID, spawnPosition) {
		return "", physics.Vector2{}, false
	}
	if !target.SpawnPlayerShip(playerID, spawnPosition, DummyPlayerCameraConfig()) {
		return "", physics.Vector2{}, false
	}

	return playerID, spawnPosition, true
}
