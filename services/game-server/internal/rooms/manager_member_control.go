package rooms

func (manager *RoomManager) AddBot(roomID string, ownerSessionID string) (*Room, RoomMember, *RoomDomainError) {
	roomID = NormalizeRoomID(roomID)
	manager.mu.Lock()
	room, ok := manager.rooms[roomID]
	manager.mu.Unlock()
	if !ok {
		return nil, RoomMember{}, &RoomDomainError{Code: RoomErrorRoomNotFound, Message: "Room was not found."}
	}

	member, roomErr := room.AddBotForOwnerSession(ownerSessionID)
	if roomErr != nil {
		return nil, RoomMember{}, roomErr
	}
	return room, member, nil
}

func (manager *RoomManager) RemoveRoomMember(roomID string, ownerSessionID string, targetPlayerID string) (*Room, RoomMember, *RoomDomainError) {
	roomID = NormalizeRoomID(roomID)
	manager.mu.Lock()
	room, ok := manager.rooms[roomID]
	manager.mu.Unlock()
	if !ok {
		return nil, RoomMember{}, &RoomDomainError{Code: RoomErrorRoomNotFound, Message: "Room was not found."}
	}

	member, roomErr := room.RemoveMemberForOwnerSession(ownerSessionID, targetPlayerID)
	if roomErr != nil {
		return nil, RoomMember{}, roomErr
	}
	return room, member, nil
}
