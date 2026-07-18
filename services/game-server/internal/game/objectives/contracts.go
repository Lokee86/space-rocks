package objectives

import "fmt"

type DefinitionID string
type InstanceID string

type Scope string

const (
	ScopePlayer     Scope = "player"
	ScopeTeam       Scope = "team"
	ScopeMatch      Scope = "match"
	ScopeCollection Scope = "collection"
	ScopeCustom     Scope = "definition_specific"
)

type Status string

const (
	StatusUndiscovered Status = "undiscovered"
	StatusDiscovered   Status = "discovered"
	StatusInactive     Status = "inactive"
	StatusActive       Status = "active"
	StatusCompleted    Status = "completed"
	StatusFailed       Status = "failed"
	StatusCancelled    Status = "cancelled"
	StatusRetired      Status = "retired"
)

type VisibilityPolicy string

const (
	VisibilityPublic                     VisibilityPolicy = "public"
	VisibilityOwnerOnly                  VisibilityPolicy = "owner_only"
	VisibilityHiddenUntilDiscovered      VisibilityPolicy = "hidden_until_discovered"
	VisibilityOwnerHiddenUntilDiscovered VisibilityPolicy = "owner_hidden_until_discovered"
)

type ConditionKind string

const (
	ConditionManual   ConditionKind = "manual"
	ConditionBoolean  ConditionKind = "boolean"
	ConditionNumeric  ConditionKind = "numeric"
	ConditionSet      ConditionKind = "set"
	ConditionSequence ConditionKind = "sequence"
	ConditionMaintain ConditionKind = "maintain"
)

type OverflowPolicy string

const (
	OverflowClamp  OverflowPolicy = "clamp"
	OverflowRetain OverflowPolicy = "retain"
)

type FactOperation string

const (
	FactSignal       FactOperation = "signal"
	FactIncrement    FactOperation = "increment"
	FactSet          FactOperation = "set"
	FactReset        FactOperation = "reset"
	FactAddMember    FactOperation = "add_member"
	FactRemoveMember FactOperation = "remove_member"
)

type AttributionKind string

const (
	AttributionOneHit      AttributionKind = "one_hit"
	AttributionInGame      AttributionKind = "in_game"
	AttributionInEncounter AttributionKind = "in_encounter"
)

type Attribution struct {
	Kind     AttributionKind
	PlayerID string
	TeamID   string
}

type Fact struct {
	Key         string
	Operation   FactOperation
	Number      float64
	Boolean     *bool
	Member      string
	Attribution Attribution
}

func Bool(value bool) *bool { return &value }

type EventType string

const (
	EventCreated         EventType = "created"
	EventDiscovered      EventType = "discovered"
	EventActivated       EventType = "activated"
	EventProgressChanged EventType = "progress_changed"
	EventTimerExpired    EventType = "timer_expired"
	EventCompleted       EventType = "completed"
	EventFailed          EventType = "failed"
	EventCancelled       EventType = "cancelled"
	EventRetired         EventType = "retired"
	EventReset           EventType = "reset"
)

type Event struct {
	Type           EventType
	DefinitionID   DefinitionID
	InstanceID     InstanceID
	PreviousStatus Status
	Status         Status
	Progress       float64
	Target         float64
	FailureReason  string
	FactKey        string
}

func validScope(scope Scope) bool {
	switch scope {
	case ScopePlayer, ScopeTeam, ScopeMatch, ScopeCollection, ScopeCustom:
		return true
	default:
		return false
	}
}

func terminal(status Status) bool {
	return status == StatusCompleted || status == StatusFailed || status == StatusCancelled || status == StatusRetired
}

func requireID(label, value string) error {
	if value == "" {
		return fmt.Errorf("%s is required", label)
	}
	return nil
}
