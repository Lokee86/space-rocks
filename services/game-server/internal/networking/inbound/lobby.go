package inbound

import "github.com/Lokee86/space-rocks/services/game-server/internal/game"

type lobbySession interface {
	HandleCreateRoomRequest(traceID string)
	HandleJoinRoomRequest(roomCode string, traceID string)
	HandleLeaveRoomRequest()
	HandleSetReadyRequest(ready bool)
	HandleStartGameRequest()
	HandleStartSinglePlayerRequest(localProfileID string, traceID string)
	HandleReturnToLobbyRequest()
}

func HandleLobbyPacket(session lobbySession, packet game.ClientPacket) bool {
	switch packet.Type {
	case game.PacketTypeCreateRoomRequest:
		session.HandleCreateRoomRequest(packet.TraceID)
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
	case game.PacketTypeStartGameRequest:
		session.HandleStartGameRequest()
		return true
	case game.PacketTypeStartSinglePlayerRequest:
		session.HandleStartSinglePlayerRequest(packet.LocalProfileID, packet.TraceID)
		return true
	case game.PacketTypeReturnToLobbyRequest:
		session.HandleReturnToLobbyRequest()
		return true
	default:
		return false
	}
}
