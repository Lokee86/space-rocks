package networkingtests

import (
	"testing"

	"github.com/Lokee86/space-rocks/services/game-server/internal/game"
	"github.com/Lokee86/space-rocks/services/game-server/internal/networking"
	"github.com/Lokee86/space-rocks/services/game-server/internal/rooms"
)

func TestBuildRoomSnapshotIncludesRoomStateAndCapacity(t *testing.T) {
	room := rooms.NewRoom("TEST", rooms.RoomStateLobby, nil)
	member := room.AddMemberSessionID("session-1")

	snapshot := networking.BuildRoomSnapshot(room, "session-1")

	if snapshot.Type != game.PacketTypeRoomSnapshot {
		t.Fatalf("expected snapshot packet type %q, got %q", game.PacketTypeRoomSnapshot, snapshot.Type)
	}
	if snapshot.RoomCode != "TEST" {
		t.Fatalf("expected room code %q, got %q", "TEST", snapshot.RoomCode)
	}
	if snapshot.RoomState != string(rooms.RoomStateLobby) {
		t.Fatalf("expected room state %q, got %q", rooms.RoomStateLobby, snapshot.RoomState)
	}
	if snapshot.LocalPlayerID != member.PlayerID {
		t.Fatalf("expected local player id %q, got %q", member.PlayerID, snapshot.LocalPlayerID)
	}
	if snapshot.OwnerID != member.PlayerID {
		t.Fatalf("expected owner id %q, got %q", member.PlayerID, snapshot.OwnerID)
	}
	if snapshot.MaxPlayers != rooms.MaxPlayersPerRoom {
		t.Fatalf("expected max players %d, got %d", rooms.MaxPlayersPerRoom, snapshot.MaxPlayers)
	}
}

func TestBuildRoomSnapshotExposesConsecutiveMatchIDs(t *testing.T) {
	room := rooms.NewRoom("TEST", rooms.RoomStateLobby, nil)
	owner := room.AddMember(rooms.NewRoomMember("session-owner"))
	owner.SetReady(true)

	if err := room.StartGameForMember(owner.PlayerID, func() *game.Game { return game.New() }); err != nil {
		t.Fatalf("expected first start to succeed, got %v", err)
	}
	firstSnapshot := networking.BuildRoomSnapshot(room, owner.SessionID)
	firstMatchID := firstSnapshot.CurrentMatchID
	if firstMatchID == "" {
		t.Fatal("expected first snapshot to expose an active match ID")
	}
	if err := room.MarkGameOver(); err != nil {
		t.Fatalf("expected game over transition to succeed, got %v", err)
	}
	if err := room.ResetToLobby(owner.PlayerID); err != nil {
		t.Fatalf("expected reset to lobby to succeed, got %v", err)
	}
	owner.SetReady(true)
	if err := room.StartGameForMember(owner.PlayerID, func() *game.Game { return game.New() }); err != nil {
		t.Fatalf("expected second start to succeed, got %v", err)
	}
	secondSnapshot := networking.BuildRoomSnapshot(room, owner.SessionID)
	if secondSnapshot.CurrentMatchID == "" || secondSnapshot.CurrentMatchID == firstMatchID {
		t.Fatalf("expected second snapshot to expose a different active match ID, first=%q second=%q", firstMatchID, secondSnapshot.CurrentMatchID)
	}
	room.GameInstance().Stop()
}

func TestBuildRoomSnapshotIncludesMembersAndReadyStates(t *testing.T) {
	room := rooms.NewRoom("TEST", rooms.RoomStateLobby, nil)
	secondSessionMember := room.AddMemberSessionID("session-2")
	setReadyInLobbyBySession(t, room, "session-2", false)
	firstSessionMember := room.AddMemberSessionID("session-1")
	setReadyInLobbyBySession(t, room, "session-1", true)

	snapshot := networking.BuildRoomSnapshot(room, "session-1")

	if len(snapshot.Members) != 2 {
		t.Fatalf("expected 2 snapshot members, got %d", len(snapshot.Members))
	}
	if snapshot.LocalPlayerID != firstSessionMember.PlayerID {
		t.Fatalf("expected local player id %q, got %q", firstSessionMember.PlayerID, snapshot.LocalPlayerID)
	}
	if snapshot.OwnerID != secondSessionMember.PlayerID {
		t.Fatalf("expected owner id %q, got %q", secondSessionMember.PlayerID, snapshot.OwnerID)
	}
	if snapshot.Members[0].PlayerID != firstSessionMember.PlayerID || !snapshot.Members[0].Ready {
		t.Fatalf("expected first sorted member to be ready with player id %q, got %#v", firstSessionMember.PlayerID, snapshot.Members[0])
	}
	if snapshot.Members[1].PlayerID != secondSessionMember.PlayerID || snapshot.Members[1].Ready {
		t.Fatalf("expected second sorted member to be not-ready with player id %q, got %#v", secondSessionMember.PlayerID, snapshot.Members[1])
	}
	for _, member := range snapshot.Members {
		if !member.Connected {
			t.Fatalf("expected snapshot member to be connected, got %#v", member)
		}
	}
}

func setReadyInLobbyBySession(t *testing.T, room *rooms.Room, sessionID string, ready bool) {
	t.Helper()

	playerID, ok := room.PlayerIDForSession(sessionID)
	if !ok {
		t.Fatalf("expected session %q to resolve to a player ID", sessionID)
	}
	if err := room.SetReadyInLobby(playerID, ready); err != nil {
		t.Fatalf("set ready for session %q: %v", sessionID, err)
	}
}
