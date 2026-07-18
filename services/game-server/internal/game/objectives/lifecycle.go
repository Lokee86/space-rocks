package objectives

import "fmt"

func (runtime *Runtime) Discover(id InstanceID) ([]Event, error) {
	objective, err := runtime.mutableInstance(id)
	if err != nil {
		return nil, err
	}
	if !objective.Definition.Lifecycle.Discoverable {
		return nil, fmt.Errorf("objective %q is not discoverable", id)
	}
	if objective.Status != StatusUndiscovered {
		return nil, nil
	}
	objective.Discovered = true
	return []Event{transition(objective, StatusDiscovered, EventDiscovered, "")}, nil
}

func (runtime *Runtime) Activate(id InstanceID) ([]Event, error) {
	objective, err := runtime.mutableInstance(id)
	if err != nil {
		return nil, err
	}
	if objective.Status == StatusActive {
		return nil, nil
	}
	if objective.Status != StatusInactive && objective.Status != StatusDiscovered {
		return nil, fmt.Errorf("objective %q cannot activate from %q", id, objective.Status)
	}
	return []Event{transition(objective, StatusActive, EventActivated, "")}, nil
}

func (runtime *Runtime) Complete(id InstanceID) ([]Event, error) {
	objective, err := runtime.mutableInstance(id)
	if err != nil {
		return nil, err
	}
	if objective.Status == StatusCompleted {
		return nil, nil
	}
	if terminal(objective.Status) {
		return nil, fmt.Errorf("objective %q cannot complete from %q", id, objective.Status)
	}
	return []Event{transition(objective, StatusCompleted, EventCompleted, "")}, nil
}

func (runtime *Runtime) Fail(id InstanceID, reason string) ([]Event, error) {
	objective, err := runtime.mutableInstance(id)
	if err != nil {
		return nil, err
	}
	if !objective.Definition.Lifecycle.Failable {
		return nil, fmt.Errorf("objective %q is not failable", id)
	}
	if reason == "" {
		return nil, fmt.Errorf("objective failure reason is required")
	}
	if objective.Status == StatusFailed {
		return nil, nil
	}
	if terminal(objective.Status) {
		return nil, fmt.Errorf("objective %q cannot fail from %q", id, objective.Status)
	}
	return []Event{transition(objective, StatusFailed, EventFailed, reason)}, nil
}

func (runtime *Runtime) Cancel(id InstanceID) ([]Event, error) {
	objective, err := runtime.mutableInstance(id)
	if err != nil {
		return nil, err
	}
	if objective.Status == StatusCancelled {
		return nil, nil
	}
	if terminal(objective.Status) {
		return nil, fmt.Errorf("objective %q cannot cancel from %q", id, objective.Status)
	}
	return []Event{transition(objective, StatusCancelled, EventCancelled, "")}, nil
}

func (runtime *Runtime) RetireInstance(id InstanceID, reason string) ([]Event, error) {
	objective, err := runtime.mutableInstance(id)
	if err != nil {
		return nil, err
	}
	if objective.Status == StatusRetired {
		return nil, nil
	}
	if terminal(objective.Status) {
		return nil, fmt.Errorf("objective %q cannot retire from %q", id, objective.Status)
	}
	return []Event{transition(objective, StatusRetired, EventRetired, reason)}, nil
}

func (runtime *Runtime) Reset(id InstanceID) ([]Event, error) {
	objective, err := runtime.mutableInstance(id)
	if err != nil {
		return nil, err
	}
	if objective.Status == StatusRetired || runtime.retired[objective.Definition.ID] {
		return nil, fmt.Errorf("retired objective %q cannot reset", id)
	}
	previous := objective.Status
	objective.Success = newConditionState(objective.Definition.Success)
	objective.Failure = conditionState{}
	if objective.Definition.Failure != nil {
		objective.Failure = newConditionState(*objective.Definition.Failure)
	}
	objective.FailureReason = ""
	initializeInstance(objective)
	event := eventFor(objective, EventReset, "")
	event.PreviousStatus = previous
	return []Event{event}, nil
}

func (runtime *Runtime) AddProgress(id InstanceID, amount float64) ([]Event, error) {
	objective, err := runtime.mutableInstance(id)
	if err != nil {
		return nil, err
	}
	condition := objective.Definition.Success
	if condition.Kind != ConditionNumeric && condition.Kind != ConditionMaintain {
		return nil, fmt.Errorf("objective %q does not use numeric progress", id)
	}
	return runtime.ApplyFacts(id, []Fact{{Key: condition.FactKey, Operation: FactIncrement, Number: amount}})
}

func (runtime *Runtime) SetProgress(id InstanceID, value float64) ([]Event, error) {
	objective, err := runtime.mutableInstance(id)
	if err != nil {
		return nil, err
	}
	if objective.Status != StatusActive {
		return nil, fmt.Errorf("objective %q is not active", id)
	}
	condition := objective.Definition.Success
	if condition.Kind != ConditionNumeric && condition.Kind != ConditionMaintain {
		return nil, fmt.Errorf("objective %q does not use numeric progress", id)
	}
	before := objective.Success.Progress
	objective.Success.Progress = normalizedProgress(condition, value)
	if before == objective.Success.Progress {
		return runtime.resolveConditions(objective), nil
	}
	events := []Event{eventFor(objective, EventProgressChanged, condition.FactKey)}
	return append(events, runtime.resolveConditions(objective)...), nil
}
