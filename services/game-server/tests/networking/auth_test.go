package networkingtests

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/Lokee86/space-rocks/server/internal/authclient"
	servergame "github.com/Lokee86/space-rocks/server/internal/game"
	"github.com/Lokee86/space-rocks/server/internal/networking"
	"github.com/Lokee86/space-rocks/server/internal/rooms"
	"github.com/gorilla/websocket"
)

type fakeTokenVerifier struct {
	receivedToken string
	result        authclient.VerifyResult
	err           error
}

func (v *fakeTokenVerifier) VerifyToken(_ context.Context, rawToken string) (authclient.VerifyResult, error) {
	v.receivedToken = rawToken
	return v.result, v.err
}

func TestWebSocketAuthenticateRequestReturnsAuthenticatedResult(t *testing.T) {
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

	var result struct {
		Type          string `json:"type"`
		Authenticated bool   `json:"authenticated"`
		UserID        int64  `json:"user_id"`
		DisplayName   string `json:"display_name"`
		ErrorCode     string `json:"error_code"`
		Message       string `json:"message"`
	}
	readJSON(t, conn, &result)

	if result.Type != "authenticate_result" {
		t.Fatalf("expected authenticate_result packet, got %q", result.Type)
	}
	if !result.Authenticated {
		t.Fatal("expected authenticated result")
	}
	if result.UserID != 123 {
		t.Fatalf("expected user id 123, got %d", result.UserID)
	}
	if result.DisplayName != "Ada" {
		t.Fatalf("expected display name Ada, got %q", result.DisplayName)
	}
	if verifier.receivedToken != "submitted-token" {
		t.Fatalf("expected verifier to receive submitted token, got %q", verifier.receivedToken)
	}
}

func TestWebSocketAuthenticateRequestRejectsInvalidToken(t *testing.T) {
	manager := networking.NewRoomManager()
	defer manager.StopAll()

	verifier := &fakeTokenVerifier{
		result: authclient.VerifyResult{Valid: false},
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

	var result struct {
		Type          string `json:"type"`
		Authenticated bool   `json:"authenticated"`
		ErrorCode     string `json:"error_code"`
	}
	readJSON(t, conn, &result)

	if result.Type != "authenticate_result" {
		t.Fatalf("expected authenticate_result packet, got %q", result.Type)
	}
	if result.Authenticated {
		t.Fatal("expected unauthenticated result")
	}
	if result.ErrorCode != "invalid_token" {
		t.Fatalf("expected error_code invalid_token, got %q", result.ErrorCode)
	}
}

func TestWebSocketAuthenticateRequestWithoutVerifierReturnsUnavailable(t *testing.T) {
	manager := networking.NewRoomManager()
	defer manager.StopAll()

	server := httptest.NewServer(networking.WebSocketHandler(manager))
	defer server.Close()

	conn, _, err := websocket.DefaultDialer.Dial(webSocketURL(server.URL), nil)
	if err != nil {
		t.Fatalf("dial websocket: %v", err)
	}
	defer conn.Close()

	if err := conn.WriteJSON(servergame.ClientPacket{Type: servergame.PacketTypeAuthenticateRequest, Token: "submitted-token"}); err != nil {
		t.Fatalf("write authenticate request: %v", err)
	}

	var result struct {
		Type          string `json:"type"`
		Authenticated bool   `json:"authenticated"`
		ErrorCode     string `json:"error_code"`
	}
	readJSON(t, conn, &result)

	if result.Type != "authenticate_result" {
		t.Fatalf("expected authenticate_result packet, got %q", result.Type)
	}
	if result.Authenticated {
		t.Fatal("expected unauthenticated result")
	}
	if result.ErrorCode != "token_verification_unavailable" {
		t.Fatalf("expected error_code token_verification_unavailable, got %q", result.ErrorCode)
	}
}

func TestWebSocketStartSinglePlayerRequestStillWorksWithoutAuthentication(t *testing.T) {
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
