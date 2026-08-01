package modes

import (
	"testing"

	"github.com/Lokee86/space-rocks/services/game-server/internal/game/rules"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/teams"
)

func TestScoreAttackLocksFirstSuccessfulParticipant(t *testing.T) {
	resolved, _ := Resolve(RoomModeConfig{PresetID: PresetScoreAttack, StartingLives: 3, TargetScore: 100}, teams.Config{Structure: teams.StructureFFA})
	decision := EvaluateMatch(resolved, MatchFacts{HadParticipants: true, Players: []PlayerFact{
		{ID: "player-b", Active: true, Status: rules.PlayerActive, Score: 100, CompletionTime: 4, SuccessOrder: 2},
		{ID: "player-a", Active: true, Status: rules.PlayerActive, Score: 100, CompletionTime: 3, SuccessOrder: 1},
	}})
	if !decision.IsOver || decision.EndReason != string(EndTargetScoreReached) || len(decision.WinningPlayerIDs) != 1 || decision.WinningPlayerIDs[0] != "player-a" {
		t.Fatalf("decision = %+v", decision)
	}
	if decision.TerminalStatus != rules.TerminalCompleted {
		t.Fatalf("terminal = %q", decision.TerminalStatus)
	}
}

func TestDeathmatchLocksFirstPlayerToKillTarget(t *testing.T) {
	resolved, _ := Resolve(RoomModeConfig{PresetID: PresetDeathmatch, TargetKills: 10}, teams.Config{Structure: teams.StructureFFA})
	decision := EvaluateMatch(resolved, MatchFacts{HadParticipants: true, Players: []PlayerFact{
		{ID: "player-b", Active: true, Status: rules.PlayerActive, Score: 10, CompletionTime: 4, SuccessOrder: 2},
		{ID: "player-a", Active: true, Status: rules.PlayerActive, Score: 10, CompletionTime: 3, SuccessOrder: 1},
	}})
	if !decision.IsOver || decision.EndReason != string(EndTargetKillsReached) || len(decision.WinningPlayerIDs) != 1 || decision.WinningPlayerIDs[0] != "player-a" {
		t.Fatalf("decision = %+v", decision)
	}
	if len(decision.Players) != 2 || decision.Players[0].TargetValue != 10 || decision.Players[1].TargetValue != 10 {
		t.Fatalf("target values = %+v", decision.Players)
	}
}

func TestScoreAttackFailsWhenNoActiveParticipantsReachTarget(t *testing.T) {
	resolved, _ := Resolve(RoomModeConfig{PresetID: PresetScoreAttack, StartingLives: 3, TargetScore: 100}, teams.Config{Structure: teams.StructureFFA})
	decision := EvaluateMatch(resolved, MatchFacts{HadParticipants: true, Players: []PlayerFact{{ID: "player-a", Status: rules.PlayerEliminated, Score: 90}}})
	if !decision.IsOver || decision.TerminalStatus != rules.TerminalFailed || decision.Players[0].Outcome != rules.OutcomeFailed {
		t.Fatalf("decision = %+v", decision)
	}
}
