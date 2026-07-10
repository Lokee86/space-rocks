package devtools

func handleDebugClearBullets(controller *Controller, playerID string, command DebugCommand) bool {
	if controller == nil || controller.target == nil || controller.streams == nil {
		return false
	}

	controller.target.ClearBullets()
	controller.streams.ClearContinuousBulletStreams()
	return true
}

func handleDebugClearAsteroids(target ClearTarget, playerID string, command DebugCommand) bool {
	if target == nil {
		return false
	}

	target.ClearAsteroids()
	return true
}
