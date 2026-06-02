package roomrules

// StartMember captures the room-member state needed to decide whether a game can start.
type StartMember struct {
	PlayerID   string
	Ready      bool
	Connected  bool
}

// StartInput captures the data needed to decide whether a game can start.
type StartInput struct {
	State              string
	OwnerID            string
	RequestingPlayerID string
	Members            []StartMember
}

// DecideStart applies the start policy without touching room state directly.
func DecideStart(input StartInput) Decision {
	foundRequester := false
	for _, member := range input.Members {
		if member.PlayerID != input.RequestingPlayerID {
			continue
		}
		foundRequester = true
		break
	}
	if !foundRequester {
		return Reject("not_in_room", "Member is not in the room.")
	}

	if input.RequestingPlayerID != input.OwnerID {
		return Reject("not_room_owner", "Only the room owner can start the game.")
	}

	switch input.State {
	case "Lobby":
	case "Starting", "InGame":
		return Reject("room_in_game", "Room is already in game.")
	default:
		return Reject("invalid_room_state", "Game can only be started from the lobby.")
	}

	for _, member := range input.Members {
		if member.Connected && !member.Ready {
			return Reject("not_ready", "All connected members must be ready.")
		}
	}

	return Allow()
}
