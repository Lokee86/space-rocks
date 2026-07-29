package rooms

import (
	"testing"

	"github.com/Lokee86/space-rocks/services/game-server/internal/game/rules"
	"github.com/Lokee86/space-rocks/services/game-server/internal/matchresults"
	"github.com/Lokee86/space-rocks/services/game-server/internal/playerdata"
)

func TestRoomMatchEndOfMatchFlowRunsOnceAndPreservesPresetPersistence(t *testing.T) {
	roomMatch := newRoomMatch(nil)
	preset := playerdata.MatchResultSummary{MatchID: "preset-match", Mode: playerdata.MatchModeMultiplayer}
	roomMatch.SetResolvedSummary(preset)
	input := matchresults.BuildInput{
		MatchID: "match-1", Session: matchresults.SessionMultiplayer,
		Participants: []matchresults.ParticipantFact{{
			PlayerRef: matchresults.PlayerRef{GamePlayerID: "player-1"}, Score: 10,
		}},
		LockedDecision: rules.MatchDecision{
			TerminalStatus: rules.TerminalCompleted,
			EndReason:      "simulation_complete",
			Players:        []rules.PlayerDecision{{ID: "player-1", Outcome: rules.OutcomeCompleted}},
		},
	}

	first, dispatch, emitted, err := roomMatch.ResolveEndOfMatch(input)
	if err != nil || !emitted {
		t.Fatalf("first resolve emitted=%v err=%v", emitted, err)
	}
	if first.MatchID != "match-1" {
		t.Fatalf("unexpected full summary: %+v", first)
	}
	if dispatch.Persistence.MatchID != "preset-match" {
		t.Fatalf("preset persistence was replaced: %+v", dispatch.Persistence)
	}

	input.Participants[0].Score = 999
	second, _, emitted, err := roomMatch.ResolveEndOfMatch(input)
	if err != nil || emitted {
		t.Fatalf("duplicate resolve emitted=%v err=%v", emitted, err)
	}
	if second.Participants[0].FinalScore != 10 {
		t.Fatalf("duplicate changed summary: %+v", second.Participants)
	}
	second.Participants[0].FinalScore = 500
	stored, ok := roomMatch.MatchSummary()
	if !ok || stored.Participants[0].FinalScore != 10 {
		t.Fatalf("stored summary was mutable: %+v", stored.Participants)
	}
}
