package weapons

import "testing"

func TestStepSlotStateClampsCooldownAndPreservesAmmo(t *testing.T) {
	state := SlotState{
		CooldownRemaining: 1.5,
		AmmoRemaining:     7,
	}

	stepped := StepSlotState(state, 2)

	if stepped.CooldownRemaining != 0 {
		t.Fatalf("CooldownRemaining = %v, want 0", stepped.CooldownRemaining)
	}
	if stepped.AmmoRemaining != 7 {
		t.Fatalf("AmmoRemaining = %v, want 7", stepped.AmmoRemaining)
	}
}

