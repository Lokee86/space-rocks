package networking

import (
	"sort"
	"testing"

	"github.com/Lokee86/space-rocks/services/game-server/internal/game"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/teams"
	"github.com/Lokee86/space-rocks/services/game-server/internal/rooms"
)

func TestBuildRoomSnapshotMatchResultIsEmptyWithoutResolvedSummary(t *testing.T) {
	room := rooms.NewRoom("room", rooms.RoomStateLobby, nil)

	snapshot := BuildRoomSnapshot(room, "")

	if snapshot.MatchResult.MatchID != "" {
		t.Fatalf("expected empty MatchID, got %q", snapshot.MatchResult.MatchID)
	}
	if snapshot.MatchResult.Mode != "" {
		t.Fatalf("expected empty Mode, got %q", snapshot.MatchResult.Mode)
	}
	if len(snapshot.MatchResult.Players) != 0 {
		t.Fatalf("expected 0 match result players, got %d", len(snapshot.MatchResult.Players))
	}
}

func TestBuildRoomSnapshotProjectsTeamIDFromMemberIDAssignment(t *testing.T) {
	room, err := rooms.NewRoomWithConfig("room", rooms.RoomStateLobby, nil, rooms.RoomCreationConfig{
		TeamConfig: teams.Config{Structure: teams.StructureCustom, AssignmentMode: teams.AssignmentOwnerAssigned},
		MaxPlayers: rooms.MaxPlayersPerRoom,
	})
	if err != nil {
		t.Fatalf("create room: %v", err)
	}
	member := room.AddMember(rooms.NewRoomMember("session-owner"))
	if err := room.SetTeamAssignment(member.SessionID, member.PlayerID, teams.Team2); err != nil {
		t.Fatalf("set assignment: %v", err)
	}

	snapshot := BuildRoomSnapshot(room, member.SessionID)
	if len(snapshot.Members) != 1 || snapshot.Members[0].TeamID != string(teams.Team2) {
		t.Fatalf("expected projected team assignment, got %+v", snapshot.Members)
	}
}

func TestBuildRoomSnapshotProjectsStructureAwareTeamCount(t *testing.T) {
	tests := []struct {
		name   string
		config teams.Config
		want   int
	}{
		{name: "ffa", config: teams.Config{Structure: teams.StructureFFA}, want: 0},
		{name: "co-op", config: teams.Config{Structure: teams.StructureCoOp}, want: 1},
		{name: "custom", config: teams.Config{Structure: teams.StructureCustom, AssignmentMode: teams.AssignmentOwnerAssigned}, want: 8},
		{name: "auto-balanced", config: teams.Config{Structure: teams.StructureAutoBalanced, AutoTeamCount: 3}, want: 3},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			room, err := rooms.NewRoomWithConfig("room", rooms.RoomStateLobby, nil, rooms.RoomCreationConfig{
				TeamConfig: test.config,
				MaxPlayers: rooms.MaxPlayersPerRoom,
			})
			if err != nil {
				t.Fatalf("create room: %v", err)
			}
			if got := BuildRoomSnapshot(room, "").TeamCount; got != test.want {
				t.Fatalf("team count = %d, want %d", got, test.want)
			}
		})
	}
}

func TestBuildRoomSnapshotIncludesResolvedMatchResult(t *testing.T) {
	room := rooms.NewRoom("room", rooms.RoomStateLobby, nil)
	room.SetJoinable(false)
	room.AddMember(rooms.NewRoomMember("session-owner"))

	if err := room.StartSinglePlayerGame(func() *game.Game { return game.New() }); err != nil {
		t.Fatalf("expected single-player start to succeed, got %v", err)
	}

	gameInstance := room.GameInstance()
	if gameInstance == nil {
		t.Fatal("expected game instance to be created")
	}
	playerID := gameInstance.AddPlayer()
	if playerID != "player-1" {
		t.Fatalf("expected player-1, got %q", playerID)
	}
	gameInstance.SetPlayerScore("player-1", 450)

	if err := room.MarkGameOver(); err != nil {
		t.Fatalf("expected game over transition to succeed, got %v", err)
	}

	snapshot := BuildRoomSnapshot(room, "session-owner")
	if snapshot.MatchResult.MatchID == "" {
		t.Fatal("expected MatchID to be populated")
	}
	if snapshot.MatchResult.Mode != "single_player" {
		t.Fatalf("expected Mode %q, got %q", "single_player", snapshot.MatchResult.Mode)
	}
	if len(snapshot.MatchResult.Players) != 1 {
		t.Fatalf("expected 1 match result player, got %d", len(snapshot.MatchResult.Players))
	}

	player := snapshot.MatchResult.Players[0]
	if player.GamePlayerID != "player-1" {
		t.Fatalf("expected GamePlayerID %q, got %q", "player-1", player.GamePlayerID)
	}
	if player.Score != 450 {
		t.Fatalf("expected Score 450, got %d", player.Score)
	}
	if player.ShipDeaths != 0 {
		t.Fatalf("expected ShipDeaths 0, got %d", player.ShipDeaths)
	}
	if player.Won {
		t.Fatal("expected Won to be false")
	}

	gameInstance.Stop()
}

func TestBuildRoomSnapshotUsesCoherentProjectionWithoutMutatingRoomMembers(t *testing.T) {
	room := rooms.NewRoom("room-1", rooms.RoomStateLobby, nil)
	room.AddMember(rooms.NewRoomMember("session-z"))
	room.AddMember(rooms.NewRoomMember("session-a"))

	beforeMembers := room.MembersSnapshot()
	snapshot := BuildRoomSnapshot(room, "session-z")
	if snapshot.RoomCode != "room-1" || snapshot.RoomState != string(rooms.RoomStateLobby) {
		t.Fatalf("unexpected room projection: %+v", snapshot)
	}
	if snapshot.CurrentMatchID != "" {
		t.Fatalf("expected empty current match, got %q", snapshot.CurrentMatchID)
	}
	if snapshot.LocalPlayerID == "" || snapshot.OwnerID == "" {
		t.Fatalf("expected local player and owner IDs: %+v", snapshot)
	}
	if len(snapshot.Members) != 2 || snapshot.Members[0].PlayerID == snapshot.Members[1].PlayerID {
		t.Fatalf("unexpected snapshot members: %+v", snapshot.Members)
	}

	roomMembers := room.MembersSnapshot()
	sort.Slice(beforeMembers, func(left, right int) bool { return beforeMembers[left].SessionID < beforeMembers[right].SessionID })
	sort.Slice(roomMembers, func(left, right int) bool { return roomMembers[left].SessionID < roomMembers[right].SessionID })
	for index := range beforeMembers {
		if roomMembers[index] != beforeMembers[index] {
			t.Fatalf("BuildRoomSnapshot mutated room member order: before=%+v after=%+v", beforeMembers, roomMembers)
		}
	}
}
