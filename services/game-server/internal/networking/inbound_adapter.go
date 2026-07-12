package networking

import (
	"github.com/Lokee86/space-rocks/services/game-server/internal/protocol/realtime"
	"github.com/Lokee86/space-rocks/services/game-server/internal/rooms"
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

func (a inboundSessionAdapter) CurrentRoomID() string {
	return a.session.currentRoomID
}

func (a inboundSessionAdapter) CurrentRoom() *rooms.Room {
	return a.session.room
}

func (a inboundSessionAdapter) CurrentGamePlayerID() string {
	return a.session.currentGamePlayerID
}

func (a inboundSessionAdapter) SessionID() string {
	return a.session.sessionID
}

func (a inboundSessionAdapter) EnqueueOutboundMessage(message []byte) {
	a.session.outbound <- message
}

func (a inboundSessionAdapter) EnqueueResyncRequest(request realtime.ResyncRequest) bool {
	roomID := ""
	if a.session.room != nil {
		roomID = a.session.room.ID
	}
	matchID := ""
	if a.session.room != nil {
		matchID = a.session.room.CurrentMatchID()
	}
	select {
	case a.session.resyncRequests <- queuedResyncRequest{Request: request, RoomID: roomID, ReceiverID: a.session.currentGamePlayerID, MatchID: matchID}:
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
