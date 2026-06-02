package networking

type inboundSessionAdapter struct {
	session *webSocketSession
}

func newInboundSessionAdapter(session *webSocketSession) inboundSessionAdapter {
	return inboundSessionAdapter{session: session}
}

func (a inboundSessionAdapter) CurrentRoomID() string {
	return a.session.currentRoomID
}

func (a inboundSessionAdapter) CurrentGamePlayerID() string {
	return a.session.currentGamePlayerID
}

func (a inboundSessionAdapter) SessionID() string {
	return a.session.sessionID
}

func (a inboundSessionAdapter) OutboundMessages() chan<- []byte {
	return a.session.outbound
}

func (a inboundSessionAdapter) LogLobbyPacketReceived(message string, roomCode string) {
	a.session.logLobbyPacketReceived(message, roomCode)
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
