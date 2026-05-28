package rooms

func (room *Room) JoinMember(sessionID string) *RoomDomainError {
	room.mu.Lock()
	defer room.mu.Unlock()

	switch room.State {
	case RoomStateLobby:
	case RoomStateStarting, RoomStateInGame:
		return &RoomDomainError{Code: RoomErrorRoomInGame, Message: "Room is already in game."}
	case RoomStateClosed:
		return &RoomDomainError{Code: RoomErrorRoomClosed, Message: "Room is closed."}
	default:
		return &RoomDomainError{Code: RoomErrorInvalidRoomState, Message: "Room is not joinable."}
	}

	if !room.Joinable {
		return &RoomDomainError{Code: RoomErrorInvalidRoomState, Message: "Room is not joinable."}
	}

	if len(room.Members) >= MaxPlayersPerRoom {
		return &RoomDomainError{Code: RoomErrorRoomFull, Message: "Room is full."}
	}

	room.addMemberLocked(NewRoomMember(sessionID))
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
