package gametests

import (
	"testing"

	"github.com/Lokee86/space-rocks/server/internal/game/rules"
)

func TestEvaluateMatchCurrentGameOverSemantics(t *testing.T) {
	tests := []struct {
		name     string
		snapshot rules.MatchSnapshot
		wantOver bool
	}{
		{
			name:     "no players",
			snapshot: rules.MatchSnapshot{},
			wantOver: false,
		},
		{
			name: "remaining lives",
			snapshot: rules.MatchSnapshot{Players: []rules.PlayerSnapshot{
				{ID: "player-1", HasRemainingLives: true},
			}},
			wantOver: false,
		},
		{
			name: "active ship",
			snapshot: rules.MatchSnapshot{Players: []rules.PlayerSnapshot{
				{ID: "player-1", HasActiveShip: true},
			}},
			wantOver: false,
		},
		{
			name: "all eliminated",
			snapshot: rules.MatchSnapshot{Players: []rules.PlayerSnapshot{
				{ID: "player-1"},
				{ID: "player-2"},
			}},
			wantOver: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			decision := rules.EvaluateMatch(test.snapshot)
			if decision.IsOver != test.wantOver {
				t.Fatalf("expected IsOver %t, got %t", test.wantOver, decision.IsOver)
			}
		})
	}
}
