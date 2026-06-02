package rooms

func (room *Room) AddMember(member *RoomMember) *RoomMember {
	room.mu.Lock()
	defer room.mu.Unlock()

	return room.addMemberLocked(member)
}

func (room *Room) addMemberLocked(member *RoomMember) *RoomMember {
	member.PlayerID = room.nextAvailablePlayerIDLocked()
	room.Members[member.PlayerID] = member
	if room.OwnerID == "" {
		room.OwnerID = member.PlayerID
	}

	return member
}

func (room *Room) AddMemberSessionID(sessionID string) *RoomMember {
	return room.AddMember(NewRoomMember(sessionID))
}

func (room *Room) PlayerIDForSession(sessionID string) (string, bool) {
	room.mu.Lock()
	defer room.mu.Unlock()

	for _, member := range room.Members {
		if member.SessionID == sessionID {
			return member.PlayerID, true
		}
	}
	return "", false
}

func (room *Room) RemoveMember(playerID string) {
	room.mu.Lock()
	defer room.mu.Unlock()

	delete(room.Members, playerID)
	if room.OwnerID == playerID {
		room.OwnerID = ""
		room.assignNextOwnerLocked()
	}
}

func (room *Room) assignNextOwnerLocked() {
	for remainingPlayerID := range room.Members {
		if room.OwnerID == "" || remainingPlayerID < room.OwnerID {
			room.OwnerID = remainingPlayerID
		}
	}
}

func (room *Room) MemberCount() int {
	room.mu.Lock()
	defer room.mu.Unlock()

	return len(room.Members)
}

func (room *Room) IsFull() bool {
	return room.MemberCount() >= MaxPlayersPerRoom
}

func (room *Room) IsEmpty() bool {
	room.mu.Lock()
	defer room.mu.Unlock()

	return room.ActivePlayers == 0 && len(room.Members) == 0
}

func (room *Room) MembersSnapshot() []RoomMember {
	room.mu.Lock()
	defer room.mu.Unlock()

	members := make([]RoomMember, 0, len(room.Members))
	for _, member := range room.Members {
		members = append(members, *member)
	}

	return members
}
