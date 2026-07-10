package devtools

type PlayerTargetSource interface {
	TargetPlayerIDs() []string
}

func resolveCommandTargetPlayerIDs(target PlayerTargetSource, requestingPlayerID string, command DebugCommand) []string {
	if command.TargetScope == targetScopeAllPlayers {
		if target == nil {
			return []string{}
		}
		return target.TargetPlayerIDs()
	}

	targetPlayerID := command.TargetPlayerID
	if targetPlayerID == "" {
		targetPlayerID = requestingPlayerID
	}

	return []string{targetPlayerID}
}
