package main

import (
	"context"
	"testing"
)

func TestBuildAuthVerifierPrefersRuntimeScenarioVerifier(t *testing.T) {
	t.Setenv(runtimeScenarioAuthEnv, "1")
	t.Setenv("API_SERVER_BASE_URL", "https://example.invalid")
	t.Setenv("GAME_SERVER_INTERNAL_TOKEN", "ordinary-internal-token")

	verifier := buildAuthVerifierFromEnv("trace-1")
	result, err := verifier.VerifyToken(context.Background(), runtimeScenarioTokenPrefix+"coordinator-1")
	if err != nil {
		t.Fatalf("verify: %v", err)
	}
	if !result.Valid || result.Identity.DisplayName != "Scenario coordinator-1" {
		t.Fatalf("unexpected identity: %+v", result)
	}
}
