package outbound

import (
	"encoding/json"
	"testing"

	"github.com/Lokee86/space-rocks/server/internal/game"
	"github.com/Lokee86/space-rocks/server/internal/game/physics"
	"github.com/Lokee86/space-rocks/server/internal/rooms"
)

func TestBuildDebugStatusResponseIncludesDebugStatusPayload(t *testing.T) {
	const (
		roomID   = "room-1"
		playerID = "player-1"
	)

	gameInstance := game.New()
	if !gameInstance.DevtoolsEnsurePlayerSession(playerID, physics.Vector2{}) {
		t.Fatal("expected DevtoolsEnsurePlayerSession to succeed")
	}

	room := rooms.NewRoom(roomID, rooms.RoomStateInGame, gameInstance)

	response, ok := BuildDebugStatusResponse(room, playerID, roomID, "127.0.0.1:1234")
	if !ok {
		t.Fatal("expected debug status response to build")
	}

	var payload map[string]any
	if err := json.Unmarshal(response, &payload); err != nil {
		t.Fatalf("expected response to decode as json: %v", err)
	}

	if got := payload["type"]; got != "debug_status" {
		t.Fatalf("expected packet type %q, got %v", "debug_status", got)
	}

	if _, ok := payload["debug_status"]; !ok {
		t.Fatal("expected debug_status to exist")
	}

	if _, ok := payload["debug_statuses"]; !ok {
		t.Fatal("expected debug_statuses to exist")
	}

	if _, ok := payload["debug_collision_bodies"]; ok {
		t.Fatal("expected debug_collision_bodies to be absent")
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
