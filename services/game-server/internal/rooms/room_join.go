package rooms

func (room *Room) SetJoinable(joinable bool) {
	room.mu.Lock()
	defer room.mu.Unlock()

	room.Joinable = joinable
}

func (room *Room) IsJoinable() bool {
	room.mu.Lock()
	defer room.mu.Unlock()

	return room.Joinable
}

func (room *Room) ValidateJoin() *RoomDomainError {
	room.mu.Lock()
	defer room.mu.Unlock()

	switch room.State {
	case RoomStateLobby:
	case RoomStateStarting, RoomStateInGame:
		return &RoomDomainError{
			Code:    RoomErrorRoomInGame,
			Message: "Room is already in game.",
		}
	case RoomStateClosed:
		return &RoomDomainError{
			Code:    RoomErrorRoomClosed,
			Message: "Room is closed.",
		}
	default:
		return &RoomDomainError{
			Code:    RoomErrorInvalidRoomState,
			Message: "Room is not joinable.",
		}
	}

	if !room.Joinable {
		return &RoomDomainError{
			Code:    RoomErrorInvalidRoomState,
			Message: "Room is not joinable.",
		}
	}

	if len(room.Members) >= MaxPlayersPerRoom {
		return &RoomDomainError{
			Code:    RoomErrorRoomFull,
			Message: "Room is full.",
		}
	}

	return nil
}

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
