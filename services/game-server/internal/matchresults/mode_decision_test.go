package matchresults

import (
	"testing"

	"github.com/Lokee86/space-rocks/services/game-server/internal/game/rules"
)

func TestResolveDecisionUsesLockedScoreAttackResult(t *testing.T) {
	decision, err := ResolveDecision(BuildInput{
		ModeID: "score_attack", Session: SessionMultiplayer,
		Participants: []ParticipantFact{
			{PlayerRef: PlayerRef{GamePlayerID: "player-1"}, Score: 500},
			{PlayerRef: PlayerRef{GamePlayerID: "player-2"}, Score: 1000},
		},
		LockedDecision: rules.MatchDecision{
			IsOver: true, TerminalStatus: rules.TerminalCompleted, EndReason: "target_score_reached",
			WinningPlayerIDs: []string{"player-1"},
			Players: []rules.PlayerDecision{
				{ID: "player-1", Outcome: rules.OutcomeWon, Placement: 1, CompletionTime: 8.5, TargetValue: 500},
				{ID: "player-2", Outcome: rules.OutcomeLost, TargetValue: 500},
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	playerOne := participantResult(t, decision, "player-1")
	if playerOne.Outcome != OutcomeWon || playerOne.Placement != 1 || playerOne.CompletionTime != 8.5 || playerOne.TargetValue != 500 {
		t.Fatalf("player one = %+v", playerOne)
	}
	if len(decision.WinningPlayerRefs) != 1 || decision.WinningPlayerRefs[0].GamePlayerID != "player-1" {
		t.Fatalf("winners = %+v", decision.WinningPlayerRefs)
	}
}

func participantResult(t *testing.T, decision MatchDecision, playerID string) ParticipantResult {
	t.Helper()
	for _, participant := range decision.Participants {
		if participant.PlayerRef.GamePlayerID == playerID {
			return participant
		}
	}
	t.Fatalf("missing participant %q", playerID)
	return ParticipantResult{}
}
