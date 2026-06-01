package targeting

import "testing"

func TestValidateRequestedTargetEmptyRequestedClearsWhenRequesterExists(t *testing.T) {
	acceptedTarget, ok := ValidateRequestedTarget(
		"player-1",
		"",
		func(playerID string) bool {
			return playerID == "player-1"
		},
	)

	if !ok {
		t.Fatalf("expected empty requested target to be accepted")
	}
	if acceptedTarget != "" {
		t.Fatalf("expected accepted target to be empty, got %q", acceptedTarget)
	}
}

func TestValidateRequestedTargetNonEmptyExistingTargetSucceeds(t *testing.T) {
	acceptedTarget, ok := ValidateRequestedTarget(
		"player-1",
		"player-2",
		func(playerID string) bool {
			return playerID == "player-1" || playerID == "player-2"
		},
	)

	if !ok {
		t.Fatalf("expected existing target to be accepted")
	}
	if acceptedTarget != "player-2" {
		t.Fatalf("expected accepted target player-2, got %q", acceptedTarget)
	}
}

func TestValidateRequestedTargetMissingRequesterFails(t *testing.T) {
	acceptedTarget, ok := ValidateRequestedTarget(
		"player-1",
		"",
		func(playerID string) bool {
			return false
		},
	)

	if ok {
		t.Fatalf("expected missing requester to fail")
	}
	if acceptedTarget != "" {
		t.Fatalf("expected empty accepted target on failure, got %q", acceptedTarget)
	}
}

func TestValidateRequestedTargetMissingTargetFails(t *testing.T) {
	acceptedTarget, ok := ValidateRequestedTarget(
		"player-1",
		"player-2",
		func(playerID string) bool {
			return playerID == "player-1"
		},
	)

	if ok {
		t.Fatalf("expected missing requested target to fail")
	}
	if acceptedTarget != "" {
		t.Fatalf("expected empty accepted target on failure, got %q", acceptedTarget)
	}
}

func TestValidateRequestedTargetSelfTargetAllowed(t *testing.T) {
	acceptedTarget, ok := ValidateRequestedTarget(
		"player-1",
		"player-1",
		func(playerID string) bool {
			return playerID == "player-1"
		},
	)

	if !ok {
		t.Fatalf("expected self-target to be accepted")
	}
	if acceptedTarget != "player-1" {
		t.Fatalf("expected accepted self target player-1, got %q", acceptedTarget)
	}
}
