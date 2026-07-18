package game

import (
	"testing"

	"github.com/Lokee86/space-rocks/services/game-server/internal/game/lives"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/teams"
)

func TestGameCreatesIsolatedLivesRuntimeAndRegistersTeamIdentity(t *testing.T) {
	first := NewWithSeed(1)
	second := NewWithSeed(2)
	if first.lifeRuntime == nil || second.lifeRuntime == nil {
		t.Fatal("expected each game to construct a lives runtime")
	}
	if first.lifeRuntime == second.lifeRuntime {
		t.Fatal("expected games to own isolated lives runtimes")
	}

	playerID := first.AddPlayerWithTeam(teams.Team3)
	if playerID == "" {
		t.Fatal("expected player registration to succeed")
	}
	state, ok := first.lifeRuntime.ParticipantSnapshot(playerID)
	if !ok || state.TeamID != teams.Team3 {
		t.Fatalf("unexpected registered participant: %+v, ok=%t", state, ok)
	}
	if _, ok := second.lifeRuntime.ParticipantSnapshot(playerID); ok {
		t.Fatal("participant leaked into another game's lives runtime")
	}
}

func TestGameConstructionAcceptsResolvedSharedTeamPolicy(t *testing.T) {
	policy := lives.NewBaselinePolicy()
	policy.Model = lives.LifeModelSharedTeamPool
	policy.StartingLives = 0
	policy.TeamPool = &lives.TeamPoolPolicy{PoolID: "team-pool-1", StartingLives: 2}

	game, err := NewWithSeedAndPolicy(1, policy)
	if err != nil {
		t.Fatalf("NewWithSeedAndPolicy() error = %v", err)
	}
	playerID := game.AddPlayerWithTeam(teams.Team1)
	if playerID == "" {
		t.Fatal("expected shared-team player registration to succeed")
	}
	state, ok := game.lifeRuntime.ParticipantSnapshot(playerID)
	if !ok || state.EffectiveLives != 2 {
		t.Fatalf("unexpected shared participant state: %+v, ok=%t", state, ok)
	}
	pool, ok := game.lifeRuntime.TeamPoolSnapshot(teams.Team1)
	if !ok || pool.RemainingLives != 2 {
		t.Fatalf("unexpected shared team pool: %+v, ok=%t", pool, ok)
	}
}

func TestGameRollsBackPlayerAddWhenLivesRegistrationFails(t *testing.T) {
	game := New()
	policy := lives.NewBaselinePolicy()
	policy.Model = lives.LifeModelSharedTeamPool
	policy.StartingLives = 0
	policy.TeamPool = &lives.TeamPoolPolicy{
		PoolID:        "team-pool-1",
		StartingLives: 2,
	}
	lifeRuntime, err := lives.NewRuntime(policy)
	if err != nil {
		t.Fatalf("NewRuntime() error = %v", err)
	}
	game.lifeRuntime = lifeRuntime

	if playerID := game.AddPlayerWithTeam(teams.NoTeam); playerID != "" {
		t.Fatalf("expected failed player add to return empty ID, got %q", playerID)
	}
	if game.nextID != 0 || len(game.playerSessions) != 0 || len(game.entities.Players) != 0 || len(game.participantRecords) != 0 {
		t.Fatalf("failed player add left partial state: nextID=%d sessions=%d players=%d records=%d", game.nextID, len(game.playerSessions), len(game.entities.Players), len(game.participantRecords))
	}
}

func TestGameRollbackPlayerAddDeletesRuntimeRegistration(t *testing.T) {
	game := New()
	playerID := game.AddPlayer()
	if _, ok := game.lifeRuntime.ParticipantSnapshot(playerID); !ok {
		t.Fatalf("expected player %q to be registered before rollback", playerID)
	}

	game.RollbackPlayerAdd(playerID)
	if _, ok := game.lifeRuntime.ParticipantSnapshot(playerID); ok {
		t.Fatalf("expected runtime registration %q to be deleted by rollback", playerID)
	}
	if _, ok := game.playerSessions[playerID]; ok {
		t.Fatalf("expected player session %q to be removed by rollback", playerID)
	}
}
