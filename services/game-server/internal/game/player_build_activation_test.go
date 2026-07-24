package game

import (
	"testing"

	"github.com/Lokee86/space-rocks/services/game-server/internal/game/playerbuild"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/teams"
)

func TestAddPlayerWithTeamAndBuildAppliesBuildBeforeSpawn(t *testing.T) {
	game := New()
	build := playerbuild.DefaultResolvedBuild("template-player")
	build.ModeID = "score_attack"

	playerID := game.AddPlayerWithTeamAndBuild(teams.NoTeam, build)
	if playerID == "" {
		t.Fatal("expected player to be added")
	}
	resolved, ok := game.PlayerResolvedBuild(playerID)
	if !ok {
		t.Fatal("expected resolved build")
	}
	if resolved.PlayerID != playerID {
		t.Fatalf("expected resolved player %q, got %q", playerID, resolved.PlayerID)
	}
	if resolved.ModeID != "score_attack" {
		t.Fatalf("expected selected mode to survive activation, got %q", resolved.ModeID)
	}
	ship := game.entities.Players[playerID]
	if ship == nil || ship.ShipTypeID != resolved.ShipID {
		t.Fatalf("ship was not spawned from resolved build: ship=%#v build=%#v", ship, resolved)
	}
}
