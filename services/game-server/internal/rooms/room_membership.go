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
	if membership.ownerID == "" && !member.IsBot {
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

func (membership *roomMembership) setMemberPlayerIDForSession(sessionID string, playerID string) bool {
	for currentPlayerID, member := range membership.members {
		if member.SessionID != sessionID {
			continue
		}

		if currentPlayerID == playerID {
			return true
		}

		delete(membership.members, currentPlayerID)
		member.PlayerID = playerID
		membership.members[playerID] = member
		if membership.ownerID == currentPlayerID {
			membership.ownerID = playerID
		}
		return true
	}

	return false
}

func (membership *roomMembership) memberCount() int {
	return len(membership.members)
}

func (membership *roomMembership) humanMemberCount() int {
	count := 0
	for _, member := range membership.members {
		if !member.IsBot {
			count++
		}
	}
	return count
}

func (membership *roomMembership) removeBots() []RoomMember {
	botPlayerIDs := make([]string, 0)
	for playerID, member := range membership.members {
		if member.IsBot {
			botPlayerIDs = append(botPlayerIDs, playerID)
		}
	}

	removed := make([]RoomMember, 0, len(botPlayerIDs))
	for _, playerID := range botPlayerIDs {
		member, ok := membership.members[playerID]
		if !ok {
			continue
		}
		removed = append(removed, *member)
		membership.removeMember(playerID)
	}
	return removed
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
		member.SetReady(ready || member.IsBot)
	}
}

func (membership *roomMembership) nextOwnerID() string {
	ownerID := ""
	for remainingPlayerID, member := range membership.members {
		if member.IsBot {
			continue
		}
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
