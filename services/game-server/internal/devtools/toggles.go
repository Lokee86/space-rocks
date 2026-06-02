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
	targetPlayerIDs := resolveCommandTargetPlayerIDs(target, playerID, command)
	if command.TargetScope == targetScopeAllPlayers {
		setDebugInvincibleForPlayers(target, playerID, targetPlayerIDs, nextAllPlayersToggleState(targetPlayerIDs, target.DevtoolsPlayerInvincible))
		return true
	}

	for _, targetPlayerID := range targetPlayerIDs {
		toggleDebugInvincibleForPlayer(target, playerID, targetPlayerID)
	}
	return true
}

func toggleDebugInvincibleForPlayer(target *game.Game, playerID string, targetPlayerID string) {
	current, _ := target.DevtoolsPlayerInvincible(targetPlayerID)
	enabled := !current
	setDebugInvincibleForPlayer(target, playerID, targetPlayerID, enabled)
}

func setDebugInvincibleForPlayers(target *game.Game, playerID string, targetPlayerIDs []string, enabled bool) {
	for _, targetPlayerID := range targetPlayerIDs {
		setDebugInvincibleForPlayer(target, playerID, targetPlayerID, enabled)
	}
}

func setDebugInvincibleForPlayer(target *game.Game, playerID string, targetPlayerID string, enabled bool) {
	target.DevtoolsSetPlayerInvincible(targetPlayerID, enabled)
	logging.Game.Info("debug invincibility set",
		logging.FieldPlayerID, playerID,
		"target_player_id", targetPlayerID,
		"enabled", enabled,
	)
}

func handleToggleDebugInfiniteLives(target *game.Game, playerID string, command DebugCommand) bool {
	targetPlayerIDs := resolveCommandTargetPlayerIDs(target, playerID, command)
	if command.TargetScope == targetScopeAllPlayers {
		setDebugInfiniteLivesForPlayers(target, playerID, targetPlayerIDs, nextAllPlayersToggleState(targetPlayerIDs, target.DevtoolsInfiniteLives))
		return true
	}

	for _, targetPlayerID := range targetPlayerIDs {
		toggleDebugInfiniteLivesForPlayer(target, playerID, targetPlayerID)
	}
	return true
}

func toggleDebugInfiniteLivesForPlayer(target *game.Game, playerID string, targetPlayerID string) {
	current, _ := target.DevtoolsInfiniteLives(targetPlayerID)
	enabled := !current
	setDebugInfiniteLivesForPlayer(target, playerID, targetPlayerID, enabled)
}

func setDebugInfiniteLivesForPlayers(target *game.Game, playerID string, targetPlayerIDs []string, enabled bool) {
	for _, targetPlayerID := range targetPlayerIDs {
		setDebugInfiniteLivesForPlayer(target, playerID, targetPlayerID, enabled)
	}
}

func setDebugInfiniteLivesForPlayer(target *game.Game, playerID string, targetPlayerID string, enabled bool) {
	target.DevtoolsSetInfiniteLives(targetPlayerID, enabled)
	logging.Game.Info("debug infinite lives set",
		logging.FieldPlayerID, playerID,
		"target_player_id", targetPlayerID,
		"enabled", enabled,
	)
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
	targetPlayerIDs := resolveCommandTargetPlayerIDs(target, playerID, command)
	if command.TargetScope == targetScopeAllPlayers {
		setDebugFreezePlayerForPlayers(target, playerID, targetPlayerIDs, nextAllPlayersToggleState(targetPlayerIDs, target.DevtoolsPlayerFrozen))
		return true
	}

	for _, targetPlayerID := range targetPlayerIDs {
		toggleDebugFreezePlayerForPlayer(target, playerID, targetPlayerID)
	}
	return true
}

func toggleDebugFreezePlayerForPlayer(target *game.Game, playerID string, targetPlayerID string) {
	current, _ := target.DevtoolsPlayerFrozen(targetPlayerID)
	enabled := !current
	setDebugFreezePlayerForPlayer(target, playerID, targetPlayerID, enabled)
}

func setDebugFreezePlayerForPlayers(target *game.Game, playerID string, targetPlayerIDs []string, enabled bool) {
	for _, targetPlayerID := range targetPlayerIDs {
		setDebugFreezePlayerForPlayer(target, playerID, targetPlayerID, enabled)
	}
}

func setDebugFreezePlayerForPlayer(target *game.Game, playerID string, targetPlayerID string, enabled bool) {
	target.DevtoolsSetPlayerFrozen(targetPlayerID, enabled)
	logging.Game.Info("debug player freeze set",
		logging.FieldPlayerID, playerID,
		"target_player_id", targetPlayerID,
		"enabled", enabled,
	)
}

func nextAllPlayersToggleState(targetPlayerIDs []string, status func(string) (bool, bool)) bool {
	if len(targetPlayerIDs) == 0 {
		return false
	}

	for _, targetPlayerID := range targetPlayerIDs {
		enabled, found := status(targetPlayerID)
		if !found || !enabled {
			return true
		}
	}

	return false
}

func handleDebugKillPlayer(target *game.Game, playerID string, command DebugCommand) bool {
	if target == nil {
		return true
	}

	for _, targetPlayerID := range resolveCommandTargetPlayerIDs(target, playerID, command) {
		killDebugPlayerTarget(target, playerID, targetPlayerID)
	}
	return true
}

func killDebugPlayerTarget(target *game.Game, playerID string, targetPlayerID string) {
	isPlayerAlive := false
	for _, player := range target.MatchDecision().Players {
		if player.ID == targetPlayerID {
			isPlayerAlive = player.Status == "active"
			break
		}
	}

	if !isPlayerAlive {
		return
	}

	target.DevtoolsKillPlayer(playerID, targetPlayerID)
}
