package game

import (
	"testing"

	"github.com/Lokee86/space-rocks/services/game-server/internal/game/awards"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/damage"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/lives"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/scoring"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/teams"
)

func TestDestructionAwardsAreIdempotentAndApplyCombo(t *testing.T) {
	game := NewWithSeed(1)
	playerID := game.AddPlayer()

	game.mu.Lock()
	game.applyDestructionAwardsLocked("asteroid-1", []scoring.Award{{PlayerID: playerID, Points: 120, Reason: scoring.EventAsteroidDestroyed}})
	game.applyDestructionAwardsLocked("asteroid-1", []scoring.Award{{PlayerID: playerID, Points: 120, Reason: scoring.EventAsteroidDestroyed}})
	game.applyDestructionAwardsLocked("asteroid-2", []scoring.Award{{PlayerID: playerID, Points: 120, Reason: scoring.EventAsteroidDestroyed}})
	game.mu.Unlock()

	owner := awards.Owner{Scope: awards.ScopePlayer, ID: playerID}
	if score, _ := game.GameplayCounter(owner, awards.CounterScore); score != 270 {
		t.Fatalf("score = %v, want 270", score)
	}
	if kills, _ := game.GameplayCounter(owner, awards.CounterKills); kills != 2 {
		t.Fatalf("kills = %v, want 2", kills)
	}
	snapshot := game.GameplayAwardSnapshot()
	if len(snapshot.Combos) != 1 || snapshot.Combos[0].State.Multiplier != 1.5 {
		t.Fatalf("unexpected combo snapshot: %+v", snapshot.Combos)
	}
}

func TestDamageContributionProducesCountersAndAssistDistribution(t *testing.T) {
	game := NewWithSeed(2)
	killerID := game.AddPlayer()
	assistID := game.AddPlayer()
	policy := game.AwardPolicy()
	policy.Assists.Enabled = true
	policy.AssistScore = 10
	game.SetAwardPolicy(policy)

	game.mu.Lock()
	game.recordDamageAwardConsequences(damage.DamageResolutionRequest{
		Source: damage.DamageSource{ResponsiblePlayerID: killerID},
		Target: damage.DamageTarget{EntityID: "asteroid-x", EntityType: damage.EntityTypeAsteroid},
	}, damage.DamageResult{Kind: damage.DamageResultKindDamage, AppliedToHealth: 90})
	game.recordDamageAwardConsequences(damage.DamageResolutionRequest{
		Source: damage.DamageSource{ResponsiblePlayerID: assistID},
		Target: damage.DamageTarget{EntityID: "asteroid-x", EntityType: damage.EntityTypeAsteroid},
	}, damage.DamageResult{Kind: damage.DamageResultKindDamage, AppliedToHealth: 10})
	game.applyDestructionAwardsLocked("asteroid-x", []scoring.Award{{PlayerID: killerID, Points: 120, Reason: scoring.EventAsteroidDestroyed}})
	game.mu.Unlock()

	killer := awards.Owner{Scope: awards.ScopePlayer, ID: killerID}
	assistant := awards.Owner{Scope: awards.ScopePlayer, ID: assistID}
	assertGameplayCounter(t, game, killer, awards.CounterDamageDealt, 90)
	assertGameplayCounter(t, game, assistant, awards.CounterDamageDealt, 10)
	assertGameplayCounter(t, game, killer, awards.CounterKills, 1)
	assertGameplayCounter(t, game, assistant, awards.CounterAssists, 1)
	assertGameplayCounter(t, game, assistant, awards.CounterScore, 10)
	if count := game.awardsRuntime().ContributionCount("asteroid-x"); count != 0 {
		t.Fatalf("contribution history count = %d, want 0", count)
	}
}

func TestDeathAwardsResetVictimProgressAndAdvanceKillerStreak(t *testing.T) {
	game := NewWithSeed(3)
	killerID := game.AddPlayer()
	victimID := game.AddPlayer()
	victim := awards.Owner{Scope: awards.ScopePlayer, ID: victimID}

	game.mu.Lock()
	_, _ = game.awardsRuntime().ApplyCombo(victim, 0)
	_, _ = game.awardsRuntime().IncrementStreak(victim, "survival")
	game.applyDeathAwardsLocked(lives.DeathFact{
		Accepted: true, PlayerID: victimID, DeathCount: 1, ReasonCode: "weapon",
		Input: lives.DeathInput{Attribution: lives.AttributionPlayerCaused, KillerPlayerID: killerID},
	})
	game.mu.Unlock()

	assertGameplayCounter(t, game, victim, awards.CounterDeaths, 1)
	assertGameplayCounter(t, game, awards.Owner{Scope: awards.ScopePlayer, ID: killerID}, awards.CounterKills, 1)
	snapshot := game.GameplayAwardSnapshot()
	if len(snapshot.Combos) != 0 {
		t.Fatalf("victim combo was not reset: %+v", snapshot.Combos)
	}
	if len(snapshot.Streaks) != 1 || snapshot.Streaks[0].Owner.ID != killerID || snapshot.Streaks[0].Count != 1 {
		t.Fatalf("unexpected streak snapshot: %+v", snapshot.Streaks)
	}
}

func TestRemovedPlayerCountersRemainInDerivedTeamTotal(t *testing.T) {
	game := NewWithSeed(4)
	player1 := game.AddPlayerWithTeam(teams.Team1)
	player2 := game.AddPlayerWithTeam(teams.Team1)
	_, err := game.ApplyGameplayAwardEvent("team-score", []awards.Mutation{
		{Owner: awards.Owner{Scope: awards.ScopePlayer, ID: player1}, CounterID: awards.CounterScore, Operation: awards.MutationIncrement, Value: 10},
		{Owner: awards.Owner{Scope: awards.ScopePlayer, ID: player2}, CounterID: awards.CounterScore, Operation: awards.MutationIncrement, Value: 20},
	})
	if err != nil {
		t.Fatalf("apply award event: %v", err)
	}
	game.RemovePlayer(player1)
	if _, err := game.ApplyGameplayAwardEvent("late-score", []awards.Mutation{{
		Owner: awards.Owner{Scope: awards.ScopePlayer, ID: player1}, CounterID: awards.CounterScore,
		Operation: awards.MutationIncrement, Value: 100,
	}}); err == nil {
		t.Fatal("removed player accepted a new award")
	}

	snapshot := game.GameplayAwardSnapshot()
	found := false
	for _, counter := range snapshot.TeamTotals {
		if counter.Owner.ID == string(teams.Team1) && counter.CounterID == awards.CounterScore {
			found = true
			if counter.Value != 30 {
				t.Fatalf("team score = %v, want 30", counter.Value)
			}
		}
	}
	if !found {
		t.Fatal("derived team score not found")
	}
}

func TestAwardObserverReceivesEffectiveDistributionOnce(t *testing.T) {
	game := NewWithSeed(5)
	playerID := game.AddPlayer()
	observed := 0
	game.AddGameplayAwardObserver(func(result awards.EventResult) {
		if result.EventID == "observer-event" {
			observed++
		}
	})
	mutation := []awards.Mutation{{
		Owner: awards.Owner{Scope: awards.ScopePlayer, ID: playerID}, CounterID: awards.CounterScore,
		Operation: awards.MutationIncrement, Value: 5,
	}}
	if _, err := game.ApplyGameplayAwardEvent("observer-event", mutation); err != nil {
		t.Fatalf("first event: %v", err)
	}
	if result, err := game.ApplyGameplayAwardEvent("observer-event", mutation); err != nil || !result.Duplicate {
		t.Fatalf("duplicate event result = %+v, err = %v", result, err)
	}
	if observed != 1 {
		t.Fatalf("observer calls = %d, want 1", observed)
	}
}

func TestGameplayAwardResetClearsMatchStateAndScoreProjection(t *testing.T) {
	game := NewWithSeed(6)
	playerID := game.AddPlayer()
	if _, err := game.ApplyGameplayAwardEvent("score-before-reset", []awards.Mutation{{
		Owner: awards.Owner{Scope: awards.ScopePlayer, ID: playerID}, CounterID: awards.CounterScore,
		Operation: awards.MutationIncrement, Value: 25,
	}}); err != nil {
		t.Fatalf("apply score: %v", err)
	}
	game.ResetGameplayAwards()
	assertGameplayCounter(t, game, awards.Owner{Scope: awards.ScopePlayer, ID: playerID}, awards.CounterScore, 0)
	game.mu.Lock()
	projectedScore := game.playerSessions[playerID].Score
	game.mu.Unlock()
	if projectedScore != 0 {
		t.Fatalf("score projection after reset = %d, want 0", projectedScore)
	}
}

func assertGameplayCounter(t *testing.T, game *Game, owner awards.Owner, id awards.CounterID, want float64) {
	t.Helper()
	got, _ := game.GameplayCounter(owner, id)
	if got != want {
		t.Fatalf("%s for %s = %v, want %v", id, owner.ID, got, want)
	}
}
