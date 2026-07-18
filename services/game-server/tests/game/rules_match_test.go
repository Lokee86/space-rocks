package gametests

import (
	"testing"

	playerstate "github.com/Lokee86/space-rocks/services/game-server/internal/game/player"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/rules"
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
			name: "no active players after participation",
			snapshot: rules.MatchSnapshot{
				HadParticipants: true,
			},
			wantOver:    true,
			wantPlayers: []rules.PlayerDecision{},
		},
		{
			name: "active player",
			snapshot: rules.MatchSnapshot{Players: []rules.PlayerSnapshot{
				{ID: "player-1", Status: playerstate.StatusActive, HasActiveShip: true},
			}},
			wantOver: false,
			wantPlayers: []rules.PlayerDecision{
				{ID: "player-1", Status: playerstate.StatusActive},
			},
		},
		{
			name: "pending respawn",
			snapshot: rules.MatchSnapshot{Players: []rules.PlayerSnapshot{
				{ID: "player-1", Status: playerstate.StatusPendingRespawn, HasRemainingLives: true},
			}},
			wantOver: false,
			wantPlayers: []rules.PlayerDecision{
				{ID: "player-1", Status: playerstate.StatusPendingRespawn},
			},
		},
		{
			name: "eliminated player",
			snapshot: rules.MatchSnapshot{Players: []rules.PlayerSnapshot{
				{ID: "player-1", Status: playerstate.StatusEliminated},
			}},
			wantOver: true,
			wantPlayers: []rules.PlayerDecision{
				{ID: "player-1", Status: playerstate.StatusEliminated},
			},
		},
		{
			name: "all eliminated",
			snapshot: rules.MatchSnapshot{Players: []rules.PlayerSnapshot{
				{ID: "player-1", Status: playerstate.StatusEliminated},
				{ID: "player-2", Status: playerstate.StatusEliminated},
			}},
			wantOver: true,
			wantPlayers: []rules.PlayerDecision{
				{ID: "player-1", Status: playerstate.StatusEliminated},
				{ID: "player-2", Status: playerstate.StatusEliminated},
			},
		},
		{
			name: "mixed participating players preserve order",
			snapshot: rules.MatchSnapshot{Players: []rules.PlayerSnapshot{
				{ID: "player-1", Status: playerstate.StatusEliminated},
				{ID: "player-2", Status: playerstate.StatusPendingRespawn, HasRemainingLives: true},
				{ID: "player-3", Status: playerstate.StatusActive, HasActiveShip: true},
			}},
			wantOver: false,
			wantPlayers: []rules.PlayerDecision{
				{ID: "player-1", Status: playerstate.StatusEliminated},
				{ID: "player-2", Status: playerstate.StatusPendingRespawn},
				{ID: "player-3", Status: playerstate.StatusActive},
			},
		},
		{
			name: "authoritative status wins over projection booleans",
			snapshot: rules.MatchSnapshot{Players: []rules.PlayerSnapshot{
				{ID: "player-1", Status: playerstate.StatusEliminated, HasActiveShip: true, HasRemainingLives: true},
			}},
			wantOver: true,
			wantPlayers: []rules.PlayerDecision{
				{ID: "player-1", Status: playerstate.StatusEliminated},
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
