package roomstests

import (
	"testing"

	"github.com/Lokee86/space-rocks/server/internal/rooms"
)

func TestRoomManagerUsesDefaultRoomForBlankID(t *testing.T) {
	manager := rooms.NewRoomManager()
	defer manager.StopAll()

	defaultRoom := manager.DefaultRoom()
	blankRoom := manager.GetOrCreate("")
	spaceRoom := manager.GetOrCreate("   ")

	if blankRoom != defaultRoom {
		t.Fatal("expected blank room id to use default room")
	}
	if spaceRoom != defaultRoom {
		t.Fatal("expected whitespace room id to use default room")
	}
}

func TestRoomManagerCreatesAndFindsCompatibilityRooms(t *testing.T) {
	manager := rooms.NewRoomManager()
	defer manager.StopAll()

	first := manager.GetOrCreate("abc")
	again := manager.GetOrCreate("abc")
	second := manager.GetOrCreate("xyz")

	if first != again {
		t.Fatal("expected same room for same id")
	}
	if first == second {
		t.Fatal("expected different room for different id")
	}
	if first.State != rooms.RoomStateInGame {
		t.Fatalf("expected compatibility room state %q, got %q", rooms.RoomStateInGame, first.State)
	}
	if first.Game == nil {
		t.Fatal("expected compatibility room to create a game")
	}

	found, ok := manager.Find("abc")
	if !ok {
		t.Fatal("expected room to be found")
	}
	if found != first {
		t.Fatal("expected found room to match created room")
	}
}

func TestRoomManagerCreateLobbyRoom(t *testing.T) {
	manager := rooms.NewRoomManager()
	defer manager.StopAll()

	room, err := manager.CreateLobbyRoom()
	if err != nil {
		t.Fatalf("create lobby room: %v", err)
	}

	if room.State != rooms.RoomStateLobby {
		t.Fatalf("expected lobby room state %q, got %q", rooms.RoomStateLobby, room.State)
	}
	if room.Game != nil {
		t.Fatal("expected lobby room not to create a game")
	}
	if !rooms.IsValidRoomCode(room.ID) {
		t.Fatalf("expected generated room code, got %q", room.ID)
	}

	found, ok := manager.Find(room.ID)
	if !ok {
		t.Fatal("expected created lobby room to be found")
	}
	if found != room {
		t.Fatal("expected found lobby room to match created room")
	}
}

func TestRoomManagerJoinRoomAddsMember(t *testing.T) {
	manager := rooms.NewRoomManager()
	defer manager.StopAll()

	room, err := manager.CreateLobbyRoom()
	if err != nil {
		t.Fatalf("create lobby room: %v", err)
	}

	joinedRoom, roomErr := manager.JoinRoom("session-1", room.ID)
	if roomErr != nil {
		t.Fatalf("join room: %v", roomErr)
	}
	if joinedRoom != room {
		t.Fatal("expected joined room to match lobby room")
	}
	if !room.HasMember("session-1") {
		t.Fatal("expected joined member to be added")
	}
	if count := room.MemberCount(); count != 1 {
		t.Fatalf("expected member count 1, got %d", count)
	}
}

func TestRoomManagerJoinRoomBelowCapacitySucceeds(t *testing.T) {
	manager := rooms.NewRoomManager()
	defer manager.StopAll()

	room, err := manager.CreateLobbyRoom()
	if err != nil {
		t.Fatalf("create lobby room: %v", err)
	}
	for index := 0; index < rooms.MaxPlayersPerRoom-1; index++ {
		room.AddMemberID(string(rune('a' + index)))
	}

	joinedRoom, roomErr := manager.JoinRoom("last", room.ID)
	if roomErr != nil {
		t.Fatalf("join below capacity: %v", roomErr)
	}
	if joinedRoom != room {
		t.Fatal("expected joined room to match lobby room")
	}
	if count := room.MemberCount(); count != rooms.MaxPlayersPerRoom {
		t.Fatalf("expected member count %d, got %d", rooms.MaxPlayersPerRoom, count)
	}
}

func TestRoomManagerJoinRoomAtCapacityFails(t *testing.T) {
	manager := rooms.NewRoomManager()
	defer manager.StopAll()

	room, err := manager.CreateLobbyRoom()
	if err != nil {
		t.Fatalf("create lobby room: %v", err)
	}
	for index := 0; index < rooms.MaxPlayersPerRoom; index++ {
		room.AddMemberID(string(rune('a' + index)))
	}

	joinedRoom, roomErr := manager.JoinRoom("overflow", room.ID)
	if roomErr == nil {
		t.Fatal("expected room_full error")
	}
	if joinedRoom != nil {
		t.Fatal("expected no joined room")
	}
	if roomErr.Code != rooms.RoomErrorRoomFull {
		t.Fatalf("expected error code %q, got %q", rooms.RoomErrorRoomFull, roomErr.Code)
	}
	if count := room.MemberCount(); count != rooms.MaxPlayersPerRoom {
		t.Fatalf("expected failed join to preserve member count %d, got %d", rooms.MaxPlayersPerRoom, count)
	}
	if room.HasMember("overflow") {
		t.Fatal("expected failed join not to add member")
	}
}

func TestRoomManagerLeaveRoomRemovesMember(t *testing.T) {
	manager := rooms.NewRoomManager()
	defer manager.StopAll()

	room, err := manager.CreateLobbyRoom()
	if err != nil {
		t.Fatalf("create lobby room: %v", err)
	}
	if _, roomErr := manager.JoinRoom("session-1", room.ID); roomErr != nil {
		t.Fatalf("join room: %v", roomErr)
	}

	result, roomErr := manager.LeaveRoom(room.ID, "session-1")
	if roomErr != nil {
		t.Fatalf("leave room: %v", roomErr)
	}
	if result.Room != room {
		t.Fatal("expected leave result room to match lobby room")
	}
	if result.RoomID != room.ID {
		t.Fatalf("expected leave result room id %q, got %q", room.ID, result.RoomID)
	}
	if result.MemberID != "session-1" {
		t.Fatalf("expected leave result member id %q, got %q", "session-1", result.MemberID)
	}
	if result.RemainingMembers != 0 {
		t.Fatalf("expected remaining members 0, got %d", result.RemainingMembers)
	}
	if room.HasMember("session-1") {
		t.Fatal("expected member to be removed")
	}
}

func TestRoomManagerSetReadyUpdatesLobbyMember(t *testing.T) {
	manager := rooms.NewRoomManager()
	defer manager.StopAll()

	room, err := manager.CreateLobbyRoom()
	if err != nil {
		t.Fatalf("create lobby room: %v", err)
	}
	if _, roomErr := manager.JoinRoom("session-1", room.ID); roomErr != nil {
		t.Fatalf("join room: %v", roomErr)
	}

	updatedRoom, roomErr := manager.SetReady(room.ID, "session-1", true)
	if roomErr != nil {
		t.Fatalf("set ready: %v", roomErr)
	}
	if updatedRoom != room {
		t.Fatal("expected updated room to match lobby room")
	}

	members := room.MembersSnapshot()
	if len(members) != 1 {
		t.Fatalf("expected 1 member, got %d", len(members))
	}
	if !members[0].Ready {
		t.Fatal("expected member to be ready")
	}
}

func TestRoomManagerSetReadyRejectsMissingMember(t *testing.T) {
	manager := rooms.NewRoomManager()
	defer manager.StopAll()

	room, err := manager.CreateLobbyRoom()
	if err != nil {
		t.Fatalf("create lobby room: %v", err)
	}

	updatedRoom, roomErr := manager.SetReady(room.ID, "missing", true)
	if roomErr == nil {
		t.Fatal("expected not_in_room error")
	}
	if updatedRoom != nil {
		t.Fatal("expected no updated room")
	}
	if roomErr.Code != rooms.RoomErrorNotInRoom {
		t.Fatalf("expected error code %q, got %q", rooms.RoomErrorNotInRoom, roomErr.Code)
	}
}

func TestRoomManagerSetReadyRejectsNonLobbyRoom(t *testing.T) {
	manager := rooms.NewRoomManager()
	defer manager.StopAll()

	room := manager.GetOrCreate("ABCDEF")
	room.AddMemberID("session-1")

	updatedRoom, roomErr := manager.SetReady(room.ID, "session-1", true)
	if roomErr == nil {
		t.Fatal("expected invalid_room_state error")
	}
	if updatedRoom != nil {
		t.Fatal("expected no updated room")
	}
	if roomErr.Code != rooms.RoomErrorInvalidRoomState {
		t.Fatalf("expected error code %q, got %q", rooms.RoomErrorInvalidRoomState, roomErr.Code)
	}
}
