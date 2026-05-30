package networking

import (
	"github.com/Lokee86/space-rocks/server/internal/game"
	"github.com/Lokee86/space-rocks/server/internal/logging"
	"github.com/Lokee86/space-rocks/server/internal/protocol/packetcodec"
)

func readClientInput(
	session *webSocketSession,
	remoteAddr string,
	readErr chan<- error,
) {
	for {
		_, msg, err := session.conn.ReadMessage()
		if err != nil {
			readErr <- err
			return
		}

		var packet game.ClientPacket
		if err := packetcodec.Decode(msg, &packet); err != nil {
			logging.Network.Warn("websocket packet decode failed",
				logging.FieldError, err,
				logging.FieldRoomID, session.currentRoomID,
				logging.FieldPlayerID, session.currentGamePlayerID,
				"session_id", session.sessionID,
				logging.FieldRemoteAddr, remoteAddr,
			)
			continue
		}

		if packet.Type == game.PacketTypeCreateRoomRequest {
			session.logLobbyPacketReceived("CreateRoomRequest received", "")
			session.handleCreateRoomRequest()
			continue
		}
		if packet.Type == game.PacketTypeJoinRoomRequest {
			session.logLobbyPacketReceived("JoinRoomRequest received", packet.RoomCode)
			session.handleJoinRoomRequest(packet.RoomCode)
			continue
		}
		if packet.Type == game.PacketTypeLeaveRoomRequest {
			session.handleLeaveRoomRequest()
			continue
		}
		if packet.Type == game.PacketTypeSetReadyRequest {
			session.handleSetReadyRequest(packet.Ready)
			continue
		}
		if packet.Type == game.PacketTypeStartGameRequest {
			session.handleStartGameRequest()
			continue
		}
		if packet.Type == game.PacketTypeStartSinglePlayerRequest {
			session.handleStartSinglePlayerRequest()
			continue
		}
		if packet.Type == game.PacketTypeReturnToLobbyRequest {
			session.handleReturnToLobbyRequest()
			continue
		}

		if session.room == nil || session.currentGamePlayerID == "" {
			continue
		}

		session.room.Game.HandlePacket(session.currentGamePlayerID, packet)
		if isPauseStateRequest(packet.Type) {
			session.EnqueuePlayerPauseState()
		}
	}
}

func isPauseStateRequest(packetType string) bool {
	return packetType == game.PacketTypePauseRequest
}

func (session *webSocketSession) logLobbyPacketReceived(message string, roomCode string) {
	args := []any{
		logging.FieldRoomID, session.currentRoomID,
		logging.FieldPlayerID, session.currentGamePlayerID,
		"session_id", session.sessionID,
		"current_room_id", session.currentRoomID,
	}
	if roomCode != "" {
		args = append(args, "room_code", roomCode)
	}

	logging.Network.Debug(message, args...)
}
