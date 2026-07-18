package damage

import "testing"

func TestResolveModifiedAmountPreservesSignedValues(t *testing.T) {
	result := ResolveModifiedAmount(-10, []DamageModifier{
		{Operation: DamageModifierOperationAdd, Value: 2},
		{Operation: DamageModifierOperationMultiply, Value: 0.5},
	}, DamageTypeThermal)

	if result.ModifiedAmount != -4 {
		t.Fatalf("expected signed modified amount %d, got %d", -4, result.ModifiedAmount)
	}
}

func TestResolveSingleHealingCapsAtMaximumHealth(t *testing.T) {
	result := ResolveSingle(DamageResolutionRequest{
		Target: DamageTarget{
			EntityID:  "player-1",
			Health:    4,
			MaxHealth: 10,
		},
		Spec: DamageSpec{
			Amount:                 -10,
			Type:                   DamageTypeEnergy,
			RestorationDestination: RestorationDestinationHealth,
		},
	})

	if result.Kind != DamageResultKindHealing {
		t.Fatalf("expected healing result, got %q", result.Kind)
	}
	if result.RestoredToHealth != 6 || result.RemainingHealth != 10 {
		t.Fatalf("health restoration = %d, remaining health = %d; want 6 and 10", result.RestoredToHealth, result.RemainingHealth)
	}
	if result.RemainingShield != 0 {
		t.Fatalf("expected shield to remain 0, got %d", result.RemainingShield)
	}
}

func TestResolveSingleRepairCapsAtMaximumShield(t *testing.T) {
	result := ResolveSingle(DamageResolutionRequest{
		Target: DamageTarget{
			EntityID:  "player-1",
			Health:    10,
			Shield:    2,
			MaxShield: 5,
		},
		Spec: DamageSpec{
			Amount:                 -10,
			Type:                   DamageTypeEnergy,
			RestorationDestination: RestorationDestinationShield,
		},
	})

	if result.Kind != DamageResultKindRepair {
		t.Fatalf("expected repair result, got %q", result.Kind)
	}
	if result.RestoredToShield != 3 || result.RemainingShield != 5 {
		t.Fatalf("shield restoration = %d, remaining shield = %d; want 3 and 5", result.RestoredToShield, result.RemainingShield)
	}
	if result.RemainingHealth != 10 {
		t.Fatalf("expected health to remain 10, got %d", result.RemainingHealth)
	}
}

func TestResolveSingleBothRestorationUsesOneSignedAmount(t *testing.T) {
	result := ResolveSingle(DamageResolutionRequest{
		Target: DamageTarget{
			EntityID:  "player-1",
			Health:    8,
			MaxHealth: 10,
			Shield:    4,
			MaxShield: 7,
		},
		Spec: DamageSpec{
			Amount:                 -5,
			Type:                   DamageTypeEnergy,
			RestorationDestination: RestorationDestinationBoth,
		},
	})

	if result.Kind != DamageResultKindHealing {
		t.Fatalf("expected healing result, got %q", result.Kind)
	}
	if result.RestoredToHealth != 2 || result.RestoredToShield != 3 {
		t.Fatalf("restoration = health %d, shield %d; want 2 and 3", result.RestoredToHealth, result.RestoredToShield)
	}
	if result.RemainingHealth != 10 || result.RemainingShield != 7 {
		t.Fatalf("remaining = health %d, shield %d; want 10 and 7", result.RemainingHealth, result.RemainingShield)
	}
}

func TestResolveSingleShieldOverflowPassesThroughByDefault(t *testing.T) {
	result := ResolveSingle(DamageResolutionRequest{
		Target: DamageTarget{Health: 10, Shield: 2},
		Spec:   DamageSpec{Amount: 5, Type: DamageTypeThermal},
	})

	if result.Kind != DamageResultKindDamage {
		t.Fatalf("expected damage result, got %q", result.Kind)
	}
	if result.AbsorbedByShield != 2 || result.AppliedToHealth != 3 {
		t.Fatalf("damage = shield %d, health %d; want 2 and 3", result.AbsorbedByShield, result.AppliedToHealth)
	}
}

func TestResolveSingleShieldOverflowCanDiscardHealthOverflow(t *testing.T) {
	result := ResolveSingle(DamageResolutionRequest{
		Target: DamageTarget{Health: 10, Shield: 2},
		Spec: DamageSpec{
			Amount:         5,
			Type:           DamageTypeThermal,
			OverflowPolicy: ShieldOverflowDiscard,
		},
	})

	if result.Kind != DamageResultKindDamage {
		t.Fatalf("expected damage result, got %q", result.Kind)
	}
	if result.AbsorbedByShield != 2 || result.AppliedToHealth != 0 {
		t.Fatalf("damage = shield %d, health %d; want 2 and 0", result.AbsorbedByShield, result.AppliedToHealth)
	}
	if result.RemainingShield != 0 || result.RemainingHealth != 10 {
		t.Fatalf("remaining = shield %d, health %d; want 0 and 10", result.RemainingShield, result.RemainingHealth)
	}
}

func TestResolveSingleBypassShieldLeavesShieldUnchanged(t *testing.T) {
	result := ResolveSingle(DamageResolutionRequest{
		Target: DamageTarget{Health: 10, Shield: 5},
		Spec: DamageSpec{
			Amount:       3,
			Type:         DamageTypeThermal,
			BypassShield: true,
		},
	})

	if result.AbsorbedByShield != 0 || result.AppliedToHealth != 3 {
		t.Fatalf("damage = shield %d, health %d; want 0 and 3", result.AbsorbedByShield, result.AppliedToHealth)
	}
	if result.RemainingShield != 5 || result.RemainingHealth != 7 {
		t.Fatalf("remaining = shield %d, health %d; want 5 and 7", result.RemainingShield, result.RemainingHealth)
	}
}

func TestResolveSingleIneffectiveRestorationDoesNotMutate(t *testing.T) {
	result := ResolveSingle(DamageResolutionRequest{
		Target: DamageTarget{
			Health:    10,
			MaxHealth: 10,
			Shield:    5,
			MaxShield: 5,
		},
		Spec: DamageSpec{
			Amount:                 -5,
			Type:                   DamageTypeEnergy,
			RestorationDestination: RestorationDestinationBoth,
		},
	})

	if result.Kind != DamageResultKindIneffective || !result.Ignored {
		t.Fatalf("expected ineffective ignored result, got kind %q ignored %t", result.Kind, result.Ignored)
	}
	if result.RemainingHealth != 10 || result.RemainingShield != 5 {
		t.Fatalf("remaining = health %d, shield %d; want 10 and 5", result.RemainingHealth, result.RemainingShield)
	}
	if result.RestoredToHealth != 0 || result.RestoredToShield != 0 {
		t.Fatalf("expected no restoration, got health %d shield %d", result.RestoredToHealth, result.RestoredToShield)
	}
}

func TestResolveSingleAlreadyLethalTargetIsDiscarded(t *testing.T) {
	result := ResolveSingle(DamageResolutionRequest{
		Target: DamageTarget{Health: 0, Shield: 5, MaxHealth: 10, MaxShield: 5},
		Spec: DamageSpec{
			Amount:                 -5,
			Type:                   DamageTypeEnergy,
			RestorationDestination: RestorationDestinationBoth,
		},
	})

	if result.Kind != DamageResultKindDiscardedLethalTarget || !result.Discarded || !result.Ignored {
		t.Fatalf("expected discarded lethal-target result, got kind %q discarded %t ignored %t", result.Kind, result.Discarded, result.Ignored)
	}
	if result.RemainingHealth != 0 || result.RemainingShield != 5 {
		t.Fatalf("remaining = health %d, shield %d; want 0 and 5", result.RemainingHealth, result.RemainingShield)
	}
	if result.RestoredToHealth != 0 || result.RestoredToShield != 0 {
		t.Fatalf("expected no restoration, got health %d shield %d", result.RestoredToHealth, result.RestoredToShield)
	}
}
