package game

import (
	"testing"

	"github.com/Lokee86/space-rocks/services/game-server/internal/game/modes"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/rules"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/teams"
)

func TestScoreAttackLocksOnFirstTargetScore(t *testing.T) {
	resolved, err := modes.Resolve(
		modes.RoomModeConfig{PresetID: modes.PresetScoreAttack, StartingLives: 4, TargetScore: 100},
		teams.Config{Structure: teams.StructureFFA},
	)
	if err != nil {
		t.Fatal(err)
	}
	game := New()
	if err := game.ConfigureMatchRules(resolved); err != nil {
		t.Fatal(err)
	}
	playerOne := game.AddPlayer()
	playerTwo := game.AddPlayer()
	game.matchElapsed = 2.5

	if change := game.SetPlayerScore(playerOne, 99); change.After != 99 || game.MatchDecision().IsOver {
		t.Fatalf("pre-target change = %+v decision = %+v", change, game.MatchDecision())
	}
	if change := game.SetPlayerScore(playerTwo, 100); change.After != 100 {
		t.Fatalf("target change = %+v", change)
	}
	decision := game.MatchDecision()
	if !decision.IsOver || decision.TerminalStatus != rules.TerminalCompleted || decision.EndReason != string(modes.EndTargetScoreReached) {
		t.Fatalf("decision = %+v", decision)
	}
	if len(decision.WinningPlayerIDs) != 1 || decision.WinningPlayerIDs[0] != playerTwo {
		t.Fatalf("winners = %+v", decision.WinningPlayerIDs)
	}
	winner := decisionPlayer(t, decision, playerTwo)
	if winner.Outcome != rules.OutcomeWon || winner.Placement != 1 || winner.CompletionTime != 2.5 || winner.TargetValue != 100 {
		t.Fatalf("winner = %+v", winner)
	}
	if change := game.AddPlayerScore(playerOne, 50); change.After != 99 {
		t.Fatalf("post-lock mutation = %+v", change)
	}
}

func TestResolvedInfiniteLivesAreAppliedToPlayers(t *testing.T) {
	resolved, err := modes.Resolve(
		modes.RoomModeConfig{PresetID: modes.PresetArcadeSurvival, InfiniteLives: true},
		teams.Config{Structure: teams.StructureCoOp},
	)
	if err != nil {
		t.Fatal(err)
	}
	game := New()
	if err := game.ConfigureMatchRules(resolved); err != nil {
		t.Fatal(err)
	}
	playerID := game.AddPlayerWithTeam(teams.ID("team-1"))
	session := game.playerSessions[playerID]
	if session == nil || !session.LifeOptions.InfiniteLives || session.Lives <= 0 {
		t.Fatalf("session = %+v", session)
	}
	if game.ResolvedMatchRules().TeamConfig.Structure != teams.StructureCoOp {
		t.Fatalf("rules = %+v", game.ResolvedMatchRules())
	}
}

func decisionPlayer(t *testing.T, decision rules.MatchDecision, playerID string) rules.PlayerDecision {
	t.Helper()
	for _, player := range decision.Players {
		if player.ID == playerID {
			return player
		}
	}
	t.Fatalf("missing decision player %q", playerID)
	return rules.PlayerDecision{}
}
