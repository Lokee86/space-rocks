package inbound

import "github.com/Lokee86/space-rocks/server/internal/game"

type lobbySession interface {
	LogLobbyPacketReceived(message string, roomCode string)
	HandleCreateRoomRequest()
	HandleJoinRoomRequest(roomCode string)
	HandleLeaveRoomRequest()
	HandleSetReadyRequest(ready bool)
	HandleStartGameRequest()
}

func HandleLobbyPacket(session lobbySession, packet game.ClientPacket) bool {
	switch packet.Type {
	case game.PacketTypeCreateRoomRequest:
		session.LogLobbyPacketReceived("CreateRoomRequest received", "")
		session.HandleCreateRoomRequest()
		return true
	case game.PacketTypeJoinRoomRequest:
		session.LogLobbyPacketReceived("JoinRoomRequest received", packet.RoomCode)
		session.HandleJoinRoomRequest(packet.RoomCode)
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
	default:
		return false
	}
}
