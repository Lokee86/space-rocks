package devtools

import (
	"github.com/Lokee86/space-rocks/server/internal/game"
	"github.com/Lokee86/space-rocks/server/internal/logging"
)

func handleToggleDebugInvincible(target *game.Game, playerID string, command DebugCommand) bool {
	targetPlayerID := command.TargetPlayerID
	if targetPlayerID == "" {
		targetPlayerID = playerID
	}

	current, _ := target.DevtoolsPlayerInvincible(targetPlayerID)
	enabled := !current
	target.DevtoolsSetPlayerInvincible(targetPlayerID, enabled)
	logging.Game.Info("debug invincibility toggled",
		logging.FieldPlayerID, playerID,
		"target_player_id", targetPlayerID,
		"enabled", enabled,
	)
	return true
}

func handleToggleDebugInfiniteLives(target *game.Game, playerID string, command DebugCommand) bool {
	targetPlayerID := command.TargetPlayerID
	if targetPlayerID == "" {
		targetPlayerID = playerID
	}

	current, _ := target.DevtoolsInfiniteLives(targetPlayerID)
	enabled := !current
	target.DevtoolsSetInfiniteLives(targetPlayerID, enabled)
	logging.Game.Info("debug infinite lives toggled",
		logging.FieldPlayerID, playerID,
		"target_player_id", targetPlayerID,
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

func handleToggleDebugFreezePlayer(target *game.Game, playerID string, command DebugCommand) bool {
	targetPlayerID := command.TargetPlayerID
	if targetPlayerID == "" {
		targetPlayerID = playerID
	}

	current, _ := target.DevtoolsPlayerFrozen(targetPlayerID)
	enabled := !current
	target.DevtoolsSetPlayerFrozen(targetPlayerID, enabled)
	logging.Game.Info("debug player freeze toggled",
		logging.FieldPlayerID, playerID,
		"target_player_id", targetPlayerID,
		"enabled", enabled,
	)
	return true
}

func handleDebugKillPlayer(target *game.Game, playerID string, command DebugCommand) bool {
	if target == nil {
		return true
	}

	targetPlayerID := command.TargetPlayerID
	if targetPlayerID == "" {
		targetPlayerID = playerID
	}

	isPlayerAlive := false
	for _, player := range target.MatchDecision().Players {
		if player.ID == targetPlayerID {
			isPlayerAlive = player.Status == "active"
			break
		}
	}

	if !isPlayerAlive {
		return true
	}

	target.DevtoolsKillPlayer(playerID, targetPlayerID)
	return true
}
