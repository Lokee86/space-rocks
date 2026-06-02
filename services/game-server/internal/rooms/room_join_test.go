package rooms

import (
	"fmt"
	"testing"
)

func TestValidateJoinAcceptsLobbyRoom(t *testing.T) {
	room := NewRoom("room", RoomStateLobby, nil)

	if err := room.ValidateJoin(); err != nil {
		t.Fatalf("expected lobby room join validation to succeed, got %v", err)
	}
}

func TestValidateJoinRejectsNonJoinableRoom(t *testing.T) {
	room := NewRoom("room", RoomStateLobby, nil)
	room.SetJoinable(false)

	err := room.ValidateJoin()
	if err == nil {
		t.Fatal("expected non-joinable room join validation to fail")
	}
	if err.Code != RoomErrorInvalidRoomState {
		t.Fatalf("expected error code %q, got %q", RoomErrorInvalidRoomState, err.Code)
	}
}

func TestValidateJoinRejectsFullRoom(t *testing.T) {
	room := NewRoom("room", RoomStateLobby, nil)
	for index := 0; index < MaxPlayersPerRoom; index++ {
		room.AddMember(NewRoomMember(fmt.Sprintf("session-%d", index+1)))
	}

	err := room.ValidateJoin()
	if err == nil {
		t.Fatal("expected full room join validation to fail")
	}
	if err.Code != RoomErrorRoomFull {
		t.Fatalf("expected error code %q, got %q", RoomErrorRoomFull, err.Code)
	}
}

func TestValidateJoinRejectsStartingAndInGameRooms(t *testing.T) {
	tests := []struct {
		name  string
		state RoomState
	}{
		{
			name:  "starting",
			state: RoomStateStarting,
		},
		{
			name:  "in-game",
			state: RoomStateInGame,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			room := NewRoom("room", test.state, nil)

			err := room.ValidateJoin()
			if err == nil {
				t.Fatalf("expected %s room join validation to fail", test.name)
			}
			if err.Code != RoomErrorRoomInGame {
				t.Fatalf("expected error code %q, got %q", RoomErrorRoomInGame, err.Code)
			}
		})
	}
}
