package outbound

import (
	"encoding/json"
	"testing"

	"github.com/Lokee86/space-rocks/server/internal/game"
	"github.com/Lokee86/space-rocks/server/internal/rooms"
)

func TestBuildDebugShapeCatalogResponseIncludesShapeCatalogPayload(t *testing.T) {
	room := rooms.NewRoom("room-1", rooms.RoomStateInGame, game.New())

	response, ok := BuildDebugShapeCatalogResponse(room, "room-1", "127.0.0.1:1234")
	if !ok {
		t.Fatal("expected debug shape catalog response to build")
	}

	var payload map[string]any
	if err := json.Unmarshal(response, &payload); err != nil {
		t.Fatalf("expected response to decode as json: %v", err)
	}

	if got := payload["type"]; got != "debug_shape_catalog" {
		t.Fatalf("expected packet type %q, got %v", "debug_shape_catalog", got)
	}

	shapes, ok := payload["shapes"].(map[string]any)
	if !ok {
		t.Fatal("expected shapes to exist as an object")
	}
	if len(shapes) == 0 {
		t.Fatal("expected shapes to be non-empty")
	}

	if _, ok := payload["debug_collision_bodies"]; ok {
		t.Fatal("expected debug_collision_bodies to be absent")
	}
	if _, ok := payload["players"]; ok {
		t.Fatal("expected players to be absent")
	}
	if _, ok := payload["asteroids"]; ok {
		t.Fatal("expected asteroids to be absent")
	}
	if _, ok := payload["bullets"]; ok {
		t.Fatal("expected bullets to be absent")
	}
	if _, ok := payload["pickups"]; ok {
		t.Fatal("expected pickups to be absent")
	}
}
