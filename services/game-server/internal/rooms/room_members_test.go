package rooms

import "testing"

func TestAddMemberSetsFirstMemberAsOwner(t *testing.T) {
	room := NewRoom("room", RoomStateLobby, nil)

	room.AddMember(NewRoomMember("session-1"))

	if room.OwnerID != "Player-1" {
		t.Fatalf("expected OwnerID Player-1, got %q", room.OwnerID)
	}
}

func TestRemoveMemberReassignsOwner(t *testing.T) {
	room := NewRoom("room", RoomStateLobby, nil)
	room.AddMember(NewRoomMember("session-1"))
	room.AddMember(NewRoomMember("session-2"))

	room.RemoveMember("Player-1")

	if room.OwnerID != "Player-2" {
		t.Fatalf("expected OwnerID Player-2, got %q", room.OwnerID)
	}
}

func TestPlayerIDForSessionFindsExpectedPlayer(t *testing.T) {
	room := NewRoom("room", RoomStateLobby, nil)
	room.AddMember(NewRoomMember("session-1"))

	playerID, ok := room.PlayerIDForSession("session-1")
	if !ok {
		t.Fatal("expected session-1 to resolve")
	}
	if playerID != "Player-1" {
		t.Fatalf("expected Player-1, got %q", playerID)
	}
}

func TestMembersSnapshotReturnsValueCopies(t *testing.T) {
	room := NewRoom("room", RoomStateLobby, nil)
	room.AddMember(NewRoomMember("session-1"))

	snapshot := room.MembersSnapshot()
	if len(snapshot) != 1 {
		t.Fatalf("expected 1 snapshot member, got %d", len(snapshot))
	}

	snapshot[0].SetReady(true)

	if room.Members["Player-1"].Ready {
		t.Fatal("expected snapshot mutation not to affect room member")
	}
}
