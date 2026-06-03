package rooms

import (
	"testing"

	"github.com/Lokee86/space-rocks/server/internal/game"
)

func TestStartGameForMemberMovesLobbyRoomToInGame(t *testing.T) {
	room := NewRoom("room", RoomStateLobby, nil)
	owner := room.AddMember(NewRoomMember("session-owner"))
	owner.SetReady(true)
	peer := room.AddMember(NewRoomMember("session-peer"))
	peer.SetReady(true)

	newGame := func() *game.Game { return game.New() }

	if err := room.StartGameForMember("Player-1", newGame); err != nil {
		t.Fatalf("expected start to succeed, got %v", err)
	}
	if room.State != RoomStateInGame {
		t.Fatalf("expected room state %q, got %q", RoomStateInGame, room.State)
	}
	if room.GameInstance() == nil {
		t.Fatal("expected game to be created")
	}
	room.GameInstance().Stop()
}

func TestStartGameForMemberRejectsNonOwner(t *testing.T) {
	room := NewRoom("room", RoomStateLobby, nil)
	owner := room.AddMember(NewRoomMember("session-owner"))
	owner.SetReady(true)
	peer := room.AddMember(NewRoomMember("session-peer"))
	peer.SetReady(true)

	err := room.StartGameForMember("Player-2", func() *game.Game { return game.New() })
	if err == nil {
		t.Fatal("expected non-owner start to fail")
	}
	if err.Code != RoomErrorNotRoomOwner {
		t.Fatalf("expected error code %q, got %q", RoomErrorNotRoomOwner, err.Code)
	}
}

func TestStartGameForMemberRejectsUnreadyConnectedMember(t *testing.T) {
	room := NewRoom("room", RoomStateLobby, nil)
	owner := room.AddMember(NewRoomMember("session-owner"))
	owner.SetReady(true)
	peer := room.AddMember(NewRoomMember("session-peer"))
	peer.SetReady(false)

	err := room.StartGameForMember("Player-1", func() *game.Game { return game.New() })
	if err == nil {
		t.Fatal("expected unready connected member to block start")
	}
	if err.Code != RoomErrorNotReady {
		t.Fatalf("expected error code %q, got %q", RoomErrorNotReady, err.Code)
	}
}

func TestStartGameForMemberRejectsNonLobbyRoom(t *testing.T) {
	room := NewRoom("room", RoomStateStarting, nil)
	room.AddMember(NewRoomMember("session-owner"))

	err := room.StartGameForMember("Player-1", func() *game.Game { return game.New() })
	if err == nil {
		t.Fatal("expected non-lobby start to fail")
	}
	if err.Code != RoomErrorRoomInGame {
		t.Fatalf("expected error code %q, got %q", RoomErrorRoomInGame, err.Code)
	}
}

func TestStartSinglePlayerGameMovesLobbyRoomWithMemberToInGame(t *testing.T) {
	room := NewRoom("room", RoomStateLobby, nil)
	room.AddMember(NewRoomMember("session-owner"))

	if err := room.StartSinglePlayerGame(func() *game.Game { return game.New() }); err != nil {
		t.Fatalf("expected single-player start to succeed, got %v", err)
	}
	if room.State != RoomStateInGame {
		t.Fatalf("expected room state %q, got %q", RoomStateInGame, room.State)
	}
	if room.GameInstance() == nil {
		t.Fatal("expected game to be created")
	}
	room.GameInstance().Stop()
}

func TestStartSinglePlayerGameRejectsRoomWithoutMembers(t *testing.T) {
	room := NewRoom("room", RoomStateLobby, nil)

	err := room.StartSinglePlayerGame(func() *game.Game { return game.New() })
	if err == nil {
		t.Fatal("expected single-player start without members to fail")
	}
	if err.Code != RoomErrorNotInRoom {
		t.Fatalf("expected error code %q, got %q", RoomErrorNotInRoom, err.Code)
	}
}

func TestStartSinglePlayerGameRejectsNonLobbyRoom(t *testing.T) {
	room := NewRoom("room", RoomStateStarting, nil)
	room.AddMember(NewRoomMember("session-owner"))

	err := room.StartSinglePlayerGame(func() *game.Game { return game.New() })
	if err == nil {
		t.Fatal("expected single-player start from non-lobby room to fail")
	}
	if err.Code != RoomErrorRoomInGame {
		t.Fatalf("expected error code %q, got %q", RoomErrorRoomInGame, err.Code)
	}
}

func TestResetToLobbyOnlyWorksFromGameOver(t *testing.T) {
	room := NewRoom("room", RoomStateLobby, nil)
	room.AddMember(NewRoomMember("session-owner"))

	err := room.ResetToLobby("Player-1")
	if err == nil {
		t.Fatal("expected reset from non-game-over state to fail")
	}
	if err.Code != RoomErrorInvalidRoomState {
		t.Fatalf("expected error code %q, got %q", RoomErrorInvalidRoomState, err.Code)
	}
}

func TestResetToLobbyClearsReadyState(t *testing.T) {
	room := NewRoom("room", RoomStateGameOver, nil)
	member := room.AddMember(NewRoomMember("session-owner"))
	member.SetReady(true)

	if err := room.ResetToLobby("Player-1"); err != nil {
		t.Fatalf("expected reset to lobby to succeed, got %v", err)
	}
	if room.State != RoomStateLobby {
		t.Fatalf("expected room state %q, got %q", RoomStateLobby, room.State)
	}
	if member.Ready {
		t.Fatal("expected ready state to be cleared")
	}
}
