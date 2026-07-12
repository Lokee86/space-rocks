package networking

import (
	"github.com/Lokee86/space-rocks/services/game-server/internal/logging"
	"github.com/Lokee86/space-rocks/services/game-server/internal/networking/inbound"
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

		envelope, err := inbound.DecodeClientPacketEnvelope(msg)
		if err != nil {
			context := session.sessionContext()
			logging.Network.Warn("websocket packet envelope decode failed",
				logging.FieldError, err,
				logging.FieldRoomID, context.RoomID,
				logging.FieldPlayerID, context.GamePlayerID,
				"session_id", session.sessionID,
				logging.FieldRemoteAddr, remoteAddr,
			)
			continue
		}
		handleClientPacket(session, remoteAddr, msg, envelope)
	}
}

func (session *webSocketSession) logLobbyPacketReceived(message string, roomCode string) {
	context := session.sessionContext()
	args := []any{
		logging.FieldRoomID, context.RoomID,
		logging.FieldPlayerID, context.GamePlayerID,
		"session_id", session.sessionID,
		"current_room_id", context.RoomID,
	}
	if roomCode != "" {
		args = append(args, "room_code", roomCode)
	}

	logging.Network.Debug(message, args...)
}
