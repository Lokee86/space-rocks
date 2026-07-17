package networking

import (
	"time"

	"github.com/Lokee86/space-rocks/services/game-server/internal/constants"
	"github.com/Lokee86/space-rocks/services/game-server/internal/rooms"
)

func tickSessionGameplayLifecycle(session *webSocketSession, done <-chan struct{}) {
	ticker := time.NewTicker(time.Second / time.Duration(constants.ServerTickRate))
	defer ticker.Stop()

	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			context := session.sessionContext()
			if context.Room == nil || context.GamePlayerID == "" {
				continue
			}

			if rooms.TickRoomGameOverLifecycle(context.Room, BroadcastRoomSnapshot) {
				rooms.ReportResolvedMatchResultOnce(context.Room, session.matchResultReporter)
			}
		}
	}
}
