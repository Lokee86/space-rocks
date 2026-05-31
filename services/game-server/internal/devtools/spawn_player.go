package devtools

import (
	"github.com/Lokee86/space-rocks/server/internal/game"
	"github.com/Lokee86/space-rocks/server/internal/game/physics"
	"github.com/Lokee86/space-rocks/server/internal/game/space"
)

func debugPlayerSpawnPosition(request SpawnEntityRequest) physics.Vector2 {
	return space.NormalizePosition(request.Position())
}

func resolveDebugSpawnPlayerID(target *game.Game, request SpawnEntityRequest) (string, bool) {
	if target == nil {
		return "", false
	}

	if request.TargetPlayerID != "" {
		normalizedID, ok := NormalizeDebugSpawnPlayerID(request.TargetPlayerID)
		if !ok {
			return "", false
		}
		if !target.DevtoolsReservePlayerID(normalizedID) {
			return "", false
		}
		return normalizedID, true
	}

	return AllocateDebugGameplayPlayerID(
		target.DevtoolsPlayerIDOccupied,
		target.DevtoolsReservePlayerID,
	)
}

func applyDebugSpawnPlayer(target *game.Game, request SpawnEntityRequest) (string, physics.Vector2, bool) {
	playerID, ok := resolveDebugSpawnPlayerID(target, request)
	if !ok {
		return "", physics.Vector2{}, false
	}

	spawnPosition := debugPlayerSpawnPosition(request)
	if !target.DevtoolsEnsurePlayerSession(playerID, spawnPosition) {
		return "", physics.Vector2{}, false
	}
	if !target.DevtoolsSpawnPlayerShip(playerID, spawnPosition) {
		return "", physics.Vector2{}, false
	}

	return playerID, spawnPosition, true
}
