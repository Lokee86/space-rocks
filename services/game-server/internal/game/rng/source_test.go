package rng

import "testing"

func TestNewProducesEqualSequencesForEqualSeeds(t *testing.T) {
	left := New(12345)
	right := New(12345)

	for i, want := range []int{left.Intn(1000), left.Intn(1000), left.Intn(1000), left.Intn(1000), left.Intn(1000)} {
		if got := right.Intn(1000); got != want {
			t.Fatalf("sequence mismatch at step %d: got %d, want %d", i, got, want)
		}
	}

	for i, want := range []float64{left.Float64(), left.Float64(), left.Float64(), left.Float64(), left.Float64()} {
		if got := right.Float64(); got != want {
			t.Fatalf("float sequence mismatch at step %d: got %v, want %v", i, got, want)
		}
	}
}

func TestNewProducesDifferentRepresentativeSequencesForDifferentSeeds(t *testing.T) {
	left := New(12345)
	right := New(54321)

	leftInts := []int{
		left.Intn(1000),
		left.Intn(1000),
		left.Intn(1000),
		left.Intn(1000),
		left.Intn(1000),
	}
	rightInts := []int{
		right.Intn(1000),
		right.Intn(1000),
		right.Intn(1000),
		right.Intn(1000),
		right.Intn(1000),
	}

	leftFloats := []float64{left.Float64(), left.Float64(), left.Float64()}
	rightFloats := []float64{right.Float64(), right.Float64(), right.Float64()}

	if equalIntSequences(leftInts, rightInts) && equalFloatSequences(leftFloats, rightFloats) {
		t.Fatal("representative sequences unexpectedly matched for different seeds")
	}
}

func TestSeedReportsOriginalValue(t *testing.T) {
	const seed int64 = -987654321

	source := New(seed)

	if got := source.Seed(); got != seed {
		t.Fatalf("Seed() = %d, want %d", got, seed)
	}
}

func equalIntSequences(left, right []int) bool {
	if len(left) != len(right) {
		return false
	}

	for i := range left {
		if left[i] != right[i] {
			return false
		}
	}

	return true
}

func equalFloatSequences(left, right []float64) bool {
	if len(left) != len(right) {
		return false
	}

	for i := range left {
		if left[i] != right[i] {
			return false
		}
	}

	return true
}
