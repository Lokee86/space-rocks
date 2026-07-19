package outbound

import (
	"testing"

	"github.com/Lokee86/space-rocks/services/game-server/internal/devtools"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game"
	"github.com/Lokee86/space-rocks/services/game-server/internal/rooms"
)

func TestBuildDebugShapeCatalogPacketIncludesCorrelatedShapeCatalog(t *testing.T) {
	if !devtools.Enabled() {
		t.Skip("debug shape catalog output is disabled by this build")
	}
	const requestID = "request-1"
	room := rooms.NewRoom("room-1", rooms.RoomStateInGame, game.New())

	packet, ok := BuildDebugShapeCatalogPacket(room, "room-1", requestID)
	if !ok {
		t.Fatal("expected debug shape catalog packet to build")
	}
	if packet.Type != devtools.PacketTypeDebugShapeCatalog {
		t.Fatalf("type = %q, want %q", packet.Type, devtools.PacketTypeDebugShapeCatalog)
	}
	if packet.RequestID != requestID {
		t.Fatalf("request_id = %q, want %q", packet.RequestID, requestID)
	}
	if len(packet.Shapes) == 0 {
		t.Fatal("expected shapes to be non-empty")
	}
}
