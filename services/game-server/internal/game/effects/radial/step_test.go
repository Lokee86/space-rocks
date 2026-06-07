package radial

import (
	"testing"

	"github.com/Lokee86/space-rocks/server/internal/game/physics"
)

func TestStepRespectsTargetFilterForAsteroidsAndEnemies(t *testing.T) {
	effect := Effect{
		ID:         "effect-1",
		Origin:     zeroVector(),
		Spec:       annularSimultaneousSpec(TargetFilter{Asteroids: true, Enemies: true}),
		Zones:      buildZones(annularSimultaneousSpec(TargetFilter{Asteroids: true, Enemies: true})),
		AgeSeconds: 0,
	}

	result := Step(&effect, 0.1, []Candidate{
		{ID: "asteroid-1", Kind: TargetAsteroid, Position: vectorAtDistance(5)},
		{ID: "enemy-1", Kind: TargetEnemy, Position: vectorAtDistance(5)},
		{ID: "player-1", Kind: TargetPlayer, Position: vectorAtDistance(5)},
		{ID: "projectile-1", Kind: TargetProjectile, Position: vectorAtDistance(5)},
		{ID: "pickup-1", Kind: TargetPickup, Position: vectorAtDistance(5)},
	})

	assertHitIDs(t, result.Hits, []string{"asteroid-1", "enemy-1"})
}

func TestStepCanIncludePlayersWhenConfigured(t *testing.T) {
	effect := Effect{
		ID:         "effect-2",
		Origin:     zeroVector(),
		Spec:       annularSimultaneousSpec(TargetFilter{Players: true}),
		Zones:      buildZones(annularSimultaneousSpec(TargetFilter{Players: true})),
		AgeSeconds: 0,
	}

	result := Step(&effect, 0.1, []Candidate{
		{ID: "player-1", Kind: TargetPlayer, Position: vectorAtDistance(5)},
	})

	assertHitIDs(t, result.Hits, []string{"player-1"})
}

func TestStepAnnularSimultaneousBloomSequence(t *testing.T) {
	spec := Spec{
		CoverageMode:        CoverageAnnularWave,
		ExpirationMode:      ExpirationSimultaneous,
		TargetFilter:        TargetFilter{Asteroids: true},
		ZoneCount:           4,
		ZoneWidth:           10,
		ZoneSpawnSeconds:    0.1,
		TickSeconds:         0.1,
		TotalSeconds:        0.4,
		ZoneLifetimeSeconds: 0,
	}

	effect := Effect{
		ID:     "effect-3",
		Origin: zeroVector(),
		Spec:   spec,
		Zones:  buildZones(spec),
	}

	candidates := []Candidate{
		{ID: "zone-1", Kind: TargetAsteroid, Position: vectorAtDistance(5)},
		{ID: "zone-2", Kind: TargetAsteroid, Position: vectorAtDistance(15)},
		{ID: "zone-3", Kind: TargetAsteroid, Position: vectorAtDistance(25)},
		{ID: "zone-4", Kind: TargetAsteroid, Position: vectorAtDistance(35)},
	}

	step := func() StepResult {
		return Step(&effect, 0.1, candidates)
	}

	result := step()
	assertHitIDs(t, result.Hits, []string{"zone-1"})
	if result.Expired {
		t.Fatal("first step expired, want false")
	}

	result = step()
	assertHitIDs(t, result.Hits, []string{"zone-1", "zone-2"})
	if result.Expired {
		t.Fatal("second step expired, want false")
	}

	result = step()
	assertHitIDs(t, result.Hits, []string{"zone-1", "zone-2", "zone-3"})
	if result.Expired {
		t.Fatal("third step expired, want false")
	}

	result = step()
	assertHitIDs(t, result.Hits, []string{"zone-1", "zone-2", "zone-3", "zone-4"})
	if result.Expired {
		t.Fatal("fourth step expired, want false")
	}

	result = step()
	assertHitIDs(t, result.Hits, nil)
	if !result.Expired {
		t.Fatal("fifth step expired = false, want true")
	}
}

func TestStepAnnularZoneHitsCandidateWhoseRadiusOverlapsZone(t *testing.T) {
	effect := Effect{
		ID:         "effect-radius",
		Origin:     zeroVector(),
		Spec:       annularSimultaneousSpec(TargetFilter{Asteroids: true}),
		Zones:      buildZones(annularSimultaneousSpec(TargetFilter{Asteroids: true})),
		AgeSeconds: 0,
	}

	result := Step(&effect, 0.1, []Candidate{
		{ID: "overlap", Kind: TargetAsteroid, Position: vectorAtDistance(12), Radius: 3},
		{ID: "miss", Kind: TargetAsteroid, Position: vectorAtDistance(12), Radius: 2},
	})

	assertHitIDs(t, result.Hits, []string{"overlap"})
}

func TestStepExpandingFillEarlyRadiusHitsNearTarget(t *testing.T) {
	spec := Spec{
		CoverageMode:        CoverageExpandingFill,
		ExpirationMode:      ExpirationSimultaneous,
		TargetFilter:        TargetFilter{Asteroids: true},
		ZoneCount:           4,
		ZoneWidth:           10,
		ZoneSpawnSeconds:    0.1,
		TickSeconds:         0.1,
		TotalSeconds:        0.4,
		ZoneLifetimeSeconds: 0,
	}

	effect := Effect{ID: "fill-1", Origin: zeroVector(), Spec: spec, Zones: buildZones(spec)}
	result := Step(&effect, 0.1, []Candidate{
		{ID: "near", Kind: TargetAsteroid, Position: vectorAtDistance(5)},
		{ID: "far", Kind: TargetAsteroid, Position: vectorAtDistance(15)},
	})

	assertHitIDs(t, result.Hits, []string{"near"})
}

func TestStepExpandingFillLaterRadiusHitsFarTarget(t *testing.T) {
	spec := Spec{
		CoverageMode:        CoverageExpandingFill,
		ExpirationMode:      ExpirationSimultaneous,
		TargetFilter:        TargetFilter{Asteroids: true},
		ZoneCount:           4,
		ZoneWidth:           10,
		ZoneSpawnSeconds:    0.1,
		TickSeconds:         0.1,
		TotalSeconds:        0.4,
		ZoneLifetimeSeconds: 0,
	}

	effect := Effect{ID: "fill-2", Origin: zeroVector(), Spec: spec, Zones: buildZones(spec)}
	Step(&effect, 0.1, []Candidate{{ID: "near", Kind: TargetAsteroid, Position: vectorAtDistance(5)}})
	result := Step(&effect, 0.1, []Candidate{
		{ID: "near", Kind: TargetAsteroid, Position: vectorAtDistance(5)},
		{ID: "far", Kind: TargetAsteroid, Position: vectorAtDistance(15)},
	})

	assertHitIDs(t, result.Hits, []string{"near", "far"})
}

func TestStepExpandingFillHitsCenterOncePerTick(t *testing.T) {
	spec := Spec{
		CoverageMode:        CoverageExpandingFill,
		ExpirationMode:      ExpirationSimultaneous,
		TargetFilter:        TargetFilter{Asteroids: true},
		ZoneCount:           4,
		ZoneWidth:           10,
		ZoneSpawnSeconds:    0.1,
		TickSeconds:         0.1,
		TotalSeconds:        0.4,
		ZoneLifetimeSeconds: 0,
	}

	effect := Effect{ID: "fill-3", Origin: zeroVector(), Spec: spec, Zones: buildZones(spec)}
	center := Candidate{ID: "center", Kind: TargetAsteroid, Position: vectorAtDistance(5)}

	result := Step(&effect, 0.1, []Candidate{center})
	assertHitIDs(t, result.Hits, []string{"center"})

	result = Step(&effect, 0.1, []Candidate{center})
	if got := countHitsForTarget(result.Hits, "center"); got != 1 {
		t.Fatalf("center hits on second tick = %d, want 1", got)
	}
}

func TestStepExpandingFillHitsCandidateWhoseRadiusOverlapsFill(t *testing.T) {
	spec := Spec{
		CoverageMode:        CoverageExpandingFill,
		ExpirationMode:      ExpirationSimultaneous,
		TargetFilter:        TargetFilter{Asteroids: true},
		ZoneCount:           1,
		ZoneWidth:           10,
		ZoneSpawnSeconds:    0,
		TickSeconds:         0.1,
		TotalSeconds:        0.4,
		ZoneLifetimeSeconds: 0,
	}

	effect := Effect{ID: "fill-radius", Origin: zeroVector(), Spec: spec, Zones: buildZones(spec)}
	result := Step(&effect, 0.1, []Candidate{
		{ID: "overlap", Kind: TargetAsteroid, Position: vectorAtDistance(12), Radius: 3},
		{ID: "miss", Kind: TargetAsteroid, Position: vectorAtDistance(12), Radius: 2},
	})

	assertHitIDs(t, result.Hits, []string{"overlap"})
}

func annularSimultaneousSpec(filter TargetFilter) Spec {
	return Spec{
		CoverageMode:        CoverageAnnularWave,
		ExpirationMode:      ExpirationSimultaneous,
		TargetFilter:        filter,
		ZoneCount:           1,
		ZoneWidth:           10,
		ZoneSpawnSeconds:    0,
		TickSeconds:         0.1,
		TotalSeconds:        0.4,
		ZoneLifetimeSeconds: 0,
	}
}

func vectorAtDistance(distance float64) physics.Vector2 {
	return physics.Vector2{X: distance}
}

func zeroVector() physics.Vector2 {
	return physics.Vector2{}
}

func assertHitIDs(t *testing.T, hits []Hit, want []string) {
	t.Helper()
	if got, wantLen := len(hits), len(want); got != wantLen {
		t.Fatalf("len(hits) = %d, want %d", got, wantLen)
	}
	for i := range want {
		if hits[i].TargetID != want[i] {
			t.Fatalf("hit %d target id = %q, want %q", i, hits[i].TargetID, want[i])
		}
	}
}

func countHitsForTarget(hits []Hit, targetID string) int {
	count := 0
	for _, hit := range hits {
		if hit.TargetID == targetID {
			count++
		}
	}
	return count
}
