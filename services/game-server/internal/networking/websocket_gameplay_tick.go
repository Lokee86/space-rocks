package networking

import (
	"time"

	"github.com/Lokee86/space-rocks/server/internal/constants"
	"github.com/Lokee86/space-rocks/server/internal/logging"
	"github.com/Lokee86/space-rocks/server/internal/rooms"
)

func tickSessionGameplayLifecycle(session *webSocketSession, done <-chan struct{}) {
	ticker := time.NewTicker(time.Second / time.Duration(constants.ServerTickRate))
	defer ticker.Stop()

	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			if session.currentGamePlayerID == "" {
				continue
			}

			if rooms.TickRoomGameOverLifecycle(session.room, BroadcastRoomSnapshot) {
				logging.Rooms.Info("room game-over lifecycle advanced; reporting match result",
					logging.FieldRoomID, session.currentRoomID,
					logging.FieldPlayerID, session.currentGamePlayerID,
					"session_id", session.sessionID,
				)
				rooms.ReportResolvedMatchResultOnce(session.room, session.matchResultReporter)
			}
		}
	}
}
