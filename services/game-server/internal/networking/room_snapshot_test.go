package networking

import (
	"testing"

	"github.com/Lokee86/space-rocks/server/internal/game"
	"github.com/Lokee86/space-rocks/server/internal/rooms"
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
