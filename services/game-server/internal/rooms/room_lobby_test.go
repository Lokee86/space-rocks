package rooms

import "testing"

func TestValidateStartAllowsOwnerWhenAllConnectedMembersAreReady(t *testing.T) {
	room := NewRoom("room", RoomStateLobby, nil)
	owner := room.AddMember(NewRoomMember("session-owner"))
	owner.SetReady(true)
	peer := room.AddMember(NewRoomMember("session-peer"))
	peer.SetReady(true)

	if err := room.ValidateStart("Player-1"); err != nil {
		t.Fatalf("expected owner start validation to succeed, got %v", err)
	}
}

func TestValidateStartRejectsNonOwner(t *testing.T) {
	room := NewRoom("room", RoomStateLobby, nil)
	room.AddMember(NewRoomMember("session-owner"))
	room.AddMember(NewRoomMember("session-peer"))

	err := room.ValidateStart("Player-2")
	if err == nil {
		t.Fatal("expected non-owner start validation to fail")
	}
	if err.Code != RoomErrorNotRoomOwner {
		t.Fatalf("expected error code %q, got %q", RoomErrorNotRoomOwner, err.Code)
	}
}

func TestValidateStartBlocksConnectedUnreadyMember(t *testing.T) {
	room := NewRoom("room", RoomStateLobby, nil)
	owner := room.AddMember(NewRoomMember("session-owner"))
	owner.SetReady(true)
	peer := room.AddMember(NewRoomMember("session-peer"))
	peer.Connected = true
	peer.SetReady(false)

	err := room.ValidateStart("Player-1")
	if err == nil {
		t.Fatal("expected connected unready member to block start")
	}
	if err.Code != RoomErrorNotReady {
		t.Fatalf("expected error code %q, got %q", RoomErrorNotReady, err.Code)
	}
}

func TestValidateStartIgnoresDisconnectedUnreadyMember(t *testing.T) {
	room := NewRoom("room", RoomStateLobby, nil)
	owner := room.AddMember(NewRoomMember("session-owner"))
	owner.SetReady(true)
	peer := room.AddMember(NewRoomMember("session-peer"))
	peer.Connected = false
	peer.SetReady(false)

	if err := room.ValidateStart("Player-1"); err != nil {
		t.Fatalf("expected disconnected unready member not to block start, got %v", err)
	}
}

func TestValidateStartRejectsNonMember(t *testing.T) {
	room := NewRoom("room", RoomStateLobby, nil)
	room.AddMember(NewRoomMember("session-owner"))

	err := room.ValidateStart("Player-2")
	if err == nil {
		t.Fatal("expected non-member start validation to fail")
	}
	if err.Code != RoomErrorNotInRoom {
		t.Fatalf("expected error code %q, got %q", RoomErrorNotInRoom, err.Code)
	}
}
