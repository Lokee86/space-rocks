package rooms

import "testing"

func TestAddMemberSetsFirstMemberAsOwner(t *testing.T) {
	room := NewRoom("room", RoomStateLobby, nil)

	room.AddMember(NewRoomMember("session-1"))

	if ownerID := room.OwnerID(); ownerID != "Player-1" {
		t.Fatalf("expected OwnerID Player-1, got %q", ownerID)
	}
}

func TestRemoveMemberReassignsOwner(t *testing.T) {
	room := NewRoom("room", RoomStateLobby, nil)
	room.AddMember(NewRoomMember("session-1"))
	room.AddMember(NewRoomMember("session-2"))

	room.RemoveMember("Player-1")

	if ownerID := room.OwnerID(); ownerID != "Player-2" {
		t.Fatalf("expected OwnerID Player-2, got %q", ownerID)
	}
}

func TestRemoveNonOwnerPreservesOwner(t *testing.T) {
	room := NewRoom("room", RoomStateLobby, nil)
	room.AddMember(NewRoomMember("session-1"))
	room.AddMember(NewRoomMember("session-2"))

	room.RemoveMember("Player-2")

	if ownerID := room.OwnerID(); ownerID != "Player-1" {
		t.Fatalf("expected OwnerID Player-1, got %q", ownerID)
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

func TestSetMemberAccountIDForSession(t *testing.T) {
	room := NewRoom("room", RoomStateLobby, nil)
	member := room.AddMember(NewRoomMember("session-1"))

	if !room.SetMemberAccountIDForSession("session-1", "account-1") {
		t.Fatal("expected SetMemberAccountIDForSession to succeed")
	}
	if member.AccountID != "account-1" {
		t.Fatalf("expected AccountID account-1, got %q", member.AccountID)
	}
}

func TestSetMemberLocalProfileIDForSession(t *testing.T) {
	room := NewRoom("room", RoomStateLobby, nil)
	member := room.AddMember(NewRoomMember("session-1"))

	if !room.SetMemberLocalProfileIDForSession("session-1", "local-profile-1") {
		t.Fatal("expected SetMemberLocalProfileIDForSession to succeed")
	}
	if member.LocalProfileID != "local-profile-1" {
		t.Fatalf("expected LocalProfileID local-profile-1, got %q", member.LocalProfileID)
	}
}

func TestSetMemberIdentityForMissingSessionReturnsFalse(t *testing.T) {
	room := NewRoom("room", RoomStateLobby, nil)
	room.AddMember(NewRoomMember("session-1"))

	if room.SetMemberAccountIDForSession("missing-session", "account-1") {
		t.Fatal("expected SetMemberAccountIDForSession to fail for missing session")
	}
	if room.SetMemberLocalProfileIDForSession("missing-session", "local-profile-1") {
		t.Fatal("expected SetMemberLocalProfileIDForSession to fail for missing session")
	}
}

func TestMembersSnapshotReturnsValueCopies(t *testing.T) {
	room := NewRoom("room", RoomStateLobby, nil)
	member := room.AddMember(NewRoomMember("session-1"))
	member.SetReady(true)

	snapshot := room.MembersSnapshot()
	if len(snapshot) != 1 {
		t.Fatalf("expected 1 snapshot member, got %d", len(snapshot))
	}
	if snapshot[0].PlayerID != "Player-1" {
		t.Fatalf("expected snapshot PlayerID Player-1, got %q", snapshot[0].PlayerID)
	}
	if snapshot[0].SessionID != "session-1" {
		t.Fatalf("expected snapshot SessionID session-1, got %q", snapshot[0].SessionID)
	}
	if !snapshot[0].Ready {
		t.Fatal("expected snapshot member to be ready")
	}

	snapshot[0].SetReady(false)

	if !member.Ready {
		t.Fatal("expected snapshot mutation not to affect room member")
	}
}

func TestMemberCountAndIsFullUseMembership(t *testing.T) {
	room := NewRoom("room", RoomStateLobby, nil)

	if got := room.MemberCount(); got != 0 {
		t.Fatalf("expected empty member count 0, got %d", got)
	}
	if room.IsFull() {
		t.Fatal("expected empty room not to be full")
	}

	for index := 0; index < MaxPlayersPerRoom; index++ {
		room.AddMember(NewRoomMember("session-" + formatPlayerID(index+1)))
	}

	if got := room.MemberCount(); got != MaxPlayersPerRoom {
		t.Fatalf("expected member count %d, got %d", MaxPlayersPerRoom, got)
	}
	if !room.IsFull() {
		t.Fatal("expected room to be full")
	}
}
