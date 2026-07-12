package networking

import (
	"github.com/Lokee86/space-rocks/services/game-server/internal/networking/inbound"
	"github.com/Lokee86/space-rocks/services/game-server/internal/protocol/realtime"
)

type inboundSessionAdapter struct {
	session *webSocketSession
}

type queuedResyncRequest struct {
	Request    realtime.ResyncRequest
	RoomID     string
	ReceiverID string
	MatchID    string
}

func newInboundSessionAdapter(session *webSocketSession) inboundSessionAdapter {
	return inboundSessionAdapter{session: session}
}

func (a inboundSessionAdapter) CurrentSessionContext() inbound.SessionContext {
	context := a.session.sessionContext()
	return inbound.SessionContext{Room: context.Room, RoomID: context.RoomID, GamePlayerID: context.GamePlayerID}
}

func (a inboundSessionAdapter) SessionID() string {
	return a.session.sessionID
}

func (a inboundSessionAdapter) EnqueueOutboundMessage(message []byte) {
	a.session.outbound <- message
}

func (a inboundSessionAdapter) EnqueueResyncRequest(context inbound.SessionContext, request realtime.ResyncRequest) bool {
	select {
	case a.session.resyncRequests <- queuedResyncRequest{Request: request, RoomID: context.RoomID, ReceiverID: context.GamePlayerID, MatchID: request.MatchID}:
		return true
	default:
		return false
	}
}

func (a inboundSessionAdapter) LogLobbyPacketReceived(message string, roomCode string) {
	a.session.logLobbyPacketReceived(message, roomCode)
}

func (a inboundSessionAdapter) HandleAuthenticateRequest(token string) {
	a.session.handleAuthenticateRequest(token)
}

func (a inboundSessionAdapter) HandleCreateRoomRequest() {
	a.session.handleCreateRoomRequest()
}

func (a inboundSessionAdapter) HandleJoinRoomRequest(roomCode string) {
	a.session.handleJoinRoomRequest(roomCode)
}

func (a inboundSessionAdapter) HandleLeaveRoomRequest() {
	a.session.handleLeaveRoomRequest()
}

func (a inboundSessionAdapter) HandleSetReadyRequest(ready bool) {
	a.session.handleSetReadyRequest(ready)
}

func (a inboundSessionAdapter) HandleStartGameRequest() {
	a.session.handleStartGameRequest()
}

func (a inboundSessionAdapter) HandleStartSinglePlayerRequest(localProfileID string) {
	a.session.handleStartSinglePlayerRequest(localProfileID)
}

func (a inboundSessionAdapter) HandleReturnToLobbyRequest() {
	a.session.handleReturnToLobbyRequest()
}

func (a inboundSessionAdapter) EnqueuePlayerPauseState() {
	a.session.EnqueuePlayerPauseState()
}

func (a inboundSessionAdapter) HandleWebRTCOffer(descriptionType string, sdp string) {
	a.session.HandleWebRTCOffer(descriptionType, sdp)
}

func (a inboundSessionAdapter) HandleWebRTCIceCandidate(media string, index int, name string) {
	a.session.HandleWebRTCIceCandidate(media, index, name)
}

func (a inboundSessionAdapter) HandleWebRTCSmoke(smokeID string, origin string, message string) {
	a.session.HandleWebRTCSmoke(smokeID, origin, message)
}

func (a inboundSessionAdapter) HandleWebRTCFailed(errorCode string, message string) {
	a.session.HandleWebRTCFailed(errorCode, message)
}
