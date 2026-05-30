package rooms

type LeaveMemberResult struct {
	Room                    *Room
	RoomID                  string
	SessionID               string
	PlayerID                string
	RemainingMembers        int
	ActivePlayers           int
	PlayerRemoved           bool
	CleanupScheduled        bool
	ShouldBroadcastSnapshot bool
}

func (manager *RoomManager) LeaveMember(roomID string, sessionID string, playerID string) (*LeaveMemberResult, *RoomDomainError) {
	leaveResult, roomErr := manager.LeaveRoom(roomID, sessionID)
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
		SessionID:               sessionID,
		PlayerID:                playerID,
		RemainingMembers:        remainingMembers,
		ActivePlayers:           room.ActivePlayers,
		PlayerRemoved:           playerRemoved,
		CleanupScheduled:        cleanupScheduled,
		ShouldBroadcastSnapshot: remainingMembers > 0,
	}, nil
}
