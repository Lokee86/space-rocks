package objectives

import (
	"fmt"
	"math"
)

type Condition struct {
	Kind               ConditionKind
	FactKey            string
	Target             float64
	Expected           bool
	RequiredMembers    []string
	Sequence           []string
	AllowDecrease      bool
	AllowReset         bool
	Overflow           OverflowPolicy
	AllowedAttribution []AttributionKind
}

type TimerDefinition struct {
	DurationSeconds float64
	ExpiryStatus    Status
	FailureReason   string
}

type LifecycleDefinition struct {
	Discoverable        bool
	InitiallyDiscovered bool
	InitiallyActive     bool
	Failable            bool
	Visibility          VisibilityPolicy
}

type Definition struct {
	ID              DefinitionID
	Scope           Scope
	Success         Condition
	Failure         *Condition
	Timer           *TimerDefinition
	Lifecycle       LifecycleDefinition
	AssociationKeys []string
}

func (definition Definition) Validate() error {
	if err := requireID("objective definition ID", string(definition.ID)); err != nil {
		return err
	}
	if !validScope(definition.Scope) {
		return fmt.Errorf("unsupported objective scope %q", definition.Scope)
	}
	if err := definition.Success.Validate(); err != nil {
		return fmt.Errorf("success condition: %w", err)
	}
	if definition.Failure != nil {
		if !definition.Lifecycle.Failable {
			return fmt.Errorf("failure condition requires a failable objective")
		}
		if err := definition.Failure.Validate(); err != nil {
			return fmt.Errorf("failure condition: %w", err)
		}
	}
	if err := definition.Lifecycle.validate(); err != nil {
		return err
	}
	if definition.Timer != nil {
		if err := definition.Timer.validate(definition.Lifecycle.Failable); err != nil {
			return err
		}
	}
	seen := map[string]struct{}{}
	for _, key := range definition.AssociationKeys {
		if key == "" {
			return fmt.Errorf("objective association key cannot be empty")
		}
		if _, exists := seen[key]; exists {
			return fmt.Errorf("duplicate objective association key %q", key)
		}
		seen[key] = struct{}{}
	}
	return nil
}

func (condition Condition) Validate() error {
	switch condition.Kind {
	case ConditionManual:
	case ConditionBoolean:
		if condition.FactKey == "" {
			return fmt.Errorf("boolean condition fact key is required")
		}
	case ConditionNumeric, ConditionMaintain:
		if condition.FactKey == "" {
			return fmt.Errorf("numeric condition fact key is required")
		}
		if !finitePositive(condition.Target) {
			return fmt.Errorf("numeric condition target must be finite and positive")
		}
	case ConditionSet:
		if condition.FactKey == "" || len(condition.RequiredMembers) == 0 {
			return fmt.Errorf("set condition requires a fact key and members")
		}
	case ConditionSequence:
		if len(condition.Sequence) == 0 {
			return fmt.Errorf("sequence condition requires stages")
		}
	default:
		return fmt.Errorf("unsupported objective condition kind %q", condition.Kind)
	}
	if condition.Overflow != "" && condition.Overflow != OverflowClamp && condition.Overflow != OverflowRetain {
		return fmt.Errorf("unsupported objective overflow policy %q", condition.Overflow)
	}
	for _, kind := range condition.AllowedAttribution {
		switch kind {
		case AttributionOneHit, AttributionInGame, AttributionInEncounter:
		default:
			return fmt.Errorf("unsupported attribution kind %q", kind)
		}
	}
	return nil
}

func (lifecycle LifecycleDefinition) validate() error {
	if lifecycle.InitiallyDiscovered && !lifecycle.Discoverable {
		return fmt.Errorf("only discoverable objectives may begin discovered")
	}
	switch lifecycle.Visibility {
	case "", VisibilityPublic, VisibilityOwnerOnly, VisibilityHiddenUntilDiscovered, VisibilityOwnerHiddenUntilDiscovered:
		return nil
	default:
		return fmt.Errorf("unsupported objective visibility policy %q", lifecycle.Visibility)
	}
}

func (timer TimerDefinition) validate(failable bool) error {
	if !finitePositive(timer.DurationSeconds) {
		return fmt.Errorf("objective timer duration must be finite and positive")
	}
	status := timer.ExpiryStatus
	if status == "" {
		status = StatusFailed
	}
	if status != StatusCompleted && status != StatusFailed && status != StatusCancelled {
		return fmt.Errorf("unsupported objective timer expiry status %q", status)
	}
	if status == StatusFailed && !failable {
		return fmt.Errorf("failed timer expiry requires a failable objective")
	}
	return nil
}

func finitePositive(value float64) bool {
	return value > 0 && !math.IsNaN(value) && !math.IsInf(value, 0)
}

func (definition Definition) clone() Definition {
	clone := definition
	clone.Success = definition.Success.clone()
	if definition.Failure != nil {
		failure := definition.Failure.clone()
		clone.Failure = &failure
	}
	if definition.Timer != nil {
		timer := *definition.Timer
		clone.Timer = &timer
	}
	clone.AssociationKeys = append([]string(nil), definition.AssociationKeys...)
	return clone
}

func (condition Condition) clone() Condition {
	condition.RequiredMembers = append([]string(nil), condition.RequiredMembers...)
	condition.Sequence = append([]string(nil), condition.Sequence...)
	condition.AllowedAttribution = append([]AttributionKind(nil), condition.AllowedAttribution...)
	return condition
}
