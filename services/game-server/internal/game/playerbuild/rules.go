package playerbuild

import (
	"fmt"

	"github.com/Lokee86/space-rocks/services/game-server/internal/game/modes"
)

func RulesForMatch(matchRules modes.ResolvedMatchRules) Rules {
	return NormalizeRules(Rules{ModeID: string(matchRules.ModeID)})
}

func NormalizeRules(rules Rules) Rules {
	if rules.ModuleActivationPolicy == "" {
		rules.ModuleActivationPolicy = ModulesAny
	}
	if rules.HardwiredPolicy == "" {
		rules.HardwiredPolicy = HardwiredAllowed
	}
	return rules
}

func ValidateRules(raw Rules) error {
	rules := NormalizeRules(raw)
	switch rules.ModuleActivationPolicy {
	case ModulesAny, ModulesPassiveOnly:
	default:
		return fmt.Errorf("unsupported module activation policy %q", rules.ModuleActivationPolicy)
	}
	switch rules.HardwiredPolicy {
	case HardwiredAllowed, HardwiredDisabled, HardwiredNormalized:
	default:
		return fmt.Errorf("unsupported hardwired policy %q", rules.HardwiredPolicy)
	}
	return nil
}
