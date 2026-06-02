package networking

import (
	"github.com/Lokee86/space-rocks/server/internal/networking/inbound"
	"github.com/Lokee86/space-rocks/server/internal/logging"
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
			logging.Network.Warn("websocket packet envelope decode failed",
				logging.FieldError, err,
				logging.FieldRoomID, session.currentRoomID,
				logging.FieldPlayerID, session.currentGamePlayerID,
				"session_id", session.sessionID,
				logging.FieldRemoteAddr, remoteAddr,
			)
			continue
		}
		handleClientPacket(session, remoteAddr, msg, envelope)
	}
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
