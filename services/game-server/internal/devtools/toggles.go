package devtools

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

func handleToggleDebugInvincible(target Target, playerID string, command DebugCommand) bool {
	targetPlayerIDs := resolveCommandTargetPlayerIDs(target, playerID, command)
	if command.TargetScope == targetScopeAllPlayers {
		setDebugInvincibleForPlayers(target, playerID, targetPlayerIDs, nextAllPlayersToggleState(targetPlayerIDs, target.PlayerInvincible))
		return true
	}

	for _, targetPlayerID := range targetPlayerIDs {
		toggleDebugInvincibleForPlayer(target, playerID, targetPlayerID)
	}
	return true
}

func toggleDebugInvincibleForPlayer(target Target, playerID string, targetPlayerID string) {
	current, _ := target.PlayerInvincible(targetPlayerID)
	enabled := !current
	setDebugInvincibleForPlayer(target, playerID, targetPlayerID, enabled)
}

func setDebugInvincibleForPlayers(target Target, playerID string, targetPlayerIDs []string, enabled bool) {
	for _, targetPlayerID := range targetPlayerIDs {
		setDebugInvincibleForPlayer(target, playerID, targetPlayerID, enabled)
	}
}

func setDebugInvincibleForPlayer(target Target, playerID string, targetPlayerID string, enabled bool) {
	target.SetPlayerInvincible(targetPlayerID, enabled)
}

func handleToggleDebugInfiniteLives(target Target, playerID string, command DebugCommand) bool {
	targetPlayerIDs := resolveCommandTargetPlayerIDs(target, playerID, command)
	if command.TargetScope == targetScopeAllPlayers {
		setDebugInfiniteLivesForPlayers(target, playerID, targetPlayerIDs, nextAllPlayersToggleState(targetPlayerIDs, target.InfiniteLives))
		return true
	}

	for _, targetPlayerID := range targetPlayerIDs {
		toggleDebugInfiniteLivesForPlayer(target, playerID, targetPlayerID)
	}
	return true
}

func toggleDebugInfiniteLivesForPlayer(target Target, playerID string, targetPlayerID string) {
	current, _ := target.InfiniteLives(targetPlayerID)
	enabled := !current
	setDebugInfiniteLivesForPlayer(target, playerID, targetPlayerID, enabled)
}

func setDebugInfiniteLivesForPlayers(target Target, playerID string, targetPlayerIDs []string, enabled bool) {
	for _, targetPlayerID := range targetPlayerIDs {
		setDebugInfiniteLivesForPlayer(target, playerID, targetPlayerID, enabled)
	}
}

func setDebugInfiniteLivesForPlayer(target Target, playerID string, targetPlayerID string, enabled bool) {
	target.SetInfiniteLives(targetPlayerID, enabled)
}

func handleToggleDebugFreezeWorld(target Target, playerID string, command DebugCommand) bool {
	freezeTarget := freezeTargetFromCommand(command)

	if freezeTarget == freezeTargetAll {
		target.ToggleFreezeWorld()
		return true
	}

	switch freezeTarget {
	case freezeTargetAsteroids:
		target.ToggleFreezeAsteroids()
	case freezeTargetBullets:
		target.ToggleFreezeBullets()
	case freezeTargetSpawning, freezeTargetSpawns:
		target.ToggleFreezeSpawning()
	case freezeTargetCollisions:
		target.ToggleFreezeCollisions()
	default:
		return true
	}

	return true
}

func handleToggleDebugFreezePlayer(target Target, playerID string, command DebugCommand) bool {
	targetPlayerIDs := resolveCommandTargetPlayerIDs(target, playerID, command)
	if command.TargetScope == targetScopeAllPlayers {
		setDebugFreezePlayerForPlayers(target, playerID, targetPlayerIDs, nextAllPlayersToggleState(targetPlayerIDs, target.PlayerFrozen))
		return true
	}

	for _, targetPlayerID := range targetPlayerIDs {
		toggleDebugFreezePlayerForPlayer(target, playerID, targetPlayerID)
	}
	return true
}

func toggleDebugFreezePlayerForPlayer(target Target, playerID string, targetPlayerID string) {
	current, _ := target.PlayerFrozen(targetPlayerID)
	enabled := !current
	setDebugFreezePlayerForPlayer(target, playerID, targetPlayerID, enabled)
}

func setDebugFreezePlayerForPlayers(target Target, playerID string, targetPlayerIDs []string, enabled bool) {
	for _, targetPlayerID := range targetPlayerIDs {
		setDebugFreezePlayerForPlayer(target, playerID, targetPlayerID, enabled)
	}
}

func setDebugFreezePlayerForPlayer(target Target, playerID string, targetPlayerID string, enabled bool) {
	target.SetPlayerFrozen(targetPlayerID, enabled)
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

func handleDebugKillPlayer(target Target, playerID string, command DebugCommand) bool {
	if target == nil {
		return true
	}

	for _, targetPlayerID := range resolveCommandTargetPlayerIDs(target, playerID, command) {
		killDebugPlayerTarget(target, playerID, targetPlayerID)
	}
	return true
}

func killDebugPlayerTarget(target Target, playerID string, targetPlayerID string) bool {
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

	if !target.ApplyPlayerDefeat(playerID, targetPlayerID) {
		return false
	}
	return true
}
