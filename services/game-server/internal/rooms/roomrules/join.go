package roomrules

// JoinInput captures the data needed to decide whether a room can be joined.
type JoinInput struct {
	State      string
	Joinable   bool
	MemberCount int
	MaxMembers int
}

// DecideJoin applies the join policy without touching room state directly.
func DecideJoin(input JoinInput) Decision {
	switch input.State {
	case "Lobby":
		// Continue to the remaining checks.
	case "Starting", "InGame":
		return Reject("room_in_game", "Room is already in game.")
	case "Closed":
		return Reject("room_closed", "Room is closed.")
	default:
		return Reject("invalid_room_state", "Room is not joinable.")
	}

	if !input.Joinable {
		return Reject("invalid_room_state", "Room is not joinable.")
	}

	if input.MemberCount >= input.MaxMembers {
		return Reject("room_full", "Room is full.")
	}

	return Allow()
}
