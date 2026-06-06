package damage

type DamageModifierCategory string

const (
	DamageModifierCategoryOutgoing     DamageModifierCategory = "outgoing"
	DamageModifierCategoryResistance   DamageModifierCategory = "resistance"
	DamageModifierCategoryVulnerability DamageModifierCategory = "vulnerability"
	DamageModifierCategoryGeneric      DamageModifierCategory = "generic"
)

type DamageModifierOperation string

const (
	DamageModifierOperationAdd      DamageModifierOperation = "add"
	DamageModifierOperationMultiply DamageModifierOperation = "multiply"
	DamageModifierOperationSet      DamageModifierOperation = "set"
)

type DamageModifier struct {
	Type      DamageType
	Category  DamageModifierCategory
	Operation DamageModifierOperation
	Value     float64
}

type AppliedDamageModifier struct {
	Modifier DamageModifier
	Value    float64
}

type ModifiedDamageAmount struct {
	BaseAmount       float64
	ModifiedAmount   int
	AppliedModifiers []AppliedDamageModifier
}

func normalizedDamageModifierCategory(modifier DamageModifier) DamageModifierCategory {
	if modifier.Category == "" {
		return DamageModifierCategoryGeneric
	}
	return modifier.Category
}

func damageModifierAppliesToType(modifier DamageModifier, damageType DamageType) bool {
	if modifier.Type == "" {
		return true
	}
	return modifier.Type == damageType
}

func isDamageModifierValid(modifier DamageModifier) bool {
	switch normalizedDamageModifierCategory(modifier) {
	case DamageModifierCategoryResistance:
		return modifier.Operation == DamageModifierOperationMultiply && modifier.Value >= 0 && modifier.Value < 1
	case DamageModifierCategoryVulnerability:
		return modifier.Operation == DamageModifierOperationMultiply && modifier.Value > 1
	case DamageModifierCategoryOutgoing, DamageModifierCategoryGeneric:
		switch modifier.Operation {
		case DamageModifierOperationAdd, DamageModifierOperationMultiply, DamageModifierOperationSet:
			return true
		default:
			return false
		}
	default:
		return false
	}
}

func FilterDamageModifiersByKind(modifiers []DamageModifier, kind DamageType) []DamageModifier {
	filtered := make([]DamageModifier, 0, len(modifiers))
	for _, modifier := range modifiers {
		if !damageModifierAppliesToType(modifier, kind) {
			continue
		}
		if !isDamageModifierValid(modifier) {
			continue
		}
		filtered = append(filtered, modifier)
	}
	return filtered
}

func ResolveModifiedAmount(baseAmount int, modifiers []DamageModifier, kind DamageType) ModifiedDamageAmount {
	result := ModifiedDamageAmount{
		BaseAmount:     float64(baseAmount),
		ModifiedAmount: baseAmount,
	}

	applicable := FilterDamageModifiersByKind(modifiers, kind)
	modified := float64(baseAmount)

	for _, modifier := range applicable {
		switch modifier.Operation {
		case DamageModifierOperationAdd:
			modified += modifier.Value
		}
	}

	for _, modifier := range applicable {
		switch modifier.Operation {
		case DamageModifierOperationMultiply:
			category := normalizedDamageModifierCategory(modifier)
			if category == DamageModifierCategoryOutgoing || category == DamageModifierCategoryGeneric {
				modified *= modifier.Value
			}
		}
	}

	for _, modifier := range applicable {
		switch modifier.Operation {
		case DamageModifierOperationMultiply:
			if normalizedDamageModifierCategory(modifier) == DamageModifierCategoryResistance {
				modified *= (1 - modifier.Value)
			}
		}
	}

	for _, modifier := range applicable {
		switch modifier.Operation {
		case DamageModifierOperationMultiply:
			if normalizedDamageModifierCategory(modifier) == DamageModifierCategoryVulnerability {
				modified *= modifier.Value
			}
		}
	}

	for _, modifier := range applicable {
		switch modifier.Operation {
		case DamageModifierOperationSet:
			modified = modifier.Value
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
