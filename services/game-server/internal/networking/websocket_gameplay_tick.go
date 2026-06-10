package networking

import (
	"time"

	"github.com/Lokee86/space-rocks/server/internal/constants"
	"github.com/Lokee86/space-rocks/server/internal/networking/outbound"
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
			if session.currentGamePlayerID == "" || !outbound.CanSendGameplayPresentationState(session.room) {
				continue
			}

			if rooms.TickRoomGameOverLifecycle(session.room, BroadcastRoomSnapshot) {
				rooms.ReportResolvedMatchResultOnce(session.room, session.matchResultReporter)
			}
		}
	}
}
