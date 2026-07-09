package devtools

import (
	"log/slog"

	"github.com/Lokee86/space-rocks/server/internal/devtools/streamruntime"
)

type clearEntitiesTarget interface {
	ClearTarget
	TargetPlayerIDs() []string
}

func handleDebugClearBullets(target ClearTarget, playerID string, command DebugCommand) bool {
	if target == nil {
		return false
	}

	target.ClearBullets()
	streamruntime.DefaultRuntime.ClearContinuousBulletStreams()
	return true
}

func handleDebugClearAsteroids(target ClearTarget, playerID string, command DebugCommand) bool {
	if target == nil {
		return false
	}

	target.ClearAsteroids()
	return true
}

func handleClearEntities(target clearEntitiesTarget, playerID string, command DebugCommand) bool {
	if target == nil {
		return false
	}

	cleared := false
	for _, targetPlayerID := range resolveCommandTargetPlayerIDs(target, playerID, command) {
		if clearEntitiesForPlayer(target, targetPlayerID) {
			cleared = true
		}
	}
	if !cleared {
		slog.Debug("devtools clear entities target not found", "source_player_id", playerID, "target_scope", command.TargetScope, "target_player_id", command.TargetPlayerID)
	}
	return cleared
}

func clearEntitiesForPlayer(target ClearTarget, targetPlayerID string) bool {
	clearedBullets := target.ClearBullets()
	clearedAsteroids := target.ClearAsteroids()
	return clearedBullets+clearedAsteroids > 0
}
