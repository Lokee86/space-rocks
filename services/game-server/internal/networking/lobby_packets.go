package networking

import (
	"github.com/Lokee86/space-rocks/server/internal/game"
)

func handleLobbyPacket(session *webSocketSession, packet game.ClientPacket) bool {
	switch packet.Type {
	case game.PacketTypeCreateRoomRequest:
		session.logLobbyPacketReceived("CreateRoomRequest received", "")
		session.handleCreateRoomRequest()
		return true
	case game.PacketTypeJoinRoomRequest:
		session.logLobbyPacketReceived("JoinRoomRequest received", packet.RoomCode)
		session.handleJoinRoomRequest(packet.RoomCode)
		return true
	case game.PacketTypeLeaveRoomRequest:
		session.handleLeaveRoomRequest()
		return true
	case game.PacketTypeSetReadyRequest:
		session.handleSetReadyRequest(packet.Ready)
		return true
	case game.PacketTypeStartGameRequest:
		session.handleStartGameRequest()
		return true
	case game.PacketTypeStartSinglePlayerRequest:
		session.handleStartSinglePlayerRequest()
		return true
	case game.PacketTypeReturnToLobbyRequest:
		session.handleReturnToLobbyRequest()
		return true
	default:
		return false
	}
}
