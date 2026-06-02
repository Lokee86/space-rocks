package devtools

import "github.com/Lokee86/space-rocks/server/internal/game"

func handleDebugSetScore(target *game.Game, playerID string, command DebugCommand) bool {
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

func handleDebugAddScore(target *game.Game, playerID string, command DebugCommand) bool {
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

func setDebugScoreForPlayer(target *game.Game, targetPlayerID string, score int) bool {
	change := target.DevtoolsSetPlayerScore(targetPlayerID, score)
	return change.Found
}

func addDebugScoreForPlayer(target *game.Game, targetPlayerID string, amount int) bool {
	change := target.DevtoolsAddPlayerScore(targetPlayerID, amount)
	return change.Found
}

func handleDebugSetLives(target *game.Game, playerID string, command DebugCommand) bool {
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

func handleDebugAddLives(target *game.Game, playerID string, command DebugCommand) bool {
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

func setDebugLivesForPlayer(target *game.Game, targetPlayerID string, lives int) bool {
	change := target.DevtoolsSetPlayerLives(targetPlayerID, lives)
	return change.Found
}

func addDebugLivesForPlayer(target *game.Game, targetPlayerID string, amount int) bool {
	change := target.DevtoolsAddPlayerLives(targetPlayerID, amount)
	return change.Found
}
