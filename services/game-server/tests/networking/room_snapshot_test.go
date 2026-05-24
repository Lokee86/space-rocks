package networkingtests

import (
	"testing"

	"github.com/Lokee86/space-rocks/server/internal/game"
	"github.com/Lokee86/space-rocks/server/internal/networking"
	"github.com/Lokee86/space-rocks/server/internal/rooms"
)

func TestBuildRoomSnapshotIncludesRoomStateAndCapacity(t *testing.T) {
	room := rooms.NewRoom("TEST", rooms.RoomStateLobby, nil)

	snapshot := networking.BuildRoomSnapshot(room, "player-1")

	if snapshot.Type != game.PacketTypeRoomSnapshot {
		t.Fatalf("expected snapshot packet type %q, got %q", game.PacketTypeRoomSnapshot, snapshot.Type)
	}
	if snapshot.RoomCode != "TEST" {
		t.Fatalf("expected room code %q, got %q", "TEST", snapshot.RoomCode)
	}
	if snapshot.RoomState != string(rooms.RoomStateLobby) {
		t.Fatalf("expected room state %q, got %q", rooms.RoomStateLobby, snapshot.RoomState)
	}
	if snapshot.LocalMemberID != "player-1" {
		t.Fatalf("expected local member id %q, got %q", "player-1", snapshot.LocalMemberID)
	}
	if snapshot.MaxPlayers != rooms.MaxPlayersPerRoom {
		t.Fatalf("expected max players %d, got %d", rooms.MaxPlayersPerRoom, snapshot.MaxPlayers)
	}
}

func TestBuildRoomSnapshotIncludesMembersAndReadyStates(t *testing.T) {
	room := rooms.NewRoom("TEST", rooms.RoomStateLobby, nil)
	room.AddMemberID("player-2")
	room.SetMemberReady("player-2", false)
	room.AddMemberID("player-1")
	room.SetMemberReady("player-1", true)

	snapshot := networking.BuildRoomSnapshot(room, "player-1")

	if len(snapshot.Members) != 2 {
		t.Fatalf("expected 2 snapshot members, got %d", len(snapshot.Members))
	}
	if snapshot.Members[0].MemberID != "player-1" || !snapshot.Members[0].Ready {
		t.Fatalf("expected first sorted member to be ready player-1, got %#v", snapshot.Members[0])
	}
	if snapshot.Members[1].MemberID != "player-2" || snapshot.Members[1].Ready {
		t.Fatalf("expected second sorted member to be not-ready player-2, got %#v", snapshot.Members[1])
	}
	for _, member := range snapshot.Members {
		if !member.Connected {
			t.Fatalf("expected snapshot member to be connected, got %#v", member)
		}
	}
}
