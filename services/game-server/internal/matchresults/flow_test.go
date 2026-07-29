package matchresults

import (
	"testing"

	"github.com/Lokee86/space-rocks/services/game-server/internal/game/rules"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/teams"
)

func TestEndOfMatchFlowRunsOnceAndLocksSummary(t *testing.T) {
	flow := NewEndOfMatchFlow()
	input := BuildInput{
		MatchID: "match-1", TraceID: "trace-1", ModeID: "baseline", Session: SessionMultiplayer,
		Participants: []ParticipantFact{
			{PlayerRef: PlayerRef{GamePlayerID: "player-1"}, Score: 10},
			{PlayerRef: PlayerRef{GamePlayerID: "player-2"}, Score: 20},
		},
		LockedDecision: rules.MatchDecision{
			TerminalStatus:   rules.TerminalCompleted,
			EndReason:        "simulation_complete",
			WinningPlayerIDs: []string{"player-2"},
			Players: []rules.PlayerDecision{
				{ID: "player-1", Outcome: rules.OutcomeLost, Placement: 2},
				{ID: "player-2", Outcome: rules.OutcomeWon, Placement: 1},
			},
		},
	}
	first, emitted, err := flow.Run(input)
	if err != nil || !emitted {
		t.Fatalf("first run emitted=%v err=%v", emitted, err)
	}
	input.Participants[1].Score = 999
	second, emitted, err := flow.Run(input)
	if err != nil || emitted {
		t.Fatalf("duplicate run emitted=%v err=%v", emitted, err)
	}
	if second.Participants[0].PlayerRef.GamePlayerID != "player-2" || second.Participants[0].FinalScore != 20 {
		t.Fatalf("locked summary changed: %+v", second.Participants)
	}
	first.Participants[0].FinalScore = 500
	third, _ := flow.Summary()
	if third.Participants[0].FinalScore != 20 {
		t.Fatalf("returned summary aliased stored summary: %+v", third.Participants)
	}
}

func TestResolveDecisionCopiesLockedOutcomesAndPlacements(t *testing.T) {
	decision, err := ResolveDecision(BuildInput{
		Session: SessionMultiplayer,
		Participants: []ParticipantFact{
			{PlayerRef: PlayerRef{GamePlayerID: "low"}, Score: 5},
			{PlayerRef: PlayerRef{GamePlayerID: "high"}, Score: 10},
		},
		LockedDecision: rules.MatchDecision{
			TerminalStatus:   rules.TerminalCompleted,
			EndReason:        "resolved_by_mode",
			WinningPlayerIDs: []string{"low"},
			Players: []rules.PlayerDecision{
				{ID: "low", Outcome: rules.OutcomeWon, Placement: 1},
				{ID: "high", Outcome: rules.OutcomeLost, Placement: 2},
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	low := participantResult(t, decision, "low")
	high := participantResult(t, decision, "high")
	if low.Outcome != OutcomeWon || low.Placement != 1 || high.Outcome != OutcomeLost || high.Placement != 2 {
		t.Fatalf("locked decision was not preserved: %+v", decision.Participants)
	}
	if len(decision.WinningPlayerRefs) != 1 || decision.WinningPlayerRefs[0].GamePlayerID != "low" {
		t.Fatalf("unexpected winners: %+v", decision.WinningPlayerRefs)
	}
}

func TestResolveDecisionRejectsIncompleteLockedDecision(t *testing.T) {
	_, err := ResolveDecision(BuildInput{
		Session: SessionMultiplayer,
		Participants: []ParticipantFact{
			{PlayerRef: PlayerRef{GamePlayerID: "a"}, Score: 10},
			{PlayerRef: PlayerRef{GamePlayerID: "b"}, Score: 10},
		},
		LockedDecision: rules.MatchDecision{
			TerminalStatus: rules.TerminalCompleted,
			EndReason:      "incomplete",
			Players:        []rules.PlayerDecision{{ID: "a", Outcome: rules.OutcomeDraw}},
		},
	})
	if err == nil {
		t.Fatal("expected incomplete locked decision to fail")
	}
}

func TestResolveDecisionProducesTeamResultsFromLockedOutcomes(t *testing.T) {
	decision, err := ResolveDecision(BuildInput{
		Session: SessionMultiplayer, TeamStructure: teams.StructureCustom,
		Participants: []ParticipantFact{
			{PlayerRef: PlayerRef{GamePlayerID: "a"}, TeamID: teams.Team1, Score: 4},
			{PlayerRef: PlayerRef{GamePlayerID: "b"}, TeamID: teams.Team1, Score: 6},
			{PlayerRef: PlayerRef{GamePlayerID: "c"}, TeamID: teams.Team2, Score: 8},
		},
		LockedDecision: rules.MatchDecision{
			TerminalStatus: rules.TerminalCompleted,
			EndReason:      "resolved_by_mode",
			Players: []rules.PlayerDecision{
				{ID: "a", Outcome: rules.OutcomeWon},
				{ID: "b", Outcome: rules.OutcomeWon},
				{ID: "c", Outcome: rules.OutcomeLost},
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(decision.Teams) != 2 || decision.Teams[0].TeamID != teams.Team1 || decision.Teams[0].FinalScore != 10 || decision.Teams[0].Outcome != OutcomeWon {
		t.Fatalf("unexpected team decision: %+v", decision.Teams)
	}
	if len(decision.WinningTeamRefs) != 1 || decision.WinningTeamRefs[0] != teams.Team1 {
		t.Fatalf("unexpected team winners: %+v", decision.WinningTeamRefs)
	}
}
