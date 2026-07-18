package radial

import (
	"testing"

	"github.com/Lokee86/space-rocks/services/game-server/internal/game/damage"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/physics"
)

func TestStepAppliesLinearFalloffForFillAndAnnularCoverage(t *testing.T) {
	for _, coverage := range []CoverageMode{CoverageExpandingFill, CoverageAnnularWave} {
		t.Run(string(coverage), func(t *testing.T) {
			effect := NewEffect(SpawnRequest{
				ID:     "effect-1",
				Origin: physics.Vector2{},
				Spec: Spec{
					CoverageMode:        coverage,
					ExpirationMode:      ExpirationSimultaneous,
					FalloffMode:         FalloffLinear,
					MinimumMultiplier:   0.25,
					TargetFilter:        TargetFilter{Asteroids: true},
					ZoneCount:           2,
					ZoneWidth:           50,
					TickSeconds:         1,
					TotalSeconds:        10,
					ZoneLifetimeSeconds: 10,
					Damage:              damage.DamageSpec{Amount: 100},
				},
			})
			effect.AgeSeconds = 1
			result := Step(&effect, 0, []Candidate{{
				ID:       "asteroid-1",
				Kind:     TargetAsteroid,
				Position: physics.Vector2{X: 50},
			}})
			if len(result.Hits) != 1 {
				t.Fatalf("hit count = %d, want 1", len(result.Hits))
			}
			if got := result.Hits[0].Damage.Amount; got != 63 {
				t.Fatalf("falloff damage = %d, want 63", got)
			}
		})
	}
}

func TestStepPreservesDamageWhenFalloffIsNotSelected(t *testing.T) {
	effect := NewEffect(SpawnRequest{
		ID:     "effect-1",
		Origin: physics.Vector2{},
		Spec: Spec{
			CoverageMode:        CoverageExpandingFill,
			ExpirationMode:      ExpirationSimultaneous,
			TargetFilter:        TargetFilter{Asteroids: true},
			ZoneCount:           2,
			ZoneWidth:           50,
			TotalSeconds:        10,
			ZoneLifetimeSeconds: 10,
			Damage:              damage.DamageSpec{Amount: 100},
		},
	})
	effect.AgeSeconds = 1
	result := Step(&effect, 0, []Candidate{{ID: "asteroid-1", Kind: TargetAsteroid, Position: physics.Vector2{X: 50}}})
	if len(result.Hits) != 1 || result.Hits[0].Damage.Amount != 100 {
		t.Fatalf("hits = %+v", result.Hits)
	}
}
