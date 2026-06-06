package damage

import "testing"

func TestResolveModifiedAmountNoModifiers(t *testing.T) {
	result := ResolveModifiedAmount(5, nil, DamageTypeThermal)

	if result.BaseAmount != 5 {
		t.Fatalf("expected base amount %v, got %v", 5, result.BaseAmount)
	}
	if result.ModifiedAmount != 5 {
		t.Fatalf("expected modified amount %d, got %d", 5, result.ModifiedAmount)
	}
	if len(result.AppliedModifiers) != 0 {
		t.Fatalf("expected no applied modifiers, got %d", len(result.AppliedModifiers))
	}
}

func TestResolveModifiedAmountAdd(t *testing.T) {
	result := ResolveModifiedAmount(5, []DamageModifier{
		{Operation: DamageModifierOperationAdd, Value: 2},
	}, DamageTypeThermal)

	if result.ModifiedAmount != 7 {
		t.Fatalf("expected modified amount %d, got %d", 7, result.ModifiedAmount)
	}
}

func TestResolveModifiedAmountMultiply(t *testing.T) {
	result := ResolveModifiedAmount(5, []DamageModifier{
		{Operation: DamageModifierOperationMultiply, Value: 2},
	}, DamageTypeThermal)

	if result.ModifiedAmount != 10 {
		t.Fatalf("expected modified amount %d, got %d", 10, result.ModifiedAmount)
	}
}

func TestResolveModifiedAmountthermalResistance025(t *testing.T) {
	result := ResolveModifiedAmount(100, []DamageModifier{
		{Type: DamageTypeThermal, Category: DamageModifierCategoryResistance, Operation: DamageModifierOperationMultiply, Value: 0.25},
	}, DamageTypeThermal)

	if result.ModifiedAmount != 75 {
		t.Fatalf("expected modified amount %d, got %d", 75, result.ModifiedAmount)
	}
}

func TestResolveModifiedAmountthermalResistance025And020(t *testing.T) {
	result := ResolveModifiedAmount(100, []DamageModifier{
		{Type: DamageTypeThermal, Category: DamageModifierCategoryResistance, Operation: DamageModifierOperationMultiply, Value: 0.25},
		{Type: DamageTypeThermal, Category: DamageModifierCategoryResistance, Operation: DamageModifierOperationMultiply, Value: 0.20},
	}, DamageTypeThermal)

	if result.ModifiedAmount != 60 {
		t.Fatalf("expected modified amount %d, got %d", 60, result.ModifiedAmount)
	}
}

func TestResolveModifiedAmountthermalVulnerability125(t *testing.T) {
	result := ResolveModifiedAmount(100, []DamageModifier{
		{Type: DamageTypeThermal, Category: DamageModifierCategoryVulnerability, Operation: DamageModifierOperationMultiply, Value: 1.25},
	}, DamageTypeThermal)

	if result.ModifiedAmount != 125 {
		t.Fatalf("expected modified amount %d, got %d", 125, result.ModifiedAmount)
	}
}

func TestResolveModifiedAmountthermalResistanceDoesNotAffectradioactive(t *testing.T) {
	result := ResolveModifiedAmount(100, []DamageModifier{
		{Type: DamageTypeThermal, Category: DamageModifierCategoryResistance, Operation: DamageModifierOperationMultiply, Value: 0.25},
	}, DamageTypeRadioactive)

	if result.ModifiedAmount != 100 {
		t.Fatalf("expected modified amount %d, got %d", 100, result.ModifiedAmount)
	}
}

func TestResolveModifiedAmountthermalVulnerabilityDoesNotAffectradioactive(t *testing.T) {
	result := ResolveModifiedAmount(100, []DamageModifier{
		{Type: DamageTypeThermal, Category: DamageModifierCategoryVulnerability, Operation: DamageModifierOperationMultiply, Value: 1.25},
	}, DamageTypeRadioactive)

	if result.ModifiedAmount != 100 {
		t.Fatalf("expected modified amount %d, got %d", 100, result.ModifiedAmount)
	}
}

func TestResolveModifiedAmountEmptyKindResistanceAppliesGlobally(t *testing.T) {
	result := ResolveModifiedAmount(100, []DamageModifier{
		{Category: DamageModifierCategoryResistance, Operation: DamageModifierOperationMultiply, Value: 0.25},
	}, DamageTypeRadioactive)

	if result.ModifiedAmount != 75 {
		t.Fatalf("expected modified amount %d, got %d", 75, result.ModifiedAmount)
	}
}

func TestResolveModifiedAmountRadioactiveIgnoresThermalResistance(t *testing.T) {
	result := ResolveModifiedAmount(100, []DamageModifier{
		{Type: DamageTypeThermal, Category: DamageModifierCategoryResistance, Operation: DamageModifierOperationMultiply, Value: 0.25},
	}, DamageTypeRadioactive)

	if result.ModifiedAmount != 100 {
		t.Fatalf("expected modified amount %d, got %d", 100, result.ModifiedAmount)
	}
}

func TestResolveModifiedAmountEmptyTypeResistanceAppliesGlobally(t *testing.T) {
	result := ResolveModifiedAmount(100, []DamageModifier{
		{Category: DamageModifierCategoryResistance, Operation: DamageModifierOperationMultiply, Value: 0.25},
	}, DamageTypeThermal)

	if result.ModifiedAmount != 75 {
		t.Fatalf("expected modified amount %d, got %d", 75, result.ModifiedAmount)
	}
}

func TestResolveModifiedAmountMultiplyHalf(t *testing.T) {
	result := ResolveModifiedAmount(5, []DamageModifier{
		{Operation: DamageModifierOperationMultiply, Value: 0.5},
	}, DamageTypeThermal)

	if result.ModifiedAmount != 3 {
		t.Fatalf("expected modified amount %d, got %d", 3, result.ModifiedAmount)
	}
}

func TestResolveModifiedAmountMultiplyOnePointFive(t *testing.T) {
	result := ResolveModifiedAmount(5, []DamageModifier{
		{Operation: DamageModifierOperationMultiply, Value: 1.5},
	}, DamageTypeThermal)

	if result.ModifiedAmount != 8 {
		t.Fatalf("expected modified amount %d, got %d", 8, result.ModifiedAmount)
	}
}

func TestResolveModifiedAmountAddBeforeMultiply(t *testing.T) {
	result := ResolveModifiedAmount(5, []DamageModifier{
		{Operation: DamageModifierOperationAdd, Value: 3},
		{Operation: DamageModifierOperationMultiply, Value: 2},
	}, DamageTypeThermal)

	if result.ModifiedAmount != 16 {
		t.Fatalf("expected modified amount %d, got %d", 16, result.ModifiedAmount)
	}
}

func TestResolveModifiedAmountSetLast(t *testing.T) {
	result := ResolveModifiedAmount(5, []DamageModifier{
		{Operation: DamageModifierOperationAdd, Value: 3},
		{Operation: DamageModifierOperationMultiply, Value: 2},
		{Operation: DamageModifierOperationSet, Value: 4},
	}, DamageTypeThermal)

	if result.ModifiedAmount != 4 {
		t.Fatalf("expected modified amount %d, got %d", 4, result.ModifiedAmount)
	}
}

func TestResolveModifiedAmountWrongKindIgnored(t *testing.T) {
	result := ResolveModifiedAmount(5, []DamageModifier{
		{Type: DamageTypeRadioactive, Operation: DamageModifierOperationAdd, Value: 3},
	}, DamageTypeThermal)

	if result.ModifiedAmount != 5 {
		t.Fatalf("expected modified amount %d, got %d", 5, result.ModifiedAmount)
	}
}

func TestResolveModifiedAmountZeroMultiplier(t *testing.T) {
	result := ResolveModifiedAmount(5, []DamageModifier{
		{Operation: DamageModifierOperationMultiply, Value: 0},
	}, DamageTypeThermal)

	if result.ModifiedAmount != 0 {
		t.Fatalf("expected modified amount %d, got %d", 0, result.ModifiedAmount)
	}
}

func TestResolveModifiedAmountNegativeClamp(t *testing.T) {
	result := ResolveModifiedAmount(5, []DamageModifier{
		{Operation: DamageModifierOperationAdd, Value: -10},
	}, DamageTypeThermal)

	if result.ModifiedAmount != 0 {
		t.Fatalf("expected modified amount %d, got %d", 0, result.ModifiedAmount)
	}
}
