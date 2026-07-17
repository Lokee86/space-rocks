package networking

import (
	game "github.com/Lokee86/space-rocks/services/game-server/internal/game"
	"github.com/Lokee86/space-rocks/services/game-server/internal/logging"
	"github.com/Lokee86/space-rocks/services/game-server/internal/protocol/packetcodec"
	observability "github.com/Lokee86/space-rocks/shared/go/observabilityevent"
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
		logging.Emit(observability.Request{
			Event: observability.EventNameOutboundPacketEncodeFailed,
			Context: observability.Context{
				TraceID:    session.connectionTraceID,
				SessionID:  session.sessionID,
				RoomID:     context.RoomID,
				PlayerID:   context.GamePlayerID,
				PacketType: game.PacketTypePlayerPauseState,
			},
			Fields: observability.Fields{
				"error_code":   "player_pause_state_encode_failed",
				"failure_mode": "player_pause_state_encode_failed",
			},
		})
		return
	}

	if !session.sessionContextMatches(context) || !context.Room.GameplayContextMatches(gameplayContext) {
		return
	}

	session.enqueue(payload)
}
