package objectives

import "fmt"

type CompositionKind string

const (
	CompositionSingle   CompositionKind = "single"
	CompositionMultiple CompositionKind = "multiple"
	CompositionMeta     CompositionKind = "meta"
)

type ComponentRef struct {
	ID   string
	Kind CompositionKind
}

type Composition struct {
	ID      string
	Kind    CompositionKind
	Members []ComponentRef
}

func (composition Composition) Validate() error {
	if composition.ID == "" {
		return fmt.Errorf("objective composition ID is required")
	}
	switch composition.Kind {
	case CompositionSingle:
		if len(composition.Members) != 0 {
			return fmt.Errorf("single objective composition cannot contain members")
		}
	case CompositionMultiple:
		if err := validateCompositionMembers(composition.Members, CompositionSingle); err != nil {
			return err
		}
	case CompositionMeta:
		if err := validateCompositionMembers(composition.Members, CompositionMultiple); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unsupported objective composition kind %q", composition.Kind)
	}
	return nil
}

func validateCompositionMembers(members []ComponentRef, expected CompositionKind) error {
	if len(members) == 0 {
		return fmt.Errorf("objective composition requires members")
	}
	seen := map[string]struct{}{}
	for _, member := range members {
		if member.ID == "" || member.Kind != expected {
			return fmt.Errorf("objective composition members must be non-empty %s references", expected)
		}
		if _, exists := seen[member.ID]; exists {
			return fmt.Errorf("duplicate objective composition member %q", member.ID)
		}
		seen[member.ID] = struct{}{}
	}
	return nil
}
