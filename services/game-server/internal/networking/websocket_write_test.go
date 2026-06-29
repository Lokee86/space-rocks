package networking

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Lokee86/space-rocks/server/internal/game"
	"github.com/Lokee86/space-rocks/server/internal/rooms"
	"github.com/gorilla/websocket"
)

func TestMaybeWriteDebugShapeCatalogSendsOncePerRoomSession(t *testing.T) {
	originalCanSend := canSendDebugShapeCatalog
	originalBuilder := buildDebugShapeCatalogResponse
	canSendDebugShapeCatalog = func(room *rooms.Room) bool {
		return true
	}
	buildDebugShapeCatalogResponse = func(room *rooms.Room, roomID string, remoteAddr string) ([]byte, bool) {
		return []byte(`{"type":"debug_shape_catalog","shapes":{}}`), true
	}
	t.Cleanup(func() {
		canSendDebugShapeCatalog = originalCanSend
		buildDebugShapeCatalogResponse = originalBuilder
	})

	serverConn, clientConn := newWebSocketTestConn(t)
	defer serverConn.Close()
	defer clientConn.Close()

	room := rooms.NewRoom("room-1", rooms.RoomStateInGame, game.New())
	session := &webSocketSession{
		conn:      serverConn,
		room:      room,
		rooms:     rooms.NewRoomManager(),
		currentRoomID: "room-1",
	}

	if !maybeWriteDebugShapeCatalog(session, "127.0.0.1:1234") {
		t.Fatal("expected first debug shape catalog write to succeed")
	}
	assertDebugShapeCatalogPacket(t, clientConn)

	if maybeWriteDebugShapeCatalog(session, "127.0.0.1:1234") {
		// no-op send still returns true; verify no duplicate packet instead.
	}
	assertNoMessageWithin(t, clientConn)

	room2 := rooms.NewRoom("room-2", rooms.RoomStateInGame, game.New())
	session.room = room2
	session.currentRoomID = "room-2"
	session.resetDebugShapeCatalogSent()

	if !maybeWriteDebugShapeCatalog(session, "127.0.0.1:1234") {
		t.Fatal("expected debug shape catalog write to succeed after reset")
	}
	assertDebugShapeCatalogPacket(t, clientConn)
}

func newWebSocketTestConn(t *testing.T) (*websocket.Conn, *websocket.Conn) {
	t.Helper()

	upgrader := websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
	serverConnCh := make(chan *websocket.Conn, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Fatalf("upgrade: %v", err)
		}
		serverConnCh <- conn
	}))
	t.Cleanup(server.Close)

	clientConn, _, err := websocket.DefaultDialer.Dial("ws"+server.URL[4:], nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	serverConn := <-serverConnCh
	return serverConn, clientConn
}

func assertDebugShapeCatalogPacket(t *testing.T, conn *websocket.Conn) {
	t.Helper()

	_, msg, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("expected debug shape catalog packet: %v", err)
	}

	var payload map[string]any
	if err := json.Unmarshal(msg, &payload); err != nil {
		t.Fatalf("expected valid json packet: %v", err)
	}
	if got := payload["type"]; got != "debug_shape_catalog" {
		t.Fatalf("expected debug shape catalog packet, got %v", got)
	}
}

func assertNoMessageWithin(t *testing.T, conn *websocket.Conn) {
	t.Helper()

	_ = conn.SetReadDeadline(time.Now().Add(50 * time.Millisecond))
	defer conn.SetReadDeadline(time.Time{})
	if _, _, err := conn.ReadMessage(); err == nil {
		t.Fatal("expected no duplicate debug shape catalog packet")
	}
}
