package lives

import (
	"testing"

	playerstate "github.com/Lokee86/space-rocks/services/game-server/internal/game/player"
)

func TestEvaluateRespawnFacts(t *testing.T) {
	policy := NewBaselinePolicy()
	tests := []struct {
		name           string
		status         playerstate.Status
		remainingLives int
		cooldown       float64
		accepted       bool
		resulting      playerstate.Status
		reason         string
	}{
		{
			name:           "accepted pending respawn",
			status:         playerstate.StatusPendingRespawn,
			remainingLives: 2,
			accepted:       true,
			resulting:      playerstate.StatusActive,
		},
		{
			name:           "active",
			status:         playerstate.StatusActive,
			remainingLives: 2,
			reason:         "already_active",
		},
		{
			name:   "eliminated",
			status: playerstate.StatusEliminated,
			reason: "respawn_cooldown_or_lives_exhausted",
		},
		{
			name:           "no lives",
			status:         playerstate.StatusPendingRespawn,
			remainingLives: 0,
			reason:         "respawn_cooldown_or_lives_exhausted",
		},
		{
			name:           "cooldown",
			status:         playerstate.StatusPendingRespawn,
			remainingLives: 2,
			cooldown:       1,
			reason:         "respawn_cooldown_or_lives_exhausted",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fact := policy.EvaluateRespawn("player-1", test.status, test.remainingLives, test.cooldown)
			if fact.Accepted != test.accepted {
				t.Fatalf("Accepted = %t, want %t; fact = %+v", fact.Accepted, test.accepted, fact)
			}
			if fact.PreviousStatus != test.status {
				t.Fatalf("PreviousStatus = %q, want %q", fact.PreviousStatus, test.status)
			}
			wantResultingStatus := test.resulting
			if !test.accepted {
				wantResultingStatus = test.status
			}
			if fact.ResultingStatus != wantResultingStatus {
				t.Fatalf("ResultingStatus = %q, want %q", fact.ResultingStatus, wantResultingStatus)
			}
			if fact.ReasonCode != test.reason {
				t.Fatalf("ReasonCode = %q, want %q", fact.ReasonCode, test.reason)
			}
		})
	}
}

func TestRejectedRespawnFactForMissingSession(t *testing.T) {
	fact := RejectedRespawn("player-1", "session_missing")
	if fact.Accepted {
		t.Fatal("expected missing-session respawn fact to be rejected")
	}
	if fact.PlayerID != "player-1" || fact.ReasonCode != "session_missing" {
		t.Fatalf("unexpected missing-session respawn fact: %+v", fact)
	}
}
