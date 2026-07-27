package networking

import (
	"testing"

	"github.com/Lokee86/space-rocks/services/game-server/internal/game"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/teams"
	"github.com/Lokee86/space-rocks/services/game-server/internal/rooms"
)

func TestBuildRoomSnapshotPreservesTeamIDInResolvedMatchResult(t *testing.T) {
	room := rooms.NewRoom("room", rooms.RoomStateLobby, nil)
	room.SetJoinable(false)
	room.AddMember(rooms.NewRoomMember("session-owner"))
	if err := room.StartSinglePlayerGame(game.New); err != nil {
		t.Fatalf("start game: %v", err)
	}
	gameInstance := room.GameInstance()
	gameInstance.SetTeamStructure(teams.StructureCustom)
	playerID := gameInstance.AddPlayerWithTeam(teams.Team3)
	gameInstance.SetPlayerScore(playerID, 250)
	if err := room.MarkGameOver(); err != nil {
		t.Fatalf("mark game over: %v", err)
	}

	snapshot := BuildRoomSnapshot(room, "session-owner")
	if len(snapshot.MatchResult.Players) != 1 {
		t.Fatalf("match result players = %+v, want one", snapshot.MatchResult.Players)
	}
	player := snapshot.MatchResult.Players[0]
	if player.TeamID != string(teams.Team3) {
		t.Fatalf("team ID = %q, want %q", player.TeamID, teams.Team3)
	}
	if player.Score != 250 {
		t.Fatalf("score = %d, want 250", player.Score)
	}
	gameInstance.Stop()
}
