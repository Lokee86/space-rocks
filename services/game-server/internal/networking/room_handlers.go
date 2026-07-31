package networking

import (
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/modes"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/teams"
	"github.com/Lokee86/space-rocks/services/game-server/internal/logging"
	"github.com/Lokee86/space-rocks/services/game-server/internal/rooms"
	observability "github.com/Lokee86/space-rocks/shared/go/observabilityevent"
)

func (session *webSocketSession) handleCreateRoomRequest(traceID string, teamStructure string, teamAssignmentMode string, teamCount int, maxPlayers int, presetID string, startingLives int, infiniteLives bool, targetScore int) {
	if !requireAuthenticatedAccount(session, traceID) {
		return
	}

	context := session.sessionContext()
	if context.RoomID != "" {
		session.EnqueueRoomError(traceID, rooms.RoomErrorAlreadyInRoom, "Session is already in a room.")
		return
	}

	room, err := session.rooms.CreateLobbyRoomWithConfig(rooms.RoomCreationConfig{
		ModeConfig: modes.RoomModeConfig{
			PresetID: modes.PresetID(presetID), StartingLives: startingLives,
			InfiniteLives: infiniteLives, TargetScore: targetScore,
		},
		TeamConfig: teams.Config{
			Structure:      teams.Structure(teamStructure),
			AssignmentMode: teams.AssignmentMode(teamAssignmentMode),
			AutoTeamCount:  teamCount,
		},
		MaxPlayers: maxPlayers,
	})
	if err != nil {
		failureTraceID := traceID
		if failureTraceID == "" {
			failureTraceID = session.connectionTraceID
		}
		logging.Emit(observability.Request{
			Event: observability.EventNameRoomCreationFailed,
			Context: observability.Context{
				TraceID:   failureTraceID,
				SessionID: session.sessionID,
			},
			Fields: observability.Fields{
				"error_code":   "room_creation_failed",
				"failure_mode": "create_lobby_room_failed",
				"reason_code":  rooms.RoomErrorInvalidRoomState,
			},
		})
		session.EnqueueRoomError(traceID, rooms.RoomErrorInvalidRoomState, "Could not create room.")
		return
	}

	if traceID == "" {
		traceID = session.connectionTraceID
	}
	logging.Emit(observability.Request{
		Event: observability.EventNameRoomCreated,
		Context: observability.Context{
			TraceID:   traceID,
			SessionID: session.sessionID,
			RoomID:    room.ID,
		},
		Fields: observability.Fields{"reason_code": "lobby_created"},
	})
	addSessionMember(room, session.sessionID, session)
	session.bindRoom(room)
	session.EnqueueRoomSnapshot(room)
}

func (session *webSocketSession) handleJoinRoomRequest(roomCode string, traceID string) {
	if !requireAuthenticatedAccount(session, traceID) {
		return
	}

	context := session.sessionContext()
	if context.RoomID != "" {
		session.EnqueueRoomError(traceID, rooms.RoomErrorAlreadyInRoom, "Session is already in a room.")
		return
	}

	room, roomErr := session.rooms.JoinRoom(session.sessionID, roomCode)
	if roomErr != nil {
		session.EnqueueRoomError(traceID, roomErr.Code, roomErr.Message)
		return
	}

	attachRoomSession(room, session.sessionID, session)
	if accountID := accountIDForSession(session); accountID != "" {
		room.SetMemberAccountIDForSession(session.sessionID, accountID)
	}
	session.bindRoom(room)
	BroadcastRoomSnapshot(room)
}

func (session *webSocketSession) handleLeaveRoomRequest() {
	session.leaveRequestedRoom()
}

func (session *webSocketSession) handleSetReadyRequest(ready bool) {
	context := session.sessionContext()
	if context.RoomID == "" || session.sessionID == "" {
		session.EnqueueRoomError("", rooms.RoomErrorNotInRoom, "Session is not in a room.")
		return
	}

	room, roomErr := session.rooms.SetReady(context.RoomID, session.sessionID, ready)
	if roomErr != nil {
		session.EnqueueRoomError("", roomErr.Code, roomErr.Message)
		return
	}

	BroadcastRoomSnapshot(room)
}

func (session *webSocketSession) handleSetTeamAssignmentRequest(targetPlayerID string, teamID string, traceID string) {
	context := session.sessionContext()
	if context.RoomID == "" || session.sessionID == "" {
		session.EnqueueRoomError(traceID, rooms.RoomErrorNotInRoom, "Session is not in a room.")
		return
	}

	room, roomErr := session.rooms.SetTeamAssignment(context.RoomID, session.sessionID, targetPlayerID, teams.ID(teamID))
	if roomErr != nil {
		session.EnqueueRoomError(traceID, roomErr.Code, roomErr.Message)
		return
	}

	BroadcastRoomSnapshot(room)
}

func (session *webSocketSession) handleStartGameRequest() {
	context := session.sessionContext()
	if context.Room == nil || context.RoomID == "" || session.sessionID == "" {
		session.EnqueueRoomError("", rooms.RoomErrorNotInRoom, "Session is not in a room.")
		return
	}

	room, roomErr := session.rooms.StartRoomGame(context.RoomID, session.sessionID)
	if roomErr != nil {
		session.EnqueueRoomError("", roomErr.Code, roomErr.Message)
		return
	}

	activateRoomPlayers(room)
	BroadcastRoomSnapshot(room)
}

func (session *webSocketSession) handleStartSinglePlayerRequest(localProfileID string, traceID string, presetID string, startingLives int, infiniteLives bool, targetScore int) {
	context := session.sessionContext()
	if traceID == "" {
		traceID = session.connectionTraceID
	}

	if context.RoomID != "" {
		session.EnqueueRoomError(traceID, rooms.RoomErrorAlreadyInRoom, "Session is already in a room.")
		return
	}

	room, roomErr := session.rooms.CreateStartedSinglePlayerRoomWithModeConfig(session.sessionID, modes.RoomModeConfig{
		PresetID: modes.PresetID(presetID), StartingLives: startingLives,
		InfiniteLives: infiniteLives, TargetScore: targetScore,
	})
	if roomErr != nil {
		failureTraceID := traceID
		if failureTraceID == "" {
			failureTraceID = session.connectionTraceID
		}
		logging.Emit(observability.Request{
			Event: observability.EventNameRoomCreationFailed,
			Context: observability.Context{
				TraceID:   failureTraceID,
				SessionID: session.sessionID,
			},
			Fields: observability.Fields{
				"error_code":   roomErr.Code,
				"failure_mode": "create_single_player_room_failed",
				"reason_code":  roomErr.Code,
			},
		})
		session.EnqueueRoomError(traceID, roomErr.Code, roomErr.Message)
		return
	}

	logging.Emit(observability.Request{
		Event: observability.EventNameRoomCreated,
		Context: observability.Context{
			TraceID:   traceID,
			SessionID: session.sessionID,
			RoomID:    room.ID,
		},
		Fields: observability.Fields{"reason_code": "single_player_created"},
	})
	attachRoomSession(room, session.sessionID, session)
	session.bindRoom(room)
	if localProfileID != "" {
		room.SetMemberLocalProfileIDForSession(session.sessionID, localProfileID)
	}

	activateRoomPlayers(room)
	BroadcastRoomSnapshot(room)
}

func (session *webSocketSession) handleReturnToLobbyRequest() {
	context := session.sessionContext()
	if context.Room == nil || context.RoomID == "" || session.sessionID == "" {
		session.EnqueueRoomError("", rooms.RoomErrorNotInRoom, "Session is not in a room.")
		return
	}

	room, roomErr := session.rooms.ReturnRoomToLobby(context.RoomID, session.sessionID)
	if roomErr != nil {
		session.EnqueueRoomError("", roomErr.Code, roomErr.Message)
		return
	}

	deactivateRoomPlayers(room)
	BroadcastRoomSnapshot(room)
}
