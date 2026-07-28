package networking

import (
	"github.com/Lokee86/space-rocks/services/game-server/internal/measurement"
	toolingrouter "github.com/Lokee86/space-rocks/services/game-server/internal/networking/tooling"
)

type packetObserver interface {
	ObservePacketWrite(toolingrouter.Context, string, string, int)
}

type receiverObserver interface {
	ObserveReceiverTick(toolingrouter.Context, measurement.ReceiverTickObservation)
}

func packetObserverFor(controller toolingrouter.MeasurementController) packetObserver {
	observer, _ := controller.(packetObserver)
	return observer
}

func receiverObserverFor(controller toolingrouter.MeasurementController) receiverObserver {
	observer, _ := controller.(receiverObserver)
	return observer
}

func (session *webSocketSession) observePacketWrite(lane string, packetFamily string, encodedBytes int) {
	if session == nil || session.packetObserver == nil {
		return
	}
	context := session.sessionContext()
	session.packetObserver.ObservePacketWrite(toolingrouter.Context{
		SessionID:    session.sessionID,
		RoomID:       context.RoomID,
		GamePlayerID: context.GamePlayerID,
	}, lane, packetFamily, encodedBytes)
}
