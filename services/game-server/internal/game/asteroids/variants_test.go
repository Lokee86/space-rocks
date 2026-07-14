package asteroids

import (
	"fmt"
	"testing"
)

func TestCountReturnsEight(t *testing.T) {
	if got, want := Count(), 8; got != want {
		t.Fatalf("Count() = %d, want %d", got, want)
	}
}

func TestCurrentVariantIndexesAreZeroBased(t *testing.T) {
	if got, want := Variants[0].Index, 0; got != want {
		t.Fatalf("Variants[0].Index = %d, want %d", got, want)
	}
	if got, want := Variants[7].Index, 7; got != want {
		t.Fatalf("Variants[7].Index = %d, want %d", got, want)
	}
}

func TestByIndexWrapsAcrossVariantCatalog(t *testing.T) {
	if got, want := ByIndex(0).ID, "asteroid_1"; got != want {
		t.Fatalf("ByIndex(0).ID = %q, want %q", got, want)
	}
	if got, want := ByIndex(7).ID, "asteroid_8"; got != want {
		t.Fatalf("ByIndex(7).ID = %q, want %q", got, want)
	}
	if got, want := ByIndex(8).ID, "asteroid_1"; got != want {
		t.Fatalf("ByIndex(8).ID = %q, want %q", got, want)
	}
}

func TestTimedSpawnVariantsReturnAllCurrentVariants(t *testing.T) {
	assertCurrentVariants(t, TimedSpawnVariants())
}

func TestFragmentSpawnVariantsReturnAllCurrentVariants(t *testing.T) {
	assertCurrentVariants(t, FragmentSpawnVariants())
}

func TestDebugSpawnVariantsReturnAllCurrentVariants(t *testing.T) {
	assertCurrentVariants(t, DebugSpawnVariants())
}

func TestCurrentVariantsKeepRequiredFieldsAndWeights(t *testing.T) {
	for i, variant := range Variants {
		wantID := fmt.Sprintf("asteroid_%d", i+1)
		if variant.ID != wantID {
			t.Fatalf("variant %d ID = %q, want %q", i, variant.ID, wantID)
		}
		if variant.CollisionShape == "" {
			t.Fatalf("variant %d CollisionShape is empty", i)
		}
		if variant.StatsProfile == "" {
			t.Fatalf("variant %d StatsProfile is empty", i)
		}
		if variant.DropTable == "" {
			t.Fatalf("variant %d DropTable is empty", i)
		}
		if variant.TimedSpawnWeight != 1.0 {
			t.Fatalf("variant %d TimedSpawnWeight = %v, want %v", i, variant.TimedSpawnWeight, 1.0)
		}
		if variant.FragmentSpawnWeight != 1.0 {
			t.Fatalf("variant %d FragmentSpawnWeight = %v, want %v", i, variant.FragmentSpawnWeight, 1.0)
		}
		if variant.DebugSpawnWeight != 1.0 {
			t.Fatalf("variant %d DebugSpawnWeight = %v, want %v", i, variant.DebugSpawnWeight, 1.0)
		}
	}
}

func TestTimedSpawnVariantIndexUsesDeterministicWeightedSelection(t *testing.T) {
	withVariants(t, []Variant{
		{ID: "asteroid_one", Index: 0, TimedSpawnWeight: 1.0},
		{ID: "asteroid_two", Index: 1, TimedSpawnWeight: 2.0},
		{ID: "asteroid_three", Index: 2, TimedSpawnWeight: 3.0},
	}, func(t *testing.T) {
		tests := []struct {
			name string
			roll float64
			want int
		}{
			{name: "first bucket", roll: 0.0, want: 0},
			{name: "second bucket", roll: 0.25, want: 1},
			{name: "third bucket", roll: 0.50, want: 2},
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				if got := TimedSpawnVariantIndex(tc.roll); got != tc.want {
					t.Fatalf("TimedSpawnVariantIndex(%v) = %d, want %d", tc.roll, got, tc.want)
				}
			})
		}
	})
}

func TestTimedSpawnVariantIndexSkipsZeroWeightVariants(t *testing.T) {
	withVariants(t, []Variant{
		{ID: "asteroid_zero", Index: 0, TimedSpawnWeight: 0.0},
		{ID: "asteroid_one", Index: 1, TimedSpawnWeight: 1.0},
	}, func(t *testing.T) {
		if got, want := TimedSpawnVariantIndex(0.0), 1; got != want {
			t.Fatalf("TimedSpawnVariantIndex(0.0) = %d, want %d", got, want)
		}
	})
}

func TestTimedSpawnVariantIndexUsesFirstAndLastBoundaries(t *testing.T) {
	withVariants(t, []Variant{
		{ID: "asteroid_one", Index: 0, TimedSpawnWeight: 1.0},
		{ID: "asteroid_two", Index: 1, TimedSpawnWeight: 1.0},
		{ID: "asteroid_three", Index: 2, TimedSpawnWeight: 1.0},
	}, func(t *testing.T) {
		if got, want := TimedSpawnVariantIndex(0.0), 0; got != want {
			t.Fatalf("TimedSpawnVariantIndex(0.0) = %d, want %d", got, want)
		}
		if got, want := TimedSpawnVariantIndex(1.0), 2; got != want {
			t.Fatalf("TimedSpawnVariantIndex(1.0) = %d, want %d", got, want)
		}
	})
}

func TestTimedSpawnVariantIndexFallsBackWhenAllWeightsAreZero(t *testing.T) {
	withVariants(t, []Variant{
		{ID: "asteroid_zero", Index: 0, TimedSpawnWeight: 0.0},
		{ID: "asteroid_one", Index: 1, TimedSpawnWeight: 0.0},
	}, func(t *testing.T) {
		if got, want := TimedSpawnVariantIndex(0.5), 0; got != want {
			t.Fatalf("TimedSpawnVariantIndex(0.5) = %d, want %d", got, want)
		}
	})
}

func TestSpawnVariantSelectorsUseTheirOwnWeightFields(t *testing.T) {
	withVariants(t, []Variant{
		{ID: "asteroid_one", Index: 0, TimedSpawnWeight: 3.0, FragmentSpawnWeight: 0.0, DebugSpawnWeight: 0.0},
		{ID: "asteroid_two", Index: 1, TimedSpawnWeight: 0.0, FragmentSpawnWeight: 3.0, DebugSpawnWeight: 0.0},
		{ID: "asteroid_three", Index: 2, TimedSpawnWeight: 0.0, FragmentSpawnWeight: 0.0, DebugSpawnWeight: 3.0},
	}, func(t *testing.T) {
		const roll = 0.5

		if got, want := TimedSpawnVariantIndex(roll), 0; got != want {
			t.Fatalf("TimedSpawnVariantIndex(%v) = %d, want %d", roll, got, want)
		}
		if got, want := FragmentSpawnVariantIndex(roll), 1; got != want {
			t.Fatalf("FragmentSpawnVariantIndex(%v) = %d, want %d", roll, got, want)
		}
		if got, want := DebugSpawnVariantIndex(roll), 2; got != want {
			t.Fatalf("DebugSpawnVariantIndex(%v) = %d, want %d", roll, got, want)
		}
	})
}

func assertCurrentVariants(t *testing.T, variants []Variant) {
	t.Helper()

	if got, want := len(variants), 8; got != want {
		t.Fatalf("len(variants) = %d, want %d", got, want)
	}

	for i, variant := range variants {
		wantID := fmt.Sprintf("asteroid_%d", i+1)
		if variant.ID != wantID {
			t.Fatalf("variant %d ID = %q, want %q", i, variant.ID, wantID)
		}
	}
}

func withVariants(t *testing.T, variants []Variant, fn func(t *testing.T)) {
	t.Helper()

	originalVariants := Variants
	Variants = variants
	t.Cleanup(func() {
		Variants = originalVariants
	})

	fn(t)
}
