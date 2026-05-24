package gametests

import (
	"testing"

	"github.com/Lokee86/space-rocks/server/internal/game/rules"
)

func TestEvaluateMatchCurrentGameOverSemantics(t *testing.T) {
	tests := []struct {
		name        string
		snapshot    rules.MatchSnapshot
		wantOver    bool
		wantPlayers []rules.PlayerDecision
	}{
		{
			name:        "no players",
			snapshot:    rules.MatchSnapshot{},
			wantOver:    false,
			wantPlayers: []rules.PlayerDecision{},
		},
		{
			name: "active player",
			snapshot: rules.MatchSnapshot{Players: []rules.PlayerSnapshot{
				{ID: "player-1", HasActiveShip: true},
			}},
			wantOver: false,
			wantPlayers: []rules.PlayerDecision{
				{ID: "player-1", Status: rules.PlayerActive},
			},
		},
		{
			name: "pending respawn",
			snapshot: rules.MatchSnapshot{Players: []rules.PlayerSnapshot{
				{ID: "player-1", HasRemainingLives: true},
			}},
			wantOver: false,
			wantPlayers: []rules.PlayerDecision{
				{ID: "player-1", Status: rules.PlayerPendingRespawn},
			},
		},
		{
			name: "eliminated player",
			snapshot: rules.MatchSnapshot{Players: []rules.PlayerSnapshot{
				{ID: "player-1"},
			}},
			wantOver: true,
			wantPlayers: []rules.PlayerDecision{
				{ID: "player-1", Status: rules.PlayerEliminated},
			},
		},
		{
			name: "all eliminated",
			snapshot: rules.MatchSnapshot{Players: []rules.PlayerSnapshot{
				{ID: "player-1"},
				{ID: "player-2"},
			}},
			wantOver: true,
			wantPlayers: []rules.PlayerDecision{
				{ID: "player-1", Status: rules.PlayerEliminated},
				{ID: "player-2", Status: rules.PlayerEliminated},
			},
		},
		{
			name: "mixed participating players preserve order",
			snapshot: rules.MatchSnapshot{Players: []rules.PlayerSnapshot{
				{ID: "player-1"},
				{ID: "player-2", HasRemainingLives: true},
				{ID: "player-3", HasActiveShip: true},
			}},
			wantOver: false,
			wantPlayers: []rules.PlayerDecision{
				{ID: "player-1", Status: rules.PlayerEliminated},
				{ID: "player-2", Status: rules.PlayerPendingRespawn},
				{ID: "player-3", Status: rules.PlayerActive},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			decision := rules.EvaluateMatch(test.snapshot)
			if decision.IsOver != test.wantOver {
				t.Fatalf("expected IsOver %t, got %t", test.wantOver, decision.IsOver)
			}
			if len(decision.Players) != len(test.wantPlayers) {
				t.Fatalf("expected %d player decisions, got %d", len(test.wantPlayers), len(decision.Players))
			}
			for index, wantPlayer := range test.wantPlayers {
				gotPlayer := decision.Players[index]
				if gotPlayer != wantPlayer {
					t.Fatalf("expected player decision %d to be %+v, got %+v", index, wantPlayer, gotPlayer)
				}
			}
		})
	}
}
