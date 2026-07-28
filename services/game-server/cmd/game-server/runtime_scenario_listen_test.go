package main

import "testing"

func TestRuntimeScenarioListenAddressUsesConfiguredPort(t *testing.T) {
	t.Setenv(runtimeScenarioPortEnv, "19080")
	if got := runtimeScenarioListenAddress(":8080"); got != "127.0.0.1:19080" {
		t.Fatalf("listen address = %q", got)
	}
}

func TestRuntimeScenarioListenAddressKeepsDefaultForInvalidPort(t *testing.T) {
	t.Setenv(runtimeScenarioPortEnv, "invalid")
	if got := runtimeScenarioListenAddress(":8080"); got != ":8080" {
		t.Fatalf("listen address = %q", got)
	}
}
