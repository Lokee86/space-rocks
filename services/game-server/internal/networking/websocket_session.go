package networking

import (
	"strconv"
	"sync/atomic"

	"github.com/Lokee86/space-rocks/server/internal/protocol/realtime"
	"github.com/Lokee86/space-rocks/server/internal/rooms"
	"github.com/gorilla/websocket"
)

var nextSessionID atomic.Uint64

type webSocketSession struct {
	conn                *websocket.Conn
	sessionID           string
	currentRoomID       string
	currentGamePlayerID string
	room                *rooms.Room
	rooms               *rooms.RoomManager
	outbound            chan []byte
	identity            SessionIdentity
	authVerifier        TokenVerifier
	matchResultReporter rooms.MatchResultReporter
	realtimeState               realtime.RealtimeSessionState
	debugShapeCatalogSentRoomID string
}

func newWebSocketSession(conn *websocket.Conn, roomManager *rooms.RoomManager, authVerifier TokenVerifier, reporter rooms.MatchResultReporter) *webSocketSession {
	sessionNumber := nextSessionID.Add(1)
	if reporter == nil {
		reporter = rooms.NoopMatchResultReporter{}
	}

	return &webSocketSession{
		conn:                 conn,
		sessionID:            "session-" + strconv.FormatUint(sessionNumber, 10),
		rooms:                roomManager,
		outbound:             make(chan []byte, 16),
		identity:             NewGuestSessionIdentity(),
		authVerifier:         authVerifier,
		matchResultReporter: reporter,
	}
}

func (session *webSocketSession) SessionIdentity() SessionIdentity {
	return session.identity
}

func (session *webSocketSession) SetAuthenticatedAccountIdentity(userID int64, accountID string, displayName string) {
	session.identity = NewAuthenticatedAccountIdentity(userID, accountID, displayName)
}

func (session *webSocketSession) resetDebugShapeCatalogSent() {
	session.debugShapeCatalogSentRoomID = ""
}
