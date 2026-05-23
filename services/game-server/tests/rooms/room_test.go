package roomstests

import (
	"testing"

	"github.com/Lokee86/space-rocks/server/internal/game"
	"github.com/Lokee86/space-rocks/server/internal/rooms"
)

func TestRoomMemberAccessors(t *testing.T) {
	room := rooms.NewRoom("TEST", rooms.RoomStateLobby, nil)
	member := rooms.NewRoomMember("session-1")

	stored := room.AddMember(member)
	if stored != member {
		t.Fatal("expected AddMember to return stored member")
	}
	if !room.HasMember("session-1") {
		t.Fatal("expected room to have member")
	}
	if count := room.MemberCount(); count != 1 {
		t.Fatalf("expected member count 1, got %d", count)
	}

	room.RemoveMember("session-1")
	if room.HasMember("session-1") {
		t.Fatal("expected room member to be removed")
	}
}

func TestRoomSetMemberReady(t *testing.T) {
	room := rooms.NewRoom("TEST", rooms.RoomStateLobby, nil)
	room.AddMemberID("session-1")

	if !room.SetMemberReady("session-1", true) {
		t.Fatal("expected SetMemberReady to find member")
	}

	members := room.MembersSnapshot()
	if len(members) != 1 {
		t.Fatalf("expected 1 member snapshot, got %d", len(members))
	}
	if !members[0].Ready {
		t.Fatal("expected member to be ready")
	}

	if room.SetMemberReady("missing", true) {
		t.Fatal("expected SetMemberReady to fail for missing member")
	}
}

func TestRoomValidateStartAllowsOneReadyMember(t *testing.T) {
	room := rooms.NewRoom("TEST", rooms.RoomStateLobby, nil)
	room.AddMemberID("session-1")
	room.SetMemberReady("session-1", true)

	if roomErr := room.ValidateStart("session-1"); roomErr != nil {
		t.Fatalf("expected ready solo member to start, got %s", roomErr.Code)
	}
}

func TestRoomValidateStartRequiresRequesterInRoom(t *testing.T) {
	room := rooms.NewRoom("TEST", rooms.RoomStateLobby, nil)
	room.AddMemberID("session-1")
	room.SetMemberReady("session-1", true)

	roomErr := room.ValidateStart("missing")
	if roomErr == nil {
		t.Fatal("expected missing requester to be rejected")
	}
	if roomErr.Code != rooms.RoomErrorNotInRoom {
		t.Fatalf("expected error code %q, got %q", rooms.RoomErrorNotInRoom, roomErr.Code)
	}
}

func TestRoomValidateStartRequiresLobbyState(t *testing.T) {
	room := rooms.NewRoom("TEST", rooms.RoomStateGameOver, nil)
	room.AddMemberID("session-1")
	room.SetMemberReady("session-1", true)

	roomErr := room.ValidateStart("session-1")
	if roomErr == nil {
		t.Fatal("expected non-lobby room to be rejected")
	}
	if roomErr.Code != rooms.RoomErrorInvalidRoomState {
		t.Fatalf("expected error code %q, got %q", rooms.RoomErrorInvalidRoomState, roomErr.Code)
	}
}

func TestRoomValidateStartRejectsDoubleStartStates(t *testing.T) {
	for _, state := range []rooms.RoomState{rooms.RoomStateStarting, rooms.RoomStateInGame} {
		room := rooms.NewRoom("TEST", state, nil)
		room.AddMemberID("session-1")
		room.SetMemberReady("session-1", true)

		roomErr := room.ValidateStart("session-1")
		if roomErr == nil {
			t.Fatalf("expected state %q to reject double start", state)
		}
		if roomErr.Code != rooms.RoomErrorRoomInGame {
			t.Fatalf("expected error code %q for state %q, got %q", rooms.RoomErrorRoomInGame, state, roomErr.Code)
		}
	}
}

func TestRoomValidateStartRequiresAllConnectedMembersReady(t *testing.T) {
	room := rooms.NewRoom("TEST", rooms.RoomStateLobby, nil)
	room.AddMemberID("session-1")
	room.AddMemberID("session-2")
	room.SetMemberReady("session-1", true)

	roomErr := room.ValidateStart("session-1")
	if roomErr == nil {
		t.Fatal("expected unready connected member to block start")
	}
	if roomErr.Code != rooms.RoomErrorNotReady {
		t.Fatalf("expected error code %q, got %q", rooms.RoomErrorNotReady, roomErr.Code)
	}
}

func TestRoomValidateStartIgnoresDisconnectedUnreadyMembers(t *testing.T) {
	room := rooms.NewRoom("TEST", rooms.RoomStateLobby, nil)
	room.AddMemberID("session-1")
	disconnected := room.AddMemberID("session-2")
	disconnected.MarkDisconnected()
	room.SetMemberReady("session-1", true)

	if roomErr := room.ValidateStart("session-1"); roomErr != nil {
		t.Fatalf("expected disconnected unready member not to block start, got %s", roomErr.Code)
	}
}

func TestRoomMarkStartingMovesLobbyToStarting(t *testing.T) {
	room := rooms.NewRoom("TEST", rooms.RoomStateLobby, nil)

	if roomErr := room.MarkStarting(); roomErr != nil {
		t.Fatalf("expected lobby room to mark starting, got %s", roomErr.Code)
	}
	if room.State != rooms.RoomStateStarting {
		t.Fatalf("expected room state %q, got %q", rooms.RoomStateStarting, room.State)
	}
}

func TestRoomMarkStartingRejectsInvalidState(t *testing.T) {
	for _, state := range []rooms.RoomState{
		rooms.RoomStateStarting,
		rooms.RoomStateInGame,
		rooms.RoomStateGameOver,
		rooms.RoomStateClosed,
	} {
		room := rooms.NewRoom("TEST", state, nil)

		roomErr := room.MarkStarting()
		if roomErr == nil {
			t.Fatalf("expected state %q to reject mark starting", state)
		}
		if roomErr.Code != rooms.RoomErrorInvalidRoomState {
			t.Fatalf("expected error code %q for state %q, got %q", rooms.RoomErrorInvalidRoomState, state, roomErr.Code)
		}
		if room.State != state {
			t.Fatalf("expected rejected room to stay %q, got %q", state, room.State)
		}
	}
}

func TestRoomMembersSnapshotIsCopy(t *testing.T) {
	room := rooms.NewRoom("TEST", rooms.RoomStateLobby, nil)
	room.AddMemberID("session-1")

	members := room.MembersSnapshot()
	members[0].SetReady(true)

	freshMembers := room.MembersSnapshot()
	if freshMembers[0].Ready {
		t.Fatal("expected member snapshot mutation not to mutate room")
	}
}

func TestRoomIsGameOverReturnsFalseWithoutExactGameState(t *testing.T) {
	var missingRoom *rooms.Room
	if missingRoom.IsGameOver() {
		t.Fatal("expected nil room not to be game over")
	}

	lobbyRoom := rooms.NewRoom("TEST", rooms.RoomStateLobby, nil)
	if lobbyRoom.IsGameOver() {
		t.Fatal("expected lobby room not to be game over")
	}

	gameRoom := rooms.NewRoom("TEST", rooms.RoomStateInGame, game.New())
	if gameRoom.IsGameOver() {
		t.Fatal("expected game-over seam to default false until game state is exposed")
	}
}
