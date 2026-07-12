package rooms

type RoomPopulation struct {
	Members       int
	ActivePlayers int
}

func (room *Room) Population() RoomPopulation {
	room.mu.Lock()
	defer room.mu.Unlock()
	return RoomPopulation{Members: room.membership.memberCount(), ActivePlayers: room.match.ActivePlayers()}
}

func (room *Room) ActivateMemberPlayer(expected GameplayContext, sessionID, playerID string) bool {
	room.mu.Lock()
	defer room.mu.Unlock()

	actual := room.gameplayContextLocked()
	if actual != expected || sessionID == "" || playerID == "" || expected.Game == nil || expected.State != RoomStateInGame {
		return false
	}

	member, ok := room.memberForSessionLocked(sessionID)
	if !ok {
		return false
	}
	if existing, exists := room.membership.memberByPlayerID(playerID); exists && existing != member {
		return false
	}

	if room.match.SessionActive(sessionID) {
		return member.PlayerID == playerID
	}

	if member.PlayerID != playerID {
		if member.PlayerID != "" {
			oldPlayerID := member.PlayerID
			delete(room.membership.members, oldPlayerID)
			if room.membership.ownerID == oldPlayerID {
				room.membership.ownerID = playerID
			}
		}
		member.PlayerID = playerID
		room.membership.members[playerID] = member
	}

	return room.match.ActivateSession(sessionID)
}

// DecrementActivePlayerCount is retained for legacy count-only callers.
func (room *Room) DecrementActivePlayerCount() int {
	room.mu.Lock()
	defer room.mu.Unlock()
	if room.match.ActivePlayers() > 0 {
		room.match.SetActivePlayers(room.match.ActivePlayers() - 1)
	}
	return room.match.ActivePlayers()
}

func (room *Room) ResetActivePlayerCount(expected *GameplayContext) bool {
	room.mu.Lock()
	defer room.mu.Unlock()
	if expected != nil && room.gameplayContextLocked() != *expected {
		return false
	}
	room.match.ResetActiveSessions()
	return true
}

func (room *Room) DeactivateMemberPlayer(sessionID string) bool {
	room.mu.Lock()
	defer room.mu.Unlock()
	return room.match.DeactivateSession(sessionID)
}

func (room *Room) RemoveMemberForSession(sessionID string) (removed RoomMember, remaining int, ok bool) {
	room.mu.Lock()
	defer room.mu.Unlock()

	member, ok := room.memberForSessionLocked(sessionID)
	if !ok {
		return RoomMember{}, room.membership.memberCount(), false
	}
	removed = *member
	room.membership.removeMember(member.PlayerID)
	return removed, room.membership.memberCount(), true
}

func (room *Room) memberForSessionLocked(sessionID string) (*RoomMember, bool) {
	playerID, ok := room.membership.playerIDForSession(sessionID)
	if !ok {
		return nil, false
	}
	return room.membership.memberByPlayerID(playerID)
}
