package networking

import (
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/teams"
	"github.com/Lokee86/space-rocks/services/game-server/internal/rooms"
)

var activateMemberPlayerCall = func(room *rooms.Room, expected rooms.GameplayContext, sessionID, playerID string) bool {
	return room.ActivateMemberPlayer(expected, sessionID, playerID)
}

func activateRoomPlayers(room *rooms.Room) {
	// Websocket sessions keep the per-connection player ID, so human activation stays in networking.
	gameplayContext := room.GameplayContext()
	if gameplayContext.Game == nil {
		return
	}
	teamSnapshot, ok := room.TeamStartSnapshot()
	if !ok {
		return
	}
	gameplayContext.Game.SetTeamStructure(teamSnapshot.Config.Structure)
	memberSnapshot := room.MembersSnapshot()
	memberIDs := make([]string, 0, len(memberSnapshot))
	teamBySessionID := make(map[string]teams.ID, len(memberSnapshot))
	for _, member := range memberSnapshot {
		if !member.Connected || member.IsBot {
			continue
		}
		memberIDs = append(memberIDs, member.SessionID)
		teamID, assigned := teamSnapshot.Assignments[member.MemberID]
		if assigned {
			teamBySessionID[member.SessionID] = teamID
		}
	}

	sessions := snapshotRoomSessions(room, memberIDs)
	for _, session := range sessions {
		if session == nil {
			continue
		}
		context := session.sessionContext()
		if context.Room != room || context.GamePlayerID != "" {
			continue
		}

		teamID, assigned := teamBySessionID[session.sessionID]
		if !assigned {
			continue
		}
		playerID := gameplayContext.Game.AddPlayerWithTeamAndBuild(teamID, session.resolvedBuildTemplate())
		if !session.setGamePlayerIDForRoom(room, playerID) {
			gameplayContext.Game.RollbackPlayerAdd(playerID)
			continue
		}
		if !activateMemberPlayerCall(room, gameplayContext, session.sessionID, playerID) {
			session.clearGamePlayerIDForRoom(room)
			gameplayContext.Game.RollbackPlayerAdd(playerID)
		}
	}

	for _, member := range memberSnapshot {
		if !member.IsBot || !member.Connected {
			continue
		}
		teamID, assigned := teamSnapshot.Assignments[member.MemberID]
		if !assigned {
			continue
		}
		playerID := gameplayContext.Game.AddBotWithTeam(teamID)
		if !activateMemberPlayerCall(room, gameplayContext, member.SessionID, playerID) {
			gameplayContext.Game.RemovePlayer(playerID)
		}
	}
}

func deactivateRoomPlayers(room *rooms.Room) {
	gameplayContext := room.GameplayContext()
	memberSnapshot := room.MembersSnapshot()
	memberIDs := make([]string, 0, len(memberSnapshot))
	for _, member := range memberSnapshot {
		if !member.Connected || member.IsBot {
			continue
		}
		memberIDs = append(memberIDs, member.SessionID)
	}

	sessions := snapshotRoomSessions(room, memberIDs)
	for _, session := range sessions {
		if session == nil {
			continue
		}
		session.clearGamePlayerIDForRoom(room)
	}
	room.ResetActivePlayerCount(&gameplayContext)
}
