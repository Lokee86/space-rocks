package damage

type DamageModifierCategory string

const (
	DamageModifierCategoryOutgoing     DamageModifierCategory = "outgoing"
	DamageModifierCategoryResistance    DamageModifierCategory = "resistance"
	DamageModifierCategoryVulnerability DamageModifierCategory = "vulnerability"
	DamageModifierCategoryGeneric       DamageModifierCategory = "generic"
)

type DamageModifierOperation string

const (
	DamageModifierOperationAdd      DamageModifierOperation = "add"
	DamageModifierOperationMultiply DamageModifierOperation = "multiply"
	DamageModifierOperationSet      DamageModifierOperation = "set"
)

type DamageModifier struct {
	Kind      DamageKind
	Category  DamageModifierCategory
	Operation DamageModifierOperation
	Value     int
}

type AppliedDamageModifier struct {
	Modifier DamageModifier
	Value    int
}

type ModifiedDamageAmount struct {
	BaseAmount       float64
	ModifiedAmount   int
	AppliedModifiers []AppliedDamageModifier
}

func FilterDamageModifiersByKind(modifiers []DamageModifier, kind DamageKind) []DamageModifier {
	filtered := make([]DamageModifier, 0, len(modifiers))
	for _, modifier := range modifiers {
		if modifier.Kind != "" && modifier.Kind != kind {
			continue
		}
		filtered = append(filtered, modifier)
	}
	return filtered
}

func ResolveModifiedAmount(baseAmount int, modifiers []DamageModifier, kind DamageKind) ModifiedDamageAmount {
	result := ModifiedDamageAmount{
		BaseAmount:     float64(baseAmount),
		ModifiedAmount: baseAmount,
	}

	applicable := FilterDamageModifiersByKind(modifiers, kind)
	modified := float64(baseAmount)

	for _, modifier := range applicable {
		switch modifier.Operation {
		case DamageModifierOperationAdd:
			modified += float64(modifier.Value)
		}
	}

	for _, modifier := range applicable {
		switch modifier.Operation {
		case DamageModifierOperationMultiply:
			modified *= float64(modifier.Value)
		}
	}

	for _, modifier := range applicable {
		switch modifier.Operation {
		case DamageModifierOperationSet:
			modified = float64(modifier.Value)
		}
	}

	if modified < 0 {
		modified = 0
	}

	result.ModifiedAmount = int(modified + 0.5)
	for _, modifier := range applicable {
		result.AppliedModifiers = append(result.AppliedModifiers, AppliedDamageModifier{
			Modifier: modifier,
			Value:    modifier.Value,
		})
	}

	return result
}
