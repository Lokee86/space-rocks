package game

import (
	"fmt"

	"github.com/Lokee86/space-rocks/services/game-server/internal/game/objectives"
)

func (target *Control) ForceObjectiveProgress(instanceID objectives.InstanceID, amount float64) error {
	target.game.mu.Lock()
	defer target.game.mu.Unlock()
	if target.game.lockedFinalMatchState != nil {
		return fmt.Errorf("match results are locked")
	}
	events, err := target.game.objectivesRuntime().AddProgress(instanceID, amount)
	target.game.publishObjectiveEventsLocked(events)
	return err
}

func (target *Control) SetObjectiveProgress(instanceID objectives.InstanceID, value float64) error {
	target.game.mu.Lock()
	defer target.game.mu.Unlock()
	if target.game.lockedFinalMatchState != nil {
		return fmt.Errorf("match results are locked")
	}
	events, err := target.game.objectivesRuntime().SetProgress(instanceID, value)
	target.game.publishObjectiveEventsLocked(events)
	return err
}

func (target *Control) ActivateObjective(instanceID objectives.InstanceID) error {
	return target.applyObjectiveAction(func(runtime *objectives.Runtime) ([]objectives.Event, error) {
		return runtime.Activate(instanceID)
	})
}

func (target *Control) DiscoverObjective(instanceID objectives.InstanceID) error {
	return target.applyObjectiveAction(func(runtime *objectives.Runtime) ([]objectives.Event, error) {
		return runtime.Discover(instanceID)
	})
}

func (target *Control) CompleteObjective(instanceID objectives.InstanceID) error {
	return target.applyObjectiveAction(func(runtime *objectives.Runtime) ([]objectives.Event, error) {
		return runtime.Complete(instanceID)
	})
}

func (target *Control) FailObjective(instanceID objectives.InstanceID, reason string) error {
	return target.applyObjectiveAction(func(runtime *objectives.Runtime) ([]objectives.Event, error) {
		return runtime.Fail(instanceID, reason)
	})
}

func (target *Control) CancelObjective(instanceID objectives.InstanceID) error {
	return target.applyObjectiveAction(func(runtime *objectives.Runtime) ([]objectives.Event, error) {
		return runtime.Cancel(instanceID)
	})
}

func (target *Control) RetireObjective(instanceID objectives.InstanceID, reason string) error {
	return target.applyObjectiveAction(func(runtime *objectives.Runtime) ([]objectives.Event, error) {
		return runtime.RetireInstance(instanceID, reason)
	})
}

func (target *Control) ResetObjective(instanceID objectives.InstanceID) error {
	return target.applyObjectiveAction(func(runtime *objectives.Runtime) ([]objectives.Event, error) {
		return runtime.Reset(instanceID)
	})
}

func (target *Control) applyObjectiveAction(
	action func(*objectives.Runtime) ([]objectives.Event, error),
) error {
	target.game.mu.Lock()
	defer target.game.mu.Unlock()
	if target.game.lockedFinalMatchState != nil {
		return fmt.Errorf("match results are locked")
	}
	events, err := action(target.game.objectivesRuntime())
	target.game.publishObjectiveEventsLocked(events)
	return err
}
