package networking

import (
	"github.com/Lokee86/space-rocks/server/internal/logging"
	"github.com/Lokee86/space-rocks/server/internal/protocol/packetcodec"
)

func (session *webSocketSession) EnqueuePlayerPauseState() {
	if session.room == nil {
		return
	}
	if session.room.Game == nil {
		return
	}
	if session.currentPlayerID == "" {
		return
	}

	packet, ok := session.room.Game.PlayerPauseStatePacket(session.currentPlayerID)
	if !ok {
		return
	}

	payload, err := packetcodec.Encode(packet)
	if err != nil {
		logging.Network.Error("player pause state marshal failed", err,
			logging.FieldRoomID, session.currentRoomID,
			logging.FieldPlayerID, session.currentPlayerID,
			"session_id", session.sessionID,
		)
		return
	}

	session.enqueue(payload)
}
