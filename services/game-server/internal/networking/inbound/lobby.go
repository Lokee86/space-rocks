package inbound

import "github.com/Lokee86/space-rocks/services/game-server/internal/game"

type lobbySession interface {
	HandleCreateRoomRequest(traceID string, teamStructure string, teamAssignmentMode string, teamCount int, maxPlayers int, presetID string, startingLives int, infiniteLives bool, targetScore int)
	HandleJoinRoomRequest(roomCode string, traceID string)
	HandleLeaveRoomRequest()
	HandleSetReadyRequest(ready bool)
	HandleSetTeamAssignmentRequest(targetPlayerID string, teamID string, traceID string)
	HandleStartGameRequest()
	HandleAddBotRequest()
	HandleRemoveRoomMemberRequest(playerID string)
	HandleStartSinglePlayerRequest(localProfileID string, traceID string, presetID string, startingLives int, infiniteLives bool, targetScore int)
	HandleReturnToLobbyRequest()
}

func HandleLobbyPacket(session lobbySession, packet game.ClientPacket) bool {
	switch packet.Type {
	case game.PacketTypeCreateRoomRequest:
		session.HandleCreateRoomRequest(packet.TraceID, packet.TeamStructure, packet.TeamAssignmentMode, packet.TeamCount, packet.MaxPlayers, packet.PresetID, packet.StartingLives, packet.InfiniteLives, packet.TargetScore)
		return true
	case game.PacketTypeJoinRoomRequest:
		session.HandleJoinRoomRequest(packet.RoomCode, packet.TraceID)
		return true
	case game.PacketTypeLeaveRoomRequest:
		session.HandleLeaveRoomRequest()
		return true
	case game.PacketTypeSetReadyRequest:
		session.HandleSetReadyRequest(packet.Ready)
		return true
	case game.PacketTypeSetTeamAssignmentRequest:
		session.HandleSetTeamAssignmentRequest(packet.TargetPlayerID, packet.TeamID, packet.TraceID)
		return true
	case game.PacketTypeStartGameRequest:
		session.HandleStartGameRequest()
		return true
	case game.PacketTypeAddBotRequest:
		session.HandleAddBotRequest()
		return true
	case game.PacketTypeRemoveRoomMemberRequest:
		session.HandleRemoveRoomMemberRequest(packet.PlayerID)
		return true
	case game.PacketTypeStartSinglePlayerRequest:
		session.HandleStartSinglePlayerRequest(packet.LocalProfileID, packet.TraceID, packet.PresetID, packet.StartingLives, packet.InfiniteLives, packet.TargetScore)
		return true
	case game.PacketTypeReturnToLobbyRequest:
		session.HandleReturnToLobbyRequest()
		return true
	default:
		return false
	}
}
