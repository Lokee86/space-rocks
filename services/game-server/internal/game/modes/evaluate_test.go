package modes

import (
	"testing"

	playerstate "github.com/Lokee86/space-rocks/services/game-server/internal/game/player"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/rules"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/teams"
)

func TestArcadeSurvivalLocksCompleteNonCompetitiveOutcomes(t *testing.T) {
	resolved, _ := Resolve(RoomModeConfig{PresetID: PresetArcadeSurvival, StartingLives: 3}, teams.Config{Structure: teams.StructureFFA})
	decision := EvaluateMatch(resolved, MatchFacts{HadParticipants: true, Players: []PlayerFact{
		{ID: "player-b", Status: playerstate.StatusEliminated, Score: 20},
		{ID: "player-a", Status: playerstate.StatusEliminated, Score: 10},
	}})
	if !decision.IsOver || decision.TerminalStatus != rules.TerminalCompleted || decision.EndReason != string(EndNoActivePlayers) {
		t.Fatalf("decision = %+v", decision)
	}
	if len(decision.Players) != 2 || decision.Players[0].ID != "player-a" || decision.Players[1].ID != "player-b" {
		t.Fatalf("players = %+v", decision.Players)
	}
	for _, player := range decision.Players {
		if player.Outcome != rules.OutcomeCompleted || player.Placement != 0 {
			t.Fatalf("arcade survival invented competitive result: %+v", player)
		}
	}
}

func TestScoreAttackLocksFirstSuccessfulParticipant(t *testing.T) {
	resolved, _ := Resolve(RoomModeConfig{PresetID: PresetScoreAttack, StartingLives: 3, TargetScore: 100}, teams.Config{Structure: teams.StructureFFA})
	decision := EvaluateMatch(resolved, MatchFacts{HadParticipants: true, Players: []PlayerFact{
		{ID: "player-b", Active: true, Status: playerstate.StatusActive, Score: 100, CompletionTime: 4, SuccessOrder: 2},
		{ID: "player-a", Active: true, Status: playerstate.StatusActive, Score: 100, CompletionTime: 3, SuccessOrder: 1},
	}})
	if !decision.IsOver || decision.EndReason != string(EndTargetScoreReached) || len(decision.WinningPlayerIDs) != 1 || decision.WinningPlayerIDs[0] != "player-a" {
		t.Fatalf("decision = %+v", decision)
	}
	if decision.TerminalStatus != rules.TerminalCompleted {
		t.Fatalf("terminal = %q", decision.TerminalStatus)
	}
}

func TestScoreAttackFailsWhenNoActiveParticipantsReachTarget(t *testing.T) {
	resolved, _ := Resolve(RoomModeConfig{PresetID: PresetScoreAttack, StartingLives: 3, TargetScore: 100}, teams.Config{Structure: teams.StructureFFA})
	decision := EvaluateMatch(resolved, MatchFacts{HadParticipants: true, Players: []PlayerFact{{ID: "player-a", Status: playerstate.StatusEliminated, Score: 90}}})
	if !decision.IsOver || decision.TerminalStatus != rules.TerminalFailed || decision.Players[0].Outcome != rules.OutcomeFailed {
		t.Fatalf("decision = %+v", decision)
	}
}
