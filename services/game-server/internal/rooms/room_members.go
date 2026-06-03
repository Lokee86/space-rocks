package rooms

func (room *Room) AddMember(member *RoomMember) *RoomMember {
	room.mu.Lock()
	defer room.mu.Unlock()

	return room.addMemberLocked(member)
}

func (room *Room) addMemberLocked(member *RoomMember) *RoomMember {
	return room.membership.addMember(member)
}

func (room *Room) AddMemberSessionID(sessionID string) *RoomMember {
	return room.AddMember(NewRoomMember(sessionID))
}

func (room *Room) PlayerIDForSession(sessionID string) (string, bool) {
	room.mu.Lock()
	defer room.mu.Unlock()

	return room.membership.playerIDForSession(sessionID)
}

func (room *Room) OwnerID() string {
	room.mu.Lock()
	defer room.mu.Unlock()

	return room.membership.ownerIDValue()
}

func (room *Room) RemoveMember(playerID string) {
	room.mu.Lock()
	defer room.mu.Unlock()

	room.membership.removeMember(playerID)
}

func (room *Room) MemberCount() int {
	room.mu.Lock()
	defer room.mu.Unlock()

	return room.membership.memberCount()
}

func (room *Room) IsFull() bool {
	return room.MemberCount() >= MaxPlayersPerRoom
}

func (room *Room) IsEmpty() bool {
	room.mu.Lock()
	defer room.mu.Unlock()

	return room.ActivePlayers == 0 && room.membership.memberCount() == 0
}

func (room *Room) MembersSnapshot() []RoomMember {
	room.mu.Lock()
	defer room.mu.Unlock()

	return room.membership.membersSnapshot()
}
