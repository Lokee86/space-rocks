package networking

import (
	"encoding/json"
	"testing"

	"github.com/Lokee86/space-rocks/services/game-server/internal/game"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/modes"
	"github.com/Lokee86/space-rocks/services/game-server/internal/rooms"
)

func TestHandleStartSinglePlayerRequestCreatesRoom(t *testing.T) {
	session := &webSocketSession{
		sessionID:         "session-1",
		connectionTraceID: "550e8400-e29b-41d4-a716-446655440033",
		rooms:             rooms.NewRoomManagerWithCleanupDelay(0),
		outbound:          make(chan []byte, 1),
	}

	session.handleStartSinglePlayerRequest("", "", "", 0, false, 0)

	if session.sessionContext().RoomID == "" {
		t.Fatal("expected room to be created")
	}
	if session.sessionContext().Room == nil {
		t.Fatal("expected room reference to be stored")
	}
	if session.sessionContext().Room.State != rooms.RoomStateInGame {
		t.Fatalf("expected room state %q, got %q", rooms.RoomStateInGame, session.sessionContext().Room.State)
	}
	assertNoQueuedRoomErrorPacket(t, session.outbound)
}

func TestHandleStartSinglePlayerRequestAppliesSelectedMode(t *testing.T) {
	session := &webSocketSession{
		sessionID:         "session-1",
		connectionTraceID: "550e8400-e29b-41d4-a716-446655440034",
		rooms:             rooms.NewRoomManagerWithCleanupDelay(0),
		outbound:          make(chan []byte, 4),
	}

	session.handleStartSinglePlayerRequest("pilot-1", "", string(modes.PresetScoreAttack), 5, false, 2500)

	room := session.sessionContext().Room
	if room == nil {
		t.Fatal("expected room to be created")
	}
	config := room.ModeConfig()
	if config.PresetID != modes.PresetScoreAttack || config.StartingLives != 5 || config.InfiniteLives || config.TargetScore != 2500 {
		t.Fatalf("unexpected room mode config: %+v", config)
	}
	resolved, ok := room.ResolvedMatchRules()
	if !ok {
		t.Fatal("expected match rules to resolve at single-player start")
	}
	if resolved.ModeID != modes.ModeScoreAttack || resolved.ObjectivePolicy.TargetScore != 2500 || resolved.LivesPolicy.StartingLives != 5 {
		t.Fatalf("unexpected resolved match rules: %+v", resolved)
	}
	assertNoQueuedRoomErrorPacket(t, session.outbound)
}

func TestHandleStartGameRequestStartsRoom(t *testing.T) {
	manager := rooms.NewRoomManagerWithCleanupDelay(0)
	room, err := manager.CreateLobbyRoom()
	if err != nil {
		t.Fatalf("expected lobby room creation to succeed, got %v", err)
	}

	session := &webSocketSession{
		sessionID: "session-1",
		context:   SessionContext{Room: room, RoomID: room.ID, GamePlayerID: "player-1"},
		rooms:     manager,
		outbound:  make(chan []byte, 1),
	}
	addSessionMember(room, session.sessionID, session)
	if _, roomErr := manager.SetReady(room.ID, session.sessionID, true); roomErr != nil {
		t.Fatalf("expected ready state update to succeed, got %v", roomErr)
	}

	session.handleStartGameRequest()

	if session.sessionContext().Room.State != rooms.RoomStateInGame {
		t.Fatalf("expected room to enter in-game state, got %q", session.sessionContext().Room.State)
	}
	assertNoQueuedRoomErrorPacket(t, session.outbound)
}

func assertQueuedRoomError(t *testing.T, outbound chan []byte, wantCode string, wantMessage string) {
	t.Helper()

	select {
	case payload := <-outbound:
		var packet game.RoomError
		if err := json.Unmarshal(payload, &packet); err != nil {
			t.Fatalf("decode room error packet: %v", err)
		}
		if packet.Type != game.PacketTypeRoomError {
			t.Fatalf("expected room error type %q, got %q", game.PacketTypeRoomError, packet.Type)
		}
		if packet.ErrorCode != wantCode {
			t.Fatalf("expected error code %q, got %q", wantCode, packet.ErrorCode)
		}
		if packet.Message != wantMessage {
			t.Fatalf("expected message %q, got %q", wantMessage, packet.Message)
		}
	default:
		t.Fatal("expected a queued room error")
	}
}

func assertNoQueuedRoomErrorPacket(t *testing.T, outbound chan []byte) {
	t.Helper()

	for {
		select {
		case payload := <-outbound:
			var packet struct {
				Type string `json:"type"`
			}
			if err := json.Unmarshal(payload, &packet); err != nil {
				t.Fatalf("decode queued packet: %v", err)
			}
			if packet.Type == game.PacketTypeRoomError {
				t.Fatalf("expected no room error packet, got %s", string(payload))
			}
		default:
			return
		}
	}
}
