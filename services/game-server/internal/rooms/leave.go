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
	gameInstance := room.GameInstance()
	if playerID != "" && gameInstance != nil {
		gameInstance.RemovePlayer(playerID)
		playerRemoved = true
		activePlayers := room.match.ActivePlayers()
		if activePlayers > 0 {
			room.match.SetActivePlayers(activePlayers - 1)
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
		ActivePlayers:           room.ActivePlayerCount(),
		PlayerRemoved:           playerRemoved,
		CleanupScheduled:        cleanupScheduled,
		ShouldBroadcastSnapshot: remainingMembers > 0,
	}, nil
}
