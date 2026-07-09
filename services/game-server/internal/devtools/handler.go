package devtools

func HandleCommand(target Target, playerID string, command DebugCommand) bool {
	if target == nil {
		return false
	}
	return NewController(Dependencies{Target: target}).HandleCommand(playerID, command)
}
