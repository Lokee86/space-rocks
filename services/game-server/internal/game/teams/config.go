package teams

import "fmt"

func ValidateConfig(config Config) error {
	switch config.Structure {
	case StructureFFA, StructureCoOp:
		if config.AssignmentMode != "" || config.AutoTeamCount != 0 {
			return fmt.Errorf("%s does not accept assignment mode or team count", config.Structure)
		}
	case StructureCustom:
		if config.AssignmentMode != AssignmentPlayerSelected && config.AssignmentMode != AssignmentOwnerAssigned {
			return fmt.Errorf("custom requires a valid assignment mode")
		}
		if config.AutoTeamCount != 0 {
			return fmt.Errorf("custom does not accept team count")
		}
	case StructureAutoBalanced:
		if config.AutoTeamCount < 2 || config.AutoTeamCount > len(OrderedIDs()) {
			return fmt.Errorf("auto-balanced team count must be between 2 and %d", len(OrderedIDs()))
		}
		if config.AssignmentMode != "" {
			return fmt.Errorf("auto-balanced does not accept assignment mode")
		}
	default:
		return fmt.Errorf("unknown team structure %q", config.Structure)
	}
	return nil
}
