package game

import (
	"testing"

	"github.com/Lokee86/space-rocks/services/game-server/internal/game/lives"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/teams"
)

func TestSharedLivesUseEffectivePoolValueAcrossGameProjections(t *testing.T) {
	policy := lives.NewBaselinePolicy()
	policy.Model = lives.LifeModelSharedTeamPool
	policy.StartingLives = 0
	policy.TeamPool = &lives.TeamPoolPolicy{PoolID: "team-pool-1", StartingLives: 2}
	game, err := NewWithSeedAndPolicy(1, policy)
	if err != nil {
		t.Fatal(err)
	}
	playerID := game.AddPlayerWithTeam(teams.Team1)
	if playerID == "" {
		t.Fatal("expected player registration")
	}
	if state := game.playerSessionStateLocked(game.playerSessions[playerID]); state.Lives != 2 {
		t.Fatalf("session projection lives = %d, want 2", state.Lives)
	}
	world, ok := game.playerWorldStateLocked(playerID)
	if !ok || world.Lives != 2 {
		t.Fatalf("world projection = %+v, ok=%t", world, ok)
	}
	if got := game.playerLives(playerID); got != 2 {
		t.Fatalf("playerLives = %d, want 2", got)
	}
	if snapshot := game.matchSnapshot(); len(snapshot.Players) != 1 || !snapshot.Players[0].HasRemainingLives {
		t.Fatalf("match projection = %+v", snapshot)
	}
}
