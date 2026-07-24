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

func (manager *RoomManager) LeaveMember(roomID, sessionID, _ string) (*LeaveMemberResult, *RoomDomainError) {
	leaveResult, roomErr := manager.LeaveRoom(roomID, sessionID)
	if roomErr != nil {
		return nil, roomErr
	}

	room := leaveResult.Room
	removedPlayerID := leaveResult.RemovedMember.PlayerID
	gameInstance := room.GameInstance()
	playerRemoved := false
	if leaveResult.MemberRemoved && removedPlayerID != "" && gameInstance != nil && room.DeactivateMemberPlayer(sessionID) {
		gameInstance.RemovePlayer(removedPlayerID)
		playerRemoved = true
	}
	for _, removedBot := range leaveResult.RemovedBots {
		if gameInstance == nil || removedBot.PlayerID == "" {
			continue
		}
		room.DeactivateMemberPlayer(removedBot.SessionID)
		gameInstance.RemovePlayer(removedBot.PlayerID)
	}

	population := room.Population()
	cleanupScheduled := population.Members == 0 && population.ActivePlayers == 0
	manager.ScheduleCleanupIfEmpty(leaveResult.RoomID)

	return &LeaveMemberResult{
		Room:                    room,
		RoomID:                  leaveResult.RoomID,
		SessionID:               sessionID,
		PlayerID:                removedPlayerID,
		RemainingMembers:        population.Members,
		ActivePlayers:           population.ActivePlayers,
		PlayerRemoved:           playerRemoved,
		CleanupScheduled:        cleanupScheduled,
		ShouldBroadcastSnapshot: population.Members > 0,
	}, nil
}
