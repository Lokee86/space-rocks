package devtools

import (
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/physics"
)

func resolveDebugRespawnTargetPlayerID(request RespawnPlayerRequest) (string, bool) {
	if request.TargetPlayerID == "" {
		return "", false
	}
	return NormalizeDebugSpawnPlayerID(request.TargetPlayerID)
}

func applyDebugRespawnPlayer(target Target, request RespawnPlayerRequest) (string, physics.Vector2, bool) {
	if target == nil {
		return "", physics.Vector2{}, false
	}

	playerID, ok := resolveDebugRespawnTargetPlayerID(request)
	if !ok {
		return "", physics.Vector2{}, false
	}

	spawnPosition, ok := target.SafeRespawnPosition(playerID)
	if !ok {
		return "", physics.Vector2{}, false
	}

	if !target.ForceRespawnPlayer(playerID, spawnPosition, DummyPlayerCameraConfig()) {
		return "", physics.Vector2{}, false
	}

	return playerID, spawnPosition, true
}
