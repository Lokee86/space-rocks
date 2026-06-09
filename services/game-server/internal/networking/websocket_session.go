package networking

import (
	"strconv"
	"sync/atomic"

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
}

func newWebSocketSession(conn *websocket.Conn, rooms *rooms.RoomManager, authVerifier TokenVerifier) *webSocketSession {
	sessionNumber := nextSessionID.Add(1)

	return &webSocketSession{
		conn:         conn,
		sessionID:    "session-" + strconv.FormatUint(sessionNumber, 10),
		rooms:        rooms,
		outbound:     make(chan []byte, 16),
		identity:     NewGuestSessionIdentity(),
		authVerifier: authVerifier,
	}
}

func (session *webSocketSession) SessionIdentity() SessionIdentity {
	return session.identity
}

func (session *webSocketSession) SetAuthenticatedAccountIdentity(userID int64, displayName string) {
	session.identity = NewAuthenticatedAccountIdentity(userID, displayName)
}
