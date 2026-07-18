package matchresults

import "testing"

func TestDispatcherProducesCompatibilityAndPresentationSafeSlices(t *testing.T) {
	summary := MatchSummary{
		MatchID: "match-1", TraceID: "trace-1", ModeID: "score_attack", Session: SessionMultiplayer,
		TerminalStatus: TerminalCompleted, EndReason: "target_reached",
		Participants: []ParticipantResult{{
			PlayerRef: PlayerRef{GamePlayerID: "player-1", AccountID: "account-secret", LocalProfileID: "local-secret"},
			Outcome:   OutcomeWon, Placement: 1, FinalScore: 100, ShipDeaths: 2,
		}},
		Objectives: []ObjectiveResolution{
			{InstanceID: "visible", Status: "completed", Discovered: true},
			{InstanceID: "hidden", Status: "active", Discovered: false},
		},
	}
	dispatch := (MatchSummaryDispatcher{}).Dispatch(summary)
	if len(dispatch.Persistence.Players) != 1 || dispatch.Persistence.Players[0].AccountID != "account-secret" || !dispatch.Persistence.Players[0].Won {
		t.Fatalf("unexpected persistence slice: %+v", dispatch.Persistence)
	}
	if len(dispatch.Presentation.Players) != 1 || dispatch.Presentation.Players[0].GamePlayerID != "player-1" {
		t.Fatalf("unexpected presentation players: %+v", dispatch.Presentation.Players)
	}
	if len(dispatch.Presentation.Objectives) != 1 || dispatch.Presentation.Objectives[0].InstanceID != "visible" {
		t.Fatalf("hidden objective leaked: %+v", dispatch.Presentation.Objectives)
	}
	if len(dispatch.Progression) != 1 || dispatch.Progression[0].PlayerRef.AccountID != "account-secret" {
		t.Fatalf("unexpected progression slice: %+v", dispatch.Progression)
	}
	foundWin := false
	foundObjective := false
	for _, fact := range dispatch.AchievementFacts {
		foundWin = foundWin || fact.Kind == "match_won"
		foundObjective = foundObjective || (fact.Kind == "objective_completed" && fact.ObjectiveID == "visible")
	}
	if !foundWin || !foundObjective {
		t.Fatalf("missing achievement facts: %+v", dispatch.AchievementFacts)
	}
}
