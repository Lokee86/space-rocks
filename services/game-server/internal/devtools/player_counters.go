package devtools

type counterCommandTarget interface {
	CounterTarget
	TargetPlayerIDs() []string
}

func handleDebugSetScore(target counterCommandTarget, playerID string, command DebugCommand) bool {
	if target == nil {
		return false
	}

	foundAny := false
	for _, targetPlayerID := range resolveCommandTargetPlayerIDs(target, playerID, command) {
		if setDebugScoreForPlayer(target, targetPlayerID, command.Score) {
			foundAny = true
		}
	}
	return foundAny
}

func handleDebugAddScore(target counterCommandTarget, playerID string, command DebugCommand) bool {
	if target == nil {
		return false
	}

	foundAny := false
	for _, targetPlayerID := range resolveCommandTargetPlayerIDs(target, playerID, command) {
		if addDebugScoreForPlayer(target, targetPlayerID, command.Amount) {
			foundAny = true
		}
	}
	return foundAny
}

func setDebugScoreForPlayer(target CounterTarget, targetPlayerID string, score int) bool {
	return target.SetPlayerScore(targetPlayerID, score)
}

func addDebugScoreForPlayer(target CounterTarget, targetPlayerID string, amount int) bool {
	return target.AddPlayerScore(targetPlayerID, amount)
}

func handleDebugSetLives(target counterCommandTarget, playerID string, command DebugCommand) bool {
	if target == nil {
		return false
	}

	foundAny := false
	for _, targetPlayerID := range resolveCommandTargetPlayerIDs(target, playerID, command) {
		if setDebugLivesForPlayer(target, targetPlayerID, command.Lives) {
			foundAny = true
		}
	}
	return foundAny
}

func handleDebugAddLives(target counterCommandTarget, playerID string, command DebugCommand) bool {
	if target == nil {
		return false
	}

	foundAny := false
	for _, targetPlayerID := range resolveCommandTargetPlayerIDs(target, playerID, command) {
		if addDebugLivesForPlayer(target, targetPlayerID, command.Amount) {
			foundAny = true
		}
	}
	return foundAny
}

func setDebugLivesForPlayer(target CounterTarget, targetPlayerID string, lives int) bool {
	return target.SetPlayerLives(targetPlayerID, lives)
}

func addDebugLivesForPlayer(target CounterTarget, targetPlayerID string, amount int) bool {
	return target.AddPlayerLives(targetPlayerID, amount)
}
