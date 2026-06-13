package main

import (
	"path/filepath"
	"testing"
)

func TestBuildPlayerDataRuntime(t *testing.T) {
	t.Setenv("PLAYER_DATA_RAILS_BASE_URL", "")
	t.Setenv("PLAYER_DATA_RAILS_INTERNAL_TOKEN", "")
	t.Setenv("PLAYER_DATA_SQLITE_PATH", "")

	runtime, err := buildPlayerDataRuntime()
	if err != nil {
		t.Fatalf("buildPlayerDataRuntime returned error: %v", err)
	}
	if runtime == nil {
		t.Fatal("buildPlayerDataRuntime returned nil runtime")
	}
}

func TestBuildPlayerDataRuntimeWithSQLitePath(t *testing.T) {
	tempDir := t.TempDir()
	tempPath := filepath.Join(tempDir, "player-data.sqlite")

	t.Setenv("PLAYER_DATA_RAILS_BASE_URL", "")
	t.Setenv("PLAYER_DATA_RAILS_INTERNAL_TOKEN", "")
	t.Setenv("PLAYER_DATA_SQLITE_PATH", tempPath)

	runtime, err := buildPlayerDataRuntime()
	if err != nil {
		t.Fatalf("buildPlayerDataRuntime returned error: %v", err)
	}
	if runtime == nil {
		t.Fatal("buildPlayerDataRuntime returned nil runtime")
	}
}
