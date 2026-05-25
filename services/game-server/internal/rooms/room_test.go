package rooms

import "testing"

func TestRoomFirstAddedMemberBecomesOwner(t *testing.T) {
	room := NewRoom("room-1")

	room.AddMember("session-b")

	if room.OwnerID != "session-b" {
		t.Fatalf("expected first member to become owner, got %q", room.OwnerID)
	}
}

func TestRoomRemovingNonOwnerKeepsOwner(t *testing.T) {
	room := NewRoom("room-1")
	room.AddMember("session-a")
	room.AddMember("session-b")

	room.RemoveMember("session-b")

	if room.OwnerID != "session-a" {
		t.Fatalf("expected owner to remain session-a, got %q", room.OwnerID)
	}
}

func TestRoomRemovingOwnerReassignsSmallestRemainingMemberID(t *testing.T) {
	room := NewRoom("room-1")
	room.AddMember("session-c")
	room.AddMember("session-b")
	room.AddMember("session-a")

	room.RemoveMember("session-c")

	if room.OwnerID != "session-a" {
		t.Fatalf("expected owner to be reassigned to session-a, got %q", room.OwnerID)
	}
}

func TestRoomRemovingFinalMemberClearsOwner(t *testing.T) {
	room := NewRoom("room-1")
	room.AddMember("session-a")

	room.RemoveMember("session-a")

	if room.OwnerID != "" {
		t.Fatalf("expected owner to be cleared, got %q", room.OwnerID)
	}
}
