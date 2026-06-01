package targeting

// ValidateRequestedTarget returns the accepted target_player_id and whether
// the request is valid.
func ValidateRequestedTarget(
	requesterPlayerID string,
	requestedTargetPlayerID string,
	playerExists func(playerID string) bool,
) (string, bool) {
	if playerExists == nil {
		return "", false
	}

	if !playerExists(requesterPlayerID) {
		return "", false
	}

	if requestedTargetPlayerID == "" {
		return "", true
	}

	if !playerExists(requestedTargetPlayerID) {
		return "", false
	}

	return requestedTargetPlayerID, true
}
