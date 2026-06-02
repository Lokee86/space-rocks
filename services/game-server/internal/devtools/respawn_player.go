package devtools

import (
	"github.com/Lokee86/space-rocks/server/internal/game"
	"github.com/Lokee86/space-rocks/server/internal/game/physics"
)

func resolveDebugRespawnTargetPlayerID(request RespawnPlayerRequest) (string, bool) {
	if request.TargetPlayerID == "" {
		return "", false
	}
	return NormalizeDebugSpawnPlayerID(request.TargetPlayerID)
}

func applyDebugRespawnPlayer(target *game.Game, request RespawnPlayerRequest) (string, physics.Vector2, bool) {
	if target == nil {
		return "", physics.Vector2{}, false
	}

	playerID, ok := resolveDebugRespawnTargetPlayerID(request)
	if !ok {
		return "", physics.Vector2{}, false
	}

	spawnPosition, ok := target.DevtoolsSafeRespawnPosition(playerID)
	if !ok {
		return "", physics.Vector2{}, false
	}

	if !target.DevtoolsForceRespawnPlayer(playerID, spawnPosition, DummyPlayerCameraConfig()) {
		return "", physics.Vector2{}, false
	}

	return playerID, spawnPosition, true
}
