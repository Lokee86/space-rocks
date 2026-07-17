package networking

import (
	"testing"
	"time"

	"github.com/Lokee86/space-rocks/services/game-server/internal/rooms"
)

func TestEnqueueSaturatedSessionReturnsPromptly(t *testing.T) {
	session := &webSocketSession{
		connectionTraceID: "550e8400-e29b-41d4-a716-446655440030",
		outbound:           make(chan []byte, 1),
	}
	session.outbound <- []byte("already queued")

	done := make(chan struct{})
	go func() {
		session.enqueue([]byte("overflow"))
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(1 * time.Second):
		t.Fatal("saturated session enqueue blocked")
	}
}

func TestBroadcastRoomSnapshotContinuesAfterFullSession(t *testing.T) {
	room := rooms.NewRoom("room", rooms.RoomStateLobby, nil)
	early := &webSocketSession{
		sessionID:         "session-1",
		connectionTraceID: "550e8400-e29b-41d4-a716-446655440031",
		outbound:          make(chan []byte, 1),
	}
	healthy := &webSocketSession{
		sessionID:         "session-2",
		connectionTraceID: "550e8400-e29b-41d4-a716-446655440032",
		outbound:          make(chan []byte, 1),
	}
	room.AddMemberSessionID(early.sessionID)
	room.AddMemberSessionID(healthy.sessionID)
	attachRoomSession(room, early.sessionID, early)
	attachRoomSession(room, healthy.sessionID, healthy)
	defer detachRoomSession(room, early.sessionID)
	defer detachRoomSession(room, healthy.sessionID)
	early.outbound <- []byte("full")

	done := make(chan struct{})
	go func() {
		BroadcastRoomSnapshot(room)
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(1 * time.Second):
		t.Fatal("room snapshot broadcast blocked on full session")
	}

	select {
	case <-healthy.outbound:
	default:
		t.Fatal("healthy session did not receive room snapshot")
	}
}
