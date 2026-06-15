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

func TestWeightedSelectionSkipsZeroWeightVariants(t *testing.T) {
	originalVariants := Variants
	t.Cleanup(func() {
		Variants = originalVariants
	})

	Variants = []Variant{
		{
			ID:               "asteroid_zero",
			Index:            0,
			TimedSpawnWeight: 0.0,
		},
		{
			ID:               "asteroid_one",
			Index:            1,
			TimedSpawnWeight: 1.0,
		},
	}

	if got, want := randomWeightedVariantIndex(func(variant Variant) float64 {
		return variant.TimedSpawnWeight
	}), 1; got != want {
		t.Fatalf("randomWeightedVariantIndex() = %d, want %d", got, want)
	}
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
