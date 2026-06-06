package damage

import "testing"

func TestResolveSingleHealthOneDamageOneDestroysTarget(t *testing.T) {
	result := ResolveSingle(DamageResolutionRequest{
		Source: DamageSource{
			EntityID:   "bullet-1",
			EntityType: EntityTypeProjectile,
			Cause:      DamageCauseProjectile,
		},
		Target: DamageTarget{
			EntityID:   "asteroid-1",
			EntityType: EntityTypeAsteroid,
			Health:     1,
		},
		Spec: DamageSpec{
			Amount: 1,
			Type:   DamageTypeKinetic,
		},
	})

	if result.BaseAmount != 1 {
		t.Fatalf("expected base amount %d, got %d", 1, result.BaseAmount)
	}
	if result.SourceEntityID != "bullet-1" {
		t.Fatalf("expected source entity id %q, got %q", "bullet-1", result.SourceEntityID)
	}
	if result.SourceEntityType != EntityTypeProjectile {
		t.Fatalf("expected source entity type %q, got %q", EntityTypeProjectile, result.SourceEntityType)
	}
	if result.ModifiedAmount != 1 {
		t.Fatalf("expected modified amount %d, got %d", 1, result.ModifiedAmount)
	}
	if result.AppliedToHealth != 1 {
		t.Fatalf("expected applied health damage 1, got %d", result.AppliedToHealth)
	}
	if result.AbsorbedByShield != 0 {
		t.Fatalf("expected absorbed shield 0, got %d", result.AbsorbedByShield)
	}
	if result.RemainingHealth != 0 {
		t.Fatalf("expected remaining health 0, got %d", result.RemainingHealth)
	}
	if !result.Destroyed {
		t.Fatal("expected target to be destroyed")
	}
}

func TestResolveSinglePlayerDestroyedIsFatal(t *testing.T) {
	result := ResolveSingle(DamageResolutionRequest{
		Target: DamageTarget{
			EntityID:   "player-1",
			EntityType: EntityTypePlayer,
			Health:     1,
			Shield:     0,
		},
		Spec: DamageSpec{
			Amount: 1,
			Type:   DamageTypeThermal,
		},
	})

	if !result.Destroyed {
		t.Fatal("expected target to be destroyed")
	}
	if !result.Fatal {
		t.Fatal("expected destroyed player result to be fatal")
	}
}

func TestResolveSingleModifiedAmountUsesModifiers(t *testing.T) {
	result := ResolveSingle(DamageResolutionRequest{
		Target: DamageTarget{
			EntityID:   "asteroid-1",
			EntityType: EntityTypeAsteroid,
			Health:     10,
		},
		Spec: DamageSpec{
			Amount: 5,
			Type:   DamageTypeThermal,
		},
		Modifiers: []DamageModifier{
			{Type: DamageTypeThermal, Category: DamageModifierCategoryResistance, Operation: DamageModifierOperationMultiply, Value: 0.5},
		},
	})

	if result.ModifiedAmount != 3 {
		t.Fatalf("expected modified amount %d, got %d", 3, result.ModifiedAmount)
	}
}

func TestResolveSingleZeroDamageIgnored(t *testing.T) {
	result := ResolveSingle(DamageResolutionRequest{
		Target: DamageTarget{
			EntityID:   "asteroid-1",
			EntityType: EntityTypeAsteroid,
			Health:     10,
		},
		Spec: DamageSpec{
			Amount: 0,
			Type:   DamageTypeThermal,
		},
	})

	if !result.Ignored {
		t.Fatal("expected zero damage result to be ignored")
	}
}

func TestResolveSingleDeadTargetIgnored(t *testing.T) {
	result := ResolveSingle(DamageResolutionRequest{
		Target: DamageTarget{
			EntityID:   "asteroid-1",
			EntityType: EntityTypeAsteroid,
			Health:     0,
		},
		Spec: DamageSpec{
			Amount: 5,
			Type:   DamageTypeThermal,
		},
	})

	if !result.Ignored {
		t.Fatal("expected dead target result to be ignored")
	}
}

func TestResolveSingleFatalPlayer(t *testing.T) {
	result := ResolveSingle(DamageResolutionRequest{
		Target: DamageTarget{
			EntityID:   "player-1",
			EntityType: EntityTypePlayer,
			Health:     1,
			Shield:     0,
		},
		Spec: DamageSpec{
			Amount: 1,
			Type:   DamageTypeThermal,
		},
	})

	if !result.Destroyed {
		t.Fatal("expected target to be destroyed")
	}
	if !result.Fatal {
		t.Fatal("expected destroyed player result to be fatal")
	}
}

func TestResolveSingleNonfatalAsteroidDestruction(t *testing.T) {
	result := ResolveSingle(DamageResolutionRequest{
		Target: DamageTarget{
			EntityID:   "asteroid-1",
			EntityType: EntityTypeAsteroid,
			Health:     1,
			Shield:     0,
		},
		Spec: DamageSpec{
			Amount: 1,
			Type:   DamageTypeThermal,
		},
	})

	if !result.Destroyed {
		t.Fatal("expected target to be destroyed")
	}
	if result.Fatal {
		t.Fatal("expected destroyed asteroid result not to be fatal")
	}
}

func TestResolveSingleReportsAppliedResistanceModifier(t *testing.T) {
	result := ResolveSingle(DamageResolutionRequest{
		Target: DamageTarget{
			EntityID:   "asteroid-1",
			EntityType: EntityTypeAsteroid,
			Health:     10,
		},
		Spec: DamageSpec{
			Amount: 5,
			Type:   DamageTypeThermal,
		},
		Modifiers: []DamageModifier{
			{Type: DamageTypeThermal, Category: DamageModifierCategoryResistance, Operation: DamageModifierOperationMultiply, Value: 0.5},
		},
	})

	if len(result.AppliedModifiers) != 1 {
		t.Fatalf("expected 1 applied modifier, got %d", len(result.AppliedModifiers))
	}
	if result.AppliedModifiers[0].Modifier.Category != DamageModifierCategoryResistance {
		t.Fatalf("expected resistance modifier, got %q", result.AppliedModifiers[0].Modifier.Category)
	}
}

func TestResolveSingleReportsAppliedVulnerabilityModifier(t *testing.T) {
	result := ResolveSingle(DamageResolutionRequest{
		Target: DamageTarget{
			EntityID:   "asteroid-1",
			EntityType: EntityTypeAsteroid,
			Health:     10,
		},
		Spec: DamageSpec{
			Amount: 5,
			Type:   DamageTypeThermal,
		},
		Modifiers: []DamageModifier{
			{Type: DamageTypeThermal, Category: DamageModifierCategoryVulnerability, Operation: DamageModifierOperationMultiply, Value: 2},
		},
	})

	if len(result.AppliedModifiers) != 1 {
		t.Fatalf("expected 1 applied modifier, got %d", len(result.AppliedModifiers))
	}
	if result.AppliedModifiers[0].Modifier.Category != DamageModifierCategoryVulnerability {
		t.Fatalf("expected vulnerability modifier, got %q", result.AppliedModifiers[0].Modifier.Category)
	}
}
