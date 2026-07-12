package rooms

import (
	"sync/atomic"
	"testing"
	"time"
)

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

func TestCleanupEmptyRoomReleasesManagerLockBeforeStoppingGame(t *testing.T) {
	manager := NewRoomManagerWithCleanupDelay(0)
	room := NewRoom("cleanup-room", RoomStateLobby, nil)

	manager.mu.Lock()
	manager.rooms[room.ID] = room
	manager.mu.Unlock()

	room.mu.Lock()
	cleanupVersion := room.cleanup.IncrementVersion()
	room.mu.Unlock()

	var stopCalls atomic.Int32
	roomCount := make(chan int, 1)
	cleanupComplete := make(chan struct{})
	previousStopRoomForCleanup := stopRoomForCleanup
	t.Cleanup(func() { stopRoomForCleanup = previousStopRoomForCleanup })
	stopRoomForCleanup = func(room *Room) {
		roomCount <- manager.RoomCount()
		stopCalls.Add(1)
		close(cleanupComplete)
	}

	go manager.cleanupEmptyRoom(room.ID, cleanupVersion)

	select {
	case <-cleanupComplete:
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for room cleanup")
	}

	if got := <-roomCount; got != 0 {
		t.Fatalf("expected manager room count to be zero before stopping game, got %d", got)
	}
	if got := stopCalls.Load(); got != 1 {
		t.Fatalf("expected stop seam to be invoked once, got %d", got)
	}
	if _, ok := manager.Find(room.ID); ok {
		t.Fatal("expected cleaned up room to be absent from manager")
	}
}

func TestRoomManagerSetReadyRejectsMissingSessionAfterPlayerIDReuse(t *testing.T) {
	manager := NewRoomManagerWithCleanupDelay(0)
	room, err := manager.CreateLobbyRoom()
	if err != nil {
		t.Fatalf("create room: %v", err)
	}
	t.Cleanup(manager.StopAll)
	if room.AddMember(NewRoomMember("session-old")) == nil {
		t.Fatal("add old member")
	}
	if _, _, removed := room.RemoveMemberForSession("session-old"); !removed {
		t.Fatal("remove old member")
	}
	replacement := room.AddMember(NewRoomMember("session-replacement"))
	if replacement == nil {
		t.Fatal("add replacement member")
	}
	joined, roomErr := manager.SetReady(room.ID, "session-old", true)
	if joined != nil || roomErr == nil || roomErr.Code != RoomErrorNotInRoom {
		t.Fatalf("expected nil room and not_in_room, got %v %v", joined, roomErr)
	}
	if replacement.Ready {
		t.Fatal("replacement should remain not ready")
	}
}

func TestRoomManagerStartRoomGameRejectsMissingSessionAfterPlayerIDReuse(t *testing.T) {
	manager := NewRoomManagerWithCleanupDelay(0)
	room, err := manager.CreateLobbyRoom()
	if err != nil {
		t.Fatalf("create room: %v", err)
	}
	t.Cleanup(manager.StopAll)
	if room.AddMember(NewRoomMember("session-old")) == nil {
		t.Fatal("add old member")
	}
	if _, _, removed := room.RemoveMemberForSession("session-old"); !removed {
		t.Fatal("remove old member")
	}
	replacement := room.AddMember(NewRoomMember("session-replacement"))
	if replacement == nil {
		t.Fatal("add replacement member")
	}
	replacement.SetReady(true)
	joined, roomErr := manager.StartRoomGame(room.ID, "session-old")
	if joined != nil || roomErr == nil || roomErr.Code != RoomErrorNotInRoom {
		t.Fatalf("expected nil room and not_in_room, got %v %v", joined, roomErr)
	}
	if room.State != RoomStateLobby || room.GameInstance() != nil || room.CurrentMatchID() != "" {
		t.Fatal("room should remain unchanged lobby")
	}
}

func TestRoomManagerReturnToLobbyRejectsMissingSessionAfterPlayerIDReuse(t *testing.T) {
	manager := NewRoomManagerWithCleanupDelay(0)
	room, err := manager.CreateLobbyRoom()
	if err != nil {
		t.Fatalf("create room: %v", err)
	}
	t.Cleanup(manager.StopAll)
	if room.AddMember(NewRoomMember("session-old")) == nil {
		t.Fatal("add old member")
	}
	if _, _, removed := room.RemoveMemberForSession("session-old"); !removed {
		t.Fatal("remove old member")
	}
	replacement := room.AddMember(NewRoomMember("session-replacement"))
	if replacement == nil {
		t.Fatal("add replacement member")
	}
	replacement.SetReady(true)
	startedRoom, roomErr := manager.StartRoomGame(room.ID, "session-replacement")
	if roomErr != nil || startedRoom == nil {
		t.Fatalf("start replacement game: %v", roomErr)
	}
	gameInstance := room.GameInstance()
	t.Cleanup(func() { gameInstance.Stop() })
	matchID := room.CurrentMatchID()
	if err := room.MarkGameOver(); err != nil {
		t.Fatalf("mark game over: %v", err)
	}
	returnedRoom, returnErr := manager.ReturnRoomToLobby(room.ID, "session-old")
	if returnedRoom != nil || returnErr == nil || returnErr.Code != RoomErrorNotInRoom {
		t.Fatalf("expected nil room and not_in_room, got %v %v", returnedRoom, returnErr)
	}
	if room.State != RoomStateGameOver || room.GameInstance() != gameInstance || room.CurrentMatchID() != matchID {
		t.Fatal("room game-over state should remain unchanged")
	}
}
