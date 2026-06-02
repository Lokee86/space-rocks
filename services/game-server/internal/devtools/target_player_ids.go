package devtools

import "github.com/Lokee86/space-rocks/server/internal/game"

func resolveCommandTargetPlayerIDs(target *game.Game, requestingPlayerID string, command DebugCommand) []string {
	if command.TargetScope == targetScopeAllPlayers {
		if target == nil {
			return []string{}
		}
		return target.DevtoolsTargetPlayerIDs()
	}

	targetPlayerID := command.TargetPlayerID
	if targetPlayerID == "" {
		targetPlayerID = requestingPlayerID
	}

	return []string{targetPlayerID}
}
