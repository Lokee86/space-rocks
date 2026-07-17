package networking

import (
	"github.com/Lokee86/space-rocks/services/game-server/internal/game"
	"github.com/Lokee86/space-rocks/services/game-server/internal/logging"
	"github.com/Lokee86/space-rocks/services/game-server/internal/protocol/packetcodec"
	observability "github.com/Lokee86/space-rocks/shared/go/observabilityevent"
)

func (session *webSocketSession) EnqueueRoomError(traceID string, errorCode string, message string) {
	packet := game.RoomError{
		Type:      game.PacketTypeRoomError,
		TraceID:   traceID,
		ErrorCode: errorCode,
		Message:   message,
	}
	payload, err := packetcodec.Encode(packet)
	if err != nil {
		if traceID == "" {
			traceID = session.connectionTraceID
		}
		context := session.sessionContext()
		logging.Emit(observability.Request{
			Event: observability.EventNameOutboundPacketEncodeFailed,
			Context: observability.Context{
				TraceID:    traceID,
				SessionID:  session.sessionID,
				RoomID:     context.RoomID,
				PacketType: game.PacketTypeRoomError,
			},
			Fields: observability.Fields{
				"error_code":   "room_error_encode_failed",
				"failure_mode": "room_error_encode_failed",
			},
		})
		return
	}

	session.enqueue(payload)
}
