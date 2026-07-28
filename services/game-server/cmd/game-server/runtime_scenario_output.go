package main

import (
	"os"
	"path/filepath"
	"strings"
)

const runtimeScenarioOutputEnv = "SPACE_ROCKS_RUNTIME_SCENARIO_OUTPUT"

func runtimeScenarioMeasurementPath(defaultPath string) string {
	outputRoot := strings.TrimSpace(os.Getenv(runtimeScenarioOutputEnv))
	if outputRoot == "" {
		return defaultPath
	}
	return filepath.Join(outputRoot, "measurements", "game-server")
}
