package networkingtests

import (
	"net/http/httptest"
	"testing"

	"github.com/Lokee86/space-rocks/server/internal/authclient"
	servergame "github.com/Lokee86/space-rocks/server/internal/game"
	"github.com/Lokee86/space-rocks/server/internal/networking"
	"github.com/Lokee86/space-rocks/server/internal/rooms"
	"github.com/gorilla/websocket"
)

func TestWebSocketCreateRoomRequiresAuthentication(t *testing.T) {
	manager := networking.NewRoomManager()
	defer manager.StopAll()

	server := httptest.NewServer(networking.WebSocketHandlerWithAuth(manager, &fakeTokenVerifier{}))
	defer server.Close()

	conn, _, err := websocket.DefaultDialer.Dial(webSocketURL(server.URL), nil)
	if err != nil {
		t.Fatalf("dial websocket: %v", err)
	}
	defer conn.Close()

	if err := conn.WriteJSON(servergame.ClientPacket{Type: servergame.PacketTypeCreateRoomRequest}); err != nil {
		t.Fatalf("write create room request: %v", err)
	}

	var roomError struct {
		Type      string `json:"type"`
		ErrorCode string `json:"error_code"`
		Message   string `json:"message"`
	}
	readJSON(t, conn, &roomError)

	if roomError.Type != servergame.PacketTypeRoomError {
		t.Fatalf("expected room error packet, got %q", roomError.Type)
	}
	if roomError.ErrorCode != rooms.RoomErrorAuthRequired {
		t.Fatalf("expected error code %q, got %q", rooms.RoomErrorAuthRequired, roomError.ErrorCode)
	}
}

func TestWebSocketJoinRoomRequiresAuthentication(t *testing.T) {
	manager := networking.NewRoomManager()
	defer manager.StopAll()

	server := httptest.NewServer(networking.WebSocketHandlerWithAuth(manager, &fakeTokenVerifier{}))
	defer server.Close()

	conn, _, err := websocket.DefaultDialer.Dial(webSocketURL(server.URL), nil)
	if err != nil {
		t.Fatalf("dial websocket: %v", err)
	}
	defer conn.Close()

	if err := conn.WriteJSON(servergame.ClientPacket{Type: servergame.PacketTypeJoinRoomRequest, RoomCode: "ABC123"}); err != nil {
		t.Fatalf("write join room request: %v", err)
	}

	var roomError struct {
		Type      string `json:"type"`
		ErrorCode string `json:"error_code"`
		Message   string `json:"message"`
	}
	readJSON(t, conn, &roomError)

	if roomError.Type != servergame.PacketTypeRoomError {
		t.Fatalf("expected room error packet, got %q", roomError.Type)
	}
	if roomError.ErrorCode != rooms.RoomErrorAuthRequired {
		t.Fatalf("expected error code %q, got %q", rooms.RoomErrorAuthRequired, roomError.ErrorCode)
	}
}

func TestWebSocketStartSinglePlayerRequestStillWorksWithoutAuthenticationAdmission(t *testing.T) {
	manager := networking.NewRoomManager()
	defer manager.StopAll()

	server := httptest.NewServer(networking.WebSocketHandler(manager))
	defer server.Close()

	conn, _, err := websocket.DefaultDialer.Dial(webSocketURL(server.URL), nil)
	if err != nil {
		t.Fatalf("dial websocket: %v", err)
	}
	defer conn.Close()

	if err := conn.WriteJSON(servergame.ClientPacket{Type: servergame.PacketTypeStartSinglePlayerRequest}); err != nil {
		t.Fatalf("write start single player request: %v", err)
	}

	var snapshot servergame.RoomSnapshot
	readJSON(t, conn, &snapshot)
	if snapshot.Type != servergame.PacketTypeRoomSnapshot {
		t.Fatalf("expected room snapshot packet, got %q", snapshot.Type)
	}
	if snapshot.RoomState != string(rooms.RoomStateInGame) {
		t.Fatalf("expected room state %q, got %q", rooms.RoomStateInGame, snapshot.RoomState)
	}
}

func TestAuthenticatedCreateRoomCreatesLobbyRoom(t *testing.T) {
	manager := networking.NewRoomManager()
	defer manager.StopAll()

	verifier := &fakeTokenVerifier{
		result: authclient.VerifyResult{
			Valid: true,
			Identity: authclient.Identity{
				UserID:      123,
				DisplayName: "Ada",
			},
		},
	}

	server := httptest.NewServer(networking.WebSocketHandlerWithAuth(manager, verifier))
	defer server.Close()

	conn, _, err := websocket.DefaultDialer.Dial(webSocketURL(server.URL), nil)
	if err != nil {
		t.Fatalf("dial websocket: %v", err)
	}
	defer conn.Close()

	if err := conn.WriteJSON(servergame.ClientPacket{Type: servergame.PacketTypeAuthenticateRequest, Token: "submitted-token"}); err != nil {
		t.Fatalf("write authenticate request: %v", err)
	}
	var authResult struct {
		Type          string `json:"type"`
		Authenticated bool   `json:"authenticated"`
		UserID        int64  `json:"user_id"`
		DisplayName   string `json:"display_name"`
		ErrorCode     string `json:"error_code"`
	}
	readJSON(t, conn, &authResult)

	if !authResult.Authenticated {
		t.Fatal("expected authenticated websocket")
	}

	if err := conn.WriteJSON(servergame.ClientPacket{Type: servergame.PacketTypeCreateRoomRequest}); err != nil {
		t.Fatalf("write create room request: %v", err)
	}

	var snapshot servergame.RoomSnapshot
	readJSON(t, conn, &snapshot)
	if snapshot.Type != servergame.PacketTypeRoomSnapshot {
		t.Fatalf("expected room snapshot packet, got %q", snapshot.Type)
	}
	if snapshot.RoomState != string(rooms.RoomStateLobby) {
		t.Fatalf("expected room state %q, got %q", rooms.RoomStateLobby, snapshot.RoomState)
	}
}

func TestAuthenticatedJoinRoomCanJoinExistingLobbyRoom(t *testing.T) {
	manager := networking.NewRoomManager()
	defer manager.StopAll()

	verifier := &fakeTokenVerifier{
		result: authclient.VerifyResult{
			Valid: true,
			Identity: authclient.Identity{
				UserID:      123,
				DisplayName: "Ada",
			},
		},
	}

	server := httptest.NewServer(networking.WebSocketHandlerWithAuth(manager, verifier))
	defer server.Close()

	creatorConn, _, err := websocket.DefaultDialer.Dial(webSocketURL(server.URL), nil)
	if err != nil {
		t.Fatalf("dial creator websocket: %v", err)
	}
	defer creatorConn.Close()

	if err := creatorConn.WriteJSON(servergame.ClientPacket{Type: servergame.PacketTypeAuthenticateRequest, Token: "creator-token"}); err != nil {
		t.Fatalf("write creator authenticate request: %v", err)
	}
	var creatorAuthResult struct {
		Type          string `json:"type"`
		Authenticated bool   `json:"authenticated"`
		UserID        int64  `json:"user_id"`
		DisplayName   string `json:"display_name"`
		ErrorCode     string `json:"error_code"`
	}
	readJSON(t, creatorConn, &creatorAuthResult)

	if err := creatorConn.WriteJSON(servergame.ClientPacket{Type: servergame.PacketTypeCreateRoomRequest}); err != nil {
		t.Fatalf("write creator create room request: %v", err)
	}
	var creatorSnapshot servergame.RoomSnapshot
	readJSON(t, creatorConn, &creatorSnapshot)
	if creatorSnapshot.RoomCode == "" {
		t.Fatal("expected generated room code")
	}

	joinerConn, _, err := websocket.DefaultDialer.Dial(webSocketURL(server.URL), nil)
	if err != nil {
		t.Fatalf("dial joiner websocket: %v", err)
	}
	defer joinerConn.Close()

	if err := joinerConn.WriteJSON(servergame.ClientPacket{Type: servergame.PacketTypeAuthenticateRequest, Token: "joiner-token"}); err != nil {
		t.Fatalf("write joiner authenticate request: %v", err)
	}
	var joinerAuthResult struct {
		Type          string `json:"type"`
		Authenticated bool   `json:"authenticated"`
		UserID        int64  `json:"user_id"`
		DisplayName   string `json:"display_name"`
		ErrorCode     string `json:"error_code"`
	}
	readJSON(t, joinerConn, &joinerAuthResult)

	if err := joinerConn.WriteJSON(servergame.ClientPacket{Type: servergame.PacketTypeJoinRoomRequest, RoomCode: creatorSnapshot.RoomCode}); err != nil {
		t.Fatalf("write join room request: %v", err)
	}

	var joinerSnapshot servergame.RoomSnapshot
	readJSON(t, joinerConn, &joinerSnapshot)
	if joinerSnapshot.Type != servergame.PacketTypeRoomSnapshot {
		t.Fatalf("expected room snapshot packet, got %q", joinerSnapshot.Type)
	}
	if joinerSnapshot.RoomCode != creatorSnapshot.RoomCode {
		t.Fatalf("expected room code %q, got %q", creatorSnapshot.RoomCode, joinerSnapshot.RoomCode)
	}
}
