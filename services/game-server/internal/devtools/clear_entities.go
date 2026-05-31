package devtools

import "github.com/Lokee86/space-rocks/server/internal/game"

func handleDebugClearBullets(target *game.Game, playerID string, command DebugCommand) bool {
	if target == nil {
		return false
	}

	target.DevtoolsClearBullets()
	return true
}

func handleDebugClearAsteroids(target *game.Game, playerID string, command DebugCommand) bool {
	if target == nil {
		return false
	}

	target.DevtoolsClearAsteroids()
	return true
}
