package networking

import (
	"testing"

	"github.com/Lokee86/space-rocks/server/internal/rooms"
)

func TestAddSessionMemberStoresRailsAccountID(t *testing.T) {
	room := rooms.NewRoom("room-1", rooms.RoomStateLobby, nil)
	session := &webSocketSession{
		identity: SessionIdentity{
			State:         SessionIdentityStateAuthenticatedAccount,
			AccountUserID: 1,
			AccountID:     "439e2746-9a06-45f1-b36b-b741b5bcfb12",
			DisplayName:   "Ada",
		},
	}

	addSessionMember(room, "session-1", session)

	playerID, ok := room.PlayerIDForSession("session-1")
	if !ok {
		t.Fatal("expected room member to be created")
	}

	found := false
	for _, member := range room.MembersSnapshot() {
		if member.PlayerID != playerID {
			continue
		}

		found = true
		if member.AccountID != "439e2746-9a06-45f1-b36b-b741b5bcfb12" {
			t.Fatalf("expected account id uuid, got %q", member.AccountID)
		}
		if member.AccountID == "1" {
			t.Fatal("expected account id to not use numeric user id")
		}
	}
	if !found {
		t.Fatalf("expected snapshot member for %q", playerID)
	}
}
