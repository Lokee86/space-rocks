package networking

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Lokee86/space-rocks/server/internal/game"
	"github.com/Lokee86/space-rocks/server/internal/game/physics"
	"github.com/Lokee86/space-rocks/server/internal/game/runtime"
	"github.com/Lokee86/space-rocks/server/internal/protocol/realtime"
	"github.com/Lokee86/space-rocks/server/internal/rooms"
	"github.com/gorilla/websocket"
)

func TestMaybeWriteDebugShapeCatalogSendsOnlyOnceForSameRoom(t *testing.T) {
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
		conn:          serverConn,
		room:          room,
		rooms:         rooms.NewRoomManager(),
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
}

func TestMaybeWriteDebugShapeCatalogSendsAgainForNewRoomAfterReset(t *testing.T) {
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

	session := &webSocketSession{
		conn:                    serverConn,
		room:                    rooms.NewRoom("room-2", rooms.RoomStateInGame, game.New()),
		rooms:                   rooms.NewRoomManager(),
		currentRoomID:           "room-2",
		debugShapeCatalogSentRoomID: "room-1",
	}
	session.resetDebugShapeCatalogSent()

	if !maybeWriteDebugShapeCatalog(session, "127.0.0.1:1234") {
		t.Fatal("expected debug shape catalog write to succeed after reset")
	}
	assertDebugShapeCatalogPacket(t, clientConn)
}

func TestWriteGameplayLaneProtocolMessageWritesLanePacket(t *testing.T) {
	originalCanSend := canSendDebugShapeCatalog
	canSendDebugShapeCatalog = func(room *rooms.Room) bool {
		return false
	}
	t.Cleanup(func() {
		canSendDebugShapeCatalog = originalCanSend
	})

	serverConn, clientConn := newWebSocketTestConn(t)
	defer serverConn.Close()
	defer clientConn.Close()

	gameInstance := game.New()
	playerID := "player-1"
	if !gameInstance.DevtoolsEnsurePlayerSession(playerID, physics.Vector2{}) {
		t.Fatal("expected DevtoolsEnsurePlayerSession to succeed")
	}
	if !gameInstance.DevtoolsSpawnPlayerShip(playerID, physics.Vector2{}, runtime.ClientConfig{
		VisibleWorldWidth:  1280,
		VisibleWorldHeight: 720,
	}) {
		t.Fatal("expected DevtoolsSpawnPlayerShip to succeed")
	}

	room := rooms.NewRoom("room-1", rooms.RoomStateInGame, gameInstance)
	session := &webSocketSession{
		conn:                serverConn,
		room:                room,
		rooms:               rooms.NewRoomManager(),
		currentRoomID:       room.ID,
		currentGamePlayerID: playerID,
	}

	if !writeGameplayLaneProtocolMessage(session, "127.0.0.1:1234") {
		t.Fatal("expected lane protocol write to succeed")
	}

	assertLanePacket(t, clientConn)
}

func TestWriteGameplayLaneProtocolMessageStoresBaselineProjectionAfterSuccessfulWrite(t *testing.T) {
	originalCanSend := canSendDebugShapeCatalog
	canSendDebugShapeCatalog = func(room *rooms.Room) bool {
		return false
	}
	t.Cleanup(func() {
		canSendDebugShapeCatalog = originalCanSend
	})

	serverConn, clientConn := newWebSocketTestConn(t)
	defer serverConn.Close()
	defer clientConn.Close()

	gameInstance := game.New()
	playerID := "player-1"
	if !gameInstance.DevtoolsEnsurePlayerSession(playerID, physics.Vector2{}) {
		t.Fatal("expected DevtoolsEnsurePlayerSession to succeed")
	}
	if !gameInstance.DevtoolsSpawnPlayerShip(playerID, physics.Vector2{}, runtime.ClientConfig{
		VisibleWorldWidth:  1280,
		VisibleWorldHeight: 720,
	}) {
		t.Fatal("expected DevtoolsSpawnPlayerShip to succeed")
	}

	state := realtime.NewRealtimeSessionState(playerID)
	state.UpdateLane(realtime.LaneWorld, realtime.Metadata{Lane: realtime.LaneWorld, Sequence: 1, BaselineID: "world-baseline", SnapshotID: "world-baseline", SnapshotKind: realtime.SnapshotKind("full"), IsFinalChunk: true})
	state.MarkBaselineReady(realtime.LaneWorld)
	state.UpdateLane(realtime.LaneOverlay, realtime.Metadata{Lane: realtime.LaneOverlay, Sequence: 1, BaselineID: "overlay-baseline", SnapshotID: "overlay-baseline", SnapshotKind: realtime.SnapshotKind("full"), IsFinalChunk: true})
	state.MarkBaselineReady(realtime.LaneOverlay)
	state.UpdateLane(realtime.LaneSession, realtime.Metadata{Lane: realtime.LaneSession, Sequence: 1, BaselineID: "session-baseline", SnapshotID: "session-baseline", SnapshotKind: realtime.SnapshotKind("full"), IsFinalChunk: true})
	state.MarkBaselineReady(realtime.LaneSession)

	if projection, ok := state.BaselineProjection(realtime.LaneWorld); ok || projection != nil {
		t.Fatalf("expected no stored world projection before write, got %#v, %t", projection, ok)
	}
	if projection, ok := state.BaselineProjection(realtime.LaneOverlay); ok || projection != nil {
		t.Fatalf("expected no stored overlay projection before write, got %#v, %t", projection, ok)
	}
	if projection, ok := state.BaselineProjection(realtime.LaneSession); ok || projection != nil {
		t.Fatalf("expected no stored session projection before write, got %#v, %t", projection, ok)
	}

	room := rooms.NewRoom("room-1", rooms.RoomStateInGame, gameInstance)
	session := &webSocketSession{
		conn:                serverConn,
		room:                room,
		rooms:               rooms.NewRoomManager(),
		currentRoomID:       room.ID,
		currentGamePlayerID: playerID,
		realtimeState:       state,
	}

	if !writeGameplayLaneProtocolMessage(session, "127.0.0.1:1234") {
		t.Fatal("expected lane protocol write to succeed")
	}

	assertLanePacket(t, clientConn)
	assertStoredBaselineProjectionType(t, session.realtimeState, realtime.LaneWorld, "world_full")
	assertStoredBaselineProjectionType(t, session.realtimeState, realtime.LaneOverlay, "overlay_full")
	assertStoredBaselineProjectionType(t, session.realtimeState, realtime.LaneSession, "session_full")
}


func assertStoredBaselineProjectionType(t *testing.T, state realtime.RealtimeSessionState, lane realtime.Lane, wantType string) {
	t.Helper()
	projection, ok := state.BaselineProjection(lane)
	if !ok {
		t.Fatalf("expected stored projection for lane=%q", lane)
	}

	switch lane {
	case realtime.LaneWorld:
		packet, ok := projection.(realtime.WorldFullPacket)
		if !ok {
			t.Fatalf("expected world projection to be realtime.WorldFullPacket, got %#v", projection)
		}
		if packet.Type != wantType {
			t.Fatalf("expected world projection type=%q, got %q", wantType, packet.Type)
		}
	case realtime.LaneOverlay:
		packet, ok := projection.(realtime.OverlayFullPacket)
		if !ok {
			t.Fatalf("expected overlay projection to be realtime.OverlayFullPacket, got %#v", projection)
		}
		if packet.Type != wantType {
			t.Fatalf("expected overlay projection type=%q, got %q", wantType, packet.Type)
		}
	case realtime.LaneSession:
		packet, ok := projection.(realtime.SessionFullPacket)
		if !ok {
			t.Fatalf("expected session projection to be realtime.SessionFullPacket, got %#v", projection)
		}
		if packet.Type != wantType {
			t.Fatalf("expected session projection type=%q, got %q", wantType, packet.Type)
		}
	default:
		t.Fatalf("unexpected lane for stored projection assertion: %q", lane)
	}
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

func assertLanePacket(t *testing.T, conn *websocket.Conn) {
	t.Helper()

	_, msg, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("expected lane packet: %v", err)
	}

	var payload map[string]any
	if err := json.Unmarshal(msg, &payload); err != nil {
		t.Fatalf("expected valid json packet: %v", err)
	}

	packetType, _ := payload["type"].(string)
	switch packetType {
	case "world_full", "world_delta", "overlay_full", "overlay_delta", "session_full", "session_delta", "event_batch", "resync_request", "resync_required":
	default:
		t.Fatalf("expected lane packet type, got %v", packetType)
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
