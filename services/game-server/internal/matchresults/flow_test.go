package matchresults

import (
	"testing"

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

func TestResolveDecisionSupportsWinnerDrawAndPlacements(t *testing.T) {
	decision, err := ResolveDecision(BuildInput{
		Session: SessionMultiplayer,
		Participants: []ParticipantFact{
			{PlayerRef: PlayerRef{GamePlayerID: "low"}, Score: 5},
			{PlayerRef: PlayerRef{GamePlayerID: "high"}, Score: 10},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if decision.Participants[0].Outcome != OutcomeWon || decision.Participants[0].Placement != 1 ||
		decision.Participants[1].Outcome != OutcomeLost || decision.Participants[1].Placement != 2 {
		t.Fatalf("unexpected decision: %+v", decision.Participants)
	}

	draw, err := ResolveDecision(BuildInput{
		Session: SessionMultiplayer,
		Participants: []ParticipantFact{
			{PlayerRef: PlayerRef{GamePlayerID: "a"}, Score: 10},
			{PlayerRef: PlayerRef{GamePlayerID: "b"}, Score: 10},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, participant := range draw.Participants {
		if participant.Outcome != OutcomeDraw || participant.Placement != 1 {
			t.Fatalf("unexpected draw participant: %+v", participant)
		}
	}
	if len(draw.WinningPlayerRefs) != 0 {
		t.Fatalf("draw has winners: %+v", draw.WinningPlayerRefs)
	}
}

func TestResolveDecisionProducesTeamResults(t *testing.T) {
	decision, err := ResolveDecision(BuildInput{
		Session: SessionMultiplayer, TeamStructure: teams.StructureCustom,
		Participants: []ParticipantFact{
			{PlayerRef: PlayerRef{GamePlayerID: "a"}, TeamID: teams.Team1, Score: 4},
			{PlayerRef: PlayerRef{GamePlayerID: "b"}, TeamID: teams.Team1, Score: 6},
			{PlayerRef: PlayerRef{GamePlayerID: "c"}, TeamID: teams.Team2, Score: 8},
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
