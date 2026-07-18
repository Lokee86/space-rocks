package damage

import "testing"

func testActiveDamageOverTime(policy DamageOverTimeStackingPolicy) ActiveDamageOverTime {
	return ActiveDamageOverTime{
		Source:          DamageSource{EntityID: "source-1"},
		Target:          DamageTargetRef{EntityID: "target-1", EntityType: EntityTypePlayer},
		AmountPerTick:   2,
		TickSeconds:     1,
		DurationSeconds: 3,
		Type:            DamageTypeThermal,
		StackKey:        "burn",
		StackingPolicy:  policy,
	}
}

func TestDamageOverTimeRuntimeStacksByDefaultAndTicksDeterministically(t *testing.T) {
	runtime := NewDamageOverTimeRuntime()
	first := testActiveDamageOverTime("")
	second := first
	second.Source.EntityID = "source-2"
	second.StackKey = "burn-2"
	runtime.Add(first)
	runtime.Add(second)
	if got := runtime.CountTarget("target-1"); got != 2 {
		t.Fatalf("stack count = %d, want 2", got)
	}
	ticks := runtime.Step(1, false)
	if len(ticks) != 2 || ticks[0].Effect.Source.EntityID != "source-1" || ticks[1].Effect.Source.EntityID != "source-2" {
		t.Fatalf("ticks = %+v", ticks)
	}
}

func TestDamageOverTimeRuntimeReplaceRefreshAndLimit(t *testing.T) {
	t.Run("replace", func(t *testing.T) {
		runtime := NewDamageOverTimeRuntime()
		runtime.Add(testActiveDamageOverTime(DamageOverTimeStack))
		outcome := runtime.Add(testActiveDamageOverTime(DamageOverTimeReplace))
		if !outcome.Replaced || runtime.CountTarget("target-1") != 1 {
			t.Fatalf("outcome = %+v, count = %d", outcome, runtime.CountTarget("target-1"))
		}
	})
	t.Run("refresh", func(t *testing.T) {
		runtime := NewDamageOverTimeRuntime()
		first := runtime.Add(testActiveDamageOverTime(DamageOverTimeStack))
		runtime.Step(0.5, false)
		outcome := runtime.Add(testActiveDamageOverTime(DamageOverTimeRefresh))
		if !outcome.Refreshed || outcome.EffectID != first.EffectID {
			t.Fatalf("outcome = %+v, first = %+v", outcome, first)
		}
		if ticks := runtime.Step(0.5, false); len(ticks) != 0 {
			t.Fatalf("refresh should reset tick timer, got %+v", ticks)
		}
	})
	t.Run("limit", func(t *testing.T) {
		runtime := NewDamageOverTimeRuntime()
		effect := testActiveDamageOverTime(DamageOverTimeLimit)
		effect.MaxStacks = 1
		runtime.Add(effect)
		outcome := runtime.Add(effect)
		if !outcome.Dropped || runtime.CountTarget("target-1") != 1 {
			t.Fatalf("outcome = %+v, count = %d", outcome, runtime.CountTarget("target-1"))
		}
	})
}

func TestDamageOverTimeRuntimePausesAndRemovesOnTargetCleanup(t *testing.T) {
	runtime := NewDamageOverTimeRuntime()
	runtime.Add(testActiveDamageOverTime(DamageOverTimeStack))
	if ticks := runtime.Step(2, true); len(ticks) != 0 {
		t.Fatalf("paused ticks = %+v", ticks)
	}
	if got := runtime.RemoveTarget("target-1"); got != 1 {
		t.Fatalf("removed = %d, want 1", got)
	}
	if runtime.CountTarget("target-1") != 0 {
		t.Fatal("expected target effects removed")
	}
}
