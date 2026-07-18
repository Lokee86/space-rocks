package objectives

import "sort"

type Viewer struct {
	PlayerID      string
	TeamID        string
	IncludeHidden bool
}

type Snapshot struct {
	DefinitionID    DefinitionID
	InstanceID      InstanceID
	Scope           Scope
	OwnerID         string
	Associations    map[string]string
	Status          Status
	Discovered      bool
	Progress        float64
	Target          float64
	TimerRemaining  float64
	FailureReason   string
	SequenceIndex   int
	Members         []string
	LastAttribution Attribution
}

func (runtime *Runtime) Snapshot(id InstanceID, viewer Viewer) (Snapshot, bool) {
	objective, ok := runtime.instances[id]
	if !ok || !visibleTo(objective, viewer) {
		return Snapshot{}, false
	}
	return snapshotFor(objective), true
}

func (runtime *Runtime) Snapshots(viewer Viewer) []Snapshot {
	snapshots := make([]Snapshot, 0, len(runtime.instances))
	for _, id := range runtime.instanceIDs() {
		if snapshot, ok := runtime.Snapshot(id, viewer); ok {
			snapshots = append(snapshots, snapshot)
		}
	}
	return snapshots
}

func visibleTo(objective *instance, viewer Viewer) bool {
	if viewer.IncludeHidden {
		return true
	}
	visibility := objective.Definition.Lifecycle.Visibility
	if visibility == "" {
		visibility = VisibilityPublic
	}
	if (visibility == VisibilityHiddenUntilDiscovered || visibility == VisibilityOwnerHiddenUntilDiscovered) && !objective.Discovered {
		return false
	}
	if visibility != VisibilityOwnerOnly && visibility != VisibilityOwnerHiddenUntilDiscovered {
		return true
	}
	switch objective.Definition.Scope {
	case ScopePlayer:
		return viewer.PlayerID != "" && viewer.PlayerID == objective.OwnerID
	case ScopeTeam:
		return viewer.TeamID != "" && viewer.TeamID == objective.OwnerID
	default:
		return viewer.PlayerID != "" && viewer.PlayerID == objective.OwnerID
	}
}

func snapshotFor(objective *instance) Snapshot {
	members := make([]string, 0, len(objective.Success.Members))
	for member := range objective.Success.Members {
		members = append(members, member)
	}
	sort.Strings(members)
	return Snapshot{
		DefinitionID:    objective.Definition.ID,
		InstanceID:      objective.ID,
		Scope:           objective.Definition.Scope,
		OwnerID:         objective.OwnerID,
		Associations:    cloneAssociations(objective.Associations),
		Status:          objective.Status,
		Discovered:      objective.Discovered,
		Progress:        objective.Success.Progress,
		Target:          conditionTarget(objective.Definition.Success),
		TimerRemaining:  objective.TimerRemaining,
		FailureReason:   objective.FailureReason,
		SequenceIndex:   objective.Success.SequenceIndex,
		Members:         members,
		LastAttribution: objective.Success.Attribution,
	}
}
