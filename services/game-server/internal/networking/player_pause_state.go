package networking

import (
	"github.com/Lokee86/space-rocks/server/internal/logging"
	"github.com/Lokee86/space-rocks/server/internal/protocol/packetcodec"
)

func (session *webSocketSession) EnqueuePlayerPauseState() {
	if session.room == nil {
		return
	}
	gameInstance := session.room.GameInstance()
	if gameInstance == nil {
		return
	}
	if session.currentGamePlayerID == "" {
		return
	}

	packet, ok := gameInstance.PlayerPauseStatePacket(session.currentGamePlayerID)
	if !ok {
		return
	}

	payload, err := packetcodec.Encode(packet)
	if err != nil {
		logging.Network.Error("player pause state marshal failed", err,
			logging.FieldRoomID, session.currentRoomID,
			logging.FieldPlayerID, session.currentGamePlayerID,
			"session_id", session.sessionID,
		)
		return
	}

	session.enqueue(payload)
}
