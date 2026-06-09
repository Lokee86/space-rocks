package main

import "testing"

func TestBuildPlayerDataRuntime(t *testing.T) {
	runtime, err := buildPlayerDataRuntime()
	if err != nil {
		t.Fatalf("buildPlayerDataRuntime returned error: %v", err)
	}
	if runtime == nil {
		t.Fatal("buildPlayerDataRuntime returned nil runtime")
	}
}
