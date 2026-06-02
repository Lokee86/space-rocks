package rooms

func (room *Room) ValidateStart(playerID string) *RoomDomainError {
	room.mu.Lock()
	defer room.mu.Unlock()

	return room.validateStartLocked(playerID)
}

func (room *Room) validateStartLocked(playerID string) *RoomDomainError {
	if _, ok := room.Members[playerID]; !ok {
		return &RoomDomainError{Code: RoomErrorNotInRoom, Message: "Member is not in the room."}
	}

	if playerID != room.OwnerID {
		return &RoomDomainError{Code: RoomErrorNotRoomOwner, Message: "Only the room owner can start the game."}
	}

	switch room.State {
	case RoomStateLobby:
	case RoomStateStarting, RoomStateInGame:
		return &RoomDomainError{Code: RoomErrorRoomInGame, Message: "Room is already in game."}
	default:
		return &RoomDomainError{Code: RoomErrorInvalidRoomState, Message: "Game can only be started from the lobby."}
	}

	for _, connectedMember := range room.Members {
		if connectedMember.Connected && !connectedMember.Ready {
			return &RoomDomainError{Code: RoomErrorNotReady, Message: "All connected members must be ready."}
		}
	}

	return nil
}

func (room *Room) SetReadyInLobby(playerID string, ready bool) *RoomDomainError {
	room.mu.Lock()
	defer room.mu.Unlock()

	if room.State != RoomStateLobby {
		return &RoomDomainError{Code: RoomErrorInvalidRoomState, Message: "Ready state can only be changed in the lobby."}
	}

	member, ok := room.Members[playerID]
	if !ok {
		return &RoomDomainError{Code: RoomErrorNotInRoom, Message: "Member is not in the room."}
	}

	member.SetReady(ready)
	return nil
}
