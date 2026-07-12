package networking

import (
	"encoding/json"

	"github.com/Lokee86/space-rocks/services/game-server/internal/networking/outbound"
	"github.com/Lokee86/space-rocks/services/game-server/internal/protocol/realtime"
)

var writeResyncMessage = outbound.WriteServerMessage

func writeResyncRequiredAndApply(session *webSocketSession, queued queuedResyncRequest, remoteAddr string) bool {
	if session == nil {
		return true
	}
	context := session.sessionContext()
	if context.Room == nil || context.GamePlayerID == "" {
		return true
	}
	gameplayContext := context.Room.GameplayContext()
	if !realtime.IsBaselineLane(queued.Request.Lane) || gameplayContext.Game == nil || queued.RoomID != context.RoomID || queued.ReceiverID != context.GamePlayerID || queued.MatchID == "" || queued.Request.MatchID == "" || queued.Request.MatchID != queued.MatchID {
		return true
	}
	matchID := gameplayContext.MatchID
	if queued.Request.MatchID != matchID {
		return true
	}
	resetRealtimeStateForContext(session, context, matchID)
	message, err := json.Marshal(map[string]any{
		"type":        realtime.PacketFamilyResyncRequired,
		"match_id":    matchID,
		"lane":        queued.Request.Lane,
		"baseline_id": queued.Request.BaselineID,
		"sequence":    queued.Request.Sequence,
		"reason":      queued.Request.Reason,
	})
	if err != nil || !session.sessionContextMatches(context) || !context.Room.GameplayContextMatches(gameplayContext) {
		return err == nil
	}
	if !writeResyncMessage(session.conn, message, func(err error) {
		logWebSocketWriteClose(err, context.RoomID, context.GamePlayerID, remoteAddr)
	}) {
		return false
	}
	if !session.sessionContextMatches(context) || !context.Room.GameplayContextMatches(gameplayContext) {
		return true
	}
	return session.realtimeState.RequireFullBaseline(queued.Request.Lane)
}
