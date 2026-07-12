package networking

import (
	"github.com/Lokee86/space-rocks/services/game-server/internal/game"
	"github.com/Lokee86/space-rocks/services/game-server/internal/logging"
	"github.com/Lokee86/space-rocks/services/game-server/internal/protocol/packetcodec"
)

func (session *webSocketSession) EnqueueRoomError(errorCode string, message string) {
	packet := game.RoomError{
		Type:      game.PacketTypeRoomError,
		ErrorCode: errorCode,
		Message:   message,
	}
	payload, err := packetcodec.Encode(packet)
	if err != nil {
		logging.Network.Error("room error marshal failed", err,
			"session_id", session.sessionID,
			"error_code", errorCode,
		)
		return
	}

	session.enqueue(payload)
}
