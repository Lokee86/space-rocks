package networking

import (
	"encoding/json"
	"testing"

	"github.com/Lokee86/space-rocks/server/internal/game"
	"github.com/Lokee86/space-rocks/server/internal/rooms"
)

func TestEnqueueRoomErrorUsesOutboundQueue(t *testing.T) {
	session := &webSocketSession{
		sessionID: "session-test",
		outbound:  make(chan []byte, 1),
	}

	session.EnqueueRoomError(RoomErrorRoomFull, "Room is full.")

	select {
	case payload := <-session.outbound:
		var packet game.RoomError
		if err := json.Unmarshal(payload, &packet); err != nil {
			t.Fatalf("decode room error packet: %v", err)
		}
		if packet.Type != game.PacketTypeRoomError {
			t.Fatalf("expected room error type %q, got %q", game.PacketTypeRoomError, packet.Type)
		}
		if packet.ErrorCode != RoomErrorRoomFull {
			t.Fatalf("expected error code %q, got %q", RoomErrorRoomFull, packet.ErrorCode)
		}
		if packet.Message != "Room is full." {
			t.Fatalf("expected room error message, got %q", packet.Message)
		}
	default:
		t.Fatal("expected room error to be queued")
	}
}

func TestRoomErrorCodes(t *testing.T) {
	codes := map[string]string{
		RoomErrorRoomNotFound:     rooms.RoomErrorRoomNotFound,
		RoomErrorRoomClosed:       rooms.RoomErrorRoomClosed,
		RoomErrorRoomInGame:       rooms.RoomErrorRoomInGame,
		RoomErrorRoomFull:         rooms.RoomErrorRoomFull,
		RoomErrorAlreadyInRoom:    rooms.RoomErrorAlreadyInRoom,
		RoomErrorNotInRoom:        rooms.RoomErrorNotInRoom,
		RoomErrorInvalidRoomCode:  rooms.RoomErrorInvalidRoomCode,
		RoomErrorNotReady:         rooms.RoomErrorNotReady,
		RoomErrorInvalidRoomState: rooms.RoomErrorInvalidRoomState,
	}

	for code, roomCode := range codes {
		if code == "" {
			t.Fatal("expected room error code to be non-empty")
		}
		if code != roomCode {
			t.Fatalf("expected networking error code %q to match rooms code %q", code, roomCode)
		}
	}
}
