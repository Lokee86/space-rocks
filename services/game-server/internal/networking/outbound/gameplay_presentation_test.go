package outbound

import (
	"encoding/json"
	"testing"

	"github.com/Lokee86/space-rocks/server/internal/game"
	"github.com/Lokee86/space-rocks/server/internal/game/physics"
	"github.com/Lokee86/space-rocks/server/internal/rooms"
)

func TestBuildGameplayPresentationStateResponseExcludesDevtoolsFields(t *testing.T) {
	const (
		roomID   = "room-1"
		playerID = "player-1"
	)

	gameInstance := game.New()
	if !gameInstance.DevtoolsEnsurePlayerSession(playerID, physics.Vector2{}) {
		t.Fatal("expected DevtoolsEnsurePlayerSession to succeed")
	}

	room := rooms.NewRoom(roomID, rooms.RoomStateInGame, gameInstance)

	response, metrics, ok := BuildGameplayPresentationStateResponse(room, playerID, roomID, "127.0.0.1:1234")
	if !ok {
		t.Fatal("expected gameplay presentation state response to build")
	}

	if metrics.PacketSize != len(response) {
		t.Fatalf("expected packet size %d, got %d", len(response), metrics.PacketSize)
	}

	if metrics.PacketSize <= 0 {
		t.Fatalf("expected packet size to be positive, got %d", metrics.PacketSize)
	}

	if metrics.Contributors.RoomState != string(room.State) {
		t.Fatalf("expected room state %q, got %q", string(room.State), metrics.Contributors.RoomState)
	}

	if metrics.Contributors.PlayerSessions <= 0 {
		t.Fatalf("expected player sessions contributor to be greater than 0, got %d", metrics.Contributors.PlayerSessions)
	}

	if metrics.BuildDuration < 0 {
		t.Fatalf("expected non-negative build duration, got %v", metrics.BuildDuration)
	}

	if metrics.EncodeDuration < 0 {
		t.Fatalf("expected non-negative encode duration, got %v", metrics.EncodeDuration)
	}

	var payload map[string]any
	if err := json.Unmarshal(response, &payload); err != nil {
		t.Fatalf("expected response to decode as json: %v", err)
	}

	if got := payload["type"]; got != "state" {
		t.Fatalf("expected packet type %q, got %v", "state", got)
	}

	for _, key := range []string{"debug_status", "debug_statuses", "debug_collision_bodies"} {
		if _, ok := payload[key]; ok {
			t.Fatalf("expected packet to omit %q", key)
		}
	}
}
