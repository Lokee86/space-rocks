package main

import (
	"context"
	"testing"
)

func TestRuntimeScenarioAuthVerifierRequiresExplicitEnvironmentFlag(t *testing.T) {
	t.Setenv(runtimeScenarioAuthEnv, "")
	if verifier := runtimeScenarioAuthVerifierFromEnv(); verifier != nil {
		t.Fatal("expected scenario verifier to remain disabled")
	}
	t.Setenv(runtimeScenarioAuthEnv, "1")
	if verifier := runtimeScenarioAuthVerifierFromEnv(); verifier == nil {
		t.Fatal("expected scenario verifier")
	}
}

func TestRuntimeScenarioAuthVerifierAcceptsScenarioTokens(t *testing.T) {
	verifier := runtimeScenarioTokenVerifier{}
	result, err := verifier.VerifyToken(context.Background(), runtimeScenarioTokenPrefix+"participant-1")
	if err != nil {
		t.Fatalf("verify: %v", err)
	}
	if !result.Valid || result.Identity.AccountID == "" || result.Identity.DisplayName != "Scenario participant-1" {
		t.Fatalf("unexpected identity: %+v", result)
	}
}

func TestRuntimeScenarioAuthVerifierAcceptsBearerWrappedScenarioTokens(t *testing.T) {
	verifier := runtimeScenarioTokenVerifier{}
	result, err := verifier.VerifyToken(context.Background(), "Bearer "+runtimeScenarioTokenPrefix+"participant-2")
	if err != nil {
		t.Fatalf("verify: %v", err)
	}
	if !result.Valid || result.Identity.DisplayName != "Scenario participant-2" {
		t.Fatalf("unexpected identity: %+v", result)
	}
}

func TestRuntimeScenarioAuthVerifierRejectsOtherTokens(t *testing.T) {
	verifier := runtimeScenarioTokenVerifier{}
	result, err := verifier.VerifyToken(context.Background(), "ordinary-token")
	if err != nil {
		t.Fatalf("verify: %v", err)
	}
	if result.Valid {
		t.Fatal("expected ordinary token rejection")
	}
}
