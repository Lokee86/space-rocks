package devtools

import (
	"github.com/Lokee86/space-rocks/server/internal/game"
	"github.com/Lokee86/space-rocks/server/internal/logging"
)

const (
	freezeTargetAll        = "all"
	freezeTargetAsteroids  = "asteroids"
	freezeTargetBullets    = "bullets"
	freezeTargetSpawning   = "spawning"
	freezeTargetSpawns     = "spawns"
	freezeTargetCollisions = "collisions"
)

func freezeTargetFromCommand(command DebugCommand) string {
	if command.FreezeTarget == "" {
		return freezeTargetAll
	}
	return command.FreezeTarget
}

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

func handleToggleDebugFreezeWorld(target *game.Game, playerID string, command DebugCommand) bool {
	freezeTarget := freezeTargetFromCommand(command)

	if freezeTarget == freezeTargetAll {
		enabled := target.DevtoolsToggleFreezeWorld()
		logging.Game.Info("debug world freeze toggled",
			logging.FieldPlayerID, playerID,
			"freeze_target", freezeTarget,
			"enabled", enabled,
		)
		return true
	}

	enabled := false
	switch freezeTarget {
	case freezeTargetAsteroids:
		enabled = target.DevtoolsToggleFreezeAsteroids()
	case freezeTargetBullets:
		enabled = target.DevtoolsToggleFreezeBullets()
	case freezeTargetSpawning, freezeTargetSpawns:
		enabled = target.DevtoolsToggleFreezeSpawning()
	case freezeTargetCollisions:
		enabled = target.DevtoolsToggleFreezeCollisions()
	default:
		logging.Game.Info("debug world freeze target ignored",
			logging.FieldPlayerID, playerID,
			"freeze_target", freezeTarget,
		)
		return true
	}

	logging.Game.Info("debug world freeze toggled",
		logging.FieldPlayerID, playerID,
		"freeze_target", freezeTarget,
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
