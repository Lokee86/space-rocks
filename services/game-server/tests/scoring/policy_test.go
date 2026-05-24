package scoringtests

import (
	"testing"

	"github.com/Lokee86/space-rocks/server/internal/constants"
	"github.com/Lokee86/space-rocks/server/internal/game/scoring"
)

func TestAsteroidDestroyedAwardsBaseScoreByAsteroidSize(t *testing.T) {
	policy := scoring.NewDefaultPolicy()

	awards := policy.Evaluate(scoring.Event{
		Kind:         scoring.EventAsteroidDestroyed,
		PlayerID:     "player-1",
		TargetID:     "asteroid-1",
		AsteroidSize: 3,
	})

	if len(awards) != 1 {
		t.Fatalf("expected 1 award, got %d", len(awards))
	}
	if awards[0].PlayerID != "player-1" {
		t.Fatalf("expected player-1 award, got %q", awards[0].PlayerID)
	}
	if awards[0].Points != constants.BaseScore/3 {
		t.Fatalf("expected %d points, got %d", constants.BaseScore/3, awards[0].Points)
	}
	if awards[0].Reason != scoring.EventAsteroidDestroyed {
		t.Fatalf("expected asteroid destroyed reason, got %q", awards[0].Reason)
	}
}

func TestAsteroidDestroyedWithoutPlayerIDReturnsNoAward(t *testing.T) {
	policy := scoring.NewDefaultPolicy()

	awards := policy.Evaluate(scoring.Event{
		Kind:         scoring.EventAsteroidDestroyed,
		AsteroidSize: 1,
	})

	if len(awards) != 0 {
		t.Fatalf("expected no awards, got %d", len(awards))
	}
}

func TestAsteroidDestroyedWithNonPositiveSizeReturnsNoAward(t *testing.T) {
	policy := scoring.NewDefaultPolicy()

	for _, asteroidSize := range []int{0, -1} {
		awards := policy.Evaluate(scoring.Event{
			Kind:         scoring.EventAsteroidDestroyed,
			PlayerID:     "player-1",
			AsteroidSize: asteroidSize,
		})

		if len(awards) != 0 {
			t.Fatalf("expected no awards for asteroid size %d, got %d", asteroidSize, len(awards))
		}
	}
}

func TestUnknownEventKindReturnsNoAward(t *testing.T) {
	policy := scoring.NewDefaultPolicy()

	awards := policy.Evaluate(scoring.Event{
		Kind:         scoring.EventKind("unknown"),
		PlayerID:     "player-1",
		AsteroidSize: 1,
	})

	if len(awards) != 0 {
		t.Fatalf("expected no awards, got %d", len(awards))
	}
}
