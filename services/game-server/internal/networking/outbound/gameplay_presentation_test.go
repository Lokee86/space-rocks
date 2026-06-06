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

	response, ok := BuildGameplayPresentationStateResponse(room, playerID, roomID, "127.0.0.1:1234")
	if !ok {
		t.Fatal("expected gameplay presentation state response to build")
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
