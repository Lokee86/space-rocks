package networking

import (
	"github.com/Lokee86/space-rocks/services/game-server/internal/logging"
	observability "github.com/Lokee86/space-rocks/shared/go/observabilityevent"
)

func (session *webSocketSession) enqueue(payload []byte) {
	select {
	case session.outbound <- payload:
		return
	default:
		session.outboundOverflowOnce.Do(func() {
			context := session.sessionContext()
			logging.Emit(observability.Request{
				Event: observability.EventNameGameServerClientDisconnected,
				Context: observability.Context{
					TraceID:   session.traceID,
					SessionID: session.sessionID,
					RoomID:    context.RoomID,
					PlayerID:  context.GamePlayerID,
				},
				Fields: observability.Fields{
					"failure_mode": "outbound_queue_full",
				},
			})
			if session.conn != nil {
				_ = session.conn.Close()
			}
		})
	}
}
