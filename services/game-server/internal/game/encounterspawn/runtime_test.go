package encounterspawn

import "testing"

func TestRuntimeContinuousSchedulingPreservesRemainder(t *testing.T) {
	runtime := NewRuntime()
	config := validConfig()
	if err := runtime.Configure(config); err != nil {
		t.Fatal(err)
	}

	opportunities, err := runtime.Step(7, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(opportunities) != 2 {
		t.Fatalf("opportunities = %d, want 2", len(opportunities))
	}
	snapshot, _ := runtime.Snapshot(config.ID)
	if snapshot.ElapsedSeconds != 1 {
		t.Fatalf("elapsed remainder = %v, want 1", snapshot.ElapsedSeconds)
	}
}

func TestRuntimePauseDeactivateStopAndReset(t *testing.T) {
	runtime := NewRuntime()
	config := validConfig()
	if err := runtime.Configure(config); err != nil {
		t.Fatal(err)
	}
	if _, err := runtime.Step(1, true); err != nil {
		t.Fatal(err)
	}
	if snapshot, _ := runtime.Snapshot(config.ID); snapshot.ElapsedSeconds != 0 {
		t.Fatalf("paused elapsed = %v, want 0", snapshot.ElapsedSeconds)
	}
	if _, err := runtime.Step(1, false); err != nil {
		t.Fatal(err)
	}
	if err := runtime.Deactivate(config.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := runtime.Step(3, false); err != nil {
		t.Fatal(err)
	}
	if snapshot, _ := runtime.Snapshot(config.ID); snapshot.ElapsedSeconds != 1 {
		t.Fatalf("deactivated elapsed = %v, want 1", snapshot.ElapsedSeconds)
	}
	if err := runtime.Activate(config.ID); err != nil {
		t.Fatal(err)
	}
	runtime.Stop()
	if _, err := runtime.Step(3, false); err != nil {
		t.Fatal(err)
	}
	if snapshot, _ := runtime.Snapshot(config.ID); snapshot.ElapsedSeconds != 1 || !snapshot.RuntimeStopped {
		t.Fatalf("unexpected stopped snapshot: %+v", snapshot)
	}
	runtime.Resume()
	if err := runtime.ResetProgress(config.ID); err != nil {
		t.Fatal(err)
	}
	if snapshot, _ := runtime.Snapshot(config.ID); snapshot.ElapsedSeconds != 0 || snapshot.RuntimeStopped {
		t.Fatalf("unexpected reset snapshot: %+v", snapshot)
	}
}

func TestRuntimeOrdersDueProfilesByPriorityThenID(t *testing.T) {
	runtime := NewRuntime()
	configs := []Config{
		{ID: "z-low", ScheduleKind: ScheduleContinuous, IntervalSeconds: 1, BatchSize: 1, Priority: 1, InitiallyActive: true},
		{ID: "z-high", ScheduleKind: ScheduleContinuous, IntervalSeconds: 1, BatchSize: 1, Priority: 2, InitiallyActive: true},
		{ID: "a-high", ScheduleKind: ScheduleContinuous, IntervalSeconds: 1, BatchSize: 1, Priority: 2, InitiallyActive: true},
	}
	for _, config := range configs {
		if err := runtime.Configure(config); err != nil {
			t.Fatal(err)
		}
	}
	opportunities, err := runtime.Step(1, false)
	if err != nil {
		t.Fatal(err)
	}
	want := []ProfileID{"a-high", "z-high", "z-low"}
	for index, profileID := range want {
		if opportunities[index].ProfileID != profileID {
			t.Fatalf("opportunity %d = %q, want %q", index, opportunities[index].ProfileID, profileID)
		}
	}
}

func TestRuntimeSnapshotsDefensivelyCopyConfig(t *testing.T) {
	runtime := NewRuntime()
	config := validConfig()
	if err := runtime.Configure(config); err != nil {
		t.Fatal(err)
	}
	config.SpawnTypeWeightedLimits["asteroid"] = 1
	snapshot, _ := runtime.Snapshot(config.ID)
	if snapshot.Config.SpawnTypeWeightedLimits["asteroid"] != 60 {
		t.Fatal("configured limits were mutated by caller")
	}
	snapshot.Config.SpawnTypeWeightedLimits["asteroid"] = 2
	second, _ := runtime.Snapshot(config.ID)
	if second.Config.SpawnTypeWeightedLimits["asteroid"] != 60 {
		t.Fatal("snapshot mutation leaked into runtime")
	}
}
