package objectives

import (
	"fmt"
	"math"
)

type Registration struct {
	InstanceID   InstanceID
	OwnerID      string
	Associations map[string]string
}

type Runtime struct {
	definitions map[DefinitionID]Definition
	retired     map[DefinitionID]bool
	instances   map[InstanceID]*instance
	nextID      int
}

type instance struct {
	ID             InstanceID
	Definition     Definition
	OwnerID        string
	Associations   map[string]string
	Status         Status
	Discovered     bool
	Success        conditionState
	Failure        conditionState
	TimerRemaining float64
	FailureReason  string
}

func NewRuntime() *Runtime {
	return &Runtime{
		definitions: make(map[DefinitionID]Definition),
		retired:     make(map[DefinitionID]bool),
		instances:   make(map[InstanceID]*instance),
	}
}

func (runtime *Runtime) RegisterDefinition(definition Definition) error {
	if err := definition.Validate(); err != nil {
		return err
	}
	if runtime.definitions == nil {
		*runtime = *NewRuntime()
	}
	if _, exists := runtime.definitions[definition.ID]; exists {
		return fmt.Errorf("objective definition %q already exists", definition.ID)
	}
	runtime.definitions[definition.ID] = definition.clone()
	return nil
}

func (runtime *Runtime) Definition(id DefinitionID) (Definition, bool) {
	definition, ok := runtime.definitions[id]
	return definition.clone(), ok
}

func (runtime *Runtime) CreateInstance(definitionID DefinitionID, registration Registration) (InstanceID, []Event, error) {
	definition, ok := runtime.definitions[definitionID]
	if !ok {
		return "", nil, fmt.Errorf("objective definition %q does not exist", definitionID)
	}
	if runtime.retired[definitionID] {
		return "", nil, fmt.Errorf("objective definition %q is retired", definitionID)
	}
	if (definition.Scope == ScopePlayer || definition.Scope == ScopeTeam) && registration.OwnerID == "" {
		return "", nil, fmt.Errorf("objective owner ID is required for scope %q", definition.Scope)
	}
	if err := validateAssociations(definition, registration.Associations); err != nil {
		return "", nil, err
	}
	instanceID := registration.InstanceID
	if instanceID == "" {
		runtime.nextID++
		instanceID = InstanceID(fmt.Sprintf("objective-%d", runtime.nextID))
	}
	if _, exists := runtime.instances[instanceID]; exists {
		return "", nil, fmt.Errorf("objective instance %q already exists", instanceID)
	}
	created := &instance{
		ID:           instanceID,
		Definition:   definition.clone(),
		OwnerID:      registration.OwnerID,
		Associations: cloneAssociations(registration.Associations),
		Success:      newConditionState(definition.Success),
	}
	if definition.Failure != nil {
		created.Failure = newConditionState(*definition.Failure)
	}
	initializeInstance(created)
	runtime.instances[instanceID] = created
	return instanceID, []Event{eventFor(created, EventCreated, "")}, nil
}

func (runtime *Runtime) ApplyFacts(instanceID InstanceID, facts []Fact) ([]Event, error) {
	objective, err := runtime.mutableInstance(instanceID)
	if err != nil {
		return nil, err
	}
	if objective.Status != StatusActive || len(facts) == 0 {
		return nil, nil
	}
	before := objective.Success.Progress
	changed, factKey := applyFacts(objective.Definition.Success, &objective.Success, facts)
	if objective.Definition.Failure != nil {
		failureChanged, failureKey := applyFacts(*objective.Definition.Failure, &objective.Failure, facts)
		changed = changed || failureChanged
		if factKey == "" {
			factKey = failureKey
		}
	}
	events := make([]Event, 0, 2)
	if changed || before != objective.Success.Progress {
		events = append(events, eventFor(objective, EventProgressChanged, factKey))
	}
	return append(events, runtime.resolveConditions(objective)...), nil
}

func (runtime *Runtime) ApplyFactsToScope(scope Scope, ownerID string, facts []Fact) ([]Event, error) {
	ids := runtime.instanceIDs()
	events := make([]Event, 0)
	for _, id := range ids {
		objective := runtime.instances[id]
		if objective.Definition.Scope != scope || objective.OwnerID != ownerID {
			continue
		}
		applied, err := runtime.ApplyFacts(id, facts)
		if err != nil {
			return events, err
		}
		events = append(events, applied...)
	}
	return events, nil
}

func (runtime *Runtime) Step(delta float64, simulationPaused bool) []Event {
	if simulationPaused || delta <= 0 || math.IsNaN(delta) || math.IsInf(delta, 0) {
		return nil
	}
	events := make([]Event, 0)
	for _, id := range runtime.instanceIDs() {
		objective := runtime.instances[id]
		if objective.Status != StatusActive {
			continue
		}
		progressChanged := advanceMaintain(objective.Definition.Success, &objective.Success, delta)
		if objective.Definition.Failure != nil {
			progressChanged = advanceMaintain(*objective.Definition.Failure, &objective.Failure, delta) || progressChanged
		}
		if progressChanged {
			events = append(events, eventFor(objective, EventProgressChanged, objective.Definition.Success.FactKey))
		}
		timerExpired := false
		if objective.Definition.Timer != nil {
			objective.TimerRemaining -= delta
			if objective.TimerRemaining <= 0 {
				objective.TimerRemaining = 0
				timerExpired = true
				events = append(events, eventFor(objective, EventTimerExpired, "timer"))
			}
		}
		resolved := runtime.resolveConditions(objective)
		if len(resolved) == 0 && timerExpired {
			resolved = runtime.resolveTimerExpiry(objective)
		}
		events = append(events, resolved...)
	}
	return events
}

func (runtime *Runtime) RetireDefinition(id DefinitionID) ([]Event, error) {
	if _, ok := runtime.definitions[id]; !ok {
		return nil, fmt.Errorf("objective definition %q does not exist", id)
	}
	if runtime.retired[id] {
		return nil, nil
	}
	runtime.retired[id] = true
	events := make([]Event, 0)
	for _, instanceID := range runtime.instanceIDs() {
		objective := runtime.instances[instanceID]
		if objective.Definition.ID != id || terminal(objective.Status) {
			continue
		}
		events = append(events, transition(objective, StatusRetired, EventRetired, "definition_retired"))
	}
	return events, nil
}

func (runtime *Runtime) IsDefinitionRetired(id DefinitionID) bool { return runtime.retired[id] }
