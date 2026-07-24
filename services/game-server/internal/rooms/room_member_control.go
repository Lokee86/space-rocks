package rooms

func (room *Room) AddBotForOwnerSession(sessionID string) (RoomMember, *RoomDomainError) {
	room.mu.Lock()
	defer room.mu.Unlock()

	if room.State != RoomStateLobby {
		return RoomMember{}, &RoomDomainError{Code: RoomErrorInvalidRoomState, Message: "Bots can only be added in the lobby."}
	}
	requester, ok := room.memberForSessionLocked(sessionID)
	if !ok {
		return RoomMember{}, &RoomDomainError{Code: RoomErrorNotInRoom, Message: "Member is not in the room."}
	}
	if requester.PlayerID != room.membership.ownerIDValue() {
		return RoomMember{}, &RoomDomainError{Code: RoomErrorNotRoomOwner, Message: "Only the room owner can add bots."}
	}
	if room.membership.memberCount() >= MaxPlayersPerRoom {
		return RoomMember{}, &RoomDomainError{Code: RoomErrorRoomFull, Message: "Room is full."}
	}

	member := room.addMemberLocked(NewBotRoomMember())
	return *member, nil
}

func (room *Room) RemoveMemberForOwnerSession(sessionID string, targetPlayerID string) (RoomMember, *RoomDomainError) {
	room.mu.Lock()
	defer room.mu.Unlock()

	if room.State != RoomStateLobby {
		return RoomMember{}, &RoomDomainError{Code: RoomErrorInvalidRoomState, Message: "Members can only be removed in the lobby."}
	}
	requester, ok := room.memberForSessionLocked(sessionID)
	if !ok {
		return RoomMember{}, &RoomDomainError{Code: RoomErrorNotInRoom, Message: "Member is not in the room."}
	}
	if requester.PlayerID != room.membership.ownerIDValue() {
		return RoomMember{}, &RoomDomainError{Code: RoomErrorNotRoomOwner, Message: "Only the room owner can remove members."}
	}
	if targetPlayerID == room.membership.ownerIDValue() {
		return RoomMember{}, &RoomDomainError{Code: RoomErrorCannotRemoveOwner, Message: "The room owner cannot be removed."}
	}
	target, ok := room.membership.memberByPlayerID(targetPlayerID)
	if !ok {
		return RoomMember{}, &RoomDomainError{Code: RoomErrorNotInRoom, Message: "Member is not in the room."}
	}

	removed := *target
	room.membership.removeMember(targetPlayerID)
	return removed, nil
}
