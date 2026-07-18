package objectives

import "testing"

func TestNumericObjectiveProgressCompletesAndClamps(t *testing.T) {
	runtime := NewRuntime()
	definition := numericDefinition("score-100", ScopePlayer, "counter:SCORE", 100)
	if err := runtime.RegisterDefinition(definition); err != nil {
		t.Fatal(err)
	}
	id, _, err := runtime.CreateInstance(definition.ID, Registration{OwnerID: "player-1"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := runtime.Activate(id); err != nil {
		t.Fatal(err)
	}
	events, err := runtime.ApplyFacts(id, []Fact{{Key: "counter:SCORE", Operation: FactIncrement, Number: 125}})
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 2 || events[0].Type != EventProgressChanged || events[1].Type != EventCompleted {
		t.Fatalf("events = %#v", events)
	}
	snapshot, ok := runtime.Snapshot(id, Viewer{PlayerID: "player-1"})
	if !ok || snapshot.Progress != 100 || snapshot.Status != StatusCompleted {
		t.Fatalf("snapshot = %#v, ok=%v", snapshot, ok)
	}
}

func TestSuccessWinsWhenFailureAlsoBecomesTrue(t *testing.T) {
	runtime := NewRuntime()
	failure := Condition{Kind: ConditionBoolean, FactKey: "failed", Expected: true}
	definition := Definition{
		ID:      "tie",
		Scope:   ScopeMatch,
		Success: Condition{Kind: ConditionBoolean, FactKey: "won", Expected: true},
		Failure: &failure,
		Lifecycle: LifecycleDefinition{
			InitiallyActive: true,
			Failable:        true,
			Visibility:      VisibilityPublic,
		},
	}
	if err := runtime.RegisterDefinition(definition); err != nil {
		t.Fatal(err)
	}
	id, _, err := runtime.CreateInstance(definition.ID, Registration{})
	if err != nil {
		t.Fatal(err)
	}
	events, err := runtime.ApplyFacts(id, []Fact{
		{Key: "won", Operation: FactSignal, Boolean: Bool(true)},
		{Key: "failed", Operation: FactSignal, Boolean: Bool(true)},
	})
	if err != nil {
		t.Fatal(err)
	}
	if events[len(events)-1].Type != EventCompleted {
		t.Fatalf("events = %#v", events)
	}
}

func TestTimerExpiryEmitsBeforeFailureAndPausesOnlyWhenRequested(t *testing.T) {
	runtime := NewRuntime()
	definition := Definition{
		ID:      "timed",
		Scope:   ScopeMatch,
		Success: Condition{Kind: ConditionManual},
		Timer:   &TimerDefinition{DurationSeconds: 2},
		Lifecycle: LifecycleDefinition{
			InitiallyActive: true,
			Failable:        true,
			Visibility:      VisibilityPublic,
		},
	}
	if err := runtime.RegisterDefinition(definition); err != nil {
		t.Fatal(err)
	}
	id, _, err := runtime.CreateInstance(definition.ID, Registration{})
	if err != nil {
		t.Fatal(err)
	}
	if events := runtime.Step(1, true); len(events) != 0 {
		t.Fatalf("paused events = %#v", events)
	}
	if snapshot, _ := runtime.Snapshot(id, Viewer{}); snapshot.TimerRemaining != 2 {
		t.Fatalf("timer after paused step = %v", snapshot.TimerRemaining)
	}
	events := runtime.Step(2, false)
	if len(events) != 2 || events[0].Type != EventTimerExpired || events[1].Type != EventFailed {
		t.Fatalf("events = %#v", events)
	}
	if events[1].FailureReason != "timer_expired" {
		t.Fatalf("failure reason = %q", events[1].FailureReason)
	}
}

func TestRetiringDefinitionRetiresActiveInstancesAndBlocksCreation(t *testing.T) {
	runtime := NewRuntime()
	definition := numericDefinition("retired", ScopeMatch, "progress", 1)
	definition.Lifecycle.InitiallyActive = true
	if err := runtime.RegisterDefinition(definition); err != nil {
		t.Fatal(err)
	}
	first, _, _ := runtime.CreateInstance(definition.ID, Registration{})
	second, _, _ := runtime.CreateInstance(definition.ID, Registration{})
	events, err := runtime.RetireDefinition(definition.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 2 || events[0].InstanceID != first || events[1].InstanceID != second {
		t.Fatalf("events = %#v", events)
	}
	if _, _, err := runtime.CreateInstance(definition.ID, Registration{}); err == nil {
		t.Fatal("expected retired definition creation to fail")
	}
}

func TestVisibilityHidesUndiscoveredAndNonOwnerSnapshots(t *testing.T) {
	runtime := NewRuntime()
	definition := numericDefinition("private", ScopePlayer, "progress", 1)
	definition.Lifecycle.Discoverable = true
	definition.Lifecycle.Visibility = VisibilityOwnerHiddenUntilDiscovered
	if err := runtime.RegisterDefinition(definition); err != nil {
		t.Fatal(err)
	}
	id, _, err := runtime.CreateInstance(definition.ID, Registration{OwnerID: "player-1"})
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := runtime.Snapshot(id, Viewer{PlayerID: "player-1"}); ok {
		t.Fatal("undiscovered objective leaked")
	}
	if _, err := runtime.Discover(id); err != nil {
		t.Fatal(err)
	}
	if _, ok := runtime.Snapshot(id, Viewer{PlayerID: "player-2"}); ok {
		t.Fatal("private objective leaked to another player")
	}
	if _, ok := runtime.Snapshot(id, Viewer{PlayerID: "player-1"}); !ok {
		t.Fatal("owner could not see discovered objective")
	}
}

func numericDefinition(id DefinitionID, scope Scope, key string, target float64) Definition {
	return Definition{
		ID:    id,
		Scope: scope,
		Success: Condition{
			Kind:     ConditionNumeric,
			FactKey:  key,
			Target:   target,
			Overflow: OverflowClamp,
		},
		Lifecycle: LifecycleDefinition{Visibility: VisibilityPublic},
	}
}
