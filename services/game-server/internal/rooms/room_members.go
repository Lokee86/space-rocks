package rooms

func (room *Room) AddMember(member *RoomMember) *RoomMember {
	room.mu.Lock()
	defer room.mu.Unlock()

	return room.addMemberLocked(member)
}

func (room *Room) addMemberLocked(member *RoomMember) *RoomMember {
	member = room.membership.addMember(member)
	if _, exists := room.roomTeams.assignments[member.MemberID]; !exists {
		room.roomTeams.assignments[member.MemberID] = defaultAssignmentForStructure(room.roomTeams.rules.Structure)
	}
	_ = room.rebalanceAutoBalancedLocked()
	return member
}

func (room *Room) AddMemberSessionID(sessionID string) *RoomMember {
	return room.AddMember(NewRoomMember(sessionID))
}

func (room *Room) PlayerIDForSession(sessionID string) (string, bool) {
	room.mu.Lock()
	defer room.mu.Unlock()

	return room.membership.playerIDForSession(sessionID)
}

func (room *Room) SetMemberAccountIDForSession(sessionID string, accountID string) bool {
	room.mu.Lock()
	defer room.mu.Unlock()

	playerID, ok := room.membership.playerIDForSession(sessionID)
	if !ok {
		return false
	}

	member, ok := room.membership.memberByPlayerID(playerID)
	if !ok {
		return false
	}

	member.AccountID = accountID
	return true
}

func (room *Room) SetMemberPlayerIDForSession(sessionID string, playerID string) bool {
	room.mu.Lock()
	defer room.mu.Unlock()

	return room.membership.setMemberPlayerIDForSession(sessionID, playerID)
}

func (room *Room) SetMemberLocalProfileIDForSession(sessionID string, localProfileID string) bool {
	room.mu.Lock()
	defer room.mu.Unlock()

	playerID, ok := room.membership.playerIDForSession(sessionID)
	if !ok {
		return false
	}

	member, ok := room.membership.memberByPlayerID(playerID)
	if !ok {
		return false
	}

	member.LocalProfileID = localProfileID
	return true
}

func (room *Room) OwnerID() string {
	room.mu.Lock()
	defer room.mu.Unlock()

	return room.membership.ownerIDValue()
}

func (room *Room) RemoveMember(playerID string) {
	room.mu.Lock()
	defer room.mu.Unlock()

	_, _ = room.removeMemberLocked(playerID)
}

func (room *Room) removeMemberLocked(playerID string) (RoomMember, bool) {
	member, ok := room.membership.memberByPlayerID(playerID)
	if !ok {
		return RoomMember{}, false
	}

	removed := *member
	room.match.RememberParticipantIdentity(removed)
	delete(room.roomTeams.assignments, member.MemberID)
	room.membership.removeMember(playerID)
	_ = room.rebalanceAutoBalancedLocked()
	return removed, true
}

func (room *Room) MemberCount() int {
	room.mu.Lock()
	defer room.mu.Unlock()

	return room.membership.memberCount()
}

func (room *Room) IsFull() bool {
	room.mu.Lock()
	defer room.mu.Unlock()
	return room.membership.memberCount() >= room.MaxPlayers
}

func (room *Room) IsEmpty() bool {
	room.mu.Lock()
	defer room.mu.Unlock()

	return room.match.ActivePlayers() == 0 && room.membership.memberCount() == 0
}

func (room *Room) MembersSnapshot() []RoomMember {
	room.mu.Lock()
	defer room.mu.Unlock()

	return room.membership.membersSnapshot()
}
