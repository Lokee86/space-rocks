package rooms

import (
	"testing"

	"github.com/Lokee86/space-rocks/server/internal/game"
)

func TestFormatPlayerID(t *testing.T) {
	tests := []struct {
		name   string
		number int
		want   string
	}{
		{
			name:   "one",
			number: 1,
			want:   "Player-1",
		},
		{
			name:   "eight",
			number: 8,
			want:   "Player-8",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := formatPlayerID(test.number); got != test.want {
				t.Fatalf("expected %q, got %q", test.want, got)
			}
		})
	}
}

func TestNextAvailablePlayerIDLocked(t *testing.T) {
	tests := []struct {
		name      string
		playerIDs []string
		want      string
	}{
		{
			name: "empty room",
			want: "Player-1",
		},
		{
			name:      "player one occupied",
			playerIDs: []string{"Player-1"},
			want:      "Player-2",
		},
		{
			name:      "fills gap",
			playerIDs: []string{"Player-1", "Player-3"},
			want:      "Player-2",
		},
		{
			name: "continues past eight",
			playerIDs: []string{
				"Player-1",
				"Player-2",
				"Player-3",
				"Player-4",
				"Player-5",
				"Player-6",
				"Player-7",
				"Player-8",
			},
			want: "Player-9",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			room := roomWithPlayerIDs(test.playerIDs...)

			if got := room.nextAvailablePlayerIDLocked(); got != test.want {
				t.Fatalf("expected %q, got %q", test.want, got)
			}
		})
	}
}

func TestAddMemberAssignsPlayerID(t *testing.T) {
	room := NewRoom("room", RoomStateLobby, nil)

	first := room.AddMember(NewRoomMember("session-1"))
	second := room.AddMember(NewRoomMember("session-2"))

	if first.PlayerID != "Player-1" {
		t.Fatalf("expected first member PlayerID Player-1, got %q", first.PlayerID)
	}
	if second.PlayerID != "Player-2" {
		t.Fatalf("expected second member PlayerID Player-2, got %q", second.PlayerID)
	}
}

func TestAddMemberStoresByPlayerIDAndSetsOwnerID(t *testing.T) {
	room := NewRoom("room", RoomStateLobby, nil)

	room.AddMember(NewRoomMember("session-1"))

	if _, ok := room.Members["Player-1"]; !ok {
		t.Fatal("expected room.Members to contain key Player-1")
	}
	if room.OwnerID != "Player-1" {
		t.Fatalf("expected OwnerID Player-1, got %q", room.OwnerID)
	}
}

func TestPlayerIDForSession(t *testing.T) {
	room := NewRoom("room", RoomStateLobby, nil)
	room.AddMember(NewRoomMember("session-1"))

	playerID, ok := room.PlayerIDForSession("session-1")
	if !ok {
		t.Fatal("expected session-1 to resolve")
	}
	if playerID != "Player-1" {
		t.Fatalf("expected Player-1, got %q", playerID)
	}

	playerID, ok = room.PlayerIDForSession("missing-session")
	if ok {
		t.Fatal("expected missing session not to resolve")
	}
	if playerID != "" {
		t.Fatalf("expected empty player ID, got %q", playerID)
	}
}

func TestRemoveMemberRecalculatesOwnerIDFromPlayerIDKeys(t *testing.T) {
	room := NewRoom("room", RoomStateLobby, nil)
	room.AddMember(NewRoomMember("session-1"))
	room.AddMember(NewRoomMember("session-2"))

	room.RemoveMember("Player-1")

	if room.OwnerID != "Player-2" {
		t.Fatalf("expected OwnerID Player-2, got %q", room.OwnerID)
	}
}

func TestSetReadyInLobbyLooksUpByPlayerID(t *testing.T) {
	room := NewRoom("room", RoomStateLobby, nil)
	member := room.AddMember(NewRoomMember("session-1"))

	if err := room.SetReadyInLobby("Player-1", true); err != nil {
		t.Fatalf("expected ready update to succeed, got %v", err)
	}
	if !member.Ready {
		t.Fatal("expected member to be ready")
	}
}

func TestValidateStartLooksUpOwnerByPlayerID(t *testing.T) {
	room := NewRoom("room", RoomStateLobby, nil)
	first := room.AddMember(NewRoomMember("session-1"))
	second := room.AddMember(NewRoomMember("session-2"))
	first.SetReady(true)
	second.SetReady(true)

	if err := room.ValidateStart("Player-1"); err != nil {
		t.Fatalf("expected owner start validation to succeed, got %v", err)
	}

	err := room.ValidateStart("Player-2")
	if err == nil {
		t.Fatal("expected non-owner start validation to fail")
	}
	if err.Code != RoomErrorNotRoomOwner {
		t.Fatalf("expected error code %q, got %q", RoomErrorNotRoomOwner, err.Code)
	}
}

func TestStartGameForMemberLooksUpOwnerByPlayerID(t *testing.T) {
	room := NewRoom("room", RoomStateLobby, nil)
	first := room.AddMember(NewRoomMember("session-1"))
	second := room.AddMember(NewRoomMember("session-2"))
	first.SetReady(true)
	second.SetReady(true)

	err := room.StartGameForMember("Player-2", game.New)
	if err == nil {
		t.Fatal("expected non-owner start to fail")
	}
	if err.Code != RoomErrorNotRoomOwner {
		t.Fatalf("expected error code %q, got %q", RoomErrorNotRoomOwner, err.Code)
	}

	if err := room.StartGameForMember("Player-1", game.New); err != nil {
		t.Fatalf("expected owner start to succeed, got %v", err)
	}
	if room.Game != nil {
		room.Game.Stop()
	}
}

func TestResetToLobbyLooksUpByPlayerID(t *testing.T) {
	room := NewRoom("room", RoomStateLobby, nil)
	room.AddMember(NewRoomMember("session-1"))
	room.State = RoomStateGameOver

	if err := room.ResetToLobby("Player-1"); err != nil {
		t.Fatalf("expected reset to lobby to succeed, got %v", err)
	}

	missingRoom := NewRoom("room", RoomStateGameOver, nil)
	missingRoom.AddMember(NewRoomMember("session-1"))

	err := missingRoom.ResetToLobby("missing-player")
	if err == nil {
		t.Fatal("expected missing player reset to fail")
	}
	if err.Code != RoomErrorNotInRoom {
		t.Fatalf("expected error code %q, got %q", RoomErrorNotInRoom, err.Code)
	}
}

func roomWithPlayerIDs(playerIDs ...string) *Room {
	room := NewRoom("room", RoomStateLobby, nil)
	for index, playerID := range playerIDs {
		sessionID := formatPlayerID(index + 100)
		room.Members[playerID] = &RoomMember{
			SessionID: sessionID,
			PlayerID:  playerID,
		}
	}
	return room
}
