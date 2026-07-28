package networking

import (
	"github.com/Lokee86/space-rocks/services/game-server/internal/measurement"
	toolingrouter "github.com/Lokee86/space-rocks/services/game-server/internal/networking/tooling"
)

func (session *webSocketSession) observeReceiverTick(observation measurement.ReceiverTickObservation) {
	if session == nil || session.receiverObserver == nil {
		return
	}
	context := session.sessionContext()
	session.receiverObserver.ObserveReceiverTick(toolingrouter.Context{
		SessionID:    session.sessionID,
		RoomID:       context.RoomID,
		GamePlayerID: context.GamePlayerID,
	}, observation)
}
