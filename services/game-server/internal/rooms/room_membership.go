package rooms

type roomMembership struct {
	members map[string]*RoomMember
	ownerID string
}

func newRoomMembership() *roomMembership {
	return &roomMembership{
		members: make(map[string]*RoomMember),
	}
}

func (membership *roomMembership) addMember(member *RoomMember) *RoomMember {
	member.PlayerID = membership.nextAvailablePlayerID()
	membership.members[member.PlayerID] = member
	if membership.ownerID == "" {
		membership.ownerID = member.PlayerID
	}

	return member
}

func (membership *roomMembership) removeMember(playerID string) {
	delete(membership.members, playerID)
	if membership.ownerID == playerID {
		membership.ownerID = membership.nextOwnerID()
	}
}

func (membership *roomMembership) playerIDForSession(sessionID string) (string, bool) {
	for _, member := range membership.members {
		if member.SessionID == sessionID {
			return member.PlayerID, true
		}
	}
	return "", false
}

func (membership *roomMembership) memberByPlayerID(playerID string) (*RoomMember, bool) {
	member, ok := membership.members[playerID]
	return member, ok
}

func (membership *roomMembership) memberCount() int {
	return len(membership.members)
}

func (membership *roomMembership) membersSnapshot() []RoomMember {
	members := make([]RoomMember, 0, len(membership.members))
	for _, member := range membership.members {
		members = append(members, *member)
	}

	return members
}

func (membership *roomMembership) ownerIDValue() string {
	return membership.ownerID
}

func (membership *roomMembership) setAllReady(ready bool) {
	for _, member := range membership.members {
		member.SetReady(ready)
	}
}

func (membership *roomMembership) nextOwnerID() string {
	ownerID := ""
	for remainingPlayerID := range membership.members {
		if ownerID == "" || remainingPlayerID < ownerID {
			ownerID = remainingPlayerID
		}
	}
	return ownerID
}

func (membership *roomMembership) occupiedPlayerIDs() map[string]bool {
	occupied := make(map[string]bool, len(membership.members))
	for _, member := range membership.members {
		if member.PlayerID == "" {
			continue
		}
		occupied[member.PlayerID] = true
	}
	return occupied
}

func (membership *roomMembership) nextAvailablePlayerID() string {
	occupied := membership.occupiedPlayerIDs()
	for number := 1; ; number++ {
		playerID := formatPlayerID(number)
		if !occupied[playerID] {
			return playerID
		}
	}
}
