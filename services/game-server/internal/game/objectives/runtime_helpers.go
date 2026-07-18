package objectives

import (
	"fmt"
	"sort"
)

func (runtime *Runtime) mutableInstance(id InstanceID) (*instance, error) {
	objective, ok := runtime.instances[id]
	if !ok {
		return nil, fmt.Errorf("objective instance %q does not exist", id)
	}
	return objective, nil
}

func (runtime *Runtime) instanceIDs() []InstanceID {
	ids := make([]InstanceID, 0, len(runtime.instances))
	for id := range runtime.instances {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(left, right int) bool { return ids[left] < ids[right] })
	return ids
}

func initializeInstance(objective *instance) {
	definition := objective.Definition
	objective.Discovered = !definition.Lifecycle.Discoverable ||
		definition.Lifecycle.InitiallyDiscovered || definition.Lifecycle.InitiallyActive
	switch {
	case definition.Lifecycle.InitiallyActive:
		objective.Status = StatusActive
	case definition.Lifecycle.Discoverable && !objective.Discovered:
		objective.Status = StatusUndiscovered
	case definition.Lifecycle.Discoverable:
		objective.Status = StatusDiscovered
	default:
		objective.Status = StatusInactive
	}
	if definition.Timer != nil {
		objective.TimerRemaining = definition.Timer.DurationSeconds
	}
}

func validateAssociations(definition Definition, associations map[string]string) error {
	for _, key := range definition.AssociationKeys {
		if associations[key] == "" {
			return fmt.Errorf("objective association %q is required", key)
		}
	}
	return nil
}

func cloneAssociations(source map[string]string) map[string]string {
	clone := make(map[string]string, len(source))
	for key, value := range source {
		clone[key] = value
	}
	return clone
}
