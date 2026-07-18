package objectives

func (runtime *Runtime) resolveConditions(objective *instance) []Event {
	if objective.Status != StatusActive {
		return nil
	}
	if conditionSatisfied(objective.Definition.Success, objective.Success) {
		return []Event{transition(objective, StatusCompleted, EventCompleted, "")}
	}
	if objective.Definition.Lifecycle.Failable && objective.Definition.Failure != nil &&
		conditionSatisfied(*objective.Definition.Failure, objective.Failure) {
		return []Event{transition(objective, StatusFailed, EventFailed, "condition_failed")}
	}
	return nil
}

func (runtime *Runtime) resolveTimerExpiry(objective *instance) []Event {
	timer := objective.Definition.Timer
	if timer == nil || objective.Status != StatusActive {
		return nil
	}
	status := timer.ExpiryStatus
	if status == "" {
		status = StatusFailed
	}
	reason := timer.FailureReason
	if status == StatusFailed && reason == "" {
		reason = "timer_expired"
	}
	switch status {
	case StatusCompleted:
		return []Event{transition(objective, status, EventCompleted, "")}
	case StatusCancelled:
		return []Event{transition(objective, status, EventCancelled, reason)}
	default:
		return []Event{transition(objective, StatusFailed, EventFailed, reason)}
	}
}

func transition(objective *instance, status Status, eventType EventType, reason string) Event {
	previous := objective.Status
	objective.Status = status
	objective.FailureReason = reason
	event := eventFor(objective, eventType, "")
	event.PreviousStatus = previous
	return event
}

func eventFor(objective *instance, eventType EventType, factKey string) Event {
	return Event{
		Type:          eventType,
		DefinitionID:  objective.Definition.ID,
		InstanceID:    objective.ID,
		Status:        objective.Status,
		Progress:      objective.Success.Progress,
		Target:        conditionTarget(objective.Definition.Success),
		FailureReason: objective.FailureReason,
		FactKey:       factKey,
	}
}
