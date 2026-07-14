package game

import "testing"

func TestNewWithSeedPreservesSimulationSeed(t *testing.T) {
	const seed int64 = 8675309

	game := NewWithSeed(seed)
	if game == nil {
		t.Fatal("NewWithSeed returned nil")
	}
	if game.rngSource == nil {
		t.Fatal("NewWithSeed did not initialize rng source")
	}
	if game.spawner == nil {
		t.Fatal("NewWithSeed did not initialize spawner")
	}
	if got := game.SimulationSeed(); got != seed {
		t.Fatalf("SimulationSeed() = %d, want %d", got, seed)
	}
}

func TestNewCreatesInitializedProductionSeedSourcePath(t *testing.T) {
	game := New()
	if game == nil {
		t.Fatal("New returned nil")
	}
	if game.rngSource == nil {
		t.Fatal("New did not initialize rng source")
	}
	if game.spawner == nil {
		t.Fatal("New did not initialize spawner")
	}
	if got, want := game.SimulationSeed(), game.rngSource.Seed(); got != want {
		t.Fatalf("SimulationSeed() = %d, want %d", got, want)
	}
}
