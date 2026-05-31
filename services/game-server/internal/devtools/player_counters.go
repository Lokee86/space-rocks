package devtools

import "github.com/Lokee86/space-rocks/server/internal/game"

func handleDebugSetScore(target *game.Game, playerID string, command DebugCommand) bool {
	if target == nil {
		return false
	}

	targetPlayerID := resolveCounterTargetPlayerID(playerID, command)
	change := target.DevtoolsSetPlayerScore(targetPlayerID, command.Score)
	return change.Found
}

func handleDebugAddScore(target *game.Game, playerID string, command DebugCommand) bool {
	if target == nil {
		return false
	}

	targetPlayerID := resolveCounterTargetPlayerID(playerID, command)
	change := target.DevtoolsAddPlayerScore(targetPlayerID, command.Amount)
	return change.Found
}

func handleDebugSetLives(target *game.Game, playerID string, command DebugCommand) bool {
	if target == nil {
		return false
	}

	targetPlayerID := resolveCounterTargetPlayerID(playerID, command)
	change := target.DevtoolsSetPlayerLives(targetPlayerID, command.Lives)
	return change.Found
}

func handleDebugAddLives(target *game.Game, playerID string, command DebugCommand) bool {
	if target == nil {
		return false
	}

	targetPlayerID := resolveCounterTargetPlayerID(playerID, command)
	change := target.DevtoolsAddPlayerLives(targetPlayerID, command.Amount)
	return change.Found
}

func resolveCounterTargetPlayerID(callingPlayerID string, command DebugCommand) string {
	if command.TargetPlayerID != "" {
		return command.TargetPlayerID
	}

	return callingPlayerID
}
