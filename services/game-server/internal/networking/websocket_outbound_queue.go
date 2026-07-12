package networking

import "github.com/Lokee86/space-rocks/services/game-server/internal/logging"

func (session *webSocketSession) enqueue(payload []byte) {
	select {
	case session.outbound <- payload:
		return
	default:
		session.outboundOverflowOnce.Do(func() {
			context := session.sessionContext()
			logging.Network.Warn("websocket outbound queue full; closing slow session",
				logging.FieldRoomID, context.RoomID,
				logging.FieldPlayerID, context.GamePlayerID,
				"session_id", session.sessionID,
			)
			if session.conn != nil {
				_ = session.conn.Close()
			}
		})
	}
}
