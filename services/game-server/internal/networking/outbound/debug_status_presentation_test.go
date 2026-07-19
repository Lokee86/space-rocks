package outbound

import (
	"testing"

	"github.com/Lokee86/space-rocks/services/game-server/internal/devtools"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/physics"
	"github.com/Lokee86/space-rocks/services/game-server/internal/rooms"
)

func TestBuildDebugStatusPacketIncludesCorrelatedStatusPayload(t *testing.T) {
	if !devtools.Enabled() {
		t.Skip("debug status output is disabled by this build")
	}
	const (
		roomID    = "room-1"
		playerID  = "player-1"
		requestID = "request-1"
	)

	gameInstance := game.New()
	control := game.NewControl(gameInstance)
	if !control.EnsurePlayerSession(playerID, physics.Vector2{}) {
		t.Fatal("expected EnsurePlayerSession to succeed")
	}

	room := rooms.NewRoom(roomID, rooms.RoomStateInGame, gameInstance)
	packet, ok := BuildDebugStatusPacket(room, playerID, requestID)
	if !ok {
		t.Fatal("expected debug status packet to build")
	}
	if packet.Type != devtools.PacketTypeDebugStatus {
		t.Fatalf("type = %q, want %q", packet.Type, devtools.PacketTypeDebugStatus)
	}
	if packet.RequestID != requestID {
		t.Fatalf("request_id = %q, want %q", packet.RequestID, requestID)
	}
	if len(packet.DebugStatuses) == 0 {
		t.Fatal("expected debug statuses to be populated")
	}
}

func TestCanSendDebugStatusRejectsNilInputs(t *testing.T) {
	if CanSendDebugStatus(nil) {
		t.Fatal("expected nil room to be rejected")
	}

	if CanSendDebugStatus(rooms.NewRoom("room-1", rooms.RoomStateInGame, nil)) {
		t.Fatal("expected nil game instance to be rejected")
	}
}
