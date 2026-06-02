package rooms

import "testing"

func TestRoomManagerJoinRoomRejectsInvalidRoomCode(t *testing.T) {
	manager := NewRoomManagerWithCleanupDelay(0)

	room, err := manager.JoinRoom("session-1", "bad")
	if err == nil {
		t.Fatal("expected invalid room code to fail")
	}
	if room != nil {
		t.Fatalf("expected room to be nil, got %#v", room)
	}
	if err.Code != RoomErrorInvalidRoomCode {
		t.Fatalf("expected error code %q, got %q", RoomErrorInvalidRoomCode, err.Code)
	}
}

func TestRoomManagerJoinRoomRejectsMissingRoom(t *testing.T) {
	manager := NewRoomManagerWithCleanupDelay(0)

	room, err := manager.JoinRoom("session-1", "ABCDEF")
	if err == nil {
		t.Fatal("expected missing room to fail")
	}
	if room != nil {
		t.Fatalf("expected room to be nil, got %#v", room)
	}
	if err.Code != RoomErrorRoomNotFound {
		t.Fatalf("expected error code %q, got %q", RoomErrorRoomNotFound, err.Code)
	}
}

func TestRoomManagerJoinRoomAcceptsLobbyRoom(t *testing.T) {
	manager := NewRoomManagerWithCleanupDelay(0)

	room, err := manager.CreateLobbyRoom()
	if err != nil {
		t.Fatalf("expected lobby room creation to succeed, got %v", err)
	}

	joinedRoom, roomErr := manager.JoinRoom("session-2", room.ID)
	if roomErr != nil {
		t.Fatalf("expected lobby join to succeed, got %v", roomErr)
	}
	if joinedRoom == nil {
		t.Fatal("expected joined room to be non-nil")
	}
	if got := joinedRoom.MemberCount(); got != 1 {
		t.Fatalf("expected member count 1, got %d", got)
	}
	playerID, ok := joinedRoom.PlayerIDForSession("session-2")
	if !ok {
		t.Fatal("expected joined session to resolve to a player ID")
	}
	if playerID == "" {
		t.Fatal("expected resolved player ID to be non-empty")
	}
}

func TestRoomManagerJoinRoomRejectsNonJoinableRoom(t *testing.T) {
	manager := NewRoomManagerWithCleanupDelay(0)

	room, err := manager.CreateSinglePlayerRoom("session-1")
	if err != nil {
		t.Fatalf("expected single-player room creation to succeed, got %v", err)
	}

	memberCount := room.MemberCount()
	joinedRoom, roomErr := manager.JoinRoom("session-2", room.ID)
	if roomErr == nil {
		t.Fatal("expected non-joinable room join to fail")
	}
	if joinedRoom != nil {
		t.Fatalf("expected room to be nil, got %#v", joinedRoom)
	}
	if roomErr.Code != RoomErrorInvalidRoomState {
		t.Fatalf("expected error code %q, got %q", RoomErrorInvalidRoomState, roomErr.Code)
	}
	if got := room.MemberCount(); got != memberCount {
		t.Fatalf("expected member count to remain %d, got %d", memberCount, got)
	}
}

func TestRoomManagerJoinRoomRejectsFullRoom(t *testing.T) {
	manager := NewRoomManagerWithCleanupDelay(0)

	room, err := manager.CreateLobbyRoom()
	if err != nil {
		t.Fatalf("expected lobby room creation to succeed, got %v", err)
	}

	for index := 0; index < MaxPlayersPerRoom; index++ {
		sessionID := "session-" + string(rune('a'+index))
		joinedRoom, roomErr := manager.JoinRoom(sessionID, room.ID)
		if roomErr != nil {
			t.Fatalf("expected join %d to succeed, got %v", index+1, roomErr)
		}
		if joinedRoom == nil {
			t.Fatalf("expected join %d to return a room", index+1)
		}
	}

	joinedRoom, roomErr := manager.JoinRoom("session-overflow", room.ID)
	if roomErr == nil {
		t.Fatal("expected full room join to fail")
	}
	if joinedRoom != nil {
		t.Fatalf("expected room to be nil, got %#v", joinedRoom)
	}
	if roomErr.Code != RoomErrorRoomFull {
		t.Fatalf("expected error code %q, got %q", RoomErrorRoomFull, roomErr.Code)
	}
	if got := room.MemberCount(); got != MaxPlayersPerRoom {
		t.Fatalf("expected member count %d, got %d", MaxPlayersPerRoom, got)
	}
}

func TestRoomManagerJoinRoomRejectsStartingRoom(t *testing.T) {
	manager := NewRoomManagerWithCleanupDelay(0)

	room, err := manager.CreateLobbyRoom()
	if err != nil {
		t.Fatalf("expected lobby room creation to succeed, got %v", err)
	}
	room.State = RoomStateStarting

	memberCount := room.MemberCount()
	joinedRoom, roomErr := manager.JoinRoom("session-2", room.ID)
	if roomErr == nil {
		t.Fatal("expected starting room join to fail")
	}
	if joinedRoom != nil {
		t.Fatalf("expected room to be nil, got %#v", joinedRoom)
	}
	if roomErr.Code != RoomErrorRoomInGame {
		t.Fatalf("expected error code %q, got %q", RoomErrorRoomInGame, roomErr.Code)
	}
	if got := room.MemberCount(); got != memberCount {
		t.Fatalf("expected member count to remain %d, got %d", memberCount, got)
	}
}

func TestRoomManagerJoinRoomRejectsInGameRoom(t *testing.T) {
	manager := NewRoomManagerWithCleanupDelay(0)

	room, err := manager.CreateLobbyRoom()
	if err != nil {
		t.Fatalf("expected lobby room creation to succeed, got %v", err)
	}
	room.State = RoomStateInGame

	memberCount := room.MemberCount()
	joinedRoom, roomErr := manager.JoinRoom("session-2", room.ID)
	if roomErr == nil {
		t.Fatal("expected in-game room join to fail")
	}
	if joinedRoom != nil {
		t.Fatalf("expected room to be nil, got %#v", joinedRoom)
	}
	if roomErr.Code != RoomErrorRoomInGame {
		t.Fatalf("expected error code %q, got %q", RoomErrorRoomInGame, roomErr.Code)
	}
	if got := room.MemberCount(); got != memberCount {
		t.Fatalf("expected member count to remain %d, got %d", memberCount, got)
	}
}

func TestRoomManagerJoinRoomRejectsClosedRoom(t *testing.T) {
	manager := NewRoomManagerWithCleanupDelay(0)

	room, err := manager.CreateLobbyRoom()
	if err != nil {
		t.Fatalf("expected lobby room creation to succeed, got %v", err)
	}
	room.State = RoomStateClosed

	memberCount := room.MemberCount()
	joinedRoom, roomErr := manager.JoinRoom("session-2", room.ID)
	if roomErr == nil {
		t.Fatal("expected closed room join to fail")
	}
	if joinedRoom != nil {
		t.Fatalf("expected room to be nil, got %#v", joinedRoom)
	}
	if roomErr.Code != RoomErrorRoomClosed {
		t.Fatalf("expected error code %q, got %q", RoomErrorRoomClosed, roomErr.Code)
	}
	if got := room.MemberCount(); got != memberCount {
		t.Fatalf("expected member count to remain %d, got %d", memberCount, got)
	}
}
