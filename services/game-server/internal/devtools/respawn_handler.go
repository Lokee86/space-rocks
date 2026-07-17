package devtools

func handleDebugRespawnPlayer(target Target, playerID string, command DebugCommand) bool {
	if command.TargetScope == targetScopeAllPlayers {
		for _, targetPlayerID := range resolveCommandTargetPlayerIDs(target, playerID, command) {
			handleDebugRespawnPlayerTarget(target, playerID, RespawnPlayerRequest{
				TargetPlayerID: targetPlayerID,
				X:              command.X,
				Y:              command.Y,
			})
		}
		return true
	}

	return handleDebugRespawnPlayerTarget(target, playerID, RespawnPlayerRequest{
		TargetPlayerID: command.TargetPlayerID,
		X:              command.X,
		Y:              command.Y,
	})
}

func handleDebugRespawnPlayerTarget(target Target, playerID string, request RespawnPlayerRequest) bool {
	if target == nil {
		return true
	}

	normalizedTargetPlayerID, ok := resolveDebugRespawnTargetPlayerID(request)
	if !ok {
		return true
	}

	isPlayerAlive := false
	for _, player := range target.MatchDecision().Players {
		if player.ID == normalizedTargetPlayerID {
			isPlayerAlive = player.Status == "active"
			break
		}
	}

	if isPlayerAlive {
		return true
	}

	request.TargetPlayerID = normalizedTargetPlayerID

	_, _, ok = applyDebugRespawnPlayer(target, request)
	if !ok {
		return true
	}

	return true
}
