package main

import "testing"

func TestRuntimeScenarioGameFactoryFromEnv(t *testing.T) {
	t.Setenv(runtimeScenarioSeedEnv, "314159")
	factory, err := runtimeScenarioGameFactoryFromEnv()
	if err != nil {
		t.Fatalf("factory: %v", err)
	}
	if factory == nil {
		t.Fatal("expected configured factory")
	}
	if got := factory().SimulationSeed(); got != 314159 {
		t.Fatalf("simulation seed = %d, want 314159", got)
	}
}

func TestRuntimeScenarioGameFactoryFromEnvRejectsInvalidSeed(t *testing.T) {
	t.Setenv(runtimeScenarioSeedEnv, "not-a-seed")
	if _, err := runtimeScenarioGameFactoryFromEnv(); err == nil {
		t.Fatal("expected invalid seed error")
	}
}

func TestRuntimeScenarioGameFactoryFromEnvAllowsUnsetSeed(t *testing.T) {
	t.Setenv(runtimeScenarioSeedEnv, "")
	factory, err := runtimeScenarioGameFactoryFromEnv()
	if err != nil {
		t.Fatalf("factory: %v", err)
	}
	if factory != nil {
		t.Fatal("expected normal unconfigured factory")
	}
}
