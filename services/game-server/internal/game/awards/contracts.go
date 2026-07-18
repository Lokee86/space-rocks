package awards

import (
	"fmt"
	"regexp"
)

type CounterID string

const (
	CounterScore              CounterID = "SCORE"
	CounterKills              CounterID = "KILLS"
	CounterAssists            CounterID = "ASSISTS"
	CounterDeaths             CounterID = "DEATHS"
	CounterDamageDealt        CounterID = "DAMAGE_DEALT"
	CounterDamageTaken        CounterID = "DAMAGE_TAKEN"
	CounterObjectiveProgress  CounterID = "OBJECTIVE_PROGRESS"
	CounterResourcesCollected CounterID = "RESOURCES_COLLECTED"
	CounterCompletionTime     CounterID = "COMPLETION_TIME"
)

var fixedCounters = map[CounterID]struct{}{
	CounterScore: {}, CounterKills: {}, CounterAssists: {}, CounterDeaths: {},
	CounterDamageDealt: {}, CounterDamageTaken: {}, CounterObjectiveProgress: {},
	CounterResourcesCollected: {}, CounterCompletionTime: {},
}

var customCounterPattern = regexp.MustCompile(`^[A-Z][A-Z0-9_]{0,63}$`)

func IsFixedCounter(id CounterID) bool {
	_, ok := fixedCounters[id]
	return ok
}

func ValidateCustomCounterID(id CounterID) error {
	if IsFixedCounter(id) {
		return fmt.Errorf("counter %q is fixed", id)
	}
	if !customCounterPattern.MatchString(string(id)) {
		return fmt.Errorf("invalid custom counter id %q", id)
	}
	return nil
}

type Scope string

const (
	ScopePlayer    Scope = "player"
	ScopeTeam      Scope = "team"
	ScopeMatch     Scope = "match"
	ScopeObjective Scope = "objective"
)

type Owner struct {
	Scope Scope
	ID    string
}

func (owner Owner) Validate() error {
	switch owner.Scope {
	case ScopePlayer, ScopeTeam, ScopeMatch, ScopeObjective:
	default:
		return fmt.Errorf("invalid counter scope %q", owner.Scope)
	}
	if owner.ID == "" {
		return fmt.Errorf("counter owner id is required")
	}
	return nil
}

type Visibility string

const (
	VisibilityHidden           Visibility = "hidden"
	VisibilityHUD              Visibility = "hud"
	VisibilityScoreboard       Visibility = "scoreboard"
	VisibilityResultsOnly      Visibility = "results_only"
	VisibilityTeamOnly         Visibility = "team_only"
	VisibilityPlayerPrivate    Visibility = "player_private"
	VisibilitySpectatorVisible Visibility = "spectator_visible"
)

type MutationOperation string

const (
	MutationIncrement       MutationOperation = "increment"
	MutationDecrement       MutationOperation = "decrement"
	MutationSet             MutationOperation = "set"
	MutationMinimum         MutationOperation = "minimum"
	MutationMaximum         MutationOperation = "maximum"
	MutationTimedAccumulate MutationOperation = "timed_accumulation"
)

type Mutation struct {
	Owner     Owner
	CounterID CounterID
	Operation MutationOperation
	Value     float64
	Source    string
}

type Change struct {
	Owner     Owner
	CounterID CounterID
	Operation MutationOperation
	Before    float64
	After     float64
	Delta     float64
	Source    string
}

type EventResult struct {
	EventID   string
	Applied   bool
	Duplicate bool
	Changes   []Change
}

type CounterDefinition struct {
	ID         CounterID
	Visibility Visibility
}

type CounterSnapshot struct {
	Owner      Owner
	CounterID  CounterID
	Value      float64
	Visibility Visibility
}

type ComboSnapshot struct {
	Owner Owner
	State ComboState
}

type StreakSnapshot struct {
	Owner Owner
	Name  string
	Count int
}
