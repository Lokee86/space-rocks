package devtools

import (
	"github.com/Lokee86/space-rocks/server/internal/game"
	"github.com/Lokee86/space-rocks/server/internal/logging"
)

func handleToggleDebugInvincible(target *game.Game, playerID string) bool {
	current, _ := target.DevtoolsPlayerInvincible(playerID)
	enabled := !current
	target.DevtoolsSetPlayerInvincible(playerID, enabled)
	logging.Game.Info("debug invincibility toggled",
		logging.FieldPlayerID, playerID,
		"enabled", enabled,
	)
	return true
}

func handleToggleDebugInfiniteLives(target *game.Game, playerID string) bool {
	current, _ := target.DevtoolsInfiniteLives(playerID)
	enabled := !current
	target.DevtoolsSetInfiniteLives(playerID, enabled)
	logging.Game.Info("debug infinite lives toggled",
		logging.FieldPlayerID, playerID,
		"enabled", enabled,
	)
	return true
}

func handleToggleDebugFreezeWorld(target *game.Game, playerID string) bool {
	enabled := !target.DevtoolsWorldFrozen()
	target.DevtoolsSetWorldFrozen(enabled)
	logging.Game.Info("debug world freeze toggled",
		logging.FieldPlayerID, playerID,
		"enabled", enabled,
	)
	return true
}

func handleToggleDebugFreezePlayer(target *game.Game, playerID string) bool {
	current, _ := target.DevtoolsPlayerFrozen(playerID)
	enabled := !current
	target.DevtoolsSetPlayerFrozen(playerID, enabled)
	logging.Game.Info("debug player freeze toggled",
		logging.FieldPlayerID, playerID,
		"enabled", enabled,
	)
	return true
}

func handleDebugKillPlayer(target *game.Game, playerID string, command DebugCommand) bool {
	targetPlayerID := command.TargetPlayerID
	if targetPlayerID == "" {
		targetPlayerID = playerID
	}
	target.DevtoolsKillPlayer(playerID, targetPlayerID)
	return true
}
