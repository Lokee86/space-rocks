package networking

import (
	"encoding/json"

	"github.com/Lokee86/space-rocks/services/game-server/internal/networking/outbound"
	"github.com/Lokee86/space-rocks/services/game-server/internal/protocol/realtime"
)

var writeResyncMessage = outbound.WriteServerMessage

func writeResyncRequiredAndApply(session *webSocketSession, queued queuedResyncRequest, remoteAddr string) bool {
	request := queued.Request
	if !realtime.IsBaselineLane(request.Lane) || session == nil || session.room == nil || session.room.GameInstance() == nil || session.currentGamePlayerID == "" || queued.RoomID != session.room.ID || queued.ReceiverID != session.currentGamePlayerID {
		return true
	}
	if session.realtimeState.ReceiverID == "" || session.realtimeState.ReceiverID != session.currentGamePlayerID {
		session.realtimeState = realtime.NewRealtimeSessionState(session.currentGamePlayerID)
	}
	message, err := json.Marshal(map[string]any{
		"type":        realtime.PacketFamilyResyncRequired,
		"lane":        request.Lane,
		"baseline_id": request.BaselineID,
		"sequence":    request.Sequence,
		"reason":      request.Reason,
	})
	if err != nil {
		return false
	}
	if !writeResyncMessage(session.conn, message, func(err error) {
		logWebSocketWriteClose(err, session.currentRoomID, session.currentGamePlayerID, remoteAddr)
	}) {
		return false
	}
	return session.realtimeState.RequireFullBaseline(request.Lane)
}
