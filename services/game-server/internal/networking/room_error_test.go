package networking

import (
	"encoding/json"
	"testing"

	"github.com/Lokee86/space-rocks/services/game-server/internal/game"
	"github.com/Lokee86/space-rocks/services/game-server/internal/rooms"
)

func TestEnqueueRoomErrorUsesOutboundQueue(t *testing.T) {
	session := &webSocketSession{
		sessionID: "session-test",
		outbound:  make(chan []byte, 1),
	}

	session.EnqueueRoomError("", rooms.RoomErrorRoomFull, "Room is full.")

	select {
	case payload := <-session.outbound:
		var packet game.RoomError
		if err := json.Unmarshal(payload, &packet); err != nil {
			t.Fatalf("decode room error packet: %v", err)
		}
		if packet.Type != game.PacketTypeRoomError {
			t.Fatalf("expected room error type %q, got %q", game.PacketTypeRoomError, packet.Type)
		}
		if packet.ErrorCode != rooms.RoomErrorRoomFull {
			t.Fatalf("expected error code %q, got %q", rooms.RoomErrorRoomFull, packet.ErrorCode)
		}
		if packet.Message != "Room is full." {
			t.Fatalf("expected room error message, got %q", packet.Message)
		}
	default:
		t.Fatal("expected room error to be queued")
	}
}
func TestEnqueueRoomErrorEchoesTraceID(t *testing.T) {
	session := &webSocketSession{
		sessionID: "session-trace",
		outbound:  make(chan []byte, 1),
	}

	session.EnqueueRoomError("trace-room-error", rooms.RoomErrorRoomFull, "Room is full.")

	select {
	case payload := <-session.outbound:
		var packet game.RoomError
		if err := json.Unmarshal(payload, &packet); err != nil {
			t.Fatalf("decode room error packet: %v", err)
		}
		if packet.TraceID != "trace-room-error" {
			t.Fatalf("expected room error trace %q, got %q", "trace-room-error", packet.TraceID)
		}
	default:
		t.Fatal("expected room error to be queued")
	}
}
