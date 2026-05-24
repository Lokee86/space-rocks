package rooms

type LeaveMemberResult struct {
	Room                    *Room
	RoomID                  string
	MemberID                string
	PlayerID                string
	RemainingMembers        int
	ActivePlayers           int
	PlayerRemoved           bool
	CleanupScheduled        bool
	ShouldBroadcastSnapshot bool
}

func (manager *RoomManager) LeaveMember(roomID string, memberID string, playerID string) (*LeaveMemberResult, *RoomDomainError) {
	leaveResult, roomErr := manager.LeaveRoom(roomID, memberID)
	if roomErr != nil {
		return nil, roomErr
	}

	room := leaveResult.Room
	playerRemoved := false
	if playerID != "" && room.Game != nil {
		room.Game.RemovePlayer(playerID)
		playerRemoved = true
		if room.ActivePlayers > 0 {
			room.ActivePlayers--
		}
	}
	cleanupScheduled := room.ShouldCleanup()
	manager.ScheduleCleanupIfEmpty(leaveResult.RoomID)

	remainingMembers := room.MemberCount()
	return &LeaveMemberResult{
		Room:                    room,
		RoomID:                  leaveResult.RoomID,
		MemberID:                memberID,
		PlayerID:                playerID,
		RemainingMembers:        remainingMembers,
		ActivePlayers:           room.ActivePlayers,
		PlayerRemoved:           playerRemoved,
		CleanupScheduled:        cleanupScheduled,
		ShouldBroadcastSnapshot: remainingMembers > 0,
	}, nil
}
