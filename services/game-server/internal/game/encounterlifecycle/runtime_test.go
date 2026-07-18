package encounterlifecycle

import "testing"

func runtimeRegistration(profileID ProfileID, cost WeightedPopulationCost) Registration {
	return Registration{
		Origin: OriginMetadata{
			ProfileID:              profileID,
			SpawnType:              "ambush",
			LifecyclePolicyID:      "standard",
			Priority:               1,
			WeightedPopulationCost: cost,
		},
		Capabilities: EntityCapabilities{SupportsSoftRetire: true, SupportsHardRemove: true},
	}
}

func TestRuntimeRegisterLookupAndDuplicateRejection(t *testing.T) {
	runtime := NewRuntime()
	registration := runtimeRegistration("profile-a", 4)
	if err := runtime.Register("entity-b", registration); err != nil {
		t.Fatalf("register failed: %v", err)
	}
	if err := runtime.Register("entity-b", registration); err == nil {
		t.Fatal("expected duplicate registration to fail")
	}
	if err := runtime.Register("", registration); err == nil {
		t.Fatal("expected empty entity ID to fail")
	}

	entry, ok := runtime.Snapshot("entity-b")
	if !ok || entry.EntityID != "entity-b" || entry.Registration != registration {
		t.Fatalf("unexpected snapshot: %+v, found=%v", entry, ok)
	}
	entry.Registration.Origin.ProfileID = "mutated-copy"
	stored, _ := runtime.Snapshot("entity-b")
	if stored.Registration.Origin.ProfileID != "profile-a" {
		t.Fatal("snapshot exposed mutable registry state")
	}
}

func TestRuntimeRegisterValidatesRegistration(t *testing.T) {
	runtime := NewRuntime()
	if err := runtime.Register("entity-invalid", Registration{}); err == nil {
		t.Fatal("expected invalid registration to fail")
	}

	registration := runtimeRegistration("profile-a", 4)
	registration.Policy = Policy{
		LifetimeExpiry:  TriggerPolicy{Enabled: true, Disposition: DispositionHardRemove},
		LifetimeSeconds: 10,
	}
	registration.Capabilities = EntityCapabilities{SupportsSoftRetire: true}
	if err := runtime.Register("entity-incompatible", registration); err == nil {
		t.Fatal("expected policy/capability mismatch to fail")
	}
}

func TestRuntimeAdvanceHonorsPauseAndActiveLifetime(t *testing.T) {
	runtime := NewRuntime()
	if err := runtime.Register("entity-a", runtimeRegistration("profile-a", 2)); err != nil {
		t.Fatal(err)
	}
	if err := runtime.Advance(2.5, false); err != nil {
		t.Fatal(err)
	}
	if err := runtime.Advance(10, true); err != nil {
		t.Fatal(err)
	}
	entry, _ := runtime.Snapshot("entity-a")
	if entry.ElapsedLifetimeSeconds != 2.5 {
		t.Fatalf("paused advance changed lifetime to %v", entry.ElapsedLifetimeSeconds)
	}
	if err := runtime.Advance(-1, false); err == nil {
		t.Fatal("expected negative delta to fail")
	}
	if err := runtime.BeginRetirement("entity-a", EvaluationResult{Trigger: TriggerLifetimeExpiry, Disposition: DispositionSoftRetire}); err != nil {
		t.Fatal(err)
	}
	if err := runtime.Advance(5, false); err != nil {
		t.Fatal(err)
	}
	entry, _ = runtime.Snapshot("entity-a")
	if entry.ElapsedLifetimeSeconds != 2.5 {
		t.Fatalf("retiring entity lifetime advanced to %v", entry.ElapsedLifetimeSeconds)
	}
}

func TestRuntimeEntityIDsAreSorted(t *testing.T) {
	runtime := NewRuntime()
	for _, entityID := range []string{"entity-10", "entity-2", "entity-1"} {
		if err := runtime.Register(entityID, runtimeRegistration("profile-a", 1)); err != nil {
			t.Fatal(err)
		}
	}
	got := runtime.EntityIDs()
	want := []string{"entity-1", "entity-10", "entity-2"}
	for index := range want {
		if got[index] != want[index] {
			t.Fatalf("entity ID %d: got %q, want %q", index, got[index], want[index])
		}
	}
}

func TestRuntimeRetirementIsIdempotentAndRetainsResult(t *testing.T) {
	runtime := NewRuntime()
	if err := runtime.Register("entity-a", runtimeRegistration("profile-a", 3)); err != nil {
		t.Fatal(err)
	}
	result := EvaluationResult{Trigger: TriggerScriptedCleanup, Disposition: DispositionHardRemove}
	if err := runtime.BeginRetirement("entity-a", result); err != nil {
		t.Fatalf("begin retirement failed: %v", err)
	}
	entry, _ := runtime.Snapshot("entity-a")
	if entry.RetirementState != RetirementStateBegun || entry.Retirement != result {
		t.Fatalf("unexpected retirement snapshot: %+v", entry)
	}
	if err := runtime.BeginRetirement("entity-a", result); err == nil {
		t.Fatal("expected second retirement request to fail")
	}
}

func TestRuntimeRemoveUpdatesProfileAccountingAfterRemoval(t *testing.T) {
	runtime := NewRuntime()
	registrations := map[string]Registration{
		"entity-a": runtimeRegistration("profile-a", 3),
		"entity-b": runtimeRegistration("profile-a", 5),
		"entity-c": runtimeRegistration("profile-b", 2),
	}
	for entityID, registration := range registrations {
		if err := runtime.Register(entityID, registration); err != nil {
			t.Fatal(err)
		}
	}

	totals := runtime.ProfileWeightedPopulationTotals()
	if totals["profile-a"] != 8 || totals["profile-b"] != 2 {
		t.Fatalf("unexpected initial totals: %+v", totals)
	}
	totals["profile-a"] = 999
	if runtime.ProfileWeightedPopulationTotals()["profile-a"] != 8 {
		t.Fatal("profile totals exposed mutable internal state")
	}

	removed, ok := runtime.Remove("entity-a")
	if !ok || removed.EntityID != "entity-a" {
		t.Fatalf("unexpected removed entry: %+v, found=%v", removed, ok)
	}
	if _, stillRegistered := runtime.Snapshot("entity-a"); stillRegistered {
		t.Fatal("removed entry remained registered")
	}
	if runtime.ProfileWeightedPopulationTotals()["profile-a"] != 5 {
		t.Fatalf("accounting did not update after removal: %+v", runtime.ProfileWeightedPopulationTotals())
	}

	if _, ok := runtime.Remove("entity-b"); !ok {
		t.Fatal("expected second profile-a removal")
	}
	if _, exists := runtime.ProfileWeightedPopulationTotals()["profile-a"]; exists {
		t.Fatal("zero profile total was not removed")
	}
	if _, ok := runtime.Remove("missing"); ok {
		t.Fatal("missing removal unexpectedly succeeded")
	}
}
