package awards

import "testing"

func TestComboAppliesCurrentTierThenAdvancesAndTimesOut(t *testing.T) {
	runtime := NewRuntime()
	owner := Owner{Scope: ScopePlayer, ID: "player-1"}
	first, err := runtime.ApplyCombo(owner, 1)
	if err != nil {
		t.Fatalf("first combo: %v", err)
	}
	if first.AppliedMultiplier != 1 || first.State.Multiplier != 1.25 || first.State.Tier != 1 {
		t.Fatalf("unexpected first combo: %+v", first)
	}
	second, _ := runtime.ApplyCombo(owner, 1.5)
	if second.AppliedMultiplier != 1.25 || second.State.Multiplier != 1.5 {
		t.Fatalf("unexpected second combo: %+v", second)
	}
	state := runtime.Combo(owner, 2.3)
	if state.Active || state.Multiplier != 1 {
		t.Fatalf("combo should time out: %+v", state)
	}
}

func TestDeathResetClearsComboAndAllStreaks(t *testing.T) {
	runtime := NewRuntime()
	owner := Owner{Scope: ScopePlayer, ID: "player-1"}
	_, _ = runtime.ApplyCombo(owner, 1)
	_, _ = runtime.IncrementStreak(owner, "pvp_kills")
	_, _ = runtime.IncrementStreak(owner, "survival")
	runtime.ResetProgress(owner)
	if state := runtime.Combo(owner, 1.1); state.Active {
		t.Fatalf("combo remained active: %+v", state)
	}
	if streaks := runtime.StreakSnapshot(); len(streaks) != 0 {
		t.Fatalf("streaks remained after reset: %+v", streaks)
	}
}
