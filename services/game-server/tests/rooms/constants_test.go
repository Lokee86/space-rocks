package roomstests

import (
	"testing"

	"github.com/Lokee86/space-rocks/server/internal/rooms"
)

func TestRoomDomainConstantsPreserveCurrentValues(t *testing.T) {
	if rooms.DefaultRoomID != "default" {
		t.Fatalf("expected default room id %q, got %q", "default", rooms.DefaultRoomID)
	}
	if rooms.MaxPlayersPerRoom != 8 {
		t.Fatalf("expected max players %d, got %d", 8, rooms.MaxPlayersPerRoom)
	}
	if rooms.RoomCodeLength != 6 {
		t.Fatalf("expected room code length %d, got %d", 6, rooms.RoomCodeLength)
	}
	if rooms.RoomCodeAlphabet == "" {
		t.Fatal("expected room code alphabet")
	}
}

func TestRoomStateWireValuesRemainStable(t *testing.T) {
	tests := map[rooms.RoomState]string{
		rooms.RoomStateLobby:    "Lobby",
		rooms.RoomStateStarting: "Starting",
		rooms.RoomStateInGame:   "InGame",
		rooms.RoomStateGameOver: "GameOver",
		rooms.RoomStateClosed:   "Closed",
	}

	for state, expected := range tests {
		if string(state) != expected {
			t.Fatalf("expected room state %q, got %q", expected, state)
		}
	}
}

func TestRoomErrorCodes(t *testing.T) {
	codes := []string{
		rooms.RoomErrorRoomNotFound,
		rooms.RoomErrorRoomClosed,
		rooms.RoomErrorRoomInGame,
		rooms.RoomErrorRoomFull,
		rooms.RoomErrorAlreadyInRoom,
		rooms.RoomErrorNotInRoom,
		rooms.RoomErrorInvalidRoomCode,
		rooms.RoomErrorNotReady,
		rooms.RoomErrorInvalidRoomState,
	}

	for _, code := range codes {
		if code == "" {
			t.Fatal("expected room error code")
		}
	}
}

func TestRoomCodeHelpers(t *testing.T) {
	if rooms.NormalizeRoomID("  ") != rooms.DefaultRoomID {
		t.Fatal("expected blank room id to normalize to default")
	}
	if rooms.NormalizeRoomCode(" abcd23 ") != "ABCD23" {
		t.Fatal("expected room code to trim and uppercase")
	}
	if !rooms.IsValidRoomCode("ABCD23") {
		t.Fatal("expected valid room code")
	}
	if rooms.IsValidRoomCode("abc") {
		t.Fatal("expected short room code to be invalid")
	}
}

func TestRoomMemberDefaults(t *testing.T) {
	member := rooms.RoomMember{SessionID: "session-1", Connected: true}
	if member.SessionID != "session-1" {
		t.Fatalf("expected session id, got %q", member.SessionID)
	}
	if member.Ready {
		t.Fatal("expected room member to default to not ready")
	}
	if !member.Connected {
		t.Fatal("expected room member to be connected")
	}
}
