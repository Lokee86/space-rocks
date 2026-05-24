package networking

import "github.com/Lokee86/space-rocks/server/internal/rooms"

func activateRoomPlayers(room *rooms.Room) {
	// Websocket sessions keep the per-connection player ID, so activation stays in networking.
	memberSnapshot := room.MembersSnapshot()
	memberIDs := make([]string, 0, len(memberSnapshot))
	for _, member := range memberSnapshot {
		if !member.Connected {
			continue
		}
		memberIDs = append(memberIDs, member.SessionID)
	}

	sessions := snapshotRoomSessions(room, memberIDs)
	for _, session := range sessions {
		if session == nil || session.currentPlayerID != "" {
			continue
		}

		playerID := room.Game.AddPlayer()
		session.currentPlayerID = playerID
		room.ActivePlayers++
	}
}

func deactivateRoomPlayers(room *rooms.Room) {
	memberSnapshot := room.MembersSnapshot()
	memberIDs := make([]string, 0, len(memberSnapshot))
	for _, member := range memberSnapshot {
		if !member.Connected {
			continue
		}
		memberIDs = append(memberIDs, member.SessionID)
	}

	sessions := snapshotRoomSessions(room, memberIDs)
	for _, session := range sessions {
		if session == nil {
			continue
		}
		session.currentPlayerID = ""
	}
	room.ActivePlayers = 0
}
