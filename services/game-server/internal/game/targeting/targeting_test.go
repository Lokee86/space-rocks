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

func TestEmptyTargetIsEmpty(t *testing.T) {
	target := EmptyTarget()

	if !target.IsEmpty() {
		t.Fatal("expected EmptyTarget() to be empty")
	}
}

func TestNonEmptyTargetRefIsNotEmpty(t *testing.T) {
	target := TargetRef{
		Kind: TargetKindPlayer,
		ID:   "player-1",
	}

	if target.IsEmpty() {
		t.Fatal("expected non-empty TargetRef to not be empty")
	}
}

func TestTargetKindPriorityOrder(t *testing.T) {
	playerPriority := TargetKindPriority(TargetKindPlayer)
	enemyPriority := TargetKindPriority(TargetKindEnemy)
	asteroidPriority := TargetKindPriority(TargetKindAsteroid)
	bulletPriority := TargetKindPriority(TargetKindBullet)

	if playerPriority <= enemyPriority {
		t.Fatalf("expected player priority (%d) > enemy priority (%d)", playerPriority, enemyPriority)
	}
	if enemyPriority <= asteroidPriority {
		t.Fatalf("expected enemy priority (%d) > asteroid priority (%d)", enemyPriority, asteroidPriority)
	}
	if asteroidPriority <= bulletPriority {
		t.Fatalf("expected asteroid priority (%d) > bullet priority (%d)", asteroidPriority, bulletPriority)
	}
}

func TestUnknownTargetKindPriorityBelowBullet(t *testing.T) {
	unknownPriority := TargetKindPriority(TargetKind("unknown"))
	bulletPriority := TargetKindPriority(TargetKindBullet)

	if unknownPriority >= bulletPriority {
		t.Fatalf("expected unknown kind priority (%d) < bullet priority (%d)", unknownPriority, bulletPriority)
	}
}
