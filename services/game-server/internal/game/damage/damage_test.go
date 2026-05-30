package damage

import "testing"

func TestResolveHealthOneDamageOneDestroysTarget(t *testing.T) {
	result := Resolve(DamageRequest{
		TargetEntityID:   "asteroid-1",
		TargetEntityType: EntityTypeAsteroid,
		SourceEntityID:   "bullet-1",
		SourceEntityType: EntityTypeProjectile,
		CurrentHealth:    1,
		Amount:           1,
	})

	if result.AppliedToHealth != 1 {
		t.Fatalf("expected applied health damage 1, got %d", result.AppliedToHealth)
	}
	if result.RemainingHealth != 0 {
		t.Fatalf("expected remaining health 0, got %d", result.RemainingHealth)
	}
	if !result.Destroyed {
		t.Fatal("expected target to be destroyed")
	}
}

func TestResolveHealthThreeDamageOneLeavesRemainingTwo(t *testing.T) {
	result := Resolve(DamageRequest{
		TargetEntityID:   "asteroid-1",
		TargetEntityType: EntityTypeAsteroid,
		SourceEntityID:   "bullet-1",
		SourceEntityType: EntityTypeProjectile,
		CurrentHealth:    3,
		Amount:           1,
	})

	if result.AppliedToHealth != 1 {
		t.Fatalf("expected applied health damage 1, got %d", result.AppliedToHealth)
	}
	if result.RemainingHealth != 2 {
		t.Fatalf("expected remaining health 2, got %d", result.RemainingHealth)
	}
	if result.Destroyed {
		t.Fatal("expected target not to be destroyed")
	}
}

func TestResolvePlayerDestroyedIsFatal(t *testing.T) {
	result := Resolve(DamageRequest{
		TargetEntityID:   "player-1",
		TargetEntityType: EntityTypePlayer,
		SourceEntityID:   "asteroid-1",
		SourceEntityType: EntityTypeAsteroid,
		CurrentHealth:    1,
		Amount:           1,
	})

	if !result.Destroyed {
		t.Fatal("expected target to be destroyed")
	}
	if !result.Fatal {
		t.Fatal("expected destroyed player result to be fatal")
	}
}

func TestResolveAsteroidDestroyedIsNotFatal(t *testing.T) {
	result := Resolve(DamageRequest{
		TargetEntityID:   "asteroid-1",
		TargetEntityType: EntityTypeAsteroid,
		SourceEntityID:   "bullet-1",
		SourceEntityType: EntityTypeProjectile,
		CurrentHealth:    1,
		Amount:           1,
	})

	if !result.Destroyed {
		t.Fatal("expected target to be destroyed")
	}
	if result.Fatal {
		t.Fatal("expected destroyed asteroid result not to be fatal")
	}
}

func TestResolveZeroDamageIsIgnored(t *testing.T) {
	result := Resolve(DamageRequest{
		TargetEntityID:   "asteroid-1",
		TargetEntityType: EntityTypeAsteroid,
		SourceEntityID:   "bullet-1",
		SourceEntityType: EntityTypeProjectile,
		CurrentHealth:    3,
		Amount:           0,
	})

	if !result.Ignored {
		t.Fatal("expected zero damage result to be ignored")
	}
}
