package main

import (
	"path/filepath"
	"testing"
)

func TestRuntimeScenarioMeasurementPathUsesConfiguredOutput(t *testing.T) {
	outputRoot := t.TempDir()
	t.Setenv(runtimeScenarioOutputEnv, outputRoot)
	got := runtimeScenarioMeasurementPath("measurement-results/game-server")
	want := filepath.Join(outputRoot, "measurements", "game-server")
	if got != want {
		t.Fatalf("measurement path = %q, want %q", got, want)
	}
}

func TestRuntimeScenarioMeasurementPathKeepsDefaultWithoutHarnessOutput(t *testing.T) {
	t.Setenv(runtimeScenarioOutputEnv, "")
	const defaultPath = "measurement-results/game-server"
	if got := runtimeScenarioMeasurementPath(defaultPath); got != defaultPath {
		t.Fatalf("measurement path = %q, want %q", got, defaultPath)
	}
}
