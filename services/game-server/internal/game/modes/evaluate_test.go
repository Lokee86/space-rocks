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

func TestTeamDeathmatchAwardsEveryMemberOfFirstTeamToTarget(t *testing.T) {
	resolved, _ := Resolve(
		RoomModeConfig{PresetID: PresetTeamDeathmatch, TargetKills: 10},
		teams.Config{Structure: teams.StructureAutoBalanced, AutoTeamCount: 2},
	)
	decision := EvaluateMatch(resolved, MatchFacts{
		HadParticipants: true,
		Players: []PlayerFact{
			{ID: "blue-a", TeamID: teams.Team1, Active: true, Status: rules.PlayerActive, Score: 6},
			{ID: "blue-b", TeamID: teams.Team1, Active: true, Status: rules.PlayerActive, Score: 4},
			{ID: "red-a", TeamID: teams.Team2, Active: true, Status: rules.PlayerActive, Score: 9},
		},
		Teams: []TeamFact{
			{ID: teams.Team1, Score: 10, CompletionTime: 7, SuccessOrder: 1},
			{ID: teams.Team2, Score: 9},
		},
	})
	if !decision.IsOver || len(decision.WinningPlayerIDs) != 2 {
		t.Fatalf("decision = %+v", decision)
	}
	if decision.WinningPlayerIDs[0] != "blue-a" || decision.WinningPlayerIDs[1] != "blue-b" {
		t.Fatalf("winning players = %+v", decision.WinningPlayerIDs)
	}
	if decision.Players[0].Outcome != rules.OutcomeWon || decision.Players[1].Outcome != rules.OutcomeWon || decision.Players[2].Outcome != rules.OutcomeLost {
		t.Fatalf("player outcomes = %+v", decision.Players)
	}
}

func TestScoreAttackFailsWhenNoActiveParticipantsReachTarget(t *testing.T) {
	resolved, _ := Resolve(RoomModeConfig{PresetID: PresetScoreAttack, StartingLives: 3, TargetScore: 100}, teams.Config{Structure: teams.StructureFFA})
	decision := EvaluateMatch(resolved, MatchFacts{HadParticipants: true, Players: []PlayerFact{{ID: "player-a", Status: rules.PlayerEliminated, Score: 90}}})
	if !decision.IsOver || decision.TerminalStatus != rules.TerminalFailed || decision.Players[0].Outcome != rules.OutcomeFailed {
		t.Fatalf("decision = %+v", decision)
	}
}
