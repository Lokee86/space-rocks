package damage

import "testing"

func TestFilterDamageModifiersByKindEmptyKindApplies(t *testing.T) {
	modifiers := []DamageModifier{
		{Category: DamageModifierCategoryGeneric, Operation: DamageModifierOperationAdd, Value: 1},
	}

	filtered := FilterDamageModifiersByKind(modifiers, DamageTypeThermal)

	if len(filtered) != 1 {
		t.Fatalf("expected 1 modifier, got %d", len(filtered))
	}
	if filtered[0] != modifiers[0] {
		t.Fatal("expected empty kind modifier to be preserved")
	}
}

func TestFilterDamageModifiersByKindMatchingKindApplies(t *testing.T) {
	modifiers := []DamageModifier{
		{Type: DamageTypeThermal, Category: DamageModifierCategoryOutgoing, Operation: DamageModifierOperationMultiply, Value: 2},
	}

	filtered := FilterDamageModifiersByKind(modifiers, DamageTypeThermal)

	if len(filtered) != 1 {
		t.Fatalf("expected 1 modifier, got %d", len(filtered))
	}
	if filtered[0] != modifiers[0] {
		t.Fatal("expected matching kind modifier to be preserved")
	}
}

func TestFilterDamageModifiersByKindNonMatchingKindDoesNotApply(t *testing.T) {
	modifiers := []DamageModifier{
		{Type: DamageTypeRadioactive, Category: DamageModifierCategoryResistance, Operation: DamageModifierOperationMultiply, Value: 0.5},
	}

	filtered := FilterDamageModifiersByKind(modifiers, DamageTypeThermal)

	if len(filtered) != 0 {
		t.Fatalf("expected 0 modifiers, got %d", len(filtered))
	}
}

func TestFilterDamageModifiersByKindPreservesInputOrder(t *testing.T) {
	modifiers := []DamageModifier{
		{Category: DamageModifierCategoryGeneric, Operation: DamageModifierOperationAdd, Value: 1},
		{Type: DamageTypeThermal, Category: DamageModifierCategoryOutgoing, Operation: DamageModifierOperationMultiply, Value: 2},
		{Type: DamageTypeThermal, Category: DamageModifierCategoryGeneric, Operation: DamageModifierOperationSet, Value: 3},
		{Type: DamageTypeRadioactive, Category: DamageModifierCategoryResistance, Operation: DamageModifierOperationMultiply, Value: 0.5},
	}

	filtered := FilterDamageModifiersByKind(modifiers, DamageTypeThermal)

	if len(filtered) != 3 {
		t.Fatalf("expected 3 modifiers, got %d", len(filtered))
	}
	if filtered[0] != modifiers[0] || filtered[1] != modifiers[1] || filtered[2] != modifiers[2] {
		t.Fatal("expected filtered modifiers to preserve input order")
	}
}

func TestFilterDamageModifiersByKindSkipsInvalidResistanceModifiers(t *testing.T) {
	modifiers := []DamageModifier{
		{Type: DamageTypeThermal, Category: DamageModifierCategoryResistance, Operation: DamageModifierOperationMultiply, Value: 1.0},
		{Type: DamageTypeThermal, Category: DamageModifierCategoryResistance, Operation: DamageModifierOperationMultiply, Value: 1.25},
	}

	filtered := FilterDamageModifiersByKind(modifiers, DamageTypeThermal)

	if len(filtered) != 0 {
		t.Fatalf("expected 0 modifiers, got %d", len(filtered))
	}
}

func TestFilterDamageModifiersByKindSkipsInvalidVulnerabilityModifiers(t *testing.T) {
	modifiers := []DamageModifier{
		{Type: DamageTypeThermal, Category: DamageModifierCategoryVulnerability, Operation: DamageModifierOperationMultiply, Value: 1.0},
		{Type: DamageTypeThermal, Category: DamageModifierCategoryVulnerability, Operation: DamageModifierOperationMultiply, Value: 0.75},
	}

	filtered := FilterDamageModifiersByKind(modifiers, DamageTypeThermal)

	if len(filtered) != 0 {
		t.Fatalf("expected 0 modifiers, got %d", len(filtered))
	}
}

func TestFilterDamageModifiersByKindOutgoingValidation(t *testing.T) {
	modifiers := []DamageModifier{
		{Type: DamageTypeThermal, Category: "", Operation: DamageModifierOperationAdd, Value: 1},
		{Type: DamageTypeThermal, Category: DamageModifierCategoryGeneric, Operation: DamageModifierOperationMultiply, Value: 2},
		{Type: DamageTypeThermal, Category: DamageModifierCategoryOutgoing, Operation: DamageModifierOperationSet, Value: 3},
	}

	filtered := FilterDamageModifiersByKind(modifiers, DamageTypeThermal)

	if len(filtered) != len(modifiers) {
		t.Fatalf("expected %d modifiers, got %d", len(modifiers), len(filtered))
	}
}
