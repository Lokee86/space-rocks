package playerbuild

import "testing"

func TestNormalizeRulesUsesBroadDefaults(t *testing.T) {
	rules := NormalizeRules(Rules{})
	if rules.ModuleActivationPolicy != ModulesAny {
		t.Fatalf("expected modules-any default, got %q", rules.ModuleActivationPolicy)
	}
	if rules.HardwiredPolicy != HardwiredAllowed {
		t.Fatalf("expected hardwired-allowed default, got %q", rules.HardwiredPolicy)
	}
}

func TestValidateRulesRejectsUnknownPolicy(t *testing.T) {
	if err := ValidateRules(Rules{HardwiredPolicy: "mystery"}); err == nil {
		t.Fatal("expected unknown hardwired policy rejection")
	}
	if err := ValidateRules(Rules{ModuleActivationPolicy: "mystery"}); err == nil {
		t.Fatal("expected unknown module activation policy rejection")
	}
}
