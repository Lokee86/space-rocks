package roomstests

import (
	"testing"

	"github.com/Lokee86/space-rocks/server/internal/rooms"
)

func TestNewRoomMember(t *testing.T) {
	member := rooms.NewRoomMember("session-1")

	if member.SessionID != "session-1" {
		t.Fatalf("expected session id, got %q", member.SessionID)
	}
	if member.Ready {
		t.Fatal("expected new room member to start not ready")
	}
	if !member.Connected {
		t.Fatal("expected new room member to start connected")
	}
}

func TestRoomMemberSetReady(t *testing.T) {
	member := rooms.NewRoomMember("session-1")

	member.SetReady(true)
	if !member.Ready {
		t.Fatal("expected room member to be ready")
	}

	member.SetReady(false)
	if member.Ready {
		t.Fatal("expected room member to be not ready")
	}
}

func TestRoomMemberConnectionMarkers(t *testing.T) {
	member := rooms.NewRoomMember("session-1")

	member.MarkDisconnected()
	if member.Connected {
		t.Fatal("expected room member to be disconnected")
	}

	member.MarkConnected()
	if !member.Connected {
		t.Fatal("expected room member to be connected")
	}
}
