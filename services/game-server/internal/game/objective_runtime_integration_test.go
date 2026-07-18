package game

import (
	"testing"

	"github.com/Lokee86/space-rocks/services/game-server/internal/game/awards"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/objectives"
)

func TestAwardCountersFeedMatchingPlayerObjective(t *testing.T) {
	game := New()
	playerID := game.AddPlayer()
	definition := objectives.Definition{
		ID:    "score-objective",
		Scope: objectives.ScopePlayer,
		Success: objectives.Condition{
			Kind:     objectives.ConditionNumeric,
			FactKey:  "counter:SCORE",
			Target:   25,
			Overflow: objectives.OverflowClamp,
		},
		Lifecycle: objectives.LifecycleDefinition{
			InitiallyActive: true,
			Visibility:      objectives.VisibilityOwnerOnly,
		},
	}
	if err := game.RegisterObjectiveDefinition(definition); err != nil {
		t.Fatal(err)
	}
	instanceID, err := game.CreateObjectiveInstance(definition.ID, objectives.Registration{OwnerID: playerID})
	if err != nil {
		t.Fatal(err)
	}

	var observed []objectives.Event
	game.AddObjectiveEventObserver(func(event objectives.Event) { observed = append(observed, event) })
	_, err = game.ApplyGameplayAwardEvent("objective-score", []awards.Mutation{{
		Owner:     awards.Owner{Scope: awards.ScopePlayer, ID: playerID},
		CounterID: awards.CounterScore,
		Operation: awards.MutationIncrement,
		Value:     25,
	}})
	if err != nil {
		t.Fatal(err)
	}

	snapshot, ok := game.ObjectiveSnapshot(instanceID, objectives.Viewer{PlayerID: playerID})
	if !ok || snapshot.Status != objectives.StatusCompleted || snapshot.Progress != 25 {
		t.Fatalf("snapshot = %#v, ok=%v", snapshot, ok)
	}
	if len(observed) != 2 || observed[0].Type != objectives.EventProgressChanged || observed[1].Type != objectives.EventCompleted {
		t.Fatalf("observed = %#v", observed)
	}
}

func TestObjectiveTimersIgnoreWorldFreezeButPauseWithoutConnectedPlayers(t *testing.T) {
	game := New()
	definition := timedObjectiveDefinition("timer")
	if err := game.RegisterObjectiveDefinition(definition); err != nil {
		t.Fatal(err)
	}
	instanceID, err := game.CreateObjectiveInstance(definition.ID, objectives.Registration{})
	if err != nil {
		t.Fatal(err)
	}

	game.Step(2)
	snapshot, _ := game.ObjectiveSnapshot(instanceID, objectives.Viewer{})
	if snapshot.TimerRemaining != 2 || snapshot.Status != objectives.StatusActive {
		t.Fatalf("timer advanced without connected players: %#v", snapshot)
	}

	game.AddPlayer()
	game.worldSimulationOptions.SetFreezeWorld(true)
	game.Step(2)
	snapshot, _ = game.ObjectiveSnapshot(instanceID, objectives.Viewer{})
	if snapshot.Status != objectives.StatusFailed || snapshot.FailureReason != "timer_expired" {
		t.Fatalf("world freeze incorrectly paused objective timer: %#v", snapshot)
	}
}

func TestPlayerDisconnectPreservesObjectiveState(t *testing.T) {
	game := New()
	playerID := game.AddPlayer()
	definition := objectives.Definition{
		ID:    "preserved",
		Scope: objectives.ScopePlayer,
		Success: objectives.Condition{
			Kind:    objectives.ConditionNumeric,
			FactKey: "progress",
			Target:  10,
		},
		Lifecycle: objectives.LifecycleDefinition{InitiallyActive: true, Visibility: objectives.VisibilityOwnerOnly},
	}
	if err := game.RegisterObjectiveDefinition(definition); err != nil {
		t.Fatal(err)
	}
	instanceID, err := game.CreateObjectiveInstance(definition.ID, objectives.Registration{OwnerID: playerID})
	if err != nil {
		t.Fatal(err)
	}
	if err := game.ApplyObjectiveFacts(instanceID, []objectives.Fact{{Key: "progress", Number: 4}}); err != nil {
		t.Fatal(err)
	}

	game.RemovePlayer(playerID)
	snapshot, ok := game.ObjectiveSnapshot(instanceID, objectives.Viewer{IncludeHidden: true})
	if !ok || snapshot.Progress != 4 || snapshot.Status != objectives.StatusActive {
		t.Fatalf("snapshot after disconnect = %#v, ok=%v", snapshot, ok)
	}
}

func TestObjectiveControlExercisesFoundationActions(t *testing.T) {
	game := New()
	definition := objectives.Definition{
		ID:    "devtools",
		Scope: objectives.ScopeMatch,
		Success: objectives.Condition{
			Kind:    objectives.ConditionNumeric,
			FactKey: "progress",
			Target:  5,
		},
		Lifecycle: objectives.LifecycleDefinition{
			Discoverable: true,
			Failable:     true,
			Visibility:   objectives.VisibilityHiddenUntilDiscovered,
		},
	}
	if err := game.RegisterObjectiveDefinition(definition); err != nil {
		t.Fatal(err)
	}
	instanceID, err := game.CreateObjectiveInstance(definition.ID, objectives.Registration{})
	if err != nil {
		t.Fatal(err)
	}
	control := NewControl(game)
	if err := control.DiscoverObjective(instanceID); err != nil {
		t.Fatal(err)
	}
	if err := control.ActivateObjective(instanceID); err != nil {
		t.Fatal(err)
	}
	if err := control.SetObjectiveProgress(instanceID, 5); err != nil {
		t.Fatal(err)
	}
	snapshot, _ := game.ObjectiveSnapshot(instanceID, objectives.Viewer{})
	if snapshot.Status != objectives.StatusCompleted {
		t.Fatalf("snapshot = %#v", snapshot)
	}
	if err := control.ResetObjective(instanceID); err != nil {
		t.Fatal(err)
	}
	snapshot, _ = game.ObjectiveSnapshot(instanceID, objectives.Viewer{IncludeHidden: true})
	if snapshot.Status != objectives.StatusUndiscovered || snapshot.Progress != 0 {
		t.Fatalf("reset snapshot = %#v", snapshot)
	}
}

func TestArcadeBaselineHasNoObjectives(t *testing.T) {
	game := New()
	if snapshots := game.ObjectiveSnapshots(objectives.Viewer{IncludeHidden: true}); len(snapshots) != 0 {
		t.Fatalf("baseline objectives = %#v", snapshots)
	}
}

func timedObjectiveDefinition(id objectives.DefinitionID) objectives.Definition {
	return objectives.Definition{
		ID:      id,
		Scope:   objectives.ScopeMatch,
		Success: objectives.Condition{Kind: objectives.ConditionManual},
		Timer:   &objectives.TimerDefinition{DurationSeconds: 2},
		Lifecycle: objectives.LifecycleDefinition{
			InitiallyActive: true,
			Failable:        true,
			Visibility:      objectives.VisibilityPublic,
		},
	}
}
