package networking

import (
	"github.com/Lokee86/space-rocks/services/game-server/internal/logging"
	"github.com/Lokee86/space-rocks/services/game-server/internal/protocol/packetcodec"
)

func (session *webSocketSession) EnqueuePlayerPauseState() {
	context := session.sessionContext()
	if context.Room == nil || context.GamePlayerID == "" {
		return
	}
	gameplayContext := context.Room.GameplayContext()
	if gameplayContext.Game == nil {
		return
	}

	packet, ok := gameplayContext.Game.PlayerPauseStatePacket(context.GamePlayerID)
	if !ok {
		return
	}

	payload, err := packetcodec.Encode(packet)
	if err != nil {
		logging.Network.Error("player pause state marshal failed", err,
			logging.FieldRoomID, context.RoomID,
			logging.FieldPlayerID, context.GamePlayerID,
			"session_id", session.sessionID,
		)
		return
	}

	if !session.sessionContextMatches(context) || !context.Room.GameplayContextMatches(gameplayContext) {
		return
	}

	session.enqueue(payload)
}
