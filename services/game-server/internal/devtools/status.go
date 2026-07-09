package devtools

func StatusFor(target Target, playerID string) DebugStatus {
	controller := NewController(Dependencies{Target: target})
	return controller.StatusFor(playerID)
}

func StatusesForAllPlayers(target Target) map[string]DebugStatus {
	controller := NewController(Dependencies{Target: target})
	return controller.StatusesForAllPlayers()
}
